/*
Package services - IRIS Benefits Administration Service

==============================================================================
FILE: internal/services/benefits_service.go
==============================================================================

DESCRIPTION:
    Business logic for Benefits Administration including benefit plans,
    enrollment, dependents, and life events.

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

// BenefitsService provides business logic for benefits administration
type BenefitsService struct {
	db *gorm.DB
}

// NewBenefitsService creates a new BenefitsService
func NewBenefitsService(db *gorm.DB) *BenefitsService {
	return &BenefitsService{db: db}
}

// === Benefit Plan DTOs ===

type CreateBenefitPlanDTO struct {
	CompanyID            uuid.UUID
	Name                 string
	Code                 string
	Description          string
	BenefitType          models.BenefitType
	ProviderName         string
	ProviderCode         string
	GroupNumber          string
	EffectiveDate        *time.Time
	TerminationDate      *time.Time
	EligibleAfterDays    int
	EmployeeCostMonthly  float64
	EmployeeCostBiweekly float64
	EmployerCostMonthly  float64
	IsTaxable            bool
	RequiresEvidence     bool
	AllowMidYearChange   bool
	CreatedByID          *uuid.UUID
}

type BenefitPlanFilters struct {
	CompanyID   uuid.UUID
	BenefitType string
	IsActive    *bool
	Page        int
	Limit       int
}

type PaginatedBenefitPlans struct {
	Data       []models.BenefitPlan `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// CreateBenefitPlan creates a new benefit plan
func (s *BenefitsService) CreateBenefitPlan(dto CreateBenefitPlanDTO) (*models.BenefitPlan, error) {
	if dto.Name == "" {
		return nil, errors.New("plan name is required")
	}

	plan := &models.BenefitPlan{
		CompanyID:            dto.CompanyID,
		Name:                 dto.Name,
		Code:                 dto.Code,
		Description:          dto.Description,
		BenefitType:          dto.BenefitType,
		ProviderName:         dto.ProviderName,
		ProviderCode:         dto.ProviderCode,
		GroupNumber:          dto.GroupNumber,
		EffectiveDate:        dto.EffectiveDate,
		TerminationDate:      dto.TerminationDate,
		EligibleAfterDays:    dto.EligibleAfterDays,
		EmployeeCostMonthly:  dto.EmployeeCostMonthly,
		EmployeeCostBiweekly: dto.EmployeeCostBiweekly,
		EmployerCostMonthly:  dto.EmployerCostMonthly,
		IsTaxable:            dto.IsTaxable,
		RequiresEvidence:     dto.RequiresEvidence,
		AllowMidYearChange:   dto.AllowMidYearChange,
		IsActive:             true,
		CreatedByID:          dto.CreatedByID,
	}

	if err := s.db.Create(plan).Error; err != nil {
		return nil, err
	}

	return plan, nil
}

// GetBenefitPlanByID retrieves a benefit plan by ID
func (s *BenefitsService) GetBenefitPlanByID(id uuid.UUID) (*models.BenefitPlan, error) {
	var plan models.BenefitPlan
	err := s.db.Preload("Options").First(&plan, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("benefit plan not found")
		}
		return nil, err
	}
	return &plan, nil
}

// ListBenefitPlans retrieves benefit plans with filters
func (s *BenefitsService) ListBenefitPlans(filters BenefitPlanFilters) (*PaginatedBenefitPlans, error) {
	query := s.db.Model(&models.BenefitPlan{}).Where("company_id = ?", filters.CompanyID)

	if filters.BenefitType != "" {
		query = query.Where("benefit_type = ?", filters.BenefitType)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
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

	var plans []models.BenefitPlan
	if err := query.Offset(offset).Limit(filters.Limit).Preload("Options").Order("name ASC").Find(&plans).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	return &PaginatedBenefitPlans{
		Data:       plans,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// === Enrollment Methods ===

type EnrollInBenefitDTO struct {
	CompanyID         uuid.UUID
	EmployeeID        uuid.UUID
	PlanID            uuid.UUID
	OptionID          *uuid.UUID
	PeriodID          *uuid.UUID
	CoverageLevel     string
	EffectiveDate     time.Time
	DependentIDs      []uuid.UUID
	LifeEventID       *uuid.UUID
	EnrollmentType    string // self, assigned
	CreatedByID       *uuid.UUID
}

// EnrollInBenefit enrolls an employee in a benefit plan
func (s *BenefitsService) EnrollInBenefit(dto EnrollInBenefitDTO) (*models.BenefitEnrollment, error) {
	// Validate plan exists and is active
	plan, err := s.GetBenefitPlanByID(dto.PlanID)
	if err != nil {
		return nil, err
	}
	if !plan.IsActive {
		return nil, errors.New("benefit plan is not active")
	}

	// Check if already enrolled in this plan type
	var existing models.BenefitEnrollment
	err = s.db.Where("employee_id = ? AND plan_id = ? AND status IN (?)",
		dto.EmployeeID, dto.PlanID, []string{"pending", "active", "pending_approval"}).
		First(&existing).Error
	if err == nil {
		return nil, errors.New("employee already enrolled in this plan")
	}

	// Get option cost if specified
	var employeeCost, employerCost float64
	if dto.OptionID != nil {
		var option models.BenefitOption
		if err := s.db.First(&option, "id = ?", dto.OptionID).Error; err == nil {
			employeeCost = option.EmployeeCostBiweekly
			employerCost = option.EmployerCostMonthly
		}
	} else {
		employeeCost = plan.EmployeeCostBiweekly
		employerCost = plan.EmployerCostMonthly
	}

	enrollment := &models.BenefitEnrollment{
		CompanyID:          dto.CompanyID,
		EmployeeID:         dto.EmployeeID,
		PlanID:             dto.PlanID,
		OptionID:           dto.OptionID,
		PeriodID:           dto.PeriodID,
		CoverageLevel:      dto.CoverageLevel,
		EffectiveDate:      dto.EffectiveDate,
		EmployeeCost:       employeeCost,
		EmployerCost:       employerCost,
		DeductionFrequency: "biweekly",
		LifeEventID:        dto.LifeEventID,
		Status:             models.BenefitEnrollmentPending,
		CreatedByID:        dto.CreatedByID,
	}

	if plan.RequiresEvidence {
		enrollment.RequiresApproval = true
		enrollment.Status = models.BenefitEnrollmentPendingApproval
	}

	if err := s.db.Create(enrollment).Error; err != nil {
		return nil, err
	}

	// Add dependents to enrollment
	for _, depID := range dto.DependentIDs {
		benefitDep := &models.BenefitDependent{
			EnrollmentID:  enrollment.ID,
			DependentID:   depID,
			IsCovered:     true,
			EffectiveDate: dto.EffectiveDate,
		}
		s.db.Create(benefitDep)
	}

	return enrollment, nil
}

// GetEnrollmentByID retrieves an enrollment by ID
func (s *BenefitsService) GetEnrollmentByID(id uuid.UUID) (*models.BenefitEnrollment, error) {
	var enrollment models.BenefitEnrollment
	err := s.db.Preload("Plan").Preload("Option").Preload("Dependents").Preload("Dependents.Dependent").
		First(&enrollment, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("enrollment not found")
		}
		return nil, err
	}
	return &enrollment, nil
}

// GetEmployeeEnrollmentsByBenefit gets all benefit enrollments for an employee
func (s *BenefitsService) GetEmployeeEnrollmentsByBenefit(employeeID uuid.UUID, status string) ([]models.BenefitEnrollment, error) {
	query := s.db.Where("employee_id = ?", employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var enrollments []models.BenefitEnrollment
	if err := query.Preload("Plan").Preload("Option").Find(&enrollments).Error; err != nil {
		return nil, err
	}
	return enrollments, nil
}

// ApproveEnrollment approves a pending enrollment
func (s *BenefitsService) ApproveEnrollment(id uuid.UUID, approverID uuid.UUID) (*models.BenefitEnrollment, error) {
	enrollment, err := s.GetEnrollmentByID(id)
	if err != nil {
		return nil, err
	}

	if enrollment.Status != models.BenefitEnrollmentPendingApproval {
		return nil, errors.New("enrollment is not pending approval")
	}

	now := time.Now()
	enrollment.Status = models.BenefitEnrollmentActive
	enrollment.ApprovedByID = &approverID
	enrollment.ApprovedAt = &now

	if err := s.db.Save(enrollment).Error; err != nil {
		return nil, err
	}

	return enrollment, nil
}

// DeclineEnrollment declines an enrollment
func (s *BenefitsService) DeclineEnrollment(id uuid.UUID, reason string) (*models.BenefitEnrollment, error) {
	enrollment, err := s.GetEnrollmentByID(id)
	if err != nil {
		return nil, err
	}

	enrollment.Status = models.BenefitEnrollmentDeclined
	enrollment.RejectionReason = reason

	if err := s.db.Save(enrollment).Error; err != nil {
		return nil, err
	}

	return enrollment, nil
}

// TerminateEnrollment terminates an enrollment
func (s *BenefitsService) TerminateEnrollment(id uuid.UUID, terminationDate time.Time) (*models.BenefitEnrollment, error) {
	enrollment, err := s.GetEnrollmentByID(id)
	if err != nil {
		return nil, err
	}

	enrollment.Status = models.BenefitEnrollmentTerminated
	enrollment.TerminationDate = &terminationDate

	if err := s.db.Save(enrollment).Error; err != nil {
		return nil, err
	}

	return enrollment, nil
}

// WaiveBenefit records a benefit waiver
func (s *BenefitsService) WaiveBenefit(companyID, employeeID, planID uuid.UUID, reason string) (*models.BenefitEnrollment, error) {
	enrollment := &models.BenefitEnrollment{
		CompanyID:    companyID,
		EmployeeID:   employeeID,
		PlanID:       planID,
		Status:       models.BenefitEnrollmentDeclined,
		IsWaived:     true,
		WaiverReason: reason,
		EffectiveDate: time.Now(),
	}

	if err := s.db.Create(enrollment).Error; err != nil {
		return nil, err
	}

	return enrollment, nil
}

// === Dependent Methods ===

type CreateDependentDTO struct {
	CompanyID    uuid.UUID
	EmployeeID   uuid.UUID
	FirstName    string
	LastName     string
	DateOfBirth  time.Time
	Gender       string
	Relationship string
	SSN          string
	Email        string
	Phone        string
}

// CreateDependent creates a new dependent
func (s *BenefitsService) CreateDependent(dto CreateDependentDTO) (*models.Dependent, error) {
	if dto.FirstName == "" || dto.LastName == "" {
		return nil, errors.New("first name and last name are required")
	}

	dependent := &models.Dependent{
		CompanyID:    dto.CompanyID,
		EmployeeID:   dto.EmployeeID,
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		DateOfBirth:  dto.DateOfBirth,
		Gender:       dto.Gender,
		Relationship: dto.Relationship,
		SSN:          dto.SSN,
		Email:        dto.Email,
		Phone:        dto.Phone,
		IsActive:     true,
	}

	if err := s.db.Create(dependent).Error; err != nil {
		return nil, err
	}

	return dependent, nil
}

// GetEmployeeDependents gets all dependents for an employee
func (s *BenefitsService) GetEmployeeDependents(employeeID uuid.UUID) ([]models.Dependent, error) {
	var dependents []models.Dependent
	if err := s.db.Where("employee_id = ? AND is_active = ?", employeeID, true).Find(&dependents).Error; err != nil {
		return nil, err
	}
	return dependents, nil
}

// VerifyDependent verifies a dependent
func (s *BenefitsService) VerifyDependent(id uuid.UUID, verifierID uuid.UUID) (*models.Dependent, error) {
	var dependent models.Dependent
	if err := s.db.First(&dependent, "id = ?", id).Error; err != nil {
		return nil, errors.New("dependent not found")
	}

	now := time.Now()
	dependent.IsVerified = true
	dependent.VerifiedAt = &now
	dependent.VerifiedByID = &verifierID

	if err := s.db.Save(&dependent).Error; err != nil {
		return nil, err
	}

	return &dependent, nil
}

// === Life Event Methods ===

type CreateLifeEventDTO struct {
	CompanyID   uuid.UUID
	EmployeeID  uuid.UUID
	EventType   string
	EventDate   time.Time
	Description string
	DocumentURL string
}

// CreateLifeEvent creates a life event record
func (s *BenefitsService) CreateLifeEvent(dto CreateLifeEventDTO) (*models.LifeEvent, error) {
	if dto.EventType == "" {
		return nil, errors.New("event type is required")
	}

	// Default enrollment window is 30 days
	enrollmentDeadline := dto.EventDate.AddDate(0, 0, 30)

	event := &models.LifeEvent{
		CompanyID:          dto.CompanyID,
		EmployeeID:         dto.EmployeeID,
		EventType:          dto.EventType,
		EventDate:          dto.EventDate,
		Description:        dto.Description,
		DocumentURL:        dto.DocumentURL,
		EnrollmentDeadline: enrollmentDeadline,
		Status:             "pending",
	}

	if err := s.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

// ApproveLifeEvent approves a life event
func (s *BenefitsService) ApproveLifeEvent(id uuid.UUID, approverID uuid.UUID) (*models.LifeEvent, error) {
	var event models.LifeEvent
	if err := s.db.First(&event, "id = ?", id).Error; err != nil {
		return nil, errors.New("life event not found")
	}

	now := time.Now()
	event.Status = "approved"
	event.ApprovedByID = &approverID
	event.ApprovedAt = &now

	if err := s.db.Save(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEmployeeLifeEvents gets life events for an employee
func (s *BenefitsService) GetEmployeeLifeEvents(employeeID uuid.UUID) ([]models.LifeEvent, error) {
	var events []models.LifeEvent
	if err := s.db.Where("employee_id = ?", employeeID).Order("event_date DESC").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// === Enrollment Period Methods ===

type CreateEnrollmentPeriodDTO struct {
	CompanyID         uuid.UUID
	Name              string
	Description       string
	StartDate         time.Time
	EndDate           time.Time
	CoverageStartDate time.Time
	CoverageEndDate   time.Time
	Year              int
	PlanIDs           []string
	CreatedByID       *uuid.UUID
}

// CreateEnrollmentPeriod creates an open enrollment period
func (s *BenefitsService) CreateEnrollmentPeriod(dto CreateEnrollmentPeriodDTO) (*models.EnrollmentPeriod, error) {
	if dto.Name == "" {
		return nil, errors.New("period name is required")
	}

	period := &models.EnrollmentPeriod{
		CompanyID:         dto.CompanyID,
		Name:              dto.Name,
		Description:       dto.Description,
		StartDate:         dto.StartDate,
		EndDate:           dto.EndDate,
		CoverageStartDate: dto.CoverageStartDate,
		CoverageEndDate:   dto.CoverageEndDate,
		Year:              dto.Year,
		PlanIDs:           dto.PlanIDs,
		Status:            models.EnrollmentPeriodStatusDraft,
		CreatedByID:       dto.CreatedByID,
	}

	if period.Year == 0 {
		period.Year = time.Now().Year()
	}

	if err := s.db.Create(period).Error; err != nil {
		return nil, err
	}

	return period, nil
}

// OpenEnrollmentPeriod opens an enrollment period
func (s *BenefitsService) OpenEnrollmentPeriod(id uuid.UUID) (*models.EnrollmentPeriod, error) {
	var period models.EnrollmentPeriod
	if err := s.db.First(&period, "id = ?", id).Error; err != nil {
		return nil, errors.New("enrollment period not found")
	}

	period.Status = models.EnrollmentPeriodStatusOpen

	if err := s.db.Save(&period).Error; err != nil {
		return nil, err
	}

	return &period, nil
}

// GetActiveEnrollmentPeriod gets the currently active enrollment period
func (s *BenefitsService) GetActiveEnrollmentPeriod(companyID uuid.UUID) (*models.EnrollmentPeriod, error) {
	var period models.EnrollmentPeriod
	now := time.Now()
	err := s.db.Where("company_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		companyID, models.EnrollmentPeriodStatusOpen, now, now).
		First(&period).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &period, nil
}

// === Beneficiary Methods ===

type CreateBeneficiaryDTO struct {
	CompanyID       uuid.UUID
	EmployeeID      uuid.UUID
	EnrollmentID    uuid.UUID
	FirstName       string
	LastName        string
	Relationship    string
	DateOfBirth     *time.Time
	SSN             string
	Address         string
	City            string
	State           string
	ZipCode         string
	Phone           string
	BeneficiaryType string
	PercentageShare float64
}

// CreateBeneficiary creates a beneficiary
func (s *BenefitsService) CreateBeneficiary(dto CreateBeneficiaryDTO) (*models.Beneficiary, error) {
	if dto.FirstName == "" || dto.LastName == "" {
		return nil, errors.New("first name and last name are required")
	}
	if dto.PercentageShare <= 0 || dto.PercentageShare > 100 {
		return nil, errors.New("percentage share must be between 0 and 100")
	}

	// Validate total percentage doesn't exceed 100 for the type
	var totalShare float64
	s.db.Model(&models.Beneficiary{}).
		Where("enrollment_id = ? AND beneficiary_type = ? AND is_active = ?", dto.EnrollmentID, dto.BeneficiaryType, true).
		Select("COALESCE(SUM(percentage_share), 0)").
		Scan(&totalShare)

	if totalShare+dto.PercentageShare > 100 {
		return nil, errors.New("total percentage share exceeds 100%")
	}

	beneficiary := &models.Beneficiary{
		CompanyID:       dto.CompanyID,
		EmployeeID:      dto.EmployeeID,
		EnrollmentID:    dto.EnrollmentID,
		FirstName:       dto.FirstName,
		LastName:        dto.LastName,
		Relationship:    dto.Relationship,
		DateOfBirth:     dto.DateOfBirth,
		SSN:             dto.SSN,
		Address:         dto.Address,
		City:            dto.City,
		State:           dto.State,
		ZipCode:         dto.ZipCode,
		Phone:           dto.Phone,
		BeneficiaryType: dto.BeneficiaryType,
		PercentageShare: dto.PercentageShare,
		IsActive:        true,
	}

	if beneficiary.BeneficiaryType == "" {
		beneficiary.BeneficiaryType = "primary"
	}

	if err := s.db.Create(beneficiary).Error; err != nil {
		return nil, err
	}

	return beneficiary, nil
}

// GetEnrollmentBeneficiaries gets beneficiaries for an enrollment
func (s *BenefitsService) GetEnrollmentBeneficiaries(enrollmentID uuid.UUID) ([]models.Beneficiary, error) {
	var beneficiaries []models.Beneficiary
	if err := s.db.Where("enrollment_id = ? AND is_active = ?", enrollmentID, true).
		Order("beneficiary_type ASC, percentage_share DESC").
		Find(&beneficiaries).Error; err != nil {
		return nil, err
	}
	return beneficiaries, nil
}
