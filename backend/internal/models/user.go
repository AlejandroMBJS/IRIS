/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/user.go
==============================================================================

DESCRIPTION:
    Defines the User model for authentication and authorization.
    Users are people who can log into the system to manage payroll operations.
    A user may optionally be linked to an Employee record (e.g., payroll staff
    who are also employees of the company).

USER PERSPECTIVE:
    - Users log into the system with email and password
    - Different roles have different permissions:
        * admin: Full access to all features
        * payroll_manager: Can manage payroll but not system settings
        * payroll_operator: Can create/edit payroll entries
        * viewer: Read-only access
    - Passwords must meet security requirements (8+ chars, upper, lower, digit, special)

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields, new roles (update enums/roles.go too)
    ⚠️  CAUTION: Password hashing, authentication methods
    ❌  DO NOT modify: SetPassword() bcrypt cost, password validation rules
    📝  Security-sensitive file - review changes carefully

SYNTAX EXPLANATION:
    - `json:"-"`: Field is NEVER included in JSON output (security)
    - enums.UserRole: Custom type for role validation
    - bcrypt.GenerateFromPassword(): Industry-standard password hashing
    - bcrypt.DefaultCost: Work factor for hashing (10 rounds)
    - *time.Time: Pointer allows null values in database

PASSWORD REQUIREMENTS:
    - Minimum 8 characters
    - At least one uppercase letter
    - At least one lowercase letter
    - At least one digit
    - At least one special character (!@#$%^&*())

RELATIONS:
    - Company: Belongs to one company
    - Employee: Optional link to employee record (for payroll team members)

==============================================================================
*/
package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"backend/internal/models/enums"
)

// User represents a user in the system.
// Users authenticate to access payroll features based on their role.
type User struct {
	BaseModel
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         enums.UserRole `gorm:"type:text;not null" json:"role"`
	FullName     string         `gorm:"type:varchar(255);not null" json:"full_name"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CompanyID    uuid.UUID      `gorm:"type:text;not null" json:"company_id"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`

	// Link to Employee if this user is also an employee (e.g., payroll team members)
	EmployeeID   *uuid.UUID     `gorm:"type:text" json:"employee_id,omitempty"`

	// Supervisor and General Manager relationships for approval workflow
	SupervisorID     *uuid.UUID `gorm:"type:text" json:"supervisor_id,omitempty"`
	GeneralManagerID *uuid.UUID `gorm:"type:text" json:"general_manager_id,omitempty"`
	Department       string     `gorm:"type:varchar(100)" json:"department,omitempty"`
	Area             string     `gorm:"type:varchar(100)" json:"area,omitempty"`

	// Relations
	Company        *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Employee       *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Supervisor     *User     `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
	GeneralManager *User     `gorm:"foreignKey:GeneralManagerID" json:"general_manager,omitempty"`
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// SetPassword hashes the password and sets it to the PasswordHash field.
func (u *User) SetPassword(password string) error {
	// Password validation (e.g., minimum length, complexity)
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one digit")
	}
	if !regexp.MustCompile(`[!@#$%^&*()]`).MatchString(password) {
		return errors.New("password must contain at least one special character (!@#$%^&*())")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword compares a plaintext password with the hashed password.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Validate validates user data
func (u *User) Validate() error {
	var validationErrors []string

	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$`).MatchString(u.Email) {
		validationErrors = append(validationErrors, "invalid email format")
	}
	if strings.TrimSpace(u.FullName) == "" {
		validationErrors = append(validationErrors, "full name is required")
	}
	if !u.Role.IsValid() {
		validationErrors = append(validationErrors, "invalid role")
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}
	return nil
}

// BeforeSave hook to validate user data before saving.
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	return u.Validate()
}

// ToResponseDTO converts the User model to a map suitable for API response,
// excluding sensitive information like PasswordHash.
func (u *User) ToResponseDTO() map[string]interface{} {
	dto := map[string]interface{}{
		"id":          u.ID.String(),
		"email":       u.Email,
		"role":        u.Role.String(),
		"full_name":   u.FullName,
		"is_active":   u.IsActive,
		"company_id":  u.CompanyID,
		"created_at":  u.CreatedAt,
		"updated_at":  u.UpdatedAt,
		"last_login_at": u.LastLoginAt,
	}
	if u.EmployeeID != nil {
		dto["employee_id"] = u.EmployeeID.String()
	}
	return dto
}
