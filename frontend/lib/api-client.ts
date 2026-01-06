/**
 * @file lib/api-client.ts
 * @description Core API client for the IRIS Payroll System frontend. Provides centralized HTTP request handling,
 * authentication token management, error handling, and type-safe API endpoints for all backend communication.
 *
 * USER PERSPECTIVE:
 *   - Ensures secure authentication with automatic token refresh and session management
 *   - Provides user-friendly error messages for network issues and API failures
 *   - Automatically redirects to login when session expires (401 errors)
 *   - Handles file uploads/downloads for employee imports and payroll documents
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify:
 *     - Add new API endpoint functions (follow existing patterns in authApi, employeeApi, etc.)
 *     - Add new TypeScript interfaces for request/response types
 *     - Update error messages in getErrorMessage() for better UX
 *     - Add new query parameters or filters to existing endpoints
 *
 *   CAUTION:
 *     - apiRequest() function contains critical error handling logic - test thoroughly
 *     - API_BASE_URL configuration - ensure environment variable is properly set
 *     - File upload/download endpoints use different content-type headers (not JSON)
 *
 *   DO NOT modify:
 *     - Core authentication flow without updating auth.ts accordingly
 *     - ApiError class structure (used throughout the app for error handling)
 *     - Response type interfaces that match backend contracts (will break type safety)
 *     - credentials: 'include' setting - required for httpOnly cookie auth
 *     - localStorage key 'user' (used for UI display, non-sensitive data only)
 *
 * EXPORTS:
 *   - Type Interfaces: Employee, PayrollPeriod, PayrollConcept, PayrollSummary, AuthResponse, etc.
 *   - API Clients: authApi, employeeApi, payrollApi, catalogApi, reportApi, incidenceApi, evidenceApi, notificationApi
 *   - Error Handling: ApiError class with status codes and error categorization
 *   - Health Check: healthApi.isServerAvailable(), healthApi.getHealth()
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"

// Types
export interface Employee {
  id: string
  employee_number: string
  first_name: string
  last_name: string
  mother_last_name?: string
  full_name: string // Derived from backend
  date_of_birth: string // Use string for Date since JSON serialization
  age: number // Derived from backend
  gender: string
  rfc: string
  curp: string
  nss?: string
  infonavit_credit?: string
  personal_email?: string
  personal_phone?: string
  emergency_contact?: string
  emergency_phone?: string
  street?: string
  exterior_number?: string
  interior_number?: string
  neighborhood?: string
  municipality?: string
  state?: string
  postal_code?: string
  country?: string
  hire_date: string // Use string for Date
  seniority: number // Derived from backend
  termination_date?: string // Use string for Date
  employment_status: string // Mismatch from backend 'status'
  employee_type: string
  collar_type: string // white_collar, blue_collar, gray_collar
  pay_frequency: string // weekly, biweekly, monthly
  payment_frequency?: string // weekly, biweekly, monthly (alias)
  contract_type?: string // indefinite, temporary, training, seasonal
  is_sindicalizado: boolean // For blue collar unionized workers
  daily_salary: number
  integrated_daily_salary: number
  sdi?: number // Salario Diario Integrado
  payment_method: string
  bank_name?: string
  bank_account?: string
  clabe?: string
  imss_registration_date?: string // Use string for Date
  regime?: string
  tax_regime?: string
  department_id?: string // Use string for UUID
  position_id?: string // Use string for UUID
  cost_center_id?: string // Use string for UUID
  shift_id?: string // Use string for UUID
  supervisor_id?: string // Use string for UUID
  shift_name?: string // Derived from backend
  supervisor_name?: string // Derived from backend
  created_at: string // Use string for Date
  updated_at: string // Use string for Date
}

export interface PayrollPeriod {
  id: string
  year: number
  period_number: number
  start_date: string
  end_date: string
  payment_date: string
  frequency: string
  period_type: string
  period_code: string
  description?: string
  working_days?: number
  status: "open" | "calculated" | "approved" | "paid" | "closed" | "cancelled"
}

export interface PayrollConcept {
  id: string
  name: string
  category: "EARNING" | "DEDUCTION" | "EMPLOYER_CONTRIBUTION"
  input_type: string
  affects_tax_base: boolean
  affects_social_security: boolean
  affects_integrated_salary: boolean
}

export interface PayrollSummary {
  period_id: string
  total_gross: number
  total_deductions: number
  total_net: number
  employer_contributions: number
  employees: Array<{
    employee_id: string
    employee_name: string
    gross: number
    deductions: number
    net: number
    status: string
  }>
}

export interface PayrollCalculationResponse {
  id: string;
  employee_id: string;
  employee_name: string;
  employee_number: string;
  payroll_period_id: string;
  period_code: string;

  // Income
  regular_salary: number;
  overtime_amount: number;
  vacation_premium: number;
  aguinaldo: number;
  other_extras: number;

  // Deductions
  isr_withholding: number;
  imss_employee: number;
  infonavit_employee: number;
  retirement_savings: number;

  // Other deductions
  loan_deductions: number;
  advance_deductions: number;
  other_deductions: number;

  // Subsidies and benefits
  food_vouchers: number;
  savings_fund: number;
  employment_subsidy?: number;

  // Salary data
  sdi?: number;
  days_worked?: number;

  // Employer contribution amounts
  imss_employer?: number;
  infonavit_employer?: number;

  // Totals
  total_gross_income: number;
  total_deductions: number;
  total_net_pay: number;

  // Employer contributions
  employer_contributions: EmployerContributionResponse;

  // Metadata
  calculation_status: string;
  calculation_date: string; // Use string for Date
  payroll_status: string;
}

export interface EmployerContributionResponse {
  total_imss: number;
  total_infonavit: number;
  total_retirement: number;
  total_contributions: number;
}


// Custom error class for API errors with additional context
export class ApiError extends Error {
  public status: number;
  public isNetworkError: boolean;
  public isServerError: boolean;
  public isAuthError: boolean;

  constructor(message: string, status: number = 0, isNetworkError: boolean = false) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.isNetworkError = isNetworkError;
    this.isServerError = status >= 500;
    this.isAuthError = status === 401 || status === 403;
  }
}

// User-friendly error messages for common HTTP status codes
function getErrorMessage(status: number, fallbackMessage?: string): string {
  switch (status) {
    case 400:
      return fallbackMessage || 'Invalid request. Please check your input.';
    case 401:
      return 'Session expired. Please log in again.';
    case 403:
      return 'You do not have permission to perform this action.';
    case 404:
      return 'The requested resource was not found.';
    case 409:
      return fallbackMessage || 'A conflict occurred. The resource may already exist.';
    case 422:
      return fallbackMessage || 'Validation failed. Please check your input.';
    case 429:
      return 'Too many requests. Please wait a moment and try again.';
    case 500:
      return 'Server error. Please try again later.';
    case 502:
    case 503:
    case 504:
      return 'Service temporarily unavailable. Please try again later.';
    default:
      return fallbackMessage || `An error occurred (${status})`;
  }
}

// Helper to get CSRF token from cookie
function getCSRFToken(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : null;
}

// API request wrapper with comprehensive error handling
async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }

  // Add CSRF token for state-changing requests
  const method = options.method?.toUpperCase() || 'GET';
  if (['POST', 'PUT', 'DELETE', 'PATCH'].includes(method)) {
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }
  }

  let response: Response;

  try {
    response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
      credentials: 'include',
    })
  } catch (error) {
    // Network error - backend is unreachable (no internet, server down, CORS issue, etc.)
    console.error('Network error:', error);
    throw new ApiError(
      'Unable to connect to server. Please check your internet connection and try again.',
      0,
      true
    );
  }

  if (!response.ok) {
    // Try to extract error message from response body
    let errorMessage: string | undefined;
    try {
      const errorData = await response.json();
      // Prefer 'message' over 'error' since 'message' contains the detailed error
      // Backend returns: {"error": "Category", "message": "Detailed error description"}
      errorMessage = errorData.message || errorData.error || errorData.detail;
    } catch {
      // Response body is not JSON or empty
    }

    // Handle authentication errors
    if (response.status === 401) {
      // Only redirect if we're in a browser context and not on auth pages
      if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/auth')) {
        window.location.href = '/auth/login'
      }
    }

    throw new ApiError(
      getErrorMessage(response.status, errorMessage),
      response.status
    );
  }

  // Handle empty responses (e.g., 204 No Content)
  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    return {} as T;
  }

  try {
    return await response.json();
  } catch {
    // Response body is not valid JSON
    throw new ApiError('Invalid response from server', response.status);
  }
}

// Auth response interface matching backend
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: {
    id: string;
    email: string;
    role: string;
    full_name: string;
    is_active: boolean;
    created_at: string;
  };
}

// User profile interface
export interface UserProfile {
  id: string;
  email: string;
  role: string;
  full_name: string;
  is_active: boolean;
  created_at: string;
  company_id: string;
}

// Auth API
export const authApi = {
  login: (email: string, password: string) =>
    apiRequest<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  register: (companyData: any) =>
    apiRequest<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(companyData),
    }),

  refreshToken: (refreshToken: string) =>
    apiRequest<AuthResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),

  logout: () =>
    apiRequest<{ message: string }>('/auth/logout', {
      method: 'POST',
    }),

  changePassword: (currentPassword: string, newPassword: string) =>
    apiRequest<{ message: string }>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    }),

  forgotPassword: (email: string) =>
    apiRequest<{ message: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  resetPassword: (token: string, newPassword: string) =>
    apiRequest<{ message: string }>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, new_password: newPassword }),
    }),

  getProfile: () =>
    apiRequest<UserProfile>('/auth/profile'),

  updateProfile: (fullName: string) =>
    apiRequest<UserProfile>('/auth/profile', {
      method: 'PUT',
      body: JSON.stringify({ full_name: fullName }),
    }),
}

// User management types
export interface User {
  id: string;
  email: string;
  role: string;
  full_name: string;
  is_active: boolean;
  company_id: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  full_name: string;
  role: string;
}

export interface UpdateUserRequest {
  full_name?: string;
  role?: string;
  password?: string;
}

// User API (admin only)
export const userApi = {
  getUsers: () =>
    apiRequest<User[]>('/users'),

  createUser: (data: CreateUserRequest) =>
    apiRequest<User>('/users', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateUser: (id: string, data: UpdateUserRequest) =>
    apiRequest<User>(`/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteUser: (id: string) =>
    apiRequest<{ message: string }>(`/users/${id}`, {
      method: 'DELETE',
    }),

  toggleUserActive: (id: string) =>
    apiRequest<User>(`/users/${id}/toggle-active`, {
      method: 'PATCH',
    }),
}

// Employee stats interface
export interface EmployeeStats {
  total_employees: number;
  active_employees: number;
  inactive_employees: number;
  by_collar_type: { [key: string]: number };
  by_employee_type: { [key: string]: number };
  by_pay_frequency: { [key: string]: number };
}

// Employee termination request
export interface TerminationRequest {
  termination_date: string;
  reason: string;
  comments?: string;
}

// Salary update request
export interface SalaryUpdateRequest {
  new_daily_salary: number;
  effective_date: string;
}

// Employee list response from backend
export interface EmployeeListResponse {
  employees: Employee[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Employee API
export const employeeApi = {
  getEmployees: (filters?: { page?: number; page_size?: number; search?: string; status?: string }) =>
    apiRequest<EmployeeListResponse>(`/employees?page=${filters?.page || 1}&page_size=${filters?.page_size || 1000}`, {
      method: 'GET',
    }),

  getEmployee: (id: string) =>
    apiRequest<Employee>(`/employees/${id}`),

  createEmployee: (data: Partial<Employee> | Record<string, any>) =>
    apiRequest<Employee>('/employees', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateEmployee: (id: string, data: Partial<Employee> | Record<string, any>) =>
    apiRequest<Employee>(`/employees/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteEmployee: (id: string) =>
    apiRequest<{ success: boolean }>(`/employees/${id}`, {
      method: 'DELETE',
    }),

  terminateEmployee: (id: string, data: TerminationRequest) =>
    apiRequest<Employee>(`/employees/${id}/terminate`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateSalary: (id: string, data: SalaryUpdateRequest) =>
    apiRequest<Employee>(`/employees/${id}/salary`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  getStats: () =>
    apiRequest<EmployeeStats>('/employees/stats'),

  validateMexicanIds: (rfc?: string, curp?: string, nss?: string) => {
    const params = new URLSearchParams();
    if (rfc) params.append('rfc', rfc);
    if (curp) params.append('curp', curp);
    if (nss) params.append('nss', nss);
    return apiRequest<{ valid: boolean; errors: string[] }>(`/employees/validate-ids?${params.toString()}`, {
      method: 'POST',
    });
  },

  // Import employees from Excel/CSV file
  importEmployees: async (file: File): Promise<ImportEmployeesResponse> => {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch(`${API_BASE_URL}/employees/import`, {
      method: 'POST',
      body: formData,
      credentials: 'include'
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.message || `Import failed: ${response.status}`)
    }

    return response.json()
  },

  // Download import template
  downloadTemplate: async (): Promise<Blob> => {
    const response = await fetch(`${API_BASE_URL}/employees/import/template`, {
    credentials: 'include'
    })

    if (!response.ok) {
      throw new Error(`Download failed: ${response.status}`)
    }

    return response.blob()
  },

  // Portal User Management
  getPortalUser: (employeeId: string) =>
    apiRequest<PortalUserResponse>(`/employees/${employeeId}/portal-user`),

  createPortalUser: (employeeId: string, data: CreatePortalUserRequest) =>
    apiRequest<PortalUserResponse>(`/employees/${employeeId}/portal-user`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updatePortalUser: (employeeId: string, data: UpdatePortalUserRequest) =>
    apiRequest<PortalUserResponse>(`/employees/${employeeId}/portal-user`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deletePortalUser: (employeeId: string) =>
    apiRequest<{ success: boolean }>(`/employees/${employeeId}/portal-user`, {
      method: 'DELETE',
    }),
}

// Employee import response
export interface ImportError {
  error: string
  row: number
}

export interface ImportEmployeesResponse {
  message: string
  total_rows: number
  imported: number
  created: number
  updated: number
  failed: number
  errors: (string | ImportError)[]
}

// Portal User Management Types
export interface PortalUserResponse {
  id: string
  email: string
  role: string
  is_active: boolean
  full_name: string
  supervisor_id?: string
  general_manager_id?: string
  department?: string
  area?: string
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface CreatePortalUserRequest {
  email: string
  password: string
  role: string
  supervisor_id?: string
  general_manager_id?: string
  department?: string
  area?: string
}

export interface UpdatePortalUserRequest {
  email?: string
  password?: string
  role?: string
  is_active?: boolean
  supervisor_id?: string
  general_manager_id?: string
  department?: string
  area?: string
}

// Payroll calculation request
export interface PayrollCalculationRequest {
  employee_id: string;
  payroll_period_id: string;
  calculate_sdi?: boolean;
}

// Generate periods response
export interface GeneratePeriodsResponse {
  message: string
  periods: PayrollPeriod[]
  count: number
}

// Payroll API
export const payrollApi = {
  getPeriods: (filters?: any) =>
    apiRequest<PayrollPeriod[]>('/payroll/periods', {
      method: 'GET',
    }),

  getPeriod: (id: string) =>
    apiRequest<PayrollPeriod>(`/payroll/periods/${id}`),

  createPeriod: (data: any) =>
    apiRequest<PayrollPeriod>('/payroll/periods', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  generateCurrentPeriods: () =>
    apiRequest<GeneratePeriodsResponse>('/payroll/periods/generate', {
      method: 'POST',
    }),

  calculatePayroll: (employeeId: string, periodId: string, calculateSdi: boolean = false) =>
    apiRequest<PayrollCalculationResponse>('/payroll/calculate', {
      method: 'POST',
      body: JSON.stringify({
        employee_id: employeeId,
        payroll_period_id: periodId,
        calculate_sdi: calculateSdi
      }),
    }),

  bulkCalculatePayroll: (periodId: string, employeeIds?: string[], calculateAll: boolean = false) =>
    apiRequest<any>('/payroll/bulk-calculate', {
      method: 'POST',
      body: JSON.stringify({ payroll_period_id: periodId, employee_ids: employeeIds, calculate_all: calculateAll }),
    }),

  getPayrollCalculation: (periodId: string, employeeId: string) =>
    apiRequest<PayrollCalculationResponse>(`/payroll/calculation/${periodId}/${employeeId}`),

  approvePayroll: (periodId: string) =>
    apiRequest<{ message: string }>(`/payroll/approve/${periodId}`, {
      method: 'POST',
    }),

  getPayrollSummary: (periodId: string) =>
    apiRequest<PayrollSummary>(`/payroll/summary/${periodId}`),

  getPayslip: (periodId: string, employeeId: string, format: 'pdf' | 'xml' | 'html' = 'pdf') =>
    apiRequest<any>(`/payroll/payslip/${periodId}/${employeeId}?format=${format}`),

  getConceptTotals: (periodId: string) =>
    apiRequest<any[]>(`/payroll/concept-totals/${periodId}`),

  getPayrollByPeriod: (periodId: string) =>
    apiRequest<PayrollCalculationResponse[]>(`/payroll/period/${periodId}`),

  processPayment: (periodId: string) =>
    apiRequest<{ message: string }>(`/payroll/payment/${periodId}`, {
      method: 'POST',
    }),

  getPaymentStatus: (periodId: string) =>
    apiRequest<{ status: string }>(`/payroll/payment/${periodId}`),
}

// Catalog API
export const catalogApi = {
  getPayrollConcepts: () =>
    apiRequest<PayrollConcept[]>('/catalogs/concepts'),

  getIncidenceTypes: () =>
    apiRequest<any[]>('/catalogs/incidence-types'),

  createConcept: (data: Partial<PayrollConcept>) =>
    apiRequest<PayrollConcept>('/catalogs/concepts', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
}

// Report API
export const reportApi = {
  generateReport: (reportType: string, payrollPeriodId: string, format: string = "json") =>
    apiRequest<any>(`/reports/generate`, {
      method: 'POST',
      body: JSON.stringify({ report_type: reportType, payroll_period_id: payrollPeriodId, format: format }),
    }),

  getReportHistory: () =>
    apiRequest<any[]>('/reports/history'),
}

// Form Field Types for dynamic forms
export interface FormFieldOption {
  value: string;
  label: string;
}

export interface FormField {
  name: string;
  type: 'text' | 'textarea' | 'number' | 'date' | 'time' | 'boolean' | 'select' | 'multiselect' | 'shift_select';
  label: string;
  label_en?: string;
  required: boolean;
  min?: number;
  max?: number;
  step?: number;
  placeholder?: string;
  default_value?: unknown;
  options?: FormFieldOption[];
  display_order: number;
  help_text?: string;
}

export interface FormFieldsConfig {
  fields: FormField[];
}

// Incidence Category (parent grouping for incidence types)
export interface IncidenceCategory {
  id: string;
  name: string;
  code: string;
  description?: string;
  color?: string;
  icon?: string;
  display_order: number;
  is_requestable: boolean;
  is_system: boolean;
  is_active: boolean;
  incidence_types?: IncidenceType[];
  created_at: string;
  updated_at: string;
}

export interface CreateIncidenceCategoryRequest {
  name: string;
  code: string;
  description?: string;
  color?: string;
  icon?: string;
  display_order?: number;
  is_requestable?: boolean;
  is_active?: boolean;
}

// Incidence Types
export interface IncidenceType {
  id: string;
  name: string;
  category_id?: string;
  category: 'absence' | 'sick' | 'vacation' | 'overtime' | 'delay' | 'bonus' | 'deduction' | 'other';
  effect_type: 'positive' | 'negative' | 'neutral';
  is_calculated: boolean;
  calculation_method?: string;
  default_value: number;
  requires_evidence: boolean;
  description?: string;
  form_fields?: FormFieldsConfig;
  is_requestable: boolean;
  approval_flow: string;
  display_order: number;
  incidence_category?: IncidenceCategory;
  created_at: string;
  updated_at: string;
}

export interface CreateIncidenceTypeRequest {
  name: string;
  category_id?: string;
  category: string;
  effect_type: string;
  is_calculated: boolean;
  calculation_method?: string;
  default_value?: number;
  requires_evidence?: boolean;
  description?: string;
  form_fields?: FormFieldsConfig;
  is_requestable?: boolean;
  approval_flow?: string;
  display_order?: number;
}

// Incidence
export interface Incidence {
  id: string;
  employee_id: string;
  payroll_period_id: string;
  incidence_type_id: string;
  start_date: string;
  end_date: string;
  quantity: number;
  calculated_amount: number;
  comments?: string;
  status: 'pending' | 'approved' | 'rejected' | 'processed';
  approved_by?: string;
  approved_at?: string;
  employee?: Employee;
  payroll_period?: PayrollPeriod;
  incidence_type?: IncidenceType;
  created_at: string;
  updated_at: string;
}

export interface CreateIncidenceRequest {
  employee_id: string;
  payroll_period_id?: string;
  incidence_type_id: string;
  start_date: string;
  end_date: string;
  quantity: number;
  comments?: string;
}

export interface UpdateIncidenceRequest {
  start_date?: string;
  end_date?: string;
  quantity?: number;
  comments?: string;
  status?: string;
}

export interface VacationBalance {
  employee_id: string;
  years_of_service: number;
  entitled_days: number;
  used_days: number;
  pending_days: number;
  available_days: number;
  year: number;
}

export interface AbsenceSummary {
  employee_id: string;
  year: number;
  by_category: Array<{
    category: string;
    days: number;
    count: number;
  }>;
}

// Incidence Category API
export const incidenceCategoryApi = {
  getAll: () =>
    apiRequest<IncidenceCategory[]>('/incidence-categories'),

  get: (id: string) =>
    apiRequest<IncidenceCategory>(`/incidence-categories/${id}`),

  create: (data: CreateIncidenceCategoryRequest) =>
    apiRequest<IncidenceCategory>('/incidence-categories', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: CreateIncidenceCategoryRequest) =>
    apiRequest<IncidenceCategory>(`/incidence-categories/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    apiRequest<{ message: string }>(`/incidence-categories/${id}`, {
      method: 'DELETE',
    }),

  updateDisplayOrder: (id: string, displayOrder: number) =>
    apiRequest<{ message: string }>(`/incidence-categories/${id}/reorder`, {
      method: 'PATCH',
      body: JSON.stringify({ display_order: displayOrder }),
    }),
}

// Incidence Types API
export const incidenceTypeApi = {
  getAll: () =>
    apiRequest<IncidenceType[]>('/incidence-types'),

  getRequestable: () =>
    apiRequest<IncidenceType[]>('/requestable-incidence-types'),

  create: (data: CreateIncidenceTypeRequest) =>
    apiRequest<IncidenceType>('/incidence-types', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: CreateIncidenceTypeRequest) =>
    apiRequest<IncidenceType>(`/incidence-types/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    apiRequest<{ message: string }>(`/incidence-types/${id}`, {
      method: 'DELETE',
    }),
}

// Incidences API
export const incidenceApi = {
  getAll: (employeeId?: string, periodId?: string, status?: string) => {
    const params = new URLSearchParams();
    if (employeeId) params.append('employee_id', employeeId);
    if (periodId) params.append('period_id', periodId);
    if (status) params.append('status', status);
    const queryString = params.toString();
    return apiRequest<Incidence[]>(`/incidences${queryString ? `?${queryString}` : ''}`);
  },

  get: (id: string) =>
    apiRequest<Incidence>(`/incidences/${id}`),

  create: (data: CreateIncidenceRequest) =>
    apiRequest<Incidence>('/incidences', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: UpdateIncidenceRequest) =>
    apiRequest<Incidence>(`/incidences/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    apiRequest<{ message: string }>(`/incidences/${id}`, {
      method: 'DELETE',
    }),

  approve: (id: string) =>
    apiRequest<Incidence>(`/incidences/${id}/approve`, {
      method: 'POST',
    }),

  reject: (id: string) =>
    apiRequest<Incidence>(`/incidences/${id}/reject`, {
      method: 'POST',
    }),

  getByEmployee: (employeeId: string) =>
    apiRequest<Incidence[]>(`/employees/${employeeId}/incidences`),

  getVacationBalance: (employeeId: string) =>
    apiRequest<VacationBalance>(`/employees/${employeeId}/vacation-balance`),

  getAbsenceSummary: (employeeId: string, year?: number) => {
    const params = year ? `?year=${year}` : '';
    return apiRequest<AbsenceSummary>(`/employees/${employeeId}/absence-summary${params}`);
  },
}

// Incidence Evidence types
export interface IncidenceEvidence {
  id: string
  incidence_id: string
  file_name: string
  original_name: string
  content_type: string
  file_size: number
  file_path: string
  uploaded_by: string
  created_at: string
}

// Evidence/Upload API - routes use /evidence base path to avoid conflicts
export const evidenceApi = {
  // Upload evidence file for an incidence
  upload: async (incidenceId: string, file: File): Promise<{ message: string; evidence: IncidenceEvidence }> => {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch(`${API_BASE_URL}/evidence/incidence/${incidenceId}`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.message || `Upload failed: ${response.status}`)
    }

    return response.json()
  },

  // List all evidence for an incidence
  list: (incidenceId: string) =>
    apiRequest<IncidenceEvidence[]>(`/evidence/incidence/${incidenceId}`),

  // Get single evidence details
  get: (evidenceId: string) =>
    apiRequest<IncidenceEvidence>(`/evidence/${evidenceId}`),

  // Download evidence file
  download: async (evidenceId: string): Promise<Blob> => {
    const response = await fetch(`${API_BASE_URL}/evidence/${evidenceId}/download`, {
    credentials: 'include',
    })

    if (!response.ok) {
      throw new Error(`Download failed: ${response.status}`)
    }

    return response.blob()
  },

  // Delete evidence
  delete: (evidenceId: string) =>
    apiRequest<{ message: string }>(`/evidence/${evidenceId}`, {
      method: 'DELETE',
    }),
}

// Health check API for checking server availability
export interface HealthCheckResponse {
  status: string;
  timestamp: number;
  service: string;
}

export const healthApi = {
  /**
   * Check if the backend server is available and responding.
   * Returns true if the server is healthy, false otherwise.
   */
  isServerAvailable: async (): Promise<boolean> => {
    try {
      const response = await fetch(`${API_BASE_URL}/health`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
      });
      return response.ok;
    } catch {
      return false;
    }
  },

  /**
   * Get the full health check response from the server.
   * Throws an ApiError if the server is unavailable.
   */
  getHealth: (): Promise<HealthCheckResponse> =>
    apiRequest<HealthCheckResponse>('/health'),
}

