/**
 * @file app/payroll/export/page.tsx
 * @description Payroll export page for downloading dual Excel templates (Vacation + Absences and Extras)
 *
 * USER PERSPECTIVE:
 *   - Select a payroll period from dropdown
 *   - See preview of export counts (vacation, absences_extras, late approvals)
 *   - Download ZIP file containing both Excel templates
 *   - View warnings for late approvals and data quality issues
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: UI layout, preview information, warning messages
 *   CAUTION: Download logic, file naming convention
 *   DO NOT modify: API endpoints, Excel format structure
 *
 * KEY FEATURES:
 *   - Period selection with validation
 *   - Real-time preview before download
 *   - Warning badges for data quality issues
 *   - One-click ZIP download
 *
 * API ENDPOINTS USED:
 *   - GET /api/payroll-export/preview/:periodID (via payrollExportApi.getPreview)
 *   - GET /api/payroll-export/dual/:periodID (via payrollExportApi.downloadDualExport)
 *   - GET /api/payroll/periods (via payrollApi.getPeriods)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Download, FileSpreadsheet, AlertCircle, CheckCircle, Calendar,
  Clock, FileText, Package
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import {
  payrollApi, payrollExportApi, PayrollPeriod, ExportPreview, ApiError
} from "@/lib/api-client"
import { canProcessPayroll } from "@/lib/auth"

export default function PayrollExportPage() {
  const router = useRouter()
  const [periods, setPeriods] = useState<PayrollPeriod[]>([])
  const [selectedPeriodId, setSelectedPeriodId] = useState<string>("")
  const [preview, setPreview] = useState<ExportPreview | null>(null)
  const [loading, setLoading] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [error, setError] = useState<string | null>(null)

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

  async function loadPreview(periodId: string) {
    if (!periodId) {
      setPreview(null)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data = await payrollExportApi.getPreview(periodId)
      setPreview(data)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load export preview")
      }
      setPreview(null)
    } finally {
      setLoading(false)
    }
  }

  async function handleDownload() {
    if (!selectedPeriodId) {
      setError("Please select a payroll period")
      return
    }

    try {
      setDownloading(true)
      setError(null)

      const blob = await payrollExportApi.downloadDualExport(selectedPeriodId)

      // Create download link
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `payroll_export_${selectedPeriodId}.zip`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)

      // Show success message
      alert("Export downloaded successfully")
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to download export")
      }
    } finally {
      setDownloading(false)
    }
  }

  const handlePeriodChange = (periodId: string) => {
    setSelectedPeriodId(periodId)
    loadPreview(periodId)
  }

  const selectedPeriod = periods.find(p => p.id === selectedPeriodId)

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-blue-50 rounded-lg">
              <Package className="w-6 h-6 text-blue-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Dual Excel Export</h1>
              <p className="text-sm text-gray-500">Download Vacation.xlsx + Absences_and_Extras.xlsx</p>
            </div>
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-red-900">Error</h3>
              <p className="text-sm text-red-700">{error}</p>
            </div>
          </div>
        )}

        {/* Period Selector */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center gap-2 mb-4">
            <Calendar className="w-5 h-5 text-gray-500" />
            <h2 className="text-lg font-semibold text-gray-900">Select Period</h2>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Payroll Period
              </label>
              <select
                value={selectedPeriodId}
                onChange={(e) => handlePeriodChange(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={loading}
              >
                <option value="">-- Select a period --</option>
                {periods.map((period) => (
                  <option key={period.id} value={period.id}>
                    {period.period_code} ({period.start_date} - {period.end_date})
                    {period.status !== 'open' && ` [${period.status.toUpperCase()}]`}
                  </option>
                ))}
              </select>
            </div>

            {selectedPeriod && (
              <div className="grid grid-cols-3 gap-4 p-4 bg-gray-50 rounded-lg">
                <div>
                  <p className="text-xs text-gray-500">Start Date</p>
                  <p className="text-sm font-medium text-gray-900">{selectedPeriod.start_date}</p>
                </div>
                <div>
                  <p className="text-xs text-gray-500">End Date</p>
                  <p className="text-sm font-medium text-gray-900">{selectedPeriod.end_date}</p>
                </div>
                <div>
                  <p className="text-xs text-gray-500">Payment Date</p>
                  <p className="text-sm font-medium text-gray-900">{selectedPeriod.payment_date}</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Preview Panel */}
        {preview && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div className="flex items-center gap-2 mb-4">
              <FileText className="w-5 h-5 text-gray-500" />
              <h2 className="text-lg font-semibold text-gray-900">Export Preview</h2>
            </div>

            <div className="grid grid-cols-4 gap-4 mb-6">
              <div className="bg-blue-50 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <FileSpreadsheet className="w-5 h-5 text-blue-600" />
                  <span className="text-2xl font-bold text-blue-900">{preview.vacaciones_count}</span>
                </div>
                <p className="text-sm text-blue-700 font-medium">Vacation</p>
                <p className="text-xs text-blue-600">Records in Vacation.xlsx</p>
              </div>

              <div className="bg-green-50 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <FileSpreadsheet className="w-5 h-5 text-green-600" />
                  <span className="text-2xl font-bold text-green-900">{preview.faltas_extras_count}</span>
                </div>
                <p className="text-sm text-green-700 font-medium">Absences and Extras</p>
                <p className="text-xs text-green-600">Records in Absences_and_Extras.xlsx</p>
              </div>

              <div className="bg-purple-50 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <CheckCircle className="w-5 h-5 text-purple-600" />
                  <span className="text-2xl font-bold text-purple-900">{preview.total_incidences}</span>
                </div>
                <p className="text-sm text-purple-700 font-medium">Total Incidences</p>
                <p className="text-xs text-purple-600">Approved for export</p>
              </div>

              <div className={`rounded-lg p-4 ${preview.late_approval_count > 0 ? 'bg-yellow-50' : 'bg-gray-50'}`}>
                <div className="flex items-center justify-between mb-2">
                  <Clock className="w-5 h-5 text-yellow-600" />
                  <span className={`text-2xl font-bold ${preview.late_approval_count > 0 ? 'text-yellow-900' : 'text-gray-400'}`}>
                    {preview.late_approval_count}
                  </span>
                </div>
                <p className={`text-sm font-medium ${preview.late_approval_count > 0 ? 'text-yellow-700' : 'text-gray-500'}`}>
                  Late Approvals
                </p>
                <p className={`text-xs ${preview.late_approval_count > 0 ? 'text-yellow-600' : 'text-gray-400'}`}>
                  Marked with warning
                </p>
              </div>
            </div>

            {/* Warnings */}
            {preview.late_approval_count > 0 && (
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-yellow-900 mb-1">Warning: Late Approvals</h3>
                    <p className="text-sm text-yellow-700">
                      There are {preview.late_approval_count} incidence(s) approved after the payroll cutoff.
                      These will appear marked with "⚠️ LATE APPROVAL" in the Observations column of the Excel file.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {preview.total_incidences === 0 && (
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-blue-900 mb-1">Information</h3>
                    <p className="text-sm text-blue-700">
                      There are no approved incidences in this period. The ZIP file will contain empty templates.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Export Info */}
            <div className="border-t border-gray-200 pt-4">
              <h3 className="text-sm font-semibold text-gray-900 mb-2">Files to Export:</h3>
              <ul className="space-y-1 text-sm text-gray-600">
                <li className="flex items-center gap-2">
                  <FileSpreadsheet className="w-4 h-4 text-blue-600" />
                  <span><strong>Vacation.xlsx</strong> - {preview.vacaciones_count} record(s)</span>
                </li>
                <li className="flex items-center gap-2">
                  <FileSpreadsheet className="w-4 h-4 text-green-600" />
                  <span><strong>Absences_and_Extras.xlsx</strong> - {preview.faltas_extras_count} record(s)</span>
                </li>
              </ul>
            </div>
          </div>
        )}

        {/* Download Button */}
        {selectedPeriodId && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-1">Ready to Download</h3>
                <p className="text-sm text-gray-600">
                  Download a ZIP file containing both Excel templates
                </p>
              </div>
              <Button
                onClick={handleDownload}
                disabled={downloading || loading}
                className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg flex items-center gap-2"
              >
                {downloading ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent" />
                    Downloading...
                  </>
                ) : (
                  <>
                    <Download className="w-5 h-5" />
                    Download Export
                  </>
                )}
              </Button>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
