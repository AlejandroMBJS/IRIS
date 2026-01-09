/**
 * @file app/incidences/page.tsx
 * @description Incidence management page for tracking absences, overtime, bonuses, and deductions
 *
 * USER PERSPECTIVE:
 *   - Create incidences for one employee, multiple employees, or all employees
 *   - Filter incidences by period (week), status, and employee
 *   - Approve or reject pending incidences
 *   - Attach evidence files to incidences
 *   - View employee details by clicking on their name
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Incidence types, filter options, bulk creation logic
 *   CAUTION: Period selection defaults to current open period
 *   DO NOT modify: Evidence upload without updating file storage backend
 *
 * REFACTORED STRUCTURE:
 *   - IncidenceStats: Summary cards for incidence counts
 *   - IncidenceFilters: Period and status/employee filters
 *   - IncidenceTable: Main data table with actions
 *   - CreateIncidenceDialog: Modal for creating new incidences
 *   - EvidenceDialog: Modal for managing incidence evidence
 *   - EmployeeInfoDialog: Modal for viewing employee details
 */

"use client"

import { useState, useEffect } from "react"
import { ClipboardList, FileDown, UserPlus, Users } from "lucide-react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import {
  incidenceApi,
  incidenceTypeApi,
  employeeApi,
  payrollApi,
  evidenceApi,
  absenceRequestApi,
  Incidence,
  IncidenceType,
  IncidenceEvidence,
  Employee,
  PayrollPeriod,
  CreateIncidenceRequest,
} from "@/lib/api-client"
import {
  SelectionMode,
  IncidenceStats,
  IncidenceFilters,
  IncidenceTable,
  CreateIncidenceDialog,
  EvidenceDialog,
  EmployeeInfoDialog,
} from "@/components/incidences"

