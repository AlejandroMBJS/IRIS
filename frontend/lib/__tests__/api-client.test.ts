/**
 * @file lib/__tests__/api-client.test.ts
 * @description Unit tests for the API client module
 */
import {
  ApiError,
  authApi,
  userApi,
  employeeApi,
  payrollApi,
  catalogApi,
  reportApi,
  healthApi,
} from '../api-client'

// Mock fetch globally
const mockFetch = jest.fn()
global.fetch = mockFetch

// Helper to create mock responses
function createMockResponse(data: any, status = 200, ok = true) {
  return {
    ok,
    status,
    headers: {
      get: (name: string) => {
        if (name === 'content-type') return 'application/json'
        return null
      },
    },
    json: jest.fn().mockResolvedValue(data),
    blob: jest.fn().mockResolvedValue(new Blob()),
  }
}

function createMockErrorResponse(status: number, errorData: any = {}) {
  return {
    ok: false,
    status,
    headers: {
      get: () => 'application/json',
    },
    json: jest.fn().mockResolvedValue(errorData),
  }
}

describe('api-client.ts', () => {
  beforeEach(() => {
    mockFetch.mockClear()
    // Mock document.cookie for CSRF token
    Object.defineProperty(document, 'cookie', {
      writable: true,
      value: 'csrf_token=test-csrf-token',
    })
  })

  // ============================================================================
  // ApiError Class Tests
  // ============================================================================
  describe('ApiError', () => {
    it('creates error with default values', () => {
      const error = new ApiError('Test error')

      expect(error.message).toBe('Test error')
      expect(error.name).toBe('ApiError')
      expect(error.status).toBe(0)
      expect(error.isNetworkError).toBe(false)
      expect(error.isServerError).toBe(false)
      expect(error.isAuthError).toBe(false)
    })

    it('creates error with custom status', () => {
      const error = new ApiError('Not found', 404)

      expect(error.status).toBe(404)
      expect(error.isServerError).toBe(false)
      expect(error.isAuthError).toBe(false)
    })

    it('identifies server errors (5xx)', () => {
      const error500 = new ApiError('Server error', 500)
      const error502 = new ApiError('Bad gateway', 502)
      const error503 = new ApiError('Unavailable', 503)

      expect(error500.isServerError).toBe(true)
      expect(error502.isServerError).toBe(true)
      expect(error503.isServerError).toBe(true)
    })

    it('identifies auth errors (401, 403)', () => {
      const error401 = new ApiError('Unauthorized', 401)
      const error403 = new ApiError('Forbidden', 403)

      expect(error401.isAuthError).toBe(true)
      expect(error403.isAuthError).toBe(true)
    })

    it('identifies network errors', () => {
      const error = new ApiError('Network error', 0, true)

      expect(error.isNetworkError).toBe(true)
      expect(error.status).toBe(0)
    })

    it('is instance of Error', () => {
      const error = new ApiError('Test')
      expect(error instanceof Error).toBe(true)
    })
  })

  // ============================================================================
  // Auth API Tests
  // ============================================================================
  describe('authApi', () => {
    describe('login', () => {
      it('sends login request with credentials', async () => {
        const mockResponse = {
          access_token: 'mock-token',
          refresh_token: 'mock-refresh',
          user: { id: '123', email: 'test@example.com', role: 'admin' },
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse))

        const result = await authApi.login('test@example.com', 'password123')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/login'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ email: 'test@example.com', password: 'password123' }),
            credentials: 'include',
          })
        )
        expect(result.access_token).toBe('mock-token')
      })

      it('includes CSRF token in header for POST request', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ access_token: 'token' }))

        await authApi.login('test@example.com', 'password')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.any(String),
          expect.objectContaining({
            headers: expect.objectContaining({
              'X-CSRF-Token': 'test-csrf-token',
            }),
          })
        )
      })

      it('throws ApiError on failed login', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockErrorResponse(401, { message: 'Invalid credentials' })
        )

        await expect(authApi.login('test@example.com', 'wrong')).rejects.toThrow(ApiError)
      })
    })

    describe('register', () => {
      it('sends registration request', async () => {
        const mockResponse = {
          access_token: 'new-token',
          user: { id: '456', email: 'new@company.com' },
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse))

        const companyData = {
          company_name: 'New Company',
          email: 'new@company.com',
          password: 'SecurePass123!',
        }

        const result = await authApi.register(companyData)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/register'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(companyData),
          })
        )
        expect(result.access_token).toBe('new-token')
      })
    })

    describe('logout', () => {
      it('sends logout request', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ message: 'Logged out' }))

        const result = await authApi.logout()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/logout'),
          expect.objectContaining({ method: 'POST' })
        )
        expect(result.message).toBe('Logged out')
      })
    })

    describe('refreshToken', () => {
      it('sends refresh token request', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ access_token: 'new-access-token' })
        )

        await authApi.refreshToken('old-refresh-token')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/refresh'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ refresh_token: 'old-refresh-token' }),
          })
        )
      })
    })

    describe('changePassword', () => {
      it('sends change password request', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ message: 'Password changed' }))

        await authApi.changePassword('oldPass', 'newPass')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/change-password'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ current_password: 'oldPass', new_password: 'newPass' }),
          })
        )
      })
    })

    describe('getProfile', () => {
      it('fetches user profile', async () => {
        const mockProfile = {
          id: '123',
          email: 'user@example.com',
          role: 'admin',
          full_name: 'Test User',
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockProfile))

        const result = await authApi.getProfile()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/profile'),
          expect.objectContaining({
            headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
          })
        )
        expect(result.email).toBe('user@example.com')
      })
    })
  })

  // ============================================================================
  // User API Tests
  // ============================================================================
  describe('userApi', () => {
    describe('getUsers', () => {
      it('fetches user list', async () => {
        const mockUsers = [
          { id: '1', email: 'user1@test.com', role: 'admin' },
          { id: '2', email: 'user2@test.com', role: 'hr' },
        ]
        mockFetch.mockResolvedValueOnce(createMockResponse(mockUsers))

        const result = await userApi.getUsers()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/users'),
          expect.any(Object)
        )
        expect(result).toHaveLength(2)
      })
    })

    describe('createUser', () => {
      it('creates a new user', async () => {
        const newUser = {
          email: 'newuser@test.com',
          password: 'password123',
          full_name: 'New User',
          role: 'hr',
        }
        mockFetch.mockResolvedValueOnce(createMockResponse({ id: '3', ...newUser }))

        const result = await userApi.createUser(newUser)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/users'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(newUser),
          })
        )
        expect(result.email).toBe('newuser@test.com')
      })
    })

    describe('updateUser', () => {
      it('updates existing user', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ id: '1', full_name: 'Updated Name' })
        )

        await userApi.updateUser('1', { full_name: 'Updated Name' })

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/users/1'),
          expect.objectContaining({
            method: 'PUT',
            body: JSON.stringify({ full_name: 'Updated Name' }),
          })
        )
      })
    })

    describe('deleteUser', () => {
      it('deletes user', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ message: 'User deleted' }))

        await userApi.deleteUser('1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/users/1'),
          expect.objectContaining({ method: 'DELETE' })
        )
      })
    })

    describe('toggleUserActive', () => {
      it('toggles user active status', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ id: '1', is_active: false })
        )

        await userApi.toggleUserActive('1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/users/1/toggle-active'),
          expect.objectContaining({ method: 'PATCH' })
        )
      })
    })
  })

  // ============================================================================
  // Employee API Tests
  // ============================================================================
  describe('employeeApi', () => {
    describe('getEmployees', () => {
      it('fetches employees with default pagination', async () => {
        const mockResponse = {
          employees: [{ id: '1', first_name: 'John' }],
          total: 1,
          page: 1,
          page_size: 1000,
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse))

        const result = await employeeApi.getEmployees()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees?page=1&page_size=1000'),
          expect.any(Object)
        )
        expect(result.employees).toHaveLength(1)
      })

      it('fetches employees with custom pagination', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ employees: [], total: 0, page: 2, page_size: 50 })
        )

        await employeeApi.getEmployees({ page: 2, page_size: 50 })

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees?page=2&page_size=50'),
          expect.any(Object)
        )
      })
    })

    describe('getEmployee', () => {
      it('fetches single employee by ID', async () => {
        const mockEmployee = { id: '123', first_name: 'John', last_name: 'Doe' }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockEmployee))

        const result = await employeeApi.getEmployee('123')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/123'),
          expect.any(Object)
        )
        expect(result.first_name).toBe('John')
      })
    })

    describe('createEmployee', () => {
      it('creates new employee', async () => {
        const newEmployee = {
          first_name: 'Jane',
          last_name: 'Smith',
          rfc: 'SMIJ900101ABC',
          curp: 'SMIJ900101MDFXXX00',
        }
        mockFetch.mockResolvedValueOnce(createMockResponse({ id: 'new-id', ...newEmployee }))

        const result = await employeeApi.createEmployee(newEmployee)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(newEmployee),
          })
        )
        expect(result.first_name).toBe('Jane')
      })
    })

    describe('updateEmployee', () => {
      it('updates employee', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ id: '123', daily_salary: 600 })
        )

        await employeeApi.updateEmployee('123', { daily_salary: 600 })

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/123'),
          expect.objectContaining({
            method: 'PUT',
            body: JSON.stringify({ daily_salary: 600 }),
          })
        )
      })
    })

    describe('deleteEmployee', () => {
      it('deletes employee', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ success: true }))

        const result = await employeeApi.deleteEmployee('123')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/123'),
          expect.objectContaining({ method: 'DELETE' })
        )
        expect(result.success).toBe(true)
      })
    })

    describe('terminateEmployee', () => {
      it('terminates employee with reason', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ id: '123', employment_status: 'terminated' })
        )

        await employeeApi.terminateEmployee('123', {
          termination_date: '2025-01-15',
          reason: 'Voluntary resignation',
        })

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/123/terminate'),
          expect.objectContaining({
            method: 'POST',
            body: expect.stringContaining('termination_date'),
          })
        )
      })
    })

    describe('updateSalary', () => {
      it('updates employee salary', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ id: '123', daily_salary: 700 })
        )

        await employeeApi.updateSalary('123', {
          new_daily_salary: 700,
          effective_date: '2025-02-01',
        })

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/123/salary'),
          expect.objectContaining({
            method: 'PUT',
          })
        )
      })
    })

    describe('getStats', () => {
      it('fetches employee statistics', async () => {
        const mockStats = {
          total_employees: 100,
          active_employees: 95,
          inactive_employees: 5,
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockStats))

        const result = await employeeApi.getStats()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/stats'),
          expect.any(Object)
        )
        expect(result.total_employees).toBe(100)
      })
    })

    describe('validateMexicanIds', () => {
      it('validates RFC, CURP, and NSS', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ valid: true, errors: [] }))

        const result = await employeeApi.validateMexicanIds(
          'PEGJ900101ABC',
          'PEGJ900101HSPLRN09',
          '12345678901'
        )

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/employees/validate-ids'),
          expect.objectContaining({ method: 'POST' })
        )
        expect(result.valid).toBe(true)
      })

      it('validates only provided IDs', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ valid: true, errors: [] }))

        await employeeApi.validateMexicanIds('PEGJ900101ABC')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('rfc=PEGJ900101ABC'),
          expect.any(Object)
        )
      })
    })
  })

  // ============================================================================
  // Payroll API Tests
  // ============================================================================
  describe('payrollApi', () => {
    describe('getPeriods', () => {
      it('fetches payroll periods', async () => {
        const mockPeriods = [
          { id: '1', period_code: '2025-BW01', status: 'open' },
          { id: '2', period_code: '2025-BW02', status: 'closed' },
        ]
        mockFetch.mockResolvedValueOnce(createMockResponse(mockPeriods))

        const result = await payrollApi.getPeriods()

        expect(result).toHaveLength(2)
        expect(result[0].period_code).toBe('2025-BW01')
      })
    })

    describe('getPeriod', () => {
      it('fetches single period', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ id: '1', period_code: '2025-BW01' })
        )

        const result = await payrollApi.getPeriod('1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/periods/1'),
          expect.any(Object)
        )
        expect(result.period_code).toBe('2025-BW01')
      })
    })

    describe('calculatePayroll', () => {
      it('calculates payroll for employee', async () => {
        const mockResult = {
          id: 'calc-1',
          employee_id: 'emp-1',
          total_net_pay: 8500.00,
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockResult))

        const result = await payrollApi.calculatePayroll('emp-1', 'period-1', true)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/calculate'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({
              employee_id: 'emp-1',
              payroll_period_id: 'period-1',
              calculate_sdi: true,
            }),
          })
        )
        expect(result.total_net_pay).toBe(8500.00)
      })

      it('calculates payroll without SDI calculation', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ id: 'calc-1' }))

        await payrollApi.calculatePayroll('emp-1', 'period-1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.any(String),
          expect.objectContaining({
            body: expect.stringContaining('"calculate_sdi":false'),
          })
        )
      })
    })

    describe('bulkCalculatePayroll', () => {
      it('bulk calculates for all employees', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ message: 'Calculated 50 employees' })
        )

        await payrollApi.bulkCalculatePayroll('period-1', undefined, true)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/bulk-calculate'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({
              payroll_period_id: 'period-1',
              employee_ids: undefined,
              calculate_all: true,
            }),
          })
        )
      })

      it('bulk calculates for specific employees', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ message: 'Calculated' }))

        await payrollApi.bulkCalculatePayroll('period-1', ['emp-1', 'emp-2'])

        expect(mockFetch).toHaveBeenCalledWith(
          expect.any(String),
          expect.objectContaining({
            body: expect.stringContaining('"employee_ids":["emp-1","emp-2"]'),
          })
        )
      })
    })

    describe('approvePayroll', () => {
      it('approves payroll period', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ message: 'Payroll approved' })
        )

        const result = await payrollApi.approvePayroll('period-1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/approve/period-1'),
          expect.objectContaining({ method: 'POST' })
        )
        expect(result.message).toBe('Payroll approved')
      })
    })

    describe('getPayrollSummary', () => {
      it('fetches payroll summary', async () => {
        const mockSummary = {
          period_id: 'period-1',
          total_gross: 100000,
          total_net: 85000,
        }
        mockFetch.mockResolvedValueOnce(createMockResponse(mockSummary))

        const result = await payrollApi.getPayrollSummary('period-1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/summary/period-1'),
          expect.any(Object)
        )
        expect(result.total_gross).toBe(100000)
      })
    })

    describe('getPayslip', () => {
      it('fetches payslip in PDF format', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ data: 'pdf-data' }))

        await payrollApi.getPayslip('period-1', 'emp-1', 'pdf')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/payslip/period-1/emp-1?format=pdf'),
          expect.any(Object)
        )
      })

      it('defaults to PDF format', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ data: 'pdf-data' }))

        await payrollApi.getPayslip('period-1', 'emp-1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('format=pdf'),
          expect.any(Object)
        )
      })
    })

    describe('processPayment', () => {
      it('processes payment for period', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ message: 'Payment processed' })
        )

        await payrollApi.processPayment('period-1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/payroll/payment/period-1'),
          expect.objectContaining({ method: 'POST' })
        )
      })
    })
  })

  // ============================================================================
  // Catalog API Tests
  // ============================================================================
  describe('catalogApi', () => {
    describe('getPayrollConcepts', () => {
      it('fetches payroll concepts', async () => {
        const mockConcepts = [
          { id: '1', name: 'Salary', category: 'EARNING' },
          { id: '2', name: 'ISR', category: 'DEDUCTION' },
        ]
        mockFetch.mockResolvedValueOnce(createMockResponse(mockConcepts))

        const result = await catalogApi.getPayrollConcepts()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/catalogs/concepts'),
          expect.any(Object)
        )
        expect(result).toHaveLength(2)
      })
    })

    describe('getIncidenceTypes', () => {
      it('fetches incidence types', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse([{ id: '1', name: 'Vacation' }])
        )

        await catalogApi.getIncidenceTypes()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/catalogs/incidence-types'),
          expect.any(Object)
        )
      })
    })

    describe('createConcept', () => {
      it('creates new concept', async () => {
        const newConcept = { name: 'Bonus', category: 'EARNING' }
        mockFetch.mockResolvedValueOnce(createMockResponse({ id: '3', ...newConcept }))

        await catalogApi.createConcept(newConcept)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/catalogs/concepts'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(newConcept),
          })
        )
      })
    })
  })

  // ============================================================================
  // Report API Tests
  // ============================================================================
  describe('reportApi', () => {
    describe('generateReport', () => {
      it('generates report with default format', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ data: 'report-data' }))

        await reportApi.generateReport('payroll_summary', 'period-1')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/reports/generate'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({
              report_type: 'payroll_summary',
              payroll_period_id: 'period-1',
              format: 'json',
            }),
          })
        )
      })

      it('generates report with custom format', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ data: 'excel-data' }))

        await reportApi.generateReport('payroll_summary', 'period-1', 'excel')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.any(String),
          expect.objectContaining({
            body: expect.stringContaining('"format":"excel"'),
          })
        )
      })
    })

    describe('getReportHistory', () => {
      it('fetches report history', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse([{ id: '1', type: 'payroll' }]))

        await reportApi.getReportHistory()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/reports/history'),
          expect.any(Object)
        )
      })
    })
  })

  // ============================================================================
  // Error Handling Tests
  // ============================================================================
  describe('Error Handling', () => {
    it('throws ApiError on network failure', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      try {
        await authApi.login('test@test.com', 'pass')
        fail('Should have thrown')
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError)
        expect((error as ApiError).isNetworkError).toBe(true)
      }
    })

    it('throws ApiError with server error message', async () => {
      mockFetch.mockResolvedValueOnce(
        createMockErrorResponse(400, { message: 'Email already exists' })
      )

      try {
        await authApi.register({ email: 'existing@test.com' })
        fail('Should have thrown')
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError)
        expect((error as ApiError).status).toBe(400)
      }
    })

    it('handles 401 error', async () => {
      mockFetch.mockResolvedValueOnce(createMockErrorResponse(401))

      try {
        await employeeApi.getEmployees()
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError)
        expect((error as ApiError).isAuthError).toBe(true)
      }
    })

    it('handles 403 forbidden error', async () => {
      mockFetch.mockResolvedValueOnce(createMockErrorResponse(403))

      try {
        await userApi.deleteUser('123')
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError)
        expect((error as ApiError).isAuthError).toBe(true)
      }
    })

    it('handles 404 not found error', async () => {
      mockFetch.mockResolvedValueOnce(createMockErrorResponse(404))

      try {
        await employeeApi.getEmployee('nonexistent')
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError)
        expect((error as ApiError).status).toBe(404)
      }
    })

    it('handles 500 server error', async () => {
      mockFetch.mockResolvedValueOnce(createMockErrorResponse(500))

      try {
        await payrollApi.getPeriods()
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError)
        expect((error as ApiError).isServerError).toBe(true)
      }
    })

    it('handles empty response body', async () => {
      const emptyResponse = {
        ok: true,
        status: 204,
        headers: {
          get: () => null, // No content-type header
        },
        json: jest.fn(),
      }
      mockFetch.mockResolvedValueOnce(emptyResponse)

      const result = await authApi.logout()

      expect(result).toEqual({})
    })
  })

  // ============================================================================
  // Health API Tests
  // ============================================================================
  describe('healthApi', () => {
    describe('getHealth', () => {
      it('fetches health status', async () => {
        mockFetch.mockResolvedValueOnce(
          createMockResponse({ status: 'healthy', version: '1.0.0' })
        )

        const result = await healthApi.getHealth()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/health'),
          expect.any(Object)
        )
        expect(result.status).toBe('healthy')
      })
    })

    describe('isServerAvailable', () => {
      it('returns true when server is available', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse({ status: 'healthy' }))

        const result = await healthApi.isServerAvailable()

        expect(result).toBe(true)
      })

      it('returns false when server is unavailable', async () => {
        mockFetch.mockRejectedValueOnce(new Error('Network error'))

        const result = await healthApi.isServerAvailable()

        expect(result).toBe(false)
      })
    })
  })
})
