package models

import (
	"time"
)

type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"index"`
	Username   string    `json:"username" gorm:"index"`
	Action     string    `json:"action" gorm:"index"`
	Resource   string    `json:"resource" gorm:"index"`
	ResourceID string    `json:"resource_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	StatusCode int       `json:"status_code"`
	Error      string    `json:"error,omitempty"`
	RequestID  string    `json:"request_id" gorm:"index"`
	Metadata   string    `json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AuditLogEntry struct {
	Action     string
	Resource   string
	ResourceID string
	Metadata   map[string]interface{}
}
