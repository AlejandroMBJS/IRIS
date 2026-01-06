/*
Package services - Authentication Service

==============================================================================
FILE: internal/services/auth_service.go
==============================================================================

DESCRIPTION:
    Handles user authentication and authorization including registration, login,
    password management, and JWT token generation/validation.

USER PERSPECTIVE:
    - Users can register new companies with admin accounts
    - Secure login with password verification and JWT tokens
    - Password reset and change functionality
    - Profile management for authenticated users

DEVELOPER GUIDELINES:
    OK to modify: Password validation rules, token expiration times
    CAUTION: JWT token generation logic, ensure proper security
    DO NOT modify: Core authentication flow without security review
    Note: Always hash passwords using bcrypt, never store plain text

SYNTAX EXPLANATION:
    - RegisterRequest creates both Company and User in a transaction
    - Login returns JWT access and refresh tokens
    - JWT tokens contain UserID, Email, and Role claims
    - CheckPassword uses bcrypt.CompareHashAndPassword for verification

==============================================================================
*/
package services

import (
    "fmt"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "backend/internal/config"
    "backend/internal/dtos"
    apperr "backend/internal/errors"
    "backend/internal/models"
    "backend/internal/models/enums"
    "backend/internal/repositories"
    "backend/internal/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo    *repositories.UserRepository
	companyRepo *repositories.CompanyRepository
	jwtConfig   *utils.JWTConfig
	db          *gorm.DB
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, appConfig *config.AppConfig) *AuthService {
	jwtConfig := utils.NewJWTConfig(
		appConfig.JWTSecret,
		appConfig.JWTExpirationHours,
		appConfig.JWTRefreshHours,
	)

	return &AuthService{
		userRepo:    repositories.NewUserRepository(db),
		companyRepo: repositories.NewCompanyRepository(db),
		jwtConfig:   jwtConfig,
		db:          db,
	}
}

// Register handles the creation of a new company and its first admin user.
func (s *AuthService) Register(req dtos.RegisterRequest) (*dtos.LoginResponse, error) {
	// Check if user email already exists
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, apperr.ErrEmailAlreadyExists
	}

	// Check if company RFC already exists
	if _, err := s.companyRepo.FindByRFC(req.CompanyRFC); err == nil {
		return nil, apperr.ErrRFCAlreadyExists
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Create new company
	company := &models.Company{
		Name: req.CompanyName,
		RFC:  req.CompanyRFC,
	}
	if err := tx.Create(company).Error; err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// Create new user, ensuring it's an admin for the new company
	user := &models.User{
		Email:     req.Email,
		Role:      enums.RoleAdmin, // First user of a company is always an admin
		FullName:  req.FullName,
		IsActive:  true,
		CompanyID: company.ID,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	if err := tx.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.generateLoginResponse(user)
}

// Login authenticates a user
func (s *AuthService) Login(req dtos.LoginRequest) (*dtos.LoginResponse, error) {
    // Find user by email
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, apperr.ErrInvalidCredentials
        }
        return nil, apperr.Wrap(err, apperr.ErrDatabaseOperation)
    }

    // Check if user is active
    if !user.IsActive {
        return nil, apperr.ErrAccountDeactivated
    }

    // Verify password
    if !user.CheckPassword(req.Password) {
        return nil, apperr.ErrInvalidCredentials
    }
    
    // Update last login
    now := time.Now()
    user.LastLoginAt = &now
    if err := s.userRepo.Update(user); err != nil {
        return nil, fmt.Errorf("failed to update last login: %w", err)
    }
    
    // Generate tokens
    return s.generateLoginResponse(user)
}

// RefreshToken refreshes an access token
func (s *AuthService) RefreshToken(refreshToken string) (*dtos.LoginResponse, error) {
    // Validate refresh token
    claims, err := s.jwtConfig.ValidateRefreshToken(refreshToken)
    if err != nil {
        return nil, apperr.Wrap(err, apperr.ErrRefreshTokenInvalid)
    }

    // Find user
    user, err := s.userRepo.FindByID(claims.UserID)
    if err != nil {
        return nil, apperr.Wrap(err, apperr.ErrNotFound)
    }

    // Check if user is active
    if !user.IsActive {
        return nil, apperr.ErrAccountDeactivated
    }

    // Generate new tokens
    return s.generateLoginResponse(user)
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID uuid.UUID, req dtos.ChangePasswordRequest) error {
    // Find user
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return apperr.Wrap(err, apperr.ErrNotFound)
    }

    // Verify current password
    if !user.CheckPassword(req.CurrentPassword) {
        return apperr.ErrPasswordMismatch
    }

    // Set new password
    if err := user.SetPassword(req.NewPassword); err != nil {
        return apperr.Wrap(err, apperr.ErrPasswordTooWeak)
    }

    // Update user
    if err := s.userRepo.Update(user); err != nil {
        return apperr.Wrap(err, apperr.ErrDatabaseOperation)
    }

    return nil
}

