---
title: Forgot Password Guide
weight: 2
---

# Forgot Password Workflow Guide

## Overview

The forgot-password feature allows users to reset their password via email when they can't remember it.

## 🔄 How It Works

### Flow Diagram
```
User → Forgot Password API → Generate Token → Send Email → MailHog
                                    ↓
                            Store Token Hash in DB
                                    ↓
User Receives Email → Clicks Link → Reset Password API → Update Password
```

### Step-by-Step Process

1. **User Requests Password Reset**
   - POST `/api/v1/auth/forgot-password` with email
   - System generates a secure, cryptographically random token (48 bytes, URL-safe)
   - Token is hashed (SHA-256) and stored in database
   - Raw token is sent to user's email
   - Token expires in **1 hour**

2. **User Receives Email**
   - Email contains a reset link: `{BASE_URL}/reset-password?token={TOKEN}`
   - User clicks the link

3. **User Resets Password**
   - POST `/api/v1/auth/reset-password` with token and new password
   - System validates token (checks hash, expiry)
   - Password is updated with Argon2id hash
   - Token is deleted (one-time use)

## 🧪 Testing Locally

### Prerequisites

1. **MailHog must be running** (SMTP test server)
2. **Backend must be connected to MailHog**

### Start MailHog

```bash
make mailhog-up
```

This starts:
- **SMTP Server**: `localhost:1025` (receives emails)
- **Web UI**: `http://localhost:8025` (view emails)

### Test Workflow

#### Step 1: Request Password Reset

```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rohimjoy70@gmail.com"
  }'
```

**Response:**
```json
{
  "message": "If an account with that email exists, a password reset link has been sent."
}
```

> **Note**: The response is always the same (security best practice) to prevent email enumeration attacks.

#### Step 2: Check MailHog for Email

1. Open MailHog UI: `http://localhost:8025`
2. You should see an email from "Detect Price by Photo"
3. Click the email to view it
4. Copy the reset token from the URL (e.g., `http://yourapp.com/reset-password?token=ABC123...`)

#### Step 3: Reset Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "PASTE_TOKEN_FROM_EMAIL_HERE",
    "new_password": "MyNewSecurePassword123!"
  }'
```

**Success Response:**
```json
{
  "message": "Password has been reset successfully. You can now log in with your new password."
}
```

**Error Response (invalid/expired token):**
```json
{
  "error": "Invalid or expired token"
}
```

#### Step 4: Login with New Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/signin/email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rohimjoy70@gmail.com",
    "password": "MyNewSecurePassword123!"
  }'
```

## 📋 API Reference

### Forgot Password

**POST** `/api/v1/auth/forgot-password`

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Validation:**
- `email` (required, valid email format)

**Response:** Always `200 OK`
```json
{
  "message": "If an account with that email exists, a password reset link has been sent."
}
```

### Reset Password

**POST** `/api/v1/auth/reset-password`

**Request:**
```json
{
  "token": "reset_token_from_email",
  "new_password": "NewSecurePassword123!"
}
```

**Validation:**
- `token` (required)
- `new_password` (required, minimum 12 characters)

**Success Response:** `200 OK`
```json
{
  "message": "Password has been reset successfully. You can now log in with your new password."
}
```

**Error Response:** `400 Bad Request`
```json
{
  "error": "Invalid or expired token"
}
```

## 🔒 Security Features

### 1. Token Security
- **Cryptographically Secure**: Uses `crypto/rand` for token generation
- **URL-Safe**: Base64 URL encoding (no special characters)
- **Hashed Storage**: Only SHA-256 hash stored in database
- **One-Time Use**: Token deleted after successful password reset
- **Short Expiry**: 1 hour validity

### 2. Email Enumeration Prevention
- Always returns success response
- Doesn't reveal if email exists or not
- Prevents attackers from discovering user accounts

### 3. Rate Limiting
- Existing valid tokens are reused (updates `last_sent_at`)
- Can't spam token requests for same email
- Old tokens are cleaned up when new ones are created

### 4. Password Requirements
- Minimum 12 characters
- Stored as Argon2id hash (not plain text)

## 🛠️ Configuration

### Environment Variables

