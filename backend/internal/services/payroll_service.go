/*
Package services - Main Payroll Calculation Service

==============================================================================
FILE: internal/services/payroll_service.go
==============================================================================

DESCRIPTION:
    Core payroll calculation engine that processes final payroll including income,
    statutory deductions (ISR, IMSS, INFONAVIT), employer contributions, and
    generates payslips in PDF/XML formats. Integrates with tax calculation and
    CFDI services for Mexican compliance.

USER PERSPECTIVE:
    - Calculate complete payroll for employees
    - Generate PDF payslips with detailed breakdown
    - Create CFDI XML receipts for SAT compliance
    - Approve and process payments
    - Bulk calculate payroll for entire periods
    - View payroll summaries and concept totals

DEVELOPER GUIDELINES:
    OK to modify: Income component calculations, add new benefit types
    CAUTION: Statutory deduction calculations must use TaxCalculationService
    DO NOT modify: Payment approval workflow without financial review
    Note: Payroll requires approved prenomina before calculation

SYNTAX EXPLANATION:
    - CalculatePayroll processes one employee using prenomina metrics
    - CalculatePayrollDirect skips prenomina (for simplified flow)
    - CalculateStatutoryDeductions uses ISR tables and IMSS rates
    - TotalNetPay = GrossIncome - StatutoryDeductions - OtherDeductions
    - ApprovePayroll locks payroll for payment processing
    - ProcessPayment marks payroll as paid and updates period status

==============================================================================
*/
package services

import (
	"bytes"
	"errors"
	"fmt"

	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"

	"backend/internal/config"
	config_payroll "backend/internal/config/payroll"

	"backend/internal/dtos"
	"backend/internal/models"
	"backend/internal/repositories"
)

// ConvertToPayrollResponse converts a PayrollCalculation model to a PayrollCalculationResponse DTO.
func (s *PayrollService) ConvertToPayrollResponse(
    payrollCalc *models.PayrollCalculation,
    employee *models.Employee,
    period *models.PayrollPeriod,
    employerContrib *models.EmployerContribution,
) *dtos.PayrollCalculationResponse {
    if payrollCalc == nil || employee == nil || period == nil || employerContrib == nil {
        return nil
    }

    response := &dtos.PayrollCalculationResponse{
        ID:                  payrollCalc.ID,
        EmployeeID:          employee.ID,
        EmployeeName:        fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
        EmployeeNumber:      employee.EmployeeNumber,
        PayrollPeriodID:     period.ID,
        PeriodCode:          period.PeriodCode,
        PaymentDate:         period.PaymentDate,

        // Income
        RegularSalary:     payrollCalc.RegularSalary,
        OvertimeAmount:    payrollCalc.OvertimeAmount,
        VacationPremium:   payrollCalc.VacationPremium,
        Aguinaldo:         payrollCalc.Aguinaldo,
        OtherExtras:       payrollCalc.OtherExtras,

        // Deductions
        ISRWithholding:    payrollCalc.ISRWithholding,
        IMSSEmployee:      payrollCalc.IMSSEmployee,
        InfonavitEmployee: payrollCalc.InfonavitEmployee,
        RetirementSavings: payrollCalc.RetirementSavings,

        // Other deductions
        LoanDeductions:    payrollCalc.LoanDeductions,
        AdvanceDeductions: payrollCalc.AdvanceDeductions,
        OtherDeductions:   payrollCalc.OtherDeductions,

        // Subsidies and benefits
        FoodVouchers:      payrollCalc.FoodVouchers,
        SavingsFund:       payrollCalc.SavingsFund,

        // Totals
        TotalGrossIncome:    payrollCalc.TotalGrossIncome,
        TotalDeductions:     payrollCalc.TotalStatutoryDeductions + payrollCalc.TotalOtherDeductions, // Calculated from model fields
        TotalNetPay:         payrollCalc.TotalNetPay,

        // Employer contributions
        EmployerContributions: dtos.EmployerContributionResponse{
            TotalIMSS:          employerContrib.TotalIMSS,
            TotalInfonavit:     employerContrib.InfonavitEmployer,
            TotalRetirement:    employerContrib.RetirementSAR,
            TotalContributions: employerContrib.TotalContributions,
        },

        // Metadata
        CalculationStatus: payrollCalc.CalculationStatus,
        CalculationDate:   *payrollCalc.CalculationDate, // Use the actual time.Time field
        PayrollStatus:     payrollCalc.PayrollStatus,
    }

    return response
}

// CreatePayrollDetails creates detailed payroll entries based on the calculation.
func (s *PayrollService) CreatePayrollDetails(payrollCalc *models.PayrollCalculation) {
    // Placeholder: In a real system, this would create individual entries for each income and deduction.
    // For now, it just ensures the method exists.
}

// SavePayrollCalculation saves the payroll calculation and its associated employer contributions.
func (s *PayrollService) SavePayrollCalculation(
    payrollCalc *models.PayrollCalculation,
    prenominaMetric *models.PrenominaMetric,
    employerContrib *models.EmployerContribution,
) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if payrollCalc.ID == uuid.Nil {
            payrollCalc.ID = uuid.New()
            if err := tx.Create(payrollCalc).Error; err != nil {
                return err
            }
        } else {
            if err := tx.Save(payrollCalc).Error; err != nil {
                return err
            }
        }

        employerContrib.PayrollCalculationID = payrollCalc.ID
        if employerContrib.ID == uuid.Nil {
            employerContrib.ID = uuid.New()
            if err := tx.Create(employerContrib).Error; err != nil {
                return err
            }
        } else {
            if err := tx.Save(employerContrib).Error; err != nil {
                return err
            }
        }

        return nil
    })
}

