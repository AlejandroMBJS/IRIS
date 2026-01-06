"use client"

import { useState } from "react"
import {
  FileText,
  CheckCircle,
  XCircle,
  Clock,
  User,
  Calendar,
  MessageSquare,
  RefreshCw,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  AbsenceRequest,
  REQUEST_TYPE_LABELS,
  ApprovalStage,
  absenceRequestApi,
} from "@/lib/api-client"
import { useToast } from "@/hooks/use-toast"

interface ApprovalListProps {
  requests: AbsenceRequest[]
  stage: ApprovalStage
  stageName: string
  onRefresh: () => void
  loading: boolean
}

export function ApprovalList({ requests, stage, stageName, onRefresh, loading }: ApprovalListProps) {
  const { toast } = useToast()
  const [selectedRequest, setSelectedRequest] = useState<AbsenceRequest | null>(null)
  const [actionType, setActionType] = useState<'APPROVED' | 'DECLINED' | null>(null)
  const [comments, setComments] = useState('')
  const [processing, setProcessing] = useState(false)

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

  const getTypeColor = (type: string) => {
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

  const handleApprove = (request: AbsenceRequest) => {
    setSelectedRequest(request)
    setActionType('APPROVED')
    setComments('')
  }

  const handleDecline = (request: AbsenceRequest) => {
    setSelectedRequest(request)
    setActionType('DECLINED')
    setComments('')
  }

  const handleSubmitAction = async () => {
    if (!selectedRequest || !actionType) return

    setProcessing(true)
    try {
      await absenceRequestApi.approve(selectedRequest.id, {
        action: actionType,
        stage: stage,
        comments: comments,
      })

      toast({
        title: actionType === 'APPROVED' ? "Solicitud Aprobada" : "Solicitud Rechazada",
        description: `La solicitud ha sido ${actionType === 'APPROVED' ? 'aprobada' : 'rechazada'} correctamente`,
      })

      setSelectedRequest(null)
      setActionType(null)
      setComments('')
      onRefresh()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "No se pudo procesar la solicitud",
        variant: "destructive",
      })
    } finally {
      setProcessing(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
          <p className="text-slate-400">Cargando solicitudes...</p>
        </div>
      </div>
    )
  }

  if (requests.length === 0) {
    return (
      <div className="p-12 text-center">
        <CheckCircle size={48} className="mx-auto text-emerald-500/50 mb-4" />
        <p className="text-slate-400">No hay solicitudes pendientes</p>
        <p className="text-sm text-slate-500 mt-2">
          Todas las solicitudes de {stageName} han sido procesadas
        </p>
        <Button
          variant="ghost"
          onClick={onRefresh}
          className="mt-4 text-slate-400"
        >
          <RefreshCw size={18} className="mr-2" />
          Actualizar
        </Button>
      </div>
    )
  }

  return (
    <>
      <div className="divide-y divide-slate-700/50">
        {requests.map((request) => (
          <div
            key={request.id}
            className="p-6 hover:bg-slate-800/30 transition-colors"
          >
            <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-6">
              {/* Request Info */}
              <div className="flex-1 space-y-4">
                <div className="flex items-start gap-4">
                  <div className={`p-3 rounded-xl ${getTypeColor(request.request_type)}`}>
                    <FileText size={24} />
                  </div>
                  <div className="flex-1">
                    <h4 className="text-white font-semibold text-lg">
                      {REQUEST_TYPE_LABELS[request.request_type] || request.request_type}
                    </h4>
                    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 mt-2 text-sm text-slate-400">
                      <span className="flex items-center gap-1">
                        <User size={14} />
                        {request.employee?.full_name || 'Empleado'}
                      </span>
                      <span className="flex items-center gap-1">
                        <Calendar size={14} />
                        {formatDate(request.start_date)} - {formatDate(request.end_date)}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock size={14} />
                        {request.total_days} dia{request.total_days !== 1 ? 's' : ''}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Reason */}
                <div className="bg-slate-800/50 rounded-xl p-4 ml-16">
                  <p className="text-slate-400 text-sm font-medium mb-1 flex items-center gap-2">
                    <MessageSquare size={14} />
                    Motivo:
                  </p>
                  <p className="text-slate-300">{request.reason}</p>
                </div>

                {/* Additional Details */}
                {(request.hours_per_day || request.paid_days || request.unpaid_days || request.shift_details) && (
                  <div className="ml-16 grid grid-cols-2 md:grid-cols-4 gap-4">
                    {request.hours_per_day && (
                      <div className="bg-slate-800/30 rounded-lg p-3">
                        <p className="text-slate-500 text-xs">Horas/Dia</p>
                        <p className="text-white font-medium">{request.hours_per_day}</p>
                      </div>
                    )}
                    {request.paid_days && (
                      <div className="bg-slate-800/30 rounded-lg p-3">
                        <p className="text-slate-500 text-xs">Dias con Goce</p>
                        <p className="text-white font-medium">{request.paid_days}</p>
                      </div>
                    )}
                    {request.unpaid_days && (
                      <div className="bg-slate-800/30 rounded-lg p-3">
                        <p className="text-slate-500 text-xs">Dias sin Goce</p>
                        <p className="text-white font-medium">{request.unpaid_days}</p>
                      </div>
                    )}
                  </div>
                )}

                {request.shift_details && (
                  <div className="ml-16 bg-slate-800/30 rounded-lg p-3">
                    <p className="text-slate-500 text-xs mb-1">Detalle de Turno</p>
                    <p className="text-slate-300 text-sm">{request.shift_details}</p>
                  </div>
                )}

                {/* Employee Info */}
                {request.employee && (
                  <div className="ml-16 flex items-center gap-2 text-xs text-slate-500">
                    <span>Departamento: {request.employee.department || 'N/A'}</span>
                    <span>-</span>
                    <span>Area: {request.employee.area || 'N/A'}</span>
                  </div>
                )}
              </div>

              {/* Actions */}
              <div className="flex lg:flex-col gap-3 lg:min-w-[140px]">
                <Button
                  onClick={() => handleApprove(request)}
                  className="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white"
                >
                  <CheckCircle size={18} className="mr-2" />
                  Aprobar
                </Button>
                <Button
                  onClick={() => handleDecline(request)}
                  variant="outline"
                  className="flex-1 border-red-500/50 text-red-400 hover:bg-red-500/10"
                >
                  <XCircle size={18} className="mr-2" />
                  Rechazar
                </Button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Approval/Decline Dialog */}
      <Dialog open={!!selectedRequest && !!actionType} onOpenChange={() => { setSelectedRequest(null); setActionType(null); }}>
        <DialogContent className="bg-slate-900 border-slate-700">
          <DialogHeader>
            <DialogTitle className="text-white">
              {actionType === 'APPROVED' ? 'Aprobar Solicitud' : 'Rechazar Solicitud'}
            </DialogTitle>
            <DialogDescription className="text-slate-400">
              {actionType === 'APPROVED'
                ? 'Confirma que deseas aprobar esta solicitud. Se notificara al empleado.'
                : 'Confirma que deseas rechazar esta solicitud. Por favor proporciona un motivo.'}
            </DialogDescription>
          </DialogHeader>

          {selectedRequest && (
            <div className="py-4">
              <div className="bg-slate-800/50 rounded-xl p-4 mb-4">
                <p className="text-white font-medium">
                  {REQUEST_TYPE_LABELS[selectedRequest.request_type]}
                </p>
                <p className="text-slate-400 text-sm mt-1">
                  {selectedRequest.employee?.full_name} - {formatDate(selectedRequest.start_date)} al {formatDate(selectedRequest.end_date)}
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-slate-300 text-sm font-medium">
                  Comentarios {actionType === 'DECLINED' ? '(requerido)' : '(opcional)'}
                </label>
                <Textarea
                  value={comments}
                  onChange={(e) => setComments(e.target.value)}
                  placeholder={actionType === 'APPROVED' ? "Comentarios adicionales..." : "Motivo del rechazo..."}
                  className="bg-slate-800/50 border-slate-700 text-white"
                  rows={3}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="ghost"
              onClick={() => { setSelectedRequest(null); setActionType(null); }}
              disabled={processing}
              className="text-slate-400"
            >
              Cancelar
            </Button>
            <Button
              onClick={handleSubmitAction}
              disabled={processing || (actionType === 'DECLINED' && !comments)}
              className={actionType === 'APPROVED'
                ? "bg-emerald-600 hover:bg-emerald-700"
                : "bg-red-600 hover:bg-red-700"
              }
            >
              {processing ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                  Procesando...
                </>
              ) : (
                <>
                  {actionType === 'APPROVED' ? <CheckCircle size={18} className="mr-2" /> : <XCircle size={18} className="mr-2" />}
                  {actionType === 'APPROVED' ? 'Aprobar' : 'Rechazar'}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
