---
title: Backend PRD
weight: 1
---

# Backend PRD: Detect Price by Photo

## 1. Executive Summary

**Product Name:** Detect Price by Photo (Backend Services)

**Version:** 1.0 MVP

**Status:** In Development

**Target Release:** Q1 2026

**Target Market:** Indonesia

**Primary Language:** Bahasa Indonesia (UI), English (API Documentation)

This document defines the backend API, database architecture, business logic, and infrastructure requirements for the **Detect Price by Photo** SaaS platform. The backend is responsible for user management, subscription handling, payment processing, AI-powered price estimation with RAG, and administrative operations.

---

## 2. Product Vision & Objectives

### 2.1 Vision Statement

Empower Indonesian online sellers, resellers, and traders to instantly estimate accurate market prices for products using AI-powered image analysis and market intelligence, enabling data-driven pricing decisions and maximizing revenue.

### 2.2 Business Objectives

1. **Revenue Growth:** Generate recurring revenue through tiered subscription models (Free, Premium, Enterprise)
2. **Market Penetration:** Capture 5% of Indonesian online sellers within 12 months
3. **Accuracy & Trust:** Deliver price estimates with >85% confidence score to build user trust
4. **Operational Efficiency:** Minimize inference costs through model optimization and decision-tree logic
5. **Scalability:** Support 100K+ concurrent users with sub-2-second response times

### 2.3 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Response Time (P95) | <2 seconds | Server logs, APM |
| Price Estimation Accuracy | >85% confidence | Historical comparison with actual sales |
| Subscription Retention Rate | >80% month-over-month | Billing system logs |
| Daily Active Users | 10,000+ | User activity database |
| System Uptime | >99.5% | Infrastructure monitoring |
| Cost per Inference | <Rp 50 | OpenRouter billing records |
| User Churn Rate | <5% per month | Subscription tracking |

---

## 3. Target Audience & User Personas

### 3.1 Primary Users

**Persona 1: Online Reseller (Agam, 28)**
- Sells secondhand electronics on marketplace platforms
- Uploads 50-100 product photos daily
- Needs quick, accurate price estimates to compete
- Willing to pay Rp 20K-50K/month for efficiency

**Persona 2: Marketplace Seller (Siti, 35)**
- Manages inventory across multiple platforms (Tokopedia, Shopee, OLX)
- Needs bulk price estimation tools
- Values API integration and batch processing
- Target segment: Premium tier (Rp 65K/month)

**Persona 3: Casual Trader (Budi, 22)**
- Occasionally sells used items (phones, laptops, furniture)
- Highly price-sensitive, prefers free tier
- May upgrade if value is proven
- Target segment: Free tier → Premium upsell

### 3.2 Secondary Users

**Admin/Support Team:**
- Manages user accounts, subscription disputes, fraud detection
- Accesses analytics dashboards and system metrics
- Manages AI model behavior and pricing configurations

---

## 4. Market Analysis & Competitive Landscape

### 4.1 Market Size

- **Indonesia e-commerce market:** USD 90.35 billion (2025), growing at 15.51% CAGR
- **Online sellers:** ~2.5 million active merchants
- **Target addressable market:** ~500K sellers who list >10 items daily
- **Pricing tool market opportunity:** Estimated 15-20% willingness to adopt AI pricing tools

### 4.2 Competitive Advantages

| Advantage | Description |
|-----------|-------------|
| **Indonesia-First Approach** | Local currency (IDR), local market knowledge, Midtrans payment integration |
| **Accuracy via RAG** | Combines image analysis + market knowledge base + real-time data |
| **Cost Optimization** | Tree-decision logic reduces API calls by 40-60% vs generic LLM solutions |
| **Multi-Model Support** | Fast/Accurate/Knowledge-based modes for different user needs |
| **Developer-Friendly** | RESTful API, comprehensive webhooks, easy integrations |

### 4.3 Competitive Positioning

- **vs. Manual pricing:** 100x faster, no human bias, scalable
- **vs. Generic price comparison tools:** Specialized for secondhand/used items with condition assessment
- **vs. International SaaS:** Localized, Rupiah-denominated, supports GoPay/QRIS

---

## 5. Core Features & Functional Requirements

### 5.1 User Management & Authentication

