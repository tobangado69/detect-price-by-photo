package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/detect-price-by-photo/backend/internal/admin/models"
	"github.com/detect-price-by-photo/backend/internal/admin/repository"
	userModels "github.com/detect-price-by-photo/backend/internal/user/user/models"
	userRepo "github.com/detect-price-by-photo/backend/internal/user/user/repository"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

// UserHandler handles admin user management endpoints.
type UserHandler struct {
	userRepo     userRepo.UserRepositoryInterface
	auditLogRepo *repository.AuditLogRepository
	logger       interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userRepo userRepo.UserRepositoryInterface, auditLogRepo *repository.AuditLogRepository, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) *UserHandler {
	return &UserHandler{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// ListUsers handles GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c echo.Context) error {
	var filter userModels.FilterUser
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid query parameters",
		})
	}

	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100 // Max limit
	}

	users, err := h.userRepo.ListUsers(c.Request().Context(), &filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to list users",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users":  users,
		"total":  len(users),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetUser handles GET /api/v1/admin/users/{id}
func (h *UserHandler) GetUser(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid user ID",
		})
	}

	user, err := h.userRepo.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		if err == userRepo.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get user",
		})
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	PlanID     *uuid.UUID `json:"plan_id"`
	Status     *string    `json:"status"` // "active", "suspended"
	Role       *string    `json:"role"`   // "user", "admin"
	BanReason  *string    `json:"ban_reason"`
	BanExpires *string    `json:"ban_expires"` // RFC3339 timestamp
}

// UpdateUser handles PUT /api/v1/admin/users/{id}
func (h *UserHandler) UpdateUser(c echo.Context) error {
	// Get admin user from context (set by RequireAdmin middleware)
	adminUser, ok := c.Get("admin_user").(interface{ GetID() uuid.UUID })
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Admin authentication required",
		})
	}

	userIDStr := c.Param("id")
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid user ID",
		})
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Get existing user
	user, err := h.userRepo.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		if err == userRepo.ErrNotFound {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get user",
		})
	}

	// Track changes for audit log
	changes := make(map[string]interface{})

	// Helper to capture old values
	currentStatus := userStatus(user)
	oldBanReason := interface{}(nil)
	if user.BanReason != nil {
		oldBanReason = *user.BanReason
	}
	oldBanExpires := interface{}(nil)
	if user.BanExpires != nil {
		oldBanExpires = user.BanExpires.Format(time.RFC3339)
	}
	oldBannedAt := interface{}(nil)
	if user.BannedAt != nil {
		oldBannedAt = user.BannedAt.Format(time.RFC3339)
	}

	// Update role if provided
	if req.Role != nil {
		if *req.Role != "user" && *req.Role != "admin" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "Invalid role. Must be 'user' or 'admin'",
			})
		}
		changes["role"] = map[string]interface{}{
			"old": user.Role,
			"new": *req.Role,
		}
		user.Role = *req.Role
	}

	// Validate ban fields combination
	if (req.BanReason != nil || req.BanExpires != nil) && (req.Status == nil && currentStatus != "suspended") {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "ban_reason and ban_expires can only be set when status is 'suspended'",
		})
	}

	if req.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*req.Status))
		if status != "active" && status != "suspended" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "Invalid status. Must be 'active' or 'suspended'",
			})
		}

		if status != currentStatus {
			changes["status"] = map[string]interface{}{
				"old": currentStatus,
				"new": status,
			}
		}

		switch status {
		case "suspended":
			reason := ""
			if req.BanReason != nil {
				reason = strings.TrimSpace(*req.BanReason)
			}
			if reason == "" {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error": "ban_reason is required when suspending a user",
				})
			}
			if oldBanReason != reason {
				changes["ban_reason"] = map[string]interface{}{
					"old": oldBanReason,
					"new": reason,
				}
			}
			user.BanReason = &reason

			now := time.Now()
			user.BannedAt = &now
			changes["banned_at"] = map[string]interface{}{
				"old": oldBannedAt,
				"new": now.Format(time.RFC3339),
			}

			if req.BanExpires != nil && strings.TrimSpace(*req.BanExpires) != "" {
				expiryStr := strings.TrimSpace(*req.BanExpires)
				expiry, err := time.Parse(time.RFC3339, expiryStr)
				if err != nil {
					return c.JSON(http.StatusBadRequest, map[string]interface{}{
						"error": "ban_expires must be a valid RFC3339 timestamp",
					})
				}
				user.BanExpires = &expiry
				if oldBanExpires != expiryStr {
					changes["ban_expires"] = map[string]interface{}{
						"old": oldBanExpires,
						"new": expiryStr,
					}
				}
			} else {
				if oldBanExpires != nil {
					changes["ban_expires"] = map[string]interface{}{
						"old": oldBanExpires,
						"new": nil,
					}
				}
				user.BanExpires = nil
			}
		case "active":
			if currentStatus == "suspended" {
				changes["banned_at"] = map[string]interface{}{
					"old": oldBannedAt,
					"new": nil,
				}
				if oldBanReason != nil {
					changes["ban_reason"] = map[string]interface{}{
						"old": oldBanReason,
						"new": nil,
					}
				}
				if oldBanExpires != nil {
					changes["ban_expires"] = map[string]interface{}{
						"old": oldBanExpires,
						"new": nil,
					}
				}
			}
			user.BannedAt = nil
			user.BanExpires = nil
			user.BanReason = nil
		}
	}

	// Update user
	if err := h.userRepo.UpdateUser(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update user",
		})
	}

	// Log audit entry
	auditLog := &models.AuditLog{
		AdminID:    adminUser.GetID(),
		Action:     "update_user",
		TargetType: "user",
		TargetID:   &userID,
		Changes:    changes,
		IPAddress:  stringPtr(c.RealIP()),
		UserAgent:  stringPtr(c.Request().UserAgent()),
		CreatedAt:  time.Now(),
	}
	if err := h.auditLogRepo.CreateAuditLog(c.Request().Context(), auditLog); err != nil {
		h.logger.Error("failed to create audit log", "error", err)
	}

	return c.JSON(http.StatusOK, user)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func userStatus(user *userModels.User) string {
	if user.BannedAt != nil {
		if user.BanExpires == nil || user.BanExpires.After(time.Now()) {
			return "suspended"
		}
	}
	return "active"
}