// ForgotPassword initiates password reset
func (s *AuthService) ForgotPassword(email string) (string, error) {
    // Find user by email
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        // Don't reveal if user exists or not
        return "", nil
    }
    
    // Generate reset token
    resetToken, err := s.jwtConfig.GeneratePasswordResetToken(user.ID, user.Email)
    if err != nil {
        return "", fmt.Errorf("failed to generate reset token: %w", err)
    }
    
    // In production, send email with reset token
    // For now, just return the token (in production, this would be sent via email)
    
    return resetToken, nil
}

// ResetPassword resets password using reset token
func (s *AuthService) ResetPassword(req dtos.ResetPasswordRequest) error {
    // Validate reset token
    claims, err := s.jwtConfig.ValidateToken(req.Token)
    if err != nil {
        return apperr.Wrap(err, apperr.ErrInvalidToken)
    }

    if claims.TokenType != "password_reset" {
        return apperr.ErrInvalidToken.WithMessage("Invalid token type")
    }

    // Find user
    user, err := s.userRepo.FindByID(claims.UserID)
    if err != nil {
        return apperr.Wrap(err, apperr.ErrNotFound)
    }

    // Set new password
    if err := user.SetPassword(req.NewPassword); err != nil {
        return apperr.Wrap(err, apperr.ErrPasswordTooWeak)
    }

    // Update user
    if err := s.userRepo.Update(user); err != nil {
        return apperr.Wrap(err, apperr.ErrDatabaseOperation)
    }

    return nil
}

// GetUserProfile gets user profile
func (s *AuthService) GetUserProfile(userID uuid.UUID) (*dtos.UserResponse, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, apperr.Wrap(err, apperr.ErrNotFound)
    }
    
    return &dtos.UserResponse{
        ID:        user.ID.String(),
        Email:     user.Email,
        Role:      user.Role.String(),
        FullName:  user.FullName,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
    }, nil
}

// UpdateUserProfile updates user profile
func (s *AuthService) UpdateUserProfile(userID uuid.UUID, fullName string) error {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return apperr.Wrap(err, apperr.ErrNotFound)
    }

    user.FullName = fullName
    if err := s.userRepo.Update(user); err != nil {
        return apperr.Wrap(err, apperr.ErrDatabaseOperation)
    }

    return nil
}

// generateLoginResponse generates login response with tokens
func (s *AuthService) generateLoginResponse(user *models.User) (*dtos.LoginResponse, error) {
    // Generate tokens
    accessToken, refreshToken, err := s.jwtConfig.GenerateTokenPair(user.ID, user.Email, user.Role)
    if err != nil {
        return nil, fmt.Errorf("failed to generate tokens: %w", err)
    }

    // Convert EmployeeID to string pointer if it exists
    var employeeIDStr *string
    if user.EmployeeID != nil {
        str := user.EmployeeID.String()
        employeeIDStr = &str
    }

    // Create response
    return &dtos.LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    int(s.jwtConfig.AccessTokenExpiry.Seconds()),
        User: dtos.UserResponse{
            ID:         user.ID.String(),
            Email:      user.Email,
            Role:       user.Role.String(),
            FullName:   user.FullName,
            IsActive:   user.IsActive,
            EmployeeID: employeeIDStr,
            CreatedAt:  user.CreatedAt,
        },
    }, nil
}

// Logout logs out a user (in production, you might blacklist tokens)
func (s *AuthService) Logout(userID uuid.UUID) error {
    // In a production system, you might:
    // 1. Add the token to a blacklist
    // 2. Invalidate refresh tokens
    // 3. Track logout in audit log
    
    // For this implementation, we'll just update last logout time
    // In a real system with token blacklisting, you'd need a different approach
    
    return nil
}

// VerifyToken verifies an access token and returns user
func (s *AuthService) VerifyToken(accessToken string) (*models.User, error) {
    claims, err := s.jwtConfig.ValidateAccessToken(accessToken)
    if err != nil {
        return nil, apperr.Wrap(err, apperr.ErrInvalidToken)
    }

    user, err := s.userRepo.FindByID(claims.UserID)
    if err != nil {
        return nil, apperr.Wrap(err, apperr.ErrNotFound)
    }

    if !user.IsActive {
        return nil, apperr.ErrAccountDeactivated
    }

    return user, nil
}