| Feature | Requirement | Priority |
|---------|-------------|----------|
| User Registration | Email/Phone signup, email verification, basic validation | P0 |
| Authentication | JWT-based auth with refresh tokens (15min access, 7day refresh) | P0 |
| Password Management | Secure hashing (bcrypt), forgot password flow, MFA support (future) | P0 |
| User Profiles | Name, email, phone, company name, billing address, photo quota tracking | P1 |
| Roles & Permissions | User role (basic), Admin role (advanced management), API-only keys for integrations | P1 |

**API Endpoints:**
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh-token`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/forgot-password`
- `GET /api/v1/users/me`
- `PUT /api/v1/users/profile`

### 5.2 Subscription & Billing Management

#### 5.2.1 Subscription Plans

| Plan | Photos/Day | Price (IDR/month) | Features | Target User |
|------|-----------|------------------|----------|------------|
| **Free** | 10 | 0 | 10 daily analyses, 7-day history, basic stats | Casual traders |
| **Premium** | 1,000 | 65,000 | 1K daily analyses, 90-day history, batch API, team features | Active sellers |
| **Enterprise** | Unlimited | Custom | Unlimited analyses, 1-year history, advanced API, dedicated support | Large merchants |

#### 5.2.2 Billing Features

- **Tiered pricing:** Transparent per-plan pricing visible in dashboard
- **Invoice generation:** Automatic invoice PDF generation per billing cycle
- **Payment methods:** Credit/Debit card, GoPay, QRIS (via Midtrans)
- **Refund policy:** 30-day money-back guarantee for annual plans
- **Overage handling:** Hard limit at tier's daily quota (rolls over next day)

**API Endpoints:**
- `GET /api/v1/subscriptions/plans`
- `POST /api/v1/subscriptions/subscribe` (tier upgrade/downgrade)
- `GET /api/v1/subscriptions/current`
- `GET /api/v1/invoices`
- `POST /api/v1/payments/create-transaction` (Midtrans integration)
- `POST /api/v1/payments/webhook` (Midtrans callback)

### 5.3 Photo Upload & Price Estimation

#### 5.3.1 Photo Upload

| Feature | Requirement | Priority |
|---------|-------------|----------|
| Single Upload | Max 10MB JPEG/PNG, 5MB limit after compression | P0 |
| Batch Upload | 1-20 photos per request via API | P1 |
| Cloud Storage | Store in S3/GCS, URL expiry 30 days | P0 |
| Metadata Extraction | Auto-extract EXIF data (date, dimensions, orientation) | P1 |
| Virus Scanning | ClamAV scan on upload | P1 |

**API Endpoints:**
- `POST /api/v1/analyses/upload` (single photo, `multipart/form-data` with field `photo`)
- `POST /api/v1/analyses/batch-upload` (multiple photos, `multipart/form-data` with repeated `photos[]`)
- `DELETE /api/v1/analyses/{analysis_id}`

#### 5.3.2 Price Estimation Engine

**Input Parameters:**
- `image` (primary: `multipart/form-data` file captured via browser camera/file picker; fallback: `image_base64` or `image_url` for legacy clients)
- `product_name` (required, e.g., "iPhone 13 Pro Max")
- `condition` (enum: new, like_new, good, fair, poor)
- `mode` (enum: fast, accurate, knowledge_based)

> **Note:** No `model` parameter is accepted from clients. All requests use the admin-selected default model and routing rules.

**Output Response:**
```json
{
  "analysis_id": "uuid",
  "estimated_price_min": 4500000,
  "estimated_price_max": 6200000,
  "estimated_price_median": 5350000,
  "currency": "IDR",
  "confidence_score": 0.87,
  "condition": "good",
  "product_name": "iPhone 13 Pro Max",
  "reasoning": "Based on 250+ recent sales on marketplace...",
  "market_data": {
    "total_listings_found": 250,
    "average_price": 5400000,
    "price_range": [4200000, 6800000],
    "condition_distribution": {...}
  },
  "image_analysis": {
    "detected_brand": "Apple",
    "detected_model": "iPhone 13 Pro Max",
    "physical_condition_assessment": "Minor scratches on frame, screen mint condition",
    "wear_score": 0.25
  },
  "created_at": "2025-11-08T10:30:00Z",
  "processing_time_ms": 1850
}
```

