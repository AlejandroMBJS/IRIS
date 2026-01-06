/*
Package enums - IRIS Payroll System Enumeration Types

==============================================================================
FILE: internal/models/enums/roles.go
==============================================================================

DESCRIPTION:
    Defines the UserRole type and constants for role-based access control.
    Roles determine what features and data each user can access.

USER PERSPECTIVE:
    - Roles are assigned when creating users
    - Each role has different permissions:
        * admin: Full system access, can manage users and settings
        * hr: Human resources access, can manage employees
        * accountant: Financial reports and payroll approval
        * payroll_staff: Day-to-day payroll operations
        * viewer: Read-only access to reports

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new roles (update IsValid() too)
    ⚠️  CAUTION: Changing existing role names (breaks database data)
    ❌  DO NOT modify: Remove existing roles without migration
    📝  Update middleware/authorization when adding new roles

SYNTAX EXPLANATION:
    - type UserRole string: Type alias for type safety
    - const block with iota-like pattern: Define all role values
    - IsValid(): Validates role value is one of the constants
    - MarshalText/UnmarshalText: JSON serialization support
    - strings.ToLower: Case-insensitive deserialization

AUTHORIZATION:
    - Roles are stored in JWT token claims
    - Middleware checks role for protected endpoints
    - API handlers may have role-specific behavior
    - See internal/api/middleware/auth.go for role checks

ROLE HIERARCHY:
    admin > hr > accountant > payroll_staff > viewer

==============================================================================
*/
package enums

import "strings"

// UserRole represents the role of a user in the system.
type UserRole string

const (
	RoleAdmin        UserRole = "admin"
	RoleHR           UserRole = "hr"
	RoleAccountant   UserRole = "accountant"
	RolePayrollStaff UserRole = "payroll_staff"
	RoleViewer       UserRole = "viewer"
	// New roles for approval workflow
	RoleSupervisor   UserRole = "supervisor"
	RoleManager      UserRole = "manager"
	RoleEmployee     UserRole = "employee"
	RoleHRAndPR      UserRole = "hr_and_pr"      // HR + Payroll combined
	RoleSupAndGM     UserRole = "sup_and_gm"     // Supervisor + General Manager combined
	RoleHRBlueGray   UserRole = "hr_blue_gray"   // HR for blue_collar and gray_collar employees only
	RoleHRWhite      UserRole = "hr_white"       // HR for white_collar employees only
)

// IsValid checks if the user role is valid.
func (ur UserRole) IsValid() bool {
	switch ur {
	case RoleAdmin, RoleHR, RoleAccountant, RolePayrollStaff, RoleViewer,
		RoleSupervisor, RoleManager, RoleEmployee, RoleHRAndPR, RoleSupAndGM,
		RoleHRBlueGray, RoleHRWhite:
		return true
	}
	return false
}

// String returns the string representation of the user role.
func (ur UserRole) String() string {
	return string(ur)
}

// MarshalText implements encoding.TextMarshaler for JSON serialization.
func (ur UserRole) MarshalText() ([]byte, error) {
	return []byte(ur.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON deserialization.
func (ur *UserRole) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch s {
	case "admin":
		*ur = RoleAdmin
	case "hr":
		*ur = RoleHR
	case "accountant":
		*ur = RoleAccountant
	case "payroll_staff":
		*ur = RolePayrollStaff
	case "viewer":
		*ur = RoleViewer
	case "supervisor":
		*ur = RoleSupervisor
	case "manager":
		*ur = RoleManager
	case "employee":
		*ur = RoleEmployee
	case "hr_and_pr":
		*ur = RoleHRAndPR
	case "sup_and_gm":
		*ur = RoleSupAndGM
	case "hr_blue_gray":
		*ur = RoleHRBlueGray
	case "hr_white":
		*ur = RoleHRWhite
	default:
		*ur = "" // Invalid role
	}
	return nil
}
