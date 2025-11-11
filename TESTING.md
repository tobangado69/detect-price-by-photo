# 🧪 Testing Guide

## 🚀 Quick Start

**1. Start services:**
```bash
make migrate  # Run migrations (first time only)
make seed     # Seed database (first time only)
make mailhog-up
make backend-up
```

**2. Sign up a new user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Test User",
    "email": "test.user@example.com",
    "password": "StrongPassword123",
    "password_confirmation": "StrongPassword123"
  }'
```

**3. Verify signup email (optional but recommended):**
- Open: http://localhost:8025
- Find the verification email and copy the token
- Confirm via:
```bash
curl "http://localhost:8080/api/v1/auth/verify-email?token=PASTE_TOKEN_HERE"
```

**4. Test forgot password:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@detectprice.com"}'
```

**5. Check email:**
- Open: http://localhost:8025 (MailHog UI)
- You should see the password reset email

**6. Reset password with token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "YOUR_TOKEN_FROM_EMAIL",
    "new_password": "YourNewPassword123"
  }'
```

**Note:** Password must be at least **12 characters** long.

---

## 🔒 Suspend and Reinstate Users (Admin)

1. **Suspend a user:**
   ```bash
   curl -X PUT http://localhost:8080/api/v1/admin/users/USER_ID \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <admin_access_token>" \
     -d '{
       "status": "suspended",
       "ban_reason": "Fraudulent activity detected",
       "ban_expires": "2025-12-31T23:59:59Z"
     }'
   ```
2. **Attempt to sign in** with the suspended account → expect `403 Account is banned`.
3. **Reinstate user:**
   ```bash
   curl -X PUT http://localhost:8080/api/v1/admin/users/USER_ID \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <admin_access_token>" \
     -d '{"status": "active"}'
   ```

Expired bans are automatically cleared on the next successful login.

---

## 📋 Available Test Users

| Email | Password | Role |
|-------|----------|------|
| admin@detectprice.com | admin123 | admin |
| johndoe@example.com | secure.password | user |

---

## 🔧 Useful Commands

**View backend logs:**
```bash
docker logs docker-backend-1 -f
```

**Restart backend:**
```bash
cd docker && docker-compose --env-file .env.dev restart backend
```

**Stop all services:**
```bash
cd docker && docker-compose --env-file .env.dev down
```

---

## 🌐 Service URLs

- **API**: http://localhost:8080
- **MailHog UI**: http://localhost:8025
- **PGWeb (DB UI)**: http://localhost:8081
- **Database**: localhost:5432
- **Redis**: localhost:6379

## ✅ Test Result
```
✅ password reset email sent successfully email=admin@detectprice.com
```

**Check MailHog**: http://localhost:8025 to see the email!
