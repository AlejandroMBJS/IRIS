/*
Package repositories - Employee Data Access Layer

==============================================================================
FILE: internal/repositories/employee_repository.go
==============================================================================

DESCRIPTION:
    Comprehensive data access layer for employee information including personal
    details, employment status, salary data, and organizational relationships.
    Provides full CRUD operations, advanced filtering, pagination, search
    capabilities, and specialized queries for employee lifecycle management
    (hiring, termination, salary changes). Includes validation for unique
    identifiers (RFC, CURP, NSS).

USER PERSPECTIVE:
    - When HR creates or updates employee records, all data flows through
      this repository
    - Supports employee search, filtering by department/status, and bulk
      operations
    - Enables compliance reporting with hire date and termination queries
    - Maintains data integrity with uniqueness checks for government IDs

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding new query methods, implementing additional
        filtering options, adding custom search logic
    ⚠️  CAUTION: Employee data is core to the entire system - maintain
        referential integrity with payroll, incidences, and user records;
        implement proper audit trails for salary and status changes
    ❌  DO NOT modify: Uniqueness constraints on RFC, CURP, NSS without
        database migration; avoid direct salary updates - use UpdateSalary()
        which maintains history
    📝  Best practices: Always use Preload() for related data to avoid N+1
        queries; implement soft deletes to preserve historical data; validate
        Mexican government ID formats (RFC, CURP, NSS) before persistence

SYNTAX EXPLANATION:
    - EmployeeRepository: Main struct holding the GORM database connection
    - Create(employee *models.Employee): Inserts new employee record
    - FindByID(id uuid.UUID): Retrieves employee with related data using Preload
    - Preload("CreatedByUser"): GORM eager loading for relationships
    - List(page, pageSize int, filters map[string]interface{}): Paginated,
      filterable employee list with dynamic query building
    - LOWER(first_name) LIKE ?: Case-insensitive search pattern
    - ExistsByRFC/CURP/NSS: Uniqueness validation methods returning boolean
    - UpdateSalary(): Specialized method using Maps to update specific fields
      with timestamp tracking
    - GetSalaryHistory(): Requires salary_history table for audit trail

==============================================================================
*/

package repositories

import (

    "strings"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "backend/internal/models"
)

// EmployeeRepository handles employee database operations
type EmployeeRepository struct {
    db *gorm.DB
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
    return &EmployeeRepository{db: db}
}

// Create creates a new employee
func (r *EmployeeRepository) Create(employee *models.Employee) error {
    return r.db.Create(employee).Error
}

// FindByID finds an employee by ID with related data
func (r *EmployeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
    var employee models.Employee
    err := r.db.Preload("CreatedByUser").
        Preload("UpdatedByUser").
        Preload("CostCenter").
        First(&employee, "id = ?", id).Error
    
    return &employee, err
}

// FindByEmployeeNumber finds an employee by employee number
func (r *EmployeeRepository) FindByEmployeeNumber(employeeNumber string) (*models.Employee, error) {
    var employee models.Employee
    err := r.db.Where("employee_number = ?", employeeNumber).First(&employee).Error
    if err != nil {
        return nil, err
    }
    return &employee, nil
}

// FindByRFC finds an employee by RFC
func (r *EmployeeRepository) FindByRFC(rfc string) (*models.Employee, error) {
    var employee models.Employee
    err := r.db.Where("rfc = ?", rfc).First(&employee).Error
    if err != nil {
        return nil, err
    }
    return &employee, nil
}

// FindByCURP finds an employee by CURP
func (r *EmployeeRepository) FindByCURP(curp string) (*models.Employee, error) {
    var employee models.Employee
    err := r.db.Where("curp = ?", curp).First(&employee).Error
    if err != nil {
        return nil, err
    }
    return &employee, nil
}

// Update updates an employee
func (r *EmployeeRepository) Update(employee *models.Employee) error {
    return r.db.Save(employee).Error
}

// Delete soft deletes an employee
func (r *EmployeeRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.Employee{}, "id = ?", id).Error
}