// Notification types - matching backend models
export type NotificationType =
  | 'employee_created'
  | 'employee_updated'
  | 'incidence_created'
  | 'incidence_approved'
  | 'incidence_rejected'
  | 'payroll_calculated'
  | 'period_created'
  | 'user_created';

export interface NotificationUser {
  id: string;
  email: string;
  full_name: string;
}

export interface Notification {
  id: string;
  company_id: string;
  actor_user_id: string;
  target_user_id?: string;
  type: NotificationType;
  title: string;
  message: string;
  resource_type?: string;
  resource_id?: string;
  created_at: string;
  actor_user?: NotificationUser;
  read: boolean;
  read_at?: string;
}

export interface NotificationResponse {
  notifications: Notification[];
}

export interface UnreadCountResponse {
  unread_count: number;
}

// Notification API - Real implementation
export const notificationApi = {
  /**
   * Get recent notifications for the current user.
   * Excludes notifications created by the user themselves.
   */
  getNotifications: async (limit: number = 20): Promise<Notification[]> => {
    const response = await apiRequest<NotificationResponse>(`/notifications?limit=${limit}`);
    return response.notifications || [];
  },

  /**
   * Get count of unread notifications for the current user.
   */
  getUnreadCount: async (): Promise<number> => {
    const response = await apiRequest<UnreadCountResponse>('/notifications/unread-count');
    return response.unread_count || 0;
  },

  /**
   * Mark a specific notification as read.
   */
  markAsRead: async (notificationId: string): Promise<void> => {
    await apiRequest<{ message: string }>(`/notifications/${notificationId}/read`, {
      method: 'POST',
    });
  },

  /**
   * Mark all notifications as read for the current user.
   */
  markAllAsRead: async (): Promise<void> => {
    await apiRequest<{ message: string }>('/notifications/read-all', {
      method: 'POST',
    });
  },
};

