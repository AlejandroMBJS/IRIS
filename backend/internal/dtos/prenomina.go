/*
Package dtos - Pre-Payroll (Prenomina) Data Transfer Objects

==============================================================================
FILE: internal/dtos/prenomina.go
==============================================================================

DESCRIPTION:
    Defines structures for pre-payroll metrics used in prenomina calculations.
    Prenomina is the preliminary payroll calculation phase where work hours,
    absences, and other metrics are collected before final payroll processing.

USER PERSPECTIVE:
    - Shapes prenomina metrics displayed in the HR dashboard
    - Tracks worked days, overtime, absences, and delays
    - Provides bulk calculation request/response formats

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new metric fields, improve calculation details
    ⚠️  CAUTION: Changing field names (affects frontend displays)
    ❌  DO NOT modify: Core metric calculations without payroll review
    📝  Keep fields aligned with Mexican labor law requirements

SYNTAX EXPLANATION:
    - json tags: Define API response field names
    - PrenominaMetricResponse: Complete metrics for one employee/period
    - PrenominaBulkCalculateRequest: Calculate multiple employees at once

PRENOMINA WORKFLOW:
    1. HR enters work metrics (hours, absences, etc.)
    2. System calculates preliminary amounts
    3. HR reviews and approves
    4. Prenomina feeds into final payroll calculation

==============================================================================
*/
package dtos

import (
	"time"

	"github.com/google/uuid"
)

// PrenominaMetricResponse represents prenomina metrics for an employee in a period
type PrenominaMetricResponse struct {
	ID                    uuid.UUID  `json:"id"`
	EmployeeID            uuid.UUID  `json:"employee_id"`
	EmployeeName          string     `json:"employee_name"`
	EmployeeNumber        string     `json:"employee_number"`
	PayrollPeriodID       uuid.UUID  `json:"payroll_period_id"`
	PeriodCode            string     `json:"period_code"`

	// Work metrics
	WorkedDays            float64    `json:"worked_days"`
	RegularHours          float64    `json:"regular_hours"`
	OvertimeHours         float64    `json:"overtime_hours"`
	DoubleOvertimeHours   float64    `json:"double_overtime_hours"`
	TripleOvertimeHours   float64    `json:"triple_overtime_hours"`

	// Leave metrics
	AbsenceDays           float64    `json:"absence_days"`
	SickDays              float64    `json:"sick_days"`
	VacationDays          float64    `json:"vacation_days"`
	UnpaidLeaveDays       float64    `json:"unpaid_leave_days"`

	// Other metrics
	DelaysCount           int        `json:"delays_count"`
	DelayMinutes          float64    `json:"delay_minutes"`
	DelayDeduction        float64    `json:"delay_deduction"`
	EarlyDeparturesCount  int        `json:"early_departures_count"`

	// Monetary amounts
	RegularSalary         float64    `json:"regular_salary"`
	OvertimeAmount        float64    `json:"overtime_amount"`
	DoubleOvertimeAmount  float64    `json:"double_overtime_amount"`
	TripleOvertimeAmount  float64    `json:"triple_overtime_amount"`
	BonusAmount           float64    `json:"bonus_amount"`
	CommissionAmount      float64    `json:"commission_amount"`
	OtherExtraAmount      float64    `json:"other_extra_amount"`
	TotalExtras           float64    `json:"total_extras"`
	LoanDeduction         float64    `json:"loan_deduction"`
	AdvanceDeduction      float64    `json:"advance_deduction"`
	OtherDeduction        float64    `json:"other_deduction"`
	TotalDeductions       float64    `json:"total_deductions"`

	// Summary
	GrossIncome           float64    `json:"gross_income"`
	NetIncome             float64    `json:"net_income"`

	// Metadata
	CalculationStatus     string     `json:"calculation_status"` // e.g., "calculated", "approved", "rejected"
	CalculatedAt          time.Time  `json:"calculated_at"`
	CalculatedBy          string     `json:"calculated_by"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// PrenominaBulkCalculateRequest for bulk calculation
type PrenominaBulkCalculateRequest struct {
	PayrollPeriodID uuid.UUID   `json:"payroll_period_id" binding:"required"`
	EmployeeIDs     []uuid.UUID `json:"employee_ids"`
	CalculateAll    bool        `json:"calculate_all"` // if true, calculate for all active employees
}

// PrenominaBulkCalculateResponse for bulk calculation results
type PrenominaBulkCalculateResponse struct {
	PeriodCode      string                     `json:"period_code"`
	TotalCalculated int                        `json:"total_calculated"`
	TotalSuccess    int                        `json:"total_success"`
	TotalFailed     int                        `json:"total_failed"`
	Results         []PrenominaCalculationResult `json:"results"`
	TotalGross      float64                    `json:"total_gross"`
	TotalNet        float64                    `json:"total_net"`
}

// PrenominaCalculationResult for individual calculation results in bulk
type PrenominaCalculationResult struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	GrossIncome  float64   `json:"gross_income,omitempty"`
	NetIncome    float64   `json:"net_income,omitempty"`
}

// PrenominaListResponse for listing prenomina metrics
type PrenominaListResponse struct {
	Metrics    []PrenominaMetricResponse `json:"metrics"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}
