/*
Package repositories - Pre-Payroll Metrics Storage Layer

==============================================================================
FILE: internal/repositories/prenomina_repository.go
==============================================================================

DESCRIPTION:
    Stores and retrieves pre-payroll metrics which are intermediate calculation
    results before final payroll processing. These metrics include worked days,
    attendance records, calculated incidences, and preliminary amounts. The
    prenomina phase allows for review and corrections before finalizing payroll.
    This repository provides CRUD operations and paginated queries.

USER PERSPECTIVE:
    - When users preview payroll before processing, they see data from this
      repository
    - Allows corrections and adjustments before committing to final payroll
    - Supports "what-if" scenarios and payroll validation workflows

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding calculation fields, implementing bulk update
        operations, adding validation logic before persistence
    ⚠️  CAUTION: Prenomina data must stay synchronized with employee and
        period data - implement transactional updates when related data changes
    ❌  DO NOT modify: Existing prenomina records after they've been promoted
        to final payroll - maintain separation between draft and final data
    📝  Best practices: Implement versioning or audit trails for prenomina
        changes; consider adding approval workflows; use pagination for
        large datasets to avoid memory issues

SYNTAX EXPLANATION:
    - PrenominaRepository: Main struct holding the GORM database connection
    - FindByEmployeeAndPeriod(employeeID, periodID uuid.UUID): Retrieves
      prenomina metrics for a specific employee in a specific period
    - Create(metrics *models.PrenominaMetric): Inserts new prenomina record
    - Update(metrics *models.PrenominaMetric): Updates existing prenomina
      using Save() which updates all fields
    - FindByPeriod(periodID uuid.UUID, page, pageSize int): Paginated query
      returning metrics and total count for a period
    - offset := (page - 1) * pageSize: Standard pagination offset calculation
    - query.Count(&total): GORM method to get total record count before
      applying pagination

==============================================================================
*/

package repositories

import (
	"github.com/google/uuid"
	"backend/internal/models"
	"gorm.io/gorm"
)

type PrenominaRepository struct {
	db *gorm.DB
}

func NewPrenominaRepository(db *gorm.DB) *PrenominaRepository {
	return &PrenominaRepository{db: db}
}

func (r *PrenominaRepository) FindByEmployeeAndPeriod(employeeID, periodID uuid.UUID) (*models.PrenominaMetric, error) {
	var metrics models.PrenominaMetric
	err := r.db.Where("employee_id = ? AND payroll_period_id = ?", employeeID, periodID).First(&metrics).Error
	return &metrics, err
}

func (r *PrenominaRepository) Create(metrics *models.PrenominaMetric) error {
	return r.db.Create(metrics).Error
}

func (r *PrenominaRepository) Update(metrics *models.PrenominaMetric) error {
	return r.db.Save(metrics).Error
}

func (r *PrenominaRepository) FindByPeriod(periodID uuid.UUID, page, pageSize int) ([]models.PrenominaMetric, int64, error) {
	var metrics []models.PrenominaMetric
	var total int64
	
	query := r.db.Model(&models.PrenominaMetric{}).Where("payroll_period_id = ?", periodID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Find(&metrics).Error

	return metrics, total, err
}
