/**
 * @file app/dashboard/hr/page.tsx
 * @description HR Dashboard - Employee management, incidences, approvals, HR calendar
 *
 * USER PERSPECTIVE:
 *   - View employees (filtered by collar type for hr_white / hr_blue_gray)
 *   - Manage incidences and absences
 *   - HR calendar and announcements
 *   - Employee lifecycle management
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Add HR-specific widgets, employee lifecycle stats
 *   CAUTION: hr_white sees only white_collar, hr_blue_gray sees blue/gray collar
 *   DO NOT modify: Access control - HR roles only
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Users, Calendar, FileText, AlertCircle, CheckCircle2,
  TrendingUp, UserPlus, Clock, ChevronRight, Briefcase, Bell
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { isAuthenticated, getCurrentUser, isHR, getUserRole } from "@/lib/auth"
import { employeeApi, Employee } from "@/lib/api-client"

export default function HRDashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [userRole, setUserRole] = useState<string>("")

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    const currentUser = getCurrentUser()
    const role = getUserRole()

    if (!isHR()) {
      router.push("/dashboard")
      return
    }

    setUser(currentUser)
    setUserRole(role || "")
    loadDashboardData()
  }, [router])

  async function loadDashboardData() {
    try {
      // Get employees - backend automatically filters by collar type for hr_white/hr_blue_gray
      const empResponse = await employeeApi.getEmployees().catch(() => ({ employees: [] }))
      setEmployees(empResponse.employees || [])
    } catch (error) {
      console.error("Error loading dashboard data:", error)
    } finally {
      setLoading(false)
    }
  }

  const stats = {
    totalEmployees: employees.length,
    activeEmployees: employees.filter(e => e.employment_status === "active").length,
    whiteCollar: employees.filter(e => e.collar_type === "white_collar").length,
    blueCollar: employees.filter(e => e.collar_type === "blue_collar").length,
    grayCollar: employees.filter(e => e.collar_type === "gray_collar").length,
    pendingApprovals: 0, // TODO: Fetch from approvals API
    recentHires: employees.filter(e => {
      const hireDate = new Date(e.hire_date)
      const thirtyDaysAgo = new Date()
      thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)
      return hireDate >= thirtyDaysAgo
    }).length,
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
      minimumFractionDigits: 2,
    }).format(amount || 0)
  }

  const getRoleDisplay = () => {
    if (userRole === "hr_white") return "HR - White Collar"
    if (userRole === "hr_blue_gray") return "HR - Blue/Gray Collar"
    if (userRole === "hr_and_pr") return "HR & Payroll"
    return "Human Resources"
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Loading dashboard...</p>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-8">
        {/* Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white">HR Dashboard</h1>
            <p className="text-slate-400 mt-1">
              Welcome, {user?.full_name || user?.email} - {getRoleDisplay()}
            </p>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={() => router.push("/employees/new")}
              className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-lg shadow-blue-500/20"
            >
              <UserPlus size={18} className="mr-2" />
              New Employee
            </Button>
            <Button
              onClick={() => router.push("/hr/calendar")}
              variant="outline"
              className="border-slate-600 text-slate-300 hover:bg-slate-800"
            >
              <Calendar size={18} className="mr-2" />
              HR Calendar
            </Button>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          {/* Total Employees */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-blue-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Employees</p>
                <p className="text-4xl font-bold text-white">{stats.activeEmployees}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-emerald-400 flex items-center">
                    <CheckCircle2 size={14} className="mr-1" />
                    {stats.totalEmployees} total
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl shadow-lg shadow-blue-500/30">
                <Users size={24} className="text-white" />
              </div>
            </div>
            <div className="mt-4 pt-4 border-t border-slate-700/50">
              <div className="flex justify-between text-xs">
                <span className="text-slate-500">White: {stats.whiteCollar}</span>
                <span className="text-slate-500">Blue: {stats.blueCollar}</span>
                <span className="text-slate-500">Gray: {stats.grayCollar}</span>
              </div>
            </div>
          </div>

          {/* Recent Hires */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-emerald-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Hires</p>
                <p className="text-4xl font-bold text-white">{stats.recentHires}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-emerald-400 flex items-center">
                    <TrendingUp size={14} className="mr-1" />
                    Last 30 days
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/30">
                <UserPlus size={24} className="text-white" />
              </div>
            </div>
          </div>

          {/* Pending Approvals */}
          <div
            className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-amber-500/10 cursor-pointer"
            onClick={() => router.push("/approvals")}
          >
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Approvals</p>
                <p className="text-4xl font-bold text-white">{stats.pendingApprovals}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-amber-400 flex items-center">
                    <Clock size={14} className="mr-1" />
                    Pending
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl shadow-lg shadow-amber-500/30">
                <AlertCircle size={24} className="text-white" />
              </div>
            </div>
          </div>

          {/* Incidences */}
          <div
            className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-purple-500/10 cursor-pointer"
            onClick={() => router.push("/incidences")}
          >
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Incidences</p>
                <p className="text-4xl font-bold text-white">0</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-purple-400 flex items-center">
                    <FileText size={14} className="mr-1" />
                    This month
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl shadow-lg shadow-purple-500/30">
                <FileText size={24} className="text-white" />
              </div>
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6">
          <h3 className="text-lg font-semibold text-white mb-6 flex items-center gap-2">
            <Briefcase size={20} className="text-amber-400" />
            HR Actions
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <button
              onClick={() => router.push("/employees/new")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-blue-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <UserPlus size={24} className="text-blue-400" />
              </div>
              <span className="text-white font-medium">New Employee</span>
              <span className="text-xs text-slate-400 mt-1">Add to system</span>
            </button>

            <button
              onClick={() => router.push("/incidences")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-emerald-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <FileText size={24} className="text-emerald-400" />
              </div>
              <span className="text-white font-medium">Incidences</span>
              <span className="text-xs text-slate-400 mt-1">Manage absences</span>
            </button>

            <button
              onClick={() => router.push("/hr/calendar")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-purple-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <Calendar size={24} className="text-purple-400" />
              </div>
              <span className="text-white font-medium">Calendar</span>
              <span className="text-xs text-slate-400 mt-1">HR Events</span>
            </button>

            <button
              onClick={() => router.push("/announcements")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-amber-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <Bell size={24} className="text-amber-400" />
              </div>
              <span className="text-white font-medium">Announcements</span>
              <span className="text-xs text-slate-400 mt-1">Communications</span>
            </button>
          </div>
        </div>

        {/* Recent Employees */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
          <div className="p-6 border-b border-slate-700/50 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Users size={20} className="text-blue-400" />
              Recent Employees
            </h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push("/employees")}
              className="text-slate-400 hover:text-white"
            >
              View all
              <ChevronRight size={16} className="ml-1" />
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Employee</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Type</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Hire Date</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {employees.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-6 py-12 text-center">
                      <Users size={48} className="mx-auto text-slate-600 mb-4" />
                      <p className="text-slate-400">No employees registered</p>
                    </td>
                  </tr>
                ) : (
                  employees.slice(0, 10).map((emp) => (
                    <tr
                      key={emp.id}
                      className="hover:bg-slate-800/30 transition-colors cursor-pointer"
                      onClick={() => router.push(`/employees/${emp.id}`)}
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white font-semibold">
                            {emp.first_name?.[0]}{emp.last_name?.[0]}
                          </div>
                          <div>
                            <p className="text-white font-medium">{emp.full_name}</p>
                            <p className="text-sm text-slate-400">{emp.employee_number}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-full ${
                          emp.collar_type === "white_collar"
                            ? "bg-blue-500/20 text-blue-400"
                            : emp.collar_type === "blue_collar"
                            ? "bg-indigo-500/20 text-indigo-400"
                            : "bg-slate-500/20 text-slate-400"
                        }`}>
                          {emp.collar_type?.replace("_", " ") || "N/A"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-slate-300 text-sm">
                          {new Date(emp.hire_date).toLocaleDateString("es-MX")}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-full ${
                          emp.employment_status === "active"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-slate-500/20 text-slate-400"
                        }`}>
                          <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${
                            emp.employment_status === "active" ? "bg-emerald-400" : "bg-slate-400"
                          }`}></span>
                          {emp.employment_status === "active" ? "Active" : emp.employment_status}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </DashboardLayout>
  )
}
