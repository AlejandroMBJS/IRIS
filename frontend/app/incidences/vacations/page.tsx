/**
 * @file app/incidences/vacations/page.tsx
 * @description Vacation tracking and request management following Mexican Labor Law (LFT)
 *
 * USER PERSPECTIVE:
 *   - View vacation balances for all employees
 *   - See days entitled based on years of service (per LFT)
 *   - Track used, pending, and available vacation days
 *   - Request vacation days for employees
 *   - View LFT vacation schedule reference
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Vacation calculation display, request form fields
 *   CAUTION: LFT vacation schedule is legally mandated in Mexico
 *   DO NOT modify: Vacation entitlement calculations without legal review
 *
 * KEY COMPONENTS:
 *   - Vacation balance table: Per-employee vacation tracking
 *   - Summary cards: Total days available, used, pending across all employees
 *   - LFT reference: Legal vacation day requirements by years of service
 *   - Request dialog: Vacation request form with date range
 *
 * API ENDPOINTS USED:
 *   - GET /api/employees (via employeeApi.getEmployees)
 *   - GET /api/incidence-types (via incidenceTypeApi.getAll)
 *   - GET /api/incidences/vacation-balance/:employeeId (via incidenceApi.getVacationBalance)
 *   - POST /api/incidences (via incidenceApi.create)
 */

"use client"

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Palmtree, Calendar, Plus, User, Clock, Check, Paperclip, FileText, X } from "lucide-react"
import { FileUpload } from "@/components/ui/file-upload"
import {
  incidenceApi,
  incidenceTypeApi,
  employeeApi,
  evidenceApi,
  Employee,
  VacationBalance,
  IncidenceType,
  CreateIncidenceRequest,
} from "@/lib/api-client"

