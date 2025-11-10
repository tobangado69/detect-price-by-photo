package models

import (
	"time"
)

const AIModelTable = "public.ai_models"

// AIModelStatus represents the status of an AI model.
type AIModelStatus string

const (
	AIModelStatusActive     AIModelStatus = "active"
	AIModelStatusDisabled   AIModelStatus = "disabled"
	AIModelStatusDeprecated AIModelStatus = "deprecated"
)

// AIModel represents an AI model configuration in the database.
type AIModel struct {
	ID                string         `json:"id" db:"id"`
	DisplayName       string         `json:"display_name" db:"display_name"`
	Provider          string         `json:"provider" db:"provider"`
	CostPer1kTokensUSD float64       `json:"cost_per_1k_tokens_usd" db:"cost_per_1k_tokens_usd"`
	AverageLatencyMs  int            `json:"average_latency_ms" db:"average_latency_ms"`
	ModesSupported    []string       `json:"modes_supported" db:"modes_supported"`
	FallbackChain     []string       `json:"fallback_chain" db:"fallback_chain"`
	Status            AIModelStatus `json:"status" db:"status"`
	IsDefault         bool           `json:"is_default" db:"is_default"`
	MaxTokens         int            `json:"max_tokens" db:"max_tokens"`
	Temperature       float64        `json:"temperature" db:"temperature"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         *time.Time     `json:"updated_at" db:"updated_at"`
}

// IsActive checks if the model is active.
func (m *AIModel) IsActive() bool {
	return m.Status == AIModelStatusActive
}

// SupportsMode checks if the model supports a given mode.
func (m *AIModel) SupportsMode(mode string) bool {
	for _, supportedMode := range m.ModesSupported {
		if supportedMode == mode {
			return true
		}
	}
	return false
}

