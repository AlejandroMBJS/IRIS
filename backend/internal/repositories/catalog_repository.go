/*
Package repositories - Payroll Catalog Data Access Layer

==============================================================================
FILE: internal/repositories/catalog_repository.go
==============================================================================

DESCRIPTION:
    Manages payroll catalog data including payroll concepts (perception,
    deduction types) and incidence types (absence, vacation, overtime
    classifications). This repository provides read and write access to
    the system's reference data used throughout payroll processing.

USER PERSPECTIVE:
    - When users configure payroll concepts or incidence types, this
      repository ensures data is consistently stored and retrieved
    - Supports dropdown lists and validation in the UI for payroll
      configuration screens
    - Enables standardized categorization of pay components

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding new catalog query methods, implementing
        additional filtering or sorting options
    ⚠️  CAUTION: Catalog data is referenced across the system - ensure
        referential integrity when adding delete/update operations
    ❌  DO NOT modify: The GORM db instance directly - always use
        repository methods for consistency
    📝  Best practices: Cache catalog data when appropriate as it
        changes infrequently but is accessed often

SYNTAX EXPLANATION:
    - CatalogRepository: Main struct holding the GORM database connection
    - GetPayrollConcepts(): Retrieves all payroll concept definitions
      (perceptions, deductions, etc.)
    - GetIncidenceTypes(): Fetches all HR incidence type definitions
      (absence codes, vacation types, etc.)
    - CreatePayrollConcept(): Adds new payroll concept to the catalog
    - r.db.Find(&concepts).Error: GORM pattern for fetching all records,
      returns error if query fails

==============================================================================
*/

package repositories

import (
	"gorm.io/gorm"
	"backend/internal/models"
)

type CatalogRepository struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) GetPayrollConcepts() ([]models.PayrollConcept, error) {
	var concepts []models.PayrollConcept
	err := r.db.Find(&concepts).Error
	return concepts, err
}

func (r *CatalogRepository) GetIncidenceTypes() ([]models.IncidenceType, error) {
	var incidenceTypes []models.IncidenceType
	err := r.db.Find(&incidenceTypes).Error
	return incidenceTypes, err
}

func (r *CatalogRepository) CreatePayrollConcept(concept *models.PayrollConcept) error {
	return r.db.Create(concept).Error
}
