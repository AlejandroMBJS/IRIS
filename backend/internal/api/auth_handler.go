/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/auth_handler.go
==============================================================================

DESCRIPTION:
    Handles all authentication-related endpoints: login, registration,
    password management, and user profile operations.

USER PERSPECTIVE:
    - Login/logout functionality
    - Initial company and admin user registration
    - Password change and reset flows
    - User profile viewing and editing

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new auth-related endpoints
    ⚠️  CAUTION: Token handling, password validation
    ❌  DO NOT modify: JWT token structure, security middleware
    📝  All responses use consistent error format

SYNTAX EXPLANATION:
    - @Summary/@Description: Swagger documentation annotations
    - c.ShouldBindJSON(): Parses and validates JSON request body
    - c.JSON(): Returns JSON response with status code
    - middleware.GetUserFromContext(): Extracts user from JWT

ENDPOINTS:
    POST /auth/register - Register new company + admin user
    POST /auth/login - Authenticate and get tokens
    POST /auth/refresh - Refresh expired access token
    POST /auth/logout - Invalidate tokens (requires auth)
    POST /auth/change-password - Change password (requires auth)
    POST /auth/forgot-password - Request password reset
    POST /auth/reset-password - Reset password with token
    GET  /auth/profile - Get user profile (requires auth)
    PUT  /auth/profile - Update user profile (requires auth)

SECURITY:
    - Passwords hashed with bcrypt
    - JWT tokens with configurable expiration
    - Refresh tokens for session management

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/config"
	"backend/internal/dtos"
	apperr "backend/internal/errors"
	"backend/internal/middleware"
	"backend/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
    authService  *services.AuthService
    auditService *services.AuditService
    appConfig    *config.AppConfig
}

// NewAuthHandler creates new authentication handler
func NewAuthHandler(authService *services.AuthService, auditService *services.AuditService, appConfig *config.AppConfig) *AuthHandler {
    return &AuthHandler{
        authService:  authService,
        auditService: auditService,
        appConfig:    appConfig,
    }
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
    // Create rate limiter for auth endpoints (10 requests/minute)
    authRateLimiter := middleware.AuthRateLimiter(h.appConfig)

    auth := router.Group("/auth")
    {
        // Rate-limited endpoints (unauthenticated, vulnerable to brute force)
        auth.POST("/register", authRateLimiter.Limit(), h.Register)
        auth.POST("/login", authRateLimiter.Limit(), h.Login)
        auth.POST("/refresh", authRateLimiter.Limit(), h.RefreshToken)
        auth.POST("/forgot-password", authRateLimiter.Limit(), h.ForgotPassword)
        auth.POST("/reset-password", authRateLimiter.Limit(), h.ResetPassword)

        // Authenticated endpoints (less vulnerable, normal rate limiting)
        auth.POST("/logout", middleware.NewAuthMiddleware(h.authService).RequireAuth(), h.Logout)
        auth.POST("/change-password", middleware.NewAuthMiddleware(h.authService).RequireAuth(), h.ChangePassword)
        auth.GET("/profile", middleware.NewAuthMiddleware(h.authService).RequireAuth(), h.GetProfile)
        auth.PUT("/profile", middleware.NewAuthMiddleware(h.authService).RequireAuth(), h.UpdateProfile)
    }
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user in the system
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.RegisterRequest true "Registration data"
// @Success 201 {object} dtos.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
    var req dtos.RegisterRequest
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Register user
    response, err := h.authService.Register(req)
    if err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        c.JSON(status, gin.H{
            "error":   "Registration Failed",
            "code":    code,
            "message": message,
        })
        return
    }
    
    c.JSON(http.StatusCreated, response)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.LoginRequest true "Login credentials"
