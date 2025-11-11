#!/bin/bash

# Script to fix password hash validation issue
# This script runs the migration and re-seeds users

set -e  # Exit on error

echo "======================================"
echo "Password Hash Fix Script"
echo "======================================"
echo ""

# Check if we're in the backend directory
if [ ! -f "go.mod" ]; then
    echo "Error: Please run this script from apps/backend directory"
    exit 1
fi

# Check if backend is running
if [ -f "../../docker/docker-compose.yml" ]; then
    echo "Step 1: Stopping backend service..."
    cd ../../docker
    docker-compose stop backend
    cd ../apps/backend
    echo "✓ Backend stopped"
    echo ""
fi

# Run migration
echo "Step 2: Running migration to change password_hash to TEXT..."
if command -v go &> /dev/null; then
    go run cmd/server/main.go migrate up
    echo "✓ Migration completed"
else
    echo "⚠ Go not found. Please run migration manually:"
    echo "  go run cmd/server/main.go migrate up"
fi
echo ""

# Re-seed users
echo "Step 3: Re-seeding users with correct password hashes..."
echo "⚠ This will reset passwords for admin and johndoe users"
read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if command -v go &> /dev/null; then
        go run cmd/server/main.go seed
        echo "✓ Users re-seeded"
    else
        echo "⚠ Go not found. Please run seed manually:"
        echo "  go run cmd/server/main.go seed"
    fi
else
    echo "⚠ Skipping user re-seed"
    echo "  You'll need to manually create users or update existing password hashes"
fi
echo ""

# Restart backend
if [ -f "../../docker/docker-compose.yml" ]; then
    echo "Step 4: Starting backend service..."
    cd ../../docker
    docker-compose up -d backend
    cd ../apps/backend
    echo "✓ Backend started"
    echo ""
fi

echo "======================================"
echo "Fix Applied Successfully!"
echo "======================================"
echo ""
echo "Test with these credentials:"
echo "  Admin User:"
echo "    Username: admin"
echo "    Email: admin@detectprice.com"
echo "    Password: admin123"
echo ""
echo "  Regular User:"
echo "    Username: johndoe"
echo "    Email: johndoe@example.com"
echo "    Password: secure.password"
echo ""
echo "API Endpoint: POST http://localhost:9871/api/v1/auth/signin/email"
echo ""
echo "See docs/PASSWORD_HASH_FIX.md for more details"

