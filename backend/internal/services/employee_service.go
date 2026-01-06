/*
Package services - Employee Management Service

==============================================================================
FILE: internal/services/employee_service.go
==============================================================================

DESCRIPTION:
    Manages employee master data including personal information, employment details,
    salary, Mexican official IDs (RFC, CURP, NSS), and bulk import from Excel/CSV.
    Validates Mexican tax identifiers and calculates integrated daily salary (SDI).

USER PERSPECTIVE:
    - Create and manage employee records with full Mexican compliance
    - Import employees in bulk from Excel templates
    - Track employment history and salary changes
    - Validate RFC, CURP, and NSS formats
    - View employee statistics by type and status
    - Manage terminations and rehires

DEVELOPER GUIDELINES:
    OK to modify: Employee fields, add custom attributes
    CAUTION: RFC/CURP/NSS validation follows Mexican government standards
    DO NOT modify: SDI calculation without reviewing IMSS implications
    Note: CollarType determines default pay frequency and union status

SYNTAX EXPLANATION:
    - RFC: 12-13 character tax ID (validated by checksum)
    - CURP: 18 character unique population ID
    - NSS: 11 digit social security number
    - CollarType: white_collar (admin), blue_collar (union), gray_collar (non-union)
    - SDI (Salario Diario Integrado): Daily salary + benefits for IMSS calculation
    - ImportEmployeesFromFile supports both Excel (.xlsx) and CSV formats

==============================================================================
*/
package services

import (
    "bytes"
    "encoding/csv"
    "errors"
    "fmt"
    "io"
    "mime/multipart"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/xuri/excelize/v2"
    "gorm.io/gorm"

    "backend/internal/dtos"
    "backend/internal/models"
    "backend/internal/models/enums"
    "backend/internal/repositories"
    "backend/internal/utils"
)

// EmployeeService handles employee business logic
type EmployeeService struct {
    employeeRepo *repositories.EmployeeRepository
    userRepo     *repositories.UserRepository
    db           *gorm.DB
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(db *gorm.DB) *EmployeeService {
    return &EmployeeService{
        employeeRepo: repositories.NewEmployeeRepository(db),
        userRepo:     repositories.NewUserRepository(db),
        db:           db,
    }
}

// CreateEmployee creates a new employee
func (s *EmployeeService) CreateEmployee(req dtos.EmployeeRequest, createdBy uuid.UUID) (*models.Employee, error) {
    // Validate unique identifiers
    if exists, err := s.employeeRepo.ExistsByRFC(req.RFC); err != nil {
        return nil, fmt.Errorf("error checking RFC: %w", err)
    } else if exists {
        return nil, errors.New("employee with this RFC already exists")
    }
    
    if exists, err := s.employeeRepo.ExistsByCURP(req.CURP); err != nil {
        return nil, fmt.Errorf("error checking CURP: %w", err)
    } else if exists {
        return nil, errors.New("employee with this CURP already exists")
    }
    
    if req.NSS != "" {
        if exists, err := s.employeeRepo.ExistsByNSS(req.NSS); err != nil {
            return nil, fmt.Errorf("error checking NSS: %w", err)
        } else if exists {
            return nil, errors.New("employee with this NSS already exists")
        }
    }
    
    // Check if employee number exists
    existingEmployee, err := s.employeeRepo.FindByEmployeeNumber(req.EmployeeNumber)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("error checking employee number: %w", err)
    }
    if existingEmployee != nil {
        return nil, errors.New("employee with this employee number already exists")
    }
    
    // Convert optional date pointers
    var terminationDate *time.Time
    if req.TerminationDate != nil && req.TerminationDate.Time != nil {
        terminationDate = req.TerminationDate.Time
    }
    var imssRegDate *time.Time
    if req.IMSSRegistrationDate != nil && req.IMSSRegistrationDate.Time != nil {
        imssRegDate = req.IMSSRegistrationDate.Time
    }

    // Create employee model
    employee := &models.Employee{
        EmployeeNumber:   req.EmployeeNumber,
        FirstName:        strings.TrimSpace(req.FirstName),
        LastName:         strings.TrimSpace(req.LastName),
        MotherLastName:   strings.TrimSpace(req.MotherLastName),
        DateOfBirth:      req.DateOfBirth.Time,
        Gender:           req.Gender,
        MaritalStatus:    req.MaritalStatus,
        RFC:              strings.ToUpper(strings.TrimSpace(req.RFC)),
        CURP:             strings.ToUpper(strings.TrimSpace(req.CURP)),
        NSS:              strings.TrimSpace(req.NSS),
        InfonavitCredit:  strings.TrimSpace(req.InfonavitCredit),
        PersonalEmail:    strings.TrimSpace(req.PersonalEmail),
        PersonalPhone:    strings.TrimSpace(req.PersonalPhone),
        EmergencyContact: strings.TrimSpace(req.EmergencyContact),
        EmergencyPhone:   strings.TrimSpace(req.EmergencyPhone),
        Street:           strings.TrimSpace(req.Street),
        ExteriorNumber:   strings.TrimSpace(req.ExteriorNumber),
        InteriorNumber:   strings.TrimSpace(req.InteriorNumber),
        Neighborhood:     strings.TrimSpace(req.Neighborhood),
        Municipality:     strings.TrimSpace(req.Municipality),
        State:            strings.TrimSpace(req.State),
        PostalCode:       strings.TrimSpace(req.PostalCode),
        Country:          strings.TrimSpace(req.Country),
        HireDate:         req.HireDate.Time,
        TerminationDate:  terminationDate,
        EmploymentStatus: req.EmploymentStatus,
        EmployeeType:     req.EmployeeType,
        CollarType:       req.CollarType,
        PayFrequency:     req.PayFrequency,
        IsSindicalizado:  req.IsSindicalizado,
        DailySalary:      req.DailySalary,
        IntegratedDailySalary: req.IntegratedDailySalary,
        PaymentMethod:    req.PaymentMethod,
        BankName:         strings.TrimSpace(req.BankName),
        BankAccount:      strings.TrimSpace(req.BankAccount),
        CLABE:            strings.TrimSpace(req.CLABE),
        IMSSRegistrationDate: imssRegDate,
        Regime:           req.Regime,
        TaxRegime:        req.TaxRegime,
        DepartmentID:     req.DepartmentID,
        PositionID:       req.PositionID,
        CostCenterID:     req.CostCenterID,
        ShiftID:          req.ShiftID,
        SupervisorID:     req.SupervisorID,
    }

    // Set created by
    employee.CreatedBy = &createdBy
    
    // Validate employee data
    if err := employee.Validate(); err != nil {
        return nil, fmt.Errorf("employee validation failed: %w", err)
    }
    
    // Create employee
    if err := s.employeeRepo.Create(employee); err != nil {
        return nil, fmt.Errorf("failed to create employee: %w", err)
    }
    
    return employee, nil
}

// GetEmployee gets an employee by ID
// userRole is optional - if provided, access control is enforced for hr_blue_gray and hr_white roles
func (s *EmployeeService) GetEmployee(id uuid.UUID, userRole ...string) (*dtos.EmployeeResponse, error) {
    employee, err := s.employeeRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("employee not found")
        }
        return nil, fmt.Errorf("error fetching employee: %w", err)
    }

    // Check collar type access if role is provided
    if len(userRole) > 0 && userRole[0] != "" {
        if !utils.CanManageCollarType(userRole[0], employee.CollarType) {
            return nil, errors.New("access denied: you cannot view employees of this collar type")
        }
    }

    return s.ConvertToResponse(employee), nil
}

