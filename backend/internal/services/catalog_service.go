/*
Package services - Catalog Service

==============================================================================
FILE: internal/services/catalog_service.go
==============================================================================

DESCRIPTION:
    Manages payroll catalogs including payroll concepts (perceptions and deductions)
    and incidence types used across the payroll system.

USER PERSPECTIVE:
    - View and manage payroll concepts (salary, overtime, bonuses, taxes)
    - Configure which concepts are taxable or affect IMSS
    - Manage incidence types for HR operations

DEVELOPER GUIDELINES:
    OK to modify: Add new catalog types, extend concept properties
    CAUTION: Changing concept categories affects payroll calculations
    DO NOT modify: Core concept types without updating payroll calculation logic
    Note: Payroll concepts are referenced throughout the calculation system

SYNTAX EXPLANATION:
    - PayrollConcept defines income/deduction types with tax implications
    - ConceptType: 'perception' (income) or 'deduction' (withholding)
    - IsTaxable: indicates if concept is subject to ISR
    - IsIMSSBase: indicates if concept is included in IMSS calculation base

==============================================================================
*/
package services

import (
	"gorm.io/gorm"
	"backend/internal/dtos"
	"backend/internal/models"
	"backend/internal/repositories"
)

type CatalogService struct {
	repo *repositories.CatalogRepository
}

func NewCatalogService(db *gorm.DB) *CatalogService {
	return &CatalogService{
		repo: repositories.NewCatalogRepository(db),
	}
}

func (s *CatalogService) GetPayrollConcepts() ([]models.PayrollConcept, error) {
	return s.repo.GetPayrollConcepts()
}

func (s *CatalogService) GetIncidenceTypes() ([]models.IncidenceType, error) {
	return s.repo.GetIncidenceTypes()
}

func (s *CatalogService) CreatePayrollConcept(req dtos.CreatePayrollConceptRequest) (*models.PayrollConcept, error) {
	concept := &models.PayrollConcept{
		Name:                   req.Name,
		Category:               req.Category,
		ConceptType:            req.ConceptType,
		IsTaxable:              req.IsTaxable,
		IsIMSSBase:             req.IsIMSSBase,
		IsIntegratedSalary:     req.IsIntegratedSalary,
	}

	err := s.repo.CreatePayrollConcept(concept)
	if err != nil {
		return nil, err
	}
	return concept, nil
}
