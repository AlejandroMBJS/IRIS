/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/payroll_concept.go
==============================================================================

DESCRIPTION:
    Defines PayrollConcept model - a catalog of income and deduction types
    that can appear on payroll calculations. Concepts define how items are
    classified for tax purposes and CFDI generation.

USER PERSPECTIVE:
    - Concepts are predefined items that appear on payslips
    - Examples: "Salario Base", "ISR", "IMSS", "Vales de Despensa"
    - Each concept has a category (income, deduction, benefit)
    - SAT codes link concepts to official Mexican tax classifications

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new concepts, update descriptions
    ⚠️  CAUTION: Changing SAT codes (affects CFDI compliance)
    ❌  DO NOT modify: Category values (database constraints)
    📝  This is seed/catalog data - add via migrations, not code

SYNTAX EXPLANATION:
    - Category: 'income', 'deduction', 'employer_contribution', 'benefit'
    - ConceptType: 'fixed' (same each period) or 'variable' (changes)
    - IsTaxable: Whether it counts toward ISR calculation
    - IsIMSSBase: Whether it's part of IMSS contribution base
    - IsIntegratedSalary: Whether it's part of SDI (Salario Diario Integrado)
    - SATCode: Official SAT code for CFDI Nomina compliance

MEXICAN TAX CONTEXT:
    - SAT (Servicio de Administracion Tributaria) = Mexican IRS
    - CFDI Nomina requires specific codes for each concept
    - Concepts marked IsIntegratedSalary affect SDI calculation
    - SDI is used to calculate IMSS and Infonavit contributions

==============================================================================
*/
package models

import (
	"gorm.io/gorm"
)

// PayrollConcept represents a payroll concept (e.g., "Base Salary", "ISR", "Food Vouchers").
type PayrollConcept struct {
	BaseModel
	Name               string  `gorm:"type:varchar(255);not null" json:"name"`
	Category           string  `gorm:"type:varchar(50);not null;check:category IN ('income','deduction','employer_contribution','benefit')" json:"category"`
	ConceptType        string  `gorm:"type:varchar(50);not null;check:concept_type IN ('fixed','variable')" json:"concept_type"`
	IsTaxable          bool    `gorm:"default:true" json:"is_taxable"`
	IsIMSSBase         bool    `gorm:"default:true" json:"is_imss_base"`
	IsIntegratedSalary bool    `gorm:"default:false" json:"is_integrated_salary"` // Whether it's part of Integrated Daily Salary calculation
	SATCode            string  `gorm:"type:varchar(20)" json:"sat_code,omitempty"` // SAT code for CFDI Nomina
	Description        string  `gorm:"type:text" json:"description,omitempty"`
}

// TableName specifies the table name
func (PayrollConcept) TableName() string {
	return "payroll_concepts"
}

// BeforeCreate hook to validate PayrollConcept data before creation.
func (pc *PayrollConcept) BeforeCreate(tx *gorm.DB) (err error) {
	if pc.Name == "" {
		return ErrNameRequired
	}
	if pc.Category == "" {
		return ErrCategoryRequired
	}
	if pc.ConceptType == "" {
		return ErrConceptTypeRequired
	}
	return
}