**AI System Requirements:**
- **Primary Model:** OpenRouter integration (configurable model)
- **RAG Layer:** PostgreSQL + pgvector for market knowledge base
- **Decision Tree Logic:** 3-tier routing (fast/accurate/knowledge)
- **Fallback System:** Cache-based estimation if API unavailable

**API Endpoints:**
- `POST /api/v1/analyses/estimate`
- `GET /api/v1/analyses/{analysis_id}`
- `GET /api/v1/analyses/history`

### 5.4 Admin Dashboard & Management

#### 5.4.1 User Management

- List/search all users with filters (plan, status, signup date)
- Edit user plans, reset quotas, suspend accounts
- View payment history per user
- Manual billing adjustments

#### 5.4.2 Pricing & Plan Configuration

- Create/edit/delete subscription plans
- Set daily photo quota per plan
- Adjust pricing dynamically (geolocation-based, promotional)
- View pricing tier adoption and revenue metrics

#### 5.4.3 AI Model Management

- Curate the catalog of allowed models (e.g., `gpt-4o-mini`, `claude-3.5-haiku`, `llama-3.1-405b`)
- Select the **global default model** that all customer requests use (users cannot override)
- Configure per-mode routing rules (fast/accurate/knowledge) tied to the admin-selected models
- Monitor cost per inference, token usage, and success rate per model
- Set rate limits, fallback chains, and emergency disable/enable toggles

#### 5.4.4 Analytics & Reporting

- Daily active users, subscription breakdown by tier
- Revenue dashboard (MRR, ARR, LTV, CAC)
- Price estimation accuracy metrics
- API performance (latency, error rates)
- Payment transaction logs and dispute tracking

**API Endpoints (Admin Only):**
- `GET /api/v1/admin/users` (list, search, filter)
- `PUT /api/v1/admin/users/{user_id}` (modify plan, status)
- `GET /api/v1/admin/analytics/dashboard`
- `POST /api/v1/admin/plans` (create/edit subscription plans)
- `GET /api/v1/admin/models` (list allowed models + metadata)
- `PUT /api/v1/admin/models/default` (set current default model)
- `PUT /api/v1/admin/models/{model_id}` (update routing config, fallback order)
- `GET /api/v1/admin/analytics/revenue`

---

## 6. Technical Architecture Overview

### 6.1 System Components

```
┌─────────────────────────────────────────────────────────┐
│            Frontend (Vite + React + Shadcn)             │
└──────────────────────┬──────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼────────┐  ┌──▼──────────┐   │
│  Auth API      │  │ Analysis API│   │  Admin API
│  Subscription  │  │ Estimates   │   │  Reports
└────────────────┘  └─────────────┘   │
        │              │              │
        └──────────────┼──────────────┘
                       │
        ┌──────────────▼──────────────┐
        │   Backend (Golang)          │
        │   - HTTP Server             │
        │   - Business Logic          │
        │   - RAG Pipeline            │
        └──────────────┬──────────────┘
                       │
        ┌──────────────┼──────────────┬────────────┐
        │              │              │            │
   ┌────▼──────┐ ┌────▼──────┐ ┌────▼──────┐ ┌──▼────────┐
   │ PostgreSQL│ │ pgvector  │ │  Redis    │ │ S3/GCS    │
   │ (Users,   │ │ (Market   │ │ (Cache,   │ │ (Images)  │
   │ Billing)  │ │ Knowledge)│ │ Sessions) │ │           │
   └───────────┘ └───────────┘ └───────────┘ └───────────┘
        │
        └────────────────┬────────────────┐
                         │                │
                    ┌────▼────────┐ ┌────▼──────────┐
                    │ OpenRouter  │ │  Midtrans     │
                    │ (AI Models) │ │  (Payments)   │
                    └─────────────┘ └───────────────┘
```

### 6.2 Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Backend** | Golang + stdlib + raw SQL | Fast, minimal dependencies, excellent concurrency |
| **Database** | PostgreSQL 15+ with pgvector | Relational + vector search, one system to manage |
| **Cache** | Redis | In-memory caching for sessions, rate limiting, results |
| **Storage** | S3/GCS | Scalable image storage with CDN support |
| **AI/ML** | OpenRouter API | 400+ model access, flexible switching, cost optimization |
| **Payment** | Midtrans REST API | Primary payment gateway for Indonesia |
| **Containerization** | Docker + docker-compose | Dev/prod parity, easy deployment |
| **Monitoring** | Prometheus + Grafana (optional) | Performance insights and alerting |

