/*
Package repositories - Shift Data Access Layer

==============================================================================
FILE: internal/repositories/shift_repository.go
==============================================================================

DESCRIPTION:
    Provides data access for shift management in the IRIS Payroll System.
    Handles CRUD operations for work shifts and employee shift assignments.

USER PERSPECTIVE:
    - HR creates and manages shifts for employees
    - Shifts define work schedules (start/end times, break duration)
    - Employees are assigned to shifts for scheduling purposes

DEVELOPER GUIDELINES:
    OK to modify: Add new query methods for shift filtering
    CAUTION: Ensure company_id filtering for multi-tenant security
    DO NOT modify: Core CRUD method signatures

==============================================================================
*/
package repositories

import (
	"backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

// Create creates a new shift
func (r *ShiftRepository) Create(shift *models.Shift) error {
	return r.db.Create(shift).Error
}

// FindByID finds a shift by ID
func (r *ShiftRepository) FindByID(id uuid.UUID) (*models.Shift, error) {
	var shift models.Shift
	err := r.db.First(&shift, "id = ?", id).Error
	return &shift, err
}

// FindByIDAndCompany finds a shift by ID and company (for security)
func (r *ShiftRepository) FindByIDAndCompany(id, companyID uuid.UUID) (*models.Shift, error) {
	var shift models.Shift
	err := r.db.Where("id = ? AND company_id = ?", id, companyID).First(&shift).Error
	return &shift, err
}

// FindByCode finds a shift by code within a company
func (r *ShiftRepository) FindByCode(code string, companyID uuid.UUID) (*models.Shift, error) {
	var shift models.Shift
	err := r.db.Where("code = ? AND company_id = ?", code, companyID).First(&shift).Error
	return &shift, err
}

// FindAllByCompany returns all shifts for a company
func (r *ShiftRepository) FindAllByCompany(companyID uuid.UUID) ([]models.Shift, error) {
	var shifts []models.Shift
	err := r.db.Where("company_id = ?", companyID).
		Order("display_order ASC, name ASC").
		Find(&shifts).Error
	return shifts, err
}

// FindActiveByCompany returns all active shifts for a company
func (r *ShiftRepository) FindActiveByCompany(companyID uuid.UUID) ([]models.Shift, error) {
	var shifts []models.Shift
	err := r.db.Where("company_id = ? AND is_active = ?", companyID, true).
		Order("display_order ASC, name ASC").
		Find(&shifts).Error
	return shifts, err
}

// Update updates a shift
func (r *ShiftRepository) Update(shift *models.Shift) error {
	return r.db.Save(shift).Error
}

// Delete soft-deletes a shift
func (r *ShiftRepository) Delete(id, companyID uuid.UUID) error {
	return r.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.Shift{}).Error
}

// CountEmployeesWithShift returns the number of employees assigned to a shift
func (r *ShiftRepository) CountEmployeesWithShift(shiftID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Employee{}).Where("shift_id = ?", shiftID).Count(&count).Error
	return count, err
}

// ExistsByCodeAndCompany checks if a shift code already exists for a company
func (r *ShiftRepository) ExistsByCodeAndCompany(code string, companyID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.Model(&models.Shift{}).Where("code = ? AND company_id = ?", code, companyID)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
