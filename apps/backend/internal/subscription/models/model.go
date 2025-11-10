// Package models contains struct definitions related to the database layer for the subscription module.
// This file defines models that map to database tables and are used for database operations.

package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Define table names
const PlanTable = "public.plans"
const SubscriptionTable = "public.subscriptions"

// Plan represents a subscription plan in the database
type Plan struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	DailyPhotoLimit int        `json:"daily_photo_limit" db:"daily_photo_limit"`
	PriceIDR        int64      `json:"price_idr" db:"price_idr"`
	PriceUSD        float64    `json:"price_usd" db:"price_usd"`
	Description     *string    `json:"description" db:"description"`
	Features        []string   `json:"features" db:"features"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
}

// Subscription represents a user's subscription in the database
type Subscription struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	PlanID             uuid.UUID  `json:"plan_id" db:"plan_id"`
	Status             string     `json:"status" db:"status"` // active, paused, cancelled
	DailyPhotoLimit    int        `json:"daily_photo_limit" db:"daily_photo_limit"`
	CurrentDayUsage    int        `json:"current_day_usage" db:"current_day_usage"`
	UsageResetAt       time.Time  `json:"usage_reset_at" db:"usage_reset_at"`
	StartedAt          time.Time  `json:"started_at" db:"started_at"`
	CurrentPeriodStart time.Time  `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end" db:"current_period_end"`
	CancelledAt        *time.Time `json:"cancelled_at" db:"cancelled_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
}

// SubscriptionWithPlan includes plan details
type SubscriptionWithPlan struct {
	Subscription
	Plan Plan `json:"plan"`
}

// HasQuotaRemaining checks if subscription has quota remaining
func (s *Subscription) HasQuotaRemaining() bool {
	return s.CurrentDayUsage < s.DailyPhotoLimit
}

// GetRemainingQuota returns remaining quota
func (s *Subscription) GetRemainingQuota() int {
	remaining := s.DailyPhotoLimit - s.CurrentDayUsage
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ParsePlanID parses a string to uuid.UUID, returns error if invalid.
func ParsePlanID(s string) (uuid.UUID, error) {
	return uuid.FromString(s)
}

// ParseSubscriptionID parses a string to uuid.UUID, returns error if invalid.
func ParseSubscriptionID(s string) (uuid.UUID, error) {
	return uuid.FromString(s)
}
