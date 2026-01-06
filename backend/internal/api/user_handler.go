/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/user_handler.go
==============================================================================

DESCRIPTION:
    Handles user management endpoints for administrators. Allows creating,
    listing, deleting, and toggling active status of users within a company.

USER PERSPECTIVE:
    - Admin can view all users in their company
    - Admin can create new users with specific roles
    - Admin can deactivate/reactivate users
    - Admin can delete users (except themselves)

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add role assignment endpoints
    ⚠️  CAUTION: Role validation, self-modification prevention
    ❌  DO NOT modify: Admin-only authorization
    📝  Users can only manage users in their own company

SYNTAX EXPLANATION:
    - RequireRole("admin"): Middleware restricting to admin users
    - ToggleUserActive: Enables/disables user login
    - companyID from context: Ensures multi-tenancy isolation

ENDPOINTS (Admin Only):
    GET    /users - List all users in company
    POST   /users - Create new user
    DELETE /users/:id - Delete user
    PATCH  /users/:id/toggle-active - Enable/disable user

SECURITY:
    - All endpoints require admin role
    - Cannot delete or deactivate yourself
    - Cannot create admin users via this endpoint

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/services"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService *services.UserService
	authService *services.AuthService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, authService *services.AuthService) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
	}
}

// RegisterRoutes registers user management routes (admin only)
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	users := router.Group("/users")
	users.Use(authMiddleware.RequireAuth())
	users.Use(authMiddleware.RequireRole("admin"))
	{
		users.GET("", h.ListUsers)
		users.POST("", h.CreateUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.PATCH("/:id/toggle-active", h.ToggleUserActive)
	}
}

// ListUsers returns all users in the admin's company
func (h *UserHandler) ListUsers(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID // not used but kept for consistency

	users, err := h.userService.GetUsersByCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get users",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// CreateUser creates a new user in the admin's company
func (h *UserHandler) CreateUser(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(companyID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "a user with this email already exists" {
			status = http.StatusConflict
		} else if err.Error() == "invalid role" || err.Error() == "cannot create admin users through this endpoint" {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to create user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates a user's role and name
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": err.Error(),
		})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(userID, targetID, companyID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cannot change your own role" || err.Error() == "user not found" ||
		   err.Error() == "user not found in your company" || err.Error() == "invalid role" ||
		   err.Error() == "cannot change role to admin" {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to update user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user from the admin's company
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": err.Error(),
		})
		return
	}

	if err := h.userService.DeleteUser(userID, targetID, companyID); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cannot delete yourself" || err.Error() == "user not found" || err.Error() == "user not found in your company" {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to delete user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ToggleUserActive toggles a user's active status
func (h *UserHandler) ToggleUserActive(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": err.Error(),
		})
		return
	}

	user, err := h.userService.ToggleUserActive(userID, targetID, companyID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cannot deactivate yourself" || err.Error() == "user not found" || err.Error() == "user not found in your company" {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to toggle user status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
