/*
Package services - Pre-Payroll (Prenomina) Service

==============================================================================
FILE: internal/services/prenomina_service.go
==============================================================================

DESCRIPTION:
    Calculates pre-payroll metrics including worked days, overtime, absences,
    bonuses, and deductions before final payroll processing. This is the first
    step in the Mexican payroll calculation workflow.

USER PERSPECTIVE:
    - Review attendance metrics before final payroll
    - Process incidences (overtime, absences, bonuses) for the period
    - Approve or reject prenomina before payroll calculation
    - Bulk calculate prenomina for all employees in a period

DEVELOPER GUIDELINES:
    OK to modify: Incidence processing rules, metric calculations
    CAUTION: SDI (Integrated Daily Salary) calculation affects IMSS contributions
    DO NOT modify: Prenomina approval workflow without updating payroll service
    Note: Prenomina must be approved before final payroll calculation

SYNTAX EXPLANATION:
    - CalculatePrenomina processes attendance and incidences for one employee
    - processIncidences converts HR incidences into payroll metrics
    - calculateDefaultMetrics sets base worked days and hours
    - WorkedDays = Period days - Absences - Sick days - Vacation days

==============================================================================
*/
package services

import (
    "errors"
    "fmt"
    "time"

    "github.com/google/uuid"
        "gorm.io/gorm"

        	"backend/internal/config" // Import for config.AppConfig
        		config_payroll "backend/internal/config/payroll"
        		"backend/internal/dtos"
        		"backend/internal/models"
        		"backend/internal/repositories"
        	)
        	
        	// PrenominaService handles prenomina calculation business logic
        	type PrenominaService struct {
        	    prenominaRepo  *repositories.PrenominaRepository
        	    employeeRepo   *repositories.EmployeeRepository
        	    periodRepo     *repositories.PayrollPeriodRepository
        	    incidenceRepo  *repositories.IncidenceRepository
        	    config         *config_payroll.PayrollConfig
        		db             *gorm.DB
        	}
// NewPrenominaService creates a new prenomina service
func NewPrenominaService(
    db *gorm.DB,
    appConfig *config.AppConfig,
) *PrenominaService {
    return &PrenominaService{
        prenominaRepo: repositories.NewPrenominaRepository(db),
        employeeRepo:  repositories.NewEmployeeRepository(db),
        periodRepo:    repositories.NewPayrollPeriodRepository(db),
        incidenceRepo: repositories.NewIncidenceRepository(db),
        config:        appConfig.PayrollConfig,
        db:            db,
    }
}

// CalculatePrenomina calculates prenomina metrics for an employee
func (s *PrenominaService) CalculatePrenomina(
    employeeID, periodID uuid.UUID,
    calculateSDI bool,
    calculatedBy uuid.UUID,
) (*dtos.PrenominaMetricResponse, error) {
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
    
    // Get existing prenomina metrics or create new
    metrics, err := s.prenominaRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("error fetching prenomina metrics: %w", err)
    }
    
    var prenominaMetric *models.PrenominaMetric
    if metrics == nil {
        // Create new metrics
        prenominaMetric = &models.PrenominaMetric{
            EmployeeID:      employeeID,
            PayrollPeriodID: periodID,
            CalculationStatus: "calculated",
        }
    } else {
        prenominaMetric = metrics
        prenominaMetric.CalculationStatus = "calculated"
    }
    
    // Calculate SDI if requested
    if calculateSDI {
        employee.IntegratedDailySalary = employee.CalculateIntegratedDailySalary()
        if err := s.employeeRepo.Update(employee); err != nil {
            return nil, fmt.Errorf("error updating SDI: %w", err)
        }
    }
    
    // Get incidences for the period
    incidences, err := s.incidenceRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("error fetching incidences: %w", err)
    }
    
    // Process incidences to populate metrics
    s.processIncidences(prenominaMetric, incidences, employee.DailySalary)
    
    // Calculate default metrics if not set by incidences
    s.calculateDefaultMetrics(prenominaMetric, period, employee)
    
    // Calculate amounts
    s.calculateAmounts(prenominaMetric, employee)
    
    // Calculate summary
    s.calculateSummary(prenominaMetric)
    
    // Save metrics
    if metrics == nil {
        if err := s.prenominaRepo.Create(prenominaMetric); err != nil {
            return nil, fmt.Errorf("error saving prenomina metrics: %w", err)
        }
    } else {
        if err := s.prenominaRepo.Update(prenominaMetric); err != nil {
            return nil, fmt.Errorf("error updating prenomina metrics: %w", err)
        }
    }
    
    // Convert to response
    return s.convertToResponse(prenominaMetric, employee, period), nil
}

