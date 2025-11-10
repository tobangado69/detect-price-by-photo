package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/detect-price-by-photo/backend/internal/subscription/models"
)

// Sentinel errors
var (
	ErrPlanNotFound      = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// PlanRepositoryInterface defines the contract for plan data access.
type PlanRepositoryInterface interface {
	GetPlans(ctx context.Context, activeOnly bool) ([]*models.Plan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*models.Plan, error)
	GetPlanByName(ctx context.Context, name string) (*models.Plan, error)
}

// SubscriptionRepositoryInterface defines the contract for subscription data access.
type SubscriptionRepositoryInterface interface {
	GetUserSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	CreateSubscription(ctx context.Context, subscription *models.Subscription) error
	UpdateSubscription(ctx context.Context, subscription *models.Subscription) error
	IncrementUsage(ctx context.Context, userID uuid.UUID) error
	ResetDailyUsage(ctx context.Context) error
}

// Ensure repositories implement interfaces
var _ PlanRepositoryInterface = (*PlanRepository)(nil)
var _ SubscriptionRepositoryInterface = (*SubscriptionRepository)(nil)

// PlanRepository is an implementation of PlanRepositoryInterface using pgxpool.
type PlanRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// SubscriptionRepository is an implementation of SubscriptionRepositoryInterface using pgxpool.
type SubscriptionRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewPlanRepository creates a new PlanRepository with pgxpool and slog logger.
func NewPlanRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *PlanRepository {
	return &PlanRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// NewSubscriptionRepository creates a new SubscriptionRepository with pgxpool and slog logger.
func NewSubscriptionRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// GetPlans retrieves all plans, optionally filtering by active status.
func (r *PlanRepository) GetPlans(ctx context.Context, activeOnly bool) ([]*models.Plan, error) {
	query := `
		SELECT id, name, daily_photo_limit, price_idr, price_usd, description, features, is_active, created_at, updated_at
		FROM public.plans
	`
	args := []interface{}{}
	
	if activeOnly {
		query += " WHERE is_active = true"
	}
	
	query += " ORDER BY price_idr ASC"

	rows, err := r.pgPool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to query plans", "error", err)
		return nil, err
	}
	defer rows.Close()

	var plans []*models.Plan
	for rows.Next() {
		var plan models.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.DailyPhotoLimit,
			&plan.PriceIDR,
			&plan.PriceUSD,
			&plan.Description,
			&plan.Features,
			&plan.IsActive,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan plan", "error", err)
			return nil, err
		}
		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating plans", "error", err)
		return nil, err
	}

	return plans, nil
}

// GetPlanByID retrieves a plan by ID.
func (r *PlanRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	query := `
		SELECT id, name, daily_photo_limit, price_idr, price_usd, description, features, is_active, created_at, updated_at
		FROM public.plans
		WHERE id = $1
	`

	var plan models.Plan
	err := r.pgPool.QueryRow(ctx, query, id).Scan(
		&plan.ID,
		&plan.Name,
		&plan.DailyPhotoLimit,
		&plan.PriceIDR,
		&plan.PriceUSD,
		&plan.Description,
		&plan.Features,
		&plan.IsActive,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		r.logger.Error("failed to get plan by ID", "id", id, "error", err)
		return nil, err
	}

	return &plan, nil
}

// GetPlanByName retrieves a plan by name.
func (r *PlanRepository) GetPlanByName(ctx context.Context, name string) (*models.Plan, error) {
	query := `
		SELECT id, name, daily_photo_limit, price_idr, price_usd, description, features, is_active, created_at, updated_at
		FROM public.plans
		WHERE LOWER(name) = LOWER($1)
	`

	var plan models.Plan
	err := r.pgPool.QueryRow(ctx, query, name).Scan(
		&plan.ID,
		&plan.Name,
		&plan.DailyPhotoLimit,
		&plan.PriceIDR,
		&plan.PriceUSD,
		&plan.Description,
		&plan.Features,
		&plan.IsActive,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		r.logger.Error("failed to get plan by name", "name", name, "error", err)
		return nil, err
	}

	return &plan, nil
}

// GetUserSubscription retrieves the active subscription for a user.
func (r *SubscriptionRepository) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, plan_id, status, daily_photo_limit, current_day_usage, usage_reset_at,
		       started_at, current_period_start, current_period_end, cancelled_at, created_at, updated_at
		FROM public.subscriptions
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var subscription models.Subscription
	err := r.pgPool.QueryRow(ctx, query, userID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.Status,
		&subscription.DailyPhotoLimit,
		&subscription.CurrentDayUsage,
		&subscription.UsageResetAt,
		&subscription.StartedAt,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		r.logger.Error("failed to get user subscription", "user_id", userID, "error", err)
		return nil, err
	}

	return &subscription, nil
}

// CreateSubscription creates a new subscription.
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, subscription *models.Subscription) error {
	query := `
		INSERT INTO public.subscriptions (
			id, user_id, plan_id, status, daily_photo_limit, current_day_usage, usage_reset_at,
			started_at, current_period_start, current_period_end, cancelled_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.pgPool.Exec(ctx, query,
		subscription.ID,
		subscription.UserID,
		subscription.PlanID,
		subscription.Status,
		subscription.DailyPhotoLimit,
		subscription.CurrentDayUsage,
		subscription.UsageResetAt,
		subscription.StartedAt,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelledAt,
	)

	if err != nil {
		r.logger.Error("failed to create subscription", "error", err)
		return err
	}

	return nil
}

// UpdateSubscription updates an existing subscription.
func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, subscription *models.Subscription) error {
	query := `
		UPDATE public.subscriptions
		SET plan_id = $2, status = $3, daily_photo_limit = $4, current_day_usage = $5,
		    usage_reset_at = $6, current_period_start = $7, current_period_end = $8,
		    cancelled_at = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pgPool.Exec(ctx, query,
		subscription.ID,
		subscription.PlanID,
		subscription.Status,
		subscription.DailyPhotoLimit,
		subscription.CurrentDayUsage,
		subscription.UsageResetAt,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelledAt,
	)

	if err != nil {
		r.logger.Error("failed to update subscription", "id", subscription.ID, "error", err)
		return err
	}

	return nil
}

// IncrementUsage increments the daily usage counter for a user's subscription.
func (r *SubscriptionRepository) IncrementUsage(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE public.subscriptions
		SET current_day_usage = current_day_usage + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND status = 'active'
	`

	result, err := r.pgPool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to increment usage", "user_id", userID, "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

// ResetDailyUsage resets daily usage for all subscriptions where usage_reset_at has passed.
// This should be called by a cron job daily at 00:00 UTC+7.
func (r *SubscriptionRepository) ResetDailyUsage(ctx context.Context) error {
	query := `
		UPDATE public.subscriptions
		SET current_day_usage = 0,
		    usage_reset_at = (CURRENT_DATE + INTERVAL '1 day' AT TIME ZONE 'Asia/Jakarta'),
		    updated_at = CURRENT_TIMESTAMP
		WHERE usage_reset_at <= CURRENT_TIMESTAMP AND status = 'active'
	`

	result, err := r.pgPool.Exec(ctx, query)
	if err != nil {
		r.logger.Error("failed to reset daily usage", "error", err)
		return err
	}

	r.logger.Info("reset daily usage", "rows_affected", result.RowsAffected())
	return nil
}