// CalculateEmployerContributions calculates the employer's contributions for an employee.
func (s *PayrollService) CalculateEmployerContributions(
    employee *models.Employee,
    payrollCalc *models.PayrollCalculation,
    period *models.PayrollPeriod,
) (*models.EmployerContribution, error) {
    // Use SDI (Integrated Daily Salary) for employer contributions
    sdi := employee.IntegratedDailySalary
    if sdi == 0 {
        // Calculate SDI with proper integration factor based on seniority
        seniority := employee.CalculateSeniority()
        integrationFactor := 1.0493 // Minimum factor for 1st year
        if seniority >= 1 {
            // Adjust factor based on seniority (increases with years of service)
            integrationFactor = 1.0493 + (float64(seniority-1) * 0.0052)
            if integrationFactor > 1.25 { // Cap at maximum
                integrationFactor = 1.25
            }
        }
        sdi = employee.DailySalary * integrationFactor
    }

    workingDays := float64(period.GetWorkingDays())
    baseForContributions := sdi * workingDays

    employerContrib := &models.EmployerContribution{
        PayrollCalculationID: payrollCalc.ID,
        EmployeeID:           employee.ID,
        PayrollPeriodID:      period.ID,
    }

    // TODO: Replace with actual Mexican IMSS rates that vary by risk classification
    // Current rates are approximations - actual rates depend on:
    // - Company risk classification (Class I-V)
    // - Salary brackets (UMA multiples)
    // - Current year rates published by IMSS

    // Disease & Maternity (Enfermedad y Maternidad): ~20.4% employer, 0.4% employee
    employerContrib.IMSSDiseaseMaternity = baseForContributions * 0.204

    // Work Risk (Riesgo de Trabajo): Varies by company risk class (0.5% to 15%)
    // Using Class I (lowest risk) as default
    employerContrib.IMSSWorkRisk = baseForContributions * 0.00540

    // Disability & Life (Invalidez y Vida): 1.75% employer, 0.625% employee
    employerContrib.IMSSDisabilityLife = baseForContributions * 0.0175

    // Retirement (Retiro): 2% employer only
    employerContrib.IMSSRetirement = baseForContributions * 0.02

    // Childcare (Guardería y Prestaciones Sociales): 1% employer only
    employerContrib.IMSSChildcare = baseForContributions * 0.01

    // INFONAVIT: 5% employer only
    employerContrib.InfonavitEmployer = baseForContributions * 0.05

    // SAR (Retirement Savings): 2% employer only
    employerContrib.RetirementSAR = baseForContributions * 0.02

    // Calculate totals
    employerContrib.TotalIMSS = employerContrib.IMSSDiseaseMaternity +
        employerContrib.IMSSDisabilityLife +
        employerContrib.IMSSRetirement +
        employerContrib.IMSSChildcare +
        employerContrib.IMSSWorkRisk

    employerContrib.TotalContributions = employerContrib.TotalIMSS +
        employerContrib.InfonavitEmployer +
        employerContrib.RetirementSAR

    return employerContrib, nil
}

// CalculateTotals calculates the total gross income, total deductions, and total net pay.
func (s *PayrollService) CalculateTotals(payrollCalc *models.PayrollCalculation) {
    payrollCalc.TotalGrossIncome = payrollCalc.RegularSalary + payrollCalc.OvertimeAmount + payrollCalc.VacationPremium + payrollCalc.Aguinaldo + payrollCalc.OtherExtras + payrollCalc.FoodVouchers + payrollCalc.SavingsFund
    payrollCalc.TotalStatutoryDeductions = payrollCalc.ISRWithholding + payrollCalc.IMSSEmployee + payrollCalc.InfonavitEmployee + payrollCalc.RetirementSavings
    payrollCalc.TotalOtherDeductions = payrollCalc.LoanDeductions + payrollCalc.AdvanceDeductions + payrollCalc.OtherDeductions
    totalDeductions := payrollCalc.TotalStatutoryDeductions + payrollCalc.TotalOtherDeductions // Calculate total deductions for net pay calculation
    payrollCalc.TotalNetPay = payrollCalc.TotalGrossIncome - totalDeductions
}

// CalculateSubsidiesAndBenefits calculates subsidies and benefits for a payroll.
func (s *PayrollService) CalculateSubsidiesAndBenefits(
    payrollCalc *models.PayrollCalculation,
    employee *models.Employee,
) {
    // Calculate food vouchers from actual benefit enrollments
    payrollCalc.FoodVouchers = 0.0

    // Query active benefit enrollments for food vouchers/meal benefits
    var benefitEnrollments []models.BenefitEnrollment
    err := s.db.Joins("JOIN benefit_plans ON benefit_plans.id = benefit_enrollments.plan_id").
        Where("benefit_enrollments.employee_id = ?", employee.ID).
        Where("benefit_enrollments.status = ?", "active").
        Where("benefit_plans.benefit_type IN (?)", []string{"wellness", "other"}).
        Where("benefit_plans.code LIKE ?", "%food%").
        Find(&benefitEnrollments).Error

    if err == nil {
        for _, enrollment := range benefitEnrollments {
            // Add the employee cost as a benefit (typically employer-provided)
            payrollCalc.FoodVouchers += enrollment.EmployeeCost
        }
    }

    // If no food voucher benefits found, leave as 0 instead of arbitrary default
    // This ensures accurate payroll calculation
}

// CalculateOtherDeductions calculates other deductions (e.g., loans, advances) for a payroll.
func (s *PayrollService) CalculateOtherDeductions(
    payrollCalc *models.PayrollCalculation,
    prenominaMetric *models.PrenominaMetric,
) {
    // Initialize all deductions to 0
    payrollCalc.LoanDeductions = 0.0
    payrollCalc.AdvanceDeductions = 0.0
    payrollCalc.OtherDeductions = 0.0

    // Use prenomina metrics for other deductions
    if prenominaMetric != nil {
        payrollCalc.OtherDeductions = prenominaMetric.OtherDeduction
    }

    // TODO: In future, integrate with loans/advances module to calculate:
    // - Active loan installments
    // - Salary advances taken
    // - Court-ordered garnishments
    // - Union dues
    // For now, rely on prenomina data or manual adjustments
}

// CalculateStatutoryDeductions calculates all statutory deductions (e.g., ISR, IMSS) for a payroll.
func (s *PayrollService) CalculateStatutoryDeductions(
    payrollCalc *models.PayrollCalculation,
    employee *models.Employee,
    period *models.PayrollPeriod,
) {
    // Determine periodicity based on payroll period frequency
    periodicity := "biweekly" // Default
    switch period.Frequency {
    case "weekly":
        periodicity = "weekly"
    case "biweekly":
        periodicity = "biweekly"
    case "monthly":
        periodicity = "monthly"
    }

    // Calculate taxable income (gross income minus exempt items)
    // For simplicity, we use the gross income as taxable base
    // In a full implementation, this would exclude exempt amounts per Mexican tax law
    taxableIncome := payrollCalc.RegularSalary + payrollCalc.OvertimeAmount + payrollCalc.VacationPremium + payrollCalc.Aguinaldo + payrollCalc.OtherExtras

    // Calculate ISR using the tax calculation service with proper tax brackets
    if s.taxCalcService != nil {
        // Use net ISR calculation (after applying employment subsidy)
        payrollCalc.ISRWithholding = s.taxCalcService.CalculateNetISR(taxableIncome, periodicity)

        // Calculate IMSS employee contribution
        // Use SDI (Integrated Daily Salary) for IMSS calculations
        sdi := employee.IntegratedDailySalary
        if sdi == 0 {
            // Calculate SDI with proper integration factor based on seniority
            // Integration factor increases with years of service per Mexican labor law
            seniority := employee.CalculateSeniority()
            integrationFactor := 1.0493 // Minimum factor for 1st year
            if seniority >= 1 {
                // Factor increases ~0.52% per year of seniority
                integrationFactor = 1.0493 + (float64(seniority-1) * 0.0052)
                if integrationFactor > 1.25 { // Cap at maximum 25% integration
                    integrationFactor = 1.25
                }
            }
            sdi = employee.DailySalary * integrationFactor
        }
        workingDays := period.GetWorkingDays()
        payrollCalc.IMSSEmployee = s.taxCalcService.CalculateIMSSEmployee(sdi, workingDays)

        // Calculate INFONAVIT employee deduction (if applicable)
        // INFONAVIT deductions depend on whether employee has an active credit
        if employee.InfonavitCredit != "" && employee.InfonavitCredit != "none" {
            // Assume percentage-based deduction as default
            // In production, this would lookup the credit type and value from employee record
            payrollCalc.InfonavitEmployee = s.taxCalcService.CalculateINFONAVITEmployee(
                sdi, workingDays, "porcentaje", 20.0, // Default 20% for illustration
            )
        }
    } else {
        // Fallback to basic calculation if tax service not available
        payrollCalc.ISRWithholding = taxableIncome * 0.10
        payrollCalc.IMSSEmployee = payrollCalc.RegularSalary * 0.02
    }
}

