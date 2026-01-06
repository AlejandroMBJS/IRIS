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
function mapToBackendEmployee(data: any): Partial<Employee> {
  return {
    employee_number: data.employeeNumber || `EMP-${Date.now()}`,
    first_name: data.firstName,
    last_name: data.lastName,
    mother_last_name: data.secondLastName || data.motherLastName,
    date_of_birth: data.birthDate || data.dateOfBirth,
    gender: data.gender?.toLowerCase() || 'other',
    rfc: data.rfc,
    curp: data.curp,
    nss: data.nss,
    personal_email: data.email || data.personalEmail,
    personal_phone: data.phone || data.personalPhone,
    emergency_contact: data.emergencyContactName || data.emergencyContact,
    emergency_phone: data.emergencyContactPhone || data.emergencyPhone,
    street: data.address?.street || data.street,
    exterior_number: data.address?.exteriorNumber || data.exteriorNumber,
    interior_number: data.address?.interiorNumber || data.interiorNumber,
    neighborhood: data.address?.neighborhood || data.neighborhood,
    municipality: data.address?.municipality || data.municipality,
    state: data.address?.state || data.state,
    postal_code: data.address?.postalCode || data.postalCode,
    country: data.address?.country || data.country || 'México',
    hire_date: data.hireDate,
    employment_status: data.status?.toLowerCase() || 'active',
    employee_type: data.employmentType?.toLowerCase() || 'permanent',
    daily_salary: data.dailySalary || data.sdi || 0,
    integrated_daily_salary: data.sdi || data.integratedDailySalary,
    payment_method: data.paymentMethod?.toLowerCase() || 'bank_transfer',
    bank_name: data.bankName,
    bank_account: data.accountNumber || data.bankAccount,
    clabe: data.clabe,
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
