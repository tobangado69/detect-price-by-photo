package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/detect-price-by-photo/backend/internal/user/auth/models"
	"github.com/detect-price-by-photo/backend/internal/user/auth/services"
	apputils "github.com/detect-price-by-photo/backend/internal/utils"

	"github.com/labstack/echo/v4"
)

// SignUp handles public user registration.
// @Summary      User sign up
// @Description  Creates a new user account and sends a verification email
// @Tags         Auth - Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      models.SignUpRequest  true  "Sign up payload"
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      409   {object}  map[string]any
// @Failure      500   {object}  map[string]any
// @Router       /api/v1/auth/signup [post]
func (h *Handler) SignUp(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "Validation failed",
			"details": apputils.ValidationErrorsToMap(err, req),
		})
	}

	user, err := h.authService.SignUp(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyRegistered):
			return c.JSON(http.StatusConflict, map[string]any{
				"error":   "Email already registered",
				"details": "The provided email address is already in use.",
			})
		default:
			h.logger.Error("Failed to sign up user", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error":   "Failed to create account",
				"details": err.Error(),
			})
		}
	}

	response := map[string]any{
		"message": "Account created successfully. Please verify your email.",
		"user": map[string]any{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"username":     user.Username,
			"role":         user.Role,
		},
	}

	return c.JSON(http.StatusCreated, response)
}
