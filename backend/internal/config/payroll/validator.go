/*
Package payroll - Payroll Configuration Validator

==============================================================================
FILE: internal/config/payroll/validator.go
==============================================================================

DESCRIPTION:
    Validates payroll configuration loaded from JSON files to ensure
    all required values are present and within acceptable ranges.
    Called automatically when loading configuration.

USER PERSPECTIVE:
    - Ensures payroll config is valid before calculations
    - Catches configuration errors at startup
    - Prevents invalid payroll calculations

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new validation rules, improve error messages
    ⚠️  CAUTION: Relaxing validation rules (may allow invalid configs)
    ❌  DO NOT modify: Remove existing validations without review
    📝  Add validations for any new configuration fields

SYNTAX EXPLANATION:
    - Validator struct: Stateless validator pattern
    - Validate(): Main entry point, calls all sub-validators
    - Each validateXxx(): Validates a specific config section
    - Returns aggregated errors with all validation failures

VALIDATED RULES:
    - UMA daily value must be positive
    - Minimum wages must be positive
    - IMSS cap multiplier must be positive
    - Fiscal year within valid range (2020-2030)
    - State payroll tax rate between 0 and 1
    - Vacation days and bonus percentages within legal limits

==============================================================================
*/
package payroll

import (
    "fmt"
    "strings"

    	"backend/internal/config/payroll/types")

// Validator validates payroll configuration
type Validator struct{}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
    return &Validator{}
}

// Validate validates the complete payroll configuration
func (v *Validator) Validate(config *PayrollConfig) error {
    var validationErrors []string
    
    // Validate official values
    if err := v.validateOfficialValues(config.OfficialValues); err != nil {
        validationErrors = append(validationErrors, err.Error())
    }
    
    // Validate regional configuration
    if err := v.validateRegionalConfig(config.Regional); err != nil {
        validationErrors = append(validationErrors, err.Error())
    }
    
    // Validate contribution rates
    if err := v.validateContributionRates(config.ContributionRates); err != nil {
        validationErrors = append(validationErrors, err.Error())
    }
    
    // Validate labor concepts
    if err := v.validateLaborConcepts(config.LaborConcepts); err != nil {
        validationErrors = append(validationErrors, err.Error())
    }
    
    if len(validationErrors) > 0 {
        return fmt.Errorf("configuration validation failed:\n  - %s", 
            strings.Join(validationErrors, "\n  - "))
    }
    
    return nil
}

// validateOfficialValues validates official values
func (v *Validator) validateOfficialValues(ov types.OfficialValues) error {
    if ov.UMA.DailyValue <= 0 {
        return fmt.Errorf("UMA daily value must be positive")
    }
    
    if ov.MinimumWages.General.DailyValue <= 0 {
        return fmt.Errorf("general minimum wage must be positive")
    }
    
    if ov.Limits.IMSSCap.Multiplier <= 0 {
        return fmt.Errorf("IMSS cap multiplier must be positive")
    }
    
    if ov.FiscalYear < 2020 || ov.FiscalYear > 2030 {
        return fmt.Errorf("fiscal year %d is outside valid range", ov.FiscalYear)
    }
    
    return nil
}

// validateRegionalConfig validates regional configuration
func (v *Validator) validateRegionalConfig(rc types.RegionalConfig) error {
    if rc.State.Name == "" {
        return fmt.Errorf("state name cannot be empty")
    }
    
    if rc.StatePayrollTax.Enabled {
        if rc.StatePayrollTax.Rate < 0 || rc.StatePayrollTax.Rate > 1 {
            return fmt.Errorf("state payroll tax rate must be between 0 and 1")
        }
        
        if len(rc.StatePayrollTax.CalculationBase) == 0 {
            return fmt.Errorf("state payroll tax calculation base cannot be empty")
        }
    }
    
    return nil
}

// validateContributionRates validates contribution rates
func (v *Validator) validateContributionRates(cr types.ContributionRates) error {
    // Validate employer rates
    if cr.IMSS.Employer.DiseaseMaternityInsurance < 0 {
        return fmt.Errorf("employer sickness/maternity cash benefits rate cannot be negative")
    }
    
    if cr.IMSS.Employer.DisabilityLife < 0 {
        return fmt.Errorf("employer disability/life rate cannot be negative")
    }
    
    if cr.IMSS.Employer.Retirement < 0 {
        return fmt.Errorf("employer retirement rate cannot be negative")
    }
    
    // Validate employee rates
    if cr.IMSS.Employee.DiseaseMaternityInsurance < 0 {
        return fmt.Errorf("employee sickness/maternity rate cannot be negative")
    }
    
    if cr.IMSS.Employee.DisabilityLife < 0 {
        return fmt.Errorf("employee disability/life rate cannot be negative")
    }
    
    if cr.Infonavit.EmployerContributionRate < 0 {
        return fmt.Errorf("employer INFONAVIT rate cannot be negative")
    }
    
    return nil
}

// validateLaborConcepts validates labor concepts
func (v *Validator) validateLaborConcepts(lc types.LaborConcepts) error {
    if lc.ChristmasBonus.MinimumDays <= 0 {
        return fmt.Errorf("christmas bonus minimum days must be positive")
    }
    
    if lc.Vacations.FirstYearDays <= 0 {
        return fmt.Errorf("vacation first year days must be positive")
    }
    
    if lc.Vacations.VacationBonusPercentage < 0 || lc.Vacations.VacationBonusPercentage > 1 {
        return fmt.Errorf("vacation bonus percentage must be between 0 and 1")
    }
    
    if lc.WorkSchedule.DailyHours <= 0 {
        return fmt.Errorf("daily work hours must be positive")
    }
    
    if lc.WorkSchedule.WeeklyDays <= 0 || lc.WorkSchedule.WeeklyDays > 7 {
        return fmt.Errorf("weekly work days must be between 1 and 7")
    }
    
    if lc.Overtime.DoubleTimePercentage < 0 {
        return fmt.Errorf("overtime double time percentage cannot be negative")
    }
    
    if lc.Overtime.TripleTimePercentage < 0 {
        return fmt.Errorf("overtime triple time percentage cannot be negative")
    }
    
    if lc.SundayPremium.Percentage < 0 {
        return fmt.Errorf("sunday premium percentage cannot be negative")
    }
    
    return nil
}
