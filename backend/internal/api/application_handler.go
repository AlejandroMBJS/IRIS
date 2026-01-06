/*
Package api - IRIS Payroll System Application Handler

==============================================================================
FILE: internal/api/application_handler.go
==============================================================================

DESCRIPTION:
    Handles all job application-related endpoints for the recruitment module.
    Manages the recruitment pipeline and application workflow.

ENDPOINTS:
    POST   /recruitment/applications              - Create application
    GET    /recruitment/applications              - List applications
    GET    /recruitment/applications/:id          - Get application by ID
    PUT    /recruitment/applications/:id          - Update application
    DELETE /recruitment/applications/:id          - Delete application
    POST   /recruitment/applications/:id/stage    - Move to stage
    POST   /recruitment/applications/:id/screen   - Screen application
    POST   /recruitment/applications/:id/reject   - Reject application
    POST   /recruitment/applications/:id/withdraw - Withdraw application
    POST   /recruitment/applications/:id/hire     - Hire candidate
    GET    /recruitment/applications/stats        - Get application stats

==============================================================================
*/
package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
)

// ApplicationHandler handles application endpoints
type ApplicationHandler struct {
	applicationService *services.ApplicationService
	candidateService   *services.CandidateService
	authService        *services.AuthService
}

// NewApplicationHandler creates a new application handler
func NewApplicationHandler(
	applicationService *services.ApplicationService,
	candidateService *services.CandidateService,
	authService *services.AuthService,
) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
		candidateService:   candidateService,
		authService:        authService,
	}
}

// RegisterRoutes registers application routes
func (h *ApplicationHandler) RegisterRoutes(router *gin.RouterGroup) {
	authMiddleware := middleware.NewAuthMiddleware(h.authService)

	applications := router.Group("/applications")
	applications.Use(authMiddleware.RequireAuth())
	{
		applications.POST("", h.Create)
		applications.GET("", h.List)
		applications.GET("/:id", h.GetByID)
		applications.PUT("/:id", h.Update)
		applications.DELETE("/:id", h.Delete)

		// Workflow actions
		applications.POST("/:id/stage", h.MoveToStage)
		applications.POST("/:id/screen", h.Screen)
		applications.POST("/:id/reject", h.Reject)
		applications.POST("/:id/withdraw", h.Withdraw)
		applications.POST("/:id/hire", h.Hire)

		// Statistics
		applications.GET("/stats", h.GetStats)
	}
}

// CreateApplicationRequest represents the request body for creating an application
type CreateApplicationRequest struct {
	CandidateID    string  `json:"candidate_id" binding:"required"`
	JobPostingID   string  `json:"job_posting_id" binding:"required"`
	ExpectedSalary float64 `json:"expected_salary,omitempty"`
}

