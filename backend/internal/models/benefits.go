/*
Package models - IRIS Payroll System Benefits Administration Module Models

==============================================================================
FILE: internal/models/benefits.go
==============================================================================

DESCRIPTION:
    Data models for Benefits Administration including benefit plans,
    enrollment, dependents, and life events.

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// BenefitType represents the type of benefit
type BenefitType string

const (
	BenefitTypeHealth       BenefitType = "health"
	BenefitTypeDental       BenefitType = "dental"
	BenefitTypeVision       BenefitType = "vision"
	BenefitTypeLife         BenefitType = "life"
	BenefitTypeDisability   BenefitType = "disability"
	BenefitTypeRetirement   BenefitType = "retirement"
	BenefitTypeFSA          BenefitType = "fsa"
	BenefitTypeHSA          BenefitType = "hsa"
	BenefitTypeWellness     BenefitType = "wellness"
	BenefitTypeOther        BenefitType = "other"
)

// EnrollmentPeriodStatus represents the status of an enrollment period
type EnrollmentPeriodStatus string

const (
	EnrollmentPeriodStatusDraft     EnrollmentPeriodStatus = "draft"
	EnrollmentPeriodStatusScheduled EnrollmentPeriodStatus = "scheduled"
	EnrollmentPeriodStatusOpen      EnrollmentPeriodStatus = "open"
	EnrollmentPeriodStatusClosed    EnrollmentPeriodStatus = "closed"
)

// BenefitEnrollmentStatus represents enrollment status
type BenefitEnrollmentStatus string

const (
	BenefitEnrollmentPending    BenefitEnrollmentStatus = "pending"
	BenefitEnrollmentActive     BenefitEnrollmentStatus = "active"
	BenefitEnrollmentDeclined   BenefitEnrollmentStatus = "declined"
	BenefitEnrollmentTerminated BenefitEnrollmentStatus = "terminated"
	BenefitEnrollmentPendingApproval BenefitEnrollmentStatus = "pending_approval"
)

// BenefitPlan represents a benefit plan offered by the company
type BenefitPlan struct {
	BaseModel
	CompanyID    uuid.UUID   `gorm:"type:text;not null;index" json:"company_id"`

	// Plan Info
	Name         string      `gorm:"size:255;not null" json:"name"`
	Code         string      `gorm:"size:50;index" json:"code"`
	Description  string      `gorm:"type:text" json:"description"`
	BenefitType  BenefitType `gorm:"size:50;not null" json:"benefit_type"`

	// Provider
	ProviderName string      `gorm:"size:255" json:"provider_name"`
	ProviderCode string      `gorm:"size:100" json:"provider_code"`
	GroupNumber  string      `gorm:"size:100" json:"group_number"`

	// Coverage Dates
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	TerminationDate *time.Time `json:"termination_date,omitempty"`

	// Eligibility
	EligibleRoles pq.StringArray `gorm:"type:text[]" json:"eligible_roles"`
	EligibleAfterDays int       `gorm:"default:0" json:"eligible_after_days"` // Waiting period
	EligibleCollarTypes pq.StringArray `gorm:"type:text[]" json:"eligible_collar_types"`

	// Costs
	EmployeeCostMonthly  float64 `gorm:"default:0" json:"employee_cost_monthly"`
	EmployeeCostBiweekly float64 `gorm:"default:0" json:"employee_cost_biweekly"`
	EmployerCostMonthly  float64 `gorm:"default:0" json:"employer_cost_monthly"`

	// Coverage Tiers
	HasTiers     bool        `gorm:"default:false" json:"has_tiers"`
	TiersConfig  string      `gorm:"type:text" json:"tiers_config"` // JSON: [{name: "Employee Only", cost: 100}, ...]

	// Documents
	SummaryDocURL string     `gorm:"size:500" json:"summary_doc_url"`
	PlanDocURL    string     `gorm:"size:500" json:"plan_doc_url"`

	// Status
	IsActive     bool        `gorm:"default:true" json:"is_active"`
	IsTaxable    bool        `gorm:"default:false" json:"is_taxable"`

	// Settings
	RequiresEvidence bool    `gorm:"default:false" json:"requires_evidence"` // Evidence of insurability
	AllowMidYearChange bool  `gorm:"default:false" json:"allow_mid_year_change"`

	CreatedByID  *uuid.UUID  `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Options      []BenefitOption     `gorm:"foreignKey:PlanID" json:"options,omitempty"`
	Enrollments  []BenefitEnrollment `gorm:"foreignKey:PlanID" json:"enrollments,omitempty"`
}

func (BenefitPlan) TableName() string {
	return "benefit_plans"
}

// BenefitOption represents coverage options within a plan
type BenefitOption struct {
	BaseModel
	PlanID       uuid.UUID  `gorm:"type:text;not null;index" json:"plan_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`

	// Coverage
	CoverageLevel string    `gorm:"size:100" json:"coverage_level"` // employee_only, employee_spouse, employee_children, family
	CoverageAmount float64  `gorm:"default:0" json:"coverage_amount"`

	// Costs
	EmployeeCostMonthly  float64 `gorm:"default:0" json:"employee_cost_monthly"`
	EmployeeCostBiweekly float64 `gorm:"default:0" json:"employee_cost_biweekly"`
	EmployerCostMonthly  float64 `gorm:"default:0" json:"employer_cost_monthly"`

	// Deductible/Limits
	Deductible   float64    `gorm:"default:0" json:"deductible"`
	OutOfPocketMax float64  `gorm:"default:0" json:"out_of_pocket_max"`

	IsDefault    bool       `gorm:"default:false" json:"is_default"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	// Relationships
	Plan         *BenefitPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (BenefitOption) TableName() string {
	return "benefit_options"
}

// EnrollmentPeriod represents an open enrollment period
type EnrollmentPeriod struct {
	BaseModel
	CompanyID    uuid.UUID              `gorm:"type:text;not null;index" json:"company_id"`

	Name         string                 `gorm:"size:255;not null" json:"name"`
	Description  string                 `gorm:"type:text" json:"description"`

	// Dates
	StartDate    time.Time              `gorm:"not null" json:"start_date"`
	EndDate      time.Time              `gorm:"not null" json:"end_date"`

	// Coverage Period
	CoverageStartDate time.Time         `gorm:"not null" json:"coverage_start_date"`
	CoverageEndDate   time.Time         `gorm:"not null" json:"coverage_end_date"`
	Year         int                    `gorm:"not null" json:"year"`

	// Status
	Status       EnrollmentPeriodStatus `gorm:"size:50;default:'draft'" json:"status"`

	// Plans available in this period
	PlanIDs      pq.StringArray         `gorm:"type:text[]" json:"plan_ids"`

	// Reminders
	ReminderSent bool                   `gorm:"default:false" json:"reminder_sent"`
	ReminderSentAt *time.Time           `json:"reminder_sent_at,omitempty"`

	CreatedByID  *uuid.UUID             `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (EnrollmentPeriod) TableName() string {
	return "benefit_enrollment_periods"
}

// BenefitEnrollment represents an employee's enrollment in a benefit plan
type BenefitEnrollment struct {
	BaseModel
	CompanyID    uuid.UUID               `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID               `gorm:"type:text;not null;index" json:"employee_id"`
	PlanID       uuid.UUID               `gorm:"type:text;not null;index" json:"plan_id"`
	OptionID     *uuid.UUID              `gorm:"type:text" json:"option_id,omitempty"`
	PeriodID     *uuid.UUID              `gorm:"type:text" json:"period_id,omitempty"`

	// Status
	Status       BenefitEnrollmentStatus `gorm:"size:50;default:'pending'" json:"status"`

	// Coverage
	CoverageLevel string                 `gorm:"size:100" json:"coverage_level"`
	CoverageAmount float64               `gorm:"default:0" json:"coverage_amount"`

	// Dates
	EnrolledAt   time.Time              `gorm:"autoCreateTime" json:"enrolled_at"`
	EffectiveDate time.Time             `gorm:"not null" json:"effective_date"`
	TerminationDate *time.Time          `json:"termination_date,omitempty"`

	// Costs
	EmployeeCost float64                `gorm:"default:0" json:"employee_cost"`
	EmployerCost float64                `gorm:"default:0" json:"employer_cost"`
	DeductionFrequency string           `gorm:"size:50" json:"deduction_frequency"` // biweekly, monthly

	// Life Event (if applicable)
	LifeEventID  *uuid.UUID             `gorm:"type:text" json:"life_event_id,omitempty"`

	// Approval
	RequiresApproval bool               `gorm:"default:false" json:"requires_approval"`
	ApprovedByID *uuid.UUID             `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time             `json:"approved_at,omitempty"`
	RejectionReason string              `gorm:"size:255" json:"rejection_reason"`

	// Waiver
	IsWaived     bool                   `gorm:"default:false" json:"is_waived"`
	WaiverReason string                 `gorm:"type:text" json:"waiver_reason"`

	CreatedByID  *uuid.UUID             `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Employee     *Employee              `gorm:"foreignKey:EmployeeID;constraint:OnDelete:RESTRICT" json:"employee,omitempty"`
	Plan         *BenefitPlan           `gorm:"foreignKey:PlanID;constraint:OnDelete:RESTRICT" json:"plan,omitempty"`
	Option       *BenefitOption         `gorm:"foreignKey:OptionID;constraint:OnDelete:SET NULL" json:"option,omitempty"`
	Dependents   []BenefitDependent     `gorm:"foreignKey:EnrollmentID;constraint:OnDelete:CASCADE" json:"dependents,omitempty"`
}

func (BenefitEnrollment) TableName() string {
	return "benefit_enrollments"
}

// Dependent represents a dependent of an employee
type Dependent struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Personal Info
	FirstName    string     `gorm:"size:100;not null" json:"first_name"`
	LastName     string     `gorm:"size:100;not null" json:"last_name"`
	MiddleName   string     `gorm:"size:100" json:"middle_name"`
	DateOfBirth  time.Time  `gorm:"not null" json:"date_of_birth"`
	Gender       string     `gorm:"size:20" json:"gender"`
	SSN          string     `gorm:"size:20" json:"ssn"` // Encrypted
	Relationship string     `gorm:"size:50;not null" json:"relationship"` // spouse, child, domestic_partner, etc.

	// Contact
	Email        string     `gorm:"size:255" json:"email"`
	Phone        string     `gorm:"size:20" json:"phone"`

	// Address (if different from employee)
	AddressDifferent bool   `gorm:"default:false" json:"address_different"`
	Address      string     `gorm:"size:255" json:"address"`
	City         string     `gorm:"size:100" json:"city"`
	State        string     `gorm:"size:50" json:"state"`
	ZipCode      string     `gorm:"size:20" json:"zip_code"`

	// Status
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsDisabled   bool       `gorm:"default:false" json:"is_disabled"`
	IsStudent    bool       `gorm:"default:false" json:"is_student"`

	// Verification
	VerificationDocURL string `gorm:"size:500" json:"verification_doc_url"`
	IsVerified   bool       `gorm:"default:false" json:"is_verified"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	VerifiedByID *uuid.UUID `gorm:"type:text" json:"verified_by_id,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (Dependent) TableName() string {
	return "benefit_dependents"
}

// BenefitDependent links dependents to benefit enrollments
type BenefitDependent struct {
	BaseModel
	EnrollmentID uuid.UUID  `gorm:"type:text;not null;index" json:"enrollment_id"`
	DependentID  uuid.UUID  `gorm:"type:text;not null;index" json:"dependent_id"`

	// Coverage
	IsCovered    bool       `gorm:"default:true" json:"is_covered"`
	EffectiveDate time.Time `gorm:"not null" json:"effective_date"`
	TerminationDate *time.Time `json:"termination_date,omitempty"`

	// Relationships
	Enrollment   *BenefitEnrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
	Dependent    *Dependent         `gorm:"foreignKey:DependentID" json:"dependent,omitempty"`
}

func (BenefitDependent) TableName() string {
	return "benefit_enrollment_dependents"
}

// LifeEvent represents a qualifying life event
type LifeEvent struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Event Info
	EventType    string     `gorm:"size:100;not null" json:"event_type"` // marriage, birth, adoption, divorce, death, etc.
	EventDate    time.Time  `gorm:"not null" json:"event_date"`
	Description  string     `gorm:"type:text" json:"description"`

	// Evidence
	DocumentURL  string     `gorm:"size:500" json:"document_url"`

	// Window
	EnrollmentDeadline time.Time `gorm:"not null" json:"enrollment_deadline"` // Usually 30-60 days from event

	// Status
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, approved, rejected, expired
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	RejectionReason string  `gorm:"size:255" json:"rejection_reason"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (LifeEvent) TableName() string {
	return "benefit_life_events"
}

// BenefitClaim represents a benefit claim
type BenefitClaim struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	EnrollmentID uuid.UUID  `gorm:"type:text;not null;index" json:"enrollment_id"`

	// Claim Info
	ClaimNumber  string     `gorm:"size:100;uniqueIndex" json:"claim_number"`
	ServiceDate  time.Time  `gorm:"not null" json:"service_date"`
	ProviderName string     `gorm:"size:255" json:"provider_name"`
	Description  string     `gorm:"type:text" json:"description"`

	// Amounts
	BilledAmount float64    `gorm:"default:0" json:"billed_amount"`
	AllowedAmount float64   `gorm:"default:0" json:"allowed_amount"`
	PaidAmount   float64    `gorm:"default:0" json:"paid_amount"`
	EmployeeResponsibility float64 `gorm:"default:0" json:"employee_responsibility"`

	// Status
	Status       string     `gorm:"size:50;default:'submitted'" json:"status"` // submitted, processing, approved, denied, paid
	DenialReason string     `gorm:"size:255" json:"denial_reason"`

	// Documents
	DocumentURL  string     `gorm:"size:500" json:"document_url"`

	// Processing
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`

	// Relationships
	Employee     *Employee          `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Enrollment   *BenefitEnrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
}

func (BenefitClaim) TableName() string {
	return "benefit_claims"
}

// Beneficiary represents a beneficiary for life insurance/retirement
type Beneficiary struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	EnrollmentID uuid.UUID  `gorm:"type:text;not null;index" json:"enrollment_id"`

	// Beneficiary Info
	FirstName    string     `gorm:"size:100;not null" json:"first_name"`
	LastName     string     `gorm:"size:100;not null" json:"last_name"`
	Relationship string     `gorm:"size:50;not null" json:"relationship"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty"`
	SSN          string     `gorm:"size:20" json:"ssn"` // Encrypted

	// Contact
	Address      string     `gorm:"size:255" json:"address"`
	City         string     `gorm:"size:100" json:"city"`
	State        string     `gorm:"size:50" json:"state"`
	ZipCode      string     `gorm:"size:20" json:"zip_code"`
	Phone        string     `gorm:"size:20" json:"phone"`

	// Allocation
	BeneficiaryType string  `gorm:"size:50;not null" json:"beneficiary_type"` // primary, contingent
	PercentageShare float64 `gorm:"default:0" json:"percentage_share"` // Must sum to 100 per type

	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Relationships
	Employee     *Employee          `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Enrollment   *BenefitEnrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
}

func (Beneficiary) TableName() string {
	return "benefit_beneficiaries"
}
