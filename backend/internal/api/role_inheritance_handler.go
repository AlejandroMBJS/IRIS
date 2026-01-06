/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/role_inheritance_handler.go
==============================================================================

DESCRIPTION:
    Handles role inheritance configuration endpoints.
    Provides CRUD operations for managing role inheritance relationships.

USER PERSPECTIVE:
    - Configure which roles inherit from others
    - View inheritance hierarchy
    - See resolved permissions for each role

DEVELOPER GUIDELINES:
    ✅  OK to modify: Response formats, validation rules
    ⚠️  CAUTION: Authorization (admin only)
    ❌  DO NOT modify: Circular dependency detection logic
    📝  All endpoints require admin authentication

ENDPOINTS:
    GET    /role-inheritance - List all inheritances
    GET    /role-inheritance/hierarchy - Get hierarchy tree
    GET    /role-inheritance/resolve/:role - Resolve inherited roles
    POST   /role-inheritance - Create inheritance
    PUT    /role-inheritance/:id - Update inheritance
    DELETE /role-inheritance/:id - Delete inheritance

==============================================================================
*/
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
)

// RoleInheritanceHandler handles role inheritance endpoints
type RoleInheritanceHandler struct {
	service *services.RoleInheritanceService
}

// NewRoleInheritanceHandler creates new role inheritance handler
func NewRoleInheritanceHandler(service *services.RoleInheritanceService) *RoleInheritanceHandler {
	return &RoleInheritanceHandler{service: service}
}

// RegisterRoutes registers role inheritance routes (admin only)
func (h *RoleInheritanceHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	inheritance := router.Group("/role-inheritance")
	// Enforce admin-only access for all role inheritance operations
	inheritance.Use(authMiddleware.RequireRole("admin"))
	{
		inheritance.GET("", h.GetAllInheritances)
		inheritance.GET("/hierarchy", h.GetHierarchy)
		inheritance.GET("/resolve/:role", h.ResolveRoles)
		inheritance.GET("/valid-roles", h.GetValidRoles)
		inheritance.POST("", h.CreateInheritance)
		inheritance.PUT("/:id", h.UpdateInheritance)
		inheritance.DELETE("/:id", h.DeleteInheritance)
	}
}

// GetAllInheritances handles fetching all role inheritances
// @Summary List all role inheritances
// @Description Returns all role inheritance relationships
// @Tags Role Inheritance
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.RoleInheritance
// @Failure 500 {object} map[string]string
// @Router /role-inheritance [get]
func (h *RoleInheritanceHandler) GetAllInheritances(c *gin.Context) {
	inheritances, err := h.service.GetAllInheritances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inheritances", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inheritances)
}

// GetHierarchy handles fetching the role inheritance hierarchy
// @Summary Get role hierarchy
// @Description Returns a hierarchical map of role inheritances
// @Tags Role Inheritance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]string
// @Failure 500 {object} map[string]string
// @Router /role-inheritance/hierarchy [get]
func (h *RoleInheritanceHandler) GetHierarchy(c *gin.Context) {
	hierarchy, err := h.service.GetInheritanceHierarchy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get hierarchy", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

// ResolveRoles handles resolving all roles (direct + inherited) for a given role
// @Summary Resolve inherited roles
// @Description Returns all roles a user with the given role effectively has
// @Tags Role Inheritance
// @Produce json
// @Security BearerAuth
// @Param role path string true "Role name"
// @Success 200 {object} map[string][]string
// @Failure 400 {object} map[string]string "Invalid role"
// @Failure 500 {object} map[string]string
// @Router /role-inheritance/resolve/{role} [get]
func (h *RoleInheritanceHandler) ResolveRoles(c *gin.Context) {
	role := c.Param("role")

	if !models.IsValidRole(role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role name"})
		return
	}

	roles, err := h.service.ResolveInheritedRoles(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve roles", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role":           role,
		"inherited_roles": roles,
	})
}

