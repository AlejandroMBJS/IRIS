/**
 * @file app/dashboard/page.tsx
 * @description Main dashboard with payroll system overview and statistics
 *
 * USER PERSPECTIVE:
 *   - View key metrics: total employees, active/inactive, payroll totals
 *   - See employee distribution by collar type (white, blue, gray)
 *   - View recent payroll periods and their status
 *   - Quick access to recent employees and quick actions
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Statistics displayed, card layouts, quick action buttons
 *   CAUTION: Calculations depend on data structure from API, validate format changes
 *   DO NOT modify: API client methods without updating interface definitions
 *
 * KEY COMPONENTS:
 *   - DashboardLayout: Standard layout wrapper with navigation
 *   - KPI cards: Employee stats, payroll totals, salary metrics
 *   - Period list: Recent payroll periods with status
 *   - Employee table: Recent employees with details
 *
 * API ENDPOINTS USED:
 *   - GET /api/employees (via employeeApi.getEmployees)
 *   - GET /api/payroll/periods (via payrollApi.getPeriods)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Users, DollarSign, FileText, TrendingUp, Calendar,
  Building2, Briefcase, Clock, AlertCircle, CheckCircle2,
  ArrowUpRight, ArrowDownRight, ChevronRight, Activity
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { isAuthenticated, getCurrentUser, getUserRole, isTeamFocusedRole } from "@/lib/auth"
import { employeeApi, payrollApi, Employee, PayrollPeriod as ApiPayrollPeriod } from "@/lib/api-client"
import { TeamDashboard } from "@/components/dashboard/team-dashboard"
import { DashboardLayout } from "@/components/layout/dashboard-layout"

// Extended interface with optional dashboard fields
interface PayrollPeriod extends ApiPayrollPeriod {
  total_gross?: number
  total_net?: number
  total_deductions?: number
}

export default function DashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [periods, setPeriods] = useState<PayrollPeriod[]>([])
  const [loading, setLoading] = useState(true)
  const [isTeamRole, setIsTeamRole] = useState(false)

  // Calculate statistics from real data
  const stats = {
    totalEmployees: employees.length,
    activeEmployees: employees.filter(e => e.employment_status === "active").length,
    whiteCollar: employees.filter(e => e.collar_type === "white_collar").length,
    blueCollar: employees.filter(e => e.collar_type === "blue_collar").length,
    grayCollar: employees.filter(e => e.collar_type === "gray_collar").length,
    sindicalizado: employees.filter(e => e.is_sindicalizado).length,
    totalDailySalary: employees.reduce((sum, e) => sum + (e.daily_salary || 0), 0),
    avgSalary: employees.length > 0 ? employees.reduce((sum, e) => sum + (e.daily_salary || 0), 0) / employees.length : 0,
    openPeriods: periods.filter(p => p.status === "open").length,
    totalGross: periods.reduce((sum, p) => sum + (p.total_gross || 0), 0),
    totalNet: periods.reduce((sum, p) => sum + (p.total_net || 0), 0),
    totalDeductions: periods.reduce((sum, p) => sum + (p.total_deductions || 0), 0),
  }

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    const currentUser = getCurrentUser()
    setUser(currentUser)

    // Redirect to role-specific dashboard
    const role = getUserRole()
    if (role === "supervisor") {
      router.push("/dashboard/supervisor")
      return
    }
    if (role === "manager") {
      router.push("/dashboard/manager")
      return
    }
    if (["hr", "hr_and_pr", "hr_blue_gray", "hr_white"].includes(role || "")) {
      router.push("/dashboard/hr")
      return
    }

    setIsTeamRole(isTeamFocusedRole())
    loadDashboardData()
  }, [router])

  async function loadDashboardData() {
    try {
      const [empResponse, periodData] = await Promise.all([
        employeeApi.getEmployees().catch(() => ({ employees: [] })),
        payrollApi.getPeriods().catch(() => []),
      ])

      // Get employees array from response
      setEmployees(empResponse.employees || [])
      setPeriods(Array.isArray(periodData) ? periodData : [])
    } catch (error) {
      console.error("Error loading dashboard data:", error)
    } finally {
      setLoading(false)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
      minimumFractionDigits: 2,
    }).format(amount || 0)
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-"
    try {
      return new Date(dateStr).toLocaleDateString("es-MX", {
        day: "2-digit",
        month: "short",
        year: "numeric"
      })
    } catch {
      return dateStr
    }
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

  // Show team-focused dashboard for supervisor/manager roles
  if (isTeamRole) {
    return (
      <DashboardLayout>
        <TeamDashboard employees={employees} user={user} />
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-8">
        {/* Page Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white">Dashboard</h1>
            <p className="text-slate-400 mt-1">
              Payroll system overview • {new Date().toLocaleDateString("es-MX", { weekday: "long", year: "numeric", month: "long", day: "numeric" })}
            </p>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={() => router.push("/payroll/periods")}
              className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-lg shadow-blue-500/20"
            >
              <Calendar size={18} className="mr-2" />
              View Periods
            </Button>
            <Button
              onClick={() => router.push("/employees")}
              variant="outline"
              className="border-slate-600 text-slate-300 hover:bg-slate-800"
            >
              <Users size={18} className="mr-2" />
              View Employees
            </Button>
          </div>
        </div>

        {/* Main KPI Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          {/* Total Employees */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-blue-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Active Employees</p>
                <p className="text-4xl font-bold text-white">{stats.activeEmployees}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-emerald-400 flex items-center">
                    <ArrowUpRight size={14} className="mr-1" />
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
                <span className="text-slate-500">White Collar: {stats.whiteCollar}</span>
                <span className="text-slate-500">Blue: {stats.blueCollar}</span>
                <span className="text-slate-500">Gray: {stats.grayCollar}</span>
              </div>
            </div>
          </div>

          {/* Total Payroll */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-emerald-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Total Payroll</p>
                <p className="text-4xl font-bold text-white">{formatCurrency(stats.totalGross)}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-emerald-400 flex items-center">
                    <CheckCircle2 size={14} className="mr-1" />
                    Net: {formatCurrency(stats.totalNet)}
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/30">
                <DollarSign size={24} className="text-white" />
              </div>
            </div>
            <div className="mt-4 pt-4 border-t border-slate-700/50">
              <div className="flex justify-between text-xs">
                <span className="text-slate-500">Deductions: {formatCurrency(stats.totalDeductions)}</span>
              </div>
            </div>
          </div>

          {/* Open Periods */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-amber-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Open Periods</p>
                <p className="text-4xl font-bold text-white">{stats.openPeriods}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-amber-400 flex items-center">
                    <Clock size={14} className="mr-1" />
                    Pending closure
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl shadow-lg shadow-amber-500/30">
                <FileText size={24} className="text-white" />
              </div>
            </div>
            <div className="mt-4 pt-4 border-t border-slate-700/50">
              <div className="flex justify-between text-xs">
                <span className="text-slate-500">Total periods: {periods.length}</span>
              </div>
            </div>
          </div>

          {/* Daily Salary Total */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-purple-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Avg. Daily Salary</p>
                <p className="text-4xl font-bold text-white">{formatCurrency(stats.avgSalary)}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-purple-400 flex items-center">
                    <TrendingUp size={14} className="mr-1" />
                    Total: {formatCurrency(stats.totalDailySalary)}
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl shadow-lg shadow-purple-500/30">
                <Activity size={24} className="text-white" />
              </div>
            </div>
            <div className="mt-4 pt-4 border-t border-slate-700/50">
              <div className="flex justify-between text-xs">
                <span className="text-slate-500">Unionized: {stats.sindicalizado}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Two Column Layout */}
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
          {/* Employee Distribution */}
          <div className="xl:col-span-1 bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
            <div className="p-6 border-b border-slate-700/50">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <Building2 size={20} className="text-blue-400" />
                Distribution by Type
              </h3>
            </div>
            <div className="p-6 space-y-4">
              {/* White Collar */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-300 flex items-center gap-2">
                    <span className="w-3 h-3 rounded-full bg-blue-500"></span>
                    White Collar (Administrative)
                  </span>
                  <span className="text-white font-medium">{stats.whiteCollar}</span>
                </div>
                <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-blue-500 to-blue-400 rounded-full transition-all duration-500"
                    style={{ width: `${stats.totalEmployees > 0 ? (stats.whiteCollar / stats.totalEmployees) * 100 : 0}%` }}
                  />
                </div>
              </div>

              {/* Blue Collar */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-300 flex items-center gap-2">
                    <span className="w-3 h-3 rounded-full bg-indigo-500"></span>
                    Blue Collar (Unionized)
                  </span>
                  <span className="text-white font-medium">{stats.blueCollar}</span>
                </div>
                <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 rounded-full transition-all duration-500"
                    style={{ width: `${stats.totalEmployees > 0 ? (stats.blueCollar / stats.totalEmployees) * 100 : 0}%` }}
                  />
                </div>
              </div>

              {/* Gray Collar */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-300 flex items-center gap-2">
                    <span className="w-3 h-3 rounded-full bg-slate-400"></span>
                    Gray Collar (Non-Unionized)
                  </span>
                  <span className="text-white font-medium">{stats.grayCollar}</span>
                </div>
                <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-slate-400 to-slate-300 rounded-full transition-all duration-500"
                    style={{ width: `${stats.totalEmployees > 0 ? (stats.grayCollar / stats.totalEmployees) * 100 : 0}%` }}
                  />
                </div>
              </div>

              <div className="pt-4 mt-4 border-t border-slate-700/50">
                <div className="grid grid-cols-2 gap-4 text-center">
                  <div className="p-3 bg-slate-800/50 rounded-lg">
                    <p className="text-2xl font-bold text-white">{stats.sindicalizado}</p>
                    <p className="text-xs text-slate-400">Unionized</p>
                  </div>
                  <div className="p-3 bg-slate-800/50 rounded-lg">
                    <p className="text-2xl font-bold text-white">{stats.totalEmployees - stats.sindicalizado}</p>
                    <p className="text-xs text-slate-400">Non-Unionized</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Recent Payroll Periods */}
          <div className="xl:col-span-2 bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
            <div className="p-6 border-b border-slate-700/50 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <Calendar size={20} className="text-emerald-400" />
                Payroll Periods
              </h3>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => router.push("/payroll/periods")}
                className="text-slate-400 hover:text-white"
              >
                View all
                <ChevronRight size={16} className="ml-1" />
              </Button>
            </div>
            <div className="divide-y divide-slate-700/50">
              {periods.length === 0 ? (
                <div className="p-12 text-center">
                  <FileText size={48} className="mx-auto text-slate-600 mb-4" />
                  <p className="text-slate-400">No payroll periods</p>
                  <p className="text-sm text-slate-500 mt-2">Periods are automatically generated by the system</p>
                </div>
              ) : (
                periods.slice(0, 4).map((period) => (
                  <div
                    key={period.id}
                    className="p-4 hover:bg-slate-800/30 transition-colors cursor-pointer"
                    onClick={() => router.push(`/payroll/${period.id}`)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className={`p-3 rounded-xl ${
                          period.frequency === "weekly"
                            ? "bg-blue-500/20 text-blue-400"
                            : period.frequency === "biweekly"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-purple-500/20 text-purple-400"
                        }`}>
                          <Calendar size={20} />
                        </div>
                        <div>
                          <p className="text-white font-medium">{period.period_code}</p>
                          <p className="text-sm text-slate-400">
                            {formatDate(period.start_date)} - {formatDate(period.end_date)}
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <p className="text-white font-semibold">{formatCurrency(period.total_gross || 0)}</p>
                        <span className={`inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full ${
                          period.status === "open"
                            ? "bg-amber-500/20 text-amber-400"
                            : period.status === "paid"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-slate-500/20 text-slate-400"
                        }`}>
                          {period.status === "open" ? "Open" : period.status === "paid" ? "Paid" : period.status}
                        </span>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6">
          <h3 className="text-lg font-semibold text-white mb-6 flex items-center gap-2">
            <Briefcase size={20} className="text-amber-400" />
            Quick Actions
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <button
              onClick={() => router.push("/employees/new")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-blue-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <Users size={24} className="text-blue-400" />
              </div>
              <span className="text-white font-medium">New Employee</span>
              <span className="text-xs text-slate-400 mt-1">Add to system</span>
            </button>

            <button
              onClick={() => router.push("/payroll/periods")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-emerald-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <Calendar size={24} className="text-emerald-400" />
              </div>
              <span className="text-white font-medium">View Periods</span>
              <span className="text-xs text-slate-400 mt-1">Check payrolls</span>
            </button>

            <button
              onClick={() => router.push("/employees")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-purple-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <FileText size={24} className="text-purple-400" />
              </div>
              <span className="text-white font-medium">Employee List</span>
              <span className="text-xs text-slate-400 mt-1">View directory</span>
            </button>

            <button
              onClick={() => router.push("/payroll")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-amber-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <DollarSign size={24} className="text-amber-400" />
              </div>
              <span className="text-white font-medium">Calculate Payroll</span>
              <span className="text-xs text-slate-400 mt-1">Process payments</span>
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
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Frequency</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Daily Salary</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {employees.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center">
                      <Users size={48} className="mx-auto text-slate-600 mb-4" />
                      <p className="text-slate-400">No employees registered</p>
                      <Button
                        onClick={() => router.push("/employees/new")}
                        className="mt-4 bg-blue-600 hover:bg-blue-700"
                      >
                        Add Employee
                      </Button>
                    </td>
                  </tr>
                ) : (
                  employees.slice(0, 5).map((emp) => (
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
                        <span className="text-slate-300 text-sm capitalize">
                          {emp.pay_frequency === "weekly" ? "Weekly" : emp.pay_frequency === "biweekly" ? "Biweekly" : emp.pay_frequency || "N/A"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white font-medium">{formatCurrency(emp.daily_salary)}</span>
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
