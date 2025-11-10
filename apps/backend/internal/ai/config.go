package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/detect-price-by-photo/backend/internal/ai/models"
	"github.com/detect-price-by-photo/backend/internal/ai/repository"
)

// ModelConfig represents an AI model configuration.
type ModelConfig struct {
	ID          string
	Name        string
	Provider    string
	Mode        string // fast, accurate, knowledge_based
	IsDefault   bool
	FallbackTo  []string // Fallback chain
	MaxTokens   int
	Temperature float64
}

// ModelRegistry manages AI model configurations.
type ModelRegistry struct {
	models      map[string]*ModelConfig
	defaultModel *ModelConfig
	modeRouting  map[string]*ModelConfig // mode -> model
	logger      *slog.Logger
}

// NewModelRegistry creates a new model registry from config (legacy method for backward compatibility).
func NewModelRegistry(defaultModel string, modeRouting string, logger *slog.Logger) (*ModelRegistry, error) {
	registry := &ModelRegistry{
		models:     make(map[string]*ModelConfig),
		modeRouting: make(map[string]*ModelConfig),
		logger:     logger,
	}

	// Parse default model
	defaultCfg := &ModelConfig{
		ID:          defaultModel,
		Name:        defaultModel,
		Provider:    extractProvider(defaultModel),
		IsDefault:   true,
		MaxTokens:   2000,
		Temperature: 0.7,
	}
	registry.models[defaultModel] = defaultCfg
	registry.defaultModel = defaultCfg

	// Parse mode routing: "fast:model1|accurate:model2|knowledge:model3"
	if modeRouting != "" {
		routes := strings.Split(modeRouting, "|")
		for _, route := range routes {
			parts := strings.SplitN(route, ":", 2)
			if len(parts) == 2 {
				mode := strings.TrimSpace(parts[0])
				modelID := strings.TrimSpace(parts[1])

				cfg := &ModelConfig{
					ID:          modelID,
					Name:        modelID,
					Provider:    extractProvider(modelID),
					Mode:        mode,
					MaxTokens:   2000,
					Temperature: 0.7,
				}
				registry.models[modelID] = cfg
				registry.modeRouting[mode] = cfg
			}
		}
	}

	return registry, nil
}

// NewModelRegistryFromDB creates a new model registry from database.
func NewModelRegistryFromDB(repo repository.AIModelRepositoryInterface, logger *slog.Logger) (*ModelRegistry, error) {
	registry := &ModelRegistry{
		models:     make(map[string]*ModelConfig),
		modeRouting: make(map[string]*ModelConfig),
		logger:     logger,
	}

	// Load all models from database
	aiModels, err := repo.GetCatalog(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load models from database: %w", err)
	}

	// Convert database models to runtime configs
	for _, aiModel := range aiModels {
		if aiModel.Status != models.AIModelStatusActive {
			continue // Skip inactive models
		}

		cfg := &ModelConfig{
			ID:          aiModel.ID,
			Name:        aiModel.DisplayName,
			Provider:    aiModel.Provider,
			IsDefault:   aiModel.IsDefault,
			FallbackTo:  aiModel.FallbackChain,
			MaxTokens:   aiModel.MaxTokens,
			Temperature: aiModel.Temperature,
		}
		registry.models[aiModel.ID] = cfg

		if aiModel.IsDefault {
			registry.defaultModel = cfg
		}

		// Build mode routing from modes_supported
		for _, mode := range aiModel.ModesSupported {
			// Use first model that supports this mode (or override if already set)
			if _, exists := registry.modeRouting[mode]; !exists {
				cfgCopy := *cfg
				cfgCopy.Mode = mode
				registry.modeRouting[mode] = &cfgCopy
			}
		}
	}

	// Ensure we have a default model
	if registry.defaultModel == nil && len(registry.models) > 0 {
		// Use first active model as default
		for _, cfg := range registry.models {
			registry.defaultModel = cfg
			cfg.IsDefault = true
			break
		}
		logger.Warn("no default model found in database, using first active model", "model_id", registry.defaultModel.ID)
	}

	return registry, nil
}

// extractProvider extracts provider from model ID (e.g., "openai/gpt-4o-mini" -> "openai").
func extractProvider(modelID string) string {
	parts := strings.SplitN(modelID, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// GetDefaultModel returns the default model configuration.
func (r *ModelRegistry) GetDefaultModel() *ModelConfig {
	return r.defaultModel
}

// GetModelForMode returns the model configuration for a specific mode.
func (r *ModelRegistry) GetModelForMode(mode string) *ModelConfig {
	if cfg, ok := r.modeRouting[mode]; ok {
		return cfg
	}
	return r.defaultModel
}

// GetModel returns a model configuration by ID.
func (r *ModelRegistry) GetModel(modelID string) (*ModelConfig, error) {
	if cfg, ok := r.models[modelID]; ok {
		return cfg, nil
	}
	return nil, fmt.Errorf("model not found: %s", modelID)
}

