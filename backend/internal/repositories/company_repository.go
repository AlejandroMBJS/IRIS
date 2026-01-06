/*
Package repositories - Multi-Tenant Company Data Access Layer

==============================================================================
FILE: internal/repositories/company_repository.go
==============================================================================

DESCRIPTION:
    Provides data access for company (tenant) information in the multi-tenant
    payroll system. Each company represents a separate business entity with
    its own employees, payroll data, and configuration. This repository
    handles company lookups by RFC (Mexico's tax ID) and UUID.

USER PERSPECTIVE:
    - When users log in, their company context is loaded using this repository
    - All payroll processing is scoped to a specific company to ensure
      data isolation between tenants
    - Company RFC validation during onboarding uses these lookup methods

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding new query methods for company search/filtering,
        implementing company-specific configuration retrieval
    ⚠️  CAUTION: Company is the root of multi-tenancy - ensure all queries
        respect tenant boundaries and never leak data between companies
    ❌  DO NOT modify: Database connection patterns - maintain consistency
        with other repositories
    📝  Best practices: Always validate company access permissions before
        returning company data; consider caching company lookups by ID

SYNTAX EXPLANATION:
    - CompanyRepository: Main struct holding the GORM database connection
    - FindByRFC(rfc string): Looks up company by RFC (Registro Federal de
      Contribuyentes - Mexico's tax ID system)
    - FindByID(id uuid.UUID): Retrieves company by unique identifier
    - r.db.Where("rfc = ?", rfc).First(&company): GORM query with parameterized
      WHERE clause, prevents SQL injection
    - Returns (*models.Company, error): Pointer to company or error if not found

==============================================================================
*/

package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"backend/internal/models"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) FindByRFC(rfc string) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("rfc = ?", rfc).First(&company).Error
	return &company, err
}

func (r *CompanyRepository) FindByID(id uuid.UUID) (*models.Company, error) {
	var company models.Company
	err := r.db.First(&company, "id = ?", id).Error
	return &company, err
}
