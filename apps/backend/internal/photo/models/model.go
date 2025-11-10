package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Condition represents product condition.
type Condition string

const (
	ConditionNew      Condition = "new"
	ConditionLikeNew  Condition = "like_new"
	ConditionGood     Condition = "good"
	ConditionFair     Condition = "fair"
	ConditionPoor     Condition = "poor"
)

// Mode represents estimation mode.
type Mode string

const (
	ModeFast          Mode = "fast"
	ModeAccurate      Mode = "accurate"
	ModeKnowledgeBased Mode = "knowledge_based"
)

// Status represents analysis status.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// AnalysisTable is the table name for price analyses.
const AnalysisTable = "public.price_analyses"

// Analysis represents a price analysis record.
type Analysis struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	ImageURL            string
	ProductName         string
	Condition           Condition
	Mode                Mode
	
	// Price estimation results
	EstimatedPriceMin   *float64
	EstimatedPriceMax   *float64
	EstimatedPriceMedian *float64
	ConfidenceScore     *float64
	Reasoning           *string
	
	// Processing metadata
	ProcessingTimeMs    *int
	CostInUSD           *float64
	ModelUsed           *string
	
	// Status tracking
	Status              Status
	ErrorMessage        *string
	
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	DeletedAt           *time.Time
}

// HasPriceEstimate returns true if the analysis has price estimates.
func (a *Analysis) HasPriceEstimate() bool {
	return a.EstimatedPriceMin != nil && a.EstimatedPriceMax != nil
}

// IsCompleted returns true if the analysis is completed.
func (a *Analysis) IsCompleted() bool {
	return a.Status == StatusCompleted
}

// IsFailed returns true if the analysis failed.
func (a *Analysis) IsFailed() bool {
	return a.Status == StatusFailed
}

// PhotoUploadRequest represents the upload request payload.
type PhotoUploadRequest struct {
	ProductName string
	Condition   Condition
	Mode        Mode
}

// AnalysisResponse represents the API response for an analysis.
type AnalysisResponse struct {
	AnalysisID          string   `json:"analysis_id"`
	ImageURL            string   `json:"image_url"`
	ProductName         string   `json:"product_name"`
	Condition           string   `json:"condition"`
	Mode                string   `json:"mode"`
	EstimatedPriceMin   *float64 `json:"estimated_price_min,omitempty"`
	EstimatedPriceMax   *float64 `json:"estimated_price_max,omitempty"`
	EstimatedPriceMedian *float64 `json:"estimated_price_median,omitempty"`
	ConfidenceScore     *float64 `json:"confidence_score,omitempty"`
	Reasoning           *string   `json:"reasoning,omitempty"`
	Status              string   `json:"status"`
	CreatedAt           string   `json:"created_at"`
}

// AnalysisListResponse represents paginated analysis list.
type AnalysisListResponse struct {
	Analyses []AnalysisResponse `json:"analyses"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	Limit    int                 `json:"limit"`
}

