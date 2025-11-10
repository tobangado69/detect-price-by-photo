package handler

import (
	"net/http"
	"time"

	"github.com/detect-price-by-photo/backend/internal/admin/models"
	"github.com/detect-price-by-photo/backend/internal/admin/repository"
	subModels "github.com/detect-price-by-photo/backend/internal/subscription/models"
	subRepo "github.com/detect-price-by-photo/backend/internal/subscription/repository"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

// PlanHandler handles admin plan management endpoints.
type PlanHandler struct {
	planRepo     subRepo.PlanRepositoryInterface
	auditLogRepo *repository.AuditLogRepository
	logger       interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewPlanHandler creates a new PlanHandler.
func NewPlanHandler(planRepo subRepo.PlanRepositoryInterface, auditLogRepo *repository.AuditLogRepository, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) *PlanHandler {
	return &PlanHandler{
		planRepo:     planRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// ListPlans handles GET /api/v1/admin/plans
func (h *PlanHandler) ListPlans(c echo.Context) error {
	activeOnly := c.QueryParam("active_only") == "true"
	plans, err := h.planRepo.GetPlans(c.Request().Context(), activeOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to list plans",
		})
	}

	return c.JSON(http.StatusOK, plans)
}

// CreatePlanRequest represents the request body for creating a plan.
type CreatePlanRequest struct {
	Name           string   `json:"name" validate:"required"`
	DailyPhotoLimit int     `json:"daily_photo_limit" validate:"required"`
	PriceIDR       float64  `json:"price_idr" validate:"required"`
	PriceUSD       float64  `json:"price_usd" validate:"required"`
	Description    string   `json:"description"`
	Features       []string `json:"features"`
	IsActive       bool     `json:"is_active"`
}

// CreatePlan handles POST /api/v1/admin/plans
func (h *PlanHandler) CreatePlan(c echo.Context) error {
	// Get admin user from context (set by RequireAdmin middleware)
	adminUser, ok := c.Get("admin_user").(interface{ GetID() uuid.UUID })
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Admin authentication required",
		})
	}

	var req CreatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	plan := &subModels.Plan{
		ID:             uuid.Must(uuid.NewV7()),
		Name:           req.Name,
		DailyPhotoLimit: req.DailyPhotoLimit,
		PriceIDR:       int64(req.PriceIDR),
		PriceUSD:       req.PriceUSD,
		Description:    stringPtr(req.Description),
		Features:       req.Features,
		IsActive:       req.IsActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      timePtr(time.Now()),
	}

	// Note: PlanRepository doesn't have CreatePlan method yet, so we'll need to add it
	// For now, return an error indicating this needs to be implemented
	return c.JSON(http.StatusNotImplemented, map[string]interface{}{
		"error": "Plan creation not yet implemented in repository",
	})

	// Log audit entry
	auditLog := &models.AuditLog{
		AdminID:    adminUser.GetID(),
		Action:     "create_plan",
		TargetType: "plan",
		TargetID:   &plan.ID,
		Changes: map[string]interface{}{
			"plan": plan,
		},
		IPAddress: stringPtr(c.RealIP()),
		UserAgent: stringPtr(c.Request().UserAgent()),
		CreatedAt: time.Now(),
	}
	if err := h.auditLogRepo.CreateAuditLog(c.Request().Context(), auditLog); err != nil {
		h.logger.Error("failed to create audit log", "error", err)
	}

	return c.JSON(http.StatusCreated, plan)
}

// UpdatePlanRequest represents the request body for updating a plan.
type UpdatePlanRequest struct {
	Name           *string  `json:"name"`
	DailyPhotoLimit *int     `json:"daily_photo_limit"`
	PriceIDR       *float64  `json:"price_idr"`
	PriceUSD       *float64  `json:"price_usd"`
	Description    *string  `json:"description"`
	Features       []string `json:"features"`
	IsActive       *bool    `json:"is_active"`
}

