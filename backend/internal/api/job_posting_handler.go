/*
Package api - IRIS Payroll System Job Posting Handler

==============================================================================
FILE: internal/api/job_posting_handler.go
==============================================================================

DESCRIPTION:
    Handles all job posting-related endpoints for the recruitment module.

ENDPOINTS:
    POST   /recruitment/job-postings         - Create job posting
    GET    /recruitment/job-postings         - List job postings
    GET    /recruitment/job-postings/:id     - Get job posting by ID
    PUT    /recruitment/job-postings/:id     - Update job posting
    DELETE /recruitment/job-postings/:id     - Delete job posting
    POST   /recruitment/job-postings/:id/publish - Publish job posting
    POST   /recruitment/job-postings/:id/pause   - Pause job posting
    POST   /recruitment/job-postings/:id/resume  - Resume job posting
    POST   /recruitment/job-postings/:id/close   - Close job posting
    GET    /recruitment/job-postings/:id/stats   - Get job posting stats

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

// JobPostingHandler handles job posting endpoints
type JobPostingHandler struct {
	jobPostingService *services.JobPostingService
	authService       *services.AuthService
}

// NewJobPostingHandler creates a new job posting handler
func NewJobPostingHandler(jobPostingService *services.JobPostingService, authService *services.AuthService) *JobPostingHandler {
	return &JobPostingHandler{
		jobPostingService: jobPostingService,
		authService:       authService,
	}
}

// RegisterRoutes registers job posting routes
func (h *JobPostingHandler) RegisterRoutes(router *gin.RouterGroup) {
	authMiddleware := middleware.NewAuthMiddleware(h.authService)

	jobPostings := router.Group("/job-postings")
	jobPostings.Use(authMiddleware.RequireAuth())
	{
		jobPostings.POST("", h.Create)
		jobPostings.GET("", h.List)
		jobPostings.GET("/:id", h.GetByID)
		jobPostings.PUT("/:id", h.Update)
		jobPostings.DELETE("/:id", h.Delete)

		// Workflow actions
		jobPostings.POST("/:id/publish", h.Publish)
		jobPostings.POST("/:id/pause", h.Pause)
		jobPostings.POST("/:id/resume", h.Resume)
		jobPostings.POST("/:id/close", h.Close)

		// Statistics
		jobPostings.GET("/:id/stats", h.GetStats)
	}
}

// CreateJobPostingRequest represents the request body for creating a job posting
type CreateJobPostingRequest struct {
	Title              string  `json:"title" binding:"required"`
	Description        string  `json:"description" binding:"required"`
	Requirements       string  `json:"requirements,omitempty"`
	Responsibilities   string  `json:"responsibilities,omitempty"`
	DepartmentID       *string `json:"department_id,omitempty"`
	CostCenterID       *string `json:"cost_center_id,omitempty"`
	PositionLevel      string  `json:"position_level,omitempty"`
	EmploymentType     string  `json:"employment_type" binding:"required"`
	CollarType         string  `json:"collar_type,omitempty"`
	SalaryMin          float64 `json:"salary_min,omitempty"`
	SalaryMax          float64 `json:"salary_max,omitempty"`
	SalaryCurrency     string  `json:"salary_currency,omitempty"`
	SalaryFrequency    string  `json:"salary_frequency,omitempty"`
	ShowSalary         bool    `json:"show_salary,omitempty"`
	PositionsAvailable int     `json:"positions_available,omitempty"`
	Location           string  `json:"location,omitempty"`
	IsRemote           bool    `json:"is_remote,omitempty"`
	RemoteType         string  `json:"remote_type,omitempty"`
	ClosesAt           *string `json:"closes_at,omitempty"`
	HiringManagerID    *string `json:"hiring_manager_id,omitempty"`
	RecruiterID        *string `json:"recruiter_id,omitempty"`
}

// Create creates a new job posting
// @Summary Create job posting
// @Tags recruitment
// @Accept json
// @Produce json
// @Param request body CreateJobPostingRequest true "Job posting data"
// @Success 201 {object} models.JobPosting
// @Router /recruitment/job-postings [post]
func (h *JobPostingHandler) Create(c *gin.Context) {
	var req CreateJobPostingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	dto := services.CreateJobPostingDTO{
		CompanyID:          companyID,
		Title:              req.Title,
		Description:        req.Description,
		Requirements:       req.Requirements,
		Responsibilities:   req.Responsibilities,
		PositionLevel:      req.PositionLevel,
		EmploymentType:     models.EmploymentType(req.EmploymentType),
		CollarType:         req.CollarType,
		SalaryMin:          req.SalaryMin,
		SalaryMax:          req.SalaryMax,
		SalaryCurrency:     req.SalaryCurrency,
		SalaryFrequency:    req.SalaryFrequency,
		ShowSalary:         req.ShowSalary,
		PositionsAvailable: req.PositionsAvailable,
		Location:           req.Location,
		IsRemote:           req.IsRemote,
		RemoteType:         models.RemoteType(req.RemoteType),
		CreatedByUserID:    &userID,
	}

	// Parse optional UUIDs
	if req.DepartmentID != nil {
		id, err := uuid.Parse(*req.DepartmentID)
		if err == nil {
			dto.DepartmentID = &id
		}
	}
	if req.CostCenterID != nil {
		id, err := uuid.Parse(*req.CostCenterID)
		if err == nil {
			dto.CostCenterID = &id
		}
	}
	if req.HiringManagerID != nil {
		id, err := uuid.Parse(*req.HiringManagerID)
		if err == nil {
			dto.HiringManagerID = &id
		}
	}
	if req.RecruiterID != nil {
		id, err := uuid.Parse(*req.RecruiterID)
		if err == nil {
			dto.RecruiterID = &id
		}
	}
	if req.ClosesAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ClosesAt)
		if err == nil {
			dto.ClosesAt = &t
		}
	}

	jobPosting, err := h.jobPostingService.Create(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, jobPosting)
}

// List retrieves job postings with filters
// @Summary List job postings
// @Tags recruitment
// @Produce json
// @Param status query string false "Filter by status"
// @Param employment_type query string false "Filter by employment type"
// @Param search query string false "Search in title/description"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} services.PaginatedJobPostings
// @Router /recruitment/job-postings [get]
func (h *JobPostingHandler) List(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	filters := services.JobPostingFilters{
		CompanyID:      companyID,
		Status:         c.Query("status"),
		EmploymentType: c.Query("employment_type"),
		Search:         c.Query("search"),
		SortBy:         c.Query("sort_by"),
		SortOrder:      c.Query("sort_order"),
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		filters.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filters.Limit = limit
	}

	result, err := h.jobPostingService.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID retrieves a job posting by ID
// @Summary Get job posting
// @Tags recruitment
// @Produce json
// @Param id path string true "Job posting ID"
// @Success 200 {object} models.JobPosting
// @Router /recruitment/job-postings/{id} [get]
func (h *JobPostingHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	jobPosting, err := h.jobPostingService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobPosting)
}

// UpdateJobPostingRequest represents the request body for updating a job posting
type UpdateJobPostingRequest struct {
	Title              *string  `json:"title,omitempty"`
	Description        *string  `json:"description,omitempty"`
	Requirements       *string  `json:"requirements,omitempty"`
	Responsibilities   *string  `json:"responsibilities,omitempty"`
	DepartmentID       *string  `json:"department_id,omitempty"`
	CostCenterID       *string  `json:"cost_center_id,omitempty"`
	PositionLevel      *string  `json:"position_level,omitempty"`
	EmploymentType     *string  `json:"employment_type,omitempty"`
	CollarType         *string  `json:"collar_type,omitempty"`
	SalaryMin          *float64 `json:"salary_min,omitempty"`
	SalaryMax          *float64 `json:"salary_max,omitempty"`
	SalaryCurrency     *string  `json:"salary_currency,omitempty"`
	SalaryFrequency    *string  `json:"salary_frequency,omitempty"`
	ShowSalary         *bool    `json:"show_salary,omitempty"`
	PositionsAvailable *int     `json:"positions_available,omitempty"`
	Location           *string  `json:"location,omitempty"`
	IsRemote           *bool    `json:"is_remote,omitempty"`
	RemoteType         *string  `json:"remote_type,omitempty"`
	ClosesAt           *string  `json:"closes_at,omitempty"`
	HiringManagerID    *string  `json:"hiring_manager_id,omitempty"`
	RecruiterID        *string  `json:"recruiter_id,omitempty"`
}

// Update updates a job posting
// @Summary Update job posting
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Job posting ID"
// @Param request body UpdateJobPostingRequest true "Job posting data"
// @Success 200 {object} models.JobPosting
// @Router /recruitment/job-postings/{id} [put]
func (h *JobPostingHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	var req UpdateJobPostingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.UpdateJobPostingDTO{
		Title:              req.Title,
		Description:        req.Description,
		Requirements:       req.Requirements,
		Responsibilities:   req.Responsibilities,
		PositionLevel:      req.PositionLevel,
		CollarType:         req.CollarType,
		SalaryMin:          req.SalaryMin,
		SalaryMax:          req.SalaryMax,
		SalaryCurrency:     req.SalaryCurrency,
		SalaryFrequency:    req.SalaryFrequency,
		ShowSalary:         req.ShowSalary,
		PositionsAvailable: req.PositionsAvailable,
		Location:           req.Location,
		IsRemote:           req.IsRemote,
	}

	// Parse employment type
	if req.EmploymentType != nil {
		et := models.EmploymentType(*req.EmploymentType)
		dto.EmploymentType = &et
	}

	// Parse remote type
	if req.RemoteType != nil {
		rt := models.RemoteType(*req.RemoteType)
		dto.RemoteType = &rt
	}

	// Parse optional UUIDs
	if req.DepartmentID != nil {
		uid, err := uuid.Parse(*req.DepartmentID)
		if err == nil {
			dto.DepartmentID = &uid
		}
	}
	if req.CostCenterID != nil {
		uid, err := uuid.Parse(*req.CostCenterID)
		if err == nil {
			dto.CostCenterID = &uid
		}
	}
	if req.HiringManagerID != nil {
		uid, err := uuid.Parse(*req.HiringManagerID)
		if err == nil {
			dto.HiringManagerID = &uid
		}
	}
	if req.RecruiterID != nil {
		uid, err := uuid.Parse(*req.RecruiterID)
		if err == nil {
			dto.RecruiterID = &uid
		}
	}
	if req.ClosesAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ClosesAt)
		if err == nil {
			dto.ClosesAt = &t
		}
	}

	jobPosting, err := h.jobPostingService.Update(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobPosting)
}

// Delete deletes a job posting
// @Summary Delete job posting
// @Tags recruitment
// @Param id path string true "Job posting ID"
// @Success 204
// @Router /recruitment/job-postings/{id} [delete]
func (h *JobPostingHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	if err := h.jobPostingService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Publish publishes a job posting
// @Summary Publish job posting
// @Tags recruitment
// @Param id path string true "Job posting ID"
// @Success 200 {object} models.JobPosting
// @Router /recruitment/job-postings/{id}/publish [post]
func (h *JobPostingHandler) Publish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	jobPosting, err := h.jobPostingService.Publish(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobPosting)
}

// Pause pauses a job posting
// @Summary Pause job posting
// @Tags recruitment
// @Param id path string true "Job posting ID"
// @Success 200 {object} models.JobPosting
// @Router /recruitment/job-postings/{id}/pause [post]
func (h *JobPostingHandler) Pause(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	jobPosting, err := h.jobPostingService.Pause(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobPosting)
}

// Resume resumes a paused job posting
// @Summary Resume job posting
// @Tags recruitment
// @Param id path string true "Job posting ID"
// @Success 200 {object} models.JobPosting
// @Router /recruitment/job-postings/{id}/resume [post]
func (h *JobPostingHandler) Resume(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	jobPosting, err := h.jobPostingService.Resume(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobPosting)
}

// CloseJobPostingRequest represents the request body for closing a job posting
type CloseJobPostingRequest struct {
	Reason string `json:"reason,omitempty"`
}

// Close closes a job posting
// @Summary Close job posting
// @Tags recruitment
// @Accept json
// @Param id path string true "Job posting ID"
// @Param request body CloseJobPostingRequest false "Close reason"
// @Success 200 {object} models.JobPosting
// @Router /recruitment/job-postings/{id}/close [post]
func (h *JobPostingHandler) Close(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	var req CloseJobPostingRequest
	c.ShouldBindJSON(&req) // Optional body

	jobPosting, err := h.jobPostingService.Close(id, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobPosting)
}

// GetStats retrieves statistics for a job posting
// @Summary Get job posting stats
// @Tags recruitment
// @Produce json
// @Param id path string true "Job posting ID"
// @Success 200 {object} services.JobPostingStats
// @Router /recruitment/job-postings/{id}/stats [get]
func (h *JobPostingHandler) GetStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job posting ID"})
		return
	}

	stats, err := h.jobPostingService.GetStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
