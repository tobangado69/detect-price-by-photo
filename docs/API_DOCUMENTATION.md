# Detect Price by Photo - API Documentation

**Base URL:** `http://localhost:8000` (development)  
**API Version:** v1  
**Base Path:** `/api/v1`

## Table of Contents

1. [Authentication](#authentication)
2. [General Endpoints](#general-endpoints)
3. [User Management](#user-management)
4. [Subscriptions](#subscriptions)
5. [Photo Analysis](#photo-analysis)
6. [Payments](#payments)
7. [Admin](#admin)

---

## Authentication

### JWT Token Format

Protected endpoints require a JWT token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

### Access Token Expiry

- **Access Token:** 24 hours (default)
- **Refresh Token:** 7 days (default)

---

## General Endpoints

### Health Check

**GET** `/healthz`

Check service health status.

**Response:**
```json
{
  "status": "ok",
  "checks": {
    "database": {
      "status": "ok"
    }
  }
}
```

### OpenAPI Spec

**GET** `/api/openapi.json`

Get OpenAPI specification (placeholder).

### File Serving (Development)

**GET** `/files/*`

Serve uploaded files from local storage (development only).

**Example:** `GET /files/uploads/2024/01/15/abc123.jpg`

---

## Authentication

### Public Endpoints

#### Sign In with Email

**POST** `/api/v1/auth/signin/email`

Authenticate user with email and password.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "John Doe"
  },
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "token_expiry": "2024-01-15T10:30:00Z"
}
```

#### Sign In with Username

**POST** `/api/v1/auth/signin/username`

Authenticate user with username and password.

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "securePassword123"
}
```

**Response:** Same as email signin

#### Forgot Password

**POST** `/api/v1/auth/forgot-password`

Initiate password reset flow. Sends reset token to email.

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "message": "If an account with that email exists, a password reset link has been sent."
}
```

#### Reset Password

**POST** `/api/v1/auth/reset-password`

Reset password using token from email.

**Request Body:**
```json
{
  "token": "reset_token_from_email",
  "new_password": "newSecurePassword123"
}
```

**Response:**
```json
{
  "message": "Password has been reset successfully. You can now log in with your new password."
}
```

#### Verify Email (Link)

**GET** `/api/v1/auth/verify-email?token=<verification_token>`

Verify email address via link.

**Query Parameters:**
- `token` (required): Email verification token

**Response:**
```json
{
  "message": "Email verified successfully"
}
```

#### Initiate Email Verification

**POST** `/api/v1/auth/verification/email/initiate`

Request email verification token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "redirect_to": "https://app.example.com/dashboard" // optional
}
```

#### Validate Email Verification

**POST** `/api/v1/auth/verification/email/validate`

Validate email verification token.

**Request Body:**
```json
{
  "token": "verification_token"
}
```

#### Create Refresh Token

**POST** `/api/v1/auth/refresh-token`

Create a new refresh token.

**Request Body:**
```json
{
  "user_id": "uuid",
  "session_id": "uuid" // optional
}
```

#### Update Refresh Token

**PUT** `/api/v1/auth/refresh-token`

Update an existing refresh token.

**Request Body:**
```json
{
  "token_id": "uuid",
  "expires_at": "2024-01-22T10:30:00Z"
}
```

#### Get Refresh Token

**GET** `/api/v1/auth/refresh-token/:tokenId`

Get refresh token details.

**Path Parameters:**
- `tokenId` (required): Refresh token ID

#### Delete Refresh Token

**DELETE** `/api/v1/auth/refresh-token/:tokenId`

Revoke a refresh token.

**Path Parameters:**
- `tokenId` (required): Refresh token ID

### Protected Endpoints (Require JWT)

#### Set User Password

**POST** `/api/v1/auth/password`

Set password for authenticated user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "user_id": "uuid",
  "password": "newPassword123"
}
```

#### Update User Password

**PUT** `/api/v1/auth/password/:userId`

Update password for user (requires current password).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `userId` (required): User ID

**Request Body:**
```json
{
  "current_password": "oldPassword123",
  "new_password": "newPassword123"
}
```

#### Create Session

**POST** `/api/v1/auth/session`

Create a new user session.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "user_id": "uuid",
  "user_agent": "Mozilla/5.0...",
  "device_name": "Chrome Browser",
  "ip_address": "192.168.1.1"
}
```

#### Update Session

**PUT** `/api/v1/auth/session`

Update an existing session.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "session_id": "uuid",
  "refreshed_at": "2024-01-15T10:30:00Z"
}
```

#### Get Session

**GET** `/api/v1/auth/session/:sessionId`

Get session details.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `sessionId` (required): Session ID

#### Delete Session

**DELETE** `/api/v1/auth/session/:sessionId`

Revoke a session.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `sessionId` (required): Session ID

#### Revoke Email Verification

**POST** `/api/v1/auth/verification/email/revoke`

Revoke email verification token.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "token": "verification_token"
}
```

#### Resend Email Verification

**POST** `/api/v1/auth/verification/email/resend`

Resend email verification.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "redirect_to": "https://app.example.com/dashboard" // optional
}
```

---

## User Management

All endpoints require JWT authentication.

### Create User

**POST** `/api/v1/users`

Create a new user account.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "email": "newuser@example.com",
  "display_name": "Jane Doe",
  "username": "janedoe", // optional, auto-generated if not provided
  "phone": "+6281234567890", // optional
  "metadata": {
    "timezone": "Asia/Jakarta"
  }
}
```

**Response:**
```json
{
  "id": "uuid",
  "email": "newuser@example.com",
  "display_name": "Jane Doe",
  "username": "janedoe",
  "role": "user",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### List Users

**GET** `/api/v1/users`

List all users (with pagination and filters).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)
- `search` (optional): Search by email or display name
- `role` (optional): Filter by role (`user` or `admin`)

**Response:**
```json
{
  "users": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "display_name": "John Doe",
      "role": "user",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

### Get User

**GET** `/api/v1/users/:userId`

Get user details by ID.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `userId` (required): User ID

**Response:**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "display_name": "John Doe",
  "username": "johndoe",
  "role": "user",
  "created_at": "2024-01-15T10:30:00Z",
  "email_verified_at": "2024-01-15T11:00:00Z"
}
```

### Update User

**PUT** `/api/v1/users/:userId`

Update user information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `userId` (required): User ID

**Request Body:**
```json
{
  "display_name": "John Updated",
  "phone": "+6281234567890",
  "avatar_url": "https://example.com/avatar.jpg",
  "metadata": {
    "timezone": "Asia/Jakarta",
    "company_name": "Acme Corp"
  }
}
```

**Response:** Updated user object

### Delete User

**DELETE** `/api/v1/users/:userId`

Delete a user account.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `userId` (required): User ID

**Response:**
```json
{
  "message": "User deleted successfully"
}
```

---

## Subscriptions

### List Plans (Public)

**GET** `/api/v1/subscriptions/plans`

Get all available subscription plans.

**Response:**
```json
{
  "plans": [
    {
      "id": "uuid",
      "name": "free",
      "display_name": "Free",
      "daily_photo_limit": 10,
      "price_idr": 0,
      "price_usd": 0,
      "description": "Perfect for casual users",
      "features": ["10 daily analyses", "7-day history"],
      "is_active": true
    },
    {
      "id": "uuid",
      "name": "premium",
      "display_name": "Premium",
      "daily_photo_limit": 1000,
      "price_idr": 65000,
      "price_usd": 4.50,
      "description": "For active sellers",
      "features": ["1K daily analyses", "90-day history", "Batch API"],
      "is_active": true
    }
  ]
}
```

### Get Current Subscription (Protected)

**GET** `/api/v1/subscriptions/current`

Get current user's subscription details.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "plan_id": "uuid",
  "plan_name": "premium",
  "status": "active",
  "daily_photo_limit": 1000,
  "current_day_usage": 45,
  "usage_reset_at": "2024-01-16T00:00:00+07:00",
  "current_period_start": "2024-01-01T00:00:00Z",
  "current_period_end": "2024-02-01T00:00:00Z"
}
```

---

## Photo Analysis

All endpoints require JWT authentication.

### Upload Photo

**POST** `/api/v1/analyses/upload`

Upload a photo for price analysis.

**Headers:**
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**Form Data:**
- `photo` (file, required): Image file (JPEG/PNG, max 10MB)
- `product_name` (string, required): Product name (e.g., "iPhone 13 Pro Max")
- `condition` (string, optional): Product condition (`new`, `like_new`, `good`, `fair`, `poor`). Default: `good`
- `mode` (string, optional): Analysis mode (`fast`, `accurate`, `knowledge_based`). Default: `fast`

**Response:**
```json
{
  "analysis_id": "uuid",
  "image_url": "http://localhost:8000/files/uploads/2024/01/15/abc123.jpg",
  "product_name": "iPhone 13 Pro Max",
  "condition": "good",
  "mode": "fast",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Trigger Price Estimation

**POST** `/api/v1/analyses/:id/estimate`

Trigger AI price estimation for an uploaded photo.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (required): Analysis ID

**Response:**
```json
{
  "id": "uuid",
  "estimated_price_min": 4500000,
  "estimated_price_max": 6200000,
  "estimated_price_median": 5350000,
  "confidence_score": 0.87,
  "reasoning": "Based on 250+ recent sales on marketplace...",
  "model_used": "openai/gpt-4o-mini",
  "processing_time_ms": 1850,
  "cost_in_usd": 0.0025,
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### List User Analyses

**GET** `/api/v1/analyses/history`

Get user's analysis history.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "analyses": [
    {
      "id": "uuid",
      "product_name": "iPhone 13 Pro Max",
      "estimated_price_median": 5350000,
      "confidence_score": 0.87,
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

### Get Analysis Details

**GET** `/api/v1/analyses/:id`

Get detailed analysis information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (required): Analysis ID

**Response:**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "image_url": "http://localhost:8000/files/uploads/2024/01/15/abc123.jpg",
  "product_name": "iPhone 13 Pro Max",
  "condition": "good",
  "mode": "fast",
  "estimated_price_min": 4500000,
  "estimated_price_max": 6200000,
  "estimated_price_median": 5350000,
  "confidence_score": 0.87,
  "reasoning": "Based on 250+ recent sales...",
  "model_used": "openai/gpt-4o-mini",
  "processing_time_ms": 1850,
  "cost_in_usd": 0.0025,
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Delete Analysis

**DELETE** `/api/v1/analyses/:id`

Delete an analysis record.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (required): Analysis ID

**Response:**
```json
{
  "message": "Analysis deleted successfully"
}
```

---

## Payments

### Create Subscription Payment (Protected)

**POST** `/api/v1/subscriptions/subscribe`

Create a subscription payment transaction via Midtrans.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "plan_id": "uuid"
}
```

**Response:**
```json
{
  "snap_token": "midtrans_snap_token",
  "redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/...",
  "transaction_id": "midtrans_transaction_id",
  "invoice_id": "uuid",
  "due_at": "2024-01-22T10:30:00Z"
}
```

### List User Invoices (Protected)

**GET** `/api/v1/invoices`

Get user's invoice history.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)
- `status` (optional): Filter by status (`pending`, `paid`, `cancelled`, `refunded`)

**Response:**
```json
{
  "invoices": [
    {
      "id": "uuid",
      "subscription_id": "uuid",
      "amount_idr": 65000,
      "amount_usd": 4.50,
      "status": "paid",
      "period_start": "2024-01-01T00:00:00Z",
      "period_end": "2024-02-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "paid_at": "2024-01-01T10:30:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

### Payment Webhook (Public)

**POST** `/api/v1/payments/webhook`

Midtrans payment webhook endpoint (no auth required, signature validation).

**Request Body:** (Midtrans webhook payload)

**Response:**
```json
{
  "status": "ok"
}
```

---

## Admin

All admin endpoints require:
1. JWT authentication
2. Admin role (`role: "admin"`)

### User Management

#### List Users (Admin)

**GET** `/api/v1/admin/users`

List all users with admin filters.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**
- `page` (optional): Page number
- `limit` (optional): Items per page
- `search` (optional): Search by email/name
- `role` (optional): Filter by role
- `status` (optional): Filter by status

**Response:** Same as regular user list endpoint

#### Get User (Admin)

**GET** `/api/v1/admin/users/:id`

Get user details (admin view).

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): User ID

**Response:** User object with full details

#### Update User (Admin)

**PUT** `/api/v1/admin/users/:id`

Update user (admin can change role, status, etc.).

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): User ID

**Request Body:**
```json
{
  "role": "admin", // Can change role
  "display_name": "Updated Name",
  "metadata": {
    "banned": false
  }
}
```

**Response:** Updated user object

### Analytics

#### Get Dashboard Metrics

**GET** `/api/v1/admin/analytics/dashboard`

Get admin dashboard metrics.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Response:**
```json
{
  "total_users": 1250,
  "active_subscriptions": {
    "free": 800,
    "premium": 400,
    "enterprise": 50
  },
  "mrr_idr": 52500000,
  "mrr_trend": "+12%",
  "daily_active_users": 1500,
  "total_analyses": 10000,
  "avg_response_time_ms": 1850,
  "error_rate": "0.02%",
  "total_revenue_idr": 120000000,
  "total_revenue_usd": 8000,
  "total_payments": 250,
  "successful_payments": 245,
  "failed_payments": 5,
  "avg_order_value_idr": 489795.92,
  "avg_order_value_usd": 32.65
}
```

#### Get Revenue Metrics

**GET** `/api/v1/admin/analytics/revenue`

Get detailed revenue analytics.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**
- `period` (optional): `monthly`, `quarterly` (default: `monthly`)

**Response:**
```json
{
  "period": "monthly",
  "total_revenue_idr": 120000000,
  "total_revenue_usd": 8000,
  "mrr": 52500000,
  "arr": 630000000,
  "ltv": 1500000,
  "cac": 200000,
  "trends": {
    "month_over_month": "+5%",
    "quarter_over_quarter": "+15%"
  }
}
```

### Plan Management

#### List Plans (Admin)

**GET** `/api/v1/admin/plans`

List all subscription plans (admin view).

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Response:** Same as public plans endpoint

#### Create Plan (Admin)

**POST** `/api/v1/admin/plans`

Create a new subscription plan.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Request Body:**
```json
{
  "name": "starter",
  "display_name": "Starter Plan",
  "daily_photo_limit": 50,
  "price_idr": 30000,
  "price_usd": 2.00,
  "description": "For small sellers",
  "features": ["50 daily analyses", "30-day history"],
  "is_active": true
}
```

**Response:** Created plan object

#### Update Plan (Admin)

**PUT** `/api/v1/admin/plans/:id`

Update an existing subscription plan.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): Plan ID

**Request Body:**
```json
{
  "price_idr": 70000,
  "is_active": true
}
```

**Response:** Updated plan object

#### Delete Plan (Admin)

**DELETE** `/api/v1/admin/plans/:id`

Soft delete a subscription plan (sets `is_active: false`).

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): Plan ID

**Response:**
```json
{
  "message": "Plan deleted successfully"
}
```

### AI Model Management

#### List AI Models (Admin)

**GET** `/api/v1/admin/models`

List all configured AI models.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Response:**
```json
{
  "models": [
    {
      "id": "openai/gpt-4o-mini",
      "display_name": "GPT-4o Mini",
      "provider": "openai",
      "cost_per_1k_tokens_usd": 0.00015,
      "average_latency_ms": 1800,
      "modes_supported": ["fast", "accurate"],
      "fallback_chain": ["anthropic/claude-3-5-haiku"],
      "status": "active",
      "is_default": true,
      "max_tokens": 4000,
      "temperature": 0.7
    }
  ]
}
```

#### Set Default Model (Admin)

**PUT** `/api/v1/admin/models/default`

Set the default AI model for all user requests.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Request Body:**
```json
{
  "model_id": "anthropic/claude-3-5-haiku"
}
```

**Response:**
```json
{
  "message": "Default model updated successfully",
  "model_id": "anthropic/claude-3-5-haiku"
}
```

#### Update Model Configuration (Admin)

**PUT** `/api/v1/admin/models/:id`

Update AI model configuration.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): Model ID (e.g., `openai/gpt-4o-mini`)

