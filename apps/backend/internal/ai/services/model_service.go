package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/detect-price-by-photo/backend/internal/ai/models"
	"github.com/detect-price-by-photo/backend/internal/ai/repository"
)

// ModelServiceInterface defines the contract for AI model management.
type ModelServiceInterface interface {
	GetCatalog(ctx context.Context) ([]*models.AIModel, error)
	SetDefault(ctx context.Context, modelID string) error
	UpdateModel(ctx context.Context, modelID string, updates *models.AIModel) error
	GetDefaultModel(ctx context.Context) (*models.AIModel, error)
	GetModelForMode(ctx context.Context, mode string) (*models.AIModel, error)
}

// Ensure ModelService implements ModelServiceInterface
var _ ModelServiceInterface = (*ModelService)(nil)

// ModelService provides AI model management functionality.
type ModelService struct {
	repo   repository.AIModelRepositoryInterface
	logger *slog.Logger
}

// NewModelService creates a new ModelService.
func NewModelService(repo repository.AIModelRepositoryInterface, logger *slog.Logger) *ModelService {
	return &ModelService{
		repo:   repo,
		logger: logger,
	}
}

// GetCatalog returns all AI models.
func (s *ModelService) GetCatalog(ctx context.Context) ([]*models.AIModel, error) {
	return s.repo.GetCatalog(ctx)
}

// SetDefault sets the default AI model.
func (s *ModelService) SetDefault(ctx context.Context, modelID string) error {
	// Verify model exists and is active
	model, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		return err
	}

	if !model.IsActive() {
		return errors.New("cannot set inactive model as default")
	}

	return s.repo.SetDefault(ctx, modelID)
}

// UpdateModel updates an AI model configuration.
func (s *ModelService) UpdateModel(ctx context.Context, modelID string, updates *models.AIModel) error {
	// Get existing model
	existing, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		return err
	}

	// Apply updates
	if updates.DisplayName != "" {
		existing.DisplayName = updates.DisplayName
	}
	if updates.Provider != "" {
		existing.Provider = updates.Provider
	}
	if updates.CostPer1kTokensUSD > 0 {
		existing.CostPer1kTokensUSD = updates.CostPer1kTokensUSD
	}
	if updates.AverageLatencyMs > 0 {
		existing.AverageLatencyMs = updates.AverageLatencyMs
	}
	if updates.ModesSupported != nil {
		existing.ModesSupported = updates.ModesSupported
	}
	if updates.FallbackChain != nil {
		existing.FallbackChain = updates.FallbackChain
	}
	if updates.Status != "" {
		existing.Status = updates.Status
	}
	if updates.MaxTokens > 0 {
		existing.MaxTokens = updates.MaxTokens
	}
	if updates.Temperature > 0 {
		existing.Temperature = updates.Temperature
	}

	// Validate: if setting as default, ensure it's active
	if existing.IsDefault && existing.Status != models.AIModelStatusActive {
		return errors.New("default model must be active")
	}

	return s.repo.Update(ctx, existing)
}

// GetDefaultModel returns the default AI model.
func (s *ModelService) GetDefaultModel(ctx context.Context) (*models.AIModel, error) {
	return s.repo.GetDefault(ctx)
}

// GetModelForMode returns the model configured for a specific mode, or default if not found.
func (s *ModelService) GetModelForMode(ctx context.Context, mode string) (*models.AIModel, error) {
	// Get all models
	catalog, err := s.repo.GetCatalog(ctx)
	if err != nil {
		return nil, err
	}

	// Find model that supports this mode and is active
	for _, model := range catalog {
		if model.IsActive() && model.SupportsMode(mode) {
			return model, nil
		}
	}

	// Fallback to default model
	return s.repo.GetDefault(ctx)
}

