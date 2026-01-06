/*
Package services - Shift Service

==============================================================================
FILE: internal/services/shift_service.go
==============================================================================

DESCRIPTION:
    Manages shift operations - CRUD for work schedules and shift assignments.
    Handles business logic for employee scheduling.

USER PERSPECTIVE:
    - HR creates shifts like "Turno Matutino", "Turno Vespertino"
    - Employees are assigned to shifts
    - Shifts appear in calendar and affect attendance tracking

DEVELOPER GUIDELINES:
    OK to modify: Add new shift operations
    CAUTION: Validate company_id for multi-tenant security
    DO NOT modify: Shift code uniqueness validation

==============================================================================
*/
package services

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
	"backend/internal/repositories"
	"backend/internal/utils"
)

// ShiftService handles shift business logic
type ShiftService struct {
	db       *gorm.DB
	shiftRepo *repositories.ShiftRepository
}

// NewShiftService creates a new shift service
func NewShiftService(db *gorm.DB) *ShiftService {
	return &ShiftService{
		db:       db,
		shiftRepo: repositories.NewShiftRepository(db),
	}
}

// ShiftRequest represents request for creating/updating shifts
type ShiftRequest struct {
	Name            string   `json:"name" binding:"required"`
	Code            string   `json:"code" binding:"required"`
	Description     string   `json:"description"`
	StartTime       string   `json:"start_time" binding:"required"`
	EndTime         string   `json:"end_time" binding:"required"`
	BreakMinutes    int      `json:"break_minutes"`
	BreakStartTime  string   `json:"break_start_time"`
	WorkHoursPerDay float64  `json:"work_hours_per_day"`
	WorkDays        []int    `json:"work_days"`
	Color           string   `json:"color"`
	DisplayOrder    int      `json:"display_order"`
	IsActive        bool     `json:"is_active"`
	IsNightShift    bool     `json:"is_night_shift"`
	CollarTypes     []string `json:"collar_types"` // e.g., ["white_collar", "blue_collar"]
}

// ShiftResponse represents the shift response with employee count
type ShiftResponse struct {
	models.Shift
	EmployeeCount int64 `json:"employee_count"`
}

// GetAllShifts returns all shifts for a company, filtered by user role's collar type access
func (s *ShiftService) GetAllShifts(companyID uuid.UUID, userRole string) ([]ShiftResponse, error) {
	shifts, err := s.shiftRepo.FindAllByCompany(companyID)
	if err != nil {
		return nil, err
	}

	// Get allowed collar types for the user's role
	allowedCollarTypes := utils.GetAllowedCollarTypes(userRole)

	// Filter shifts and add employee count
	var responses []ShiftResponse
	for _, shift := range shifts {
		// If user has restricted collar types, filter shifts
		if allowedCollarTypes != nil {
			// Parse shift's collar types
			var shiftCollarTypes []string
			if shift.CollarTypes != "" && shift.CollarTypes != "[]" {
				json.Unmarshal([]byte(shift.CollarTypes), &shiftCollarTypes)
			}

			// If shift has specific collar types, check if any match user's allowed types
			if len(shiftCollarTypes) > 0 {
				hasMatch := false
				for _, allowed := range allowedCollarTypes {
					for _, shiftType := range shiftCollarTypes {
						if allowed == shiftType {
							hasMatch = true
							break
						}
					}
					if hasMatch {
						break
					}
				}
				if !hasMatch {
					continue // Skip this shift
				}
			}
			// If shift has no collar types (empty), it's available to all - include it
		}

		count, _ := s.shiftRepo.CountEmployeesWithShift(shift.ID)
		responses = append(responses, ShiftResponse{
			Shift:         shift,
			EmployeeCount: count,
		})
	}

	return responses, nil
}

