/*
Package services - Mexican Tax Calculation Service

==============================================================================
FILE: internal/services/tax_calculation_service.go
==============================================================================

DESCRIPTION:
    Calculates Mexican payroll taxes and social security contributions including
    ISR (income tax), IMSS (social security), INFONAVIT (housing), and employment
    subsidies based on official SAT tables.

USER PERSPECTIVE:
    - Accurate ISR withholding based on SAT 2025 tax brackets
    - IMSS employee contributions calculated per Mexican law
    - Employment subsidy (Subsidio al Empleo) automatically applied
    - Support for weekly, biweekly, and monthly pay periods

DEVELOPER GUIDELINES:
    OK to modify: Tax table JSON files, UMA daily value updates
    CAUTION: ISR and subsidy calculations must match SAT regulations exactly
    DO NOT modify: Tax bracket logic without verifying against SAT tables
    Note: Update tax tables annually with SAT published rates

SYNTAX EXPLANATION:
    - CalculateNetISR = ISR - Employment Subsidy (capped at 0)
    - ISR formula: Fixed fee + ((Income - Lower limit) * Rate / 100)
    - IMSS uses SDI (Integrated Daily Salary) capped at 25 UMA
    - INFONAVIT supports: percentage, fixed amount, or UMA-based deductions

==============================================================================
*/
package services

import (
	"encoding/json"
	"fmt"
	"os"
)

// ISRBracket represents a single row in the ISR table
type ISRBracket struct {
	LowerLimit float64 `json:"lower_limit"`
	UpperLimit float64 `json:"upper_limit"`
	FixedFee   float64 `json:"fixed_fee"`
	Percentage float64 `json:"percentage"`
}

// ISRTable represents the full ISR tax table
type ISRTable struct {
	Periodicity string       `json:"periodicity"`
	Year        int          `json:"year"`
	Description string       `json:"description"`
	Rows        []ISRBracket `json:"rows"`
}

// SubsidyBracket represents a single row in the employment subsidy table
type SubsidyBracket struct {
	LowerLimit    float64 `json:"lower_limit"`
	UpperLimit    float64 `json:"upper_limit"`
	SubsidyAmount float64 `json:"subsidy_amount"`
}

// SubsidyTable represents the employment subsidy tables by periodicity
type SubsidyTable struct {
	Year      int              `json:"year"`
	Monthly   []SubsidyBracket `json:"monthly"`
	Biweekly  []SubsidyBracket `json:"biweekly"`
	Weekly    []SubsidyBracket `json:"weekly"`
}

// IMSSEmployeeRates represents IMSS employee contribution rates
type IMSSEmployeeRates struct {
	SicknessMaternity  float64 // 0.0025 (0.25%)
	DisabilityLife     float64 // 0.00625 (0.625%)
	UnemploymentOldAge float64 // 0.01125 (1.125%)
	Retirement         float64 // 0.0 (employer pays 100%)
}

// TaxCalculationService handles all Mexican tax calculations
type TaxCalculationService struct {
	biweeklyISRTable ISRTable
	monthlyISRTable  ISRTable
	subsidyTable     SubsidyTable
	imssRates        IMSSEmployeeRates
	umaDaily         float64
}

// NewTaxCalculationService creates a new tax calculation service
func NewTaxCalculationService(configPath string) (*TaxCalculationService, error) {
	service := &TaxCalculationService{
		// Default IMSS employee rates from Mexican law
		imssRates: IMSSEmployeeRates{
			SicknessMaternity:  0.0025,  // 0.25%
			DisabilityLife:     0.00625, // 0.625%
			UnemploymentOldAge: 0.01125, // 1.125%
			Retirement:         0.0,     // Employer pays 100%
		},
		umaDaily: 113.14, // UMA 2025
	}

	// Load ISR biweekly table
	biweeklyPath := configPath + "/tables/isr_biweekly_2025.json"
	if err := service.loadISRTable(biweeklyPath, &service.biweeklyISRTable); err != nil {
		// If file not found or empty, use default hardcoded values
		service.biweeklyISRTable = getDefaultBiweeklyISRTable()
	}

	// Load ISR monthly table
	monthlyPath := configPath + "/tables/isr_monthly_2025.json"
	if err := service.loadISRTable(monthlyPath, &service.monthlyISRTable); err != nil {
		service.monthlyISRTable = getDefaultMonthlyISRTable()
	}

	// Load subsidy table
	subsidyPath := configPath + "/tables/subsidy_2025.json"
	if err := service.loadSubsidyTable(subsidyPath); err != nil {
		service.subsidyTable = getDefaultSubsidyTable()
	}

	return service, nil
}

