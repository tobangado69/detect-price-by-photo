---
title: Project Structure
weight: 5
---

# Backend Project Structure

## Directory Layout

```
apps/backend/
├── cmd/                    # Application entry point
│   ├── commands/          # CLI commands (migrate, serve, seed)
│   └── main.go            # Main entry point
├── internal/              # Private application code
│   ├── adapter/           # External adapters (Postgres, Redis)
│   ├── admin/             # Admin module (user management, analytics)
│   ├── ai/                # AI service integration (OpenRouter)
│   ├── cache/             # Cache layer (Redis)
│   ├── config/            # Configuration management
│   ├── db/                # Database migrations and seeders
│   ├── middleware/        # HTTP middleware
│   ├── notification/      # Email notification service
│   ├── payment/           # Payment module (Midtrans)
│   ├── photo/             # Photo analysis module
│   ├── server/            # HTTP server setup
│   ├── storage/           # File storage (local/S3)
│   ├── subscription/      # Subscription management
│   ├── user/              # User module
│   │   ├── auth/          # Authentication (JWT, password reset)
│   │   └── user/          # User management
│   └── utils/             # Shared utilities
├── templates/             # Email templates
├── scripts/               # Utility scripts
└── test/                  # Integration tests
```

## Key Modules

### Authentication (`internal/user/auth/`)
- JWT token generation and validation
- Password hashing (Argon2id)
- Email verification
- Password reset flow
- Session management

### User Management (`internal/user/user/`)
- User CRUD operations
- Role-based access control (RBAC)
- User suspension/banning

### Admin (`internal/admin/`)
- User administration
- Analytics dashboard
- Plan management
- AI model configuration

### Payment (`internal/payment/`)
- Midtrans integration
- Invoice management
- Webhook handling

### Photo Analysis (`internal/photo/`)
- Image upload and storage
- AI-powered price estimation
- RAG (Retrieval-Augmented Generation)

## Database Schema

See [Database Setup](/backend/database-setup/) for migration details.

## API Structure

All API endpoints follow RESTful conventions:
- `/api/v1/auth/*` - Authentication endpoints
- `/api/v1/users/*` - User management (admin only)
- `/api/v1/admin/*` - Admin operations (admin only)
- `/api/v1/subscriptions/*` - Subscription management
- `/api/v1/photos/*` - Photo analysis
- `/api/v1/payments/*` - Payment processing