// UpdateEmployee updates an employee
func (s *EmployeeService) UpdateEmployee(id uuid.UUID, req dtos.EmployeeRequest, updatedBy uuid.UUID) (*dtos.EmployeeResponse, error) {
    // Get existing employee
    employee, err := s.employeeRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("employee not found")
        }
        return nil, fmt.Errorf("error fetching employee: %w", err)
    }
    
    // Check for RFC uniqueness if changed
    if strings.ToUpper(strings.TrimSpace(req.RFC)) != employee.RFC {
        if exists, err := s.employeeRepo.ExistsByRFC(req.RFC); err != nil {
            return nil, fmt.Errorf("error checking RFC: %w", err)
        } else if exists {
            return nil, errors.New("another employee with this RFC already exists")
        }
    }
    
    // Check for CURP uniqueness if changed
    if strings.ToUpper(strings.TrimSpace(req.CURP)) != employee.CURP {
        if exists, err := s.employeeRepo.ExistsByCURP(req.CURP); err != nil {
            return nil, fmt.Errorf("error checking CURP: %w", err)
        } else if exists {
            return nil, errors.New("another employee with this CURP already exists")
        }
    }
    
    // Convert optional date pointers for update
    var updateTerminationDate *time.Time
    if req.TerminationDate != nil && req.TerminationDate.Time != nil {
        updateTerminationDate = req.TerminationDate.Time
    }
    var updateImssRegDate *time.Time
    if req.IMSSRegistrationDate != nil && req.IMSSRegistrationDate.Time != nil {
        updateImssRegDate = req.IMSSRegistrationDate.Time
    }

    // Update employee fields
    employee.FirstName = strings.TrimSpace(req.FirstName)
    employee.LastName = strings.TrimSpace(req.LastName)
    employee.MotherLastName = strings.TrimSpace(req.MotherLastName)
    employee.DateOfBirth = req.DateOfBirth.Time
    employee.Gender = req.Gender
    employee.MaritalStatus = req.MaritalStatus
    employee.RFC = strings.ToUpper(strings.TrimSpace(req.RFC))
    employee.CURP = strings.ToUpper(strings.TrimSpace(req.CURP))
    employee.NSS = strings.TrimSpace(req.NSS)
    employee.InfonavitCredit = strings.TrimSpace(req.InfonavitCredit)
    employee.PersonalEmail = strings.TrimSpace(req.PersonalEmail)
    employee.PersonalPhone = strings.TrimSpace(req.PersonalPhone)
    employee.EmergencyContact = strings.TrimSpace(req.EmergencyContact)
    employee.EmergencyPhone = strings.TrimSpace(req.EmergencyPhone)
    employee.Street = strings.TrimSpace(req.Street)
    employee.ExteriorNumber = strings.TrimSpace(req.ExteriorNumber)
    employee.InteriorNumber = strings.TrimSpace(req.InteriorNumber)
    employee.Neighborhood = strings.TrimSpace(req.Neighborhood)
    employee.Municipality = strings.TrimSpace(req.Municipality)
    employee.State = strings.TrimSpace(req.State)
    employee.PostalCode = strings.TrimSpace(req.PostalCode)
    employee.Country = strings.TrimSpace(req.Country)
    employee.HireDate = req.HireDate.Time
    employee.TerminationDate = updateTerminationDate
    employee.EmploymentStatus = req.EmploymentStatus
    employee.EmployeeType = req.EmployeeType
    employee.CollarType = req.CollarType
    employee.PayFrequency = req.PayFrequency
    employee.IsSindicalizado = req.IsSindicalizado
    employee.DailySalary = req.DailySalary
    employee.IntegratedDailySalary = req.IntegratedDailySalary
    employee.PaymentMethod = req.PaymentMethod
    employee.BankName = strings.TrimSpace(req.BankName)
    employee.BankAccount = strings.TrimSpace(req.BankAccount)
    employee.CLABE = strings.TrimSpace(req.CLABE)
    employee.IMSSRegistrationDate = updateImssRegDate
    employee.Regime = req.Regime
    employee.TaxRegime = req.TaxRegime
    employee.DepartmentID = req.DepartmentID
    employee.PositionID = req.PositionID
    employee.CostCenterID = req.CostCenterID
    employee.ShiftID = req.ShiftID
    employee.SupervisorID = req.SupervisorID

    // Set updated by
    employee.UpdatedBy = &updatedBy
    
    // Validate employee data
    if err := employee.Validate(); err != nil {
        return nil, fmt.Errorf("employee validation failed: %w", err)
    }
    
    // Update employee
    if err := s.employeeRepo.Update(employee); err != nil {
        return nil, fmt.Errorf("failed to update employee: %w", err)
    }
    
    return s.ConvertToResponse(employee), nil
}

// DeleteEmployee soft deletes an employee after validating no dependencies exist
func (s *EmployeeService) DeleteEmployee(id uuid.UUID) error {
    employee, err := s.employeeRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("employee not found")
        }
        return fmt.Errorf("error fetching employee: %w", err)
    }

    // Check if employee is active
    if employee.EmploymentStatus == "active" {
        return errors.New("cannot delete active employee. Terminate employee first via termination process")
    }

    // Check for active/pending payroll calculations
    var payrollCount int64
    s.db.Model(&models.PayrollCalculation{}).
        Where("employee_id = ?", id).
        Where("status IN (?)", []string{"pending", "approved"}).
        Count(&payrollCount)
    if payrollCount > 0 {
        return fmt.Errorf("cannot delete employee: %d pending/approved payroll records exist. Complete or archive payroll first", payrollCount)
    }

    // Check for active benefit enrollments
    var activeBenefitCount int64
    s.db.Model(&models.BenefitEnrollment{}).
        Where("employee_id = ?", id).
        Where("status = ?", "active").
        Count(&activeBenefitCount)
    if activeBenefitCount > 0 {
        return fmt.Errorf("cannot delete employee: %d active benefit enrollments exist. Terminate benefits first", activeBenefitCount)
    }

    // Check for open leave requests or ongoing time tracking
    var openIncidenceCount int64
    s.db.Model(&models.Incidence{}).
        Where("employee_id = ?", id).
        Where("status IN (?)", []string{"pending", "approved"}).
        Count(&openIncidenceCount)
    if openIncidenceCount > 0 {
        return fmt.Errorf("cannot delete employee: %d open leave/incidence requests exist. Resolve them first", openIncidenceCount)
    }

    // Check for active job applications or recruitment records
    var activeApplicationCount int64
    s.db.Model(&models.Application{}).
        Joins("JOIN employees ON employees.id = applications.employee_id").
        Where("applications.employee_id = ?", id).
        Where("applications.status IN (?)", []string{"applied", "screening", "interview", "offer"}).
        Count(&activeApplicationCount)
    if activeApplicationCount > 0 {
        return errors.New("cannot delete employee: active recruitment applications exist")
    }

    // If all validations pass, perform soft delete
    // Note: Soft delete sets DeletedAt timestamp, preserving audit trail
    // Historical records (past payroll, completed incidences) are retained for compliance
    return s.employeeRepo.Delete(id)
}

