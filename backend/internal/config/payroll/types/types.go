/*
Package types - Mexican Payroll Configuration Type Definitions

==============================================================================
FILE: internal/config/payroll/types/types.go
==============================================================================

DESCRIPTION:
    Defines all data structures for Mexican payroll configuration including
    official government values, tax tables, social security rates, and
    labor law parameters. These types are populated from JSON config files.

USER PERSPECTIVE:
    - Defines the structure of payroll rules loaded from config files
    - Contains official Mexican government values (UMA, minimum wages, IMSS)
    - Supports multi-zone wage configurations (general, northern border)

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields when regulations change
    ⚠️  CAUTION: Changing field names (breaks JSON loading)
    ❌  DO NOT modify: Remove fields without migration plan
    📝  Keep JSON tags matching config file structure

SYNTAX EXPLANATION:
    - json tags: Map to payroll config JSON files
    - Nested structs: Mirror the hierarchical config structure
    - Brackets: Define ISR/subsidy tables for tax calculations

KEY MEXICAN PAYROLL CONCEPTS:
    - UMA: Unidad de Medida y Actualización (unit for calculating benefits)
    - SMG: Salario Mínimo General (general minimum wage)
    - IMSS: Instituto Mexicano del Seguro Social (social security)
    - ISR: Impuesto Sobre la Renta (income tax)
    - Aguinaldo: Christmas bonus (minimum 15 days salary)
    - Prima Vacacional: Vacation premium (25% of vacation pay)

==============================================================================
*/
package types

// OfficialValues holds official values like UMA, minimum wages, and IMSS limits.
type OfficialValues struct {
	FiscalYear   int           `json:"fiscal_year"`
	UMA          UMAValues     `json:"uma"`
	MinimumWages MinimumWages  `json:"minimum_wages"`
	Limits       OfficialLimits `json:"limits"`
}

// UMAValues holds UMA (Unidad de Medida y Actualización) values.
type UMAValues struct {
	DailyValue   float64 `json:"daily_value"`
	MonthlyValue float64 `json:"monthly_value"`
	AnnualValue  float64 `json:"annual_value"`
}

// MinimumWages holds general and professional minimum wage values.
type MinimumWages struct {
	DefaultZone          string                `json:"default_zone"` // e.g., "general", "northern_border_free_zone"
	General              MinimumWageZone       `json:"general"`
	NorthernBorderFreeZone MinimumWageZone       `json:"northern_border_free_zone"`
	Historical           MinimumWageHistorical `json:"historical"`
}

// MinimumWageZone defines minimum wage values for a specific zone.
type MinimumWageZone struct {
	DailyValue        float64 `json:"daily_value"`
	ProfessionalDaily map[string]float64 `json:"professional_daily,omitempty"`
}

// MinimumWageHistorical contains historical or phased-out zone data.
type MinimumWageHistorical struct {
	ZoneB MinimumWageZoneHistorical `json:"zone_b"`
	ZoneC MinimumWageZoneHistorical `json:"json_c"`
}

// MinimumWageZoneHistorical for zones that might no longer be active.
type MinimumWageZoneHistorical struct {
	Active            bool    `json:"active"`
	DailyValue        float64 `json:"daily_value"`
	ProfessionalDaily map[string]float64 `json:"professional_daily,omitempty"`
}

// OfficialLimits holds various official limits, e.g., IMSS caps.
type OfficialLimits struct {
	IMSSCap IMSSCapLimit `json:"imss_cap"`
}

// IMSSCapLimit defines the IMSS contribution cap.
type IMSSCapLimit struct {
	Multiplier float64 `json:"multiplier"` // e.g., 25 times UMA
}


// RegionalConfig holds regional-specific configuration.
type RegionalConfig struct {
	State           StateConfig          `json:"state"`
	StatePayrollTax StatePayrollTaxConfig `json:"state_payroll_tax"`
	LocalHolidays   []LocalHoliday       `json:"local_holidays"`
}

