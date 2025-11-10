# Technical Implementation Guide: Detect Price by Photo

## 1. Project Initialization & Setup

### 1.1 Backend Setup (Golang)

**Create project structure:**
```bash
mkdir detect-price-api
cd detect-price-api

# Initialize Go modules
go mod init github.com/your-org/detect-price-api

# Create directory structure
mkdir -p cmd/server internal/{ai,payment,user,photo,db,config,utils}
mkdir -p internal/utils/testutils
mkdir -p migrations test
touch .env.example Dockerfile docker-compose.yml
```

**Install core dependencies:**
```bash
go get github.com/jackc/pgx/v5                        # PostgreSQL driver / pool
go get github.com/redis/go-redis/v9                   # Redis client
go get go.opentelemetry.io/otel/sdk                   # Observability
go get github.com/lestrrat-go/jwx/v2                  # JWT handling
go get github.com/pressly/goose/v3/cmd/goose          # Migrations
go get github.com/labstack/echo/v4                    # HTTP server
```

**Environment configuration (.env):**
```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/detect_price_db

# Redis
REDIS_URL=redis://localhost:6379

# API Keys
OPENROUTER_API_KEY=your_openrouter_key
MIDTRANS_SERVER_KEY=your_midtrans_server_key
MIDTRANS_CLIENT_KEY=your_midtrans_client_key
MIDTRANS_ENV=sandbox  # or 'production'

# JWT
JWT_SECRET=your_jwt_secret_min_32_chars_long
JWT_EXPIRY_ACCESS=900          # 15 minutes
JWT_EXPIRY_REFRESH=604800      # 7 days

# Server
PORT=8080
ENV=development  # or 'production'
CORS_ORIGIN=http://localhost:5173,https://detectpricebyphoto.com

# Cloud Storage
AWS_S3_BUCKET=detect-price-uploads
AWS_REGION=ap-southeast-1
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret

# AI Configuration (admin-controlled defaults)
AI_DEFAULT_MODEL=openai/gpt-4o-mini
AI_ALLOWED_MODELS=openai/gpt-4o-mini,anthropic/claude-3-5-haiku,meta/llama3.1-405b
AI_MODE_ROUTING=fast:openai/gpt-4o-mini|accurate:anthropic/claude-3-5-haiku|knowledge:meta/llama3.1-405b
```

### 1.2 Frontend Setup (Vite + React)

**Create project:**
```bash
npm create vite@latest detect-price-web -- --template react-ts
cd detect-price-web

# Install dependencies
npm install
npm install -D typescript @types/react @types/react-dom
npm install react-router-dom zustand axios react-hook-form zod
npm install shadcn-ui
npm install react-i18next i18next
npm install -D tailwindcss postcss autoprefixer
npm install lucide-react recharts
npx shadcn-ui@latest init
```

**Project structure:**
```
detect-price-web/
├── src/
│   ├── components/
│   │   ├── ui/              # Shadcn UI components
│   │   ├── upload/          # Upload feature components
│   │   ├── auth/            # Auth components
│   │   ├── dashboard/       # Dashboard components
│   │   └── admin/           # Admin components
│   ├── pages/
│   ├── services/            # API calls (Axios)
│   ├── stores/              # Zustand stores
│   ├── hooks/               # Custom React hooks
│   ├── types/               # TypeScript types
│   ├── locales/             # i18n translation files
│   ├── styles/              # Global CSS
│   └── App.tsx
├── .env.local
├── vite.config.ts
└── tailwind.config.js
```

---

## 2. Database Design & Setup

### 2.1 PostgreSQL Installation & Configuration

**Docker setup (docker-compose.yml for dev):**
```yaml
version: '3.9'

services:
  postgres:
    image: postgres:16-alpine
    container_name: detect_price_db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: detect_price_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    command:
      - "postgres"
      - "-c"
      - "shared_preload_libraries=pgvector"

  pgvector:
    image: pgvector/pgvector:pg16
    container_name: detect_price_pgvector
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: detect_price_db
    ports:
      - "5433:5432"
    volumes:
      - pgvector_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: detect_price_cache
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  pgvector_data:
  redis_data:
```

**Run database:**
```bash
docker-compose up -d
```

### 2.2 Database Migrations (Golang)

