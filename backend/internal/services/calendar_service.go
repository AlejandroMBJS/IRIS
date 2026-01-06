/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/calendar_service.go
==============================================================================

DESCRIPTION:
    Aggregates calendar events from multiple sources (AbsenceRequests, Incidences,
    ShiftExceptions) into a unified calendar view for HR. Supports filtering by
    collar type, employee, event type, and date range.

USER PERSPECTIVE:
    - HR users see all employee events on one calendar
    - Events are color-coded by employee
    - Can filter by collar type (white/blue/gray)
    - Can filter by specific employees
    - Can filter by event type (absence, incidence, shift)

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new event sources, filters
    ⚠️  CAUTION: Query performance with large date ranges
    ❌  DO NOT modify: Event type constants (breaks frontend)
    📝  Consider pagination for very large datasets

==============================================================================
*/
package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/dtos"
	"backend/internal/models"
)

// CalendarService handles calendar event aggregation
type CalendarService struct {
	db *gorm.DB
}

// NewCalendarService creates a new CalendarService
func NewCalendarService(db *gorm.DB) *CalendarService {
	return &CalendarService{db: db}
}

// GetCalendarEvents aggregates events from all sources within the date range
func (s *CalendarService) GetCalendarEvents(req dtos.CalendarEventRequest) (*dtos.CalendarEventsResponse, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %v", err)
	}

	// Get filtered employee IDs based on collar types and explicit employee filter
	employeeIDs, err := s.getFilteredEmployeeIDs(req.EmployeeIDs, req.CollarTypes, req.DepartmentID)
	if err != nil {
		return nil, err
	}

	var events []dtos.CalendarEventResponse
	var summary dtos.CalendarSummary

	// Determine which event types to fetch
	fetchAbsences := s.shouldFetchEventType(req.EventTypes, "absence")
	fetchIncidences := s.shouldFetchEventType(req.EventTypes, "incidence")
	fetchShiftChanges := s.shouldFetchEventType(req.EventTypes, "shift_change")

	// 1. Fetch AbsenceRequests
	if fetchAbsences {
		absenceEvents, err := s.fetchAbsenceRequests(startDate, endDate, employeeIDs, req.Status)
		if err != nil {
			return nil, fmt.Errorf("error fetching absence requests: %v", err)
		}
		events = append(events, absenceEvents...)
		summary.TotalAbsences = len(absenceEvents)
	}

	// 2. Fetch Incidences
	if fetchIncidences {
		incidenceEvents, err := s.fetchIncidences(startDate, endDate, employeeIDs, req.Status)
		if err != nil {
			return nil, fmt.Errorf("error fetching incidences: %v", err)
		}
		events = append(events, incidenceEvents...)
		summary.TotalIncidences = len(incidenceEvents)
	}

	// 3. Fetch ShiftExceptions
	if fetchShiftChanges {
		shiftEvents, err := s.fetchShiftExceptions(startDate, endDate, employeeIDs)
		if err != nil {
			return nil, fmt.Errorf("error fetching shift exceptions: %v", err)
		}
		events = append(events, shiftEvents...)
		summary.TotalShiftChanges = len(shiftEvents)
	}

	// Calculate summary statistics
	s.calculateSummary(&summary, events)

	return &dtos.CalendarEventsResponse{
		Events:     events,
		TotalCount: len(events),
		Summary:    summary,
	}, nil
}

// GetEmployeesForCalendar returns employees with auto-assigned colors
func (s *CalendarService) GetEmployeesForCalendar(collarTypes []string, departmentID string) (*dtos.CalendarEmployeesResponse, error) {
	query := s.db.Model(&models.Employee{}).
		Where("employment_status = ?", "active")

	// Filter by collar types
	if len(collarTypes) > 0 {
		query = query.Where("collar_type IN ?", collarTypes)
	}

	// Filter by department
	if departmentID != "" {
		if deptUUID, err := uuid.Parse(departmentID); err == nil {
			query = query.Where("department_id = ?", deptUUID)
		}
	}

	// Order by employee number for consistent color assignment
	query = query.Order("employee_number ASC")

	var employees []models.Employee
	if err := query.Find(&employees).Error; err != nil {
		return nil, fmt.Errorf("error fetching employees: %v", err)
	}

	// Convert to DTO with color assignment
	result := make([]dtos.CalendarEmployee, len(employees))
	for i, emp := range employees {
		fullName := emp.FirstName + " " + emp.LastName
		if emp.MotherLastName != "" {
			fullName += " " + emp.MotherLastName
		}

		result[i] = dtos.CalendarEmployee{
			ID:             emp.ID,
			EmployeeNumber: emp.EmployeeNumber,
			FullName:       fullName,
			CollarType:     emp.CollarType,
			DepartmentID:   emp.DepartmentID,
			Color:          dtos.GetColorForIndex(i),
		}
	}

	return &dtos.CalendarEmployeesResponse{
		Employees: result,
	}, nil
}

