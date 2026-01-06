/*
Package models - IRIS Payroll System Time Tracking Module Models

==============================================================================
FILE: internal/models/timetracking.go
==============================================================================

DESCRIPTION:
    Data models for Time Tracking including timesheets, time entries,
    projects, and time-off tracking.

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TimesheetStatus represents the status of a timesheet
type TimesheetStatus string

const (
	TimesheetStatusDraft     TimesheetStatus = "draft"
	TimesheetStatusSubmitted TimesheetStatus = "submitted"
	TimesheetStatusApproved  TimesheetStatus = "approved"
	TimesheetStatusRejected  TimesheetStatus = "rejected"
	TimesheetStatusProcessed TimesheetStatus = "processed"
)

// TimeEntryType represents the type of time entry
type TimeEntryType string

const (
	TimeEntryTypeRegular  TimeEntryType = "regular"
	TimeEntryTypeOvertime TimeEntryType = "overtime"
	TimeEntryTypeBreak    TimeEntryType = "break"
	TimeEntryTypePTO      TimeEntryType = "pto"
	TimeEntryTypeSick     TimeEntryType = "sick"
	TimeEntryTypeHoliday  TimeEntryType = "holiday"
	TimeEntryTypeTraining TimeEntryType = "training"
)

// Project represents a project for time tracking
type Project struct {
	BaseModel
	CompanyID    uuid.UUID      `gorm:"type:text;not null;index" json:"company_id"`

	// Project Info
	Name         string         `gorm:"size:255;not null" json:"name"`
	Code         string         `gorm:"size:50;uniqueIndex" json:"code"`
	Description  string         `gorm:"type:text" json:"description"`

	// Client
	ClientName   string         `gorm:"size:255" json:"client_name"`
	ClientCode   string         `gorm:"size:100" json:"client_code"`

	// Classification
	DepartmentID *uuid.UUID     `gorm:"type:text" json:"department_id,omitempty"`
	CostCenterID *uuid.UUID     `gorm:"type:text" json:"cost_center_id,omitempty"`

	// Dates
	StartDate    *time.Time     `json:"start_date,omitempty"`
	EndDate      *time.Time     `json:"end_date,omitempty"`

	// Budget
	BudgetHours  float64        `gorm:"default:0" json:"budget_hours"`
	BudgetAmount float64        `gorm:"default:0" json:"budget_amount"`

	// Status
	Status       string         `gorm:"size:50;default:'active'" json:"status"` // active, on_hold, completed, cancelled
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	IsBillable   bool           `gorm:"default:true" json:"is_billable"`

	// Manager
	ProjectManagerID *uuid.UUID `gorm:"type:text" json:"project_manager_id,omitempty"`

	// Settings
	RequireTask  bool           `gorm:"default:false" json:"require_task"`
	RequireNotes bool           `gorm:"default:false" json:"require_notes"`

	CreatedByID  *uuid.UUID     `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Tasks        []ProjectTask  `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	Members      []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
}

func (Project) TableName() string {
	return "time_projects"
}

// ProjectTask represents a task within a project
type ProjectTask struct {
	BaseModel
	ProjectID    uuid.UUID  `gorm:"type:text;not null;index" json:"project_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	Code         string     `gorm:"size:50" json:"code"`

	// Budget
	BudgetHours  float64    `gorm:"default:0" json:"budget_hours"`

	// Status
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsBillable   bool       `gorm:"default:true" json:"is_billable"`

	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	// Relationships
	Project      *Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (ProjectTask) TableName() string {
	return "time_project_tasks"
}

// ProjectMember represents team members assigned to a project
type ProjectMember struct {
	BaseModel
	ProjectID    uuid.UUID  `gorm:"type:text;not null;index" json:"project_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	Role         string     `gorm:"size:100" json:"role"` // developer, designer, manager, etc.
	HourlyRate   float64    `gorm:"default:0" json:"hourly_rate"`
	BudgetHours  float64    `gorm:"default:0" json:"budget_hours"`

	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`

	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CanApprove   bool       `gorm:"default:false" json:"can_approve"` // Can approve time for this project

	// Relationships
	Project      *Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Employee     *Employee  `gorm:"foreignKey:EmployeeID;constraint:OnDelete:RESTRICT" json:"employee,omitempty"`
}

func (ProjectMember) TableName() string {
	return "time_project_members"
}

// Timesheet represents a weekly/biweekly timesheet
type Timesheet struct {
	BaseModel
	CompanyID    uuid.UUID       `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID       `gorm:"type:text;not null;index" json:"employee_id"`

	// Period
	PeriodStart  time.Time       `gorm:"not null;index" json:"period_start"`
	PeriodEnd    time.Time       `gorm:"not null" json:"period_end"`
	PayrollPeriodID *uuid.UUID   `gorm:"type:text;index" json:"payroll_period_id,omitempty"`

	// Hours Summary
	TotalRegularHours   float64 `gorm:"default:0" json:"total_regular_hours"`
	TotalOvertimeHours  float64 `gorm:"default:0" json:"total_overtime_hours"`
	TotalDoubleTimeHours float64 `gorm:"default:0" json:"total_double_time_hours"`
	TotalPTOHours       float64 `gorm:"default:0" json:"total_pto_hours"`
	TotalHolidayHours   float64 `gorm:"default:0" json:"total_holiday_hours"`
	TotalHours          float64 `gorm:"default:0" json:"total_hours"`
	TotalBillableHours  float64 `gorm:"default:0" json:"total_billable_hours"`

	// Status
	Status       TimesheetStatus `gorm:"size:50;default:'draft'" json:"status"`

	// Submission
	SubmittedAt  *time.Time      `json:"submitted_at,omitempty"`
	SubmissionNotes string       `gorm:"type:text" json:"submission_notes"`

	// Approval
	ApprovedByID *uuid.UUID      `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time      `json:"approved_at,omitempty"`
	ApprovalNotes string         `gorm:"type:text" json:"approval_notes"`

	// Rejection
	RejectedByID *uuid.UUID      `gorm:"type:text" json:"rejected_by_id,omitempty"`
	RejectedAt   *time.Time      `json:"rejected_at,omitempty"`
	RejectionReason string       `gorm:"type:text" json:"rejection_reason"`

	// Processing
	ProcessedAt  *time.Time      `json:"processed_at,omitempty"`

	// Relationships
	Employee     *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Entries      []TimeEntry     `gorm:"foreignKey:TimesheetID" json:"entries,omitempty"`
}

func (Timesheet) TableName() string {
	return "time_timesheets"
}

// TimeEntry represents a single time entry
type TimeEntry struct {
	BaseModel
	CompanyID    uuid.UUID     `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID     `gorm:"type:text;not null;index" json:"employee_id"`
	TimesheetID  *uuid.UUID    `gorm:"type:text;index" json:"timesheet_id,omitempty"`

	// Date/Time
	EntryDate    time.Time     `gorm:"not null;index" json:"entry_date"`
	StartTime    *time.Time    `json:"start_time,omitempty"`
	EndTime      *time.Time    `json:"end_time,omitempty"`

	// Duration
	Hours        float64       `gorm:"default:0" json:"hours"`
	BreakMinutes int           `gorm:"default:0" json:"break_minutes"`

	// Type
	EntryType    TimeEntryType `gorm:"size:50;default:'regular'" json:"entry_type"`
	IsBillable   bool          `gorm:"default:true" json:"is_billable"`

	// Project/Task
	ProjectID    *uuid.UUID    `gorm:"type:text;index" json:"project_id,omitempty"`
	TaskID       *uuid.UUID    `gorm:"type:text" json:"task_id,omitempty"`

	// Description
	Description  string        `gorm:"type:text" json:"description"`
	Notes        string        `gorm:"type:text" json:"notes"`

	// Cost Center (if not project-based)
	CostCenterID *uuid.UUID    `gorm:"type:text" json:"cost_center_id,omitempty"`

	// Rates
	HourlyRate   float64       `gorm:"default:0" json:"hourly_rate"`
	OvertimeRate float64       `gorm:"default:0" json:"overtime_rate"`

	// Status
	Status       string        `gorm:"size:50;default:'draft'" json:"status"` // draft, submitted, approved

	// Source
	Source       string        `gorm:"size:50" json:"source"` // manual, clock_in, import

	// Lock
	IsLocked     bool          `gorm:"default:false" json:"is_locked"` // Cannot be edited after approval

	// Relationships
	Employee     *Employee     `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Timesheet    *Timesheet    `gorm:"foreignKey:TimesheetID" json:"timesheet,omitempty"`
	Project      *Project      `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Task         *ProjectTask  `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

func (TimeEntry) TableName() string {
	return "time_entries"
}

// ClockRecord represents clock in/out records
type ClockRecord struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Clock In
	ClockInTime  time.Time  `gorm:"not null" json:"clock_in_time"`
	ClockInSource string    `gorm:"size:50" json:"clock_in_source"` // app, web, terminal, manual
	ClockInLocation string  `gorm:"size:255" json:"clock_in_location"` // GPS coordinates or terminal ID
	ClockInNotes string     `gorm:"type:text" json:"clock_in_notes"`
	ClockInIP    string     `gorm:"size:50" json:"clock_in_ip"`

	// Clock Out
	ClockOutTime *time.Time `json:"clock_out_time,omitempty"`
	ClockOutSource string   `gorm:"size:50" json:"clock_out_source"`
	ClockOutLocation string `gorm:"size:255" json:"clock_out_location"`
	ClockOutNotes string    `gorm:"type:text" json:"clock_out_notes"`
	ClockOutIP   string     `gorm:"size:50" json:"clock_out_ip"`

	// Breaks
	BreakMinutes int        `gorm:"default:0" json:"break_minutes"`
	BreakStart   *time.Time `json:"break_start,omitempty"`
	BreakEnd     *time.Time `json:"break_end,omitempty"`

	// Calculated
	WorkedHours  float64    `gorm:"default:0" json:"worked_hours"`
	IsComplete   bool       `gorm:"default:false" json:"is_complete"`

	// Status
	Status       string     `gorm:"size:50;default:'active'" json:"status"` // active, completed, adjusted
	IsAdjusted   bool       `gorm:"default:false" json:"is_adjusted"`
	AdjustmentReason string `gorm:"size:255" json:"adjustment_reason"`
	AdjustedByID *uuid.UUID `gorm:"type:text" json:"adjusted_by_id,omitempty"`

	// Link to time entry
	TimeEntryID  *uuid.UUID `gorm:"type:text" json:"time_entry_id,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (ClockRecord) TableName() string {
	return "time_clock_records"
}

// OvertimeRequest represents an overtime request
type OvertimeRequest struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Request Details
	RequestDate  time.Time  `gorm:"not null" json:"request_date"`
	StartTime    time.Time  `gorm:"not null" json:"start_time"`
	EndTime      time.Time  `gorm:"not null" json:"end_time"`
	EstimatedHours float64  `gorm:"default:0" json:"estimated_hours"`
	Reason       string     `gorm:"type:text;not null" json:"reason"`

	// Project (if applicable)
	ProjectID    *uuid.UUID `gorm:"type:text" json:"project_id,omitempty"`

	// Status
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, approved, rejected, completed

	// Approval
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovalNotes string    `gorm:"type:text" json:"approval_notes"`

	// Rejection
	RejectedByID *uuid.UUID `gorm:"type:text" json:"rejected_by_id,omitempty"`
	RejectedAt   *time.Time `json:"rejected_at,omitempty"`
	RejectionReason string  `gorm:"size:255" json:"rejection_reason"`

	// Actual Hours
	ActualHours  float64    `gorm:"default:0" json:"actual_hours"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Project      *Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (OvertimeRequest) TableName() string {
	return "time_overtime_requests"
}

// TimeOffBalance represents an employee's time-off balances
type TimeOffBalance struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	Year         int        `gorm:"not null;index" json:"year"`

	// Vacation
	VacationEntitled  float64 `gorm:"default:0" json:"vacation_entitled"` // Annual entitlement
	VacationUsed      float64 `gorm:"default:0" json:"vacation_used"`
	VacationPending   float64 `gorm:"default:0" json:"vacation_pending"` // Approved but not taken
	VacationCarryOver float64 `gorm:"default:0" json:"vacation_carry_over"` // From previous year
	VacationBalance   float64 `gorm:"default:0" json:"vacation_balance"` // Available

	// Sick Leave
	SickEntitled      float64 `gorm:"default:0" json:"sick_entitled"`
	SickUsed          float64 `gorm:"default:0" json:"sick_used"`
	SickPending       float64 `gorm:"default:0" json:"sick_pending"`
	SickBalance       float64 `gorm:"default:0" json:"sick_balance"`

	// Personal Days
	PersonalEntitled  float64 `gorm:"default:0" json:"personal_entitled"`
	PersonalUsed      float64 `gorm:"default:0" json:"personal_used"`
	PersonalPending   float64 `gorm:"default:0" json:"personal_pending"`
	PersonalBalance   float64 `gorm:"default:0" json:"personal_balance"`

	// Other PTO
	OtherEntitled     float64 `gorm:"default:0" json:"other_entitled"`
	OtherUsed         float64 `gorm:"default:0" json:"other_used"`
	OtherPending      float64 `gorm:"default:0" json:"other_pending"`
	OtherBalance      float64 `gorm:"default:0" json:"other_balance"`

	// Last calculation
	LastCalculatedAt  *time.Time `json:"last_calculated_at,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (TimeOffBalance) TableName() string {
	return "time_off_balances"
}

// TimeOffAccrual represents an accrual of time-off
type TimeOffAccrual struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Accrual Info
	AccrualDate  time.Time  `gorm:"not null" json:"accrual_date"`
	TimeOffType  string     `gorm:"size:50;not null" json:"time_off_type"` // vacation, sick, personal
	Hours        float64    `gorm:"not null" json:"hours"`
	Reason       string     `gorm:"size:255" json:"reason"` // monthly_accrual, annual_grant, adjustment, etc.

	// Source
	PayrollPeriodID *uuid.UUID `gorm:"type:text" json:"payroll_period_id,omitempty"`
	IsManual     bool       `gorm:"default:false" json:"is_manual"`
	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (TimeOffAccrual) TableName() string {
	return "time_off_accruals"
}

// TimePolicy represents company time policies
type TimePolicy struct {
	BaseModel
	CompanyID    uuid.UUID      `gorm:"type:text;not null;index" json:"company_id"`

	Name         string         `gorm:"size:255;not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`

	// Work Hours
	StandardWeeklyHours float64 `gorm:"default:40" json:"standard_weekly_hours"`
	StandardDailyHours  float64 `gorm:"default:8" json:"standard_daily_hours"`

	// Overtime
	OvertimeThresholdDaily  float64 `gorm:"default:8" json:"overtime_threshold_daily"`
	OvertimeThresholdWeekly float64 `gorm:"default:40" json:"overtime_threshold_weekly"`
	OvertimeMultiplier     float64 `gorm:"default:1.5" json:"overtime_multiplier"`
	DoubleTimeThreshold    float64 `gorm:"default:12" json:"double_time_threshold"`
	DoubleTimeMultiplier   float64 `gorm:"default:2" json:"double_time_multiplier"`

	// Breaks
	RequiredBreakAfterHours float64 `gorm:"default:6" json:"required_break_after_hours"`
	MinBreakMinutes        int     `gorm:"default:30" json:"min_break_minutes"`

	// Rounding
	RoundingIntervalMinutes int    `gorm:"default:15" json:"rounding_interval_minutes"` // 0 = no rounding

	// Vacation Accrual
	VacationAccrualRate   float64 `gorm:"default:0" json:"vacation_accrual_rate"` // Hours per pay period
	VacationMaxAccrual    float64 `gorm:"default:0" json:"vacation_max_accrual"` // Cap
	VacationMaxCarryOver  float64 `gorm:"default:0" json:"vacation_max_carry_over"`

	// Sick Leave
	SickAccrualRate       float64 `gorm:"default:0" json:"sick_accrual_rate"`
	SickMaxAccrual        float64 `gorm:"default:0" json:"sick_max_accrual"`

	// Applicability
	ApplicableRoles      pq.StringArray `gorm:"type:text[]" json:"applicable_roles"`
	ApplicableCollarTypes pq.StringArray `gorm:"type:text[]" json:"applicable_collar_types"`

	IsDefault    bool           `gorm:"default:false" json:"is_default"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`

	CreatedByID  *uuid.UUID     `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (TimePolicy) TableName() string {
	return "time_policies"
}

// Holiday represents company holidays
type Holiday struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Date         time.Time  `gorm:"not null;index" json:"date"`
	Year         int        `gorm:"not null;index" json:"year"`

	// Type
	HolidayType  string     `gorm:"size:50" json:"holiday_type"` // federal, company, optional
	IsPaid       bool       `gorm:"default:true" json:"is_paid"`
	PaidHours    float64    `gorm:"default:8" json:"paid_hours"`

	// Applicability
	ApplicableCollarTypes pq.StringArray `gorm:"type:text[]" json:"applicable_collar_types"`
	ApplicableLocations   pq.StringArray `gorm:"type:text[]" json:"applicable_locations"`

	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsRecurring  bool       `gorm:"default:true" json:"is_recurring"` // Repeats annually

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (Holiday) TableName() string {
	return "time_holidays"
}