### 6.3 API Architecture

**Version:** v1

**Base URL:** `https://api.detectpricebyphoto.com/api/v1` (Production)

**Authentication:** Bearer token (JWT), API key for integrations

**Response Format:** JSON

**Rate Limiting:**
- Free tier: 5 requests/day
- Premium tier: 100 requests/day
- Enterprise: Custom

---

## 7. Database Schema & Data Models

### 7.1 Core Tables

#### Users Table
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  phone VARCHAR(20),
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  company_name VARCHAR(255),
  status ENUM('active', 'suspended', 'deleted'),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  last_login_at TIMESTAMP
);
```

#### Subscriptions Table
```sql
CREATE TABLE subscriptions (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  plan_id UUID REFERENCES plans(id),
  status ENUM('active', 'paused', 'cancelled'),
  daily_photo_limit INTEGER,
  current_day_usage INTEGER,
  usage_reset_at TIMESTAMP,
  started_at TIMESTAMP DEFAULT NOW(),
  current_period_start TIMESTAMP,
  current_period_end TIMESTAMP,
  cancelled_at TIMESTAMP
);
```

#### Price Analyses Table
```sql
CREATE TABLE price_analyses (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  image_url TEXT,
  product_name VARCHAR(255) NOT NULL,
  condition ENUM('new', 'like_new', 'good', 'fair', 'poor'),
  mode ENUM('fast', 'accurate', 'knowledge_based'),
  estimated_price_min BIGINT,
  estimated_price_max BIGINT,
  estimated_price_median BIGINT,
  confidence_score FLOAT,
  reasoning TEXT,
  processing_time_ms INTEGER,
  cost_in_usd FLOAT,
  model_used VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW()
);
```

#### Invoices Table
```sql
CREATE TABLE invoices (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  subscription_id UUID REFERENCES subscriptions(id),
  amount_idr BIGINT,
  amount_usd FLOAT,
  period_start DATE,
  period_end DATE,
  status ENUM('draft', 'sent', 'paid', 'overdue', 'cancelled'),
  payment_method VARCHAR(50),
  created_at TIMESTAMP DEFAULT NOW(),
  due_at TIMESTAMP,
  paid_at TIMESTAMP
);
```

#### Market Knowledge Base (Embeddings)
```sql
CREATE TABLE market_knowledge (
  id UUID PRIMARY KEY,
  product_name VARCHAR(255),
  condition VARCHAR(50),
  embedding vector(1536), -- OpenAI 3-small embedding
  market_data JSONB,  -- Contains: price range, avg price, listings count, etc.
  source VARCHAR(100), -- 'marketplace', 'manual', 'scrape'
  last_updated TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  
  CONSTRAINT market_knowledge_embedding_idx
    USING hnsw (embedding vector_cosine_ops)
);
```

#### Payments (Midtrans Integration)
```sql
CREATE TABLE payments (
  id UUID PRIMARY KEY,
  invoice_id UUID REFERENCES invoices(id),
  transaction_id VARCHAR(100) UNIQUE, -- Midtrans transaction_id
  amount_idr BIGINT,
  status ENUM('pending', 'completed', 'failed', 'expired', 'cancelled'),
  payment_method VARCHAR(50),
  payment_channel VARCHAR(50),
  midtrans_response JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  completed_at TIMESTAMP,
  expires_at TIMESTAMP
);
```

#### Subscription Plans (Admin Configurable)
```sql
CREATE TABLE plans (
  id UUID PRIMARY KEY,
  name VARCHAR(50),
  daily_photo_limit INTEGER,
  price_idr BIGINT,
  price_usd FLOAT,
  description TEXT,
  features JSONB,  -- JSON array of feature flags
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

#### Admin Audit Logs
```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY,
  admin_id UUID REFERENCES users(id),
  action VARCHAR(100), -- 'user_plan_updated', 'model_config_changed', etc.
  target_type VARCHAR(50), -- 'user', 'payment', 'model', etc.
  target_id UUID,
  changes JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);
```

### 7.2 Indexing Strategy

| Table | Index | Purpose |
|-------|-------|---------|
| users | email, phone | Auth lookup |
| subscriptions | user_id, status | User subscription query |
| price_analyses | user_id, created_at | History pagination |
| market_knowledge | embedding (HNSW) | Vector similarity search |
| payments | transaction_id | Idempotency check |

---

## 8. Business Logic & Workflows

### 8.1 Subscription Workflow

1. **User Registration**
   - User signs up with email/phone
   - Auto-enrolled in Free tier
   - Quota reset daily at 00:00 UTC+7

2. **Upgrade Flow**
   - User selects Premium plan
   - System creates payment transaction via Midtrans
   - User redirected to payment gateway
   - Webhook updates subscription status on payment success

3. **Quota Management**
   - System tracks daily usage per user
   - On each analysis: increment counter, check against limit
   - If exceeded: return 429 error with upgrade prompt
   - Daily reset runs at 00:00 UTC+7 (scheduled job)

### 8.2 Price Estimation Workflow

1. **User Uploads Photo**
   - Validate file size, format
   - Scan for viruses
   - Upload to S3/GCS
   - Store metadata in DB

2. **AI Analysis Decision Tree**
   - **Fast Mode (5-10s):** Quick cache check + basic rules (brand/model detection)
   - **Accurate Mode (15-25s):** Full LLM call to GPT-4o with image analysis
   - **Knowledge Mode (10-20s):** RAG query + market knowledge base + LLM synthesis

3. **RAG Pipeline**
   - Extract query embedding: product name + condition
   - Vector search (cosine similarity) in pgvector
   - Retrieve top-5 similar market records
   - Pass to LLM as context
   - LLM synthesizes final estimate

4. **Response & Caching**
   - Cache result in Redis for 24 hours (key: image_hash + product_name + condition)
   - Store detailed analysis in PostgreSQL
   - Decrement daily quota
   - Return estimate to user

### 8.3 Payment & Webhook Handling

1. **Create Transaction**
   - Generate Midtrans token
   - Create pending payment record
   - Return snap token to frontend

2. **Payment Webhook**
   - Validate webhook signature (Midtrans secret)
   - Check for duplicate processing (transaction_id idempotency)
   - Update payment status in DB
   - If SUCCESS: activate subscription, update user's active_subscription_id
   - If FAILED/EXPIRED: notify user, allow retry

3. **Dispute Handling**
   - Manual intervention by admin
   - Refund workflow (mark invoice as cancelled)
   - Revert subscription to previous tier

---

## 9. API Specifications

### 9.1 Authentication Endpoints

**POST /api/v1/auth/register**
```json
Request:
{
  "email": "seller@example.com",
  "password": "SecurePass123!",
  "phone": "08123456789",
  "first_name": "Agam"
}

Response: 201 Created
{
  "user_id": "uuid",
  "email": "seller@example.com",
  "subscription": {
    "plan": "free",
    "daily_limit": 10
  },
  "auth_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

**POST /api/v1/auth/login**
```json
Request:
{
  "email": "seller@example.com",
  "password": "SecurePass123!"
}

Response: 200 OK
{
  "auth_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "seller@example.com",
    "subscription_plan": "free"
  }
}
```

### 9.2 Analysis Endpoints

**POST /api/v1/analyses/estimate**
```json
Request:
{
  "image": "base64_encoded_or_url",
  "product_name": "iPhone 13 Pro Max",
  "condition": "good",
  "mode": "accurate"
}

