package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/detect-price-by-photo/backend/internal/ai/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrModelNotFound = errors.New("model not found")
	ErrNoDefaultModel = errors.New("no default model configured")
)

// AIModelRepositoryInterface defines the contract for AI model data access.
type AIModelRepositoryInterface interface {
	GetCatalog(ctx context.Context) ([]*models.AIModel, error)
	GetByID(ctx context.Context, id string) (*models.AIModel, error)
	GetDefault(ctx context.Context) (*models.AIModel, error)
	SetDefault(ctx context.Context, modelID string) error
	Update(ctx context.Context, model *models.AIModel) error
}

// AIModelRepository is an implementation of AIModelRepositoryInterface using pgxpool.
type AIModelRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewAIModelRepository creates a new AIModelRepository.
func NewAIModelRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *AIModelRepository {
	return &AIModelRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// GetCatalog returns all AI models.
func (r *AIModelRepository) GetCatalog(ctx context.Context) ([]*models.AIModel, error) {
	query := `
		SELECT id, display_name, provider, cost_per_1k_tokens_usd, average_latency_ms,
		       modes_supported, fallback_chain, status, is_default, max_tokens, temperature,
		       created_at, updated_at
		FROM ` + models.AIModelTable + `
		ORDER BY is_default DESC, display_name ASC
	`

	rows, err := r.pgPool.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to get AI models catalog", "error", err)
		return nil, err
	}
	defer rows.Close()

	var aiModels []*models.AIModel
	for rows.Next() {
		model := &models.AIModel{}
		err := rows.Scan(
			&model.ID,
			&model.DisplayName,
			&model.Provider,
			&model.CostPer1kTokensUSD,
			&model.AverageLatencyMs,
			&model.ModesSupported,
			&model.FallbackChain,
			&model.Status,
			&model.IsDefault,
			&model.MaxTokens,
			&model.Temperature,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan AI model row", "error", err)
			continue
		}
		aiModels = append(aiModels, model)
	}

	return aiModels, nil
}

// GetByID returns an AI model by ID.
func (r *AIModelRepository) GetByID(ctx context.Context, id string) (*models.AIModel, error) {
	query := `
		SELECT id, display_name, provider, cost_per_1k_tokens_usd, average_latency_ms,
		       modes_supported, fallback_chain, status, is_default, max_tokens, temperature,
		       created_at, updated_at
		FROM ` + models.AIModelTable + `
		WHERE id = $1
	`

	row := r.pgPool.QueryRow(ctx, query, id)
	model := &models.AIModel{}
	err := row.Scan(
		&model.ID,
		&model.DisplayName,
		&model.Provider,
		&model.CostPer1kTokensUSD,
		&model.AverageLatencyMs,
		&model.ModesSupported,
		&model.FallbackChain,
		&model.Status,
		&model.IsDefault,
		&model.MaxTokens,
		&model.Temperature,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModelNotFound
		}
		r.logger.Error("failed to get AI model by ID", "id", id, "error", err)
		return nil, err
	}

	return model, nil
}

// GetDefault returns the default AI model.
func (r *AIModelRepository) GetDefault(ctx context.Context) (*models.AIModel, error) {
	query := `
		SELECT id, display_name, provider, cost_per_1k_tokens_usd, average_latency_ms,
		       modes_supported, fallback_chain, status, is_default, max_tokens, temperature,
		       created_at, updated_at
		FROM ` + models.AIModelTable + `
		WHERE is_default = TRUE AND status = 'active'
		LIMIT 1
	`

	row := r.pgPool.QueryRow(ctx, query)
	model := &models.AIModel{}
	err := row.Scan(
		&model.ID,
		&model.DisplayName,
		&model.Provider,
		&model.CostPer1kTokensUSD,
		&model.AverageLatencyMs,
		&model.ModesSupported,
		&model.FallbackChain,
		&model.Status,
		&model.IsDefault,
		&model.MaxTokens,
		&model.Temperature,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoDefaultModel
		}
		r.logger.Error("failed to get default AI model", "error", err)
		return nil, err
	}

	return model, nil
}

// SetDefault sets a model as the default (and unsets the previous default).
func (r *AIModelRepository) SetDefault(ctx context.Context, modelID string) error {
	tx, err := r.pgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		r.logger.Error("failed to begin transaction", "error", err)
		return err
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			r.logger.Warn("failed to rollback transaction", "error", rbErr)
		}
	}()

	// Unset all defaults
	_, err = tx.Exec(ctx, `UPDATE `+models.AIModelTable+` SET is_default = FALSE WHERE is_default = TRUE`)
	if err != nil {
		r.logger.Error("failed to unset defaults", "error", err)
		return err
	}

	// Set new default
	_, err = tx.Exec(ctx, `UPDATE `+models.AIModelTable+` SET is_default = TRUE WHERE id = $1`, modelID)
	if err != nil {
		r.logger.Error("failed to set default model", "model_id", modelID, "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("failed to commit transaction", "error", err)
		return err
	}

	r.logger.Info("default model updated", "model_id", modelID)
	return nil
}

// Update updates an AI model configuration.
func (r *AIModelRepository) Update(ctx context.Context, model *models.AIModel) error {
	query := `
		UPDATE ` + models.AIModelTable + `
		SET display_name = $1, provider = $2, cost_per_1k_tokens_usd = $3, average_latency_ms = $4,
		    modes_supported = $5, fallback_chain = $6, status = $7, max_tokens = $8, temperature = $9,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
	`

	cmd, err := r.pgPool.Exec(ctx, query,
		model.DisplayName,
		model.Provider,
		model.CostPer1kTokensUSD,
		model.AverageLatencyMs,
		model.ModesSupported,
		model.FallbackChain,
		model.Status,
		model.MaxTokens,
		model.Temperature,
		model.ID,
	)
	if err != nil {
		r.logger.Error("failed to update AI model", "model_id", model.ID, "error", err)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrModelNotFound
	}

	r.logger.Info("AI model updated", "model_id", model.ID)
	return nil
}

