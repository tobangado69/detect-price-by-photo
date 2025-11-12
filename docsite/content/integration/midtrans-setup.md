---
title: Midtrans Setup
weight: 1
---

# Midtrans Configuration

## Environment Variables

The following Midtrans credentials have been configured in `docker/.env.dev`:

```env
MIDTRANS_SERVER_KEY=SB-Mid-server-8u5BAzgmche547jrjzElSMHX
MIDTRANS_CLIENT_KEY=SB-Mid-client-g319kiulB1K-LLy_
MIDTRANS_ENV=sandbox
MIDTRANS_MERCHANT_ID=G369803106
```

## Configuration Details

- **Environment:** `sandbox` (for testing)
- **Merchant ID:** `G369803106`
- **Server Key:** Used for server-side operations (webhooks, transaction status)
- **Client Key:** Used for client-side operations (Snap API)

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