**Using goose for migrations:**

**Install goose:**
```bash
go get github.com/pressly/goose/v3/cmd/goose
```

**Create migration files:**
```bash
cd migrations
goose create create_users_table sql
goose create create_subscriptions_table sql
goose create create_price_analyses_table sql
goose create create_market_knowledge_table sql
goose create create_payments_table sql
goose create create_plans_table sql
```

**Example migration (001_create_users_table.sql):**
```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  phone VARCHAR(20),
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  company_name VARCHAR(255),
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_login_at TIMESTAMP,
  last_login_ip VARCHAR(45)
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_status;
DROP TABLE users;
```

**Run migrations:**
```bash
cd migrations
goose postgres "postgresql://postgres:postgres@localhost:5432/detect_price_db" up
```

### 2.3 pgvector Setup

**Install pgvector extension:**
```sql
CREATE EXTENSION IF NOT EXISTS vector;

-- Create market knowledge table with vector column
CREATE TABLE market_knowledge (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_name VARCHAR(255),
  condition VARCHAR(50),
  embedding vector(1536),
  market_data JSONB,
  source VARCHAR(100),
  last_updated TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create HNSW index for fast similarity search
CREATE INDEX ON market_knowledge 
  USING hnsw (embedding vector_cosine_ops);
```

---

## 3. Backend Implementation (Golang)

### 3.1 Core Application Structure

**cmd/server/main.go:**
```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/your-org/detect-price-api/internal/handlers"
	"github.com/your-org/detect-price-api/pkg/database"
	"github.com/your-org/detect-price-api/pkg/cache"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize database
	db, err := database.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Redis cache
	redisClient := cache.NewRedisClient(os.Getenv("REDIS_URL"))
	defer redisClient.Close()

	// Register routes
	mux := http.NewServeMux()
	
	// Health check
	mux.HandleFunc("GET /health", handlers.HealthCheck)

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/register", handlers.Register(db, redisClient))
	mux.HandleFunc("POST /api/v1/auth/login", handlers.Login(db))
	mux.HandleFunc("POST /api/v1/auth/refresh-token", handlers.RefreshToken)

	// Analysis routes
	mux.HandleFunc("POST /api/v1/analyses/estimate", handlers.EstimatePrice(db, redisClient))
	mux.HandleFunc("GET /api/v1/analyses/history", handlers.GetHistory(db))

	// Subscription routes
	mux.HandleFunc("GET /api/v1/subscriptions/plans", handlers.GetPlans(db))
	mux.HandleFunc("POST /api/v1/subscriptions/subscribe", handlers.Subscribe(db))

	// Payment routes
	mux.HandleFunc("POST /api/v1/payments/webhook", handlers.PaymentWebhook(db))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("Server error", slog.Any("error", err))
		os.Exit(1)
	}
}
```

### 3.2 Database Package

**pkg/database/postgres.go:**
```go
package database

import (
	"database/sql"
	"context"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

// Helper method for prepared statements
func (db *DB) PreparedStmt(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.PrepareContext(ctx, query)
}
```

### 3.3 Authentication Implementation

**internal/services/auth.go:**
```go
package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/your-org/detect-price-api/pkg/database"
)

type User struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Status    string
}

type AuthService struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthService(db *database.DB, jwtSecret string) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret}
}

// Register creates a new user and enrolls in Free tier
func (s *AuthService) Register(ctx context.Context, email, password, firstName string) (*User, error) {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	userID := uuid.New().String()
	query := `
		INSERT INTO users (id, email, password_hash, first_name, status, created_at)
		VALUES ($1, $2, $3, $4, 'active', CURRENT_TIMESTAMP)
		RETURNING id, email, first_name, status
	`
	
	user := &User{ID: userID}
	err = s.db.QueryRowContext(ctx, query, userID, email, hash, firstName).
		Scan(&user.ID, &user.Email, &user.FirstName, &user.Status)
	if err != nil {
		return nil, err
	}

	// Auto-enroll in Free tier
	subscriptionID := uuid.New().String()
	subscribeQuery := `
		INSERT INTO subscriptions (id, user_id, plan_id, status, daily_photo_limit, started_at)
		VALUES ($1, $2, $3, 'active', 10, CURRENT_TIMESTAMP)
	`
	// You'd fetch the Free plan ID from database or config
	if _, err := s.db.ExecContext(ctx, subscribeQuery, subscriptionID, userID, "free-plan-id"); err != nil {
		return nil, err
	}

	return user, nil
}

// GenerateTokens creates JWT access and refresh tokens
func (s *AuthService) GenerateTokens(userID string) (accessToken, refreshToken string, err error) {
	// Access token (15 minutes)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (7 days)
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
```

