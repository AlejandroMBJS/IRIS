/*
Package services - HR Incidence Service

==============================================================================
FILE: internal/services/incidence_service.go
==============================================================================

DESCRIPTION:
    Manages HR incidences (attendance events) including absences, overtime, sick
    leave, vacations, bonuses, and deductions. Handles incidence types, approval
    workflow, and vacation balance calculations per Mexican labor law.

USER PERSPECTIVE:
    - Record employee incidences (absences, overtime, sick days, vacations)
    - Approve or reject incidence requests
    - Track vacation balances based on seniority
    - Upload evidence files for incidence documentation
    - View absence summaries and statistics

DEVELOPER GUIDELINES:
    OK to modify: Incidence categories, calculation methods
    CAUTION: Vacation day calculations follow Mexican Federal Labor Law
    DO NOT modify: Approved incidences cannot be deleted or edited
    Note: Incidences are processed during prenomina calculation

SYNTAX EXPLANATION:
    - Category types: absence, sick, vacation, overtime, delay, bonus, deduction
    - EffectType: positive (adds money), negative (deducts), neutral (informative)
    - CalculationMethod: daily_rate, hourly_rate, fixed_amount, percentage
    - Status flow: pending -> approved/rejected -> processed
    - Vacation days per Mexican law: Year 1=12, Year 2=14, increases with seniority

==============================================================================
*/
package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// IncidenceService handles incidence business logic
type IncidenceService struct {
	db *gorm.DB
}

// NewIncidenceService creates a new incidence service
func NewIncidenceService(db *gorm.DB) *IncidenceService {
	return &IncidenceService{db: db}
}

// IncidenceTypeRequest represents request for creating/updating incidence types
type IncidenceTypeRequest struct {
	Name              string                   `json:"name" binding:"required"`
	CategoryID        string                   `json:"category_id"`      // UUID of the parent category
	Category          string                   `json:"category"`         // Legacy field (required for backward compatibility)
	EffectType        string                   `json:"effect_type" binding:"required"`
	IsCalculated      bool                     `json:"is_calculated"`
	CalculationMethod string                   `json:"calculation_method"`
	DefaultValue      float64                  `json:"default_value"`
	RequiresEvidence  bool                     `json:"requires_evidence"`
	Description       string                   `json:"description"`
	FormFields        *models.FormFieldsConfig `json:"form_fields"`      // Custom form field definitions
	IsRequestable     bool                     `json:"is_requestable"`   // Can employees request this type?
	ApprovalFlow      string                   `json:"approval_flow"`    // Approval workflow type
	DisplayOrder      int                      `json:"display_order"`    // Display order in UI
}

// CreateIncidenceRequest represents request for creating an incidence
type CreateIncidenceRequest struct {
	EmployeeID      string  `json:"employee_id" binding:"required"`
	PayrollPeriodID string  `json:"payroll_period_id"`
	IncidenceTypeID string  `json:"incidence_type_id" binding:"required"`
	StartDate       string  `json:"start_date" binding:"required"`
	EndDate         string  `json:"end_date" binding:"required"`
	Quantity        float64 `json:"quantity" binding:"required"`
	Comments        string  `json:"comments"`
}

// UpdateIncidenceRequest represents request for updating an incidence
type UpdateIncidenceRequest struct {
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	Quantity   float64 `json:"quantity"`
	Comments   string  `json:"comments"`
	Status     string  `json:"status"`
}

// === Incidence Type Operations ===

// GetAllIncidenceTypes returns all incidence types
func (s *IncidenceService) GetAllIncidenceTypes() ([]models.IncidenceType, error) {
	var types []models.IncidenceType
	err := s.db.Preload("IncidenceCategory").Order("display_order, category, name").Find(&types).Error
	return types, err
}

// GetRequestableIncidenceTypes returns incidence types that employees can request
func (s *IncidenceService) GetRequestableIncidenceTypes() ([]models.IncidenceType, error) {
	var types []models.IncidenceType
	err := s.db.Preload("IncidenceCategory").
		Where("is_requestable = ?", true).
		Order("display_order, category, name").
		Find(&types).Error
	return types, err
}

