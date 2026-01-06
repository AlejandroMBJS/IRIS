/**
 * @file app/payroll/[periodId]/payslip/[employeeId]/page.tsx
 * @description Individual employee payslip view with print and download options
 *
 * USER PERSPECTIVE:
 *   - View detailed payslip for a specific employee and period
 *   - See earnings breakdown (salary, overtime, bonuses, etc.)
 *   - See deductions breakdown (ISR, IMSS, loans, etc.)
 *   - View net pay amount
 *   - Print or download PDF version of payslip
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Payslip layout, earnings/deductions display format
 *   CAUTION: Print styles use print: classes, test thoroughly
 *   DO NOT modify: Calculation display without verifying backend data structure
 *
 * KEY COMPONENTS:
 *   - Payslip header: Company info, period, employee details
 *   - Earnings table: All income items
 *   - Deductions table: All deduction items
 *   - Net pay display: Final amount to be paid
 *   - Print/download buttons
 *
 * API ENDPOINTS USED:
 *   - GET /api/employees/:id (via employeeApi.getEmployee)
 *   - GET /api/payroll/periods/:periodId/employees/:employeeId (via payrollApi.getPayslip)
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import { ArrowLeft, Download, Printer, FileText } from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { payrollApi, employeeApi, Employee, PayrollCalculationResponse, PayrollPeriod, ApiError } from "@/lib/api-client"

export default function PayslipPage() {
  const router = useRouter()
  const params = useParams()
  const periodId = params.periodId as string
  const employeeId = params.employeeId as string

  const [employee, setEmployee] = useState<Employee | null>(null)
  const [payroll, setPayroll] = useState<PayrollCalculationResponse | null>(null)
  const [period, setPeriod] = useState<PayrollPeriod | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (periodId && employeeId) {
      loadData()
    }
  }, [periodId, employeeId])

  async function loadData() {
    try {
      setLoading(true)
      setError(null)

      const [employeeData, payslipData, periodData] = await Promise.all([
        employeeApi.getEmployee(employeeId).catch(() => null),
        payrollApi.getPayrollCalculation(periodId, employeeId).catch(() => null),
        payrollApi.getPeriod(periodId).catch(() => null),
      ])

      setEmployee(employeeData)
      setPayroll(payslipData)
      setPeriod(periodData)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load payslip data")
      }
    } finally {
      setLoading(false)
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
        year: "numeric",
        month: "long",
        day: "numeric",
      })
    } catch {
      return dateStr
    }
  }

  const [downloading, setDownloading] = useState(false)
  const [printing, setPrinting] = useState(false)

  const handlePrint = async () => {
    try {
      setPrinting(true)

      // Fetch the PDF as a blob (auth via httpOnly cookies)
      const response = await fetch(`/api/v1/payroll/payslip/${periodId}/${employeeId}?format=pdf`, {
        method: "GET",
        credentials: "include",
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: "Failed to load PDF for printing" }))
        throw new Error(errorData.message || "Failed to load PDF for printing")
      }

      // Get the blob and create URL
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)

      // Open PDF in new window and print
      const printWindow = window.open(url, "_blank")
      if (printWindow) {
        printWindow.addEventListener("load", () => {
          printWindow.print()
        })
      } else {
        // Fallback: create iframe and print
        const iframe = document.createElement("iframe")
        iframe.style.display = "none"
        iframe.src = url
        document.body.appendChild(iframe)
        iframe.onload = () => {
          iframe.contentWindow?.print()
          setTimeout(() => {
            document.body.removeChild(iframe)
            window.URL.revokeObjectURL(url)
          }, 1000)
        }
      }
    } catch (err) {
      console.error("Print error:", err)
      setError(err instanceof Error ? err.message : "Failed to print PDF")
    } finally {
      setPrinting(false)
    }
  }

  const handleDownload = async () => {
    try {
      setDownloading(true)

      const employeeNumber = employee?.employee_number || employeeId
      const baseFilename = `recibo_${employeeNumber}_${periodId}`

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

      // Check responses
      if (!pdfResponse.ok) {
        const errorData = await pdfResponse.json().catch(() => ({ message: "Failed to download PDF" }))
        throw new Error(errorData.message || "Failed to download PDF")
      }

      // Download PDF
      const pdfBlob = await pdfResponse.blob()
      const pdfUrl = window.URL.createObjectURL(pdfBlob)
      const pdfLink = document.createElement("a")
      pdfLink.href = pdfUrl

      // Try to get filename from Content-Disposition header
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

        // Try to get filename from Content-Disposition header
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
      } else {
        console.warn("XML/CFDI download not available, only PDF was downloaded")
      }
    } catch (err) {
      console.error("Download error:", err)
      setError(err instanceof Error ? err.message : "Failed to download files")
    } finally {
      setDownloading(false)
    }
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="flex items-center gap-2 text-slate-400">
            <div className="w-5 h-5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin" />
            Loading payslip...
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
            onClick={() => router.push(`/payroll/${periodId}`)}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={20} />
            Back to Payroll Summary
          </button>
          <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 text-red-400">
            {error}
          </div>
        </div>
      </DashboardLayout>
    )
  }

  // Build earnings and deductions arrays from payroll data with SAT codes
  const earnings = [
    { code: "001", concept: "Sueldos, Salarios", amount: payroll?.regular_salary || 0 },
    payroll?.overtime_amount ? { code: "019", concept: "Horas Extra", amount: payroll.overtime_amount } : null,
    payroll?.vacation_premium ? { code: "021", concept: "Prima Vacacional", amount: payroll.vacation_premium } : null,
    payroll?.aguinaldo ? { code: "002", concept: "Aguinaldo", amount: payroll.aguinaldo } : null,
    payroll?.food_vouchers ? { code: "029", concept: "Vales de Despensa", amount: payroll.food_vouchers } : null,
    payroll?.savings_fund ? { code: "005", concept: "Fondo de Ahorro", amount: payroll.savings_fund } : null,
    payroll?.other_extras ? { code: "028", concept: "Bonos y Otros", amount: payroll.other_extras } : null,
  ].filter((e): e is { code: string; concept: string; amount: number } => e !== null && e.amount > 0)

  const deductions = [
    { code: "002", concept: "ISR (Impuesto Sobre la Renta)", amount: payroll?.isr_withholding || 0 },
    { code: "001", concept: "IMSS (Cuotas Obrero)", amount: payroll?.imss_employee || 0 },
    payroll?.infonavit_employee ? { code: "010", concept: "INFONAVIT", amount: payroll.infonavit_employee } : null,
    payroll?.retirement_savings ? { code: "017", concept: "Aportaciones SAR", amount: payroll.retirement_savings } : null,
    payroll?.loan_deductions ? { code: "004", concept: "Préstamos", amount: payroll.loan_deductions } : null,
    payroll?.advance_deductions ? { code: "012", concept: "Anticipos", amount: payroll.advance_deductions } : null,
    payroll?.other_deductions ? { code: "006", concept: "Otras Deducciones", amount: payroll.other_deductions } : null,
  ].filter((d): d is { code: string; concept: string; amount: number } => d !== null && d.amount > 0)

  const employerContributions = payroll?.employer_contributions
    ? [
        { concept: "IMSS Patronal", amount: payroll.employer_contributions.total_imss || 0 },
        { concept: "INFONAVIT Patronal", amount: payroll.employer_contributions.total_infonavit || 0 },
        { concept: "SAR/Retiro", amount: payroll.employer_contributions.total_retirement || 0 },
      ].filter((c) => c.amount > 0)
    : []

  const getContractTypeLabel = (type?: string) => {
    const labels: Record<string, string> = {
      "indefinite": "Contrato Indefinido",
      "temporary": "Contrato Temporal",
      "training": "Contrato de Capacitación",
      "seasonal": "Contrato por Temporada"
    }
    return labels[type || ""] || type || "-"
  }

  const getRegimeLabel = (regime?: string) => {
    const labels: Record<string, string> = {
      "02": "Sueldos",
      "05": "Libre Ejercicio",
      "08": "Comisionistas",
      "09": "Honorarios"
    }
    return labels[regime || ""] || regime || "02 - Sueldos"
  }

  const getFrequencyLabel = (frequency?: string) => {
    const labels: Record<string, string> = {
      "weekly": "Semanal",
      "biweekly": "Quincenal",
      "monthly": "Mensual"
    }
    return labels[frequency?.toLowerCase() || ""] || frequency || "-"
  }

  return (
    <DashboardLayout>
      <div className="space-y-6 print:space-y-4">
        {/* Header - Hide on print */}
        <div className="flex items-center justify-between print:hidden">
          <div className="flex items-center gap-4">
            <button
              onClick={() => router.push(`/payroll/${periodId}`)}
              className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
            >
              <ArrowLeft size={20} />
            </button>
            <div>
              <h1 className="text-3xl font-bold text-white">Recibo de Nómina</h1>
              <p className="text-slate-400 mt-1">{employee?.full_name || payroll?.employee_name}</p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button
              onClick={handlePrint}
              variant="outline"
              className="border-slate-600"
              disabled={printing || downloading}
            >
              {printing ? (
                <>
                  <div className="w-4 h-4 border-2 border-slate-400 border-t-transparent rounded-full animate-spin mr-2" />
                  Cargando...
                </>
              ) : (
                <>
                  <Printer size={20} className="mr-2" />
                  Imprimir
                </>
              )}
            </Button>
            <Button
              onClick={handleDownload}
              className="bg-blue-600 hover:bg-blue-700"
              disabled={printing || downloading}
            >
              {downloading ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                  Descargando...
                </>
              ) : (
                <>
                  <Download size={20} className="mr-2" />
                  Descargar PDF y XML
                </>
              )}
            </Button>
          </div>
        </div>

        {/* Payslip Content */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-lg border border-slate-700 p-6 print:bg-white print:text-black print:border-gray-300">
          {/* Company Header */}
          <div className="flex justify-between items-start mb-6 pb-4 border-b border-slate-700 print:border-gray-300">
            <div>
              <h2 className="text-2xl font-bold text-white print:text-black">IRIS Talent S.A. de C.V.</h2>
              <p className="text-slate-400 print:text-gray-600 text-sm mt-1">RFC: EKU9003173C9</p>
              <p className="text-slate-400 print:text-gray-600 text-sm">Reg. Patronal IMSS: Y6479105106</p>
            </div>
            <div className="text-right">
              <div className="px-4 py-2 bg-blue-500/20 border border-blue-500/30 rounded-lg">
                <p className="text-xs text-blue-300 print:text-blue-600">RECIBO DE NÓMINA</p>
                <p className="text-lg font-bold text-white print:text-black">{payroll?.period_code || periodId}</p>
              </div>
            </div>
          </div>

          {/* Period Information */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6 pb-4 border-b border-slate-700 print:border-gray-300 bg-slate-900/30 p-4 rounded-lg">
            <div>
              <p className="text-xs text-slate-500 print:text-gray-500">Frecuencia</p>
              <p className="text-sm text-white print:text-black font-medium">{getFrequencyLabel(period?.frequency)}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500 print:text-gray-500">Fecha Inicio</p>
              <p className="text-sm text-white print:text-black font-medium">{formatDate(period?.start_date)}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500 print:text-gray-500">Fecha Fin</p>
              <p className="text-sm text-white print:text-black font-medium">{formatDate(period?.end_date)}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500 print:text-gray-500">Fecha de Pago</p>
              <p className="text-sm text-white print:text-black font-medium">{formatDate(period?.payment_date)}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500 print:text-gray-500">Días Trabajados</p>
              <p className="text-sm text-white print:text-black font-medium">{payroll?.days_worked || period?.working_days || "-"}</p>
            </div>
          </div>

          {/* Employee Information - Detailed */}
          <div className="mb-6 pb-4 border-b border-slate-700 print:border-gray-300">
            <h3 className="text-sm font-semibold text-slate-400 print:text-gray-600 mb-3">DATOS DEL TRABAJADOR</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">No. Empleado</p>
                <p className="text-sm text-white print:text-black font-medium font-mono">
                  {employee?.employee_number || payroll?.employee_number}
                </p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">RFC</p>
                <p className="text-sm text-white print:text-black font-medium font-mono">{employee?.rfc || "-"}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">NSS (IMSS)</p>
                <p className="text-sm text-white print:text-black font-medium font-mono">{employee?.nss || "-"}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">CURP</p>
                <p className="text-sm text-white print:text-black font-medium font-mono">{employee?.curp || "-"}</p>
              </div>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
              <div className="col-span-2">
                <p className="text-xs text-slate-500 print:text-gray-500">Nombre Completo</p>
                <p className="text-sm text-white print:text-black font-medium">
                  {employee?.full_name || payroll?.employee_name}
                </p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">Tipo Contrato</p>
                <p className="text-sm text-white print:text-black font-medium">{getContractTypeLabel(employee?.contract_type)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">Régimen</p>
                <p className="text-sm text-white print:text-black font-medium">{getRegimeLabel(employee?.tax_regime)}</p>
              </div>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">Periodicidad</p>
                <p className="text-sm text-white print:text-black font-medium">{getFrequencyLabel(employee?.payment_frequency)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">Salario Diario</p>
                <p className="text-sm text-emerald-400 print:text-emerald-700 font-medium">{formatCurrency(employee?.daily_salary)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">SDI</p>
                <p className="text-sm text-emerald-400 print:text-emerald-700 font-medium">{formatCurrency(payroll?.sdi || employee?.sdi)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500 print:text-gray-500">Fecha Ingreso</p>
                <p className="text-sm text-white print:text-black font-medium">{formatDate(employee?.hire_date)}</p>
              </div>
            </div>
          </div>

          {/* Earnings and Deductions Tables */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            {/* PERCEPCIONES (Earnings) */}
            <div className="border border-slate-700 print:border-gray-300 rounded-lg overflow-hidden">
              <div className="bg-emerald-900/30 print:bg-emerald-50 px-4 py-3 border-b border-slate-700 print:border-gray-300">
                <h3 className="text-lg font-semibold text-emerald-400 print:text-emerald-700">PERCEPCIONES</h3>
              </div>
              <table className="w-full">
                <thead className="bg-slate-900/50 print:bg-gray-100">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-slate-400 print:text-gray-600">Clave</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-slate-400 print:text-gray-600">Concepto</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-slate-400 print:text-gray-600">Importe</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-700 print:divide-gray-200">
                  {earnings.map((earning, i) => (
                    <tr key={i}>
                      <td className="px-3 py-2 text-xs text-slate-400 print:text-gray-500 font-mono">{earning.code}</td>
                      <td className="px-3 py-2 text-sm text-slate-300 print:text-gray-700">{earning.concept}</td>
                      <td className="px-3 py-2 text-sm text-white print:text-black text-right font-medium">
                        {formatCurrency(earning.amount)}
                      </td>
                    </tr>
                  ))}
                  {earnings.length === 0 && (
                    <tr>
                      <td colSpan={3} className="px-3 py-4 text-center text-sm text-slate-500">
                        Sin percepciones
                      </td>
                    </tr>
                  )}
                </tbody>
                <tfoot className="bg-emerald-900/20 print:bg-emerald-50">
                  <tr>
                    <td colSpan={2} className="px-3 py-3 text-sm font-bold text-emerald-400 print:text-emerald-700">
                      Total Percepciones
                    </td>
                    <td className="px-3 py-3 text-sm font-bold text-emerald-400 print:text-emerald-700 text-right">
                      {formatCurrency(payroll?.total_gross_income)}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>

            {/* DEDUCCIONES (Deductions) */}
            <div className="border border-slate-700 print:border-gray-300 rounded-lg overflow-hidden">
              <div className="bg-red-900/30 print:bg-red-50 px-4 py-3 border-b border-slate-700 print:border-gray-300">
                <h3 className="text-lg font-semibold text-red-400 print:text-red-700">DEDUCCIONES</h3>
              </div>
              <table className="w-full">
                <thead className="bg-slate-900/50 print:bg-gray-100">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-slate-400 print:text-gray-600">Clave</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-slate-400 print:text-gray-600">Concepto</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-slate-400 print:text-gray-600">Importe</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-700 print:divide-gray-200">
                  {deductions.map((deduction, i) => (
                    <tr key={i}>
                      <td className="px-3 py-2 text-xs text-slate-400 print:text-gray-500 font-mono">{deduction.code}</td>
                      <td className="px-3 py-2 text-sm text-slate-300 print:text-gray-700">{deduction.concept}</td>
                      <td className="px-3 py-2 text-sm text-white print:text-black text-right font-medium">
                        {formatCurrency(deduction.amount)}
                      </td>
                    </tr>
                  ))}
                  {deductions.length === 0 && (
                    <tr>
                      <td colSpan={3} className="px-3 py-4 text-center text-sm text-slate-500">
                        Sin deducciones
                      </td>
                    </tr>
                  )}
                </tbody>
                <tfoot className="bg-red-900/20 print:bg-red-50">
                  <tr>
                    <td colSpan={2} className="px-3 py-3 text-sm font-bold text-red-400 print:text-red-700">
                      Total Deducciones
                    </td>
                    <td className="px-3 py-3 text-sm font-bold text-red-400 print:text-red-700 text-right">
                      {formatCurrency(payroll?.total_deductions)}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>

          {/* Net Pay - Prominent */}
          <div className="bg-gradient-to-r from-blue-900/40 to-cyan-900/40 print:bg-blue-50 border border-blue-600/50 print:border-blue-300 rounded-xl p-6 mb-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-blue-300 print:text-blue-700">NETO A PAGAR</h3>
                <p className="text-xs text-blue-400/70 print:text-blue-600 mt-1">Total percepciones menos total deducciones</p>
              </div>
              <p className="text-4xl font-bold text-white print:text-blue-700">
                {formatCurrency(payroll?.total_net_pay)}
              </p>
            </div>
          </div>

          {/* Subsidio al Empleo (if applicable) */}
          {payroll?.employment_subsidy && payroll.employment_subsidy > 0 && (
            <div className="bg-amber-900/20 print:bg-amber-50 border border-amber-600/30 print:border-amber-300 rounded-lg p-4 mb-6">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-sm font-semibold text-amber-400 print:text-amber-700">SUBSIDIO PARA EL EMPLEO</h4>
                  <p className="text-xs text-amber-400/70 print:text-amber-600">Art. Dec. Trans. LISR</p>
                </div>
                <p className="text-xl font-bold text-amber-400 print:text-amber-700">
                  {formatCurrency(payroll.employment_subsidy)}
                </p>
              </div>
            </div>
          )}

          {/* Employer Contributions */}
          {employerContributions.length > 0 && (
            <div className="bg-purple-900/20 print:bg-purple-50 border border-purple-600/30 print:border-purple-300 rounded-lg p-4 mb-6">
              <h4 className="text-sm font-semibold text-purple-400 print:text-purple-700 mb-3">APORTACIONES PATRONALES (Informativo)</h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {employerContributions.map((contrib, i) => (
                  <div key={i}>
                    <p className="text-xs text-purple-400/70 print:text-purple-600">{contrib.concept}</p>
                    <p className="text-sm text-purple-300 print:text-purple-700 font-medium">{formatCurrency(contrib.amount)}</p>
                  </div>
                ))}
                <div>
                  <p className="text-xs text-purple-400/70 print:text-purple-600">Total Patronal</p>
                  <p className="text-sm text-purple-400 print:text-purple-700 font-bold">
                    {formatCurrency(payroll?.employer_contributions?.total_contributions)}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Legal Notice */}
          <div className="mt-6 pt-4 border-t border-slate-700 print:border-gray-300">
            <p className="text-xs text-slate-500 print:text-gray-500 text-center">
              Este documento es un comprobante de pago emitido conforme a lo dispuesto en la Ley Federal del Trabajo Art. 804.
              <br />
              El CFDI de nómina correspondiente se genera de acuerdo a las disposiciones del SAT.
            </p>
          </div>
        </div>

        {/* Status Footer */}
        <div className="flex justify-between items-center text-sm text-slate-500 print:hidden">
          <span>Estatus: {payroll?.payroll_status || "Calculado"}</span>
          <span>Generado por IRIS Payroll System</span>
        </div>
      </div>
    </DashboardLayout>
  )
}
