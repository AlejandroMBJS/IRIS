/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/incidence_handler.go
==============================================================================

DESCRIPTION:
    Handles incidence (absence, overtime, vacation, etc.) endpoints for
    both incidence types (catalog) and individual incidence records.

USER PERSPECTIVE:
    - Manage incidence types (absence, vacation, overtime, bonus)
    - Create, update, and track employee incidences
    - Approve or reject pending incidences
    - View employee vacation balance and absence summary

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new incidence categories
    ⚠️  CAUTION: Approval workflow, status transitions
    ❌  DO NOT modify: Effect type values (positive/negative/neutral)
    📝  Incidences must be approved before payroll calculation

SYNTAX EXPLANATION:
    - IncidenceType: Catalog entry (e.g., "Vacaciones", "Falta")
    - Incidence: Individual record linked to employee and period
    - ApproveIncidence: Changes status from pending to approved
    - GetEmployeeVacationBalance: Calculates remaining vacation days

ENDPOINTS - Incidence Types (Catalog):
    GET    /incidence-types - List all types
    POST   /incidence-types - Create new type
    PUT    /incidence-types/:id - Update type
    DELETE /incidence-types/:id - Delete type

ENDPOINTS - Incidences:
    GET    /incidences - List with filters
    GET    /incidences/:id - Get details
    POST   /incidences - Create new
    PUT    /incidences/:id - Update
    DELETE /incidences/:id - Delete
    POST   /incidences/:id/approve - Approve
    POST   /incidences/:id/reject - Reject

ENDPOINTS - Employee Incidences:
    GET /employees/:id/incidences - Employee's incidences
    GET /employees/:id/vacation-balance - Vacation days remaining
    GET /employees/:id/absence-summary - Yearly absence summary

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

// IncidenceHandler handles incidence-related endpoints
type IncidenceHandler struct {
	incidenceService *services.IncidenceService
}

// NewIncidenceHandler creates a new incidence handler
func NewIncidenceHandler(incidenceService *services.IncidenceService) *IncidenceHandler {
	return &IncidenceHandler{incidenceService: incidenceService}
}

// RegisterRoutes registers incidence routes
func (h *IncidenceHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Incidence Types routes (catalogs)
	types := router.Group("/incidence-types")
	{
		types.GET("", h.ListIncidenceTypes)
		types.POST("", h.CreateIncidenceType)
		types.PUT("/:id", h.UpdateIncidenceType)
		types.DELETE("/:id", h.DeleteIncidenceType)
	}

	// Requestable incidence types (for employee portal)
	router.GET("/requestable-incidence-types", h.ListRequestableIncidenceTypes)

	// Incidences routes
	incidences := router.Group("/incidences")
	{
		incidences.GET("", h.ListIncidences)
		incidences.GET("/:id", h.GetIncidence)
		incidences.POST("", h.CreateIncidence)
		incidences.PUT("/:id", h.UpdateIncidence)
		incidences.DELETE("/:id", h.DeleteIncidence)
		incidences.POST("/:id/approve", h.ApproveIncidence)
		incidences.POST("/:id/reject", h.RejectIncidence)
	}

	// Employee-specific incidence routes
	employees := router.Group("/employees")
	{
		employees.GET("/:id/incidences", h.GetEmployeeIncidences)
		employees.GET("/:id/vacation-balance", h.GetEmployeeVacationBalance)
		employees.GET("/:id/absence-summary", h.GetEmployeeAbsenceSummary)
	}
}

// === Incidence Types ===

// ListIncidenceTypes returns all incidence types
func (h *IncidenceHandler) ListIncidenceTypes(c *gin.Context) {
	types, err := h.incidenceService.GetAllIncidenceTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get incidence types",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types)
}

// ListRequestableIncidenceTypes returns incidence types that employees can request
// @Summary List requestable incidence types
// @Tags Incidence Types
// @Produce json
// @Success 200 {array} models.IncidenceType
// @Router /requestable-incidence-types [get]
func (h *IncidenceHandler) ListRequestableIncidenceTypes(c *gin.Context) {
	types, err := h.incidenceService.GetRequestableIncidenceTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get requestable incidence types",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types)
}

// CreateIncidenceType creates a new incidence type
func (h *IncidenceHandler) CreateIncidenceType(c *gin.Context) {
	var req services.IncidenceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	incidenceType, err := h.incidenceService.CreateIncidenceType(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid category" || err.Error() == "invalid effect type" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to create incidence type",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, incidenceType)
}

