"use client"

/**
 * Absence Dialog Component
 *
 * Dialog for viewing absence request details or creating new absences from the calendar.
 * Shows request type, dates, status, and approval history for existing events.
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
  Calendar,
  User,
  Clock,
  FileText,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
  Plus
} from "lucide-react"
import { toast } from "sonner"

interface AbsenceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  event: CalendarEvent | null
  employees: CalendarEmployee[]
  selectedDate: Date | null
  onSave: () => void
  onClose: () => void
}

// Request type labels
const REQUEST_TYPE_LABELS: Record<string, string> = {
  'PAID_LEAVE': 'Paid Leave',
  'UNPAID_LEAVE': 'Unpaid Leave',
  'VACATION': 'Vacation',
  'LATE_ENTRY': 'Late Entry Pass',
  'EARLY_EXIT': 'Early Exit Pass',
  'SHIFT_CHANGE': 'Shift Change',
  'TIME_FOR_TIME': 'Time for Time',
  'SICK_LEAVE': 'Sick Leave',
  'PERSONAL': 'Personal',
  'OTHER': 'Other',
}

// Status badge styling
const getStatusBadge = (status: string) => {
  const s = status.toUpperCase()
  if (s === 'PENDING') {
    return (
      <Badge variant="outline" className="bg-yellow-500/20 text-yellow-400 border-yellow-500">
        <AlertCircle className="h-3 w-3 mr-1" />
        Pending
      </Badge>
    )
  }
  if (s === 'APPROVED' || s === 'COMPLETED') {
    return (
      <Badge variant="outline" className="bg-green-500/20 text-green-400 border-green-500">
        <CheckCircle className="h-3 w-3 mr-1" />
        Approved
      </Badge>
    )
  }
  if (s === 'DECLINED' || s === 'REJECTED') {
    return (
      <Badge variant="outline" className="bg-red-500/20 text-red-400 border-red-500">
        <XCircle className="h-3 w-3 mr-1" />
        Declined
      </Badge>
    )
  }
  return (
    <Badge variant="outline" className="bg-slate-500/20 text-slate-400 border-slate-500">
      {status}
    </Badge>
  )
}

// Approval stage labels
const APPROVAL_STAGE_LABELS: Record<string, string> = {
  'SUPERVISOR': 'Supervisor',
  'MANAGER': 'Manager',
  'HR': 'Human Resources',
  'PAYROLL': 'Payroll',
  'COMPLETED': 'Completed',
}

export function AbsenceDialog({
  open,
  onOpenChange,
  event,
  employees,
  selectedDate,
  onSave,
  onClose
}: AbsenceDialogProps) {
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

  // Load incidence types for absence category when in create mode
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
      // Filter for absence-related types (category = absence, vacation, sick)
      const absenceTypes = types.filter(t =>
        ['absence', 'vacation', 'sick'].includes(t.category?.toLowerCase() || '')
      )
      setIncidenceTypes(absenceTypes.length > 0 ? absenceTypes : types)
    } catch (error) {
      console.error('Error loading incidence types:', error)
      toast.error('Error loading absence types')
    } finally {
      setLoadingTypes(false)
    }
  }

  const handleSubmit = async () => {
    // Validation
    if (!selectedEmployeeId) {
      toast.error('Please select an employee')
      return
    }
    if (!selectedTypeId) {
      toast.error('Please select an absence type')
      return
    }
    if (!startDate || !endDate) {
      toast.error('Please select start and end dates')
      return
    }
    if (new Date(endDate) < new Date(startDate)) {
      toast.error('End date must be after start date')
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
      toast.success('Absence created successfully')
      onSave()
    } catch (error) {
      console.error('Error creating absence:', error)
      toast.error('Error creating absence')
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
              <Plus className="h-5 w-5 text-blue-400" />
              New Absence
            </DialogTitle>
          </DialogHeader>

          {loadingTypes ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
            </div>
          ) : (
            <div className="space-y-4 mt-4">
              {/* Employee Selection */}
              <div className="space-y-2">
                <Label className="text-slate-300">Employee</Label>
                <Select value={selectedEmployeeId} onValueChange={setSelectedEmployeeId}>
                  <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                    <SelectValue placeholder="Select employee..." />
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

              {/* Absence Type Selection */}
              <div className="space-y-2">
                <Label className="text-slate-300">Absence Type</Label>
                <Select value={selectedTypeId} onValueChange={setSelectedTypeId}>
                  <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                    <SelectValue placeholder="Select type..." />
                  </SelectTrigger>
                  <SelectContent className="bg-slate-800 border-slate-700">
                    {incidenceTypes.map(type => (
                      <SelectItem
                        key={type.id}
                        value={type.id}
                        className="text-white hover:bg-slate-700"
                      >
                        {type.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Date Range */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label className="text-slate-300">Start Date</Label>
                  <input
                    type="date"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-slate-300">End Date</Label>
                  <input
                    type="date"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* Quantity */}
              <div className="space-y-2">
                <Label className="text-slate-300">Quantity (days)</Label>
                <input
                  type="number"
                  min="1"
                  value={quantity}
                  onChange={(e) => setQuantity(parseInt(e.target.value) || 1)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* Comments */}
              <div className="space-y-2">
                <Label className="text-slate-300">Comments (optional)</Label>
                <Textarea
                  value={comments}
                  onChange={(e) => setComments(e.target.value)}
                  placeholder="Reason for absence..."
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
                  Cancel
                </Button>
                <Button
                  onClick={handleSubmit}
                  className="bg-blue-600 hover:bg-blue-700"
                  disabled={submitting}
                >
                  {submitting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      <Plus className="h-4 w-4 mr-2" />
                      Create Absence
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

  const requestTypeLabel = REQUEST_TYPE_LABELS[event.request_type || ''] || event.request_type
  const approvalStageLabel = APPROVAL_STAGE_LABELS[event.approval_stage || ''] || event.approval_stage

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <Calendar className="h-5 w-5 text-blue-400" />
            Absence Details
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 mt-4">
          {/* Status Badge */}
          <div className="flex items-center justify-between">
            {getStatusBadge(event.status)}
            {event.approval_stage && event.status.toUpperCase() === 'PENDING' && (
              <span className="text-sm text-slate-400">
                Stage: {approvalStageLabel}
              </span>
            )}
          </div>

          {/* Request Type */}
          <div className="bg-slate-900/50 rounded-lg p-4">
            <div className="text-sm text-slate-400 mb-1">Request Type</div>
            <div className="text-lg font-semibold text-white">
              {requestTypeLabel || event.title}
            </div>
          </div>

          {/* Employee Info */}
          <div className="flex items-center gap-3 bg-slate-900/50 rounded-lg p-4">
            <User className="h-5 w-5 text-slate-400" />
            <div>
              <div className="text-sm text-slate-400">Employee</div>
              <div className="font-medium text-white">{event.employee_name}</div>
              <div className="text-xs text-slate-500">{event.employee_number}</div>
            </div>
          </div>

          {/* Date Range */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="text-sm text-slate-400 mb-1">Start Date</div>
              <div className="font-medium text-white">
                {format(parseISO(event.start_date), 'MMM dd, yyyy')}
              </div>
            </div>
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="text-sm text-slate-400 mb-1">End Date</div>
              <div className="font-medium text-white">
                {format(parseISO(event.end_date), 'MMM dd, yyyy')}
              </div>
            </div>
          </div>

          {/* Total Days */}
          {event.total_days && (
            <div className="flex items-center gap-3 bg-slate-900/50 rounded-lg p-4">
              <Clock className="h-5 w-5 text-slate-400" />
              <div>
                <div className="text-sm text-slate-400">Total Days</div>
                <div className="font-medium text-white">{event.total_days} day{event.total_days !== 1 ? 's' : ''}</div>
              </div>
            </div>
          )}

          {/* Reason */}
          {event.reason && (
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-slate-400 mb-2">
                <FileText className="h-4 w-4" />
                Reason
              </div>
              <div className="text-white">{event.reason}</div>
            </div>
          )}

          {/* Collar Type Badge */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-slate-400">Employee Type:</span>
            <Badge variant="outline" className="border-slate-600">
              {event.collar_type === 'white_collar' && 'Administrative'}
              {event.collar_type === 'blue_collar' && 'Operative'}
              {event.collar_type === 'gray_collar' && 'Mixed'}
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
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