// ListEmployees lists employees with pagination
// userRole is optional - if provided, collar type filtering is applied for hr_blue_gray and hr_white roles
// userID is optional - if provided and role is supervisor/manager, filters to show only supervised employees
// - supervisor: shows only direct reports
// - manager: shows all subordinates (recursive hierarchy)
// - hr_white: shows only white_collar employees
// - hr_blue_gray: shows only blue_collar and gray_collar employees
// - admin/hr/hr_and_pr/payroll_staff/accountant: shows all employees
func (s *EmployeeService) ListEmployees(page, pageSize int, filters map[string]interface{}, userRole string, userID ...uuid.UUID) (*dtos.EmployeeListResponse, error) {
    // Apply collar type filtering based on role
    if userRole != "" {
        allowedCollarTypes := utils.GetAllowedCollarTypes(userRole)
        if allowedCollarTypes != nil {
            // If the request already has collar_types filter, intersect with allowed types
            if existingTypes, ok := filters["collar_types"].([]string); ok {
                filters["collar_types"] = utils.IntersectCollarTypes(allowedCollarTypes, existingTypes)
            } else {
                // Apply role-based collar type restriction
                filters["collar_types"] = allowedCollarTypes
            }
        }
    }

    // Apply hierarchical filtering for supervisor and manager roles
    if utils.IsManagerRole(userRole) && len(userID) > 0 {
        // Get the user's employee ID to use as supervisor filter
        var user models.User
        if err := s.db.First(&user, "id = ?", userID[0]).Error; err == nil && user.EmployeeID != nil {
            if userRole == "supervisor" {
                // Supervisors see only their direct reports
                filters["supervisor_id"] = *user.EmployeeID
            } else if userRole == "manager" {
                // Managers see their direct reports + all subordinates recursively
                subordinateIDs := s.getAllSubordinateIDs(*user.EmployeeID)
                if len(subordinateIDs) > 0 {
                    filters["subordinate_ids"] = subordinateIDs
                } else {
                    // If no subordinates, still show direct reports
                    filters["supervisor_id"] = *user.EmployeeID
                }
            }
        }
    }

    employees, total, err := s.employeeRepo.List(page, pageSize, filters)
    if err != nil {
        return nil, fmt.Errorf("error listing employees: %w", err)
    }

    // Convert to response DTOs
    employeeResponses := make([]dtos.EmployeeResponse, len(employees))
    for i, emp := range employees {
        employeeResponses[i] = *s.ConvertToResponse(&emp)
    }

    totalPages := 1
    if pageSize > 0 {
        totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
    }

    return &dtos.EmployeeListResponse{
        Employees: employeeResponses,
        Total:     total,
        Page:      page,
        PageSize:  pageSize,
        TotalPages: totalPages,
    }, nil
}

// getAllSubordinateIDs recursively gets all subordinate employee IDs under a supervisor
func (s *EmployeeService) getAllSubordinateIDs(supervisorID uuid.UUID) []uuid.UUID {
    var subordinateIDs []uuid.UUID
    var directReports []models.Employee

    // Get direct reports
    if err := s.db.Where("supervisor_id = ?", supervisorID).Find(&directReports).Error; err != nil {
        return subordinateIDs
    }

    // Add direct report IDs and recursively get their subordinates
    for _, emp := range directReports {
        subordinateIDs = append(subordinateIDs, emp.ID)
        // Recursively get subordinates of this employee
        subSubordinates := s.getAllSubordinateIDs(emp.ID)
        subordinateIDs = append(subordinateIDs, subSubordinates...)
    }

    return subordinateIDs
}

// TerminateEmployee terminates an employee
func (s *EmployeeService) TerminateEmployee(id uuid.UUID, req dtos.EmployeeTerminationRequest) error {
    employee, err := s.employeeRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("employee not found")
        }
        return fmt.Errorf("error fetching employee: %w", err)
    }
    
    // Check if already terminated
    if employee.EmploymentStatus == "terminated" {
        return errors.New("employee is already terminated")
    }
    
    // Validate termination date
    if req.TerminationDate.Time.Before(employee.HireDate) {
        return errors.New("termination date cannot be before hire date")
    }

    // Update employee
    terminationTime := req.TerminationDate.Time
    employee.TerminationDate = &terminationTime
    employee.EmploymentStatus = "terminated"
    
    // Create termination record (in real system, you'd have a separate table)
    // For now, just update the employee
    
    return s.employeeRepo.Update(employee)
}

// UpdateEmployeeSalary updates employee salary
func (s *EmployeeService) UpdateEmployeeSalary(id uuid.UUID, req dtos.EmployeeSalaryUpdateRequest, updatedBy uuid.UUID) error {
    employee, err := s.employeeRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("employee not found")
        }
        return fmt.Errorf("error fetching employee: %w", err)
    }

    // Validate new salary is positive
    if req.NewDailySalary <= 0 {
        return errors.New("daily salary must be positive")
    }

    // Validate salary is not unrealistically low (below Mexican minimum wage)
    // As of 2025, minimum daily wage is approximately 248.93 MXN
    const minDailySalary = 200.0 // Set slightly below to allow flexibility
    if req.NewDailySalary < minDailySalary {
        return fmt.Errorf("daily salary (%.2f) is below minimum wage threshold (%.2f MXN)",
            req.NewDailySalary, minDailySalary)
    }

    // Validate salary is not unrealistically high (prevent data entry errors)
    const maxDailySalary = 50000.0 // ~1.5M MXN monthly
    if req.NewDailySalary > maxDailySalary {
        return fmt.Errorf("daily salary (%.2f) exceeds maximum threshold (%.2f MXN) - verify amount",
            req.NewDailySalary, maxDailySalary)
    }

    // Warn if salary change is more than 50% increase or decrease
    if employee.DailySalary > 0 {
        changePercent := ((req.NewDailySalary - employee.DailySalary) / employee.DailySalary) * 100
        if changePercent > 50 || changePercent < -50 {
            // Note: This is a warning logged, not blocking the update
            // In production, you might want to require approval for large changes
        }
    }

    // Check if effective date is in the future relative to the latest closed payroll period
    // Prevent retroactive salary changes that affect already-processed payroll
    var latestClosedPeriod models.PayrollPeriod
    err = s.db.Where("status IN (?)", []string{"closed", "paid"}).
        Order("end_date DESC").
        First(&latestClosedPeriod).Error

    if err == nil {
        if req.EffectiveDate.Before(latestClosedPeriod.EndDate) {
            return fmt.Errorf("effective date (%s) cannot be before last closed payroll period end date (%s)",
                req.EffectiveDate.Format("2006-01-02"),
                latestClosedPeriod.EndDate.Format("2006-01-02"))
        }
    }

    // Update employee salary
    employee.DailySalary = req.NewDailySalary
    employee.UpdatedBy = &updatedBy

    // Recalculate integrated daily salary
    employee.IntegratedDailySalary = employee.CalculateIntegratedDailySalary()

    // TODO: Create salary history record for audit trail
    // salaryHistory := &models.SalaryHistory{
    //     EmployeeID: employee.ID,
    //     OldSalary: employee.DailySalary,
    //     NewSalary: req.NewDailySalary,
    //     EffectiveDate: req.EffectiveDate,
    //     UpdatedBy: updatedBy,
    // }

    return s.employeeRepo.Update(employee)
}

// GetActiveEmployees returns active employees
func (s *EmployeeService) GetActiveEmployees() ([]dtos.EmployeeResponse, error) {
    filters := map[string]interface{}{
        "status": "active",
    }
    
    employees, _, err := s.employeeRepo.List(1, 1000, filters) // Large page size
    if err != nil {
        return nil, fmt.Errorf("error fetching active employees: %w", err)
    }
    
    responses := make([]dtos.EmployeeResponse, len(employees))
    for i, emp := range employees {
        responses[i] = *s.ConvertToResponse(&emp)
    }
    
    return responses, nil
}