### 3.4 Price Estimation Service

**internal/services/estimation.go:**
```go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/your-org/detect-price-api/pkg/database"
)

type EstimationRequest struct {
	Image       string `json:"image"`
	ProductName string `json:"product_name"`
	Condition   string `json:"condition"`
	Mode        string `json:"mode"` // fast, accurate, knowledge_based
}

type EstimationResult struct {
	AnalysisID         string  `json:"analysis_id"`
	EstimatedPriceMin  int64   `json:"estimated_price_min"`
	EstimatedPriceMax  int64   `json:"estimated_price_max"`
	EstimatedPriceMedian int64 `json:"estimated_price_median"`
	ConfidenceScore    float64 `json:"confidence_score"`
	Reasoning          string  `json:"reasoning"`
	ProcessingTimeMs   int64   `json:"processing_time_ms"`
}

type EstimationService struct {
	db          *database.DB
	redis       *redis.Client
	openrouterKey string
}

func NewEstimationService(db *database.DB, redis *redis.Client, openrouterKey string) *EstimationService {
	return &EstimationService{
		db:          db,
		redis:       redis,
		openrouterKey: openrouterKey,
	}
}

// EstimatePrice handles the price estimation logic with decision tree
func (s *EstimationService) EstimatePrice(ctx context.Context, req EstimationRequest, userID string) (*EstimationResult, error) {
	startTime := time.Now()

	// Step 1: Check cache
	cacheKey := fmt.Sprintf("estimate:%s:%s:%s", req.ProductName, req.Condition, hashImage(req.Image))
	cachedResult, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var result EstimationResult
		json.Unmarshal([]byte(cachedResult), &result)
		return &result, nil
	}

	// Step 2: Decision tree - choose inference path
	var estimatedPrice int64
	var confidence float64
	var reasoning string

	switch req.Mode {
	case "fast":
		// Fast mode: Quick heuristic (2-5 seconds)
		estimatedPrice, confidence, reasoning = s.fastEstimate(ctx, req)
	case "accurate":
		// Accurate mode: Full LLM call (15-25 seconds)
		estimatedPrice, confidence, reasoning = s.accurateEstimate(ctx, req)
	case "knowledge_based":
		// Knowledge mode: RAG + LLM (10-20 seconds)
		estimatedPrice, confidence, reasoning = s.ragEstimate(ctx, req)
	default:
		estimatedPrice, confidence, reasoning = s.fastEstimate(ctx, req)
	}

	// Step 3: Cache result for 24 hours
	result := &EstimationResult{
		AnalysisID:         generateAnalysisID(),
		EstimatedPriceMin:  int64(float64(estimatedPrice) * 0.85),
		EstimatedPriceMax:  int64(float64(estimatedPrice) * 1.15),
		EstimatedPriceMedian: estimatedPrice,
		ConfidenceScore:    confidence,
		Reasoning:          reasoning,
		ProcessingTimeMs:   time.Since(startTime).Milliseconds(),
	}

	// Cache the result
	resultJSON, _ := json.Marshal(result)
	s.redis.Set(ctx, cacheKey, string(resultJSON), 24*time.Hour)

	// Step 4: Store in database
	s.storeAnalysis(ctx, userID, result)

	return result, nil
}

// ragEstimate uses RAG (Retrieval-Augmented Generation) approach
func (s *EstimationService) ragEstimate(ctx context.Context, req EstimationRequest) (int64, float64, string) {
	// 1. Generate embedding for query
	embedding := s.generateEmbedding(ctx, req.ProductName+" "+req.Condition)

	// 2. Vector search in pgvector for similar products
	marketData := s.vectorSearch(ctx, embedding)

	// 3. Create context from market data
	context_str := fmt.Sprintf("Similar market data:\n%s\nProduct: %s, Condition: %s\nEstimate price based on market data.",
		marketData, req.ProductName, req.Condition)

	// 4. Call LLM with context
	response := s.callOpenRouter(ctx, context_str, req)

	// Parse LLM response
	price, confidence := parseOpenRouterResponse(response)

	return price, confidence, response
}

// callOpenRouter makes API call to OpenRouter
func (s *EstimationService) callOpenRouter(ctx context.Context, context_str string, req EstimationRequest) string {
	payload := map[string]interface{}{
		"model": "openai/gpt-4o",
		"messages": []map[string]string{
			{
				"role": "user",
				"content": fmt.Sprintf(`You are a professional appraiser. Analyze this product image and estimate its market price.
				Product: %s
				Condition: %s
				Context: %s
				Image: %s
				Respond with JSON: {"estimated_price": <number>, "confidence": <0-1>, "reasoning": "<string>"}`,
					req.ProductName, req.Condition, context_str, req.Image),
			},
		},
		"temperature": 0.7,
		"max_tokens":  500,
	}

	reqBody, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", 
		io.NopCloser(bytes.NewBuffer(reqBody)))
	httpReq.Header.Set("Authorization", "Bearer "+s.openrouterKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(httpReq)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func (s *EstimationService) vectorSearch(ctx context.Context, embedding []float32) string {
	// Query pgvector for similar market records
	query := `
		SELECT market_data FROM market_knowledge
		WHERE 1 - (embedding <=> $1::vector) > 0.7
		ORDER BY embedding <=> $1::vector
		LIMIT 5
	`
	
	// Execute query and return results
	// (implementation details omitted for brevity)
	return ""
}
```

