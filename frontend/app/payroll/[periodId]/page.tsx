/**
 * @file app/payroll/[periodId]/page.tsx
 * @description Payroll period detail and calculation page
 *
 * USER PERSPECTIVE:
 *   - View summary of payroll for a specific period
 *   - See totals: gross income, deductions, net pay, employer contributions
 *   - Calculate payroll for all employees in the period
 *   - Approve/post the payroll when calculations are complete
 *   - View individual employee payroll details with link to payslip
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Summary cards, table columns, calculation button behavior
 *   CAUTION: Approval is irreversible, ensure proper confirmation dialogs
 *   DO NOT modify: Calculation logic without updating backend payroll engine
 *
 * KEY COMPONENTS:
 *   - Summary cards: Total gross, deductions, net, employer contributions
 *   - Tax breakdown: ISR, IMSS, other deductions
 *   - Employee table: Per-employee calculations with payslip links
 *   - Action buttons: Calculate, approve, refresh
 *
 * API ENDPOINTS USED:
 *   - GET /api/payroll/periods/:id (via payrollApi.getPeriod)
 *   - GET /api/payroll/periods/:id/summary (via payrollApi.getPayrollSummary)
 *   - GET /api/payroll/periods/:id/calculations (via payrollApi.getPayrollByPeriod)
 *   - POST /api/payroll/periods/:id/calculate (via payrollApi.bulkCalculatePayroll)
 *   - POST /api/payroll/periods/:id/approve (via payrollApi.approvePayroll)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import {
  ArrowLeft,
  Calculator,
  Check,
  DollarSign,
  Users,
  TrendingUp,
  TrendingDown,
  Eye,
  RefreshCw,
  Calendar,
  Clock,
  Play,
  Building2,
  AlertCircle,
  FileText,
  Wallet,
  Download
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import {
  payrollApi,
  employeeApi,
  PayrollPeriod,
  PayrollSummary,
  PayrollCalculationResponse,
  ApiError,
  Employee,
} from "@/lib/api-client"
import { canProcessPayroll } from "@/lib/auth"

export default function PayrollSummaryPage() {
  const router = useRouter()
  const params = useParams()
  const periodId = params.periodId as string

  const [period, setPeriod] = useState<PayrollPeriod | null>(null)
  const [summary, setSummary] = useState<PayrollSummary | null>(null)
  const [calculations, setCalculations] = useState<PayrollCalculationResponse[]>([])
  const [eligibleEmployees, setEligibleEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [calculating, setCalculating] = useState(false)
  const [approving, setApproving] = useState(false)
  const [processing, setProcessing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Check authorization - only admin and payroll roles can access
    if (!canProcessPayroll()) {
      router.push("/dashboard")
      return
    }
    if (periodId) {
      loadData()
    }
  }, [periodId, router])

  async function loadData() {
    try {
      setLoading(true)
      setError(null)

      const [periodData, summaryData, calcData, employeesResponse] = await Promise.all([
        payrollApi.getPeriod(periodId).catch(() => null),
        payrollApi.getPayrollSummary(periodId).catch(() => null),
        payrollApi.getPayrollByPeriod(periodId).catch(() => []),
        employeeApi.getEmployees().catch(() => ({ employees: [] })),
      ])

      setPeriod(periodData)
      setSummary(summaryData)
      setCalculations(Array.isArray(calcData) ? calcData : [])

      // Filter eligible employees based on period frequency
      const allEmployees = employeesResponse.employees || []
      const activeEmployees = allEmployees.filter(e => e.employment_status === "active")

      if (periodData) {
        let filtered: Employee[] = []
        switch (periodData.frequency) {
          case "weekly":
            filtered = activeEmployees.filter(e => e.collar_type === "blue_collar" || e.collar_type === "gray_collar")
            break
          case "biweekly":
            filtered = activeEmployees.filter(e => e.collar_type === "white_collar")
            break
          case "monthly":
            filtered = activeEmployees // All employees for monthly
            break
          default:
            filtered = activeEmployees
        }
        setEligibleEmployees(filtered)
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load payroll data")
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleCalculate() {
    try {
      setCalculating(true)
      setError(null)
      await payrollApi.bulkCalculatePayroll(periodId, undefined, true)
      await loadData()
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to calculate payroll")
      }
    } finally {
      setCalculating(false)
    }
  }

  async function handleApprove() {
    if (!confirm("Are you sure you want to approve this payroll? This action cannot be undone.")) {
      return
    }
    try {
      setApproving(true)
      setError(null)
      await payrollApi.approvePayroll(periodId)
      await loadData()
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to approve payroll")
      }
    } finally {
      setApproving(false)
    }
  }

  async function handleProcessPayment() {
    if (!confirm("Are you sure you want to process the payment? This action will mark the payroll as paid.")) {
      return
    }
    try {
      setProcessing(true)
      setError(null)
      await payrollApi.processPayment(periodId)
      await loadData()
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to process payment")
      }
    } finally {
      setProcessing(false)
    }
  }

  async function handleDownloadPayslip(employeeId: string, employeeNumber?: string) {
    try {
      const baseFilename = `recibo_${employeeNumber || employeeId}_${periodId}`

      // Fetch both PDF and XML in parallel (auth via httpOnly cookies)
      const [pdfResponse, xmlResponse] = await Promise.all([
        fetch(`/api/v1/payroll/payslip/${periodId}/${employeeId}?format=pdf`, {
          method: "GET",
          credentials: "include",
        }),
        fetch(`/api/v1/payroll/payslip/${periodId}/${employeeId}?format=xml`, {
          method: "GET",
          credentials: "include",
        }),
      ])

      if (!pdfResponse.ok) {
        const errorData = await pdfResponse.json().catch(() => ({ message: "Failed to download PDF" }))
        throw new Error(errorData.message || "Failed to download PDF")
      }

      // Download PDF
      const pdfBlob = await pdfResponse.blob()
      const pdfUrl = window.URL.createObjectURL(pdfBlob)
      const pdfLink = document.createElement("a")
      pdfLink.href = pdfUrl

      const pdfContentDisposition = pdfResponse.headers.get("Content-Disposition")
      let pdfFilename = `${baseFilename}.pdf`
      if (pdfContentDisposition) {
        const filenameMatch = pdfContentDisposition.match(/filename=(.+)/)
        if (filenameMatch && filenameMatch[1]) {
          pdfFilename = filenameMatch[1].replace(/"/g, "")
        }
      }

      pdfLink.download = pdfFilename
      document.body.appendChild(pdfLink)
      pdfLink.click()
      document.body.removeChild(pdfLink)
      window.URL.revokeObjectURL(pdfUrl)

      // Download XML (CFDI) if available
      if (xmlResponse.ok) {
        const xmlBlob = await xmlResponse.blob()
        const xmlUrl = window.URL.createObjectURL(xmlBlob)
        const xmlLink = document.createElement("a")
        xmlLink.href = xmlUrl

        const xmlContentDisposition = xmlResponse.headers.get("Content-Disposition")
        let xmlFilename = `${baseFilename}_cfdi.xml`
        if (xmlContentDisposition) {
          const filenameMatch = xmlContentDisposition.match(/filename=(.+)/)
          if (filenameMatch && filenameMatch[1]) {
            xmlFilename = filenameMatch[1].replace(/"/g, "")
          }
        }

        xmlLink.download = xmlFilename
        document.body.appendChild(xmlLink)

        // Small delay to ensure both downloads start properly
        setTimeout(() => {
          xmlLink.click()
          document.body.removeChild(xmlLink)
          window.URL.revokeObjectURL(xmlUrl)
        }, 100)
      }
    } catch (err) {
      console.error("Download error:", err)
      setError(err instanceof Error ? err.message : "Failed to download files")
    }
  }

  const formatCurrency = (amount: number | undefined) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
    }).format(amount || 0)
  }

  const formatDate = (dateStr: string | undefined) => {
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
      case "closed":
        return "bg-gradient-to-r from-amber-500/20 to-orange-500/20 text-amber-300 border border-amber-500/30"
      case "approved":
        return "bg-gradient-to-r from-purple-500/20 to-pink-500/20 text-purple-300 border border-purple-500/30"
      case "paid":
        return "bg-gradient-to-r from-emerald-500/20 to-green-500/20 text-emerald-300 border border-emerald-500/30"
      default:
        return "bg-slate-700/50 text-slate-400"
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "open": return <Play size={14} />
      case "calculated":
      case "closed": return <Clock size={14} />
      case "approved": return <Check size={14} />
      case "paid": return <Check size={14} />
      default: return null
    }
  }

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "open": return "Open"
      case "calculated": return "Calculated"
      case "approved": return "Approved"
      case "closed": return "Closed"
      case "paid": return "Paid"
      default: return status
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

  // Calculate totals from calculations if summary isn't available
  const totals = {
    totalGross: summary?.total_gross || calculations.reduce((sum, c) => sum + (c.total_gross_income || 0), 0),
    totalDeductions: summary?.total_deductions || calculations.reduce((sum, c) => sum + (c.total_deductions || 0), 0),
    totalNet: summary?.total_net || calculations.reduce((sum, c) => sum + (c.total_net_pay || 0), 0),
    employerContributions:
      summary?.employer_contributions ||
      calculations.reduce((sum, c) => sum + (c.employer_contributions?.total_contributions || 0), 0),
    totalISR: calculations.reduce((sum, c) => sum + (c.isr_withholding || 0), 0),
    totalIMSS: calculations.reduce((sum, c) => sum + (c.imss_employee || 0), 0),
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="flex flex-col items-center gap-3 text-slate-400">
            <div className="w-8 h-8 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
            Loading payroll data...
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <button
              onClick={() => router.push("/payroll")}
              className="p-2.5 text-slate-400 hover:text-white hover:bg-slate-800/50 rounded-xl border border-slate-700/50 transition-all"
            >
              <ArrowLeft size={20} />
            </button>
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
                Payroll Summary
              </h1>
              <p className="text-slate-400 mt-1">
                {period ? `${period.year} - Period ${period.period_number}` : `Period ${periodId}`}
              </p>
            </div>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={loadData}
              variant="outline"
              size="sm"
              className="border-slate-600 text-slate-300 hover:bg-slate-700"
            >
              <RefreshCw size={16} className="mr-2" />
              Refresh
            </Button>
            {/* Calculate Button - available when open or calculated */}
            {(period?.status === "open" || period?.status === "calculated") && (
              <Button
                onClick={handleCalculate}
                disabled={calculating}
                variant="outline"
                className="border-blue-600/50 text-blue-300 hover:bg-blue-600/20 hover:border-blue-500"
              >
                {calculating ? (
                  <RefreshCw size={18} className="mr-2 animate-spin" />
                ) : (
                  <Calculator size={18} className="mr-2" />
                )}
                {calculating ? "Calculating..." : "Calculate Payroll"}
              </Button>
            )}
            {/* Approve Button - available when open or calculated, and has calculations */}
            {(period?.status === "open" || period?.status === "calculated") && (
              <Button
                onClick={handleApprove}
                disabled={approving || calculations.length === 0}
                className="bg-gradient-to-r from-emerald-600 to-green-600 hover:from-emerald-700 hover:to-green-700 shadow-lg shadow-emerald-500/25"
              >
                {approving ? (
                  <RefreshCw size={18} className="mr-2 animate-spin" />
                ) : (
                  <Check size={18} className="mr-2" />
                )}
                {approving ? "Approving..." : "Approve Payroll"}
              </Button>
            )}
            {/* Process Payment Button - available when approved */}
            {period?.status === "approved" && (
              <Button
                onClick={handleProcessPayment}
                disabled={processing}
                className="bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 shadow-lg shadow-purple-500/25"
              >
                {processing ? (
                  <RefreshCw size={18} className="mr-2 animate-spin" />
                ) : (
                  <DollarSign size={18} className="mr-2" />
                )}
                {processing ? "Processing..." : "Process Payment"}
              </Button>
            )}
            {/* Paid status indicator */}
            {period?.status === "paid" && (
              <div className="flex items-center gap-2 px-4 py-2 bg-emerald-500/20 border border-emerald-500/30 rounded-lg">
                <Check size={18} className="text-emerald-400" />
                <span className="text-emerald-300 font-medium">Payroll Paid</span>
              </div>
            )}
          </div>
        </div>

        {/* Period Info */}
        {period && (
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-5 border border-slate-700/50">
            <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-purple-500/20 rounded-lg">
                  <Calendar size={16} className="text-purple-400" />
                </div>
                <div>
                  <p className="text-xs text-slate-400">Frequency</p>
                  <p className="text-sm font-medium text-white">{getFrequencyLabel(period.frequency)}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-500/20 rounded-lg">
                  <Calendar size={16} className="text-blue-400" />
                </div>
                <div>
                  <p className="text-xs text-slate-400">Start</p>
                  <p className="text-sm font-medium text-white">{formatDate(period.start_date)}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-cyan-500/20 rounded-lg">
                  <Calendar size={16} className="text-cyan-400" />
                </div>
                <div>
                  <p className="text-xs text-slate-400">End</p>
                  <p className="text-sm font-medium text-white">{formatDate(period.end_date)}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-amber-500/20 rounded-lg">
                  <Wallet size={16} className="text-amber-400" />
                </div>
                <div>
                  <p className="text-xs text-slate-400">Payment</p>
                  <p className="text-sm font-medium text-white">{formatDate(period.payment_date)}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-indigo-500/20 rounded-lg">
                  <Users size={16} className="text-indigo-400" />
                </div>
                <div>
                  <p className="text-xs text-slate-400">Employees</p>
                  <p className="text-sm font-medium text-white">
                    {calculations.length > 0 ? calculations.length : eligibleEmployees.length}
                    {calculations.length === 0 && eligibleEmployees.length > 0 && (
                      <span className="text-xs text-slate-400 ml-1">(pending)</span>
                    )}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <span className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg ${getStatusBadge(period.status)}`}>
                  {getStatusIcon(period.status)}
                  {getStatusLabel(period.status)}
                </span>
              </div>
            </div>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="bg-red-900/20 border border-red-700/50 rounded-xl p-4 text-red-400 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <AlertCircle size={20} />
              <span>{error}</span>
            </div>
            <button onClick={loadData} className="px-3 py-1 bg-red-500/20 hover:bg-red-500/30 rounded-lg transition-colors">
              Retry
            </button>
          </div>
        )}

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-gradient-to-br from-emerald-600/20 to-green-600/20 backdrop-blur-sm rounded-xl p-6 border border-emerald-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-emerald-200 text-sm font-medium">Total Gross</p>
                <p className="text-2xl font-bold text-white mt-2">{formatCurrency(totals.totalGross)}</p>
                <p className="text-emerald-300/70 text-xs mt-1">Total earnings</p>
              </div>
              <div className="p-4 bg-emerald-500/20 rounded-xl">
                <TrendingUp size={28} className="text-emerald-400" />
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-red-600/20 to-rose-600/20 backdrop-blur-sm rounded-xl p-6 border border-red-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-red-200 text-sm font-medium">Total Deductions</p>
                <p className="text-2xl font-bold text-white mt-2">{formatCurrency(totals.totalDeductions)}</p>
                <p className="text-red-300/70 text-xs mt-1">ISR + IMSS + Other</p>
              </div>
              <div className="p-4 bg-red-500/20 rounded-xl">
                <TrendingDown size={28} className="text-red-400" />
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-blue-600/20 to-cyan-600/20 backdrop-blur-sm rounded-xl p-6 border border-blue-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-blue-200 text-sm font-medium">Total Net Pay</p>
                <p className="text-2xl font-bold text-white mt-2">{formatCurrency(totals.totalNet)}</p>
                <p className="text-blue-300/70 text-xs mt-1">Amount to deposit</p>
              </div>
              <div className="p-4 bg-blue-500/20 rounded-xl">
                <DollarSign size={28} className="text-blue-400" />
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-purple-600/20 to-pink-600/20 backdrop-blur-sm rounded-xl p-6 border border-purple-500/30">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-purple-200 text-sm font-medium">Employer Contributions</p>
                <p className="text-2xl font-bold text-white mt-2">{formatCurrency(totals.employerContributions)}</p>
                <p className="text-purple-300/70 text-xs mt-1">IMSS, INFONAVIT, etc.</p>
              </div>
              <div className="p-4 bg-purple-500/20 rounded-xl">
                <Building2 size={28} className="text-purple-400" />
              </div>
            </div>
          </div>
        </div>

        {/* Tax Summary */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-400">ISR Withheld</p>
                <p className="text-xl font-bold text-red-400 mt-1">{formatCurrency(totals.totalISR)}</p>
              </div>
              <div className="p-2 bg-red-500/20 rounded-lg">
                <FileText size={20} className="text-red-400" />
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-400">IMSS Employee</p>
                <p className="text-xl font-bold text-amber-400 mt-1">{formatCurrency(totals.totalIMSS)}</p>
              </div>
              <div className="p-2 bg-amber-500/20 rounded-lg">
                <Building2 size={20} className="text-amber-400" />
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-400">Other Deductions</p>
                <p className="text-xl font-bold text-slate-300 mt-1">
                  {formatCurrency(totals.totalDeductions - totals.totalISR - totals.totalIMSS)}
                </p>
              </div>
              <div className="p-2 bg-slate-500/20 rounded-lg">
                <TrendingDown size={20} className="text-slate-400" />
              </div>
            </div>
          </div>
        </div>

        {/* Employee Payroll Table */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-500/20 rounded-lg">
                <Users size={18} className="text-indigo-400" />
              </div>
              <h2 className="text-lg font-semibold text-white">Employee Details</h2>
            </div>
            <span className="px-2.5 py-1 text-xs font-medium bg-indigo-500/20 text-indigo-300 rounded-full">
              {calculations.length} employees
            </span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/80 border-b border-slate-700/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Employee
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Regular Salary
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Gross Income
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    ISR
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    IMSS
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Total Deductions
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Net Pay
                  </th>
                  <th className="px-6 py-4 text-center text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Payslip
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {calculations.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="px-6 py-12 text-center text-slate-400">
                      <div className="flex flex-col items-center gap-3">
                        <Calculator className="w-12 h-12 text-slate-600" />
                        <p>No payroll calculations for this period</p>
                        <p className="text-sm text-slate-500">Click "Calculate Payroll" to generate the calculations</p>
                        <Button
                          onClick={handleCalculate}
                          disabled={calculating}
                          className="mt-2 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700"
                        >
                          <Calculator size={16} className="mr-2" />
                          Calculate Payroll
                        </Button>
                      </div>
                    </td>
                  </tr>
                ) : (
                  calculations.map((calc) => (
                    <tr key={calc.id} className="hover:bg-slate-700/30 transition-colors group">
                      <td className="px-6 py-4">
                        <div>
                          <div className="text-sm text-white font-medium">{calc.employee_name}</div>
                          <div className="text-xs text-slate-500 font-mono">{calc.employee_number}</div>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-300 text-right font-mono">
                        {formatCurrency(calc.regular_salary)}
                      </td>
                      <td className="px-6 py-4 text-sm text-emerald-400 text-right font-semibold">
                        {formatCurrency(calc.total_gross_income)}
                      </td>
                      <td className="px-6 py-4 text-sm text-red-400 text-right font-mono">
                        {formatCurrency(calc.isr_withholding)}
                      </td>
                      <td className="px-6 py-4 text-sm text-amber-400 text-right font-mono">
                        {formatCurrency(calc.imss_employee)}
                      </td>
                      <td className="px-6 py-4 text-sm text-red-400 text-right font-semibold">
                        {formatCurrency(calc.total_deductions)}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className="text-sm font-bold text-white bg-blue-500/20 px-3 py-1.5 rounded-lg">
                          {formatCurrency(calc.total_net_pay)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <div className="flex items-center justify-center gap-1">
                          <button
                            onClick={() => router.push(`/payroll/${periodId}/payslip/${calc.employee_id}`)}
                            className="p-2 text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition-all"
                            title="View Payslip"
                          >
                            <Eye size={18} />
                          </button>
                          <button
                            onClick={() => handleDownloadPayslip(calc.employee_id, calc.employee_number)}
                            className="p-2 text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10 rounded-lg transition-all"
                            title="Download PDF and XML"
                          >
                            <Download size={18} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
              {calculations.length > 0 && (
                <tfoot className="bg-slate-900/80 border-t-2 border-slate-600">
                  <tr>
                    <td className="px-6 py-4 text-sm text-white font-bold">
                      TOTALS ({calculations.length} employees)
                    </td>
                    <td className="px-6 py-4 text-sm text-slate-300 text-right font-mono font-medium">
                      {formatCurrency(calculations.reduce((sum, c) => sum + (c.regular_salary || 0), 0))}
                    </td>
                    <td className="px-6 py-4 text-sm text-emerald-400 text-right font-bold">
                      {formatCurrency(totals.totalGross)}
                    </td>
                    <td className="px-6 py-4 text-sm text-red-400 text-right font-mono font-medium">
                      {formatCurrency(totals.totalISR)}
                    </td>
                    <td className="px-6 py-4 text-sm text-amber-400 text-right font-mono font-medium">
                      {formatCurrency(totals.totalIMSS)}
                    </td>
                    <td className="px-6 py-4 text-sm text-red-400 text-right font-bold">
                      {formatCurrency(totals.totalDeductions)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <span className="text-sm font-bold text-white bg-emerald-500/30 px-3 py-1.5 rounded-lg">
                        {formatCurrency(totals.totalNet)}
                      </span>
                    </td>
                    <td></td>
                  </tr>
                </tfoot>
              )}
            </table>
          </div>
        </div>

        {/* Footer Info */}
        {calculations.length > 0 && (
          <div className="flex items-center justify-between text-sm text-slate-400 px-2">
            <span>Payroll generated on {new Date().toLocaleDateString("en-US")}</span>
            <div className="flex items-center gap-4">
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-emerald-400 rounded-full"></span>
                Gross: {formatCurrency(totals.totalGross)}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-red-400 rounded-full"></span>
                Deductions: {formatCurrency(totals.totalDeductions)}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-blue-400 rounded-full"></span>
                Net: {formatCurrency(totals.totalNet)}
              </span>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