// @Success 200 {object} dtos.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    var req dtos.LoginRequest

    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }

    // Login user
    response, err := h.authService.Login(req)
    if err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        // Log failed login attempt
        h.auditService.LogLoginAttempt(c, req.Email, false, &message, nil)

        c.JSON(status, gin.H{
            "error":   "Login Failed",
            "code":    code,
            "message": message,
        })
        return
    }

    // Log successful login - parse user ID to UUID
    userIDParsed, _ := uuid.Parse(response.User.ID)
    h.auditService.LogLoginAttempt(c, req.Email, true, nil, &userIDParsed)

    // Create login session
    sessionID, err := h.auditService.CreateLoginSession(
        userIDParsed,
        response.User.Email,
        c.ClientIP(),
        c.Request.UserAgent(),
    )
    if err == nil {
        response.SessionID = sessionID
    }

		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("access_token", response.AccessToken, 900, "/", "", h.appConfig.IsProduction(), true)
		c.SetCookie("refresh_token", response.RefreshToken, 604800, "/api/auth", "", h.appConfig.IsProduction(), true)
    c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh expired access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dtos.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req dtos.RefreshTokenRequest

    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }

    // Refresh token
    response, err := h.authService.RefreshToken(req.RefreshToken)
    if err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        c.JSON(status, gin.H{
            "error":   "Token Refresh Failed",
            "code":    code,
            "message": message,
        })
        return
    }

    // Log token refresh - parse user ID to UUID
    refreshUserID, _ := uuid.Parse(response.User.ID)
    h.auditService.LogTokenRefresh(c, refreshUserID, response.User.Email, true)

		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("access_token", response.AccessToken, 900, "/", "", h.appConfig.IsProduction(), true)
		c.SetCookie("refresh_token", response.RefreshToken, 604800, "/api/auth", "", h.appConfig.IsProduction(), true)
    c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout user and invalidate tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
    var userID uuid.UUID
    userID, email, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }

    // Get session ID from header or generate one
    sessionID := c.GetHeader("X-Session-ID")
    if sessionID == "" {
        sessionID = "unknown"
    }

    // Log logout event
    h.auditService.LogLogout(c, userID, email, sessionID)

    // Logout user
    if err := h.authService.Logout(userID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Logout Failed",
            "message": err.Error(),
        })
        return
    }

		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("access_token", "", -1, "/", "", h.appConfig.IsProduction(), true)
		c.SetCookie("refresh_token", "", -1, "/api/auth", "", h.appConfig.IsProduction(), true)
    c.JSON(http.StatusOK, gin.H{
        "message": "Successfully logged out",
    })
}

// ChangePassword handles password change
// @Summary Change password
// @Description Change user password
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.ChangePasswordRequest true "Password change data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
    var req dtos.ChangePasswordRequest

    // Get user from context
    var userID uuid.UUID
    userID, email, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }

    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }

    // Change password
    if err := h.authService.ChangePassword(userID, req); err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        // Log failed password change
        h.auditService.LogPasswordChange(c, userID, email, false)

        c.JSON(status, gin.H{
            "error":   "Password Change Failed",
            "code":    code,
            "message": message,
        })
        return
    }

    // Log successful password change
    h.auditService.LogPasswordChange(c, userID, email, true)

    c.JSON(http.StatusOK, gin.H{
        "message": "Password changed successfully",
    })
}

// ForgotPassword handles forgot password request
// @Summary Forgot password
// @Description Request password reset
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.ForgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
    var req dtos.ForgotPasswordRequest
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Generate reset token
    resetToken, err := h.authService.ForgotPassword(req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Password Reset Failed",
            "message": err.Error(),
        })
        return
    }
    
    // In production, send email with reset token
    // For development/testing, return the token
    if gin.Mode() == gin.DebugMode {
        c.JSON(http.StatusOK, gin.H{
            "message": "If an account exists with this email, a password reset link has been sent",
            "reset_token": resetToken, // Only in development
        })
    } else {
        c.JSON(http.StatusOK, gin.H{
            "message": "If an account exists with this email, a password reset link has been sent",
        })
    }
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset password using reset token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.ResetPasswordRequest true "Reset password data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
    var req dtos.ResetPasswordRequest
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Reset password
    if err := h.authService.ResetPassword(req); err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        c.JSON(status, gin.H{
            "error":   "Password Reset Failed",
            "code":    code,
            "message": message,
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Password reset successfully",
    })
}

// GetProfile gets user profile
// @Summary Get user profile
// @Description Get current user profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
    var userID uuid.UUID
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    // Get profile
    profile, err := h.authService.GetUserProfile(userID)
    if err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        c.JSON(status, gin.H{
            "error":   "Profile Retrieval Failed",
            "code":    code,
            "message": message,
        })
        return
    }

    c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates user profile
// @Summary Update user profile
// @Description Update current user profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Profile update data"
// @Success 200 {object} dtos.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
    var req struct {
        FullName string `json:"full_name" binding:"required,min=2"`
    }
    
    // Get user from context
    var userID uuid.UUID
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Update profile
    if err := h.authService.UpdateUserProfile(userID, req.FullName); err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        c.JSON(status, gin.H{
            "error":   "Profile Update Failed",
            "code":    code,
            "message": message,
        })
        return
    }

    // Get updated profile
    profile, err := h.authService.GetUserProfile(userID)
    if err != nil {
        status := apperr.GetHTTPStatus(err)
        code := apperr.GetErrorCode(err)
        message := apperr.GetErrorMessage(err)

        c.JSON(status, gin.H{
            "error":   "Profile Retrieval Failed",
            "code":    code,
            "message": message,
        })
        return
    }

    c.JSON(http.StatusOK, profile)
}
