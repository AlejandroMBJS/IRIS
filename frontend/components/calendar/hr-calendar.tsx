"use client"

/**
 * HR Calendar Component
 *
 * Displays a month, week, or day view calendar grid with color-coded events.
 * Events are grouped by date and displayed with employee colors.
 */

import { useMemo } from "react"
import { CalendarEvent } from "@/lib/api-client"
import {
  format,
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  startOfDay,
  endOfDay,
  eachDayOfInterval,
  eachHourOfInterval,
  isSameMonth,
  isSameDay,
  isToday,
  parseISO,
  setHours,
  setMinutes,
  setSeconds,
  setMilliseconds
} from "date-fns"
import { cn } from "@/lib/utils"
import { CalendarEventItem } from "./calendar-event"
import { Loader2, Calendar, Clock } from "lucide-react"

interface HRCalendarProps {
  viewMode: 'month' | 'week' | 'day'
  currentDate: Date
  events: CalendarEvent[]
  employeeColorMap: Map<string, string>
  onEventClick: (event: CalendarEvent) => void
  onDateClick: (date: Date) => void
  loading: boolean
}

export function HRCalendar({
  viewMode,
  currentDate,
  events,
  employeeColorMap,
  onEventClick,
  onDateClick,
  loading
}: HRCalendarProps) {

  // Calculate calendar days
  const calendarDays = useMemo(() => {
    if (viewMode === 'month') {
      const monthStart = startOfMonth(currentDate)
      const monthEnd = endOfMonth(currentDate)
      const calStart = startOfWeek(monthStart, { weekStartsOn: 1 })
      const calEnd = endOfWeek(monthEnd, { weekStartsOn: 1 })
      return eachDayOfInterval({ start: calStart, end: calEnd })
    } else if (viewMode === 'week') {
      const weekStart = startOfWeek(currentDate, { weekStartsOn: 1 })
      const weekEnd = endOfWeek(currentDate, { weekStartsOn: 1 })
      return eachDayOfInterval({ start: weekStart, end: weekEnd })
    } else {
      // Day view - just the current day
      return [currentDate]
    }
  }, [currentDate, viewMode])

  // Group events by date (with proper date handling - no mutation)
  const eventsByDate = useMemo(() => {
    const map = new Map<string, CalendarEvent[]>()

    events.forEach(event => {
      const eventStartDate = parseISO(event.start_date)
      const eventEndDate = parseISO(event.end_date)

      // Normalize dates without mutating
      const eventStart = setMilliseconds(setSeconds(setMinutes(setHours(eventStartDate, 0), 0), 0), 0)
      const eventEnd = setMilliseconds(setSeconds(setMinutes(setHours(eventEndDate, 23), 59), 59), 999)

      // Add event to each day it spans
      calendarDays.forEach(calDay => {
        // Create a fresh date for comparison without mutating calDay
        const dayStart = setMilliseconds(setSeconds(setMinutes(setHours(new Date(calDay), 0), 0), 0), 0)

        if (dayStart >= eventStart && dayStart <= eventEnd) {
          const key = format(calDay, 'yyyy-MM-dd')
          const existing = map.get(key) || []
          if (!existing.find(e => e.id === event.id)) {
            map.set(key, [...existing, event])
          }
        }
      })
    })

    return map
  }, [events, calendarDays])

  // Generate hours for day view
  const dayHours = useMemo(() => {
    if (viewMode !== 'day') return []
    const dayStart = startOfDay(currentDate)
    const dayEnd = endOfDay(currentDate)
    return eachHourOfInterval({ start: dayStart, end: dayEnd })
  }, [currentDate, viewMode])

  const weekDays = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="w-12 h-12 text-blue-500 animate-spin" />
      </div>
    )
  }

  // Day view rendering
  if (viewMode === 'day') {
    const dateKey = format(currentDate, 'yyyy-MM-dd')
    const dayEvents = eventsByDate.get(dateKey) || []

    return (
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 overflow-hidden h-full flex flex-col">
        {/* Day header */}
        <div className="bg-slate-900/50 border-b border-slate-700 p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={cn(
                "w-12 h-12 rounded-lg flex items-center justify-center text-xl font-bold",
                isToday(currentDate) ? "bg-blue-600 text-white" : "bg-slate-700 text-slate-300"
              )}>
                {format(currentDate, 'd')}
              </div>
              <div>
                <div className="text-lg font-semibold text-white">
                  {format(currentDate, 'EEEE')}
                </div>
                <div className="text-sm text-slate-400">
                  {format(currentDate, 'MMMM yyyy')}
                </div>
              </div>
            </div>
            <div className="text-sm text-slate-400">
              {dayEvents.length} event{dayEvents.length !== 1 ? 's' : ''}
            </div>
          </div>
        </div>

        {/* Day content - All events */}
        <div className="flex-1 overflow-auto p-4">
          {dayEvents.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-slate-400">
              <Calendar className="w-16 h-16 mb-4 opacity-50" />
              <p className="text-lg font-medium">No events</p>
              <p className="text-sm">Click on a date to add an event</p>
            </div>
          ) : (
            <div className="space-y-3">
              {dayEvents.map(event => (
                <div
                  key={event.id}
                  onClick={() => onEventClick(event)}
                  className={cn(
                    "p-4 rounded-lg cursor-pointer transition-all hover:scale-[1.02] border-l-4",
                    event.status?.toUpperCase() === 'PENDING' && "border-l-yellow-500 bg-yellow-500/10",
                    (event.status?.toUpperCase() === 'APPROVED' || event.status?.toUpperCase() === 'COMPLETED') && "border-l-green-500 bg-green-500/10",
                    (event.status?.toUpperCase() === 'DECLINED' || event.status?.toUpperCase() === 'REJECTED') && "border-l-red-500 bg-red-500/10",
                    !['PENDING', 'APPROVED', 'COMPLETED', 'DECLINED', 'REJECTED'].includes(event.status?.toUpperCase() || '') && "border-l-slate-500 bg-slate-700/30"
                  )}
                >
                  <div className="flex items-start gap-3">
                    <div
                      className="w-3 h-3 rounded-full mt-1.5 flex-shrink-0"
                      style={{ backgroundColor: employeeColorMap.get(event.employee_id) || '#6366f1' }}
                    />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-semibold text-white">{event.employee_name}</span>
                        <span className={cn(
                          "text-xs px-2 py-0.5 rounded-full",
                          event.status?.toUpperCase() === 'PENDING' && "bg-yellow-500/20 text-yellow-400",
                          (event.status?.toUpperCase() === 'APPROVED' || event.status?.toUpperCase() === 'COMPLETED') && "bg-green-500/20 text-green-400",
                          (event.status?.toUpperCase() === 'DECLINED' || event.status?.toUpperCase() === 'REJECTED') && "bg-red-500/20 text-red-400"
                        )}>
                          {event.status}
                        </span>
                      </div>
                      <div className="text-sm text-slate-300 mb-2">{event.title}</div>
                      <div className="flex items-center gap-4 text-xs text-slate-400">
                        <span className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          {format(parseISO(event.start_date), 'MMM d')} - {format(parseISO(event.end_date), 'MMM d')}
                        </span>
                        {event.total_days && (
                          <span>{event.total_days} day{event.total_days !== 1 ? 's' : ''}</span>
                        )}
                        <span className="capitalize">{event.event_type.replace('_', ' ')}</span>
                      </div>
                      {event.reason && (
                        <div className="mt-2 text-sm text-slate-400 line-clamp-2">
                          {event.reason}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    )
  }

  // Calculate number of rows
  const numberOfRows = Math.ceil(calendarDays.length / 7)

  // Month/Week view rendering
  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 h-full flex flex-col overflow-hidden">
      {/* Week day headers */}
      <div className="grid grid-cols-7 bg-slate-900/50 border-b border-slate-700 flex-shrink-0">
        {weekDays.map(day => (
          <div key={day} className="p-2 text-center text-sm font-medium text-slate-400">
            {day}
          </div>
        ))}
      </div>

      {/* Calendar grid */}
      <div
        className="grid grid-cols-7 flex-1 min-h-0"
        style={{ gridTemplateRows: `repeat(${numberOfRows}, minmax(0, 1fr))` }}
      >
        {calendarDays.map((day, idx) => {
          const dateKey = format(day, 'yyyy-MM-dd')
          const dayEvents = eventsByDate.get(dateKey) || []
          const isCurrentMonth = viewMode === 'week' || isSameMonth(day, currentDate)
          const isCurrentDay = isToday(day)
          const maxEventsToShow = viewMode === 'week' ? 10 : 3

          return (
            <div
              key={idx}
              onClick={() => onDateClick(day)}
              className={cn(
                "p-1 border-r border-b border-slate-700 cursor-pointer transition-colors overflow-hidden flex flex-col",
                isCurrentMonth ? "bg-slate-800/30" : "bg-slate-900/50",
                isCurrentDay && "ring-2 ring-inset ring-blue-500",
                "hover:bg-slate-700/30"
              )}
            >
              {/* Day number */}
              <div className={cn(
                "text-sm font-medium mb-1 px-1 flex-shrink-0",
                isCurrentMonth ? "text-white" : "text-slate-600",
                isCurrentDay && "text-blue-400"
              )}>
                {format(day, 'd')}
              </div>

              {/* Events */}
              <div className="space-y-0.5 overflow-hidden flex-1 min-h-0">
                {dayEvents.slice(0, maxEventsToShow).map(event => (
                  <CalendarEventItem
                    key={event.id}
                    event={event}
                    color={employeeColorMap.get(event.employee_id) || '#6366f1'}
                    onClick={() => onEventClick(event)}
                    compact={viewMode === 'month'}
                  />
                ))}
                {dayEvents.length > maxEventsToShow && (
                  <div className="text-xs text-slate-400 px-1 py-0.5">
                    +{dayEvents.length - maxEventsToShow} more
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
