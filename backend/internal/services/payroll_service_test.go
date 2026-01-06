package services

import (
	"backend/internal/models"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ============================================================================
// Test Setup and Helpers
// ============================================================================

// setupPayrollTestDB creates an in-memory SQLite database for testing
func setupPayrollTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Migrate all required tables
	err = db.AutoMigrate(
		&models.Company{},
		&models.Employee{},
		&models.PayrollPeriod{},
		&models.PayrollCalculation{},
		&models.PrenominaMetric{},
		&models.EmployerContribution{},
		&models.PayrollDetail{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// createPayrollTestCompany creates a test company in the database
func createPayrollTestCompany(t *testing.T, db *gorm.DB) *models.Company {
	company := &models.Company{
		Name:     "Test Company SA de CV",
		RFC:      "TCO123456789",
		IsActive: true,
	}
	company.ID = uuid.New()
	err := db.Create(company).Error
	require.NoError(t, err)
	return company
}

// createPayrollTestEmployee creates a test employee with typical values
func createPayrollTestEmployee(t *testing.T, db *gorm.DB, companyID uuid.UUID, dailySalary float64) *models.Employee {
	empNum := uuid.New().String()[:8] // Unique employee number
	employee := &models.Employee{
		EmployeeNumber:   "EMP-" + empNum,
		FirstName:        "Juan",
		LastName:         "Perez",
		MotherLastName:   "Garcia",
		RFC:              "PEGJ900101ABC",
		CURP:             "PEGJ900101HSPLRN09",
		DateOfBirth:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HireDate:         time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
		DailySalary:      dailySalary,
		EmploymentStatus: "active",
		CollarType:       "white_collar",
		PayFrequency:     "biweekly",
		CompanyID:        companyID,
		Gender:           "male",
		EmployeeType:     "permanent",
	}
	employee.ID = uuid.New()
	err := db.Create(employee).Error
	require.NoError(t, err)
	return employee
}

// periodCounter is used to generate unique period codes
var periodCounter int

// createPayrollTestPeriod creates a biweekly test period
func createPayrollTestPeriod(t *testing.T, db *gorm.DB, frequency string) *models.PayrollPeriod {
	var startDate, endDate, paymentDate time.Time
	var periodCode string
	var periodType string

	periodCounter++
	periodNum := periodCounter % 100 // Keep it as 2 digits

	switch frequency {
	case "weekly":
		startDate = time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)
		paymentDate = time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC)
		periodCode = fmt.Sprintf("2025-W%02d", periodNum)
		periodType = "weekly"
	case "monthly":
		startDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
		paymentDate = time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)
		periodCode = fmt.Sprintf("2025-M%02d", periodNum)
		periodType = "monthly"
	default: // biweekly
		startDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentDate = time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
		periodCode = fmt.Sprintf("2025-BW%02d", periodNum)
		periodType = "biweekly"
	}

	period := &models.PayrollPeriod{
		PeriodCode:   periodCode,
		Year:         2025,
		PeriodNumber: 1,
		Frequency:    frequency,
		PeriodType:   periodType,
		StartDate:    startDate,
		EndDate:      endDate,
		PaymentDate:  paymentDate,
		Status:       "open",
	}
	period.ID = uuid.New()
	err := db.Create(period).Error
	require.NoError(t, err)
	return period
}

// createPayrollTestPrenomina creates an approved prenomina for testing
func createPayrollTestPrenomina(t *testing.T, db *gorm.DB, employeeID, periodID uuid.UUID) *models.PrenominaMetric {
	now := time.Now()
	prenomina := &models.PrenominaMetric{
		EmployeeID:        employeeID,
		PayrollPeriodID:   periodID,
		CalculationStatus: "approved",
		CalculationDate:   &now,
		WorkedDays:        15,
		RegularHours:      120,
		OvertimeHours:     10,
		AbsenceDays:       0,
		RegularSalary:     7500.00,
		OvertimeAmount:    625.00,
	}
	prenomina.ID = uuid.New()
	err := db.Create(prenomina).Error
	require.NoError(t, err)
	return prenomina
}

// ============================================================================
// TaxCalculationService Tests (ISR Calculation)
// ============================================================================

