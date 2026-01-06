/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/payroll_period_handler.go
==============================================================================

DESCRIPTION:
    Handles payroll period management endpoints: creating new periods,
    listing existing periods, and retrieving period details.

USER PERSPECTIVE:
    - Create weekly, biweekly, or monthly payroll periods
    - View all periods with filtering options
    - Get details of a specific period including totals

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add period update/close endpoints
    ⚠️  CAUTION: Period date validation logic
    ❌  DO NOT modify: Period code format validation
    📝  Periods are created per pay frequency type

SYNTAX EXPLANATION:
    - GetPeriods: Returns filtered list of periods
    - GetPeriod: Returns single period by ID
    - CreatePeriod: Creates new period with validation

ENDPOINTS:
    GET  /payroll/periods - List all periods
    GET  /payroll/periods/:id - Get period details
    POST /payroll/periods - Create new period

PERIOD TYPES:
    - weekly: For blue/gray collar workers
    - biweekly: For white collar workers
    - monthly: For special cases

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/dtos"
	"backend/internal/services"
)

type PayrollPeriodHandler struct {
	service *services.PayrollPeriodService
}

func NewPayrollPeriodHandler(service *services.PayrollPeriodService) *PayrollPeriodHandler {
	return &PayrollPeriodHandler{service: service}
}

func (h *PayrollPeriodHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/payroll/periods")
	group.GET("", h.GetPeriods)
	group.GET("/:id", h.GetPeriod)
	group.POST("", h.CreatePeriod)
	group.POST("/generate", h.GenerateCurrentPeriods)
}

func (h *PayrollPeriodHandler) GetPeriods(c *gin.Context) {
	filters := make(map[string]interface{}) // In a real app, parse from query params
	periods, err := h.service.GetPeriods(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payroll periods"})
		return
	}
	c.JSON(http.StatusOK, periods)
}

func (h *PayrollPeriodHandler) GetPeriod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}
	period, err := h.service.GetPeriod(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll period not found"})
		return
	}
	c.JSON(http.StatusOK, period)
}

func (h *PayrollPeriodHandler) CreatePeriod(c *gin.Context) {
	var req dtos.CreatePayrollPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	period, err := h.service.CreatePeriod(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll period"})
		return
	}

	c.JSON(http.StatusCreated, period)
}

// GenerateCurrentPeriods automatically generates the current payroll periods
// @Summary Generate current payroll periods
// @Description Automatically creates weekly, biweekly, and monthly periods for the current date
// @Tags Payroll Periods
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.PayrollPeriod
// @Failure 500 {object} map[string]string
// @Router /payroll/periods/generate [post]
func (h *PayrollPeriodHandler) GenerateCurrentPeriods(c *gin.Context) {
	periods, err := h.service.GenerateCurrentPeriods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate periods",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Periods generated successfully",
		"periods": periods,
		"count":   len(periods),
	})
}