// processIncidences processes incidences and populates metrics
func (s *PrenominaService) processIncidences(
    metrics *models.PrenominaMetric,
    incidences []models.Incidence,
    dailySalary float64,
) {
    for _, incidence := range incidences {
        if incidence.Status != "approved" && incidence.Status != "processed" {
            continue
        }
        
        switch incidence.IncidenceType.Category {
        case "absence":
            if incidence.IncidenceType.EffectType == "negative" {
                metrics.AbsenceDays += incidence.Quantity
            }
        case "sick":
            metrics.SickDays += incidence.Quantity
        case "vacation":
            metrics.VacationDays += incidence.Quantity
        case "overtime":
            switch incidence.IncidenceType.CalculationMethod {
            case "hourly":
                metrics.OvertimeHours += incidence.Quantity
            case "hourly_double":
                metrics.DoubleOvertimeHours += incidence.Quantity
            case "hourly_triple":
                metrics.TripleOvertimeHours += incidence.Quantity
            }
        case "delay":
            metrics.DelaysCount++
            metrics.DelayMinutes += incidence.Quantity
        case "bonus":
            metrics.BonusAmount += incidence.CalculatedAmount
        case "deduction":
            metrics.OtherDeduction += incidence.CalculatedAmount
        }
    }
}

// calculateDefaultMetrics calculates default work metrics
func (s *PrenominaService) calculateDefaultMetrics(
    metrics *models.PrenominaMetric,
    period *models.PayrollPeriod,
    employee *models.Employee,
) {
    // Calculate worked days based on period
    periodDays := period.CalculateDays()
    
    // Default: assume employee worked all days minus absences
    if metrics.WorkedDays == 0 {
        metrics.WorkedDays = float64(periodDays) - 
            metrics.AbsenceDays - 
            metrics.SickDays - 
            metrics.VacationDays - 
            metrics.UnpaidLeaveDays
    }
    
    // Calculate regular hours (8 hours per worked day)
    if metrics.RegularHours == 0 {
        metrics.RegularHours = metrics.WorkedDays * 8
    }
    
    // Calculate Sunday premium if applicable
    // (In Mexico, working on Sunday typically adds 25% premium)
    // This would require knowing which days were Sundays
}

// calculateAmounts calculates monetary amounts
func (s *PrenominaService) calculateAmounts(
    metrics *models.PrenominaMetric,
    employee *models.Employee,
) {
    // Calculate regular salary
    hourlyRate := employee.DailySalary / 8
    
    // Regular salary for worked hours
    metrics.RegularSalary = metrics.RegularHours * hourlyRate
    
    // Overtime amounts
    metrics.OvertimeAmount = metrics.OvertimeHours * hourlyRate * 
        (1 + s.config.LaborConcepts.Overtime.DoubleTimePercentage)
    
    metrics.DoubleOvertimeAmount = metrics.DoubleOvertimeHours * hourlyRate * 
        (1 + s.config.LaborConcepts.Overtime.TripleTimePercentage)
    
    metrics.TripleOvertimeAmount = metrics.TripleOvertimeHours * hourlyRate * 3
    
    // Delay deduction (proportional to minutes)
    if metrics.DelayMinutes > 0 {
        minuteRate := hourlyRate / 60
        metrics.DelayDeduction = metrics.DelayMinutes * minuteRate
    }
}

