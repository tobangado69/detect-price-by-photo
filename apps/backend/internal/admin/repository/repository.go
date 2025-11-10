package repository

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/detect-price-by-photo/backend/internal/admin/models"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLogRepository handles audit log operations.
type AuditLogRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *AuditLogRepository {
	return &AuditLogRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// CreateAuditLog creates a new audit log entry.
func (r *AuditLogRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.Must(uuid.NewV7())
	}

	changesJSON, _ := json.Marshal(log.Changes)
	if changesJSON == nil {
		changesJSON = []byte("{}")
	}

	query := `
		INSERT INTO ` + models.AuditLogTable + ` (
			id, admin_id, action, target_type, target_id, changes, ip_address, user_agent, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.pgPool.Exec(ctx, query,
		log.ID,
		log.AdminID,
		log.Action,
		log.TargetType,
		log.TargetID,
		changesJSON,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create audit log", "error", err)
		return err
	}

	return nil
}