// CalculateIncomeComponents calculates all income-related components for a payroll.
func (s *PayrollService) CalculateIncomeComponents(
    payrollCalc *models.PayrollCalculation,
    prenominaMetric *models.PrenominaMetric,
    employee *models.Employee,
    period *models.PayrollPeriod,
) {
    // Placeholder implementation: Assign a basic regular salary for now
    payrollCalc.RegularSalary = employee.DailySalary * float64(period.GetWorkingDays())
    // Other income components would be calculated here based on prenominaMetric, employee, and period.
}

// PayrollService handles payroll calculation business logic
type PayrollService struct {
	payrollRepo    *repositories.PayrollRepository
	employeeRepo   *repositories.EmployeeRepository
	periodRepo     *repositories.PayrollPeriodRepository
	prenominaRepo  *repositories.PrenominaRepository
	incidenceRepo  *repositories.IncidenceRepository
	config         *config_payroll.PayrollConfig
	taxConfig      *config_payroll.MexicanTaxConfig
	taxCalcService *TaxCalculationService
	cfdiService    *CfdiService
	db             *gorm.DB
}

// NewPayrollService creates a new payroll service
func NewPayrollService(
	db *gorm.DB,
	appConfig *config.AppConfig,
) *PayrollService {
	// Initialize tax calculation service with config path
	taxCalcService, err := NewTaxCalculationService("configs")
	if err != nil {
		// Log warning but continue - service has fallback defaults
		fmt.Printf("Warning: Could not load tax config files: %v\n", err)
	}

	return &PayrollService{
		payrollRepo:    repositories.NewPayrollRepository(db),
		employeeRepo:   repositories.NewEmployeeRepository(db),
		periodRepo:     repositories.NewPayrollPeriodRepository(db),
		prenominaRepo:  repositories.NewPrenominaRepository(db),
		incidenceRepo:  repositories.NewIncidenceRepository(db),
		config:         appConfig.PayrollConfig,
		taxConfig:      &appConfig.PayrollConfig.MexicanTaxConfig,
		taxCalcService: taxCalcService,
		cfdiService:    NewCfdiService("path/to/cert.cer", "path/to/key.key", "password"),
		db:             db,
	}
}

// CalculatePayroll calculates complete payroll for an employee
func (s *PayrollService) CalculatePayroll(
    employeeID, periodID uuid.UUID,
    calculateSDI bool,
    calculatedBy uuid.UUID,
) (*dtos.PayrollCalculationResponse, error) {
    // Get employee
    employee, err := s.employeeRepo.FindByID(employeeID)
    if err != nil {
        return nil, fmt.Errorf("employee not found: %w", err)
    }
    
    // Get payroll period
    period, err := s.periodRepo.FindByID(periodID)
    if err != nil {
        return nil, fmt.Errorf("payroll period not found: %w", err)
    }
    
    // Check if period is open for calculation
    if !period.IsOpen() && period.Status != "calculated" {
        return nil, errors.New("payroll period is not open for calculation")
    }
    
    // Get prenomina metrics
    prenominaMetric, err := s.prenominaRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil {
        return nil, fmt.Errorf("prenomina metrics not found: %w", err)
    }
    
    // Check if prenomina is approved
    if prenominaMetric.CalculationStatus != "approved" {
        return nil, errors.New("prenomina must be approved before payroll calculation")
    }
    
    // Calculate SDI if requested
    if calculateSDI {
        employee.IntegratedDailySalary = employee.CalculateIntegratedDailySalary()
        if err := s.employeeRepo.Update(employee); err != nil {
            return nil, fmt.Errorf("error updating SDI: %w", err)
        }
    }
    
    // Get existing payroll calculation or create new
    payrollCalc, err := s.payrollRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("error fetching payroll calculation: %w", err)
    }
    
    if payrollCalc == nil {
        payrollCalc = &models.PayrollCalculation{
            EmployeeID:        employeeID,
            PayrollPeriodID:   periodID,
            PrenominaMetricID: prenominaMetric.ID,
            CalculationStatus: "calculated",
        }
        now := time.Now()
        payrollCalc.CalculationDate = &now
    } else {
        payrollCalc.CalculationStatus = "calculated"
        now := time.Now()
        payrollCalc.CalculationDate = &now
    }
    
    // Calculate payroll components
    s.CalculateIncomeComponents(payrollCalc, prenominaMetric, employee, period)
    s.CalculateStatutoryDeductions(payrollCalc, employee, period)
    s.CalculateOtherDeductions(payrollCalc, prenominaMetric)
    s.CalculateSubsidiesAndBenefits(payrollCalc, employee)
    s.CalculateTotals(payrollCalc)
    
    // Calculate employer contributions
    employerContrib, err := s.CalculateEmployerContributions(employee, payrollCalc, period)
    if err != nil {
        return nil, fmt.Errorf("error calculating employer contributions: %w", err)
    }
    
    // Save payroll calculation
    if err := s.SavePayrollCalculation(payrollCalc, prenominaMetric, employerContrib); err != nil {
        return nil, fmt.Errorf("error saving payroll calculation: %w", err)
    }
    
    // Create payroll details
    s.CreatePayrollDetails(payrollCalc)
    
    return s.ConvertToPayrollResponse(payrollCalc, employee, period, employerContrib), nil
}

