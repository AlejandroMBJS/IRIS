"use client"

import { useEffect, useState, useCallback } from "react"
import { useRouter } from "next/navigation"
import {
  FileText,
  Clock,
  CheckCircle,
  XCircle,
  Calendar,
  ArrowRight,
  Plus,
  AlertCircle,
  Palmtree,
  Bell,
  User,
  MessageCircle,
  Image as ImageIcon,
  ChevronRight,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { isAuthenticated, getCurrentUser, hasApprovalRole } from "@/lib/auth"
import { PortalLayout } from "@/components/layout/portal-layout"
import {
  absenceRequestApi,
  AbsenceRequest,
  REQUEST_TYPE_LABELS,
  REQUEST_STATUS_LABELS,
  APPROVAL_STAGE_LABELS,
  employeeApi,
  VacationBalance,
  announcementApi,
  Announcement,
} from "@/lib/api-client"

export default function DashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [requests, setRequests] = useState<AbsenceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [vacationBalance, setVacationBalance] = useState<VacationBalance | null>(null)
  const [announcements, setAnnouncements] = useState<Announcement[]>([])
  const [announcementLoading, setAnnouncementLoading] = useState(true)
  const [unreadCount, setUnreadCount] = useState(0)

  // Calculate statistics
  const stats = {
    total: requests.length,
    pending: requests.filter(r => r.status === 'PENDING').length,
    approved: requests.filter(r => r.status === 'APPROVED').length,
    declined: requests.filter(r => r.status === 'DECLINED').length,
  }

  // Load announcements
  const loadAnnouncements = useCallback(async () => {
    try {
      setAnnouncementLoading(true)
      const [announcementsData, unreadData] = await Promise.all([
        announcementApi.getAll(),
        announcementApi.getUnreadCount()
      ])
      setAnnouncements(announcementsData.announcements || [])
      setUnreadCount(unreadData.unread_count || 0)
    } catch (error) {
      console.error("Error loading announcements:", error)
    } finally {
      setAnnouncementLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    setUser(getCurrentUser())
    loadDashboardData()
    loadAnnouncements()
  }, [router, loadAnnouncements])

  // Refresh announcements when page becomes visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        loadAnnouncements()
      }
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [loadAnnouncements])

  async function loadDashboardData() {
    try {
      const response = await absenceRequestApi.getMyRequests()
      setRequests(response.requests || [])

      // Load vacation balance if user has employee_id
      const currentUser = getCurrentUser()
      if (currentUser?.employee_id) {
        try {
          const balance = await employeeApi.getVacationBalance(currentUser.employee_id)
          setVacationBalance(balance)
        } catch (err) {
          console.error("Error loading vacation balance:", err)
        }
      }
    } catch (error) {
      console.error("Error loading dashboard data:", error)
    } finally {
      setLoading(false)
    }
  }

  // Mark announcement as read
  const handleMarkAsRead = async (id: string) => {
    try {
      await announcementApi.markAsRead(id)
      setUnreadCount(prev => Math.max(0, prev - 1))
    } catch (error) {
      console.error("Error marking announcement as read:", error)
    }
  }

  // Format date for display
  const formatAnnouncementDate = (dateStr: string) => {
    if (!dateStr) return "-"
    try {
      const date = new Date(dateStr)
      const now = new Date()
      const diffMs = now.getTime() - date.getTime()
      const diffHours = diffMs / (1000 * 60 * 60)
      const diffDays = diffMs / (1000 * 60 * 60 * 24)

      if (diffHours < 1) {
        return "Hace unos minutos"
      } else if (diffHours < 24) {
        return `Hace ${Math.floor(diffHours)} hora${Math.floor(diffHours) > 1 ? 's' : ''}`
      } else if (diffDays < 7) {
        return `Hace ${Math.floor(diffDays)} dia${Math.floor(diffDays) > 1 ? 's' : ''}`
      } else {
        return date.toLocaleDateString("es-MX", {
          day: "2-digit",
          month: "short",
          year: "numeric"
        })
      }
    } catch {
      return dateStr
    }
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

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'PENDING':
        return 'bg-amber-500/20 text-amber-400'
      case 'APPROVED':
        return 'bg-emerald-500/20 text-emerald-400'
      case 'DECLINED':
        return 'bg-red-500/20 text-red-400'
      case 'ARCHIVED':
        return 'bg-slate-500/20 text-slate-400'
      default:
        return 'bg-slate-500/20 text-slate-400'
    }
  }

  if (loading) {
    return (
      <PortalLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Cargando...</p>
          </div>
        </div>
      </PortalLayout>
    )
  }

  return (
    <PortalLayout>
      <div className="space-y-8">
        {/* Page Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white">
              Bienvenido, {user?.full_name?.split(' ')[0] || 'Usuario'}
            </h1>
            <p className="text-slate-400 mt-1">
              Portal del Empleado - {new Date().toLocaleDateString("es-MX", { weekday: "long", year: "numeric", month: "long", day: "numeric" })}
            </p>
          </div>
          <Button
            onClick={() => router.push("/requests/new")}
            className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-lg shadow-blue-500/20"
          >
            <Plus size={18} className="mr-2" />
            Nueva Solicitud
          </Button>
        </div>

        {/* Announcements Section */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
          <div className="p-6 border-b border-slate-700/50 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl shadow-lg shadow-blue-500/30">
                <Bell size={20} className="text-white" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                  Anuncios
                  {unreadCount > 0 && (
                    <span className="px-2 py-0.5 text-xs font-medium bg-blue-500 text-white rounded-full">
                      {unreadCount} nuevo{unreadCount > 1 ? 's' : ''}
                    </span>
                  )}
                </h3>
                <p className="text-sm text-slate-400">Noticias y comunicados de la empresa</p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push("/announcements")}
              className="text-slate-400 hover:text-white"
            >
              Ver todos
              <ChevronRight size={16} className="ml-1" />
            </Button>
          </div>

          <div className="divide-y divide-slate-700/50">
            {announcementLoading ? (
              <div className="p-8 text-center">
                <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
                <p className="text-slate-400 text-sm">Cargando anuncios...</p>
              </div>
            ) : announcements.length === 0 ? (
              <div className="p-12 text-center">
                <Bell size={48} className="mx-auto text-slate-600 mb-4" />
                <p className="text-slate-400">No hay anuncios</p>
                <p className="text-sm text-slate-500 mt-2">Los anuncios de la empresa apareceran aqui</p>
              </div>
            ) : (
              announcements.slice(0, 3).map((announcement, index) => (
                <div
                  key={announcement.id}
                  className="p-5 hover:bg-slate-800/30 transition-colors cursor-pointer group"
                  onClick={() => {
                    handleMarkAsRead(announcement.id)
                    router.push("/announcements")
                  }}
                >
                  <div className="flex gap-4">
                    {/* Image or Icon */}
                    <div className="flex-shrink-0">
                      {announcement.image_data ? (
                        <div className="w-20 h-20 rounded-xl overflow-hidden border border-slate-700">
                          <img
                            src={`data:image/jpeg;base64,${announcement.image_data}`}
                            alt={announcement.title}
                            className="w-full h-full object-cover"
                          />
                        </div>
                      ) : (
                        <div className="w-20 h-20 rounded-xl bg-gradient-to-br from-slate-700 to-slate-800 flex items-center justify-center border border-slate-600">
                          <Bell size={28} className="text-slate-400" />
                        </div>
                      )}
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1 min-w-0">
                          <h4 className="text-white font-semibold truncate group-hover:text-blue-400 transition-colors">
                            {announcement.title}
                          </h4>
                          <p className="text-slate-400 text-sm mt-1 line-clamp-2">
                            {announcement.message}
                          </p>
                        </div>
                        <ArrowRight size={16} className="text-slate-500 group-hover:text-blue-400 transition-colors flex-shrink-0 mt-1" />
                      </div>

                      {/* Meta info */}
                      <div className="flex items-center gap-4 mt-3 text-xs text-slate-500">
                        <div className="flex items-center gap-1">
                          <User size={12} />
                          <span>{announcement.created_by?.full_name || "Sistema"}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Clock size={12} />
                          <span>{formatAnnouncementDate(announcement.created_at)}</span>
                        </div>
                        {announcement.image_data && (
                          <div className="flex items-center gap-1 text-blue-400">
                            <ImageIcon size={12} />
                            <span>Imagen adjunta</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Quick Stats Footer */}
          {announcements.length > 0 && (
            <div className="p-4 bg-slate-800/30 border-t border-slate-700/50">
              <div className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-6">
                  <div className="flex items-center gap-2 text-slate-400">
                    <Bell size={14} className="text-blue-400" />
                    <span>{announcements.length} anuncio{announcements.length > 1 ? 's' : ''} activo{announcements.length > 1 ? 's' : ''}</span>
                  </div>
                  {unreadCount > 0 && (
                    <div className="flex items-center gap-2 text-blue-400">
                      <div className="w-2 h-2 bg-blue-400 rounded-full animate-pulse" />
                      <span>{unreadCount} sin leer</span>
                    </div>
                  )}
                </div>
                <Button
                  variant="link"
                  size="sm"
                  onClick={() => router.push("/announcements")}
                  className="text-blue-400 hover:text-blue-300 p-0 h-auto"
                >
                  Ver todos los anuncios
                  <ChevronRight size={14} className="ml-1" />
                </Button>
              </div>
            </div>
          )}
        </div>

        {/* Mini Request Stats - Compact version */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-slate-800/50 backdrop-blur rounded-xl p-4 border border-slate-700/50 hover:border-blue-500/30 transition-all">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/20 rounded-lg">
                <FileText size={18} className="text-blue-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.total}</p>
                <p className="text-xs text-slate-400">Total Solicitudes</p>
              </div>
            </div>
          </div>
          <div className="bg-slate-800/50 backdrop-blur rounded-xl p-4 border border-slate-700/50 hover:border-amber-500/30 transition-all">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-500/20 rounded-lg">
                <Clock size={18} className="text-amber-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.pending}</p>
                <p className="text-xs text-slate-400">Pendientes</p>
              </div>
            </div>
          </div>
          <div className="bg-slate-800/50 backdrop-blur rounded-xl p-4 border border-slate-700/50 hover:border-emerald-500/30 transition-all">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-500/20 rounded-lg">
                <CheckCircle size={18} className="text-emerald-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.approved}</p>
                <p className="text-xs text-slate-400">Aprobadas</p>
              </div>
            </div>
          </div>
          <div className="bg-slate-800/50 backdrop-blur rounded-xl p-4 border border-slate-700/50 hover:border-red-500/30 transition-all">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-red-500/20 rounded-lg">
                <XCircle size={18} className="text-red-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.declined}</p>
                <p className="text-xs text-slate-400">Rechazadas</p>
              </div>
            </div>
          </div>
        </div>

        {/* Vacation Balance Widget */}
        {vacationBalance && (
          <div className="bg-gradient-to-br from-emerald-900/30 to-teal-900/30 backdrop-blur-xl rounded-2xl border border-emerald-700/50 overflow-hidden">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                  <Palmtree size={20} className="text-emerald-400" />
                  Balance de Vacaciones {vacationBalance.year}
                </h3>
                <Button
                  size="sm"
                  onClick={() => router.push("/requests/new?type=VACATION")}
                  className="bg-emerald-600 hover:bg-emerald-700"
                >
                  <Plus size={16} className="mr-1" />
                  Solicitar
                </Button>
              </div>

              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {/* Entitled Days */}
                <div className="bg-slate-800/50 rounded-xl p-4 border border-slate-700/50">
                  <p className="text-xs text-slate-400 uppercase tracking-wider mb-1">Corresponden</p>
                  <p className="text-2xl font-bold text-white">{vacationBalance.entitled_days}</p>
                  <p className="text-xs text-slate-500 mt-1">dias</p>
                </div>

                {/* Used Days */}
                <div className="bg-slate-800/50 rounded-xl p-4 border border-slate-700/50">
                  <p className="text-xs text-slate-400 uppercase tracking-wider mb-1">Usados</p>
                  <p className="text-2xl font-bold text-amber-400">{vacationBalance.used_days}</p>
                  <p className="text-xs text-slate-500 mt-1">dias</p>
                </div>

                {/* Pending Days */}
                <div className="bg-slate-800/50 rounded-xl p-4 border border-slate-700/50">
                  <p className="text-xs text-slate-400 uppercase tracking-wider mb-1">Pendientes</p>
                  <p className="text-2xl font-bold text-blue-400">{vacationBalance.pending_days}</p>
                  <p className="text-xs text-slate-500 mt-1">dias</p>
                </div>

                {/* Available Days */}
                <div className="bg-gradient-to-br from-emerald-600/20 to-emerald-700/20 rounded-xl p-4 border border-emerald-500/50">
                  <p className="text-xs text-emerald-300 uppercase tracking-wider mb-1">Disponibles</p>
                  <p className="text-2xl font-bold text-emerald-400">{vacationBalance.available_days}</p>
                  <p className="text-xs text-emerald-300 mt-1">dias</p>
                </div>
              </div>

              {/* Years of Service */}
              <div className="mt-4 pt-4 border-t border-slate-700/50">
                <p className="text-sm text-slate-400">
                  Antiguedad: <span className="text-white font-medium">{vacationBalance.years_of_service.toFixed(1)} anos</span>
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Quick Actions */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6">
          <h3 className="text-lg font-semibold text-white mb-6 flex items-center gap-2">
            <Calendar size={20} className="text-blue-400" />
            Acciones Rapidas
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <button
              onClick={() => router.push("/requests/new?type=VACATION")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-blue-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <Calendar size={24} className="text-blue-400" />
              </div>
              <span className="text-white font-medium">Vacaciones</span>
              <span className="text-xs text-slate-400 mt-1">Solicitar dias</span>
            </button>

            <button
              onClick={() => router.push("/requests/new?type=PAID_LEAVE")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-emerald-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <CheckCircle size={24} className="text-emerald-400" />
              </div>
              <span className="text-white font-medium">Permiso con Goce</span>
              <span className="text-xs text-slate-400 mt-1">Permiso pagado</span>
            </button>

            <button
              onClick={() => router.push("/requests/new?type=LATE_ENTRY")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-amber-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <Clock size={24} className="text-amber-400" />
              </div>
              <span className="text-white font-medium">Pase de Entrada</span>
              <span className="text-xs text-slate-400 mt-1">Llegada tarde</span>
            </button>

            <button
              onClick={() => router.push("/requests/new?type=EARLY_EXIT")}
              className="group relative flex flex-col items-center p-6 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300"
            >
              <div className="p-4 bg-purple-500/20 rounded-xl mb-4 group-hover:scale-110 transition-transform">
                <ArrowRight size={24} className="text-purple-400" />
              </div>
              <span className="text-white font-medium">Pase de Salida</span>
              <span className="text-xs text-slate-400 mt-1">Salida temprano</span>
            </button>
          </div>
        </div>

        {/* Recent Requests */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
          <div className="p-6 border-b border-slate-700/50 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <FileText size={20} className="text-blue-400" />
              Solicitudes Recientes
            </h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push("/requests")}
              className="text-slate-400 hover:text-white"
            >
              Ver todas
              <ArrowRight size={16} className="ml-1" />
            </Button>
          </div>
          <div className="divide-y divide-slate-700/50">
            {requests.length === 0 ? (
              <div className="p-12 text-center">
                <FileText size={48} className="mx-auto text-slate-600 mb-4" />
                <p className="text-slate-400">No tienes solicitudes</p>
                <p className="text-sm text-slate-500 mt-2">Crea tu primera solicitud de ausencia</p>
                <Button
                  onClick={() => router.push("/requests/new")}
                  className="mt-4 bg-blue-600 hover:bg-blue-700"
                >
                  Nueva Solicitud
                </Button>
              </div>
            ) : (
              requests.slice(0, 5).map((request) => (
                <div
                  key={request.id}
                  className="p-4 hover:bg-slate-800/30 transition-colors cursor-pointer"
                  onClick={() => router.push(`/requests/${request.id}`)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className={`p-3 rounded-xl ${
                        request.request_type === 'VACATION'
                          ? "bg-blue-500/20 text-blue-400"
                          : request.request_type === 'PAID_LEAVE'
                          ? "bg-emerald-500/20 text-emerald-400"
                          : request.request_type === 'SICK_LEAVE'
                          ? "bg-red-500/20 text-red-400"
                          : "bg-amber-500/20 text-amber-400"
                      }`}>
                        <FileText size={20} />
                      </div>
                      <div>
                        <p className="text-white font-medium">
                          {REQUEST_TYPE_LABELS[request.request_type] || request.request_type}
                        </p>
                        <p className="text-sm text-slate-400">
                          {formatDate(request.start_date)} - {formatDate(request.end_date)} ({request.total_days} dias)
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      <span className={`inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-full ${getStatusColor(request.status)}`}>
                        {REQUEST_STATUS_LABELS[request.status] || request.status}
                      </span>
                      {request.status === 'PENDING' && (
                        <p className="text-xs text-slate-500 mt-1">
                          En: {APPROVAL_STAGE_LABELS[request.current_approval_stage] || request.current_approval_stage}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Info Banner for Approvers */}
        {hasApprovalRole() && (
          <div className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/30 rounded-xl p-4">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-blue-400 mt-0.5" />
              <div>
                <p className="text-white font-medium">Tienes solicitudes pendientes de aprobar</p>
                <p className="text-sm text-slate-400 mt-1">
                  Revisa la seccion de Aprobaciones en el menu lateral para ver las solicitudes que requieren tu atencion.
                </p>
              </div>
            </div>
          </div>
        )}
      </div>
    </PortalLayout>
  )
}