// getFilteredEmployeeIDs returns employee IDs based on filters
func (s *CalendarService) getFilteredEmployeeIDs(employeeIDs []string, collarTypes []string, departmentID string) ([]uuid.UUID, error) {
	// If specific employee IDs provided, use them directly
	if len(employeeIDs) > 0 {
		result := make([]uuid.UUID, 0, len(employeeIDs))
		for _, idStr := range employeeIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				result = append(result, id)
			}
		}
		return result, nil
	}

	// Otherwise, query based on collar types and department
	query := s.db.Model(&models.Employee{}).
		Where("employment_status = ?", "active").
		Select("id")

	if len(collarTypes) > 0 {
		query = query.Where("collar_type IN ?", collarTypes)
	}

	if departmentID != "" {
		if deptUUID, err := uuid.Parse(departmentID); err == nil {
			query = query.Where("department_id = ?", deptUUID)
		}
	}

	var ids []uuid.UUID
	if err := query.Find(&ids).Error; err != nil {
		return nil, err
	}

	return ids, nil
}

// shouldFetchEventType determines if an event type should be fetched
func (s *CalendarService) shouldFetchEventType(eventTypes []string, eventType string) bool {
	if len(eventTypes) == 0 {
		return true // Fetch all types if none specified
	}
	for _, t := range eventTypes {
		if strings.EqualFold(t, eventType) {
			return true
		}
	}
	return false
}

// fetchAbsenceRequests retrieves absence requests within the date range
func (s *CalendarService) fetchAbsenceRequests(startDate, endDate time.Time, employeeIDs []uuid.UUID, status string) ([]dtos.CalendarEventResponse, error) {
	query := s.db.Model(&models.AbsenceRequest{}).
		Preload("Employee").
		Where("(start_date <= ? AND end_date >= ?)", endDate, startDate) // Overlapping date range

	if len(employeeIDs) > 0 {
		query = query.Where("employee_id IN ?", employeeIDs)
	}

	if status != "" {
		query = query.Where("status = ?", strings.ToUpper(status))
	}

	var requests []models.AbsenceRequest
	if err := query.Find(&requests).Error; err != nil {
		return nil, err
	}

	events := make([]dtos.CalendarEventResponse, len(requests))
	for i, req := range requests {
		employeeName := ""
		employeeNumber := ""
		collarType := ""
		var departmentID *uuid.UUID

		if req.Employee != nil {
			employeeName = req.Employee.FirstName + " " + req.Employee.LastName
			employeeNumber = req.Employee.EmployeeNumber
			collarType = req.Employee.CollarType
			departmentID = req.Employee.DepartmentID
		}

		// Get Spanish label for request type
		title := dtos.GetRequestTypeLabel(string(req.RequestType)) + " - " + employeeName

		events[i] = dtos.CalendarEventResponse{
			ID:                req.ID,
			EventType:         dtos.EventTypeAbsence,
			Title:             title,
			Description:       req.Reason,
			StartDate:         req.StartDate,
			EndDate:           req.EndDate,
			AllDay:            true,
			Status:            string(req.Status),
			EmployeeID:        req.EmployeeID,
			EmployeeName:      employeeName,
			EmployeeNumber:    employeeNumber,
			CollarType:        collarType,
			DepartmentID:      departmentID,
			SourceID:          req.ID,
			SourceType:        "absence_request",
			RequestType:       string(req.RequestType),
			TotalDays:         req.TotalDays,
			ApprovalStage:     string(req.CurrentApprovalStage),
			Reason:            req.Reason,
			CreatedAt:         req.CreatedAt,
			UpdatedAt:         req.UpdatedAt,
		}
	}

	return events, nil
}

