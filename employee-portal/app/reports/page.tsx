"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  BarChart3,
  TrendingUp,
  Calendar,
  Clock,
  CheckCircle,
  XCircle,
  Palmtree,
  FileText,
  Activity,
  Users,
} from "lucide-react"
import { Card } from "@/components/ui/card"
import { isAuthenticated, getCurrentUser } from "@/lib/auth"
import { PortalLayout } from "@/components/layout/portal-layout"
import {
  absenceRequestApi,
  AbsenceRequest,
  REQUEST_TYPE_LABELS,
  REQUEST_STATUS_LABELS,
  employeeApi,
  VacationBalance,
} from "@/lib/api-client"

export default function ReportsPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [requests, setRequests] = useState<AbsenceRequest[]>([])
  const [vacationBalance, setVacationBalance] = useState<VacationBalance | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    const currentUser = getCurrentUser()
    setUser(currentUser)
    loadReportData(currentUser)
  }, [router])

  async function loadReportData(currentUser: any) {
    try {
      setLoading(true)
      const response = await absenceRequestApi.getMyRequests()
      setRequests(response.requests || [])

      // Load vacation balance if user has employee_id
      if (currentUser?.employee_id) {
        try {
          const balance = await employeeApi.getVacationBalance(currentUser.employee_id)
          setVacationBalance(balance)
        } catch (err) {
          console.error("Error loading vacation balance:", err)
        }
      }
    } catch (error) {
      console.error("Error loading report data:", error)
    } finally {
      setLoading(false)
    }
  }

  // Calculate statistics
  const currentYear = new Date().getFullYear()
  const yearRequests = requests.filter(r => {
    const year = new Date(r.created_at).getFullYear()
    return year === currentYear
  })

  const stats = {
    total: yearRequests.length,
    pending: yearRequests.filter(r => r.status === 'PENDING').length,
    approved: yearRequests.filter(r => r.status === 'APPROVED').length,
    declined: yearRequests.filter(r => r.status === 'DECLINED').length,
  }

  // Request type distribution
  const requestTypeStats = Object.entries(
    yearRequests.reduce((acc, req) => {
      acc[req.request_type] = (acc[req.request_type] || 0) + 1
      return acc
    }, {} as Record<string, number>)
  ).map(([type, count]) => ({
    type,
    label: REQUEST_TYPE_LABELS[type as keyof typeof REQUEST_TYPE_LABELS] || type,
    count,
    percentage: ((count / yearRequests.length) * 100).toFixed(1)
  }))

  // Monthly distribution
  const monthlyStats = Array.from({ length: 12 }, (_, i) => {
    const month = i
    const monthRequests = yearRequests.filter(r => {
      const reqMonth = new Date(r.created_at).getMonth()
      return reqMonth === month
    })
    return {
      month: new Date(2000, i, 1).toLocaleDateString('es-MX', { month: 'short' }),
      count: monthRequests.length,
      approved: monthRequests.filter(r => r.status === 'APPROVED').length
    }
  })

  const maxMonthlyCount = Math.max(...monthlyStats.map(m => m.count), 1)

  // Average processing time
  const processedRequests = yearRequests.filter(r => r.status !== 'PENDING')
  const avgProcessingDays = processedRequests.length > 0
    ? processedRequests.reduce((sum, req) => {
        const created = new Date(req.created_at)
        const updated = new Date(req.updated_at)
        const days = Math.ceil((updated.getTime() - created.getTime()) / (1000 * 60 * 60 * 24))
        return sum + days
      }, 0) / processedRequests.length
    : 0

  if (loading) {
    return (
      <PortalLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Cargando reportes...</p>
          </div>
        </div>
      </PortalLayout>
    )
  }

  return (
    <PortalLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold text-white flex items-center gap-3">
            <BarChart3 className="h-8 w-8 text-blue-400" />
            Reportes y Analíticas
          </h1>
          <p className="text-slate-400 mt-2">
            Estadísticas y tendencias de tus solicitudes ({currentYear})
          </p>
        </div>

        {/* Key Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          <Card className="bg-gradient-to-br from-blue-900/30 to-blue-800/30 backdrop-blur-sm border-blue-700/50 p-6">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-blue-300 text-sm font-medium">Total Solicitudes</p>
                <p className="text-4xl font-bold text-white">{stats.total}</p>
                <p className="text-xs text-blue-300/70">Este año</p>
              </div>
              <div className="p-3 bg-blue-500/20 rounded-xl">
                <FileText className="h-6 w-6 text-blue-400" />
              </div>
            </div>
          </Card>

          <Card className="bg-gradient-to-br from-amber-900/30 to-amber-800/30 backdrop-blur-sm border-amber-700/50 p-6">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-amber-300 text-sm font-medium">Pendientes</p>
                <p className="text-4xl font-bold text-white">{stats.pending}</p>
                <p className="text-xs text-amber-300/70">En revisión</p>
              </div>
              <div className="p-3 bg-amber-500/20 rounded-xl">
                <Clock className="h-6 w-6 text-amber-400" />
              </div>
            </div>
          </Card>

          <Card className="bg-gradient-to-br from-emerald-900/30 to-emerald-800/30 backdrop-blur-sm border-emerald-700/50 p-6">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-emerald-300 text-sm font-medium">Aprobadas</p>
                <p className="text-4xl font-bold text-white">{stats.approved}</p>
                <p className="text-xs text-emerald-300/70">
                  {stats.total > 0 ? ((stats.approved / stats.total) * 100).toFixed(0) : 0}% del total
                </p>
              </div>
              <div className="p-3 bg-emerald-500/20 rounded-xl">
                <CheckCircle className="h-6 w-6 text-emerald-400" />
              </div>
            </div>
          </Card>

          <Card className="bg-gradient-to-br from-red-900/30 to-red-800/30 backdrop-blur-sm border-red-700/50 p-6">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-red-300 text-sm font-medium">Rechazadas</p>
                <p className="text-4xl font-bold text-white">{stats.declined}</p>
                <p className="text-xs text-red-300/70">
                  {stats.total > 0 ? ((stats.declined / stats.total) * 100).toFixed(0) : 0}% del total
                </p>
              </div>
              <div className="p-3 bg-red-500/20 rounded-xl">
                <XCircle className="h-6 w-6 text-red-400" />
              </div>
            </div>
          </Card>
        </div>

        {/* Vacation Balance Summary */}
        {vacationBalance && (
          <Card className="bg-gradient-to-br from-teal-900/30 to-cyan-900/30 backdrop-blur-sm border-teal-700/50 p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="p-3 bg-teal-500/20 rounded-xl">
                  <Palmtree className="h-6 w-6 text-teal-400" />
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-white">Balance de Vacaciones</h3>
                  <p className="text-sm text-teal-300/70">Resumen anual</p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-slate-900/30 rounded-lg p-4 border border-slate-700/50">
                <p className="text-xs text-slate-400 uppercase mb-1">Corresponden</p>
                <p className="text-2xl font-bold text-white">{vacationBalance.entitled_days}</p>
              </div>
              <div className="bg-slate-900/30 rounded-lg p-4 border border-slate-700/50">
                <p className="text-xs text-slate-400 uppercase mb-1">Usados</p>
                <p className="text-2xl font-bold text-amber-400">{vacationBalance.used_days}</p>
              </div>
              <div className="bg-slate-900/30 rounded-lg p-4 border border-slate-700/50">
                <p className="text-xs text-slate-400 uppercase mb-1">Pendientes</p>
                <p className="text-2xl font-bold text-blue-400">{vacationBalance.pending_days}</p>
              </div>
              <div className="bg-gradient-to-br from-emerald-600/20 to-emerald-700/20 rounded-lg p-4 border border-emerald-500/50">
                <p className="text-xs text-emerald-300 uppercase mb-1">Disponibles</p>
                <p className="text-2xl font-bold text-emerald-400">{vacationBalance.available_days}</p>
              </div>
            </div>

            {/* Progress Bar */}
            <div className="mt-4">
              <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
                <span>Uso de vacaciones</span>
                <span>
                  {vacationBalance.entitled_days > 0
                    ? ((vacationBalance.used_days / vacationBalance.entitled_days) * 100).toFixed(0)
                    : 0}%
                </span>
              </div>
              <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-teal-500 to-cyan-500 transition-all"
                  style={{
                    width: `${Math.min((vacationBalance.used_days / Math.max(vacationBalance.entitled_days, 1)) * 100, 100)}%`
                  }}
                />
              </div>
            </div>
          </Card>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Monthly Trend Chart */}
          <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 p-6">
            <div className="flex items-center gap-3 mb-6">
              <TrendingUp className="h-5 w-5 text-blue-400" />
              <h3 className="text-lg font-semibold text-white">Tendencia Mensual</h3>
            </div>

            <div className="space-y-3">
              {monthlyStats.map((month, idx) => (
                <div key={idx} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-slate-300 font-medium">{month.month}</span>
                    <div className="flex items-center gap-3">
                      <span className="text-emerald-400 text-xs">
                        {month.approved} aprobadas
                      </span>
                      <span className="text-white font-semibold">{month.count}</span>
                    </div>
                  </div>
                  <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-blue-500 to-purple-500"
                      style={{ width: `${(month.count / maxMonthlyCount) * 100}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </Card>

          {/* Request Type Distribution */}
          <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 p-6">
            <div className="flex items-center gap-3 mb-6">
              <Activity className="h-5 w-5 text-purple-400" />
              <h3 className="text-lg font-semibold text-white">Distribución por Tipo</h3>
            </div>

            {requestTypeStats.length === 0 ? (
              <div className="text-center py-8 text-slate-500">
                <FileText className="h-12 w-12 mx-auto mb-3 opacity-50" />
                <p>No hay solicitudes este año</p>
              </div>
            ) : (
              <div className="space-y-4">
                {requestTypeStats.map((item, idx) => (
                  <div key={idx} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-slate-300 font-medium">{item.label}</span>
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-slate-400">{item.percentage}%</span>
                        <span className="text-sm text-white font-semibold">{item.count}</span>
                      </div>
                    </div>
                    <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                      <div
                        className={`h-full ${
                          idx === 0
                            ? "bg-gradient-to-r from-blue-500 to-blue-600"
                            : idx === 1
                            ? "bg-gradient-to-r from-emerald-500 to-emerald-600"
                            : idx === 2
                            ? "bg-gradient-to-r from-amber-500 to-amber-600"
                            : "bg-gradient-to-r from-purple-500 to-purple-600"
                        }`}
                        style={{ width: `${item.percentage}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>

        {/* Additional Insights */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 p-6">
            <div className="flex items-center gap-3 mb-4">
              <Calendar className="h-5 w-5 text-cyan-400" />
              <h3 className="text-lg font-semibold text-white">Tiempo de Procesamiento</h3>
            </div>
            <div className="flex items-baseline gap-2">
              <p className="text-4xl font-bold text-white">{avgProcessingDays.toFixed(1)}</p>
              <p className="text-slate-400">días promedio</p>
            </div>
            <p className="text-sm text-slate-500 mt-2">
              Tiempo promedio desde solicitud hasta resolución
            </p>
          </Card>

          <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 p-6">
            <div className="flex items-center gap-3 mb-4">
              <CheckCircle className="h-5 w-5 text-green-400" />
              <h3 className="text-lg font-semibold text-white">Tasa de Aprobación</h3>
            </div>
            <div className="flex items-baseline gap-2">
              <p className="text-4xl font-bold text-white">
                {processedRequests.length > 0
                  ? ((stats.approved / processedRequests.length) * 100).toFixed(0)
                  : 0}%
              </p>
              <p className="text-slate-400">aprobadas</p>
            </div>
            <p className="text-sm text-slate-500 mt-2">
              De {processedRequests.length} solicitudes procesadas
            </p>
          </Card>
        </div>
      </div>
    </PortalLayout>
  )
}