### 3.5 Payment Integration (Midtrans)

**pkg/payment/midtrans.go:**
```go
package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type MidtransClient struct {
	serverKey string
	clientKey string
	env       string // "sandbox" or "production"
}

func NewMidtransClient(serverKey, clientKey, env string) *MidtransClient {
	return &MidtransClient{
		serverKey: serverKey,
		clientKey: clientKey,
		env:       env,
	}
}

type CreateTransactionRequest struct {
	OrderID      string
	Amount       int64
	Email        string
	Phone        string
	FirstName    string
	LastName     string
	Description  string
}

type CreateTransactionResponse struct {
	Token      string `json:"token"`
	RedirectURL string `json:"redirect_url"`
	TransactionID string `json:"transaction_id"`
}

func (m *MidtransClient) CreateTransaction(req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	url := fmt.Sprintf("https://app.%s.midtrans.com/snap/v1/transactions", m.env)

	payload := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     req.OrderID,
			"gross_amount": req.Amount,
		},
		"customer_details": map[string]interface{}{
			"email":      req.Email,
			"phone":      req.Phone,
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		},
		"item_details": []map[string]interface{}{
			{
				"id":       "premium-subscription",
				"price":    req.Amount,
				"quantity": 1,
				"name":     req.Description,
			},
		},
	}

	body, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(m.serverKey, "")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result CreateTransactionResponse
	json.Unmarshal(respBody, &result)

	return &result, nil
}

// ValidateWebhook verifies Midtrans webhook signature
func (m *MidtransClient) ValidateWebhook(orderId, statusCode, signedKey string) bool {
	expectedSignature := fmt.Sprintf("%s%s%s", orderId, statusCode, m.serverKey)
	// Hash with SHA512 and compare with signedKey
	// (implementation omitted for brevity)
	return true
}
```

---

## 4. Frontend Implementation (React + Vite)

### 4.1 API Client Setup

**src/services/api.ts:**
```typescript
import axios, { AxiosInstance } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add JWT token to requests
apiClient.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      try {
        const refreshToken = sessionStorage.getItem('refresh_token');
        const response = await axios.post(`${API_BASE_URL}/auth/refresh-token`, {
          refresh_token: refreshToken,
        });
        sessionStorage.setItem('access_token', response.data.access_token);
        // Retry original request
        return apiClient(error.config);
      } catch {
        sessionStorage.clear();
        window.location.href = '/auth/login';
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

### 4.2 Zustand Store Setup

**src/stores/authStore.ts:**
```typescript
import { create } from 'zustand';
import apiClient from '@/services/api';

