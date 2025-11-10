package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/detect-price-by-photo/backend/internal/photo/models"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors
var (
	ErrAnalysisNotFound = errors.New("analysis not found")
)

// AnalysisRepositoryInterface defines the contract for analysis data access.
type AnalysisRepositoryInterface interface {
	CreateAnalysis(ctx context.Context, analysis *models.Analysis) error
	GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error)
	ListUserAnalyses(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Analysis, int, error)
	UpdateAnalysis(ctx context.Context, analysis *models.Analysis) error
	DeleteAnalysis(ctx context.Context, id uuid.UUID) error
}

// Ensure AnalysisRepository implements AnalysisRepositoryInterface
var _ AnalysisRepositoryInterface = (*AnalysisRepository)(nil)

// AnalysisRepository is an implementation of AnalysisRepositoryInterface using pgxpool.
type AnalysisRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewAnalysisRepository creates a new AnalysisRepository with pgxpool and slog logger.
func NewAnalysisRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *AnalysisRepository {
	return &AnalysisRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// CreateAnalysis creates a new analysis record.
func (r *AnalysisRepository) CreateAnalysis(ctx context.Context, analysis *models.Analysis) error {
	if analysis.ID == uuid.Nil {
		analysis.ID = uuid.Must(uuid.NewV7())
	}
	analysis.CreatedAt = time.Now()

	query := `
		INSERT INTO public.price_analyses (
			id, user_id, image_url, product_name, condition, mode, status, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.pgPool.Exec(ctx, query,
		analysis.ID,
		analysis.UserID,
		analysis.ImageURL,
		analysis.ProductName,
		string(analysis.Condition),
		string(analysis.Mode),
		string(analysis.Status),
		analysis.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create analysis", "error", err, "user_id", analysis.UserID)
		return err
	}

	r.logger.Info("analysis created", "analysis_id", analysis.ID, "user_id", analysis.UserID)
	return nil
}

// GetAnalysis retrieves an analysis by ID.
func (r *AnalysisRepository) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	query := `
		SELECT id, user_id, image_url, product_name, condition, mode,
		       estimated_price_min, estimated_price_max, estimated_price_median,
		       confidence_score, reasoning, processing_time_ms, cost_in_usd,
		       model_used, status, error_message, created_at, updated_at, deleted_at
		FROM public.price_analyses
		WHERE id = $1 AND deleted_at IS NULL
	`

	var analysis models.Analysis
	err := r.pgPool.QueryRow(ctx, query, id).Scan(
		&analysis.ID,
		&analysis.UserID,
		&analysis.ImageURL,
		&analysis.ProductName,
		&analysis.Condition,
		&analysis.Mode,
		&analysis.EstimatedPriceMin,
		&analysis.EstimatedPriceMax,
		&analysis.EstimatedPriceMedian,
		&analysis.ConfidenceScore,
		&analysis.Reasoning,
		&analysis.ProcessingTimeMs,
		&analysis.CostInUSD,
		&analysis.ModelUsed,
		&analysis.Status,
		&analysis.ErrorMessage,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
		&analysis.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAnalysisNotFound
		}
		r.logger.Error("failed to get analysis", "id", id, "error", err)
		return nil, err
	}

	return &analysis, nil
}

// ListUserAnalyses retrieves paginated analyses for a user.
func (r *AnalysisRepository) ListUserAnalyses(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Analysis, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM public.price_analyses
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	var total int
	err := r.pgPool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count analyses", "user_id", userID, "error", err)
		return nil, 0, err
	}

	// Get analyses
	query := `
		SELECT id, user_id, image_url, product_name, condition, mode,
		       estimated_price_min, estimated_price_max, estimated_price_median,
		       confidence_score, reasoning, processing_time_ms, cost_in_usd,
		       model_used, status, error_message, created_at, updated_at, deleted_at
		FROM public.price_analyses
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pgPool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("failed to list analyses", "user_id", userID, "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var analyses []*models.Analysis
	for rows.Next() {
		var analysis models.Analysis
		err := rows.Scan(
			&analysis.ID,
			&analysis.UserID,
			&analysis.ImageURL,
			&analysis.ProductName,
			&analysis.Condition,
			&analysis.Mode,
			&analysis.EstimatedPriceMin,
			&analysis.EstimatedPriceMax,
			&analysis.EstimatedPriceMedian,
			&analysis.ConfidenceScore,
			&analysis.Reasoning,
			&analysis.ProcessingTimeMs,
			&analysis.CostInUSD,
			&analysis.ModelUsed,
			&analysis.Status,
			&analysis.ErrorMessage,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
			&analysis.DeletedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan analysis", "error", err)
			return nil, 0, err
		}
		analyses = append(analyses, &analysis)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating analyses", "error", err)
		return nil, 0, err
	}

	return analyses, total, nil
}

// UpdateAnalysis updates an existing analysis record.
func (r *AnalysisRepository) UpdateAnalysis(ctx context.Context, analysis *models.Analysis) error {
	query := `
		UPDATE public.price_analyses
		SET estimated_price_min = $2, estimated_price_max = $3, estimated_price_median = $4,
		    confidence_score = $5, reasoning = $6, processing_time_ms = $7,
		    cost_in_usd = $8, model_used = $9, status = $10, error_message = $11,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pgPool.Exec(ctx, query,
		analysis.ID,
		analysis.EstimatedPriceMin,
		analysis.EstimatedPriceMax,
		analysis.EstimatedPriceMedian,
		analysis.ConfidenceScore,
		analysis.Reasoning,
		analysis.ProcessingTimeMs,
		analysis.CostInUSD,
		analysis.ModelUsed,
		string(analysis.Status),
		analysis.ErrorMessage,
	)

	if err != nil {
		r.logger.Error("failed to update analysis", "id", analysis.ID, "error", err)
		return err
	}

	return nil
}

// DeleteAnalysis soft deletes an analysis.
func (r *AnalysisRepository) DeleteAnalysis(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE public.price_analyses
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pgPool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete analysis", "id", id, "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAnalysisNotFound
	}

	return nil
}

