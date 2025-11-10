package repository

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/detect-price-by-photo/backend/internal/subscription/models"
	"github.com/detect-price-by-photo/backend/internal/utils/testutils"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*testutils.TestEnv, func()) {
	te := testutils.NewTestEnv(t)
	pgPool, _, err := te.SetupPostgres()
	require.NoError(t, err)
	te.SetupConfig()
	te.RunAppMigrations()
	return te, func() { pgPool.Close() }
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestPlanRepository_GetPlans(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlanRepository(te.PGPool, newTestLogger())

	t.Run("Get all plans", func(t *testing.T) {
		plans, err := repo.GetPlans(ctx, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(plans), 3) // Should have at least Free, Premium, Enterprise
	})

	t.Run("Get active plans only", func(t *testing.T) {
		plans, err := repo.GetPlans(ctx, true)
		require.NoError(t, err)
		for _, plan := range plans {
			assert.True(t, plan.IsActive)
		}
	})
}

func TestPlanRepository_GetPlanByID(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlanRepository(te.PGPool, newTestLogger())

	t.Run("Get existing plan", func(t *testing.T) {
		// Get Free plan by name first
		freePlan, err := repo.GetPlanByName(ctx, "free")
		require.NoError(t, err)

		plan, err := repo.GetPlanByID(ctx, freePlan.ID)
		require.NoError(t, err)
		assert.Equal(t, "free", plan.Name)
		assert.Equal(t, 10, plan.DailyPhotoLimit)
	})

	t.Run("Get non-existent plan", func(t *testing.T) {
		fakeID := uuid.Must(uuid.NewV7())
		plan, err := repo.GetPlanByID(ctx, fakeID)
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.ErrorIs(t, err, ErrPlanNotFound)
	})
}

func TestPlanRepository_GetPlanByName(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlanRepository(te.PGPool, newTestLogger())

	t.Run("Get Free plan", func(t *testing.T) {
		plan, err := repo.GetPlanByName(ctx, "free")
		require.NoError(t, err)
		assert.Equal(t, "free", plan.Name)
		assert.Equal(t, 10, plan.DailyPhotoLimit)
	})

	t.Run("Get Premium plan", func(t *testing.T) {
		plan, err := repo.GetPlanByName(ctx, "premium")
		require.NoError(t, err)
		assert.Equal(t, "premium", plan.Name)
		assert.Equal(t, 1000, plan.DailyPhotoLimit)
	})

	t.Run("Get non-existent plan", func(t *testing.T) {
		plan, err := repo.GetPlanByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.ErrorIs(t, err, ErrPlanNotFound)
	})
}

func TestSubscriptionRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := NewSubscriptionRepository(te.PGPool, newTestLogger())

	// Get Free plan
	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := now.In(loc).AddDate(0, 0, 1)
	usageResetAt := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)

	subscription := &models.Subscription{
		ID:                 uuid.Must(uuid.NewV7()),
		UserID:             userID,
		PlanID:             freePlan.ID,
		Status:             "active",
		DailyPhotoLimit:    freePlan.DailyPhotoLimit,
		CurrentDayUsage:    0,
		UsageResetAt:       usageResetAt,
		StartedAt:          now,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 0, 30),
		CancelledAt:        nil,
	}

	t.Run("Create subscription", func(t *testing.T) {
		err := subRepo.CreateSubscription(ctx, subscription)
		require.NoError(t, err)
	})

	t.Run("Get user subscription", func(t *testing.T) {
		sub, err := subRepo.GetUserSubscription(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, userID, sub.UserID)
		assert.Equal(t, freePlan.ID, sub.PlanID)
		assert.Equal(t, "active", sub.Status)
		assert.Equal(t, 0, sub.CurrentDayUsage)
	})

	t.Run("Get non-existent subscription", func(t *testing.T) {
		fakeUserID := uuid.Must(uuid.NewV7())
		sub, err := subRepo.GetUserSubscription(ctx, fakeUserID)
		assert.Error(t, err)
		assert.Nil(t, sub)
		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

func TestSubscriptionRepository_IncrementUsage(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := NewSubscriptionRepository(te.PGPool, newTestLogger())

	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := now.In(loc).AddDate(0, 0, 1)
	usageResetAt := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)

	subscription := &models.Subscription{
		ID:                 uuid.Must(uuid.NewV7()),
		UserID:             userID,
		PlanID:             freePlan.ID,
		Status:             "active",
		DailyPhotoLimit:    freePlan.DailyPhotoLimit,
		CurrentDayUsage:    0,
		UsageResetAt:       usageResetAt,
		StartedAt:          now,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 0, 30),
	}

	err = subRepo.CreateSubscription(ctx, subscription)
	require.NoError(t, err)

	t.Run("Increment usage", func(t *testing.T) {
		err := subRepo.IncrementUsage(ctx, userID)
		require.NoError(t, err)

		sub, err := subRepo.GetUserSubscription(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, sub.CurrentDayUsage)
	})

	t.Run("Increment multiple times", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			err := subRepo.IncrementUsage(ctx, userID)
			require.NoError(t, err)
		}

		sub, err := subRepo.GetUserSubscription(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 6, sub.CurrentDayUsage) // 1 from previous test + 5 new
	})
}

func TestSubscriptionRepository_UpdateSubscription(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := NewSubscriptionRepository(te.PGPool, newTestLogger())

	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := now.In(loc).AddDate(0, 0, 1)
	usageResetAt := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)

	subscription := &models.Subscription{
		ID:                 uuid.Must(uuid.NewV7()),
		UserID:             userID,
		PlanID:             freePlan.ID,
		Status:             "active",
		DailyPhotoLimit:    freePlan.DailyPhotoLimit,
		CurrentDayUsage:    0,
		UsageResetAt:       usageResetAt,
		StartedAt:          now,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 0, 30),
	}

	err = subRepo.CreateSubscription(ctx, subscription)
	require.NoError(t, err)

	t.Run("Update subscription status", func(t *testing.T) {
		cancelledAt := time.Now()
		subscription.Status = "cancelled"
		subscription.CancelledAt = &cancelledAt

		err := subRepo.UpdateSubscription(ctx, subscription)
		require.NoError(t, err)

		// Should not find active subscription
		sub, err := subRepo.GetUserSubscription(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, sub)
	})
}
