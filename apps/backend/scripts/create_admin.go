package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/detect-price-by-photo/backend/internal/adapter"
	"github.com/detect-price-by-photo/backend/internal/config"
	"github.com/detect-price-by-photo/backend/internal/utils"
	authRepo "github.com/detect-price-by-photo/backend/internal/user/auth/repository"
	userRepo "github.com/detect-price-by-photo/backend/internal/user/user/repository"
	authModels "github.com/detect-price-by-photo/backend/internal/user/auth/models"
	userModels "github.com/detect-price-by-photo/backend/internal/user/user/models"
	"github.com/gofrs/uuid/v5"
)

func main() {
	// Parse command line flags
	email := flag.String("email", "admin@detectprice.com", "User email address")
	password := flag.String("password", "", "User password (required)")
	name := flag.String("name", "Admin User", "User display name")
	role := flag.String("role", "admin", "User role (admin or user)")
	cfgFile := flag.String("config", "", "Config file path (optional)")
	flag.Parse()

	if *password == "" {
		fmt.Fprintf(os.Stderr, "Error: Password is required. Use -password flag\n")
		fmt.Fprintf(os.Stderr, "Usage: go run create_admin.go -email user@example.com -password yourpassword -name \"User Name\" -role admin\n")
		os.Exit(1)
	}

	// Validate role
	if *role != "admin" && *role != "user" {
		fmt.Fprintf(os.Stderr, "Error: Role must be either 'admin' or 'user'\n")
		os.Exit(1)
	}

	// Load configuration
	// Try to load .env file from current directory or parent directories
	if *cfgFile == "" {
		// Try loading .env from current directory
		if _, err := os.Stat(".env"); err == nil {
			*cfgFile = ".env"
		} else if _, err := os.Stat("../.env"); err == nil {
			*cfgFile = "../.env"
		} else if _, err := os.Stat("../../docker/.env.dev"); err == nil {
			*cfgFile = "../../docker/.env.dev"
		}
	}
	
	cfg, err := config.Load(*cfgFile)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Override DATABASE_URL if running locally (not in Docker)
	// Check if DATABASE_URL points to 'db' hostname (Docker internal)
	dbURL := cfg.Database.PostgresURL
	if strings.Contains(dbURL, "@db:") {
		// Replace 'db' with 'localhost' for local execution
		dbURL = strings.Replace(dbURL, "@db:", "@localhost:", 1)
		slog.Info("Adjusted DATABASE_URL for local execution", "url", dbURL)
	}
	
	// Also check environment variable (highest priority)
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		dbURL = envURL
		slog.Info("Using DATABASE_URL from environment variable")
	}

	// Connect to database
	pg, err := adapter.NewPostgres(adapter.PostgresConfig{
		URL: dbURL,
	})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pg.Pool.Close()

	ctx := context.Background()

	// Initialize repositories
	userRepository := userRepo.NewUserRepository(pg.Pool, slog.Default())
	authRepository := authRepo.NewAuthRepository(pg.Pool, slog.Default())

	// Check if user already exists
	existingUser, err := userRepository.GetUserByEmail(ctx, *email)
	if err == nil && existingUser != nil {
		// User exists, update role if different
		if existingUser.Role == *role {
			fmt.Printf("User %s already has role '%s'.\n", *email, *role)
		} else {
			existingUser.Role = *role
			if err := userRepository.UpdateUser(ctx, existingUser); err != nil {
				slog.Error("Failed to update user role", "error", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Updated existing user %s to role '%s'\n", *email, *role)
		}
	} else {
		// Create new user
		// Generate username from email
		username := strings.Split(*email, "@")[0]
		username = strings.ToLower(strings.ReplaceAll(username, ".", "_"))

		user := &userModels.User{
			ID:          uuid.Must(uuid.NewV7()),
			Email:       *email,
			DisplayName: *name,
			Username:    &username,
			Role:        *role,
			Metadata: &userModels.UserMetadata{
				Timezone: "Asia/Jakarta",
			},
		}

		if err := userRepository.CreateUser(ctx, user); err != nil {
			slog.Error("Failed to create user", "error", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Created user with role '%s': %s (%s)\n", *role, *email, user.ID)
	}

	// Set password
	hasher := apputils.NewPasswordHasher()
	hashedPassword, err := hasher.Hash(*password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		os.Exit(1)
	}

	// Get user ID
	user, err := userRepository.GetUserByEmail(ctx, *email)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		os.Exit(1)
	}

	userPassword := &authModels.UserPassword{
		UserID:       user.ID,
		PasswordHash: hashedPassword,
	}

	// Try to update password first, if fails, create new
	if err := authRepository.UpdateUserPassword(ctx, user.ID, hashedPassword); err != nil {
		if err := authRepository.SetUserPassword(ctx, userPassword); err != nil {
			slog.Error("Failed to set password", "error", err)
			os.Exit(1)
		}
	}

	fmt.Printf("✓ Password set for user\n")
	fmt.Printf("\n✅ User account created successfully!\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Email:    %s\n", *email)
	fmt.Printf("Password: %s\n", *password)
	fmt.Printf("Role:     %s\n", *role)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	if *role == "admin" {
		fmt.Printf("\nYou can now use this account to access admin endpoints.\n")
	}
}

