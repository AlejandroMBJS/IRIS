/*
Package models - IRIS Payroll System Onboarding Models

==============================================================================
FILE: internal/models/onboarding.go
==============================================================================

DESCRIPTION:
    Defines data models for the employee onboarding module including templates,
    checklists, tasks, and document requirements for new hires.

USER PERSPECTIVE:
    - HR can create reusable onboarding templates
    - New employees see their personalized onboarding checklist
    - Managers can track onboarding progress
    - Documents can be collected and signed electronically

DEVELOPER GUIDELINES:
    OK to modify:
      - Add new fields as onboarding features expand
      - Add new task types
      - Add validation rules

    DO NOT modify:
      - Primary key structure (ID field)
      - Foreign key relationships without migration

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OnboardingTaskType represents the type of onboarding task
type OnboardingTaskType string

const (
	OnboardingTaskTypeDocument    OnboardingTaskType = "document"     // Upload/sign document
	OnboardingTaskTypeForm        OnboardingTaskType = "form"         // Fill out form
	OnboardingTaskTypeTraining    OnboardingTaskType = "training"     // Complete training
	OnboardingTaskTypeMeeting     OnboardingTaskType = "meeting"      // Attend meeting
	OnboardingTaskTypeEquipment   OnboardingTaskType = "equipment"    // Receive equipment
	OnboardingTaskTypeAccess      OnboardingTaskType = "access"       // Get system access
	OnboardingTaskTypeAcknowledge OnboardingTaskType = "acknowledge"  // Acknowledge policy
	OnboardingTaskTypeCustom      OnboardingTaskType = "custom"       // Custom task
)

// OnboardingTaskStatus represents the status of a task
type OnboardingTaskStatus string

const (
	OnboardingTaskStatusPending    OnboardingTaskStatus = "pending"
	OnboardingTaskStatusInProgress OnboardingTaskStatus = "in_progress"
	OnboardingTaskStatusCompleted  OnboardingTaskStatus = "completed"
	OnboardingTaskStatusSkipped    OnboardingTaskStatus = "skipped"
	OnboardingTaskStatusOverdue    OnboardingTaskStatus = "overdue"
)

// OnboardingChecklistStatus represents the overall checklist status
type OnboardingChecklistStatus string

const (
	OnboardingChecklistStatusNotStarted OnboardingChecklistStatus = "not_started"
	OnboardingChecklistStatusInProgress OnboardingChecklistStatus = "in_progress"
	OnboardingChecklistStatusCompleted  OnboardingChecklistStatus = "completed"
	OnboardingChecklistStatusCancelled  OnboardingChecklistStatus = "cancelled"
)

// OnboardingTemplate represents a reusable onboarding checklist template
type OnboardingTemplate struct {
	BaseModel

	CompanyID uuid.UUID `gorm:"type:text;not null;index" json:"company_id"`
	Company   *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	// Template Info
	Name        string `gorm:"size:255;not null" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Target Audience
	DepartmentID  *uuid.UUID `gorm:"type:text;index" json:"department_id,omitempty"`
	PositionLevel string     `gorm:"size:50" json:"position_level,omitempty"` // entry, mid, senior, etc.
	CollarType    string     `gorm:"size:20" json:"collar_type,omitempty"`    // white_collar, blue_collar

	// Configuration
	EstimatedDays int  `gorm:"default:30" json:"estimated_days"` // Expected days to complete
	IsDefault     bool `gorm:"default:false" json:"is_default"`  // Default template for company
	IsActive      bool `gorm:"default:true" json:"is_active"`

	// Metadata
	CreatedByID *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
	CreatedBy   *User      `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// Relationships
	TaskTemplates []OnboardingTaskTemplate `gorm:"foreignKey:TemplateID" json:"task_templates,omitempty"`
}

// TableName specifies the table name for OnboardingTemplate
func (OnboardingTemplate) TableName() string {
	return "onboarding_templates"
}

// OnboardingTaskTemplate represents a task template within an onboarding template
type OnboardingTaskTemplate struct {
	BaseModel

	TemplateID uuid.UUID           `gorm:"type:text;not null;index" json:"template_id"`
	Template   *OnboardingTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`

	// Task Info
	Title       string             `gorm:"size:255;not null" json:"title"`
	Description string             `gorm:"type:text" json:"description,omitempty"`
	TaskType    OnboardingTaskType `gorm:"size:50;not null" json:"task_type"`

	// Timing
	DueAfterDays int `gorm:"default:0" json:"due_after_days"` // Days after start date
	DisplayOrder int `gorm:"default:0" json:"display_order"`

	// Assignment
	AssigneeRole    string     `gorm:"size:50" json:"assignee_role,omitempty"`       // employee, manager, hr, it
	AssigneeUserID  *uuid.UUID `gorm:"type:text" json:"assignee_user_id,omitempty"`  // Specific user
	NotifyOnOverdue bool       `gorm:"default:true" json:"notify_on_overdue"`

	// Requirements
	IsRequired       bool           `gorm:"default:true" json:"is_required"`
	RequiresApproval bool           `gorm:"default:false" json:"requires_approval"`
	ApproverRole     string         `gorm:"size:50" json:"approver_role,omitempty"` // manager, hr
	DocumentURL      string         `gorm:"size:500" json:"document_url,omitempty"` // For document type tasks
	FormFields       pq.StringArray `gorm:"type:text[]" json:"form_fields,omitempty"`

	// Dependencies
	DependsOnTaskID *uuid.UUID `gorm:"type:text" json:"depends_on_task_id,omitempty"`

	IsActive bool `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for OnboardingTaskTemplate
func (OnboardingTaskTemplate) TableName() string {
	return "onboarding_task_templates"
}

// OnboardingChecklist represents an active onboarding process for an employee
type OnboardingChecklist struct {
	BaseModel

	CompanyID  uuid.UUID `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID uuid.UUID `gorm:"type:text;not null;index" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	// Template (optional - can be created from scratch)
	TemplateID *uuid.UUID          `gorm:"type:text;index" json:"template_id,omitempty"`
	Template   *OnboardingTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`

	// Checklist Info
	Title       string `gorm:"size:255;not null" json:"title"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Timeline
	StartDate      time.Time  `gorm:"not null" json:"start_date"`
	TargetEndDate  time.Time  `gorm:"not null" json:"target_end_date"`
	ActualEndDate  *time.Time `json:"actual_end_date,omitempty"`

	// Status
	Status           OnboardingChecklistStatus `gorm:"size:50;default:'not_started'" json:"status"`
	ProgressPercent  int                       `gorm:"default:0" json:"progress_percent"`
	CompletedTasks   int                       `gorm:"default:0" json:"completed_tasks"`
	TotalTasks       int                       `gorm:"default:0" json:"total_tasks"`

	// Assigned Resources
	HRContactID      *uuid.UUID `gorm:"type:text" json:"hr_contact_id,omitempty"`
	HRContact        *User      `gorm:"foreignKey:HRContactID" json:"hr_contact,omitempty"`
	BuddyEmployeeID  *uuid.UUID `gorm:"type:text" json:"buddy_employee_id,omitempty"`
	BuddyEmployee    *Employee  `gorm:"foreignKey:BuddyEmployeeID" json:"buddy_employee,omitempty"`

	// Notes
	Notes string `gorm:"type:text" json:"notes,omitempty"`

	// Metadata
	CreatedByID *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
	CreatedBy   *User      `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// Relationships
	Tasks []OnboardingTask `gorm:"foreignKey:ChecklistID" json:"tasks,omitempty"`
}

// TableName specifies the table name for OnboardingChecklist
func (OnboardingChecklist) TableName() string {
	return "onboarding_checklists"
}

// IsOverdue checks if the checklist is past its target end date
func (c *OnboardingChecklist) IsOverdue() bool {
	if c.Status == OnboardingChecklistStatusCompleted || c.Status == OnboardingChecklistStatusCancelled {
		return false
	}
	return time.Now().After(c.TargetEndDate)
}

// OnboardingTask represents an individual task in an employee's onboarding checklist
type OnboardingTask struct {
	BaseModel

	ChecklistID uuid.UUID            `gorm:"type:text;not null;index" json:"checklist_id"`
	Checklist   *OnboardingChecklist `gorm:"foreignKey:ChecklistID" json:"checklist,omitempty"`

	// From Template (optional)
	TaskTemplateID *uuid.UUID              `gorm:"type:text" json:"task_template_id,omitempty"`
	TaskTemplate   *OnboardingTaskTemplate `gorm:"foreignKey:TaskTemplateID" json:"task_template,omitempty"`

	// Task Info
	Title       string             `gorm:"size:255;not null" json:"title"`
	Description string             `gorm:"type:text" json:"description,omitempty"`
	TaskType    OnboardingTaskType `gorm:"size:50;not null" json:"task_type"`

	// Timing
	DueDate      *time.Time `json:"due_date,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	// Status
	Status OnboardingTaskStatus `gorm:"size:50;default:'pending'" json:"status"`

	// Assignment
	AssigneeID   *uuid.UUID `gorm:"type:text" json:"assignee_id,omitempty"`
	Assignee     *User      `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	AssigneeRole string     `gorm:"size:50" json:"assignee_role,omitempty"` // employee, manager, hr, it

	// Requirements
	IsRequired       bool   `gorm:"default:true" json:"is_required"`
	RequiresApproval bool   `gorm:"default:false" json:"requires_approval"`
	ApproverRole     string `gorm:"size:50" json:"approver_role,omitempty"`

	// Completion Data
	CompletedByID   *uuid.UUID `gorm:"type:text" json:"completed_by_id,omitempty"`
	CompletedBy     *User      `gorm:"foreignKey:CompletedByID" json:"completed_by,omitempty"`
	CompletionNotes string     `gorm:"type:text" json:"completion_notes,omitempty"`

	// For document tasks
	DocumentURL       string     `gorm:"size:500" json:"document_url,omitempty"`
	UploadedFileID    *uuid.UUID `gorm:"type:text" json:"uploaded_file_id,omitempty"`
	SignedAt          *time.Time `json:"signed_at,omitempty"`

	// For form tasks
	FormData string `gorm:"type:text" json:"form_data,omitempty"` // JSON blob

	// Approval (if requires_approval)
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedBy   *User      `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovalNotes string    `gorm:"type:text" json:"approval_notes,omitempty"`

	// Dependencies
	DependsOnTaskID *uuid.UUID `gorm:"type:text" json:"depends_on_task_id,omitempty"`
}

// TableName specifies the table name for OnboardingTask
func (OnboardingTask) TableName() string {
	return "onboarding_tasks"
}

// IsOverdue checks if the task is past its due date
func (t *OnboardingTask) IsOverdue() bool {
	if t.Status == OnboardingTaskStatusCompleted || t.Status == OnboardingTaskStatusSkipped {
		return false
	}
	if t.DueDate == nil {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// CanComplete checks if the task can be marked as completed
func (t *OnboardingTask) CanComplete() bool {
	// Check if dependencies are met
	// This is a simple check - in reality you'd check the dependent task status
	return t.Status == OnboardingTaskStatusPending || t.Status == OnboardingTaskStatusInProgress
}

// OnboardingNote represents notes/comments on an onboarding process
type OnboardingNote struct {
	BaseModel

	ChecklistID uuid.UUID            `gorm:"type:text;not null;index" json:"checklist_id"`
	Checklist   *OnboardingChecklist `gorm:"foreignKey:ChecklistID" json:"checklist,omitempty"`
	TaskID      *uuid.UUID           `gorm:"type:text;index" json:"task_id,omitempty"` // Optional - can be for specific task
	Task        *OnboardingTask      `gorm:"foreignKey:TaskID" json:"task,omitempty"`

	// Note Content
	Content    string `gorm:"type:text;not null" json:"content"`
	NoteType   string `gorm:"size:50" json:"note_type,omitempty"` // comment, issue, resolution
	IsInternal bool   `gorm:"default:false" json:"is_internal"`   // HR-only note

	// Author
	AuthorID uuid.UUID `gorm:"type:text;not null" json:"author_id"`
	Author   *User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// TableName specifies the table name for OnboardingNote
func (OnboardingNote) TableName() string {
	return "onboarding_notes"
}