// GetValidRoles handles fetching the list of valid role names
// @Summary Get valid roles
// @Description Returns all valid role names in the system
// @Tags Role Inheritance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /role-inheritance/valid-roles [get]
func (h *RoleInheritanceHandler) GetValidRoles(c *gin.Context) {
	roles := models.ValidRoles()

	// Add permission levels for UI sorting/display
	rolesWithLevels := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		rolesWithLevels[i] = map[string]interface{}{
			"name":  role,
			"level": models.RolePermissionLevel(role),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": rolesWithLevels,
	})
}

// CreateInheritanceRequest represents the request body for creating an inheritance
type CreateInheritanceRequest struct {
	ChildRole  string `json:"child_role" binding:"required"`
	ParentRole string `json:"parent_role" binding:"required"`
	Priority   int    `json:"priority"`
	Notes      string `json:"notes"`
}

// CreateInheritance handles creating a new role inheritance
// @Summary Create role inheritance
// @Description Creates a new role inheritance relationship
// @Tags Role Inheritance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateInheritanceRequest true "Inheritance data"
// @Success 201 {object} models.RoleInheritance
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 409 {object} map[string]string "Circular dependency or duplicate"
// @Failure 500 {object} map[string]string
// @Router /role-inheritance [post]
func (h *RoleInheritanceHandler) CreateInheritance(c *gin.Context) {
	var req CreateInheritanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "message": err.Error()})
		return
	}

	inheritance := &models.RoleInheritance{
		ChildRole:  req.ChildRole,
		ParentRole: req.ParentRole,
		Priority:   req.Priority,
		IsActive:   true,
		Notes:      req.Notes,
	}

	err := h.service.CreateInheritance(inheritance)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid") ||
		   strings.Contains(err.Error(), "cannot inherit from itself") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "already exists") ||
		          strings.Contains(err.Error(), "circular dependency") {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": "Failed to create inheritance", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inheritance)
}

// UpdateInheritanceRequest represents the request body for updating an inheritance
type UpdateInheritanceRequest struct {
	IsActive bool   `json:"is_active"`
	Priority int    `json:"priority"`
	Notes    string `json:"notes"`
}

// UpdateInheritance handles updating an existing role inheritance
// @Summary Update role inheritance
// @Description Updates an existing role inheritance relationship
// @Tags Role Inheritance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Inheritance ID (UUID)"
// @Param request body UpdateInheritanceRequest true "Updated data"
// @Success 200 {object} models.RoleInheritance
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Inheritance not found"
// @Failure 500 {object} map[string]string
// @Router /role-inheritance/{id} [put]
func (h *RoleInheritanceHandler) UpdateInheritance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inheritance ID"})
		return
	}

	var req UpdateInheritanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "message": err.Error()})
		return
	}

	updates := &models.RoleInheritance{
		IsActive: req.IsActive,
		Priority: req.Priority,
		Notes:    req.Notes,
	}

	err = h.service.UpdateInheritance(id, updates)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to update inheritance", "message": err.Error()})
		return
	}

	// Get updated inheritance
	inheritances, _ := h.service.GetAllInheritances()
	for _, inheritance := range inheritances {
		if inheritance.ID == id {
			c.JSON(http.StatusOK, inheritance)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inheritance updated successfully"})
}

// DeleteInheritance handles deleting a role inheritance
// @Summary Delete role inheritance
// @Description Deletes a role inheritance relationship
// @Tags Role Inheritance
// @Produce json
// @Security BearerAuth
// @Param id path string true "Inheritance ID (UUID)"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Inheritance not found"
// @Failure 500 {object} map[string]string
// @Router /role-inheritance/{id} [delete]
func (h *RoleInheritanceHandler) DeleteInheritance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inheritance ID"})
		return
	}

	err = h.service.DeleteInheritance(id)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to delete inheritance", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inheritance deleted successfully"})
}
