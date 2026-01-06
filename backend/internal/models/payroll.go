/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/payroll.go
==============================================================================

DESCRIPTION:
    Defines all payroll-related models including calculations, pre-nomina metrics,
    and employer contributions. This is the heart of the payroll system.

USER PERSPECTIVE:
    - PayrollCalculation: The final payroll result for each employee per period
    - PrenominaMetric: Work metrics (hours, days, overtime) before final calculation
    - PayrollDetail: Line-by-line breakdown of income/deductions
    - EmployerContribution: IMSS/Infonavit contributions the company pays

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new income/deduction fields, new statuses
    ⚠️  CAUTION: Calculation totals logic, status transitions
    ❌  DO NOT modify: Field types for monetary amounts (decimal precision)
    📝  Monetary fields use decimal(15,2) for Mexican pesos (2 decimal places)

SYNTAX EXPLANATION:
    - decimal(15,2): Database decimal type, 15 total digits, 2 after decimal
    - *time.Time: Pointer to time, allows NULL (e.g., ApprovedAt before approval)
    - check:X IN (...): Database constraint ensuring valid values only

WORKFLOW:
    1. System creates PrenominaMetric with work metrics from incidences
    2. Payroll calculation runs, creates PayrollCalculation
    3. Status moves: pending → calculated → approved → processed → paid
    4. EmployerContribution tracks what company owes to IMSS/Infonavit

MEXICAN PAYROLL CONCEPTS:
    - ISR (Impuesto Sobre la Renta): Income tax withholding
    - IMSS: Social security contributions (employee + employer)
    - Infonavit: Housing fund contributions
    - Aguinaldo: Christmas bonus (15 days minimum)
    - Prima Vacacional: Vacation premium (25% of vacation pay)
    - SDI (Salario Diario Integrado): Integrated daily salary for IMSS

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
)

// PayrollCalculation represents a single payroll calculation for an employee in a period.
// This is the main payroll record showing all income, deductions, and net pay.
type PayrollCalculation struct {
	BaseModel
	EmployeeID         uuid.UUID  `gorm:"type:text;not null" json:"employee_id"`
	PayrollPeriodID    uuid.UUID  `gorm:"type:text;not null" json:"payroll_period_id"`
	PrenominaMetricID  uuid.UUID  `gorm:"type:text" json:"prenomina_metric_id"`
	CalculationDate    *time.Time `json:"calculation_date"`
	CalculationStatus  string     `gorm:"type:varchar(50);default:'pending';check:calculation_status IN ('pending','calculated','approved','rejected')" json:"calculation_status"`
	PayrollStatus      string     `gorm:"type:varchar(50);default:'pending';check:payroll_status IN ('pending','processed','paid','cancelled')" json:"payroll_status"` // 'processed' means waiting for payment
	ApprovedBy         *uuid.UUID `gorm:"type:text" json:"approved_by,omitempty"`
	ApprovedAt         *time.Time `json:"approved_at,omitempty"`
	ProcessedBy        *uuid.UUID `gorm:"type:text" json:"processed_by,omitempty"`
	ProcessedAt        *time.Time `json:"processed_at,omitempty"`

	// Incomes
	RegularSalary      float64 `gorm:"type:decimal(15,2);default:0" json:"regular_salary"`
	OvertimeAmount     float64 `gorm:"type:decimal(15,2);default:0" json:"overtime_amount"`
	DoubleOvertimeAmount float64 `gorm:"type:decimal(15,2);default:0" json:"double_overtime_amount"`
	TripleOvertimeAmount float64 `gorm:"type:decimal(15,2);default:0" json:"triple_overtime_amount"`
	VacationPremium    float64 `gorm:"type:decimal(15,2);default:0" json:"vacation_premium"`
	Aguinaldo          float64 `gorm:"type:decimal(15,2);default:0" json:"aguinaldo"`
	BonusAmount        float64 `gorm:"type:decimal(15,2);default:0" json:"bonus_amount"`
	CommissionAmount   float64 `gorm:"type:decimal(15,2);default:0" json:"commission_amount"`
	OtherExtras        float64 `gorm:"type:decimal(15,2);default:0" json:"other_extras"`

	// Deductions
	ISRWithholding     float64 `gorm:"type:decimal(15,2);default:0" json:"isr_withholding"`
	IMSSEmployee       float64 `gorm:"type:decimal(15,2);default:0" json:"imss_employee"`
	InfonavitEmployee  float64 `gorm:"type:decimal(15,2);default:0" json:"infonavit_employee"`
	RetirementSavings  float64 `gorm:"type:decimal(15,2);default:0" json:"retirement_savings"`
	LoanDeductions     float64 `gorm:"type:decimal(15,2);default:0" json:"loan_deductions"`
	AdvanceDeductions  float64 `gorm:"type:decimal(15,2);default:0" json:"advance_deductions"`
	OtherDeductions    float64 `gorm:"type:decimal(15,2);default:0" json:"other_deductions"`

	// Benefits / Subsidies
	FoodVouchers       float64 `gorm:"type:decimal(15,2);default:0" json:"food_vouchers"`
	SavingsFund        float64 `gorm:"type:decimal(15,2);default:0" json:"savings_fund"`
	EmploymentSubsidy  float64 `gorm:"type:decimal(15,2);default:0" json:"employment_subsidy"` // ISR Employment Subsidy

	// Totals
	TotalGrossIncome   float64 `gorm:"type:decimal(15,2);default:0" json:"total_gross_income"`
	TotalStatutoryDeductions float64 `gorm:"type:decimal(15,2);default:0" json:"total_statutory_deductions"`
	TotalOtherDeductions float64 `gorm:"type:decimal(15,2);default:0" json:"total_other_deductions"`
	TotalNetPay        float64 `gorm:"type:decimal(15,2);default:0" json:"total_net_pay"`

	// Employer Contributions (duplicated for easy access, also in separate table)
	IMSSEmployer       float64 `gorm:"type:decimal(15,2);default:0" json:"imss_employer"`
	InfonavitEmployer  float64 `gorm:"type:decimal(15,2);default:0" json:"infonavit_employer"`

	// Relations
	Employee           *Employee          `gorm:"foreignKey:EmployeeID;constraint:OnDelete:RESTRICT" json:"employee,omitempty"`
	PayrollPeriod      *PayrollPeriod     `gorm:"foreignKey:PayrollPeriodID;constraint:OnDelete:RESTRICT" json:"payroll_period,omitempty"`
	PrenominaMetric    *PrenominaMetric   `gorm:"foreignKey:PrenominaMetricID;constraint:OnDelete:CASCADE" json:"prenomina_metric,omitempty"`
	ApprovedByUser     *User              `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL" json:"approved_by_user,omitempty"`
	PayrollDetails     []PayrollDetail    `gorm:"foreignKey:PayrollCalculationID;constraint:OnDelete:CASCADE" json:"payroll_details,omitempty"`
	EmployerContribution *EmployerContribution `gorm:"foreignKey:PayrollCalculationID;constraint:OnDelete:CASCADE" json:"employer_contribution,omitempty"`
}

