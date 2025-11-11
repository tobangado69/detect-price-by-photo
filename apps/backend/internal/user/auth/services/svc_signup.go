package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/detect-price-by-photo/backend/internal/user/auth/models"
	user_models "github.com/detect-price-by-photo/backend/internal/user/user/models"
	userrepo "github.com/detect-price-by-photo/backend/internal/user/user/repository"
)

// SignUp creates a new user account with the provided credentials.
// It creates the user record, sets the password, and optionally triggers email verification.
func (s *AuthService) SignUp(ctx context.Context, req *models.SignUpRequest) (*user_models.User, error) {
	if req == nil {
		return nil, errors.New("signup request is required")
	}

	// Prevent duplicate registrations
	if existing, err := s.userService.GetUserByEmail(ctx, req.Email); err == nil && existing != nil {
		return nil, ErrEmailAlreadyRegistered
	} else if err != nil && !errors.Is(err, userrepo.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		parts := strings.SplitN(req.Email, "@", 2)
		base := strings.TrimSpace(strings.ReplaceAll(parts[0], ".", " "))
		switch {
		case base == "":
			displayName = "New User"
		case len(base) == 1:
			displayName = strings.ToUpper(base)
		default:
			displayName = strings.ToUpper(base[:1]) + base[1:]
		}
	}

	user := &user_models.User{
		DisplayName: displayName,
		Email:       strings.ToLower(strings.TrimSpace(req.Email)),
	}

	if err := s.userService.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Ensure we clean up user record if password assignment fails
	if err := s.SetUserPassword(ctx, &models.UserPassword{
		UserID:       user.ID,
		PasswordHash: req.Password,
	}); err != nil {
		if delErr := s.userService.DeleteUser(ctx, user.ID); delErr != nil {
			if s.logger != nil {
				s.logger.Error("failed to rollback user after password error", "user_id", user.ID, "error", delErr.Error())
			}
		}
		return nil, fmt.Errorf("failed to set user password: %w", err)
	}

	// Trigger email verification where possible, but don't fail sign up if mailer not configured
	if err := s.InitiateEmailVerification(ctx, user.Email, req.RedirectTo); err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to initiate email verification after signup", "user_id", user.ID, "error", err.Error())
		}
	}

	return user, nil
}
