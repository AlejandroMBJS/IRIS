/*
Package services - IRIS Time Tracking Service

==============================================================================
FILE: internal/services/timetracking_service.go
==============================================================================

DESCRIPTION:
    Business logic for Time Tracking including timesheets, time entries,
    projects, and time-off tracking.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TimeTrackingService provides business logic for time tracking
type TimeTrackingService struct {
	db *gorm.DB
}

// NewTimeTrackingService creates a new TimeTrackingService
func NewTimeTrackingService(db *gorm.DB) *TimeTrackingService {
	return &TimeTrackingService{db: db}
}

// === Project DTOs ===

type CreateProjectDTO struct {
	CompanyID        uuid.UUID
	Name             string
	Code             string
	Description      string
	ClientName       string
	ClientCode       string
	DepartmentID     *uuid.UUID
	CostCenterID     *uuid.UUID
	StartDate        *time.Time
	EndDate          *time.Time
	BudgetHours      float64
	BudgetAmount     float64
	IsBillable       bool
	ProjectManagerID *uuid.UUID
	RequireTask      bool
	RequireNotes     bool
	CreatedByID      *uuid.UUID
}

type ProjectFilters struct {
	CompanyID    uuid.UUID
	Status       string
	ClientName   string
	DepartmentID *uuid.UUID
	IsBillable   *bool
	Search       string
	Page         int
	Limit        int
}

type PaginatedProjects struct {
	Data       []models.Project `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// CreateProject creates a new project
func (s *TimeTrackingService) CreateProject(dto CreateProjectDTO) (*models.Project, error) {
	if dto.Name == "" {
		return nil, errors.New("project name is required")
	}

	project := &models.Project{
		CompanyID:        dto.CompanyID,
		Name:             dto.Name,
		Code:             dto.Code,
		Description:      dto.Description,
		ClientName:       dto.ClientName,
		ClientCode:       dto.ClientCode,
		DepartmentID:     dto.DepartmentID,
		CostCenterID:     dto.CostCenterID,
		StartDate:        dto.StartDate,
		EndDate:          dto.EndDate,
		BudgetHours:      dto.BudgetHours,
		BudgetAmount:     dto.BudgetAmount,
		IsBillable:       dto.IsBillable,
		ProjectManagerID: dto.ProjectManagerID,
		RequireTask:      dto.RequireTask,
		RequireNotes:     dto.RequireNotes,
		Status:           "active",
		IsActive:         true,
		CreatedByID:      dto.CreatedByID,
	}

	if err := s.db.Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// GetProjectByID retrieves a project by ID
func (s *TimeTrackingService) GetProjectByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("Tasks").Preload("Members").First(&project, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return &project, nil
}

// ListProjects retrieves projects with filters
func (s *TimeTrackingService) ListProjects(filters ProjectFilters) (*PaginatedProjects, error) {
	query := s.db.Model(&models.Project{}).Where("company_id = ?", filters.CompanyID)

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.ClientName != "" {
		query = query.Where("client_name LIKE ?", "%"+filters.ClientName+"%")
	}
	if filters.DepartmentID != nil {
		query = query.Where("department_id = ?", *filters.DepartmentID)
	}
	if filters.IsBillable != nil {
		query = query.Where("is_billable = ?", *filters.IsBillable)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("name LIKE ? OR code LIKE ? OR client_name LIKE ?", searchTerm, searchTerm, searchTerm)
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

	var projects []models.Project
	if err := query.Offset(offset).Limit(filters.Limit).Order("name ASC").Find(&projects).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	return &PaginatedProjects{
		Data:       projects,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// AddProjectMember adds a member to a project
func (s *TimeTrackingService) AddProjectMember(projectID, employeeID uuid.UUID, role string, hourlyRate, budgetHours float64, canApprove bool) (*models.ProjectMember, error) {
	// Check if already a member
	var existing models.ProjectMember
	err := s.db.Where("project_id = ? AND employee_id = ?", projectID, employeeID).First(&existing).Error
	if err == nil {
		return nil, errors.New("employee is already a project member")
	}

	member := &models.ProjectMember{
		ProjectID:   projectID,
		EmployeeID:  employeeID,
		Role:        role,
		HourlyRate:  hourlyRate,
		BudgetHours: budgetHours,
		CanApprove:  canApprove,
		IsActive:    true,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, err
	}

	return member, nil
}

// CreateProjectTask creates a task for a project
func (s *TimeTrackingService) CreateProjectTask(projectID uuid.UUID, name, description, code string, budgetHours float64, isBillable bool, order int) (*models.ProjectTask, error) {
	if name == "" {
		return nil, errors.New("task name is required")
	}

	task := &models.ProjectTask{
		ProjectID:    projectID,
		Name:         name,
		Description:  description,
		Code:         code,
		BudgetHours:  budgetHours,
		IsBillable:   isBillable,
		DisplayOrder: order,
		IsActive:     true,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// === Timesheet Methods ===

type CreateTimesheetDTO struct {
	CompanyID       uuid.UUID
	EmployeeID      uuid.UUID
	PeriodStart     time.Time
	PeriodEnd       time.Time
	PayrollPeriodID *uuid.UUID
}

// CreateTimesheet creates a new timesheet
func (s *TimeTrackingService) CreateTimesheet(dto CreateTimesheetDTO) (*models.Timesheet, error) {
	// Check if timesheet already exists for this period
	var existing models.Timesheet
	err := s.db.Where("employee_id = ? AND period_start = ? AND period_end = ?",
		dto.EmployeeID, dto.PeriodStart, dto.PeriodEnd).First(&existing).Error
	if err == nil {
		return &existing, nil
	}

	timesheet := &models.Timesheet{
		CompanyID:       dto.CompanyID,
		EmployeeID:      dto.EmployeeID,
		PeriodStart:     dto.PeriodStart,
		PeriodEnd:       dto.PeriodEnd,
		PayrollPeriodID: dto.PayrollPeriodID,
		Status:          models.TimesheetStatusDraft,
	}

	if err := s.db.Create(timesheet).Error; err != nil {
		return nil, err
	}

	return timesheet, nil
}

// GetTimesheetByID retrieves a timesheet by ID
func (s *TimeTrackingService) GetTimesheetByID(id uuid.UUID) (*models.Timesheet, error) {
	var timesheet models.Timesheet
	err := s.db.Preload("Entries").Preload("Employee").First(&timesheet, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("timesheet not found")
		}
		return nil, err
	}
	return &timesheet, nil
}

// GetEmployeeTimesheets gets timesheets for an employee
func (s *TimeTrackingService) GetEmployeeTimesheets(employeeID uuid.UUID, status string) ([]models.Timesheet, error) {
	query := s.db.Where("employee_id = ?", employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var timesheets []models.Timesheet
	if err := query.Order("period_start DESC").Find(&timesheets).Error; err != nil {
		return nil, err
	}
	return timesheets, nil
}

// SubmitTimesheet submits a timesheet for approval
func (s *TimeTrackingService) SubmitTimesheet(id uuid.UUID, notes string) (*models.Timesheet, error) {
	timesheet, err := s.GetTimesheetByID(id)
	if err != nil {
		return nil, err
	}

	if timesheet.Status != models.TimesheetStatusDraft {
		return nil, errors.New("timesheet is not in draft status")
	}

	// Calculate totals
	s.calculateTimesheetTotals(timesheet)

	now := time.Now()
	timesheet.Status = models.TimesheetStatusSubmitted
	timesheet.SubmittedAt = &now
	timesheet.SubmissionNotes = notes

	if err := s.db.Save(timesheet).Error; err != nil {
		return nil, err
	}

	return timesheet, nil
}

// ApproveTimesheet approves a timesheet
func (s *TimeTrackingService) ApproveTimesheet(id uuid.UUID, approverID uuid.UUID, notes string) (*models.Timesheet, error) {
	timesheet, err := s.GetTimesheetByID(id)
	if err != nil {
		return nil, err
	}

	if timesheet.Status != models.TimesheetStatusSubmitted {
		return nil, errors.New("timesheet is not submitted")
	}

	now := time.Now()
	timesheet.Status = models.TimesheetStatusApproved
	timesheet.ApprovedByID = &approverID
	timesheet.ApprovedAt = &now
	timesheet.ApprovalNotes = notes

	// Lock all entries
	s.db.Model(&models.TimeEntry{}).Where("timesheet_id = ?", id).Update("is_locked", true)

	if err := s.db.Save(timesheet).Error; err != nil {
		return nil, err
	}

	return timesheet, nil
}

// RejectTimesheet rejects a timesheet
func (s *TimeTrackingService) RejectTimesheet(id uuid.UUID, rejectorID uuid.UUID, reason string) (*models.Timesheet, error) {
	timesheet, err := s.GetTimesheetByID(id)
	if err != nil {
		return nil, err
	}

	if timesheet.Status != models.TimesheetStatusSubmitted {
		return nil, errors.New("timesheet is not submitted")
	}

	now := time.Now()
	timesheet.Status = models.TimesheetStatusRejected
	timesheet.RejectedByID = &rejectorID
	timesheet.RejectedAt = &now
	timesheet.RejectionReason = reason

	if err := s.db.Save(timesheet).Error; err != nil {
		return nil, err
	}

	return timesheet, nil
}

func (s *TimeTrackingService) calculateTimesheetTotals(timesheet *models.Timesheet) {
	var entries []models.TimeEntry
	s.db.Where("timesheet_id = ?", timesheet.ID).Find(&entries)

	var regular, overtime, doubleTime, pto, holiday, billable, total float64

	for _, entry := range entries {
		switch entry.EntryType {
		case models.TimeEntryTypeRegular:
			regular += entry.Hours
		case models.TimeEntryTypeOvertime:
			overtime += entry.Hours
		case models.TimeEntryTypePTO:
			pto += entry.Hours
		case models.TimeEntryTypeSick:
			pto += entry.Hours
		case models.TimeEntryTypeHoliday:
			holiday += entry.Hours
		}
		total += entry.Hours
		if entry.IsBillable {
			billable += entry.Hours
		}
	}

	timesheet.TotalRegularHours = regular
	timesheet.TotalOvertimeHours = overtime
	timesheet.TotalDoubleTimeHours = doubleTime
	timesheet.TotalPTOHours = pto
	timesheet.TotalHolidayHours = holiday
	timesheet.TotalBillableHours = billable
	timesheet.TotalHours = total
}

// === Time Entry Methods ===

type CreateTimeEntryDTO struct {
	CompanyID    uuid.UUID
	EmployeeID   uuid.UUID
	TimesheetID  *uuid.UUID
	EntryDate    time.Time
	StartTime    *time.Time
	EndTime      *time.Time
	Hours        float64
	BreakMinutes int
	EntryType    models.TimeEntryType
	IsBillable   bool
	ProjectID    *uuid.UUID
	TaskID       *uuid.UUID
	Description  string
	Notes        string
	CostCenterID *uuid.UUID
	Source       string
}

// CreateTimeEntry creates a time entry
func (s *TimeTrackingService) CreateTimeEntry(dto CreateTimeEntryDTO) (*models.TimeEntry, error) {
	if dto.Hours <= 0 && dto.StartTime == nil {
		return nil, errors.New("hours or start/end time is required")
	}

	entry := &models.TimeEntry{
		CompanyID:    dto.CompanyID,
		EmployeeID:   dto.EmployeeID,
		TimesheetID:  dto.TimesheetID,
		EntryDate:    dto.EntryDate,
		StartTime:    dto.StartTime,
		EndTime:      dto.EndTime,
		Hours:        dto.Hours,
		BreakMinutes: dto.BreakMinutes,
		EntryType:    dto.EntryType,
		IsBillable:   dto.IsBillable,
		ProjectID:    dto.ProjectID,
		TaskID:       dto.TaskID,
		Description:  dto.Description,
		Notes:        dto.Notes,
		CostCenterID: dto.CostCenterID,
		Source:       dto.Source,
		Status:       "draft",
	}

	if entry.EntryType == "" {
		entry.EntryType = models.TimeEntryTypeRegular
	}
	if entry.Source == "" {
		entry.Source = "manual"
	}

	// Calculate hours from start/end time if not provided
	if entry.Hours == 0 && entry.StartTime != nil && entry.EndTime != nil {
		duration := entry.EndTime.Sub(*entry.StartTime)
		entry.Hours = duration.Hours() - float64(entry.BreakMinutes)/60
	}

	if err := s.db.Create(entry).Error; err != nil {
		return nil, err
	}

	// Update timesheet totals if linked
	if entry.TimesheetID != nil {
		var timesheet models.Timesheet
		if err := s.db.First(&timesheet, "id = ?", entry.TimesheetID).Error; err == nil {
			s.calculateTimesheetTotals(&timesheet)
			s.db.Save(&timesheet)
		}
	}

	return entry, nil
}

// GetTimeEntryByID retrieves a time entry by ID
func (s *TimeTrackingService) GetTimeEntryByID(id uuid.UUID) (*models.TimeEntry, error) {
	var entry models.TimeEntry
	err := s.db.Preload("Project").Preload("Task").First(&entry, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("time entry not found")
		}
		return nil, err
	}
	return &entry, nil
}

// GetEmployeeTimeEntries gets time entries for an employee
func (s *TimeTrackingService) GetEmployeeTimeEntries(employeeID uuid.UUID, startDate, endDate time.Time) ([]models.TimeEntry, error) {
	var entries []models.TimeEntry
	if err := s.db.Where("employee_id = ? AND entry_date >= ? AND entry_date <= ?",
		employeeID, startDate, endDate).
		Preload("Project").
		Order("entry_date DESC, start_time ASC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// === Clock In/Out Methods ===

// ClockIn clocks in an employee
func (s *TimeTrackingService) ClockIn(companyID, employeeID uuid.UUID, source, location, ip string, notes string) (*models.ClockRecord, error) {
	// Check if already clocked in
	var existing models.ClockRecord
	err := s.db.Where("employee_id = ? AND is_complete = ?", employeeID, false).First(&existing).Error
	if err == nil {
		return nil, errors.New("already clocked in")
	}

	record := &models.ClockRecord{
		CompanyID:       companyID,
		EmployeeID:      employeeID,
		ClockInTime:     time.Now(),
		ClockInSource:   source,
		ClockInLocation: location,
		ClockInIP:       ip,
		ClockInNotes:    notes,
		Status:          "active",
		IsComplete:      false,
	}

	if record.ClockInSource == "" {
		record.ClockInSource = "app"
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

// ClockOut clocks out an employee
func (s *TimeTrackingService) ClockOut(employeeID uuid.UUID, source, location, ip string, notes string) (*models.ClockRecord, error) {
	var record models.ClockRecord
	err := s.db.Where("employee_id = ? AND is_complete = ?", employeeID, false).First(&record).Error
	if err != nil {
		return nil, errors.New("not clocked in")
	}

	now := time.Now()
	record.ClockOutTime = &now
	record.ClockOutSource = source
	record.ClockOutLocation = location
	record.ClockOutIP = ip
	record.ClockOutNotes = notes
	record.IsComplete = true
	record.Status = "completed"

	// Calculate worked hours
	duration := now.Sub(record.ClockInTime)
	record.WorkedHours = duration.Hours() - float64(record.BreakMinutes)/60

	if err := s.db.Save(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// StartBreak starts a break
func (s *TimeTrackingService) StartBreak(employeeID uuid.UUID) (*models.ClockRecord, error) {
	var record models.ClockRecord
	err := s.db.Where("employee_id = ? AND is_complete = ?", employeeID, false).First(&record).Error
	if err != nil {
		return nil, errors.New("not clocked in")
	}

	now := time.Now()
	record.BreakStart = &now

	if err := s.db.Save(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// EndBreak ends a break
func (s *TimeTrackingService) EndBreak(employeeID uuid.UUID) (*models.ClockRecord, error) {
	var record models.ClockRecord
	err := s.db.Where("employee_id = ? AND is_complete = ? AND break_start IS NOT NULL", employeeID, false).First(&record).Error
	if err != nil {
		return nil, errors.New("not on break")
	}

	now := time.Now()
	record.BreakEnd = &now

	// Calculate break duration
	if record.BreakStart != nil {
		breakDuration := now.Sub(*record.BreakStart)
		record.BreakMinutes += int(breakDuration.Minutes())
	}

	record.BreakStart = nil
	record.BreakEnd = nil

	if err := s.db.Save(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// GetActiveClockRecord gets the current clock record for an employee
func (s *TimeTrackingService) GetActiveClockRecord(employeeID uuid.UUID) (*models.ClockRecord, error) {
	var record models.ClockRecord
	err := s.db.Where("employee_id = ? AND is_complete = ?", employeeID, false).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// === Time Off Balance Methods ===

// GetTimeOffBalance gets time off balance for an employee
func (s *TimeTrackingService) GetTimeOffBalance(employeeID uuid.UUID, year int) (*models.TimeOffBalance, error) {
	if year == 0 {
		year = time.Now().Year()
	}

	var balance models.TimeOffBalance
	err := s.db.Where("employee_id = ? AND year = ?", employeeID, year).First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &balance, nil
}

// UpdateTimeOffBalance updates time off balance
func (s *TimeTrackingService) UpdateTimeOffBalance(employeeID uuid.UUID, year int, balanceType string, hours float64, isUsage bool) (*models.TimeOffBalance, error) {
	balance, err := s.GetTimeOffBalance(employeeID, year)
	if err != nil {
		return nil, err
	}

	// Get company ID from employee
	var employee models.Employee
	if err := s.db.First(&employee, "id = ?", employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}

	if balance == nil {
		balance = &models.TimeOffBalance{
			CompanyID:  employee.CompanyID,
			EmployeeID: employeeID,
			Year:       year,
		}
	}

	switch balanceType {
	case "vacation":
		if isUsage {
			balance.VacationUsed += hours
		} else {
			balance.VacationEntitled += hours
		}
		balance.VacationBalance = balance.VacationEntitled + balance.VacationCarryOver - balance.VacationUsed - balance.VacationPending
	case "sick":
		if isUsage {
			balance.SickUsed += hours
		} else {
			balance.SickEntitled += hours
		}
		balance.SickBalance = balance.SickEntitled - balance.SickUsed - balance.SickPending
	case "personal":
		if isUsage {
			balance.PersonalUsed += hours
		} else {
			balance.PersonalEntitled += hours
		}
		balance.PersonalBalance = balance.PersonalEntitled - balance.PersonalUsed - balance.PersonalPending
	}

	now := time.Now()
	balance.LastCalculatedAt = &now

	if balance.ID == uuid.Nil {
		if err := s.db.Create(balance).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Save(balance).Error; err != nil {
			return nil, err
		}
	}

	return balance, nil
}

// === Holiday Methods ===

// CreateHoliday creates a company holiday
func (s *TimeTrackingService) CreateHoliday(companyID uuid.UUID, name string, date time.Time, holidayType string, isPaid bool, paidHours float64, createdByID *uuid.UUID) (*models.Holiday, error) {
	if name == "" {
		return nil, errors.New("holiday name is required")
	}

	holiday := &models.Holiday{
		CompanyID:   companyID,
		Name:        name,
		Date:        date,
		Year:        date.Year(),
		HolidayType: holidayType,
		IsPaid:      isPaid,
		PaidHours:   paidHours,
		IsActive:    true,
		IsRecurring: true,
		CreatedByID: createdByID,
	}

	if holiday.HolidayType == "" {
		holiday.HolidayType = "company"
	}
	if holiday.PaidHours == 0 {
		holiday.PaidHours = 8
	}

	if err := s.db.Create(holiday).Error; err != nil {
		return nil, err
	}

	return holiday, nil
}

// GetHolidays gets holidays for a company and year
func (s *TimeTrackingService) GetHolidays(companyID uuid.UUID, year int) ([]models.Holiday, error) {
	var holidays []models.Holiday
	if err := s.db.Where("company_id = ? AND year = ? AND is_active = ?", companyID, year, true).
		Order("date ASC").Find(&holidays).Error; err != nil {
		return nil, err
	}
	return holidays, nil
}
