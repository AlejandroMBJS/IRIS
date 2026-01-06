"use client"

import { Calendar, Filter } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Employee, PayrollPeriod } from "@/lib/api-client"
import { formatDate } from "./constants"

interface IncidenceFiltersProps {
  periods: PayrollPeriod[]
  employees: Employee[]
  filterPeriod: string
  filterStatus: string
  filterEmployee: string
  onPeriodChange: (value: string) => void
  onStatusChange: (value: string) => void
  onEmployeeChange: (value: string) => void
  onClearFilters: () => void
}

export function IncidenceFilters({
  periods,
  employees,
  filterPeriod,
  filterStatus,
  filterEmployee,
  onPeriodChange,
  onStatusChange,
  onEmployeeChange,
  onClearFilters,
}: IncidenceFiltersProps) {
  return (
    <>
      {/* Period Selector - Primary filter */}
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="flex items-center gap-4">
          <Calendar className="h-5 w-5 text-blue-400" />
          <div className="flex-1">
            <Label className="text-slate-300 mb-2 block">Payroll Period (Week)</Label>
            <Select value={filterPeriod} onValueChange={onPeriodChange}>
              <SelectTrigger className="bg-slate-900 border-slate-600 max-w-md">
                <SelectValue placeholder="Select period..." />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-600">
                <SelectItem value="all">All periods</SelectItem>
                {periods.map((period) => (
                  <SelectItem key={period.id} value={period.id}>
                    Week {period.period_number} - {formatDate(period.start_date)} to {formatDate(period.end_date)}
                    {period.status === "open" && " (Active)"}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Additional Filters */}
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="flex items-center gap-4">
          <Filter className="h-5 w-5 text-slate-400" />
          <div className="flex-1 grid grid-cols-1 md:grid-cols-3 gap-4">
            <Select value={filterStatus} onValueChange={onStatusChange}>
              <SelectTrigger className="bg-slate-900 border-slate-600">
                <SelectValue placeholder="Filter by status..." />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-600">
                <SelectItem value="all">All statuses</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="rejected">Rejected</SelectItem>
                <SelectItem value="processed">Processed</SelectItem>
              </SelectContent>
            </Select>

            <Select value={filterEmployee} onValueChange={onEmployeeChange}>
              <SelectTrigger className="bg-slate-900 border-slate-600">
                <SelectValue placeholder="Filter by employee..." />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-600">
                <SelectItem value="all">All employees</SelectItem>
                {employees.map((emp) => (
                  <SelectItem key={emp.id} value={emp.id}>
                    {emp.first_name} {emp.last_name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Button
              variant="outline"
              onClick={onClearFilters}
              className="border-slate-600"
            >
              Clear Filters
            </Button>
          </div>
        </div>
      </div>
    </>
  )
}
