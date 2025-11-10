package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/detect-price-by-photo/backend/internal/photo/models"
	"github.com/detect-price-by-photo/backend/internal/photo/repository"
	"github.com/detect-price-by-photo/backend/internal/storage"
	subServices "github.com/detect-price-by-photo/backend/internal/subscription/services"
	"github.com/gofrs/uuid/v5"
)

// EstimatorServiceInterface defines the interface for price estimation (to avoid circular dependency).
type EstimatorServiceInterface interface {
	EstimatePrice(ctx context.Context, imageURL string, productName string, condition string, mode string) (*PriceEstimateResult, error)
}

// PriceEstimateResult represents the result from AI estimation.
// This matches the ai.PriceEstimateResult structure.
type PriceEstimateResult struct {
	EstimatedPriceMin   float64
	EstimatedPriceMax   float64
	EstimatedPriceMedian float64
	ConfidenceScore     float64
	Reasoning           string
	ModelUsed           string
	ProcessingTimeMs    int
	CostInUSD           float64
	MarketData          map[string]interface{}
}

// PhotoServiceInterface defines the contract for photo business logic.
type PhotoServiceInterface interface {
	UploadPhoto(ctx context.Context, userID uuid.UUID, file io.Reader, fileName string, contentType string, metadata *models.PhotoUploadRequest) (*models.Analysis, error)
	GetAnalysis(ctx context.Context, analysisID uuid.UUID) (*models.Analysis, error)
	ListUserAnalyses(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Analysis, int, error)
	DeleteAnalysis(ctx context.Context, userID uuid.UUID, analysisID uuid.UUID) error
	GetSignedImageURL(ctx context.Context, imageURL string, expiry time.Duration) (string, error)
	EstimatePrice(ctx context.Context, analysisID uuid.UUID) (*models.Analysis, error)
}

// Ensure PhotoService implements PhotoServiceInterface
var _ PhotoServiceInterface = (*PhotoService)(nil)

// PhotoService implements photo business logic.
type PhotoService struct {
	analysisRepo      repository.AnalysisRepositoryInterface
	storageClient     storage.StorageClientInterface
	subscriptionService subServices.SubscriptionServiceInterface
	estimatorService  EstimatorServiceInterface
	logger            *slog.Logger
}

type PhotoServiceOpts struct {
	AnalysisRepo       repository.AnalysisRepositoryInterface
	StorageClient      storage.StorageClientInterface
	SubscriptionService subServices.SubscriptionServiceInterface
	EstimatorService   EstimatorServiceInterface
	Logger             *slog.Logger
}

// NewPhotoService creates a new PhotoService.
func NewPhotoService(opts PhotoServiceOpts) *PhotoService {
	return &PhotoService{
		analysisRepo:       opts.AnalysisRepo,
		storageClient:      opts.StorageClient,
		subscriptionService: opts.SubscriptionService,
		estimatorService:   opts.EstimatorService,
		logger:             opts.Logger,
	}
}

