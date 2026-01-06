/*
Package services - User Management Service

==============================================================================
FILE: internal/services/user_service.go
==============================================================================

DESCRIPTION:
    Manages system users within a company including user creation, deletion,
    and status management. Enforces company-level user isolation and role-based
    access control.

USER PERSPECTIVE:
    - Admins can create users in their company
    - Users are isolated by company (multi-tenant)
    - Activate/deactivate user accounts
    - Delete inactive users

DEVELOPER GUIDELINES:
    OK to modify: User validation rules, add new roles
    CAUTION: Company isolation is critical for security
    DO NOT modify: Admin self-deletion protection
    Note: Only non-admin roles can be created through this service

SYNTAX EXPLANATION:
    - CreateUser requires company admin authentication
    - Users belong to a single CompanyID (multi-tenant isolation)
    - Role enum: admin, manager, accountant, hr, readonly
    - ToggleUserActive prevents admins from deactivating themselves
    - ToResponseDTO hides sensitive data (password hash)

==============================================================================
*/
package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
	"backend/internal/models/enums"
	"backend/internal/repositories"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Password string `json:"password,omitempty"`
}

// UserService handles user management business logic
type UserService struct {
	userRepo *repositories.UserRepository
	db       *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userRepo: repositories.NewUserRepository(db),
		db:       db,
	}
}

// GetUsersByCompany returns all users in a company
func (s *UserService) GetUsersByCompany(companyID uuid.UUID) ([]map[string]interface{}, error) {
	users, err := s.userRepo.FindByCompanyID(companyID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = user.ToResponseDTO()
	}
	return result, nil
}

// CreateUser creates a new user in the same company as the admin
func (s *UserService) CreateUser(companyID uuid.UUID, req CreateUserRequest) (map[string]interface{}, error) {
	// Validate role - only allow non-admin roles
	role := enums.UserRole(req.Role)
	if !role.IsValid() {
		return nil, errors.New("invalid role")
	}
	if role == enums.RoleAdmin {
		return nil, errors.New("cannot create admin users through this endpoint")
	}

	// Check if email already exists (active users only)
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("a user with this email already exists")
	}

	// Check if a soft-deleted user exists with this email and remove it
	// This allows re-registering users that were previously deleted
	existsDeleted, err := s.userRepo.ExistsByEmailIncludingDeleted(req.Email)
	if err != nil {
		return nil, err
	}
	if existsDeleted {
		// Hard delete the soft-deleted user to free up the email
		if err := s.userRepo.HardDeleteByEmail(req.Email); err != nil {
			return nil, err
		}
	}

	// Create new user
	user := &models.User{
		Email:     req.Email,
		Role:      role,
		FullName:  req.FullName,
		IsActive:  true,
		CompanyID: companyID,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user.ToResponseDTO(), nil
}

// UpdateUser updates a user's role, name and/or password
func (s *UserService) UpdateUser(adminID, userID, companyID uuid.UUID, req UpdateUserRequest) (map[string]interface{}, error) {
	// Get the user to update
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Verify user belongs to the same company
	if user.CompanyID != companyID {
		return nil, errors.New("user not found in your company")
	}

	// Update full name if provided
	if req.FullName != "" {
		user.FullName = req.FullName
	}

	// Update role if provided
	if req.Role != "" {
		// Cannot change your own role
		if adminID == userID {
			return nil, errors.New("cannot change your own role")
		}

		role := enums.UserRole(req.Role)
		if !role.IsValid() {
			return nil, errors.New("invalid role")
		}
		if role == enums.RoleAdmin {
			return nil, errors.New("cannot change role to admin")
		}
		user.Role = role
	}

	// Update password if provided (admin resetting user password)
	if req.Password != "" {
		if len(req.Password) < 8 {
			return nil, errors.New("password must be at least 8 characters")
		}
		if err := user.SetPassword(req.Password); err != nil {
			return nil, errors.New("failed to set password")
		}
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user.ToResponseDTO(), nil
}

// DeleteUser deletes a user (admin cannot delete themselves)
func (s *UserService) DeleteUser(adminID, userID, companyID uuid.UUID) error {
	// Cannot delete yourself
	if adminID == userID {
		return errors.New("cannot delete yourself")
	}

	// Get the user to delete
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Verify user belongs to the same company
	if user.CompanyID != companyID {
		return errors.New("user not found in your company")
	}

	return s.userRepo.Delete(userID)
}

// ToggleUserActive toggles user active status
func (s *UserService) ToggleUserActive(adminID, userID, companyID uuid.UUID) (map[string]interface{}, error) {
	// Cannot toggle yourself
	if adminID == userID {
		return nil, errors.New("cannot deactivate yourself")
	}

	// Get the user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Verify user belongs to the same company
	if user.CompanyID != companyID {
		return nil, errors.New("user not found in your company")
	}

	// Toggle status
	user.IsActive = !user.IsActive
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user.ToResponseDTO(), nil
}
