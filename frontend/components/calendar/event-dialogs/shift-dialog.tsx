"use client"

/**
 * Shift Dialog Component
 *
 * Dialog for viewing shift change details from the calendar.
 * Shows shift information, employee, and date.
 */

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { CalendarEvent, CalendarEmployee } from "@/lib/api-client"
import { format, parseISO } from "date-fns"
import {
  Calendar,
  User,
  Clock,
  RefreshCw,
  CheckCircle
} from "lucide-react"

interface ShiftDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  event: CalendarEvent | null
  employees: CalendarEmployee[]
  selectedDate: Date | null
  onSave: () => void
  onClose: () => void
}

export function ShiftDialog({
  open,
  onOpenChange,
  event,
  employees,
  selectedDate,
  onSave,
  onClose
}: ShiftDialogProps) {

  if (!event) {
    return null
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <RefreshCw className="h-5 w-5 text-cyan-400" />
            Shift Change Details
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 mt-4">
          {/* Status Badge */}
          <div className="flex items-center justify-between">
            <Badge variant="outline" className="bg-green-500/20 text-green-400 border-green-500">
              <CheckCircle className="h-3 w-3 mr-1" />
              Applied
            </Badge>
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

          {/* Shift Info */}
          <div className="bg-slate-900/50 rounded-lg p-4">
            <div className="text-sm text-slate-400 mb-2">New Shift</div>
            <div className="flex items-center gap-3">
              <div className="p-3 bg-cyan-500/20 rounded-lg">
                <Clock className="h-6 w-6 text-cyan-400" />
              </div>
              <div>
                <div className="text-lg font-semibold text-white">
                  {event.shift_name || 'Assigned Shift'}
                </div>
                {event.shift_code && (
                  <div className="text-sm text-slate-400">
                    Code: {event.shift_code}
                  </div>
                )}
                {event.shift_time && (
                  <div className="text-sm text-cyan-400 mt-1">
                    {event.shift_time}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Date */}
          <div className="bg-slate-900/50 rounded-lg p-4">
            <div className="flex items-center gap-2 text-sm text-slate-400 mb-1">
              <Calendar className="h-4 w-4" />
              Change Date
            </div>
            <div className="font-medium text-white">
              {format(parseISO(event.start_date), 'EEEE, MMMM dd, yyyy')}
            </div>
          </div>

          {/* Description/Comments */}
          {event.description && (
            <div className="bg-slate-900/50 rounded-lg p-4">
              <div className="text-sm text-slate-400 mb-2">Details</div>
              <div className="text-white">{event.description}</div>
            </div>
          )}

          {/* Collar Type Badge */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-slate-400">Employee Type:</span>
            <Badge variant="outline" className="border-slate-600">
              {event.collar_type === 'white_collar' && 'White Collar'}
              {event.collar_type === 'blue_collar' && 'Blue Collar'}
              {event.collar_type === 'gray_collar' && 'Gray Collar'}
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