// fetchIncidences retrieves incidences within the date range
func (s *CalendarService) fetchIncidences(startDate, endDate time.Time, employeeIDs []uuid.UUID, status string) ([]dtos.CalendarEventResponse, error) {
	query := s.db.Model(&models.Incidence{}).
		Preload("Employee").
		Preload("IncidenceType").
		Where("(start_date <= ? AND end_date >= ?)", endDate, startDate)

	if len(employeeIDs) > 0 {
		query = query.Where("employee_id IN ?", employeeIDs)
	}

	if status != "" {
		query = query.Where("status = ?", strings.ToLower(status))
	}

	var incidences []models.Incidence
	if err := query.Find(&incidences).Error; err != nil {
		return nil, err
	}

	events := make([]dtos.CalendarEventResponse, len(incidences))
	for i, inc := range incidences {
		employeeName := ""
		employeeNumber := ""
		collarType := ""
		var departmentID *uuid.UUID

		if inc.Employee != nil {
			employeeName = inc.Employee.FirstName + " " + inc.Employee.LastName
			employeeNumber = inc.Employee.EmployeeNumber
			collarType = inc.Employee.CollarType
			departmentID = inc.Employee.DepartmentID
		}

		incidenceTypeName := ""
		category := ""
		effectType := ""
		if inc.IncidenceType != nil {
			incidenceTypeName = inc.IncidenceType.Name
			category = inc.IncidenceType.Category
			effectType = inc.IncidenceType.EffectType
		}

		// Build title
		categoryLabel := dtos.GetIncidenceCategoryLabel(category)
		title := categoryLabel + " - " + employeeName

		events[i] = dtos.CalendarEventResponse{
			ID:               inc.ID,
			EventType:        dtos.EventTypeIncidence,
			Title:            title,
			Description:      inc.Comments,
			StartDate:        inc.StartDate,
			EndDate:          inc.EndDate,
			AllDay:           true,
			Status:           inc.Status,
			EmployeeID:       inc.EmployeeID,
			EmployeeName:     employeeName,
			EmployeeNumber:   employeeNumber,
			CollarType:       collarType,
			DepartmentID:     departmentID,
			SourceID:         inc.ID,
			SourceType:       "incidence",
			Category:         category,
			EffectType:       effectType,
			Quantity:         inc.Quantity,
			CalculatedAmount: inc.CalculatedAmount,
			IncidenceType:    incidenceTypeName,
			CreatedAt:        inc.CreatedAt,
			UpdatedAt:        inc.UpdatedAt,
		}
	}

	return events, nil
}

// fetchShiftExceptions retrieves shift changes within the date range
func (s *CalendarService) fetchShiftExceptions(startDate, endDate time.Time, employeeIDs []uuid.UUID) ([]dtos.CalendarEventResponse, error) {
	query := s.db.Model(&models.ShiftException{}).
		Preload("Employee").
		Preload("Shift").
		Where("date >= ? AND date <= ?", startDate, endDate)

	if len(employeeIDs) > 0 {
		query = query.Where("employee_id IN ?", employeeIDs)
	}

	var exceptions []models.ShiftException
	if err := query.Find(&exceptions).Error; err != nil {
		return nil, err
	}

	events := make([]dtos.CalendarEventResponse, len(exceptions))
	for i, exc := range exceptions {
		employeeName := ""
		employeeNumber := ""
		collarType := ""
		var departmentID *uuid.UUID

		if exc.Employee != nil {
			employeeName = exc.Employee.FirstName + " " + exc.Employee.LastName
			employeeNumber = exc.Employee.EmployeeNumber
			collarType = exc.Employee.CollarType
			departmentID = exc.Employee.DepartmentID
		}

		shiftName := ""
		shiftCode := ""
		shiftTime := ""
		if exc.Shift != nil {
			shiftName = exc.Shift.Name
			shiftCode = exc.Shift.Code
			shiftTime = exc.Shift.StartTime + " - " + exc.Shift.EndTime
		}

		title := "Cambio de Turno - " + employeeName

		events[i] = dtos.CalendarEventResponse{
			ID:             exc.ID,
			EventType:      dtos.EventTypeShiftChange,
			Title:          title,
			Description:    fmt.Sprintf("Turno: %s (%s)", shiftName, shiftTime),
			StartDate:      exc.Date,
			EndDate:        exc.Date,
			AllDay:         true,
			Status:         "approved", // Shift exceptions are always approved when created
			EmployeeID:     exc.EmployeeID,
			EmployeeName:   employeeName,
			EmployeeNumber: employeeNumber,
			CollarType:     collarType,
			DepartmentID:   departmentID,
			SourceID:       exc.ID,
			SourceType:     "shift_exception",
			ShiftName:      shiftName,
			ShiftCode:      shiftCode,
			ShiftTime:      shiftTime,
			CreatedAt:      exc.CreatedAt,
			UpdatedAt:      exc.UpdatedAt,
		}
	}

	return events, nil
}

// calculateSummary computes statistics from the events
func (s *CalendarService) calculateSummary(summary *dtos.CalendarSummary, events []dtos.CalendarEventResponse) {
	for _, event := range events {
		// Count by status
		switch strings.ToUpper(event.Status) {
		case "PENDING":
			summary.Pending++
		case "APPROVED", "COMPLETED":
			summary.Approved++
		case "DECLINED", "REJECTED":
			summary.Declined++
		}

		// Count by collar type
		switch event.CollarType {
		case "white_collar":
			summary.WhiteCollar++
		case "blue_collar":
			summary.BlueCollar++
		case "gray_collar":
			summary.GrayCollar++
		}
	}
}
