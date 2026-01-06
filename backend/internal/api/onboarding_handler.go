/*
Package api - IRIS Payroll System Onboarding Handler

==============================================================================
FILE: internal/api/onboarding_handler.go
==============================================================================

DESCRIPTION:
    Handles all onboarding-related endpoints including templates, checklists,
    and task management.

ENDPOINTS:
    Templates:
    POST   /onboarding/templates              - Create template
    GET    /onboarding/templates              - List templates
    GET    /onboarding/templates/:id          - Get template
    PUT    /onboarding/templates/:id          - Update template
    DELETE /onboarding/templates/:id          - Delete template
    POST   /onboarding/templates/:id/duplicate - Duplicate template
    POST   /onboarding/templates/:id/tasks    - Add task template
    PUT    /onboarding/templates/:id/tasks/:taskId - Update task template
    DELETE /onboarding/templates/:id/tasks/:taskId - Delete task template

    Checklists:
    POST   /onboarding/checklists             - Create checklist
    GET    /onboarding/checklists             - List checklists
    GET    /onboarding/checklists/:id         - Get checklist
    PUT    /onboarding/checklists/:id         - Update checklist
    DELETE /onboarding/checklists/:id         - Delete checklist
    POST   /onboarding/checklists/:id/cancel  - Cancel checklist
    POST   /onboarding/checklists/:id/tasks   - Add task
    GET    /onboarding/checklists/:id/notes   - Get notes
    POST   /onboarding/checklists/:id/notes   - Add note

    Tasks:
    GET    /onboarding/tasks/:id              - Get task
    POST   /onboarding/tasks/:id/complete     - Complete task
    POST   /onboarding/tasks/:id/approve      - Approve task
    POST   /onboarding/tasks/:id/skip         - Skip task
    DELETE /onboarding/tasks/:id              - Delete task

    Statistics:
    GET    /onboarding/stats                  - Get stats
    GET    /onboarding/overdue                - Get overdue tasks

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

// OnboardingHandler handles onboarding endpoints
type OnboardingHandler struct {
	templateService  *services.OnboardingTemplateService
	checklistService *services.OnboardingChecklistService
	authService      *services.AuthService
}

// NewOnboardingHandler creates a new onboarding handler
func NewOnboardingHandler(
	templateService *services.OnboardingTemplateService,
	checklistService *services.OnboardingChecklistService,
	authService *services.AuthService,
) *OnboardingHandler {
	return &OnboardingHandler{
		templateService:  templateService,
		checklistService: checklistService,
		authService:      authService,
	}
}

// RegisterRoutes registers onboarding routes
func (h *OnboardingHandler) RegisterRoutes(router *gin.RouterGroup) {
	authMiddleware := middleware.NewAuthMiddleware(h.authService)

	onboarding := router.Group("/onboarding")
	onboarding.Use(authMiddleware.RequireAuth())
	{
		// Template routes
		templates := onboarding.Group("/templates")
		{
			templates.POST("", h.CreateTemplate)
			templates.GET("", h.ListTemplates)
			templates.GET("/:id", h.GetTemplate)
			templates.PUT("/:id", h.UpdateTemplate)
			templates.DELETE("/:id", h.DeleteTemplate)
			templates.POST("/:id/duplicate", h.DuplicateTemplate)
			templates.POST("/:id/tasks", h.CreateTaskTemplate)
			templates.PUT("/:id/tasks/:taskId", h.UpdateTaskTemplate)
			templates.DELETE("/:id/tasks/:taskId", h.DeleteTaskTemplate)
			templates.POST("/:id/tasks/reorder", h.ReorderTaskTemplates)
		}

		// Checklist routes
		checklists := onboarding.Group("/checklists")
		{
			checklists.POST("", h.CreateChecklist)
			checklists.GET("", h.ListChecklists)
			checklists.GET("/:id", h.GetChecklist)
			checklists.PUT("/:id", h.UpdateChecklist)
			checklists.DELETE("/:id", h.DeleteChecklist)
			checklists.POST("/:id/cancel", h.CancelChecklist)
			checklists.POST("/:id/tasks", h.AddTask)
			checklists.GET("/:id/notes", h.GetNotes)
			checklists.POST("/:id/notes", h.AddNote)
		}

		// Task routes
		tasks := onboarding.Group("/tasks")
		{
			tasks.GET("/:id", h.GetTask)
			tasks.POST("/:id/complete", h.CompleteTask)
			tasks.POST("/:id/approve", h.ApproveTask)
			tasks.POST("/:id/skip", h.SkipTask)
			tasks.DELETE("/:id", h.DeleteTask)
		}

		// Stats and reports
		onboarding.GET("/stats", h.GetStats)
		onboarding.GET("/overdue", h.GetOverdueTasks)
		onboarding.GET("/employee/:employeeId", h.GetEmployeeOnboarding)
	}
}

// === Template Handlers ===

// CreateTemplateRequest represents the request for creating a template
type CreateTemplateRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description,omitempty"`
	DepartmentID  *string `json:"department_id,omitempty"`
	PositionLevel string  `json:"position_level,omitempty"`
	CollarType    string  `json:"collar_type,omitempty"`
	EstimatedDays int     `json:"estimated_days,omitempty"`
	IsDefault     bool    `json:"is_default,omitempty"`
}

// CreateTemplate creates a new onboarding template
func (h *OnboardingHandler) CreateTemplate(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.CreateTemplateDTO{
		CompanyID:     companyID,
		Name:          req.Name,
		Description:   req.Description,
		PositionLevel: req.PositionLevel,
		CollarType:    req.CollarType,
		EstimatedDays: req.EstimatedDays,
		IsDefault:     req.IsDefault,
		CreatedByID:   &userID,
	}

	if req.DepartmentID != nil {
		id, err := uuid.Parse(*req.DepartmentID)
		if err == nil {
			dto.DepartmentID = &id
		}
	}

	template, err := h.templateService.Create(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// ListTemplates lists all templates
func (h *OnboardingHandler) ListTemplates(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	filters := services.TemplateFilters{
		CompanyID:     companyID,
		PositionLevel: c.Query("position_level"),
		CollarType:    c.Query("collar_type"),
		Search:        c.Query("search"),
	}

	if departmentID := c.Query("department_id"); departmentID != "" {
		id, err := uuid.Parse(departmentID)
		if err == nil {
			filters.DepartmentID = &id
		}
	}
	if isActive := c.Query("is_active"); isActive != "" {
		active := isActive == "true"
		filters.IsActive = &active
	}
	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		filters.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filters.Limit = limit
	}

	result, err := h.templateService.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTemplate retrieves a template by ID
func (h *OnboardingHandler) GetTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := h.templateService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateTemplateRequest represents the request for updating a template
type UpdateTemplateRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	DepartmentID  *string `json:"department_id,omitempty"`
	PositionLevel *string `json:"position_level,omitempty"`
	CollarType    *string `json:"collar_type,omitempty"`
	EstimatedDays *int    `json:"estimated_days,omitempty"`
	IsDefault     *bool   `json:"is_default,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// UpdateTemplate updates a template
func (h *OnboardingHandler) UpdateTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.UpdateTemplateDTO{
		Name:          req.Name,
		Description:   req.Description,
		PositionLevel: req.PositionLevel,
		CollarType:    req.CollarType,
		EstimatedDays: req.EstimatedDays,
		IsDefault:     req.IsDefault,
		IsActive:      req.IsActive,
	}

	if req.DepartmentID != nil {
		uid, err := uuid.Parse(*req.DepartmentID)
		if err == nil {
			dto.DepartmentID = &uid
		}
	}

	template, err := h.templateService.Update(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate deletes a template
func (h *OnboardingHandler) DeleteTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	if err := h.templateService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DuplicateTemplateRequest represents the request for duplicating a template
type DuplicateTemplateRequest struct {
	NewName string `json:"new_name" binding:"required"`
}

// DuplicateTemplate duplicates a template
func (h *OnboardingHandler) DuplicateTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req DuplicateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	template, err := h.templateService.DuplicateTemplate(id, req.NewName, &userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// CreateTaskTemplateRequest represents the request for creating a task template
type CreateTaskTemplateRequest struct {
	Title            string   `json:"title" binding:"required"`
	Description      string   `json:"description,omitempty"`
	TaskType         string   `json:"task_type" binding:"required"`
	DueAfterDays     int      `json:"due_after_days,omitempty"`
	DisplayOrder     int      `json:"display_order,omitempty"`
	AssigneeRole     string   `json:"assignee_role,omitempty"`
	AssigneeUserID   *string  `json:"assignee_user_id,omitempty"`
	NotifyOnOverdue  bool     `json:"notify_on_overdue,omitempty"`
	IsRequired       bool     `json:"is_required,omitempty"`
	RequiresApproval bool     `json:"requires_approval,omitempty"`
	ApproverRole     string   `json:"approver_role,omitempty"`
	DocumentURL      string   `json:"document_url,omitempty"`
	FormFields       []string `json:"form_fields,omitempty"`
	DependsOnTaskID  *string  `json:"depends_on_task_id,omitempty"`
}

// CreateTaskTemplate creates a task template
func (h *OnboardingHandler) CreateTaskTemplate(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req CreateTaskTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.CreateTaskTemplateDTO{
		TemplateID:       templateID,
		Title:            req.Title,
		Description:      req.Description,
		TaskType:         models.OnboardingTaskType(req.TaskType),
		DueAfterDays:     req.DueAfterDays,
		DisplayOrder:     req.DisplayOrder,
		AssigneeRole:     req.AssigneeRole,
		NotifyOnOverdue:  req.NotifyOnOverdue,
		IsRequired:       req.IsRequired,
		RequiresApproval: req.RequiresApproval,
		ApproverRole:     req.ApproverRole,
		DocumentURL:      req.DocumentURL,
		FormFields:       req.FormFields,
	}

	if req.AssigneeUserID != nil {
		id, err := uuid.Parse(*req.AssigneeUserID)
		if err == nil {
			dto.AssigneeUserID = &id
		}
	}
	if req.DependsOnTaskID != nil {
		id, err := uuid.Parse(*req.DependsOnTaskID)
		if err == nil {
			dto.DependsOnTaskID = &id
		}
	}

	taskTemplate, err := h.templateService.CreateTaskTemplate(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, taskTemplate)
}

// UpdateTaskTemplate updates a task template
func (h *OnboardingHandler) UpdateTaskTemplate(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task template ID"})
		return
	}

	var req CreateTaskTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	taskType := models.OnboardingTaskType(req.TaskType)
	dto := services.UpdateTaskTemplateDTO{
		Title:            &req.Title,
		Description:      &req.Description,
		TaskType:         &taskType,
		DueAfterDays:     &req.DueAfterDays,
		DisplayOrder:     &req.DisplayOrder,
		AssigneeRole:     &req.AssigneeRole,
		NotifyOnOverdue:  &req.NotifyOnOverdue,
		IsRequired:       &req.IsRequired,
		RequiresApproval: &req.RequiresApproval,
		ApproverRole:     &req.ApproverRole,
		DocumentURL:      &req.DocumentURL,
		FormFields:       req.FormFields,
	}

	taskTemplate, err := h.templateService.UpdateTaskTemplate(taskID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taskTemplate)
}

// DeleteTaskTemplate deletes a task template
func (h *OnboardingHandler) DeleteTaskTemplate(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task template ID"})
		return
	}

	if err := h.templateService.DeleteTaskTemplate(taskID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderTaskTemplatesRequest represents the request for reordering tasks
type ReorderTaskTemplatesRequest struct {
	TaskOrder []string `json:"task_order" binding:"required"`
}

// ReorderTaskTemplates reorders task templates
func (h *OnboardingHandler) ReorderTaskTemplates(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req ReorderTaskTemplatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	var taskOrder []uuid.UUID
	for _, idStr := range req.TaskOrder {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID in order"})
			return
		}
		taskOrder = append(taskOrder, id)
	}

	if err := h.templateService.ReorderTaskTemplates(templateID, taskOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tasks reordered successfully"})
}

// === Checklist Handlers ===

// CreateChecklistRequest represents the request for creating a checklist
type CreateChecklistRequest struct {
	EmployeeID      string  `json:"employee_id" binding:"required"`
	TemplateID      *string `json:"template_id,omitempty"`
	Title           string  `json:"title,omitempty"`
	Description     string  `json:"description,omitempty"`
	StartDate       string  `json:"start_date" binding:"required"`
	TargetEndDate   string  `json:"target_end_date,omitempty"`
	HRContactID     *string `json:"hr_contact_id,omitempty"`
	BuddyEmployeeID *string `json:"buddy_employee_id,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

// CreateChecklist creates a new checklist
func (h *OnboardingHandler) CreateChecklist(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CreateChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee ID"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
		return
	}

	// Default target end date to 30 days after start
	targetEndDate := startDate.AddDate(0, 0, 30)
	if req.TargetEndDate != "" {
		targetEndDate, err = time.Parse("2006-01-02", req.TargetEndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target end date format, use YYYY-MM-DD"})
			return
		}
	}

	dto := services.CreateChecklistDTO{
		CompanyID:     companyID,
		EmployeeID:    employeeID,
		Title:         req.Title,
		Description:   req.Description,
		StartDate:     startDate,
		TargetEndDate: targetEndDate,
		Notes:         req.Notes,
		CreatedByID:   &userID,
	}

	if req.TemplateID != nil {
		id, err := uuid.Parse(*req.TemplateID)
		if err == nil {
			dto.TemplateID = &id
		}
	}
	if req.HRContactID != nil {
		id, err := uuid.Parse(*req.HRContactID)
		if err == nil {
			dto.HRContactID = &id
		}
	}
	if req.BuddyEmployeeID != nil {
		id, err := uuid.Parse(*req.BuddyEmployeeID)
		if err == nil {
			dto.BuddyEmployeeID = &id
		}
	}

	checklist, err := h.checklistService.Create(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, checklist)
}

// ListChecklists lists all checklists
func (h *OnboardingHandler) ListChecklists(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	filters := services.ChecklistFilters{
		CompanyID: companyID,
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if employeeID := c.Query("employee_id"); employeeID != "" {
		id, err := uuid.Parse(employeeID)
		if err == nil {
			filters.EmployeeID = &id
		}
	}
	if hrContactID := c.Query("hr_contact_id"); hrContactID != "" {
		id, err := uuid.Parse(hrContactID)
		if err == nil {
			filters.HRContactID = &id
		}
	}
	if isOverdue := c.Query("is_overdue"); isOverdue == "true" {
		overdue := true
		filters.IsOverdue = &overdue
	}
	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		filters.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filters.Limit = limit
	}

	result, err := h.checklistService.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetChecklist retrieves a checklist by ID
func (h *OnboardingHandler) GetChecklist(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	checklist, err := h.checklistService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, checklist)
}

// UpdateChecklistRequest represents the request for updating a checklist
type UpdateChecklistRequest struct {
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	TargetEndDate   *string `json:"target_end_date,omitempty"`
	HRContactID     *string `json:"hr_contact_id,omitempty"`
	BuddyEmployeeID *string `json:"buddy_employee_id,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

// UpdateChecklist updates a checklist
func (h *OnboardingHandler) UpdateChecklist(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	var req UpdateChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.UpdateChecklistDTO{
		Title:       req.Title,
		Description: req.Description,
		Notes:       req.Notes,
	}

	if req.TargetEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.TargetEndDate)
		if err == nil {
			dto.TargetEndDate = &t
		}
	}
	if req.HRContactID != nil {
		id, err := uuid.Parse(*req.HRContactID)
		if err == nil {
			dto.HRContactID = &id
		}
	}
	if req.BuddyEmployeeID != nil {
		id, err := uuid.Parse(*req.BuddyEmployeeID)
		if err == nil {
			dto.BuddyEmployeeID = &id
		}
	}

	checklist, err := h.checklistService.Update(id, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, checklist)
}

// DeleteChecklist deletes a checklist
func (h *OnboardingHandler) DeleteChecklist(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	if err := h.checklistService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CancelChecklistRequest represents the request for cancelling a checklist
type CancelChecklistRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CancelChecklist cancels a checklist
func (h *OnboardingHandler) CancelChecklist(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	var req CancelChecklistRequest
	c.ShouldBindJSON(&req)

	checklist, err := h.checklistService.CancelChecklist(id, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, checklist)
}

// AddTaskRequest represents the request for adding a task
type AddTaskRequest struct {
	Title            string  `json:"title" binding:"required"`
	Description      string  `json:"description,omitempty"`
	TaskType         string  `json:"task_type" binding:"required"`
	DueDate          *string `json:"due_date,omitempty"`
	DisplayOrder     int     `json:"display_order,omitempty"`
	AssigneeID       *string `json:"assignee_id,omitempty"`
	AssigneeRole     string  `json:"assignee_role,omitempty"`
	IsRequired       bool    `json:"is_required,omitempty"`
	RequiresApproval bool    `json:"requires_approval,omitempty"`
	ApproverRole     string  `json:"approver_role,omitempty"`
	DocumentURL      string  `json:"document_url,omitempty"`
	DependsOnTaskID  *string `json:"depends_on_task_id,omitempty"`
}

// AddTask adds a task to a checklist
func (h *OnboardingHandler) AddTask(c *gin.Context) {
	checklistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	var req AddTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	dto := services.CreateTaskDTO{
		ChecklistID:      checklistID,
		Title:            req.Title,
		Description:      req.Description,
		TaskType:         models.OnboardingTaskType(req.TaskType),
		DisplayOrder:     req.DisplayOrder,
		AssigneeRole:     req.AssigneeRole,
		IsRequired:       req.IsRequired,
		RequiresApproval: req.RequiresApproval,
		ApproverRole:     req.ApproverRole,
		DocumentURL:      req.DocumentURL,
	}

	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			dto.DueDate = &t
		}
	}
	if req.AssigneeID != nil {
		id, err := uuid.Parse(*req.AssigneeID)
		if err == nil {
			dto.AssigneeID = &id
		}
	}
	if req.DependsOnTaskID != nil {
		id, err := uuid.Parse(*req.DependsOnTaskID)
		if err == nil {
			dto.DependsOnTaskID = &id
		}
	}

	task, err := h.checklistService.AddTask(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetNotes retrieves notes for a checklist
func (h *OnboardingHandler) GetNotes(c *gin.Context) {
	checklistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	// Check if user is HR to show internal notes
	role := middleware.GetUserRoleFromContext(c)
	includeInternal := role == "admin" || role == "hr"

	notes, err := h.checklistService.GetNotes(checklistID, includeInternal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

// AddNoteRequest represents the request for adding a note
type AddNoteRequest struct {
	Content    string  `json:"content" binding:"required"`
	TaskID     *string `json:"task_id,omitempty"`
	NoteType   string  `json:"note_type,omitempty"`
	IsInternal bool    `json:"is_internal,omitempty"`
}

// AddNote adds a note to a checklist
func (h *OnboardingHandler) AddNote(c *gin.Context) {
	checklistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist ID"})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Error", "message": err.Error()})
		return
	}

	var taskID *uuid.UUID
	if req.TaskID != nil {
		id, err := uuid.Parse(*req.TaskID)
		if err == nil {
			taskID = &id
		}
	}

	note, err := h.checklistService.AddNote(checklistID, taskID, req.Content, req.NoteType, req.IsInternal, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// === Task Handlers ===

// GetTask retrieves a task by ID
func (h *OnboardingHandler) GetTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := h.checklistService.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CompleteTaskRequest represents the request for completing a task
type CompleteTaskRequest struct {
	CompletionNotes string  `json:"completion_notes,omitempty"`
	UploadedFileID  *string `json:"uploaded_file_id,omitempty"`
	FormData        string  `json:"form_data,omitempty"`
}

// CompleteTask completes a task
func (h *OnboardingHandler) CompleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CompleteTaskRequest
	c.ShouldBindJSON(&req)

	dto := services.CompleteTaskDTO{
		CompletedByID:   userID,
		CompletionNotes: req.CompletionNotes,
		FormData:        req.FormData,
	}

	if req.UploadedFileID != nil {
		id, err := uuid.Parse(*req.UploadedFileID)
		if err == nil {
			dto.UploadedFileID = &id
		}
	}

	task, err := h.checklistService.CompleteTask(taskID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ApproveTaskRequest represents the request for approving a task
type ApproveTaskRequest struct {
	ApprovalNotes string `json:"approval_notes,omitempty"`
}

// ApproveTask approves a task
func (h *OnboardingHandler) ApproveTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req ApproveTaskRequest
	c.ShouldBindJSON(&req)

	dto := services.ApproveTaskDTO{
		ApprovedByID:  userID,
		ApprovalNotes: req.ApprovalNotes,
	}

	task, err := h.checklistService.ApproveTask(taskID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// SkipTaskRequest represents the request for skipping a task
type SkipTaskRequest struct {
	Reason string `json:"reason,omitempty"`
}

// SkipTask skips a task
func (h *OnboardingHandler) SkipTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req SkipTaskRequest
	c.ShouldBindJSON(&req)

	task, err := h.checklistService.SkipTask(taskID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task
func (h *OnboardingHandler) DeleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	if err := h.checklistService.DeleteTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// === Stats & Reports ===

// GetStats retrieves onboarding statistics
func (h *OnboardingHandler) GetStats(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	stats, err := h.checklistService.GetStats(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetOverdueTasks retrieves all overdue tasks
func (h *OnboardingHandler) GetOverdueTasks(c *gin.Context) {
	_, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	tasks, err := h.checklistService.GetOverdueTasks(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetEmployeeOnboarding retrieves onboarding history for an employee
func (h *OnboardingHandler) GetEmployeeOnboarding(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee ID"})
		return
	}

	checklists, err := h.checklistService.GetEmployeeOnboardingHistory(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, checklists)
}