// calculateSummary calculates summary totals
func (s *PrenominaService) calculateSummary(metrics *models.PrenominaMetric) {
    // Total extras
    metrics.TotalExtras = metrics.BonusAmount + 
        metrics.CommissionAmount + 
        metrics.OtherExtraAmount +
        metrics.OvertimeAmount +
        metrics.DoubleOvertimeAmount +
        metrics.TripleOvertimeAmount
    
    // Total deductions
    metrics.TotalDeductions = metrics.LoanDeduction +
        metrics.AdvanceDeduction +
        metrics.OtherDeduction +
        metrics.DelayDeduction
    
    // Gross income (before deductions)
    metrics.GrossIncome = metrics.RegularSalary + metrics.TotalExtras
    
    // Net income (after pre-payroll deductions)
    metrics.NetIncome = metrics.GrossIncome - metrics.TotalDeductions
}

// BulkCalculatePrenomina calculates prenomina for multiple employees
func (s *PrenominaService) BulkCalculatePrenomina(
    req dtos.PrenominaBulkCalculateRequest,
    calculatedBy uuid.UUID,
) (*dtos.PrenominaBulkCalculateResponse, error) {
    // Get payroll period
    period, err := s.periodRepo.FindByID(req.PayrollPeriodID)
    if err != nil {
        return nil, fmt.Errorf("payroll period not found: %w", err)
    }
    
    // Get employees to calculate
    var employees []models.Employee
    if req.CalculateAll {
        // Get all active employees
        filters := map[string]interface{}{
            "status": "active",
        }
        employees, _, err = s.employeeRepo.List(1, 10000, filters) // Large limit
        if err != nil {
            return nil, fmt.Errorf("error fetching active employees: %w", err)
        }
    } else {
        // Get specific employees
        for _, employeeID := range req.EmployeeIDs {
            employee, err := s.employeeRepo.FindByID(employeeID)
            if err != nil {
                continue // Skip not found employees
            }
            employees = append(employees, *employee)
        }
    }
    
    // Calculate for each employee
    results := make([]dtos.PrenominaCalculationResult, 0, len(employees))
    totalSuccess := 0
    totalGross := 0.0
    totalNet := 0.0
    
    for _, employee := range employees {
        result := dtos.PrenominaCalculationResult{
            EmployeeID:   employee.ID,
            EmployeeName: fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
        }
        
        // Calculate prenomina
        response, err := s.CalculatePrenomina(
            employee.ID,
            req.PayrollPeriodID,
            false, // Don't recalculate SDI in bulk
            calculatedBy,
        )
        
        if err != nil {
            result.Success = false
            result.Error = err.Error()
        } else {
            result.Success = true
            result.GrossIncome = response.GrossIncome
            result.NetIncome = response.NetIncome
            
            totalSuccess++
            totalGross += response.GrossIncome
            totalNet += response.NetIncome
        }
        
        results = append(results, result)
    }
    
    return &dtos.PrenominaBulkCalculateResponse{
        PeriodCode:      period.PeriodCode,
        TotalCalculated: len(employees),
        TotalSuccess:    totalSuccess,
        TotalFailed:     len(employees) - totalSuccess,
        Results:         results,
        TotalGross:      totalGross,
        TotalNet:        totalNet,
    }, nil
}

// GetPrenominaMetrics gets prenomina metrics
func (s *PrenominaService) GetPrenominaMetrics(
    employeeID, periodID uuid.UUID,
) (*dtos.PrenominaMetricResponse, error) {
    metrics, err := s.prenominaRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("prenomina metrics not found")
        }
        return nil, fmt.Errorf("error fetching prenomina metrics: %w", err)
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
    
    return s.convertToResponse(metrics, employee, period), nil
}