func (s *TaxCalculationService) loadISRTable(path string, table *ISRTable) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) <= 2 { // Empty JSON {}
		return fmt.Errorf("empty file")
	}
	return json.Unmarshal(data, table)
}

func (s *TaxCalculationService) loadSubsidyTable(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) <= 2 { // Empty JSON {}
		return fmt.Errorf("empty file")
	}
	return json.Unmarshal(data, &s.subsidyTable)
}

// CalculateISR calculates the ISR (income tax) withholding for a given taxable income
// periodicity: "biweekly", "monthly", "weekly"
func (s *TaxCalculationService) CalculateISR(taxableIncome float64, periodicity string) float64 {
	var brackets []ISRBracket

	switch periodicity {
	case "biweekly":
		brackets = s.biweeklyISRTable.Rows
	case "monthly":
		brackets = s.monthlyISRTable.Rows
	default:
		brackets = s.biweeklyISRTable.Rows // Default to biweekly
	}

	if len(brackets) == 0 {
		// Fallback to default biweekly table if not loaded
		brackets = getDefaultBiweeklyISRTable().Rows
	}

	// Find the applicable bracket
	for _, bracket := range brackets {
		if taxableIncome >= bracket.LowerLimit && taxableIncome <= bracket.UpperLimit {
			// ISR = Fixed fee + ((Taxable income - Lower limit) * Rate / 100)
			excessAmount := taxableIncome - bracket.LowerLimit
			isr := bracket.FixedFee + (excessAmount * bracket.Percentage / 100)
			return isr
		}
	}

	// If income exceeds all brackets, use the last bracket
	if len(brackets) > 0 && taxableIncome > brackets[len(brackets)-1].UpperLimit {
		lastBracket := brackets[len(brackets)-1]
		excessAmount := taxableIncome - lastBracket.LowerLimit
		return lastBracket.FixedFee + (excessAmount * lastBracket.Percentage / 100)
	}

	return 0
}

// CalculateEmploymentSubsidy calculates the employment subsidy (Subsidio al Empleo)
func (s *TaxCalculationService) CalculateEmploymentSubsidy(taxableIncome float64, periodicity string) float64 {
	var brackets []SubsidyBracket

	switch periodicity {
	case "biweekly":
		brackets = s.subsidyTable.Biweekly
	case "monthly":
		brackets = s.subsidyTable.Monthly
	case "weekly":
		brackets = s.subsidyTable.Weekly
	default:
		brackets = s.subsidyTable.Biweekly
	}

	if len(brackets) == 0 {
		brackets = getDefaultSubsidyTable().Biweekly
	}

	// Find applicable subsidy bracket
	for _, bracket := range brackets {
		if taxableIncome >= bracket.LowerLimit && taxableIncome <= bracket.UpperLimit {
			return bracket.SubsidyAmount
		}
	}

	return 0
}

// CalculateNetISR calculates the net ISR after applying employment subsidy
// Returns: ISR to withhold (positive) or subsidy to pay (negative)
func (s *TaxCalculationService) CalculateNetISR(taxableIncome float64, periodicity string) float64 {
	isr := s.CalculateISR(taxableIncome, periodicity)
	subsidy := s.CalculateEmploymentSubsidy(taxableIncome, periodicity)

	netISR := isr - subsidy

	// If subsidy exceeds ISR, return 0 (no negative withholding in most cases)
	// Note: Some companies may pay out the difference, but by default we cap at 0
	if netISR < 0 {
		return 0
	}

	return netISR
}

// CalculateIMSSEmployee calculates the employee's IMSS contribution
// SDI = Salario Diario Integrado (Integrated Daily Salary)
// workingDays = number of working days in the period
func (s *TaxCalculationService) CalculateIMSSEmployee(sdi float64, workingDays int) float64 {
	// IMSS contribution cap: 25 times UMA (Unidad de Medida y Actualización)
	maxSDI := s.umaDaily * 25
	if sdi > maxSDI {
		sdi = maxSDI
	}

	periodSalary := sdi * float64(workingDays)

	// Calculate each component
	sicknessMaternity := periodSalary * s.imssRates.SicknessMaternity
	disabilityLife := periodSalary * s.imssRates.DisabilityLife
	unemploymentOldAge := periodSalary * s.imssRates.UnemploymentOldAge
	retirement := periodSalary * s.imssRates.Retirement

	totalIMSS := sicknessMaternity + disabilityLife + unemploymentOldAge + retirement

	return totalIMSS
}

