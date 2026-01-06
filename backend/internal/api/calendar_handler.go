/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/calendar_handler.go
==============================================================================

DESCRIPTION:
    HTTP handlers for the HR Calendar feature. Provides endpoints to aggregate
    calendar events from multiple sources and list employees with colors.

ENDPOINTS:
    GET /calendar/events     - Get aggregated calendar events
    GET /calendar/employees  - Get employees with assigned colors

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/dtos"
	"backend/internal/services"
)

// CalendarHandler handles HTTP requests for calendar
type CalendarHandler struct {
	service *services.CalendarService
}

// NewCalendarHandler creates a new handler
func NewCalendarHandler(service *services.CalendarService) *CalendarHandler {
	return &CalendarHandler{service: service}
}

// RegisterRoutes registers all calendar routes
func (h *CalendarHandler) RegisterRoutes(rg *gin.RouterGroup) {
	calendar := rg.Group("/calendar")
	{
		calendar.GET("/events", h.GetEvents)
		calendar.GET("/employees", h.GetEmployees)
	}
}

// GetEvents handles GET /calendar/events
// @Summary Get calendar events
// @Description Retrieves aggregated calendar events from absence requests, incidences, and shift exceptions
// @Tags Calendar
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param employee_ids[] query []string false "Filter by employee IDs"
// @Param collar_types[] query []string false "Filter by collar types (white_collar, blue_collar, gray_collar)"
// @Param event_types[] query []string false "Filter by event types (absence, incidence, shift_change)"
// @Param department_id query string false "Filter by department ID"
// @Param status query string false "Filter by status (pending, approved, declined)"
// @Success 200 {object} dtos.CalendarEventsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calendar/events [get]
func (h *CalendarHandler) GetEvents(c *gin.Context) {
	// Build request from query parameters
	req := dtos.CalendarEventRequest{
		StartDate:    c.Query("start_date"),
		EndDate:      c.Query("end_date"),
		EmployeeIDs:  c.QueryArray("employee_ids[]"),
		CollarTypes:  c.QueryArray("collar_types[]"),
		EventTypes:   c.QueryArray("event_types[]"),
		DepartmentID: c.Query("department_id"),
		Status:       c.Query("status"),
	}

	// Also support non-array format for single values
	if len(req.EmployeeIDs) == 0 {
		if id := c.Query("employee_ids"); id != "" {
			req.EmployeeIDs = []string{id}
		}
	}
	if len(req.CollarTypes) == 0 {
		if ct := c.Query("collar_types"); ct != "" {
			req.CollarTypes = []string{ct}
		}
	}
	if len(req.EventTypes) == 0 {
		if et := c.Query("event_types"); et != "" {
			req.EventTypes = []string{et}
		}
	}

	// Validate required fields
	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date are required",
		})
		return
	}

	// Get events from service
	response, err := h.service.GetCalendarEvents(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetEmployees handles GET /calendar/employees
// @Summary Get employees for calendar
// @Description Retrieves list of active employees with auto-assigned colors for calendar display
// @Tags Calendar
// @Accept json
// @Produce json
// @Param collar_types[] query []string false "Filter by collar types"
// @Param department_id query string false "Filter by department ID"
// @Success 200 {object} dtos.CalendarEmployeesResponse
// @Failure 500 {object} map[string]string
// @Router /calendar/employees [get]
func (h *CalendarHandler) GetEmployees(c *gin.Context) {
	collarTypes := c.QueryArray("collar_types[]")
	departmentID := c.Query("department_id")

	// Also support non-array format
	if len(collarTypes) == 0 {
		if ct := c.Query("collar_types"); ct != "" {
			collarTypes = []string{ct}
		}
	}

	response, err := h.service.GetEmployeesForCalendar(collarTypes, departmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
