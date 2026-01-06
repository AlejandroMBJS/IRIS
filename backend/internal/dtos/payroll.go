/*
Package dtos - Payroll Calculation Data Transfer Objects

==============================================================================
FILE: internal/dtos/payroll.go
==============================================================================

DESCRIPTION:
    Defines all payroll calculation request and response structures including
    individual calculations, bulk processing, payslip generation, and
    financial reports. Core DTOs for the payroll processing system.

USER PERSPECTIVE:
    - Shapes payroll calculation results displayed in the UI
    - Defines report formats for HR and Finance teams
    - Contains all income, deduction, and employer contribution breakdowns

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new report types, enhance calculation details
    ⚠️  CAUTION: Changing calculation field names (affects reports)
    ❌  DO NOT modify: Tax/IMSS calculation structures without fiscal review
    📝  Keep totals fields consistent across all response types

SYNTAX EXPLANATION:
    - PayrollCalculationRequest: Input for single employee calculation
    - PayrollCalculationResponse: Complete breakdown of one payroll
    - PayrollBulkCalculateRequest: Process multiple employees at once
    - CollarTypeSummary: Aggregate data grouped by employee type

CALCULATION BREAKDOWN:
    Income:
        - RegularSalary, OvertimeAmount, VacationPremium, Aguinaldo

    Deductions (Employee):
        - ISRWithholding, IMSSEmployee, InfonavitEmployee

    Employer Contributions:
        - TotalIMSS, TotalInfonavit, TotalRetirement

REPORT TYPES:
    - Payslip (PDF/XML): Individual employee payslip
    - Financial Export: Aggregated data for accounting
    - Period Summary: Overview of entire payroll period

==============================================================================
*/
package dtos

import (
	"time"

	"github.com/google/uuid"
)

// PayrollCalculationRequest represents request for payroll calculation
type PayrollCalculationRequest struct {
	EmployeeID    uuid.UUID `json:"employee_id" binding:"required"`
	PayrollPeriodID uuid.UUID `json:"payroll_period_id" binding:"required"`
	CalculateSDI    bool      `json:"calculate_sdi"` // Recalculate Integrated Daily Salary
}

// PayrollCalculationResponse represents payroll calculation results
type PayrollCalculationResponse struct {
	ID                    uuid.UUID              `json:"id"`
	EmployeeID            uuid.UUID              `json:"employee_id"`
	EmployeeName          string                 `json:"employee_name"`
	EmployeeNumber        string                 `json:"employee_number"`
	PayrollPeriodID       uuid.UUID              `json:"payroll_period_id"`
	PeriodCode            string                 `json:"period_code"`
	PaymentDate           time.Time              `json:"payment_date"`

	// Income
	RegularSalary         float64                `json:"regular_salary"`
	OvertimeAmount        float64                `json:"overtime_amount"`
	VacationPremium       float64                `json:"vacation_premium"`
	Aguinaldo             float64                `json:"aguinaldo"`
	OtherExtras           float64                `json:"other_extras"`

	// Deductions
	ISRWithholding        float64                `json:"isr_withholding"`
	IMSSEmployee          float64                `json:"imss_employee"`
	InfonavitEmployee     float64                `json:"infonavit_employee"`
	RetirementSavings     float64                `json:"retirement_savings"`

	// Other deductions
	LoanDeductions        float64                `json:"loan_deductions"`
	AdvanceDeductions     float64                `json:"advance_deductions"`
	OtherDeductions       float64                `json:"other_deductions"`

	// Subsidies and benefits
	FoodVouchers          float64                `json:"food_vouchers"`
	SavingsFund           float64                `json:"savings_fund"`

	// Totals
	TotalGrossIncome      float64                `json:"total_gross_income"`
	TotalDeductions       float64                `json:"total_deductions"`
	TotalNetPay           float64                `json:"total_net_pay"`

	// Employer contributions
	EmployerContributions EmployerContributionResponse `json:"employer_contributions"`

	// Metadata
	CalculationStatus     string                 `json:"calculation_status"`
	CalculationDate       time.Time              `json:"calculation_date"`
	PayrollStatus         string                 `json:"payroll_status"`
}

// EmployerContributionResponse represents employer contribution details
type EmployerContributionResponse struct {
	TotalIMSS          float64 `json:"total_imss"`
	TotalInfonavit     float64 `json:"total_infonavit"`
	TotalRetirement    float64 `json:"total_retirement"`
	TotalContributions float64 `json:"total_contributions"`
}

// PayrollBulkCalculateRequest for bulk calculation
type PayrollBulkCalculateRequest struct {
	PayrollPeriodID uuid.UUID   `json:"payroll_period_id" binding:"required"`
	EmployeeIDs     []uuid.UUID `json:"employee_ids"`
	CalculateAll    bool        `json:"calculate_all"`
}

// PayrollBulkCalculationResponse for bulk calculation results
type PayrollBulkCalculationResponse struct {
	PeriodCode      string                   `json:"period_code"`
	TotalCalculated int                      `json:"total_calculated"`
	TotalSuccess    int                      `json:"total_success"`
	TotalFailed     int                      `json:"total_failed"`
	Results         []PayrollCalculationResult `json:"results"`
	TotalGross      float64                  `json:"total_gross"`
	TotalNet        float64                  `json:"total_net"`
}