// StateConfig defines basic state information.
type StateConfig struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// StatePayrollTaxConfig defines state payroll tax rules.
type StatePayrollTaxConfig struct {
	Enabled         bool    `json:"enabled"`
	Rate            float64 `json:"rate"`
	CalculationBase []string `json:"calculation_base"` // e.g., "gross_salary", "taxable_salary"
}

// LocalHoliday defines a local holiday.
type LocalHoliday struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	AppliesTo   string `json:"applies_to"` // e.g., "all", "state", "municipality"
}


// ContributionRates holds all social security and other contribution rates.
type ContributionRates struct {
	IMSS      IMSSRatesConfig    `json:"imss"`
	Infonavit InfonavitRatesConfig `json:"infonavit"`
}

// IMSSRatesConfig contains IMSS rates for employer and employee.
type IMSSRatesConfig struct {
	Employer IMSSPartyRatesConfig `json:"employer"`
	Employee IMSSPartyRatesConfig `json:"employee"`
	MaximumBases IMSSMaximumBases `json:"maximum_bases"`
}

// IMSSPartyRatesConfig defines IMSS rates for either employer or employee.
type IMSSPartyRatesConfig struct {
	// Sickness and Maternity Insurance
	DiseaseMaternityInsurance float64 `json:"disease_maternity_insurance"`
	// Disability and Life Insurance
	DisabilityLife float64 `json:"disability_life"`
	// Retirement, Severance, and Old Age (RCV)
	Retirement       float64 `json:"retirement"`
	SeveranceOldAge  SeveranceOldAgeRates `json:"severance_old_age"`
	// Childcare and Social Benefits
	ChildcareSocialBenefits float64 `json:"childcare_social_benefits"`
	// Work Risk Insurance (varies by activity class)
	WorkRisk WorkRiskRates `json:"work_risk"`
	// Housing Fund (INFONAVIT)
	Housing float64 `json:"housing"`
}

// IMSSMaximumBases defines the maximum bases for IMSS calculations.
type IMSSMaximumBases struct {
	DiseaseMaternityInsurance float64 `json:"disease_maternity_insurance"`
	DisabilityLifeInsurance   float64 `json:"disability_life_insurance"`
	Retirement                float64 `json:"retirement"`
}

// SeveranceOldAgeRates holds specific rates for Severance and Old Age.
type SeveranceOldAgeRates struct {
	Base    float64 `json:"base"`
	Credits float64 `json:"credits"`
}

// WorkRiskRates holds specific rates for Work Risk Insurance per class.
type WorkRiskRates struct {
	ClassI   float64 `json:"class_i"`
	ClassII  float64 `json:"class_ii"`
	ClassIII float64 `json:"class_iii"`
	ClassIV  float64 `json:"class_iv"`
	ClassV   float64 `json:"class_v"`
}

// InfonavitRatesConfig contains Infonavit rates for employer and employee.
type InfonavitRatesConfig struct {
	EmployerContributionRate float64 `json:"employer_contribution_rate"`
	EmployeeContributionRate float64 `json:"employee_contribution_rate"`
	MaximumBase              float64 `json:"maximum_base"` // Max times UMA
}


// LaborConcepts defines rules and parameters for various labor concepts.
type LaborConcepts struct {
	ChristmasBonus AguinaldoRulesConfig `json:"christmas_bonus"`
	Vacations      VacationRulesConfig      `json:"vacations"`
	WorkSchedule   WorkSchedule   `json:"work_schedule"`
	Overtime       Overtime       `json:"overtime"`
	SundayPremium  SundayPremium  `json:"sunday_premium"`
	SavingsFund    SavingsFund    `json:"savings_fund"`
	FoodVouchers   FoodVouchers   `json:"food_vouchers"`
}

