# Backend Initialization - VDD Plan

**Feature ID:** `backend-initialize`  
**Status:** Planning  
**Created:** 2025-11-08  
**Priority:** P0 (MVP Foundation)

## Business Context

Build the complete backend API for "Detect Price by Photo" SaaS platform following Vertical Driven Development principles. The backend must support:

1. **User Authentication** ✅ (Already implemented)
2. **Subscription Management** - Tiered plans, quota tracking, daily limits
3. **Photo Upload & Storage** - S3 integration, validation, virus scanning
4. **Price Estimation** - AI-powered analysis with OpenRouter, RAG, decision tree routing
5. **Payment Processing** - Midtrans integration, webhook handling, invoice generation
6. **Admin Dashboard** - User management, analytics, system configuration
7. **AI Model Management** - Admin-only model selection, routing configuration

## Current State

✅ **Completed:**
- User authentication (signin, signup, password reset, email verification)
- JWT token management (access + refresh tokens)
- Database migrations (users, sessions, passwords, tokens)
- Server infrastructure (Echo framework, middleware, config)
- PostgreSQL adapter with connection pooling

❌ **Missing:**
- Subscription/plan management
- Photo upload endpoints
- AI integration (OpenRouter)
- Payment processing (Midtrans)
- Admin APIs
- AI model management

## Vertical Slice Breakdown

Following VDD principles, we organize by **business features**, not technical layers. Each slice is independently deployable and testable.

### Slice 1: Subscription Management
**Business Value:** Enable tiered pricing (Free/Premium/Enterprise) with quota tracking

**Components:**
- Database: `plans`, `subscriptions` tables
- API: `GET /api/v1/subscriptions/plans`, `GET /api/v1/subscriptions/current`
- Business Logic: Daily quota tracking, reset at midnight UTC+7
- Integration: Auto-enroll new users in Free tier

**Dependencies:** User module (already exists)

### Slice 2: Photo Upload & Storage
**Business Value:** Accept product photos from users, validate, store in S3

**Components:**
- API: `POST /api/v1/analyses/upload` (multipart/form-data)
- Storage: S3/MinIO integration
- Validation: File size (max 10MB), format (JPEG/PNG), virus scanning
- Database: `price_analyses` table (image_url, metadata)

**Dependencies:** User auth (for user_id), Subscription (for quota check)

### Slice 3: Price Estimation Engine
**Business Value:** Core feature - analyze photos and estimate market prices

**Components:**
- API: `POST /api/v1/analyses/estimate`
- AI Integration: OpenRouter API client
- Decision Tree: Fast/Accurate/Knowledge modes
- RAG Pipeline: pgvector for market knowledge base
- Caching: Redis for result caching (24h TTL)
- Database: Store analysis results, track costs

**Dependencies:** Photo upload, Subscription (quota), AI model config

### Slice 4: Payment Integration
**Business Value:** Process subscription payments via Midtrans

**Components:**
- API: `POST /api/v1/subscriptions/subscribe`, `POST /api/v1/payments/webhook`
- Midtrans Client: Create transactions, validate webhooks
- Database: `payments`, `invoices` tables
- Business Logic: Activate subscription on payment success

**Dependencies:** Subscription module, User module

### Slice 5: Admin Dashboard APIs
**Business Value:** Admin tools for user management, analytics, system monitoring

**Components:**
- API: `GET /api/v1/admin/users`, `PUT /api/v1/admin/users/{id}`, `GET /api/v1/admin/analytics/dashboard`
- Authorization: Admin role check middleware
- Database: Audit logs table
- Business Logic: User suspension, plan changes, quota resets

**Dependencies:** User module, Subscription module

### Slice 6: AI Model Management
**Business Value:** Admin control over AI models (default selection, routing rules)

**Components:**
- Database: `ai_models` table
- API: `GET /api/v1/admin/models`, `PUT /api/v1/admin/models/default`, `PUT /api/v1/admin/models/{id}`
- Config: Runtime model registry, fallback chains
- Integration: Used by Price Estimation engine

**Dependencies:** Admin auth, Price Estimation (consumes model config)

## Implementation Order (VDD Vertical Development)

**Phase 1: Foundation (Week 1-2)**
1. ✅ User Auth (already done)
2. Slice 1: Subscription Management
3. Slice 2: Photo Upload & Storage

**Phase 2: Core Feature (Week 3-4)**
4. Slice 6: AI Model Management (needed by estimation)
5. Slice 3: Price Estimation Engine

**Phase 3: Monetization (Week 5-6)**
6. Slice 4: Payment Integration

**Phase 4: Operations (Week 7-8)**
7. Slice 5: Admin Dashboard APIs

## Technical Decisions

### Database Schema
- Follow existing migration pattern (`internal/db/migrations/`)
- Use raw SQL (no ORM) as per PRD
- Add pgvector extension for RAG

### API Structure
- Follow existing pattern: `internal/{module}/handler/`, `internal/{module}/services/`, `internal/{module}/repository/`
- Use Echo framework (already in place)
- Version: `/api/v1/`

### AI Integration
- OpenRouter API client in `internal/ai/`
- Model registry pattern for runtime model selection
- Admin-controlled default model (no user override)

### Storage
- S3/MinIO for photo storage
- Configurable via environment variables
- Generate signed URLs for 30-day expiry

## Success Criteria

- [ ] All 6 vertical slices implemented and tested
- [ ] API endpoints match PRD specifications
- [ ] Database migrations complete
- [ ] Integration tests pass
- [ ] Performance targets met (P95 <2s)
- [ ] Admin can configure AI models
- [ ] Payment webhooks process correctly

## Next Steps

1. Review this plan with team
2. Start with Slice 1 (Subscription Management)
3. Use `/brief` for each slice to create detailed implementation plans
4. Execute slices sequentially, testing each independently

---

**References:**
- Backend PRD: `docs/prd/backend-prd.md`
- Technical Guide: `docs/technical/backend-technical.md`
- VDD Principles: `.cursor/rules/vdd-core-principles.mdc`

