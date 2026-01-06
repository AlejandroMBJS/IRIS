package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EscalationLog tracks the audit trail for auto-escalations in the approval workflow
// When a request sits pending for 24+ hours, it auto-escalates to the next level
type EscalationLog struct {
	BaseModel
	AbsenceRequestID  uuid.UUID      `gorm:"type:text;not null;index" json:"absence_request_id"`
	FromStage         string         `gorm:"type:varchar(50);not null" json:"from_stage"`
	ToStage           string         `gorm:"type:varchar(50);not null" json:"to_stage"`
	EscalatedAt       time.Time      `gorm:"not null" json:"escalated_at"`
	EscalationReason  string         `gorm:"type:varchar(255);default:'24 hours without approval'" json:"escalation_reason"`
	NotifiedUserIDs   datatypes.JSON `gorm:"type:jsonb" json:"notified_user_ids"` // JSON array of user IDs (SQLite and PostgreSQL compatible)

	// Relationships
	AbsenceRequest AbsenceRequest `gorm:"foreignKey:AbsenceRequestID;constraint:OnDelete:CASCADE" json:"absence_request,omitempty"`
}

// BeforeCreate hook to set UUID and default escalation time
func (e *EscalationLog) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.EscalatedAt.IsZero() {
		e.EscalatedAt = time.Now()
	}
	if e.EscalationReason == "" {
		e.EscalationReason = "24 hours without approval"
	}
	return nil
}

// TableName specifies the table name for GORM
func (EscalationLog) TableName() string {
	return "escalation_logs"
}

// GetNotifiedUserIDsSlice returns the notified user IDs as a Go slice
func (e *EscalationLog) GetNotifiedUserIDsSlice() []string {
	if e.NotifiedUserIDs == nil || len(e.NotifiedUserIDs) == 0 {
		return []string{}
	}

	var ids []string
	if err := json.Unmarshal(e.NotifiedUserIDs, &ids); err != nil {
		return []string{}
	}
	return ids
}

// AddNotifiedUser adds a user ID to the notified users list
func (e *EscalationLog) AddNotifiedUser(userID string) {
	ids := e.GetNotifiedUserIDsSlice()
	ids = append(ids, userID)

	data, err := json.Marshal(ids)
	if err != nil {
		return
	}
	e.NotifiedUserIDs = data
}

// SetNotifiedUserIDs sets the notified user IDs from a slice
func (e *EscalationLog) SetNotifiedUserIDs(userIDs []string) error {
	data, err := json.Marshal(userIDs)
	if err != nil {
		return err
	}
	e.NotifiedUserIDs = data
	return nil
}
