# Quick Fix: Password Hash Validation Error

## TL;DR

Run this command to fix the "invalid hash format" error:

### Using Makefile (Recommended - Works on Windows/Linux/Mac)
```bash
make fix-password
```

### Alternative: Using Scripts

#### Windows (PowerShell)
```powershell
cd apps\backend
.\scripts\fix_password_hash.ps1
```

#### Linux/Mac (Bash)
```bash
cd apps/backend
chmod +x scripts/fix_password_hash.sh
./scripts/fix_password_hash.sh
```

## Or Manual Steps Using Makefile

### 1. Run Migration
```bash
make migrate
```

### 2. Re-seed Users
```bash
make seed
```

### 3. Restart Backend
```bash
make backend-up
```

### 4. Test Login
```bash
POST http://localhost:9871/api/v1/auth/signin/email
Content-Type: application/json

{
  "email": "admin@detectprice.com",
  "password": "admin123"
}
```

## What Was Fixed?

- Changed `password_hash` column from `BYTEA` to `TEXT`
- Removed PostgreSQL `convert_to()`/`convert_from()` functions
- Password hashes now stored as plain text strings (Argon2 PHC format)

## Default Test Credentials

After re-seeding:

| User | Email | Password | Role |
|------|-------|----------|------|
| admin | admin@detectprice.com | admin123 | admin |
| johndoe | johndoe@example.com | secure.password | user |

## Need More Details?

See [PASSWORD_HASH_FIX.md](./PASSWORD_HASH_FIX.md) for complete documentation.

