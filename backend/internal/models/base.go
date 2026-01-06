/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/base.go
==============================================================================

DESCRIPTION:
    Defines the BaseModel struct that provides common fields (ID, timestamps,
    soft delete) for all database models in the IRIS Payroll System. Every
    other model embeds this base model.

USER PERSPECTIVE:
    - Every record in the system has an ID, creation date, and update date
    - Deleted records are "soft deleted" (not permanently removed from database)
    - This allows data recovery and audit trails for compliance

DEVELOPER GUIDELINES:
    ❌  DO NOT MODIFY - This is a foundational file
    ⚠️  All models MUST embed BaseModel as the first field
    📝  Example: type MyModel struct { BaseModel; OtherFields... }

SYNTAX EXPLANATION:
    - uuid.UUID: Universally Unique Identifier, 128-bit identifier
    - gorm.DeletedAt: Enables soft delete (NULL = active, timestamp = deleted)
    - `gorm:"..."`: GORM ORM tags for database column configuration
    - `json:"..."`: JSON serialization tags for API responses
    - BeforeCreate(): GORM hook called automatically before INSERT operations

DATABASE IMPACT:
    These fields are added to every table:
    - id (TEXT): Primary key, UUID format
    - created_at (DATETIME): Auto-set on INSERT
    - updated_at (DATETIME): Auto-updated on UPDATE
    - deleted_at (DATETIME): NULL for active, timestamp for soft-deleted

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models.
// All models in the system MUST embed this struct to ensure consistent
// ID generation, timestamps, and soft delete functionality.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate generates a new UUID for the ID field if it's not already set.
func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}
