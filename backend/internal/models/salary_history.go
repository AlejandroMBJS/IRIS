/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/salary_history.go
==============================================================================

DESCRIPTION:
    Tracks historical changes to employee salaries. Every time an employee's
    DailySalary changes, a record is created here for audit and reporting.

USER PERSPECTIVE:
    - Visible in employee profile under "Historial de Salario"
    - Shows timeline of salary increases/decreases
    - Includes reason for change and who made it
    - Useful for annual reviews and compliance audits

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add additional fields (e.g., percentage change)
    ⚠️  CAUTION: This is audit data - treat as append-only
    ❌  DO NOT modify: Existing records should never be updated/deleted
    📝  Create new records in employee service when salary changes

SYNTAX EXPLANATION:
    - EffectiveDate: When the new salary took effect
    - OldDailySalary/NewDailySalary: Before and after values
    - RecordedBy: User who made the change (audit trail)
    - *uuid.UUID for RecordedBy: Pointer allows NULL for system changes

USAGE:
    - Called from employee.Update() when DailySalary changes
    - Used in annual salary reports
    - Important for retroactive calculations if needed

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
)

// SalaryHistory records changes in an employee's salary.
type SalaryHistory struct {
	BaseModel
	EmployeeID    uuid.UUID  `gorm:"type:text;not null" json:"employee_id"`
	EffectiveDate time.Time  `gorm:"type:date;not null" json:"effective_date"`
	OldDailySalary float64    `gorm:"type:decimal(12,2);not null" json:"old_daily_salary"`
	NewDailySalary float64    `gorm:"type:decimal(12,2);not null" json:"new_daily_salary"`
	Reason        string     `gorm:"type:text" json:"reason,omitempty"`
	RecordedBy    *uuid.UUID `gorm:"type:text" json:"recorded_by,omitempty"` // User who made the change
	Employee      *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	RecordedByUser *User     `gorm:"foreignKey:RecordedBy" json:"recorded_by_user,omitempty"`
}

// TableName specifies the table name
func (SalaryHistory) TableName() string {
	return "salary_history"
}
