/**
 * @file lib/api/employees.ts
 * @description Employee API service layer for the IRIS Payroll System. Provides high-level functions
 * for employee CRUD operations with data mapping between frontend form format and backend API format.
 * Acts as an adapter layer between UI components and the core API client.
 *
 * USER PERSPECTIVE:
 *   - Enables creating, viewing, updating, and deleting employee records
 *   - Handles data validation and error messages from the backend
 *   - Ensures employee data is correctly formatted for Mexican payroll compliance
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify:
 *     - Update mapToBackendEmployee() when form fields change
 *     - Add new employee operations (terminate, update salary, etc.)
 *     - Enhance error messages for better user feedback
 *     - Add data validation before sending to backend
 *
 *   CAUTION:
 *     - mapToBackendEmployee() must stay in sync with backend Employee model
 *     - Field name mappings (e.g., firstName -> first_name) are critical for data integrity
 *     - All functions return ApiResponse<T> wrapper for consistent error handling
 *     - Default values (e.g., country: 'México') should match business rules
 *
 *   DO NOT modify:
 *     - ApiResponse interface structure without updating all consuming components
 *     - Employee type interface (defined in api-client.ts) without backend coordination
 *     - Remove required field mappings (causes data loss or validation errors)
 *
 * EXPORTS:
 *   - createEmployee(): Creates new employee with form data mapping
 *   - getEmployees(): Fetches all employees
 *   - getEmployee(): Fetches single employee by ID
 *   - updateEmployee(): Updates employee with form data mapping
 *   - deleteEmployee(): Deletes employee by ID
 *   - ApiResponse<T>: Standardized response wrapper with success/error handling
 */

import { employeeApi, Employee } from '../api-client'

interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

// Map frontend form data to backend expected format
// Handles both camelCase (form) and snake_case (already converted) formats
function mapToBackendEmployee(data: any): Record<string, any> {
  return {
    employee_number: data.employee_number || data.employeeNumber || `EMP-${Date.now()}`,
    first_name: data.first_name || data.firstName,
    last_name: data.last_name || data.lastName,
    mother_last_name: data.mother_last_name || data.secondLastName || data.motherLastName,
    date_of_birth: data.date_of_birth || data.birthDate || data.dateOfBirth,
    gender: (data.gender || 'other')?.toLowerCase(),
    marital_status: data.marital_status || data.maritalStatus,
    nationality: data.nationality,
    birth_state: data.birth_state || data.birthState,
    education_level: data.education_level || data.educationLevel,
    rfc: data.rfc,
    curp: data.curp,
    nss: data.nss,
    infonavit_credit: data.infonavit_credit || data.infonavitCredit,
    personal_email: data.personal_email || data.email || data.personalEmail,
    personal_phone: data.personal_phone || data.phone || data.personalPhone,
    cell_phone: data.cell_phone || data.cellPhone,
    emergency_contact: data.emergency_contact || data.emergencyContactName || data.emergencyContact,
    emergency_phone: data.emergency_phone || data.emergencyContactPhone || data.emergencyPhone,
    emergency_relationship: data.emergency_relationship || data.emergencyRelationship,
    street: data.address?.street || data.street,
    exterior_number: data.exterior_number || data.address?.exteriorNumber || data.exteriorNumber,
    interior_number: data.interior_number || data.address?.interiorNumber || data.interiorNumber,
    neighborhood: data.address?.neighborhood || data.neighborhood,
    municipality: data.address?.municipality || data.municipality,
    state: data.address?.state || data.state,
    postal_code: data.postal_code || data.address?.postalCode || data.postalCode,
    country: data.address?.country || data.country || 'México',
    hire_date: data.hire_date || data.hireDate,
    termination_date: data.termination_date || data.terminationDate,
    employment_status: (data.employment_status || data.status || 'active')?.toLowerCase(),
    employee_type: (data.employee_type || data.employmentType || 'permanent')?.toLowerCase(),
    collar_type: data.collar_type || data.collarType,
    pay_frequency: data.pay_frequency || data.payFrequency,
    is_sindicalizado: data.is_sindicalizado ?? data.isSindicalizado ?? false,
    contract_start_date: data.contract_start_date || data.contractStartDate,
    contract_end_date: data.contract_end_date || data.contractEndDate,
    department_id: data.department_id || data.departmentId,
    position_id: data.position_id || data.positionId,
    cost_center_id: data.cost_center_id || data.costCenterId,
    production_area: data.production_area || data.productionArea,
    location: data.location,
    patronal_registry: data.patronal_registry || data.patronalRegistry,
    company_name: data.company_name || data.companyName,
    shift_id: data.shift_id || data.shiftId,
    supervisor_id: data.supervisor_id || data.supervisorId,
    team_name: data.team_name || data.teamName,
    package_code: data.package_code || data.packageCode,
    route: data.route,
    transport_stop: data.transport_stop || data.transportStop,
    recruitment_source: data.recruitment_source || data.recruitmentSource,
    daily_salary: data.daily_salary || data.dailySalary || 0,
    integrated_daily_salary: data.integrated_daily_salary || data.sdi || data.integratedDailySalary,
    payment_method: (data.payment_method || data.paymentMethod || 'bank_transfer')?.toLowerCase(),
    bank_name: data.bank_name || data.bankName,
    bank_account: data.bank_account || data.accountNumber || data.bankAccount,
    clabe: data.clabe,
    imss_registration_date: data.imss_registration_date || data.imssRegistrationDate,
    regime: data.regime,
    tax_regime: data.tax_regime || data.taxRegime,
    fiscal_postal_code: data.fiscal_postal_code || data.fiscalPostalCode,
    fiscal_name: data.fiscal_name || data.fiscalName,
    child1_gender: data.child1_gender || data.child1Gender,
    child2_gender: data.child2_gender || data.child2Gender,
    child3_gender: data.child3_gender || data.child3Gender,
    child4_gender: data.child4_gender || data.child4Gender,
  }
}

export async function createEmployee(data: any): Promise<ApiResponse<Employee>> {
  try {
    const backendData = mapToBackendEmployee(data)
    const result = await employeeApi.createEmployee(backendData)
    return {
      success: true,
      data: result,
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Error creating employee',
    }
  }
}

export async function getEmployees(): Promise<ApiResponse<Employee[]>> {
  try {
    const result = await employeeApi.getEmployees()
    return {
      success: true,
      data: result.employees || [],
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Error fetching employees',
    }
  }
}

export async function getEmployee(id: string): Promise<ApiResponse<Employee>> {
  try {
    const result = await employeeApi.getEmployee(id)
    return {
      success: true,
      data: result,
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Error fetching employee',
    }
  }
}

export async function updateEmployee(id: string, data: any): Promise<ApiResponse<Employee>> {
  try {
    const backendData = mapToBackendEmployee(data)
    const result = await employeeApi.updateEmployee(id, backendData)
    return {
      success: true,
      data: result,
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Error updating employee',
    }
  }
}

export async function deleteEmployee(id: string): Promise<ApiResponse<void>> {
  try {
    await employeeApi.deleteEmployee(id)
    return {
      success: true,
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Error deleting employee',
    }
  }
}