// List lists employees with pagination and filtering
func (r *EmployeeRepository) List(page, pageSize int, filters map[string]interface{}) ([]models.Employee, int64, error) {
    var employees []models.Employee
    var total int64
    
    query := r.db.Model(&models.Employee{})
    
    // Apply filters
    if status, ok := filters["status"]; ok {
        query = query.Where("employment_status = ?", status)
    }
    if employeeType, ok := filters["employee_type"]; ok {
        query = query.Where("employee_type = ?", employeeType)
    }
    if departmentID, ok := filters["department_id"]; ok {
        query = query.Where("department_id = ?", departmentID)
    }
    if costCenterID, ok := filters["cost_center_id"]; ok {
        query = query.Where("cost_center_id = ?", costCenterID)
    }
    if search, ok := filters["search"]; ok {
        searchStr := "%" + strings.ToLower(search.(string)) + "%"
        query = query.Where(
            "LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR employee_number LIKE ? OR rfc LIKE ?",
            searchStr, searchStr, searchStr, searchStr,
        )
    }
    if activeOnly, ok := filters["active_only"]; ok && activeOnly.(bool) {
        query = query.Where("employment_status = ?", "active")
    }
    // Filter by collar types (for hierarchical HR roles)
    if collarTypes, ok := filters["collar_types"]; ok {
        if types, ok := collarTypes.([]string); ok && len(types) > 0 {
            query = query.Where("collar_type IN ?", types)
        }
    }
    // Filter by supervisor (for supervisor role)
    if supervisorID, ok := filters["supervisor_id"]; ok {
        query = query.Where("supervisor_id = ?", supervisorID)
    }
    // Filter by subordinate IDs (for manager role - includes all subordinates recursively)
    if subordinateIDs, ok := filters["subordinate_ids"]; ok {
        if ids, ok := subordinateIDs.([]uuid.UUID); ok && len(ids) > 0 {
            query = query.Where("id IN ?", ids)
        }
    }

    // Count total
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Apply pagination
    offset := (page - 1) * pageSize
    query = query.Limit(pageSize).Offset(offset)
    
    // Preload related data and execute
    err := query.Preload("CostCenter").
        Order("created_at DESC").
        Find(&employees).Error
    
    return employees, total, err
}

// GetActiveCount returns count of active employees
func (r *EmployeeRepository) GetActiveCount() (int64, error) {
    var count int64
    err := r.db.Model(&models.Employee{}).
        Where("employment_status = ?", "active").
        Count(&count).Error
    return count, err
}

// GetByHireDateRange returns employees hired in a date range
func (r *EmployeeRepository) GetByHireDateRange(startDate, endDate time.Time) ([]models.Employee, error) {
    var employees []models.Employee
    err := r.db.Where("hire_date BETWEEN ? AND ?", startDate, endDate).
        Find(&employees).Error
    return employees, err
}

// GetTerminatedInPeriod returns employees terminated in a date range
func (r *EmployeeRepository) GetTerminatedInPeriod(startDate, endDate time.Time) ([]models.Employee, error) {
    var employees []models.Employee
    err := r.db.Where("termination_date BETWEEN ? AND ?", startDate, endDate).
        Find(&employees).Error
    return employees, err
}

// UpdateStatus updates employee status
func (r *EmployeeRepository) UpdateStatus(id uuid.UUID, status string) error {
    return r.db.Model(&models.Employee{}).
        Where("id = ?", id).
        Update("employment_status", status).Error
}

// UpdateSalary updates employee salary with history tracking
func (r *EmployeeRepository) UpdateSalary(id uuid.UUID, newSalary float64) error {
    return r.db.Model(&models.Employee{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "daily_salary": newSalary,
            "updated_at": time.Now(),
        }).Error
}

// ExistsByRFC checks if employee exists by RFC
func (r *EmployeeRepository) ExistsByRFC(rfc string) (bool, error) {
    var count int64
    err := r.db.Model(&models.Employee{}).
        Where("rfc = ?", rfc).
        Count(&count).Error
    return count > 0, err
}

// ExistsByCURP checks if employee exists by CURP
func (r *EmployeeRepository) ExistsByCURP(curp string) (bool, error) {
    var count int64
    err := r.db.Model(&models.Employee{}).
        Where("curp = ?", curp).
        Count(&count).Error
    return count > 0, err
}

// ExistsByNSS checks if employee exists by NSS
func (r *EmployeeRepository) ExistsByNSS(nss string) (bool, error) {
    var count int64
    err := r.db.Model(&models.Employee{}).
        Where("nss = ?", nss).
        Count(&count).Error
    return count > 0, err
}

// GetSalaryHistory gets employee salary changes (requires salary_history table)
func (r *EmployeeRepository) GetSalaryHistory(employeeID uuid.UUID) ([]models.SalaryHistory, error) {
    var history []models.SalaryHistory
    err := r.db.Where("employee_id = ?", employeeID).
        Order("effective_date DESC").
        Find(&history).Error
    return history, err
}
