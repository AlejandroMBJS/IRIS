/*
Package dtos - Authentication Data Transfer Objects

==============================================================================
FILE: internal/dtos/auth.go
==============================================================================

DESCRIPTION:
    Defines request and response structures for authentication endpoints
    including login, registration, password management, and token refresh.

USER PERSPECTIVE:
    - Shapes the JSON format for login/register API requests
    - Defines what user data is returned after authentication
    - Controls password change and reset flows

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields for additional user data
    ⚠️  CAUTION: Changing validation rules (affects frontend)
    ❌  DO NOT modify: Password field requirements without security review
    📝  Keep binding tags aligned with frontend validation

SYNTAX EXPLANATION:
    - binding:"required": Field must be present in request
    - binding:"email": Validates email format
    - binding:"min=8": Minimum string length validation
    - binding:"oneof=...": Enum-like validation

==============================================================================
*/
package dtos

import (
    "time"

    "backend/internal/models/enums"
)

// LoginRequest represents login request data
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	CompanyName string         `json:"company_name" binding:"required"`
	CompanyRFC  string         `json:"company_rfc" binding:"required"`
	Email       string         `json:"email" binding:"required,email"`
	Password    string         `json:"password" binding:"required,min=8"`
	Role        enums.UserRole `json:"role" binding:"required,oneof=admin hr accountant viewer"`
	FullName    string         `json:"full_name" binding:"required,min=2"`
}

// LoginResponse represents login response data
type LoginResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    TokenType    string       `json:"token_type"`
    ExpiresIn    int          `json:"expires_in"`
    User         UserResponse `json:"user"`
    SessionID    string       `json:"session_id,omitempty"`
}

// UserResponse represents user data in responses
type UserResponse struct {
    ID         string    `json:"id"`
    Email      string    `json:"email"`
    Role       string    `json:"role"`
    FullName   string    `json:"full_name"`
    IsActive   bool      `json:"is_active"`
    EmployeeID *string   `json:"employee_id,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" binding:"required"`
    NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
    Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
    Token       string `json:"token" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}