// CreateIncidenceType creates a new incidence type
func (s *IncidenceService) CreateIncidenceType(req IncidenceTypeRequest) (*models.IncidenceType, error) {
	// Validate category (legacy) - still required for backward compatibility
	validCategories := map[string]bool{
		"absence": true, "sick": true, "vacation": true, "overtime": true,
		"delay": true, "bonus": true, "deduction": true, "other": true,
	}

	// If CategoryID is provided, get the category code from it
	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		id, err := uuid.Parse(req.CategoryID)
		if err != nil {
			return nil, errors.New("invalid category ID")
		}
		categoryID = &id

		// Get category to populate legacy field
		var category models.IncidenceCategory
		if err := s.db.First(&category, "id = ?", id).Error; err != nil {
			return nil, errors.New("category not found")
		}
		req.Category = category.Code
	}

	if req.Category == "" || !validCategories[req.Category] {
		return nil, errors.New("invalid category")
	}

	// Validate effect type
	validEffects := map[string]bool{"positive": true, "negative": true, "neutral": true}
	if !validEffects[req.EffectType] {
		return nil, errors.New("invalid effect type")
	}

	incidenceType := &models.IncidenceType{
		Name:              req.Name,
		CategoryID:        categoryID,
		Category:          req.Category,
		EffectType:        req.EffectType,
		IsCalculated:      req.IsCalculated,
		CalculationMethod: req.CalculationMethod,
		DefaultValue:      req.DefaultValue,
		RequiresEvidence:  req.RequiresEvidence,
		Description:       req.Description,
		IsRequestable:     req.IsRequestable,
		ApprovalFlow:      req.ApprovalFlow,
		DisplayOrder:      req.DisplayOrder,
	}

	// Set form fields if provided
	if req.FormFields != nil {
		if err := incidenceType.SetFormFieldsConfig(req.FormFields); err != nil {
			return nil, errors.New("invalid form fields configuration")
		}
	}

	// Default approval flow if not specified
	if incidenceType.ApprovalFlow == "" {
		incidenceType.ApprovalFlow = "standard"
	}

	if err := s.db.Create(incidenceType).Error; err != nil {
		return nil, err
	}

	// Reload with category relationship
	s.db.Preload("IncidenceCategory").First(incidenceType, "id = ?", incidenceType.ID)

	return incidenceType, nil
}

// UpdateIncidenceType updates an incidence type
func (s *IncidenceService) UpdateIncidenceType(id uuid.UUID, req IncidenceTypeRequest) (*models.IncidenceType, error) {
	var incidenceType models.IncidenceType
	if err := s.db.First(&incidenceType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("incidence type not found")
		}
		return nil, err
	}

	// Validate category (legacy) - still required for backward compatibility
	validCategories := map[string]bool{
		"absence": true, "sick": true, "vacation": true, "overtime": true,
		"delay": true, "bonus": true, "deduction": true, "other": true,
	}

	// If CategoryID is provided, get the category code from it
	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err != nil {
			return nil, errors.New("invalid category ID")
		}
		categoryID = &catID

		// Get category to populate legacy field
		var category models.IncidenceCategory
		if err := s.db.First(&category, "id = ?", catID).Error; err != nil {
			return nil, errors.New("category not found")
		}
		req.Category = category.Code
	}

	if req.Category == "" || !validCategories[req.Category] {
		return nil, errors.New("invalid category")
	}

	// Validate effect type
	validEffects := map[string]bool{"positive": true, "negative": true, "neutral": true}
	if !validEffects[req.EffectType] {
		return nil, errors.New("invalid effect type")
	}

	incidenceType.Name = req.Name
	incidenceType.CategoryID = categoryID
	incidenceType.Category = req.Category
	incidenceType.EffectType = req.EffectType
	incidenceType.IsCalculated = req.IsCalculated
	incidenceType.CalculationMethod = req.CalculationMethod
	incidenceType.DefaultValue = req.DefaultValue
	incidenceType.RequiresEvidence = req.RequiresEvidence
	incidenceType.Description = req.Description
	incidenceType.IsRequestable = req.IsRequestable
	incidenceType.DisplayOrder = req.DisplayOrder

	// Update approval flow if provided
	if req.ApprovalFlow != "" {
		incidenceType.ApprovalFlow = req.ApprovalFlow
	}

	// Set form fields if provided
	if req.FormFields != nil {
		if err := incidenceType.SetFormFieldsConfig(req.FormFields); err != nil {
			return nil, errors.New("invalid form fields configuration")
		}
	}

	if err := s.db.Save(&incidenceType).Error; err != nil {
		return nil, err
	}

	// Reload with category relationship
	s.db.Preload("IncidenceCategory").First(&incidenceType, "id = ?", incidenceType.ID)

	return &incidenceType, nil
}

