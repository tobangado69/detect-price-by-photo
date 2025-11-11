# Detect Price by Photo - API Documentation

**Base URL:** `http://localhost:8080` (development - Docker port mapping)  
**API Version:** v1  
**Base Path:** `/api/v1`

> **Note:** The backend runs on port 8000 inside Docker, but is exposed on port 8080 on the host machine.

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

**Note:** In production, files should be served via S3/CDN, not through the API server.

---

## Authentication

### Public Endpoints

#### Sign Up

**POST** `/api/v1/auth/signup`

Register a new user account. No authentication required.

**Request Body:**
```json
{
  "display_name": "John Doe",
  "email": "user@example.com",
  "password": "StrongPassword123",
  "password_confirmation": "StrongPassword123",
  "redirect_to": "https://app.example.com/welcome"
}
```

**Response:**
```json
{
  "message": "Account created successfully. Please verify your email address.",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "John Doe",
    "username": "user",
    "role": "user"
  }
}
```

**Notes:**
- `display_name` is optional; if omitted, it is derived from the email.
- A verification email is sent automatically (if SMTP is configured).
- Password must be at least 12 characters and match `password_confirmation`.

#### Sign In with Email

**POST** `/api/v1/auth/signin/email`

Authenticate user with email and password.

**Request Body:**
```json
{
  "email": "admin@detectprice.com",
  "password": "admin123"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "admin@detectprice.com",
    "display_name": "System Administrator"
  },
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "token_expiry": "2024-01-15T10:30:00Z"
}
```

**Status Codes:**
- `200 OK`: Authenticated successfully
- `401 Unauthorized`: Invalid credentials or email not verified
- `403 Forbidden`: Account suspended (see `ban_reason` in admin panel)
- `500 Internal Server Error`: Unexpected failure

**Status Codes:**
- `200 OK`: Authenticated successfully
- `401 Unauthorized`: Invalid credentials or email not verified
- `403 Forbidden`: Account suspended
- `500 Internal Server Error`: Unexpected failure

#### Sign In with Username

**POST** `/api/v1/auth/signin/username`

Authenticate user with username and password.

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response:** Same as email signin

#### Forgot Password

**POST** `/api/v1/auth/forgot-password`

Initiate password reset flow. Sends reset token to email.

**Request Body:**
```json
{
  "email": "admin@detectprice.com"
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

**Validation:**
- `token` (required): Password reset token from email
- `new_password` (required, min: 12 characters): New password

**Response (Success):**
```json
{
  "message": "Password has been reset successfully. You can now log in with your new password."
}
```

**Response (Error):**
```json
{
  "error": "Invalid or expired token"
}
```

**Status Codes:**
- `200 OK`: Password reset successful
- `400 Bad Request`: Invalid token or validation error

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
  "email": "admin@detectprice.com",
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

Revoke email verification token (protected endpoint).

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

**Response:**
```json
{
  "message": "Verification token revoked"
}
```

#### Resend Email Verification

**POST** `/api/v1/auth/verification/email/resend`

Resend email verification (protected endpoint).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "email": "admin@detectprice.com",
  "redirect_to": "https://app.example.com/dashboard" // optional
}
```

**Response:**
```json
{
  "message": "Verification email resent if the email is registered"
}
```

**Status Codes:**
- `200 OK`: Email sent successfully
- `400 Bad Request`: Validation error
- `404 Not Found`: User not found
- `409 Conflict`: Token still valid (check email)

---

## User Management

All endpoints require JWT authentication and are intended for authenticated/admin usage. Public self-service registration should use **POST `/api/v1/auth/signup`**.

### Create User

**POST** `/api/v1/users` *(Authenticated/Admin)*

