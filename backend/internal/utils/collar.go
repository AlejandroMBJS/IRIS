/*
Package utils - Collar Type Access Control Utilities

==============================================================================
FILE: internal/utils/collar.go
==============================================================================

DESCRIPTION:
    Provides utility functions for collar-type based access control.
    Used by HR role filtering to restrict access to employees based on their
    collar type (white_collar, blue_collar, gray_collar).

USER PERSPECTIVE:
    - hr_blue_gray role can ONLY see blue_collar and gray_collar employees
    - hr_white role can ONLY see white_collar employees
    - hr and admin roles see ALL employees

DEVELOPER GUIDELINES:
    OK to modify: Add new collar types, modify role mappings
    CAUTION: Ensure consistency with roles defined in models/enums/roles.go
    DO NOT modify: Function signatures (breaks callers)

==============================================================================
*/
package utils

// GetAllowedCollarTypes returns the collar types a role is allowed to access.
// Returns nil for roles with unrestricted access (admin, hr, etc.)
// Returns specific collar types for restricted HR roles.
func GetAllowedCollarTypes(role string) []string {
	switch role {
	case "hr_blue_gray":
		return []string{"blue_collar", "gray_collar"}
	case "hr_white":
		return []string{"white_collar"}
	default:
		// All other roles have unrestricted collar type access
		// Their access is controlled by other means (e.g., supervisor relationship)
		return nil
	}
}

// IsHRRole checks if a role is any HR role (including sub-HR roles)
func IsHRRole(role string) bool {
	switch role {
	case "hr", "hr_and_pr", "hr_blue_gray", "hr_white":
		return true
	default:
		return false
	}
}

// CanManageCollarType checks if a role can manage employees of a specific collar type
func CanManageCollarType(role string, collarType string) bool {
	allowedTypes := GetAllowedCollarTypes(role)
	if allowedTypes == nil {
		// nil means unrestricted access
		return true
	}
	for _, allowed := range allowedTypes {
		if allowed == collarType {
			return true
		}
	}
	return false
}

// IntersectCollarTypes returns the intersection of two collar type slices.
// Used when combining role-based restrictions with request filters.
func IntersectCollarTypes(a, b []string) []string {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	set := make(map[string]bool)
	for _, v := range b {
		set[v] = true
	}
	var result []string
	for _, v := range a {
		if set[v] {
			result = append(result, v)
		}
	}
	return result
}

// IsManagerRole checks if a role requires hierarchical filtering (supervisor, manager)
func IsManagerRole(role string) bool {
	switch role {
	case "supervisor", "manager":
		return true
	default:
		return false
	}
}

// RequiresEmployeeFiltering checks if a role requires filtering to only show specific employees
// (either by hierarchy or by collar type)
func RequiresEmployeeFiltering(role string) bool {
	return IsManagerRole(role) || IsHRRole(role)
}

// CanViewAllEmployees checks if a role can view all employees without restrictions
func CanViewAllEmployees(role string) bool {
	switch role {
	case "admin", "hr", "hr_and_pr", "payroll_staff", "accountant":
		return true
	default:
		return false
	}
}
