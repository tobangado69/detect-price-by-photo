package services

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/detect-price-by-photo/backend/internal/user/user/models"
	"github.com/detect-price-by-photo/backend/internal/user/user/repository"

	"github.com/gofrs/uuid/v5"
)

// SubscriptionServiceInterface defines the interface for subscription operations.
// This is used to avoid circular imports.
type SubscriptionServiceInterface interface {
	CreateSubscription(ctx context.Context, userID uuid.UUID, planID uuid.UUID) error
	GetPlanByName(ctx context.Context, name string) (uuid.UUID, error)
}

// UserServiceInterface defines the contract for user business logic.
type UserServiceInterface interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListUsers(ctx context.Context, filter *models.FilterUser) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
}

// Ensure UserService implements UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)

// UserService implements user business logic using a UserRepositoryInterface.
type UserService struct {
	userRepo            repository.UserRepositoryInterface
	subscriptionService SubscriptionServiceInterface
	logger              *slog.Logger
}

type UserServiceOpts struct {
	UserRepo            repository.UserRepositoryInterface
	SubscriptionService SubscriptionServiceInterface // Optional: for auto-enrollment
	Logger              *slog.Logger                 // Optional: for logging
}

// NewUserService creates a new UserService.
func NewUserService(opts UserServiceOpts) *UserService {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(nil, nil))
	}

	return &UserService{
		userRepo:            opts.UserRepo,
		subscriptionService: opts.SubscriptionService,
		logger:              logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.Must(uuid.NewV7())
	}

	if user.Metadata == nil {
		user.Metadata = &models.UserMetadata{
			Timezone: "UTC", // Set default timezone if not provided
		}
	}

	// Generate username from email if not provided
	if user.Email != "" && (user.Username == nil || *user.Username == "") {
		base := strings.SplitN(user.Email, "@", 2)[0]
		// Remove non-alphanumeric and non-underscore, make lowercase
		re := regexp.MustCompile(`[^a-z0-9_]+`)
		sanitized := re.ReplaceAllString(strings.ToLower(base), "")
		if sanitized == "" {
			sanitized = "user"
		}
		username := sanitized
		// Ensure username is unique, add suffix if needed
		suffix := 1
		for {
			exists, err := s.userRepo.UsernameExists(ctx, username)
			if err != nil {
				return err
			}
			if !exists {
				break
			}
			username = sanitized + "_" + strconv.Itoa(suffix)
			suffix++
		}
		user.Username = &username
	}

	// Create user in database
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	// Auto-enroll user in Free plan if subscription service is available
	if s.subscriptionService != nil {
		freePlanID, err := s.subscriptionService.GetPlanByName(ctx, "free")
		if err != nil {
			s.logger.Warn("failed to get free plan for auto-enrollment", "user_id", user.ID, "error", err)
			// Don't fail user creation if subscription enrollment fails
		} else {
			err := s.subscriptionService.CreateSubscription(ctx, user.ID, freePlanID)
			if err != nil {
				s.logger.Warn("failed to auto-enroll user in free plan", "user_id", user.ID, "error", err)
				// Don't fail user creation if subscription enrollment fails
			} else {
				s.logger.Info("auto-enrolled user in free plan", "user_id", user.ID)
			}
		}
	}

	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, filter *models.FilterUser) ([]*models.User, error) {
	return s.userRepo.ListUsers(ctx, filter)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.DeleteUser(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.userRepo.GetUserByUsername(ctx, username)
}

func (s *UserService) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	now := time.Now()
	user.EmailVerifiedAt = &now
	return s.userRepo.UpdateUser(ctx, user)
}
