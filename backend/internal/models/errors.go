/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/errors.go
==============================================================================

DESCRIPTION:
    Defines common validation errors used across model BeforeCreate and
    BeforeUpdate hooks. Centralizes error messages for consistency.

USER PERSPECTIVE:
    - These errors appear when validation fails during create/update
    - Clear messages help users understand what went wrong
    - Consistent error format across all API endpoints

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new error variables for new validations
    ⚠️  CAUTION: Changing existing error messages (may break tests)
    ❌  DO NOT modify: Variable names (used throughout codebase)
    📝  Keep messages user-friendly and actionable

SYNTAX EXPLANATION:
    - var: Package-level variables (exported if capitalized)
    - errors.New(): Creates a simple error with message
    - Err prefix: Convention for error variables in Go
    - Used in BeforeCreate/BeforeUpdate hooks for validation

USAGE:
    if model.Name == "" {
        return ErrNameRequired
    }

==============================================================================
*/
package models

import "errors"

// Common errors
var (
	ErrNameRequired        = errors.New("name is required")
	ErrCompanyIDRequired   = errors.New("company ID is required")
	ErrCategoryRequired    = errors.New("category is required")
	ErrEffectTypeRequired  = errors.New("effect type is required")
	ErrConceptTypeRequired = errors.New("concept type is required")
	ErrCodeRequired        = errors.New("code is required")
)
