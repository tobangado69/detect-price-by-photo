package validator

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
)

const (
	MaxFileSize      = 10 * 1024 * 1024 // 10MB
	AllowedMimeTypes = "image/jpeg,image/png,image/jpg"
)

var (
	ErrFileTooLarge   = errors.New("file size exceeds maximum allowed (10MB)")
	ErrInvalidMimeType = errors.New("file type not allowed (only JPEG/PNG supported)")
	ErrEmptyFile      = errors.New("file is empty")
)

// ValidateFile validates file size and MIME type.
func ValidateFile(file io.Reader, size int64, contentType string) error {
	// Check file size
	if size > MaxFileSize {
		return ErrFileTooLarge
	}

	if size == 0 {
		return ErrEmptyFile
	}

	// Parse and validate MIME type
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("invalid content type: %w", err)
	}

	// Normalize MIME type (image/jpg -> image/jpeg)
	if mediaType == "image/jpg" {
		mediaType = "image/jpeg"
	}

	// Check if MIME type is allowed
	allowedTypes := strings.Split(AllowedMimeTypes, ",")
	isAllowed := false
	for _, allowed := range allowedTypes {
		if strings.TrimSpace(allowed) == mediaType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return ErrInvalidMimeType
	}

	return nil
}

// ValidateProductName validates product name input.
func ValidateProductName(productName string) error {
	if len(strings.TrimSpace(productName)) == 0 {
		return errors.New("product name is required")
	}
	if len(productName) > 200 {
		return errors.New("product name must be less than 200 characters")
	}
	return nil
}

// ValidateCondition validates product condition.
func ValidateCondition(condition string) error {
	validConditions := []string{"new", "like_new", "good", "fair", "poor"}
	for _, valid := range validConditions {
		if condition == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid condition: must be one of %v", validConditions)
}

// ValidateMode validates estimation mode.
func ValidateMode(mode string) error {
	validModes := []string{"fast", "accurate", "knowledge_based"}
	for _, valid := range validModes {
		if mode == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid mode: must be one of %v", validModes)
}

