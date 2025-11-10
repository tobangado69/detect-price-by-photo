# Backend Implementation Plan - Vertical Slices

**Feature:** Backend Initialization  
**Approach:** Vertical Driven Development (VDD)  
**Created:** 2025-11-08

## Overview

This plan breaks down the backend into 6 vertical slices, each representing a complete business feature. Each slice is implemented end-to-end (API → Logic → Data) before moving to the next.

## Slice 1: Subscription Management

### Business Goal
Enable tiered subscription plans (Free/Premium/Enterprise) with daily quota tracking and automatic resets.

### Tasks

#### 1.1 Database Schema
- [ ] Create migration: `00007_create_plans_table.sql`
  - Fields: id, name, daily_photo_limit, price_idr, price_usd, features (JSONB), is_active, created_at, updated_at
- [ ] Create migration: `00008_create_subscriptions_table.sql`
  - Fields: id, user_id, plan_id, status, daily_photo_limit, current_day_usage, usage_reset_at, started_at, current_period_start, current_period_end, cancelled_at
- [ ] Add indexes: user_id, plan_id, status
- [ ] Seed initial plans (Free, Premium, Enterprise)

#### 1.2 Domain Models
- [ ] Create `internal/subscription/models/schema.go` (SQL structs)
- [ ] Create `internal/subscription/models/model.go` (domain models)
- [ ] Plan struct: ID, Name, DailyLimit, PriceIDR, Features
- [ ] Subscription struct: ID, UserID, PlanID, Status, DailyLimit, CurrentUsage, ResetAt

#### 1.3 Repository Layer
- [ ] Create `internal/subscription/repository/repository.go`
- [ ] Methods: `GetPlans()`, `GetPlanByID()`, `GetUserSubscription()`, `CreateSubscription()`, `UpdateSubscription()`, `ResetDailyUsage()`
- [ ] Use raw SQL queries with prepared statements

#### 1.4 Service Layer
- [ ] Create `internal/subscription/services/services.go`
- [ ] `SubscriptionService` struct with methods:
  - `ListPlans(ctx)` - Get all active plans
  - `GetCurrentSubscription(ctx, userID)` - Get user's active subscription
  - `CreateSubscription(ctx, userID, planID)` - Enroll user in plan
  - `CheckQuota(ctx, userID)` - Check if user has quota remaining
  - `IncrementUsage(ctx, userID)` - Increment daily usage counter
- [ ] Auto-enroll new users in Free tier (hook into user registration)

#### 1.5 API Handlers
- [ ] Create `internal/subscription/handler/handler.go`
- [ ] `GET /api/v1/subscriptions/plans` - List all plans
- [ ] `GET /api/v1/subscriptions/current` - Get current user's subscription (requires auth)
- [ ] Response format matches PRD spec

#### 1.6 Module Registration
- [ ] Create `internal/subscription/module.go`
- [ ] Register routes in `internal/server/loader.go`
- [ ] Wire dependencies (db, services, handlers)

#### 1.7 Testing
- [ ] Unit tests for repository methods
- [ ] Integration tests for API endpoints
- [ ] Test quota reset logic (daily reset at UTC+7 midnight)

---

## Slice 2: Photo Upload & Storage

### Business Goal
Accept product photos from users, validate them, and store securely in S3/MinIO.

### Tasks

