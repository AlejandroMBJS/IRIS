"use client"

/**
 * Calendar Toolbar Component
 *
 * Navigation controls, view mode toggle, and date picker for the HR calendar.
 */

import { useState, useRef, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { CalendarEmployee } from "@/lib/api-client"
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameMonth, isSameDay, isToday, startOfWeek, endOfWeek } from "date-fns"
import {
  ChevronLeft,
  ChevronRight,
  Calendar as CalendarIcon,
  LayoutGrid,
  List,
  RefreshCw,
  Download,
  CalendarDays,
  Search,
  X,
  Users
} from "lucide-react"
import { cn } from "@/lib/utils"

interface CalendarToolbarProps {
  currentDate: Date
  viewMode: 'month' | 'week' | 'day'
  onViewModeChange: (mode: 'month' | 'week' | 'day') => void
  onNavigate: (direction: 'prev' | 'next' | 'today') => void
  onDateSelect: (date: Date) => void
  onRefresh: () => void
  onExport?: () => void
  employees?: CalendarEmployee[]
  selectedEmployeeIds?: string[]
  onEmployeeSelect?: (ids: string[]) => void
}

export function CalendarToolbar({
  currentDate,
  viewMode,
  onViewModeChange,
  onNavigate,
  onDateSelect,
  onRefresh,
  onExport,
  employees = [],
  selectedEmployeeIds = [],
  onEmployeeSelect
}: CalendarToolbarProps) {
  const [showDatePicker, setShowDatePicker] = useState(false)
  const [pickerMonth, setPickerMonth] = useState(currentDate)
  const [employeeSearch, setEmployeeSearch] = useState("")
  const [showEmployeeDropdown, setShowEmployeeDropdown] = useState(false)
  const datePickerRef = useRef<HTMLDivElement>(null)
  const employeeSearchRef = useRef<HTMLDivElement>(null)

  // Filter employees based on search
  const filteredEmployees = employees.filter(emp => {
    if (!employeeSearch) return true
    const search = employeeSearch.toLowerCase()
    return (
      emp.full_name?.toLowerCase().includes(search) ||
      emp.employee_number?.toLowerCase().includes(search)
    )
  })

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (datePickerRef.current && !datePickerRef.current.contains(event.target as Node)) {
        setShowDatePicker(false)
      }
      if (employeeSearchRef.current && !employeeSearchRef.current.contains(event.target as Node)) {
        setShowEmployeeDropdown(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  // Update picker month when currentDate changes
  useEffect(() => {
    setPickerMonth(currentDate)
  }, [currentDate])

  // Format the title based on view mode
  const getTitle = () => {
    switch (viewMode) {
      case 'month':
        return format(currentDate, 'MMMM yyyy')
      case 'week':
        return `Week of ${format(currentDate, 'MMM d, yyyy')}`
      case 'day':
        return format(currentDate, 'EEEE, MMMM d, yyyy')
      default:
        return format(currentDate, 'MMMM yyyy')
    }
  }

  // Generate mini calendar days
  const miniCalendarDays = () => {
    const monthStart = startOfMonth(pickerMonth)
    const monthEnd = endOfMonth(pickerMonth)
    const calStart = startOfWeek(monthStart, { weekStartsOn: 1 })
    const calEnd = endOfWeek(monthEnd, { weekStartsOn: 1 })
    return eachDayOfInterval({ start: calStart, end: calEnd })
  }

  const handleDateClick = (date: Date) => {
    onDateSelect(date)
    setShowDatePicker(false)
  }

  return (
    <div className="flex items-center justify-between px-4 py-3 bg-slate-800/50 border-b border-slate-700">
      {/* Left: Navigation */}
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onNavigate('prev')}
          className="border-slate-600 hover:bg-slate-700"
        >
          <ChevronLeft className="h-4 w-4" />
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={() => onNavigate('today')}
          className="border-slate-600 hover:bg-slate-700 px-4"
        >
          Today
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={() => onNavigate('next')}
          className="border-slate-600 hover:bg-slate-700"
        >
          <ChevronRight className="h-4 w-4" />
        </Button>

        {/* Date picker trigger */}
        <div className="relative" ref={datePickerRef}>
          <button
            onClick={() => setShowDatePicker(!showDatePicker)}
            className="flex items-center gap-2 ml-4 px-3 py-1.5 rounded-lg hover:bg-slate-700 transition-colors"
          >
            <CalendarIcon className="h-4 w-4 text-slate-400" />
            <h2 className="text-lg font-semibold text-white capitalize">
              {getTitle()}
            </h2>
          </button>

          {/* Mini calendar dropdown */}
          {showDatePicker && (
            <div className="absolute top-full left-0 mt-2 bg-slate-800 border border-slate-700 rounded-lg shadow-xl z-50 p-3 w-72">
              {/* Month navigation */}
              <div className="flex items-center justify-between mb-3">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setPickerMonth(prev => new Date(prev.getFullYear(), prev.getMonth() - 1))}
                  className="text-slate-400 hover:text-white h-8 w-8 p-0"
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm font-medium text-white">
                  {format(pickerMonth, 'MMMM yyyy')}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setPickerMonth(prev => new Date(prev.getFullYear(), prev.getMonth() + 1))}
                  className="text-slate-400 hover:text-white h-8 w-8 p-0"
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>

              {/* Week day headers */}
              <div className="grid grid-cols-7 mb-2">
                {['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'].map(day => (
                  <div key={day} className="text-center text-xs text-slate-500 py-1">
                    {day}
                  </div>
                ))}
              </div>

              {/* Calendar days */}
              <div className="grid grid-cols-7 gap-1">
                {miniCalendarDays().map((day, idx) => {
                  const isCurrentMonth = isSameMonth(day, pickerMonth)
                  const isSelected = isSameDay(day, currentDate)
                  const isTodayDate = isToday(day)

                  return (
                    <button
                      key={idx}
                      onClick={() => handleDateClick(day)}
                      className={cn(
                        "h-8 w-8 rounded-md text-sm transition-colors",
                        isCurrentMonth ? "text-white" : "text-slate-600",
                        isSelected && "bg-blue-600 text-white",
                        isTodayDate && !isSelected && "ring-1 ring-blue-500",
                        !isSelected && "hover:bg-slate-700"
                      )}
                    >
                      {format(day, 'd')}
                    </button>
                  )
                })}
              </div>

              {/* Quick actions */}
              <div className="mt-3 pt-3 border-t border-slate-700 flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    onDateSelect(new Date())
                    setShowDatePicker(false)
                  }}
                  className="flex-1 border-slate-600 text-xs"
                >
                  Go to Today
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Center: Employee Search */}
      {onEmployeeSelect && (
        <div className="relative flex-shrink-0" ref={employeeSearchRef}>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
            <input
              type="text"
              placeholder="Search employees..."
              value={employeeSearch}
              onChange={(e) => {
                setEmployeeSearch(e.target.value)
                setShowEmployeeDropdown(true)
              }}
              onFocus={() => setShowEmployeeDropdown(true)}
              className="w-64 bg-slate-900/50 border border-slate-700 rounded-lg pl-9 pr-8 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500"
            />
            {(employeeSearch || selectedEmployeeIds.length > 0) && (
              <button
                onClick={() => {
                  setEmployeeSearch("")
                  onEmployeeSelect([])
                }}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>

          {/* Selected count badge */}
          {selectedEmployeeIds.length > 0 && (
            <div className="absolute -top-2 -right-2 bg-blue-600 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
              {selectedEmployeeIds.length}
            </div>
          )}

          {/* Dropdown */}
          {showEmployeeDropdown && (
            <div className="absolute top-full left-0 mt-2 w-72 max-h-64 overflow-y-auto bg-slate-800 border border-slate-700 rounded-lg shadow-xl z-50">
              <div className="p-2 border-b border-slate-700 flex items-center justify-between">
                <span className="text-xs text-slate-400">
                  {filteredEmployees.length} employees
                </span>
                {selectedEmployeeIds.length > 0 && (
                  <button
                    onClick={() => onEmployeeSelect([])}
                    className="text-xs text-blue-400 hover:text-blue-300"
                  >
                    Clear all
                  </button>
                )}
              </div>
              {filteredEmployees.length === 0 ? (
                <div className="p-4 text-center text-slate-500 text-sm">
                  No employees found
                </div>
              ) : (
                filteredEmployees.slice(0, 20).map(emp => {
                  const isSelected = selectedEmployeeIds.includes(emp.id)
                  return (
                    <button
                      key={emp.id}
                      onClick={() => {
                        if (isSelected) {
                          onEmployeeSelect(selectedEmployeeIds.filter(id => id !== emp.id))
                        } else {
                          onEmployeeSelect([...selectedEmployeeIds, emp.id])
                        }
                      }}
                      className={cn(
                        "w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-slate-700 transition-colors",
                        isSelected && "bg-blue-600/20"
                      )}
                    >
                      <div
                        className="w-3 h-3 rounded-full flex-shrink-0"
                        style={{ backgroundColor: emp.color || '#6366f1' }}
                      />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm text-white truncate">{emp.full_name}</div>
                        <div className="text-xs text-slate-400">{emp.employee_number}</div>
                      </div>
                      {isSelected && (
                        <div className="w-4 h-4 bg-blue-600 rounded-full flex items-center justify-center">
                          <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                          </svg>
                        </div>
                      )}
                    </button>
                  )
                })
              )}
              {filteredEmployees.length > 20 && (
                <div className="p-2 text-center text-xs text-slate-500 border-t border-slate-700">
                  +{filteredEmployees.length - 20} more - refine your search
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Right: View Mode, Export & Refresh */}
      <div className="flex items-center gap-2">
        {onExport && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onExport}
            className="text-slate-400 hover:text-white hover:bg-slate-700"
            title="Export calendar"
          >
            <Download className="h-4 w-4" />
          </Button>
        )}

        <Button
          variant="ghost"
          size="sm"
          onClick={onRefresh}
          className="text-slate-400 hover:text-white hover:bg-slate-700"
          title="Refresh"
        >
          <RefreshCw className="h-4 w-4" />
        </Button>

        <div className="flex items-center bg-slate-900/50 rounded-lg p-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onViewModeChange('month')}
            className={`px-3 ${
              viewMode === 'month'
                ? 'bg-blue-600 text-white'
                : 'text-slate-400 hover:text-white hover:bg-slate-700'
            }`}
            title="Month view"
          >
            <LayoutGrid className="h-4 w-4 mr-2" />
            Month
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onViewModeChange('week')}
            className={`px-3 ${
              viewMode === 'week'
                ? 'bg-blue-600 text-white'
                : 'text-slate-400 hover:text-white hover:bg-slate-700'
            }`}
            title="Week view"
          >
            <List className="h-4 w-4 mr-2" />
            Week
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onViewModeChange('day')}
            className={`px-3 ${
              viewMode === 'day'
                ? 'bg-blue-600 text-white'
                : 'text-slate-400 hover:text-white hover:bg-slate-700'
            }`}
            title="Day view"
          >
            <CalendarDays className="h-4 w-4 mr-2" />
            Day
          </Button>
        </div>
      </div>
    </div>
  )
}
