# 🚀 Gmail SMTP Setup - Ready to Use!

## ✅ Your Gmail App Password

```
Email: rohimjoy70@gmail.com
App Password: pycljwqaflomyuun
```

## 📝 Step 1: Update Your .env.dev File

Edit `docker/.env.dev` and add/update these lines:

```bash
# SMTP Configuration - Gmail
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=rohimjoy70@gmail.com
SMTP_PASSWORD=pycljwqaflomyuun
SMTP_SENDER_NAME=Detect Price by Photo
SMTP_SENDER_EMAIL=rohimjoy70@gmail.com
SMTP_SECURE=true

# Make sure APP_BASE_URL is set
APP_BASE_URL=http://localhost:5173
```

## 🔧 Step 2: Update Docker Compose

Your `docker-compose.yml` already has the configuration! Just make sure the backend section looks like this:

```yaml
backend:
  environment:
    - SMTP_HOST=smtp.gmail.com
    - SMTP_PORT=587
    - SMTP_USERNAME=rohimjoy70@gmail.com
    - SMTP_PASSWORD=pycljwqaflomyuun
```

## 🚀 Step 3: Rebuild and Restart

```bash
# Stop backend
docker-compose -p detect-price -f docker/docker-compose.yml down backend

# Start with new config
docker-compose -p detect-price -f docker/docker-compose.yml --env-file docker/.env.dev up -d backend
```

## 🧪 Step 4: Test It!

```bash
curl -X POST http://localhost:9871/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rohimjoy70@gmail.com"
  }'
```

**Check your Gmail inbox (rohimjoy70@gmail.com)** - You should receive a password reset email! 📧

## ✅ Verify Backend Connected

```bash
docker logs detect-price-backend-1 --tail 10 | grep "Mailer"
```

Should show:
```
INF Mailer service initialized host=smtp.gmail.com port=587
```

## 🎯 What to Do Now

1. Edit `docker/.env.dev` - Add the SMTP settings above
2. Run: `docker-compose -p detect-price -f docker/docker-compose.yml --env-file docker/.env.dev up -d backend`
3. Test forgot-password API
4. Check your Gmail inbox!

---

**Need help editing .env.dev?** Open the file in your editor and paste the SMTP configuration!