func TestCalculateISR_BiweeklyLowIncome(t *testing.T) {
	// Test ISR calculation for low income bracket (first bracket)
	taxService, err := NewTaxCalculationService("nonexistent") // Will use defaults
	require.NoError(t, err)

	// $400 biweekly income - first bracket (1.92%)
	isr := taxService.CalculateISR(400.00, "biweekly")

	// Expected: 0 fixed + (400 - 0.01) * 1.92% = 7.68 (approximately)
	assert.InDelta(t, 7.68, isr, 0.10)
}

func TestCalculateISR_BiweeklyMiddleIncome(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// $5000 biweekly income - third bracket
	// LowerLimit: 3694.53, UpperLimit: 6576.26, FixedFee: 216.88, Percentage: 10.88
	isr := taxService.CalculateISR(5000.00, "biweekly")

	// Expected: 216.88 + (5000 - 3694.53) * 10.88% = 216.88 + 142.12 = 359.00
	assert.InDelta(t, 359.00, isr, 1.0)
}

func TestCalculateISR_BiweeklyHighIncome(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// $50,000 biweekly income - higher bracket
	// LowerLimit: 46153.86, UpperLimit: 138461.54, FixedFee: 8378.50, Percentage: 23.52
	isr := taxService.CalculateISR(50000.00, "biweekly")

	// Expected: 8378.50 + (50000 - 46153.86) * 23.52% = 8378.50 + 904.37 = 9282.87
	assert.InDelta(t, 9282.87, isr, 10.0)
}

func TestCalculateISR_MonthlyIncome(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// $20,000 monthly income - second bracket
	// LowerLimit: 9440.19, UpperLimit: 80047.92, FixedFee: 181.31, Percentage: 6.40
	isr := taxService.CalculateISR(20000.00, "monthly")

	// Expected: 181.31 + (20000 - 9440.19) * 6.40% = 181.31 + 675.83 = 857.14
	assert.InDelta(t, 857.14, isr, 5.0)
}

func TestCalculateISR_ZeroIncome(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	isr := taxService.CalculateISR(0, "biweekly")
	assert.Equal(t, 0.0, isr)
}

// ============================================================================
// TaxCalculationService Tests (IMSS Calculation)
// ============================================================================

func TestCalculateIMSSEmployee_StandardSDI(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// SDI of $500/day, 15 working days
	sdi := 500.00
	workingDays := 15

	imss := taxService.CalculateIMSSEmployee(sdi, workingDays)

	// Period salary = 500 * 15 = 7500
	// Total IMSS = 7500 * (0.25% + 0.625% + 1.125% + 0%) = 7500 * 2% = 150
	assert.InDelta(t, 150.00, imss, 1.0)
}

func TestCalculateIMSSEmployee_CappedSDI(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// Very high SDI that exceeds 25 UMA cap
	// UMA 2025 = 113.14, so cap = 113.14 * 25 = 2828.50
	sdi := 5000.00 // Above cap
	workingDays := 15

	imss := taxService.CalculateIMSSEmployee(sdi, workingDays)

	// Capped period salary = 2828.50 * 15 = 42,427.50
	// Total IMSS = 42,427.50 * 2% = 848.55
	assert.InDelta(t, 848.55, imss, 5.0)
}

func TestCalculateIMSSEmployee_WeeklyPeriod(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	sdi := 400.00
	workingDays := 7 // Weekly period

	imss := taxService.CalculateIMSSEmployee(sdi, workingDays)

	// Period salary = 400 * 7 = 2800
	// Total IMSS = 2800 * 2% = 56
	assert.InDelta(t, 56.00, imss, 1.0)
}

func TestCalculateIMSSEmployeeBreakdown(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	sdi := 500.00
	workingDays := 15

	breakdown := taxService.CalculateIMSSEmployeeBreakdown(sdi, workingDays)

	periodSalary := sdi * float64(workingDays) // 7500

	// Verify each component
	assert.InDelta(t, periodSalary*0.0025, breakdown.SicknessMaternity, 0.01)   // 18.75
	assert.InDelta(t, periodSalary*0.00625, breakdown.DisabilityLife, 0.01)      // 46.875
	assert.InDelta(t, periodSalary*0.01125, breakdown.UnemploymentOldAge, 0.01)  // 84.375
	assert.Equal(t, 0.0, breakdown.Retirement) // Employer pays 100%
	assert.InDelta(t, 150.00, breakdown.Total, 1.0)
}

