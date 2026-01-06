/*
Package payroll - Mexican Payroll Configuration and Calculations

==============================================================================
FILE: internal/config/payroll/config.go
==============================================================================

DESCRIPTION:
    Main payroll configuration structure that aggregates all Mexican labor
    law parameters, tax tables, and social security rates. Provides helper
    methods for common payroll calculations.

USER PERSPECTIVE:
    - Central configuration for Mexican payroll calculations
    - Provides minimum wage lookups by zone
    - Calculates IMSS employer/employee contributions
    - Determines vacation days based on years of service

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new calculation methods
    ⚠️  CAUTION: Modifying existing calculations (affects payroll accuracy)
    ❌  DO NOT modify: Rate formulas without legal verification
    📝  All rates should come from official SAT/IMSS publications

SYNTAX EXPLANATION:
    - PayrollConfig struct: Aggregates all config types
    - GetVacationDaysForYears(): Implements LFT vacation table
    - CalculateIMSSEmployerContribution(): Sum of all IMSS employer rates

MEXICAN LAW REFERENCES:
    - LFT (Ley Federal del Trabajo): Vacation days, overtime, bonuses
    - LSS (Ley del Seguro Social): IMSS contribution rates
    - LISR (Ley del ISR): Income tax tables

==============================================================================
*/
package payroll

import (
	"backend/internal/config/payroll/types"
)

// PayrollConfig is the main payroll configuration structure
type PayrollConfig struct {
	OfficialValues    types.OfficialValues    `json:"official_values"`
	Regional          types.RegionalConfig    `json:"regional"`
	ContributionRates types.ContributionRates `json:"contribution_rates"`
	LaborConcepts     types.LaborConcepts     `json:"labor_concepts"`
	CalculationTables types.CalculationTables `json:"calculation_tables"`
	MexicanTaxConfig  MexicanTaxConfig        `yaml:"mexican_tax_config"`
}

// NewPayrollConfig creates a new empty payroll configuration
func NewPayrollConfig() *PayrollConfig {
    return &PayrollConfig{}
}

// GetDefaultSMG returns the default minimum wage for San Luis Potosí
func (pc *PayrollConfig) GetDefaultSMG() (float64, error) {
    if pc.OfficialValues.MinimumWages.DefaultZone == "general" {
        return pc.OfficialValues.MinimumWages.General.DailyValue, nil
    } else if pc.OfficialValues.MinimumWages.DefaultZone == "northern_border_free_zone" {
        return pc.OfficialValues.MinimumWages.NorthernBorderFreeZone.DailyValue, nil
    }
    return pc.OfficialValues.MinimumWages.General.DailyValue, nil
}

// GetSMGForZone returns the minimum wage for a specific zone
func (pc *PayrollConfig) GetSMGForZone(zone string) (float64, error) {
    switch zone {
    case "general":
        return pc.OfficialValues.MinimumWages.General.DailyValue, nil
    case "northern_border_free_zone":
        return pc.OfficialValues.MinimumWages.NorthernBorderFreeZone.DailyValue, nil
    case "zone_b":
        if pc.OfficialValues.MinimumWages.Historical.ZoneB.Active {
            return pc.OfficialValues.MinimumWages.Historical.ZoneB.DailyValue, nil
        }
    case "zone_c":
        if pc.OfficialValues.MinimumWages.Historical.ZoneC.Active {
            return pc.OfficialValues.MinimumWages.Historical.ZoneC.DailyValue, nil
        }
    }
    return pc.GetDefaultSMG()
}

// CalculateIMSSEmployerContribution calculates employer IMSS contribution
func (pc *PayrollConfig) CalculateIMSSEmployerContribution(baseSalary float64) float64 {
    // Sickness and maternity (cash benefits)
    sicknessMaternity := baseSalary * pc.ContributionRates.IMSS.Employer.DiseaseMaternityInsurance
    
    // Disability and life
    disabilityLife := baseSalary * pc.ContributionRates.IMSS.Employer.DisabilityLife
    
    // Retirement
    retirement := baseSalary * pc.ContributionRates.IMSS.Employer.Retirement
    
    // Unemployment and old age (base)
    unemploymentOldAge := baseSalary * pc.ContributionRates.IMSS.Employer.SeveranceOldAge.Base
    
    // Childcare and Social Benefits
    daycare := baseSalary * pc.ContributionRates.IMSS.Employer.ChildcareSocialBenefits
    
    // Work risk (using default class I for office work)
    workRisk := baseSalary * pc.ContributionRates.IMSS.Employer.WorkRisk.ClassI
    
    total := sicknessMaternity + disabilityLife + retirement + unemploymentOldAge + daycare + workRisk
    return total
}

// CalculateStatePayrollTax calculates state payroll tax for San Luis Potosí
func (pc *PayrollConfig) CalculateStatePayrollTax(taxableBase float64) float64 {
    if !pc.Regional.StatePayrollTax.Enabled {
        return 0.0
    }
    return taxableBase * pc.Regional.StatePayrollTax.Rate
}

// GetIMSSEmployerTotalRate returns the total employer IMSS rate
func (pc *PayrollConfig) GetIMSSEmployerTotalRate() float64 {
    rates := pc.ContributionRates.IMSS.Employer
    total := rates.DiseaseMaternityInsurance +
             rates.DisabilityLife +
             rates.Retirement +
             rates.SeveranceOldAge.Base +
             rates.ChildcareSocialBenefits +
             rates.WorkRisk.ClassI
    return total
}

// GetIMSSEmployeeTotalRate returns the total employee IMSS rate
func (pc *PayrollConfig) GetIMSSEmployeeTotalRate() float64 {
    rates := pc.ContributionRates.IMSS.Employee
    return rates.DiseaseMaternityInsurance + rates.DisabilityLife + rates.Retirement
}

// GetVacationDaysForYears returns vacation days based on years of service
func (pc *PayrollConfig) GetVacationDaysForYears(yearsOfService int) int {
    if yearsOfService == 1 {
        return pc.LaborConcepts.Vacations.FirstYearDays
    }
    
    // Default calculation if table is not loaded
    if yearsOfService >= 2 && yearsOfService <= 5 {
        return 14
    } else if yearsOfService >= 6 && yearsOfService <= 10 {
        return 16
    } else if yearsOfService >= 11 && yearsOfService <= 15 {
        return 18
    } else if yearsOfService >= 16 && yearsOfService <= 20 {
        return 20
    } else if yearsOfService >= 21 && yearsOfService <= 25 {
        return 22
    } else if yearsOfService >= 26 && yearsOfService <= 30 {
        return 24
    } else if yearsOfService >= 31 && yearsOfService <= 35 {
        return 26
    } else if yearsOfService >= 36 && yearsOfService <= 40 {
        return 28
    } else if yearsOfService >= 41 && yearsOfService <= 45 {
        return 30
    } else if yearsOfService >= 46 && yearsOfService <= 50 {
        return 32
    } else if yearsOfService >= 51 {
        return 34
    }
    
    return pc.LaborConcepts.Vacations.FirstYearDays
}
