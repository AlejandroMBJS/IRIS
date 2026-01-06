/*
Package payroll - Mexican Tax Configuration Structures

==============================================================================
FILE: internal/config/payroll/mexican_tax_config.go
==============================================================================

DESCRIPTION:
    Defines the MexicanTaxConfig structure that groups all tax-related
    configuration including ISR tables, IMSS rates, INFONAVIT, vacation
    rules, and aguinaldo (Christmas bonus) rules.

USER PERSPECTIVE:
    - Groups tax configuration for easier access
    - Used by payroll calculation services
    - Loaded from JSON configuration files

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new tax-related configuration groups
    ⚠️  CAUTION: Changing field types (breaks JSON unmarshaling)
    ❌  DO NOT modify: Field names without updating JSON configs
    📝  Keep structure aligned with SAT requirements

SYNTAX EXPLANATION:
    - Embeds types from the types package
    - Uses json tags for configuration file loading
    - Aggregates related tax configurations

==============================================================================
*/
package payroll

import (
	"backend/internal/config/payroll/types"
)

// MexicanTaxConfig represents the overall Mexican tax configuration.
type MexicanTaxConfig struct {
	ISRTables    types.ISRTablesConfig    `json:"isr_tables"`
	IMSSRates    types.IMSSRatesConfig    `json:"imss_rates"`
	InfonavitRates types.InfonavitRatesConfig `json:"infonavit_rates"`
	VacationRules  types.VacationRulesConfig  `json:"vacation_rules"`
	AguinaldoRules types.AguinaldoRulesConfig `json:"aguinaldo_rules"`
}