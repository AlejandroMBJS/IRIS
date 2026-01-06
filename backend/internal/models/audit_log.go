/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/audit_log.go
==============================================================================

DESCRIPTION:
    Defines the AuditLog model for tracking user authentication and activity.
    Logs all login attempts, successful logins, token refreshes, and page visits.

USER PERSPECTIVE:
    - Admins can view detailed logs of all user activity
    - Tracks successful and failed login attempts
    - Records time spent on pages
    - Helps identify security issues and user behavior

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new event types, add new fields
    ⚠️  CAUTION: Changing enum values (ensure DB migration)
    ❌  DO NOT modify: Core structure without considering existing logs
    📝  Security-sensitive file - logs must be tamper-resistant

SYNTAX EXPLANATION:
    - `gorm:"type:text;not null"`: Database column definition
    - `json:"event_type"`: JSON field name in API responses
    - *uuid.UUID: Nullable foreign key (nil for failed logins)
    - *string: Nullable string field

EVENT TYPES:
    - login_attempt: Any login attempt (success or failure)
    - login_success: Successful login
    - login_failure: Failed login attempt
    - token_refresh: Access token refreshed
    - logout: User logged out
    - page_visit: User visited a page
    - password_change: User changed password

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogEventType represents the type of audit event
type AuditLogEventType string

const (
	EventLoginAttempt    AuditLogEventType = "login_attempt"
	EventLoginSuccess    AuditLogEventType = "login_success"
	EventLoginFailure    AuditLogEventType = "login_failure"
	EventTokenRefresh    AuditLogEventType = "token_refresh"
	EventLogout          AuditLogEventType = "logout"
	EventPageVisit       AuditLogEventType = "page_visit"
	EventPasswordChange  AuditLogEventType = "password_change"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	BaseModel
	EventType   AuditLogEventType `gorm:"type:text;not null;index" json:"event_type"`
	UserID      *uuid.UUID         `gorm:"type:text;index" json:"user_id,omitempty"`
	Email       string             `gorm:"type:varchar(255);not null;index" json:"email"`
	IPAddress   string             `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent   string             `gorm:"type:text" json:"user_agent,omitempty"`
	Success     bool               `gorm:"default:false" json:"success"`
	FailureReason *string          `gorm:"type:text" json:"failure_reason,omitempty"`
	PageURL     *string            `gorm:"type:text" json:"page_url,omitempty"`
	SessionID   *string            `gorm:"type:varchar(255)" json:"session_id,omitempty"`
	Duration    *int               `gorm:"type:integer" json:"duration,omitempty"` // Duration in seconds for page visits
	Metadata    *string            `gorm:"type:text" json:"metadata,omitempty"`    // JSON metadata for additional info

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (AuditLog) TableName() string {
	return "audit_logs"
}

// LoginSession represents active login sessions with timing information
type LoginSession struct {
	BaseModel
	UserID       uuid.UUID  `gorm:"type:text;not null;index" json:"user_id"`
	Email        string     `gorm:"type:varchar(255);not null" json:"email"`
	LoginAt      time.Time  `gorm:"not null" json:"login_at"`
	LogoutAt     *time.Time `json:"logout_at,omitempty"`
	LastActivity time.Time  `gorm:"not null" json:"last_activity"`
	IPAddress    string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent    string     `gorm:"type:text" json:"user_agent,omitempty"`
	SessionID    string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"session_id"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (LoginSession) TableName() string {
	return "login_sessions"
}

// GetDuration returns the session duration in seconds
func (ls *LoginSession) GetDuration() int64 {
	if ls.LogoutAt != nil {
		return int64(ls.LogoutAt.Sub(ls.LoginAt).Seconds())
	}
	return int64(ls.LastActivity.Sub(ls.LoginAt).Seconds())
}

// PageVisit represents page visit tracking for detailed analytics
type PageVisit struct {
	BaseModel
	UserID      uuid.UUID  `gorm:"type:text;not null;index" json:"user_id"`
	SessionID   string     `gorm:"type:varchar(255);not null;index" json:"session_id"`
	PageURL     string     `gorm:"type:text;not null" json:"page_url"`
	PageTitle   *string    `gorm:"type:varchar(255)" json:"page_title,omitempty"`
	EnteredAt   time.Time  `gorm:"not null" json:"entered_at"`
	ExitedAt    *time.Time `json:"exited_at,omitempty"`
	Duration    *int       `gorm:"type:integer" json:"duration,omitempty"` // Duration in seconds

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (PageVisit) TableName() string {
	return "page_visits"
}
