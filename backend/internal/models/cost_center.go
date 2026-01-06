/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/cost_center.go
==============================================================================

DESCRIPTION:
    Defines CostCenter model for organizing employees into departments or
    cost allocation units. Used for financial reporting and payroll grouping.

USER PERSPECTIVE:
    - Cost centers appear in employee assignment dropdown
    - Used to group employees by department for reports
    - Payroll can be filtered/summarized by cost center
    - Examples: "Produccion", "Administracion", "Ventas"

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields (budget, manager, etc.)
    ⚠️  CAUTION: Deleting cost centers with assigned employees
    ❌  DO NOT modify: CompanyID isolation (multi-tenancy)
    📝  Each company manages their own cost centers independently

SYNTAX EXPLANATION:
    - CompanyID: Links to owning company (multi-tenancy)
    - Employees relation: All employees assigned to this cost center
    - BeforeCreate/BeforeUpdate: GORM hooks for validation
    - ErrNameRequired/ErrCompanyIDRequired: Defined in errors.go

MULTI-TENANCY:
    - Cost centers are company-specific
    - Users can only see cost centers from their own company
    - Always filter queries by CompanyID from JWT token

==============================================================================
*/
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CostCenter represents a cost center or department within a company.
type CostCenter struct {
	BaseModel
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	CompanyID   uuid.UUID  `gorm:"type:text;not null" json:"company_id"`
	Company     *Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Employees   []Employee `gorm:"foreignKey:CostCenterID" json:"employees,omitempty"`
}

// TableName specifies the table name
func (CostCenter) TableName() string {
	return "cost_centers"
}

// BeforeCreate hook to validate cost center data before creation.
func (cc *CostCenter) BeforeCreate(tx *gorm.DB) (err error) {
	if cc.Name == "" {
		return ErrNameRequired
	}
	if cc.CompanyID == uuid.Nil {
		return ErrCompanyIDRequired
	}
	return
}

// BeforeUpdate hook to validate cost center data before update.
func (cc *CostCenter) BeforeUpdate(tx *gorm.DB) (err error) {
	if cc.Name == "" {
		return ErrNameRequired
	}
	return
}