interface User {
  id: string;
  email: string;
  first_name: string;
}

interface AuthStore {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, firstName: string) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isAuthenticated: false,
  loading: false,

  login: async (email: string, password: string) => {
    set({ loading: true });
    try {
      const response = await apiClient.post('/auth/login', { email, password });
      sessionStorage.setItem('access_token', response.data.auth_token);
      sessionStorage.setItem('refresh_token', response.data.refresh_token);
      set({ user: response.data.user, isAuthenticated: true });
    } catch (error) {
      throw error;
    } finally {
      set({ loading: false });
    }
  },

  register: async (email: string, password: string, firstName: string) => {
    set({ loading: true });
    try {
      const response = await apiClient.post('/auth/register', { email, password, first_name: firstName });
      sessionStorage.setItem('access_token', response.data.auth_token);
      sessionStorage.setItem('refresh_token', response.data.refresh_token);
      set({ user: response.data.user, isAuthenticated: true });
    } finally {
      set({ loading: false });
    }
  },

  logout: () => {
    sessionStorage.clear();
    set({ user: null, isAuthenticated: false });
  },
}));
```

**src/stores/subscriptionStore.ts:**
```typescript
import { create } from 'zustand';
import apiClient from '@/services/api';

interface Subscription {
  plan: 'free' | 'premium' | 'enterprise';
  daily_limit: number;
  current_day_usage: number;
  usage_reset_at: string;
}

interface SubscriptionStore {
  subscription: Subscription | null;
  loading: boolean;
  fetchSubscription: () => Promise<void>;
  incrementUsage: () => void;
}

export const useSubscriptionStore = create<SubscriptionStore>((set) => ({
  subscription: null,
  loading: false,

  fetchSubscription: async () => {
    set({ loading: true });
    try {
      const response = await apiClient.get('/subscriptions/current');
      set({ subscription: response.data });
    } finally {
      set({ loading: false });
    }
  },

  incrementUsage: () => {
    set((state) => ({
      subscription: state.subscription
        ? { ...state.subscription, current_day_usage: state.subscription.current_day_usage + 1 }
        : null,
    }));
  },
}));
```

### 4.3 Component: Photo Upload

**src/components/upload/PhotoUploadCard.tsx:**
```typescript
import { useState, useRef } from 'react';
import { Upload } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';

interface PhotoUploadCardProps {
  onPhotoSelect: (file: File, preview: string) => void;
  isLoading?: boolean;
}

export const PhotoUploadCard: React.FC<PhotoUploadCardProps> = ({ onPhotoSelect, isLoading }) => {
  const [isDragActive, setIsDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { toast } = useToast();

  const handleFile = (file: File) => {
    // Validate file
    if (!file.type.startsWith('image/')) {
      toast({ variant: 'destructive', title: 'Invalid file type. Please upload an image.' });
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      toast({ variant: 'destructive', title: 'File too large. Max 10MB.' });
      return;
    }

    // Create preview
    const reader = new FileReader();
    reader.onload = (e) => {
      const preview = e.target?.result as string;
      onPhotoSelect(file, preview);
    };
    reader.readAsDataURL(file);
  };

  return (
    <div
      onDragOver={() => setIsDragActive(true)}
      onDragLeave={() => setIsDragActive(false)}
      onDrop={(e) => {
        e.preventDefault();
        const files = e.dataTransfer.files;
        if (files.length > 0) handleFile(files[0]);
      }}
      className={`border-2 border-dashed rounded-lg p-8 text-center transition ${
        isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300'
      }`}
    >
      <Upload className="mx-auto h-12 w-12 text-gray-400 mb-4" />
      <p className="text-lg font-medium mb-2">Upload Product Photo</p>
      <p className="text-sm text-gray-500 mb-4">Drag and drop or click to select</p>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        onChange={(e) => e.target.files && handleFile(e.target.files[0])}
        className="hidden"
      />
      <Button onClick={() => fileInputRef.current?.click()} disabled={isLoading}>
        Select Photo
      </Button>
    </div>
  );
};
```

### 4.4 Component: Price Result

**src/components/upload/PriceEstimateResult.tsx:**
```typescript
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Share2 } from 'lucide-react';

