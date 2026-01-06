"use client"

/**
 * Calendar Sidebar Component
 *
 * Provides filters, quick actions, and summary statistics for the HR calendar.
 * Features improved employee display with avatars, department grouping, and event counts.
 */

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Badge } from "@/components/ui/badge"
import { CalendarEmployee, CalendarSummary, CalendarEventType } from "@/lib/api-client"
import {
  Plus,
  Calendar,
  ClipboardList,
  Clock,
  Users,
  Filter,
  ChevronDown,
  ChevronRight,
  Search,
  Building2,
  UserCircle,
  LayoutGrid,
  List,
  X,
  Sparkles,
  TrendingUp,
  AlertCircle,
  CheckCircle2,
} from "lucide-react"
import { useState, useMemo } from "react"

export type CollarType = 'white_collar' | 'blue_collar' | 'gray_collar'

interface CalendarSidebarProps {
  employees: CalendarEmployee[]
  selectedEmployeeIds: string[]
  onEmployeeSelect: (ids: string[]) => void
  selectedCollarTypes: CollarType[]
  onCollarTypeChange: (types: CollarType[]) => void
  selectedEventTypes: CalendarEventType[]
  onEventTypeChange: (types: CalendarEventType[]) => void
  statusFilter: string
  onStatusChange: (status: string) => void
  summary: CalendarSummary | null
  onCreateAbsence: () => void
  onCreateIncidence: () => void
}

const COLLAR_TYPES = [
  { value: 'white_collar', label: 'Administrative', color: 'bg-blue-500', lightBg: 'bg-blue-500/20', textColor: 'text-blue-400' },
  { value: 'blue_collar', label: 'Operative', color: 'bg-indigo-500', lightBg: 'bg-indigo-500/20', textColor: 'text-indigo-400' },
  { value: 'gray_collar', label: 'Mixed', color: 'bg-slate-500', lightBg: 'bg-slate-500/20', textColor: 'text-slate-400' },
]

const EVENT_TYPES: { value: CalendarEventType; label: string; icon: typeof Calendar; color: string }[] = [
  { value: 'absence', label: 'Absences', icon: Calendar, color: 'text-red-400' },
  { value: 'incidence', label: 'Incidences', icon: ClipboardList, color: 'text-amber-400' },
  { value: 'shift_change', label: 'Shift Changes', icon: Clock, color: 'text-purple-400' },
]

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses', icon: LayoutGrid },
  { value: 'pending', label: 'Pending', icon: Clock, color: 'text-yellow-400' },
  { value: 'approved', label: 'Approved', icon: CheckCircle2, color: 'text-green-400' },
  { value: 'declined', label: 'Declined', icon: X, color: 'text-red-400' },
]

// Generate initials from name
function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map(n => n[0])
    .join('')
    .toUpperCase()
}