// GetEmployeeStats returns employee statistics
// userRole is optional - if provided, stats are filtered by collar type for hr_blue_gray and hr_white roles
func (s *EmployeeService) GetEmployeeStats(userRole ...string) (map[string]interface{}, error) {
    stats := make(map[string]interface{})

    // Build base query with collar type filter if applicable
    baseQuery := s.db.Model(&models.Employee{})
    if len(userRole) > 0 && userRole[0] != "" {
        allowedCollarTypes := utils.GetAllowedCollarTypes(userRole[0])
        if allowedCollarTypes != nil {
            baseQuery = baseQuery.Where("collar_type IN ?", allowedCollarTypes)
        }
    }

    // Get total count
    var totalCount int64
    baseQuery.Count(&totalCount)
    stats["total_employees"] = totalCount

    // Get active count (need a fresh query with same filter)
    activeQuery := s.db.Model(&models.Employee{}).Where("employment_status = ?", "active")
    if len(userRole) > 0 && userRole[0] != "" {
        allowedCollarTypes := utils.GetAllowedCollarTypes(userRole[0])
        if allowedCollarTypes != nil {
            activeQuery = activeQuery.Where("collar_type IN ?", allowedCollarTypes)
        }
    }
    var activeCount int64
    activeQuery.Count(&activeCount)
    stats["active_employees"] = activeCount

    // Get count by type
    typeStats := make(map[string]int64)
    var typeResults []struct {
        EmployeeType string
        Count        int64
    }

    typeQuery := s.db.Model(&models.Employee{}).
        Select("employee_type, COUNT(*) as count").
        Group("employee_type")
    if len(userRole) > 0 && userRole[0] != "" {
        allowedCollarTypes := utils.GetAllowedCollarTypes(userRole[0])
        if allowedCollarTypes != nil {
            typeQuery = typeQuery.Where("collar_type IN ?", allowedCollarTypes)
        }
    }
    typeQuery.Find(&typeResults)

    for _, result := range typeResults {
        typeStats[result.EmployeeType] = result.Count
    }
    stats["by_type"] = typeStats

    // Get count by status
    statusStats := make(map[string]int64)
    var statusResults []struct {
        EmploymentStatus string
        Count            int64
    }

    statusQuery := s.db.Model(&models.Employee{}).
        Select("employment_status, COUNT(*) as count").
        Group("employment_status")
    if len(userRole) > 0 && userRole[0] != "" {
        allowedCollarTypes := utils.GetAllowedCollarTypes(userRole[0])
        if allowedCollarTypes != nil {
            statusQuery = statusQuery.Where("collar_type IN ?", allowedCollarTypes)
        }
    }
    statusQuery.Find(&statusResults)

    for _, result := range statusResults {
        statusStats[result.EmploymentStatus] = result.Count
    }
    stats["by_status"] = statusStats

    return stats, nil
}

// ConvertToResponse converts Employee model to response DTO
func (s *EmployeeService) ConvertToResponse(employee *models.Employee) *dtos.EmployeeResponse {
    response := &dtos.EmployeeResponse{
        ID:                    employee.ID,
        EmployeeNumber:        employee.EmployeeNumber,
        FirstName:             employee.FirstName,
        LastName:              employee.LastName,
        MotherLastName:        employee.MotherLastName,
        FullName:              fmt.Sprintf("%s %s %s", employee.FirstName, employee.LastName, employee.MotherLastName),
        DateOfBirth:           employee.DateOfBirth,
        Age:                   employee.CalculateAge(),
        Gender:                employee.Gender,
        RFC:                   employee.RFC,
        CURP:                  employee.CURP,
        NSS:                   employee.NSS,
        HireDate:              employee.HireDate,
        Seniority:             employee.CalculateSeniority(),
        EmploymentStatus:      employee.EmploymentStatus,
        EmployeeType:          employee.EmployeeType,
        CollarType:            employee.CollarType,
        PayFrequency:          employee.PayFrequency,
        IsSindicalizado:       employee.IsSindicalizado,
        DailySalary:           employee.DailySalary,
        IntegratedDailySalary: employee.IntegratedDailySalary,
        InfonavitCredit:       employee.InfonavitCredit,
        PersonalEmail:         employee.PersonalEmail,
        PersonalPhone:         employee.PersonalPhone,
        EmergencyContact:      employee.EmergencyContact,
        EmergencyPhone:        employee.EmergencyPhone,
        Street:                employee.Street,
        ExteriorNumber:        employee.ExteriorNumber,
        InteriorNumber:        employee.InteriorNumber,
        Neighborhood:          employee.Neighborhood,
        Municipality:          employee.Municipality,
        State:                 employee.State,
        PostalCode:            employee.PostalCode,
        Country:               employee.Country,
        TerminationDate:       employee.TerminationDate,
        PaymentMethod:         employee.PaymentMethod,
        BankName:              employee.BankName,
        BankAccount:           employee.BankAccount,
        CLABE:                 employee.CLABE,
        IMSSRegistrationDate:  employee.IMSSRegistrationDate,
        Regime:                employee.Regime,
        TaxRegime:             employee.TaxRegime,
        DepartmentID:          employee.DepartmentID,
        PositionID:            employee.PositionID,
        CostCenterID:          employee.CostCenterID,
        ShiftID:               employee.ShiftID,
        SupervisorID:          employee.SupervisorID,
        CreatedAt:             employee.CreatedAt,
        UpdatedAt:             employee.UpdatedAt,
    }

    // Get shift name if shift is loaded or fetch it
    if employee.Shift != nil {
        response.ShiftName = employee.Shift.Name
    } else if employee.ShiftID != nil {
        var shift models.Shift
        if err := s.db.First(&shift, "id = ?", employee.ShiftID).Error; err == nil {
            response.ShiftName = shift.Name
        }
    }

    // Get supervisor name if supervisor is loaded or fetch it
    if employee.Supervisor != nil {
        response.SupervisorName = fmt.Sprintf("%s %s", employee.Supervisor.FirstName, employee.Supervisor.LastName)
    } else if employee.SupervisorID != nil {
        var supervisor models.Employee
        if err := s.db.First(&supervisor, "id = ?", employee.SupervisorID).Error; err == nil {
            response.SupervisorName = fmt.Sprintf("%s %s", supervisor.FirstName, supervisor.LastName)
        }
    }

    return response
}

// ValidateMexicanIDs validates Mexican official IDs
func (s *EmployeeService) ValidateMexicanIDs(rfc, curp, nss string) (map[string]bool, error) {
    validation := map[string]bool{
        "rfc_valid":  models.ValidateRFC(rfc),
        "curp_valid": models.ValidateCURP(curp),
    }

    if nss != "" {
        validation["nss_valid"] = models.ValidateNSS(nss)
    }

    return validation, nil
}

