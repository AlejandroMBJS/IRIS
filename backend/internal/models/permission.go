/*
Package models - Permission Matrix Model

==============================================================================
FILE: internal/models/permission.go
==============================================================================

DESCRIPTION:
    Defines role-based permission matrix for access control configuration.
    Allows admin to configure what each role can see and do in the system.

USER PERSPECTIVE:
    - Admin configures permissions through permission matrix UI
    - Each role (supervisor, manager, hr, etc.) has specific permissions
    - Permissions control access to features and data visibility

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new permission types, add new resources
    ⚠️  CAUTION: Changing permission keys may break frontend checks
    ❌  DO NOT modify: BaseModel fields, primary keys
    📝  Permission keys should be consistent with frontend auth checks

PERMISSION TYPES:
    - can_view: Can view the resource
    - can_create: Can create new records
    - can_edit: Can modify existing records
    - can_delete: Can delete records
    - can_export: Can export data
    - can_approve: Can approve requests

RESOURCES:
    - employees: Employee management
    - payroll: Payroll processing
    - reports: Report generation
    - configuration: System configuration
    - approvals: Absence request approvals
    - incidences: Incidence management
==============================================================================
*/
package models

import (
	"encoding/json"
)

// Permission represents a single permission entry in the permission matrix
type Permission struct {
	BaseModel
	Role        string          `gorm:"type:varchar(50);not null;uniqueIndex:idx_permission_role_resource" json:"role"`
	Resource    string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_permission_role_resource" json:"resource"`
	Permissions json.RawMessage `gorm:"type:jsonb" json:"permissions"` // JSONB for PostgreSQL, TEXT for SQLite
	Description string          `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool            `gorm:"default:true" json:"is_active"`
}

// PermissionSet defines the structure of permissions for a resource
type PermissionSet struct {
	CanView    bool `json:"can_view"`
	CanCreate  bool `json:"can_create"`
	CanEdit    bool `json:"can_edit"`
	CanDelete  bool `json:"can_delete"`
	CanExport  bool `json:"can_export"`
	CanApprove bool `json:"can_approve"`
}

// GetPermissions parses the JSONB permissions field
func (p *Permission) GetPermissions() (*PermissionSet, error) {
	var perms PermissionSet
	if err := json.Unmarshal(p.Permissions, &perms); err != nil {
		return nil, err
	}
	return &perms, nil
}

// SetPermissions sets the JSONB permissions field
func (p *Permission) SetPermissions(perms *PermissionSet) error {
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	p.Permissions = data
	return nil
}

// ValidResources returns the list of all permission resources
func ValidResources() []string {
	return []string{
		"employees",
		"payroll",
		"reports",
		"configuration",
		"approvals",
		"incidences",
		"users",
		"announcements",
		"calendar",
		"shifts",
		"inbox",
	}
}

// TableName specifies the table name for Permission model
func (Permission) TableName() string {
	return "permissions"
}