// DeleteIncidenceType deletes an incidence type
func (s *IncidenceService) DeleteIncidenceType(id uuid.UUID) error {
	// First fetch the type to check if it's protected
	var incidenceType models.IncidenceType
	if err := s.db.First(&incidenceType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("incidence type not found")
		}
		return err
	}

	// Protect system types - vacation and sick categories cannot be deleted
	protectedCategories := map[string]bool{"vacation": true, "sick": true}
	if protectedCategories[incidenceType.Category] {
		return errors.New("no se puede eliminar este tipo de incidencia porque es requerido por el sistema")
	}

	// Check if type has any incidences
	var count int64
	s.db.Model(&models.Incidence{}).Where("incidence_type_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete incidence type with existing incidences")
	}

	result := s.db.Delete(&models.IncidenceType{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("incidence type not found")
	}
	return result.Error
}

// === Employee Incidence Operations ===

// GetIncidencesByEmployee returns all incidences for an employee
func (s *IncidenceService) GetIncidencesByEmployee(employeeID uuid.UUID) ([]models.Incidence, error) {
	var incidences []models.Incidence
	err := s.db.Preload("IncidenceType").Preload("PayrollPeriod").
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		Find(&incidences).Error
	return incidences, err
}

// GetIncidencesByPeriod returns all incidences for a payroll period
func (s *IncidenceService) GetIncidencesByPeriod(periodID uuid.UUID) ([]models.Incidence, error) {
	var incidences []models.Incidence
	err := s.db.Preload("IncidenceType").Preload("Employee").
		Where("payroll_period_id = ?", periodID).
		Order("start_date DESC").
		Find(&incidences).Error
	return incidences, err
}

