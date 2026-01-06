/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/payroll_period.go
==============================================================================

DESCRIPTION:
    Defines the PayrollPeriod model which represents a time window for payroll
    processing. Each period has a type (weekly, biweekly, monthly), date range,
    and status workflow that tracks the payroll processing lifecycle.

USER PERSPECTIVE:
    - Periods appear in the "Periodos de Nomina" section
    - Users create periods to group payroll calculations for a date range
    - Period types determine which employees are included:
        * Weekly: Blue collar and gray collar workers
        * Biweekly: White collar workers
        * Monthly: Special cases
    - Status shows where the period is in the processing workflow

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new status values (update check constraint)
    ⚠️  CAUTION: Status transition logic in Close() method
    ❌  DO NOT modify: PeriodCode format validation (breaks existing data)
    📝  Period codes follow format: YYYY-BW01 (biweekly), YYYY-W01 (weekly), YYYY-M01 (monthly)

SYNTAX EXPLANATION:
    - check:status IN (...): Database constraint for valid statuses
    - uniqueIndex on PeriodCode: Only one period per code
    - Validate(): Called in BeforeSave to ensure data integrity
    - Close(): Business logic for transitioning period to closed

STATUS WORKFLOW:
    open → calculated → approved → paid → closed
                    ↘ cancelled (can happen from any state before paid)

BUSINESS RULES:
    - Period can only be closed after being paid
    - StartDate must be before EndDate
    - EndDate must be before PaymentDate
    - IsFiscalClosing marks periods for tax reporting

==============================================================================
*/
package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// PayrollPeriod represents a payroll calculation period

type PayrollPeriod struct {

    BaseModel

    

    // Identification

    PeriodCode  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"period_code"`

    Year        int       `gorm:"not null" json:"year"`

    PeriodNumber int      `gorm:"not null" json:"period_number"`

    Frequency   string    `gorm:"type:varchar(20);not null" json:"frequency"`

    PeriodType  string    `gorm:"type:varchar(20);not null;check:period_type IN ('weekly','biweekly','monthly')" json:"period_type"`

    

    // Dates

    StartDate   time.Time `gorm:"type:date;not null" json:"start_date"`

    EndDate     time.Time `gorm:"type:date;not null" json:"end_date"`

    PaymentDate time.Time `gorm:"type:date;not null" json:"payment_date"`

    

    // Status

    Status      string    `gorm:"type:varchar(20);default:'open';check:status IN ('open','calculated','approved','paid','closed','cancelled')" json:"status"`

    

    // Financial totals

    TotalGross          float64 `gorm:"type:decimal(15,2);default:0" json:"total_gross"`

    TotalDeductions     float64 `gorm:"type:decimal(15,2);default:0" json:"total_deductions"`

    TotalNet            float64 `gorm:"type:decimal(15,2);default:0" json:"total_net"`

    TotalEmployerContributions float64 `gorm:"type:decimal(15,2);default:0" json:"total_employer_contributions"`

    

    // Metadata

    Description string    `gorm:"type:text" json:"description,omitempty"`

    IsFiscalClosing bool `gorm:"default:false" json:"is_fiscal_closing"`

    

    // Foreign Keys

    CreatedBy   *uuid.UUID `gorm:"type:text" json:"created_by,omitempty"`

    ClosedBy    *uuid.UUID `gorm:"type:text" json:"closed_by,omitempty"`

    

    // Relations

    CreatedByUser *User `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`

    ClosedByUser  *User `gorm:"foreignKey:ClosedBy" json:"closed_by_user,omitempty"`

    Incidences    []Incidence `gorm:"foreignKey:PayrollPeriodID" json:"incidences,omitempty"`

    PrenominaMetrics []PrenominaMetric `gorm:"foreignKey:PayrollPeriodID" json:"prenomina_metrics,omitempty"`

    PayrollHeaders []PayrollHeader `gorm:"foreignKey:PayrollPeriodID" json:"payroll_headers,omitempty"`

    

    // Audit

    ClosedAt    *time.Time `json:"closed_at,omitempty"`

}

// TableName specifies the table name
func (PayrollPeriod) TableName() string {
    return "payroll_periods"
}

// Validate validates payroll period data
func (pp *PayrollPeriod) Validate() error {
    var validationErrors []string
    
    // Period code validation
    periodCodeRegex := regexp.MustCompile(`^\d{4}-(BW\d{2}|M\d{2}|W\d{2})$`)
    if !periodCodeRegex.MatchString(pp.PeriodCode) {
        validationErrors = append(validationErrors, "period code must be in format YYYY-BW01, YYYY-M01, or YYYY-W01")
    }
    
    // Date validation
    if pp.StartDate.IsZero() {
        validationErrors = append(validationErrors, "start date is required")
    }
    if pp.EndDate.IsZero() {
        validationErrors = append(validationErrors, "end date is required")
    }
    if pp.PaymentDate.IsZero() {
        validationErrors = append(validationErrors, "payment date is required")
    }
    
    if !pp.StartDate.Before(pp.EndDate) && !pp.StartDate.Equal(pp.EndDate) {
        validationErrors = append(validationErrors, "start date must be before or equal to end date")
    }
    if pp.EndDate.After(pp.PaymentDate) {
        validationErrors = append(validationErrors, "end date must be before or equal to payment date")
    }
    
    // Period type validation
    if pp.PeriodType != "weekly" && pp.PeriodType != "biweekly" && pp.PeriodType != "monthly" {
        validationErrors = append(validationErrors, "invalid period type")
    }
    
    if len(validationErrors) > 0 {
        return errors.New(strings.Join(validationErrors, "; "))
    }
    
    return nil
}

// GetWorkingDays returns the number of working days in the period
func (pp *PayrollPeriod) GetWorkingDays() int {
    // Placeholder for now. A real implementation would exclude weekends and holidays.
    return pp.CalculateDays()
}

// CalculateDays returns the number of days in the period
func (pp *PayrollPeriod) CalculateDays() int {
    duration := pp.EndDate.Sub(pp.StartDate)
    return int(duration.Hours()/24) + 1 // Inclusive
}

// IsOpen returns true if period is open for processing
func (pp *PayrollPeriod) IsOpen() bool {
    return pp.Status == "open"
}

// CanCalculate returns true if period can be calculated
func (pp *PayrollPeriod) CanCalculate() bool {
    return pp.Status == "open" || pp.Status == "calculated"
}

// Close closes the payroll period
func (pp *PayrollPeriod) Close(userID uuid.UUID) error {
    if pp.Status == "closed" {
        return errors.New("period is already closed")
    }
    if pp.Status != "paid" {
        return errors.New("period must be paid before closing")
    }
    
    now := time.Now()
    pp.Status = "closed"
    pp.ClosedBy = &userID
    pp.ClosedAt = &now
    
    return nil
}

// BeforeSave validates before saving
func (pp *PayrollPeriod) BeforeSave(tx *gorm.DB) error {
    return pp.Validate()
}