// GetPayrollCalculation retrieves a single payroll calculation by employee and period ID
func (s *PayrollService) GetPayrollCalculation(employeeID, periodID uuid.UUID) (*dtos.PayrollCalculationResponse, error) {
    payrollCalc, err := s.payrollRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("payroll calculation not found")
        }
        return nil, fmt.Errorf("error fetching payroll calculation: %w", err)
    }

    // Get employee and period for response
    employee, err := s.employeeRepo.FindByID(employeeID)
    if err != nil {
        return nil, fmt.Errorf("error fetching employee: %w", err)
    }
    
    period, err := s.periodRepo.FindByID(periodID)
    if err != nil {
        return nil, fmt.Errorf("error fetching payroll period: %w", err)
    }

    employerContrib, err := s.payrollRepo.FindEmployerContributionByPayrollCalculationID(payrollCalc.ID)
    if err != nil {
        return nil, fmt.Errorf("error fetching employer contribution: %w", err)
    }

    return s.ConvertToPayrollResponse(payrollCalc, employee, period, employerContrib), nil
}

// GetPayrollByPeriod retrieves all payroll calculations for a given period
func (s *PayrollService) GetPayrollByPeriod(periodID uuid.UUID) ([]dtos.PayrollCalculationResponse, error) {
	calculations, err := s.payrollRepo.FindByPeriod(periodID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve payroll calculations for period %s: %w", periodID, err)
	}

	if len(calculations) == 0 {
		return []dtos.PayrollCalculationResponse{}, nil
	}

	var responses []dtos.PayrollCalculationResponse
	for _, calc := range calculations {
        // Need to fetch employee and period for each calculation to convert to DTO
        employee, err := s.employeeRepo.FindByID(calc.EmployeeID)
        if err != nil {
            return nil, fmt.Errorf("error fetching employee for payroll calculation %s: %w", calc.ID, err)
        }
        period, err := s.periodRepo.FindByID(calc.PayrollPeriodID)
        if err != nil {
            return nil, fmt.Errorf("error fetching payroll period for payroll calculation %s: %w", calc.ID, err)
        }
        employerContrib, err := s.payrollRepo.FindEmployerContributionByPayrollCalculationID(calc.ID)
        if err != nil {
            return nil, fmt.Errorf("error fetching employer contribution for payroll calculation %s: %w", calc.ID, err)
        }
		responses = append(responses, *s.ConvertToPayrollResponse(&calc, employee, period, employerContrib))
	}

	return responses, nil
}

// BulkCalculatePayroll calculates payroll for multiple employees
func (s *PayrollService) BulkCalculatePayroll(
    periodID uuid.UUID,
    employeeIDs []uuid.UUID,
    calculateAll bool,
    calculatedBy uuid.UUID,
) (*dtos.PayrollBulkCalculationResponse, error) {
    // Get payroll period
    period, err := s.periodRepo.FindByID(periodID)
    if err != nil {
        return nil, fmt.Errorf("payroll period not found: %w", err)
    }

    // Check if period is open for calculation
    if !period.IsOpen() && period.Status != "calculated" {
        return nil, errors.New("payroll period is not open for calculation")
    }

    // Get employees to calculate based on period frequency
    var employees []models.Employee
    if calculateAll {
        // Get all active employees with matching pay frequency/collar type
        activeFilter := map[string]interface{}{"active_only": true}
        allEmployees, _, err := s.employeeRepo.List(1, 10000, activeFilter)
        if err != nil {
            return nil, fmt.Errorf("could not fetch employees: %w", err)
        }

        // Filter employees based on period frequency
        // Weekly -> blue_collar and gray_collar employees
        // Biweekly -> white_collar employees
        // Monthly -> all employees (special cases, admin decides)
        for _, emp := range allEmployees {
            switch period.Frequency {
            case "weekly":
                if emp.CollarType == "blue_collar" || emp.CollarType == "gray_collar" {
                    employees = append(employees, emp)
                }
            case "biweekly":
                if emp.CollarType == "white_collar" {
                    employees = append(employees, emp)
                }
            case "monthly":
                // Monthly periods include all employees by default (special cases)
                employees = append(employees, emp)
            default:
                // If frequency not recognized, include all
                employees = append(employees, emp)
            }
        }
    } else {
        // Get specific employees
        for _, id := range employeeIDs {
            emp, err := s.employeeRepo.FindByID(id)
            if err != nil {
                continue // Skip if employee not found
            }
            employees = append(employees, *emp)
        }
    }

    response := &dtos.PayrollBulkCalculationResponse{
        PeriodCode:      period.PeriodCode,
        TotalCalculated: len(employees),
        TotalSuccess:    0,
        TotalFailed:     0,
        Results:         make([]dtos.PayrollCalculationResult, 0),
    }

    totalGross := 0.0
    totalNet := 0.0

    for _, employee := range employees {
        result := dtos.PayrollCalculationResult{
            EmployeeID:   employee.ID,
            EmployeeName: fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
            Success:      false,
        }

        // Calculate payroll directly (simplified version that doesn't require prenomina)
        payrollCalc, err := s.CalculatePayrollDirect(&employee, period, calculatedBy)
        if err != nil {
            result.Error = err.Error()
            response.TotalFailed++
        } else {
            result.Success = true
            result.GrossIncome = payrollCalc.TotalGrossIncome
            result.NetIncome = payrollCalc.TotalNetPay
            response.TotalSuccess++
            totalGross += payrollCalc.TotalGrossIncome
            totalNet += payrollCalc.TotalNetPay
        }

        response.Results = append(response.Results, result)
    }

    response.TotalGross = totalGross
    response.TotalNet = totalNet

    // Update period totals
    period.TotalGross = totalGross
    period.TotalNet = totalNet
    period.TotalDeductions = totalGross - totalNet
    if period.Status == "open" {
        period.Status = "calculated"
    }
    s.db.Save(period)

    return response, nil
}