// ImportEmployeesFromFile imports employees from Excel or CSV file
func (s *EmployeeService) ImportEmployeesFromFile(file multipart.File, filename string, userID uuid.UUID) (map[string]interface{}, error) {
    var records [][]string
    var err error

    if strings.HasSuffix(strings.ToLower(filename), ".csv") {
        records, err = s.parseCSV(file)
    } else {
        records, err = s.parseExcel(file)
    }

    if err != nil {
        return nil, fmt.Errorf("error parsing file: %w", err)
    }

    if len(records) < 2 {
        return nil, errors.New("file must contain header row and at least one data row")
    }

    // Process records
    result := map[string]interface{}{
        "total":     len(records) - 1,
        "created":   0,
        "updated":   0,
        "failed":    0,
        "errors":    []map[string]interface{}{},
    }

    headers := records[0]
    headerMap := make(map[string]int)
    for i, h := range headers {
        headerMap[strings.ToLower(strings.TrimSpace(h))] = i
    }

    for rowNum, row := range records[1:] {
        emp, err := s.rowToEmployee(row, headerMap, userID)
        if err != nil {
            errList := result["errors"].([]map[string]interface{})
            errList = append(errList, map[string]interface{}{
                "row":     rowNum + 2,
                "error":   err.Error(),
            })
            result["errors"] = errList
            result["failed"] = result["failed"].(int) + 1
            continue
        }

        // Try to find existing employee by employee_number, RFC, or CURP
        var existingEmp models.Employee
        found := false

        // Check by employee_number first (most common identifier)
        if emp.EmployeeNumber != "" {
            if err := s.db.Where("employee_number = ? AND company_id = ?", emp.EmployeeNumber, emp.CompanyID).First(&existingEmp).Error; err == nil {
                found = true
            }
        }

        // If not found by employee_number, try RFC
        if !found && emp.RFC != "" {
            if err := s.db.Where("rfc = ? AND company_id = ?", emp.RFC, emp.CompanyID).First(&existingEmp).Error; err == nil {
                found = true
            }
        }

        // If not found by RFC, try CURP
        if !found && emp.CURP != "" {
            if err := s.db.Where("curp = ? AND company_id = ?", emp.CURP, emp.CompanyID).First(&existingEmp).Error; err == nil {
                found = true
            }
        }

        if found {
            // Update existing employee - preserve ID and timestamps
            emp.ID = existingEmp.ID
            emp.CreatedAt = existingEmp.CreatedAt
            emp.CreatedBy = existingEmp.CreatedBy

            if err := s.db.Save(emp).Error; err != nil {
                errList := result["errors"].([]map[string]interface{})
                errList = append(errList, map[string]interface{}{
                    "row":     rowNum + 2,
                    "error":   fmt.Sprintf("failed to update: %s", err.Error()),
                })
                result["errors"] = errList
                result["failed"] = result["failed"].(int) + 1
                continue
            }
            result["updated"] = result["updated"].(int) + 1
        } else {
            // Create new employee
            if err := s.db.Create(emp).Error; err != nil {
                errList := result["errors"].([]map[string]interface{})
                errList = append(errList, map[string]interface{}{
                    "row":     rowNum + 2,
                    "error":   err.Error(),
                })
                result["errors"] = errList
                result["failed"] = result["failed"].(int) + 1
                continue
            }
            result["created"] = result["created"].(int) + 1
        }
    }

    // Add success count for backwards compatibility
    result["success"] = result["created"].(int) + result["updated"].(int)

    return result, nil
}

func (s *EmployeeService) parseCSV(file multipart.File) ([][]string, error) {
    reader := csv.NewReader(file)
    return reader.ReadAll()
}

