"use client"

/**
 * Calendar Event Item Component
 *
 * Renders a single event on the calendar with color coding and status indicator.
 */

import { CalendarEvent } from "@/lib/api-client"
import { cn } from "@/lib/utils"

interface CalendarEventItemProps {
  event: CalendarEvent
  color: string
  onClick: () => void
  compact?: boolean
}

// Get status indicator style
const getStatusStyle = (status: string) => {
  const s = status.toUpperCase()
  if (s === 'PENDING') return 'border-l-yellow-500'
  if (s === 'APPROVED' || s === 'COMPLETED') return 'border-l-green-500'
  if (s === 'DECLINED' || s === 'REJECTED') return 'border-l-red-500'
  return 'border-l-slate-500'
}

// Get event type icon/badge
const getEventTypeBadge = (eventType: string) => {
  switch (eventType) {
    case 'absence':
      return 'A'
    case 'incidence':
      return 'I'
    case 'shift_change':
      return 'T'
    default:
      return '?'
  }
}

export function CalendarEventItem({
  event,
  color,
  onClick,
  compact = false
}: CalendarEventItemProps) {

  const statusStyle = getStatusStyle(event.status)

  if (compact) {
    // Compact view for month calendar
    return (
      <div
        onClick={(e) => {
          e.stopPropagation()
          onClick()
        }}
        className={cn(
          "text-xs px-1.5 py-0.5 rounded truncate cursor-pointer transition-opacity hover:opacity-80 border-l-2",
          statusStyle
        )}
        style={{ backgroundColor: `${color}20`, color: color }}
        title={`${event.title}\n${event.status}`}
      >
        <span className="font-medium">{event.employee_name.split(' ')[0]}</span>
      </div>
    )
  }

  // Detailed view for week calendar
  return (
    <div
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
      className={cn(
        "text-xs px-2 py-1 rounded cursor-pointer transition-opacity hover:opacity-80 border-l-2",
        statusStyle
      )}
      style={{ backgroundColor: `${color}20` }}
    >
      <div className="flex items-center gap-1">
        <span
          className="w-2 h-2 rounded-full flex-shrink-0"
          style={{ backgroundColor: color }}
        />
        <span className="font-medium truncate" style={{ color }}>
          {event.employee_name}
        </span>
      </div>
      <div className="text-slate-400 truncate mt-0.5">
        {event.title.split(' - ')[0]}
      </div>
    </div>
  )
}
