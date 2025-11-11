# Gmail SMTP Setup for Password Reset Emails

## 🎯 Quick Setup Guide

This guide will help you configure Gmail SMTP to send real password reset emails.

## 📋 Prerequisites

- Gmail account
- App Password (2FA must be enabled)

## 🔧 Step 1: Generate Gmail App Password

### 1.1 Enable 2-Factor Authentication

1. Go to: https://myaccount.google.com/security
2. Under "How you sign in to Google", enable **2-Step Verification**
3. Follow the setup process

### 1.2 Create App Password

1. Go to: https://myaccount.google.com/apppasswords
2. Select app: **Mail**
3. Select device: **Other (Custom name)**
4. Enter name: `Detect Price Backend`
5. Click **Generate**
6. **Copy the 16-character password** (format: `xxxx xxxx xxxx xxxx`)

**Important**: Save this password! You won't be able to see it again.

## 🔧 Step 2: Update Docker Environment

Add these to your `docker/.env.dev` file:

```bash
# Gmail SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-16-char-app-password
SMTP_SENDER_NAME=Detect Price by Photo
SMTP_SENDER_EMAIL=your-email@gmail.com
SMTP_SECURE=true

# Application Base URL
APP_BASE_URL=http://localhost:5173
```

**Replace**:
- `your-email@gmail.com` - Your Gmail address
- `your-16-char-app-password` - The app password from Step 1.2 (remove spaces)

## 🔧 Step 3: Update Docker Compose

Since we're using Gmail instead of MailHog, update `docker-compose.yml`:

The backend section should already override SMTP settings from environment variables, so no changes needed if properly configured.

## 🔧 Step 4: Restart Backend

```bash
# Restart backend to apply new SMTP settings
docker-compose -p detect-price -f docker/docker-compose.yml --env-file docker/.env.dev up -d backend
```

## 🧪 Step 5: Test It

### Send Password Reset Email

```bash
curl -X POST http://localhost:9871/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rohimjoy70@gmail.com"
  }'
```

**Check your email inbox!** You should receive a real password reset email.

## ✅ Verification

Check backend logs:

```bash
docker logs detect-price-backend-1 --tail 20 | grep -i "mailer"
```

Should show:
```
Mailer service initialized host=smtp.gmail.com port=587
```

## ⚠️ Troubleshooting

### Error: "Username and Password not accepted"

**Solution 1**: Make sure you're using an App Password, not your regular Gmail password

**Solution 2**: App password must be 16 characters without spaces:
```
# Wrong:
xxxx xxxx xxxx xxxx

# Correct:
xxxxxxxxxxxxxxxx
```

**Solution 3**: Check 2FA is enabled on your Google account

### Error: "Connection timeout"

**Solution**: Check if port 587 is blocked by your firewall/ISP

Alternative ports:
- Try port `465` with `SMTP_SECURE=true`
- Try port `25` (less common)

### Email Goes to Spam

This is normal for development. In production:
1. Use a domain email (not Gmail)
2. Configure SPF/DKIM/DMARC records
3. Use a dedicated email service (SendGrid, AWS SES)

## 🎯 Alternative: Mailtrap (Testing Inbox)

If you want a dedicated testing inbox:

1. Sign up: https://mailtrap.io (free)
2. Get your SMTP credentials
3. Update `.env.dev`:

```bash
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USERNAME=your-mailtrap-username
SMTP_PASSWORD=your-mailtrap-password
SMTP_SENDER_EMAIL=test@detectprice.com
```

Benefits:
- See all test emails in one inbox
- No risk of sending to real users
- Better for team testing

## 🎯 Alternative: Brevo (Real Email Delivery)

For real email delivery with better deliverability:

1. Sign up: https://www.brevo.com (free 300 emails/day)
2. Get SMTP credentials from dashboard
3. Update `.env.dev`:

```bash
SMTP_HOST=smtp-relay.brevo.com
SMTP_PORT=587
SMTP_USERNAME=your-brevo-login-email
SMTP_PASSWORD=your-brevo-smtp-key
SMTP_SENDER_EMAIL=your-verified-sender@yourdomain.com
```

## 📊 Comparison

| Service | Free Tier | Pros | Cons |
|---------|-----------|------|------|
| **Gmail** | Unlimited | Easy setup, real emails | May go to spam, rate limits |
| **Mailtrap** | 500/month | Testing inbox, no spam | Not real delivery |
| **Brevo** | 300/day | Professional, reliable | Requires signup |
| **MailHog** | Unlimited | Local, fast | Only localhost |

## 🔒 Security Notes

### Protect Your Credentials

**Never commit** SMTP passwords to git!

Add to `.gitignore`:
```
docker/.env.dev
docker/.env.prod
*.env
```

### Use Environment Variables

In production, use:
- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets
- Docker Secrets

### Rotate Passwords

Change your app password regularly:
1. Revoke old app password
2. Generate new one
3. Update `.env.dev`
4. Restart backend

## 📝 Complete .env.dev Example

```bash
# Database
DATABASE_URL=postgresql://postgres:postgres@db:5432/detect_price?sslmode=disable
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=detect_price

# Redis
REDIS_URL=redis://cache:6379

# SMTP (Gmail)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=yourname@gmail.com
SMTP_PASSWORD=your16charapppassword
SMTP_SENDER_NAME=Detect Price by Photo
SMTP_SENDER_EMAIL=yourname@gmail.com
SMTP_SECURE=true

# Application
APP_BASE_URL=http://localhost:5173
APP_MODE=development
JWT_SECRET_KEY=your-super-secret-jwt-key-change-this-in-production
JWT_ALGORITHM=HS256
```

## 🎉 You're Done!

Now when you test forgot-password, you'll receive real emails in your Gmail inbox!

Try it:
```bash
curl -X POST http://localhost:9871/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "rohimjoy70@gmail.com"}'
```

Check your Gmail! 📧