// GetActiveShifts returns only active shifts for a company, filtered by user role
func (s *ShiftService) GetActiveShifts(companyID uuid.UUID, userRole string) ([]models.Shift, error) {
	shifts, err := s.shiftRepo.FindActiveByCompany(companyID)
	if err != nil {
		return nil, err
	}

	// Get allowed collar types for the user's role
	allowedCollarTypes := utils.GetAllowedCollarTypes(userRole)
	if allowedCollarTypes == nil {
		return shifts, nil // No filtering needed
	}

	// Filter shifts
	var filtered []models.Shift
	for _, shift := range shifts {
		var shiftCollarTypes []string
		if shift.CollarTypes != "" && shift.CollarTypes != "[]" {
			json.Unmarshal([]byte(shift.CollarTypes), &shiftCollarTypes)
		}

		// If shift has no collar types, include it for all
		if len(shiftCollarTypes) == 0 {
			filtered = append(filtered, shift)
			continue
		}

		// Check if any collar type matches
		for _, allowed := range allowedCollarTypes {
			for _, shiftType := range shiftCollarTypes {
				if allowed == shiftType {
					filtered = append(filtered, shift)
					goto nextShift
				}
			}
		}
	nextShift:
	}

	return filtered, nil
}

// GetShiftByID returns a single shift by ID
func (s *ShiftService) GetShiftByID(id, companyID uuid.UUID) (*models.Shift, error) {
	shift, err := s.shiftRepo.FindByIDAndCompany(id, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shift not found")
		}
		return nil, err
	}
	return shift, nil
}

// CreateShift creates a new shift
func (s *ShiftService) CreateShift(req ShiftRequest, companyID uuid.UUID) (*models.Shift, error) {
	// Check if code already exists
	exists, err := s.shiftRepo.ExistsByCodeAndCompany(req.Code, companyID, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("un turno con este código ya existe")
	}

	// Set default work days if not provided
	workDays := "[1,2,3,4,5]" // Mon-Fri
	if len(req.WorkDays) > 0 {
		// Convert array to JSON string
		data, _ := json.Marshal(req.WorkDays)
		workDays = string(data)
	}

	// Set default work hours if not provided
	workHours := req.WorkHoursPerDay
	if workHours <= 0 {
		workHours = 8.0
	}

	// Set default color if not provided
	color := req.Color
	if color == "" {
		color = "#3B82F6"
	}

	// Set collar types
	collarTypes := "[]"
	if len(req.CollarTypes) > 0 {
		data, _ := json.Marshal(req.CollarTypes)
		collarTypes = string(data)
	}

	shift := &models.Shift{
		Name:            req.Name,
		Code:            req.Code,
		Description:     req.Description,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		BreakMinutes:    req.BreakMinutes,
		BreakStartTime:  req.BreakStartTime,
		WorkHoursPerDay: workHours,
		WorkDays:        workDays,
		Color:           color,
		DisplayOrder:    req.DisplayOrder,
		IsActive:        req.IsActive,
		IsNightShift:    req.IsNightShift,
		CollarTypes:     collarTypes,
		CompanyID:       companyID,
	}

	if err := s.shiftRepo.Create(shift); err != nil {
		return nil, err
	}

	return shift, nil
}

// UpdateShift updates an existing shift
func (s *ShiftService) UpdateShift(id uuid.UUID, req ShiftRequest, companyID uuid.UUID) (*models.Shift, error) {
	// Find existing shift
	shift, err := s.shiftRepo.FindByIDAndCompany(id, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shift not found")
		}
		return nil, err
	}

	// Check if code already exists (excluding current shift)
	exists, err := s.shiftRepo.ExistsByCodeAndCompany(req.Code, companyID, &id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("un turno con este código ya existe")
	}

	// Update fields
	shift.Name = req.Name
	shift.Code = req.Code
	shift.Description = req.Description
	shift.StartTime = req.StartTime
	shift.EndTime = req.EndTime
	shift.BreakMinutes = req.BreakMinutes
	shift.BreakStartTime = req.BreakStartTime
	shift.IsActive = req.IsActive
	shift.IsNightShift = req.IsNightShift
	shift.DisplayOrder = req.DisplayOrder

	if req.WorkHoursPerDay > 0 {
		shift.WorkHoursPerDay = req.WorkHoursPerDay
	}
	if len(req.WorkDays) > 0 {
		data, _ := json.Marshal(req.WorkDays)
		shift.WorkDays = string(data)
	}
	if req.Color != "" {
		shift.Color = req.Color
	}

	// Update collar types (always update, even if empty to allow clearing)
	collarTypes := "[]"
	if len(req.CollarTypes) > 0 {
		data, _ := json.Marshal(req.CollarTypes)
		collarTypes = string(data)
	}
	shift.CollarTypes = collarTypes

	if err := s.shiftRepo.Update(shift); err != nil {
		return nil, err
	}

	return shift, nil
}