Response: 200 OK
{
  "analysis_id": "uuid",
  "estimated_price_min": 4500000,
  "estimated_price_max": 6200000,
  "estimated_price_median": 5350000,
  "currency": "IDR",
  "confidence_score": 0.87,
  ...
}
```

**GET /api/v1/analyses/history?limit=20&offset=0**
```json
Response: 200 OK
{
  "analyses": [...],
  "total_count": 150,
  "has_more": true
}
```

### 9.3 Subscription Endpoints

**POST /api/v1/subscriptions/subscribe**
```json
Request:
{
  "plan_id": "premium-uuid"
}

Response: 200 OK
{
  "snap_token": "token_from_midtrans",
  "redirect_url": "https://snap.midtrans.com/...",
  "transaction_id": "uuid"
}
```

**GET /api/v1/subscriptions/current**
```json
Response: 200 OK
{
  "plan": "premium",
  "daily_limit": 1000,
  "current_day_usage": 342,
  "usage_reset_at": "2025-11-09T00:00:00Z",
  "current_period_end": "2025-12-08T00:00:00Z",
  "auto_renew": true
}
```

### 9.4 Payment Webhook (Midtrans)

**POST /api/v1/payments/webhook**
```json
Request (from Midtrans):
{
  "transaction_id": "abc123",
  "order_id": "ORDER-001",
  "payment_type": "gopay",
  "transaction_status": "settlement",
  "gross_amount": "65000.00",
  "signature_key": "hash..."
}

