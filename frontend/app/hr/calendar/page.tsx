"use client"

/**
 * HR Calendar Page
 *
 * Main page for Human Resources to view and manage employee schedules,
 * absences, incidences, and shift changes on a unified calendar.
 *
 * Features:
 * - Month, Week and Day view toggle
 * - Filter by collar type (white/blue/gray)
 * - Filter by specific employees
 * - Color-coded events per employee
 * - Quick actions to create absences, incidences, shift changes
 * - Click events to view/edit details
 * - Date picker for quick navigation
 * - Export calendar to CSV
 * - Quick action context menu on date click
 */

import { useState, useEffect, useMemo, useCallback, useRef } from "react"
import { useRouter } from "next/navigation"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { HRCalendar } from "@/components/calendar/hr-calendar"
import { CalendarSidebar, CollarType } from "@/components/calendar/calendar-sidebar"
import { CalendarToolbar } from "@/components/calendar/calendar-toolbar"
import { AbsenceDialog } from "@/components/calendar/event-dialogs/absence-dialog"
import { IncidenceDialog } from "@/components/calendar/event-dialogs/incidence-dialog"
import { ShiftDialog } from "@/components/calendar/event-dialogs/shift-dialog"
import { Button } from "@/components/ui/button"
import {
  calendarApi,
  CalendarEvent,
  CalendarEmployee,
  CalendarSummary,
  CalendarEventType
} from "@/lib/api-client"
import { isAuthenticated } from "@/lib/auth"
import {
  format,
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  startOfDay,
  endOfDay,
  addMonths,
  subMonths,
  addWeeks,
  subWeeks,
  addDays,
  subDays
} from "date-fns"
import { toast } from "sonner"
import { Plus, Calendar, ClipboardList, X } from "lucide-react"

type ViewMode = 'month' | 'week' | 'day'