// UpdateIncidenceType updates an incidence type
func (h *IncidenceHandler) UpdateIncidenceType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	var req services.IncidenceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	incidenceType, err := h.incidenceService.UpdateIncidenceType(id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "incidence type not found" {
			status = http.StatusNotFound
		} else if err.Error() == "invalid category" || err.Error() == "invalid effect type" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to update incidence type",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidenceType)
}

// DeleteIncidenceType deletes an incidence type
func (h *IncidenceHandler) DeleteIncidenceType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	if err := h.incidenceService.DeleteIncidenceType(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "incidence type not found" {
			status = http.StatusNotFound
		} else if err.Error() == "cannot delete incidence type with existing incidences" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   "Failed to delete incidence type",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incidence type deleted successfully"})
}

// === Incidences ===

// ListIncidences returns all incidences with optional filters
func (h *IncidenceHandler) ListIncidences(c *gin.Context) {
	employeeID := c.Query("employee_id")
	periodID := c.Query("period_id")
	status := c.Query("status")

	incidences, err := h.incidenceService.GetAllIncidences(employeeID, periodID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get incidences",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidences)
}

// GetIncidence returns a single incidence by ID
func (h *IncidenceHandler) GetIncidence(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	incidence, err := h.incidenceService.GetIncidenceByID(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "incidence not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Failed to get incidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidence)
}

// CreateIncidence creates a new incidence
func (h *IncidenceHandler) CreateIncidence(c *gin.Context) {
	var req services.CreateIncidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	incidence, err := h.incidenceService.CreateIncidence(req)
	if err != nil {
		status := http.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "employee not found" || errMsg == "incidence type not found" {
			status = http.StatusNotFound
		} else if errMsg == "invalid employee ID" || errMsg == "invalid incidence type ID" ||
			errMsg == "invalid start date format, use YYYY-MM-DD" ||
			errMsg == "invalid end date format, use YYYY-MM-DD" ||
			errMsg == "end date must be after start date" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to create incidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, incidence)
}

// UpdateIncidence updates an incidence
func (h *IncidenceHandler) UpdateIncidence(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	var req services.UpdateIncidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	incidence, err := h.incidenceService.UpdateIncidence(id, req)
	if err != nil {
		status := http.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "incidence not found" {
			status = http.StatusNotFound
		} else if errMsg == "cannot update a processed incidence" ||
			errMsg == "invalid start date format" ||
			errMsg == "invalid end date format" ||
			errMsg == "end date must be after start date" ||
			errMsg == "invalid status" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to update incidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidence)
}

// DeleteIncidence deletes an incidence
func (h *IncidenceHandler) DeleteIncidence(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	if err := h.incidenceService.DeleteIncidence(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "incidence not found" {
			status = http.StatusNotFound
		} else if err.Error() == "cannot delete a processed incidence" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to delete incidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incidence deleted successfully"})
}

// ApproveIncidence approves an incidence
func (h *IncidenceHandler) ApproveIncidence(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	incidence, err := h.incidenceService.ApproveIncidence(id, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "incidence not found" {
			status = http.StatusNotFound
		} else if err.Error() == "only pending incidences can be approved" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to approve incidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidence)
}

// RejectIncidence rejects an incidence
func (h *IncidenceHandler) RejectIncidence(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": err.Error(),
		})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	incidence, err := h.incidenceService.RejectIncidence(id, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "incidence not found" {
			status = http.StatusNotFound
		} else if err.Error() == "only pending incidences can be rejected" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   "Failed to reject incidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidence)
}

// === Employee Incidences ===

// GetEmployeeIncidences returns all incidences for an employee
func (h *IncidenceHandler) GetEmployeeIncidences(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid employee ID",
			"message": err.Error(),
		})
		return
	}

	incidences, err := h.incidenceService.GetIncidencesByEmployee(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get employee incidences",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incidences)
}

// GetEmployeeVacationBalance returns vacation balance for an employee
func (h *IncidenceHandler) GetEmployeeVacationBalance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid employee ID",
			"message": err.Error(),
		})
		return
	}

	balance, err := h.incidenceService.GetEmployeeVacationBalance(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "employee not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Failed to get vacation balance",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// GetEmployeeAbsenceSummary returns absence summary for an employee
func (h *IncidenceHandler) GetEmployeeAbsenceSummary(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid employee ID",
			"message": err.Error(),
		})
		return
	}

	// Default to current year
	yearStr := c.DefaultQuery("year", "2025")
	yearInt := 2025
	// Parse as integer
	n := 0
	for _, ch := range yearStr {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	if n > 0 {
		yearInt = n
	}

	summary, err := h.incidenceService.GetEmployeeAbsenceSummary(id, yearInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get absence summary",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}