// TableName specifies the table name
func (PayrollCalculation) TableName() string {
	return "payroll_calculations"
}

// PayrollDetail represents a single line item in a payroll calculation.
type PayrollDetail struct {
	BaseModel
	PayrollCalculationID uuid.UUID `gorm:"type:text;not null" json:"payroll_calculation_id"`
	Concept              string    `gorm:"type:varchar(255);not null" json:"concept"`
	ConceptType          string    `gorm:"type:varchar(50);not null;check:concept_type IN ('income','deduction','employer_contribution','benefit')" json:"concept_type"`
	Category             string    `gorm:"type:varchar(50)" json:"category,omitempty"` // e.g., 'regular', 'overtime', 'isr', 'imss'
	Amount               float64   `gorm:"type:decimal(15,2);default:0" json:"amount"`
	IsTaxable            bool      `gorm:"default:true" json:"is_taxable"`
	IsIMSSBase           bool      `gorm:"default:false" json:"is_imss_base"`
	IsInfonavitBase      bool      `gorm:"default:false" json:"is_infonavit_base"`
	SATCode              string    `gorm:"type:varchar(20)" json:"sat_code,omitempty"`
}

// TableName specifies the table name
func (PayrollDetail) TableName() string {
	return "payroll_details"
}

// PrenominaMetric stores pre-payroll calculation metrics for an employee for a period.
type PrenominaMetric struct {
	BaseModel
	EmployeeID           uuid.UUID  `gorm:"type:text;not null" json:"employee_id"`
	PayrollPeriodID      uuid.UUID  `gorm:"type:text;not null" json:"payroll_period_id"`
	CalculationStatus    string     `gorm:"type:varchar(50);default:'pending';check:calculation_status IN ('pending','calculated','approved','rejected','processed')" json:"calculation_status"`
	CalculationDate      *time.Time `json:"calculation_date"`

	// Work metrics
	WorkedDays           float64 `gorm:"type:decimal(5,2);default:0" json:"worked_days"`
	RegularHours         float64 `gorm:"type:decimal(5,2);default:0" json:"regular_hours"`
	OvertimeHours        float64 `gorm:"type:decimal(5,2);default:0" json:"overtime_hours"`
	DoubleOvertimeHours  float64 `gorm:"type:decimal(5,2);default:0" json:"double_overtime_hours"`
	TripleOvertimeHours  float64 `gorm:"type:decimal(5,2);default:0" json:"triple_overtime_hours"`

	// Leave metrics
	AbsenceDays          float64 `gorm:"type:decimal(5,2);default:0" json:"absence_days"`
	SickDays             float64 `gorm:"type:decimal(5,2);default:0" json:"sick_days"`
	VacationDays         float64 `gorm:"type:decimal(5,2);default:0" json:"vacation_days"`
	UnpaidLeaveDays      float64 `gorm:"type:decimal(5,2);default:0" json:"unpaid_leave_days"`

	// Other metrics
	DelaysCount          int     `gorm:"default:0" json:"delays_count"`
	DelayMinutes         float64 `gorm:"type:decimal(5,2);default:0" json:"delay_minutes"`
	EarlyDeparturesCount int     `gorm:"default:0" json:"early_departures_count"`

	// Monetary amounts (pre-calculated)
	RegularSalary        float64 `gorm:"type:decimal(15,2);default:0" json:"regular_salary"`
	OvertimeAmount       float64 `gorm:"type:decimal(15,2);default:0" json:"overtime_amount"`
	DoubleOvertimeAmount float64 `gorm:"type:decimal(15,2);default:0" json:"double_overtime_amount"`
	TripleOvertimeAmount float64 `gorm:"type:decimal(15,2);default:0" json:"triple_overtime_amount"`
	BonusAmount          float64 `gorm:"type:decimal(15,2);default:0" json:"bonus_amount"`
	CommissionAmount     float64 `gorm:"type:decimal(15,2);default:0" json:"commission_amount"`
	OtherExtraAmount     float64 `gorm:"type:decimal(15,2);default:0" json:"other_extra_amount"`

	// Deductions
	LoanDeduction        float64 `gorm:"type:decimal(15,2);default:0" json:"loan_deduction"`
	AdvanceDeduction     float64 `gorm:"type:decimal(15,2);default:0" json:"advance_deduction"`
	OtherDeduction       float64 `gorm:"type:decimal(15,2);default:0" json:"other_deduction"`
	DelayDeduction       float64 `gorm:"type:decimal(15,2);default:0" json:"delay_deduction"`

	// Summary Totals
	TotalExtras          float64 `gorm:"type:decimal(15,2);default:0" json:"total_extras"`
	TotalDeductions      float64 `gorm:"type:decimal(15,2);default:0" json:"total_deductions"`
	GrossIncome          float64 `gorm:"type:decimal(15,2);default:0" json:"gross_income"`
	NetIncome            float64 `gorm:"type:decimal(15,2);default:0" json:"net_income"`

	// Relations
	Employee        *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	PayrollPeriod   *PayrollPeriod `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period,omitempty"`
}

