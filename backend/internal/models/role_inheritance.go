/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/role_inheritance.go
==============================================================================

DESCRIPTION:
    Defines role inheritance relationships for permission management.
    Allows roles to inherit permissions from parent roles, simplifying
    access control configuration.

USER PERSPECTIVE:
    - Administrators can define which roles inherit from others
    - Example: admin inherits from hr, manager, payroll (has all their permissions)
    - Reduces manual permission configuration
    - Clear hierarchy visualization in UI

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new role types, inheritance validation rules
    ⚠️  CAUTION: Circular dependency detection (prevent infinite loops)
    ❌  DO NOT modify: Core role names without updating middleware
    📝  Role names must match those used in auth middleware

INHERITANCE EXAMPLES:
    admin → hr, manager, payroll, supervisor (inherits all permissions)
    hr → supervisor (HR can do everything supervisors can)
    manager → supervisor (Managers can do everything supervisors can)
    payroll → hr (Payroll has HR permissions plus payroll-specific ones)

BUSINESS RULES:
    - No circular dependencies (role A inherits from B, B cannot inherit from A)
    - Multiple inheritance allowed (admin can inherit from hr + manager + payroll)
    - Transitive inheritance (if A inherits B, and B inherits C, A gets C's permissions)
    - Active flag allows temporary disable without deletion

==============================================================================
*/
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleInheritance defines a relationship where one role inherits permissions from another
// Example: admin inherits from hr, meaning admin users have all hr permissions
type RoleInheritance struct {
	BaseModel
	ChildRole  string `gorm:"type:varchar(50);not null;index:idx_role_inheritance_child" json:"child_role"`  // Role that inherits
	ParentRole string `gorm:"type:varchar(50);not null;index:idx_role_inheritance_parent" json:"parent_role"` // Role to inherit from
	IsActive   bool   `gorm:"default:true" json:"is_active"`                                                  // Can be disabled without deletion
	Priority   int    `gorm:"default:0" json:"priority"`                                                      // Higher priority = checked first (for conflict resolution)
	Notes      string `gorm:"type:text" json:"notes,omitempty"`                                               // Admin notes
}

// TableName specifies the table name for GORM
func (RoleInheritance) TableName() string {
	return "role_inheritances"
}

// BeforeCreate hook for validation
func (r *RoleInheritance) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	// Prevent self-inheritance
	if r.ChildRole == r.ParentRole {
		return gorm.ErrInvalidValue
	}

	return nil
}

// ValidRoles returns the list of valid role names in the system
func ValidRoles() []string {
	return []string{
		"admin",
		"hr",
		"hr_and_pr",
		"hr_blue_gray",
		"hr_white",
		"manager",
		"gm", // general_manager
		"supervisor",
		"payroll",
		"accountant",
		"employee",
	}
}

// IsValidRole checks if a role name is valid
func IsValidRole(role string) bool {
	validRoles := ValidRoles()
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// RolePermissionLevel returns a numeric level for role hierarchy (higher = more permissions)
// Used for default inheritance suggestions
func RolePermissionLevel(role string) int {
	levels := map[string]int{
		"employee":      1,
		"supervisor":    2,
		"hr_blue_gray":  3,
		"hr_white":      3,
		"hr":            4,
		"manager":       4,
		"accountant":    4,
		"payroll":       5,
		"hr_and_pr":     6,
		"gm":            7,
		"admin":         10,
	}
	if level, ok := levels[role]; ok {
		return level
	}
	return 0
}
