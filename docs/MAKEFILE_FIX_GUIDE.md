# Makefile Guide: Fix Password Hash Validation Error

## 🚀 Quick Fix (One Command)

Run this single command to fix everything:

```bash
make fix-password
```

This will automatically:
1. ✅ Start the database if not running
2. ✅ Run the migration to change `password_hash` from BYTEA to TEXT
3. ✅ Re-seed users with correct password hashes
4. ✅ Restart the backend service

## 📋 Step-by-Step (Manual Control)

If you prefer to run each step manually:

### Step 1: Ensure Database is Running
```bash
make db-up
```

### Step 2: Run Migration
```bash
make migrate
```

### Step 3: Re-seed Users
```bash
make seed
```

### Step 4: Restart Backend
```bash
make backend-up
```

## 🧪 Testing

After applying the fix, test with Postman or curl:

### Using curl
```bash
curl -X POST http://localhost:9871/api/v1/auth/signin/email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@detectprice.com",
    "password": "admin123"
  }'
```

### Expected Response
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "user": {
      "id": "...",
      "email": "admin@detectprice.com",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

## 🔑 Default Test Credentials

After seeding, use these credentials:

| User Type | Email | Password | Username |
|-----------|-------|----------|----------|
| Admin | admin@detectprice.com | admin123 | admin |
| Regular | johndoe@example.com | secure.password | johndoe |

## 📝 Available Makefile Commands

### Database Commands
- `make db-up` - Start PostgreSQL + Redis
- `make db-down` - Stop PostgreSQL + Redis
- `make db-shell` - Open psql shell
- `make db-logs` - View PostgreSQL logs

### Migration Commands
- `make migrate` - Run all pending migrations
- `make migrate-down` - Roll back last migration
- `make seed` - Seed database with initial data

### Fix Commands
- `make fix-password` - **Fix password hash error (all-in-one)**

### Service Commands
- `make backend-build` - Build backend Docker image
- `make backend-up` - Start backend service
- `make frontend-build` - Build frontend Docker image
- `make frontend-up` - Start frontend service
- `make full-build` - Build both backend and frontend
- `make docker-up` - Start all services
- `make docker-down` - Stop all services

### Utility Commands
- `make pgweb-up` - Start pgweb UI for database management
- `make clean` - Remove all containers and volumes
- `make help` - Show available commands

## 🔍 Troubleshooting

### Issue: "Network not found"
**Solution:** Start the database first
```bash
make db-up
make fix-password
```

### Issue: "Migration already applied"
This is normal if you've already run the migration. Just run the seed:
```bash
make seed
```

### Issue: "Port already in use"
**Solution:** Check if services are already running
```bash
docker ps
make docker-down
make docker-up
```

### Issue: Backend won't start
**Solution:** Check logs and rebuild
```bash
docker logs detect-price-backend-1
make backend-build
make backend-up
```

## 🔧 Advanced Usage

### Run Migration Only (Without Seed)
```bash
make db-up
make migrate
```

### Re-seed Without Migration
```bash
make seed
```

### View What Changed in Database
```bash
make db-shell
```
Then in psql:
```sql
-- Check password hash format
SELECT 
    user_id,
    LEFT(password_hash, 50) AS hash_preview,
    LENGTH(password_hash) AS hash_length
FROM public.user_passwords;

-- Exit psql
\q
```

### Check Migration Status
After running `make migrate`, the migration `00016_alter_password_hash_to_text.sql` should be applied.

## 📊 What the Fix Does

### Before (Broken)
- Column type: `BYTEA`
- Storage: Using `convert_to()` function
- Retrieval: Using `convert_from()` function
- Result: ❌ Corrupted hash format

### After (Fixed)
- Column type: `TEXT`
- Storage: Direct string storage
- Retrieval: Direct string retrieval
- Result: ✅ Valid Argon2 PHC format

## 🎯 Complete Workflow Example

Here's a complete example workflow:

```bash
# 1. Stop everything
make docker-down

# 2. Start fresh
make db-up

# 3. Apply the fix
make fix-password

# 4. Start all services
make docker-up

# 5. Test the login
curl -X POST http://localhost:9871/api/v1/auth/signin/email \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@detectprice.com", "password": "admin123"}'
```

## 📖 Related Documentation

- [PASSWORD_HASH_FIX.md](./PASSWORD_HASH_FIX.md) - Detailed technical explanation
- [QUICK_FIX_PASSWORD.md](./QUICK_FIX_PASSWORD.md) - Quick reference
- [ADMIN_SEED.md](./ADMIN_SEED.md) - Admin user creation guide
- [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) - API reference

## ✅ Verification Checklist

After running `make fix-password`:

- [ ] Migration applied successfully
- [ ] Users re-seeded
- [ ] Backend restarted
- [ ] Login with admin@detectprice.com works
- [ ] Login with johndoe@example.com works
- [ ] No "invalid hash format" errors in logs

## 🆘 Need Help?

If you encounter issues:

1. Check Docker logs: `docker logs detect-price-backend-1`
2. Check database logs: `make db-logs`
3. Verify database connection: `make db-shell`
4. Review the full documentation: `docs/PASSWORD_HASH_FIX.md`

## 🔄 Rollback (If Needed)

If you need to rollback:

```bash
make migrate-down
```

**Warning:** This will revert to BYTEA storage. You'll need to re-hash all passwords.