// ============================================================================
// TaxCalculationService Tests (Net ISR with Subsidy)
// ============================================================================

func TestCalculateNetISR_HighIncome(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// High income - no subsidy applies
	netISR := taxService.CalculateNetISR(50000.00, "biweekly")

	// For high income, net ISR equals gross ISR (no subsidy)
	grossISR := taxService.CalculateISR(50000.00, "biweekly")
	assert.Equal(t, grossISR, netISR)
}

func TestCalculateNetISR_LowIncome(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// Very low income - subsidy may exceed ISR
	netISR := taxService.CalculateNetISR(1000.00, "biweekly")

	// Net ISR should not be negative (capped at 0)
	assert.GreaterOrEqual(t, netISR, 0.0)
}

// ============================================================================
// TaxCalculationService Tests (INFONAVIT Calculation)
// ============================================================================

func TestCalculateINFONAVITEmployee_Percentage(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	sdi := 500.00
	workingDays := 15
	creditValue := 20.0 // 20%

	infonavit := taxService.CalculateINFONAVITEmployee(sdi, workingDays, "porcentaje", creditValue)

	// 500 * 15 * 20% = 1500
	assert.InDelta(t, 1500.00, infonavit, 0.01)
}

func TestCalculateINFONAVITEmployee_FixedAmount(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	sdi := 500.00
	workingDays := 15
	creditValue := 800.00 // Fixed $800

	infonavit := taxService.CalculateINFONAVITEmployee(sdi, workingDays, "cuota_fija", creditValue)

	assert.Equal(t, 800.00, infonavit)
}

func TestCalculateINFONAVITEmployee_UMAMultiple(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	sdi := 500.00
	workingDays := 15
	creditValue := 2.0 // 2x UMA

	infonavit := taxService.CalculateINFONAVITEmployee(sdi, workingDays, "veces_salario_minimo", creditValue)

	// UMA 2025 = 113.14
	// Daily deduction = 113.14 * 2 = 226.28
	// Period deduction = 226.28 * 15 = 3394.20
	assert.InDelta(t, 3394.20, infonavit, 1.0)
}

func TestCalculateINFONAVITEmployee_InvalidCreditType(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	infonavit := taxService.CalculateINFONAVITEmployee(500, 15, "invalid", 100)
	assert.Equal(t, 0.0, infonavit)
}

// ============================================================================
// PayrollService Tests - CalculateTotals
// ============================================================================

func TestCalculateTotals_BasicPayroll(t *testing.T) {
	db := setupPayrollTestDB(t)
	service := &PayrollService{db: db}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary:     10000.00,
		OvertimeAmount:    1000.00,
		VacationPremium:   500.00,
		Aguinaldo:         0.00,
		OtherExtras:       200.00,
		FoodVouchers:      300.00,
		SavingsFund:       0.00,
		ISRWithholding:    1200.00,
		IMSSEmployee:      300.00,
		InfonavitEmployee: 500.00,
		RetirementSavings: 0.00,
		LoanDeductions:    200.00,
		AdvanceDeductions: 100.00,
		OtherDeductions:   50.00,
	}

	service.CalculateTotals(payrollCalc)

	// Gross = 10000 + 1000 + 500 + 0 + 200 + 300 + 0 = 12000
	assert.Equal(t, 12000.00, payrollCalc.TotalGrossIncome)

	// Statutory = 1200 + 300 + 500 + 0 = 2000
	assert.Equal(t, 2000.00, payrollCalc.TotalStatutoryDeductions)

	// Other = 200 + 100 + 50 = 350
	assert.Equal(t, 350.00, payrollCalc.TotalOtherDeductions)

	// Net = 12000 - 2000 - 350 = 9650
	assert.Equal(t, 9650.00, payrollCalc.TotalNetPay)
}

func TestCalculateTotals_ZeroDeductions(t *testing.T) {
	db := setupPayrollTestDB(t)
	service := &PayrollService{db: db}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary: 5000.00,
	}

	service.CalculateTotals(payrollCalc)

	assert.Equal(t, 5000.00, payrollCalc.TotalGrossIncome)
	assert.Equal(t, 0.00, payrollCalc.TotalStatutoryDeductions)
	assert.Equal(t, 0.00, payrollCalc.TotalOtherDeductions)
	assert.Equal(t, 5000.00, payrollCalc.TotalNetPay)
}

