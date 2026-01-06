/*
Package repositories - HR Incidence Data Access Layer

==============================================================================
FILE: internal/repositories/incidence_repository.go
==============================================================================

DESCRIPTION:
    Manages HR incidence data including employee absences, vacations, overtime,
    bonuses, and other special payroll events. Incidences affect payroll
    calculations and are tied to specific employees and payroll periods.
    This repository provides querying capabilities with related data preloading.

USER PERSPECTIVE:
    - When HR records an employee absence or vacation, data flows through
      this repository
    - Supports payroll processing by providing incidence data that affects
      pay calculations (e.g., unpaid absences reduce pay, overtime increases it)
    - Enables reporting on employee attendance and special pay events

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding query methods for different incidence filtering
        (by date range, type, status), implementing bulk operations
    ⚠️  CAUTION: Incidences directly impact payroll calculations - ensure
        data integrity and consider implementing audit trails for changes
    ❌  DO NOT modify: The relationship between incidences and payroll periods
        without updating dependent calculation logic
    📝  Best practices: Always preload IncidenceType when querying to avoid
        N+1 query problems; validate date ranges fall within payroll periods

SYNTAX EXPLANATION:
    - IncidenceRepository: Main struct holding the GORM database connection
    - FindByEmployeeAndPeriod(employeeID, periodID uuid.UUID): Retrieves all
      incidences for a specific employee in a specific payroll period
    - Preload("IncidenceType"): GORM eager loading - fetches related incidence
      type data in a single query to avoid N+1 problem
    - Where("employee_id = ? AND payroll_period_id = ?"): Compound WHERE clause
      with multiple parameterized conditions

==============================================================================
*/

package repositories

import (
	"github.com/google/uuid"
	"backend/internal/models"
	"gorm.io/gorm"
)

type IncidenceRepository struct {
	db *gorm.DB
}

func NewIncidenceRepository(db *gorm.DB) *IncidenceRepository {
	return &IncidenceRepository{db: db}
}

func (r *IncidenceRepository) FindByEmployeeAndPeriod(employeeID, periodID uuid.UUID) ([]models.Incidence, error) {
	var incidences []models.Incidence
	err := r.db.Preload("IncidenceType").Where("employee_id = ? AND payroll_period_id = ?", employeeID, periodID).Find(&incidences).Error
	return incidences, err
}