**Request Body:**
```json
{
  "display_name": "GPT-4o Mini Updated",
  "modes_supported": ["fast", "accurate", "knowledge_based"],
  "fallback_chain": ["anthropic/claude-3-5-haiku", "meta-llama/llama-3.1-405b-instruct"],
  "status": "active",
  "max_tokens": 8000,
  "temperature": 0.8
}
```

**Response:** Updated model object

---

## Error Responses

All endpoints may return the following error formats:

### 400 Bad Request
```json
{
  "error": "Validation failed",
  "details": "email is required"
}
```

### 401 Unauthorized
```json
{
  "error": "authentication required"
}
```

### 403 Forbidden
```json
{
  "error": "admin access required"
}
```

### 404 Not Found
```json
{
  "error": "resource not found"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate limit exceeded",
  "retry_after": 60
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Rate Limiting

- **Default Rate Limit:** 100 requests per minute per IP
- **Burst Size:** 20 requests
- **Rate Limit Header:** `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

## Notes

1. **File Upload:** Use `multipart/form-data` for photo uploads
2. **Token Storage:** Store access tokens securely (httpOnly cookies recommended)
3. **Token Refresh:** Use refresh token to obtain new access tokens before expiry
4. **Admin Access:** Only users with `role: "admin"` can access admin endpoints
5. **Local Storage:** File serving endpoint (`/files/*`) is for development only
6. **Webhook Security:** Payment webhook validates Midtrans signature

---

**Last Updated:** 2024-01-15  
**API Version:** 1.0.0