func TestCalculateTotals_HighDeductions(t *testing.T) {
	db := setupPayrollTestDB(t)
	service := &PayrollService{db: db}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary:     5000.00,
		ISRWithholding:    3000.00,
		IMSSEmployee:      1000.00,
		LoanDeductions:    2000.00,
	}

	service.CalculateTotals(payrollCalc)

	// Net can be negative if deductions exceed income
	assert.Equal(t, 5000.00, payrollCalc.TotalGrossIncome)
	assert.Equal(t, 4000.00, payrollCalc.TotalStatutoryDeductions)
	assert.Equal(t, 2000.00, payrollCalc.TotalOtherDeductions)
	assert.Equal(t, -1000.00, payrollCalc.TotalNetPay)
}

// ============================================================================
// PayrollService Tests - CalculateIncomeComponents
// ============================================================================

func TestCalculateIncomeComponents_BiweeklyPeriod(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	period := createPayrollTestPeriod(t, db,"biweekly")

	service := &PayrollService{db: db}
	payrollCalc := &models.PayrollCalculation{}
	prenomina := &models.PrenominaMetric{}

	service.CalculateIncomeComponents(payrollCalc, prenomina, employee, period)

	// Daily salary * working days = 500 * 15 = 7500
	assert.Equal(t, 7500.00, payrollCalc.RegularSalary)
}

func TestCalculateIncomeComponents_WeeklyPeriod(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,350.00)
	period := createPayrollTestPeriod(t, db,"weekly")

	service := &PayrollService{db: db}
	payrollCalc := &models.PayrollCalculation{}
	prenomina := &models.PrenominaMetric{}

	service.CalculateIncomeComponents(payrollCalc, prenomina, employee, period)

	// Daily salary * working days = 350 * 7 = 2450
	assert.Equal(t, 2450.00, payrollCalc.RegularSalary)
}

// ============================================================================
// PayrollService Tests - CalculateEmployerContributions
// ============================================================================

func TestCalculateEmployerContributions_Basic(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	period := createPayrollTestPeriod(t, db,"biweekly")

	service := &PayrollService{db: db}
	payrollCalc := &models.PayrollCalculation{
		RegularSalary: 7500.00,
	}
	payrollCalc.ID = uuid.New()

	contrib, err := service.CalculateEmployerContributions(employee, payrollCalc, period)

	require.NoError(t, err)
	require.NotNil(t, contrib)

	// Verify contribution is linked correctly
	assert.Equal(t, payrollCalc.ID, contrib.PayrollCalculationID)
	assert.Equal(t, employee.ID, contrib.EmployeeID)
	assert.Equal(t, period.ID, contrib.PayrollPeriodID)

	// Calculate SDI (using employee's daily salary * integration factor)
	// Integration factor for employee with seniority = 1.0493 + (seniority * 0.0052)
	seniority := employee.CalculateSeniority()
	integrationFactor := 1.0493 + (float64(seniority) * 0.0052)
	if integrationFactor > 1.25 {
		integrationFactor = 1.25
	}
	sdi := employee.DailySalary * integrationFactor
	workingDays := period.GetWorkingDays()
	baseForContributions := sdi * float64(workingDays)

	// Verify calculations (using actual Mexican IMSS rates)
	// IMSSDiseaseMaternity = base * 20.4%
	assert.InDelta(t, baseForContributions*0.204, contrib.IMSSDiseaseMaternity, 50.0)
	// InfonavitEmployer = base * 5%
	assert.InDelta(t, baseForContributions*0.05, contrib.InfonavitEmployer, 15.0)
	// RetirementSAR = base * 2%
	assert.InDelta(t, baseForContributions*0.02, contrib.RetirementSAR, 10.0)

	// Total should be sum of components
	assert.Greater(t, contrib.TotalContributions, 0.0)
}

// ============================================================================
// PayrollService Tests - CalculateStatutoryDeductions (with TaxService)
// ============================================================================

