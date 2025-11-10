package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ForgotPasswordRequest represents the request body for forgot password endpoint.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the request body for reset password endpoint.
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
// Initiates a password reset flow by generating a token and sending it via email.
func (h *Handler) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Initiate password reset (returns nil even if user doesn't exist for security)
	if err := h.authService.InitiatePasswordReset(c.Request().Context(), req.Email); err != nil {
		h.logger.Error("Failed to initiate password reset", slog.String("error", err.Error()))
		// Still return success to prevent email enumeration
	}

	// Always return success to prevent email enumeration attacks
	return c.JSON(http.StatusOK, map[string]string{
		"message": "If an account with that email exists, a password reset link has been sent.",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
// Validates the reset token and updates the user's password.
func (h *Handler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	if err := h.authService.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
		h.logger.Error("Failed to reset password", slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password has been reset successfully. You can now log in with your new password.",
	})
}

