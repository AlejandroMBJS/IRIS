/**
 * @file lib/auth.ts
 * @description Authentication service layer for the IRIS Payroll System. Manages user login, registration,
 * logout, and role-based authorization. Handles token storage and provides utility functions to check
 * user permissions across the application.
 *
 * USER PERSPECTIVE:
 *   - Enables secure login and registration with automatic session management
 *   - Maintains user session across page refreshes via localStorage
 *   - Automatically redirects to login page when session expires
 *   - Controls access to features based on user roles (admin, hr, accountant, etc.)
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify:
 *     - Add new role-based permission functions (e.g., canViewReports(), canManageIncidences())
 *     - Enhance error handling in login() and registerCompany()
 *     - Add session timeout warnings or auto-refresh logic
 *     - Update role names if backend role system changes
 *
 *   CAUTION:
 *     - logout() function clears tokens and redirects - ensure proper cleanup
 *     - getCurrentUser() reads from localStorage - can return stale data
 *     - Role permission functions are used throughout the app for conditional rendering
 *     - Registration maps frontend form fields to backend format - keep in sync
 *
 *   DO NOT modify:
 *     - Token storage keys ('auth_token', 'refresh_token', 'user') without updating api-client.ts
 *     - Login/logout flow without considering impact on navigation and protected routes
 *     - Role names without backend coordination (breaks authorization)
 *     - checkAuth() redirect logic without updating route protection
 *
 * EXPORTS:
 *   - Authentication: login(), registerCompany(), logout()
 *   - Session Management: getCurrentUser(), isAuthenticated(), checkAuth()
 *   - Authorization: hasRole(), isAdmin(), canManageUsers(), canDeleteEmployees(),
 *     canAddEmployees(), canProcessPayroll(), canExportPayroll(), canViewConfiguration()
 */

import { authApi, setAuthToken, clearAuthToken, getAuthToken, AuthResponse } from './api-client'

interface LoginResponse {
  success: boolean
  token?: string
  user?: AuthResponse['user']
  message?: string
}

interface RegisterResponse {
  success: boolean
  token?: string
  user?: AuthResponse['user']
  message?: string
}

export const login = async (email: string, password: string): Promise<LoginResponse> => {
  try {
    const response = await authApi.login(email, password)

    if (response.access_token) {
      setAuthToken(response.access_token)

      // Store refresh token and user data
      if (typeof window !== 'undefined') {
        localStorage.setItem('refresh_token', response.refresh_token)
        if (response.user) {
          localStorage.setItem('user', JSON.stringify(response.user))
        }
      }

      return {
        success: true,
        token: response.access_token,
        user: response.user
      }
    }

    return {
      success: false,
      message: 'Invalid response from server'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Login failed'
    }
  }
}

export const registerCompany = async (companyData: any): Promise<RegisterResponse> => {
  try {
    // Map frontend form fields to backend expected fields
    const backendData = {
      company_name: companyData.companyName,
      company_rfc: companyData.rfc,
      email: companyData.email,
      password: companyData.password,
      role: 'admin', // Default role for company registration
      full_name: companyData.contactName
    }

    const response = await authApi.register(backendData)

    if (response.access_token) {
      setAuthToken(response.access_token)

      // Store refresh token and user data
      if (typeof window !== 'undefined') {
        localStorage.setItem('refresh_token', response.refresh_token)
        if (response.user) {
          localStorage.setItem('user', JSON.stringify(response.user))
        }
      }

      return {
        success: true,
        token: response.access_token,
        user: response.user,
        message: 'Company registered successfully'
      }
    }

    return {
      success: false,
      message: 'Invalid response from server'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Registration failed'
    }
  }
}

export const logout = () => {
  clearAuthToken()
  if (typeof window !== 'undefined') {
    localStorage.removeItem('user')
    localStorage.removeItem('refresh_token')
    window.location.href = '/auth/login'
  }
}