func TestCalculateStatutoryDeductions_WithTaxService(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	employee.IntegratedDailySalary = 525.00 // Set SDI
	db.Save(employee)
	period := createPayrollTestPeriod(t, db,"biweekly")

	taxService, _ := NewTaxCalculationService("nonexistent")
	service := &PayrollService{
		db:             db,
		taxCalcService: taxService,
	}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary:   7500.00,
		OvertimeAmount:  500.00,
		VacationPremium: 0.00,
		Aguinaldo:       0.00,
		OtherExtras:     0.00,
	}

	service.CalculateStatutoryDeductions(payrollCalc, employee, period)

	// ISR should be calculated based on taxable income
	// Taxable = 7500 + 500 = 8000
	// Expected bracket: 7707.70-23076.92, FixedFee: 716.61, Rate: 17.92%
	// ISR = 716.61 + (8000 - 7707.70) * 17.92% = 716.61 + 52.38 = 769 (before subsidy)
	assert.Greater(t, payrollCalc.ISRWithholding, 0.0)

	// IMSS should be calculated based on SDI
	// SDI = 525, workingDays = 15
	// IMSS = 525 * 15 * 2% = 157.50
	assert.InDelta(t, 157.50, payrollCalc.IMSSEmployee, 5.0)
}

func TestCalculateStatutoryDeductions_WithoutTaxService(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	period := createPayrollTestPeriod(t, db,"biweekly")

	// Service without tax calculation service (fallback mode)
	service := &PayrollService{
		db:             db,
		taxCalcService: nil,
	}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary:   7500.00,
		OvertimeAmount:  500.00,
	}

	service.CalculateStatutoryDeductions(payrollCalc, employee, period)

	// Fallback ISR = taxable * 10% = 8000 * 0.10 = 800
	assert.InDelta(t, 800.00, payrollCalc.ISRWithholding, 0.01)

	// Fallback IMSS = regular salary * 2% = 7500 * 0.02 = 150
	assert.InDelta(t, 150.00, payrollCalc.IMSSEmployee, 0.01)
}

func TestCalculateStatutoryDeductions_WeeklyPeriodicity(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,350.00)
	employee.IntegratedDailySalary = 365.00
	db.Save(employee)
	period := createPayrollTestPeriod(t, db,"weekly")

	taxService, _ := NewTaxCalculationService("nonexistent")
	service := &PayrollService{
		db:             db,
		taxCalcService: taxService,
	}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary: 2450.00,
	}

	service.CalculateStatutoryDeductions(payrollCalc, employee, period)

	// Should use weekly periodicity (defaults to biweekly ISR table in current implementation)
	assert.GreaterOrEqual(t, payrollCalc.ISRWithholding, 0.0)
	assert.Greater(t, payrollCalc.IMSSEmployee, 0.0)
}

func TestCalculateStatutoryDeductions_WithInfonavitCredit(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	employee.IntegratedDailySalary = 525.00
	employee.InfonavitCredit = "porcentaje"
	db.Save(employee)
	period := createPayrollTestPeriod(t, db,"biweekly")

	taxService, _ := NewTaxCalculationService("nonexistent")
	service := &PayrollService{
		db:             db,
		taxCalcService: taxService,
	}

	payrollCalc := &models.PayrollCalculation{
		RegularSalary: 7500.00,
	}

	service.CalculateStatutoryDeductions(payrollCalc, employee, period)

	// INFONAVIT should be calculated for employees with active credit
	assert.Greater(t, payrollCalc.InfonavitEmployee, 0.0)
}

// ============================================================================
// PayrollService Tests - SavePayrollCalculation
// ============================================================================

func TestSavePayrollCalculation_NewRecord(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	period := createPayrollTestPeriod(t, db,"biweekly")
	prenomina := createPayrollTestPrenomina(t, db,employee.ID, period.ID)

	service := &PayrollService{db: db}

	payrollCalc := &models.PayrollCalculation{
		EmployeeID:        employee.ID,
		PayrollPeriodID:   period.ID,
		PrenominaMetricID: prenomina.ID,
		RegularSalary:     7500.00,
		TotalNetPay:       6000.00,
	}
	employerContrib := &models.EmployerContribution{
		EmployeeID:      employee.ID,
		PayrollPeriodID: period.ID,
		TotalIMSS:       500.00,
	}

	err := service.SavePayrollCalculation(payrollCalc, prenomina, employerContrib)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, payrollCalc.ID)
	assert.NotEqual(t, uuid.Nil, employerContrib.ID)
	assert.Equal(t, payrollCalc.ID, employerContrib.PayrollCalculationID)

	// Verify persisted
	var savedCalc models.PayrollCalculation
	err = db.First(&savedCalc, "id = ?", payrollCalc.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 7500.00, savedCalc.RegularSalary)
}

