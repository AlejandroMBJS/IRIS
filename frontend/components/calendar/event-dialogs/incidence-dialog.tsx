"use client"

/**
 * Incidence Dialog Component
 *
 * Dialog for viewing incidence details or creating new incidences from the calendar.
 * Shows incidence type, dates, quantity, and calculated amount for existing events.
 * Shows a creation form when no event is provided.
 */

import { useState, useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { CalendarEvent, CalendarEmployee, incidenceApi, incidenceTypeApi, IncidenceType } from "@/lib/api-client"
import { format, parseISO } from "date-fns"
import {
  ClipboardList,
  User,
  Calendar,
  DollarSign,
  Hash,
  FileText,
  CheckCircle,
  XCircle,
  AlertCircle,
  TrendingUp,
  TrendingDown,
  Minus,
  Loader2,
  Plus
} from "lucide-react"
import { toast } from "sonner"

interface IncidenceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  event: CalendarEvent | null
  employees: CalendarEmployee[]
  selectedDate: Date | null
  onSave: () => void
  onClose: () => void
}

// Category labels
const CATEGORY_LABELS: Record<string, string> = {
  'absence': 'Ausencia',
  'sick': 'Enfermedad',
  'vacation': 'Vacaciones',
  'overtime': 'Horas Extra',
  'delay': 'Retardo',
  'bonus': 'Bono',
  'deduction': 'Deducción',
  'other': 'Otro',
}

// Status badge styling
const getStatusBadge = (status: string) => {
  const s = status.toLowerCase()
  if (s === 'pending') {
    return (
      <Badge variant="outline" className="bg-yellow-500/20 text-yellow-400 border-yellow-500">
        <AlertCircle className="h-3 w-3 mr-1" />
        Pendiente
      </Badge>
    )
  }
  if (s === 'approved') {
    return (
      <Badge variant="outline" className="bg-green-500/20 text-green-400 border-green-500">
        <CheckCircle className="h-3 w-3 mr-1" />
        Aprobado
      </Badge>
    )
  }
  if (s === 'rejected') {
    return (
      <Badge variant="outline" className="bg-red-500/20 text-red-400 border-red-500">
        <XCircle className="h-3 w-3 mr-1" />
        Rechazado
      </Badge>
    )
  }
  if (s === 'processed') {
    return (
      <Badge variant="outline" className="bg-blue-500/20 text-blue-400 border-blue-500">
        <CheckCircle className="h-3 w-3 mr-1" />
        Procesado
      </Badge>
    )
  }
  return (
    <Badge variant="outline" className="bg-slate-500/20 text-slate-400 border-slate-500">
      {status}
    </Badge>
  )
}

// Effect type icon
const getEffectIcon = (effectType: string) => {
  switch (effectType) {
    case 'positive':
      return <TrendingUp className="h-4 w-4 text-green-400" />
    case 'negative':
      return <TrendingDown className="h-4 w-4 text-red-400" />
    default:
      return <Minus className="h-4 w-4 text-slate-400" />
  }
}

// Effect type label
const EFFECT_TYPE_LABELS: Record<string, string> = {
  'positive': 'Suma al salario',
  'negative': 'Descuenta del salario',
  'neutral': 'Sin efecto',
}