// TableName specifies the table name
func (PrenominaMetric) TableName() string {
	return "prenomina_metrics"
}

// EmployerContribution stores the employer's contributions for an employee for a payroll period.
type EmployerContribution struct {
	BaseModel
	PayrollCalculationID uuid.UUID `gorm:"type:text;not null;uniqueIndex" json:"payroll_calculation_id"`
	EmployeeID           uuid.UUID `gorm:"type:text;not null" json:"employee_id"`
	PayrollPeriodID      uuid.UUID `gorm:"type:text;not null" json:"payroll_period_id"`

	// IMSS Contributions
	IMSSDiseaseMaternity float64 `gorm:"type:decimal(15,2);default:0" json:"imss_disease_maternity"` // Enfermedad y Maternidad
	IMSSDisabilityLife   float64 `gorm:"type:decimal(15,2);default:0" json:"imss_disability_life"`   // Invalidez y Vida
	IMSSRetirement       float64 `gorm:"type:decimal(15,2);default:0" json:"imss_retirement"`       // Cesantía en Edad Avanzada y Vejez (RCV)
	IMSSChildcare        float64 `gorm:"type:decimal(15,2);default:0" json:"imss_childcare"`        // Guarderías y Prestaciones Sociales
	IMSSWorkRisk         float64 `gorm:"type:decimal(15,2);default:0" json:"imss_work_risk"`        // Riesgos de Trabajo

	// Infonavit Contributions
	InfonavitEmployer    float64 `gorm:"type:decimal(15,2);default:0" json:"infonavit_employer"`

	// Other Contributions
	RetirementSAR        float64 `gorm:"type:decimal(15,2);default:0" json:"retirement_sar"` // Fondo de Retiro (SAR)
	StatePayrollTax      float64 `gorm:"type:decimal(15,2);default:0" json:"state_payroll_tax"` // Impuesto Sobre Nómina Estatal

	// Benefits (Employer's portion)
	FoodVouchers         float64 `gorm:"type:decimal(15,2);default:0" json:"food_vouchers"`
	SavingsFund          float64 `gorm:"type:decimal(15,2);default:0" json:"savings_fund"`

	// Totals
	TotalIMSS            float64 `gorm:"type:decimal(15,2);default:0" json:"total_imss"`
	TotalInfonavit       float64 `gorm:"type:decimal(15,2);default:0" json:"total_infonavit"`
	TotalContributions   float64 `gorm:"type:decimal(15,2);default:0" json:"total_contributions"`

	// Relations
	PayrollCalculation   *PayrollCalculation `gorm:"foreignKey:PayrollCalculationID" json:"payroll_calculation,omitempty"`
	Employee             *Employee           `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	PayrollPeriod        *PayrollPeriod      `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period,omitempty"`
}

// TableName specifies the table name
func (EmployerContribution) TableName() string {
	return "employer_contributions"
}

// PayrollHeader for previous versions of the project.
type PayrollHeader struct {
	BaseModel
	EmployeeID uuid.UUID `gorm:"type:text;not null" json:"employee_id"`
	PayrollPeriodID uuid.UUID `gorm:"type:text;not null" json:"payroll_period_id"`
}
