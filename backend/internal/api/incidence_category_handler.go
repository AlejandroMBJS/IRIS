/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/incidence_category_handler.go
==============================================================================

DESCRIPTION:
    Handles incidence category CRUD endpoints. Categories are the top-level
    grouping for incidence types.

USER PERSPECTIVE:
    - List all categories with their incidence types
    - Create new categories for organizing incidence types
    - Update category details (name, color, icon, etc.)
    - Delete categories (only if empty and not system-protected)

DEVELOPER GUIDELINES:
    OK to modify: Add new category operations
    CAUTION: System categories cannot be deleted
    DO NOT modify: Default category codes

ENDPOINTS:
    GET    /incidence-categories - List all categories
    GET    /incidence-categories/:id - Get single category with types
    POST   /incidence-categories - Create new category
    PUT    /incidence-categories/:id - Update category
    DELETE /incidence-categories/:id - Delete category
    PATCH  /incidence-categories/:id/reorder - Update display order

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/services"
)

// IncidenceCategoryHandler handles category-related endpoints
type IncidenceCategoryHandler struct {
	categoryService *services.IncidenceCategoryService
}

// NewIncidenceCategoryHandler creates a new category handler
func NewIncidenceCategoryHandler(categoryService *services.IncidenceCategoryService) *IncidenceCategoryHandler {
	return &IncidenceCategoryHandler{categoryService: categoryService}
}

// RegisterRoutes registers category routes
func (h *IncidenceCategoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	categories := router.Group("/incidence-categories")
	{
		categories.GET("", h.ListCategories)
		categories.GET("/:id", h.GetCategory)
		categories.POST("", h.CreateCategory)
		categories.PUT("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
		categories.PATCH("/:id/reorder", h.UpdateDisplayOrder)
	}
}

// ListCategories returns all categories
// @Summary List all incidence categories
// @Tags Incidence Categories
// @Produce json
// @Success 200 {array} models.IncidenceCategory
// @Router /incidence-categories [get]
func (h *IncidenceCategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategory returns a single category by ID
// @Summary Get incidence category by ID
// @Tags Incidence Categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.IncidenceCategory
// @Failure 404 {object} map[string]string
// @Router /incidence-categories/{id} [get]
func (h *IncidenceCategoryHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid category ID",
		})
		return
	}

	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "category not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory creates a new category
// @Summary Create incidence category
// @Tags Incidence Categories
// @Accept json
// @Produce json
// @Param request body services.CategoryRequest true "Category data"
// @Success 201 {object} models.IncidenceCategory
// @Failure 400 {object} map[string]string
// @Router /incidence-categories [post]
func (h *IncidenceCategoryHandler) CreateCategory(c *gin.Context) {
	var req services.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	// Default is_active to true if not specified
	if !req.IsActive && c.Request.ContentLength > 0 {
		// Check if is_active was explicitly set to false or just missing
		var raw map[string]interface{}
		if err := c.ShouldBindJSON(&raw); err == nil {
			if _, exists := raw["is_active"]; !exists {
				req.IsActive = true
			}
		} else {
			req.IsActive = true
		}
	}

	category, err := h.categoryService.CreateCategory(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "category with this code already exists" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   "Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory updates an existing category
// @Summary Update incidence category
// @Tags Incidence Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body services.CategoryRequest true "Category data"
// @Success 200 {object} models.IncidenceCategory
// @Failure 400,404 {object} map[string]string
// @Router /incidence-categories/{id} [put]
func (h *IncidenceCategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid category ID",
		})
		return
	}

	var req services.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	category, err := h.categoryService.UpdateCategory(id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "category not found" {
			status = http.StatusNotFound
		} else if err.Error() == "category with this code already exists" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   "Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory deletes a category
// @Summary Delete incidence category
// @Tags Incidence Categories
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]string
// @Failure 400,404,409 {object} map[string]string
// @Router /incidence-categories/{id} [delete]
func (h *IncidenceCategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid category ID",
		})
		return
	}

	err = h.categoryService.DeleteCategory(id)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "category not found":
			status = http.StatusNotFound
		case "cannot delete system category":
			status = http.StatusForbidden
		case "cannot delete category with existing incidence types":
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   "Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "category deleted successfully",
	})
}

// UpdateDisplayOrder updates the display order of a category
// @Summary Update category display order
// @Tags Incidence Categories
// @Accept json
// @Param id path string true "Category ID"
// @Param request body map[string]int true "Display order"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /incidence-categories/{id}/reorder [patch]
func (h *IncidenceCategoryHandler) UpdateDisplayOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid category ID",
		})
		return
	}

	var req struct {
		DisplayOrder int `json:"display_order" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	err = h.categoryService.UpdateDisplayOrder(id, req.DisplayOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "display order updated successfully",
	})
}