```bash
# SMTP Configuration (for sending emails)
SMTP_HOST=mailhog          # Docker: mailhog, Local: localhost
SMTP_PORT=1025             # MailHog SMTP port
SMTP_USERNAME=             # Optional (not needed for MailHog)
SMTP_PASSWORD=             # Optional (not needed for MailHog)
SMTP_SENDER_NAME="Detect Price by Photo"
SMTP_SENDER_EMAIL="noreply@detectprice.com"

# Application
APP_BASE_URL=http://localhost:5173  # Frontend URL for reset links
```

### Docker Compose

MailHog is configured in `compose.yaml` (via `docker/_stacks_/mailpit.yaml`):

```yaml
mailhog:
  image: mailhog/mailhog:latest
  ports:
    - '1025:1025'  # SMTP server
    - '8025:8025'  # Web UI
```

Backend automatically connects to MailHog via environment variables:

```yaml
backend:
  environment:
    - SMTP_HOST=mailhog
    - SMTP_PORT=1025
```

## 🐛 Troubleshooting

### Problem: Email Not Received

**Check 1: Is MailHog Running?**
```bash
docker ps | grep mailhog
```

**Solution:**
```bash
make mailhog-up
```

**Check 2: Is Backend Connected to MailHog?**
```bash
docker compose -f compose.yaml logs backend | grep "Mailer service initialized"
```

Should show:
```
Mailer service initialized host=mailhog port=1025
```

If it shows `host=localhost`, restart backend:
```bash
docker compose -f compose.yaml up -d
```

**Check 3: Check MailHog UI**

Open `http://localhost:8025` and see if emails are appearing.

### Problem: Token Expired

Tokens expire after **1 hour**. Request a new reset link:
```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "your@email.com"}'
```

### Problem: Invalid Token Error

Possible causes:
1. Token already used (one-time use only)
2. Token expired (>1 hour old)
3. Token copied incorrectly (check for extra spaces)

**Solution:** Request a new reset link

### Problem: Password Validation Failed

Password must be **at least 12 characters**:
```json
{
  "error": "Validation failed",
  "details": "new_password: must be at least 12 characters"
}
```

## 📧 Email Template

The password reset email includes:
- User's display name (if available)
- Reset link with token
- Expiry time (1 hour)
- Security notice

Example email body:
```
Hello Rohim Joy,

You requested to reset your password for Detect Price by Photo.

Click the link below to reset your password:
http://localhost:5173/reset-password?token=ABC123...

This link will expire in 1 hour.

If you didn't request this password reset, please ignore this email.

Best regards,
Detect Price by Photo Team
```

## 🔧 Development Tips

### View All Emails in MailHog

Open: `http://localhost:8025`

### Clear All Emails
Click "Clear" button in MailHog UI

### Check Backend Logs
```bash
docker compose -f compose.yaml logs backend --tail 50 | grep -i "password\|email\|reset"
```

### Query Database for Tokens
```bash
make db-shell
```

Then in psql:
```sql
SELECT 
    id,
    user_id,
    subject,
    relates_to,
    created_at,
    expires_at,
    last_sent_at
FROM public.one_time_tokens
WHERE subject = 'password_reset'
ORDER BY created_at DESC;
```

## 🎯 Production Considerations

### 1. Use Real SMTP Service
Replace MailHog with a real email service:
- SendGrid
- AWS SES
- Mailgun
- Postmark

### 2. Update Environment Variables
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=YOUR_SENDGRID_API_KEY
SMTP_SENDER_EMAIL=noreply@your domain.com
```

### 3. Configure SPF/DKIM/DMARC
Ensure your domain has proper email authentication records

### 4. Use HTTPS
```bash
APP_BASE_URL=https://yourdomain.com
```

### 5. Monitor Email Delivery
- Track bounce rates
- Monitor delivery rates
- Set up alerts for failures

## 📚 Related Documentation

- [API Documentation](./API_DOCUMENTATION.md)
- [Password Hash Fix Guide](./PASSWORD_HASH_FIX.md)
- [Makefile Guide](./MAKEFILE_FIX_GUIDE.md)
- [Admin Seed Guide](./ADMIN_SEED.md)

