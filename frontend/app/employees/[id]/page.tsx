/**
 * @file app/employees/[id]/page.tsx
 * @description Employee detail view showing all employee information in read-only mode
 *
 * USER PERSPECTIVE:
 *   - View complete employee information organized by section
 *   - Sections: Personal, Employment, Bank, Emergency Contact, Address
 *   - Status badge shows active/inactive state
 *   - Edit button navigates to edit page
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Section layout, field display order, formatting
 *   CAUTION: Ensure date and currency formatting matches locale requirements
 *   DO NOT modify: Dynamic route parameter handling without testing
 *
 * KEY COMPONENTS:
 *   - InfoField: Reusable label/value display component
 *   - Sections: Personal, Employment, Bank, Emergency Contact, Address
 *   - Status badges: Visual indicators for employment status
 *
 * API ENDPOINTS USED:
 *   - GET /api/employees/:id (via employeeApi.getEmployee)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import { ArrowLeft, Edit, User as UserIcon, Building, CreditCard, Phone, MapPin, KeyRound, Plus, Trash2, Save, X, Eye, EyeOff, Loader2, Clock, Users } from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { employeeApi, Employee, ApiError, PortalUserResponse, CreatePortalUserRequest, UpdatePortalUserRequest, userApi, User } from "@/lib/api-client"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

export default function EmployeeDetailPage() {
  const router = useRouter()
  const params = useParams()
  const [employee, setEmployee] = useState<Employee | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Portal user state
  const [portalUser, setPortalUser] = useState<PortalUserResponse | null>(null)
  const [portalUserLoading, setPortalUserLoading] = useState(false)
  const [portalUserError, setPortalUserError] = useState<string | null>(null)

  // Users for supervisor/manager selection
  const [allUsers, setAllUsers] = useState<User[]>([])
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [formLoading, setFormLoading] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [formData, setFormData] = useState<CreatePortalUserRequest>({
    email: '',
    password: '',
    role: 'employee',
    department: '',
    area: '',
  })
  const [editFormData, setEditFormData] = useState<UpdatePortalUserRequest>({})

  useEffect(() => {
    if (params.id) {
      loadEmployee(params.id as string)
      loadPortalUser(params.id as string)
      loadUsers()
    }
  }, [params.id])

  async function loadUsers() {
    try {
      const users = await userApi.getUsers()
      setAllUsers(users.filter(u => u.is_active))
    } catch (err) {
      console.error('Failed to load users for supervisor/manager selection:', err)
    }
  }

  async function loadEmployee(id: string) {
    try {
      setLoading(true)
      setError(null)
      const data = await employeeApi.getEmployee(id)
      setEmployee(data)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load employee")
      }
    } finally {
      setLoading(false)
    }
  }

  async function loadPortalUser(id: string) {
    try {
      setPortalUserLoading(true)
      setPortalUserError(null)
      const data = await employeeApi.getPortalUser(id)
      setPortalUser(data)
    } catch (err) {
      // Not having a portal user is normal, not an error
      setPortalUser(null)
      if (err instanceof ApiError && err.status !== 404) {
        setPortalUserError(err.message)
      }
    } finally {
      setPortalUserLoading(false)
    }
  }

  async function handleCreatePortalUser(e: React.FormEvent) {
    e.preventDefault()
    if (!params.id) return

    setFormLoading(true)
    setFormError(null)

    try {
      const data = await employeeApi.createPortalUser(params.id as string, formData)
      setPortalUser(data)
      setIsCreateDialogOpen(false)
      setFormData({ email: '', password: '', role: 'employee', department: '', area: '' })
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message)
      } else {
        setFormError("Error creating portal user")
      }
    } finally {
      setFormLoading(false)
    }
  }

  async function handleUpdatePortalUser(e: React.FormEvent) {
    e.preventDefault()
    if (!params.id) return

    setFormLoading(true)
    setFormError(null)

    try {
      const data = await employeeApi.updatePortalUser(params.id as string, editFormData)
      setPortalUser(data)
      setIsEditDialogOpen(false)
      setEditFormData({})
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message)
      } else {
        setFormError("Error updating portal user")
      }
    } finally {
      setFormLoading(false)
    }
  }

  async function handleDeletePortalUser() {
    if (!params.id) return

    setFormLoading(true)
    setFormError(null)

    try {
      await employeeApi.deletePortalUser(params.id as string)
      setPortalUser(null)
      setIsDeleteDialogOpen(false)
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message)
      } else {
        setFormError("Error deleting portal user")
      }
    } finally {
      setFormLoading(false)
    }
  }

  function openEditDialog() {
    if (portalUser) {
      setEditFormData({
        email: portalUser.email,
        role: portalUser.role,
        is_active: portalUser.is_active,
        supervisor_id: portalUser.supervisor_id || '',
        general_manager_id: portalUser.general_manager_id || '',
        department: portalUser.department || '',
        area: portalUser.area || '',
      })
    }
    setFormError(null)
    setIsEditDialogOpen(true)
  }

  const roleLabels: Record<string, string> = {
    employee: 'Employee',
    supervisor: 'Supervisor',
    manager: 'Manager',
    hr: 'Human Resources',
    hr_and_pr: 'HR and Payroll',
    sup_and_gm: 'Supervisor and GM',
    admin: 'Administrator',
  }

  const formatDate = (dateStr: string | undefined) => {
    if (!dateStr) return "-"
    try {
      return new Date(dateStr).toLocaleDateString("es-MX", {
        year: "numeric",
        month: "long",
        day: "numeric",
      })
    } catch {
      return dateStr
    }
  }

  const formatCurrency = (amount: number | undefined) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
    }).format(amount || 0)
  }

  const InfoField = ({ label, value }: { label: string; value: string | undefined | number }) => (
    <div>
      <dt className="text-sm text-slate-400">{label}</dt>
      <dd className="text-white font-medium mt-1">{value || "-"}</dd>
    </div>
  )

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="flex items-center gap-2 text-slate-400">
            <div className="w-5 h-5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin" />
            Loading employee...
          </div>
        </div>
      </DashboardLayout>
    )
  }

  if (error) {
    return (
      <DashboardLayout>
        <div className="space-y-4">
          <button
            onClick={() => router.push("/employees")}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={20} />
            Back to Employees
          </button>
          <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 text-red-400">
            {error}
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => router.push("/employees")}
              className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
            >
              <ArrowLeft size={20} />
            </button>
            <div>
              <h1 className="text-3xl font-bold text-white">{employee?.full_name}</h1>
              <p className="text-slate-400 mt-1">Employee #{employee?.employee_number}</p>
            </div>
          </div>
          <Button
            onClick={() => router.push(`/employees/${params.id}/edit`)}
            className="bg-blue-600 hover:bg-blue-700"
          >
            <Edit size={20} className="mr-2" />
            Edit Employee
          </Button>
        </div>

        {/* Status Badge */}
        <div className="flex items-center gap-4">
          <span
            className={`inline-flex px-3 py-1.5 text-sm font-semibold rounded-full ${
              employee?.employment_status === "active"
                ? "bg-green-900/30 text-green-400 border border-green-700"
                : employee?.employment_status === "inactive"
                ? "bg-yellow-900/30 text-yellow-400 border border-yellow-700"
                : "bg-slate-700/30 text-slate-400 border border-slate-600"
            }`}
          >
            {employee?.employment_status}
          </span>
          <span className="text-slate-400">|</span>
          <span className="text-slate-300">{employee?.employee_type}</span>
        </div>

        {/* Personal Information */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-6">
            <UserIcon className="text-blue-400" size={20} />
            <h2 className="text-xl font-semibold text-white">Personal Information</h2>
          </div>
          <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <InfoField label="Full Name" value={employee?.full_name} />
            <InfoField label="Date of Birth" value={formatDate(employee?.date_of_birth)} />
            <InfoField label="Age" value={employee?.age ? `${employee.age} years` : undefined} />
            <InfoField label="Gender" value={employee?.gender} />
            <InfoField label="RFC" value={employee?.rfc} />
            <InfoField label="CURP" value={employee?.curp} />
            <InfoField label="NSS" value={employee?.nss} />
            <InfoField label="Personal Email" value={employee?.personal_email} />
            <InfoField label="Personal Phone" value={employee?.personal_phone} />
          </dl>
        </div>

        {/* Employment Information */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-6">
            <Building className="text-purple-400" size={20} />
            <h2 className="text-xl font-semibold text-white">Employment Information</h2>
          </div>
          <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <InfoField label="Hire Date" value={formatDate(employee?.hire_date)} />
            <InfoField label="Seniority" value={employee?.seniority ? `${employee.seniority} years` : undefined} />
            <InfoField label="Termination Date" value={formatDate(employee?.termination_date)} />
            <InfoField label="Employee Type" value={employee?.employee_type} />
            <InfoField label="Daily Salary" value={formatCurrency(employee?.daily_salary)} />
            <InfoField label="Integrated Daily Salary (SDI)" value={formatCurrency(employee?.integrated_daily_salary)} />
            <InfoField label="Payment Method" value={employee?.payment_method} />
            <InfoField label="Tax Regime" value={employee?.tax_regime} />
            <InfoField label="IMSS Registration Date" value={formatDate(employee?.imss_registration_date)} />
            <InfoField label="INFONAVIT Credit" value={employee?.infonavit_credit} />
          </dl>
        </div>

        {/* Organization Information */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-6">
            <Users className="text-indigo-400" size={20} />
            <h2 className="text-xl font-semibold text-white">Organization</h2>
          </div>
          <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div>
              <dt className="text-sm text-slate-400">Shift</dt>
              <dd className="mt-1">
                {employee?.shift_name ? (
                  <span className="inline-flex items-center gap-2 px-3 py-1.5 bg-indigo-900/30 text-indigo-300 border border-indigo-700 rounded-lg">
                    <Clock size={14} />
                    {employee.shift_name}
                  </span>
                ) : (
                  <span className="text-slate-500">Not assigned</span>
                )}
              </dd>
            </div>
            <div>
              <dt className="text-sm text-slate-400">Direct Supervisor</dt>
              <dd className="mt-1">
                {employee?.supervisor_name ? (
                  <span className="inline-flex items-center gap-2 px-3 py-1.5 bg-purple-900/30 text-purple-300 border border-purple-700 rounded-lg">
                    <Users size={14} />
                    {employee.supervisor_name}
                  </span>
                ) : (
                  <span className="text-slate-500">Not assigned</span>
                )}
              </dd>
            </div>
            <InfoField label="Collar Type" value={
              employee?.collar_type === 'white_collar' ? 'White Collar (Administrative)' :
              employee?.collar_type === 'blue_collar' ? 'Blue Collar (Unionized)' :
              employee?.collar_type === 'gray_collar' ? 'Gray Collar (Non-Unionized)' :
              employee?.collar_type
            } />
            <InfoField label="Pay Frequency" value={
              employee?.pay_frequency === 'weekly' ? 'Weekly' :
              employee?.pay_frequency === 'biweekly' ? 'Biweekly' :
              employee?.pay_frequency === 'monthly' ? 'Monthly' :
              employee?.pay_frequency
            } />
          </dl>
        </div>

        {/* Bank Information */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-6">
            <CreditCard className="text-green-400" size={20} />
            <h2 className="text-xl font-semibold text-white">Bank Information</h2>
          </div>
          <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <InfoField label="Bank Name" value={employee?.bank_name} />
            <InfoField label="Bank Account" value={employee?.bank_account} />
            <InfoField label="CLABE" value={employee?.clabe} />
          </dl>
        </div>

        {/* Emergency Contact */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-6">
            <Phone className="text-red-400" size={20} />
            <h2 className="text-xl font-semibold text-white">Emergency Contact</h2>
          </div>
          <dl className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <InfoField label="Contact Name" value={employee?.emergency_contact} />
            <InfoField label="Contact Phone" value={employee?.emergency_phone} />
          </dl>
        </div>

        {/* Address */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-6">
            <MapPin className="text-amber-400" size={20} />
            <h2 className="text-xl font-semibold text-white">Address</h2>
          </div>
          <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <InfoField label="Street" value={employee?.street} />
            <InfoField label="Exterior Number" value={employee?.exterior_number} />
            <InfoField label="Interior Number" value={employee?.interior_number} />
            <InfoField label="Neighborhood" value={employee?.neighborhood} />
            <InfoField label="Municipality" value={employee?.municipality} />
            <InfoField label="State" value={employee?.state} />
            <InfoField label="Postal Code" value={employee?.postal_code} />
            <InfoField label="Country" value={employee?.country} />
          </dl>
        </div>

        {/* Portal Credentials */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-2">
              <KeyRound className="text-cyan-400" size={20} />
              <h2 className="text-xl font-semibold text-white">Portal Credentials</h2>
            </div>
            {!portalUserLoading && !portalUser && (
              <Button
                onClick={() => {
                  setFormError(null)
                  setFormData({
                    email: employee?.personal_email || '',
                    password: '',
                    role: 'employee',
                    department: '',
                    area: '',
                  })
                  setIsCreateDialogOpen(true)
                }}
                className="bg-cyan-600 hover:bg-cyan-700"
                size="sm"
              >
                <Plus size={16} className="mr-2" />
                Create Portal Access
              </Button>
            )}
          </div>

          {portalUserLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 text-slate-400 animate-spin" />
              <span className="ml-2 text-slate-400">Loading...</span>
            </div>
          ) : portalUserError ? (
            <div className="text-red-400 bg-red-900/20 border border-red-700/50 rounded-lg p-4">
              {portalUserError}
            </div>
          ) : portalUser ? (
            <div className="space-y-4">
              <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <InfoField label="Email" value={portalUser.email} />
                <InfoField label="Role" value={roleLabels[portalUser.role] || portalUser.role} />
                <div>
                  <dt className="text-sm text-slate-400">Status</dt>
                  <dd className="mt-1">
                    <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
                      portalUser.is_active
                        ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                        : 'bg-red-500/20 text-red-300 border border-red-500/30'
                    }`}>
                      {portalUser.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </dd>
                </div>
                <InfoField label="Last Login" value={portalUser.last_login_at ? formatDate(portalUser.last_login_at) : 'Never'} />
                <InfoField
                  label="Direct Supervisor"
                  value={portalUser.supervisor_id ? allUsers.find(u => u.id === portalUser.supervisor_id)?.full_name || 'Assigned' : 'Not assigned'}
                />
                <InfoField
                  label="General Manager"
                  value={portalUser.general_manager_id ? allUsers.find(u => u.id === portalUser.general_manager_id)?.full_name || 'Assigned' : 'Not assigned'}
                />
                <InfoField label="Department" value={portalUser.department || '-'} />
                <InfoField label="Area" value={portalUser.area || '-'} />
              </dl>
              <div className="flex items-center gap-2 pt-4 border-t border-slate-700">
                <Button
                  onClick={openEditDialog}
                  variant="outline"
                  size="sm"
                  className="border-slate-600 text-slate-300 hover:bg-slate-700"
                >
                  <Edit size={16} className="mr-2" />
                  Edit Credentials
                </Button>
                <Button
                  onClick={() => {
                    setFormError(null)
                    setIsDeleteDialogOpen(true)
                  }}
                  variant="outline"
                  size="sm"
                  className="border-red-600 text-red-400 hover:bg-red-600/20"
                >
                  <Trash2 size={16} className="mr-2" />
                  Delete Access
                </Button>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-slate-400">
              <KeyRound className="w-12 h-12 mx-auto mb-3 opacity-30" />
              <p>This employee does not have portal access</p>
              <p className="text-sm text-slate-500 mt-1">Create credentials so they can access the employee portal</p>
            </div>
          )}
        </div>

        {/* Metadata */}
        <div className="text-sm text-slate-500 flex gap-6">
          <span>Created: {formatDate(employee?.created_at)}</span>
          <span>Updated: {formatDate(employee?.updated_at)}</span>
        </div>

        {/* Create Portal User Dialog */}
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-md">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <KeyRound className="h-5 w-5 text-cyan-400" />
                Create Portal Access
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                Create credentials so {employee?.full_name} can access the employee portal.
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleCreatePortalUser} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Email *
                </label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                  placeholder="employee@company.com"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Password *
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    className="w-full px-3 py-2 pr-10 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                    placeholder="Minimum 8 characters"
                    required
                    minLength={8}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-300"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
                <p className="text-xs text-slate-500 mt-1">
                  Must include uppercase, lowercase, number and special character (!@#$%^&*())
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Role *
                </label>
                <select
                  value={formData.role}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-cyan-500"
                  required
                >
                  <option value="employee">Employee</option>
                  <option value="supervisor">Supervisor</option>
                  <option value="manager">Manager</option>
                  <option value="hr">Human Resources</option>
                  <option value="hr_and_pr">HR and Payroll</option>
                  <option value="sup_and_gm">Supervisor and GM</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Direct Supervisor
                </label>
                <select
                  value={formData.supervisor_id || ''}
                  onChange={(e) => setFormData({ ...formData, supervisor_id: e.target.value || undefined })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-cyan-500"
                >
                  <option value="">No supervisor assigned</option>
                  {allUsers.map(u => (
                    <option key={u.id} value={u.id}>{u.full_name} ({roleLabels[u.role] || u.role})</option>
                  ))}
                </select>
                <p className="text-xs text-slate-500 mt-1">
                  Select who will approve requests in first instance
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  General Manager
                </label>
                <select
                  value={formData.general_manager_id || ''}
                  onChange={(e) => setFormData({ ...formData, general_manager_id: e.target.value || undefined })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-cyan-500"
                >
                  <option value="">No manager assigned</option>
                  {allUsers.map(u => (
                    <option key={u.id} value={u.id}>{u.full_name} ({roleLabels[u.role] || u.role})</option>
                  ))}
                </select>
                <p className="text-xs text-slate-500 mt-1">
                  Select who will approve requests after the supervisor
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Department
                  </label>
                  <input
                    type="text"
                    value={formData.department}
                    onChange={(e) => setFormData({ ...formData, department: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                    placeholder="Operations"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Area
                  </label>
                  <input
                    type="text"
                    value={formData.area}
                    onChange={(e) => setFormData({ ...formData, area: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                    placeholder="Production"
                  />
                </div>
              </div>

              {formError && (
                <div className="text-red-400 text-sm bg-red-900/20 border border-red-700/50 rounded-lg p-3">
                  {formError}
                </div>
              )}

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setIsCreateDialogOpen(false)}
                  className="border-slate-600"
                  disabled={formLoading}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  className="bg-cyan-600 hover:bg-cyan-700"
                  disabled={formLoading}
                >
                  {formLoading ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      <Save className="h-4 w-4 mr-2" />
                      Create Access
                    </>
                  )}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        {/* Edit Portal User Dialog */}
        <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-md">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <Edit className="h-5 w-5 text-amber-400" />
                Edit Credentials
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                Modify portal access credentials for {employee?.full_name}.
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleUpdatePortalUser} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Email
                </label>
                <input
                  type="email"
                  value={editFormData.email || ''}
                  onChange={(e) => setEditFormData({ ...editFormData, email: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  New Password (leave empty to keep current)
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={editFormData.password || ''}
                    onChange={(e) => setEditFormData({ ...editFormData, password: e.target.value || undefined })}
                    className="w-full px-3 py-2 pr-10 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                    placeholder="New password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-300"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Role
                </label>
                <select
                  value={editFormData.role || ''}
                  onChange={(e) => setEditFormData({ ...editFormData, role: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-cyan-500"
                >
                  <option value="employee">Employee</option>
                  <option value="supervisor">Supervisor</option>
                  <option value="manager">Manager</option>
                  <option value="hr">Human Resources</option>
                  <option value="hr_and_pr">HR and Payroll</option>
                  <option value="sup_and_gm">Supervisor and GM</option>
                </select>
              </div>

              <div>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={editFormData.is_active !== false}
                    onChange={(e) => setEditFormData({ ...editFormData, is_active: e.target.checked })}
                    className="rounded border-slate-600 bg-slate-800 text-cyan-500"
                  />
                  <span className="text-sm text-slate-300">Active user</span>
                </label>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Direct Supervisor
                </label>
                <select
                  value={editFormData.supervisor_id || ''}
                  onChange={(e) => setEditFormData({ ...editFormData, supervisor_id: e.target.value || undefined })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-cyan-500"
                >
                  <option value="">No supervisor assigned</option>
                  {allUsers.map(u => (
                    <option key={u.id} value={u.id}>{u.full_name} ({roleLabels[u.role] || u.role})</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  General Manager
                </label>
                <select
                  value={editFormData.general_manager_id || ''}
                  onChange={(e) => setEditFormData({ ...editFormData, general_manager_id: e.target.value || undefined })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-cyan-500"
                >
                  <option value="">No manager assigned</option>
                  {allUsers.map(u => (
                    <option key={u.id} value={u.id}>{u.full_name} ({roleLabels[u.role] || u.role})</option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Department
                  </label>
                  <input
                    type="text"
                    value={editFormData.department || ''}
                    onChange={(e) => setEditFormData({ ...editFormData, department: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Area
                  </label>
                  <input
                    type="text"
                    value={editFormData.area || ''}
                    onChange={(e) => setEditFormData({ ...editFormData, area: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                  />
                </div>
              </div>

              {formError && (
                <div className="text-red-400 text-sm bg-red-900/20 border border-red-700/50 rounded-lg p-3">
                  {formError}
                </div>
              )}

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setIsEditDialogOpen(false)}
                  className="border-slate-600"
                  disabled={formLoading}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  className="bg-amber-600 hover:bg-amber-700"
                  disabled={formLoading}
                >
                  {formLoading ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      <Save className="h-4 w-4 mr-2" />
                      Save Changes
                    </>
                  )}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        {/* Delete Confirmation Dialog */}
        <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-md">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-red-400">
                <Trash2 className="h-5 w-5" />
                Delete Portal Access
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                This action will permanently delete portal access credentials for {employee?.full_name}. The employee will no longer be able to log in to the portal.
              </DialogDescription>
            </DialogHeader>

            {formError && (
              <div className="text-red-400 text-sm bg-red-900/20 border border-red-700/50 rounded-lg p-3">
                {formError}
              </div>
            )}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsDeleteDialogOpen(false)}
                className="border-slate-600"
                disabled={formLoading}
              >
                Cancel
              </Button>
              <Button
                onClick={handleDeletePortalUser}
                className="bg-red-600 hover:bg-red-700"
                disabled={formLoading}
              >
                {formLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Deleting...
                  </>
                ) : (
                  <>
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete Access
                  </>
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </DashboardLayout>
  )
}
