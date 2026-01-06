/*
Package dtos - Employee Data Transfer Objects

==============================================================================
FILE: internal/dtos/employee.go
==============================================================================

DESCRIPTION:
    Defines request and response structures for employee management including
    creation, updates, termination, and salary changes. Contains all Mexican
    payroll-specific employee data fields.

USER PERSPECTIVE:
    - Shapes employee forms in the frontend
    - Contains all required Mexican payroll identifiers (RFC, CURP, NSS)
    - Handles employment status and salary management

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new employee fields
    ⚠️  CAUTION: Changing required fields (affects frontend validation)
    ❌  DO NOT modify: RFC/CURP/NSS formats without fiscal review
    📝  Keep validations aligned with SAT/IMSS requirements

SYNTAX EXPLANATION:
    - EmployeeRequest: Create/update input from frontend
    - EmployeeResponse: Full employee data returned to frontend
    - CollarType: white_collar (salaried), blue_collar (hourly), gray_collar
    - PayFrequency: weekly, biweekly, monthly payment schedule

MEXICAN PAYROLL FIELDS:
    - RFC: Tax ID (Registro Federal de Contribuyentes)
    - CURP: Population ID (Clave Única de Registro de Población)
    - NSS: Social Security Number (Número de Seguro Social)
    - CLABE: Bank account (Clave Bancaria Estandarizada)
    - InfonavitCredit: Housing credit deduction number

==============================================================================
*/
package dtos

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeRequest represents data for creating or updating an employee
type EmployeeRequest struct {
	EmployeeNumber        string     `json:"employee_number" binding:"required"`
	FirstName             string     `json:"first_name" binding:"required"`
	LastName              string     `json:"last_name" binding:"required"`
	MotherLastName        string     `json:"mother_last_name"`
	DateOfBirth           Date       `json:"date_of_birth" binding:"required"`
	Gender                string     `json:"gender" binding:"required,oneof=male female other"`
	MaritalStatus         string     `json:"marital_status"`
	RFC                   string     `json:"rfc" binding:"required"`
	CURP                  string     `json:"curp" binding:"required"`
	NSS                   string     `json:"nss"`
	InfonavitCredit       string     `json:"infonavit_credit"`
	PersonalEmail         string     `json:"personal_email,omitempty" binding:"omitempty,email"`
	PersonalPhone         string     `json:"personal_phone,omitempty"`
	EmergencyContact      string     `json:"emergency_contact,omitempty"`
	EmergencyPhone        string     `json:"emergency_phone,omitempty"`
	Street                string     `json:"street,omitempty"`
	ExteriorNumber        string     `json:"exterior_number,omitempty"`
	InteriorNumber        string     `json:"interior_number,omitempty"`
	Neighborhood          string     `json:"neighborhood,omitempty"`
	Municipality          string     `json:"municipality,omitempty"`
	State                 string     `json:"state,omitempty"`
	PostalCode            string     `json:"postal_code,omitempty"`
	Country               string     `json:"country,omitempty"`
	HireDate              Date       `json:"hire_date" binding:"required"`
	TerminationDate       *DatePtr   `json:"termination_date,omitempty"`
	EmploymentStatus      string     `json:"employment_status,omitempty" binding:"omitempty,oneof=active inactive on_leave terminated"`
	EmployeeType          string     `json:"employee_type,omitempty" binding:"omitempty,oneof=permanent temporary contractor intern"`
	CollarType            string     `json:"collar_type,omitempty" binding:"omitempty,oneof=white_collar blue_collar gray_collar"`
	PayFrequency          string     `json:"pay_frequency,omitempty" binding:"omitempty,oneof=weekly biweekly monthly"`
	IsSindicalizado       bool       `json:"is_sindicalizado"`
	DailySalary           float64    `json:"daily_salary" binding:"required,gt=0"`
	IntegratedDailySalary float64    `json:"integrated_daily_salary"`
	PaymentMethod         string     `json:"payment_method,omitempty" binding:"omitempty,oneof=bank_transfer cash check"`
	BankName              string     `json:"bank_name,omitempty"`
	BankAccount           string     `json:"bank_account,omitempty"`
	CLABE                 string     `json:"clabe,omitempty"`
	IMSSRegistrationDate  *DatePtr   `json:"imss_registration_date,omitempty"`
	Regime                string     `json:"regime,omitempty"`
	TaxRegime             string     `json:"tax_regime,omitempty"`
	DepartmentID          *uuid.UUID `json:"department_id,omitempty"`
	PositionID            *uuid.UUID `json:"position_id,omitempty"`
	CostCenterID          *uuid.UUID `json:"cost_center_id,omitempty"`
	ShiftID               *uuid.UUID `json:"shift_id,omitempty"`
	SupervisorID          *uuid.UUID `json:"supervisor_id,omitempty"`
}