// Legacy function exports for backward compatibility
export async function getNotifications(limit: number = 20): Promise<Notification[]> {
  return notificationApi.getNotifications(limit);
}

export async function markNotificationAsRead(id: string): Promise<{ success: boolean }> {
  await notificationApi.markAsRead(id);
  return { success: true };
}

export async function markAllNotificationsAsRead(): Promise<{ success: boolean }> {
  await notificationApi.markAllAsRead();
  return { success: true };
}

// =============================================================================
// CALENDAR API - HR Calendar Feature
// =============================================================================

export type CalendarEventType = 'absence' | 'incidence' | 'shift_change';

export interface CalendarEvent {
  id: string;
  event_type: CalendarEventType;
  title: string;
  description?: string;
  start_date: string;
  end_date: string;
  all_day: boolean;
  status: string;

  // Employee info
  employee_id: string;
  employee_name: string;
  employee_number: string;
  collar_type: string;
  department_id?: string;

  // Source info
  source_id: string;
  source_type: string;

  // Absence request fields
  request_type?: string;
  total_days?: number;
  approval_stage?: string;
  reason?: string;

  // Incidence fields
  category?: string;
  effect_type?: string;
  quantity?: number;
  calculated_amount?: number;
  incidence_type?: string;

  // Shift change fields
  shift_name?: string;
  shift_code?: string;
  shift_time?: string;
  original_shift?: string;