export function IncidenceDialog({
  open,
  onOpenChange,
  event,
  employees,
  selectedDate,
  onSave,
  onClose
}: IncidenceDialogProps) {
  // State for create mode
  const [isCreating, setIsCreating] = useState(false)
  const [incidenceTypes, setIncidenceTypes] = useState<IncidenceType[]>([])
  const [loadingTypes, setLoadingTypes] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  // Form state
  const [selectedEmployeeId, setSelectedEmployeeId] = useState<string>('')
  const [selectedTypeId, setSelectedTypeId] = useState<string>('')
  const [startDate, setStartDate] = useState<string>('')
  const [endDate, setEndDate] = useState<string>('')
  const [quantity, setQuantity] = useState<number>(1)
  const [comments, setComments] = useState<string>('')

  // Selected type details for showing info
  const selectedType = incidenceTypes.find(t => t.id === selectedTypeId)

  // Load incidence types when in create mode
  useEffect(() => {
    if (open && !event) {
      setIsCreating(true)
      loadIncidenceTypes()
      // Set initial date from selectedDate
      if (selectedDate) {
        const dateStr = format(selectedDate, 'yyyy-MM-dd')
        setStartDate(dateStr)
        setEndDate(dateStr)
      }
    } else {
      setIsCreating(false)
    }
  }, [open, event, selectedDate])

  // Reset form when dialog closes
  useEffect(() => {
    if (!open) {
      setSelectedEmployeeId('')
      setSelectedTypeId('')
      setStartDate('')
      setEndDate('')
      setQuantity(1)
      setComments('')
    }
  }, [open])

  const loadIncidenceTypes = async () => {
    try {
      setLoadingTypes(true)
      const types = await incidenceTypeApi.getRequestable()
      setIncidenceTypes(types)
    } catch (error) {
      console.error('Error loading incidence types:', error)
      toast.error('Error loading incidence types')
    } finally {
      setLoadingTypes(false)
    }
  }

  const handleSubmit = async () => {
    // Validation
    if (!selectedEmployeeId) {
      toast.error('Por favor seleccione un empleado')
      return
    }
    if (!selectedTypeId) {
      toast.error('Por favor seleccione un tipo de incidencia')
      return
    }
    if (!startDate || !endDate) {
      toast.error('Por favor seleccione las fechas')
      return
    }
    if (new Date(endDate) < new Date(startDate)) {
      toast.error('La fecha fin debe ser posterior a la fecha inicio')
      return
    }

    try {
      setSubmitting(true)
      await incidenceApi.create({
        employee_id: selectedEmployeeId,
        incidence_type_id: selectedTypeId,
        start_date: startDate,
        end_date: endDate,
        quantity: quantity,
        comments: comments || undefined,
      })
      toast.success('Incidencia creada exitosamente')
      onSave()
    } catch (error) {
      console.error('Error creating incidence:', error)
      toast.error('Error al crear la incidencia')
    } finally {
      setSubmitting(false)
    }
  }

  // Create mode - show form
  if (isCreating) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="bg-slate-800 border-slate-700 text-white max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-xl">
              <Plus className="h-5 w-5 text-purple-400" />
              Nueva Incidencia
            </DialogTitle>
          </DialogHeader>

          {loadingTypes ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-purple-500" />
            </div>
          ) : (
            <div className="space-y-4 mt-4">
              {/* Employee Selection */}
              <div className="space-y-2">
                <Label className="text-slate-300">Empleado</Label>
                <Select value={selectedEmployeeId} onValueChange={setSelectedEmployeeId}>
                  <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                    <SelectValue placeholder="Seleccionar empleado..." />
                  </SelectTrigger>
                  <SelectContent className="bg-slate-800 border-slate-700">
                    {employees.map(emp => (
                      <SelectItem
                        key={emp.id}
                        value={emp.id}
                        className="text-white hover:bg-slate-700"
                      >
                        {emp.full_name} ({emp.employee_number})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Incidence Type Selection */}
              <div className="space-y-2">
                <Label className="text-slate-300">Tipo de Incidencia</Label>
                <Select value={selectedTypeId} onValueChange={setSelectedTypeId}>
                  <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                    <SelectValue placeholder="Seleccionar tipo..." />
                  </SelectTrigger>
                  <SelectContent className="bg-slate-800 border-slate-700 max-h-60">
                    {incidenceTypes.map(type => (
                      <SelectItem
                        key={type.id}
                        value={type.id}
                        className="text-white hover:bg-slate-700"
                      >
                        <div className="flex items-center gap-2">
                          {type.effect_type === 'positive' && <TrendingUp className="h-3 w-3 text-green-400" />}
                          {type.effect_type === 'negative' && <TrendingDown className="h-3 w-3 text-red-400" />}
                          {type.effect_type === 'neutral' && <Minus className="h-3 w-3 text-slate-400" />}
                          {type.name}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Selected Type Info */}
              {selectedType && (
                <div className="bg-slate-900/50 rounded-lg p-3 border border-slate-700">
                  <div className="flex items-center gap-2 text-sm">
                    {getEffectIcon(selectedType.effect_type || 'neutral')}
                    <span className="text-slate-300">
                      {EFFECT_TYPE_LABELS[selectedType.effect_type || 'neutral']}
                    </span>
                    <Badge variant="outline" className="border-slate-600 ml-auto">
                      {CATEGORY_LABELS[selectedType.category || ''] || selectedType.category}
                    </Badge>
                  </div>
                  {selectedType.description && (
                    <p className="text-xs text-slate-400 mt-2">{selectedType.description}</p>
                  )}
                </div>
              )}

              {/* Date Range */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label className="text-slate-300">Fecha Inicio</Label>
                  <input
                    type="date"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-slate-300">Fecha Fin</Label>
                  <input
                    type="date"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* Quantity */}
              <div className="space-y-2">
                <Label className="text-slate-300">Cantidad</Label>
                <input
                  type="number"
                  min="0.5"
                  step="0.5"
                  value={quantity}
                  onChange={(e) => setQuantity(parseFloat(e.target.value) || 1)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>

              {/* Comments */}
              <div className="space-y-2">
                <Label className="text-slate-300">Comentarios (opcional)</Label>
                <Textarea
                  value={comments}
                  onChange={(e) => setComments(e.target.value)}
                  placeholder="Detalles adicionales de la incidencia..."
                  className="bg-slate-900 border-slate-700 text-white placeholder-slate-500 min-h-20"
                />
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-2 mt-6">
                <Button
                  variant="outline"
                  onClick={onClose}
                  className="border-slate-600 hover:bg-slate-700"
                  disabled={submitting}
                >
                  Cancelar
                </Button>
                <Button
                  onClick={handleSubmit}
                  className="bg-purple-600 hover:bg-purple-700"
                  disabled={submitting}
                >
                  {submitting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Guardando...
                    </>
                  ) : (
                    <>
                      <Plus className="h-4 w-4 mr-2" />
                      Crear Incidencia
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    )
  }

  // View mode - show existing event details
  if (!event) {
    return null
  }

  const categoryLabel = CATEGORY_LABELS[event.category || ''] || event.category
  const effectLabel = EFFECT_TYPE_LABELS[event.effect_type || ''] || event.effect_type

  // Format currency
  const formatCurrency = (amount: number | undefined) => {
    if (amount === undefined || amount === null) return '-'
    return new Intl.NumberFormat('es-MX', {
      style: 'currency',
      currency: 'MXN'
    }).format(amount)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <ClipboardList className="h-5 w-5 text-purple-400" />
            Detalles de Incidencia
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 mt-4">
          {/* Status Badge */}
          <div className="flex items-center justify-between">
            {getStatusBadge(event.status)}
          </div>

          {/* Incidence Type & Category */}
          <div className="bg-slate-900/50 rounded-lg p-4">
            <div className="text-sm text-slate-400 mb-1">Tipo de Incidencia</div>
            <div className="text-lg font-semibold text-white">
              {event.incidence_type || categoryLabel}
            </div>
            <div className="flex items-center gap-2 mt-2">
              <Badge variant="outline" className="border-slate-600">
                {categoryLabel}
              </Badge>
              {event.effect_type && (
                <div className="flex items-center gap-1 text-sm text-slate-400">
                  {getEffectIcon(event.effect_type)}
                  {effectLabel}
                </div>
              )}
            </div>
          </div>

          {/* Employee Info */}
          <div className="flex items-center gap-3 bg-slate-900/50 rounded-lg p-4">
            <User className="h-5 w-5 text-slate-400" />
            <div>
              <div className="text-sm text-slate-400">Empleado</div>
              <div className="font-medium text-white">{event.employee_name}</div>
              <div className="text-xs text-slate-500">{event.employee_number}</div>
            </div>
          </div>

          {/* Date Range */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-slate-400 mb-1">
                <Calendar className="h-4 w-4" />
                Fecha Inicio
              </div>
              <div className="font-medium text-white">
                {format(parseISO(event.start_date), 'MMM dd, yyyy')}
              </div>
            </div>
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-slate-400 mb-1">
                <Calendar className="h-4 w-4" />
                Fecha Fin
              </div>
              <div className="font-medium text-white">
                {format(parseISO(event.end_date), 'MMM dd, yyyy')}
              </div>
            </div>
          </div>

          {/* Quantity & Amount */}
          <div className="grid grid-cols-2 gap-4">
            {event.quantity !== undefined && (
              <div className="bg-slate-900/50 rounded-lg p-4">
                <div className="flex items-center gap-2 text-sm text-slate-400 mb-1">
                  <Hash className="h-4 w-4" />
                  Cantidad
                </div>
                <div className="text-xl font-semibold text-white">
                  {event.quantity}
                </div>
              </div>
            )}
            {event.calculated_amount !== undefined && (
              <div className="bg-slate-900/50 rounded-lg p-4">
                <div className="flex items-center gap-2 text-sm text-slate-400 mb-1">
                  <DollarSign className="h-4 w-4" />
                  Monto Calculado
                </div>
                <div className={`text-xl font-semibold ${
                  event.effect_type === 'positive' ? 'text-green-400' :
                  event.effect_type === 'negative' ? 'text-red-400' :
                  'text-white'
                }`}>
                  {formatCurrency(event.calculated_amount)}
                </div>
              </div>
            )}
          </div>

          {/* Description/Comments */}
          {event.description && (
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-slate-400 mb-2">
                <FileText className="h-4 w-4" />
                Comentarios
              </div>
              <div className="text-white">{event.description}</div>
            </div>
          )}

          {/* Collar Type Badge */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-slate-400">Tipo de Empleado:</span>
            <Badge variant="outline" className="border-slate-600">
              {event.collar_type === 'white_collar' && 'Administrativo'}
              {event.collar_type === 'blue_collar' && 'Operativo'}
              {event.collar_type === 'gray_collar' && 'Mixto'}
            </Badge>
          </div>
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-2 mt-6">
          <Button
            variant="outline"
            onClick={onClose}
            className="border-slate-600 hover:bg-slate-700"
          >
            Cerrar
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
