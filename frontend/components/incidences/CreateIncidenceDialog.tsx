"use client"

import { useState } from "react"
import { Check, X, Users, Paperclip, FileText, Search, User } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { FileUpload } from "@/components/ui/file-upload"
import {
  Employee,
  IncidenceType,
  PayrollPeriod,
  CreateIncidenceRequest,
} from "@/lib/api-client"
import { SelectionMode, formatDate } from "./constants"

interface CreateIncidenceDialogProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  selectionMode: SelectionMode
  employees: Employee[]
  incidenceTypes: IncidenceType[]
  periods: PayrollPeriod[]
  onSave: (
    formData: CreateIncidenceRequest,
    targetEmployees: string[],
    pendingFiles: File[]
  ) => Promise<void>
  saving: boolean
  savingProgress: { current: number; total: number }
  uploadingFiles: boolean
}

export function CreateIncidenceDialog({
  isOpen,
  onOpenChange,
  selectionMode,
  employees,
  incidenceTypes,
  periods,
  onSave,
  saving,
  savingProgress,
  uploadingFiles,
}: CreateIncidenceDialogProps) {
  const [selectedEmployees, setSelectedEmployees] = useState<string[]>([])
  const [employeeModalSearch, setEmployeeModalSearch] = useState("")
  const [pendingFiles, setPendingFiles] = useState<File[]>([])

  const defaultPeriod = periods.find(p => p.status === "open")?.id || ""

  const [formData, setFormData] = useState<CreateIncidenceRequest>({
    employee_id: "",
    payroll_period_id: defaultPeriod,
    incidence_type_id: "",
    start_date: new Date().toISOString().split("T")[0],
    end_date: new Date().toISOString().split("T")[0],
    quantity: 1,
    comments: "",
  })

  // Reset form when dialog opens
  const handleOpenChange = (open: boolean) => {
    if (open) {
      setSelectedEmployees([])
      setPendingFiles([])
      setEmployeeModalSearch("")
      setFormData({
        employee_id: "",
        payroll_period_id: defaultPeriod,
        incidence_type_id: "",
        start_date: new Date().toISOString().split("T")[0],
        end_date: new Date().toISOString().split("T")[0],
        quantity: 1,
        comments: "",
      })
    }
    onOpenChange(open)
  }

  // Filter employees for modal search
  const filteredModalEmployees = employees.filter(emp => {
    if (!employeeModalSearch) return true
    const search = employeeModalSearch.toLowerCase()
    return (
      emp.first_name?.toLowerCase().includes(search) ||
      emp.last_name?.toLowerCase().includes(search) ||
      emp.employee_number?.toLowerCase().includes(search) ||
      `${emp.first_name} ${emp.last_name}`.toLowerCase().includes(search)
    )
  })

  const handleFileSelect = (file: File) => {
    setPendingFiles(prev => [...prev, file])
  }

  const handleRemovePendingFile = (index: number) => {
    setPendingFiles(prev => prev.filter((_, i) => i !== index))
  }

  const handleEmployeeToggle = (employeeId: string) => {
    setSelectedEmployees(prev =>
      prev.includes(employeeId)
        ? prev.filter(id => id !== employeeId)
        : [...prev, employeeId]
    )
  }

  const handleSelectAll = () => {
    if (selectedEmployees.length === employees.length) {
      setSelectedEmployees([])
    } else {
      setSelectedEmployees(employees.map(e => e.id))
    }
  }

  const getTargetEmployees = (): string[] => {
    switch (selectionMode) {
      case "single":
        return formData.employee_id ? [formData.employee_id] : []
      case "multiple":
        return selectedEmployees
      case "all":
        return employees.map(e => e.id)
      default:
        return []
    }
  }

  const handleSave = async () => {
    const targetEmployees = getTargetEmployees()
    if (targetEmployees.length === 0) return
    await onSave(formData, targetEmployees, pendingFiles)
    setPendingFiles([])
  }

  const selectedType = incidenceTypes.find(t => t.id === formData.incidence_type_id)
  const isEvidenceRequired = selectedType?.requires_evidence || false
  const targetEmployeeCount = getTargetEmployees().length

  const isDisabled =
    saving ||
    !formData.incidence_type_id ||
    (selectionMode === "single" && !formData.employee_id) ||
    (selectionMode === "multiple" && selectedEmployees.length === 0) ||
    (selectionMode === "single" && isEvidenceRequired && pendingFiles.length === 0)

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {selectionMode === "single" && "New Incidence - One Employee"}
            {selectionMode === "multiple" && "New Incidence - Multiple Employees"}
            {selectionMode === "all" && "New Incidence - All Employees"}
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            {selectionMode === "single" && "Register a new incidence for one employee"}
            {selectionMode === "multiple" && `Select employees (${selectedEmployees.length} selected)`}
            {selectionMode === "all" && `Will apply to all ${employees.length} active employees`}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Period Selection */}
          <div className="space-y-2">
            <Label>Payroll Period (Week)</Label>
            <Select
              value={formData.payroll_period_id}
              onValueChange={(value) => setFormData({ ...formData, payroll_period_id: value })}
            >
              <SelectTrigger className="bg-slate-800 border-slate-600">
                <SelectValue placeholder="Select period..." />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-600">
                {periods.filter(p => p.status === "open").map((period) => (
                  <SelectItem key={period.id} value={period.id}>
                    Week {period.period_number} - {formatDate(period.start_date)} to {formatDate(period.end_date)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Employee Selection based on mode */}
          {selectionMode === "single" && (
            <div className="space-y-2">
              <Label>Employee</Label>
              <div className="space-y-2">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-400" />
                  <Input
                    placeholder="Search by name or number..."
                    value={employeeModalSearch}
                    onChange={(e) => setEmployeeModalSearch(e.target.value)}
                    className="pl-9 bg-slate-800 border-slate-600"
                  />
                </div>
                <div className="max-h-48 overflow-y-auto bg-slate-800 rounded-lg border border-slate-600">
                  {filteredModalEmployees.length === 0 ? (
                    <div className="p-3 text-center text-slate-400 text-sm">
                      No employees found
                    </div>
                  ) : (
                    filteredModalEmployees.map((emp) => (
                      <div
                        key={emp.id}
                        className={`flex items-center gap-3 p-3 cursor-pointer transition-colors ${
                          formData.employee_id === emp.id
                            ? "bg-blue-600/30 border-l-2 border-blue-500"
                            : "hover:bg-slate-700/50"
                        }`}
                        onClick={() => setFormData({ ...formData, employee_id: emp.id })}
                      >
                        <User className="h-4 w-4 text-slate-400 flex-shrink-0" />
                        <div className="min-w-0 flex-1">
                          <div className="text-sm text-white truncate">
                            {emp.first_name} {emp.last_name}
                          </div>
                          <div className="text-xs text-slate-400">{emp.employee_number}</div>
                        </div>
                        {formData.employee_id === emp.id && (
                          <Check className="h-4 w-4 text-blue-400 flex-shrink-0" />
                        )}
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          )}

          {selectionMode === "multiple" && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Employees ({selectedEmployees.length} selected)</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleSelectAll}
                  className="text-blue-400 hover:text-blue-300"
                >
                  {selectedEmployees.length === employees.length ? "Deselect all" : "Select all"}
                </Button>
              </div>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-400" />
                <Input
                  placeholder="Search by name or number..."
                  value={employeeModalSearch}
                  onChange={(e) => setEmployeeModalSearch(e.target.value)}
                  className="pl-9 bg-slate-800 border-slate-600"
                />
              </div>
              <div className="max-h-48 overflow-y-auto bg-slate-800 rounded-lg border border-slate-600 p-2">
                {filteredModalEmployees.length === 0 ? (
                  <div className="p-3 text-center text-slate-400 text-sm">
                    No employees found
                  </div>
                ) : (
                  filteredModalEmployees.map((emp) => (
                    <div
                      key={emp.id}
                      className="flex items-center gap-3 p-2 hover:bg-slate-700/50 rounded cursor-pointer"
                      onClick={() => handleEmployeeToggle(emp.id)}
                    >
                      <Checkbox
                        checked={selectedEmployees.includes(emp.id)}
                        onCheckedChange={() => handleEmployeeToggle(emp.id)}
                        className="border-slate-500"
                      />
                      <span className="text-sm text-slate-200">
                        {emp.first_name} {emp.last_name}
                      </span>
                      <span className="text-xs text-slate-500">{emp.employee_number}</span>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}

          {selectionMode === "all" && (
            <div className="bg-green-500/10 border border-green-500/30 rounded-lg p-3">
              <div className="flex items-center gap-2 text-green-400">
                <Users className="h-4 w-4" />
                <span className="font-medium">
                  This incidence will apply to all {employees.length} active employees
                </span>
              </div>
            </div>
          )}

          {/* Incidence Type */}
          <div className="space-y-2">
            <Label>Incidence Type</Label>
            <Select
              value={formData.incidence_type_id}
              onValueChange={(value) => setFormData({ ...formData, incidence_type_id: value })}
            >
              <SelectTrigger className="bg-slate-800 border-slate-600">
                <SelectValue placeholder="Select type..." />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-600">
                {incidenceTypes.map((type) => (
                  <SelectItem key={type.id} value={type.id}>
                    {type.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Dates */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="start_date">Start Date</Label>
              <Input
                id="start_date"
                type="date"
                value={formData.start_date}
                onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                className="bg-slate-800 border-slate-600"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="end_date">End Date</Label>
              <Input
                id="end_date"
                type="date"
                value={formData.end_date}
                onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                className="bg-slate-800 border-slate-600"
              />
            </div>
          </div>

          {/* Quantity */}
          <div className="space-y-2">
            <Label htmlFor="quantity">Quantity (days/hours)</Label>
            <Input
              id="quantity"
              type="number"
              min="0.5"
              step="0.5"
              value={formData.quantity}
              onChange={(e) => setFormData({ ...formData, quantity: parseFloat(e.target.value) || 1 })}
              className="bg-slate-800 border-slate-600"
            />
          </div>

          {/* Comments */}
          <div className="space-y-2">
            <Label htmlFor="comments">Comments (Optional)</Label>
            <Input
              id="comments"
              value={formData.comments}
              onChange={(e) => setFormData({ ...formData, comments: e.target.value })}
              placeholder="Additional notes about the incidence"
              className="bg-slate-800 border-slate-600"
            />
          </div>

          {/* File Upload Section - Only for single employee mode */}
          {selectionMode === "single" && (
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Paperclip className="h-4 w-4" />
                Evidence {isEvidenceRequired ? (
                  <span className="text-orange-400 text-xs font-medium">(Required)</span>
                ) : (
                  <span className="text-slate-500 text-xs">(Optional)</span>
                )}
              </Label>
              {isEvidenceRequired && pendingFiles.length === 0 && (
                <div className="bg-orange-500/10 border border-orange-500/30 rounded-lg p-3">
                  <div className="text-orange-400 text-sm">
                    This incidence type requires mandatory evidence. Please attach at least one file.
                  </div>
                </div>
              )}
              <FileUpload
                onUpload={async (file) => {
                  handleFileSelect(file)
                  return Promise.resolve()
                }}
                label=""
              />
              {pendingFiles.length > 0 && (
                <div className="space-y-2 mt-2">
                  <Label className="text-slate-400 text-sm">Files to attach ({pendingFiles.length})</Label>
                  <div className="space-y-1">
                    {pendingFiles.map((file, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between p-2 bg-slate-800/50 rounded-lg border border-slate-700"
                      >
                        <div className="flex items-center gap-2 min-w-0">
                          <FileText className="h-4 w-4 text-blue-400 flex-shrink-0" />
                          <span className="text-sm text-slate-300 truncate">{file.name}</span>
                          <span className="text-xs text-slate-500">
                            ({(file.size / 1024).toFixed(1)} KB)
                          </span>
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemovePendingFile(index)}
                          className="h-6 w-6 p-0 text-slate-400 hover:text-red-400"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {selectionMode !== "single" && (
            <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-3">
              <div className="text-yellow-400 text-sm flex items-center gap-2">
                <Paperclip className="h-4 w-4" />
                To attach evidence, use "One Employee" mode or attach later from the table.
              </div>
            </div>
          )}

          {/* Progress indicator */}
          {uploadingFiles && (
            <div className="bg-green-500/10 border border-green-500/30 rounded-lg p-3">
              <div className="text-green-400 text-sm">
                Uploading attached files...
              </div>
            </div>
          )}

          {saving && savingProgress.total > 1 && (
            <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-3">
              <div className="text-blue-400 text-sm">
                Processing {savingProgress.current} of {savingProgress.total} employees...
              </div>
              <div className="w-full bg-slate-700 rounded-full h-2 mt-2">
                <div
                  className="bg-blue-500 h-2 rounded-full transition-all"
                  style={{ width: `${(savingProgress.current / savingProgress.total) * 100}%` }}
                />
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            className="border-slate-600"
            disabled={saving}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={isDisabled}
            className="bg-blue-600 hover:bg-blue-700"
          >
            {saving
              ? savingProgress.total > 1
                ? `Saving (${savingProgress.current}/${savingProgress.total})...`
                : "Saving..."
              : selectionMode === "single"
                ? "Register"
                : `Register for ${targetEmployeeCount} employee(s)`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
