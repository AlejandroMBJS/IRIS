package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend/internal/models"
	"backend/internal/services"
)

type TrainingHandler struct {
	trainingService *services.TrainingService
}

func NewTrainingHandler(trainingService *services.TrainingService) *TrainingHandler {
	return &TrainingHandler{trainingService: trainingService}
}

// Category handlers
func (h *TrainingHandler) CreateCategory(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var req struct {
		Name        string     `json:"name" binding:"required"`
		Description string     `json:"description"`
		ParentID    *uuid.UUID `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := h.trainingService.CreateCategory(companyID.(uuid.UUID), req.Name, req.Description, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *TrainingHandler) GetCategories(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	categories, err := h.trainingService.GetCategories(companyID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Course handlers
func (h *TrainingHandler) CreateCourse(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var dto services.CreateCourseDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	uid := userID.(uuid.UUID)
	dto.CreatedByID = &uid
	course, err := h.trainingService.CreateCourse(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, course)
}

func (h *TrainingHandler) GetCourse(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}
	course, err := h.trainingService.GetCourseByID(courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *TrainingHandler) ListCourses(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	categoryIDStr := c.Query("category_id")

	filters := services.CourseFilters{
		CompanyID: companyID.(uuid.UUID),
		Status:    status,
		Page:      page,
		Limit:     limit,
	}
	if categoryIDStr != "" {
		if id, err := uuid.Parse(categoryIDStr); err == nil {
			filters.CategoryID = &id
		}
	}

	result, err := h.trainingService.ListCourses(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *TrainingHandler) UpdateCourse(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}
	var dto services.UpdateCourseDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	course, err := h.trainingService.UpdateCourse(courseID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *TrainingHandler) PublishCourse(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}
	course, err := h.trainingService.PublishCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *TrainingHandler) ArchiveCourse(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}
	course, err := h.trainingService.ArchiveCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, course)
}

// Module handlers
func (h *TrainingHandler) CreateModule(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Order       int    `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	module, err := h.trainingService.CreateModule(courseID, req.Title, req.Description, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, module)
}

func (h *TrainingHandler) CreateContent(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}
	var req struct {
		Title       string             `json:"title" binding:"required"`
		Description string             `json:"description"`
		ContentType models.ContentType `json:"content_type"`
		ContentURL  string             `json:"content_url"`
		Order       int                `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	content, err := h.trainingService.CreateContent(moduleID, req.Title, req.Description, req.ContentType, req.ContentURL, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, content)
}

// Enrollment handlers
func (h *TrainingHandler) EnrollEmployee(c *gin.Context) {
	var dto services.EnrollEmployeeDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	enrollment, err := h.trainingService.EnrollEmployee(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, enrollment)
}

func (h *TrainingHandler) GetEnrollment(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	enrollment, err := h.trainingService.GetEnrollmentByID(enrollmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enrollment not found"})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *TrainingHandler) GetEmployeeEnrollments(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	enrollments, err := h.trainingService.GetEmployeeEnrollments(companyID.(uuid.UUID), employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollments)
}

func (h *TrainingHandler) StartCourse(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	enrollment, err := h.trainingService.StartCourse(enrollmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *TrainingHandler) UpdateContentProgress(c *gin.Context) {
	var dto services.UpdateProgressDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	progress, err := h.trainingService.UpdateContentProgress(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, progress)
}

// Certificate handlers
func (h *TrainingHandler) IssueCertificate(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	certificate, err := h.trainingService.IssueCertificate(enrollmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, certificate)
}

// Statistics
func (h *TrainingHandler) GetCourseStatistics(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}
	stats, err := h.trainingService.GetCourseStats(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