export default function IncidencesPage() {
  // Data state
  const [incidences, setIncidences] = useState<Incidence[]>([])
  const [incidenceTypes, setIncidenceTypes] = useState<IncidenceType[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [periods, setPeriods] = useState<PayrollPeriod[]>([])

  // UI state
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [successMessage, setSuccessMessage] = useState("")

  // Filter state
  const [filterStatus, setFilterStatus] = useState<string>("")
  const [filterEmployee, setFilterEmployee] = useState<string>("")
  const [filterPeriod, setFilterPeriod] = useState<string>("")

  // Create dialog state
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [selectionMode, setSelectionMode] = useState<SelectionMode>("single")
  const [saving, setSaving] = useState(false)
  const [savingProgress, setSavingProgress] = useState({ current: 0, total: 0 })
  const [uploadingFiles, setUploadingFiles] = useState(false)

  // Evidence dialog state
  const [isEvidenceDialogOpen, setIsEvidenceDialogOpen] = useState(false)
  const [selectedIncidence, setSelectedIncidence] = useState<Incidence | null>(null)
  const [evidences, setEvidences] = useState<IncidenceEvidence[]>([])
  const [loadingEvidences, setLoadingEvidences] = useState(false)

  // Employee info dialog state
  const [isEmployeeDialogOpen, setIsEmployeeDialogOpen] = useState(false)
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(null)

  // Load initial data
  useEffect(() => {
    loadData()
  }, [])

  // Reload incidences when filters change
  useEffect(() => {
    loadIncidences()
  }, [filterStatus, filterEmployee, filterPeriod])

  // Clear success message after 5 seconds
  useEffect(() => {
    if (successMessage) {
      const timer = setTimeout(() => setSuccessMessage(""), 5000)
      return () => clearTimeout(timer)
    }
  }, [successMessage])

  const loadData = async () => {
    try {
      setLoading(true)
      const [types, empsResponse, prds] = await Promise.all([
        incidenceTypeApi.getAll(),
        employeeApi.getEmployees(),
        payrollApi.getPeriods(),
      ])
      setIncidenceTypes(types)
      const emps = empsResponse.employees || []
      setEmployees(emps.filter(e => e.employment_status === "active"))
      const sortedPeriods = prds.sort((a, b) =>
        new Date(b.start_date).getTime() - new Date(a.start_date).getTime()
      )
      setPeriods(sortedPeriods)
      const openPeriod = sortedPeriods.find(p => p.status === "open")
      if (openPeriod) {
        setFilterPeriod(openPeriod.id)
      }
      await loadIncidences()
    } catch (err: any) {
      setError(err.message || "Error loading data")
    } finally {
      setLoading(false)
    }
  }

  const loadIncidences = async () => {
    try {
      const data = await incidenceApi.getAll(
        filterEmployee && filterEmployee !== "all" ? filterEmployee : undefined,
        filterPeriod && filterPeriod !== "all" ? filterPeriod : undefined,
        filterStatus && filterStatus !== "all" ? filterStatus : undefined
      )
      setIncidences(data)
      setError("")
    } catch (err: any) {
      setError(err.message || "Error loading incidences")
    }
  }

  const exportToExcel = async () => {
    try {
      const blob = await absenceRequestApi.exportApproved({
        period_id: filterPeriod && filterPeriod !== "all" ? filterPeriod : undefined,
        employee_id: filterEmployee && filterEmployee !== "all" ? filterEmployee : undefined,
      })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `approved_requests_${new Date().toISOString().split('T')[0]}.xlsx`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      setSuccessMessage("Excel file exported successfully")
    } catch (err: any) {
      setError(err.message || "Error exporting to Excel")
    }
  }

  const handleOpenCreateDialog = (mode: SelectionMode) => {
    setSelectionMode(mode)
    setIsCreateDialogOpen(true)
  }

  const handleSaveIncidence = async (
    formData: CreateIncidenceRequest,
    targetEmployees: string[],
    pendingFiles: File[]
  ) => {
    try {
      setSaving(true)
      setSavingProgress({ current: 0, total: targetEmployees.length })
      setError("")

      let successCount = 0
      let errorCount = 0
      const createdIncidenceIds: string[] = []

      for (let i = 0; i < targetEmployees.length; i++) {
        const employeeId = targetEmployees[i]
        setSavingProgress({ current: i + 1, total: targetEmployees.length })

        try {
          const response = await incidenceApi.create({
            ...formData,
            employee_id: employeeId,
          })
          successCount++
          if (response && response.id) {
            createdIncidenceIds.push(response.id)
          }
        } catch (err) {
          errorCount++
          console.error(`Error creating incidence for employee ${employeeId}:`, err)
        }
      }

      // Upload pending files if we have any and only one incidence was created
      if (pendingFiles.length > 0 && createdIncidenceIds.length === 1) {
        setUploadingFiles(true)
        const incidenceId = createdIncidenceIds[0]
        for (const file of pendingFiles) {
          try {
            await evidenceApi.upload(incidenceId, file)
          } catch (err) {
            console.error(`Error uploading file ${file.name}:`, err)
          }
        }
        setUploadingFiles(false)
      }

      await loadIncidences()
      setIsCreateDialogOpen(false)

      if (errorCount > 0) {
        setSuccessMessage(`${successCount} incidences created. ${errorCount} failed.`)
      } else {
        const fileMsg = pendingFiles.length > 0 && createdIncidenceIds.length === 1
          ? ` with ${pendingFiles.length} attached file(s)`
          : ""
        setSuccessMessage(`${successCount} incidence(s) created successfully${fileMsg}.`)
      }
    } catch (err: any) {
      setError(err.message || "Error saving incidence")
    } finally {
      setSaving(false)
      setUploadingFiles(false)
      setSavingProgress({ current: 0, total: 0 })
    }
  }

  const handleApprove = async (id: string) => {
    try {
      await incidenceApi.approve(id)
      await loadIncidences()
    } catch (err: any) {
      setError(err.message || "Error approving incidence")
    }
  }

  const handleReject = async (id: string) => {
    try {
      await incidenceApi.reject(id)
      await loadIncidences()
    } catch (err: any) {
      setError(err.message || "Error rejecting incidence")
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this incidence?")) return
    try {
      await incidenceApi.delete(id)
      await loadIncidences()
    } catch (err: any) {
      setError(err.message || "Error deleting incidence")
    }
  }

  // Evidence functions
  const openEvidenceDialog = async (incidence: Incidence) => {
    setSelectedIncidence(incidence)
    setIsEvidenceDialogOpen(true)
    await loadEvidences(incidence.id)
  }

  const loadEvidences = async (incidenceId: string) => {
    try {
      setLoadingEvidences(true)
      const data = await evidenceApi.list(incidenceId)
      setEvidences(data)
    } catch (err: any) {
      console.error("Error loading evidences:", err)
      setEvidences([])
    } finally {
      setLoadingEvidences(false)
    }
  }

  const handleEvidenceUpload = async (file: File) => {
    if (!selectedIncidence) return
    await evidenceApi.upload(selectedIncidence.id, file)
    await loadEvidences(selectedIncidence.id)
  }

  const handleEvidenceDownload = async (evidenceId: string, fileName: string) => {
    try {
      const blob = await evidenceApi.download(evidenceId)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = fileName
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (err: any) {
      setError(err.message || "Error downloading file")
    }
  }

  const handleEvidenceDelete = async (evidenceId: string) => {
    if (!selectedIncidence) return
    if (!confirm("Are you sure you want to delete this evidence?")) return
    try {
      await evidenceApi.delete(evidenceId)
      await loadEvidences(selectedIncidence.id)
    } catch (err: any) {
      setError(err.message || "Error deleting evidence")
    }
  }

  const handleClearFilters = () => {
    setFilterStatus("")
    setFilterEmployee("")
    setFilterPeriod("")
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-white flex items-center gap-2">
              <ClipboardList className="h-6 w-6" />
              Incidences
            </h1>
            <p className="text-slate-400 mt-1">
              Manage absences, vacations, overtime and more
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              onClick={exportToExcel}
              variant="outline"
              className="border-emerald-600 text-emerald-400 hover:bg-emerald-600/20"
              title="Export approved incidences to Excel"
            >
              <FileDown className="h-4 w-4 sm:mr-2" />
              <span className="hidden sm:inline">Export</span>
            </Button>
            <Button
              onClick={() => handleOpenCreateDialog("single")}
              className="bg-blue-600 hover:bg-blue-700"
            >
              <UserPlus className="h-4 w-4 sm:mr-2" />
              <span className="hidden sm:inline">One Employee</span>
            </Button>
            <Button
              onClick={() => handleOpenCreateDialog("multiple")}
              variant="outline"
              className="border-blue-600 text-blue-400 hover:bg-blue-600/20"
            >
              <Users className="h-4 w-4 sm:mr-2" />
              <span className="hidden sm:inline">Multiple</span>
            </Button>
            <Button
              onClick={() => handleOpenCreateDialog("all")}
              variant="outline"
              className="border-green-600 text-green-400 hover:bg-green-600/20"
            >
              <Users className="h-4 w-4 sm:mr-2" />
              <span className="hidden sm:inline">All</span>
            </Button>
          </div>
        </div>

        {/* Stats Summary */}
        <IncidenceStats incidences={incidences} />

        {/* Filters */}
        <IncidenceFilters
          periods={periods}
          employees={employees}
          filterPeriod={filterPeriod}
          filterStatus={filterStatus}
          filterEmployee={filterEmployee}
          onPeriodChange={setFilterPeriod}
          onStatusChange={setFilterStatus}
          onEmployeeChange={setFilterEmployee}
          onClearFilters={handleClearFilters}
        />

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
        <IncidenceTable
          incidences={incidences}
          loading={loading}
          onApprove={handleApprove}
          onReject={handleReject}
          onDelete={handleDelete}
          onOpenEvidence={openEvidenceDialog}
          onOpenEmployeeInfo={(employee) => {
            setSelectedEmployee(employee)
            setIsEmployeeDialogOpen(true)
          }}
        />

        {/* Create Dialog */}
        <CreateIncidenceDialog
          isOpen={isCreateDialogOpen}
          onOpenChange={setIsCreateDialogOpen}
          selectionMode={selectionMode}
          employees={employees}
          incidenceTypes={incidenceTypes}
          periods={periods}
          onSave={handleSaveIncidence}
          saving={saving}
          savingProgress={savingProgress}
          uploadingFiles={uploadingFiles}
        />

        {/* Evidence Dialog */}
        <EvidenceDialog
          isOpen={isEvidenceDialogOpen}
          onOpenChange={setIsEvidenceDialogOpen}
          incidence={selectedIncidence}
          evidences={evidences}
          loading={loadingEvidences}
          onUpload={handleEvidenceUpload}
          onDownload={handleEvidenceDownload}
          onDelete={handleEvidenceDelete}
        />

        {/* Employee Info Dialog */}
        <EmployeeInfoDialog
          isOpen={isEmployeeDialogOpen}
          onOpenChange={setIsEmployeeDialogOpen}
          employee={selectedEmployee}
        />
      </div>
    </DashboardLayout>
  )
}
