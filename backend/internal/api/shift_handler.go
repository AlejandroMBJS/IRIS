/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/shift_handler.go
==============================================================================

DESCRIPTION:
    Handles shift management CRUD endpoints. Shifts define work schedules
    for employees including start/end times, breaks, and work days.

USER PERSPECTIVE:
    - List all shifts configured for the company
    - Create new shifts (e.g., Turno Matutino, Vespertino, Nocturno)
    - Update shift details (times, breaks, work days)
    - Delete shifts (only if no employees assigned)
    - Toggle shift active status

DEVELOPER GUIDELINES:
    OK to modify: Add new shift operations
    CAUTION: Validate company_id for multi-tenant security
    DO NOT modify: Shift deletion validation (employees check)

ENDPOINTS:
    GET    /shifts - List all shifts for company
    GET    /shifts/active - List only active shifts
    GET    /shifts/:id - Get single shift
    POST   /shifts - Create new shift
    PUT    /shifts/:id - Update shift
    DELETE /shifts/:id - Delete shift
    PATCH  /shifts/:id/toggle - Toggle active status

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

// ShiftHandler handles shift-related endpoints
type ShiftHandler struct {
	shiftService *services.ShiftService
}

// NewShiftHandler creates a new shift handler
func NewShiftHandler(shiftService *services.ShiftService) *ShiftHandler {
	return &ShiftHandler{
		shiftService: shiftService,
	}
}

// RegisterRoutes registers shift routes
func (h *ShiftHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	shifts := router.Group("/shifts")
	shifts.Use(authMiddleware.RequireAuth())
	{
		shifts.GET("", h.ListShifts)
		shifts.GET("/active", h.ListActiveShifts)
		shifts.GET("/:id", h.GetShift)

		// Admin/HR only routes
		admin := shifts.Group("")
		admin.Use(authMiddleware.RequireRole("admin", "hr", "hr_and_pr", "hr_blue_gray", "hr_white"))
		{
			admin.POST("", h.CreateShift)
			admin.PUT("/:id", h.UpdateShift)
			admin.DELETE("/:id", h.DeleteShift)
			admin.PATCH("/:id/toggle", h.ToggleShiftActive)
		}
	}
}

// getCompanyIDAndRole extracts the company ID and role from the authenticated user
func (h *ShiftHandler) getCompanyIDAndRole(c *gin.Context) (uuid.UUID, string, error) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		return uuid.Nil, "", err
	}
	role := middleware.GetUserRoleFromContext(c)
	return companyID, role, nil
}

// getCompanyID extracts the company ID from the authenticated user (legacy helper)
func (h *ShiftHandler) getCompanyID(c *gin.Context) (uuid.UUID, error) {
	companyID, _, err := h.getCompanyIDAndRole(c)
	return companyID, err
}

// ListShifts returns all shifts for the company, filtered by user's collar type access
// @Summary List all shifts
// @Tags Shifts
// @Produce json
// @Success 200 {array} services.ShiftResponse
// @Router /shifts [get]
func (h *ShiftHandler) ListShifts(c *gin.Context) {
	companyID, role, err := h.getCompanyIDAndRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	shifts, err := h.shiftService.GetAllShifts(companyID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shifts)
}

// ListActiveShifts returns only active shifts for the company, filtered by user's collar type access
// @Summary List active shifts
// @Tags Shifts
// @Produce json
// @Success 200 {array} models.Shift
// @Router /shifts/active [get]
func (h *ShiftHandler) ListActiveShifts(c *gin.Context) {
	companyID, role, err := h.getCompanyIDAndRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	shifts, err := h.shiftService.GetActiveShifts(companyID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shifts)
}

// GetShift returns a single shift by ID
// @Summary Get shift by ID
// @Tags Shifts
// @Produce json
// @Param id path string true "Shift ID"
// @Success 200 {object} models.Shift
// @Failure 404 {object} map[string]string
// @Router /shifts/{id} [get]
func (h *ShiftHandler) GetShift(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid shift ID",
		})
		return
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	shift, err := h.shiftService.GetShiftByID(id, companyID)
	if err != nil {
		if err.Error() == "shift not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "shift not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shift)
}

// CreateShift creates a new shift
// @Summary Create shift
// @Tags Shifts
// @Accept json
// @Produce json
// @Param request body services.ShiftRequest true "Shift data"
// @Success 201 {object} models.Shift
// @Failure 400 {object} map[string]string
// @Router /shifts [post]
func (h *ShiftHandler) CreateShift(c *gin.Context) {
	var req services.ShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	shift, err := h.shiftService.CreateShift(req, companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, shift)
}

// UpdateShift updates an existing shift
// @Summary Update shift
// @Tags Shifts
// @Accept json
// @Produce json
// @Param id path string true "Shift ID"
// @Param request body services.ShiftRequest true "Shift data"
// @Success 200 {object} models.Shift
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /shifts/{id} [put]
func (h *ShiftHandler) UpdateShift(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid shift ID",
		})
		return
	}

	var req services.ShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	shift, err := h.shiftService.UpdateShift(id, req, companyID)
	if err != nil {
		if err.Error() == "shift not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "shift not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shift)
}

// DeleteShift deletes a shift
// @Summary Delete shift
// @Tags Shifts
// @Param id path string true "Shift ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /shifts/{id} [delete]
func (h *ShiftHandler) DeleteShift(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid shift ID",
		})
		return
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	err = h.shiftService.DeleteShift(id, companyID)
	if err != nil {
		if err.Error() == "shift not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "shift not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "shift deleted successfully",
	})
}

// ToggleShiftActive toggles the active status of a shift
// @Summary Toggle shift active status
// @Tags Shifts
// @Param id path string true "Shift ID"
// @Success 200 {object} models.Shift
// @Failure 404 {object} map[string]string
// @Router /shifts/{id}/toggle [patch]
func (h *ShiftHandler) ToggleShiftActive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "invalid shift ID",
		})
		return
	}

	companyID, err := h.getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "could not determine company",
		})
		return
	}

	shift, err := h.shiftService.ToggleShiftActive(id, companyID)
	if err != nil {
		if err.Error() == "shift not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "shift not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shift)
}
