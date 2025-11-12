---
title: Midtrans Setup
weight: 1
---

# Midtrans Configuration

## Environment Variables

**⚠️ IMPORTANT: All Midtrans credentials must be set via environment variables. Never hardcode API keys in source code.**

Configure the following environment variables in `docker/.env.dev` (for local development) or your production environment:

```env
# Midtrans Payment Gateway Configuration
MIDTRANS_SERVER_KEY=your-midtrans-server-key-here
MIDTRANS_CLIENT_KEY=your-midtrans-client-key-here
MIDTRANS_ENV=sandbox                    # sandbox or production
MIDTRANS_MERCHANT_ID=your-merchant-id  # Optional, if required by your Midtrans setup
```

## Configuration Details

- **Environment:** Set `MIDTRANS_ENV` to `sandbox` for testing or `production` for live payments
- **Server Key:** Used for server-side operations (webhooks, transaction status, server-to-server API calls)
- **Client Key:** Used for client-side operations (Snap API integration in frontend)
- **Merchant ID:** Your Midtrans merchant identifier (if required)

## Getting Your Credentials

1. **Sandbox (Testing):**
   - Sign up at [Midtrans Dashboard](https://dashboard.sandbox.midtrans.com/)
   - Navigate to Settings → Access Keys
   - Copy Server Key and Client Key

2. **Production:**
   - Complete account verification
   - Get production credentials from Settings → Access Keys
   - Update `MIDTRANS_ENV=production`

## Backend Integration

The backend reads these environment variables from `docker/.env.dev` and uses them to:
- Initialize the Midtrans client
- Create payment transactions
- Handle payment webhooks
- Query transaction status

## Testing Payment Flow

1. **Create a subscription payment:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/subscriptions/subscribe \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "plan_id": "plan-uuid-here"
     }'
   ```

2. **The response will include:**
   - `snap_token` - Token for Midtrans Snap payment page
   - `redirect_url` - URL to redirect user to payment page
   - `transaction_id` - Midtrans transaction ID

3. **Webhook endpoint:**
   - `POST /api/v1/payments/webhook` - Midtrans will call this when payment status changes

## Production Setup

When moving to production:
1. Get production credentials from Midtrans dashboard
2. Update `MIDTRANS_ENV=production`
3. Update `MIDTRANS_SERVER_KEY` and `MIDTRANS_CLIENT_KEY` with production keys
4. Restart the backend service

## Security Notes

- ⚠️ Never commit `.env.dev` with production credentials to version control
- ⚠️ Server Key should only be used on the backend (never expose to frontend)
- ⚠️ Client Key can be used on the frontend for Snap API integration