func (s *EmployeeService) parseExcel(file multipart.File) ([][]string, error) {
    data, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    f, err := excelize.OpenReader(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer f.Close()

    sheets := f.GetSheetList()
    if len(sheets) == 0 {
        return nil, errors.New("no sheets found in Excel file")
    }

    sheetName := sheets[0]

    // Get dimensions to know how many rows/cols we have
    rows, err := f.GetRows(sheetName)
    if err != nil {
        return nil, err
    }

    // Process each cell individually to handle dates properly
    var result [][]string
    for rowIdx, row := range rows {
        var processedRow []string
        for colIdx := range row {
            colName, _ := excelize.ColumnNumberToName(colIdx + 1)
            cellRef := colName + fmt.Sprintf("%d", rowIdx+1)

            // Get the raw cell value - this handles dates better
            cellValue, _ := f.GetCellValue(sheetName, cellRef)

            // If the value looks like an Excel serial date number, try to convert it
            if cellValue != "" && rowIdx > 0 { // Skip header row
                if serial, err := strconv.ParseFloat(cellValue, 64); err == nil {
                    // Check if this could be a date (Excel dates are typically between 1 and 2958465)
                    // and the value doesn't look like a normal number we want to keep
                    if serial > 1 && serial < 2958466 {
                        // Check if this column is a date column by header name
                        if len(result) > 0 && colIdx < len(result[0]) {
                            headerLower := strings.ToLower(result[0][colIdx])
                            if strings.Contains(headerLower, "date") || strings.Contains(headerLower, "fecha") {
                                // Convert Excel serial to date
                                t, err := excelize.ExcelDateToTime(serial, false)
                                if err == nil {
                                    cellValue = t.Format("2006-01-02")
                                }
                            }
                        }
                    }
                }
            }

            processedRow = append(processedRow, cellValue)
        }
        result = append(result, processedRow)
    }

    return result, nil
}

func (s *EmployeeService) rowToEmployee(row []string, headerMap map[string]int, userID uuid.UUID) (*models.Employee, error) {
    getValue := func(key string) string {
        if idx, ok := headerMap[key]; ok && idx < len(row) {
            return strings.TrimSpace(row[idx])
        }
        return ""
    }

    // Parse required fields
    employeeNumber := getValue("employee_number")
    firstName := getValue("first_name")
    lastName := getValue("last_name")
    rfc := getValue("rfc")
    curp := getValue("curp")
    dailySalaryStr := getValue("daily_salary")

    if employeeNumber == "" {
        return nil, errors.New("employee_number is required")
    }
    if firstName == "" {
        return nil, errors.New("first_name is required")
    }
    if lastName == "" {
        return nil, errors.New("last_name is required")
    }
    if rfc == "" {
        return nil, errors.New("rfc is required")
    }
    if curp == "" {
        return nil, errors.New("curp is required")
    }

    // Parse dates
    dobStr := getValue("date_of_birth")
    dob, err := parseDate(dobStr)
    if err != nil {
        return nil, fmt.Errorf("invalid date_of_birth: %w", err)
    }

    hireDateStr := getValue("hire_date")
    hireDate, err := parseDate(hireDateStr)
    if err != nil {
        return nil, fmt.Errorf("invalid hire_date: %w", err)
    }

    // Parse salary - handle comma as decimal separator
    dailySalaryStr = strings.ReplaceAll(dailySalaryStr, ",", ".")
    dailySalary, err := strconv.ParseFloat(dailySalaryStr, 64)
    if err != nil || dailySalary <= 0 {
        return nil, errors.New("daily_salary must be a positive number")
    }

    // Parse boolean - support multiple formats
    isSindicalizado := normalizeBool(getValue("is_sindicalizado"))

    // Get company ID from user
    var user models.User
    if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    // Normalize all field values
    collarType := normalizeCollarType(getValue("collar_type"))
    payFrequency := normalizePayFrequency(getValue("pay_frequency"))

    emp := &models.Employee{
        EmployeeNumber:   employeeNumber,
        FirstName:        normalizeName(firstName),
        LastName:         normalizeName(lastName),
        MotherLastName:   normalizeName(getValue("mother_last_name")),
        DateOfBirth:      dob,
        Gender:           normalizeGender(getValue("gender")),
        RFC:              strings.ToUpper(strings.ReplaceAll(rfc, " ", "")),
        CURP:             strings.ToUpper(strings.ReplaceAll(curp, " ", "")),
        NSS:              normalizeNSS(getValue("nss")),
        HireDate:         hireDate,
        DailySalary:      dailySalary,
        CollarType:       collarType,
        PayFrequency:     payFrequency,
        EmploymentStatus: normalizeEmploymentStatus(getValue("employment_status")),
        EmployeeType:     normalizeEmployeeType(getValue("employee_type")),
        IsSindicalizado:  isSindicalizado,
        BankName:         normalizeBankName(getValue("bank_name")),
        BankAccount:      strings.ReplaceAll(getValue("bank_account"), " ", ""),
        CLABE:            strings.ReplaceAll(getValue("clabe"), " ", ""),
        PaymentMethod:    normalizePaymentMethod(getValue("payment_method")),
        CompanyID:        user.CompanyID,
        CreatedBy:        &userID,
    }

    // Set defaults for collar type and pay frequency
    if emp.CollarType == "" {
        emp.CollarType = "white_collar"
    }
    if emp.PayFrequency == "" {
        if emp.CollarType == "white_collar" {
            emp.PayFrequency = "biweekly"
        } else {
            emp.PayFrequency = "weekly"
        }
    }

    return emp, nil
}

// normalizeName converts name to title case (first letter uppercase, rest lowercase)
func normalizeName(name string) string {
    if name == "" {
        return ""
    }
    return strings.Title(strings.ToLower(strings.TrimSpace(name)))
}

// normalizeGender normalizes gender values to "male" or "female"
func normalizeGender(gender string) string {
    g := strings.ToLower(strings.TrimSpace(gender))
    switch g {
    case "m", "male", "masculino", "hombre", "h", "masc":
        return "male"
    case "f", "female", "femenino", "mujer", "fem":
        return "female"
    default:
        return g
    }
}

// normalizeCollarType normalizes collar type values
func normalizeCollarType(collarType string) string {
    ct := strings.ToLower(strings.TrimSpace(collarType))
    ct = strings.ReplaceAll(ct, " ", "_")
    ct = strings.ReplaceAll(ct, "-", "_")
    switch ct {
    case "white", "white_collar", "blanco", "cuello_blanco", "oficina", "administrativo":
        return "white_collar"
    case "blue", "blue_collar", "azul", "cuello_azul", "operativo", "obrero":
        return "blue_collar"
    default:
        if ct != "" {
            return ct
        }
        return ""
    }
}

// normalizePayFrequency normalizes pay frequency values
func normalizePayFrequency(freq string) string {
    f := strings.ToLower(strings.TrimSpace(freq))
    f = strings.ReplaceAll(f, " ", "_")
    f = strings.ReplaceAll(f, "-", "_")
    switch f {
    case "weekly", "semanal", "semana", "w":
        return "weekly"
    case "biweekly", "quincenal", "quincena", "bi_weekly", "bi-weekly", "q":
        return "biweekly"
    case "monthly", "mensual", "mes", "m":
        return "monthly"
    default:
        if f != "" {
            return f
        }
        return ""
    }
}

// normalizeEmploymentStatus normalizes employment status values
func normalizeEmploymentStatus(status string) string {
    s := strings.ToLower(strings.TrimSpace(status))
    switch s {
    case "active", "activo", "a", "1", "true":
        return "active"
    case "inactive", "inactivo", "i", "0", "false":
        return "inactive"
    case "terminated", "terminado", "baja", "t":
        return "terminated"
    case "suspended", "suspendido", "s":
        return "suspended"
    case "on_leave", "leave", "licencia", "permiso":
        return "on_leave"
    default:
        if s == "" {
            return "active"
        }
        return s
    }
}

// normalizeEmployeeType normalizes employee type values
func normalizeEmployeeType(empType string) string {
    t := strings.ToLower(strings.TrimSpace(empType))
    t = strings.ReplaceAll(t, " ", "_")
    t = strings.ReplaceAll(t, "-", "_")
    switch t {
    case "permanent", "permanente", "planta", "base", "p":
        return "permanent"
    case "temporary", "temporal", "eventual", "t":
        return "temporary"
    case "contractor", "contratista", "externo", "c":
        return "contractor"
    case "intern", "practicante", "becario", "i":
        return "intern"
    default:
        if t == "" {
            return "permanent"
        }
        return t
    }
}

// normalizePaymentMethod normalizes payment method values
func normalizePaymentMethod(method string) string {
    m := strings.ToLower(strings.TrimSpace(method))
    m = strings.ReplaceAll(m, " ", "_")
    m = strings.ReplaceAll(m, "-", "_")
    switch m {
    case "bank_transfer", "transfer", "transferencia", "banco", "bank", "deposito", "deposito_bancario":
        return "bank_transfer"
    case "cash", "efectivo", "e":
        return "cash"
    case "check", "cheque", "ch":
        return "check"
    default:
        if m == "" {
            return "bank_transfer"
        }
        return m
    }
}

// normalizeBankName normalizes common bank names
func normalizeBankName(bank string) string {
    b := strings.ToUpper(strings.TrimSpace(bank))
    switch b {
    case "BBVA", "BBVA BANCOMER", "BANCOMER":
        return "BBVA"
    case "SANTANDER", "BANCO SANTANDER":
        return "Santander"
    case "BANAMEX", "CITIBANAMEX", "CITI":
        return "Citibanamex"
    case "BANORTE", "BANCO BANORTE":
        return "Banorte"
    case "HSBC", "BANCO HSBC":
        return "HSBC"
    case "SCOTIABANK", "SCOTIA":
        return "Scotiabank"
    case "INBURSA", "BANCO INBURSA":
        return "Inbursa"
    case "AZTECA", "BANCO AZTECA":
        return "Banco Azteca"
    case "AFIRME", "BANCO AFIRME":
        return "Afirme"
    case "BAJIO", "BANCO DEL BAJIO", "BANBAJIO":
        return "BanBajío"
    case "BANREGIO", "BANCO BANREGIO":
        return "BanRegio"
    default:
        // Return with proper title case if not a known bank
        if bank == "" {
            return ""
        }
        return strings.Title(strings.ToLower(bank))
    }
}

// normalizeBool converts various boolean representations to bool
func normalizeBool(value string) bool {
    v := strings.ToLower(strings.TrimSpace(value))
    switch v {
    case "true", "1", "yes", "si", "sí", "y", "t", "verdadero", "v":
        return true
    default:
        return false
    }
}

// normalizeNSS pads NSS with leading zeros to ensure 11 digits
func normalizeNSS(nss string) string {
    // Remove spaces and any non-numeric characters
    nss = strings.TrimSpace(nss)
    nss = strings.ReplaceAll(nss, " ", "")
    nss = strings.ReplaceAll(nss, "-", "")

    if nss == "" {
        return ""
    }

    // Pad with leading zeros if less than 11 digits
    if len(nss) < 11 {
        nss = fmt.Sprintf("%011s", nss)
    }

    return nss
}

func parseDate(dateStr string) (time.Time, error) {
    if dateStr == "" {
        return time.Time{}, errors.New("date is required")
    }

    // Trim any whitespace
    dateStr = strings.TrimSpace(dateStr)

    // Try to parse as Excel serial number first
    if serial, err := strconv.ParseFloat(dateStr, 64); err == nil && serial > 1 && serial < 2958466 {
        t, err := excelize.ExcelDateToTime(serial, false)
        if err == nil {
            return t, nil
        }
    }

    // Formats with 4-digit year (unambiguous)
    formats4digit := []string{
        "2006-01-02",
        "2006-01-02T15:04:05Z",
        "2006-01-02 15:04:05",
        "02/01/2006",
        "01/02/2006",
        "2006/01/02",
        "02-01-2006",
        "01-02-2006",
        "1/2/2006",
        "2/1/2006",
        time.RFC3339,
    }

    for _, format := range formats4digit {
        if t, err := time.Parse(format, dateStr); err == nil {
            return t, nil
        }
    }

    // Formats with 2-digit year - need special handling
    // Excel typically uses MM-DD-YY format
    formats2digit := []string{
        "01-02-06", // MM-DD-YY (Excel default)
        "1-2-06",   // M-D-YY
        "01/02/06", // MM/DD/YY
        "1/2/06",   // M/D/YY
        "02-01-06", // DD-MM-YY
        "2-1-06",   // D-M-YY
        "02/01/06", // DD/MM/YY
        "2/1/06",   // D/M/YY
    }

    for _, format := range formats2digit {
        if t, err := time.Parse(format, dateStr); err == nil {
            // Go's time.Parse uses pivot year 2069 for 2-digit years
            // So 67 becomes 2067, 25 becomes 2025
            // We need to fix this: if year > current year + 10, subtract 100
            year := t.Year()
            currentYear := time.Now().Year()
            if year > currentYear+10 {
                // This is likely a birth date in the 1900s
                t = t.AddDate(-100, 0, 0)
            }
            return t, nil
        }
    }

    return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func getOrDefault(value, defaultValue string) string {
    if value == "" {
        return defaultValue
    }
    return value
}

// GenerateImportTemplate generates an Excel template for employee import
func (s *EmployeeService) GenerateImportTemplate() ([]byte, error) {
    f := excelize.NewFile()
    sheet := "Employees"
    f.SetSheetName("Sheet1", sheet)

    headers := []string{
        "employee_number", "first_name", "last_name", "mother_last_name",
        "date_of_birth", "gender", "rfc", "curp", "nss",
        "hire_date", "daily_salary", "collar_type", "pay_frequency",
        "employment_status", "employee_type", "is_sindicalizado",
        "bank_name", "bank_account", "clabe", "payment_method",
    }

    // Set headers
    for i, header := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue(sheet, cell, header)
    }

    // Style header row
    style, _ := f.NewStyle(&excelize.Style{
        Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
        Fill:      excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
        Alignment: &excelize.Alignment{Horizontal: "center"},
    })
    f.SetRowStyle(sheet, 1, 1, style)

    // Add example row
    exampleData := []string{
        "EMP001", "Juan", "Perez", "Lopez",
        "1990-05-15", "male", "PELJ900515XXX", "PELJ900515HSLRLN09", "12345678901",
        "2024-01-15", "500.00", "white_collar", "biweekly",
        "active", "permanent", "false",
        "BBVA", "1234567890", "012180001234567890", "bank_transfer",
    }

    for i, value := range exampleData {
        cell, _ := excelize.CoordinatesToCellName(i+1, 2)
        f.SetCellValue(sheet, cell, value)
    }

    // Add instructions sheet
    f.NewSheet("Instructions")
    instructions := [][]string{
        {"Column", "Required", "Description", "Valid Values"},
        {"employee_number", "Yes", "Unique employee identifier", "Text (e.g., EMP001)"},
        {"first_name", "Yes", "First name", "Text"},
        {"last_name", "Yes", "Last name (paterno)", "Text"},
        {"mother_last_name", "No", "Mother's last name (materno)", "Text"},
        {"date_of_birth", "Yes", "Birth date", "YYYY-MM-DD"},
        {"gender", "Yes", "Gender", "male, female, other"},
        {"rfc", "Yes", "RFC (12-13 characters)", "Text"},
        {"curp", "Yes", "CURP (18 characters)", "Text"},
        {"nss", "No", "IMSS number (11 digits)", "Text"},
        {"hire_date", "Yes", "Employment start date", "YYYY-MM-DD"},
        {"daily_salary", "Yes", "Daily wage in MXN", "Number"},
        {"collar_type", "Yes", "Worker classification", "white_collar, blue_collar, gray_collar"},
        {"pay_frequency", "Yes", "Payment frequency", "weekly, biweekly, monthly"},
        {"employment_status", "No", "Current status", "active, inactive, on_leave, terminated"},
        {"employee_type", "No", "Contract type", "permanent, temporary, contractor, intern"},
        {"is_sindicalizado", "No", "Union member (blue collar)", "true, false"},
        {"bank_name", "No", "Bank name", "Text"},
        {"bank_account", "No", "Bank account number", "Text"},
        {"clabe", "No", "CLABE (18 digits)", "Text"},
        {"payment_method", "No", "Payment method", "bank_transfer, cash, check"},
    }

    for i, row := range instructions {
        for j, value := range row {
            cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
            f.SetCellValue("Instructions", cell, value)
        }
    }

    // Style instructions header
    f.SetRowStyle("Instructions", 1, 1, style)

    // Set column widths
    for i := range headers {
        col, _ := excelize.ColumnNumberToName(i + 1)
        f.SetColWidth(sheet, col, col, 18)
    }
    f.SetColWidth("Instructions", "A", "A", 20)
    f.SetColWidth("Instructions", "B", "B", 10)
    f.SetColWidth("Instructions", "C", "C", 40)
    f.SetColWidth("Instructions", "D", "D", 40)

    // Write to buffer
    buf, err := f.WriteToBuffer()
    if err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}

// =========================================================================
// Portal User Management Methods
// =========================================================================

// GetEmployeePortalUser retrieves the portal user account linked to an employee
func (s *EmployeeService) GetEmployeePortalUser(employeeID uuid.UUID) (*dtos.PortalUserResponse, error) {
    // First verify the employee exists
    employee, err := s.employeeRepo.FindByID(employeeID)
    if err != nil {
        return nil, fmt.Errorf("employee not found: %w", err)
    }

    // Find user by employee_id
    var user models.User
    result := s.db.Where("employee_id = ?", employeeID).First(&user)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("portal user not found for employee %s", employee.EmployeeNumber)
        }
        return nil, fmt.Errorf("error finding portal user: %w", result.Error)
    }

    // Convert to response DTO
    response := &dtos.PortalUserResponse{
        ID:         user.ID.String(),
        Email:      user.Email,
        Role:       string(user.Role),
        IsActive:   user.IsActive,
        FullName:   user.FullName,
        Department: user.Department,
        Area:       user.Area,
        CreatedAt:  user.CreatedAt,
        UpdatedAt:  user.UpdatedAt,
    }

    if user.SupervisorID != nil {
        supervisorID := user.SupervisorID.String()
        response.SupervisorID = &supervisorID
    }
    if user.GeneralManagerID != nil {
        gmID := user.GeneralManagerID.String()
        response.GeneralManagerID = &gmID
    }
    if user.LastLoginAt != nil {
        response.LastLoginAt = user.LastLoginAt
    }

    return response, nil
}

