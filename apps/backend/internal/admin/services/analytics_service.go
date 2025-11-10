package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsService provides analytics and reporting functionality.
type AnalyticsService struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(pgPool *pgxpool.Pool, logger *slog.Logger) *AnalyticsService {
	return &AnalyticsService{
		pgPool: pgPool,
		logger: logger,
	}
}

// DashboardMetrics represents dashboard metrics.
type DashboardMetrics struct {
	TotalUsers            int64                  `json:"total_users"`
	ActiveSubscriptions   map[string]int64       `json:"active_subscriptions"` // plan_name -> count
	MRR                   float64                `json:"mrr"`                 // Monthly Recurring Revenue in IDR
	DailyActiveUsers      int64                  `json:"daily_active_users"`
	TotalAnalyses         int64                  `json:"total_analyses"`
	AnalysesToday         int64                  `json:"analyses_today"`
	AverageLatencyMs      float64                `json:"average_latency_ms"`
	ErrorRate             float64                `json:"error_rate"` // 0.0 to 1.0
}

// GetDashboard returns dashboard metrics.
func (s *AnalyticsService) GetDashboard(ctx context.Context) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{
		ActiveSubscriptions: make(map[string]int64),
	}

	// Total users
	if err := s.pgPool.QueryRow(ctx, `SELECT COUNT(*) FROM public.users`).Scan(&metrics.TotalUsers); err != nil {
		s.logger.Error("failed to get total users", "error", err)
		return nil, err
	}

	// Active subscriptions by plan
	rows, err := s.pgPool.Query(ctx, `
		SELECT p.name, COUNT(s.id)
		FROM public.subscriptions s
		JOIN public.plans p ON s.plan_id = p.id
		WHERE s.status = 'active'
		GROUP BY p.name
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var planName string
			var count int64
			if err := rows.Scan(&planName, &count); err == nil {
				metrics.ActiveSubscriptions[planName] = count
			}
		}
	}

	// MRR (Monthly Recurring Revenue)
	if err := s.pgPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.price_idr), 0)
		FROM public.subscriptions s
		JOIN public.plans p ON s.plan_id = p.id
		WHERE s.status = 'active'
	`).Scan(&metrics.MRR); err != nil {
		s.logger.Warn("failed to calculate MRR", "error", err)
	}

	// Daily active users (users who logged in today)
	today := time.Now().Format("2006-01-02")
	if err := s.pgPool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM public.sessions
		WHERE created_at::date = $1::date
	`, today).Scan(&metrics.DailyActiveUsers); err != nil {
		s.logger.Warn("failed to get daily active users", "error", err)
	}

	// Total analyses
	if err := s.pgPool.QueryRow(ctx, `SELECT COUNT(*) FROM public.price_analyses`).Scan(&metrics.TotalAnalyses); err != nil {
		s.logger.Warn("failed to get total analyses", "error", err)
	}

	// Analyses today
	if err := s.pgPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM public.price_analyses
		WHERE created_at::date = $1::date
	`, today).Scan(&metrics.AnalysesToday); err != nil {
		s.logger.Warn("failed to get analyses today", "error", err)
	}

	// Average latency (from price_analyses.processing_time_ms)
	if err := s.pgPool.QueryRow(ctx, `
		SELECT COALESCE(AVG(processing_time_ms), 0)
		FROM public.price_analyses
		WHERE processing_time_ms IS NOT NULL
	`).Scan(&metrics.AverageLatencyMs); err != nil {
		s.logger.Warn("failed to get average latency", "error", err)
	}

	// Error rate (analyses with status = 'failed' / total analyses)
	var totalAnalyses, failedAnalyses int64
	if err := s.pgPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM public.price_analyses
	`).Scan(&totalAnalyses); err == nil {
		if err := s.pgPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM public.price_analyses
			WHERE status = 'failed'
		`).Scan(&failedAnalyses); err == nil && totalAnalyses > 0 {
			metrics.ErrorRate = float64(failedAnalyses) / float64(totalAnalyses)
		}
	}

	return metrics, nil
}

// RevenueMetrics represents revenue analytics.
type RevenueMetrics struct {
	Period          string    `json:"period"` // "daily", "weekly", "monthly"
	TotalRevenueIDR float64   `json:"total_revenue_idr"`
	TotalRevenueUSD float64   `json:"total_revenue_usd"`
	Transactions    int64     `json:"transactions"`
	AverageOrderValue float64 `json:"average_order_value"`
	Trend            []RevenueDataPoint `json:"trend"`
}

// RevenueDataPoint represents a single revenue data point.
type RevenueDataPoint struct {
	Date            string  `json:"date"`
	RevenueIDR      float64 `json:"revenue_idr"`
	RevenueUSD      float64 `json:"revenue_usd"`
	Transactions    int64   `json:"transactions"`
}

// GetRevenueMetrics returns revenue metrics for a given period.
func (s *AnalyticsService) GetRevenueMetrics(ctx context.Context, period string) (*RevenueMetrics, error) {
	metrics := &RevenueMetrics{
		Period: period,
		Trend:  []RevenueDataPoint{},
	}

	// Validate period
	if period != "daily" && period != "weekly" && period != "monthly" {
		period = "monthly"
	}

	// Calculate date range
	now := time.Now()
	var startDate time.Time
	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -30) // Last 30 days
	case "weekly":
		startDate = now.AddDate(0, -3, 0) // Last 3 months
	case "monthly":
		startDate = now.AddDate(-12, 0, 0) // Last 12 months
	}

	// Total revenue
	if err := s.pgPool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(amount_idr), 0),
			COALESCE(SUM(amount_usd), 0),
			COUNT(*)
		FROM public.invoices
		WHERE status = 'paid' AND created_at >= $1
	`, startDate).Scan(&metrics.TotalRevenueIDR, &metrics.TotalRevenueUSD, &metrics.Transactions); err != nil {
		s.logger.Warn("failed to get revenue metrics", "error", err)
	}

	if metrics.Transactions > 0 {
		metrics.AverageOrderValue = metrics.TotalRevenueIDR / float64(metrics.Transactions)
	}

	// Revenue trend
	var dateFormat string
	switch period {
	case "daily":
		dateFormat = "YYYY-MM-DD"
	case "weekly":
		dateFormat = "YYYY-\"W\"WW"
	case "monthly":
		dateFormat = "YYYY-MM"
	}

	rows, err := s.pgPool.Query(ctx, `
		SELECT 
			TO_CHAR(created_at, $1) as period,
			COALESCE(SUM(amount_idr), 0) as revenue_idr,
			COALESCE(SUM(amount_usd), 0) as revenue_usd,
			COUNT(*) as transactions
		FROM public.invoices
		WHERE status = 'paid' AND created_at >= $2
		GROUP BY TO_CHAR(created_at, $1)
		ORDER BY period ASC
	`, dateFormat, startDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var point RevenueDataPoint
			if err := rows.Scan(&point.Date, &point.RevenueIDR, &point.RevenueUSD, &point.Transactions); err == nil {
				metrics.Trend = append(metrics.Trend, point)
			}
		}
	}

	return metrics, nil
}