// AguinaldoRulesConfig defines rules for Aguinaldo (Christmas bonus).
type AguinaldoRulesConfig struct {
	MinimumDays      int     `json:"minimum_days"`
	MinimumAmount int `json:"minimum_amount"` // In days of salary
	TaxExemptUMALimit float64 `json:"tax_exempt_uma_limit"` // In UMAs
}

// VacationRulesConfig defines rules for vacations.
type VacationRulesConfig struct {
	VacationPremiumRate float64 `json:"vacation_premium_rate"`
	VacationDaysTable   []VacationDaysBracket `json:"vacation_days_table"`
	FirstYearDays       int     `json:"first_year_days"`
	VacationBonusPercentage float64 `json:"vacation_bonus_percentage"`
	TaxExemptUMALimit   float64 `json:"tax_exempt_uma_limit"` // In UMAs
}

// VacationDaysBracket defines vacation days per years of service.
type VacationDaysBracket struct {
	YearsOfService int `json:"years_of_service"`
	Days           int `json:"days"`
}

// WorkSchedule defines standard work schedule parameters.
type WorkSchedule struct {
	DailyHours    float64 `json:"daily_hours"`
	WeeklyDays    float64 `json:"weekly_days"`
	BiweeklyDays  float64 `json:"biweekly_days"`
	MonthlyDays   float64 `json:"monthly_days"`
	AnnualDays    float64 `json:"annual_days"`
}

// Overtime defines rules for overtime payment.
type Overtime struct {
	DoubleTimePercentage float64 `json:"double_time_percentage"`
	TripleTimePercentage float64 `json:"triple_time_percentage"`
	TaxExemptUMALimit    float64 `json:"tax_exempt_uma_limit"` // In UMAs
}

// SundayPremium defines rules for Sunday premium.
type SundayPremium struct {
	Percentage float64 `json:"percentage"`
	TaxExemptUMALimit float64 `json:"tax_exempt_uma_limit"` // In UMAs
}

// SavingsFund defines rules for savings fund.
type SavingsFund struct {
	EmployerContributionPercentage float64 `json:"employer_contribution_percentage"`
	EmployeeContributionPercentage float64 `json:"employee_contribution_percentage"`
	MaxContributionUMALimit      float64 `json:"max_contribution_uma_limit"` // In UMAs
}

// FoodVouchers defines rules for food vouchers.
type FoodVouchers struct {
	MaxExemptUMALimit float64 `json:"max_exempt_uma_limit"` // In UMAs
}


// CalculationTables holds various tables used in payroll calculations.
type CalculationTables struct {
	ISRTables    ISRTablesConfig    `json:"isr_tables"`
	SubsidyTables SubsidyTablesConfig `json:"subsidy_tables"`
}

// ISRTablesConfig contains monthly, biweekly, and weekly ISR tables.
type ISRTablesConfig struct {
	Monthly2024      []ISRBracket    `json:"monthly_2024"`
	Biweekly2024     []ISRBracket    `json:"biweekly_2024"`
	Weekly2024       []ISRBracket    `json:"weekly_2024"`
	EmploymentSubsidy []SubsidyBracket `json:"employment_subsidy"`
}

// ISRBracket defines a single bracket for ISR calculation.
type ISRBracket struct {
	LowerLimit     float64 `json:"lower_limit"`
	UpperLimit     float64 `json:"upper_limit"`
	FixedFee       float64 `json:"fixed_fee"`
	RatePercentage float64 `json:"rate_percentage"`
}

// SubsidyTablesConfig contains monthly, biweekly, and weekly subsidy tables.
type SubsidyTablesConfig struct {
	Monthly  []SubsidyBracket `json:"monthly"`
	Biweekly []SubsidyBracket `json:"biweekly"`
	Weekly   []SubsidyBracket `json:"weekly"`
}

// SubsidyBracket defines a single bracket for employment subsidy.
type SubsidyBracket struct {
	LowerLimit    float64 `json:"lower_limit"`
	UpperLimit    float64 `json:"upper_limit"`
	SubsidyAmount float64 `json:"subsidy_amount"`
}
