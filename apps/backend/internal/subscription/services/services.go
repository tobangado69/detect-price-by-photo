package services

import (
	"context"
	"errors"
	"time"

	"github.com/detect-price-by-photo/backend/internal/subscription/models"
	"github.com/detect-price-by-photo/backend/internal/subscription/repository"

	"github.com/gofrs/uuid/v5"
)

// SubscriptionServiceInterface defines the contract for subscription business logic.
type SubscriptionServiceInterface interface {
	ListPlans(ctx context.Context, activeOnly bool) ([]*models.Plan, error)
	GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error)
	GetCurrentSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	CreateSubscription(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription *models.Subscription) error
	CheckQuota(ctx context.Context, userID uuid.UUID) (bool, error)
	IncrementUsage(ctx context.Context, userID uuid.UUID) error
	ResetDailyUsage(ctx context.Context) error
	GetPlanByName(ctx context.Context, name string) (uuid.UUID, error) // For auto-enrollment
}

// Ensure SubscriptionService implements SubscriptionServiceInterface
var _ SubscriptionServiceInterface = (*SubscriptionService)(nil)

// SubscriptionService implements subscription business logic.
type SubscriptionService struct {
	planRepo         repository.PlanRepositoryInterface
	subscriptionRepo repository.SubscriptionRepositoryInterface
}

type SubscriptionServiceOpts struct {
	PlanRepo         repository.PlanRepositoryInterface
	SubscriptionRepo repository.SubscriptionRepositoryInterface
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(opts SubscriptionServiceOpts) *SubscriptionService {
	return &SubscriptionService{
		planRepo:         opts.PlanRepo,
		subscriptionRepo: opts.SubscriptionRepo,
	}
}

// ListPlans retrieves all plans, optionally filtering by active status.
func (s *SubscriptionService) ListPlans(ctx context.Context, activeOnly bool) ([]*models.Plan, error) {
	return s.planRepo.GetPlans(ctx, activeOnly)
}

// GetCurrentSubscription retrieves the active subscription for a user.
func (s *SubscriptionService) GetCurrentSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	subscription, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			return nil, errors.New("no active subscription found")
		}
		return nil, err
	}

	// Check if usage needs to be reset (usage_reset_at has passed)
	now := time.Now()
	if subscription.UsageResetAt.Before(now) {
		// Reset usage and update reset time
		subscription.CurrentDayUsage = 0
		// Set reset time to next day at 00:00 UTC+7 (Asia/Jakarta)
		loc, _ := time.LoadLocation("Asia/Jakarta")
		tomorrow := now.In(loc).AddDate(0, 0, 1)
		subscription.UsageResetAt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
		
		if err := s.subscriptionRepo.UpdateSubscription(ctx, subscription); err != nil {
			return nil, err
		}
	}

	return subscription, nil
}

// CreateSubscription creates a new subscription for a user.
// Returns the created subscription, or error if creation fails.
// For auto-enrollment use case, the caller can ignore the return value.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*models.Subscription, error) {
	// Get plan details
	plan, err := s.planRepo.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}

	// Check if user already has an active subscription
	existing, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err == nil && existing != nil {
		// Cancel existing subscription
		now := time.Now()
		existing.Status = "cancelled"
		existing.CancelledAt = &now
		if err := s.subscriptionRepo.UpdateSubscription(ctx, existing); err != nil {
			return nil, err
		}
	}

	// Create new subscription
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := now.In(loc).AddDate(0, 0, 1)
	usageResetAt := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
	periodEnd := now.AddDate(0, 0, 30) // 30 days from now

	subscription := &models.Subscription{
		ID:                uuid.Must(uuid.NewV7()),
		UserID:            userID,
		PlanID:            planID,
		Status:            "active",
		DailyPhotoLimit:   plan.DailyPhotoLimit,
		CurrentDayUsage:   0,
		UsageResetAt:      usageResetAt,
		StartedAt:         now,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:  periodEnd,
		CancelledAt:       nil,
	}

	if err := s.subscriptionRepo.CreateSubscription(ctx, subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

// CheckQuota checks if a user has quota remaining.
func (s *SubscriptionService) CheckQuota(ctx context.Context, userID uuid.UUID) (bool, error) {
	subscription, err := s.GetCurrentSubscription(ctx, userID)
	if err != nil {
		return false, err
	}

	return subscription.HasQuotaRemaining(), nil
}

// IncrementUsage increments the daily usage counter for a user.
func (s *SubscriptionService) IncrementUsage(ctx context.Context, userID uuid.UUID) error {
	// First check quota
	hasQuota, err := s.CheckQuota(ctx, userID)
	if err != nil {
		return err
	}
	if !hasQuota {
		return errors.New("quota exceeded")
	}

	// Increment usage
	return s.subscriptionRepo.IncrementUsage(ctx, userID)
}

// ResetDailyUsage resets daily usage for all subscriptions (called by cron).
func (s *SubscriptionService) ResetDailyUsage(ctx context.Context) error {
	return s.subscriptionRepo.ResetDailyUsage(ctx)
}

// GetPlanByName retrieves a plan by name and returns its ID (for auto-enrollment).
func (s *SubscriptionService) GetPlanByName(ctx context.Context, name string) (uuid.UUID, error) {
	plan, err := s.planRepo.GetPlanByName(ctx, name)
	if err != nil {
		return uuid.Nil, err
	}
	return plan.ID, nil
}

// GetPlanByID retrieves a plan by ID.
func (s *SubscriptionService) GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error) {
	return s.planRepo.GetPlanByID(ctx, planID)
}

// UpdateSubscription updates an existing subscription.
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, subscription *models.Subscription) error {
	return s.subscriptionRepo.UpdateSubscription(ctx, subscription)
}