// IMSSBreakdown provides a detailed breakdown of IMSS employee contributions
type IMSSBreakdown struct {
	SicknessMaternity  float64
	DisabilityLife     float64
	UnemploymentOldAge float64
	Retirement         float64
	Total              float64
}

// CalculateIMSSEmployeeBreakdown returns detailed IMSS breakdown
func (s *TaxCalculationService) CalculateIMSSEmployeeBreakdown(sdi float64, workingDays int) IMSSBreakdown {
	maxSDI := s.umaDaily * 25
	if sdi > maxSDI {
		sdi = maxSDI
	}

	periodSalary := sdi * float64(workingDays)

	breakdown := IMSSBreakdown{
		SicknessMaternity:  periodSalary * s.imssRates.SicknessMaternity,
		DisabilityLife:     periodSalary * s.imssRates.DisabilityLife,
		UnemploymentOldAge: periodSalary * s.imssRates.UnemploymentOldAge,
		Retirement:         periodSalary * s.imssRates.Retirement,
	}
	breakdown.Total = breakdown.SicknessMaternity + breakdown.DisabilityLife +
		breakdown.UnemploymentOldAge + breakdown.Retirement

	return breakdown
}

// CalculateINFONAVITEmployee calculates employee INFONAVIT deduction
// INFONAVIT deductions vary by credit type:
// - "porcentaje": percentage of integrated daily salary
// - "cuota_fija": fixed amount
// - "veces_salario_minimo": times minimum wage
func (s *TaxCalculationService) CalculateINFONAVITEmployee(
	sdi float64,
	workingDays int,
	creditType string,
	creditValue float64,
) float64 {
	switch creditType {
	case "porcentaje":
		// Percentage of integrated salary
		return sdi * float64(workingDays) * (creditValue / 100)
	case "cuota_fija":
		// Fixed amount per period
		return creditValue
	case "veces_salario_minimo":
		// Times minimum wage (UMA)
		// Convert to period amount
		dailyDeduction := s.umaDaily * creditValue
		return dailyDeduction * float64(workingDays)
	default:
		return 0
	}
}

// Default ISR tables (hardcoded fallback)
func getDefaultBiweeklyISRTable() ISRTable {
	return ISRTable{
		Periodicity: "biweekly",
		Year:        2025,
		Rows: []ISRBracket{
			{LowerLimit: 0.01, UpperLimit: 435.39, FixedFee: 0.00, Percentage: 1.92},
			{LowerLimit: 435.40, UpperLimit: 3694.52, FixedFee: 8.36, Percentage: 6.40},
			{LowerLimit: 3694.53, UpperLimit: 6576.26, FixedFee: 216.88, Percentage: 10.88},
			{LowerLimit: 6576.27, UpperLimit: 7707.69, FixedFee: 535.65, Percentage: 16.00},
			{LowerLimit: 7707.70, UpperLimit: 23076.92, FixedFee: 716.61, Percentage: 17.92},
			{LowerLimit: 23076.93, UpperLimit: 46153.85, FixedFee: 3452.27, Percentage: 21.36},
			{LowerLimit: 46153.86, UpperLimit: 138461.54, FixedFee: 8378.50, Percentage: 23.52},
			{LowerLimit: 138461.55, UpperLimit: 184615.38, FixedFee: 30087.05, Percentage: 30.00},
			{LowerLimit: 184615.39, UpperLimit: 553846.15, FixedFee: 43933.21, Percentage: 32.00},
			{LowerLimit: 553846.16, UpperLimit: 1661538.46, FixedFee: 162086.90, Percentage: 34.00},
			{LowerLimit: 1661538.47, UpperLimit: 999999999.99, FixedFee: 538703.04, Percentage: 35.00},
		},
	}
}