// CalculatePayrollDirect calculates payroll directly without requiring prenomina
func (s *PayrollService) CalculatePayrollDirect(
    employee *models.Employee,
    period *models.PayrollPeriod,
    calculatedBy uuid.UUID,
) (*models.PayrollCalculation, error) {
    // Get or create payroll calculation
    payrollCalc, err := s.payrollRepo.FindByEmployeeAndPeriod(employee.ID, period.ID)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("error fetching payroll calculation: %w", err)
    }

    now := time.Now()
    isNewRecord := false
    if payrollCalc == nil {
        isNewRecord = true
        payrollCalc = &models.PayrollCalculation{
            CalculationStatus: "pending",
            PayrollStatus:     "pending",
            CalculationDate:   &now,
        }
        payrollCalc.ID = uuid.New()
        payrollCalc.EmployeeID = employee.ID
        payrollCalc.PayrollPeriodID = period.ID
    }

    // Calculate regular salary based on working days
    workingDays := period.GetWorkingDays()

    // Get approved incidences for this employee and period
    incidences, _ := s.incidenceRepo.FindByEmployeeAndPeriod(employee.ID, period.ID)

    // Calculate incidence effects on payroll
    incidenceDeductions := 0.0
    incidenceAdditions := 0.0
    absenceDays := 0.0

    for _, incidence := range incidences {
        // Only process approved incidences
        if incidence.Status != "approved" && incidence.Status != "processed" {
            continue
        }

        // Load incidence type if not preloaded
        var incType models.IncidenceType
        if incidence.IncidenceType != nil {
            incType = *incidence.IncidenceType
        } else {
            s.db.First(&incType, "id = ?", incidence.IncidenceTypeID)
        }

        // Calculate amount based on type category and effect
        switch incType.Category {
        case "vacation":
            // Vacation days are PAID days - they don't reduce salary
            // Instead, add vacation premium (25% of vacation days pay per Mexican law)
            vacationPay := employee.DailySalary * incidence.Quantity
            vacationPremium := vacationPay * 0.25 // 25% vacation premium (prima vacacional)
            payrollCalc.VacationPremium += vacationPremium
            // Vacation days are worked days, so no deduction from salary
        case "sick":
            // Sick days - typically covered by IMSS after 3 days
            // For simplicity, treat first 3 days as unpaid, rest as IMSS-covered
            if incidence.Quantity > 3 {
                absenceDays += 3 // Only first 3 days are unpaid
            } else {
                absenceDays += incidence.Quantity
            }
        case "absence", "delay":
            // Unpaid absences and delays reduce salary
            absenceDays += incidence.Quantity
            if incType.EffectType == "negative" {
                incidenceDeductions += incidence.CalculatedAmount
            }
        case "overtime", "bonus":
            // Overtime and bonuses add to pay
            if incType.EffectType == "positive" {
                incidenceAdditions += incidence.CalculatedAmount
            }
        case "deduction":
            // Other deductions
            incidenceDeductions += incidence.CalculatedAmount
        default:
            // Handle based on effect type for other categories
            switch incType.EffectType {
            case "negative":
                incidenceDeductions += incidence.CalculatedAmount
            case "positive":
                incidenceAdditions += incidence.CalculatedAmount
            // neutral has no effect on pay
            }
        }

        // Mark incidence as processed
        incidence.Status = "processed"
        s.db.Save(&incidence)
    }

    // Adjust salary for absence days
    effectiveWorkingDays := float64(workingDays) - absenceDays
    if effectiveWorkingDays < 0 {
        effectiveWorkingDays = 0
    }
    payrollCalc.RegularSalary = employee.DailySalary * effectiveWorkingDays

    // Add overtime from incidences (if any)
    payrollCalc.OvertimeAmount += incidenceAdditions

    // Calculate gross income
    payrollCalc.TotalGrossIncome = payrollCalc.RegularSalary + payrollCalc.OvertimeAmount +
        payrollCalc.VacationPremium + payrollCalc.Aguinaldo + payrollCalc.OtherExtras

    // Calculate statutory deductions using tax service
    s.CalculateStatutoryDeductions(payrollCalc, employee, period)

    // Calculate employer contributions (employee, payrollCalc, period)
    employerContrib, _ := s.CalculateEmployerContributions(employee, payrollCalc, period)

    // Add incidence-based deductions to other deductions
    payrollCalc.OtherDeductions += incidenceDeductions

    // Calculate totals
    payrollCalc.TotalStatutoryDeductions = payrollCalc.ISRWithholding + payrollCalc.IMSSEmployee +
        payrollCalc.InfonavitEmployee + payrollCalc.RetirementSavings
    payrollCalc.TotalOtherDeductions = payrollCalc.LoanDeductions + payrollCalc.AdvanceDeductions +
        payrollCalc.OtherDeductions
    payrollCalc.TotalNetPay = payrollCalc.TotalGrossIncome - payrollCalc.TotalStatutoryDeductions -
        payrollCalc.TotalOtherDeductions

    // Update status
    payrollCalc.CalculationStatus = "calculated"
    payrollCalc.PayrollStatus = "processed"
    payrollCalc.CalculationDate = &now

    // Save payroll calculation
    if isNewRecord {
        if err := s.db.Create(payrollCalc).Error; err != nil {
            return nil, fmt.Errorf("error creating payroll calculation: %w", err)
        }
    } else {
        if err := s.db.Save(payrollCalc).Error; err != nil {
            return nil, fmt.Errorf("error updating payroll calculation: %w", err)
        }
    }

    // Save employer contributions
    employerContrib.PayrollCalculationID = payrollCalc.ID
    if employerContrib.ID == uuid.Nil {
        employerContrib.ID = uuid.New()
        s.db.Create(employerContrib)
    } else {
        s.db.Save(employerContrib)
    }

    return payrollCalc, nil
}

// GetPayrollSummary returns a summary of a payroll period.
func (s *PayrollService) GetPayrollSummary(periodID uuid.UUID) (*dtos.PayrollSummary, error) {
	calculations, err := s.payrollRepo.FindByPeriod(periodID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve payroll calculations for summary: %w", err)
	}

	if len(calculations) == 0 {
		return nil, errors.New("no payroll calculations found for this period")
	}

	var totalGross, totalDeductions, totalNet, employerContributions float64
	employeeSummaries := make([]dtos.PayrollSummaryEmployee, len(calculations))

	for i, calc := range calculations {
		totalGross += calc.TotalGrossIncome
		totalDeductions += calc.TotalStatutoryDeductions + calc.TotalOtherDeductions
		totalNet += calc.TotalNetPay
		// Note: This is a simplified sum. A real implementation might need to be more careful.
		employerContributions += calc.IMSSEmployer + calc.InfonavitEmployer

		employeeSummaries[i] = dtos.PayrollSummaryEmployee{
			EmployeeID:   calc.EmployeeID.String(),
			EmployeeName: calc.Employee.FirstName + " " + calc.Employee.LastName,
			Gross:        calc.TotalGrossIncome,
			Deductions:   calc.TotalStatutoryDeductions + calc.TotalOtherDeductions,
			Net:          calc.TotalNetPay,
			Status:       calc.CalculationStatus,
		}
	}

	summary := &dtos.PayrollSummary{
		PeriodID:              periodID.String(),
		TotalGross:            totalGross,
		TotalDeductions:       totalDeductions,
		TotalNet:              totalNet,
		EmployerContributions: employerContributions,
		Employees:             employeeSummaries,
	}

	return summary, nil
}