Response: 200 OK
{
  "status": "success",
  "message": "Subscription activated"
}
```

---

## 10. Admin Console API

### 10.1 User Management

**GET /api/v1/admin/users?search=email&limit=50&offset=0**

**PUT /api/v1/admin/users/{user_id}**
```json
{
  "plan_id": "premium-uuid",
  "status": "suspended",
  "notes": "Manual quota reset"
}
```

### 10.2 Analytics

**GET /api/v1/admin/analytics/dashboard**
```json
Response:
{
  "total_users": 5000,
  "active_subscriptions": {
    "free": 4200,
    "premium": 750,
    "enterprise": 50
  },
  "mrr_idr": 52500000,
  "mrr_trend": "+12%",
  "api_performance": {
    "avg_response_time_ms": 1850,
    "p95_response_time_ms": 3200,
    "error_rate": "0.02%"
  }
}
```

### 10.3 Model Configuration

**GET /api/v1/admin/models**
```json
Response:
[
  {
    "id": "openai/gpt-4o-mini",
    "display_name": "GPT-4o mini",
    "cost_per_1k_tokens_usd": 0.002,
    "latency_ms": 1800,
    "is_default": true,
    "status": "active",
    "modes_supported": ["fast", "accurate"],
    "fallback_chain": ["anthropic/claude-3-5-haiku"]
  },
  {
    "id": "anthropic/claude-3-5-haiku",
    "display_name": "Claude 3.5 Haiku",
    "cost_per_1k_tokens_usd": 0.0035,
    "latency_ms": 2200,
    "is_default": false,
    "status": "active",
    "modes_supported": ["accurate", "knowledge"],
    "fallback_chain": ["openai/gpt-4o-mini"]
  }
]
```

**PUT /api/v1/admin/models/default**
```json
Request:
{
  "model_id": "anthropic/claude-3-5-haiku"
}

Response:
{
  "message": "Default model updated",
  "default_model": "anthropic/claude-3-5-haiku"
}
```

**PUT /api/v1/admin/models/{model_id}**
```json
Request:
{
  "modes_supported": ["fast", "accurate"],
  "fallback_chain": ["openai/gpt-4o-mini"],
  "status": "active"
}

