package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/detect-price-by-photo/backend/internal/photo/models"
	"github.com/detect-price-by-photo/backend/internal/photo/services"
	"github.com/detect-price-by-photo/backend/internal/user/auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

// HandlerInterface defines the contract for photo handlers.
type HandlerInterface interface {
	UploadPhoto(c echo.Context) error
	GetAnalysis(c echo.Context) error
	ListUserAnalyses(c echo.Context) error
	DeleteAnalysis(c echo.Context) error
	EstimatePrice(c echo.Context) error
}

// Ensure Handler implements HandlerInterface
var _ HandlerInterface = (*Handler)(nil)

// Handler holds dependencies for photo handlers.
type Handler struct {
	logger       *slog.Logger
	photoService services.PhotoServiceInterface
	validator    *validator.Validate
}

type HandlerOpts struct {
	Logger       *slog.Logger
	PhotoService services.PhotoServiceInterface
}

// NewHandler creates a new Handler instance.
func NewHandler(opts *HandlerOpts) *Handler {
	return &Handler{
		logger:       opts.Logger,
		photoService: opts.PhotoService,
		validator:    validator.New(),
	}
}

// @Summary      Upload photo for price analysis
// @Description  Uploads a product photo and creates a pending analysis record
// @Tags         Analyses
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer {token}"
// @Param        photo         formData  file    true  "Product photo (JPEG/PNG, max 10MB)"
// @Param        product_name  formData  string  true  "Product name"
// @Param        condition     formData  string  false "Product condition (new, like_new, good, fair, poor)" default(good)
// @Param        mode          formData  string  false "Estimation mode (fast, accurate, knowledge_based)" default(fast)
// @Success      201           {object}  models.AnalysisResponse
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Router       /api/v1/analyses/upload [post]
func (h *Handler) UploadPhoto(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid multipart form data"})
	}

	// Get photo file
	photos := form.File["photo"]
	if len(photos) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Photo file is required"})
	}
	photoFile := photos[0]

	// Open file
	file, err := photoFile.Open()
	if err != nil {
		h.logger.Error("Failed to open uploaded file", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process file"})
	}
	defer file.Close()

	// Get metadata
	productName := form.Value["product_name"]
	if len(productName) == 0 || productName[0] == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "product_name is required"})
	}

	condition := models.ConditionGood
	if len(form.Value["condition"]) > 0 && form.Value["condition"][0] != "" {
		condition = models.Condition(form.Value["condition"][0])
	}

	mode := models.ModeFast
	if len(form.Value["mode"]) > 0 && form.Value["mode"][0] != "" {
		mode = models.Mode(form.Value["mode"][0])
	}

	metadata := &models.PhotoUploadRequest{
		ProductName: productName[0],
		Condition:   condition,
		Mode:        mode,
	}

	// Upload photo
	analysis, err := h.photoService.UploadPhoto(ctx, userID, file, photoFile.Filename, photoFile.Header.Get("Content-Type"), metadata)
	if err != nil {
		if err.Error() == "daily quota exceeded" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Daily quota exceeded"})
		}
		h.logger.Error("Failed to upload photo", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upload photo"})
	}

	// Convert to response
	response := &models.AnalysisResponse{
		AnalysisID:  analysis.ID.String(),
		ImageURL:    analysis.ImageURL,
		ProductName: analysis.ProductName,
		Condition:   string(analysis.Condition),
		Mode:        string(analysis.Mode),
		Status:      string(analysis.Status),
		CreatedAt:   analysis.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.JSON(http.StatusCreated, response)
}

// @Summary      Get analysis details
// @Description  Retrieves details of a specific price analysis
// @Tags         Analyses
// @Security     BearerAuth
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer {token}"
// @Param        id             path      string  true  "Analysis ID"
// @Success      200            {object}  models.AnalysisResponse
// @Failure      404            {object}  map[string]string
// @Router       /api/v1/analyses/{id} [get]
func (h *Handler) GetAnalysis(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	analysisIDStr := c.Param("id")
	analysisID, err := uuid.FromString(analysisIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid analysis ID"})
	}

	analysis, err := h.photoService.GetAnalysis(ctx, analysisID)
	if err != nil {
		if err.Error() == "analysis not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Analysis not found"})
		}
		h.logger.Error("Failed to get analysis", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve analysis"})
	}

	// Verify ownership
	if analysis.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized"})
	}

	// Convert to response
	response := &models.AnalysisResponse{
		AnalysisID:          analysis.ID.String(),
		ImageURL:            analysis.ImageURL,
		ProductName:         analysis.ProductName,
		Condition:           string(analysis.Condition),
		Mode:                string(analysis.Mode),
		EstimatedPriceMin:   analysis.EstimatedPriceMin,
		EstimatedPriceMax:   analysis.EstimatedPriceMax,
		EstimatedPriceMedian: analysis.EstimatedPriceMedian,
		ConfidenceScore:     analysis.ConfidenceScore,
		Reasoning:           analysis.Reasoning,
		Status:              string(analysis.Status),
		CreatedAt:           analysis.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.JSON(http.StatusOK, response)
}