interface Analysis {
  estimated_price_min: number;
  estimated_price_max: number;
  estimated_price_median: number;
  confidence_score: number;
  reasoning: string;
}

export const PriceEstimateResult: React.FC<{ analysis: Analysis }> = ({ analysis }) => {
  const confidenceColor =
    analysis.confidence_score >= 0.8
      ? 'bg-green-100 text-green-800'
      : analysis.confidence_score >= 0.6
      ? 'bg-yellow-100 text-yellow-800'
      : 'bg-red-100 text-red-800';

  return (
    <Card>
      <CardHeader>
        <CardTitle>Price Estimate</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="bg-blue-50 p-4 rounded-lg">
          <p className="text-sm text-gray-600">Estimated Price</p>
          <p className="text-3xl font-bold text-blue-600">
            Rp {analysis.estimated_price_median.toLocaleString('id-ID')}
          </p>
          <p className="text-xs text-gray-500 mt-1">
            Range: Rp {analysis.estimated_price_min.toLocaleString('id-ID')} - Rp{' '}
            {analysis.estimated_price_max.toLocaleString('id-ID')}
          </p>
        </div>

        <div className="flex items-center justify-between">
          <p className="text-sm font-medium">Confidence Score</p>
          <Badge className={confidenceColor}>
            {(analysis.confidence_score * 100).toFixed(0)}%
          </Badge>
        </div>

        <div>
          <p className="text-sm font-medium mb-2">Analysis Reasoning</p>
          <p className="text-sm text-gray-600">{analysis.reasoning}</p>
        </div>

        <Button variant="outline" className="w-full">
          <Share2 className="h-4 w-4 mr-2" />
          Share Result
        </Button>
      </CardContent>
    </Card>
  );
};
```

### 4.5 Routing Setup

**src/App.tsx:**
```typescript
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { AppLayout } from '@/components/layout/AppLayout';
import { Dashboard } from '@/pages/Dashboard';
import { Upload } from '@/pages/Upload';
import { Login } from '@/pages/auth/Login';
import { Register } from '@/pages/auth/Register';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return isAuthenticated ? <>{children}</> : <Navigate to="/auth/login" />;
};

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/auth/login" element={<Login />} />
        <Route path="/auth/register" element={<Register />} />

        <Route
          element={
            <ProtectedRoute>
              <AppLayout />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<Dashboard />} />
          <Route path="/upload" element={<Upload />} />
          {/* More routes */}
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
```

---

## 5. Docker & Deployment

### 5.1 Backend Dockerfile

**Dockerfile:**
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY migrations ./migrations

# Expose port
EXPOSE 8080

# Run server
CMD ["./server"]
```

### 5.2 Frontend Build & Deploy

**Vite build:**
```bash
npm run build  # Creates dist/ directory
```

**Deploy to Vercel:**
```bash
npm install -g vercel
vercel deploy --prod
```

---

## 6. Running the Stack Locally

```bash
# Start database and cache
docker-compose up -d

# Run migrations
cd migrations && goose postgres "postgresql://postgres:postgres@localhost:5432/detect_price_db" up

# Start backend (in another terminal)
cd backend
go run cmd/server/main.go

# Start frontend (in another terminal)
cd frontend
npm install
npm run dev  # Vite dev server on http://localhost:5173
```

---

## 7. Testing

### 7.1 Backend Testing

**Unit test example (auth_test.go):**
```go
package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/detect-price-api/pkg/database"
)

func TestRegister(t *testing.T) {
	// Setup
	db := setupTestDB()
	defer db.Close()
	svc := NewAuthService(db, "test-secret")

	// Test
	user, err := svc.Register(context.Background(), "test@example.com", "SecurePass123!", "Test")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "active", user.Status)
}
```

### 7.2 Frontend Testing

**Component test example:**
```typescript
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PhotoUploadCard } from '@/components/upload/PhotoUploadCard';

test('uploads photo successfully', async () => {
  const onSelect = jest.fn();
  render(<PhotoUploadCard onPhotoSelect={onSelect} />);

  const button = screen.getByText('Select Photo');
  await userEvent.click(button);

  // Assert
  expect(onSelect).toHaveBeenCalled();
});
```

---

## 8. Performance Optimization Checklist

