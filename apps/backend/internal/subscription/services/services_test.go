package services

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/detect-price-by-photo/backend/internal/subscription/repository"
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

func TestSubscriptionService_ListPlans(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := repository.NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := repository.NewSubscriptionRepository(te.PGPool, newTestLogger())

	service := NewSubscriptionService(SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subRepo,
	})

	t.Run("List all plans", func(t *testing.T) {
		plans, err := service.ListPlans(ctx, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(plans), 3)
	})

	t.Run("List active plans only", func(t *testing.T) {
		plans, err := service.ListPlans(ctx, true)
		require.NoError(t, err)
		for _, plan := range plans {
			assert.True(t, plan.IsActive)
		}
	})
}

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := repository.NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := repository.NewSubscriptionRepository(te.PGPool, newTestLogger())

	service := NewSubscriptionService(SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subRepo,
	})

	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())

	t.Run("Create new subscription", func(t *testing.T) {
		sub, err := service.CreateSubscription(ctx, userID, freePlan.ID)
		require.NoError(t, err)
		assert.Equal(t, userID, sub.UserID)
		assert.Equal(t, freePlan.ID, sub.PlanID)
		assert.Equal(t, "active", sub.Status)
		assert.Equal(t, freePlan.DailyPhotoLimit, sub.DailyPhotoLimit)
		assert.Equal(t, 0, sub.CurrentDayUsage)
	})

	t.Run("Create subscription cancels existing one", func(t *testing.T) {
		// Create first subscription
		sub1, err := service.CreateSubscription(ctx, userID, freePlan.ID)
		require.NoError(t, err)

		// Create second subscription (should cancel first)
		premiumPlan, err := planRepo.GetPlanByName(ctx, "premium")
		require.NoError(t, err)

		sub2, err := service.CreateSubscription(ctx, userID, premiumPlan.ID)
		require.NoError(t, err)
		assert.Equal(t, premiumPlan.ID, sub2.PlanID)

		// Verify old subscription is cancelled
		oldSub, err := subRepo.GetUserSubscription(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, premiumPlan.ID, oldSub.PlanID) // Should be the new one
		assert.NotEqual(t, sub1.ID, oldSub.ID)
	})

	t.Run("Create subscription with invalid plan", func(t *testing.T) {
		fakePlanID := uuid.Must(uuid.NewV7())
		sub, err := service.CreateSubscription(ctx, userID, fakePlanID)
		assert.Error(t, err)
		assert.Nil(t, sub)
		assert.Contains(t, err.Error(), "plan not found")
	})
}

func TestSubscriptionService_GetCurrentSubscription(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := repository.NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := repository.NewSubscriptionRepository(te.PGPool, newTestLogger())

	service := NewSubscriptionService(SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subRepo,
	})

	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())

	t.Run("Get existing subscription", func(t *testing.T) {
		_, err := service.CreateSubscription(ctx, userID, freePlan.ID)
		require.NoError(t, err)

		sub, err := service.GetCurrentSubscription(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, userID, sub.UserID)
		assert.Equal(t, freePlan.ID, sub.PlanID)
	})

	t.Run("Get non-existent subscription", func(t *testing.T) {
		fakeUserID := uuid.Must(uuid.NewV7())
		sub, err := service.GetCurrentSubscription(ctx, fakeUserID)
		assert.Error(t, err)
		assert.Nil(t, sub)
		assert.Contains(t, err.Error(), "no active subscription found")
	})
}

func TestSubscriptionService_CheckQuota(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := repository.NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := repository.NewSubscriptionRepository(te.PGPool, newTestLogger())

	service := NewSubscriptionService(SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subRepo,
	})

	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())

	_, err = service.CreateSubscription(ctx, userID, freePlan.ID)
	require.NoError(t, err)

	t.Run("Check quota when available", func(t *testing.T) {
		hasQuota, err := service.CheckQuota(ctx, userID)
		require.NoError(t, err)
		assert.True(t, hasQuota)
	})

	t.Run("Check quota after incrementing", func(t *testing.T) {
		// Increment usage to limit
		for i := 0; i < freePlan.DailyPhotoLimit; i++ {
			err := service.IncrementUsage(ctx, userID)
			require.NoError(t, err)
		}

		hasQuota, err := service.CheckQuota(ctx, userID)
		require.NoError(t, err)
		assert.False(t, hasQuota)
	})
}

func TestSubscriptionService_IncrementUsage(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := repository.NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := repository.NewSubscriptionRepository(te.PGPool, newTestLogger())

	service := NewSubscriptionService(SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subRepo,
	})

	freePlan, err := planRepo.GetPlanByName(ctx, "free")
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())

	_, err = service.CreateSubscription(ctx, userID, freePlan.ID)
	require.NoError(t, err)

	t.Run("Increment usage successfully", func(t *testing.T) {
		err := service.IncrementUsage(ctx, userID)
		require.NoError(t, err)

		sub, err := service.GetCurrentSubscription(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, sub.CurrentDayUsage)
	})

	t.Run("Increment usage when quota exceeded", func(t *testing.T) {
		// Reset usage first
		sub, err := service.GetCurrentSubscription(ctx, userID)
		require.NoError(t, err)
		sub.CurrentDayUsage = freePlan.DailyPhotoLimit
		err = subRepo.UpdateSubscription(ctx, sub)
		require.NoError(t, err)

		err = service.IncrementUsage(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})
}

func TestSubscriptionService_GetPlanByName(t *testing.T) {
	ctx := context.Background()
	te, cleanup := setupTestDB(t)
	defer cleanup()

	planRepo := repository.NewPlanRepository(te.PGPool, newTestLogger())
	subRepo := repository.NewSubscriptionRepository(te.PGPool, newTestLogger())

	service := NewSubscriptionService(SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subRepo,
	})

	t.Run("Get Free plan ID", func(t *testing.T) {
		planID, err := service.GetPlanByName(ctx, "free")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, planID)

		plan, err := planRepo.GetPlanByID(ctx, planID)
		require.NoError(t, err)
		assert.Equal(t, "free", plan.Name)
	})

	t.Run("Get non-existent plan", func(t *testing.T) {
		planID, err := service.GetPlanByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, planID)
	})
}

