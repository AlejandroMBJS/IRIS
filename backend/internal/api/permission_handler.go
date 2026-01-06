/*
Package api - Permission Matrix API Handler

==============================================================================
FILE: internal/api/permission_handler.go
==============================================================================

DESCRIPTION:
    HTTP handlers for the permission matrix configuration API. Allows admin
    to view and configure role-based permissions.

USER PERSPECTIVE:
    - Admin views all permissions in the permission matrix
    - Admin updates permissions for specific roles and resources
    - Changes take effect immediately across the system

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new endpoints for permission analytics
    ⚠️  CAUTION: Permission changes affect all users with that role
    ❌  DO NOT modify: Only admin should access these endpoints
    📝  All permission operations are restricted to admin role only

ENDPOINTS:
    GET    /permissions - List all permissions
    GET    /permissions/role/:role - Get permissions for a role
    GET    /permissions/:id - Get specific permission
    POST   /permissions - Create new permission
    PUT    /permissions/:id - Update permission
    DELETE /permissions/:id - Delete permission
    GET    /permissions/roles - List all valid roles
    GET    /permissions/resources - List all valid resources

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

// PermissionHandler handles permission matrix endpoints
type PermissionHandler struct {
	permissionService *services.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// RegisterRoutes registers permission routes (admin only)
func (h *PermissionHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	permissions := router.Group("/permissions")
	// Enforce admin-only access for all permission matrix operations
	permissions.Use(authMiddleware.RequireRole("admin"))
	{
		permissions.GET("", h.ListPermissions)
		permissions.GET("/role/:role", h.GetPermissionsByRole)
		permissions.GET("/roles", h.ListRoles)
		permissions.GET("/resources", h.ListResources)
		permissions.GET("/:id", h.GetPermission)
		permissions.POST("", h.CreatePermission)
		permissions.PUT("/:id", h.UpdatePermission)
		permissions.DELETE("/:id", h.DeletePermission)
	}
}

// ListPermissions handles listing all permissions
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.permissionService.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GetPermissionsByRole handles getting permissions for a specific role
func (h *PermissionHandler) GetPermissionsByRole(c *gin.Context) {
	role := c.Param("role")

	permissions, err := h.permissionService.GetPermissionsByRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GetPermission handles getting a specific permission
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Invalid permission ID format",
		})
		return
	}

	permissions, err := h.permissionService.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	// Find the permission with matching ID
	for _, perm := range permissions {
		if perm.ID.String() == id {
			c.JSON(http.StatusOK, perm)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error":   "Not Found",
		"message": "Permission not found",
	})
}

// CreatePermission handles creating a new permission
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var permission models.Permission

	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	if err := h.permissionService.CreatePermission(&permission); err != nil {
		status := http.StatusInternalServerError
		// Check if it's a duplicate permission error
		if strings.Contains(err.Error(), "permission already exists") {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   "Permission Creation Failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// UpdatePermission handles updating a permission
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Invalid permission ID format",
		})
		return
	}

	var updates models.Permission
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	permission, err := h.permissionService.UpdatePermission(id, &updates)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "permission not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Permission Update Failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// DeletePermission handles deleting a permission
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Invalid permission ID format",
		})
		return
	}

	if err := h.permissionService.DeletePermission(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "permission not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Permission Deletion Failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission deleted successfully",
	})
}

// ListRoles handles listing all valid roles
func (h *PermissionHandler) ListRoles(c *gin.Context) {
	roles := models.ValidRoles()
	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// ListResources handles listing all valid resources
func (h *PermissionHandler) ListResources(c *gin.Context) {
	resources := models.ValidResources()
	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
	})
}