  created_at: string;
  updated_at: string;
}

export interface CalendarSummary {
  total_absences: number;
  total_incidences: number;
  total_shift_changes: number;
  pending: number;
  approved: number;
  declined: number;
  white_collar: number;
  blue_collar: number;
  gray_collar: number;
}

export interface CalendarEventsResponse {
  events: CalendarEvent[];
  total_count: number;
  summary: CalendarSummary;
}

export interface CalendarEmployee {
  id: string;
  employee_number: string;
  full_name: string;
  collar_type: string;
  department_id?: string;
  department_name?: string;
  color: string;
}

export interface CalendarEmployeesResponse {
  employees: CalendarEmployee[];
}

export interface CalendarEventFilters {
  start_date: string;
  end_date: string;
  employee_ids?: string[];
  collar_types?: string[];
  event_types?: CalendarEventType[];
  department_id?: string;
  status?: string;
}

// Calendar API
export const calendarApi = {
  /**
   * Get calendar events within a date range with optional filters
   */
  getEvents: async (filters: CalendarEventFilters): Promise<CalendarEventsResponse> => {
    const params = new URLSearchParams();
    params.append('start_date', filters.start_date);
    params.append('end_date', filters.end_date);

    filters.employee_ids?.forEach(id => params.append('employee_ids[]', id));
    filters.collar_types?.forEach(type => params.append('collar_types[]', type));
    filters.event_types?.forEach(type => params.append('event_types[]', type));

    if (filters.department_id) params.append('department_id', filters.department_id);
    if (filters.status) params.append('status', filters.status);

    return apiRequest<CalendarEventsResponse>(`/calendar/events?${params.toString()}`);
  },

  /**
   * Get employees with assigned colors for calendar display
   */
  getEmployees: async (collarTypes?: string[], departmentId?: string): Promise<CalendarEmployeesResponse> => {
    const params = new URLSearchParams();
    collarTypes?.forEach(type => params.append('collar_types[]', type));
    if (departmentId) params.append('department_id', departmentId);

    const queryString = params.toString();
    const url = queryString ? `/calendar/employees?${queryString}` : '/calendar/employees';
    return apiRequest<CalendarEmployeesResponse>(url);
  },
};