export const getCurrentUser = () => {
  if (typeof window === 'undefined') return null
  
  const userStr = localStorage.getItem('user')
  if (!userStr) return null
  
  try {
    return JSON.parse(userStr)
  } catch {
    return null
  }
}

export const isAuthenticated = () => {
  return !!getAuthToken()
}

export const checkAuth = () => {
  if (!isAuthenticated()) {
    window.location.href = '/auth/login'
    return false
  }
  return true
}

// Role-based authorization
export const hasRole = (role: string) => {
  const user = getCurrentUser()
  return user?.role === role
}

// Check if user is admin
export const isAdmin = () => {
  return hasRole('admin')
}

// Check if user can manage users (only admin)
export const canManageUsers = () => {
  return isAdmin()
}

// Check if user can delete employees (only admin)
export const canDeleteEmployees = () => {
  return isAdmin()
}

// Check if user can add employees (admin and all HR roles)
export const canAddEmployees = () => {
  const user = getCurrentUser()
  return ['admin', 'hr', 'hr_and_pr', 'hr_blue_gray', 'hr_white'].includes(user?.role)
}

// Check if user can process payroll (admin, accountant, payroll_staff)
export const canProcessPayroll = () => {
  const user = getCurrentUser()
  return ['admin', 'accountant', 'payroll_staff'].includes(user?.role)
}

// Check if user can export payroll (admin, accountant, payroll_staff)
export const canExportPayroll = () => {
  const user = getCurrentUser()
  return ['admin', 'accountant', 'payroll_staff'].includes(user?.role)
}

// Check if user can view configuration (only admin)
export const canViewConfiguration = () => {
  return isAdmin()
}

// ============================================================================
// Employee Portal Role Functions
// ============================================================================

// Get user role
export const getUserRole = () => {
  const user = getCurrentUser()
  return user?.role || 'employee'
}

// Check if user is a supervisor (can approve supervisor stage)
export const isSupervisor = () => {
  const user = getCurrentUser()
  return ['supervisor', 'sup_and_gm', 'admin'].includes(user?.role)
}

// Check if user is a manager (can approve manager stage)
export const isManager = () => {
  const user = getCurrentUser()
  return ['manager', 'sup_and_gm', 'admin'].includes(user?.role)
}

// Check if user is HR (can approve HR stage) - includes all HR sub-roles
export const isHR = () => {
  const user = getCurrentUser()
  return ['hr', 'hr_and_pr', 'hr_blue_gray', 'hr_white', 'admin'].includes(user?.role)
}

// Check if user is HR and Payroll combined
export const isHRAndPayroll = () => {
  const user = getCurrentUser()
  return ['hr_and_pr', 'admin'].includes(user?.role)
}

// Check if user has HR role with restricted collar type access
export const isHRWithCollarRestriction = () => {
  const user = getCurrentUser()
  return ['hr_blue_gray', 'hr_white'].includes(user?.role)
}

// Get the collar types the user can access (null means all access)
export const getHRCollarAccess = (): string[] | null => {
  const user = getCurrentUser()
  switch(user?.role) {
    case 'hr_blue_gray':
      return ['blue_collar', 'gray_collar']
    case 'hr_white':
      return ['white_collar']
    default:
      return null // All access
  }
}

// Check if user is Supervisor and General Manager combined (skips approval stages)
export const isSupervisorAndGM = () => {
  const user = getCurrentUser()
  return user?.role === 'sup_and_gm'
}

// Check if user can create absence requests (all authenticated users)
export const canCreateAbsenceRequests = () => {
  return isAuthenticated()
}

// Check if user can approve at supervisor level
export const canApproveSupervisor = () => {
  return isSupervisor()
}

// Check if user can approve at manager level
export const canApproveManager = () => {
  return isManager()
}

// Check if user can approve at HR level
export const canApproveHR = () => {
  return isHR()
}

// Check if user has any approval role
export const hasApprovalRole = () => {
  return isSupervisor() || isManager() || isHR()
}

