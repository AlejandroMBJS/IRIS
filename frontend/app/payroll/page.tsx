/**
 * @file app/payroll/page.tsx
 * @description Payroll overview page showing periods organized by status
 *
 * USER PERSPECTIVE:
 *   - View all payroll periods organized by status (open, closed, posted)
 *   - Quick stats showing count of periods in each status
 *   - Quick actions: manage periods, view employees, configure system
 *   - Click on open periods to process payroll
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Status filters, quick action buttons, card layouts
 *   CAUTION: Status logic (OPEN/CLOSED/POSTED) must match backend enum
 *   DO NOT modify: Period status filtering without backend coordination
 *
 * KEY COMPONENTS:
 *   - Status cards: Count of periods by status (open, closed, posted)
 *   - Period lists: Grouped by status with action buttons
 *   - Quick actions: Navigation to related pages
 *
 * API ENDPOINTS USED:
 *   - GET /api/payroll/periods (via payrollApi.getPeriods)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Calendar, Play, FileText, DollarSign, ArrowRight,
  Users, TrendingUp, Clock, CheckCircle2, AlertCircle,
  Calculator, RefreshCw, ChevronRight, Wallet, Briefcase,
  HardHat, UserCog
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { payrollApi, PayrollPeriod, ApiError, employeeApi, Employee } from "@/lib/api-client"
import { canProcessPayroll } from "@/lib/auth"

interface PayrollSummaryByCollar {
  whiteCollar: { count: number; totalSalary: number; frequency: string }
  blueCollar: { count: number; totalSalary: number; frequency: string }
  grayCollar: { count: number; totalSalary: number; frequency: string }
  total: { count: number; totalSalary: number }
}

export default function PayrollPage() {
  const router = useRouter()
  const [periods, setPeriods] = useState<PayrollPeriod[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [payrollSummary, setPayrollSummary] = useState<PayrollSummaryByCollar>({
    whiteCollar: { count: 0, totalSalary: 0, frequency: "biweekly" },
    blueCollar: { count: 0, totalSalary: 0, frequency: "weekly" },
    grayCollar: { count: 0, totalSalary: 0, frequency: "weekly" },
    total: { count: 0, totalSalary: 0 }
  })

  useEffect(() => {
    // Check authorization - only admin and payroll roles can access
    if (!canProcessPayroll()) {
      router.push("/dashboard")
      return
    }
    loadData()
  }, [router])

  async function loadData() {
    try {
      setLoading(true)
      setError(null)

      // First, generate current periods if they don't exist
      try {
        await payrollApi.generateCurrentPeriods()
      } catch (genErr) {
        // Ignore errors - periods might already exist
        console.log("Period generation:", genErr)
      }

      const [periodsData, employeesResponse] = await Promise.all([
        payrollApi.getPeriods(),
        employeeApi.getEmployees()
      ])
      setPeriods(Array.isArray(periodsData) ? periodsData : [])

      const emps = employeesResponse.employees || []
      setEmployees(emps)

      // Calculate payroll summary by collar type (only active employees)
      const activeEmployees = emps.filter(e => e.employment_status === "active")

      const whiteCollar = activeEmployees.filter(e => e.collar_type === "white_collar")
      const blueCollar = activeEmployees.filter(e => e.collar_type === "blue_collar")
      const grayCollar = activeEmployees.filter(e => e.collar_type === "gray_collar")

      setPayrollSummary({
        whiteCollar: {
          count: whiteCollar.length,
          totalSalary: whiteCollar.reduce((sum, e) => sum + (e.daily_salary || 0), 0),
          frequency: "biweekly"
        },
        blueCollar: {
          count: blueCollar.length,
          totalSalary: blueCollar.reduce((sum, e) => sum + (e.daily_salary || 0), 0),
          frequency: "weekly"
        },
        grayCollar: {
          count: grayCollar.length,
          totalSalary: grayCollar.reduce((sum, e) => sum + (e.daily_salary || 0), 0),
          frequency: "weekly"
        },
        total: {
          count: activeEmployees.length,
          totalSalary: activeEmployees.reduce((sum, e) => sum + (e.daily_salary || 0), 0)
        }
      })
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load payroll data")
      }
      setPeriods([])
    } finally {
      setLoading(false)
    }
  }

  async function loadPeriods() {
    loadData()
  }

  const openPeriods = periods.filter((p) => p.status === "open")
  const closedPeriods = periods.filter((p) => p.status === "closed" || p.status === "calculated")
  const pendingPaymentPeriods = periods.filter((p) => p.status === "approved")
  const postedPeriods = periods.filter((p) => p.status === "paid")

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-"
    try {
      return new Date(dateStr).toLocaleDateString("en-US", {
        day: "2-digit",
        month: "short",
        year: "numeric"
      })
    } catch {
      return dateStr
    }
  }

  const getFrequencyBadge = (frequency: string) => {
    switch (frequency?.toLowerCase()) {
      case "weekly":
        return "bg-gradient-to-r from-blue-500/20 to-cyan-500/20 text-blue-300 border border-blue-500/30"
      case "biweekly":
        return "bg-gradient-to-r from-purple-500/20 to-pink-500/20 text-purple-300 border border-purple-500/30"
      case "monthly":
        return "bg-gradient-to-r from-amber-500/20 to-orange-500/20 text-amber-300 border border-amber-500/30"
      default:
        return "bg-slate-700/50 text-slate-400"
    }
  }

  const getFrequencyLabel = (frequency: string) => {
    switch (frequency?.toLowerCase()) {
      case "weekly": return "Weekly"
      case "biweekly": return "Biweekly"
      case "monthly": return "Monthly"
      default: return frequency
    }
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              Payroll
            </h1>
            <p className="text-slate-400 mt-1">Process and manage your employees' payroll</p>
          </div>
          <div className="flex items-center gap-3">
            <Button
              onClick={loadPeriods}
              variant="outline"
              size="sm"
              className="border-slate-600 text-slate-300 hover:bg-slate-700"
            >
              <RefreshCw size={16} className="mr-2" />
              Refresh
            </Button>
            <Button
              onClick={() => router.push("/payroll/periods")}
              className="bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 shadow-lg shadow-blue-500/25"
            >
              <Calendar size={20} className="mr-2" />
              Manage Periods
            </Button>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-900/20 border border-red-700/50 rounded-xl p-4 text-red-400 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <AlertCircle size={20} />
              <span>{error}</span>
            </div>
            <button onClick={loadPeriods} className="px-3 py-1 bg-red-500/20 hover:bg-red-500/30 rounded-lg transition-colors">
              Retry
            </button>
          </div>
        )}

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-gradient-to-br from-blue-600/20 to-cyan-600/20 backdrop-blur-sm rounded-xl p-6 border border-blue-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-blue-200 text-sm font-medium">Open Periods</p>
                <p className="text-4xl font-bold text-white mt-2">{openPeriods.length}</p>
                <p className="text-blue-300/70 text-xs mt-1">Ready to process</p>
              </div>
              <div className="p-4 bg-blue-500/20 rounded-xl">
                <Play size={28} className="text-blue-400" />
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-amber-600/20 to-orange-600/20 backdrop-blur-sm rounded-xl p-6 border border-amber-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-amber-200 text-sm font-medium">Pending Approval</p>
                <p className="text-4xl font-bold text-white mt-2">{closedPeriods.length}</p>
                <p className="text-amber-300/70 text-xs mt-1">Require review</p>
              </div>
              <div className="p-4 bg-amber-500/20 rounded-xl">
                <FileText size={28} className="text-amber-400" />
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-purple-600/20 to-pink-600/20 backdrop-blur-sm rounded-xl p-6 border border-purple-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-purple-200 text-sm font-medium">Pending Payment</p>
                <p className="text-4xl font-bold text-white mt-2">{pendingPaymentPeriods.length}</p>
                <p className="text-purple-300/70 text-xs mt-1">Approved, ready to pay</p>
              </div>
              <div className="p-4 bg-purple-500/20 rounded-xl">
                <DollarSign size={28} className="text-purple-400" />
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-emerald-600/20 to-green-600/20 backdrop-blur-sm rounded-xl p-6 border border-emerald-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-emerald-200 text-sm font-medium">Paid</p>
                <p className="text-4xl font-bold text-white mt-2">{postedPeriods.length}</p>
                <p className="text-emerald-300/70 text-xs mt-1">Completed payrolls</p>
              </div>
              <div className="p-4 bg-emerald-500/20 rounded-xl">
                <CheckCircle2 size={28} className="text-emerald-400" />
              </div>
            </div>
          </div>
        </div>

        {/* Payroll Summary by Collar Type */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50">
          <div className="flex items-center justify-between mb-5">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <DollarSign size={20} className="text-emerald-400" />
              Payroll Summary by Employee Type
            </h3>
            <span className="text-sm text-slate-400">Active employees</span>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* White Collar */}
            <div className="bg-gradient-to-br from-purple-600/20 to-indigo-600/20 rounded-xl p-5 border border-purple-500/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2.5 bg-purple-500/20 rounded-lg">
                  <Briefcase size={22} className="text-purple-400" />
                </div>
                <div>
                  <p className="text-purple-200 text-sm font-medium">White Collar</p>
                  <p className="text-purple-400/70 text-xs">Biweekly</p>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Employees</span>
                  <span className="text-2xl font-bold text-white">{payrollSummary.whiteCollar.count}</span>
                </div>
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Daily Salary</span>
                  <span className="text-lg font-semibold text-purple-300">
                    ${payrollSummary.whiteCollar.totalSalary.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>
            </div>

            {/* Blue Collar */}
            <div className="bg-gradient-to-br from-blue-600/20 to-cyan-600/20 rounded-xl p-5 border border-blue-500/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2.5 bg-blue-500/20 rounded-lg">
                  <HardHat size={22} className="text-blue-400" />
                </div>
                <div>
                  <p className="text-blue-200 text-sm font-medium">Blue Collar</p>
                  <p className="text-blue-400/70 text-xs">Weekly</p>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Employees</span>
                  <span className="text-2xl font-bold text-white">{payrollSummary.blueCollar.count}</span>
                </div>
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Daily Salary</span>
                  <span className="text-lg font-semibold text-blue-300">
                    ${payrollSummary.blueCollar.totalSalary.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>
            </div>

            {/* Gray Collar */}
            <div className="bg-gradient-to-br from-slate-600/20 to-zinc-600/20 rounded-xl p-5 border border-slate-500/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2.5 bg-slate-500/20 rounded-lg">
                  <UserCog size={22} className="text-slate-400" />
                </div>
                <div>
                  <p className="text-slate-200 text-sm font-medium">Gray Collar</p>
                  <p className="text-slate-400/70 text-xs">Weekly</p>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Employees</span>
                  <span className="text-2xl font-bold text-white">{payrollSummary.grayCollar.count}</span>
                </div>
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Daily Salary</span>
                  <span className="text-lg font-semibold text-slate-300">
                    ${payrollSummary.grayCollar.totalSalary.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>
            </div>

            {/* Total */}
            <div className="bg-gradient-to-br from-emerald-600/20 to-green-600/20 rounded-xl p-5 border border-emerald-500/30">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2.5 bg-emerald-500/20 rounded-lg">
                  <Users size={22} className="text-emerald-400" />
                </div>
                <div>
                  <p className="text-emerald-200 text-sm font-medium">Grand Total</p>
                  <p className="text-emerald-400/70 text-xs">All types</p>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Employees</span>
                  <span className="text-2xl font-bold text-white">{payrollSummary.total.count}</span>
                </div>
                <div className="flex justify-between items-baseline">
                  <span className="text-slate-400 text-sm">Daily Salary</span>
                  <span className="text-lg font-semibold text-emerald-300">
                    ${payrollSummary.total.totalSalary.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-5 border border-slate-700/50">
          <h3 className="text-sm font-medium text-slate-300 mb-4">Quick Actions</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <button
              onClick={() => router.push("/payroll/periods")}
              className="flex items-center gap-3 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-blue-500/30 transition-all group"
            >
              <div className="p-2 bg-blue-500/20 rounded-lg group-hover:bg-blue-500/30 transition-colors">
                <Calendar size={20} className="text-blue-400" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium text-white">View Periods</p>
                <p className="text-xs text-slate-400">Manage cycles</p>
              </div>
            </button>
            <button
              onClick={() => router.push("/employees")}
              className="flex items-center gap-3 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-emerald-500/30 transition-all group"
            >
              <div className="p-2 bg-emerald-500/20 rounded-lg group-hover:bg-emerald-500/30 transition-colors">
                <Users size={20} className="text-emerald-400" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium text-white">Employees</p>
                <p className="text-xs text-slate-400">View list</p>
              </div>
            </button>
            <button
              onClick={() => router.push("/configuration/payroll-setup")}
              className="flex items-center gap-3 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-purple-500/30 transition-all group"
            >
              <div className="p-2 bg-purple-500/20 rounded-lg group-hover:bg-purple-500/30 transition-colors">
                <Calculator size={20} className="text-purple-400" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium text-white">Configuration</p>
                <p className="text-xs text-slate-400">ISR, IMSS, etc.</p>
              </div>
            </button>
            <button
              onClick={() => router.push("/dashboard")}
              className="flex items-center gap-3 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-amber-500/30 transition-all group"
            >
              <div className="p-2 bg-amber-500/20 rounded-lg group-hover:bg-amber-500/30 transition-colors">
                <TrendingUp size={20} className="text-amber-400" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium text-white">Dashboard</p>
                <p className="text-xs text-slate-400">View summary</p>
              </div>
            </button>
          </div>
        </div>

        {/* Open Periods - Ready to Process */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/20 rounded-lg">
                <Play size={18} className="text-blue-400" />
              </div>
              <h2 className="text-lg font-semibold text-white">Open Periods - Ready to Process</h2>
            </div>
            <span className="px-2.5 py-1 text-xs font-medium bg-blue-500/20 text-blue-300 rounded-full">
              {openPeriods.length} available
            </span>
          </div>
          {loading ? (
            <div className="px-6 py-12 text-center text-slate-400">
              <div className="flex flex-col items-center gap-3">
                <div className="w-8 h-8 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                <span>Loading periods...</span>
              </div>
            </div>
          ) : openPeriods.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <Calendar className="mx-auto mb-4 text-slate-600" size={48} />
              <p className="text-slate-400">No open periods available</p>
              <p className="text-sm text-slate-500 mt-2">Create a new period to start processing payroll</p>
              <Button
                onClick={() => router.push("/payroll/periods")}
                variant="outline"
                size="sm"
                className="mt-4 border-slate-600 hover:border-blue-500/50"
              >
                Go to Periods
              </Button>
            </div>
          ) : (
            <div className="divide-y divide-slate-700/50">
              {openPeriods.map((period) => (
                <div
                  key={period.id}
                  className="px-6 py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors group"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-slate-800/80 rounded-xl group-hover:bg-blue-500/20 transition-colors">
                      <Calendar size={24} className="text-slate-400 group-hover:text-blue-400 transition-colors" />
                    </div>
                    <div>
                      <p className="text-white font-semibold text-lg">
                        {period.year} - Period {period.period_number}
                      </p>
                      <div className="flex items-center gap-3 mt-1">
                        <span className="text-sm text-slate-400">
                          {formatDate(period.start_date)} - {formatDate(period.end_date)}
                        </span>
                        <span className="text-slate-600">|</span>
                        <span className="text-sm text-slate-500">
                          Payment: {formatDate(period.payment_date)}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className={`px-3 py-1.5 text-xs font-medium rounded-lg ${getFrequencyBadge(period.frequency)}`}>
                      {getFrequencyLabel(period.frequency)}
                    </span>
                    <Button
                      onClick={() => router.push(`/payroll/${period.id}`)}
                      className="bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 shadow-lg shadow-blue-500/25"
                    >
                      Process
                      <ArrowRight size={16} className="ml-2" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Pending Approval */}
        {closedPeriods.length > 0 && (
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-amber-500/20 rounded-lg">
                  <Clock size={18} className="text-amber-400" />
                </div>
                <h2 className="text-lg font-semibold text-white">Pending Approval</h2>
              </div>
              <span className="px-2.5 py-1 text-xs font-medium bg-amber-500/20 text-amber-300 rounded-full">
                {closedPeriods.length} pending
              </span>
            </div>
            <div className="divide-y divide-slate-700/50">
              {closedPeriods.slice(0, 5).map((period) => (
                <div
                  key={period.id}
                  onClick={() => router.push(`/payroll/${period.id}`)}
                  className="px-6 py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors cursor-pointer group"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-slate-800/80 rounded-xl group-hover:bg-amber-500/20 transition-colors">
                      <FileText size={24} className="text-slate-400 group-hover:text-amber-400 transition-colors" />
                    </div>
                    <div>
                      <p className="text-white font-semibold">
                        {period.year} - Period {period.period_number}
                      </p>
                      <p className="text-sm text-slate-400">
                        Payment: {formatDate(period.payment_date)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={`px-3 py-1.5 text-xs font-medium rounded-lg ${getFrequencyBadge(period.frequency)}`}>
                      {getFrequencyLabel(period.frequency)}
                    </span>
                    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg bg-amber-500/20 text-amber-300 border border-amber-500/30">
                      <Clock size={12} />
                      Pending
                    </span>
                    <ChevronRight size={20} className="text-slate-500 group-hover:text-amber-400 transition-colors" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Pending Payment */}
        {pendingPaymentPeriods.length > 0 && (
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-purple-500/20 rounded-lg">
                  <DollarSign size={18} className="text-purple-400" />
                </div>
                <h2 className="text-lg font-semibold text-white">Pending Payment</h2>
              </div>
              <span className="px-2.5 py-1 text-xs font-medium bg-purple-500/20 text-purple-300 rounded-full">
                {pendingPaymentPeriods.length} approved
              </span>
            </div>
            <div className="divide-y divide-slate-700/50">
              {pendingPaymentPeriods.slice(0, 5).map((period) => (
                <div
                  key={period.id}
                  onClick={() => router.push(`/payroll/${period.id}`)}
                  className="px-6 py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors cursor-pointer group"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-slate-800/80 rounded-xl group-hover:bg-purple-500/20 transition-colors">
                      <Wallet size={24} className="text-slate-400 group-hover:text-purple-400 transition-colors" />
                    </div>
                    <div>
                      <p className="text-white font-semibold">
                        {period.year} - Period {period.period_number}
                      </p>
                      <p className="text-sm text-slate-400">
                        Payment: {formatDate(period.payment_date)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={`px-3 py-1.5 text-xs font-medium rounded-lg ${getFrequencyBadge(period.frequency)}`}>
                      {getFrequencyLabel(period.frequency)}
                    </span>
                    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg bg-purple-500/20 text-purple-300 border border-purple-500/30">
                      <CheckCircle2 size={12} />
                      Approved
                    </span>
                    <ChevronRight size={20} className="text-slate-500 group-hover:text-purple-400 transition-colors" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Recent Posted Payrolls */}
        {postedPeriods.length > 0 && (
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-emerald-500/20 rounded-lg">
                  <CheckCircle2 size={18} className="text-emerald-400" />
                </div>
                <h2 className="text-lg font-semibold text-white">Paid Payrolls</h2>
              </div>
              <span className="px-2.5 py-1 text-xs font-medium bg-emerald-500/20 text-emerald-300 rounded-full">
                {postedPeriods.length} completed
              </span>
            </div>
            <div className="divide-y divide-slate-700/50">
              {postedPeriods.slice(0, 5).map((period) => (
                <div
                  key={period.id}
                  onClick={() => router.push(`/payroll/${period.id}`)}
                  className="px-6 py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors cursor-pointer group"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-slate-800/80 rounded-xl group-hover:bg-emerald-500/20 transition-colors">
                      <Wallet size={24} className="text-slate-400 group-hover:text-emerald-400 transition-colors" />
                    </div>
                    <div>
                      <p className="text-white font-semibold">
                        {period.year} - Period {period.period_number}
                      </p>
                      <p className="text-sm text-slate-400">
                        Payment: {formatDate(period.payment_date)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={`px-3 py-1.5 text-xs font-medium rounded-lg ${getFrequencyBadge(period.frequency)}`}>
                      {getFrequencyLabel(period.frequency)}
                    </span>
                    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                      <CheckCircle2 size={12} />
                      Posted
                    </span>
                    <ChevronRight size={20} className="text-slate-500 group-hover:text-emerald-400 transition-colors" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Summary Footer */}
        {!loading && periods.length > 0 && (
          <div className="flex items-center justify-between text-sm text-slate-400 px-2">
            <span>Total: {periods.length} periods</span>
            <div className="flex items-center gap-4">
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-blue-400 rounded-full"></span>
                {openPeriods.length} open
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-amber-400 rounded-full"></span>
                {closedPeriods.length} to approve
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-purple-400 rounded-full"></span>
                {pendingPaymentPeriods.length} to pay
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-emerald-400 rounded-full"></span>
                {postedPeriods.length} paid
              </span>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