// Create creates a new application
// @Summary Create application
// @Tags recruitment
// @Accept json
// @Produce json
// @Param request body CreateApplicationRequest true "Application data"
// @Success 201 {object} models.Application
// @Router /recruitment/applications [post]
func (h *ApplicationHandler) Create(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	candidateID, err := uuid.Parse(req.CandidateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate ID"})
		return
	}

	jobPostingID, err := uuid.Parse(req.JobPostingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	dto := services.CreateApplicationDTO{
		CompanyID:      companyID,
		CandidateID:    candidateID,
		JobPostingID:   jobPostingID,
		ExpectedSalary: req.ExpectedSalary,
	}

	application, err := h.applicationService.Create(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, application)
}

// List retrieves applications with filters
// @Summary List applications
// @Tags recruitment
// @Produce json
// @Param job_posting_id query string false "Filter by job posting ID"
// @Param candidate_id query string false "Filter by candidate ID"
// @Param stage query string false "Filter by stage"
// @Param status query string false "Filter by status"
// @Param search query string false "Search in candidate name/email"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} services.PaginatedApplications
// @Router /recruitment/applications [get]
func (h *ApplicationHandler) List(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	filters := services.ApplicationFilters{
		CompanyID: companyID,
		Stage:     c.Query("stage"),
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if jobPostingID := c.Query("job_posting_id"); jobPostingID != "" {
		id, err := uuid.Parse(jobPostingID)
		if err == nil {
			filters.JobPostingID = &id
		}
	}
	if candidateID := c.Query("candidate_id"); candidateID != "" {
		id, err := uuid.Parse(candidateID)
		if err == nil {
			filters.CandidateID = &id
		}
	}
	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		filters.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filters.Limit = limit
	}

	result, err := h.applicationService.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID retrieves an application by ID
// @Summary Get application
// @Tags recruitment
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id} [get]
func (h *ApplicationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	application, err := h.applicationService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// UpdateApplicationRequest represents the request body for updating an application
type UpdateApplicationRequest struct {
	ExpectedSalary *float64 `json:"expected_salary,omitempty"`
	OfferedSalary  *float64 `json:"offered_salary,omitempty"`
	StartDate      *string  `json:"start_date,omitempty"`
}

// Update updates an application
// @Summary Update application
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body UpdateApplicationRequest true "Application data"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id} [put]
func (h *ApplicationHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	var req UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.UpdateApplicationDTO{
		ExpectedSalary: req.ExpectedSalary,
		OfferedSalary:  req.OfferedSalary,
	}

	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			dto.StartDate = &t
		}
	}

	application, err := h.applicationService.Update(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// Delete deletes an application
// @Summary Delete application
// @Tags recruitment
// @Param id path string true "Application ID"
// @Success 204
// @Router /recruitment/applications/{id} [delete]
func (h *ApplicationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	if err := h.applicationService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// MoveToStageRequest represents the request for changing application stage
type MoveToStageRequest struct {
	Stage string `json:"stage" binding:"required"`
}

// MoveToStage moves an application to a new stage
// @Summary Move application to stage
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body MoveToStageRequest true "Stage data"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id}/stage [post]
func (h *ApplicationHandler) MoveToStage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	var req MoveToStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	stage := models.ApplicationStage(req.Stage)

	application, err := h.applicationService.MoveToStage(id, stage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// ScreenApplicationRequest represents the request for screening an application
type ScreenApplicationRequest struct {
	Score int    `json:"score" binding:"required,min=1,max=5"`
	Notes string `json:"notes,omitempty"`
}

// Screen screens an application
// @Summary Screen application
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body ScreenApplicationRequest true "Screening data"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id}/screen [post]
func (h *ApplicationHandler) Screen(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	var req ScreenApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	dto := services.ScreeningDTO{
		Score:  req.Score,
		Notes:  req.Notes,
		UserID: userID,
	}

	application, err := h.applicationService.Screen(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// RejectApplicationRequest represents the request for rejecting an application
type RejectApplicationRequest struct {
	Reason string `json:"reason,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

// Reject rejects an application
// @Summary Reject application
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body RejectApplicationRequest false "Rejection data"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id}/reject [post]
func (h *ApplicationHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	var req RejectApplicationRequest
	c.ShouldBindJSON(&req) // Optional body

	dto := services.RejectionDTO{
		Reason: req.Reason,
		Notes:  req.Notes,
	}

	application, err := h.applicationService.Reject(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// Withdraw withdraws an application
// @Summary Withdraw application
// @Tags recruitment
// @Param id path string true "Application ID"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id}/withdraw [post]
func (h *ApplicationHandler) Withdraw(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	application, err := h.applicationService.Withdraw(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// HireApplicationRequest represents the request for hiring a candidate
type HireApplicationRequest struct {
	StartDate string `json:"start_date" binding:"required"`
}

// Hire hires a candidate through their application
// @Summary Hire candidate
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body HireApplicationRequest true "Hire data"
// @Success 200 {object} models.Application
// @Router /recruitment/applications/{id}/hire [post]
func (h *ApplicationHandler) Hire(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application ID"})
		return
	}

	var req HireApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
		return
	}

	application, err := h.applicationService.Hire(id, startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, application)
}

// GetStats retrieves statistics for applications
// @Summary Get application stats
// @Tags recruitment
// @Produce json
// @Param job_posting_id query string true "Job posting ID"
// @Success 200 {object} services.ApplicationStats
// @Router /recruitment/applications/stats [get]
func (h *ApplicationHandler) GetStats(c *gin.Context) {
	jobPostingIDStr := c.Query("job_posting_id")
	if jobPostingIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_posting_id is required"})
		return
	}

	jobPostingID, err := uuid.Parse(jobPostingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	stats, err := h.applicationService.GetStats(jobPostingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