// GetAllIncidences returns all incidences with optional filters
func (s *IncidenceService) GetAllIncidences(employeeID, periodID string, status string) ([]models.Incidence, error) {
	query := s.db.Preload("IncidenceType").Preload("Employee").Preload("PayrollPeriod")

	if employeeID != "" {
		query = query.Where("employee_id = ?", employeeID)
	}
	if periodID != "" {
		query = query.Where("payroll_period_id = ?", periodID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var incidences []models.Incidence
	err := query.Order("created_at DESC").Find(&incidences).Error
	return incidences, err
}

// GetIncidenceByID returns a single incidence by ID
func (s *IncidenceService) GetIncidenceByID(id uuid.UUID) (*models.Incidence, error) {
	var incidence models.Incidence
	err := s.db.Preload("IncidenceType").Preload("Employee").Preload("PayrollPeriod").
		First(&incidence, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("incidence not found")
		}
		return nil, err
	}
	return &incidence, nil
}

// CreateIncidence creates a new incidence
func (s *IncidenceService) CreateIncidence(req CreateIncidenceRequest) (*models.Incidence, error) {
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("invalid employee ID")
	}

	incidenceTypeID, err := uuid.Parse(req.IncidenceTypeID)
	if err != nil {
		return nil, errors.New("invalid incidence type ID")
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Verify employee exists
	var employee models.Employee
	if err := s.db.First(&employee, "id = ?", employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}

	// Verify incidence type exists
	var incidenceType models.IncidenceType
	if err := s.db.First(&incidenceType, "id = ?", incidenceTypeID).Error; err != nil {
		return nil, errors.New("incidence type not found")
	}

	// Find current active payroll period if not provided
	var periodID uuid.UUID
	if req.PayrollPeriodID != "" {
		periodID, err = uuid.Parse(req.PayrollPeriodID)
		if err != nil {
			return nil, errors.New("invalid payroll period ID")
		}
	} else {
		// Find the active period that contains the start date
		var period models.PayrollPeriod
		err := s.db.Where("start_date <= ? AND end_date >= ? AND status = ?", startDate, startDate, "active").
			First(&period).Error
		if err != nil {
			// If no active period found, try to find any period containing this date
			err = s.db.Where("start_date <= ? AND end_date >= ?", startDate, startDate).
				First(&period).Error
			if err != nil {
				return nil, errors.New("no payroll period found for the specified date")
			}
		}
		periodID = period.ID
	}

	// Calculate amount based on incidence type
	calculatedAmount := 0.0
	if incidenceType.IsCalculated {
		switch incidenceType.CalculationMethod {
		case "daily_rate":
			calculatedAmount = employee.DailySalary * req.Quantity
		case "hourly_rate":
			calculatedAmount = (employee.DailySalary / 8) * req.Quantity
		case "fixed_amount":
			calculatedAmount = incidenceType.DefaultValue * req.Quantity
		case "percentage":
			calculatedAmount = employee.DailySalary * (incidenceType.DefaultValue / 100) * req.Quantity
		default:
			calculatedAmount = req.Quantity
		}
	}

	incidence := &models.Incidence{
		EmployeeID:       employeeID,
		PayrollPeriodID:  periodID,
		IncidenceTypeID:  incidenceTypeID,
		StartDate:        startDate,
		EndDate:          endDate,
		Quantity:         req.Quantity,
		CalculatedAmount: calculatedAmount,
		Comments:         req.Comments,
		Status:           "pending",
	}

	if err := s.db.Create(incidence).Error; err != nil {
		return nil, err
	}

	// Reload with relationships
	return s.GetIncidenceByID(incidence.ID)
}

// UpdateIncidence updates an incidence
func (s *IncidenceService) UpdateIncidence(id uuid.UUID, req UpdateIncidenceRequest) (*models.Incidence, error) {
	incidence, err := s.GetIncidenceByID(id)
	if err != nil {
		return nil, err
	}

	if incidence.Status == "processed" {
		return nil, errors.New("cannot update a processed incidence")
	}

	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, errors.New("invalid start date format")
		}
		incidence.StartDate = startDate
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, errors.New("invalid end date format")
		}
		incidence.EndDate = endDate
	}

	if incidence.EndDate.Before(incidence.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	if req.Quantity > 0 {
		incidence.Quantity = req.Quantity
	}

	if req.Comments != "" {
		incidence.Comments = req.Comments
	}

	if req.Status != "" {
		validStatuses := map[string]bool{"pending": true, "approved": true, "rejected": true}
		if !validStatuses[req.Status] {
			return nil, errors.New("invalid status")
		}
		incidence.Status = req.Status
	}

	if err := s.db.Save(incidence).Error; err != nil {
		return nil, err
	}

	return s.GetIncidenceByID(id)
}

// DeleteIncidence deletes an incidence
func (s *IncidenceService) DeleteIncidence(id uuid.UUID) error {
	incidence, err := s.GetIncidenceByID(id)
	if err != nil {
		return err
	}

	if incidence.Status == "processed" {
		return errors.New("cannot delete a processed incidence")
	}

	return s.db.Delete(incidence).Error
}

// ApproveIncidence approves an incidence
func (s *IncidenceService) ApproveIncidence(id uuid.UUID, approverID uuid.UUID) (*models.Incidence, error) {
	incidence, err := s.GetIncidenceByID(id)
	if err != nil {
		return nil, err
	}

	if incidence.Status != "pending" {
		return nil, errors.New("only pending incidences can be approved")
	}

	now := time.Now()
	incidence.Status = "approved"
	incidence.ApprovedBy = &approverID
	incidence.ApprovedAt = &now

	if err := s.db.Save(incidence).Error; err != nil {
		return nil, err
	}

	return s.GetIncidenceByID(id)
}

