/*
Package repositories - Payroll Calculation Storage Layer

==============================================================================
FILE: internal/repositories/payroll_repository.go
==============================================================================

DESCRIPTION:
    Manages finalized payroll calculation data including gross pay, deductions,
    net pay, and employer contributions (IMSS, INFONAVIT, etc.). This repository
    handles the storage and retrieval of completed payroll calculations for
    individual employees or entire payroll periods. Includes comprehensive
    relationship preloading for complete payroll record access.

USER PERSPECTIVE:
    - When users finalize payroll processing, calculations are stored via
      this repository
    - Enables payroll reports, pay stub generation, and historical payroll
      queries
    - Provides data for government reporting (IMSS, SAT declarations)
    - Supports employer contribution tracking for accounting integration

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding aggregate queries (totals, summaries), implementing
        export functionality, adding audit trail queries
    ⚠️  CAUTION: Payroll calculations are immutable once finalized - implement
        correction/reversal patterns instead of updates; ensure transactional
        integrity with employer contributions
    ❌  DO NOT modify: Finalized payroll records directly - use reversal and
        re-calculation workflows; never delete records needed for compliance
    📝  Best practices: Always preload relationships to avoid N+1 queries;
        implement versioning for payroll corrections; maintain referential
        integrity between PayrollCalculation and EmployerContribution

SYNTAX EXPLANATION:
    - PayrollRepository: Main struct holding the GORM database connection
    - FindByEmployeeAndPeriod(employeeID, periodID uuid.UUID): Retrieves
      complete payroll calculation for a specific employee and period
    - Preload("Employee").Preload("PayrollPeriod").Preload("EmployerContribution"):
      Chains multiple eager loading operations for full relationship graph
    - if err == gorm.ErrRecordNotFound: Explicit handling of not-found case
      (redundant here but shows error differentiation pattern)
    - FindByPeriod(periodID uuid.UUID): Gets all payroll calculations for
      a complete pay period
    - FindEmployerContributionByPayrollCalculationID(): Specialized query for
      employer-side payroll taxes and contributions (IMSS, INFONAVIT)

==============================================================================
*/

package repositories

import (
	"github.com/google/uuid"
	"backend/internal/models"
	"gorm.io/gorm"
)

type PayrollRepository struct {
	db *gorm.DB
}

func NewPayrollRepository(db *gorm.DB) *PayrollRepository {
	return &PayrollRepository{db: db}
}

func (r *PayrollRepository) FindByEmployeeAndPeriod(employeeID, periodID uuid.UUID) (*models.PayrollCalculation, error) {
	var calc models.PayrollCalculation
	err := r.db.
		Preload("Employee").
		Preload("PayrollPeriod").
		Preload("EmployerContribution").
		Where("employee_id = ? AND payroll_period_id = ?", employeeID, periodID).
		First(&calc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &calc, nil
}

func (r *PayrollRepository) FindByPeriod(periodID uuid.UUID) ([]models.PayrollCalculation, error) {
	var calcs []models.PayrollCalculation
	err := r.db.
		Preload("Employee").
		Preload("PayrollPeriod").
		Preload("EmployerContribution").
		Where("payroll_period_id = ?", periodID).
		Find(&calcs).Error
	return calcs, err
}

// FindEmployerContributionByPayrollCalculationID finds an employer contribution by payroll calculation ID
func (r *PayrollRepository) FindEmployerContributionByPayrollCalculationID(payrollCalculationID uuid.UUID) (*models.EmployerContribution, error) {
	var contrib models.EmployerContribution
	err := r.db.Where("payroll_calculation_id = ?", payrollCalculationID).First(&contrib).Error
	if err != nil {
		return nil, err
	}
	return &contrib, nil
}
