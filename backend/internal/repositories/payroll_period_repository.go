/*
Package repositories - Payroll Period Management Data Access Layer

==============================================================================
FILE: internal/repositories/payroll_period_repository.go
==============================================================================

DESCRIPTION:
    Manages payroll periods which define the time spans for payroll processing
    (weekly, biweekly, monthly, etc.). Each period has a status (open, closed,
    processing) and serves as the temporal boundary for all payroll calculations,
    incidences, and payments. This repository handles CRUD operations and
    filtering by year and status.

USER PERSPECTIVE:
    - When users select a pay period to process payroll, this repository
      provides the available periods
    - Period status controls whether payroll can be calculated or modified
    - Closed periods are locked to prevent retroactive changes

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding new query methods for period search, implementing
        status transition logic, adding date range queries
    ⚠️  CAUTION: Period status changes affect system-wide payroll operations -
        implement proper locking and validation before status updates
    ❌  DO NOT modify: Period dates after payroll has been processed - this
        would invalidate calculations and compliance reporting
    📝  Best practices: Implement status state machine to prevent invalid
        transitions (e.g., can't reopen a closed period without admin approval);
        consider adding audit logs for period status changes

SYNTAX EXPLANATION:
    - PayrollPeriodRepository: Main struct holding the GORM database connection
    - FindByID(id uuid.UUID): Retrieves a specific payroll period by ID
    - GetPeriods(filters map[string]interface{}): Flexible query accepting
      dynamic filters (year, status, etc.)
    - if year, ok := filters["year"]; ok: Go idiom for safely checking if
      a map key exists and extracting its value
    - query = query.Where(...): GORM method chaining - builds query progressively
    - Create(period *models.PayrollPeriod): Inserts new payroll period record

==============================================================================
*/

package repositories

import (
	"time"

	"github.com/google/uuid"
	"backend/internal/models"
	"gorm.io/gorm"
)

type PayrollPeriodRepository struct {
	db *gorm.DB
}

func NewPayrollPeriodRepository(db *gorm.DB) *PayrollPeriodRepository {
	return &PayrollPeriodRepository{db: db}
}

func (r *PayrollPeriodRepository) FindByID(id uuid.UUID) (*models.PayrollPeriod, error) {
	var period models.PayrollPeriod
	err := r.db.First(&period, "id = ?", id).Error
	return &period, err
}

func (r *PayrollPeriodRepository) GetPeriods(filters map[string]interface{}) ([]models.PayrollPeriod, error) {
	var periods []models.PayrollPeriod
	query := r.db.Model(&models.PayrollPeriod{})
	// Apply filters if any
	if year, ok := filters["year"]; ok {
		query = query.Where("year = ?", year)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&periods).Error
	return periods, err
}

func (r *PayrollPeriodRepository) Create(period *models.PayrollPeriod) error {
	return r.db.Create(period).Error
}

func (r *PayrollPeriodRepository) FindByPeriodCode(periodCode string) (*models.PayrollPeriod, error) {
	var period models.PayrollPeriod
	err := r.db.First(&period, "period_code = ?", periodCode).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

// HasOverlappingPeriod checks if there's an overlapping period for the given frequency and dates
func (r *PayrollPeriodRepository) HasOverlappingPeriod(frequency string, startDate, endDate time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.PayrollPeriod{}).
		Where("frequency = ?", frequency).
		Where("(start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?)",
			endDate, startDate, startDate, endDate).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
