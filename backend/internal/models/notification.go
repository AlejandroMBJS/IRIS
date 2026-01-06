/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/notification.go
==============================================================================

DESCRIPTION:
    Defines the Notification model for company-wide alerts and updates.
    Notifications are created when significant changes occur (employee added,
    payroll processed, incidences approved, etc.) and are shown to other
    users in the same company.

USER PERSPECTIVE:
    - Users see notifications for changes they didn't make
    - Bell icon shows unread count
    - Click notification to mark as read
    - Links to relevant pages (employee detail, incidence, etc.)

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new notification types
    ⚠️  CAUTION: Notification filtering by company
    ❌  DO NOT modify: Read/unread status logic
    📝  Notifications are per-user within a company

SYNTAX EXPLANATION:
    - NotificationType: Enum for categorizing notifications
    - actor_user_id: User who caused the notification (excluded from seeing it)
    - target_user_id: Specific user to notify (nil = all company users)
    - resource_type/id: Link to related entity for navigation

NOTIFICATION TYPES:
    - employee_created: New employee added
    - employee_updated: Employee info changed
    - incidence_created: New incidence submitted
    - incidence_approved: Incidence approved
    - incidence_rejected: Incidence rejected
    - payroll_calculated: Payroll calculation completed
    - period_created: New payroll period created

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType categorizes notifications
type NotificationType string

const (
	NotificationEmployeeCreated   NotificationType = "employee_created"
	NotificationEmployeeUpdated   NotificationType = "employee_updated"
	NotificationIncidenceCreated  NotificationType = "incidence_created"
	NotificationIncidenceApproved NotificationType = "incidence_approved"
	NotificationIncidenceRejected NotificationType = "incidence_rejected"
	NotificationPayrollCalculated NotificationType = "payroll_calculated"
	NotificationPeriodCreated     NotificationType = "period_created"
	NotificationUserCreated       NotificationType = "user_created"
)

// Notification represents an alert or update for company users
type Notification struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key" json:"id"`
	CompanyID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"company_id"`
	ActorUserID  uuid.UUID        `gorm:"type:uuid;not null" json:"actor_user_id"`      // Who caused the notification
	TargetUserID *uuid.UUID       `gorm:"type:uuid" json:"target_user_id,omitempty"`    // Specific target (nil = all)
	Type         NotificationType `gorm:"type:varchar(50);not null" json:"type"`
	Title        string           `gorm:"type:varchar(100);not null" json:"title"`
	Message      string           `gorm:"type:varchar(500);not null" json:"message"`
	ResourceType string           `gorm:"type:varchar(50)" json:"resource_type,omitempty"` // e.g., "employee", "incidence"
	ResourceID   *uuid.UUID       `gorm:"type:uuid" json:"resource_id,omitempty"`          // ID of related entity
	CreatedAt    time.Time        `json:"created_at"`

	// Relationships
	Company   Company `gorm:"foreignKey:CompanyID" json:"-"`
	ActorUser User    `gorm:"foreignKey:ActorUserID" json:"actor_user,omitempty"`
}

// NotificationRead tracks which users have read which notifications
type NotificationRead struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	NotificationID uuid.UUID `gorm:"type:uuid;not null;index" json:"notification_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ReadAt         time.Time `json:"read_at"`

	// Unique constraint: each user can only mark a notification as read once
	// Created via migration index
}

// BeforeCreate generates UUID for new notifications
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// BeforeCreate generates UUID for notification read records
func (nr *NotificationRead) BeforeCreate(tx *gorm.DB) error {
	if nr.ID == uuid.Nil {
		nr.ID = uuid.New()
	}
	return nil
}

// NotificationWithReadStatus includes read status for a specific user
type NotificationWithReadStatus struct {
	Notification
	Read   bool      `json:"read"`
	ReadAt *time.Time `json:"read_at,omitempty"`
}
