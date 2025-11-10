package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/detect-price-by-photo/backend/internal/user/auth/models"
	"github.com/detect-price-by-photo/backend/internal/utils"
)

// InitiatePasswordReset generates a password reset token and sends it via email.
// If a valid token already exists for the user, it updates last_sent_at instead of creating a new one.
func (s *AuthService) InitiatePasswordReset(ctx context.Context, email string) error {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		// Return success even if user doesn't exist to prevent email enumeration
		return nil
	}

	userID := user.ID

	// Check for an existing valid token for this user/email
	tokens, err := s.authRepo.FindAllOneTimeTokens(ctx)
	now := time.Now()
	var existingToken *models.OneTimeToken
	if err == nil {
		for _, t := range tokens {
			if t.UserID != nil && *t.UserID == userID &&
				t.Subject == models.OneTimeTokenSubjectPasswordReset &&
				t.RelatesTo == email &&
				now.Before(t.ExpiresAt) {
				existingToken = t
				break
			}
		}
	}

	if existingToken != nil {
		// If a valid token exists, update last_sent_at
		existingToken.LastSentAt = &now
		if err := s.authRepo.UpdateOneTimeTokenLastSentAt(ctx, existingToken.ID, now); err != nil {
			return err
		}
		// Note: We can't resend the exact token since we only store the hash
		// In production, you might want to generate a new token here
		return nil
	}

	// Generate a new, cryptographically secure, URL-safe token
	rawToken, err := apputils.GenerateURLSafeToken(48)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := now.Add(1 * time.Hour) // Password reset tokens expire in 1 hour

	// Remove any old tokens for this user/email
	for _, t := range tokens {
		if t.UserID != nil && *t.UserID == userID && t.Subject == models.OneTimeTokenSubjectPasswordReset {
			_ = s.authRepo.DeleteOneTimeToken(ctx, t.ID)
		}
	}

	// Store the new token hash in the database
	token := &models.OneTimeToken{
		UserID:     &userID,
		Subject:    models.OneTimeTokenSubjectPasswordReset,
		TokenHash:  tokenHash,
		RelatesTo:  email,
		Metadata:   nil,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastSentAt: &now,
	}
	if err := s.authRepo.CreateOneTimeToken(ctx, token); err != nil {
		return err
	}

	// Send the rawToken to the user's email address
	if err := s.sendPasswordResetEmail(ctx, email, rawToken); err != nil {
		// If sending fails, still return success to prevent email enumeration
		return nil
	}

	return nil
}

// ResetPassword validates the token and updates the user's password.
// The token is deleted after successful use (one-time use).
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" {
		return errors.New("token is required")
	}
	if newPassword == "" {
		return errors.New("new password is required")
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Retrieve the token from the database using its hash
	oneTimeToken, err := s.authRepo.GetOneTimeTokenByTokenHash(ctx, tokenHash)
	if err != nil || oneTimeToken == nil {
		return errors.New("invalid or expired token")
	}

	// Check expiration
	if time.Now().After(oneTimeToken.ExpiresAt) {
		return errors.New("token expired")
	}

	// Ensure token is bound to a user
	if oneTimeToken.UserID == nil {
		return errors.New("token not bound to a user")
	}
	userID := *oneTimeToken.UserID

	// Ensure token is for password reset
	if oneTimeToken.Subject != models.OneTimeTokenSubjectPasswordReset {
		return errors.New("invalid token type")
	}

	// Delete the token after successful validation (one-time use)
	_ = s.authRepo.DeleteOneTimeToken(ctx, oneTimeToken.ID)

	// Hash the new password
	hasher := apputils.NewPasswordHasher()
	hashedPassword, err := hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the user's password
	userPassword := &models.UserPassword{
		UserID:       userID,
		PasswordHash: hashedPassword,
	}
	if err := s.authRepo.UpdateUserPassword(ctx, userID, hashedPassword); err != nil {
		// If user doesn't have a password yet, create one
		if err := s.authRepo.SetUserPassword(ctx, userPassword); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}
	}

	return nil
}

// sendPasswordResetEmail constructs the password reset URL and sends the email using the injected mailer.
// If no mailer is configured, it logs the URL to stdout (useful for local dev).
func (s *AuthService) sendPasswordResetEmail(ctx context.Context, email, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	if s.mailer != nil {
		subject := "Reset Your Password - Detect Price by Photo"
		templateName := "password_reset.html" // Template name (can be plain text if template not available)
		
		// Try to fetch user to pass display name to template
		var displayName string
		if s.userService != nil {
			if user, err := s.userService.GetUserByEmail(ctx, email); err == nil && user != nil {
				displayName = user.DisplayName
			}
		}

		// Template data
		data := map[string]any{
			"Email":       email,
			"DisplayName": displayName,
			"ResetURL":    resetURL,
		}

		if err := s.mailer.SendEmail(ctx, []string{email}, subject, templateName, data); err != nil {
			// Fallback to plain text if template fails
			displayNameStr := ""
			if displayName != "" {
				displayNameStr = " " + displayName
			}
			body := fmt.Sprintf(`
Hello%s,

You requested to reset your password for Detect Price by Photo.

Click the link below to reset your password:
%s

This link will expire in 1 hour.

If you didn't request this password reset, please ignore this email.

Best regards,
Detect Price by Photo Team
`, displayNameStr, resetURL)
			
			// Try plain text email as fallback
			if err := s.mailer.SendEmail(ctx, []string{email}, subject, "", map[string]any{"Body": body}); err != nil {
				return fmt.Errorf("failed to send password reset email: %w", err)
			}
		}
	} else {
		// Development mode: log the reset URL
		fmt.Printf("Password reset URL (no mailer configured) for %s: %s\n", email, resetURL)
	}

	return nil
}
