/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/message.go
==============================================================================

DESCRIPTION:
    Defines the Message model for the internal messaging/inbox system.
    Supports direct messages between users, questions about announcements,
    and threaded conversations.

USER PERSPECTIVE:
    - Employees can send messages to HR, supervisors, or managers
    - Questions about announcements are linked to the original announcement
    - Messages can be replied to, creating a conversation thread
    - Unread messages are highlighted in the inbox

DEVELOPER GUIDELINES:
    OK to modify: Add new message types, enhance thread functionality
    CAUTION: Message visibility rules (sender, recipient, linked announcement)
    DO NOT modify: Read/unread status logic, soft delete behavior

MESSAGE TYPES:
    - direct: Direct message between two users
    - announcement_question: Question about an announcement
    - system: System-generated messages

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageType categorizes messages
type MessageType string

const (
	MessageTypeDirect              MessageType = "direct"
	MessageTypeAnnouncementQuestion MessageType = "announcement_question"
	MessageTypeSystem              MessageType = "system"
)

// MessageStatus tracks the status of a message
type MessageStatus string

const (
	MessageStatusUnread   MessageStatus = "unread"
	MessageStatusRead     MessageStatus = "read"
	MessageStatusArchived MessageStatus = "archived"
)

// Message represents a message in the internal messaging system
type Message struct {
	ID             uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
	CompanyID      uuid.UUID      `gorm:"type:text;not null;index" json:"company_id"`
	SenderID       uuid.UUID      `gorm:"type:text;not null;index" json:"sender_id"`
	RecipientID    uuid.UUID      `gorm:"type:text;not null;index" json:"recipient_id"`
	Subject        string         `gorm:"type:varchar(255);not null" json:"subject"`
	Body           string         `gorm:"type:text;not null" json:"body"`
	Type           MessageType    `gorm:"type:varchar(50);not null;default:'direct'" json:"type"`
	Status         MessageStatus  `gorm:"type:varchar(20);not null;default:'unread'" json:"status"`

	// Optional: Link to announcement for announcement_question type
	AnnouncementID *uuid.UUID     `gorm:"type:text;index" json:"announcement_id,omitempty"`

	// Optional: Link to parent message for threaded conversations
	ParentID       *uuid.UUID     `gorm:"type:text;index" json:"parent_id,omitempty"`

	// Timestamps
	ReadAt         *time.Time     `json:"read_at,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships (populated on query)
	Sender       *User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Recipient    *User         `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Announcement *Announcement `gorm:"foreignKey:AnnouncementID" json:"announcement,omitempty"`
	Parent       *Message      `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies      []Message     `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// BeforeCreate generates UUID for new messages
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for GORM
func (Message) TableName() string {
	return "messages"
}

// MessageResponse is the DTO for API responses
type MessageResponse struct {
	ID             uuid.UUID     `json:"id"`
	Subject        string        `json:"subject"`
	Body           string        `json:"body"`
	Type           MessageType   `json:"type"`
	Status         MessageStatus `json:"status"`
	AnnouncementID *uuid.UUID    `json:"announcement_id,omitempty"`
	ParentID       *uuid.UUID    `json:"parent_id,omitempty"`
	ReadAt         *time.Time    `json:"read_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`

	// Populated relationships
	Sender *struct {
		ID       uuid.UUID `json:"id"`
		FullName string    `json:"full_name"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
	} `json:"sender,omitempty"`

	Recipient *struct {
		ID       uuid.UUID `json:"id"`
		FullName string    `json:"full_name"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
	} `json:"recipient,omitempty"`

	ReplyCount int `json:"reply_count,omitempty"`
}
