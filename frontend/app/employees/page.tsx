/**
 * @file app/employees/page.tsx
 * @description Employee list page with search, filtering, and Excel import functionality
 *
 * USER PERSPECTIVE:
 *   - View all employees in a searchable, filterable table
 *   - Search by name, employee number, or RFC
 *   - Filter by employment status and collar type
 *   - Import employees from Excel file (with template download)
 *   - Add new employees individually
 *   - View employee stats: total, active, collar type distribution
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Table columns, filter options, search fields
 *   CAUTION: Excel import expects specific column format (see template)
 *   DO NOT modify: Import/export functionality without updating backend endpoints
 *
 * KEY COMPONENTS:
 *   - Table: Employee list with inline actions (view, edit, delete)
 *   - Search bar: Real-time filtering
 *   - Filter dropdowns: Status and collar type
 *   - Import dialog: Excel file upload with format guide
 *
 * API ENDPOINTS USED:
 *   - GET /api/employees (via employeeApi.getEmployees)
 *   - DELETE /api/employees/:id (via employeeApi.deleteEmployee)
 *   - GET /api/employees/template (via employeeApi.downloadTemplate)
 *   - POST /api/employees/import (via employeeApi.importEmployees)
 */

"use client"

import { useEffect, useState, useRef } from "react"
import { useRouter } from "next/navigation"
import {
  Plus, Search, Eye, Edit, Trash2, Users, UserCheck,
  UserX, Building2, Briefcase, DollarSign, Filter,
  Download, Upload, RefreshCw, ChevronDown, FileSpreadsheet,
  Info, AlertCircle, CheckCircle, X, Loader2, FileText, HelpCircle
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { employeeApi, Employee, ApiError, ImportEmployeesResponse, ImportError } from "@/lib/api-client"
import { canDeleteEmployees, canAddEmployees } from "@/lib/auth"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

export default function EmployeesPage() {
  const router = useRouter()
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [statusFilter, setStatusFilter] = useState("")
  const [collarFilter, setCollarFilter] = useState("")
  const [showFilters, setShowFilters] = useState(false)

  // Import dialog state
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false)
  const [isFormatGuideOpen, setIsFormatGuideOpen] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<ImportEmployeesResponse | null>(null)
  const [importError, setImportError] = useState<string | null>(null)
  const [downloadingTemplate, setDownloadingTemplate] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    loadEmployees()
  }, [])

  async function loadEmployees() {
    try {
      setLoading(true)
      setError(null)
      const response = await employeeApi.getEmployees()
      // Backend returns { employees: [...], total, page, page_size, total_pages }
      setEmployees(response.employees || [])
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load employees")
      }
      setEmployees([])
    } finally {
      setLoading(false)
    }
  }

  const filteredEmployees = employees.filter((emp) => {
    const matchesSearch =
      searchTerm === "" ||
      emp.full_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      emp.employee_number?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      emp.rfc?.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus =
      statusFilter === "" || emp.employment_status?.toLowerCase() === statusFilter.toLowerCase()
    const matchesCollar =
      collarFilter === "" || emp.collar_type === collarFilter
    return matchesSearch && matchesStatus && matchesCollar
  })

  // Statistics
  const stats = {
    total: employees.length,
    active: employees.filter(e => e.employment_status === "active").length,
    inactive: employees.filter(e => e.employment_status === "inactive").length,
    whiteCollar: employees.filter(e => e.collar_type === "white_collar").length,
    blueCollar: employees.filter(e => e.collar_type === "blue_collar").length,
    grayCollar: employees.filter(e => e.collar_type === "gray_collar").length,
    sindicalizado: employees.filter(e => e.is_sindicalizado).length,
    totalSalary: employees.reduce((sum, e) => sum + (e.daily_salary || 0), 0),
  }

  const handleViewEmployee = (id: string) => {
    router.push(`/employees/${id}`)
  }

  const handleEditEmployee = (id: string) => {
    router.push(`/employees/${id}/edit`)
  }

  const handleDeleteEmployee = async (id: string) => {
    if (!confirm("Are you sure you want to delete this employee?")) return
    try {
      await employeeApi.deleteEmployee(id)
      loadEmployees()
    } catch (err) {
      alert("Failed to delete employee")
    }
  }

  // Import functions
  const handleOpenImportDialog = () => {
    setImportResult(null)
    setImportError(null)
    setIsImportDialogOpen(true)
  }

  const handleDownloadTemplate = async () => {
    try {
      setDownloadingTemplate(true)
      const blob = await employeeApi.downloadTemplate()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = "employee_import_template.xlsx"
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (err) {
      setImportError("Error downloading template")
    } finally {
      setDownloadingTemplate(false)
    }
  }

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    const validExtensions = [".xlsx", ".xls", ".csv"]
    const extension = file.name.toLowerCase().substring(file.name.lastIndexOf("."))
    if (!validExtensions.includes(extension)) {
      setImportError("Invalid file format. Use .xlsx, .xls or .csv")
      return
    }

    try {
      setImporting(true)
      setImportError(null)
      setImportResult(null)
      const result = await employeeApi.importEmployees(file)
      setImportResult(result)
      if (result.imported > 0) {
        loadEmployees() // Refresh the list
      }
    } catch (err: any) {
      setImportError(err.message || "Error importing employees")
    } finally {
      setImporting(false)
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = ""
      }
    }
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-"
    try {
      return new Date(dateStr).toLocaleDateString("es-MX")
    } catch {
      return dateStr
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
    }).format(amount || 0)
  }

  const getCollarBadge = (collarType: string) => {
    switch (collarType) {
      case "white_collar":
        return "bg-gradient-to-r from-blue-500/20 to-cyan-500/20 text-blue-300 border border-blue-500/30"
      case "blue_collar":
        return "bg-gradient-to-r from-indigo-500/20 to-purple-500/20 text-indigo-300 border border-indigo-500/30"
      case "gray_collar":
        return "bg-gradient-to-r from-slate-500/20 to-zinc-500/20 text-slate-300 border border-slate-500/30"
      default:
        return "bg-slate-700/50 text-slate-400"
    }
  }

  const getCollarLabel = (collarType: string) => {
    switch (collarType) {
      case "white_collar": return "White Collar"
      case "blue_collar": return "Blue Collar"
      case "gray_collar": return "Gray Collar"
      default: return collarType
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return "bg-gradient-to-r from-emerald-500/20 to-green-500/20 text-emerald-300 border border-emerald-500/30"
      case "inactive":
        return "bg-gradient-to-r from-amber-500/20 to-yellow-500/20 text-amber-300 border border-amber-500/30"
      case "terminated":
        return "bg-gradient-to-r from-red-500/20 to-rose-500/20 text-red-300 border border-red-500/30"
      default:
        return "bg-slate-700/50 text-slate-400"
    }
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              Employees
            </h1>
            <p className="text-slate-400 mt-1">Manage employee information</p>
          </div>
          <div className="flex items-center gap-3">
            <Button
              onClick={loadEmployees}
              variant="outline"
              size="sm"
              className="border-slate-600 text-slate-300 hover:bg-slate-700"
            >
              <RefreshCw size={16} className="mr-2" />
              Refresh
            </Button>
            {canAddEmployees() && (
              <>
                <Button
                  onClick={handleOpenImportDialog}
                  variant="outline"
                  className="border-emerald-600 text-emerald-400 hover:bg-emerald-600/20"
                >
                  <Upload size={18} className="mr-2" />
                  Import Excel
                </Button>
                <Button
                  onClick={() => router.push("/employees/new")}
                  className="bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 shadow-lg shadow-blue-500/25"
                >
                  <Plus size={20} className="mr-2" />
                  New Employee
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/20 rounded-lg">
                <Users className="w-5 h-5 text-blue-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.total}</p>
                <p className="text-xs text-slate-400">Total Employees</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-500/20 rounded-lg">
                <UserCheck className="w-5 h-5 text-emerald-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.active}</p>
                <p className="text-xs text-slate-400">Active</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-500/20 rounded-lg">
                <Building2 className="w-5 h-5 text-indigo-400" />
              </div>
              <div>
                <p className="text-2xl font-bold text-white">{stats.sindicalizado}</p>
                <p className="text-xs text-slate-400">Unionized</p>
              </div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-500/20 rounded-lg">
                <DollarSign className="w-5 h-5 text-amber-400" />
              </div>
              <div>
                <p className="text-lg font-bold text-white">{formatCurrency(stats.totalSalary)}</p>
                <p className="text-xs text-slate-400">Total Daily Salary</p>
              </div>
            </div>
          </div>
        </div>

        {/* Collar Type Distribution */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
          <h3 className="text-sm font-medium text-slate-300 mb-3">Distribution by Collar Type</h3>
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center p-3 bg-blue-500/10 rounded-lg border border-blue-500/20">
              <p className="text-2xl font-bold text-blue-400">{stats.whiteCollar}</p>
              <p className="text-xs text-slate-400">White Collar</p>
              <p className="text-[10px] text-slate-500">Administrative</p>
            </div>
            <div className="text-center p-3 bg-indigo-500/10 rounded-lg border border-indigo-500/20">
              <p className="text-2xl font-bold text-indigo-400">{stats.blueCollar}</p>
              <p className="text-xs text-slate-400">Blue Collar</p>
              <p className="text-[10px] text-slate-500">Unionized</p>
            </div>
            <div className="text-center p-3 bg-slate-500/10 rounded-lg border border-slate-500/20">
              <p className="text-2xl font-bold text-slate-400">{stats.grayCollar}</p>
              <p className="text-xs text-slate-400">Gray Collar</p>
              <p className="text-[10px] text-slate-500">Non-Unionized</p>
            </div>
          </div>
        </div>

        {/* Search and Filters */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl p-4 border border-slate-700/50">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={20} />
              <input
                type="text"
                placeholder="Search by name, number or RFC..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2.5 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all"
              />
            </div>
            <div className="flex gap-3">
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="bg-slate-900/50 border border-slate-600 rounded-lg px-4 py-2 text-slate-200 focus:outline-none focus:border-blue-500 transition-all"
              >
                <option value="">All Statuses</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="terminated">Terminated</option>
              </select>
              <select
                value={collarFilter}
                onChange={(e) => setCollarFilter(e.target.value)}
                className="bg-slate-900/50 border border-slate-600 rounded-lg px-4 py-2 text-slate-200 focus:outline-none focus:border-blue-500 transition-all"
              >
                <option value="">All Types</option>
                <option value="white_collar">White Collar</option>
                <option value="blue_collar">Blue Collar</option>
                <option value="gray_collar">Gray Collar</option>
              </select>
            </div>
          </div>
          {(searchTerm || statusFilter || collarFilter) && (
            <div className="mt-3 flex items-center gap-2">
              <span className="text-xs text-slate-400">Active filters:</span>
              {searchTerm && (
                <span className="px-2 py-1 text-xs bg-blue-500/20 text-blue-300 rounded-full">
                  Search: {searchTerm}
                </span>
              )}
              {statusFilter && (
                <span className="px-2 py-1 text-xs bg-emerald-500/20 text-emerald-300 rounded-full">
                  Status: {statusFilter}
                </span>
              )}
              {collarFilter && (
                <span className="px-2 py-1 text-xs bg-indigo-500/20 text-indigo-300 rounded-full">
                  Type: {getCollarLabel(collarFilter)}
                </span>
              )}
              <button
                onClick={() => {
                  setSearchTerm("")
                  setStatusFilter("")
                  setCollarFilter("")
                }}
                className="px-2 py-1 text-xs text-red-400 hover:text-red-300 transition-colors"
              >
                Clear filters
              </button>
            </div>
          )}
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-900/20 border border-red-700/50 rounded-xl p-4 text-red-400 flex items-center justify-between">
            <span>{error}</span>
            <button onClick={loadEmployees} className="px-3 py-1 bg-red-500/20 hover:bg-red-500/30 rounded-lg transition-colors">
              Retry
            </button>
          </div>
        )}

        {/* Table */}
        <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-700/50 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/80 border-b border-slate-700/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Employee #
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    RFC
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Type
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Hire Date
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Daily Salary
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-4 text-center text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {loading ? (
                  <tr>
                    <td colSpan={8} className="px-6 py-12 text-center text-slate-400">
                      <div className="flex flex-col items-center gap-3">
                        <div className="w-8 h-8 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                        <span>Loading employees...</span>
                      </div>
                    </td>
                  </tr>
                ) : filteredEmployees.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="px-6 py-12 text-center text-slate-400">
                      {searchTerm || statusFilter || collarFilter ? (
                        <div className="flex flex-col items-center gap-3">
                          <Filter className="w-12 h-12 text-slate-600" />
                          <p>No employees match the filters</p>
                          <button
                            onClick={() => {
                              setSearchTerm("")
                              setStatusFilter("")
                              setCollarFilter("")
                            }}
                            className="text-blue-400 hover:text-blue-300 transition-colors"
                          >
                            Clear filters
                          </button>
                        </div>
                      ) : (
                        <div className="flex flex-col items-center gap-3">
                          <Users className="w-12 h-12 text-slate-600" />
                          <p>No employees registered</p>
                          <Button
                            onClick={() => router.push("/employees/new")}
                            className="bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700"
                          >
                            <Plus size={16} className="mr-2" />
                            Add first employee
                          </Button>
                        </div>
                      )}
                    </td>
                  </tr>
                ) : (
                  filteredEmployees.map((emp) => (
                    <tr
                      key={emp.id}
                      className="hover:bg-slate-700/30 transition-colors group"
                    >
                      <td className="px-6 py-4">
                        <span className="text-sm font-mono text-slate-300 bg-slate-800/50 px-2 py-1 rounded">
                          {emp.employee_number}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div>
                          <div className="text-sm text-white font-medium">{emp.full_name}</div>
                          <div className="text-xs text-slate-500 font-mono">{emp.curp}</div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-sm text-slate-300 font-mono">{emp.rfc}</span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col gap-1.5">
                          <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-lg ${getCollarBadge(emp.collar_type)}`}>
                            {getCollarLabel(emp.collar_type)}
                          </span>
                          <div className="flex items-center gap-2 text-xs text-slate-500">
                            <span className="capitalize">{emp.pay_frequency}</span>
                            {emp.is_sindicalizado && (
                              <span className="px-1.5 py-0.5 bg-indigo-500/20 text-indigo-300 rounded text-[10px]">
                                Unionized
                              </span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-300">{formatDate(emp.hire_date)}</td>
                      <td className="px-6 py-4">
                        <span className="text-sm font-semibold text-emerald-400">
                          {formatCurrency(emp.daily_salary)}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-lg capitalize ${getStatusBadge(emp.employment_status)}`}>
                          {emp.employment_status === "active" ? "Active" :
                           emp.employment_status === "inactive" ? "Inactive" :
                           emp.employment_status === "terminated" ? "Terminated" : emp.employment_status}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center justify-center gap-1 opacity-70 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => handleViewEmployee(emp.id)}
                            className="p-2 text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition-all"
                            title="View"
                          >
                            <Eye size={18} />
                          </button>
                          <button
                            onClick={() => handleEditEmployee(emp.id)}
                            className="p-2 text-slate-400 hover:text-amber-400 hover:bg-amber-500/10 rounded-lg transition-all"
                            title="Edit"
                          >
                            <Edit size={18} />
                          </button>
                          {canDeleteEmployees() && (
                            <button
                              onClick={() => handleDeleteEmployee(emp.id)}
                              className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                              title="Delete"
                            >
                              <Trash2 size={18} />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Footer Summary */}
        {!loading && filteredEmployees.length > 0 && (
          <div className="flex items-center justify-between text-sm text-slate-400 px-2">
            <span>
              Showing {filteredEmployees.length} of {employees.length} employees
            </span>
            <div className="flex items-center gap-4">
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-emerald-400 rounded-full"></span>
                {stats.active} active
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2 h-2 bg-amber-400 rounded-full"></span>
                {stats.inactive} inactive
              </span>
            </div>
          </div>
        )}

        {/* Import Dialog */}
        {isImportDialogOpen && (
        <Dialog open={isImportDialogOpen} onOpenChange={setIsImportDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-xl">
                <FileSpreadsheet className="h-6 w-6 text-emerald-400" />
                Import Employees from Excel
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                Bulk upload employees using an Excel (.xlsx, .xls) or CSV file
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6 py-4">
              {/* Instructions */}
              <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <Info className="h-5 w-5 text-blue-400 mt-0.5 flex-shrink-0" />
                  <div className="space-y-2 text-sm">
                    <p className="text-blue-300 font-medium">Instructions:</p>
                    <ol className="list-decimal list-inside space-y-1 text-slate-300">
                      <li>Download the Excel template with the correct format</li>
                      <li>Fill in employee data following the format</li>
                      <li>Select the file and upload employees</li>
                    </ol>
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div className="grid grid-cols-2 gap-4">
                {/* Download Template */}
                <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4 space-y-3">
                  <h3 className="font-medium text-slate-200 flex items-center gap-2">
                    <Download className="h-4 w-4 text-slate-400" />
                    Step 1: Download Template
                  </h3>
                  <p className="text-xs text-slate-400">
                    The template includes all required columns and sample data.
                  </p>
                  <Button
                    onClick={handleDownloadTemplate}
                    disabled={downloadingTemplate}
                    variant="outline"
                    className="w-full border-emerald-600 text-emerald-400 hover:bg-emerald-600/20"
                  >
                    {downloadingTemplate ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Downloading...
                      </>
                    ) : (
                      <>
                        <Download className="h-4 w-4 mr-2" />
                        Download Template
                      </>
                    )}
                  </Button>
                </div>

                {/* Upload File */}
                <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4 space-y-3">
                  <h3 className="font-medium text-slate-200 flex items-center gap-2">
                    <Upload className="h-4 w-4 text-slate-400" />
                    Step 2: Upload File
                  </h3>
                  <p className="text-xs text-slate-400">
                    Upload the Excel file with employee data.
                  </p>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".xlsx,.xls,.csv"
                    onChange={handleFileSelect}
                    className="hidden"
                  />
                  <Button
                    onClick={() => fileInputRef.current?.click()}
                    disabled={importing}
                    className="w-full bg-emerald-600 hover:bg-emerald-700"
                  >
                    {importing ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Importing...
                      </>
                    ) : (
                      <>
                        <Upload className="h-4 w-4 mr-2" />
                        Select File
                      </>
                    )}
                  </Button>
                </div>
              </div>

              {/* Format Guide Button */}
              <div className="flex justify-center">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsFormatGuideOpen(true)}
                  className="text-slate-400 hover:text-slate-200"
                >
                  <HelpCircle className="h-4 w-4 mr-2" />
                  View Data Format Guide
                </Button>
              </div>

              {/* Import Error */}
              {importError && (
                <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="h-5 w-5 text-red-400 mt-0.5 flex-shrink-0" />
                    <div>
                      <p className="text-red-300 font-medium">Import error</p>
                      <p className="text-sm text-red-400 mt-1">{importError}</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Import Result */}
              {importResult && (
                <div className={`rounded-lg p-4 border ${
                  importResult.failed > 0
                    ? "bg-amber-500/10 border-amber-500/30"
                    : "bg-emerald-500/10 border-emerald-500/30"
                }`}>
                  <div className="flex items-start gap-3">
                    {importResult.failed > 0 ? (
                      <AlertCircle className="h-5 w-5 text-amber-400 mt-0.5 flex-shrink-0" />
                    ) : (
                      <CheckCircle className="h-5 w-5 text-emerald-400 mt-0.5 flex-shrink-0" />
                    )}
                    <div className="flex-1">
                      <p className={`font-medium ${importResult.failed > 0 ? "text-amber-300" : "text-emerald-300"}`}>
                        Import Result
                      </p>
                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-3 text-sm">
                        <div className="text-center p-2 bg-slate-800/50 rounded-lg">
                          <p className="text-2xl font-bold text-white">{importResult.total_rows || 0}</p>
                          <p className="text-xs text-slate-400">Total Rows</p>
                        </div>
                        <div className="text-center p-2 bg-emerald-500/10 rounded-lg">
                          <p className="text-2xl font-bold text-emerald-400">{importResult.created || 0}</p>
                          <p className="text-xs text-slate-400">Created</p>
                        </div>
                        <div className="text-center p-2 bg-blue-500/10 rounded-lg">
                          <p className="text-2xl font-bold text-blue-400">{importResult.updated || 0}</p>
                          <p className="text-xs text-slate-400">Updated</p>
                        </div>
                        <div className="text-center p-2 bg-red-500/10 rounded-lg">
                          <p className="text-2xl font-bold text-red-400">{importResult.failed || 0}</p>
                          <p className="text-xs text-slate-400">Errors</p>
                        </div>
                      </div>
                      {importResult.errors && importResult.errors.length > 0 && (
                        <div className="mt-3">
                          <p className="text-xs text-amber-400 font-medium mb-2">Errors found:</p>
                          <div className="max-h-40 overflow-y-auto text-xs space-y-1">
                            {importResult.errors.map((err, idx) => {
                              const errorObj = err as ImportError
                              const isObject = typeof err === 'object' && err !== null
                              return (
                                <div key={idx} className="text-red-400 bg-red-500/10 px-2 py-1 rounded flex items-start gap-2">
                                  {isObject && errorObj.row && (
                                    <span className="bg-red-500/20 px-1.5 py-0.5 rounded text-red-300 font-mono whitespace-nowrap">
                                      Row {errorObj.row}
                                    </span>
                                  )}
                                  <span>
                                    {isObject
                                      ? (errorObj.error || JSON.stringify(err))
                                      : String(err)}
                                  </span>
                                </div>
                              )
                            })}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsImportDialogOpen(false)}
                className="border-slate-600"
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        )}

        {/* Format Guide Dialog */}
        {isFormatGuideOpen && (
        <Dialog open={isFormatGuideOpen} onOpenChange={setIsFormatGuideOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-4xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-xl">
                <FileText className="h-6 w-6 text-blue-400" />
                Data Format Guide
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                Description of each column and required format for the import file
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6 py-4">
              {/* Required Fields */}
              <div>
                <h3 className="text-sm font-semibold text-emerald-400 mb-3 flex items-center gap-2">
                  <CheckCircle className="h-4 w-4" />
                  Required Fields
                </h3>
                <div className="bg-slate-800/50 rounded-lg border border-slate-700 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-slate-800 border-b border-slate-700">
                      <tr>
                        <th className="px-4 py-2 text-left text-slate-300">Column</th>
                        <th className="px-4 py-2 text-left text-slate-300">Description</th>
                        <th className="px-4 py-2 text-left text-slate-300">Format / Example</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700">
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">employee_number</td>
                        <td className="px-4 py-2 text-slate-300">Employee number (unique)</td>
                        <td className="px-4 py-2 text-slate-400">EMP001, EMP002</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">first_name</td>
                        <td className="px-4 py-2 text-slate-300">First name(s)</td>
                        <td className="px-4 py-2 text-slate-400">Juan Carlos</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">last_name</td>
                        <td className="px-4 py-2 text-slate-300">Paternal last name</td>
                        <td className="px-4 py-2 text-slate-400">Garcia</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">date_of_birth</td>
                        <td className="px-4 py-2 text-slate-300">Date of birth</td>
                        <td className="px-4 py-2 text-slate-400">1990-05-15 (YYYY-MM-DD)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">gender</td>
                        <td className="px-4 py-2 text-slate-300">Gender</td>
                        <td className="px-4 py-2 text-slate-400">male, female, other</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">rfc</td>
                        <td className="px-4 py-2 text-slate-300">RFC with homonymy</td>
                        <td className="px-4 py-2 text-slate-400">GARP900515ABC (13 characters)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">curp</td>
                        <td className="px-4 py-2 text-slate-300">CURP</td>
                        <td className="px-4 py-2 text-slate-400">GARP900515HDFRRN09 (18 characters)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">hire_date</td>
                        <td className="px-4 py-2 text-slate-300">Hire date</td>
                        <td className="px-4 py-2 text-slate-400">2024-01-15 (YYYY-MM-DD)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">daily_salary</td>
                        <td className="px-4 py-2 text-slate-300">Daily salary</td>
                        <td className="px-4 py-2 text-slate-400">500.00 (decimal number)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-emerald-400">collar_type</td>
                        <td className="px-4 py-2 text-slate-300">Collar type</td>
                        <td className="px-4 py-2 text-slate-400">white_collar, blue_collar, gray_collar</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Optional Fields */}
              <div>
                <h3 className="text-sm font-semibold text-blue-400 mb-3 flex items-center gap-2">
                  <Info className="h-4 w-4" />
                  Optional Fields
                </h3>
                <div className="bg-slate-800/50 rounded-lg border border-slate-700 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-slate-800 border-b border-slate-700">
                      <tr>
                        <th className="px-4 py-2 text-left text-slate-300">Column</th>
                        <th className="px-4 py-2 text-left text-slate-300">Description</th>
                        <th className="px-4 py-2 text-left text-slate-300">Format / Example</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700">
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">mother_last_name</td>
                        <td className="px-4 py-2 text-slate-300">Maternal last name</td>
                        <td className="px-4 py-2 text-slate-400">Lopez</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">nss</td>
                        <td className="px-4 py-2 text-slate-300">Social Security Number</td>
                        <td className="px-4 py-2 text-slate-400">12345678901 (11 digits)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">personal_email</td>
                        <td className="px-4 py-2 text-slate-300">Personal email</td>
                        <td className="px-4 py-2 text-slate-400">juan@email.com</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">personal_phone</td>
                        <td className="px-4 py-2 text-slate-300">Personal phone</td>
                        <td className="px-4 py-2 text-slate-400">5512345678</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">bank_name</td>
                        <td className="px-4 py-2 text-slate-300">Bank name</td>
                        <td className="px-4 py-2 text-slate-400">BBVA, Santander, etc.</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">bank_account</td>
                        <td className="px-4 py-2 text-slate-300">Account number</td>
                        <td className="px-4 py-2 text-slate-400">0123456789</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">clabe</td>
                        <td className="px-4 py-2 text-slate-300">Interbank CLABE</td>
                        <td className="px-4 py-2 text-slate-400">012345678901234567 (18 digits)</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">is_sindicalizado</td>
                        <td className="px-4 py-2 text-slate-300">Is unionized</td>
                        <td className="px-4 py-2 text-slate-400">true, false, 1, 0</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">employment_status</td>
                        <td className="px-4 py-2 text-slate-300">Employment status</td>
                        <td className="px-4 py-2 text-slate-400">active, inactive, terminated</td>
                      </tr>
                      <tr className="hover:bg-slate-700/30">
                        <td className="px-4 py-2 font-mono text-blue-400">pay_frequency</td>
                        <td className="px-4 py-2 text-slate-300">Pay frequency</td>
                        <td className="px-4 py-2 text-slate-400">weekly, biweekly, monthly</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Collar Types Explanation */}
              <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
                <h3 className="text-sm font-semibold text-slate-200 mb-3">Collar Types</h3>
                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div className="p-3 bg-blue-500/10 rounded-lg border border-blue-500/20">
                    <p className="font-mono text-blue-400">white_collar</p>
                    <p className="text-slate-400 text-xs mt-1">Administrative staff, biweekly pay</p>
                  </div>
                  <div className="p-3 bg-indigo-500/10 rounded-lg border border-indigo-500/20">
                    <p className="font-mono text-indigo-400">blue_collar</p>
                    <p className="text-slate-400 text-xs mt-1">Unionized staff, weekly pay</p>
                  </div>
                  <div className="p-3 bg-slate-500/10 rounded-lg border border-slate-500/20">
                    <p className="font-mono text-slate-400">gray_collar</p>
                    <p className="text-slate-400 text-xs mt-1">Non-unionized operational staff, weekly pay</p>
                  </div>
                </div>
              </div>

              {/* Important Notes */}
              <div className="bg-amber-500/10 border border-amber-500/30 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-amber-400 mb-2 flex items-center gap-2">
                  <AlertCircle className="h-4 w-4" />
                  Important Notes
                </h3>
                <ul className="list-disc list-inside space-y-1 text-sm text-slate-300">
                  <li>RFC and CURP must be valid for Mexico</li>
                  <li>Dates must be in YYYY-MM-DD format</li>
                  <li>Employee number must be unique</li>
                  <li>Salaries must be positive numbers</li>
                  <li>The first row of the file must contain the column names</li>
                </ul>
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsFormatGuideOpen(false)}
                className="border-slate-600"
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        )}
      </div>
    </DashboardLayout>
  )
}
