/*
Package services - IRIS Payroll System Onboarding Template Service

==============================================================================
FILE: internal/services/onboarding_template_service.go
==============================================================================

DESCRIPTION:
    Provides business logic for managing onboarding templates and task templates.
    Templates define reusable onboarding checklists for different roles/positions.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OnboardingTemplateService provides business logic for onboarding templates
type OnboardingTemplateService struct {
	db *gorm.DB
}

// NewOnboardingTemplateService creates a new OnboardingTemplateService
func NewOnboardingTemplateService(db *gorm.DB) *OnboardingTemplateService {
	return &OnboardingTemplateService{db: db}
}

// CreateTemplateDTO contains data for creating an onboarding template
type CreateTemplateDTO struct {
	CompanyID     uuid.UUID
	Name          string
	Description   string
	DepartmentID  *uuid.UUID
	PositionLevel string
	CollarType    string
	EstimatedDays int
	IsDefault     bool
	CreatedByID   *uuid.UUID
}

// UpdateTemplateDTO contains data for updating an onboarding template
type UpdateTemplateDTO struct {
	Name          *string
	Description   *string
	DepartmentID  *uuid.UUID
	PositionLevel *string
	CollarType    *string
	EstimatedDays *int
	IsDefault     *bool
	IsActive      *bool
}

// CreateTaskTemplateDTO contains data for creating a task template
type CreateTaskTemplateDTO struct {
	TemplateID       uuid.UUID
	Title            string
	Description      string
	TaskType         models.OnboardingTaskType
	DueAfterDays     int
	DisplayOrder     int
	AssigneeRole     string
	AssigneeUserID   *uuid.UUID
	NotifyOnOverdue  bool
	IsRequired       bool
	RequiresApproval bool
	ApproverRole     string
	DocumentURL      string
	FormFields       []string
	DependsOnTaskID  *uuid.UUID
}

// UpdateTaskTemplateDTO contains data for updating a task template
type UpdateTaskTemplateDTO struct {
	Title            *string
	Description      *string
	TaskType         *models.OnboardingTaskType
	DueAfterDays     *int
	DisplayOrder     *int
	AssigneeRole     *string
	AssigneeUserID   *uuid.UUID
	NotifyOnOverdue  *bool
	IsRequired       *bool
	RequiresApproval *bool
	ApproverRole     *string
	DocumentURL      *string
	FormFields       []string
	DependsOnTaskID  *uuid.UUID
	IsActive         *bool
}

// TemplateFilters contains filters for listing templates
type TemplateFilters struct {
	CompanyID     uuid.UUID
	DepartmentID  *uuid.UUID
	PositionLevel string
	CollarType    string
	IsActive      *bool
	Search        string
	Page          int
	Limit         int
}

// PaginatedTemplates contains paginated template results
type PaginatedTemplates struct {
	Data       []models.OnboardingTemplate `json:"data"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

// Create creates a new onboarding template
func (s *OnboardingTemplateService) Create(dto CreateTemplateDTO) (*models.OnboardingTemplate, error) {
	if dto.Name == "" {
		return nil, errors.New("template name is required")
	}

	// If setting as default, unset other defaults
	if dto.IsDefault {
		s.db.Model(&models.OnboardingTemplate{}).
			Where("company_id = ? AND is_default = ?", dto.CompanyID, true).
			Update("is_default", false)
	}

	template := &models.OnboardingTemplate{
		CompanyID:     dto.CompanyID,
		Name:          dto.Name,
		Description:   dto.Description,
		DepartmentID:  dto.DepartmentID,
		PositionLevel: dto.PositionLevel,
		CollarType:    dto.CollarType,
		EstimatedDays: dto.EstimatedDays,
		IsDefault:     dto.IsDefault,
		IsActive:      true,
		CreatedByID:   dto.CreatedByID,
	}

	if template.EstimatedDays <= 0 {
		template.EstimatedDays = 30 // Default 30 days
	}

	if err := s.db.Create(template).Error; err != nil {
		return nil, err
	}

	return template, nil
}

// GetByID retrieves a template by ID with task templates
func (s *OnboardingTemplateService) GetByID(id uuid.UUID) (*models.OnboardingTemplate, error) {
	var template models.OnboardingTemplate
	err := s.db.Preload("TaskTemplates", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).First(&template, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return &template, nil
}

// Update updates a template
func (s *OnboardingTemplateService) Update(id uuid.UUID, dto UpdateTemplateDTO) (*models.OnboardingTemplate, error) {
	template, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// If setting as default, unset other defaults
	if dto.IsDefault != nil && *dto.IsDefault {
		s.db.Model(&models.OnboardingTemplate{}).
			Where("company_id = ? AND is_default = ? AND id != ?", template.CompanyID, true, id).
			Update("is_default", false)
	}

	if dto.Name != nil {
		template.Name = *dto.Name
	}
	if dto.Description != nil {
		template.Description = *dto.Description
	}
	if dto.DepartmentID != nil {
		template.DepartmentID = dto.DepartmentID
	}
	if dto.PositionLevel != nil {
		template.PositionLevel = *dto.PositionLevel
	}
	if dto.CollarType != nil {
		template.CollarType = *dto.CollarType
	}
	if dto.EstimatedDays != nil {
		template.EstimatedDays = *dto.EstimatedDays
	}
	if dto.IsDefault != nil {
		template.IsDefault = *dto.IsDefault
	}
	if dto.IsActive != nil {
		template.IsActive = *dto.IsActive
	}

	if err := s.db.Save(template).Error; err != nil {
		return nil, err
	}

	return template, nil
}

// Delete soft-deletes a template
func (s *OnboardingTemplateService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.OnboardingTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("template not found")
	}
	return nil
}

// List retrieves templates with filters
func (s *OnboardingTemplateService) List(filters TemplateFilters) (*PaginatedTemplates, error) {
	query := s.db.Model(&models.OnboardingTemplate{}).Where("company_id = ?", filters.CompanyID)

	if filters.DepartmentID != nil {
		query = query.Where("department_id = ?", *filters.DepartmentID)
	}
	if filters.PositionLevel != "" {
		query = query.Where("position_level = ?", filters.PositionLevel)
	}
	if filters.CollarType != "" {
		query = query.Where("collar_type = ?", filters.CollarType)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)
	query = query.Order("is_default DESC, name ASC")

	var templates []models.OnboardingTemplate
	if err := query.Preload("TaskTemplates").Find(&templates).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedTemplates{
		Data:       templates,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetDefaultTemplate gets the default template for a company
func (s *OnboardingTemplateService) GetDefaultTemplate(companyID uuid.UUID) (*models.OnboardingTemplate, error) {
	var template models.OnboardingTemplate
	err := s.db.Where("company_id = ? AND is_default = ? AND is_active = ?", companyID, true, true).
		Preload("TaskTemplates").
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// DuplicateTemplate creates a copy of a template
func (s *OnboardingTemplateService) DuplicateTemplate(id uuid.UUID, newName string, createdByID *uuid.UUID) (*models.OnboardingTemplate, error) {
	original, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Create new template
	newTemplate := &models.OnboardingTemplate{
		CompanyID:     original.CompanyID,
		Name:          newName,
		Description:   original.Description,
		DepartmentID:  original.DepartmentID,
		PositionLevel: original.PositionLevel,
		CollarType:    original.CollarType,
		EstimatedDays: original.EstimatedDays,
		IsDefault:     false, // Never default for duplicates
		IsActive:      true,
		CreatedByID:   createdByID,
	}

	if err := s.db.Create(newTemplate).Error; err != nil {
		return nil, err
	}

	// Duplicate task templates
	for _, taskTemplate := range original.TaskTemplates {
		newTaskTemplate := models.OnboardingTaskTemplate{
			TemplateID:       newTemplate.ID,
			Title:            taskTemplate.Title,
			Description:      taskTemplate.Description,
			TaskType:         taskTemplate.TaskType,
			DueAfterDays:     taskTemplate.DueAfterDays,
			DisplayOrder:     taskTemplate.DisplayOrder,
			AssigneeRole:     taskTemplate.AssigneeRole,
			AssigneeUserID:   taskTemplate.AssigneeUserID,
			NotifyOnOverdue:  taskTemplate.NotifyOnOverdue,
			IsRequired:       taskTemplate.IsRequired,
			RequiresApproval: taskTemplate.RequiresApproval,
			ApproverRole:     taskTemplate.ApproverRole,
			DocumentURL:      taskTemplate.DocumentURL,
			FormFields:       taskTemplate.FormFields,
			IsActive:         taskTemplate.IsActive,
		}
		s.db.Create(&newTaskTemplate)
	}

	return s.GetByID(newTemplate.ID)
}

// === Task Template Methods ===

// CreateTaskTemplate creates a new task template
func (s *OnboardingTemplateService) CreateTaskTemplate(dto CreateTaskTemplateDTO) (*models.OnboardingTaskTemplate, error) {
	if dto.Title == "" {
		return nil, errors.New("task title is required")
	}

	taskTemplate := &models.OnboardingTaskTemplate{
		TemplateID:       dto.TemplateID,
		Title:            dto.Title,
		Description:      dto.Description,
		TaskType:         dto.TaskType,
		DueAfterDays:     dto.DueAfterDays,
		DisplayOrder:     dto.DisplayOrder,
		AssigneeRole:     dto.AssigneeRole,
		AssigneeUserID:   dto.AssigneeUserID,
		NotifyOnOverdue:  dto.NotifyOnOverdue,
		IsRequired:       dto.IsRequired,
		RequiresApproval: dto.RequiresApproval,
		ApproverRole:     dto.ApproverRole,
		DocumentURL:      dto.DocumentURL,
		FormFields:       dto.FormFields,
		DependsOnTaskID:  dto.DependsOnTaskID,
		IsActive:         true,
	}

	if err := s.db.Create(taskTemplate).Error; err != nil {
		return nil, err
	}

	return taskTemplate, nil
}

// GetTaskTemplateByID retrieves a task template by ID
func (s *OnboardingTemplateService) GetTaskTemplateByID(id uuid.UUID) (*models.OnboardingTaskTemplate, error) {
	var taskTemplate models.OnboardingTaskTemplate
	err := s.db.First(&taskTemplate, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task template not found")
		}
		return nil, err
	}
	return &taskTemplate, nil
}

// UpdateTaskTemplate updates a task template
func (s *OnboardingTemplateService) UpdateTaskTemplate(id uuid.UUID, dto UpdateTaskTemplateDTO) (*models.OnboardingTaskTemplate, error) {
	taskTemplate, err := s.GetTaskTemplateByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Title != nil {
		taskTemplate.Title = *dto.Title
	}
	if dto.Description != nil {
		taskTemplate.Description = *dto.Description
	}
	if dto.TaskType != nil {
		taskTemplate.TaskType = *dto.TaskType
	}
	if dto.DueAfterDays != nil {
		taskTemplate.DueAfterDays = *dto.DueAfterDays
	}
	if dto.DisplayOrder != nil {
		taskTemplate.DisplayOrder = *dto.DisplayOrder
	}
	if dto.AssigneeRole != nil {
		taskTemplate.AssigneeRole = *dto.AssigneeRole
	}
	if dto.AssigneeUserID != nil {
		taskTemplate.AssigneeUserID = dto.AssigneeUserID
	}
	if dto.NotifyOnOverdue != nil {
		taskTemplate.NotifyOnOverdue = *dto.NotifyOnOverdue
	}
	if dto.IsRequired != nil {
		taskTemplate.IsRequired = *dto.IsRequired
	}
	if dto.RequiresApproval != nil {
		taskTemplate.RequiresApproval = *dto.RequiresApproval
	}
	if dto.ApproverRole != nil {
		taskTemplate.ApproverRole = *dto.ApproverRole
	}
	if dto.DocumentURL != nil {
		taskTemplate.DocumentURL = *dto.DocumentURL
	}
	if dto.FormFields != nil {
		taskTemplate.FormFields = dto.FormFields
	}
	if dto.DependsOnTaskID != nil {
		taskTemplate.DependsOnTaskID = dto.DependsOnTaskID
	}
	if dto.IsActive != nil {
		taskTemplate.IsActive = *dto.IsActive
	}

	if err := s.db.Save(taskTemplate).Error; err != nil {
		return nil, err
	}

	return taskTemplate, nil
}

// DeleteTaskTemplate soft-deletes a task template
func (s *OnboardingTemplateService) DeleteTaskTemplate(id uuid.UUID) error {
	result := s.db.Delete(&models.OnboardingTaskTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task template not found")
	}
	return nil
}

// ReorderTaskTemplates reorders task templates within a template
func (s *OnboardingTemplateService) ReorderTaskTemplates(templateID uuid.UUID, taskOrder []uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, taskID := range taskOrder {
			if err := tx.Model(&models.OnboardingTaskTemplate{}).
				Where("id = ? AND template_id = ?", taskID, templateID).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