// ============================================================================
// Message/Inbox Types and API
// ============================================================================

export type MessageType = 'direct' | 'announcement_question' | 'system';
export type MessageStatus = 'unread' | 'read' | 'archived';

export interface Message {
  id: string;
  company_id: string;
  sender_id: string;
  recipient_id: string;
  subject: string;
  body: string;
  type: MessageType;
  status: MessageStatus;
  announcement_id?: string;
  parent_id?: string;
  read_at?: string;
  created_at: string;
  updated_at: string;
  sender?: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  };
  recipient?: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  };
  announcement?: {
    id: string;
    title: string;
    message: string;
  };
  replies?: Message[];
}

export interface MessagesResponse {
  messages: Message[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface SendMessageRequest {
  recipient_id: string;
  subject: string;
  body: string;
}

export interface ReplyRequest {
  body: string;
}

export interface RecipientSuggestionsResponse {
  users: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  }[];
}

// Message API
export const messageApi = {
  // Get inbox (received messages)
  getInbox: (page: number = 1, pageSize: number = 20, status: string = 'all') =>
    apiRequest<MessagesResponse>(`/messages?page=${page}&page_size=${pageSize}&status=${status}`),

  // Get sent messages
  getSent: (page: number = 1, pageSize: number = 20) =>
    apiRequest<MessagesResponse>(`/messages/sent?page=${page}&page_size=${pageSize}`),

  // Get single message
  get: (id: string) =>
    apiRequest<Message>(`/messages/${id}`),

  // Send a new message
  send: (data: SendMessageRequest) =>
    apiRequest<Message>('/messages', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Reply to a message
  reply: (id: string, data: ReplyRequest) =>
    apiRequest<Message>(`/messages/${id}/reply`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Mark message as read
  markAsRead: (id: string) =>
    apiRequest<{ message: string }>(`/messages/${id}/read`, {
      method: 'POST',
    }),

  // Mark message as unread
  markAsUnread: (id: string) =>
    apiRequest<{ message: string }>(`/messages/${id}/unread`, {
      method: 'POST',
    }),

  // Archive message
  archive: (id: string) =>
    apiRequest<{ message: string }>(`/messages/${id}/archive`, {
      method: 'POST',
    }),

  // Delete message
  delete: (id: string) =>
    apiRequest<{ message: string }>(`/messages/${id}`, {
      method: 'DELETE',
    }),

  // Get unread count
  getUnreadCount: () =>
    apiRequest<{ unread_count: number }>('/messages/unread-count'),

  // Get recipient suggestions
  getRecipients: (search: string = '') =>
    apiRequest<RecipientSuggestionsResponse>(`/messages/recipients?search=${encodeURIComponent(search)}`),
};

// ============================================================================
// Shift Management Types and API
// ============================================================================

export interface Shift {
  id: string;
  name: string;
  code: string;
  description?: string;
  start_time: string;
  end_time: string;
  break_minutes: number;
  break_start_time?: string;
  work_hours_per_day: number;
  work_days: string; // JSON array like "[1,2,3,4,5]"
  color: string;
  display_order: number;
  is_rest_day: boolean;
  is_active: boolean;
  is_night_shift: boolean;
  collar_types: string; // JSON array like '["white_collar", "blue_collar"]'
  company_id: string;
  created_at: string;
  updated_at: string;
}

export interface ShiftResponse extends Shift {
  employee_count: number;
}

export interface CreateShiftRequest {
  name: string;
  code: string;
  description?: string;
  start_time: string;
  end_time: string;
  break_minutes?: number;
  break_start_time?: string;
  work_hours_per_day?: number;
  work_days?: number[];
  color?: string;
  display_order?: number;
  is_active?: boolean;
  is_night_shift?: boolean;
  collar_types?: string[]; // e.g., ["white_collar", "blue_collar"]
}

export interface UpdateShiftRequest extends CreateShiftRequest {}

// Shift API
export const shiftApi = {
  // Get all shifts for the company
  getAll: () =>
    apiRequest<ShiftResponse[]>('/shifts'),

  // Get only active shifts
  getActive: () =>
    apiRequest<Shift[]>('/shifts/active'),

  // Get single shift by ID
  get: (id: string) =>
    apiRequest<Shift>(`/shifts/${id}`),

  // Create a new shift
  create: (data: CreateShiftRequest) =>
    apiRequest<Shift>('/shifts', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Update an existing shift
  update: (id: string, data: UpdateShiftRequest) =>
    apiRequest<Shift>(`/shifts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  // Delete a shift
  delete: (id: string) =>
    apiRequest<{ message: string }>(`/shifts/${id}`, {
      method: 'DELETE',
    }),

  // Toggle shift active status
  toggleActive: (id: string) =>
    apiRequest<Shift>(`/shifts/${id}/toggle`, {
      method: 'PATCH',
    }),
};

// ============================================================================
// Absence Request Types and API
// ============================================================================

export type RequestType =
  | 'PAID_LEAVE'
  | 'UNPAID_LEAVE'
  | 'VACATION'
  | 'LATE_ENTRY'
  | 'EARLY_EXIT'
  | 'SHIFT_CHANGE'
  | 'TIME_FOR_TIME'
  | 'SICK_LEAVE'
  | 'PERSONAL'
  | 'OTHER';

export type RequestStatus = 'PENDING' | 'APPROVED' | 'DECLINED' | 'ARCHIVED';

export type ApprovalStage =
  | 'SUPERVISOR'
  | 'MANAGER'
  | 'GENERAL_MANAGER'
  | 'HR'
  | 'HR_BLUE_GRAY'
  | 'PAYROLL'
  | 'COMPLETED';

export type ApprovalAction = 'APPROVED' | 'DECLINED';

export interface ApprovalHistory {
  id: string;
  request_id: string;
  approver_id: string;
  approval_stage: ApprovalStage;
  action: ApprovalAction;
  comments?: string;
  created_at: string;
  approver?: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  };
}

export interface AbsenceRequest {
  id: string;
  employee_id: string;
  request_type: RequestType;
  incidence_type_id?: string;
  start_date: string;
  end_date: string;
  total_days: number;
  reason: string;
  status: RequestStatus;
  current_approval_stage: ApprovalStage;
  hours_per_day?: number;
  paid_days?: number;
  unpaid_days?: number;
  unpaid_comments?: string;
  shift_details?: string;
  new_shift_id?: string;
  employee?: {
    id: string;
    employee_number: string;
    full_name: string;
    collar_type: string;
    department_id?: string;
  };
  approval_history?: ApprovalHistory[];
  created_at: string;
  updated_at: string;
}

export interface CreateAbsenceRequestRequest {
  employee_id: string;
  request_type: RequestType;
  start_date: string;
  end_date: string;
  total_days: number;
  reason: string;
  hours_per_day?: number;
  paid_days?: number;
  unpaid_days?: number;
  unpaid_comments?: string;
  shift_details?: string;
  new_shift_id?: string;
}

export interface ApproveRequestRequest {
  action: ApprovalAction;
  stage: ApprovalStage;
  comments?: string;
}

export interface CreateAbsenceRequestResponse {
  success: boolean;
  requestId: string;
  incidenceId: string;
}

export interface PendingCountsResponse {
  supervisor_count: number;
  general_manager_count: number;
  hr_blue_gray_count: number;
  hr_white_count: number;
}

// Absence Request API
export const absenceRequestApi = {
  // Create new absence request
  create: (data: CreateAbsenceRequestRequest) =>
    apiRequest<CreateAbsenceRequestResponse>('/absence-requests', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Get user's own requests
  getMyRequests: () =>
    apiRequest<{ requests: AbsenceRequest[] }>('/absence-requests/my-requests'),

  // Get pending requests by stage (supervisor, general_manager, hr_blue_gray, etc.)
  getPending: (stage: string) =>
    apiRequest<{ requests: AbsenceRequest[] }>(`/absence-requests/pending/${stage}`),

  // Get approved requests with optional filters
  getApproved: (filters?: { period_id?: string; employee_id?: string; collar_type?: string }) => {
    const params = new URLSearchParams();
    if (filters?.period_id) params.append('period_id', filters.period_id);
    if (filters?.employee_id) params.append('employee_id', filters.employee_id);
    if (filters?.collar_type) params.append('collar_type', filters.collar_type);
    const queryString = params.toString();
    return apiRequest<{ requests: AbsenceRequest[] }>(
      `/absence-requests/approved${queryString ? `?${queryString}` : ''}`
    );
  },

  // Approve or decline a request
  approve: (id: string, data: ApproveRequestRequest) =>
    apiRequest<{ success: boolean }>(`/absence-requests/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Delete a request (employee can delete their own pending requests)
  delete: (id: string) =>
    apiRequest<{ success: boolean; message: string }>(`/absence-requests/${id}`, {
      method: 'DELETE',
    }),

  // Archive a request
  archive: (id: string) =>
    apiRequest<{ success: boolean; message: string }>(`/absence-requests/${id}/archive`, {
      method: 'PATCH',
    }),

  // Get overlapping absences for validation
  getOverlapping: (startDate: string, endDate: string, excludeRequestId?: string) => {
    const params = new URLSearchParams();
    params.append('start_date', startDate);
    params.append('end_date', endDate);
    if (excludeRequestId) params.append('exclude_request_id', excludeRequestId);
    return apiRequest<AbsenceRequest[]>(`/absence-requests/overlapping?${params.toString()}`);
  },

  // Get pending counts for the current user (role-based)
  getCounts: () =>
    apiRequest<PendingCountsResponse>('/absence-requests/counts'),

  // Export approved requests to Excel
  exportApproved: async (filters?: { period_id?: string; employee_id?: string }): Promise<Blob> => {
    const params = new URLSearchParams();
    if (filters?.period_id) params.append('period_id', filters.period_id);
    if (filters?.employee_id) params.append('employee_id', filters.employee_id);

    const queryString = params.toString();
    const response = await fetch(
      `${API_BASE_URL}/absence-requests/export${queryString ? `?${queryString}` : ''}`,
      {
      credentials: 'include',
      }
    );

    if (!response.ok) {
      throw new Error(`Export failed: ${response.status}`);
    }

    return response.blob();
  },
};

// Audit Log Types
export interface AuditLog {
  id: string;
  event_type: string;
  user_id?: string;
  email: string;
  ip_address?: string;
  user_agent?: string;
  success: boolean;
  failure_reason?: string;
  page_url?: string;
  session_id?: string;
  duration?: number;
  metadata?: string;
  created_at: string;
  updated_at: string;
  user?: {
    id: string;
    email: string;
    full_name: string;
    role: string;
  };
}

export interface LoginSession {
  id: string;
  user_id: string;
  email: string;
  login_at: string;
  logout_at?: string;
  last_activity: string;
  ip_address?: string;
  user_agent?: string;
  session_id: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  user?: {
    id: string;
    email: string;
    full_name: string;
    role: string;
  };
}

export interface PageVisit {
  id: string;
  user_id: string;
  session_id: string;
  page_url: string;
  page_title?: string;
  entered_at: string;
  exited_at?: string;
  duration?: number;
  created_at: string;
  updated_at: string;
}

export interface AuditLogsResponse {
  data: AuditLog[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface LoginHistoryResponse {
  data: LoginSession[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface PageVisitsResponse {
  data: PageVisit[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface AuditStatsResponse {
  total_attempts: number;
  successful_attempts: number;
  failed_attempts: number;
  unique_users: number;
  failure_reasons: Record<string, number>;
  success_rate: number;
}

// ============================================================================
// Announcement Types and API
// ============================================================================

export interface Announcement {
  id: string;
  company_id: string;
  title: string;
  message: string;
  scope: 'ALL' | 'TEAM';
  image_data?: string;
  created_by: string;
  expires_at?: string;
  is_read?: boolean;
  created_at: string;
  updated_at: string;
  creator?: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  };
}

export interface CreateAnnouncementRequest {
  title: string;
  message: string;
  scope: 'ALL' | 'TEAM';
  image_base64?: string;
  expires_in_days?: number;
}

export interface AnnouncementsResponse {
  data: Announcement[];
}

// Announcement API
export const announcementApi = {
  // Get all active announcements
  getAll: async (): Promise<Announcement[]> => {
    const response = await apiRequest<AnnouncementsResponse>('/announcements');
    return response.data || [];
  },

  // Get single announcement
  get: (id: string) =>
    apiRequest<Announcement>(`/announcements/${id}`),

  // Create new announcement
  create: (data: CreateAnnouncementRequest) =>
    apiRequest<Announcement>('/announcements', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Delete announcement
  delete: (id: string) =>
    apiRequest<{ message: string }>(`/announcements/${id}`, {
      method: 'DELETE',
    }),

  // Mark announcement as read
  markAsRead: (id: string) =>
    apiRequest<{ message: string }>(`/announcements/${id}/read`, {
      method: 'POST',
    }),

  // Get unread count
  getUnreadCount: async (): Promise<number> => {
    const response = await apiRequest<{ count: number }>('/announcements/unread-count');
    return response.count || 0;
  },
};

// Audit Log API
export const auditApi = {
  // Get audit logs with filters (admin only)
  getLogs: (filters?: {
    page?: number;
    page_size?: number;
    user_id?: string;
    email?: string;
    event_type?: string;
    success?: boolean;
    start_date?: string;
    end_date?: string;
  }) => {
    const params = new URLSearchParams();
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.page_size) params.append('page_size', filters.page_size.toString());
    if (filters?.user_id) params.append('user_id', filters.user_id);
    if (filters?.email) params.append('email', filters.email);
    if (filters?.event_type) params.append('event_type', filters.event_type);
    if (filters?.success !== undefined) params.append('success', filters.success.toString());
    if (filters?.start_date) params.append('start_date', filters.start_date);
    if (filters?.end_date) params.append('end_date', filters.end_date);

    const queryString = params.toString();
    return apiRequest<AuditLogsResponse>(
      `/audit/logs${queryString ? `?${queryString}` : ''}`
    );
  },

  // Get login attempts (admin only)
  getLoginAttempts: (page?: number, pageSize?: number) => {
    const params = new URLSearchParams();
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());

    const queryString = params.toString();
    return apiRequest<AuditLogsResponse>(
      `/audit/login-attempts${queryString ? `?${queryString}` : ''}`
    );
  },

  // Get login history (admin can see all, users see their own)
  getLoginHistory: (userId?: string, page?: number, pageSize?: number) => {
    const params = new URLSearchParams();
    if (userId) params.append('user_id', userId);
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());

    const queryString = params.toString();
    return apiRequest<LoginHistoryResponse>(
      `/audit/login-history${queryString ? `?${queryString}` : ''}`
    );
  },

  // Get active sessions
  getActiveSessions: (userId?: string) => {
    const params = new URLSearchParams();
    if (userId) params.append('user_id', userId);

    const queryString = params.toString();
    return apiRequest<{ data: LoginSession[] }>(
      `/audit/active-sessions${queryString ? `?${queryString}` : ''}`
    );
  },

  // Get page visits
  getPageVisits: (userId?: string, sessionId?: string, page?: number, pageSize?: number) => {
    const params = new URLSearchParams();
    if (userId) params.append('user_id', userId);
    if (sessionId) params.append('session_id', sessionId);
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());

    const queryString = params.toString();
    return apiRequest<PageVisitsResponse>(
      `/audit/page-visits${queryString ? `?${queryString}` : ''}`
    );
  },

  // Get global statistics (admin only)
  getStats: (days?: number) => {
    const params = new URLSearchParams();
    if (days) params.append('days', days.toString());

    const queryString = params.toString();
    return apiRequest<AuditStatsResponse>(
      `/audit/stats${queryString ? `?${queryString}` : ''}`
    );
  },

  // Get user activity statistics
  getUserStats: (userId: string, days?: number) => {
    const params = new URLSearchParams();
    if (days) params.append('days', days.toString());

    const queryString = params.toString();
    return apiRequest<{
      total_logins: number;
      total_page_visits: number;
      avg_session_duration_seconds: number;
    }>(`/audit/user/${userId}/stats${queryString ? `?${queryString}` : ''}`);
  },
};

// ============================================================================
// Payroll Export API - Dual Excel export for payroll processing
// ============================================================================

export interface ExportPreview {
  vacaciones_count: number;
  faltas_extras_count: number;
  late_approval_count: number;
  total_incidences: number;
  payroll_period_id: string;
}

export const payrollExportApi = {
  // Get export preview (counts and warnings)
  getPreview: (periodId: string) =>
    apiRequest<ExportPreview>(`/payroll-export/preview/${periodId}`),

  // Download dual Excel export as ZIP
  downloadDualExport: async (periodId: string): Promise<Blob> => {
    const response = await fetch(`${API_BASE_URL}/payroll-export/dual/${periodId}`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new ApiError(
        getErrorMessage(response.status, 'Failed to download export'),
        response.status
      );
    }

    return response.blob();
  },
};

// ============================================================================
// Incidence Configuration API - Tipo mappings for payroll export
// ============================================================================

export interface IncidenceTipoMapping {
  id: string;
  incidence_type_id: string;
  tipo_code?: string | null;
  motivo?: string | null;
  template_type: 'vacaciones' | 'faltas_extras';
  hours_multiplier: number;
  notes?: string | null;
  created_at: string;
  updated_at: string;
  incidence_type?: IncidenceType;
}

export interface CreateTipoMappingRequest {
  incidence_type_id: string;
  tipo_code?: string | null;
  motivo?: 'EXTRAS' | 'FALTA' | 'PERHORAS' | null;
  template_type: 'vacaciones' | 'faltas_extras';
  hours_multiplier: number;
  notes?: string | null;
}

export interface UpdateTipoMappingRequest {
  tipo_code?: string | null;
  motivo?: 'EXTRAS' | 'FALTA' | 'PERHORAS' | null;
  template_type?: 'vacaciones' | 'faltas_extras';
  hours_multiplier?: number;
  notes?: string | null;
}

export const incidenceConfigApi = {
  // Get all tipo mappings
  getAllMappings: () =>
    apiRequest<IncidenceTipoMapping[]>('/incidence-config/tipo-mappings'),

  // Get single mapping by ID
  getMapping: (id: string) =>
    apiRequest<IncidenceTipoMapping>(`/incidence-config/tipo-mappings/${id}`),

  // Create new mapping
  createMapping: (data: CreateTipoMappingRequest) =>
    apiRequest<IncidenceTipoMapping>('/incidence-config/tipo-mappings', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Update mapping
  updateMapping: (id: string, data: UpdateTipoMappingRequest) =>
    apiRequest<IncidenceTipoMapping>(`/incidence-config/tipo-mappings/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  // Delete mapping
  deleteMapping: (id: string) =>
    apiRequest<{ message: string }>(`/incidence-config/tipo-mappings/${id}`, {
      method: 'DELETE',
    }),

  // Get unmapped incidence types
  getUnmappedTypes: () =>
    apiRequest<IncidenceType[]>('/incidence-config/unmapped-types'),
};

// ============================================================================
// Role Inheritance API - Role permission inheritance configuration
// ============================================================================

export interface RoleInheritance {
  id: string;
  child_role: string;
  parent_role: string;
  is_active: boolean;
  priority: number;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleInheritanceRequest {
  child_role: string;
  parent_role: string;
  priority?: number;
  notes?: string;
}

export interface UpdateRoleInheritanceRequest {
  is_active: boolean;
  priority?: number;
  notes?: string;
}

export interface ValidRole {
  name: string;
  level: number;
}

export const roleInheritanceApi = {
  // Get all role inheritances
  getAllInheritances: () =>
    apiRequest<RoleInheritance[]>('/role-inheritance'),

  // Get role hierarchy tree
  getHierarchy: () =>
    apiRequest<Record<string, string[]>>('/role-inheritance/hierarchy'),

  // Resolve inherited roles for a specific role
  resolveRoles: (role: string) =>
    apiRequest<{ role: string; inherited_roles: string[] }>(`/role-inheritance/resolve/${role}`),

  // Get list of valid roles
  getValidRoles: () =>
    apiRequest<{ roles: ValidRole[] }>('/role-inheritance/valid-roles'),

  // Create new inheritance
  createInheritance: (data: CreateRoleInheritanceRequest) =>
    apiRequest<RoleInheritance>('/role-inheritance', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Update inheritance
  updateInheritance: (id: string, data: UpdateRoleInheritanceRequest) =>
    apiRequest<RoleInheritance>(`/role-inheritance/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  // Delete inheritance
  deleteInheritance: (id: string) =>
    apiRequest<{ message: string }>(`/role-inheritance/${id}`, {
      method: 'DELETE',
    }),
};

// ============================================================================
// Permission Matrix API - Role-based access control configuration
// ============================================================================

export interface PermissionSet {
  can_view: boolean;
  can_create: boolean;
  can_edit: boolean;
  can_delete: boolean;
  can_export: boolean;
  can_approve: boolean;
}

export interface Permission {
  id: string;
  role: string;
  resource: string;
  permissions: PermissionSet;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreatePermissionRequest {
  role: string;
  resource: string;
  permissions: PermissionSet;
  description?: string;
}

export interface UpdatePermissionRequest {
  permissions: PermissionSet;
  description?: string;
  is_active: boolean;
}

export const permissionApi = {
  // Get all permissions
  getAllPermissions: () =>
    apiRequest<Permission[]>('/permissions'),

  // Get permissions for a specific role
  getPermissionsByRole: (role: string) =>
    apiRequest<Permission[]>(`/permissions/role/${role}`),

  // Get specific permission
  getPermission: (id: string) =>
    apiRequest<Permission>(`/permissions/${id}`),

  // Get list of valid roles
  getRoles: () =>
    apiRequest<{ roles: string[] }>('/permissions/roles'),

  // Get list of valid resources
  getResources: () =>
    apiRequest<{ resources: string[] }>('/permissions/resources'),

  // Create new permission
  createPermission: (data: CreatePermissionRequest) =>
    apiRequest<Permission>('/permissions', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Update permission
  updatePermission: (id: string, data: UpdatePermissionRequest) =>
    apiRequest<Permission>(`/permissions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  // Delete permission
  deletePermission: (id: string) =>
    apiRequest<{ message: string }>(`/permissions/${id}`, {
      method: 'DELETE',
    }),
};