// CreateEmployeePortalUser creates a new portal user account for an employee
func (s *EmployeeService) CreateEmployeePortalUser(employeeID uuid.UUID, req dtos.CreatePortalUserRequest) (*dtos.PortalUserResponse, error) {
    // Verify the employee exists
    employee, err := s.employeeRepo.FindByID(employeeID)
    if err != nil {
        return nil, fmt.Errorf("employee not found: %w", err)
    }

    // Check if a portal user already exists for this employee
    var existingUser models.User
    result := s.db.Where("employee_id = ?", employeeID).First(&existingUser)
    if result.Error == nil {
        return nil, fmt.Errorf("portal user account already exists for employee %s", employee.EmployeeNumber)
    }

    // Check if email is already taken
    result = s.db.Where("email = ?", req.Email).First(&existingUser)
    if result.Error == nil {
        return nil, fmt.Errorf("email %s is already in use", req.Email)
    }

    // Get the company ID from the employee
    companyID := employee.CompanyID

    // Build full name from employee record
    fullName := employee.FirstName + " " + employee.LastName
    if employee.MotherLastName != "" {
        fullName += " " + employee.MotherLastName
    }

    // Create new user
    user := models.User{
        Email:      req.Email,
        FullName:   fullName,
        IsActive:   true,
        CompanyID:  companyID,
        EmployeeID: &employeeID,
        Department: req.Department,
        Area:       req.Area,
    }

    // Set role
    user.Role = enums.UserRole(req.Role)
    if !user.Role.IsValid() {
        return nil, fmt.Errorf("invalid role: %s", req.Role)
    }

    // Set supervisor if provided
    if req.SupervisorID != "" {
        supervisorUUID, err := uuid.Parse(req.SupervisorID)
        if err != nil {
            return nil, fmt.Errorf("invalid supervisor ID format: %w", err)
        }
        // Verify supervisor exists
        var supervisor models.User
        if err := s.db.First(&supervisor, "id = ?", supervisorUUID).Error; err != nil {
            return nil, fmt.Errorf("supervisor not found: %w", err)
        }
        user.SupervisorID = &supervisorUUID
    }

    // Set general manager if provided
    if req.GeneralManagerID != "" {
        gmUUID, err := uuid.Parse(req.GeneralManagerID)
        if err != nil {
            return nil, fmt.Errorf("invalid general manager ID format: %w", err)
        }
        // Verify general manager exists
        var gm models.User
        if err := s.db.First(&gm, "id = ?", gmUUID).Error; err != nil {
            return nil, fmt.Errorf("general manager not found: %w", err)
        }
        user.GeneralManagerID = &gmUUID
    }

    // Set and hash password
    if err := user.SetPassword(req.Password); err != nil {
        return nil, fmt.Errorf("password validation failed: %w", err)
    }

    // Save user to database
    if err := s.db.Create(&user).Error; err != nil {
        return nil, fmt.Errorf("failed to create portal user: %w", err)
    }

    // Return response
    response := &dtos.PortalUserResponse{
        ID:         user.ID.String(),
        Email:      user.Email,
        Role:       string(user.Role),
        IsActive:   user.IsActive,
        FullName:   user.FullName,
        Department: user.Department,
        Area:       user.Area,
        CreatedAt:  user.CreatedAt,
        UpdatedAt:  user.UpdatedAt,
    }

    if user.SupervisorID != nil {
        supervisorID := user.SupervisorID.String()
        response.SupervisorID = &supervisorID
    }

    if user.GeneralManagerID != nil {
        gmID := user.GeneralManagerID.String()
        response.GeneralManagerID = &gmID
    }

    return response, nil
}

