"use client"

import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { shiftApi, Shift, employeeApi, Employee } from "@/lib/api-client"
import { Search } from "lucide-react"

interface EmploymentFormProps {
  data: Record<string, unknown>
  onChange: (data: Record<string, unknown>) => void
  shifts?: Shift[]  // Optional: pre-loaded shifts
  employees?: Employee[]  // Optional: pre-loaded employees for supervisor
}

const EMPLOYMENT_STATUS_OPTIONS = [
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
  { value: "on_leave", label: "On Leave" },
  { value: "terminated", label: "Terminated" },
]

const EMPLOYEE_TYPE_OPTIONS = [
  { value: "", label: "Select..." },
  { value: "permanent", label: "Permanent" },
  { value: "temporary", label: "Temporary" },
  { value: "contractor", label: "Contractor" },
  { value: "intern", label: "Intern" },
]

const COLLAR_TYPE_OPTIONS = [
  { value: "white_collar", label: "White Collar (Administrative)" },
  { value: "blue_collar", label: "Blue Collar (Unionized Worker)" },
  { value: "gray_collar", label: "Gray Collar (Non-Unionized Worker)" },
]

const PAY_FREQUENCY_OPTIONS = [
  { value: "weekly", label: "Weekly" },
  { value: "biweekly", label: "Biweekly" },
  { value: "monthly", label: "Monthly" },
]

const PAYMENT_METHOD_OPTIONS = [
  { value: "bank_transfer", label: "Bank Transfer" },
  { value: "check", label: "Check" },
  { value: "cash", label: "Cash" },
]

const REGIME_OPTIONS = [
  { value: "salary", label: "Salary and Wages" },
  { value: "assimilated", label: "Assimilated to Salary" },
  { value: "fees", label: "Fees" },
]

const TAX_REGIME_OPTIONS = [
  { value: "", label: "Select..." },
  { value: "601", label: "601 - General Law for Legal Entities" },
  { value: "603", label: "603 - Legal Entities with Non-Profit Purposes" },
  { value: "605", label: "605 - Salary and Wages" },
  { value: "606", label: "606 - Leasing" },
  { value: "612", label: "612 - Individuals with Business Activities" },
  { value: "621", label: "621 - Tax Incorporation" },
  { value: "625", label: "625 - Business Activities Regime" },
  { value: "626", label: "626 - Simplified Trust Regime" },
]

