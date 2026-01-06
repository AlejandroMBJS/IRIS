/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/incidence_category.go
==============================================================================

DESCRIPTION:
    Defines the IncidenceCategory model for managing incidence type categories.
    Categories are the top-level grouping for incidence types, allowing admins
    to organize and manage types in a hierarchical structure.

USER PERSPECTIVE:
    - Categories appear in the admin "Incidencias > Categorias" section
    - Admins can create, edit, and organize categories
    - Each incidence type belongs to exactly one category
    - Some categories are system-protected (cannot be deleted)

DEVELOPER GUIDELINES:
    OK to modify: Add new fields, update validation logic
    CAUTION: IsSystem flag protects categories from deletion
    DO NOT modify: Existing system category codes without migration

CATEGORIES RELATIONSHIP:
    IncidenceCategory (1) → (many) IncidenceType

==============================================================================
*/
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IncidenceCategory represents a top-level category for incidence types.
// Categories organize incidence types into logical groups like "absence", "vacation", etc.
type IncidenceCategory struct {
	BaseModel
	Name          string `gorm:"type:varchar(255);not null" json:"name"`
	Code          string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Description   string `gorm:"type:text" json:"description,omitempty"`
	Color         string `gorm:"type:varchar(100)" json:"color,omitempty"`         // CSS class for badge styling
	Icon          string `gorm:"type:varchar(50)" json:"icon,omitempty"`           // Icon name for UI
	DisplayOrder  int    `gorm:"default:0" json:"display_order"`                   // For sorting in UI
	IsRequestable bool   `gorm:"default:false" json:"is_requestable"`              // Can employees create requests for this category?
	IsSystem      bool   `gorm:"default:false" json:"is_system"`                   // Protected from deletion
	IsActive      bool   `gorm:"default:true" json:"is_active"`                    // Show in UI?

	// Relations - IncidenceTypes belonging to this category
	IncidenceTypes []IncidenceType `gorm:"foreignKey:CategoryID" json:"incidence_types,omitempty"`
}

// TableName specifies the table name
func (IncidenceCategory) TableName() string {
	return "incidence_categories"
}

// BeforeCreate hook for IncidenceCategory
func (ic *IncidenceCategory) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate UUID if not set
	if ic.ID == uuid.Nil {
		ic.ID = uuid.New()
	}
	if ic.Name == "" {
		return ErrNameRequired
	}
	if ic.Code == "" {
		return ErrCodeRequired
	}
	return
}

// CanDelete checks if the category can be deleted
func (ic *IncidenceCategory) CanDelete() bool {
	return !ic.IsSystem
}

// DefaultIncidenceCategories returns the default system categories to seed
func DefaultIncidenceCategories() []IncidenceCategory {
	return []IncidenceCategory{
		{
			Name:          "Falta",
			Code:          "absence",
			Description:   "Ausencias no programadas del empleado",
			Color:         "bg-red-100 text-red-800",
			Icon:          "user-x",
			DisplayOrder:  1,
			IsRequestable: true,
			IsSystem:      false,
			IsActive:      true,
		},
		{
			Name:          "Enfermedad",
			Code:          "sick",
			Description:   "Incapacidades por enfermedad",
			Color:         "bg-orange-100 text-orange-800",
			Icon:          "heart-pulse",
			DisplayOrder:  2,
			IsRequestable: true,
			IsSystem:      true, // System-protected
			IsActive:      true,
		},
		{
			Name:          "Vacaciones",
			Code:          "vacation",
			Description:   "Dias de vacaciones del empleado",
			Color:         "bg-green-100 text-green-800",
			Icon:          "palm-tree",
			DisplayOrder:  3,
			IsRequestable: true,
			IsSystem:      true, // System-protected
			IsActive:      true,
		},
		{
			Name:          "Tiempo Extra",
			Code:          "overtime",
			Description:   "Horas extras trabajadas",
			Color:         "bg-blue-100 text-blue-800",
			Icon:          "clock",
			DisplayOrder:  4,
			IsRequestable: false,
			IsSystem:      false,
			IsActive:      true,
		},
		{
			Name:          "Retardo",
			Code:          "delay",
			Description:   "Llegadas tarde o salidas tempranas",
			Color:         "bg-yellow-100 text-yellow-800",
			Icon:          "alarm-clock",
			DisplayOrder:  5,
			IsRequestable: true,
			IsSystem:      false,
			IsActive:      true,
		},
		{
			Name:          "Bono",
			Code:          "bonus",
			Description:   "Bonos y compensaciones adicionales",
			Color:         "bg-purple-100 text-purple-800",
			Icon:          "gift",
			DisplayOrder:  6,
			IsRequestable: false,
			IsSystem:      false,
			IsActive:      true,
		},
		{
			Name:          "Deduccion",
			Code:          "deduction",
			Description:   "Deducciones adicionales",
			Color:         "bg-pink-100 text-pink-800",
			Icon:          "minus-circle",
			DisplayOrder:  7,
			IsRequestable: false,
			IsSystem:      false,
			IsActive:      true,
		},
		{
			Name:          "Otro",
			Code:          "other",
			Description:   "Otras incidencias",
			Color:         "bg-gray-100 text-gray-800",
			Icon:          "file-question",
			DisplayOrder:  8,
			IsRequestable: true,
			IsSystem:      false,
			IsActive:      true,
		},
	}
}
