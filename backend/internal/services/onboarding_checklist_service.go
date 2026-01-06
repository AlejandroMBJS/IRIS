/*
Package services - IRIS Payroll System Onboarding Checklist Service

==============================================================================
FILE: internal/services/onboarding_checklist_service.go
==============================================================================

DESCRIPTION:
    Provides business logic for managing employee onboarding checklists and tasks.
    Handles creating checklists from templates, task completion, and progress tracking.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OnboardingChecklistService provides business logic for onboarding checklists
type OnboardingChecklistService struct {
	db *gorm.DB
}

// NewOnboardingChecklistService creates a new OnboardingChecklistService
func NewOnboardingChecklistService(db *gorm.DB) *OnboardingChecklistService {
	return &OnboardingChecklistService{db: db}
}

// CreateChecklistDTO contains data for creating an onboarding checklist
type CreateChecklistDTO struct {
	CompanyID       uuid.UUID
	EmployeeID      uuid.UUID
	TemplateID      *uuid.UUID
	Title           string
	Description     string
	StartDate       time.Time
	TargetEndDate   time.Time
	HRContactID     *uuid.UUID
	BuddyEmployeeID *uuid.UUID
	Notes           string
	CreatedByID     *uuid.UUID
}

// UpdateChecklistDTO contains data for updating a checklist
type UpdateChecklistDTO struct {
	Title           *string
	Description     *string
	TargetEndDate   *time.Time
	HRContactID     *uuid.UUID
	BuddyEmployeeID *uuid.UUID
	Notes           *string
}

// CreateTaskDTO contains data for creating a task
type CreateTaskDTO struct {
	ChecklistID      uuid.UUID
	Title            string
	Description      string
	TaskType         models.OnboardingTaskType
	DueDate          *time.Time
	DisplayOrder     int
	AssigneeID       *uuid.UUID
	AssigneeRole     string
	IsRequired       bool
	RequiresApproval bool
	ApproverRole     string
	DocumentURL      string
	DependsOnTaskID  *uuid.UUID
}

// CompleteTaskDTO contains data for completing a task
type CompleteTaskDTO struct {
	CompletedByID   uuid.UUID
	CompletionNotes string
	UploadedFileID  *uuid.UUID
	FormData        string
}

// ApproveTaskDTO contains data for approving a task
type ApproveTaskDTO struct {
	ApprovedByID  uuid.UUID
	ApprovalNotes string
}

// ChecklistFilters contains filters for listing checklists
type ChecklistFilters struct {
	CompanyID   uuid.UUID
	EmployeeID  *uuid.UUID
	Status      string
	HRContactID *uuid.UUID
	IsOverdue   *bool
	Search      string
	Page        int
	Limit       int
	SortBy      string
	SortOrder   string
}

// PaginatedChecklists contains paginated checklist results
type PaginatedChecklists struct {
	Data       []models.OnboardingChecklist `json:"data"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"page_size"`
	TotalPages int                          `json:"total_pages"`
}

// ChecklistStats contains statistics for onboarding
type ChecklistStats struct {
	TotalActive       int64            `json:"total_active"`
	TotalCompleted    int64            `json:"total_completed"`
	TotalOverdue      int64            `json:"total_overdue"`
	AvgCompletionDays float64          `json:"avg_completion_days"`
	ByStatus          map[string]int64 `json:"by_status"`
}

// Create creates a new onboarding checklist
func (s *OnboardingChecklistService) Create(dto CreateChecklistDTO) (*models.OnboardingChecklist, error) {
	// Verify employee exists
	var employee models.Employee
	if err := s.db.First(&employee, "id = ?", dto.EmployeeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		return nil, err
	}

	// Check if employee already has an active checklist
	var existingCount int64
	s.db.Model(&models.OnboardingChecklist{}).
		Where("employee_id = ? AND status IN ?", dto.EmployeeID, []string{"not_started", "in_progress"}).
		Count(&existingCount)
	if existingCount > 0 {
		return nil, errors.New("employee already has an active onboarding checklist")
	}

	checklist := &models.OnboardingChecklist{
		CompanyID:       dto.CompanyID,
		EmployeeID:      dto.EmployeeID,
		TemplateID:      dto.TemplateID,
		Title:           dto.Title,
		Description:     dto.Description,
		StartDate:       dto.StartDate,
		TargetEndDate:   dto.TargetEndDate,
		HRContactID:     dto.HRContactID,
		BuddyEmployeeID: dto.BuddyEmployeeID,
		Notes:           dto.Notes,
		Status:          models.OnboardingChecklistStatusNotStarted,
		ProgressPercent: 0,
		CompletedTasks:  0,
		TotalTasks:      0,
		CreatedByID:     dto.CreatedByID,
	}

	if checklist.Title == "" {
		checklist.Title = "Onboarding - " + employee.FirstName + " " + employee.LastName
	}

	// Use transaction to create checklist and tasks from template
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(checklist).Error; err != nil {
			return err
		}

		// If template provided, create tasks from template
		if dto.TemplateID != nil {
			var template models.OnboardingTemplate
			if err := tx.Preload("TaskTemplates", "is_active = ?", true).First(&template, "id = ?", *dto.TemplateID).Error; err != nil {
				return errors.New("template not found")
			}

			for _, taskTemplate := range template.TaskTemplates {
				var dueDate *time.Time
				if taskTemplate.DueAfterDays > 0 {
					d := dto.StartDate.AddDate(0, 0, taskTemplate.DueAfterDays)
					dueDate = &d
				}

				task := models.OnboardingTask{
					ChecklistID:      checklist.ID,
					TaskTemplateID:   &taskTemplate.ID,
					Title:            taskTemplate.Title,
					Description:      taskTemplate.Description,
					TaskType:         taskTemplate.TaskType,
					DueDate:          dueDate,
					DisplayOrder:     taskTemplate.DisplayOrder,
					Status:           models.OnboardingTaskStatusPending,
					AssigneeRole:     taskTemplate.AssigneeRole,
					IsRequired:       taskTemplate.IsRequired,
					RequiresApproval: taskTemplate.RequiresApproval,
					ApproverRole:     taskTemplate.ApproverRole,
					DocumentURL:      taskTemplate.DocumentURL,
					DependsOnTaskID:  taskTemplate.DependsOnTaskID,
				}

				// Assign to specific user if specified
				if taskTemplate.AssigneeUserID != nil {
					task.AssigneeID = taskTemplate.AssigneeUserID
				}

				if err := tx.Create(&task).Error; err != nil {
					return err
				}
			}

			// Update task count
			var taskCount int64
			tx.Model(&models.OnboardingTask{}).Where("checklist_id = ?", checklist.ID).Count(&taskCount)
			checklist.TotalTasks = int(taskCount)
			tx.Save(checklist)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetByID(checklist.ID)
}

// GetByID retrieves a checklist by ID with tasks
func (s *OnboardingChecklistService) GetByID(id uuid.UUID) (*models.OnboardingChecklist, error) {
	var checklist models.OnboardingChecklist
	err := s.db.Preload("Employee").
		Preload("HRContact").
		Preload("BuddyEmployee").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Tasks.Assignee").
		Preload("Tasks.CompletedBy").
		Preload("Tasks.ApprovedBy").
		First(&checklist, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("checklist not found")
		}
		return nil, err
	}
	return &checklist, nil
}

// GetByEmployeeID retrieves the active checklist for an employee
func (s *OnboardingChecklistService) GetByEmployeeID(employeeID uuid.UUID) (*models.OnboardingChecklist, error) {
	var checklist models.OnboardingChecklist
	err := s.db.Where("employee_id = ? AND status IN ?", employeeID, []string{"not_started", "in_progress"}).
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		First(&checklist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &checklist, nil
}

// Update updates a checklist
func (s *OnboardingChecklistService) Update(id uuid.UUID, dto UpdateChecklistDTO) (*models.OnboardingChecklist, error) {
	checklist, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Title != nil {
		checklist.Title = *dto.Title
	}
	if dto.Description != nil {
		checklist.Description = *dto.Description
	}
	if dto.TargetEndDate != nil {
		checklist.TargetEndDate = *dto.TargetEndDate
	}
	if dto.HRContactID != nil {
		checklist.HRContactID = dto.HRContactID
	}
	if dto.BuddyEmployeeID != nil {
		checklist.BuddyEmployeeID = dto.BuddyEmployeeID
	}
	if dto.Notes != nil {
		checklist.Notes = *dto.Notes
	}

	if err := s.db.Save(checklist).Error; err != nil {
		return nil, err
	}

	return checklist, nil
}

// Delete soft-deletes a checklist
func (s *OnboardingChecklistService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.OnboardingChecklist{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("checklist not found")
	}
	return nil
}

// List retrieves checklists with filters
func (s *OnboardingChecklistService) List(filters ChecklistFilters) (*PaginatedChecklists, error) {
	query := s.db.Model(&models.OnboardingChecklist{}).Where("company_id = ?", filters.CompanyID)

	if filters.EmployeeID != nil {
		query = query.Where("employee_id = ?", *filters.EmployeeID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.HRContactID != nil {
		query = query.Where("hr_contact_id = ?", *filters.HRContactID)
	}
	if filters.IsOverdue != nil && *filters.IsOverdue {
		query = query.Where("target_end_date < ? AND status IN ?", time.Now(), []string{"not_started", "in_progress"})
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Joins("LEFT JOIN employees ON employees.id = onboarding_checklists.employee_id").
			Where("employees.first_name LIKE ? OR employees.last_name LIKE ? OR onboarding_checklists.title LIKE ?",
				searchTerm, searchTerm, searchTerm)
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

	sortBy := "start_date"
	sortOrder := "DESC"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}
	query = query.Order(sortBy + " " + sortOrder)

	var checklists []models.OnboardingChecklist
	if err := query.Preload("Employee").Preload("Tasks").Find(&checklists).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedChecklists{
		Data:       checklists,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// === Task Methods ===

// AddTask adds a task to a checklist
func (s *OnboardingChecklistService) AddTask(dto CreateTaskDTO) (*models.OnboardingTask, error) {
	// Verify checklist exists and is active
	var checklist models.OnboardingChecklist
	if err := s.db.First(&checklist, "id = ?", dto.ChecklistID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("checklist not found")
		}
		return nil, err
	}

	if checklist.Status == models.OnboardingChecklistStatusCompleted || checklist.Status == models.OnboardingChecklistStatusCancelled {
		return nil, errors.New("cannot add tasks to a completed or cancelled checklist")
	}

	task := &models.OnboardingTask{
		ChecklistID:      dto.ChecklistID,
		Title:            dto.Title,
		Description:      dto.Description,
		TaskType:         dto.TaskType,
		DueDate:          dto.DueDate,
		DisplayOrder:     dto.DisplayOrder,
		Status:           models.OnboardingTaskStatusPending,
		AssigneeID:       dto.AssigneeID,
		AssigneeRole:     dto.AssigneeRole,
		IsRequired:       dto.IsRequired,
		RequiresApproval: dto.RequiresApproval,
		ApproverRole:     dto.ApproverRole,
		DocumentURL:      dto.DocumentURL,
		DependsOnTaskID:  dto.DependsOnTaskID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	// Update checklist task count
	s.updateChecklistProgress(dto.ChecklistID)

	return task, nil
}

// GetTaskByID retrieves a task by ID
func (s *OnboardingChecklistService) GetTaskByID(id uuid.UUID) (*models.OnboardingTask, error) {
	var task models.OnboardingTask
	err := s.db.Preload("Checklist").Preload("Assignee").Preload("CompletedBy").Preload("ApprovedBy").
		First(&task, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, err
	}
	return &task, nil
}

// CompleteTask marks a task as completed
func (s *OnboardingChecklistService) CompleteTask(id uuid.UUID, dto CompleteTaskDTO) (*models.OnboardingTask, error) {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	if task.Status == models.OnboardingTaskStatusCompleted {
		return nil, errors.New("task is already completed")
	}

	// Check dependencies
	if task.DependsOnTaskID != nil {
		var dependentTask models.OnboardingTask
		if err := s.db.First(&dependentTask, "id = ?", *task.DependsOnTaskID).Error; err == nil {
			if dependentTask.Status != models.OnboardingTaskStatusCompleted && dependentTask.Status != models.OnboardingTaskStatusSkipped {
				return nil, errors.New("dependent task must be completed first")
			}
		}
	}

	now := time.Now()
	task.CompletedByID = &dto.CompletedByID
	task.CompletedAt = &now
	task.CompletionNotes = dto.CompletionNotes

	if dto.UploadedFileID != nil {
		task.UploadedFileID = dto.UploadedFileID
		if task.TaskType == models.OnboardingTaskTypeDocument {
			task.SignedAt = &now
		}
	}
	if dto.FormData != "" {
		task.FormData = dto.FormData
	}

	// If requires approval, mark as in progress; otherwise complete
	if task.RequiresApproval {
		task.Status = models.OnboardingTaskStatusInProgress
	} else {
		task.Status = models.OnboardingTaskStatusCompleted
	}

	if err := s.db.Save(task).Error; err != nil {
		return nil, err
	}

	// Update checklist progress
	s.updateChecklistProgress(task.ChecklistID)

	return task, nil
}

// ApproveTask approves a completed task
func (s *OnboardingChecklistService) ApproveTask(id uuid.UUID, dto ApproveTaskDTO) (*models.OnboardingTask, error) {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	if !task.RequiresApproval {
		return nil, errors.New("task does not require approval")
	}

	if task.Status != models.OnboardingTaskStatusInProgress {
		return nil, errors.New("task must be in progress to approve")
	}

	now := time.Now()
	task.Status = models.OnboardingTaskStatusCompleted
	task.ApprovedByID = &dto.ApprovedByID
	task.ApprovedAt = &now
	task.ApprovalNotes = dto.ApprovalNotes

	if err := s.db.Save(task).Error; err != nil {
		return nil, err
	}

	// Update checklist progress
	s.updateChecklistProgress(task.ChecklistID)

	return task, nil
}

// SkipTask marks a task as skipped
func (s *OnboardingChecklistService) SkipTask(id uuid.UUID, reason string) (*models.OnboardingTask, error) {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	if task.IsRequired {
		return nil, errors.New("required tasks cannot be skipped")
	}

	task.Status = models.OnboardingTaskStatusSkipped
	task.CompletionNotes = reason

	if err := s.db.Save(task).Error; err != nil {
		return nil, err
	}

	// Update checklist progress
	s.updateChecklistProgress(task.ChecklistID)

	return task, nil
}

// DeleteTask removes a task from a checklist
func (s *OnboardingChecklistService) DeleteTask(id uuid.UUID) error {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	checklistID := task.ChecklistID

	result := s.db.Delete(&models.OnboardingTask{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	// Update checklist progress
	s.updateChecklistProgress(checklistID)

	return nil
}

// updateChecklistProgress recalculates and updates checklist progress
func (s *OnboardingChecklistService) updateChecklistProgress(checklistID uuid.UUID) {
	var checklist models.OnboardingChecklist
	if err := s.db.First(&checklist, "id = ?", checklistID).Error; err != nil {
		return
	}

	var totalTasks int64
	var completedTasks int64

	s.db.Model(&models.OnboardingTask{}).Where("checklist_id = ?", checklistID).Count(&totalTasks)
	s.db.Model(&models.OnboardingTask{}).Where("checklist_id = ? AND status IN ?", checklistID,
		[]string{string(models.OnboardingTaskStatusCompleted), string(models.OnboardingTaskStatusSkipped)}).Count(&completedTasks)

	checklist.TotalTasks = int(totalTasks)
	checklist.CompletedTasks = int(completedTasks)

	if totalTasks > 0 {
		checklist.ProgressPercent = int((completedTasks * 100) / totalTasks)
	} else {
		checklist.ProgressPercent = 0
	}

	// Update status based on progress
	if checklist.ProgressPercent == 0 && checklist.Status == models.OnboardingChecklistStatusNotStarted {
		// Status remains not_started
	} else if checklist.ProgressPercent == 100 {
		checklist.Status = models.OnboardingChecklistStatusCompleted
		now := time.Now()
		checklist.ActualEndDate = &now
	} else if checklist.Status == models.OnboardingChecklistStatusNotStarted {
		checklist.Status = models.OnboardingChecklistStatusInProgress
	}

	s.db.Save(&checklist)
}

// CancelChecklist cancels an onboarding checklist
func (s *OnboardingChecklistService) CancelChecklist(id uuid.UUID, reason string) (*models.OnboardingChecklist, error) {
	checklist, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if checklist.Status == models.OnboardingChecklistStatusCompleted {
		return nil, errors.New("cannot cancel a completed checklist")
	}

	checklist.Status = models.OnboardingChecklistStatusCancelled
	checklist.Notes = reason

	if err := s.db.Save(checklist).Error; err != nil {
		return nil, err
	}

	return checklist, nil
}

// GetStats returns statistics for onboarding
func (s *OnboardingChecklistService) GetStats(companyID uuid.UUID) (*ChecklistStats, error) {
	stats := &ChecklistStats{
		ByStatus: make(map[string]int64),
	}

	// Active checklists
	s.db.Model(&models.OnboardingChecklist{}).
		Where("company_id = ? AND status = ?", companyID, models.OnboardingChecklistStatusInProgress).
		Count(&stats.TotalActive)

	// Completed checklists
	s.db.Model(&models.OnboardingChecklist{}).
		Where("company_id = ? AND status = ?", companyID, models.OnboardingChecklistStatusCompleted).
		Count(&stats.TotalCompleted)

	// Overdue checklists
	s.db.Model(&models.OnboardingChecklist{}).
		Where("company_id = ? AND target_end_date < ? AND status IN ?", companyID, time.Now(), []string{"not_started", "in_progress"}).
		Count(&stats.TotalOverdue)

	// Count by status
	var statusResults []struct {
		Status string
		Count  int64
	}
	s.db.Model(&models.OnboardingChecklist{}).
		Select("status, count(*) as count").
		Where("company_id = ?", companyID).
		Group("status").
		Scan(&statusResults)
	for _, r := range statusResults {
		stats.ByStatus[r.Status] = r.Count
	}

	// Average completion days
	s.db.Model(&models.OnboardingChecklist{}).
		Select("COALESCE(AVG(JULIANDAY(actual_end_date) - JULIANDAY(start_date)), 0)").
		Where("company_id = ? AND status = ?", companyID, models.OnboardingChecklistStatusCompleted).
		Scan(&stats.AvgCompletionDays)

	return stats, nil
}

// GetOverdueTasks returns all overdue tasks for a company
func (s *OnboardingChecklistService) GetOverdueTasks(companyID uuid.UUID) ([]models.OnboardingTask, error) {
	var tasks []models.OnboardingTask
	err := s.db.Joins("JOIN onboarding_checklists ON onboarding_checklists.id = onboarding_tasks.checklist_id").
		Where("onboarding_checklists.company_id = ?", companyID).
		Where("onboarding_tasks.due_date < ?", time.Now()).
		Where("onboarding_tasks.status NOT IN ?", []string{
			string(models.OnboardingTaskStatusCompleted),
			string(models.OnboardingTaskStatusSkipped),
		}).
		Preload("Checklist").
		Preload("Checklist.Employee").
		Preload("Assignee").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// AddNote adds a note to a checklist
func (s *OnboardingChecklistService) AddNote(checklistID uuid.UUID, taskID *uuid.UUID, content string, noteType string, isInternal bool, authorID uuid.UUID) (*models.OnboardingNote, error) {
	note := &models.OnboardingNote{
		ChecklistID: checklistID,
		TaskID:      taskID,
		Content:     content,
		NoteType:    noteType,
		IsInternal:  isInternal,
		AuthorID:    authorID,
	}

	if err := s.db.Create(note).Error; err != nil {
		return nil, err
	}

	return note, nil
}

// GetNotes retrieves notes for a checklist
func (s *OnboardingChecklistService) GetNotes(checklistID uuid.UUID, includeInternal bool) ([]models.OnboardingNote, error) {
	query := s.db.Where("checklist_id = ?", checklistID)
	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}

	var notes []models.OnboardingNote
	if err := query.Preload("Author").Order("created_at DESC").Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

// BulkAssignTasks assigns tasks to users in bulk
func (s *OnboardingChecklistService) BulkAssignTasks(taskIDs []uuid.UUID, assigneeID uuid.UUID) error {
	return s.db.Model(&models.OnboardingTask{}).
		Where("id IN ?", taskIDs).
		Update("assignee_id", assigneeID).Error
}

// GetEmployeeOnboardingHistory retrieves all onboarding checklists for an employee
func (s *OnboardingChecklistService) GetEmployeeOnboardingHistory(employeeID uuid.UUID) ([]models.OnboardingChecklist, error) {
	var checklists []models.OnboardingChecklist
	err := s.db.Where("employee_id = ?", employeeID).
		Preload("Tasks").
		Order("start_date DESC").
		Find(&checklists).Error
	if err != nil {
		return nil, err
	}
	return checklists, nil
}
