/*
Package api - IRIS Payroll System Candidate Handler

==============================================================================
FILE: internal/api/candidate_handler.go
==============================================================================

DESCRIPTION:
    Handles all candidate-related endpoints for the recruitment module.

ENDPOINTS:
    POST   /recruitment/candidates         - Create candidate
    GET    /recruitment/candidates         - List candidates
    GET    /recruitment/candidates/:id     - Get candidate by ID
    PUT    /recruitment/candidates/:id     - Update candidate
    DELETE /recruitment/candidates/:id     - Delete candidate
    POST   /recruitment/candidates/:id/hire    - Mark as hired
    POST   /recruitment/candidates/:id/reject  - Mark as rejected

==============================================================================
*/
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
)

// CandidateHandler handles candidate endpoints
type CandidateHandler struct {
	candidateService *services.CandidateService
	authService      *services.AuthService
}

// NewCandidateHandler creates a new candidate handler
func NewCandidateHandler(candidateService *services.CandidateService, authService *services.AuthService) *CandidateHandler {
	return &CandidateHandler{
		candidateService: candidateService,
		authService:      authService,
	}
}

// RegisterRoutes registers candidate routes
func (h *CandidateHandler) RegisterRoutes(router *gin.RouterGroup) {
	authMiddleware := middleware.NewAuthMiddleware(h.authService)

	candidates := router.Group("/candidates")
	candidates.Use(authMiddleware.RequireAuth())
	{
		candidates.POST("", h.Create)
		candidates.GET("", h.List)
		candidates.GET("/:id", h.GetByID)
		candidates.PUT("/:id", h.Update)
		candidates.DELETE("/:id", h.Delete)

		// Status actions
		candidates.POST("/:id/hire", h.MarkAsHired)
		candidates.POST("/:id/reject", h.MarkAsRejected)
	}
}

// CreateCandidateRequest represents the request body for creating a candidate
type CreateCandidateRequest struct {
	FirstName            string   `json:"first_name" binding:"required"`
	LastName             string   `json:"last_name" binding:"required"`
	Email                string   `json:"email" binding:"required,email"`
	Phone                string   `json:"phone,omitempty"`
	CurrentTitle         string   `json:"current_title,omitempty"`
	CurrentCompany       string   `json:"current_company,omitempty"`
	YearsExperience      int      `json:"years_experience,omitempty"`
	LinkedInURL          string   `json:"linkedin_url,omitempty"`
	PortfolioURL         string   `json:"portfolio_url,omitempty"`
	ResumeFileID         *string  `json:"resume_file_id,omitempty"`
	CoverLetter          string   `json:"cover_letter,omitempty"`
	Source               string   `json:"source,omitempty"`
	SourceDetails        string   `json:"source_details,omitempty"`
	ReferredByEmployeeID *string  `json:"referred_by_employee_id,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
}

// Create creates a new candidate
// @Summary Create candidate
// @Tags recruitment
// @Accept json
// @Produce json
// @Param request body CreateCandidateRequest true "Candidate data"
// @Success 201 {object} models.Candidate
// @Router /recruitment/candidates [post]
func (h *CandidateHandler) Create(c *gin.Context) {
	var req CreateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	dto := services.CreateCandidateDTO{
		CompanyID:       companyID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		CurrentTitle:    req.CurrentTitle,
		CurrentCompany:  req.CurrentCompany,
		YearsExperience: req.YearsExperience,
		LinkedInURL:     req.LinkedInURL,
		PortfolioURL:    req.PortfolioURL,
		CoverLetter:     req.CoverLetter,
		Source:          req.Source,
		SourceDetails:   req.SourceDetails,
		Tags:            req.Tags,
	}

	// Parse optional UUIDs
	if req.ResumeFileID != nil {
		id, err := uuid.Parse(*req.ResumeFileID)
		if err == nil {
			dto.ResumeFileID = &id
		}
	}
	if req.ReferredByEmployeeID != nil {
		id, err := uuid.Parse(*req.ReferredByEmployeeID)
		if err == nil {
			dto.ReferredByEmployeeID = &id
		}
	}

	candidate, err := h.candidateService.Create(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, candidate)
}

// List retrieves candidates with filters
// @Summary List candidates
// @Tags recruitment
// @Produce json
// @Param status query string false "Filter by status"
// @Param source query string false "Filter by source"
// @Param search query string false "Search in name/email"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} services.PaginatedCandidates
// @Router /recruitment/candidates [get]
func (h *CandidateHandler) List(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	filters := services.CandidateFilters{
		CompanyID: companyID,
		Status:    c.Query("status"),
		Source:    c.Query("source"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		filters.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filters.Limit = limit
	}

	// Parse tags
	if tags := c.Query("tags"); tags != "" {
		filters.Tags = []string{tags}
	}

	result, err := h.candidateService.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID retrieves a candidate by ID
// @Summary Get candidate
// @Tags recruitment
// @Produce json
// @Param id path string true "Candidate ID"
// @Success 200 {object} models.Candidate
// @Router /recruitment/candidates/{id} [get]
func (h *CandidateHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate ID"})
		return
	}

	candidate, err := h.candidateService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidate)
}

// UpdateCandidateRequest represents the request body for updating a candidate
type UpdateCandidateRequest struct {
	FirstName            *string  `json:"first_name,omitempty"`
	LastName             *string  `json:"last_name,omitempty"`
	Email                *string  `json:"email,omitempty"`
	Phone                *string  `json:"phone,omitempty"`
	CurrentTitle         *string  `json:"current_title,omitempty"`
	CurrentCompany       *string  `json:"current_company,omitempty"`
	YearsExperience      *int     `json:"years_experience,omitempty"`
	LinkedInURL          *string  `json:"linkedin_url,omitempty"`
	PortfolioURL         *string  `json:"portfolio_url,omitempty"`
	ResumeFileID         *string  `json:"resume_file_id,omitempty"`
	CoverLetter          *string  `json:"cover_letter,omitempty"`
	Source               *string  `json:"source,omitempty"`
	SourceDetails        *string  `json:"source_details,omitempty"`
	ReferredByEmployeeID *string  `json:"referred_by_employee_id,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	Status               *string  `json:"status,omitempty"`
}

