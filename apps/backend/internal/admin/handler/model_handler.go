package handler

import (
	"net/http"
	"time"

	"github.com/detect-price-by-photo/backend/internal/admin/models"
	"github.com/detect-price-by-photo/backend/internal/admin/repository"
	aiModels "github.com/detect-price-by-photo/backend/internal/ai/models"
	aiServices "github.com/detect-price-by-photo/backend/internal/ai/services"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

// ModelHandler handles admin AI model management endpoints.
type ModelHandler struct {
	modelService aiServices.ModelServiceInterface
	auditLogRepo *repository.AuditLogRepository
	logger       interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewModelHandler creates a new ModelHandler.
func NewModelHandler(modelService aiServices.ModelServiceInterface, auditLogRepo *repository.AuditLogRepository, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) *ModelHandler {
	return &ModelHandler{
		modelService: modelService,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// ListModels handles GET /api/v1/admin/models
func (h *ModelHandler) ListModels(c echo.Context) error {
	models, err := h.modelService.GetCatalog(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to list models",
		})
	}

	return c.JSON(http.StatusOK, models)
}

// SetDefaultModelRequest represents the request body for setting default model.
type SetDefaultModelRequest struct {
	ModelID string `json:"model_id" validate:"required"`
}

// SetDefaultModel handles PUT /api/v1/admin/models/default
func (h *ModelHandler) SetDefaultModel(c echo.Context) error {
	// Get admin user from context (set by RequireAdmin middleware)
	adminUser, ok := c.Get("admin_user").(interface{ GetID() uuid.UUID })
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Admin authentication required",
		})
	}

	var req SetDefaultModelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	if err := h.modelService.SetDefault(c.Request().Context(), req.ModelID); err != nil {
		if err.Error() == "model not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Model not found",
			})
		}
		if err.Error() == "cannot set inactive model as default" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "Cannot set inactive model as default",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to set default model",
		})
	}

	// Log audit entry
	adminID := adminUser.GetID()
	auditLog := &models.AuditLog{
		AdminID:    adminID,
		Action:     "set_default_model",
		TargetType: "ai_model",
		TargetID:   nil, // Model ID is string, not UUID, so store in changes instead
		Changes: map[string]interface{}{
			"model_id": req.ModelID,
		},
		IPAddress: stringPtr(c.RealIP()),
		UserAgent: stringPtr(c.Request().UserAgent()),
		CreatedAt: time.Now(),
	}
	if err := h.auditLogRepo.CreateAuditLog(c.Request().Context(), auditLog); err != nil {
		h.logger.Error("failed to create audit log", "error", err)
	}

	// Get updated default model
	defaultModel, err := h.modelService.GetDefaultModel(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":      "Default model updated",
			"default_model": req.ModelID,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Default model updated",
		"default_model": defaultModel,
	})
}

// UpdateModelRequest represents the request body for updating a model.
type UpdateModelRequest struct {
	DisplayName       *string         `json:"display_name"`
	ModesSupported    []string         `json:"modes_supported"`
	FallbackChain     []string         `json:"fallback_chain"`
	Status            *aiModels.AIModelStatus `json:"status"`
	MaxTokens         *int            `json:"max_tokens"`
	Temperature       *float64        `json:"temperature"`
}

// UpdateModel handles PUT /api/v1/admin/models/{id}
func (h *ModelHandler) UpdateModel(c echo.Context) error {
	// Get admin user from context (set by RequireAdmin middleware)
	adminUser, ok := c.Get("admin_user").(interface{ GetID() uuid.UUID })
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Admin authentication required",
		})
	}

	modelID := c.Param("id")
	if modelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Model ID required",
		})
	}

	var req UpdateModelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Build update model
	updates := &aiModels.AIModel{
		ID: modelID,
	}
	if req.DisplayName != nil {
		updates.DisplayName = *req.DisplayName
	}
	if req.ModesSupported != nil {
		updates.ModesSupported = req.ModesSupported
	}
	if req.FallbackChain != nil {
		updates.FallbackChain = req.FallbackChain
	}
	if req.Status != nil {
		updates.Status = *req.Status
	}
	if req.MaxTokens != nil {
		updates.MaxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		updates.Temperature = *req.Temperature
	}

	if err := h.modelService.UpdateModel(c.Request().Context(), modelID, updates); err != nil {
		if err.Error() == "model not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Model not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update model",
		})
	}

	// Log audit entry
	adminID := adminUser.GetID()
	changes := make(map[string]interface{})
	if req.DisplayName != nil {
		changes["display_name"] = *req.DisplayName
	}
	if req.ModesSupported != nil {
		changes["modes_supported"] = req.ModesSupported
	}
	if req.FallbackChain != nil {
		changes["fallback_chain"] = req.FallbackChain
	}
	if req.Status != nil {
		changes["status"] = *req.Status
	}
	changes["model_id"] = modelID // Store model ID in changes since it's a string, not UUID
	auditLog := &models.AuditLog{
		AdminID:    adminID,
		Action:     "update_model",
		TargetType: "ai_model",
		TargetID:   nil, // Model ID is string, not UUID, so store in changes instead
		Changes:    changes,
		IPAddress:  stringPtr(c.RealIP()),
		UserAgent:  stringPtr(c.Request().UserAgent()),
		CreatedAt:  time.Now(),
	}
	if err := h.auditLogRepo.CreateAuditLog(c.Request().Context(), auditLog); err != nil {
		h.logger.Error("failed to create audit log", "error", err)
	}

	// Get updated model
	updatedModel, err := h.modelService.GetCatalog(c.Request().Context())
	if err == nil {
		for _, m := range updatedModel {
			if m.ID == modelID {
				return c.JSON(http.StatusOK, m)
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Model updated successfully",
		"model_id": modelID,
	})
}