// PayrollCalculationResult for individual calculation results in bulk
type PayrollCalculationResult struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	GrossIncome  float64   `json:"gross_income,omitempty"`
	NetIncome    float64   `json:"net_income,omitempty"`
}

// PayslipRequest for payslip generation
type PayslipRequest struct {
	EmployeeID    uuid.UUID `json:"employee_id" binding:"required"`
	PayrollPeriodID uuid.UUID `json:"payroll_period_id" binding:"required"`
	Format        string    `json:"format" binding:"required,oneof=pdf xml html"`
}

// CreatePayrollPeriodRequest for creating a new payroll period
type CreatePayrollPeriodRequest struct {
	PeriodCode   string    `json:"period_code" binding:"required"`
	Year         int       `json:"year" binding:"required,gt=0"`
	PeriodNumber int       `json:"period_number" binding:"required,gt=0"`
	Frequency    string    `json:"frequency" binding:"required,oneof=weekly biweekly monthly"`
	PeriodType   string    `json:"period_type" binding:"required,oneof=weekly biweekly monthly"`
	StartDate    time.Time `json:"start_date" binding:"required"`
	EndDate      time.Time `json:"end_date" binding:"required"`
	PaymentDate  time.Time `json:"payment_date" binding:"required"`
	Description  string    `json:"description"`
}

// PayrollReportRequest for generating reports
type PayrollReportRequest struct {
	ReportType      string    `json:"report_type" binding:"required"`
	PayrollPeriodID uuid.UUID `json:"payroll_period_id" binding:"required"`
	Format          string    `json:"format" binding:"omitempty,oneof=pdf excel csv json"`
	CollarType      string    `json:"collar_type" binding:"omitempty,oneof=white_collar blue_collar gray_collar all"`
	GroupByCollar   bool      `json:"group_by_collar"`
}

// FinancialExportRequest for financial team exports
type FinancialExportRequest struct {
	PayrollPeriodID uuid.UUID `json:"payroll_period_id" binding:"required"`
	Format          string    `json:"format" binding:"required,oneof=pdf excel csv"`
	GroupByCollar   bool      `json:"group_by_collar"`
	IncludeSummary  bool      `json:"include_summary"`
}

// CollarTypeSummary for grouped payroll data by collar type
type CollarTypeSummary struct {
	CollarType          string  `json:"collar_type"`
	CollarTypeLabel     string  `json:"collar_type_label"`
	EmployeeCount       int     `json:"employee_count"`
	TotalGross          float64 `json:"total_gross"`
	TotalDeductions     float64 `json:"total_deductions"`
	TotalNet            float64 `json:"total_net"`
	TotalISR            float64 `json:"total_isr"`
	TotalIMSS           float64 `json:"total_imss"`
	TotalInfonavit      float64 `json:"total_infonavit"`
	EmployerIMSS        float64 `json:"employer_imss"`
	EmployerInfonavit   float64 `json:"employer_infonavit"`
	EmployerRetirement  float64 `json:"employer_retirement"`
}

// FinancialExportResponse for financial team export results
type FinancialExportResponse struct {
	PeriodCode        string              `json:"period_code"`
	PeriodStartDate   string              `json:"period_start_date"`
	PeriodEndDate     string              `json:"period_end_date"`
	PaymentDate       string              `json:"payment_date"`
	TotalEmployees    int                 `json:"total_employees"`
	GrandTotalGross   float64             `json:"grand_total_gross"`
	GrandTotalNet     float64             `json:"grand_total_net"`
	GrandTotalEmployer float64            `json:"grand_total_employer"`
	CollarSummaries   []CollarTypeSummary `json:"collar_summaries"`
	GeneratedAt       string              `json:"generated_at"`
}

// PayrollReportResponse for report history
type PayrollReportResponse struct {
	ReportType    string               `json:"report_type"`
	PeriodID      uuid.UUID            `json:"period_id"`
	PeriodCode    string               `json:"period_code,omitempty"`
	EmployeeCount int                  `json:"employee_count"`
	Totals        map[string]float64 `json:"totals"`
	GeneratedAt   time.Time            `json:"generated_at"`
	DownloadURL   string               `json:"download_url,omitempty"`
}

// PayrollSummary represents a summary of a payroll period
type PayrollSummary struct {
	PeriodID              string                     `json:"period_id"`
	TotalGross            float64                    `json:"total_gross"`
	TotalDeductions       float64                    `json:"total_deductions"`
	TotalNet              float64                    `json:"total_net"`
	EmployerContributions float64                    `json:"employer_contributions"`
	Employees             []PayrollSummaryEmployee `json:"employees"`
}

// PayrollSummaryEmployee represents an employee's summary in a payroll period
type PayrollSummaryEmployee struct {
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Gross        float64 `json:"gross"`
	Deductions   float64 `json:"deductions"`
	Net          float64 `json:"net"`
	Status       string  `json:"status"`
}

// PayrollConceptTotal represents the total amount for a specific payroll concept across a period.
type PayrollConceptTotal struct {
	Concept string  `json:"concept"`
	Total   float64 `json:"total"`
}