// ApprovePayroll approves all payroll calculations for a given period.
func (s *PayrollService) ApprovePayroll(periodID uuid.UUID, approvedBy uuid.UUID) error {
    calculations, err := s.payrollRepo.FindByPeriod(periodID)
    if err != nil {
        return fmt.Errorf("failed to retrieve payroll calculations for period %s: %w", periodID, err)
    }

    if len(calculations) == 0 {
        return errors.New("no payroll calculations found to approve for this period")
    }

    tx := s.db.Begin()
    if tx.Error != nil {
        return fmt.Errorf("failed to start transaction: %w", tx.Error)
    }
    defer tx.Rollback()

    for _, calc := range calculations {
        calc.CalculationStatus = "approved"
        calc.ApprovedBy = &approvedBy
        now := time.Now()
        calc.ApprovedAt = &now
        if err := tx.Save(&calc).Error; err != nil {
            return fmt.Errorf("failed to approve payroll calculation %s: %w", calc.ID, err)
        }
    }

    // Optionally update the payroll period status
    period, err := s.periodRepo.FindByID(periodID)
    if err != nil {
        return fmt.Errorf("payroll period not found: %w", err)
    }
    period.Status = "approved"
    if err := tx.Save(period).Error; err != nil {
        return fmt.Errorf("failed to update payroll period status: %w", err)
    }

    return tx.Commit().Error
}

// ProcessPayment updates the status of all payroll calculations for a period to 'paid'
func (s *PayrollService) ProcessPayment(periodID uuid.UUID, processedBy uuid.UUID) error {
	calculations, err := s.payrollRepo.FindByPeriod(periodID)
	if err != nil {
		return fmt.Errorf("failed to retrieve payroll calculations for period %s: %w", periodID, err)
	}

	if len(calculations) == 0 {
		return errors.New("no payroll calculations found to process payment for this period")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	for _, calc := range calculations {
		// Only process payments for approved payrolls
		if calc.CalculationStatus != "approved" {
			tx.Rollback()
			return fmt.Errorf("payroll calculation %s for employee %s is not approved", calc.ID, calc.EmployeeID)
		}
		calc.PayrollStatus = "paid"
		calc.ProcessedBy = &processedBy
		now := time.Now()
		calc.ProcessedAt = &now
		if err := tx.Save(&calc).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update payroll calculation %s status to paid: %w", calc.ID, err)
		}
	}

	// Optionally update the payroll period status
	period, err := s.periodRepo.FindByID(periodID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("payroll period not found: %w", err)
	}
	period.Status = "paid"
	if err := tx.Save(period).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update payroll period status to paid: %w", err)
	}

	return tx.Commit().Error
}

// GetPaymentStatus retrieves the payment status for a given payroll period
func (s *PayrollService) GetPaymentStatus(periodID uuid.UUID) (string, error) {
	period, err := s.periodRepo.FindByID(periodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("payroll period not found")
		}
		return "", fmt.Errorf("error fetching payroll period: %w", err)
	}
	return period.Status, nil
}

// GeneratePayslip generates payslip for employee
func (s *PayrollService) GeneratePayslip(
    employeeID, periodID uuid.UUID,
    format string, // pdf, xml, html
) ([]byte, error) {
    // Get payroll calculation
    payroll, err := s.payrollRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil {
        return nil, fmt.Errorf("payroll calculation not found: %w", err)
    }
    
    // Generate payslip based on format
    switch format {
    case "pdf":
        return s.GeneratePDFPayslip(payroll)
    case "xml":
        return s.cfdiService.GenerateCfdiXML(payroll)
    case "html":
        return s.GenerateHTMLPayslip(payroll)
    default:
        return nil, errors.New("unsupported format")
    }
}

// GetConceptTotals calculates and returns the totals for each payroll concept for a given period.
func (s *PayrollService) GetConceptTotals(periodID uuid.UUID) ([]dtos.PayrollConceptTotal, error) {
	// This requires querying payroll_lines or aggregating from payroll_calculations
	// For simplicity, this example will aggregate from payroll_calculations directly
	// A more robust solution would join with payroll_lines and payroll_concepts

	calculations, err := s.payrollRepo.FindByPeriod(periodID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve payroll calculations for concept totals: %w", err)
	}

	if len(calculations) == 0 {
		return nil, errors.New("no payroll calculations found for this period")
	}

	var totalGross, totalDeductions, totalNet, employerContributions float64
	employeeSummaries := make([]dtos.PayrollSummaryEmployee, len(calculations))

	for i, calc := range calculations {
		totalGross += calc.TotalGrossIncome
		totalDeductions += calc.TotalStatutoryDeductions + calc.TotalOtherDeductions
		totalNet += calc.TotalNetPay
		// Note: This is a simplified sum. A real implementation might need to be more careful.
		employerContributions += calc.IMSSEmployer + calc.InfonavitEmployer

		employeeSummaries[i] = dtos.PayrollSummaryEmployee{
			EmployeeID:   calc.EmployeeID.String(),
			EmployeeName: calc.Employee.FirstName + " " + calc.Employee.LastName,
			Gross:        calc.TotalGrossIncome,
			Deductions:   calc.TotalStatutoryDeductions + calc.TotalOtherDeductions,
			Net:          calc.TotalNetPay,
			Status:       calc.CalculationStatus,
		}
	}

	conceptTotals := make(map[string]float64)

	for _, calc := range calculations {
		conceptTotals["Regular Salary"] += calc.RegularSalary
		conceptTotals["Overtime"] += calc.OvertimeAmount + calc.DoubleOvertimeAmount + calc.TripleOvertimeAmount
		conceptTotals["Vacation Premium"] += calc.VacationPremium
		conceptTotals["Aguinaldo"] += calc.Aguinaldo
		conceptTotals["Other Income"] += calc.OtherExtras

		conceptTotals["ISR Withholding"] += calc.ISRWithholding
		conceptTotals["IMSS Employee"] += calc.IMSSEmployee
		conceptTotals["Infonavit Employee"] += calc.InfonavitEmployee
		conceptTotals["Retirement Savings"] += calc.RetirementSavings
		conceptTotals["Loan Deductions"] += calc.LoanDeductions
		conceptTotals["Advance Deductions"] += calc.AdvanceDeductions
		conceptTotals["Other Deductions"] += calc.OtherDeductions

		conceptTotals["Food Vouchers"] += calc.FoodVouchers
		conceptTotals["Savings Fund"] += calc.SavingsFund
	}

	var results []dtos.PayrollConceptTotal
	for concept, total := range conceptTotals {
		results = append(results, dtos.PayrollConceptTotal{
			Concept: concept,
			Total:   total,
		})
	}

	return results, nil
}


// GenerateDetailedReport is a placeholder for detailed report generation.
func (s *PayrollService) GenerateDetailedReport(
    calculations []models.PayrollCalculation,
) (*dtos.PayrollReportResponse, error) {
    // Implementation for detailed report
    return nil, nil
}

// GenerateStatutoryReport is a placeholder for statutory report generation.
func (s *PayrollService) GenerateStatutoryReport(
    calculations []models.PayrollCalculation,
) (*dtos.PayrollReportResponse, error) {
    // Implementation for statutory report
    return nil, nil
}

// GenerateAccountingReport is a placeholder for accounting report generation.
func (s *PayrollService) GenerateAccountingReport(
    calculations []models.PayrollCalculation,
) (*dtos.PayrollReportResponse, error) {
    // Implementation for accounting report
    return nil, nil
}

