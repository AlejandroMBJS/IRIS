/**
 * @file lib/__tests__/auth.test.ts
 * @description Unit tests for the authentication service
 */
import {
  login,
  registerCompany,
  logout,
  getCurrentUser,
  isAuthenticated,
  hasRole,
  getUserRole,
  isAdmin,
  isSupervisor,
  isManager,
  isTeamFocusedRole,
  canManageUsers,
  canDeleteEmployees,
  canAddEmployees,
  canProcessPayroll,
  canExportPayroll,
  canViewConfiguration,
  isHR,
  isHRAndPayroll,
  isHRWithCollarRestriction,
  getHRCollarAccess,
} from '../auth'

// Mock the api-client module
jest.mock('../api-client', () => ({
  authApi: {
    login: jest.fn(),
    register: jest.fn(),
  },
}))

import { authApi } from '../api-client'

const mockAuthApi = authApi as jest.Mocked<typeof authApi>

describe('auth.ts', () => {
  // Setup and teardown
  beforeEach(() => {
    localStorage.clear()
    jest.clearAllMocks()
  })

  // ============================================================================
  // getCurrentUser Tests
  // ============================================================================
  describe('getCurrentUser', () => {
    it('returns null when no user in localStorage', () => {
      expect(getCurrentUser()).toBeNull()
    })

    it('returns user object when valid user in localStorage', () => {
      const user = { id: '123', email: 'test@example.com', role: 'admin' }
      localStorage.setItem('user', JSON.stringify(user))

      expect(getCurrentUser()).toEqual(user)
    })

    it('returns null when localStorage contains invalid JSON', () => {
      localStorage.setItem('user', 'invalid-json')

      expect(getCurrentUser()).toBeNull()
    })
  })

  // ============================================================================
  // isAuthenticated Tests
  // ============================================================================
  describe('isAuthenticated', () => {
    it('returns false when no user logged in', () => {
      expect(isAuthenticated()).toBe(false)
    })

    it('returns true when user is logged in', () => {
      localStorage.setItem('user', JSON.stringify({ id: '123', role: 'admin' }))

      expect(isAuthenticated()).toBe(true)
    })
  })

  // ============================================================================
  // hasRole Tests
  // ============================================================================
  describe('hasRole', () => {
    it('returns false when no user logged in', () => {
      expect(hasRole('admin')).toBe(false)
    })

    it('returns true when user has the specified role', () => {
      localStorage.setItem('user', JSON.stringify({ id: '123', role: 'admin' }))

      expect(hasRole('admin')).toBe(true)
    })

    it('returns false when user does not have the specified role', () => {
      localStorage.setItem('user', JSON.stringify({ id: '123', role: 'employee' }))

      expect(hasRole('admin')).toBe(false)
    })
  })

  // ============================================================================
  // getUserRole Tests
  // ============================================================================
  describe('getUserRole', () => {
    it('returns null when no user logged in', () => {
      expect(getUserRole()).toBeNull()
    })

    it('returns the user role when logged in', () => {
      localStorage.setItem('user', JSON.stringify({ id: '123', role: 'hr' }))

      expect(getUserRole()).toBe('hr')
    })
  })

  // ============================================================================
  // Role Check Functions Tests
  // ============================================================================
  describe('isAdmin', () => {
    it('returns true for admin role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
      expect(isAdmin()).toBe(true)
    })

    it('returns false for non-admin roles', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(isAdmin()).toBe(false)
    })
  })

  describe('isSupervisor', () => {
    it('returns true for supervisor role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'supervisor' }))
      expect(isSupervisor()).toBe(true)
    })

    it('returns true for sup_and_gm role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'sup_and_gm' }))
      expect(isSupervisor()).toBe(true)
    })

    it('returns false for other roles', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(isSupervisor()).toBe(false)
    })
  })

  describe('isManager', () => {
    it('returns true for manager role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'manager' }))
      expect(isManager()).toBe(true)
    })

    it('returns true for sup_and_gm role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'sup_and_gm' }))
      expect(isManager()).toBe(true)
    })

    it('returns false for other roles', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'employee' }))
      expect(isManager()).toBe(false)
    })
  })

  describe('isTeamFocusedRole', () => {
    const teamFocusedRoles = ['supervisor', 'manager', 'sup_and_gm']

    teamFocusedRoles.forEach(role => {
      it(`returns true for ${role} role`, () => {
        localStorage.setItem('user', JSON.stringify({ role }))
        expect(isTeamFocusedRole()).toBe(true)
      })
    })

    it('returns false for non-team-focused roles', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'accountant' }))
      expect(isTeamFocusedRole()).toBe(false)
    })
  })

  // ============================================================================
  // Permission Functions Tests
  // ============================================================================
  describe('canManageUsers', () => {
    it('returns true only for admin', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
      expect(canManageUsers()).toBe(true)
    })

    it('returns false for non-admin', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(canManageUsers()).toBe(false)
    })
  })

  describe('canDeleteEmployees', () => {
    it('returns true only for admin', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
      expect(canDeleteEmployees()).toBe(true)
    })

    it('returns false for non-admin', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(canDeleteEmployees()).toBe(false)
    })
  })

  describe('canAddEmployees', () => {
    const allowedRoles = ['admin', 'hr', 'hr_and_pr', 'hr_blue_gray', 'hr_white', 'accountant', 'payroll_staff']

    allowedRoles.forEach(role => {
      it(`returns true for ${role} role`, () => {
        localStorage.setItem('user', JSON.stringify({ role }))
        expect(canAddEmployees()).toBe(true)
      })
    })

    it('returns false for employee role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'employee' }))
      expect(canAddEmployees()).toBe(false)
    })
  })

  describe('canProcessPayroll', () => {
    const allowedRoles = ['admin', 'accountant', 'payroll_staff']

    allowedRoles.forEach(role => {
      it(`returns true for ${role} role`, () => {
        localStorage.setItem('user', JSON.stringify({ role }))
        expect(canProcessPayroll()).toBe(true)
      })
    })

    it('returns false for hr role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(canProcessPayroll()).toBe(false)
    })
  })

  describe('canExportPayroll', () => {
    const allowedRoles = ['admin', 'accountant', 'payroll_staff']

    allowedRoles.forEach(role => {
      it(`returns true for ${role} role`, () => {
        localStorage.setItem('user', JSON.stringify({ role }))
        expect(canExportPayroll()).toBe(true)
      })
    })

    it('returns false for employee role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'employee' }))
      expect(canExportPayroll()).toBe(false)
    })
  })

  describe('canViewConfiguration', () => {
    it('returns true only for admin', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
      expect(canViewConfiguration()).toBe(true)
    })

    it('returns false for non-admin', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(canViewConfiguration()).toBe(false)
    })
  })

  // ============================================================================
  // HR Role Functions Tests
  // ============================================================================
  describe('isHR', () => {
    const hrRoles = ['hr', 'hr_and_pr', 'hr_blue_gray', 'hr_white', 'admin']

    hrRoles.forEach(role => {
      it(`returns true for ${role} role`, () => {
        localStorage.setItem('user', JSON.stringify({ role }))
        expect(isHR()).toBe(true)
      })
    })

    it('returns false for employee role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'employee' }))
      expect(isHR()).toBe(false)
    })
  })

  describe('isHRAndPayroll', () => {
    it('returns true for hr_and_pr role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr_and_pr' }))
      expect(isHRAndPayroll()).toBe(true)
    })

    it('returns true for admin role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
      expect(isHRAndPayroll()).toBe(true)
    })

    it('returns false for hr role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(isHRAndPayroll()).toBe(false)
    })
  })

  describe('isHRWithCollarRestriction', () => {
    it('returns true for hr_blue_gray role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr_blue_gray' }))
      expect(isHRWithCollarRestriction()).toBe(true)
    })

    it('returns true for hr_white role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr_white' }))
      expect(isHRWithCollarRestriction()).toBe(true)
    })

    it('returns false for hr role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(isHRWithCollarRestriction()).toBe(false)
    })
  })

  describe('getHRCollarAccess', () => {
    it('returns blue_collar and gray_collar for hr_blue_gray role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr_blue_gray' }))
      expect(getHRCollarAccess()).toEqual(['blue_collar', 'gray_collar'])
    })

    it('returns white_collar for hr_white role', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr_white' }))
      expect(getHRCollarAccess()).toEqual(['white_collar'])
    })

    it('returns null for admin (all access)', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
      expect(getHRCollarAccess()).toBeNull()
    })

    it('returns null for hr role (all access)', () => {
      localStorage.setItem('user', JSON.stringify({ role: 'hr' }))
      expect(getHRCollarAccess()).toBeNull()
    })
  })

  // ============================================================================
  // login Function Tests
  // ============================================================================
  describe('login', () => {
    it('returns success and stores user on successful login', async () => {
      const mockResponse = {
        access_token: 'mock-token',
        refresh_token: 'mock-refresh',
        user: { id: '123', email: 'test@example.com', role: 'admin' },
      }
      mockAuthApi.login.mockResolvedValue(mockResponse)

      const result = await login('test@example.com', 'password123')

      expect(result.success).toBe(true)
      expect(result.token).toBe('mock-token')
      expect(result.user).toEqual(mockResponse.user)
      expect(localStorage.getItem('user')).toBe(JSON.stringify(mockResponse.user))
    })

    it('returns failure when no access_token in response', async () => {
      mockAuthApi.login.mockResolvedValue({} as any)

      const result = await login('test@example.com', 'password123')

      expect(result.success).toBe(false)
      expect(result.message).toBe('Invalid response from server')
    })

    it('returns failure with error message on API error', async () => {
      mockAuthApi.login.mockRejectedValue(new Error('Invalid credentials'))

      const result = await login('test@example.com', 'wrong-password')

      expect(result.success).toBe(false)
      expect(result.message).toBe('Invalid credentials')
    })
  })

  // ============================================================================
  // registerCompany Function Tests
  // ============================================================================
  describe('registerCompany', () => {
    it('maps form fields correctly and returns success', async () => {
      const mockResponse = {
        access_token: 'mock-token',
        refresh_token: 'mock-refresh',
        user: { id: '123', email: 'admin@company.com', role: 'admin' },
      }
      mockAuthApi.register.mockResolvedValue(mockResponse)

      const formData = {
        companyName: 'Test Company',
        rfc: 'ABC123456789',
        email: 'admin@company.com',
        password: 'SecurePass123!',
        contactName: 'John Doe',
      }

      const result = await registerCompany(formData)

      // Verify the API was called with correctly mapped data
      expect(mockAuthApi.register).toHaveBeenCalledWith({
        company_name: 'Test Company',
        company_rfc: 'ABC123456789',
        email: 'admin@company.com',
        password: 'SecurePass123!',
        role: 'admin',
        full_name: 'John Doe',
      })

      expect(result.success).toBe(true)
      expect(result.message).toBe('Company registered successfully')
      expect(localStorage.getItem('user')).toBe(JSON.stringify(mockResponse.user))
    })

    it('returns failure when registration fails', async () => {
      mockAuthApi.register.mockRejectedValue(new Error('Email already exists'))

      const result = await registerCompany({
        companyName: 'Test',
        rfc: 'ABC123',
        email: 'existing@email.com',
        password: 'pass',
        contactName: 'John',
      })

      expect(result.success).toBe(false)
      expect(result.message).toBe('Email already exists')
    })
  })

  // ============================================================================
  // logout Function Tests
  // ============================================================================
  describe('logout', () => {
    // Note: logout redirects, which can't be fully tested in jsdom
    // We test that it clears localStorage
    it('clears user from localStorage', () => {
      localStorage.setItem('user', JSON.stringify({ id: '123' }))

      // Mock window.location.href setter
      const originalLocation = window.location
      delete (window as any).location
      window.location = { ...originalLocation, href: '' } as any

      logout()

      expect(localStorage.getItem('user')).toBeNull()

      // Restore window.location
      window.location = originalLocation
    })
  })
})
