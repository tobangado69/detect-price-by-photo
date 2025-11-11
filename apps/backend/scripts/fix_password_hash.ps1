# PowerShell Script to fix password hash validation issue
# This script runs the migration and re-seeds users

$ErrorActionPreference = "Stop"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Password Hash Fix Script" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the backend directory
if (-not (Test-Path "go.mod")) {
    Write-Host "Error: Please run this script from apps/backend directory" -ForegroundColor Red
    exit 1
}

# Check if backend is running
if (Test-Path "../../docker/docker-compose.yml") {
    Write-Host "Step 1: Stopping backend service..." -ForegroundColor Yellow
    Push-Location ../../docker
    docker-compose stop backend
    Pop-Location
    Write-Host "✓ Backend stopped" -ForegroundColor Green
    Write-Host ""
}

# Run migration
Write-Host "Step 2: Running migration to change password_hash to TEXT..." -ForegroundColor Yellow
try {
    go run cmd/server/main.go migrate up
    Write-Host "✓ Migration completed" -ForegroundColor Green
} catch {
    Write-Host "⚠ Migration failed. Please run manually:" -ForegroundColor Yellow
    Write-Host "  go run cmd/server/main.go migrate up" -ForegroundColor Yellow
}
Write-Host ""

# Re-seed users
Write-Host "Step 3: Re-seeding users with correct password hashes..." -ForegroundColor Yellow
Write-Host "⚠ This will reset passwords for admin and johndoe users" -ForegroundColor Yellow
$response = Read-Host "Continue? (y/n)"
Write-Host ""

if ($response -eq "y" -or $response -eq "Y") {
    try {
        go run cmd/server/main.go seed
        Write-Host "✓ Users re-seeded" -ForegroundColor Green
    } catch {
        Write-Host "⚠ Seed failed. Please run manually:" -ForegroundColor Yellow
        Write-Host "  go run cmd/server/main.go seed" -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠ Skipping user re-seed" -ForegroundColor Yellow
    Write-Host "  You'll need to manually create users or update existing password hashes" -ForegroundColor Yellow
}
Write-Host ""

# Restart backend
if (Test-Path "../../docker/docker-compose.yml") {
    Write-Host "Step 4: Starting backend service..." -ForegroundColor Yellow
    Push-Location ../../docker
    docker-compose up -d backend
    Pop-Location
    Write-Host "✓ Backend started" -ForegroundColor Green
    Write-Host ""
}

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Fix Applied Successfully!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test with these credentials:" -ForegroundColor White
Write-Host "  Admin User:" -ForegroundColor Cyan
Write-Host "    Username: admin"
Write-Host "    Email: admin@detectprice.com"
Write-Host "    Password: admin123"
Write-Host ""
Write-Host "  Regular User:" -ForegroundColor Cyan
Write-Host "    Username: johndoe"
Write-Host "    Email: johndoe@example.com"
Write-Host "    Password: secure.password"
Write-Host ""
Write-Host "API Endpoint: POST http://localhost:9871/api/v1/auth/signin/email" -ForegroundColor Yellow
Write-Host ""
Write-Host "See docs/PASSWORD_HASH_FIX.md for more details" -ForegroundColor Gray