// UploadPhoto uploads a photo, checks quota, and creates an analysis record.
func (s *PhotoService) UploadPhoto(ctx context.Context, userID uuid.UUID, file io.Reader, fileName string, contentType string, metadata *models.PhotoUploadRequest) (*models.Analysis, error) {
	// Check subscription quota
	hasQuota, err := s.subscriptionService.CheckQuota(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}
	if !hasQuota {
		return nil, errors.New("daily quota exceeded")
	}

	// Generate storage key: analyses/{user_id}/{timestamp}_{filename}
	timestamp := time.Now().Unix()
	ext := filepath.Ext(fileName)
	storageKey := fmt.Sprintf("analyses/%s/%d%s", userID.String(), timestamp, ext)

	// Upload to S3/MinIO
	_, err = s.storageClient.Upload(ctx, file, storageKey, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate signed URL (30 days expiry)
	imageURL, err := s.storageClient.GetSignedURL(ctx, storageKey, 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	// Set defaults
	condition := metadata.Condition
	if condition == "" {
		condition = models.ConditionGood
	}
	mode := metadata.Mode
	if mode == "" {
		mode = models.ModeFast
	}

	// Create analysis record
	analysis := &models.Analysis{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      userID,
		ImageURL:    imageURL,
		ProductName: metadata.ProductName,
		Condition:   condition,
		Mode:        mode,
		Status:      models.StatusPending,
	}

	err = s.analysisRepo.CreateAnalysis(ctx, analysis)
	if err != nil {
		// Try to clean up uploaded file on error
		_ = s.storageClient.Delete(ctx, storageKey)
		return nil, fmt.Errorf("failed to create analysis record: %w", err)
	}

	// Increment subscription usage
	err = s.subscriptionService.IncrementUsage(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to increment usage", "user_id", userID, "error", err)
		// Don't fail the upload if usage increment fails
	}

	s.logger.Info("photo uploaded", "analysis_id", analysis.ID, "user_id", userID)
	return analysis, nil
}

// GetAnalysis retrieves an analysis by ID.
func (s *PhotoService) GetAnalysis(ctx context.Context, analysisID uuid.UUID) (*models.Analysis, error) {
	return s.analysisRepo.GetAnalysis(ctx, analysisID)
}

// ListUserAnalyses retrieves paginated analyses for a user.
func (s *PhotoService) ListUserAnalyses(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Analysis, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.analysisRepo.ListUserAnalyses(ctx, userID, limit, offset)
}

// DeleteAnalysis deletes an analysis (soft delete).
func (s *PhotoService) DeleteAnalysis(ctx context.Context, userID uuid.UUID, analysisID uuid.UUID) error {
	// Verify ownership
	analysis, err := s.analysisRepo.GetAnalysis(ctx, analysisID)
	if err != nil {
		return err
	}

	if analysis.UserID != userID {
		return errors.New("unauthorized: analysis belongs to another user")
	}

	return s.analysisRepo.DeleteAnalysis(ctx, analysisID)
}

// GetSignedImageURL generates a new signed URL for an image.
func (s *PhotoService) GetSignedImageURL(ctx context.Context, imageURL string, expiry time.Duration) (string, error) {
	// Extract storage key from URL (implementation depends on storage backend)
	// For now, assume imageURL contains the key or we need to parse it
	// This is a simplified version - in production, you'd parse the URL properly
	return s.storageClient.GetSignedURL(ctx, imageURL, expiry)
}

// EstimatePrice performs AI price estimation for an existing analysis.
func (s *PhotoService) EstimatePrice(ctx context.Context, analysisID uuid.UUID) (*models.Analysis, error) {
	// Get analysis record
	analysis, err := s.analysisRepo.GetAnalysis(ctx, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	// Check if already completed
	if analysis.Status == models.StatusCompleted {
		return analysis, nil
	}

	// Update status to processing
	analysis.Status = models.StatusProcessing
	if err := s.analysisRepo.UpdateAnalysis(ctx, analysis); err != nil {
		s.logger.Warn("failed to update analysis status to processing", "error", err)
	}

	// Perform estimation
	if s.estimatorService == nil {
		return nil, fmt.Errorf("estimator service not configured")
	}

	result, err := s.estimatorService.EstimatePrice(ctx, analysis.ImageURL, analysis.ProductName, string(analysis.Condition), string(analysis.Mode))
	if err != nil {
		// Update status to failed
		analysis.Status = models.StatusFailed
		errorMsg := err.Error()
		analysis.ErrorMessage = &errorMsg
		_ = s.analysisRepo.UpdateAnalysis(ctx, analysis)
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// Update analysis with results
	analysis.EstimatedPriceMin = &result.EstimatedPriceMin
	analysis.EstimatedPriceMax = &result.EstimatedPriceMax
	analysis.EstimatedPriceMedian = &result.EstimatedPriceMedian
	analysis.ConfidenceScore = &result.ConfidenceScore
	analysis.Reasoning = &result.Reasoning
	analysis.ProcessingTimeMs = &result.ProcessingTimeMs
	analysis.CostInUSD = &result.CostInUSD
	analysis.ModelUsed = &result.ModelUsed
	analysis.Status = models.StatusCompleted

	// Update in database
	if err := s.analysisRepo.UpdateAnalysis(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to update analysis with results: %w", err)
	}

	s.logger.Info("price estimation completed", "analysis_id", analysisID, "model", result.ModelUsed)
	return analysis, nil
}

