/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/absence_request_handler.go
==============================================================================

DESCRIPTION:
    HTTP handlers for absence request approval workflow.
    Handles creation, approval, and management of time-off requests.

ENDPOINTS:
    POST   /absence-requests              - Create new request
    GET    /absence-requests/my-requests  - Get user's own requests
    GET    /absence-requests/pending/:stage - Get pending by stage
    POST   /absence-requests/:id/approve  - Approve/decline request
    DELETE /absence-requests/:id          - Delete request
    PATCH  /absence-requests/:id/archive  - Archive request
    GET    /absence-requests/overlapping  - Check overlapping absences
    GET    /absence-requests/counts       - Get pending counts

==============================================================================
*/
package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/models/enums"
	"backend/internal/services"
)

// AbsenceRequestHandler handles HTTP requests for absence requests
type AbsenceRequestHandler struct {
	service *services.AbsenceRequestService
}

// NewAbsenceRequestHandler creates a new handler
func NewAbsenceRequestHandler(service *services.AbsenceRequestService) *AbsenceRequestHandler {
	return &AbsenceRequestHandler{service: service}
}

// RegisterRoutes registers all absence request routes
func (h *AbsenceRequestHandler) RegisterRoutes(rg *gin.RouterGroup) {
	requests := rg.Group("/absence-requests")
	{
		requests.POST("", h.Create)
		requests.GET("/my-requests", h.GetMyRequests)
		requests.GET("/pending/:stage", h.GetPendingByStage)
		requests.POST("/:id/approve", h.Approve)
		requests.DELETE("/:id", h.Delete)
		requests.PATCH("/:id/archive", h.Archive)
		requests.GET("/overlapping", h.GetOverlapping)
		requests.GET("/counts", h.GetCounts)
		requests.GET("/approved", h.GetApproved)
		requests.GET("/export", h.ExportApproved)
	}
}

// CreateAbsenceRequestDTO is the request body for creating an absence request
type CreateAbsenceRequestDTO struct {
	EmployeeID     string   `json:"employee_id" binding:"required"`
	RequestType    string   `json:"request_type" binding:"required"`
	StartDate      string   `json:"start_date" binding:"required"`
	EndDate        string   `json:"end_date" binding:"required"`
	TotalDays      float64  `json:"total_days" binding:"required"`
	Reason         string   `json:"reason" binding:"required"`
	HoursPerDay    *float64 `json:"hours_per_day"`
	PaidDays       *float64 `json:"paid_days"`
	UnpaidDays     *float64 `json:"unpaid_days"`
	UnpaidComments string   `json:"unpaid_comments"`
	ShiftDetails   string   `json:"shift_details"`
	NewShiftID     *string  `json:"new_shift_id"` // For SHIFT_CHANGE requests - the target shift
}

// Create handles POST /absence-requests
func (h *AbsenceRequestHandler) Create(c *gin.Context) {
	var dto CreateAbsenceRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, err := uuid.Parse(dto.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee_id"})
		return
	}

	startDate, err := time.Parse("2006-01-02", dto.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", dto.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// Parse new_shift_id if provided
	var newShiftID *uuid.UUID
	if dto.NewShiftID != nil && *dto.NewShiftID != "" {
		shiftID, err := uuid.Parse(*dto.NewShiftID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid new_shift_id"})
			return
		}
		newShiftID = &shiftID
	}

	input := services.CreateAbsenceRequestInput{
		EmployeeID:     employeeID,
		RequestType:    models.RequestType(dto.RequestType),
		StartDate:      startDate,
		EndDate:        endDate,
		TotalDays:      dto.TotalDays,
		Reason:         dto.Reason,
		HoursPerDay:    dto.HoursPerDay,
		PaidDays:       dto.PaidDays,
		UnpaidDays:     dto.UnpaidDays,
		UnpaidComments: dto.UnpaidComments,
		ShiftDetails:   dto.ShiftDetails,
		NewShiftID:     newShiftID,
	}

	result, err := h.service.CreateAbsenceRequest(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"requestId":   result.Request.ID,
		"incidenceId": result.IncidenceID, // NEW: Return incidence ID for evidence upload
	})
}

