/*
Package middleware - Authentication and authorization middleware for HTTP requests

==============================================================================
FILE: internal/middleware/auth.go
==============================================================================

DESCRIPTION:
    This file provides JWT-based authentication middleware for the Gin web framework.
    It validates JWT tokens, extracts user information, and enforces role-based access
    control (RBAC) for protected routes. User context is stored in the request context
    for downstream handlers to access.

USER PERSPECTIVE:
    - Users must include a valid JWT token in the Authorization header to access
      protected endpoints
    - Invalid or expired tokens result in 401 Unauthorized responses
    - Insufficient permissions (wrong role) result in 403 Forbidden responses
    - Successful authentication allows seamless access to authorized resources

DEVELOPER GUIDELINES:
    ✅  OK to modify: Error messages, HTTP status codes for specific scenarios
    ✅  OK to modify: Add new context values if additional user data is needed
    ⚠️  CAUTION: Changing token extraction logic - ensure backwards compatibility
    ⚠️  CAUTION: Modifying RequireRole logic - test all role combinations thoroughly
    ❌  DO NOT modify: The order of middleware execution (RequireAuth must run before RequireRole)
    ❌  DO NOT modify: Context key names without updating all consumers
    📝  Always call c.Abort() when rejecting requests to prevent downstream handlers from running
    📝  Use c.Next() to pass control to the next middleware/handler in the chain

SYNTAX EXPLANATION:
    - gin.HandlerFunc: A Gin middleware function signature that wraps HTTP handlers
    - c.GetHeader("Authorization"): Extracts the Authorization header from the request
    - c.Set("key", value): Stores values in the Gin context for the request lifecycle
    - c.Get("key"): Retrieves values from the Gin context (returns interface{} + exists bool)
    - c.Abort(): Stops the middleware chain - no subsequent handlers will execute
    - c.Next(): Continues to the next middleware/handler in the chain
    - Type assertion: userID.(uuid.UUID) converts interface{} to concrete type
    - Variadic parameter: RequireRole(roles ...string) accepts multiple role arguments

MIDDLEWARE PATTERN:
    1. RequireAuth() validates JWT and stores user data in context
    2. RequireRole() checks if authenticated user has required permissions
    3. Route handlers use GetUserFromContext() to retrieve authenticated user info

EXAMPLE USAGE:
    authMiddleware := NewAuthMiddleware(authService)

    // Require authentication only
    router.GET("/profile", authMiddleware.RequireAuth(), profileHandler)

    // Require authentication + admin role
    router.GET("/admin/users",
        authMiddleware.RequireAuth(),
        authMiddleware.RequireRole("admin"),
        adminHandler)

    // Require authentication + one of multiple roles
    router.POST("/payroll",
        authMiddleware.RequireAuth(),
        authMiddleware.RequireRole("admin", "manager"),
        payrollHandler)

==============================================================================
*/

package middleware

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "backend/internal/services"
    "backend/internal/utils"
)

// AuthMiddleware authenticates requests using JWT tokens
type AuthMiddleware struct {
    authService *services.AuthService
}

// NewAuthMiddleware creates new authentication middleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
    return &AuthMiddleware{authService: authService}
}

// RoleMiddleware enforces role-based access control
type RoleMiddleware struct {
    allowedRoles []string
}

// NewRoleMiddleware creates new role middleware for specified roles
func NewRoleMiddleware(roles ...string) *RoleMiddleware {
    return &RoleMiddleware{allowedRoles: roles}
}

// RequireRole checks if the authenticated user has one of the allowed roles
func (m *RoleMiddleware) RequireRole() gin.HandlerFunc {
    return func(c *gin.Context) {
        roleStr := GetUserRoleFromContext(c)
        if roleStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "Unauthorized",
                "message": "User role not found",
            })
            c.Abort()
            return
        }

        hasRole := false
        for _, allowedRole := range m.allowedRoles {
            if roleStr == allowedRole {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{
                "error":   "Forbidden",
                "message": "Insufficient permissions",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireAuth requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from header
        authHeader := c.GetHeader("Authorization")
        token, err := utils.ExtractTokenFromHeader(authHeader)
				if err != nil {
					token, err = c.Cookie("access_token")
				}
				if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "Unauthorized",
                "message": err.Error(),
            })
            c.Abort()
            return
        }
        
        // Verify token and get user
        user, err := m.authService.VerifyToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "Unauthorized",
                "message": "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        // Store user in context
        c.Set("userID", user.ID)
        c.Set("userEmail", user.Email)
        c.Set("userRole", user.Role)
        c.Set("companyID", user.CompanyID)
        c.Set("user", user)

        c.Next()
    }
}

// RequireRole requires specific role
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user role from context and convert to string
        roleStr := GetUserRoleFromContext(c)
        if roleStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "Unauthorized",
                "message": "User role not found",
            })
            c.Abort()
            return
        }

        // Check if user has required role
        hasRole := false
        for _, requiredRole := range roles {
            if roleStr == requiredRole {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{
                "error":   "Forbidden",
                "message": "Insufficient permissions",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// GetUserFromContext extracts user from context
// Returns: userID, userEmail, companyID, error
func GetUserFromContext(c *gin.Context) (uuid.UUID, string, uuid.UUID, error) {
    userID, exists := c.Get("userID")
    if !exists {
        return uuid.Nil, "", uuid.Nil, http.ErrNoLocation
    }

    userEmail, _ := c.Get("userEmail")
    companyID, _ := c.Get("companyID")

    return userID.(uuid.UUID), userEmail.(string), companyID.(uuid.UUID), nil
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(c *gin.Context) string {
    userRole, exists := c.Get("userRole")
    if !exists {
        return ""
    }

    // Handle both enums.UserRole and string types for userRole
    switch r := userRole.(type) {
    case string:
        return r
    default:
        // For enums.UserRole or any other type with String() method
        return fmt.Sprintf("%v", r)
    }
}

// ==============================================================================
// Collar Type Filtering Utilities for Hierarchical HR Roles
// NOTE: Core implementations are in backend/internal/utils/collar.go
// These wrappers are kept for backward compatibility with handlers.
// ==============================================================================

// GetAllowedCollarTypes returns the collar types a role is allowed to access.
// Delegates to utils.GetAllowedCollarTypes
func GetAllowedCollarTypes(role string) []string {
    return utils.GetAllowedCollarTypes(role)
}

// IsHRRole checks if a role is any HR role (including sub-HR roles)
// Delegates to utils.IsHRRole
func IsHRRole(role string) bool {
    return utils.IsHRRole(role)
}

// CanManageCollarType checks if a role can manage employees of a specific collar type
// Delegates to utils.CanManageCollarType
func CanManageCollarType(role string, collarType string) bool {
    return utils.CanManageCollarType(role, collarType)
}

// IntersectCollarTypes returns the intersection of two collar type slices.
// Delegates to utils.IntersectCollarTypes
func IntersectCollarTypes(a, b []string) []string {
    return utils.IntersectCollarTypes(a, b)
}