// RejectIncidence rejects an incidence
func (s *IncidenceService) RejectIncidence(id uuid.UUID, approverID uuid.UUID) (*models.Incidence, error) {
	incidence, err := s.GetIncidenceByID(id)
	if err != nil {
		return nil, err
	}

	if incidence.Status != "pending" {
		return nil, errors.New("only pending incidences can be rejected")
	}

	now := time.Now()
	incidence.Status = "rejected"
	incidence.ApprovedBy = &approverID
	incidence.ApprovedAt = &now

	if err := s.db.Save(incidence).Error; err != nil {
		return nil, err
	}

	return s.GetIncidenceByID(id)
}

// GetEmployeeVacationBalance returns vacation statistics for an employee
func (s *IncidenceService) GetEmployeeVacationBalance(employeeID uuid.UUID) (map[string]interface{}, error) {
	// Get employee to calculate vacation days based on seniority
	var employee models.Employee
	if err := s.db.First(&employee, "id = ?", employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}

	// Calculate years of service
	yearsOfService := time.Since(employee.HireDate).Hours() / 24 / 365

	// Mexican labor law vacation days based on seniority
	var entitledDays int
	switch {
	case yearsOfService < 1:
		entitledDays = int(yearsOfService * 12) // Prorated first year
	case yearsOfService < 2:
		entitledDays = 12
	case yearsOfService < 3:
		entitledDays = 14
	case yearsOfService < 4:
		entitledDays = 16
	case yearsOfService < 5:
		entitledDays = 18
	case yearsOfService < 10:
		entitledDays = 20
	case yearsOfService < 15:
		entitledDays = 22
	case yearsOfService < 20:
		entitledDays = 24
	case yearsOfService < 25:
		entitledDays = 26
	case yearsOfService < 30:
		entitledDays = 28
	default:
		entitledDays = 30
	}

	// Get current year used vacation days
	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	var usedDays float64
	s.db.Model(&models.Incidence{}).
		Joins("JOIN incidence_types ON incidences.incidence_type_id = incidence_types.id").
		Where("incidences.employee_id = ?", employeeID).
		Where("incidence_types.category = ?", "vacation").
		Where("incidences.status IN (?)", []string{"approved", "processed"}).
		Where("incidences.start_date >= ? AND incidences.start_date <= ?", startOfYear, endOfYear).
		Select("COALESCE(SUM(incidences.quantity), 0)").
		Scan(&usedDays)

	// Get pending vacation requests
	var pendingDays float64
	s.db.Model(&models.Incidence{}).
		Joins("JOIN incidence_types ON incidences.incidence_type_id = incidence_types.id").
		Where("incidences.employee_id = ?", employeeID).
		Where("incidence_types.category = ?", "vacation").
		Where("incidences.status = ?", "pending").
		Select("COALESCE(SUM(incidences.quantity), 0)").
		Scan(&pendingDays)

	return map[string]interface{}{
		"employee_id":      employeeID,
		"years_of_service": yearsOfService,
		"entitled_days":    entitledDays,
		"used_days":        usedDays,
		"pending_days":     pendingDays,
		"available_days":   float64(entitledDays) - usedDays - pendingDays,
		"year":             currentYear,
	}, nil
}

// GetEmployeeAbsenceSummary returns absence statistics for an employee
func (s *IncidenceService) GetEmployeeAbsenceSummary(employeeID uuid.UUID, year int) (map[string]interface{}, error) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	// Get absences by category
	type CategoryCount struct {
		Category string  `json:"category"`
		Days     float64 `json:"days"`
		Count    int     `json:"count"`
	}

	var categoryCounts []CategoryCount
	s.db.Model(&models.Incidence{}).
		Joins("JOIN incidence_types ON incidences.incidence_type_id = incidence_types.id").
		Where("incidences.employee_id = ?", employeeID).
		Where("incidences.status IN (?)", []string{"approved", "processed"}).
		Where("incidences.start_date >= ? AND incidences.start_date <= ?", startOfYear, endOfYear).
		Group("incidence_types.category").
		Select("incidence_types.category, SUM(incidences.quantity) as days, COUNT(*) as count").
		Scan(&categoryCounts)

	return map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"by_category": categoryCounts,
	}, nil
}
