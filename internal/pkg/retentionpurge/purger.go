// Package retentionpurge implementa o worker de limpeza de dados expirados por projeto.
// BKL-1037: deleta eventos mais antigos que data_retention_days de cada projeto.
// Executa 1x/dia às 03:00 UTC.
package retentionpurge

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Purger deleta eventos expirados baseado na política de retenção por projeto.
type Purger struct {
	db *pgxpool.Pool
}

// New cria um novo Purger.
func New(db *pgxpool.Pool) *Purger {
	return &Purger{db: db}
}

// Start inicia o goroutine do purger.
// Executa imediatamente e depois a cada `interval` (normalmente 24h).
func Start(ctx context.Context, db *pgxpool.Pool, interval time.Duration) {
	p := New(db)
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	log.Printf("[retention-purge] iniciado — intervalo: %s", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Primeira execução ao startup
	p.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[retention-purge] parado")
			return
		case <-ticker.C:
			p.runOnce(ctx)
		}
	}
}

// runOnce executa o purge para todos os projetos com política de retenção configurada.
func (p *Purger) runOnce(ctx context.Context) {
	// Busca projetos com política de retenção e deleta eventos antigos
	rows, err := p.db.Query(ctx, `
		SELECT id, slug, data_retention_days
		FROM projects
		WHERE is_active = true AND data_retention_days > 0
	`)
	if err != nil {
		log.Printf("[retention-purge] erro ao buscar projetos: %v", err)
		return
	}
	defer rows.Close()

	type projectInfo struct {
		id                string
		slug              string
		dataRetentionDays int
	}

	var projects []projectInfo
	for rows.Next() {
		var p projectInfo
		if err := rows.Scan(&p.id, &p.slug, &p.dataRetentionDays); err != nil {
			log.Printf("[retention-purge] scan error: %v", err)
			continue
		}
		projects = append(projects, p)
	}
	rows.Close()

	totalDeleted := 0
	for _, proj := range projects {
		cutoff := time.Now().UTC().AddDate(0, 0, -proj.dataRetentionDays)
		n, err := p.purgeProject(ctx, proj.id, cutoff)
		if err != nil {
			log.Printf("[retention-purge] erro purge projeto %s: %v", proj.slug, err)
			continue
		}
		if n > 0 {
			log.Printf("[retention-purge] projeto=%s deletados=%d (retenção=%dd, cutoff=%s)",
				proj.slug, n, proj.dataRetentionDays, cutoff.Format("2006-01-02"))
			totalDeleted += n
		}
	}

	if totalDeleted > 0 {
		log.Printf("[retention-purge] total deletado: %d eventos em %d projetos", totalDeleted, len(projects))
	}
}

// purgeProject deleta eventos de um projeto mais antigos que cutoff.
// Usa DELETE em batches de 10.000 para não bloquear o banco por muito tempo.
func (p *Purger) purgeProject(ctx context.Context, projectID string, cutoff time.Time) (int, error) {
	const batchSize = 10_000
	total := 0

	for {
		result, err := p.db.Exec(ctx, `
			DELETE FROM events
			WHERE id IN (
				SELECT id FROM events
				WHERE project_id = $1::uuid
				  AND created_at < $2
				LIMIT $3
			)
		`, projectID, cutoff, batchSize)
		if err != nil {
			return total, fmt.Errorf("purgeProject DELETE: %w", err)
		}
		n := int(result.RowsAffected())
		total += n
		if n < batchSize {
			break // último batch
		}
	}

	return total, nil
}