#### 2.1 Storage Client
- [ ] Create `internal/storage/s3_client.go`
- [ ] Implement S3/MinIO client using AWS SDK
- [ ] Methods: `Upload(ctx, file, key)`, `GetSignedURL(ctx, key, expiry)`, `Delete(ctx, key)`
- [ ] Support both S3 and MinIO (local dev)
- [ ] Config from environment: `AWS_S3_BUCKET`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`

#### 2.2 File Validation
- [ ] Create `internal/photo/validator.go`
- [ ] Validate file size (max 10MB)
- [ ] Validate MIME type (image/jpeg, image/png)
- [ ] Extract EXIF metadata (optional, for future use)
- [ ] Virus scanning placeholder (ClamAV integration - Phase 2)

#### 2.3 Database Schema
- [ ] Create migration: `00009_create_price_analyses_table.sql`
  - Fields: id, user_id, image_url, product_name, condition, mode, estimated_price_min, estimated_price_max, estimated_price_median, confidence_score, reasoning, processing_time_ms, cost_in_usd, model_used, created_at
- [ ] Add indexes: user_id, created_at (for history pagination)

#### 2.4 Domain Models
- [ ] Create `internal/photo/models/model.go`
- [ ] `PhotoUpload` struct: File, ProductName, Condition, Mode
- [ ] `Analysis` struct: ID, UserID, ImageURL, ProductName, EstimatedPrice, Confidence, etc.

#### 2.5 Repository Layer
- [ ] Create `internal/photo/repository/repository.go`
- [ ] Methods: `CreateAnalysis()`, `GetAnalysis()`, `ListUserAnalyses()`, `DeleteAnalysis()`

#### 2.6 Service Layer
- [ ] Create `internal/photo/services/services.go`
- [ ] `PhotoService` struct:
  - `UploadPhoto(ctx, userID, file, metadata)` - Upload to S3, create DB record
  - `GetAnalysis(ctx, analysisID)` - Retrieve analysis by ID
  - `ListUserAnalyses(ctx, userID, limit, offset)` - Paginated history
- [ ] Check subscription quota before upload (call SubscriptionService)
- [ ] Generate signed URLs for image access (30-day expiry)

#### 2.7 API Handlers
- [ ] Create `internal/photo/handler/handler.go`
- [ ] `POST /api/v1/analyses/upload` - Multipart form-data upload
  - Field name: `photo` (File)
  - Fields: `product_name` (required), `condition` (optional, default: "good"), `mode` (optional, default: "fast")
- [ ] `GET /api/v1/analyses/{id}` - Get analysis details
- [ ] `GET /api/v1/analyses/history` - List user's analysis history (paginated)
- [ ] `DELETE /api/v1/analyses/{id}` - Delete analysis (soft delete)

#### 2.8 Module Registration
- [ ] Create `internal/photo/module.go`
- [ ] Register routes with auth middleware
- [ ] Wire S3 client, subscription service, photo service

#### 2.9 Testing
- [ ] Unit tests for file validation
- [ ] Integration tests for S3 upload (use MinIO in tests)
- [ ] API tests for upload endpoint
- [ ] Test quota enforcement

---

## Slice 3: Price Estimation Engine

### Business Goal
Core feature: Analyze uploaded photos using AI to estimate market prices with RAG and decision tree routing.

### Tasks

#### 3.1 AI Model Configuration
- [ ] Create `internal/ai/config.go`
- [ ] Load AI model config from database (`ai_models` table)
- [ ] Runtime registry: `GetModel(modelID)`, `GetDefaultModel()`, `GetModelForMode(mode)`
- [ ] Fallback chain support

#### 3.2 OpenRouter Client
- [ ] Create `internal/ai/openrouter_client.go`
- [ ] HTTP client for OpenRouter API
- [ ] Methods: `ChatCompletion(ctx, modelID, messages, imageURL)`
- [ ] Handle rate limits, retries, errors
- [ ] Parse response: extract price estimate, confidence, reasoning

#### 3.3 Embedding Service (RAG)
- [ ] Create `internal/ai/embedder.go`
- [ ] Generate embeddings for product queries (use OpenAI embeddings API via OpenRouter)
- [ ] Vector dimension: 1536 (OpenAI ada-002)

#### 3.4 Vector Search (RAG)
- [ ] Create `internal/ai/retriever.go`
- [ ] Query pgvector for similar market records
- [ ] SQL: `SELECT ... FROM market_knowledge WHERE embedding <=> $1 < 0.3 ORDER BY embedding <=> $1 LIMIT 5`
- [ ] Return top-5 similar products with market data

#### 3.5 Decision Tree Logic
- [ ] Create `internal/ai/estimator.go`
- [ ] `EstimatePrice(ctx, imageURL, productName, condition, mode)` method
- [ ] Decision tree:
  - **Fast mode**: Cache check → Basic rules → Quick LLM call (gpt-4o-mini)
  - **Accurate mode**: Full LLM call with image (claude-3.5-haiku)
  - **Knowledge mode**: RAG query → Market data → LLM synthesis
- [ ] Cache results in Redis (24h TTL, key: `estimate:{productName}:{condition}:{imageHash}`)

#### 3.6 Market Knowledge Base Schema
- [ ] Create migration: `00010_create_market_knowledge_table.sql`
  - Fields: id, product_name, condition, embedding (vector(1536)), market_data (JSONB), source, last_updated, created_at
- [ ] Create HNSW index: `CREATE INDEX ON market_knowledge USING hnsw (embedding vector_cosine_ops)`
- [ ] Seed initial market data (optional, for testing)

#### 3.7 Redis Integration
- [ ] Create `internal/cache/redis_client.go` (if not exists)
- [ ] Cache estimation results: `SetEstimation(key, result, ttl)`, `GetEstimation(key)`
- [ ] Cache key format: `estimate:{hash}:{product}:{condition}`

#### 3.8 Service Integration
- [ ] Update `internal/photo/services/services.go`
- [ ] Add `EstimatePrice(ctx, analysisID)` method
- [ ] Call AI estimator, store results in `price_analyses` table
- [ ] Track costs (tokens used, model used)
- [ ] Increment subscription usage

#### 3.9 API Handler Update
- [ ] Update `internal/photo/handler/handler.go`
- [ ] `POST /api/v1/analyses/estimate` - Trigger price estimation
  - Request: `{ "image_url": "...", "product_name": "...", "condition": "...", "mode": "..." }`
  - Response: Full analysis result (matches PRD spec)
- [ ] Async processing option (future: webhook notification)

#### 3.10 Testing
- [ ] Unit tests for decision tree logic
- [ ] Mock OpenRouter API responses
- [ ] Integration tests for RAG pipeline (with test pgvector data)
- [ ] Test caching behavior
- [ ] Test fallback chain

---

## Slice 4: Payment Integration

### Business Goal
Process subscription payments via Midtrans, handle webhooks, activate subscriptions on success.

### Tasks

#### 4.1 Midtrans Client
- [ ] Create `internal/payment/midtrans_client.go`
- [ ] Methods: `CreateTransaction(ctx, orderID, amount, customer)` - Returns snap token
- [ ] `ValidateWebhook(ctx, orderID, statusCode, signature)` - Verify webhook authenticity
- [ ] Config: `MIDTRANS_SERVER_KEY`, `MIDTRANS_CLIENT_KEY`, `MIDTRANS_ENV` (sandbox/production)

#### 4.2 Database Schema
- [ ] Create migration: `00011_create_invoices_table.sql`
  - Fields: id, user_id, subscription_id, amount_idr, amount_usd, period_start, period_end, status, payment_method, created_at, due_at, paid_at
- [ ] Create migration: `00012_create_payments_table.sql`
  - Fields: id, invoice_id, transaction_id (unique), amount_idr, status, payment_method, payment_channel, midtrans_response (JSONB), created_at, completed_at, expires_at
- [ ] Add indexes: transaction_id (unique), invoice_id, user_id

#### 4.3 Domain Models
- [ ] Create `internal/payment/models/model.go`
- [ ] `Invoice` struct: ID, UserID, SubscriptionID, Amount, Status, Period
- [ ] `Payment` struct: ID, InvoiceID, TransactionID, Amount, Status, MidtransResponse

#### 4.4 Repository Layer
- [ ] Create `internal/payment/repository/repository.go`
- [ ] Methods: `CreateInvoice()`, `GetInvoice()`, `UpdateInvoiceStatus()`, `CreatePayment()`, `GetPaymentByTransactionID()`, `UpdatePaymentStatus()`

#### 4.5 Service Layer
- [ ] Create `internal/payment/services/services.go`
- [ ] `PaymentService` struct:
  - `CreateSubscriptionPayment(ctx, userID, planID)` - Create invoice + Midtrans transaction
  - `HandleWebhook(ctx, payload)` - Process Midtrans webhook
  - `ActivateSubscription(ctx, paymentID)` - Activate subscription on payment success
- [ ] Idempotency: Check `transaction_id` to prevent duplicate processing

#### 4.6 API Handlers
- [ ] Create `internal/payment/handler/handler.go`
- [ ] `POST /api/v1/subscriptions/subscribe` - Create payment transaction
  - Request: `{ "plan_id": "..." }`
  - Response: `{ "snap_token": "...", "redirect_url": "...", "transaction_id": "..." }`
- [ ] `POST /api/v1/payments/webhook` - Midtrans webhook endpoint (no auth required, signature validation)
- [ ] `GET /api/v1/invoices` - List user's invoices (requires auth)

#### 4.7 Module Registration
- [ ] Create `internal/payment/module.go`
- [ ] Register routes
- [ ] Wire Midtrans client, payment service, subscription service

#### 4.8 Testing
- [ ] Unit tests for Midtrans client (mock HTTP)
- [ ] Integration tests for webhook handling
- [ ] Test idempotency (duplicate webhook handling)
- [ ] Test subscription activation flow

---

## Slice 5: Admin Dashboard APIs

### Business Goal
Admin tools for user management, analytics, system configuration.

### Tasks

#### 5.1 Admin Authorization
- [ ] Create `internal/admin/middleware.go`
- [ ] `RequireAdmin()` middleware - Check user role = "admin"
- [ ] Add `role` field to `users` table (migration: `00013_add_role_to_users.sql`)
- [ ] Seed admin user (via migration or seeder)

#### 5.2 Database Schema
- [ ] Create migration: `00014_create_audit_logs_table.sql`
  - Fields: id, admin_id, action, target_type, target_id, changes (JSONB), created_at
- [ ] Index: admin_id, target_type, created_at

#### 5.3 User Management
- [ ] Create `internal/admin/handler/user_handler.go`
- [ ] `GET /api/v1/admin/users` - List users (search, filter by plan, status, date)
- [ ] `PUT /api/v1/admin/users/{id}` - Update user (plan, status, quota reset)
- [ ] `GET /api/v1/admin/users/{id}` - Get user details
- [ ] Log all admin actions to `audit_logs`

#### 5.4 Analytics Service
- [ ] Create `internal/admin/services/analytics_service.go`
- [ ] `GetDashboard(ctx)` - Aggregate metrics:
  - Total users, active subscriptions breakdown
  - MRR (Monthly Recurring Revenue)
  - API performance (avg latency, error rate)
  - Daily active users
- [ ] `GetRevenueMetrics(ctx, period)` - Revenue trends, LTV, CAC

#### 5.5 Analytics Handlers
- [ ] Create `internal/admin/handler/analytics_handler.go`
- [ ] `GET /api/v1/admin/analytics/dashboard` - Main dashboard metrics
- [ ] `GET /api/v1/admin/analytics/revenue` - Revenue analytics
- [ ] Response format matches PRD spec

#### 5.6 Plan Management
- [ ] Create `internal/admin/handler/plan_handler.go`
- [ ] `GET /api/v1/admin/plans` - List all plans
- [ ] `POST /api/v1/admin/plans` - Create new plan
- [ ] `PUT /api/v1/admin/plans/{id}` - Update plan
- [ ] `DELETE /api/v1/admin/plans/{id}` - Soft delete plan

#### 5.7 Module Registration
- [ ] Create `internal/admin/module.go`
- [ ] Register admin routes with `RequireAdmin()` middleware
- [ ] Wire services, handlers

#### 5.8 Testing
- [ ] Unit tests for analytics calculations
- [ ] Integration tests for admin endpoints
- [ ] Test authorization (non-admin should get 403)
- [ ] Test audit logging

---

## Slice 6: AI Model Management

### Business Goal
Admin-only control over AI models: select default model, configure routing rules, manage fallback chains.

### Tasks

#### 6.1 Database Schema
- [ ] Create migration: `00015_create_ai_models_table.sql`
  - Fields: id (TEXT PRIMARY KEY), display_name, provider, cost_per_1k_tokens, average_latency_ms, modes_supported (TEXT[]), fallback_chain (TEXT[]), status, is_default (BOOLEAN), created_at, updated_at
- [ ] Unique constraint: Only one `is_default = TRUE`
- [ ] Seed initial models (gpt-4o-mini, claude-3.5-haiku, llama3.1-405b)

#### 6.2 Domain Models
- [ ] Create `internal/ai/models/model.go`
- [ ] `AIModel` struct: ID, DisplayName, Provider, Cost, Latency, ModesSupported, FallbackChain, Status, IsDefault

#### 6.3 Repository Layer
- [ ] Create `internal/ai/repository/repository.go`
- [ ] Methods: `GetCatalog()`, `GetByID()`, `GetDefault()`, `SetDefault()`, `Update()`

#### 6.4 Service Layer
- [ ] Create `internal/ai/services/model_service.go`
- [ ] `ModelService` struct:
  - `GetCatalog(ctx)` - List all models
  - `SetDefault(ctx, modelID)` - Set default model (unset previous default)
  - `Update(ctx, modelID, patch)` - Update model config (modes, fallback, status)
- [ ] Validation: Ensure default model exists and is active

#### 6.5 Runtime Registry
- [ ] Update `internal/ai/config.go`
- [ ] Load models from database on startup
- [ ] Build runtime registry: `GetModel(modelID)`, `GetDefaultModel()`, `GetModelForMode(mode)`
- [ ] Support fallback chain resolution

#### 6.6 API Handlers
- [ ] Create `internal/admin/handler/model_handler.go`
- [ ] `GET /api/v1/admin/models` - List all models with metadata
- [ ] `PUT /api/v1/admin/models/default` - Set default model
  - Request: `{ "model_id": "..." }`
- [ ] `PUT /api/v1/admin/models/{id}` - Update model config
  - Request: `{ "modes_supported": [...], "fallback_chain": [...], "status": "..." }`

#### 6.7 Integration with Estimation Engine
- [ ] Update `internal/ai/estimator.go`
- [ ] Use `ModelService.GetModelForMode(mode)` instead of hardcoded models
- [ ] Use fallback chain from model config
- [ ] Log which model was used in analysis results

#### 6.8 Module Registration
- [ ] Register admin model routes in `internal/admin/module.go`
- [ ] Wire model service

#### 6.9 Testing
- [ ] Unit tests for model service
- [ ] Integration tests for admin endpoints
- [ ] Test default model switching
- [ ] Test fallback chain execution
- [ ] Test that estimation engine uses admin-selected models

---

## Cross-Cutting Concerns

### Database Migrations
- [ ] All migrations follow naming: `000XX_description.sql`
- [ ] Include both `Up` and `Down` migrations
- [ ] Test migrations: `go run cmd/commands/migrate.go up` and `down`

### Error Handling
- [ ] Consistent error responses: `{ "error": "...", "details": {...} }`
- [ ] HTTP status codes per PRD spec
- [ ] Log errors with context (user_id, request_id, etc.)

### Configuration
- [ ] Add new env vars to `internal/config/types.go`
- [ ] Update `internal/config/default.go` with defaults
- [ ] Document in `.env.example`

### Testing Strategy
- [ ] Unit tests: Repository, Service layers (>80% coverage)
- [ ] Integration tests: API endpoints (happy path + error cases)
- [ ] Use testcontainers for PostgreSQL, Redis, MinIO in tests

### Documentation
- [ ] Update API docs (Swagger/OpenAPI) for new endpoints
- [ ] Document environment variables
- [ ] Update README with setup instructions

---

## Implementation Checklist

### Phase 1: Foundation
- [ ] Slice 1: Subscription Management
- [ ] Slice 2: Photo Upload & Storage

### Phase 2: Core Feature
- [ ] Slice 6: AI Model Management
- [ ] Slice 3: Price Estimation Engine

### Phase 3: Monetization
- [ ] Slice 4: Payment Integration

### Phase 4: Operations
- [ ] Slice 5: Admin Dashboard APIs

---

**Next Action:** Start with Slice 1 (Subscription Management). Use `/brief subscription-management` to create detailed implementation plan.

