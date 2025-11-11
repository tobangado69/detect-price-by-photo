# 🧪 Test Forgot Password - Quick Guide

## ✅ Services Running
- Backend: http://localhost:9871
- MailHog UI: http://localhost:8025

## 🚀 Test Now

**1. Send password reset request:**
```bash
curl -X POST http://localhost:9871/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "rohimjoy70@gmail.com"}'
```

**2. Check email in MailHog:**
- Open: http://localhost:8025
- You should see the password reset email

## 🔧 Troubleshooting

**Check backend logs:**
```bash
docker logs docker-backend-1 -f --tail 20
```

**Check services:**
```bash
cd docker && docker-compose --env-file .env.dev ps
```

## 📝 Notes
- MailHog is for **testing only** (emails don't actually send)
- For production, configure Gmail or Brevo SMTP in `docker-compose.yml`