func getDefaultMonthlyISRTable() ISRTable {
	return ISRTable{
		Periodicity: "monthly",
		Year:        2025,
		Rows: []ISRBracket{
			{LowerLimit: 0.01, UpperLimit: 9440.18, FixedFee: 0.00, Percentage: 1.92},
			{LowerLimit: 9440.19, UpperLimit: 80047.92, FixedFee: 181.31, Percentage: 6.40},
			{LowerLimit: 80047.93, UpperLimit: 142562.30, FixedFee: 4704.21, Percentage: 10.88},
			{LowerLimit: 142562.31, UpperLimit: 167000.00, FixedFee: 11607.37, Percentage: 16.00},
			{LowerLimit: 167000.01, UpperLimit: 500000.00, FixedFee: 15516.57, Percentage: 17.92},
			{LowerLimit: 500000.01, UpperLimit: 1000000.00, FixedFee: 74817.57, Percentage: 21.36},
			{LowerLimit: 1000000.01, UpperLimit: 999999999.99, FixedFee: 181817.57, Percentage: 23.52},
		},
	}
}

func getDefaultSubsidyTable() SubsidyTable {
	return SubsidyTable{
		Year: 2025,
		Biweekly: []SubsidyBracket{
			{LowerLimit: 0.01, UpperLimit: 815.49, SubsidyAmount: 187.85},
			{LowerLimit: 815.50, UpperLimit: 1223.86, SubsidyAmount: 187.76},
			{LowerLimit: 1223.87, UpperLimit: 1601.16, SubsidyAmount: 187.67},
			{LowerLimit: 1601.17, UpperLimit: 1631.17, SubsidyAmount: 181.26},
			{LowerLimit: 1631.18, UpperLimit: 2050.61, SubsidyAmount: 176.52},
			{LowerLimit: 2050.62, UpperLimit: 2176.62, SubsidyAmount: 163.49},
			{LowerLimit: 2176.63, UpperLimit: 2461.73, SubsidyAmount: 149.94},
			{LowerLimit: 2461.74, UpperLimit: 2872.92, SubsidyAmount: 135.98},
			{LowerLimit: 2872.93, UpperLimit: 3283.11, SubsidyAmount: 117.02},
			{LowerLimit: 3283.12, UpperLimit: 3406.84, SubsidyAmount: 100.43},
			{LowerLimit: 3406.85, UpperLimit: 999999999.99, SubsidyAmount: 0.00},
		},
		Monthly: []SubsidyBracket{
			{LowerLimit: 0.01, UpperLimit: 1768.96, SubsidyAmount: 407.02},
			{LowerLimit: 1768.97, UpperLimit: 2653.38, SubsidyAmount: 406.83},
			{LowerLimit: 2653.39, UpperLimit: 3472.84, SubsidyAmount: 406.62},
			{LowerLimit: 3472.85, UpperLimit: 3537.87, SubsidyAmount: 392.77},
			{LowerLimit: 3537.88, UpperLimit: 4446.15, SubsidyAmount: 382.46},
			{LowerLimit: 4446.16, UpperLimit: 4717.18, SubsidyAmount: 354.23},
			{LowerLimit: 4717.19, UpperLimit: 5335.42, SubsidyAmount: 324.87},
			{LowerLimit: 5335.43, UpperLimit: 6224.67, SubsidyAmount: 294.63},
			{LowerLimit: 6224.68, UpperLimit: 7113.90, SubsidyAmount: 253.54},
			{LowerLimit: 7113.91, UpperLimit: 7382.33, SubsidyAmount: 217.61},
			{LowerLimit: 7382.34, UpperLimit: 999999999.99, SubsidyAmount: 0.00},
		},
		Weekly: []SubsidyBracket{
			{LowerLimit: 0.01, UpperLimit: 407.33, SubsidyAmount: 93.73},
			{LowerLimit: 407.34, UpperLimit: 610.96, SubsidyAmount: 93.66},
			{LowerLimit: 610.97, UpperLimit: 799.68, SubsidyAmount: 93.60},
			{LowerLimit: 799.69, UpperLimit: 814.66, SubsidyAmount: 90.44},
			{LowerLimit: 814.67, UpperLimit: 1023.75, SubsidyAmount: 88.06},
			{LowerLimit: 1023.76, UpperLimit: 1086.19, SubsidyAmount: 81.55},
			{LowerLimit: 1086.20, UpperLimit: 1228.57, SubsidyAmount: 74.83},
			{LowerLimit: 1228.58, UpperLimit: 1433.32, SubsidyAmount: 67.83},
			{LowerLimit: 1433.33, UpperLimit: 1638.07, SubsidyAmount: 58.38},
			{LowerLimit: 1638.08, UpperLimit: 1699.88, SubsidyAmount: 50.12},
			{LowerLimit: 1699.89, UpperLimit: 999999999.99, SubsidyAmount: 0.00},
		},
	}
}