// @Summary      List user analyses
// @Description  Retrieves paginated list of user's price analyses
// @Tags         Analyses
// @Security     BearerAuth
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer {token}"
// @Param        page           query     int     false "Page number" default(1)
// @Param        limit          query     int     false "Items per page" default(20)
// @Success      200            {object}  models.AnalysisListResponse
// @Router       /api/v1/analyses/history [get]
func (h *Handler) ListUserAnalyses(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	// Parse pagination params
	page := 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	analyses, total, err := h.photoService.ListUserAnalyses(ctx, userID, page, limit)
	if err != nil {
		h.logger.Error("Failed to list analyses", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve analyses"})
	}

	// Convert to response
	responses := make([]models.AnalysisResponse, len(analyses))
	for i, analysis := range analyses {
		responses[i] = models.AnalysisResponse{
			AnalysisID:          analysis.ID.String(),
			ImageURL:            analysis.ImageURL,
			ProductName:         analysis.ProductName,
			Condition:           string(analysis.Condition),
			Mode:                string(analysis.Mode),
			EstimatedPriceMin:   analysis.EstimatedPriceMin,
			EstimatedPriceMax:   analysis.EstimatedPriceMax,
			EstimatedPriceMedian: analysis.EstimatedPriceMedian,
			ConfidenceScore:     analysis.ConfidenceScore,
			Reasoning:           analysis.Reasoning,
			Status:              string(analysis.Status),
			CreatedAt:           analysis.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := &models.AnalysisListResponse{
		Analyses: responses,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	return c.JSON(http.StatusOK, response)
}

// @Summary      Delete analysis
// @Description  Soft deletes a price analysis
// @Tags         Analyses
// @Security     BearerAuth
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer {token}"
// @Param        id             path      string  true  "Analysis ID"
// @Success      200            {object}  map[string]string
// @Failure      404            {object}  map[string]string
// @Router       /api/v1/analyses/{id} [delete]
func (h *Handler) DeleteAnalysis(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	analysisIDStr := c.Param("id")
	analysisID, err := uuid.FromString(analysisIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid analysis ID"})
	}

	err = h.photoService.DeleteAnalysis(ctx, userID, analysisID)
	if err != nil {
		if err.Error() == "analysis not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Analysis not found"})
		}
		if err.Error() == "unauthorized: analysis belongs to another user" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized"})
		}
		h.logger.Error("Failed to delete analysis", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete analysis"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Analysis deleted successfully"})
}

// @Summary      Estimate price for analysis
// @Description  Triggers AI price estimation for an existing analysis
// @Tags         Analyses
// @Security     BearerAuth
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer {token}"
// @Param        id             path      string  true  "Analysis ID"
// @Success      200            {object}  models.AnalysisResponse
// @Failure      404            {object}  map[string]string
// @Failure      500            {object}  map[string]string
// @Router       /api/v1/analyses/{id}/estimate [post]
func (h *Handler) EstimatePrice(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	analysisIDStr := c.Param("id")
	analysisID, err := uuid.FromString(analysisIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid analysis ID"})
	}

	// Verify ownership
	analysis, err := h.photoService.GetAnalysis(ctx, analysisID)
	if err != nil {
		if err.Error() == "analysis not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Analysis not found"})
		}
		h.logger.Error("Failed to get analysis", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve analysis"})
	}

	if analysis.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized"})
	}

	// Perform estimation
	updatedAnalysis, err := h.photoService.EstimatePrice(ctx, analysisID)
	if err != nil {
		h.logger.Error("Failed to estimate price", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to estimate price"})
	}

	// Convert to response
	response := &models.AnalysisResponse{
		AnalysisID:          updatedAnalysis.ID.String(),
		ImageURL:            updatedAnalysis.ImageURL,
		ProductName:         updatedAnalysis.ProductName,
		Condition:           string(updatedAnalysis.Condition),
		Mode:                string(updatedAnalysis.Mode),
		EstimatedPriceMin:   updatedAnalysis.EstimatedPriceMin,
		EstimatedPriceMax:   updatedAnalysis.EstimatedPriceMax,
		EstimatedPriceMedian: updatedAnalysis.EstimatedPriceMedian,
		ConfidenceScore:     updatedAnalysis.ConfidenceScore,
		Reasoning:           updatedAnalysis.Reasoning,
		Status:              string(updatedAnalysis.Status),
		CreatedAt:           updatedAnalysis.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.JSON(http.StatusOK, response)
}

