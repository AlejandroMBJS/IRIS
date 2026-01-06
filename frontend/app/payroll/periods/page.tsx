/**
 * @file app/payroll/periods/page.tsx
 * @description Payroll period management page with filtering and status tracking
 *
 * USER PERSPECTIVE:
 *   - View all payroll periods in a sortable table
 *   - Filter by year, frequency (weekly/biweekly/monthly), and status
 *   - See period dates, payment dates, and current status
 *   - Click on period to view/process details
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Filter options, table columns, status badges
 *   CAUTION: Frequency values (weekly, biweekly, monthly) are tied to backend enum
 *   DO NOT modify: Period status logic without updating calculation engine
 *
 * KEY COMPONENTS:
 *   - Filter bar: Year, frequency, and status dropdowns
 *   - Stats cards: Total, open, closed, posted period counts
 *   - Period table: Comprehensive list with status indicators
 *
 * API ENDPOINTS USED:
 *   - GET /api/payroll/periods (via payrollApi.getPeriods)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Calendar, Play, Check, Lock, RefreshCw,
  Search, Filter, ChevronRight, Clock, AlertCircle
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { payrollApi, PayrollPeriod, ApiError } from "@/lib/api-client"
import { canProcessPayroll } from "@/lib/auth"

export default function PayrollPeriodsPage() {
  const router = useRouter()
  const [periods, setPeriods] = useState<PayrollPeriod[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState("")
  const [yearFilter, setYearFilter] = useState("")
  const [frequencyFilter, setFrequencyFilter] = useState("")

  useEffect(() => {
    // Check authorization - only admin and payroll roles can access
    if (!canProcessPayroll()) {
      router.push("/dashboard")
      return
    }
    loadPeriods()
  }, [router])

  async function loadPeriods() {
    try {
      setLoading(true)
      setError(null)
      const data = await payrollApi.getPeriods()
      setPeriods(Array.isArray(data) ? data : [])
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load payroll periods")
      }
      setPeriods([])
    } finally {
      setLoading(false)
    }
  }

  const filteredPeriods = periods.filter((period) => {
    const matchesStatus = statusFilter === "" || period.status === statusFilter
    const matchesYear = yearFilter === "" || period.year.toString() === yearFilter
    const matchesFrequency = frequencyFilter === "" || period.frequency?.toLowerCase() === frequencyFilter
    return matchesStatus && matchesYear && matchesFrequency
  })

  const years = [...new Set(periods.map((p) => p.year))].sort((a, b) => b - a)

  // Stats
  const stats = {
    total: periods.length,
    open: periods.filter(p => p.status === "open").length,
    closed: periods.filter(p => ["closed", "calculated", "approved"].includes(p.status)).length,
    posted: periods.filter(p => p.status === "paid").length,
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

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "open":
        return "bg-gradient-to-r from-blue-500/20 to-cyan-500/20 text-blue-300 border border-blue-500/30"
      case "calculated":
      case "approved":
      case "closed":
        return "bg-gradient-to-r from-amber-500/20 to-orange-500/20 text-amber-300 border border-amber-500/30"
      case "paid":
        return "bg-gradient-to-r from-emerald-500/20 to-green-500/20 text-emerald-300 border border-emerald-500/30"
      default:
        return "bg-slate-700/50 text-slate-400 border border-slate-600"
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "open":
        return <Play size={12} />
      case "calculated":
      case "approved":
      case "closed":
        return <Clock size={12} />
      case "paid":
        return <Check size={12} />
      default:
        return null
    }
  }

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "open": return "Open"
      case "calculated": return "Calculated"
      case "approved": return "Approved"
      case "closed": return "Closed"
      case "paid": return "Paid"
      case "cancelled": return "Cancelled"
      default: return status
    }
  }

  const getFrequencyBadge = (frequency: string) => {
    switch (frequency?.toLowerCase()) {
      case "weekly":
        return "bg-blue-900/30 text-blue-300"
      case "biweekly":
        return "bg-purple-900/30 text-purple-300"
      case "monthly":
        return "bg-amber-900/30 text-amber-300"
      default:
        return "bg-slate-700/30 text-slate-400"
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

  const handlePeriodClick = (periodId: string) => {
    router.push(`/payroll/${periodId}`)
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              Payroll Periods
            </h1>
            <p className="text-slate-400 mt-1">Manage payment cycles and periods</p>
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
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-slate-500/20 rounded-lg">
                <Calendar className="w-5 h-5 text-slate-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.total}</p>
                <p className="text-xs text-slate-400">Total Periods</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/20 rounded-lg">
                <Play className="w-5 h-5 text-blue-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.open}</p>
                <p className="text-xs text-slate-400">Open</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-500/20 rounded-lg">
                <Clock className="w-5 h-5 text-amber-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.closed}</p>
                <p className="text-xs text-slate-400">Closed</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-500/20 rounded-lg">
                <Check className="w-5 h-5 text-emerald-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.posted}</p>
                <p className="text-xs text-slate-400">Posted</p>
              </div>
            </div>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
          <div className="flex flex-col md:flex-row gap-4">
            <select
              value={yearFilter}
              onChange={(e) => setYearFilter(e.target.value)}
              className="bg-slate-900/50 border border-slate-600 rounded-lg px-4 py-2.5 text-slate-200 focus:outline-none focus:border-blue-500 transition-all"
            >
              <option value="">All Years</option>
              {years.map((year) => (
                <option key={year} value={year}>
                  {year}
                </option>
              ))}
            </select>
            <select
              value={frequencyFilter}
              onChange={(e) => setFrequencyFilter(e.target.value)}
              className="bg-slate-900/50 border border-slate-600 rounded-lg px-4 py-2.5 text-slate-200 focus:outline-none focus:border-blue-500 transition-all"
            >
              <option value="">All Frequencies</option>
              <option value="weekly">Weekly</option>
              <option value="biweekly">Biweekly</option>
              <option value="monthly">Monthly</option>
            </select>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="bg-slate-900/50 border border-slate-600 rounded-lg px-4 py-2.5 text-slate-200 focus:outline-none focus:border-blue-500 transition-all"
            >
              <option value="">All Statuses</option>
              <option value="open">Open</option>
              <option value="calculated">Calculated</option>
              <option value="approved">Approved</option>
              <option value="paid">Paid</option>
              <option value="closed">Closed</option>
            </select>
          </div>
          {(yearFilter || statusFilter || frequencyFilter) && (
            <div className="mt-3 flex items-center gap-2">
              <span className="text-xs text-slate-400">Active filters:</span>
              {yearFilter && (
                <span className="px-2 py-1 text-xs bg-slate-700/50 text-slate-300 rounded-full">
                  Year: {yearFilter}
                </span>
              )}
              {frequencyFilter && (
                <span className="px-2 py-1 text-xs bg-purple-500/20 text-purple-300 rounded-full">
                  {getFrequencyLabel(frequencyFilter)}
                </span>
              )}
              {statusFilter && (
                <span className="px-2 py-1 text-xs bg-blue-500/20 text-blue-300 rounded-full">
                  {getStatusLabel(statusFilter)}
                </span>
              )}
              <button
                onClick={() => {
                  setStatusFilter("")
                  setYearFilter("")
                  setFrequencyFilter("")
                }}
                className="px-2 py-1 text-xs text-red-400 hover:text-red-300 transition-colors"
              >
                Clear filters
              </button>
            </div>
          )}
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

        {/* Table */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/80 border-b border-slate-700/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Year
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Period #
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Frequency
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Start Date
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    End Date
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Payment Date
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-4 text-center text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {loading ? (
                  <tr>
                    <td colSpan={8} className="px-6 py-12 text-center text-slate-400">
                      <div className="flex flex-col items-center gap-3">
                        <div className="w-8 h-8 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                        <span>Loading periods...</span>
                      </div>
                    </td>
                  </tr>
                ) : filteredPeriods.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="px-6 py-12 text-center text-slate-400">
                      {statusFilter || yearFilter || frequencyFilter ? (
                        <div className="flex flex-col items-center gap-3">
                          <Filter className="w-12 h-12 text-slate-600" />
                          <p>No periods match the filters</p>
                          <button
                            onClick={() => {
                              setStatusFilter("")
                              setYearFilter("")
                              setFrequencyFilter("")
                            }}
                            className="text-blue-400 hover:text-blue-300 transition-colors"
                          >
                            Clear filters
                          </button>
                        </div>
                      ) : (
                        <div className="flex flex-col items-center gap-3">
                          <Calendar className="w-12 h-12 text-slate-600" />
                          <p>No payroll periods</p>
                          <p className="text-sm text-slate-500">Periods are automatically generated by the system</p>
                        </div>
                      )}
                    </td>
                  </tr>
                ) : (
                  filteredPeriods.map((period) => (
                    <tr
                      key={period.id}
                      onClick={() => handlePeriodClick(period.id)}
                      className="hover:bg-slate-700/30 cursor-pointer transition-colors group"
                    >
                      <td className="px-6 py-4">
                        <span className="text-lg font-bold text-white">{period.year}</span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-sm font-mono text-slate-300 bg-slate-800/50 px-2 py-1 rounded">
                          Period {period.period_number}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-2.5 py-1 text-xs font-medium rounded-lg ${getFrequencyBadge(period.frequency)}`}>
                          {getFrequencyLabel(period.frequency)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-300">{formatDate(period.start_date)}</td>
                      <td className="px-6 py-4 text-sm text-slate-300">{formatDate(period.end_date)}</td>
                      <td className="px-6 py-4 text-sm text-slate-300">{formatDate(period.payment_date)}</td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-semibold rounded-lg ${getStatusBadge(period.status)}`}>
                          {getStatusIcon(period.status)}
                          {getStatusLabel(period.status)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <ChevronRight size={20} className="inline-block text-slate-500 group-hover:text-blue-400 transition-colors" />
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Summary */}
        {!loading && filteredPeriods.length > 0 && (
          <div className="flex items-center justify-between text-sm text-slate-400 px-2">
            <span>Showing {filteredPeriods.length} of {periods.length} periods</span>
            <div className="flex items-center gap-4">
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-blue-400 rounded-full"></span>
                {stats.open} open
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-amber-400 rounded-full"></span>
                {stats.closed} closed
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-emerald-400 rounded-full"></span>
                {stats.posted} posted
              </span>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