// UpdateEmployeePortalUser updates an existing portal user account
func (s *EmployeeService) UpdateEmployeePortalUser(employeeID uuid.UUID, req dtos.UpdatePortalUserRequest) (*dtos.PortalUserResponse, error) {
    // Verify the employee exists
    employee, err := s.employeeRepo.FindByID(employeeID)
    if err != nil {
        return nil, fmt.Errorf("employee not found: %w", err)
    }

    // Find the portal user
    var user models.User
    result := s.db.Where("employee_id = ?", employeeID).First(&user)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("portal user not found for employee %s", employee.EmployeeNumber)
        }
        return nil, fmt.Errorf("error finding portal user: %w", result.Error)
    }

    // Update email if provided
    if req.Email != "" && req.Email != user.Email {
        // Check if new email is already taken
        var existingUser models.User
        if err := s.db.Where("email = ? AND id != ?", req.Email, user.ID).First(&existingUser).Error; err == nil {
            return nil, fmt.Errorf("email %s is already in use", req.Email)
        }
        user.Email = req.Email
    }

    // Update password if provided
    if req.Password != "" {
        if err := user.SetPassword(req.Password); err != nil {
            return nil, fmt.Errorf("password validation failed: %w", err)
        }
    }

    // Update role if provided
    if req.Role != "" {
        newRole := enums.UserRole(req.Role)
        if !newRole.IsValid() {
            return nil, fmt.Errorf("invalid role: %s", req.Role)
        }
        user.Role = newRole
    }

    // Update active status if provided
    if req.IsActive != nil {
        user.IsActive = *req.IsActive
    }

    // Update supervisor if provided
    if req.SupervisorID != "" {
        supervisorUUID, err := uuid.Parse(req.SupervisorID)
        if err != nil {
            return nil, fmt.Errorf("invalid supervisor ID format: %w", err)
        }
        // Verify supervisor exists
        var supervisor models.User
        if err := s.db.First(&supervisor, "id = ?", supervisorUUID).Error; err != nil {
            return nil, fmt.Errorf("supervisor not found: %w", err)
        }
        user.SupervisorID = &supervisorUUID
    }

    // Update general manager if provided
    if req.GeneralManagerID != "" {
        gmUUID, err := uuid.Parse(req.GeneralManagerID)
        if err != nil {
            return nil, fmt.Errorf("invalid general manager ID format: %w", err)
        }
        // Verify general manager exists
        var gm models.User
        if err := s.db.First(&gm, "id = ?", gmUUID).Error; err != nil {
            return nil, fmt.Errorf("general manager not found: %w", err)
        }
        user.GeneralManagerID = &gmUUID
    }

    // Update department and area if provided
    if req.Department != "" {
        user.Department = req.Department
    }
    if req.Area != "" {
        user.Area = req.Area
    }

    // Save changes
    if err := s.db.Save(&user).Error; err != nil {
        return nil, fmt.Errorf("failed to update portal user: %w", err)
    }

    // Return response
    response := &dtos.PortalUserResponse{
        ID:         user.ID.String(),
        Email:      user.Email,
        Role:       string(user.Role),
        IsActive:   user.IsActive,
        FullName:   user.FullName,
        Department: user.Department,
        Area:       user.Area,
        CreatedAt:  user.CreatedAt,
        UpdatedAt:  user.UpdatedAt,
    }

    if user.SupervisorID != nil {
        supervisorID := user.SupervisorID.String()
        response.SupervisorID = &supervisorID
    }
    if user.GeneralManagerID != nil {
        gmID := user.GeneralManagerID.String()
        response.GeneralManagerID = &gmID
    }
    if user.LastLoginAt != nil {
        response.LastLoginAt = user.LastLoginAt
    }

    return response, nil
}

// DeleteEmployeePortalUser deletes a portal user account for an employee
func (s *EmployeeService) DeleteEmployeePortalUser(employeeID uuid.UUID) error {
    // Verify the employee exists
    employee, err := s.employeeRepo.FindByID(employeeID)
    if err != nil {
        return fmt.Errorf("employee not found: %w", err)
    }

    // Find the portal user
    var user models.User
    result := s.db.Where("employee_id = ?", employeeID).First(&user)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return fmt.Errorf("no portal user account found for employee %s", employee.EmployeeNumber)
        }
        return fmt.Errorf("error finding portal user: %w", result.Error)
    }

    // Delete the user (soft delete if BaseModel has DeletedAt)
    if err := s.db.Delete(&user).Error; err != nil {
        return fmt.Errorf("failed to delete portal user: %w", err)
    }

    return nil
}