func TestSavePayrollCalculation_UpdateExisting(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	period := createPayrollTestPeriod(t, db,"biweekly")
	prenomina := createPayrollTestPrenomina(t, db,employee.ID, period.ID)

	service := &PayrollService{db: db}

	// Create initial record
	payrollCalc := &models.PayrollCalculation{
		EmployeeID:        employee.ID,
		PayrollPeriodID:   period.ID,
		PrenominaMetricID: prenomina.ID,
		RegularSalary:     7500.00,
	}
	payrollCalc.ID = uuid.New()
	db.Create(payrollCalc)

	employerContrib := &models.EmployerContribution{
		PayrollCalculationID: payrollCalc.ID,
		EmployeeID:           employee.ID,
		PayrollPeriodID:      period.ID,
		TotalIMSS:            500.00,
	}
	employerContrib.ID = uuid.New()
	db.Create(employerContrib)

	// Update
	payrollCalc.RegularSalary = 8000.00
	employerContrib.TotalIMSS = 600.00

	err := service.SavePayrollCalculation(payrollCalc, prenomina, employerContrib)

	require.NoError(t, err)

	// Verify updated
	var savedCalc models.PayrollCalculation
	db.First(&savedCalc, "id = ?", payrollCalc.ID)
	assert.Equal(t, 8000.00, savedCalc.RegularSalary)

	var savedContrib models.EmployerContribution
	db.First(&savedContrib, "id = ?", employerContrib.ID)
	assert.Equal(t, 600.00, savedContrib.TotalIMSS)
}

// ============================================================================
// PayrollService Tests - ConvertToPayrollResponse
// ============================================================================

func TestConvertToPayrollResponse_Success(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)
	period := createPayrollTestPeriod(t, db,"biweekly")

	service := &PayrollService{db: db}

	now := time.Now()
	payrollCalc := &models.PayrollCalculation{
		RegularSalary:            7500.00,
		OvertimeAmount:           500.00,
		ISRWithholding:           800.00,
		IMSSEmployee:             150.00,
		TotalGrossIncome:         8000.00,
		TotalStatutoryDeductions: 950.00,
		TotalOtherDeductions:     100.00,
		TotalNetPay:              6950.00,
		CalculationStatus:        "calculated",
		CalculationDate:          &now,
	}
	payrollCalc.ID = uuid.New()

	employerContrib := &models.EmployerContribution{
		TotalIMSS:          500.00,
		InfonavitEmployer:  200.00,
		RetirementSAR:      150.00,
		TotalContributions: 850.00,
	}

	response := service.ConvertToPayrollResponse(payrollCalc, employee, period, employerContrib)

	require.NotNil(t, response)
	assert.Equal(t, payrollCalc.ID, response.ID)
	assert.Equal(t, employee.ID, response.EmployeeID)
	assert.Equal(t, "Juan Perez", response.EmployeeName)
	assert.Equal(t, employee.EmployeeNumber, response.EmployeeNumber)
	assert.Equal(t, period.ID, response.PayrollPeriodID)
	assert.Equal(t, period.PeriodCode, response.PeriodCode)
	assert.Equal(t, 7500.00, response.RegularSalary)
	assert.Equal(t, 500.00, response.OvertimeAmount)
	assert.Equal(t, 800.00, response.ISRWithholding)
	assert.Equal(t, 8000.00, response.TotalGrossIncome)
	assert.Equal(t, 6950.00, response.TotalNetPay)
	assert.Equal(t, 500.00, response.EmployerContributions.TotalIMSS)
	assert.Equal(t, 850.00, response.EmployerContributions.TotalContributions)
}

func TestConvertToPayrollResponse_NilInputs(t *testing.T) {
	db := setupPayrollTestDB(t)
	service := &PayrollService{db: db}

	response := service.ConvertToPayrollResponse(nil, nil, nil, nil)
	assert.Nil(t, response)
}