// GetMyRequests handles GET /absence-requests/my-requests
func (h *AbsenceRequestHandler) GetMyRequests(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	requests, err := h.service.GetMyRequests(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// GetPendingByStage handles GET /absence-requests/pending/:stage
func (h *AbsenceRequestHandler) GetPendingByStage(c *gin.Context) {
	stage := c.Param("stage")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var requests []models.AbsenceRequest
	var err error

	switch stage {
	case "supervisor", "SUPERVISOR":
		requests, err = h.service.GetPendingRequestsForSupervisor(userID.(uuid.UUID))
	case "manager", "MANAGER", "general_manager", "GENERAL_MANAGER":
		requests, err = h.service.GetPendingRequestsForGeneralManager()
	case "hr", "HR":
		requests, err = h.service.GetPendingRequestsForHR(userID.(uuid.UUID))
	case "hr_blue_gray", "HR_BLUE_GRAY":
		requests, err = h.service.GetPendingRequestsForHRBlueGray(userID.(uuid.UUID))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stage"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// ApproveRequestDTO is the request body for approving/declining
type ApproveRequestDTO struct {
	Action   string `json:"action" binding:"required"` // APPROVED or DECLINED
	Stage    string `json:"stage" binding:"required"`
	Comments string `json:"comments"`
}

// Approve handles POST /absence-requests/:id/approve
func (h *AbsenceRequestHandler) Approve(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var dto ApproveRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.ApproveRequestInput{
		RequestID:  requestID,
		ApproverID: userID.(uuid.UUID),
		Stage:      models.ApprovalStage(dto.Stage),
		Action:     models.ApprovalAction(dto.Action),
		Comments:   dto.Comments,
	}

	if err := h.service.ApproveRequest(input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete handles DELETE /absence-requests/:id
func (h *AbsenceRequestHandler) Delete(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	if err := h.service.DeleteRequest(requestID, userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Request deleted successfully"})
}

// Archive handles PATCH /absence-requests/:id/archive
func (h *AbsenceRequestHandler) Archive(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	if err := h.service.ArchiveRequest(requestID, userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Request archived successfully"})
}

// GetOverlapping handles GET /absence-requests/overlapping
func (h *AbsenceRequestHandler) GetOverlapping(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	excludeIDStr := c.Query("exclude_request_id")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	var excludeID *uuid.UUID
	if excludeIDStr != "" {
		id, err := uuid.Parse(excludeIDStr)
		if err == nil {
			excludeID = &id
		}
	}

	overlapping, err := h.service.GetOverlappingAbsences(userID.(uuid.UUID), start, end, excludeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overlapping)
}

// GetCounts handles GET /absence-requests/counts
func (h *AbsenceRequestHandler) GetCounts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	roleStr := middleware.GetUserRoleFromContext(c)
	if roleStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found in context"})
		return
	}

	role := enums.UserRole(roleStr)
	counts := h.service.GetPendingCounts(userID.(uuid.UUID), role)

	c.JSON(http.StatusOK, counts)
}

// GetApproved handles GET /absence-requests/approved
func (h *AbsenceRequestHandler) GetApproved(c *gin.Context) {
	filters := make(map[string]interface{})

	if periodID := c.Query("period_id"); periodID != "" {
		filters["period_id"] = periodID
	}
	if employeeID := c.Query("employee_id"); employeeID != "" {
		filters["employee_id"] = employeeID
	}
	if collarType := c.Query("collar_type"); collarType != "" {
		filters["collar_type"] = collarType
	}

	requests, err := h.service.GetApprovedRequests(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// ExportApproved handles GET /absence-requests/export
func (h *AbsenceRequestHandler) ExportApproved(c *gin.Context) {
	filters := make(map[string]interface{})

	if periodID := c.Query("period_id"); periodID != "" {
		filters["period_id"] = periodID
	}
	if employeeID := c.Query("employee_id"); employeeID != "" {
		filters["employee_id"] = employeeID
	}

	requests, err := h.service.GetApprovedRequests(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	excelFile, err := h.service.GenerateApprovedRequestsExcel(requests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=approved_requests.xlsx")
	c.Data(http.StatusOK, "application/octet-stream", excelFile)
}