// UpdatePlan handles PUT /api/v1/admin/plans/{id}
func (h *PlanHandler) UpdatePlan(c echo.Context) error {
	// Get admin user from context (set by RequireAdmin middleware)
	adminUser, ok := c.Get("admin_user").(interface{ GetID() uuid.UUID })
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Admin authentication required",
		})
	}

	planIDStr := c.Param("id")
	planID, err := uuid.FromString(planIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid plan ID",
		})
	}

	var req UpdatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Get existing plan
	plan, err := h.planRepo.GetPlanByID(c.Request().Context(), planID)
	if err != nil {
		if err == subRepo.ErrPlanNotFound {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Plan not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get plan",
		})
	}

	// Track changes
	changes := make(map[string]interface{})

	// Update fields if provided
	if req.Name != nil {
		changes["name"] = map[string]interface{}{"old": plan.Name, "new": *req.Name}
		plan.Name = *req.Name
	}
	if req.DailyPhotoLimit != nil {
		changes["daily_photo_limit"] = map[string]interface{}{"old": plan.DailyPhotoLimit, "new": *req.DailyPhotoLimit}
		plan.DailyPhotoLimit = *req.DailyPhotoLimit
	}
	if req.PriceIDR != nil {
		changes["price_idr"] = map[string]interface{}{"old": plan.PriceIDR, "new": int64(*req.PriceIDR)}
		plan.PriceIDR = int64(*req.PriceIDR)
	}
	if req.PriceUSD != nil {
		changes["price_usd"] = map[string]interface{}{"old": plan.PriceUSD, "new": *req.PriceUSD}
		plan.PriceUSD = *req.PriceUSD
	}
	if req.Description != nil {
		changes["description"] = map[string]interface{}{"old": plan.Description, "new": *req.Description}
		plan.Description = req.Description
	}
	if req.Features != nil {
		changes["features"] = map[string]interface{}{"old": plan.Features, "new": req.Features}
		plan.Features = req.Features
	}
	if req.IsActive != nil {
		changes["is_active"] = map[string]interface{}{"old": plan.IsActive, "new": *req.IsActive}
		plan.IsActive = *req.IsActive
	}

	plan.UpdatedAt = timePtr(time.Now())

	// Note: PlanRepository doesn't have UpdatePlan method yet
	return c.JSON(http.StatusNotImplemented, map[string]interface{}{
		"error": "Plan update not yet implemented in repository",
	})

	// Log audit entry
	auditLog := &models.AuditLog{
		AdminID:    adminUser.GetID(),
		Action:     "update_plan",
		TargetType: "plan",
		TargetID:   &planID,
		Changes:    changes,
		IPAddress:  stringPtr(c.RealIP()),
		UserAgent:  stringPtr(c.Request().UserAgent()),
		CreatedAt:  time.Now(),
	}
	if err := h.auditLogRepo.CreateAuditLog(c.Request().Context(), auditLog); err != nil {
		h.logger.Error("failed to create audit log", "error", err)
	}

	return c.JSON(http.StatusOK, plan)
}

// DeletePlan handles DELETE /api/v1/admin/plans/{id}
func (h *PlanHandler) DeletePlan(c echo.Context) error {
	// Get admin user from context (set by RequireAdmin middleware)
	adminUser, ok := c.Get("admin_user").(interface{ GetID() uuid.UUID })
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Admin authentication required",
		})
	}

	planIDStr := c.Param("id")
	planID, err := uuid.FromString(planIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid plan ID",
		})
	}

	// Get plan to verify it exists
	plan, err := h.planRepo.GetPlanByID(c.Request().Context(), planID)
	if err != nil {
		if err == subRepo.ErrPlanNotFound {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Plan not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get plan",
		})
	}

	// Soft delete: set is_active = false
	plan.IsActive = false
	plan.UpdatedAt = timePtr(time.Now())

	// Note: PlanRepository doesn't have UpdatePlan method yet
	return c.JSON(http.StatusNotImplemented, map[string]interface{}{
		"error": "Plan deletion not yet implemented in repository",
	})

	// Log audit entry
	auditLog := &models.AuditLog{
		AdminID:    adminUser.GetID(),
		Action:     "delete_plan",
		TargetType: "plan",
		TargetID:   &planID,
		Changes: map[string]interface{}{
			"plan_name": plan.Name,
			"is_active": false,
		},
		IPAddress: stringPtr(c.RealIP()),
		UserAgent: stringPtr(c.Request().UserAgent()),
		CreatedAt: time.Now(),
	}
	if err := h.auditLogRepo.CreateAuditLog(c.Request().Context(), auditLog); err != nil {
		h.logger.Error("failed to create audit log", "error", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Plan deleted successfully",
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}