// ListPrenominaMetrics lists prenomina metrics for a period
func (s *PrenominaService) ListPrenominaMetrics(
    periodID uuid.UUID,
    page, pageSize int,
) (*dtos.PrenominaListResponse, error) {
    metrics, total, err := s.prenominaRepo.FindByPeriod(periodID, page, pageSize)
    if err != nil {
        return nil, fmt.Errorf("error listing prenomina metrics: %w", err)
    }
    
    // Convert to responses
    responses := make([]dtos.PrenominaMetricResponse, len(metrics))
    for i, metric := range metrics {
        employee, _ := s.employeeRepo.FindByID(metric.EmployeeID)
        period, _ := s.periodRepo.FindByID(metric.PayrollPeriodID)
        
        responses[i] = *s.convertToResponse(&metric, employee, period)
    }
    
    totalPages := 1
    if pageSize > 0 {
        totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
    }
    
    return &dtos.PrenominaListResponse{
        Metrics:    responses,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

// ApprovePrenomina approves prenomina metrics
func (s *PrenominaService) ApprovePrenomina(
    employeeID, periodID uuid.UUID,
    approvedBy uuid.UUID,
) error {
    metrics, err := s.prenominaRepo.FindByEmployeeAndPeriod(employeeID, periodID)
    if err != nil {
        return fmt.Errorf("prenomina metrics not found: %w", err)
    }
    
    if metrics.CalculationStatus == "approved" {
        return errors.New("prenomina already approved")
    }
    
    metrics.CalculationStatus = "approved"
    return s.prenominaRepo.Update(metrics)
}

// convertToResponse converts PrenominaMetric to response DTO
func (s *PrenominaService) convertToResponse(
    metrics *models.PrenominaMetric,
    employee *models.Employee,
    period *models.PayrollPeriod,
) *dtos.PrenominaMetricResponse {
    var calculatedAt time.Time
    if metrics.CalculationDate != nil {
        calculatedAt = *metrics.CalculationDate
    }
    return &dtos.PrenominaMetricResponse{
        ID:                   metrics.ID,
        EmployeeID:           metrics.EmployeeID,
        EmployeeName:         fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
        EmployeeNumber:       employee.EmployeeNumber,
        PayrollPeriodID:      metrics.PayrollPeriodID,
        PeriodCode:           period.PeriodCode,
        
        WorkedDays:           metrics.WorkedDays,
        RegularHours:         metrics.RegularHours,
        OvertimeHours:        metrics.OvertimeHours,
        DoubleOvertimeHours:  metrics.DoubleOvertimeHours,
        TripleOvertimeHours:  metrics.TripleOvertimeHours,
        
        AbsenceDays:          metrics.AbsenceDays,
        SickDays:             metrics.SickDays,
        VacationDays:         metrics.VacationDays,
        UnpaidLeaveDays:      metrics.UnpaidLeaveDays,
        
        DelaysCount:          metrics.DelaysCount,
        DelayMinutes:         metrics.DelayMinutes,
        DelayDeduction:       metrics.DelayDeduction,
        EarlyDeparturesCount: metrics.EarlyDeparturesCount,
        
        BonusAmount:          metrics.BonusAmount,
        CommissionAmount:     metrics.CommissionAmount,
        OtherExtraAmount:     metrics.OtherExtraAmount,
        TotalExtras:          metrics.TotalExtras,
        
        LoanDeduction:        metrics.LoanDeduction,
        AdvanceDeduction:     metrics.AdvanceDeduction,
        OtherDeduction:       metrics.OtherDeduction,
        TotalDeductions:      metrics.TotalDeductions,
        
        GrossIncome:          metrics.GrossIncome,
        NetIncome:            metrics.NetIncome,
        
        CalculationStatus:    metrics.CalculationStatus,
        CalculatedAt:         calculatedAt,
        CalculatedBy:         "", // Would need to fetch user name
        CreatedAt:            metrics.CreatedAt,
        UpdatedAt:            metrics.UpdatedAt,
    }
}
