/*
Package dtos - Catalog Data Transfer Objects

==============================================================================
FILE: internal/dtos/catalog.go
==============================================================================

DESCRIPTION:
    Defines structures for payroll catalog management including payroll
    concepts (perceptions, deductions, employer contributions). These
    catalogs are used to configure what items appear on payslips.

USER PERSPECTIVE:
    - Configures what income/deduction types are available
    - Controls which concepts are taxable or affect IMSS
    - Links concepts to SAT codes for CFDI compliance

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new catalog types
    ⚠️  CAUTION: Changing category/type enums (affects existing data)
    ❌  DO NOT modify: SAT code handling without fiscal review
    📝  Validate against SAT catalog for CFDI generation

SYNTAX EXPLANATION:
    - Category: income, deduction, employer_contribution, benefit
    - ConceptType: fixed (same each period) or variable
    - IsTaxable: Affects ISR calculation
    - IsIMSSBase: Affects IMSS contribution calculation
    - IsIntegratedSalary: Included in SDI calculation

SAT COMPLIANCE:
    - SATCode must match official SAT catalogs for CFDI
    - Categories map to SAT perception/deduction types

==============================================================================
*/
package dtos

// CreatePayrollConceptRequest represents data for creating a new payroll concept
type CreatePayrollConceptRequest struct {
	Name               string  `json:"name" binding:"required"`
	Category           string  `json:"category" binding:"required,oneof=income deduction employer_contribution benefit"`
	ConceptType        string  `json:"concept_type" binding:"required,oneof=fixed variable"`
	IsTaxable          bool    `json:"is_taxable"`
	IsIMSSBase         bool    `json:"is_imss_base"`
	IsIntegratedSalary bool    `json:"is_integrated_salary"`
	SATCode            string  `json:"sat_code"`
	Description        string  `json:"description"`
}
