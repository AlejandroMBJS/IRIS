"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  FileText,
  Plus,
  Clock,
  CheckCircle,
  XCircle,
  Archive,
  Trash2,
  RefreshCw,
  Filter,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { isAuthenticated, getCurrentUser } from "@/lib/auth"
import { PortalLayout } from "@/components/layout/portal-layout"
import {
  absenceRequestApi,
  AbsenceRequest,
  REQUEST_TYPE_LABELS,
  REQUEST_STATUS_LABELS,
  APPROVAL_STAGE_LABELS,
  RequestStatus,
  RequestType,
} from "@/lib/api-client"
import { useToast } from "@/hooks/use-toast"

export default function MyRequestsPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [user, setUser] = useState<any>(null)
  const [requests, setRequests] = useState<AbsenceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState<string>('ALL')
  const [typeFilter, setTypeFilter] = useState<string>('ALL')

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    setUser(getCurrentUser())
    loadRequests()
  }, [router])

  async function loadRequests() {
    setLoading(true)
    try {
      const response = await absenceRequestApi.getMyRequests()
      setRequests(response.requests || [])
    } catch (error) {
      console.error("Error loading requests:", error)
      toast({
        title: "Error",
        description: "No se pudieron cargar las solicitudes",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (requestId: string) => {
    if (!confirm("Estas seguro de eliminar esta solicitud?")) return

    try {
      await absenceRequestApi.delete(requestId)
      toast({
        title: "Solicitud Eliminada",
        description: "La solicitud ha sido eliminada correctamente",
      })
      loadRequests()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "No se pudo eliminar la solicitud",
        variant: "destructive",
      })
    }
  }

  const handleArchive = async (requestId: string) => {
    try {
      await absenceRequestApi.archive(requestId)
      toast({
        title: "Solicitud Archivada",
        description: "La solicitud ha sido archivada correctamente",
      })
      loadRequests()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "No se pudo archivar la solicitud",
        variant: "destructive",
      })
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

  const getStatusColor = (status: RequestStatus) => {
    switch (status) {
      case 'PENDING':
        return 'bg-amber-500/20 text-amber-400 border-amber-500/30'
      case 'APPROVED':
        return 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30'
      case 'DECLINED':
        return 'bg-red-500/20 text-red-400 border-red-500/30'
      case 'ARCHIVED':
        return 'bg-slate-500/20 text-slate-400 border-slate-500/30'
      default:
        return 'bg-slate-500/20 text-slate-400 border-slate-500/30'
    }
  }

  const getStatusIcon = (status: RequestStatus) => {
    switch (status) {
      case 'PENDING':
        return <Clock size={14} />
      case 'APPROVED':
        return <CheckCircle size={14} />
      case 'DECLINED':
        return <XCircle size={14} />
      case 'ARCHIVED':
        return <Archive size={14} />
      default:
        return <FileText size={14} />
    }
  }

  const getTypeColor = (type: RequestType) => {
    switch (type) {
      case 'VACATION':
        return 'bg-blue-500/20 text-blue-400'
      case 'PAID_LEAVE':
        return 'bg-emerald-500/20 text-emerald-400'
      case 'UNPAID_LEAVE':
        return 'bg-amber-500/20 text-amber-400'
      case 'SICK_LEAVE':
        return 'bg-red-500/20 text-red-400'
      default:
        return 'bg-purple-500/20 text-purple-400'
    }
  }

  // Filter requests
  const filteredRequests = requests.filter(request => {
    if (statusFilter !== 'ALL' && request.status !== statusFilter) return false
    if (typeFilter !== 'ALL' && request.request_type !== typeFilter) return false
    return true
  })

  // Statistics
  const stats = {
    total: requests.length,
    pending: requests.filter(r => r.status === 'PENDING').length,
    approved: requests.filter(r => r.status === 'APPROVED').length,
    declined: requests.filter(r => r.status === 'DECLINED').length,
  }

  if (loading) {
    return (
      <PortalLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Cargando solicitudes...</p>
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
            <h1 className="text-3xl font-bold text-white">Mis Solicitudes</h1>
            <p className="text-slate-400 mt-1">
              Historial de solicitudes de ausencia
            </p>
          </div>
          <div className="flex gap-3">
            <Button
              variant="outline"
              onClick={loadRequests}
              className="border-slate-600 text-slate-300 hover:bg-slate-800"
            >
              <RefreshCw size={18} className="mr-2" />
              Actualizar
            </Button>
            <Button
              onClick={() => router.push("/requests/new")}
              className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-lg shadow-blue-500/20"
            >
              <Plus size={18} className="mr-2" />
              Nueva Solicitud
            </Button>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-slate-800/50 rounded-xl p-4 border border-slate-700/50">
            <p className="text-slate-400 text-sm">Total</p>
            <p className="text-2xl font-bold text-white mt-1">{stats.total}</p>
          </div>
          <div className="bg-slate-800/50 rounded-xl p-4 border border-amber-500/30">
            <p className="text-amber-400 text-sm">Pendientes</p>
            <p className="text-2xl font-bold text-white mt-1">{stats.pending}</p>
          </div>
          <div className="bg-slate-800/50 rounded-xl p-4 border border-emerald-500/30">
            <p className="text-emerald-400 text-sm">Aprobadas</p>
            <p className="text-2xl font-bold text-white mt-1">{stats.approved}</p>
          </div>
          <div className="bg-slate-800/50 rounded-xl p-4 border border-red-500/30">
            <p className="text-red-400 text-sm">Rechazadas</p>
            <p className="text-2xl font-bold text-white mt-1">{stats.declined}</p>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex items-center gap-2">
              <Filter size={18} className="text-slate-400" />
              <span className="text-slate-300 font-medium">Filtros:</span>
            </div>
            <div className="flex flex-wrap gap-4">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-40 bg-slate-800/50 border-slate-700 text-white">
                  <SelectValue placeholder="Estado" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ALL">Todos los estados</SelectItem>
                  <SelectItem value="PENDING">Pendiente</SelectItem>
                  <SelectItem value="APPROVED">Aprobada</SelectItem>
                  <SelectItem value="DECLINED">Rechazada</SelectItem>
                  <SelectItem value="ARCHIVED">Archivada</SelectItem>
                </SelectContent>
              </Select>

              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-48 bg-slate-800/50 border-slate-700 text-white">
                  <SelectValue placeholder="Tipo" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ALL">Todos los tipos</SelectItem>
                  <SelectItem value="VACATION">Vacaciones</SelectItem>
                  <SelectItem value="PAID_LEAVE">Permiso con Goce</SelectItem>
                  <SelectItem value="UNPAID_LEAVE">Permiso sin Goce</SelectItem>
                  <SelectItem value="LATE_ENTRY">Pase de Entrada</SelectItem>
                  <SelectItem value="EARLY_EXIT">Pase de Salida</SelectItem>
                  <SelectItem value="SHIFT_CHANGE">Cambio de Turno</SelectItem>
                  <SelectItem value="SICK_LEAVE">Incapacidad</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Requests List */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
          {filteredRequests.length === 0 ? (
            <div className="p-12 text-center">
              <FileText size={48} className="mx-auto text-slate-600 mb-4" />
              <p className="text-slate-400">No hay solicitudes</p>
              <p className="text-sm text-slate-500 mt-2">
                {requests.length > 0 ? "Intenta con otros filtros" : "Crea tu primera solicitud de ausencia"}
              </p>
              {requests.length === 0 && (
                <Button
                  onClick={() => router.push("/requests/new")}
                  className="mt-4 bg-blue-600 hover:bg-blue-700"
                >
                  Nueva Solicitud
                </Button>
              )}
            </div>
          ) : (
            <div className="divide-y divide-slate-700/50">
              {filteredRequests.map((request) => (
                <div
                  key={request.id}
                  className="p-6 hover:bg-slate-800/30 transition-colors"
                >
                  <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div className="flex items-start gap-4">
                      <div className={`p-3 rounded-xl ${getTypeColor(request.request_type)}`}>
                        <FileText size={24} />
                      </div>
                      <div>
                        <h4 className="text-white font-semibold">
                          {REQUEST_TYPE_LABELS[request.request_type] || request.request_type}
                        </h4>
                        <p className="text-slate-400 text-sm mt-1">
                          {formatDate(request.start_date)} - {formatDate(request.end_date)}
                          <span className="text-slate-500 ml-2">({request.total_days} dias)</span>
                        </p>
                        <p className="text-slate-500 text-sm mt-2 line-clamp-2">
                          {request.reason}
                        </p>
                      </div>
                    </div>

                    <div className="flex flex-col items-end gap-3">
                      <div className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-full border ${getStatusColor(request.status)}`}>
                        {getStatusIcon(request.status)}
                        {REQUEST_STATUS_LABELS[request.status] || request.status}
                      </div>

                      {request.status === 'PENDING' && (
                        <p className="text-xs text-slate-500">
                          En: {APPROVAL_STAGE_LABELS[request.current_approval_stage] || request.current_approval_stage}
                        </p>
                      )}

                      <div className="flex gap-2">
                        {request.status === 'PENDING' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(request.id)}
                            className="text-red-400 hover:text-red-300 hover:bg-red-500/10"
                          >
                            <Trash2 size={16} className="mr-1" />
                            Eliminar
                          </Button>
                        )}
                        {(request.status === 'APPROVED' || request.status === 'DECLINED') && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleArchive(request.id)}
                            className="text-slate-400 hover:text-slate-300 hover:bg-slate-700/50"
                          >
                            <Archive size={16} className="mr-1" />
                            Archivar
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Approval History */}
                  {request.approval_history && request.approval_history.length > 0 && (
                    <div className="mt-4 pt-4 border-t border-slate-700/50">
                      <p className="text-slate-400 text-sm font-medium mb-2">Historial de Aprobacion:</p>
                      <div className="flex flex-wrap gap-2">
                        {request.approval_history.map((history, idx) => (
                          <span
                            key={idx}
                            className={`inline-flex items-center gap-1 px-2 py-1 text-xs rounded-full ${
                              history.action === 'APPROVED'
                                ? 'bg-emerald-500/20 text-emerald-400'
                                : 'bg-red-500/20 text-red-400'
                            }`}
                          >
                            {history.action === 'APPROVED' ? <CheckCircle size={12} /> : <XCircle size={12} />}
                            {APPROVAL_STAGE_LABELS[history.approval_stage]}
                            {history.approver && ` - ${history.approver.full_name}`}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </PortalLayout>
  )
}