// Update updates a candidate
// @Summary Update candidate
// @Tags recruitment
// @Accept json
// @Produce json
// @Param id path string true "Candidate ID"
// @Param request body UpdateCandidateRequest true "Candidate data"
// @Success 200 {object} models.Candidate
// @Router /recruitment/candidates/{id} [put]
func (h *CandidateHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate ID"})
		return
	}

	var req UpdateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.UpdateCandidateDTO{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		CurrentTitle:    req.CurrentTitle,
		CurrentCompany:  req.CurrentCompany,
		YearsExperience: req.YearsExperience,
		LinkedInURL:     req.LinkedInURL,
		PortfolioURL:    req.PortfolioURL,
		CoverLetter:     req.CoverLetter,
		Source:          req.Source,
		SourceDetails:   req.SourceDetails,
		Tags:            req.Tags,
	}

	// Parse optional UUIDs
	if req.ResumeFileID != nil {
		uid, err := uuid.Parse(*req.ResumeFileID)
		if err == nil {
			dto.ResumeFileID = &uid
		}
	}
	if req.ReferredByEmployeeID != nil {
		uid, err := uuid.Parse(*req.ReferredByEmployeeID)
		if err == nil {
			dto.ReferredByEmployeeID = &uid
		}
	}
	if req.Status != nil {
		status := models.CandidateStatus(*req.Status)
		dto.Status = &status
	}

	candidate, err := h.candidateService.Update(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidate)
}

// Delete deletes a candidate
// @Summary Delete candidate
// @Tags recruitment
// @Param id path string true "Candidate ID"
// @Success 204
// @Router /recruitment/candidates/{id} [delete]
func (h *CandidateHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate ID"})
		return
	}

	if err := h.candidateService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkAsHired marks a candidate as hired
// @Summary Mark candidate as hired
// @Tags recruitment
// @Param id path string true "Candidate ID"
// @Success 200 {object} models.Candidate
// @Router /recruitment/candidates/{id}/hire [post]
func (h *CandidateHandler) MarkAsHired(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate ID"})
		return
	}

	candidate, err := h.candidateService.MarkAsHired(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidate)
}

// MarkAsRejected marks a candidate as rejected
// @Summary Mark candidate as rejected
// @Tags recruitment
// @Param id path string true "Candidate ID"
// @Success 200 {object} models.Candidate
// @Router /recruitment/candidates/{id}/reject [post]
func (h *CandidateHandler) MarkAsRejected(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate ID"})
		return
	}

	candidate, err := h.candidateService.MarkAsRejected(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidate)
}
