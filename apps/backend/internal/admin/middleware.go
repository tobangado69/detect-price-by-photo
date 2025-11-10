package admin

import (
	"net/http"

	"github.com/detect-price-by-photo/backend/internal/user/auth"
	"github.com/detect-price-by-photo/backend/internal/user/user/models"
	"github.com/detect-price-by-photo/backend/internal/user/user/repository"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

// RequireAdmin middleware checks if the authenticated user has admin role.
// It requires a UserRepository to fetch the user from the database.
func RequireAdmin(userRepo repository.UserRepositoryInterface) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user ID from JWT claims
			userIDStr, ok := auth.GetUserID(c)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			userID, err := uuid.FromString(userIDStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
			}

			// Fetch user from database to check role
			user, err := userRepo.GetUserByID(c.Request().Context(), userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
			}

			// Check if user has admin role
			if user.Role != "admin" {
				return echo.NewHTTPError(http.StatusForbidden, "admin access required")
			}

			// Store admin user in context for handlers
			c.Set("admin_user", user)

			return next(c)
		}
	}
}

// GetAdminFromContext extracts admin user from echo context (set by RequireAdmin middleware).
func GetAdminFromContext(c echo.Context) (*models.User, error) {
	adminUser, ok := c.Get("admin_user").(*models.User)
	if !ok || adminUser == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "admin authentication required")
	}

	return adminUser, nil
}

// GetAdminIDFromContext extracts admin user ID from context.
func GetAdminIDFromContext(c echo.Context) (uuid.UUID, error) {
	adminUser, err := GetAdminFromContext(c)
	if err != nil {
		return uuid.Nil, err
	}
	return adminUser.ID, nil
}