// Generate a consistent color based on string
function stringToColor(str: string): string {
  const colors = [
    'bg-rose-500', 'bg-pink-500', 'bg-fuchsia-500', 'bg-purple-500',
    'bg-violet-500', 'bg-indigo-500', 'bg-blue-500', 'bg-sky-500',
    'bg-cyan-500', 'bg-teal-500', 'bg-emerald-500', 'bg-green-500',
    'bg-lime-500', 'bg-yellow-500', 'bg-amber-500', 'bg-orange-500',
  ]
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

export function CalendarSidebar({
  employees,
  selectedEmployeeIds,
  onEmployeeSelect,
  selectedCollarTypes,
  onCollarTypeChange,
  selectedEventTypes,
  onEventTypeChange,
  statusFilter,
  onStatusChange,
  summary,
  onCreateAbsence,
  onCreateIncidence,
}: CalendarSidebarProps) {

  const [showEmployees, setShowEmployees] = useState(true)
  const [employeeSearch, setEmployeeSearch] = useState('')
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list')
  const [groupByDepartment, setGroupByDepartment] = useState(true)

  const handleCollarTypeToggle = (type: CollarType) => {
    if (selectedCollarTypes.includes(type)) {
      onCollarTypeChange(selectedCollarTypes.filter(t => t !== type))
    } else {
      onCollarTypeChange([...selectedCollarTypes, type])
    }
  }

  const handleEventTypeToggle = (type: CalendarEventType) => {
    if (selectedEventTypes.includes(type)) {
      onEventTypeChange(selectedEventTypes.filter(t => t !== type))
    } else {
      onEventTypeChange([...selectedEventTypes, type])
    }
  }

  const handleEmployeeToggle = (id: string) => {
    if (selectedEmployeeIds.includes(id)) {
      onEmployeeSelect(selectedEmployeeIds.filter(i => i !== id))
    } else {
      onEmployeeSelect([...selectedEmployeeIds, id])
    }
  }

  const clearEmployeeSelection = () => {
    onEmployeeSelect([])
  }

  const selectAllEmployees = () => {
    onEmployeeSelect(filteredEmployees.map(e => e.id))
  }

  // Filter employees by search
  const filteredEmployees = useMemo(() => {
    return employees.filter(emp =>
      emp.full_name.toLowerCase().includes(employeeSearch.toLowerCase()) ||
      emp.employee_number.toLowerCase().includes(employeeSearch.toLowerCase())
    )
  }, [employees, employeeSearch])

  // Group employees by department
  const groupedEmployees = useMemo(() => {
    if (!groupByDepartment) return { 'All': filteredEmployees }

    return filteredEmployees.reduce((groups, emp) => {
      const dept = emp.department_name || 'No Department'
      if (!groups[dept]) {
        groups[dept] = []
      }
      groups[dept].push(emp)
      return groups
    }, {} as Record<string, CalendarEmployee[]>)
  }, [filteredEmployees, groupByDepartment])

  const selectedCount = selectedEmployeeIds.length
  const isAllSelected = selectedCount === filteredEmployees.length && filteredEmployees.length > 0

  return (
    <div className="w-80 h-full bg-gradient-to-b from-slate-800/80 to-slate-900/80 border-r border-slate-700/50 flex flex-col backdrop-blur-xl overflow-hidden">
      {/* Header with Quick Actions */}
      <div className="flex-shrink-0 p-4 border-b border-slate-700/50">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-blue-400" />
            HR Calendar
          </h2>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <Button
            onClick={onCreateAbsence}
            className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-lg shadow-blue-500/20 h-11"
            size="sm"
          >
            <Plus className="h-4 w-4 mr-1.5" />
            Absence
          </Button>
          <Button
            onClick={onCreateIncidence}
            variant="outline"
            className="border-slate-600 hover:bg-slate-700 hover:border-slate-500 h-11"
            size="sm"
          >
            <Plus className="h-4 w-4 mr-1.5" />
            Incidence
          </Button>
        </div>
      </div>

      {/* Summary Stats */}
      {summary && (
        <div className="flex-shrink-0 p-4 border-b border-slate-700/50">
          <div className="flex items-center gap-2 mb-3">
            <TrendingUp className="h-4 w-4 text-emerald-400" />
            <h3 className="text-sm font-medium text-slate-300">Period Summary</h3>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-700/50">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-red-400" />
                <span className="text-xs text-slate-400">Absences</span>
              </div>
              <div className="text-2xl font-bold text-white mt-1">{summary.total_absences}</div>
            </div>
            <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-700/50">
              <div className="flex items-center gap-2">
                <ClipboardList className="h-4 w-4 text-amber-400" />
                <span className="text-xs text-slate-400">Incidences</span>
              </div>
              <div className="text-2xl font-bold text-white mt-1">{summary.total_incidences}</div>
            </div>
            <div className="bg-yellow-500/10 rounded-xl p-3 border border-yellow-500/30">
              <div className="flex items-center gap-2">
                <AlertCircle className="h-4 w-4 text-yellow-400" />
                <span className="text-xs text-yellow-400">Pending</span>
              </div>
              <div className="text-2xl font-bold text-yellow-400 mt-1">{summary.pending}</div>
            </div>
            <div className="bg-emerald-500/10 rounded-xl p-3 border border-emerald-500/30">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                <span className="text-xs text-emerald-400">Approved</span>
              </div>
              <div className="text-2xl font-bold text-emerald-400 mt-1">{summary.approved}</div>
            </div>
          </div>
        </div>
      )}

      {/* Filters Section */}
      <div className="flex-shrink-0 p-4 border-b border-slate-700/50 space-y-4">
        {/* Collar Type Filter */}
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2 flex items-center gap-2">
            <Filter className="h-3.5 w-3.5" />
            Employee Type
          </h3>
          <div className="flex flex-wrap gap-2">
            {COLLAR_TYPES.map(type => {
              const isSelected = selectedCollarTypes.length === 0 || selectedCollarTypes.includes(type.value as CollarType)
              return (
                <button
                  key={type.value}
                  onClick={() => handleCollarTypeToggle(type.value as CollarType)}
                  className={`px-3 py-1.5 rounded-full text-xs font-medium transition-all ${
                    isSelected
                      ? `${type.lightBg} ${type.textColor} ring-1 ring-current`
                      : 'bg-slate-800 text-slate-500 hover:bg-slate-700'
                  }`}
                >
                  <span className={`inline-block w-2 h-2 rounded-full ${type.color} mr-1.5`} />
                  {type.label}
                </button>
              )
            })}
          </div>
        </div>

        {/* Event Type Filter */}
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
            Event Type
          </h3>
          <div className="flex flex-wrap gap-2">
            {EVENT_TYPES.map(type => {
              const isSelected = selectedEventTypes.length === 0 || selectedEventTypes.includes(type.value)
              const Icon = type.icon
              return (
                <button
                  key={type.value}
                  onClick={() => handleEventTypeToggle(type.value)}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-all ${
                    isSelected
                      ? `bg-slate-700 ${type.color} ring-1 ring-current/30`
                      : 'bg-slate-800 text-slate-500 hover:bg-slate-700'
                  }`}
                >
                  <Icon className="h-3.5 w-3.5" />
                  {type.label}
                </button>
              )
            })}
          </div>
        </div>

        {/* Status Filter */}
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
            Status
          </h3>
          <div className="grid grid-cols-2 gap-1.5">
            {STATUS_OPTIONS.map(option => {
              const isSelected = statusFilter === option.value
              const Icon = option.icon
              return (
                <button
                  key={option.value}
                  onClick={() => onStatusChange(option.value)}
                  className={`flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium transition-all ${
                    isSelected
                      ? 'bg-slate-700 text-white ring-1 ring-slate-500'
                      : 'bg-slate-800/50 text-slate-400 hover:bg-slate-800 hover:text-slate-300'
                  }`}
                >
                  <Icon className={`h-3.5 w-3.5 ${option.color || ''}`} />
                  {option.label}
                </button>
              )
            })}
          </div>
        </div>
      </div>

      {/* Employee List Section */}
      <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
        {/* Employee Section Header */}
        <div className="p-4 pb-2">
          <button
            onClick={() => setShowEmployees(!showEmployees)}
            className="w-full flex items-center justify-between text-sm font-medium text-slate-300 hover:text-white transition-colors"
          >
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-blue-400" />
              <span>Employees</span>
              <Badge variant="secondary" className="bg-slate-700 text-slate-300 text-xs">
                {employees.length}
              </Badge>
            </div>
            {showEmployees ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </button>
        </div>

        {showEmployees && (
          <>
            {/* Search and View Controls */}
            <div className="px-4 pb-2 space-y-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                  type="text"
                  placeholder="Search employees..."
                  value={employeeSearch}
                  onChange={(e) => setEmployeeSearch(e.target.value)}
                  className="w-full bg-slate-900/50 border border-slate-700 rounded-xl pl-9 pr-3 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
                />
                {employeeSearch && (
                  <button
                    onClick={() => setEmployeeSearch('')}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
                  >
                    <X className="h-4 w-4" />
                  </button>
                )}
              </div>

              {/* View Controls */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setGroupByDepartment(!groupByDepartment)}
                    className={`flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors ${
                      groupByDepartment ? 'bg-blue-500/20 text-blue-400' : 'text-slate-500 hover:text-slate-300'
                    }`}
                  >
                    <Building2 className="h-3.5 w-3.5" />
                    Group
                  </button>
                  <button
                    onClick={() => setViewMode(viewMode === 'list' ? 'grid' : 'list')}
                    className="p-1.5 rounded text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-colors"
                  >
                    {viewMode === 'list' ? <LayoutGrid className="h-4 w-4" /> : <List className="h-4 w-4" />}
                  </button>
                </div>

                {/* Selection controls */}
                <div className="flex items-center gap-2">
                  {selectedCount > 0 && (
                    <button
                      onClick={clearEmployeeSelection}
                      className="text-xs text-slate-400 hover:text-white flex items-center gap-1"
                    >
                      <X className="h-3 w-3" />
                      Clear ({selectedCount})
                    </button>
                  )}
                  <button
                    onClick={isAllSelected ? clearEmployeeSelection : selectAllEmployees}
                    className="text-xs text-blue-400 hover:text-blue-300"
                  >
                    {isAllSelected ? 'None' : 'All'}
                  </button>
                </div>
              </div>
            </div>

            {/* Employee List */}
            <div className="flex-1 overflow-y-auto px-4 pb-4">
              {filteredEmployees.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 text-slate-500">
                  <UserCircle className="h-12 w-12 mb-2 opacity-50" />
                  <p className="text-sm">No employees found</p>
                </div>
              ) : viewMode === 'grid' ? (
                // Grid View
                <div className="grid grid-cols-3 gap-2">
                  {filteredEmployees.map(employee => {
                    const isSelected = selectedEmployeeIds.length === 0 || selectedEmployeeIds.includes(employee.id)
                    const initials = getInitials(employee.full_name)
                    const avatarColor = employee.color || stringToColor(employee.full_name)

                    return (
                      <button
                        key={employee.id}
                        onClick={() => handleEmployeeToggle(employee.id)}
                        className={`flex flex-col items-center p-2 rounded-xl transition-all ${
                          isSelected
                            ? 'bg-slate-700/50 ring-1 ring-blue-500/50'
                            : 'hover:bg-slate-800/50 opacity-50'
                        }`}
                        title={employee.full_name}
                      >
                        <div
                          className={`w-10 h-10 rounded-full flex items-center justify-center text-white text-xs font-bold shadow-lg`}
                          style={{ backgroundColor: avatarColor }}
                        >
                          {initials}
                        </div>
                        <span className="text-[10px] text-slate-400 mt-1 truncate w-full text-center">
                          {employee.full_name.split(' ')[0]}
                        </span>
                      </button>
                    )
                  })}
                </div>
              ) : (
                // List View (grouped or ungrouped)
                <div className="space-y-4">
                  {Object.entries(groupedEmployees).map(([department, deptEmployees]) => (
                    <div key={department}>
                      {groupByDepartment && Object.keys(groupedEmployees).length > 1 && (
                        <div className="flex items-center gap-2 mb-2 sticky top-0 bg-slate-900/90 backdrop-blur-sm py-1 -mx-1 px-1">
                          <Building2 className="h-3.5 w-3.5 text-slate-500" />
                          <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                            {department}
                          </span>
                          <span className="text-xs text-slate-600">({deptEmployees.length})</span>
                        </div>
                      )}
                      <div className="space-y-1">
                        {deptEmployees.map(employee => {
                          const isSelected = selectedEmployeeIds.length === 0 || selectedEmployeeIds.includes(employee.id)
                          const initials = getInitials(employee.full_name)
                          const avatarColor = employee.color || stringToColor(employee.full_name)

                          return (
                            <button
                              key={employee.id}
                              onClick={() => handleEmployeeToggle(employee.id)}
                              className={`w-full flex items-center gap-3 p-2.5 rounded-xl transition-all group ${
                                isSelected
                                  ? 'bg-slate-700/50 hover:bg-slate-700'
                                  : 'hover:bg-slate-800/50 opacity-60 hover:opacity-100'
                              }`}
                            >
                              {/* Avatar */}
                              <div className="relative">
                                <div
                                  className="w-9 h-9 rounded-full flex items-center justify-center text-white text-xs font-bold shadow-md transition-transform group-hover:scale-105"
                                  style={{ backgroundColor: avatarColor }}
                                >
                                  {initials}
                                </div>
                                {isSelected && selectedEmployeeIds.length > 0 && (
                                  <div className="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 bg-blue-500 rounded-full flex items-center justify-center">
                                    <CheckCircle2 className="h-2.5 w-2.5 text-white" />
                                  </div>
                                )}
                              </div>

                              {/* Info */}
                              <div className="flex-1 min-w-0 text-left">
                                <div className="text-sm text-white font-medium truncate group-hover:text-white">
                                  {employee.full_name}
                                </div>
                                <div className="flex items-center gap-2 text-xs text-slate-500">
                                  <span>{employee.employee_number}</span>
                                  {employee.department_name && (
                                    <>
                                      <span className="text-slate-700">•</span>
                                      <span className="truncate">{employee.department_name}</span>
                                    </>
                                  )}
                                </div>
                              </div>

                              {/* Collar type indicator */}
                              <div className="flex-shrink-0">
                                <div
                                  className={`w-2 h-2 rounded-full ${
                                    employee.collar_type === 'white_collar' ? 'bg-blue-500' :
                                    employee.collar_type === 'blue_collar' ? 'bg-indigo-500' : 'bg-slate-500'
                                  }`}
                                  title={COLLAR_TYPES.find(c => c.value === employee.collar_type)?.label}
                                />
                              </div>
                            </button>
                          )
                        })}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
