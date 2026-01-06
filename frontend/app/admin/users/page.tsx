/**
 * @file app/admin/users/page.tsx
 * @description User management page for admin-only access control
 *
 * USER PERSPECTIVE:
 *   - View all system users and their roles
 *   - Create new users with specific roles (HR, Accountant, Payroll Staff, Viewer)
 *   - Activate/deactivate user accounts
 *   - Delete non-admin users
 *   - Track user activity (last login)
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: User roles, form fields, validation rules
 *   CAUTION: Admin users cannot be deleted or deactivated
 *   DO NOT modify: Role permissions without updating auth middleware
 *
 * KEY COMPONENTS:
 *   - User table: List with role, status, and action buttons
 *   - Create user modal: Form with email, password, name, role
 *   - Stats cards: Total, active, inactive, admin counts
 *   - Role selector: Predefined roles with descriptions
 *
 * API ENDPOINTS USED:
 *   - GET /api/users (via userApi.getUsers)
 *   - POST /api/users (via userApi.createUser)
 *   - DELETE /api/users/:id (via userApi.deleteUser)
 *   - PUT /api/users/:id/toggle-active (via userApi.toggleUserActive)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Plus, Users, UserCheck, UserX, Trash2, RefreshCw,
  Shield, Eye, EyeOff, X, Edit
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { userApi, User, CreateUserRequest, UpdateUserRequest, ApiError } from "@/lib/api-client"
import { isAdmin } from "@/lib/auth"

const ROLES = [
  { value: "hr", label: "HR - All", description: "Can view and edit ALL employees" },
  { value: "hr_blue_gray", label: "HR - Operations", description: "Only sees Blue Collar and Gray Collar employees (operations and technicians)" },
  { value: "hr_white", label: "HR - Administrative", description: "Only sees White Collar employees (administrative)" },
  { value: "hr_and_pr", label: "HR + Payroll", description: "Combines HR and Payroll permissions" },
  { value: "accountant", label: "Accountant", description: "Can calculate and export payroll" },
  { value: "payroll_staff", label: "Payroll Staff", description: "Can only calculate and export payroll" },
  { value: "supervisor", label: "Supervisor", description: "Can approve requests from their employees" },
  { value: "manager", label: "General Manager", description: "Approves requests after supervisor" },
  { value: "sup_and_gm", label: "Supervisor + Manager", description: "Combines Supervisor and General Manager permissions" },
  { value: "employee", label: "Employee", description: "Employee user with access to employee portal" },
  { value: "viewer", label: "Read Only", description: "Can only view information" },
]

export default function UsersPage() {
  const router = useRouter()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showModal, setShowModal] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formData, setFormData] = useState<CreateUserRequest>({
    email: "",
    password: "",
    full_name: "",
    role: "payroll_staff",
  })
  const [showPassword, setShowPassword] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [editFormData, setEditFormData] = useState<UpdateUserRequest>({
    full_name: "",
    role: "",
    password: "",
  })
  const [showEditPassword, setShowEditPassword] = useState(false)

  useEffect(() => {
    if (!isAdmin()) {
      router.push("/dashboard")
      return
    }
    loadUsers()
  }, [router])

  async function loadUsers() {
    try {
      setLoading(true)
      setError(null)
      const data = await userApi.getUsers()
      setUsers(Array.isArray(data) ? data : [])
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Error loading users")
      }
      setUsers([])
    } finally {
      setLoading(false)
    }
  }

  async function handleCreateUser(e: React.FormEvent) {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      await userApi.createUser(formData)
      setShowModal(false)
      setFormData({ email: "", password: "", full_name: "", role: "payroll_staff" })
      loadUsers()
    } catch (err) {
      if (err instanceof ApiError) {
        alert(err.message)
      } else {
        alert("Error creating user")
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleDeleteUser(id: string, email: string) {
    if (!confirm(`Are you sure you want to delete user ${email}?`)) return

    try {
      await userApi.deleteUser(id)
      loadUsers()
    } catch (err) {
      if (err instanceof ApiError) {
        alert(err.message)
      } else {
        alert("Error deleting user")
      }
    }
  }

  async function handleToggleActive(id: string) {
    try {
      await userApi.toggleUserActive(id)
      loadUsers()
    } catch (err) {
      if (err instanceof ApiError) {
        alert(err.message)
      } else {
        alert("Error changing status")
      }
    }
  }

  function openEditModal(user: User) {
    setEditingUser(user)
    setEditFormData({
      full_name: user.full_name,
      role: user.role,
      password: "",
    })
    setShowEditPassword(false)
    setShowEditModal(true)
  }

  async function handleEditUser(e: React.FormEvent) {
    e.preventDefault()
    if (!editingUser) return
    setIsSubmitting(true)

    try {
      // Only include password if it was provided
      const updateData: UpdateUserRequest = {
        full_name: editFormData.full_name,
        role: editFormData.role,
      }
      if (editFormData.password && editFormData.password.length > 0) {
        updateData.password = editFormData.password
      }
      await userApi.updateUser(editingUser.id, updateData)
      setShowEditModal(false)
      setEditingUser(null)
      loadUsers()
    } catch (err) {
      if (err instanceof ApiError) {
        alert(err.message)
      } else {
        alert("Error updating user")
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const getRoleBadge = (role: string) => {
    switch (role) {
      case "admin":
        return "bg-gradient-to-r from-purple-500/20 to-pink-500/20 text-purple-300 border border-purple-500/30"
      case "hr":
        return "bg-gradient-to-r from-blue-500/20 to-cyan-500/20 text-blue-300 border border-blue-500/30"
      case "hr_blue_gray":
        return "bg-gradient-to-r from-sky-500/20 to-indigo-500/20 text-sky-300 border border-sky-500/30"
      case "hr_white":
        return "bg-gradient-to-r from-teal-500/20 to-cyan-500/20 text-teal-300 border border-teal-500/30"
      case "hr_and_pr":
        return "bg-gradient-to-r from-violet-500/20 to-blue-500/20 text-violet-300 border border-violet-500/30"
      case "accountant":
        return "bg-gradient-to-r from-emerald-500/20 to-green-500/20 text-emerald-300 border border-emerald-500/30"
      case "payroll_staff":
        return "bg-gradient-to-r from-amber-500/20 to-yellow-500/20 text-amber-300 border border-amber-500/30"
      case "supervisor":
        return "bg-gradient-to-r from-orange-500/20 to-red-500/20 text-orange-300 border border-orange-500/30"
      case "manager":
        return "bg-gradient-to-r from-rose-500/20 to-pink-500/20 text-rose-300 border border-rose-500/30"
      case "sup_and_gm":
        return "bg-gradient-to-r from-red-500/20 to-orange-500/20 text-red-300 border border-red-500/30"
      case "employee":
        return "bg-gradient-to-r from-lime-500/20 to-green-500/20 text-lime-300 border border-lime-500/30"
      case "viewer":
        return "bg-gradient-to-r from-slate-500/20 to-zinc-500/20 text-slate-300 border border-slate-500/30"
      default:
        return "bg-slate-700/50 text-slate-400"
    }
  }

  const getRoleLabel = (role: string) => {
    switch (role) {
      case "admin": return "Administrator"
      case "hr": return "HR - All"
      case "hr_blue_gray": return "HR - Operations"
      case "hr_white": return "HR - Administrative"
      case "hr_and_pr": return "HR + Payroll"
      case "accountant": return "Accountant"
      case "payroll_staff": return "Payroll Staff"
      case "supervisor": return "Supervisor"
      case "manager": return "General Manager"
      case "sup_and_gm": return "Supervisor + Manager"
      case "employee": return "Employee"
      case "viewer": return "Read Only"
      default: return role
    }
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-"
    try {
      return new Date(dateStr).toLocaleDateString("es-MX", {
        year: "numeric",
        month: "short",
        day: "numeric",
      })
    } catch {
      return dateStr
    }
  }

  const stats = {
    total: users.length,
    active: users.filter(u => u.is_active).length,
    inactive: users.filter(u => !u.is_active).length,
    admins: users.filter(u => u.role === "admin").length,
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              User Management
            </h1>
            <p className="text-slate-400 mt-1">Manage your company users</p>
          </div>
          <div className="flex items-center gap-3">
            <Button
              onClick={loadUsers}
              variant="outline"
              size="sm"
              className="border-slate-600 text-slate-300 hover:bg-slate-700"
            >
              <RefreshCw size={16} className="mr-2" />
              Refresh
            </Button>
            <Button
              onClick={() => setShowModal(true)}
              className="bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 shadow-lg shadow-blue-500/25"
            >
              <Plus size={20} className="mr-2" />
              New User
            </Button>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/20 rounded-lg">
                <Users className="w-5 h-5 text-blue-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.total}</p>
                <p className="text-xs text-slate-400">Total Users</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-500/20 rounded-lg">
                <UserCheck className="w-5 h-5 text-emerald-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.active}</p>
                <p className="text-xs text-slate-400">Active</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-500/20 rounded-lg">
                <UserX className="w-5 h-5 text-amber-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.inactive}</p>
                <p className="text-xs text-slate-400">Inactive</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-500/20 rounded-lg">
                <Shield className="w-5 h-5 text-purple-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.admins}</p>
                <p className="text-xs text-slate-400">Administrators</p>
              </div>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-900/20 border border-red-700/50 rounded-xl p-4 text-red-400 flex items-center justify-between">
            <span>{error}</span>
            <button onClick={loadUsers} className="px-3 py-1 bg-red-500/20 hover:bg-red-500/30 rounded-lg transition-colors">
              Retry
            </button>
          </div>
        )}

        {/* Table */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/80 border-b border-slate-700/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Role
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Last Login
                  </th>
                  <th className="px-6 py-4 text-center text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {loading ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-12 text-center text-slate-400">
                      <div className="flex flex-col items-center gap-3">
                        <div className="w-8 h-8 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                        <span>Loading users...</span>
                      </div>
                    </td>
                  </tr>
                ) : users.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-12 text-center text-slate-400">
                      <div className="flex flex-col items-center gap-3">
                        <Users className="w-12 h-12 text-slate-600" />
                        <p>No users registered</p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  users.map((user) => (
                    <tr
                      key={user.id}
                      className="hover:bg-slate-700/30 transition-colors group"
                    >
                      <td className="px-6 py-4">
                        <div>
                          <div className="text-sm text-white font-medium">{user.full_name}</div>
                          <div className="text-xs text-slate-500">{user.email}</div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-lg ${getRoleBadge(user.role)}`}>
                          {getRoleLabel(user.role)}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-lg ${
                          user.is_active
                            ? "bg-gradient-to-r from-emerald-500/20 to-green-500/20 text-emerald-300 border border-emerald-500/30"
                            : "bg-gradient-to-r from-red-500/20 to-rose-500/20 text-red-300 border border-red-500/30"
                        }`}>
                          {user.is_active ? "Active" : "Inactive"}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-300">
                        {formatDate(user.created_at)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-300">
                        {user.last_login_at ? formatDate(user.last_login_at) : "-"}
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center justify-center gap-1 opacity-70 group-hover:opacity-100 transition-opacity">
                          {user.role !== "admin" && (
                            <>
                              <button
                                onClick={() => openEditModal(user)}
                                className="p-2 text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition-all"
                                title="Edit role"
                              >
                                <Edit size={18} />
                              </button>
                              <button
                                onClick={() => handleToggleActive(user.id)}
                                className={`p-2 rounded-lg transition-all ${
                                  user.is_active
                                    ? "text-slate-400 hover:text-amber-400 hover:bg-amber-500/10"
                                    : "text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10"
                                }`}
                                title={user.is_active ? "Deactivate" : "Activate"}
                              >
                                {user.is_active ? <EyeOff size={18} /> : <Eye size={18} />}
                              </button>
                              <button
                                onClick={() => handleDeleteUser(user.id, user.email)}
                                className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                                title="Delete"
                              >
                                <Trash2 size={18} />
                              </button>
                            </>
                          )}
                          {user.role === "admin" && (
                            <span className="text-xs text-slate-500">Protected admin</span>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Create User Modal */}
        {showModal && (
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-slate-900 border border-slate-700 rounded-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-white">New User</h2>
                <button
                  onClick={() => setShowModal(false)}
                  className="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition-all"
                >
                  <X size={20} />
                </button>
              </div>

              <form onSubmit={handleCreateUser} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Full Name
                  </label>
                  <input
                    type="text"
                    value={formData.full_name}
                    onChange={(e) => setFormData({ ...formData, full_name: e.target.value })}
                    required
                    className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all"
                    placeholder="John Doe"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Email (Username)
                  </label>
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                    className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all"
                    placeholder="user@company.com"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Password
                  </label>
                  <div className="relative">
                    <input
                      type={showPassword ? "text" : "password"}
                      value={formData.password}
                      onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      required
                      minLength={8}
                      className="w-full px-4 py-2.5 pr-12 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all"
                      placeholder="Minimum 8 characters"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white transition-colors"
                    >
                      {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                  <p className="text-xs text-slate-500 mt-1">
                    Must contain uppercase, lowercase, numbers and special characters
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
                    User Role
                  </label>
                  <div className="grid grid-cols-1 gap-2 max-h-48 overflow-y-auto pr-2">
                    {ROLES.map((role) => (
                      <label
                        key={role.value}
                        className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                          formData.role === role.value
                            ? "bg-blue-500/20 border-blue-500/50"
                            : "bg-slate-800/50 border-slate-700 hover:border-slate-600"
                        }`}
                      >
                        <input
                          type="radio"
                          name="createRole"
                          value={role.value}
                          checked={formData.role === role.value}
                          onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                          className="mt-1 w-4 h-4 text-blue-500 bg-slate-800 border-slate-600 focus:ring-blue-500"
                        />
                        <div className="flex-1 min-w-0">
                          <div className={`text-sm font-medium ${formData.role === role.value ? "text-blue-300" : "text-white"}`}>
                            {role.label}
                          </div>
                          <div className="text-xs text-slate-400 mt-0.5">{role.description}</div>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>

                {/* Collar Type Info for HR roles */}
                {formData.role?.startsWith('hr') && (
                  <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
                    <p className="text-sm text-blue-300 font-medium mb-2">Access by Employee Type:</p>
                    <ul className="text-xs text-slate-400 space-y-1">
                      {formData.role === 'hr' && (
                        <li>• <span className="text-white">HR - All:</span> Sees employees of all types</li>
                      )}
                      {formData.role === 'hr_blue_gray' && (
                        <>
                          <li>• <span className="text-sky-300">Blue Collar:</span> Operators, workers, plant personnel</li>
                          <li>• <span className="text-indigo-300">Gray Collar:</span> Technicians, maintenance, support</li>
                        </>
                      )}
                      {formData.role === 'hr_white' && (
                        <li>• <span className="text-teal-300">White Collar:</span> Administrative, managers, office</li>
                      )}
                      {formData.role === 'hr_and_pr' && (
                        <li>• <span className="text-violet-300">HR + Payroll:</span> Full access to HR and payroll calculation</li>
                      )}
                    </ul>
                  </div>
                )}

                <div className="flex gap-3 pt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setShowModal(false)}
                    className="flex-1 border-slate-600 text-slate-300 hover:bg-slate-700"
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="flex-1 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700"
                  >
                    {isSubmitting ? "Creating..." : "Create User"}
                  </Button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Edit User Modal */}
        {showEditModal && editingUser && (
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-slate-900 border border-slate-700 rounded-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-white">Edit User</h2>
                <button
                  onClick={() => { setShowEditModal(false); setEditingUser(null); }}
                  className="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition-all"
                >
                  <X size={20} />
                </button>
              </div>

              <form onSubmit={handleEditUser} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Email (Username)
                  </label>
                  <input
                    type="email"
                    value={editingUser.email}
                    disabled
                    className="w-full px-4 py-2.5 bg-slate-800/50 border border-slate-700 rounded-lg text-slate-400 cursor-not-allowed"
                  />
                  <p className="text-xs text-slate-500 mt-1">Email cannot be modified</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Full Name
                  </label>
                  <input
                    type="text"
                    value={editFormData.full_name}
                    onChange={(e) => setEditFormData({ ...editFormData, full_name: e.target.value })}
                    required
                    className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    New Password (optional)
                  </label>
                  <div className="relative">
                    <input
                      type={showEditPassword ? "text" : "password"}
                      value={editFormData.password || ""}
                      onChange={(e) => setEditFormData({ ...editFormData, password: e.target.value })}
                      minLength={8}
                      className="w-full px-4 py-2.5 pr-12 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all"
                      placeholder="Leave empty to keep current"
                    />
                    <button
                      type="button"
                      onClick={() => setShowEditPassword(!showEditPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white transition-colors"
                    >
                      {showEditPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                  <p className="text-xs text-slate-500 mt-1">
                    Minimum 8 characters. Leave empty to keep current password.
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
                    User Role
                  </label>
                  <div className="grid grid-cols-1 gap-2 max-h-48 overflow-y-auto pr-2">
                    {ROLES.map((role) => (
                      <label
                        key={role.value}
                        className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                          editFormData.role === role.value
                            ? "bg-blue-500/20 border-blue-500/50"
                            : "bg-slate-800/50 border-slate-700 hover:border-slate-600"
                        }`}
                      >
                        <input
                          type="radio"
                          name="editRole"
                          value={role.value}
                          checked={editFormData.role === role.value}
                          onChange={(e) => setEditFormData({ ...editFormData, role: e.target.value })}
                          className="mt-1 w-4 h-4 text-blue-500 bg-slate-800 border-slate-600 focus:ring-blue-500"
                        />
                        <div className="flex-1 min-w-0">
                          <div className={`text-sm font-medium ${editFormData.role === role.value ? "text-blue-300" : "text-white"}`}>
                            {role.label}
                          </div>
                          <div className="text-xs text-slate-400 mt-0.5">{role.description}</div>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>

                {/* Collar Type Info for HR roles */}
                {editFormData.role?.startsWith('hr') && (
                  <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
                    <p className="text-sm text-blue-300 font-medium mb-2">Access by Employee Type:</p>
                    <ul className="text-xs text-slate-400 space-y-1">
                      {editFormData.role === 'hr' && (
                        <li>• <span className="text-white">HR - All:</span> Sees employees of all types</li>
                      )}
                      {editFormData.role === 'hr_blue_gray' && (
                        <>
                          <li>• <span className="text-sky-300">Blue Collar:</span> Operators, workers, plant personnel</li>
                          <li>• <span className="text-indigo-300">Gray Collar:</span> Technicians, maintenance, support</li>
                        </>
                      )}
                      {editFormData.role === 'hr_white' && (
                        <li>• <span className="text-teal-300">White Collar:</span> Administrative, managers, office</li>
                      )}
                      {editFormData.role === 'hr_and_pr' && (
                        <li>• <span className="text-violet-300">HR + Payroll:</span> Full access to HR and payroll calculation</li>
                      )}
                    </ul>
                  </div>
                )}

                <div className="flex gap-3 pt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => { setShowEditModal(false); setEditingUser(null); }}
                    className="flex-1 border-slate-600 text-slate-300 hover:bg-slate-700"
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="flex-1 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700"
                  >
                    {isSubmitting ? "Saving..." : "Save Changes"}
                  </Button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