Create a new user account via the administrative API. Useful for back-office tooling.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "email": "newuser@example.com",
  "display_name": "Jane Doe",
  "username": "janedoe",
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
  "amount_idr": 65000,
  "amount_usd": 4.50
}
```

**Status Codes:**
- `200 OK`: Payment transaction created successfully
- `400 Bad Request`: Invalid plan ID or validation error
- `401 Unauthorized`: Authentication required
- `500 Internal Server Error`: Payment gateway error

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

List all users with admin filters and pagination.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**
- `search` (optional): Search by email or display name
- `limit` (optional): Items per page (default: 50, max: 100)
- `offset` (optional): Offset for pagination (default: 0)

**Response:**
```json
{
  "users": [
    {
      "id": "uuid",
      "email": "admin@detectprice.com",
      "display_name": "John Doe",
      "username": "johndoe",
      "role": "user",
      "created_at": "2024-01-15T10:30:00Z",
      "email_verified_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total": 1250,
  "limit": 50,
  "offset": 0
}
```

**Status Codes:**
- `200 OK`: Users retrieved successfully
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `500 Internal Server Error`: Failed to retrieve users

#### Get User (Admin)

**GET** `/api/v1/admin/users/:id`

Get user details (admin view with full information).

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): User ID (UUID)

**Response:**
```json
{
  "id": "uuid",
  "email": "admin@detectprice.com",
  "display_name": "John Doe",
  "username": "johndoe",
  "role": "user",
  "avatar_url": null,
  "metadata": {
    "timezone": "Asia/Jakarta"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": null,
  "email_verified_at": "2024-01-15T11:00:00Z",
  "last_login_at": "2024-01-20T08:00:00Z",
  "banned_at": null,
  "ban_expires": null,
  "ban_reason": null
}
```

**Status Codes:**
- `200 OK`: User retrieved successfully
- `400 Bad Request`: Invalid user ID format
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `404 Not Found`: User not found
- `500 Internal Server Error`: Failed to retrieve user

#### Update User (Admin)

**PUT** `/api/v1/admin/users/:id`

Update user information. Admin can change role, plan, and account suspension state. All changes are logged in audit logs.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Path Parameters:**
- `id` (required): User ID (UUID)

**Request Body:**
```json
{
  "role": "admin",
  "status": "suspended",
  "ban_reason": "Fraudulent activity detected",
  "ban_expires": "2025-01-31T23:59:59Z"
}
```

**Status Behaviour**
- `status: "suspended"` requires a non-empty `ban_reason`. Optionally provide `ban_expires` (RFC3339) to auto-lift the ban.
- `status: "active"` clears any existing ban (`banned_at`, `ban_expires`, `ban_reason`).
- When a ban expires naturally, the next login attempt automatically clears the ban and proceeds if credentials are valid.

**Response:**
```json
{
  "id": "uuid",
  "email": "admin@detectprice.com",
  "display_name": "John Doe",
  "username": "johndoe",
  "role": "admin",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-20T12:00:00Z"
}
```

**Status Codes:**
- `200 OK`: User updated successfully
- `400 Bad Request`: Invalid user ID or request body (e.g., invalid role)
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `404 Not Found`: User not found
- `500 Internal Server Error`: Failed to update user

**Note:** All admin actions are automatically logged in the audit log system for accountability.

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
  "mrr": 52500000.0,
  "daily_active_users": 150,
  "total_analyses": 10000,
  "analyses_today": 245,
  "average_latency_ms": 1850.5,
  "error_rate": 0.02
}
```

**Status Codes:**
- `200 OK`: Dashboard metrics retrieved successfully
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `500 Internal Server Error`: Failed to retrieve metrics

**Note:** 
- `mrr` is Monthly Recurring Revenue in IDR (sum of all active subscription plan prices)
- `error_rate` is a decimal between 0.0 and 1.0 (e.g., 0.02 = 2%)
- `daily_active_users` counts users who created a session today

#### Get Revenue Metrics

**GET** `/api/v1/admin/analytics/revenue`

Get detailed revenue analytics.

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**
- `period` (optional): `daily`, `weekly`, or `monthly` (default: `monthly`)

**Response:**
```json
{
  "period": "monthly",
  "total_revenue_idr": 120000000.0,
  "total_revenue_usd": 8000.0,
  "transactions": 250,
  "average_order_value": 480000.0,
  "trend": [
    {
      "date": "2024-01",
      "revenue_idr": 10000000.0,
      "revenue_usd": 666.67,
      "transactions": 20
    },
    {
      "date": "2024-02",
      "revenue_idr": 12000000.0,
      "revenue_usd": 800.0,
      "transactions": 25
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Revenue metrics retrieved successfully
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `500 Internal Server Error`: Failed to retrieve revenue metrics

**Note:**
- `period` determines the date range and grouping:
  - `daily`: Last 30 days, grouped by day
  - `weekly`: Last 3 months, grouped by week
  - `monthly`: Last 12 months, grouped by month
- `average_order_value` is calculated as `total_revenue_idr / transactions`
- `trend` array contains revenue data points for the selected period

### Plan Management

#### List Plans (Admin)

**GET** `/api/v1/admin/plans`

List all subscription plans (admin view).

**Headers:**
```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**
- `active_only` (optional): `true` to show only active plans (default: `false`)

**Response:**
```json
{
  "plans": [
    {
      "id": "uuid",
      "name": "free",
      "daily_photo_limit": 10,
      "price_idr": 0,
      "price_usd": 0.0,
      "description": "Perfect for casual users",
      "features": ["10 daily analyses", "7-day history"],
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": null
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Plans retrieved successfully
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `500 Internal Server Error`: Failed to retrieve plans

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
  "daily_photo_limit": 50,
  "price_idr": 30000,
  "price_usd": 2.00,
  "description": "For small sellers",
  "features": ["50 daily analyses", "30-day history"],
  "is_active": true
}
```

**Validation:**
- `name` (required): Plan name (unique identifier)
- `daily_photo_limit` (required): Daily photo analysis limit
- `price_idr` (required): Price in Indonesian Rupiah
- `price_usd` (required): Price in US Dollars
- `description` (optional): Plan description
- `features` (optional): Array of feature strings
- `is_active` (optional): Whether plan is active (default: `true`)

**Response:**
```json
{
  "id": "uuid",
  "name": "starter",
  "daily_photo_limit": 50,
  "price_idr": 30000,
  "price_usd": 2.00,
  "description": "For small sellers",
  "features": ["50 daily analyses", "30-day history"],
  "is_active": true,
  "created_at": "2024-01-20T10:30:00Z",
  "updated_at": "2024-01-20T10:30:00Z"
}
```

**Status Codes:**
- `201 Created`: Plan created successfully
- `400 Bad Request`: Validation error or invalid request body
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `500 Internal Server Error`: Failed to create plan

**Note:** All plan creation actions are logged in audit logs.

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
  "name": "premium",              // Optional: Update plan name
  "daily_photo_limit": 1000,     // Optional: Update daily limit
  "price_idr": 70000,            // Optional: Update price in IDR
  "price_usd": 4.67,             // Optional: Update price in USD
  "description": "Updated description", // Optional: Update description
  "features": ["Updated features"],     // Optional: Update features array
  "is_active": true                      // Optional: Update active status
}
```

**Response:**
```json
{
  "id": "uuid",
  "name": "premium",
  "daily_photo_limit": 1000,
  "price_idr": 70000,
  "price_usd": 4.67,
  "description": "Updated description",
  "features": ["Updated features"],
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-20T12:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Plan updated successfully
- `400 Bad Request`: Invalid plan ID or request body
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `404 Not Found`: Plan not found
- `500 Internal Server Error`: Failed to update plan

**Note:** All plan updates are logged in audit logs with change tracking.

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

**Status Codes:**
- `200 OK`: Plan deleted successfully
- `400 Bad Request`: Invalid plan ID
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `404 Not Found`: Plan not found
- `500 Internal Server Error`: Failed to delete plan

**Note:** 
- Plan deletion is a soft delete (sets `is_active: false`)
- All deletion actions are logged in audit logs
- Existing subscriptions are not affected by plan deletion

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
[
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
    "temperature": 0.7,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": null
  },
  {
    "id": "anthropic/claude-3-5-haiku",
    "display_name": "Claude 3.5 Haiku",
    "provider": "anthropic",
    "cost_per_1k_tokens_usd": 0.00025,
    "average_latency_ms": 2200,
    "modes_supported": ["accurate", "knowledge_based"],
    "fallback_chain": [],
    "status": "active",
    "is_default": false,
    "max_tokens": 200000,
    "temperature": 0.7,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": null
  }
]
```

**Status Codes:**
- `200 OK`: Models retrieved successfully
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `500 Internal Server Error`: Failed to retrieve models

**Note:** 
- Models are ordered by `is_default DESC, display_name ASC` (default model first)
- Only active models are used for price estimation
- `status` can be: `active`, `disabled`, or `deprecated`

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
  "message": "Default model updated",
  "default_model": {
    "id": "anthropic/claude-3-5-haiku",
    "display_name": "Claude 3.5 Haiku",
    "provider": "anthropic",
    "status": "active",
    "is_default": true
  }
}
```

**Status Codes:**
- `200 OK`: Default model updated successfully
- `400 Bad Request`: Cannot set inactive model as default
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `404 Not Found`: Model not found
- `500 Internal Server Error`: Failed to update default model

**Note:** 
- Only active models can be set as default
- Setting a new default automatically unsets the previous default
- All changes are logged in audit logs

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
  "display_name": "GPT-4o Mini Updated",                    // Optional: Update display name
  "modes_supported": ["fast", "accurate", "knowledge_based"], // Optional: Update supported modes
  "fallback_chain": ["anthropic/claude-3-5-haiku"],         // Optional: Update fallback chain
  "status": "active",                                         // Optional: Update status ("active", "disabled", "deprecated")
  "max_tokens": 8000,                                         // Optional: Update max tokens
  "temperature": 0.8                                          // Optional: Update temperature
}
```

**Response:**
```json
{
  "id": "openai/gpt-4o-mini",
  "display_name": "GPT-4o Mini Updated",
  "provider": "openai",
  "cost_per_1k_tokens_usd": 0.00015,
  "average_latency_ms": 1800,
  "modes_supported": ["fast", "accurate", "knowledge_based"],
  "fallback_chain": ["anthropic/claude-3-5-haiku"],
  "status": "active",
  "is_default": true,
  "max_tokens": 8000,
  "temperature": 0.8,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-20T12:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Model updated successfully
- `400 Bad Request`: Invalid model ID or request body
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Admin access required
- `404 Not Found`: Model not found
- `500 Internal Server Error`: Failed to update model

**Note:** 
- All model updates are logged in audit logs
- Changing `status` to `disabled` or `deprecated` will prevent the model from being used
- If updating a default model to inactive status, you must set another model as default first

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

