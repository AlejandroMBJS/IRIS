/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/company.go
==============================================================================

DESCRIPTION:
    Defines the Company model. Each company is an employer entity that has
    employees and users. The system supports multi-tenancy through company
    isolation - users can only see data from their own company.

USER PERSPECTIVE:
    - Company is created during initial registration
    - All employees and users belong to exactly one company
    - Company RFC (tax ID) is used for official CFDI documents
    - Company can be deactivated to disable access without deleting data

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields (fiscal data, settings, etc.)
    ⚠️  CAUTION: Activation/deactivation logic
    ❌  DO NOT modify: RFC uniqueness constraint, multi-tenancy logic
    📝  Always filter queries by CompanyID for data isolation

SYNTAX EXPLANATION:
    - ActivatedAt/DeactivatedAt: Track company status changes
    - BeforeCreate/BeforeUpdate: GORM hooks for automatic timestamp management
    - foreignKey:CompanyID: All users/employees reference this company

MULTI-TENANCY:
    - Each company's data is isolated
    - Users can only access their company's employees, payroll, etc.
    - CompanyID is set from JWT token, not from user input
    - DO NOT trust CompanyID from request body in handlers

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Company represents a company/employer in the system.
// All employees and users belong to a company, enabling multi-tenancy.
type Company struct {
	BaseModel
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	RFC          string         `gorm:"type:varchar(13);uniqueIndex;not null" json:"rfc"`
	Address      string         `gorm:"type:varchar(255)" json:"address,omitempty"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Email        string         `gorm:"type:varchar(255)" json:"email,omitempty"`
	Website      string         `gorm:"type:varchar(255)" json:"website,omitempty"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Users        []User         `gorm:"foreignKey:CompanyID" json:"users,omitempty"`
	Employees    []Employee     `gorm:"foreignKey:CompanyID" json:"employees,omitempty"`
	CreatedBy    *uuid.UUID     `gorm:"type:text" json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID     `gorm:"type:text" json:"updated_by,omitempty"`
	CreatedByUser *User         `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
	UpdatedByUser *User         `gorm:"foreignKey:UpdatedBy" json:"updated_by_user,omitempty"`
	ActivatedAt  *time.Time     `json:"activated_at,omitempty"`
	DeactivatedAt *time.Time    `json:"deactivated_at,omitempty"`
}

// TableName specifies the table name
func (Company) TableName() string {
	return "companies"
}

// BeforeCreate hook to generate UUID and set ActivatedAt for new active companies.
func (c *Company) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate UUID if not set (important since BaseModel's hook is overridden)
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.IsActive && c.ActivatedAt == nil {
		now := time.Now()
		c.ActivatedAt = &now
	}
	return
}

// BeforeUpdate hook to manage ActivatedAt/DeactivatedAt.
func (c *Company) BeforeUpdate(tx *gorm.DB) (err error) {
	if c.IsActive && c.ActivatedAt == nil {
		now := time.Now()
		c.ActivatedAt = &now
		c.DeactivatedAt = nil
	} else if !c.IsActive && c.DeactivatedAt == nil {
		now := time.Now()
		c.DeactivatedAt = &now
	}
	return
}