export default function HRCalendarPage() {
  const router = useRouter()

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>('month')
  const [currentDate, setCurrentDate] = useState(new Date())

  // Filter state
  const [selectedCollarTypes, setSelectedCollarTypes] = useState<CollarType[]>([])
  const [selectedEmployeeIds, setSelectedEmployeeIds] = useState<string[]>([])
  const [selectedEventTypes, setSelectedEventTypes] = useState<CalendarEventType[]>([])
  const [statusFilter, setStatusFilter] = useState<string>('')

  // Data state
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [employees, setEmployees] = useState<CalendarEmployee[]>([])
  const [summary, setSummary] = useState<CalendarSummary | null>(null)
  const [loading, setLoading] = useState(true)

  // Dialog state
  const [absenceDialogOpen, setAbsenceDialogOpen] = useState(false)
  const [incidenceDialogOpen, setIncidenceDialogOpen] = useState(false)
  const [shiftDialogOpen, setShiftDialogOpen] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null)
  const [selectedDate, setSelectedDate] = useState<Date | null>(null)

  // Quick action menu state
  const [quickActionMenu, setQuickActionMenu] = useState<{ x: number; y: number; date: Date } | null>(null)
  const quickActionRef = useRef<HTMLDivElement>(null)

  // Check authentication
  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
    }
  }, [router])

  // Close quick action menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (quickActionRef.current && !quickActionRef.current.contains(event.target as Node)) {
        setQuickActionMenu(null)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't handle if user is typing in an input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return
      }

      switch (e.key) {
        case 'ArrowLeft':
          handleNavigate('prev')
          break
        case 'ArrowRight':
          handleNavigate('next')
          break
        case 't':
        case 'T':
          handleNavigate('today')
          break
        case 'm':
        case 'M':
          setViewMode('month')
          break
        case 'w':
        case 'W':
          setViewMode('week')
          break
        case 'd':
        case 'D':
          setViewMode('day')
          break
        case 'Escape':
          setQuickActionMenu(null)
          break
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [viewMode])

  // Calculate date range based on view mode
  const dateRange = useMemo(() => {
    if (viewMode === 'month') {
      const start = startOfWeek(startOfMonth(currentDate), { weekStartsOn: 1 })
      const end = endOfWeek(endOfMonth(currentDate), { weekStartsOn: 1 })
      return { start, end }
    } else if (viewMode === 'week') {
      const start = startOfWeek(currentDate, { weekStartsOn: 1 })
      const end = endOfWeek(currentDate, { weekStartsOn: 1 })
      return { start, end }
    } else {
      // Day view
      const start = startOfDay(currentDate)
      const end = endOfDay(currentDate)
      return { start, end }
    }
  }, [currentDate, viewMode])

  // Employee color map for rendering
  const employeeColorMap = useMemo(() => {
    const map = new Map<string, string>()
    employees.forEach(emp => map.set(emp.id, emp.color))
    return map
  }, [employees])

  // Load calendar data
  const loadCalendarData = useCallback(async () => {
    try {
      setLoading(true)
      const response = await calendarApi.getEvents({
        start_date: format(dateRange.start, 'yyyy-MM-dd'),
        end_date: format(dateRange.end, 'yyyy-MM-dd'),
        employee_ids: selectedEmployeeIds.length > 0 ? selectedEmployeeIds : undefined,
        collar_types: selectedCollarTypes.length > 0 ? selectedCollarTypes : undefined,
        event_types: selectedEventTypes.length > 0 ? selectedEventTypes : undefined,
        status: statusFilter || undefined,
      })
      setEvents(response.events || [])
      setSummary(response.summary)
    } catch (error) {
      console.error('Error loading calendar data:', error)
      toast.error('Error loading calendar events')
    } finally {
      setLoading(false)
    }
  }, [dateRange, selectedCollarTypes, selectedEmployeeIds, selectedEventTypes, statusFilter])

  // Load employees
  const loadEmployees = useCallback(async () => {
    try {
      const response = await calendarApi.getEmployees(
        selectedCollarTypes.length > 0 ? selectedCollarTypes : undefined
      )
      setEmployees(response.employees || [])
    } catch (error) {
      console.error('Error loading employees:', error)
      toast.error('Error loading employees')
    }
  }, [selectedCollarTypes])

  // Load data on mount and when filters change
  useEffect(() => {
    loadCalendarData()
  }, [loadCalendarData])

  useEffect(() => {
    loadEmployees()
  }, [loadEmployees])

  // Navigation handlers
  const handleNavigate = (direction: 'prev' | 'next' | 'today') => {
    if (direction === 'today') {
      setCurrentDate(new Date())
    } else if (viewMode === 'month') {
      setCurrentDate(prev => direction === 'prev' ? subMonths(prev, 1) : addMonths(prev, 1))
    } else if (viewMode === 'week') {
      setCurrentDate(prev => direction === 'prev' ? subWeeks(prev, 1) : addWeeks(prev, 1))
    } else {
      // Day view
      setCurrentDate(prev => direction === 'prev' ? subDays(prev, 1) : addDays(prev, 1))
    }
  }

  // Date selection from date picker
  const handleDateSelect = (date: Date) => {
    setCurrentDate(date)
  }

  // Event handlers
  const handleEventClick = (event: CalendarEvent) => {
    setSelectedEvent(event)
    setQuickActionMenu(null)
    if (event.event_type === 'absence') {
      setAbsenceDialogOpen(true)
    } else if (event.event_type === 'incidence') {
      setIncidenceDialogOpen(true)
    } else if (event.event_type === 'shift_change') {
      setShiftDialogOpen(true)
    }
  }

  const handleDateClick = (date: Date) => {
    setSelectedDate(date)
    // For day view, don't show quick action menu (user can use sidebar)
    if (viewMode !== 'day') {
      // Get mouse position for context menu
      const event = window.event as MouseEvent
      if (event) {
        setQuickActionMenu({
          x: event.clientX,
          y: event.clientY,
          date
        })
      }
    }
  }

  const handleCreateAbsence = (date?: Date) => {
    setSelectedEvent(null)
    setSelectedDate(date || null)
    setQuickActionMenu(null)
    setAbsenceDialogOpen(true)
  }

  const handleCreateIncidence = (date?: Date) => {
    setSelectedEvent(null)
    setSelectedDate(date || null)
    setQuickActionMenu(null)
    setIncidenceDialogOpen(true)
  }

  const handleDialogClose = () => {
    setAbsenceDialogOpen(false)
    setIncidenceDialogOpen(false)
    setShiftDialogOpen(false)
    setSelectedEvent(null)
    setSelectedDate(null)
  }

  const handleEventSaved = () => {
    handleDialogClose()
    loadCalendarData()
    toast.success('Event saved successfully')
  }

  // Export calendar to CSV
  const handleExport = () => {
    if (events.length === 0) {
      toast.error('No events to export')
      return
    }

    const headers = ['Employee', 'Employee Number', 'Event Type', 'Status', 'Start Date', 'End Date', 'Title', 'Reason']
    const rows = events.map(event => [
      event.employee_name,
      event.employee_number,
      event.event_type,
      event.status,
      event.start_date,
      event.end_date,
      event.title,
      event.reason || ''
    ])

    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `calendar-export-${format(currentDate, 'yyyy-MM')}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    toast.success('Calendar exported successfully')
  }

  return (
    <DashboardLayout>
      <div className="flex h-[calc(100vh-7rem)] overflow-hidden">
        {/* Sidebar with filters */}
        <CalendarSidebar
          employees={employees}
          selectedEmployeeIds={selectedEmployeeIds}
          onEmployeeSelect={setSelectedEmployeeIds}
          selectedCollarTypes={selectedCollarTypes}
          onCollarTypeChange={setSelectedCollarTypes}
          selectedEventTypes={selectedEventTypes}
          onEventTypeChange={setSelectedEventTypes}
          statusFilter={statusFilter}
          onStatusChange={setStatusFilter}
          summary={summary}
          onCreateAbsence={() => handleCreateAbsence()}
          onCreateIncidence={() => handleCreateIncidence()}
        />

        {/* Main calendar area */}
        <div className="flex-1 flex flex-col overflow-hidden min-w-0">
          {/* Toolbar */}
          <CalendarToolbar
            currentDate={currentDate}
            viewMode={viewMode}
            onViewModeChange={setViewMode}
            onNavigate={handleNavigate}
            onDateSelect={handleDateSelect}
            onRefresh={loadCalendarData}
            onExport={handleExport}
            employees={employees}
            selectedEmployeeIds={selectedEmployeeIds}
            onEmployeeSelect={setSelectedEmployeeIds}
          />

          {/* Calendar grid */}
          <div className="flex-1 overflow-y-auto overflow-x-hidden p-4">
            <HRCalendar
              viewMode={viewMode}
              currentDate={currentDate}
              events={events}
              employeeColorMap={employeeColorMap}
              onEventClick={handleEventClick}
              onDateClick={handleDateClick}
              loading={loading}
            />
          </div>

          {/* Keyboard shortcuts hint */}
          <div className="px-4 pb-2 text-xs text-slate-500 flex items-center gap-4">
            <span>Keyboard shortcuts:</span>
            <span><kbd className="px-1 bg-slate-800 rounded">←</kbd> <kbd className="px-1 bg-slate-800 rounded">→</kbd> Navigate</span>
            <span><kbd className="px-1 bg-slate-800 rounded">T</kbd> Today</span>
            <span><kbd className="px-1 bg-slate-800 rounded">M</kbd> Month</span>
            <span><kbd className="px-1 bg-slate-800 rounded">W</kbd> Week</span>
            <span><kbd className="px-1 bg-slate-800 rounded">D</kbd> Day</span>
          </div>
        </div>
      </div>

      {/* Quick Action Context Menu */}
      {quickActionMenu && (
        <div
          ref={quickActionRef}
          className="fixed bg-slate-800 border border-slate-700 rounded-lg shadow-xl z-50 py-2 min-w-48"
          style={{
            left: Math.min(quickActionMenu.x, window.innerWidth - 200),
            top: Math.min(quickActionMenu.y, window.innerHeight - 150)
          }}
        >
          <div className="px-3 py-2 border-b border-slate-700">
            <div className="text-sm font-medium text-white">
              {format(quickActionMenu.date, 'EEEE, MMM d')}
            </div>
          </div>
          <div className="py-1">
            <button
              onClick={() => handleCreateAbsence(quickActionMenu.date)}
              className="w-full flex items-center gap-2 px-3 py-2 text-sm text-slate-300 hover:bg-slate-700 hover:text-white"
            >
              <Calendar className="h-4 w-4" />
              New Absence
            </button>
            <button
              onClick={() => handleCreateIncidence(quickActionMenu.date)}
              className="w-full flex items-center gap-2 px-3 py-2 text-sm text-slate-300 hover:bg-slate-700 hover:text-white"
            >
              <ClipboardList className="h-4 w-4" />
              New Incidence
            </button>
          </div>
          <div className="border-t border-slate-700 pt-1">
            <button
              onClick={() => {
                setCurrentDate(quickActionMenu.date)
                setViewMode('day')
                setQuickActionMenu(null)
              }}
              className="w-full flex items-center gap-2 px-3 py-2 text-sm text-slate-300 hover:bg-slate-700 hover:text-white"
            >
              View Day
            </button>
          </div>
        </div>
      )}

      {/* Dialogs */}
      <AbsenceDialog
        open={absenceDialogOpen}
        onOpenChange={setAbsenceDialogOpen}
        event={selectedEvent?.event_type === 'absence' ? selectedEvent : null}
        employees={employees}
        selectedDate={selectedDate}
        onSave={handleEventSaved}
        onClose={handleDialogClose}
      />

      <IncidenceDialog
        open={incidenceDialogOpen}
        onOpenChange={setIncidenceDialogOpen}
        event={selectedEvent?.event_type === 'incidence' ? selectedEvent : null}
        employees={employees}
        selectedDate={selectedDate}
        onSave={handleEventSaved}
        onClose={handleDialogClose}
      />

      <ShiftDialog
        open={shiftDialogOpen}
        onOpenChange={setShiftDialogOpen}
        event={selectedEvent?.event_type === 'shift_change' ? selectedEvent : null}
        employees={employees}
        selectedDate={selectedDate}
        onSave={handleEventSaved}
        onClose={handleDialogClose}
      />
    </DashboardLayout>
  )
}