// DeleteShift deletes a shift
func (s *ShiftService) DeleteShift(id, companyID uuid.UUID) error {
	// Check if shift exists
	shift, err := s.shiftRepo.FindByIDAndCompany(id, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("shift not found")
		}
		return err
	}

	// Check if employees are assigned to this shift
	count, err := s.shiftRepo.CountEmployeesWithShift(shift.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("no se puede eliminar: hay empleados asignados a este turno")
	}

	return s.shiftRepo.Delete(id, companyID)
}

// ToggleShiftActive toggles the active status of a shift
func (s *ShiftService) ToggleShiftActive(id, companyID uuid.UUID) (*models.Shift, error) {
	shift, err := s.shiftRepo.FindByIDAndCompany(id, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shift not found")
		}
		return nil, err
	}

	shift.IsActive = !shift.IsActive

	if err := s.shiftRepo.Update(shift); err != nil {
		return nil, err
	}

	return shift, nil
}

// SeedDefaultShifts creates default shifts for a new company
func (s *ShiftService) SeedDefaultShifts(companyID uuid.UUID) error {
	// Check if shifts already exist
	shifts, _ := s.shiftRepo.FindAllByCompany(companyID)
	if len(shifts) > 0 {
		return nil // Already seeded
	}

	defaultShifts := []ShiftRequest{
		{
			Name:            "Turno Matutino",
			Code:            "T1",
			Description:     "Turno de mañana",
			StartTime:       "06:00",
			EndTime:         "14:00",
			BreakMinutes:    30,
			BreakStartTime:  "10:00",
			WorkHoursPerDay: 8,
			WorkDays:        []int{1, 2, 3, 4, 5},
			Color:           "#F59E0B",
			DisplayOrder:    1,
			IsActive:        true,
			IsNightShift:    false,
		},
		{
			Name:            "Turno Vespertino",
			Code:            "T2",
			Description:     "Turno de tarde",
			StartTime:       "14:00",
			EndTime:         "22:00",
			BreakMinutes:    30,
			BreakStartTime:  "18:00",
			WorkHoursPerDay: 8,
			WorkDays:        []int{1, 2, 3, 4, 5},
			Color:           "#3B82F6",
			DisplayOrder:    2,
			IsActive:        true,
			IsNightShift:    false,
		},
		{
			Name:            "Turno Nocturno",
			Code:            "T3",
			Description:     "Turno de noche",
			StartTime:       "22:00",
			EndTime:         "06:00",
			BreakMinutes:    30,
			BreakStartTime:  "02:00",
			WorkHoursPerDay: 8,
			WorkDays:        []int{1, 2, 3, 4, 5},
			Color:           "#8B5CF6",
			DisplayOrder:    3,
			IsActive:        true,
			IsNightShift:    true,
		},
		{
			Name:            "Horario Administrativo",
			Code:            "ADM",
			Description:     "Horario de oficina",
			StartTime:       "09:00",
			EndTime:         "18:00",
			BreakMinutes:    60,
			BreakStartTime:  "14:00",
			WorkHoursPerDay: 8,
			WorkDays:        []int{1, 2, 3, 4, 5},
			Color:           "#10B981",
			DisplayOrder:    4,
			IsActive:        true,
			IsNightShift:    false,
		},
	}

	for _, req := range defaultShifts {
		if _, err := s.CreateShift(req, companyID); err != nil {
			return err
		}
	}

	return nil
}
