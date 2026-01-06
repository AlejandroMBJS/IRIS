/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/employee.go
==============================================================================

DESCRIPTION:
    Defines the Employee model - the core entity of the payroll system.
    Contains all employee personal data, Mexican tax identifiers (RFC, CURP, NSS),
    salary information, and employment details needed for payroll processing.

USER PERSPECTIVE:
    - Stores all employee information visible in the "Empleados" section
    - Used for calculating payroll, taxes, and benefits
    - Tracks Mexican-specific data: RFC, CURP, NSS, IMSS registration
    - Supports three collar types determining payment frequency:
        * White Collar: Office/administrative staff, paid biweekly
        * Blue Collar: Unionized factory workers, paid weekly
        * Gray Collar: Non-unionized factory workers, paid weekly

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields (remember to update DTOs and API)
    ⚠️  CAUTION when modifying: Validation rules, salary calculations
    ❌  DO NOT modify: Mexican ID validation functions (RFC, CURP, NSS)
    📝  When adding fields: Also update internal/dtos/employee.go

SYNTAX EXPLANATION:
    - type Employee struct: Defines Employee as a Go struct (like a class)
    - BaseModel: Embedded struct, gives Employee all BaseModel fields
    - `gorm:"..."`: Database column configuration
        * type:varchar(50): Column type and size
        * uniqueIndex: Creates unique index (no duplicates)
        * not null: Column cannot be NULL
        * check:X IN (...): Database-level constraint for allowed values
        * default:'...'`: Default value if not provided
    - `json:"..."`: JSON field name for API responses
        * omitempty: Omit field from JSON if empty/zero

VALIDATION RULES:
    - RFC: 12-13 alphanumeric characters (Mexican tax ID)
    - CURP: 18 characters (Mexican unique population registry)
    - NSS: 11 digits (Social Security number)
    - DailySalary: Must be positive

BUSINESS LOGIC:
    - GetVacationDays(): Returns vacation days based on Mexican labor law
    - CalculateIntegratedDailySalary(): Calculates SDI for IMSS
    - CalculateSeniority(): Years of service for benefits calculation

RELATIONS:
    - Company: Belongs to one company
    - Incidences: Has many incidences (absences, vacations, etc.)
    - PayrollHeaders: Has many payroll records

==============================================================================
*/
package models

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// Employee represents an employee in the system.
// This is the central entity for all payroll operations.
type Employee struct {
    BaseModel
    
    // Personal Information
    EmployeeNumber   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"employee_number"`
    FirstName        string    `gorm:"type:varchar(100);not null" json:"first_name"`
    LastName         string    `gorm:"type:varchar(100);not null" json:"last_name"`
    MotherLastName   string    `gorm:"type:varchar(100)" json:"mother_last_name,omitempty"`
    DateOfBirth      time.Time `gorm:"type:date;not null" json:"date_of_birth"`
    Gender           string    `gorm:"type:varchar(20);check:gender IN ('male','female','other')" json:"gender"`
    MaritalStatus    string    `gorm:"type:varchar(50)" json:"marital_status,omitempty"`
    Nationality      string    `gorm:"type:varchar(100);default:'Mexicana'" json:"nationality"`
    BirthState       string    `gorm:"type:varchar(100)" json:"birth_state,omitempty"`
    EducationLevel   string    `gorm:"type:varchar(100)" json:"education_level,omitempty"`
    
    // Mexican Official IDs
    RFC              string    `gorm:"type:varchar(13);uniqueIndex;not null" json:"rfc"`
    CURP             string    `gorm:"type:varchar(18);uniqueIndex;not null" json:"curp"`
    NSS              string    `gorm:"type:varchar(11)" json:"nss,omitempty"`
    InfonavitCredit  string    `gorm:"type:varchar(50)" json:"infonavit_credit,omitempty"`
    
    // Contact Information
    PersonalEmail    string    `gorm:"type:varchar(255)" json:"personal_email,omitempty"`
    PersonalPhone    string    `gorm:"type:varchar(20)" json:"personal_phone,omitempty"`
    CellPhone        string    `gorm:"type:varchar(20)" json:"cell_phone,omitempty"`
    EmergencyContact string    `gorm:"type:varchar(255)" json:"emergency_contact,omitempty"`
    EmergencyPhone   string    `gorm:"type:varchar(20)" json:"emergency_phone,omitempty"`
    EmergencyRelationship string `gorm:"type:varchar(100)" json:"emergency_relationship,omitempty"`
    
    // Address
    Street           string    `gorm:"type:varchar(255)" json:"street,omitempty"`
    ExteriorNumber   string    `gorm:"type:varchar(20)" json:"exterior_number,omitempty"`
    InteriorNumber   string    `gorm:"type:varchar(20)" json:"interior_number,omitempty"`
    Neighborhood     string    `gorm:"type:varchar(100)" json:"neighborhood,omitempty"`
    Municipality     string    `gorm:"type:varchar(100)" json:"municipality,omitempty"`
    State            string    `gorm:"type:varchar(100);default:'San Luis Potosí'" json:"state"`
    PostalCode       string    `gorm:"type:varchar(10)" json:"postal_code,omitempty"`
    Country          string    `gorm:"type:varchar(100);default:'México'" json:"country"`
    
    // Employment Details
    HireDate         time.Time `gorm:"type:date;not null" json:"hire_date"`
    TerminationDate  *time.Time `gorm:"type:date" json:"termination_date,omitempty"`
    EmploymentStatus string    `gorm:"type:varchar(50);default:'active';check:employment_status IN ('active','inactive','on_leave','terminated')" json:"employment_status"`
    EmployeeType     string    `gorm:"type:varchar(50);check:employee_type IN ('permanent','temporary','contractor','intern')" json:"employee_type"`

    // Collar type determines payment frequency:
    // - white_collar: Administrative/office workers, biweekly payment
    // - blue_collar: Unionized factory/production workers, weekly payment
    // - gray_collar: Non-unionized factory/production workers, weekly payment
    CollarType       string    `gorm:"type:varchar(20);default:'white_collar';check:collar_type IN ('white_collar','blue_collar','gray_collar')" json:"collar_type"`
    PayFrequency     string    `gorm:"type:varchar(20);default:'biweekly';check:pay_frequency IN ('weekly','biweekly','monthly')" json:"pay_frequency"`
    IsSindicalizado  bool      `gorm:"default:false" json:"is_sindicalizado"` // For blue collar workers

    // Additional Employment Details
    ContractStartDate *time.Time `gorm:"type:date" json:"contract_start_date,omitempty"`
    ContractEndDate   *time.Time `gorm:"type:date" json:"contract_end_date,omitempty"`
    ProductionArea    string    `gorm:"type:varchar(100)" json:"production_area,omitempty"`
    Location          string    `gorm:"type:varchar(100)" json:"location,omitempty"`
    PatronalRegistry  string    `gorm:"type:varchar(50)" json:"patronal_registry,omitempty"`
    CompanyName       string    `gorm:"type:varchar(255)" json:"company_name,omitempty"` // Razón Social
    ShiftID           *uuid.UUID `gorm:"type:text" json:"shift_id,omitempty"`
    SupervisorID      *uuid.UUID `gorm:"type:text" json:"supervisor_id,omitempty"`
    TeamName          string    `gorm:"type:varchar(100)" json:"team_name,omitempty"`
    PackageCode       string    `gorm:"type:varchar(50)" json:"package_code,omitempty"`

    // Operational / Logistics
    Route             string    `gorm:"type:varchar(100)" json:"route,omitempty"`
    TransportStop     string    `gorm:"type:varchar(255)" json:"transport_stop,omitempty"`
    RecruitmentSource string    `gorm:"type:varchar(100)" json:"recruitment_source,omitempty"` // Fuente

    // Fiscal Information
    FiscalPostalCode  string    `gorm:"type:varchar(10)" json:"fiscal_postal_code,omitempty"`
    FiscalName        string    `gorm:"type:varchar(255)" json:"fiscal_name,omitempty"` // Nombre Fiscal del Empleado

    // Family Information (Children)
    Child1Gender      string    `gorm:"type:varchar(20)" json:"child1_gender,omitempty"`
    Child2Gender      string    `gorm:"type:varchar(20)" json:"child2_gender,omitempty"`
    Child3Gender      string    `gorm:"type:varchar(20)" json:"child3_gender,omitempty"`
    Child4Gender      string    `gorm:"type:varchar(20)" json:"child4_gender,omitempty"`
    
    // Financial Information
    DailySalary      float64   `gorm:"type:decimal(12,2);not null" json:"daily_salary"`
    IntegratedDailySalary float64 `gorm:"type:decimal(12,2)" json:"integrated_daily_salary"`
    PaymentMethod    string    `gorm:"type:varchar(50);default:'bank_transfer'" json:"payment_method"`
    BankName         string    `gorm:"type:varchar(100)" json:"bank_name,omitempty"`
    BankAccount      string    `gorm:"type:varchar(50)" json:"bank_account,omitempty"`
    CLABE            string    `gorm:"type:varchar(18)" json:"clabe,omitempty"`
    
    // IMSS & Tax Information
    IMSSRegistrationDate *time.Time `gorm:"type:date" json:"imss_registration_date,omitempty"`
    Regime             string    `gorm:"type:varchar(50);default:'salary'" json:"regime"`
    TaxRegime          string    `gorm:"type:varchar(50)" json:"tax_regime,omitempty"`
    
    // Foreign Keys
    CompanyID        uuid.UUID `gorm:"type:text;not null" json:"company_id"`
    DepartmentID     *uuid.UUID `gorm:"type:text" json:"department_id,omitempty"`
    PositionID       *uuid.UUID `gorm:"type:text" json:"position_id,omitempty"`
    CostCenterID     *uuid.UUID `gorm:"type:text" json:"cost_center_id,omitempty"`
    CreatedBy        *uuid.UUID `gorm:"type:text" json:"created_by,omitempty"`
    UpdatedBy        *uuid.UUID `gorm:"type:text" json:"updated_by,omitempty"`
    
    // Relations
    Company          *Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
    CreatedByUser    *User      `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
    UpdatedByUser    *User      `gorm:"foreignKey:UpdatedBy" json:"updated_by_user,omitempty"`
    CostCenter       *CostCenter `gorm:"foreignKey:CostCenterID" json:"cost_center,omitempty"`
    Shift            *Shift      `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
    Supervisor       *Employee   `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
    Incidences       []Incidence `gorm:"foreignKey:EmployeeID" json:"incidences,omitempty"`
    PrenominaMetrics []PrenominaMetric `gorm:"foreignKey:EmployeeID" json:"prenomina_metrics,omitempty"`
    PayrollHeaders   []PayrollHeader `gorm:"foreignKey:EmployeeID" json:"payroll_headers,omitempty"`
}

// TableName specifies the table name
func (Employee) TableName() string {
    return "employees"
}

// Validate validates employee data
func (e *Employee) Validate() error {
    var validationErrors []string
    
    // Required fields
    if strings.TrimSpace(e.EmployeeNumber) == "" {
        validationErrors = append(validationErrors, "employee number is required")
    }
    if strings.TrimSpace(e.FirstName) == "" {
        validationErrors = append(validationErrors, "first name is required")
    }
    if strings.TrimSpace(e.LastName) == "" {
        validationErrors = append(validationErrors, "last name is required")
    }
    
    // Mexican ID validation
    if !ValidateRFC(e.RFC) {
        validationErrors = append(validationErrors, "invalid RFC format")
    }
    if !ValidateCURP(e.CURP) {
        validationErrors = append(validationErrors, "invalid CURP format")
    }
    if e.NSS != "" && !ValidateNSS(e.NSS) {
        validationErrors = append(validationErrors, "invalid NSS format")
    }
    
    // Date validation
    if e.DateOfBirth.IsZero() {
        validationErrors = append(validationErrors, "date of birth is required")
    }
    if e.HireDate.IsZero() {
        validationErrors = append(validationErrors, "hire date is required")
    }
    if e.DateOfBirth.After(e.HireDate) {
        validationErrors = append(validationErrors, "date of birth cannot be after hire date")
    }
    
    // Financial validation
    if e.DailySalary <= 0 {
        validationErrors = append(validationErrors, "daily salary must be positive")
    }
    
    if len(validationErrors) > 0 {
        return errors.New(strings.Join(validationErrors, "; "))
    }
    
    return nil
}

// CalculateAge calculates employee age
func (e *Employee) CalculateAge() int {
    now := time.Now()
    years := now.Year() - e.DateOfBirth.Year()
    
    // Adjust if birthday hasn't occurred this year
    if now.YearDay() < e.DateOfBirth.YearDay() {
        years--
    }
    
    return years
}

// CalculateSeniority calculates years of service
func (e *Employee) CalculateSeniority() float64 {
    now := time.Now()
    if e.TerminationDate != nil && e.TerminationDate.Before(now) {
        now = *e.TerminationDate
    }
    
    duration := now.Sub(e.HireDate)
    years := duration.Hours() / 24 / 365.25
    return years
}

// CalculateIntegratedDailySalary calculates SDI
func (e *Employee) CalculateIntegratedDailySalary() float64 {
    if e.IntegratedDailySalary > 0 {
        return e.IntegratedDailySalary
    }
    
    // SDI = Daily Salary * (1 + (Aguinaldo days/365) + (Vacation days * Prima Vacacional/365))
    // Standard Mexican calculation
    aguinaldoDays := 15.0 // Minimum by law
    vacationDays := e.GetVacationDays()
    primaVacacional := 0.25 // 25% minimum by law
    
    sdi := e.DailySalary * (1 + (aguinaldoDays/365) + ((float64(vacationDays) * primaVacacional)/365))
    return roundToTwoDecimals(sdi)
}

// GetVacationDays returns vacation days based on seniority
func (e *Employee) GetVacationDays() int {
    years := int(e.CalculateSeniority())
    
    if years == 1 {
        return 12
    } else if years >= 2 && years <= 5 {
        return 14
    } else if years >= 6 && years <= 10 {
        return 16
    } else if years >= 11 && years <= 15 {
        return 18
    } else if years >= 16 && years <= 20 {
        return 20
    } else if years >= 21 && years <= 25 {
        return 22
    } else if years >= 26 && years <= 30 {
        return 24
    } else if years >= 31 && years <= 35 {
        return 26
    } else if years >= 36 && years <= 40 {
        return 28
    } else if years >= 41 && years <= 45 {
        return 30
    } else if years >= 46 && years <= 50 {
        return 32
    } else if years >= 51 {
        return 34
    }
    
    return 12
}

// BeforeSave validates and calculates before saving
func (e *Employee) BeforeSave(tx *gorm.DB) error {
    if err := e.Validate(); err != nil {
        return err
    }
    
    // Calculate SDI if not set
    if e.IntegratedDailySalary == 0 {
        e.IntegratedDailySalary = e.CalculateIntegratedDailySalary()
    }
    
    return nil
}

// Helper validation functions
func ValidateRFC(rfc string) bool {
    // RFC validation regex (12 or 13 characters)
    rfcRegex := regexp.MustCompile(`^[A-Z&Ñ]{3,4}[0-9]{6}[A-Z0-9]{3}$`)
    return rfcRegex.MatchString(rfc)
}

func ValidateCURP(curp string) bool {
    // CURP validation regex (18 characters)
    curpRegex := regexp.MustCompile(`^[A-Z]{4}[0-9]{6}[HM][A-Z]{5}[A-Z0-9]{2}$`)
    return curpRegex.MatchString(curp)
}

func ValidateNSS(nss string) bool {
    // NSS validation (11 digits)
    nssRegex := regexp.MustCompile(`^[0-9]{11}$`)
    return nssRegex.MatchString(nss)
}

// roundToTwoDecimals uses banker's rounding (round to nearest even) for precision
// This is the standard rounding method for financial calculations to avoid bias
func roundToTwoDecimals(value float64) float64 {
    return math.Round(value*100) / 100
}