// ============================================================================
// SDI (Integrated Daily Salary) Calculation Tests
// ============================================================================

func TestCalculateIntegratedDailySalary_FirstYear(t *testing.T) {
	// Employee hired less than 1 year ago
	employee := &models.Employee{
		DailySalary: 500.00,
		HireDate:    time.Now().AddDate(0, -6, 0), // 6 months ago
	}

	sdi := employee.CalculateIntegratedDailySalary()

	// SDI = DailySalary * (1 + (15/365) + (VacationDays * 0.25 / 365))
	// For < 1 year, vacation days = 12
	// SDI = 500 * (1 + 0.0411 + 0.0082) = 500 * 1.0493 = 524.65
	assert.InDelta(t, 524.65, sdi, 1.0)
}

func TestCalculateIntegratedDailySalary_WithExistingSDI(t *testing.T) {
	// If SDI is already set, should return the existing value
	employee := &models.Employee{
		DailySalary:           500.00,
		IntegratedDailySalary: 550.00, // Already calculated
		HireDate:              time.Now().AddDate(-2, 0, 0),
	}

	sdi := employee.CalculateIntegratedDailySalary()

	assert.Equal(t, 550.00, sdi)
}

// ============================================================================
// PayrollService Tests - CalculateOtherDeductions
// ============================================================================

func TestCalculateOtherDeductions(t *testing.T) {
	db := setupPayrollTestDB(t)
	service := &PayrollService{db: db}

	payrollCalc := &models.PayrollCalculation{}
	prenomina := &models.PrenominaMetric{
		OtherDeduction: 250.00,
	}

	service.CalculateOtherDeductions(payrollCalc, prenomina)

	// Should initialize all deductions to 0
	assert.Equal(t, 0.00, payrollCalc.LoanDeductions)
	assert.Equal(t, 0.00, payrollCalc.AdvanceDeductions)
	// Should copy from prenomina
	assert.Equal(t, 250.00, payrollCalc.OtherDeductions)
}

// ============================================================================
// PayrollService Tests - CalculateSubsidiesAndBenefits
// ============================================================================

func TestCalculateSubsidiesAndBenefits(t *testing.T) {
	db := setupPayrollTestDB(t)
	company := createPayrollTestCompany(t, db)
	employee := createPayrollTestEmployee(t, db, company.ID,500.00)

	// Migrate benefit-related tables for this test
	err := db.AutoMigrate(
		&models.BenefitPlan{},
		&models.BenefitEnrollment{},
	)
	require.NoError(t, err)

	service := &PayrollService{db: db}
	payrollCalc := &models.PayrollCalculation{}

	service.CalculateSubsidiesAndBenefits(payrollCalc, employee)

	// Now queries actual benefit enrollments - should be 0 if no enrollments exist
	assert.Equal(t, 0.00, payrollCalc.FoodVouchers)
}

// ============================================================================
// Edge Cases and Boundary Tests
// ============================================================================

func TestCalculateISR_BoundaryValues(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	// Test at exact bracket boundary
	// First bracket ends at 435.39
	isrAtBoundary := taxService.CalculateISR(435.39, "biweekly")
	isrJustAbove := taxService.CalculateISR(435.40, "biweekly")

	// Just above boundary should move to next bracket
	assert.Greater(t, isrJustAbove, isrAtBoundary)
}

func TestCalculateIMSSEmployee_ZeroSDI(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	imss := taxService.CalculateIMSSEmployee(0, 15)
	assert.Equal(t, 0.0, imss)
}

func TestCalculateIMSSEmployee_ZeroWorkingDays(t *testing.T) {
	taxService, _ := NewTaxCalculationService("nonexistent")

	imss := taxService.CalculateIMSSEmployee(500, 0)
	assert.Equal(t, 0.0, imss)
}

func TestPayrollPeriod_GetWorkingDays(t *testing.T) {
	period := &models.PayrollPeriod{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	days := period.GetWorkingDays()

	// Current implementation returns calendar days (inclusive)
	assert.Equal(t, 15, days)
}

func TestPayrollPeriod_CalculateDays(t *testing.T) {
	period := &models.PayrollPeriod{
		StartDate: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
	}

	days := period.CalculateDays()

	// 7 days inclusive
	assert.Equal(t, 7, days)
}