- [ ] Enable Redis caching for API responses
- [ ] Implement database query indexing
- [ ] Use prepared statements for all DB queries
- [ ] Implement connection pooling (Postgres & Redis)
- [ ] Lazy load images in history view
- [ ] Code-split React Router pages
- [ ] Monitor bundle size (<300KB gzip)
- [ ] Set up CDN for static assets
- [ ] Enable CORS caching headers

---

## 9. Monitoring & Observability

**Structured logging (slog):**
```go
logger.Info("Analysis created",
  slog.String("user_id", userID),
  slog.String("analysis_id", analysisID),
  slog.Int64("estimated_price", estimatedPrice),
  slog.Float64("confidence_score", confidenceScore),
)
```

**Frontend error tracking (Sentry):**
```typescript
import * as Sentry from "@sentry/react";

Sentry.init({
  dsn: process.env.VITE_SENTRY_DSN,
  environment: process.env.VITE_ENV,
});
```

---

## 10. Deployment Checklist

- [ ] Environment variables configured
- [ ] Database migrations run
- [ ] SSL certificates installed
- [ ] CORS origins configured
- [ ] Rate limiting enabled
- [ ] Monitoring & logging active
- [ ] Backup strategy implemented
- [ ] Disaster recovery plan documented
- [ ] Load testing completed
- [ ] Security audit passed

---

## 11. AI Model Management Reference

### 11.1 Environment Variables

```
AI_DEFAULT_MODEL=openai/gpt-4o-mini
AI_ALLOWED_MODELS=openai/gpt-4o-mini,anthropic/claude-3-5-haiku,meta/llama3.1-405b
AI_MODE_ROUTING=fast:openai/gpt-4o-mini|accurate:anthropic/claude-3-5-haiku|knowledge:meta/llama3.1-405b
```

These values are used as bootstraps. On startup the backend loads the `ai_models` table, applies overrides from the env, then warms a runtime cache (`config.AI`).
Client requests never include a `model` override; every analysis uses this central configuration.

### 11.2 Database Schema

```sql
CREATE TABLE ai_models (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  provider TEXT NOT NULL,
  cost_per_1k_tokens NUMERIC(10,4),
  average_latency_ms INTEGER,
  modes_supported TEXT[] DEFAULT '{}',
  fallback_chain TEXT[] DEFAULT '{}',
  status TEXT DEFAULT 'active',
  is_default BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ai_models_default_true
  ON ai_models (is_default) WHERE is_default = TRUE;
```

### 11.3 Service Interface

```go
type AIModel struct {
	ID             string
	DisplayName    string
	ModesSupported []string
	FallbackChain  []string
	Status         string
	IsDefault      bool
}

type ModelService interface {
	GetCatalog(ctx context.Context) ([]AIModel, error)
	SetDefault(ctx context.Context, modelID string) error
	Update(ctx context.Context, modelID string, patch UpdateModelRequest) (*AIModel, error)
}

type UpdateModelRequest struct {
	ModesSupported []string `json:"modes_supported"`
	FallbackChain  []string `json:"fallback_chain"`
	Status         string   `json:"status"`
}
```

### 11.4 Echo Handlers

```go
func (h *AdminHandler) ListModels(c echo.Context) error {
	models, err := h.modelService.GetCatalog(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, models)
}

func (h *AdminHandler) SetDefaultModel(c echo.Context) error {
	var payload struct {
		ModelID string `json:"model_id" validate:"required"`
	}
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if err := h.modelService.SetDefault(c.Request().Context(), payload.ModelID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message":       "Default model updated",
		"default_model": payload.ModelID,
	})
}
```

### 11.5 Runtime Usage

```go
modelID := cfg.AI.DefaultModel
if override, ok := cfg.AI.ModeRouting[requestedMode]; ok {
	modelID = override
}

client := registry.Get(modelID)
resp, err := client.Estimate(ctx, payload)
if err != nil {
	for _, fallback := range cfg.AI.FallbackChain(modelID) {
		if ret, fallbackErr := registry.Get(fallback).Estimate(ctx, payload); fallbackErr == nil {
			resp = ret
			break
		}
	}
}
```

---

**Document Version:** 1.1  
**Last Updated:** 2025-11-08  
**Next Review:** 2025-12-08