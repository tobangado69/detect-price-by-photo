package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

const AuditLogTable = "public.audit_logs"

// AuditLog represents an audit log entry in the database.
type AuditLog struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	AdminID    uuid.UUID              `json:"admin_id" db:"admin_id"`
	Action     string                 `json:"action" db:"action"`
	TargetType string                 `json:"target_type" db:"target_type"`
	TargetID   *uuid.UUID             `json:"target_id" db:"target_id"`
	Changes    map[string]interface{} `json:"changes" db:"changes"`
	IPAddress  *string                `json:"ip_address" db:"ip_address"`
	UserAgent  *string                `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