Response:
{
  "id": "anthropic/claude-3-5-haiku",
  "modes_supported": ["fast", "accurate"],
  "fallback_chain": ["openai/gpt-4o-mini"],
  "status": "active"
}
```

---

## 11. Error Handling & HTTP Status Codes

| Status | Scenario | Response |
|--------|----------|----------|
| 200 | Success | JSON payload |
| 201 | Resource created | Resource + location header |
| 400 | Invalid input | `{ "error": "validation_error", "details": [...] }` |
| 401 | Unauthorized | `{ "error": "invalid_token" }` |
| 403 | Forbidden | `{ "error": "insufficient_permissions" }` |
| 404 | Not found | `{ "error": "resource_not_found" }` |
| 429 | Rate limit exceeded | `{ "error": "quota_exceeded", "retry_after": 86400 }` |
| 500 | Server error | `{ "error": "internal_error", "trace_id": "uuid" }` |

---

## 12. Security & Compliance

### 12.1 Authentication & Authorization

- **JWT tokens:** 15-minute expiry for access tokens, 7-day for refresh tokens
- **Password requirements:** Minimum 12 characters, uppercase, lowercase, digit, special char
- **Rate limiting:** API endpoint protection (100 req/min per IP, escalating for auth endpoints)
- **HTTPS only:** All API traffic encrypted via TLS 1.3+

### 12.2 Data Privacy

- **GDPR/CCPA compliance:** User data export, deletion, consent tracking
- **Indonesia GDPR (PDP Law):** Compliance with local data protection regulations
- **PCI-DSS:** Payment handling via Midtrans (not storing card data)
- **Encryption at rest:** PostgreSQL encryption, S3 server-side encryption

### 12.3 API Security

- **Input validation:** All endpoints validate and sanitize inputs
- **SQL injection prevention:** Prepared statements for all queries
- **CSRF protection:** Token-based CSRF for state-changing operations
- **API key rotation:** Mandatory annual rotation for integrations

---

## 13. Performance & Scalability

### 13.1 Expected Load

- **Peak concurrent users:** 50K concurrent API requests
- **Daily API calls:** 5M+ analyses per day at maturity
- **Database transactions:** 10K+ TPS at peak

### 13.2 Performance Targets

| Metric | Target | Solution |
|--------|--------|----------|
| API P95 latency | <2s | Connection pooling, Redis caching, async processing |
| Database query latency | <100ms | Indexing, query optimization, read replicas |
| Price estimation accuracy | >85% | RAG refinement, model selection, feedback loops |

### 13.3 Scalability Approach

- **Horizontal scaling:** Stateless Golang service behind load balancer
- **Database:** Read replicas for analytics, write primary for transactions
- **Caching:** Redis cluster for session, results, rate limiting
- **Queue system:** (Future) Celery/Bull for async heavy lifting (batch processing)

---

## 14. Deployment & DevOps

### 14.1 Containerization

- **Docker images:** Alpine-based Go image for minimal size
- **docker-compose:** Dev environment with Postgres, Redis, MinIO
- **Kubernetes:** (Phase 2) EKS/GKE deployment with auto-scaling

### 14.2 CI/CD Pipeline

- **VCS:** GitHub with branch protection on main
- **Build:** GitHub Actions → Docker build → ECR registry
- **Testing:** Unit tests (>80% coverage), integration tests, E2E tests
- **Deployment:** Rolling updates, blue-green deployment strategy

---

## 15. Success Criteria & KPIs

### 15.1 Launch Criteria

- [ ] All P0 features implemented and tested
- [ ] API response times meet targets (P95 <2s)
- [ ] 99.5% uptime during beta
- [ ] Price estimation accuracy >85%
- [ ] Midtrans payment integration fully tested

### 15.2 Phase 2 Features (Post-Launch)

- Batch API for bulk analyses
- Bulk upload with webhook notifications
- Advanced analytics dashboard for sellers
- Integration with marketplace APIs (Tokopedia, Shopee)
- B2B API tier for enterprise customers
- Predictive pricing recommendations

---

## 16. Constraints & Assumptions

### 16.1 Constraints

- **Budget:** Minimal infrastructure costs (<$5K/month at launch)
- **Team size:** 1-2 backend developers
- **Timeline:** 8-10 weeks to MVP launch

### 16.2 Assumptions

- OpenRouter API availability and pricing stability
- PostgreSQL + pgvector adequate for market knowledge base
- Indonesian market ready for AI-powered pricing tools
- Midtrans reliable payment processing for Indonesia
- 70% of users on mobile, 30% on web

---

## 17. Risk Assessment & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|-----------|
| AI model inaccuracy | High | Medium | RAG refinement, user feedback loops, human review for disputes |
| Midtrans downtime | High | Low | Fallback payment methods, queue system for retries |
| Database scalability | High | Low | Read replicas, sharding strategy (Phase 2) |
| Market competition | Medium | Medium | Continuous model improvement, local features, pricing edge |
| Regulatory changes | Medium | Low | Compliance review quarterly, legal consultation |

---

## Appendix: Glossary

- **RAG:** Retrieval-Augmented Generation (LLM + knowledge base)
- **pgvector:** PostgreSQL extension for vector embeddings
- **Idempotency:** Same operation produces same result regardless of repetitions
- **MRR:** Monthly Recurring Revenue
- **P95/P99:** 95th/99th percentile latency
- **TPS:** Transactions Per Second

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-08  
**Next Review:** 2025-12-08