export default function VacationsPage() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [vacationBalances, setVacationBalances] = useState<Map<string, VacationBalance>>(new Map())
  const [incidenceTypes, setIncidenceTypes] = useState<IncidenceType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(null)
  const [vacationType, setVacationType] = useState<IncidenceType | null>(null)

  const [formData, setFormData] = useState<CreateIncidenceRequest>({
    employee_id: "",
    incidence_type_id: "",
    start_date: "",
    end_date: "",
    quantity: 1,
    comments: "",
  })
  const [saving, setSaving] = useState(false)
  const [pendingFiles, setPendingFiles] = useState<File[]>([])
  const [uploadingFiles, setUploadingFiles] = useState(false)
  const [successMessage, setSuccessMessage] = useState("")

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      setLoading(true)
      const [empsResponse, types] = await Promise.all([
        employeeApi.getEmployees(),
        incidenceTypeApi.getAll(),
      ])

      // Get employees array from response
      const employeesList = empsResponse.employees || []

      // Filter only active employees
      const activeEmployees = employeesList.filter(e => e.employment_status === "active")
      setEmployees(activeEmployees)
      setIncidenceTypes(types)

      // Find vacation type
      const vacType = types.find(t => t.category === "vacation")
      setVacationType(vacType || null)

      // Load vacation balances for each employee
      const balances = new Map<string, VacationBalance>()
      for (const emp of activeEmployees) {
        try {
          const balance = await incidenceApi.getVacationBalance(emp.id)
          balances.set(emp.id, balance)
        } catch (err) {
          // Skip employees without balance data
        }
      }
      setVacationBalances(balances)
      setError("")
    } catch (err: any) {
      setError(err.message || "Error loading data")
    } finally {
      setLoading(false)
    }
  }

  // Clear success message after 5 seconds
  useEffect(() => {
    if (successMessage) {
      const timer = setTimeout(() => setSuccessMessage(""), 5000)
      return () => clearTimeout(timer)
    }
  }, [successMessage])

  const handleRequestVacation = (employee: Employee) => {
    if (!vacationType) {
      setError("No incidence type has been configured for vacations")
      return
    }

    setSelectedEmployee(employee)
    setPendingFiles([])
    setFormData({
      employee_id: employee.id,
      incidence_type_id: vacationType.id,
      start_date: new Date().toISOString().split("T")[0],
      end_date: new Date().toISOString().split("T")[0],
      quantity: 1,
      comments: "",
    })
    setIsDialogOpen(true)
  }

  const handleFileSelect = (file: File) => {
    setPendingFiles(prev => [...prev, file])
  }

  const handleRemovePendingFile = (index: number) => {
    setPendingFiles(prev => prev.filter((_, i) => i !== index))
  }

  const calculateDays = (startDate: string, endDate: string) => {
    const start = new Date(startDate)
    const end = new Date(endDate)
    const diffTime = Math.abs(end.getTime() - start.getTime())
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1
    return diffDays
  }

  const handleDateChange = (field: "start_date" | "end_date", value: string) => {
    const newFormData = { ...formData, [field]: value }
    if (newFormData.start_date && newFormData.end_date) {
      newFormData.quantity = calculateDays(newFormData.start_date, newFormData.end_date)
    }
    setFormData(newFormData)
  }

  const handleSave = async () => {
    try {
      setSaving(true)
      setError("")
      const response = await incidenceApi.create(formData)

      // Upload pending files if we have any
      if (pendingFiles.length > 0 && response && response.id) {
        setUploadingFiles(true)
        for (const file of pendingFiles) {
          try {
            await evidenceApi.upload(response.id, file)
          } catch (err) {
            console.error(`Error uploading file ${file.name}:`, err)
          }
        }
        setUploadingFiles(false)
      }

      await loadData()
      setIsDialogOpen(false)
      setPendingFiles([])

      const fileMsg = pendingFiles.length > 0
        ? ` with ${pendingFiles.length} attached file(s)`
        : ""
      setSuccessMessage(`Vacation request registered successfully${fileMsg}.`)
    } catch (err: any) {
      setError(err.message || "Error saving vacation request")
    } finally {
      setSaving(false)
      setUploadingFiles(false)
    }
  }

  const getAvailabilityColor = (available: number, entitled: number) => {
    const percentage = (available / entitled) * 100
    if (percentage >= 75) return "text-green-400"
    if (percentage >= 50) return "text-yellow-400"
    if (percentage >= 25) return "text-orange-400"
    return "text-red-400"
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white flex items-center gap-2">
              <Palmtree className="h-6 w-6" />
              Vacation Management
            </h1>
            <p className="text-slate-400 mt-1">
              Tracking vacation days per employee according to the Federal Labor Law
            </p>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <div className="text-slate-400 text-sm">Total Employees</div>
            <div className="text-2xl font-bold text-white">{employees.length}</div>
          </div>
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <div className="text-slate-400 text-sm">Available Days (Total)</div>
            <div className="text-2xl font-bold text-green-400">
              {Array.from(vacationBalances.values()).reduce((acc, b) => acc + b.available_days, 0).toFixed(0)}
            </div>
          </div>
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <div className="text-slate-400 text-sm">Used Days (Total)</div>
            <div className="text-2xl font-bold text-blue-400">
              {Array.from(vacationBalances.values()).reduce((acc, b) => acc + b.used_days, 0).toFixed(0)}
            </div>
          </div>
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <div className="text-slate-400 text-sm">Pending Requests</div>
            <div className="text-2xl font-bold text-yellow-400">
              {Array.from(vacationBalances.values()).reduce((acc, b) => acc + b.pending_days, 0).toFixed(0)}
            </div>
          </div>
        </div>

        {/* Info Box */}
        <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
          <h3 className="text-blue-400 font-medium mb-2">Vacation Days according to LFT</h3>
          <div className="text-slate-300 text-sm grid grid-cols-2 md:grid-cols-5 gap-2">
            <div>1 year: 12 days</div>
            <div>2 years: 14 days</div>
            <div>3 years: 16 days</div>
            <div>4 years: 18 days</div>
            <div>5-9 years: 20 days</div>
            <div>10-14 years: 22 days</div>
            <div>15-19 years: 24 days</div>
            <div>20-24 years: 26 days</div>
            <div>25-29 years: 28 days</div>
            <div>30+ years: 30 days</div>
          </div>
        </div>

        {/* Success Message */}
        {successMessage && (
          <div className="bg-green-500/10 border border-green-500/50 rounded-lg p-4 text-green-400">
            {successMessage}
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/50 rounded-lg p-4 text-red-400">
            {error}
          </div>
        )}

        {/* Table */}
        <div className="bg-slate-800/50 rounded-lg border border-slate-700">
          <Table>
            <TableHeader>
              <TableRow className="border-slate-700 hover:bg-slate-800/50">
                <TableHead className="text-slate-300">Employee</TableHead>
                <TableHead className="text-slate-300">Seniority</TableHead>
                <TableHead className="text-slate-300">Entitled Days</TableHead>
                <TableHead className="text-slate-300">Used Days</TableHead>
                <TableHead className="text-slate-300">Pending</TableHead>
                <TableHead className="text-slate-300">Available</TableHead>
                <TableHead className="text-slate-300 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-slate-400 py-8">
                    Loading vacation information...
                  </TableCell>
                </TableRow>
              ) : employees.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-slate-400 py-8">
                    No active employees
                  </TableCell>
                </TableRow>
              ) : (
                employees.map((employee) => {
                  const balance = vacationBalances.get(employee.id)
                  return (
                    <TableRow key={employee.id} className="border-slate-700 hover:bg-slate-800/30">
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <User className="h-4 w-4 text-slate-400" />
                          <div>
                            <div className="text-white font-medium">
                              {employee.first_name} {employee.last_name}
                            </div>
                            <div className="text-slate-400 text-xs">
                              {employee.employee_number}
                            </div>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-slate-300">
                        <div className="flex items-center gap-2">
                          <Clock className="h-4 w-4 text-slate-400" />
                          {balance ? `${balance.years_of_service.toFixed(1)} years` : "-"}
                        </div>
                      </TableCell>
                      <TableCell className="text-white font-medium">
                        {balance?.entitled_days || "-"}
                      </TableCell>
                      <TableCell className="text-blue-400">
                        {balance?.used_days || 0}
                      </TableCell>
                      <TableCell className="text-yellow-400">
                        {balance?.pending_days || 0}
                      </TableCell>
                      <TableCell>
                        <span className={balance ? getAvailabilityColor(balance.available_days, balance.entitled_days) : "text-slate-400"}>
                          {balance?.available_days?.toFixed(0) || "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleRequestVacation(employee)}
                          disabled={!balance || balance.available_days <= 0}
                          className="border-slate-600 text-slate-300 hover:text-white"
                        >
                          <Plus className="h-4 w-4 mr-1" />
                          Request
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>

        {/* Request Vacation Dialog */}
        {isDialogOpen && (
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <Palmtree className="h-5 w-5" />
                Request Vacation
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                {selectedEmployee && (
                  <span>
                    Employee: {selectedEmployee.first_name} {selectedEmployee.last_name}
                    <br />
                    Available days: {vacationBalances.get(selectedEmployee.id)?.available_days.toFixed(0) || 0}
                  </span>
                )}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="start_date">Start Date</Label>
                  <Input
                    id="start_date"
                    type="date"
                    value={formData.start_date}
                    onChange={(e) => handleDateChange("start_date", e.target.value)}
                    className="bg-slate-800 border-slate-600"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="end_date">End Date</Label>
                  <Input
                    id="end_date"
                    type="date"
                    value={formData.end_date}
                    onChange={(e) => handleDateChange("end_date", e.target.value)}
                    className="bg-slate-800 border-slate-600"
                  />
                </div>
              </div>

              <div className="bg-slate-800 rounded-lg p-4">
                <div className="text-slate-400 text-sm">Requested days</div>
                <div className="text-2xl font-bold text-white">{formData.quantity}</div>
                {selectedEmployee && vacationBalances.get(selectedEmployee.id) && (
                  formData.quantity > (vacationBalances.get(selectedEmployee.id)?.available_days || 0) && (
                    <div className="text-red-400 text-sm mt-1">
                      Exceeds available days
                    </div>
                  )
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="comments">Comments (Optional)</Label>
                <Input
                  id="comments"
                  value={formData.comments}
                  onChange={(e) => setFormData({ ...formData, comments: e.target.value })}
                  placeholder="Additional notes"
                  className="bg-slate-800 border-slate-600"
                />
              </div>

              {/* File Upload Section */}
              <div className="space-y-2">
                {(() => {
                  const isRequired = vacationType?.requires_evidence || false
                  return (
                    <>
                      <Label className="flex items-center gap-2">
                        <Paperclip className="h-4 w-4" />
                        Evidence {isRequired ? (
                          <span className="text-orange-400 text-xs font-medium">(Required)</span>
                        ) : (
                          <span className="text-slate-500 text-xs">(Optional)</span>
                        )}
                      </Label>
                      {isRequired && pendingFiles.length === 0 && (
                        <div className="bg-orange-500/10 border border-orange-500/30 rounded-lg p-3">
                          <div className="text-orange-400 text-sm">
                            Vacation requests require mandatory evidence. Please attach at least one file.
                          </div>
                        </div>
                      )}
                    </>
                  )
                })()}
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

              {/* Progress indicator */}
              {uploadingFiles && (
                <div className="bg-green-500/10 border border-green-500/30 rounded-lg p-3">
                  <div className="text-green-400 text-sm">
                    Uploading attached files...
                  </div>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsDialogOpen(false)}
                className="border-slate-600"
                disabled={saving || uploadingFiles}
              >
                Cancel
              </Button>
              <Button
                onClick={handleSave}
                disabled={
                  saving ||
                  uploadingFiles ||
                  !formData.start_date ||
                  !formData.end_date ||
                  (selectedEmployee !== null && formData.quantity > (vacationBalances.get(selectedEmployee.id)?.available_days || 0)) ||
                  // Enforce evidence requirement if vacation type requires it
                  (vacationType?.requires_evidence && pendingFiles.length === 0)
                }
                className="bg-green-600 hover:bg-green-700"
              >
                <Check className="h-4 w-4 mr-2" />
                {saving ? "Saving..." : uploadingFiles ? "Uploading files..." : "Request Vacation"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        )}
      </div>
    </DashboardLayout>
  )
}