export function EmploymentForm({ data, onChange, shifts: propShifts, employees: propEmployees }: EmploymentFormProps) {
  const [shifts, setShifts] = useState<Shift[]>(propShifts || [])
  const [loadingShifts, setLoadingShifts] = useState(!propShifts)
  const [employees, setEmployees] = useState<Employee[]>(propEmployees || [])
  const [loadingEmployees, setLoadingEmployees] = useState(!propEmployees)
  const [supervisorSearch, setSupervisorSearch] = useState("")

  useEffect(() => {
    if (!propShifts) {
      shiftApi.getActive()
        .then((data) => setShifts(data || []))
        .catch(console.error)
        .finally(() => setLoadingShifts(false))
    }
  }, [propShifts])

  useEffect(() => {
    if (!propEmployees) {
      employeeApi.getEmployees()
        .then((response) => {
          const emps = response?.employees || []
          setEmployees(emps.filter(e => e.employment_status === "active"))
        })
        .catch(console.error)
        .finally(() => setLoadingEmployees(false))
    }
  }, [propEmployees])

  // Filter employees for supervisor selector
  const filteredSupervisors = (employees || []).filter(emp => {
    if (!supervisorSearch) return true
    const search = supervisorSearch.toLowerCase()
    return (
      emp.first_name?.toLowerCase().includes(search) ||
      emp.last_name?.toLowerCase().includes(search) ||
      emp.employee_number?.toLowerCase().includes(search) ||
      `${emp.first_name} ${emp.last_name}`.toLowerCase().includes(search)
    )
  })

  const handleChange = (field: string, value: unknown) => {
    onChange({ [field]: value })
  }

  const handleCollarTypeChange = (value: string) => {
    handleChange("collarType", value)
    // Auto-set pay frequency and sindicalizado based on collar type
    if (value === "white_collar") {
      handleChange("payFrequency", "biweekly")
      handleChange("isSindicalizado", false)
    } else if (value === "blue_collar") {
      handleChange("payFrequency", "weekly")
      handleChange("isSindicalizado", true)
    } else if (value === "gray_collar") {
      handleChange("payFrequency", "weekly")
      handleChange("isSindicalizado", false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Employment Dates */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Employment Dates</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <Label htmlFor="hireDate" className="text-slate-300">
              Hire Date <span className="text-red-400">*</span>
            </Label>
            <Input
              id="hireDate"
              type="date"
              value={(data.hireDate as string) || ""}
              onChange={(e) => handleChange("hireDate", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="terminationDate" className="text-slate-300">Termination Date</Label>
            <Input
              id="terminationDate"
              type="date"
              value={(data.terminationDate as string) || ""}
              onChange={(e) => handleChange("terminationDate", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="contractStartDate" className="text-slate-300">Contract Start Date</Label>
            <Input
              id="contractStartDate"
              type="date"
              value={(data.contractStartDate as string) || ""}
              onChange={(e) => handleChange("contractStartDate", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="contractEndDate" className="text-slate-300">Contract End Date</Label>
            <Input
              id="contractEndDate"
              type="date"
              value={(data.contractEndDate as string) || ""}
              onChange={(e) => handleChange("contractEndDate", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Employment Status */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Employment Status</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <Label htmlFor="employmentStatus" className="text-slate-300">Status</Label>
            <select
              id="employmentStatus"
              value={(data.employmentStatus as string) || "active"}
              onChange={(e) => handleChange("employmentStatus", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {EMPLOYMENT_STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="employeeType" className="text-slate-300">Employee Type</Label>
            <select
              id="employeeType"
              value={(data.employeeType as string) || ""}
              onChange={(e) => handleChange("employeeType", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {EMPLOYEE_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="collarType" className="text-slate-300">Collar Type</Label>
            <select
              id="collarType"
              value={(data.collarType as string) || "white_collar"}
              onChange={(e) => handleCollarTypeChange(e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {COLLAR_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="payFrequency" className="text-slate-300">Pay Frequency</Label>
            <select
              id="payFrequency"
              value={(data.payFrequency as string) || "biweekly"}
              onChange={(e) => handleChange("payFrequency", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {PAY_FREQUENCY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="mt-4 flex items-center gap-2">
          <input
            type="checkbox"
            id="isSindicalizado"
            className="h-4 w-4 rounded border-slate-700 bg-slate-800"
            checked={Boolean(data.isSindicalizado)}
            onChange={(e) => handleChange("isSindicalizado", e.target.checked)}
          />
          <Label htmlFor="isSindicalizado" className="text-slate-300">Unionized</Label>
        </div>
      </div>

      {/* Organization */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Organization</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="departmentId" className="text-slate-300">Department</Label>
            <Input
              id="departmentId"
              placeholder="Department ID"
              value={(data.departmentId as string) || ""}
              onChange={(e) => handleChange("departmentId", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="positionId" className="text-slate-300">Position</Label>
            <Input
              id="positionId"
              placeholder="Position ID"
              value={(data.positionId as string) || ""}
              onChange={(e) => handleChange("positionId", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="costCenterId" className="text-slate-300">Cost Center</Label>
            <Input
              id="costCenterId"
              placeholder="Cost center ID"
              value={(data.costCenterId as string) || ""}
              onChange={(e) => handleChange("costCenterId", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="productionArea" className="text-slate-300">Production Area</Label>
            <Input
              id="productionArea"
              placeholder="Production area"
              value={(data.productionArea as string) || ""}
              onChange={(e) => handleChange("productionArea", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="location" className="text-slate-300">Location</Label>
            <Input
              id="location"
              placeholder="North Plant"
              value={(data.location as string) || ""}
              onChange={(e) => handleChange("location", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="teamName" className="text-slate-300">Team</Label>
            <Input
              id="teamName"
              placeholder="Team name"
              value={(data.teamName as string) || ""}
              onChange={(e) => handleChange("teamName", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Company & Registration */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Company & Registration</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="companyName" className="text-slate-300">Company Name</Label>
            <Input
              id="companyName"
              placeholder="Company name"
              value={(data.companyName as string) || ""}
              onChange={(e) => handleChange("companyName", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="patronalRegistry" className="text-slate-300">Employer Registration</Label>
            <Input
              id="patronalRegistry"
              placeholder="IMSS employer registration"
              value={(data.patronalRegistry as string) || ""}
              onChange={(e) => handleChange("patronalRegistry", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="packageCode" className="text-slate-300">Package Code</Label>
            <Input
              id="packageCode"
              placeholder="PKG-001"
              value={(data.packageCode as string) || ""}
              onChange={(e) => handleChange("packageCode", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Shift & Supervisor */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Shift & Supervisor</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <Label htmlFor="shiftId" className="text-slate-300">Shift</Label>
            <select
              id="shiftId"
              value={(data.shiftId as string) || ""}
              onChange={(e) => handleChange("shiftId", e.target.value || null)}
              disabled={loadingShifts}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white disabled:opacity-50"
            >
              <option value="">{loadingShifts ? "Loading shifts..." : "Select shift..."}</option>
              {(shifts || []).map((shift) => (
                <option key={shift.id} value={shift.id}>
                  {shift.name} ({shift.code}) - {shift.start_time} to {shift.end_time}
                </option>
              ))}
            </select>
            {(() => {
              const shiftIdStr = data.shiftId as string
              if ((shifts || []).length > 0 && shiftIdStr) {
                const selectedShift = (shifts || []).find(s => s.id === shiftIdStr)
                if (selectedShift) {
                  return (
                    <p className="text-xs text-slate-500 mt-1">
                      {selectedShift.work_hours_per_day}h/day - {selectedShift.is_night_shift ? 'Night shift' : 'Day shift'}
                    </p>
                  )
                }
              }
              return null
            })()}
          </div>
          <div>
            <Label htmlFor="supervisorId" className="text-slate-300">Supervisor</Label>
            <div className="relative">
              <div className="relative mb-2">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-400" />
                <Input
                  placeholder="Search supervisor..."
                  value={supervisorSearch}
                  onChange={(e) => setSupervisorSearch(e.target.value)}
                  className="pl-9 bg-slate-800 border-slate-700 text-white"
                />
              </div>
              <select
                id="supervisorId"
                value={(data.supervisorId as string) || ""}
                onChange={(e) => handleChange("supervisorId", e.target.value || null)}
                disabled={loadingEmployees}
                className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white disabled:opacity-50"
              >
                <option value="">{loadingEmployees ? "Loading employees..." : "Select supervisor..."}</option>
                {filteredSupervisors.map((emp) => (
                  <option key={emp.id} value={emp.id}>
                    {emp.first_name} {emp.last_name} ({emp.employee_number})
                  </option>
                ))}
              </select>
              {(() => {
                const supervisorIdStr = data.supervisorId as string
                if ((employees || []).length > 0 && supervisorIdStr) {
                  const selectedSupervisor = (employees || []).find(e => e.id === supervisorIdStr)
                  if (selectedSupervisor) {
                    return (
                      <p className="text-xs text-slate-500 mt-1">
                        {selectedSupervisor.employee_number} - {selectedSupervisor.collar_type?.replace("_", " ")}
                      </p>
                    )
                  }
                }
                return null
              })()}
            </div>
          </div>
        </div>
      </div>

      {/* Salary Information */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Salary Information</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="dailySalary" className="text-slate-300">
              Daily Salary <span className="text-red-400">*</span>
            </Label>
            <Input
              id="dailySalary"
              type="number"
              step="0.01"
              placeholder="500.00"
              value={(data.dailySalary as string) || ""}
              onChange={(e) => handleChange("dailySalary", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="integratedDailySalary" className="text-slate-300">SDI (Integrated Daily Salary)</Label>
            <Input
              id="integratedDailySalary"
              type="number"
              step="0.01"
              placeholder="Calculated automatically"
              value={(data.integratedDailySalary as string) || ""}
              onChange={(e) => handleChange("integratedDailySalary", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
            <p className="text-xs text-slate-500 mt-1">Leave empty to calculate automatically</p>
          </div>
          <div>
            <Label htmlFor="paymentMethod" className="text-slate-300">Payment Method</Label>
            <select
              id="paymentMethod"
              value={(data.paymentMethod as string) || "bank_transfer"}
              onChange={(e) => handleChange("paymentMethod", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {PAYMENT_METHOD_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* IMSS & Tax */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">IMSS & Tax</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <Label htmlFor="imssRegistrationDate" className="text-slate-300">IMSS Registration Date</Label>
            <Input
              id="imssRegistrationDate"
              type="date"
              value={(data.imssRegistrationDate as string) || ""}
              onChange={(e) => handleChange("imssRegistrationDate", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="regime" className="text-slate-300">Regime</Label>
            <select
              id="regime"
              value={(data.regime as string) || "salary"}
              onChange={(e) => handleChange("regime", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {REGIME_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="taxRegime" className="text-slate-300">Tax Regime</Label>
            <select
              id="taxRegime"
              value={(data.taxRegime as string) || ""}
              onChange={(e) => handleChange("taxRegime", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {TAX_REGIME_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="fiscalPostalCode" className="text-slate-300">Fiscal Postal Code</Label>
            <Input
              id="fiscalPostalCode"
              placeholder="78000"
              value={(data.fiscalPostalCode as string) || ""}
              onChange={(e) => handleChange("fiscalPostalCode", e.target.value.replace(/\D/g, ""))}
              maxLength={5}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div className="md:col-span-2">
            <Label htmlFor="fiscalName" className="text-slate-300">Fiscal Name</Label>
            <Input
              id="fiscalName"
              placeholder="Employee fiscal name"
              value={(data.fiscalName as string) || ""}
              onChange={(e) => handleChange("fiscalName", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Operational / Logistics */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Operational / Logistics</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="route" className="text-slate-300">Route</Label>
            <Input
              id="route"
              placeholder="Transport route"
              value={(data.route as string) || ""}
              onChange={(e) => handleChange("route", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="transportStop" className="text-slate-300">Transport Stop</Label>
            <Input
              id="transportStop"
              placeholder="Assigned stop"
              value={(data.transportStop as string) || ""}
              onChange={(e) => handleChange("transportStop", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="recruitmentSource" className="text-slate-300">Recruitment Source</Label>
            <Input
              id="recruitmentSource"
              placeholder="LinkedIn, Indeed, Referral..."
              value={(data.recruitmentSource as string) || ""}
              onChange={(e) => handleChange("recruitmentSource", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Children Information */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Children Information</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <Label htmlFor="child1Gender" className="text-slate-300">Child 1</Label>
            <select
              id="child1Gender"
              value={(data.child1Gender as string) || ""}
              onChange={(e) => handleChange("child1Gender", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              <option value="">No record</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
            </select>
          </div>
          <div>
            <Label htmlFor="child2Gender" className="text-slate-300">Child 2</Label>
            <select
              id="child2Gender"
              value={(data.child2Gender as string) || ""}
              onChange={(e) => handleChange("child2Gender", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              <option value="">No record</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
            </select>
          </div>
          <div>
            <Label htmlFor="child3Gender" className="text-slate-300">Child 3</Label>
            <select
              id="child3Gender"
              value={(data.child3Gender as string) || ""}
              onChange={(e) => handleChange("child3Gender", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              <option value="">No record</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
            </select>
          </div>
          <div>
            <Label htmlFor="child4Gender" className="text-slate-300">Child 4</Label>
            <select
              id="child4Gender"
              value={(data.child4Gender as string) || ""}
              onChange={(e) => handleChange("child4Gender", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              <option value="">No record</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  )
}