// EmployeeResponse represents employee data returned in API responses
type EmployeeResponse struct {
	ID                    uuid.UUID  `json:"id"`
	EmployeeNumber        string     `json:"employee_number"`
	FirstName             string     `json:"first_name"`
	LastName              string     `json:"last_name"`
	MotherLastName        string     `json:"mother_last_name,omitempty"`
	FullName              string     `json:"full_name"`
	DateOfBirth           time.Time  `json:"date_of_birth"`
	Age                   int        `json:"age"`
	Gender                string     `json:"gender"`
	RFC                   string     `json:"rfc"`
	CURP                  string     `json:"curp"`
	NSS                   string     `json:"nss,omitempty"`
	InfonavitCredit       string     `json:"infonavit_credit,omitempty"`
	PersonalEmail         string     `json:"personal_email,omitempty"`
	PersonalPhone         string     `json:"personal_phone,omitempty"`
	EmergencyContact      string     `json:"emergency_contact,omitempty"`
	EmergencyPhone        string     `json:"emergency_phone,omitempty"`
	Street                string     `json:"street,omitempty"`
	ExteriorNumber        string     `json:"exterior_number,omitempty"`
	InteriorNumber        string     `json:"interior_number,omitempty"`
	Neighborhood          string     `json:"neighborhood,omitempty"`
	Municipality          string     `json:"municipality,omitempty"`
	State                 string     `json:"state,omitempty"`
	PostalCode            string     `json:"postal_code,omitempty"`
	Country               string     `json:"country,omitempty"`
	HireDate              time.Time  `json:"hire_date"`
	Seniority             float64    `json:"seniority"`
	TerminationDate       *time.Time `json:"termination_date,omitempty"`
	EmploymentStatus      string     `json:"employment_status"`
	EmployeeType          string     `json:"employee_type"`
	CollarType            string     `json:"collar_type"`
	PayFrequency          string     `json:"pay_frequency"`
	IsSindicalizado       bool       `json:"is_sindicalizado"`
	DailySalary           float64    `json:"daily_salary"`
	IntegratedDailySalary float64    `json:"integrated_daily_salary"`
	PaymentMethod         string     `json:"payment_method"`
	BankName              string     `json:"bank_name,omitempty"`
	BankAccount           string     `json:"bank_account,omitempty"`
	CLABE                 string     `json:"clabe,omitempty"`
	IMSSRegistrationDate  *time.Time `json:"imss_registration_date,omitempty"`
	Regime                string     `json:"regime,omitempty"`
	TaxRegime             string     `json:"tax_regime,omitempty"`
	DepartmentID          *uuid.UUID `json:"department_id,omitempty"`
	PositionID            *uuid.UUID `json:"position_id,omitempty"`
	CostCenterID          *uuid.UUID `json:"cost_center_id,omitempty"`
	ShiftID               *uuid.UUID `json:"shift_id,omitempty"`
	SupervisorID          *uuid.UUID `json:"supervisor_id,omitempty"`
	ShiftName             string     `json:"shift_name,omitempty"`
	SupervisorName        string     `json:"supervisor_name,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// EmployeeListResponse for listing employees
type EmployeeListResponse struct {
	Employees  []EmployeeResponse `json:"employees"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// EmployeeSearchRequest for searching employees
type EmployeeSearchRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Search       string `form:"search"`
	Status       string `form:"status"`
	EmployeeType string `form:"employee_type"`
	DepartmentID string `form:"department_id"`
}

// EmployeeTerminationRequest for terminating an employee
type EmployeeTerminationRequest struct {
	TerminationDate Date   `json:"termination_date" binding:"required"`
	Reason          string `json:"reason" binding:"required"`
	Comments        string `json:"comments,omitempty"`
}

// EmployeeSalaryUpdateRequest for updating an employee's salary
type EmployeeSalaryUpdateRequest struct {
	NewDailySalary float64 `json:"new_daily_salary" binding:"required,gt=0"`
	EffectiveDate  Date    `json:"effective_date" binding:"required"`
}

// =========================================================================
// Portal User Management DTOs
// =========================================================================

// CreatePortalUserRequest for creating a portal user account for an employee
type CreatePortalUserRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	Role             string `json:"role" binding:"required"` // employee, supervisor, manager, hr, hr_and_pr, sup_and_gm
	SupervisorID     string `json:"supervisor_id,omitempty"` // UUID of the supervisor user
	GeneralManagerID string `json:"general_manager_id,omitempty"` // UUID of the general manager user
	Department       string `json:"department,omitempty"`
	Area             string `json:"area,omitempty"`
}

// UpdatePortalUserRequest for updating a portal user account
type UpdatePortalUserRequest struct {
	Email            string `json:"email,omitempty"`
	Password         string `json:"password,omitempty"`
	Role             string `json:"role,omitempty"`
	IsActive         *bool  `json:"is_active,omitempty"`
	SupervisorID     string `json:"supervisor_id,omitempty"`
	GeneralManagerID string `json:"general_manager_id,omitempty"`
	Department       string `json:"department,omitempty"`
	Area             string `json:"area,omitempty"`
}

// PortalUserResponse for returning portal user information
type PortalUserResponse struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	Role             string     `json:"role"`
	IsActive         bool       `json:"is_active"`
	FullName         string     `json:"full_name"`
	SupervisorID     *string    `json:"supervisor_id,omitempty"`
	GeneralManagerID *string    `json:"general_manager_id,omitempty"`
	Department       string     `json:"department,omitempty"`
	Area             string     `json:"area,omitempty"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