func (s *PayrollService) GeneratePDFPayslip(payroll *models.PayrollCalculation) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Colors
	headerR, headerG, headerB := 30, 58, 138 // Dark blue

	// ==================== COMPANY HEADER ====================
	pdf.SetFillColor(headerR, headerG, headerB)
	pdf.Rect(0, 0, 210, 40, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(10, 8)
	pdf.Cell(100, 8, "RECIBO DE NOMINA")
	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(10, 18)
	pdf.Cell(100, 5, "Empresa Ficticia S.A. de C.V.")
	pdf.SetXY(10, 23)
	pdf.Cell(100, 5, "RFC: EKU9003173C9")
	pdf.SetXY(10, 28)
	pdf.Cell(100, 5, "Reg. Patronal IMSS: A1234567890")
	pdf.SetXY(10, 33)
	pdf.Cell(100, 5, "Av. Industria 100, Col. Centro, CP 78000, San Luis Potosi, SLP")

	// Period info on right
	pdf.SetXY(130, 10)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(70, 6, fmt.Sprintf("Periodo: %s", payroll.PayrollPeriod.PeriodCode))
	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(130, 17)
	pdf.Cell(70, 5, fmt.Sprintf("Del: %s", payroll.PayrollPeriod.StartDate.Format("02/01/2006")))
	pdf.SetXY(130, 22)
	pdf.Cell(70, 5, fmt.Sprintf("Al: %s", payroll.PayrollPeriod.EndDate.Format("02/01/2006")))
	pdf.SetXY(130, 27)
	pdf.Cell(70, 5, fmt.Sprintf("Pago: %s", payroll.PayrollPeriod.PaymentDate.Format("02/01/2006")))
	// Calculate working days
	workDays := payroll.PayrollPeriod.GetWorkingDays()
	pdf.SetXY(130, 32)
	pdf.Cell(70, 5, fmt.Sprintf("Dias Pagados: %d", workDays))

	// Reset colors
	pdf.SetTextColor(0, 0, 0)

	// ==================== EMPLOYEE INFORMATION ====================
	pdf.SetXY(10, 45)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(70, 130, 180) // Steel blue
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(190, 7, "DATOS DEL TRABAJADOR", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Arial", "", 9)
	y := pdf.GetY() + 1

	// Row 1
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(28, 5, "No. Empleado:")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(35, 5, payroll.Employee.EmployeeNumber)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(15, 5, "RFC:")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(35, 5, payroll.Employee.RFC)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(15, 5, "NSS:")
	pdf.SetFont("Arial", "", 9)
	nss := payroll.Employee.NSS
	if nss == "" {
		nss = "N/A"
	}
	pdf.Cell(45, 5, nss)

	// Row 2
	y += 6
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(28, 5, "Nombre:")
	pdf.SetFont("Arial", "", 9)
	fullName := fmt.Sprintf("%s %s", payroll.Employee.FirstName, payroll.Employee.LastName)
	if payroll.Employee.MotherLastName != "" {
		fullName = fmt.Sprintf("%s %s %s", payroll.Employee.FirstName, payroll.Employee.LastName, payroll.Employee.MotherLastName)
	}
	pdf.Cell(70, 5, fullName)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(15, 5, "CURP:")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(60, 5, payroll.Employee.CURP)

	// Row 3
	y += 6
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(28, 5, "Tipo Contrato:")
	pdf.SetFont("Arial", "", 9)
	tipoContrato := "Indeterminado"
	if payroll.Employee.EmployeeType == "temporary" {
		tipoContrato = "Temporal"
	}
	pdf.Cell(35, 5, tipoContrato)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(28, 5, "Tipo Regimen:")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(35, 5, "Sueldos y Salarios")
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(20, 5, "Periodicidad:")
	pdf.SetFont("Arial", "", 9)
	periodicidad := "Quincenal"
	if payroll.Employee.PayFrequency == "weekly" {
		periodicidad = "Semanal"
	} else if payroll.Employee.PayFrequency == "monthly" {
		periodicidad = "Mensual"
	}
	pdf.Cell(30, 5, periodicidad)

	// Row 4
	y += 6
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(28, 5, "Salario Diario:")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(35, 5, fmt.Sprintf("$%.2f", payroll.Employee.DailySalary))
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(28, 5, "S.D. Integrado:")
	pdf.SetFont("Arial", "", 9)
	sdi := payroll.Employee.IntegratedDailySalary
	if sdi == 0 {
		sdi = payroll.Employee.DailySalary * 1.0493 // Approximate SDI factor
	}
	pdf.Cell(35, 5, fmt.Sprintf("$%.2f", sdi))
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(20, 5, "F. Ingreso:")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(30, 5, payroll.Employee.HireDate.Format("02/01/2006"))

	// ==================== PERCEPCIONES (EARNINGS) ====================
	y += 10
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(144, 238, 144) // Light green
	pdf.SetTextColor(0, 0, 0)
	// Header with SAT codes
	pdf.CellFormat(15, 6, "Clave", "1", 0, "C", true, 0, "")
	pdf.CellFormat(55, 6, "PERCEPCIONES", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, "Gravado", "1", 0, "R", true, 0, "")

	// Deducciones header
	pdf.SetFillColor(255, 182, 193) // Light pink
	pdf.CellFormat(15, 6, "Clave", "1", 0, "C", true, 0, "")
	pdf.CellFormat(55, 6, "DEDUCCIONES", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, "Importe", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	y = pdf.GetY()
	startY := y

	// Percepciones items with SAT codes
	pdf.SetXY(10, y)
	pdf.CellFormat(15, 5, "001", "LR", 0, "C", false, 0, "")
	pdf.CellFormat(55, 5, "Sueldos, Salarios y Jornales", "R", 0, "L", false, 0, "")
	pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.RegularSalary), "R", 0, "R", false, 0, "")

	if payroll.OvertimeAmount+payroll.DoubleOvertimeAmount+payroll.TripleOvertimeAmount > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "019", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Horas Extra", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.OvertimeAmount+payroll.DoubleOvertimeAmount+payroll.TripleOvertimeAmount), "R", 0, "R", false, 0, "")
	}

	if payroll.VacationPremium > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "021", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Prima Vacacional", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.VacationPremium), "R", 0, "R", false, 0, "")
	}

	if payroll.Aguinaldo > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "002", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Aguinaldo", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.Aguinaldo), "R", 0, "R", false, 0, "")
	}

	if payroll.BonusAmount > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "028", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Comisiones/Bonos", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.BonusAmount+payroll.CommissionAmount), "R", 0, "R", false, 0, "")
	}

	if payroll.FoodVouchers > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "029", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Vales de Despensa", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.FoodVouchers), "R", 0, "R", false, 0, "")
	}

	if payroll.SavingsFund > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "005", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Fondo de Ahorro", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.SavingsFund), "R", 0, "R", false, 0, "")
	}

	if payroll.OtherExtras > 0 {
		y += 5
		pdf.SetXY(10, y)
		pdf.CellFormat(15, 5, "038", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Otros Ingresos", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.OtherExtras), "R", 0, "R", false, 0, "")
	}

	// Deducciones items (right column) with SAT codes
	y = startY
	pdf.SetXY(105, y)
	pdf.CellFormat(15, 5, "002", "LR", 0, "C", false, 0, "")
	pdf.CellFormat(55, 5, "ISR (Impuesto Sobre la Renta)", "R", 0, "L", false, 0, "")
	pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.ISRWithholding), "R", 1, "R", false, 0, "")

	y += 5
	pdf.SetXY(105, y)
	pdf.CellFormat(15, 5, "001", "LR", 0, "C", false, 0, "")
	pdf.CellFormat(55, 5, "Seguridad Social (IMSS)", "R", 0, "L", false, 0, "")
	pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.IMSSEmployee), "R", 1, "R", false, 0, "")

	if payroll.InfonavitEmployee > 0 {
		y += 5
		pdf.SetXY(105, y)
		pdf.CellFormat(15, 5, "010", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "INFONAVIT", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.InfonavitEmployee), "R", 1, "R", false, 0, "")
	}

	if payroll.RetirementSavings > 0 {
		y += 5
		pdf.SetXY(105, y)
		pdf.CellFormat(15, 5, "017", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Aportacion Voluntaria SAR", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.RetirementSavings), "R", 1, "R", false, 0, "")
	}

	if payroll.LoanDeductions > 0 {
		y += 5
		pdf.SetXY(105, y)
		pdf.CellFormat(15, 5, "004", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Prestamos", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.LoanDeductions), "R", 1, "R", false, 0, "")
	}

	if payroll.AdvanceDeductions > 0 {
		y += 5
		pdf.SetXY(105, y)
		pdf.CellFormat(15, 5, "012", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Anticipos de Salario", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.AdvanceDeductions), "R", 1, "R", false, 0, "")
	}

	if payroll.OtherDeductions > 0 {
		y += 5
		pdf.SetXY(105, y)
		pdf.CellFormat(15, 5, "004", "LR", 0, "C", false, 0, "")
		pdf.CellFormat(55, 5, "Otras Deducciones", "R", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("$%.2f", payroll.OtherDeductions), "R", 1, "R", false, 0, "")
	}

	// ==================== TOTALS SECTION ====================
	y += 12
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 230, 200) // Light green
	pdf.CellFormat(70, 7, "TOTAL PERCEPCIONES", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, fmt.Sprintf("$%.2f", payroll.TotalGrossIncome), "1", 0, "R", true, 0, "")

	pdf.SetFillColor(255, 200, 200) // Light red
	pdf.CellFormat(70, 7, "TOTAL DEDUCCIONES", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, fmt.Sprintf("$%.2f", payroll.TotalStatutoryDeductions+payroll.TotalOtherDeductions), "1", 1, "R", true, 0, "")

	// Net Pay - prominent display
	y += 12
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(headerR, headerG, headerB)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(140, 10, "NETO A PAGAR", "1", 0, "R", true, 0, "")
	pdf.SetFillColor(34, 139, 34) // Forest green
	pdf.CellFormat(50, 10, fmt.Sprintf("$%.2f MXN", payroll.TotalNetPay), "1", 1, "C", true, 0, "")

	// ==================== EMPLOYER CONTRIBUTIONS (INFORMATIVE) ====================
	pdf.SetTextColor(0, 0, 0)
	y += 15
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(230, 230, 250) // Lavender
	pdf.CellFormat(190, 6, "APORTACIONES PATRONALES (Informativo - No afectan el neto a pagar)", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	y = pdf.GetY()
	pdf.SetXY(10, y)
	pdf.Cell(35, 5, "Cuota IMSS Patron:")
	pdf.Cell(25, 5, fmt.Sprintf("$%.2f", payroll.IMSSEmployer))
	pdf.Cell(35, 5, "Aport. INFONAVIT:")
	pdf.Cell(25, 5, fmt.Sprintf("$%.2f", payroll.InfonavitEmployer))

	// SAR/AFORE and totals
	sarEmployer := 0.0
	imssRetirement := 0.0
	if payroll.EmployerContribution != nil {
		sarEmployer = payroll.EmployerContribution.RetirementSAR
		imssRetirement = payroll.EmployerContribution.IMSSRetirement
	}
	pdf.Cell(35, 5, "SAR/Retiro:")
	pdf.Cell(25, 5, fmt.Sprintf("$%.2f", sarEmployer+imssRetirement))

	totalPatron := payroll.IMSSEmployer + payroll.InfonavitEmployer + sarEmployer + imssRetirement
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(25, 5, fmt.Sprintf("Total: $%.2f", totalPatron))

	// ==================== SUBSIDIO AL EMPLEO (if applicable) ====================
	if payroll.EmploymentSubsidy > 0 {
		y = pdf.GetY() + 8
		pdf.SetXY(10, y)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(255, 255, 200) // Light yellow
		pdf.CellFormat(95, 6, "SUBSIDIO PARA EL EMPLEO (Art. Dec. Trans.)", "1", 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(95, 6, fmt.Sprintf("$%.2f", payroll.EmploymentSubsidy), "1", 1, "R", true, 0, "")
	}

	// ==================== LEGAL NOTICE ====================
	y = 255
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "I", 7)
	pdf.SetTextColor(100, 100, 100)
	pdf.MultiCell(190, 3, "Este documento es un comprobante de pago emitido conforme a lo dispuesto en la Ley Federal del Trabajo Art. 804. "+
		"Para efectos fiscales, solicite su CFDI (Comprobante Fiscal Digital por Internet) con complemento de nomina.", "", "L", false)

	// Footer with generation info
	pdf.SetXY(10, 275)
	pdf.SetFont("Arial", "", 7)
	pdf.Cell(95, 4, fmt.Sprintf("Generado: %s", time.Now().Format("02/01/2006 15:04:05")))
	pdf.Cell(95, 4, fmt.Sprintf("UUID CFDI: %s (Pendiente Timbrado)", "N/A"))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *PayrollService) GenerateHTMLPayslip(payroll *models.PayrollCalculation) ([]byte, error) {
	htmlContent := fmt.Sprintf("<h1>Payslip for %s %s</h1><p>Period: %s</p><p>Net Pay: %.2f</p>",
		payroll.Employee.FirstName, payroll.Employee.LastName,
		payroll.PayrollPeriod.PeriodCode, payroll.TotalNetPay)
	return []byte(htmlContent), nil
}
