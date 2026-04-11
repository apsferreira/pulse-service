package repository

// BKL-1079: Repositório para flag_approvals — workflow de aprovação para ativação de flags.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlagApprovalRepository struct {
	db *pgxpool.Pool
}

func NewFlagApprovalRepository(db *pgxpool.Pool) *FlagApprovalRepository {
	return &FlagApprovalRepository{db: db}
}

// ListByFlag retorna todas as solicitações de aprovação de uma flag.
func (r *FlagApprovalRepository) ListByFlag(ctx context.Context, flagKey, projectID string) ([]domain.FlagApproval, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, flag_key, project_id, requested_by, reviewed_by, status, COALESCE(notes,''), payload, created_at, reviewed_at
		 FROM flag_approvals
		 WHERE flag_key = $1 AND project_id = $2::uuid
		 ORDER BY created_at DESC`,
		flagKey, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("flag_approval_repository: ListByFlag: %w", err)
	}
	defer rows.Close()

	var approvals []domain.FlagApproval
	for rows.Next() {
		a, err := scanApproval(rows)
		if err != nil {
			return nil, fmt.Errorf("flag_approval_repository: ListByFlag scan: %w", err)
		}
		approvals = append(approvals, a)
	}
	return approvals, rows.Err()
}

// ListPending retorna todas as aprovações pendentes de um projeto.
func (r *FlagApprovalRepository) ListPending(ctx context.Context, projectID string) ([]domain.FlagApproval, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, flag_key, project_id, requested_by, reviewed_by, status, COALESCE(notes,''), payload, created_at, reviewed_at
		 FROM flag_approvals
		 WHERE project_id = $1::uuid AND status = 'pending_approval'
		 ORDER BY created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("flag_approval_repository: ListPending: %w", err)
	}
	defer rows.Close()

	var approvals []domain.FlagApproval
	for rows.Next() {
		a, err := scanApproval(rows)
		if err != nil {
			return nil, fmt.Errorf("flag_approval_repository: ListPending scan: %w", err)
		}
		approvals = append(approvals, a)
	}
	return approvals, rows.Err()
}

// Create cria uma nova solicitação de aprovação.
func (r *FlagApprovalRepository) Create(ctx context.Context, flagKey, projectID string, req domain.CreateApprovalRequest) (*domain.FlagApproval, error) {
	payloadJSON, _ := json.Marshal(req.Payload)

	a := domain.FlagApproval{
		ID:          uuid.New().String(),
		FlagKey:     flagKey,
		ProjectID:   projectID,
		RequestedBy: req.RequestedBy,
		Status:      domain.ApprovalStatusPendingApproval,
		Notes:       req.Notes,
		Payload:     req.Payload,
		CreatedAt:   time.Now().UTC(),
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO flag_approvals (id, flag_key, project_id, requested_by, status, notes, payload, created_at)
		 VALUES ($1, $2, $3::uuid, $4, 'pending_approval', $5, $6, $7)`,
		a.ID, a.FlagKey, a.ProjectID, a.RequestedBy, a.Notes, payloadJSON, a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("flag_approval_repository: Create: %w", err)
	}
	return &a, nil
}

// Review atualiza o status de uma aprovação (approved | rejected).
func (r *FlagApprovalRepository) Review(ctx context.Context, approvalID, projectID string, req domain.ReviewApprovalRequest) (*domain.FlagApproval, error) {
	if req.Status != domain.ApprovalStatusApproved && req.Status != domain.ApprovalStatusRejected {
		return nil, fmt.Errorf("flag_approval_repository: Review: status deve ser 'approved' ou 'rejected'")
	}

	now := time.Now().UTC()
	row := r.db.QueryRow(ctx,
		`UPDATE flag_approvals
		 SET reviewed_by = $1, status = $2, notes = $3, reviewed_at = $4
		 WHERE id = $5::uuid AND project_id = $6::uuid AND status = 'pending_approval'
		 RETURNING id, flag_key, project_id, requested_by, reviewed_by, status, COALESCE(notes,''), payload, created_at, reviewed_at`,
		req.ReviewedBy, string(req.Status), req.Notes, now, approvalID, projectID,
	)

	var a domain.FlagApproval
	var payloadBytes []byte
	var reviewedBy *string
	var reviewedAt *time.Time
	if err := row.Scan(&a.ID, &a.FlagKey, &a.ProjectID, &a.RequestedBy, &reviewedBy,
		&a.Status, &a.Notes, &payloadBytes, &a.CreatedAt, &reviewedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("flag_approval_repository: Review: %w", err)
	}
	a.ReviewedBy = reviewedBy
	a.ReviewedAt = reviewedAt
	if len(payloadBytes) > 0 {
		_ = json.Unmarshal(payloadBytes, &a.Payload)
	}
	return &a, nil
}

func scanApproval(rows pgx.Rows) (domain.FlagApproval, error) {
	var a domain.FlagApproval
	var payloadBytes []byte
	var reviewedBy *string
	var reviewedAt *time.Time
	err := rows.Scan(&a.ID, &a.FlagKey, &a.ProjectID, &a.RequestedBy, &reviewedBy,
		&a.Status, &a.Notes, &payloadBytes, &a.CreatedAt, &reviewedAt)
	if err != nil {
		return a, err
	}
	a.ReviewedBy = reviewedBy
	a.ReviewedAt = reviewedAt
	if len(payloadBytes) > 0 {
		_ = json.Unmarshal(payloadBytes, &a.Payload)
	}
	return a, nil
}
