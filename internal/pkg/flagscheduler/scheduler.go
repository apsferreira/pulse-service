// Package flagscheduler implementa o worker de ativação/desativação automática de feature flags.
// BKL-1040: flags com starts_at/ends_at são processadas a cada 60 segundos.
package flagscheduler

import (
	"context"
	"log"
	"time"
)

// Activator define a interface mínima necessária pelo scheduler.
type Activator interface {
	ActivateScheduledFlags(ctx context.Context) (int, int, error)
}

// Start inicia o goroutine do scheduler de flags.
// Chama repo.ActivateScheduledFlags a cada `interval` e loga o resultado.
// O contexto cancelado para o loop.
func Start(ctx context.Context, repo Activator, interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}

	log.Printf("[flag-scheduler] iniciado — intervalo: %s", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Executa imediatamente na inicialização
	runOnce(ctx, repo)

	for {
		select {
		case <-ctx.Done():
			log.Println("[flag-scheduler] parado")
			return
		case <-ticker.C:
			runOnce(ctx, repo)
		}
	}
}

func runOnce(ctx context.Context, repo Activator) {
	activated, deactivated, err := repo.ActivateScheduledFlags(ctx)
	if err != nil {
		log.Printf("[flag-scheduler] erro ao processar flags agendadas: %v", err)
		return
	}
	if activated > 0 || deactivated > 0 {
		log.Printf("[flag-scheduler] ativadas=%d desativadas=%d", activated, deactivated)
	}
}
