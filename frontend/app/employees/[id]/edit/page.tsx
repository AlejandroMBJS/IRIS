/**
 * @file app/employees/[id]/edit/page.tsx
 * @description Edit employee page - loads existing employee data and allows updates
 * Includes shift and supervisor assignment capabilities
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import { ArrowLeft, Calendar, MapPin, Building2, Loader2, Users, Clock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useToast } from "@/hooks/use-toast"
import { employeeApi, Employee, ApiError, shiftApi, ShiftResponse } from "@/lib/api-client"
import { DashboardLayout } from "@/components/layout/dashboard-layout"

export default function EditEmployeePage() {
  const router = useRouter()
  const params = useParams()
  const { toast } = useToast()
  const [activeTab, setActiveTab] = useState("personal")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Lists for dropdowns
  const [shifts, setShifts] = useState<ShiftResponse[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loadingShifts, setLoadingShifts] = useState(true)
  const [loadingEmployees, setLoadingEmployees] = useState(true)

  const [formData, setFormData] = useState({
    employee_number: "",
    first_name: "",
    last_name: "",
    mother_last_name: "",
    date_of_birth: "",
    gender: "",
    marital_status: "",
    rfc: "",
    curp: "",
    nss: "",
    infonavit_credit: "",
    personal_email: "",
    personal_phone: "",
    emergency_contact: "",
    emergency_phone: "",
    street: "",
    exterior_number: "",
    interior_number: "",
    neighborhood: "",
    municipality: "",
    state: "",
    postal_code: "",
    country: "",
    hire_date: "",
    employment_status: "active",
    employee_type: "permanent",
    collar_type: "white_collar",
    pay_frequency: "biweekly",
    is_sindicalizado: false,
    daily_salary: "",
    integrated_daily_salary: "",
    payment_method: "bank_transfer",
    bank_name: "",
    bank_account: "",
    clabe: "",
    tax_regime: "",
    shift_id: "",
    supervisor_id: "",
  })

  useEffect(() => {
    loadShifts()
    loadEmployeesList()
    if (params.id) {
      loadEmployee(params.id as string)
    }
  }, [params.id])

  async function loadShifts() {
    try {
      setLoadingShifts(true)
      const data = await shiftApi.getActive()
      setShifts((data || []) as ShiftResponse[])
    } catch (err) {
      console.error("Error loading shifts:", err)
      setShifts([])
    } finally {
      setLoadingShifts(false)
    }
  }

  async function loadEmployeesList() {
    try {
      setLoadingEmployees(true)
      const response = await employeeApi.getEmployees({ page: 1, page_size: 1000 })
      // Filter out the current employee from the supervisor list
      const emps = response?.employees || []
      const filteredEmployees = emps.filter(e => e.id !== params.id)
      setEmployees(filteredEmployees)
    } catch (err) {
      console.error("Error loading employees:", err)
      setEmployees([])
    } finally {
      setLoadingEmployees(false)
    }
  }

  async function loadEmployee(id: string) {
    try {
      setLoading(true)
      setError(null)
      const employee = await employeeApi.getEmployee(id)

      // Convert employee data to form format
      setFormData({
        employee_number: employee.employee_number || "",
        first_name: employee.first_name || "",
        last_name: employee.last_name || "",
        mother_last_name: employee.mother_last_name || "",
        date_of_birth: employee.date_of_birth ? employee.date_of_birth.split('T')[0] : "",
        gender: employee.gender || "",
        marital_status: "",
        rfc: employee.rfc || "",
        curp: employee.curp || "",
        nss: employee.nss || "",
        infonavit_credit: employee.infonavit_credit || "",
        personal_email: employee.personal_email || "",
        personal_phone: employee.personal_phone || "",
        emergency_contact: employee.emergency_contact || "",
        emergency_phone: employee.emergency_phone || "",
        street: employee.street || "",
        exterior_number: employee.exterior_number || "",
        interior_number: employee.interior_number || "",
        neighborhood: employee.neighborhood || "",
        municipality: employee.municipality || "",
        state: employee.state || "",
        postal_code: employee.postal_code || "",
        country: employee.country || "",
        hire_date: employee.hire_date ? employee.hire_date.split('T')[0] : "",
        employment_status: employee.employment_status || "active",
        employee_type: employee.employee_type || "permanent",
        collar_type: employee.collar_type || "white_collar",
        pay_frequency: employee.pay_frequency || "biweekly",
        is_sindicalizado: employee.is_sindicalizado || false,
        daily_salary: employee.daily_salary?.toString() || "",
        integrated_daily_salary: employee.integrated_daily_salary?.toString() || "",
        payment_method: employee.payment_method || "bank_transfer",
        bank_name: employee.bank_name || "",
        bank_account: employee.bank_account || "",
        clabe: employee.clabe || "",
        tax_regime: employee.tax_regime || "",
        shift_id: employee.shift_id || "",
        supervisor_id: employee.supervisor_id || "",
      })
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load employee")
      }
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async () => {
    setIsSubmitting(true)

    try {
      const updateData: Partial<Employee> = {
        employee_number: formData.employee_number,
        first_name: formData.first_name,
        last_name: formData.last_name,
        mother_last_name: formData.mother_last_name || undefined,
        date_of_birth: formData.date_of_birth,
        gender: formData.gender,
        rfc: formData.rfc,
        curp: formData.curp,
        nss: formData.nss || undefined,
        infonavit_credit: formData.infonavit_credit || undefined,
        personal_email: formData.personal_email || undefined,
        personal_phone: formData.personal_phone || undefined,
        emergency_contact: formData.emergency_contact || undefined,
        emergency_phone: formData.emergency_phone || undefined,
        street: formData.street || undefined,
        exterior_number: formData.exterior_number || undefined,
        interior_number: formData.interior_number || undefined,
        neighborhood: formData.neighborhood || undefined,
        municipality: formData.municipality || undefined,
        state: formData.state || undefined,
        postal_code: formData.postal_code || undefined,
        country: formData.country || undefined,
        hire_date: formData.hire_date,
        employment_status: formData.employment_status,
        employee_type: formData.employee_type,
        collar_type: formData.collar_type,
        pay_frequency: formData.pay_frequency,
        is_sindicalizado: formData.is_sindicalizado,
        daily_salary: parseFloat(formData.daily_salary) || 0,
        integrated_daily_salary: parseFloat(formData.integrated_daily_salary) || undefined,
        payment_method: formData.payment_method,
        bank_name: formData.bank_name || undefined,
        bank_account: formData.bank_account || undefined,
        clabe: formData.clabe || undefined,
        tax_regime: formData.tax_regime || undefined,
        shift_id: formData.shift_id || undefined,
        supervisor_id: formData.supervisor_id || undefined,
      }

      await employeeApi.updateEmployee(params.id as string, updateData)

      toast({
        title: "Employee updated",
        description: "Employee data has been updated successfully",
      })
      router.push(`/employees/${params.id}`)
    } catch (err) {
      toast({
        title: "Error",
        description: err instanceof ApiError ? err.message : "Error updating employee",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const InputField = ({
    label,
    name,
    type = "text",
    required = false,
    placeholder = ""
  }: {
    label: string
    name: keyof typeof formData
    type?: string
    required?: boolean
    placeholder?: string
  }) => (
    <div>
      <label className="block text-sm font-medium text-slate-300 mb-1.5">
        {label} {required && <span className="text-red-400">*</span>}
      </label>
      <input
        type={type}
        value={formData[name] as string}
        onChange={(e) => setFormData({ ...formData, [name]: e.target.value })}
        className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
        placeholder={placeholder}
        required={required}
      />
    </div>
  )

  const SelectField = ({
    label,
    name,
    options,
    required = false
  }: {
    label: string
    name: keyof typeof formData
    options: { value: string, label: string }[]
    required?: boolean
  }) => (
    <div>
      <label className="block text-sm font-medium text-slate-300 mb-1.5">
        {label} {required && <span className="text-red-400">*</span>}
      </label>
      <select
        value={formData[name] as string}
        onChange={(e) => setFormData({ ...formData, [name]: e.target.value })}
        className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
        required={required}
      >
        <option value="">Select...</option>
        {options.map(opt => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  )

  // Get available shifts based on employee's collar type
  const getAvailableShifts = () => {
    const shiftList = shifts || []
    if (!formData.collar_type) return shiftList.filter(s => s.is_active)

    return shiftList.filter(shift => {
      if (!shift.is_active) return false
      // If shift has no collar_types restriction, it's available to all
      if (!shift.collar_types || shift.collar_types === "[]") return true

      try {
        const collarTypes = JSON.parse(shift.collar_types)
        return collarTypes.length === 0 || collarTypes.includes(formData.collar_type)
      } catch {
        return true
      }
    })
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="flex items-center gap-2 text-slate-400">
            <Loader2 className="w-5 h-5 animate-spin" />
            Loading employee...
          </div>
        </div>
      </DashboardLayout>
    )
  }

  if (error) {
    return (
      <DashboardLayout>
        <div className="space-y-4">
          <button
            onClick={() => router.push("/employees")}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={20} />
            Back to Employees
          </button>
          <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 text-red-400">
            {error}
          </div>
        </div>
      </DashboardLayout>
    )
  }

  const availableShifts = getAvailableShifts()

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => router.push(`/employees/${params.id}`)}
              className="hover:bg-slate-800"
            >
              <ArrowLeft className="h-5 w-5" />
            </Button>
            <div>
              <h1 className="text-3xl font-bold text-white">Edit Employee</h1>
              <p className="text-slate-400">{formData.first_name} {formData.last_name} - #{formData.employee_number}</p>
            </div>
          </div>

          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => router.push(`/employees/${params.id}`)}
              className="border-slate-700 text-slate-300 hover:bg-slate-800"
            >
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={isSubmitting}
              className="bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700"
            >
              {isSubmitting ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </div>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid grid-cols-5 w-full bg-slate-800">
            <TabsTrigger value="personal" className="data-[state=active]:bg-primary">
              Personal Data
            </TabsTrigger>
            <TabsTrigger value="employment" className="data-[state=active]:bg-primary">
              <Calendar className="h-4 w-4 mr-2" />
              Employment
            </TabsTrigger>
            <TabsTrigger value="organization" className="data-[state=active]:bg-primary">
              <Users className="h-4 w-4 mr-2" />
              Organization
            </TabsTrigger>
            <TabsTrigger value="bank" className="data-[state=active]:bg-primary">
              <Building2 className="h-4 w-4 mr-2" />
              Bank Data
            </TabsTrigger>
            <TabsTrigger value="address" className="data-[state=active]:bg-primary">
              <MapPin className="h-4 w-4 mr-2" />
              Address
            </TabsTrigger>
          </TabsList>

          <TabsContent value="personal" className="space-y-4">
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Personal Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <InputField label="Employee Number" name="employee_number" required />
                  <InputField label="First Name(s)" name="first_name" required />
                  <InputField label="Paternal Last Name" name="last_name" required />
                  <InputField label="Maternal Last Name" name="mother_last_name" />
                  <InputField label="Date of Birth" name="date_of_birth" type="date" required />
                  <SelectField
                    label="Gender"
                    name="gender"
                    required
                    options={[
                      { value: "male", label: "Male" },
                      { value: "female", label: "Female" },
                      { value: "other", label: "Other" },
                    ]}
                  />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <InputField label="RFC" name="rfc" required placeholder="XXXX000000XXX" />
                  <InputField label="CURP" name="curp" required placeholder="18 characters" />
                  <InputField label="NSS (IMSS)" name="nss" placeholder="11 digits" />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <InputField label="Personal Email" name="personal_email" type="email" />
                  <InputField label="Personal Phone" name="personal_phone" />
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="employment" className="space-y-4">
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Employment Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <InputField label="Hire Date" name="hire_date" type="date" required />
                  <SelectField
                    label="Employment Status"
                    name="employment_status"
                    options={[
                      { value: "active", label: "Active" },
                      { value: "inactive", label: "Inactive" },
                      { value: "on_leave", label: "On Leave" },
                      { value: "terminated", label: "Terminated" },
                    ]}
                  />
                  <SelectField
                    label="Employee Type"
                    name="employee_type"
                    options={[
                      { value: "permanent", label: "Permanent" },
                      { value: "temporary", label: "Temporary" },
                      { value: "contractor", label: "Contractor" },
                      { value: "intern", label: "Intern" },
                    ]}
                  />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <SelectField
                    label="Collar Type"
                    name="collar_type"
                    required
                    options={[
                      { value: "white_collar", label: "White Collar (Administrative)" },
                      { value: "blue_collar", label: "Blue Collar (Unionized)" },
                      { value: "gray_collar", label: "Gray Collar (Non-Unionized)" },
                    ]}
                  />
                  <SelectField
                    label="Pay Frequency"
                    name="pay_frequency"
                    required
                    options={[
                      { value: "weekly", label: "Weekly" },
                      { value: "biweekly", label: "Biweekly" },
                      { value: "monthly", label: "Monthly" },
                    ]}
                  />
                  <div className="flex items-center gap-2 pt-7">
                    <input
                      type="checkbox"
                      checked={formData.is_sindicalizado}
                      onChange={(e) => setFormData({ ...formData, is_sindicalizado: e.target.checked })}
                      className="rounded border-slate-600 bg-slate-800"
                    />
                    <label className="text-sm text-slate-300">Is Unionized</label>
                  </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <InputField label="Daily Salary" name="daily_salary" type="number" required placeholder="0.00" />
                  <InputField label="SDI (Integrated Daily Salary)" name="integrated_daily_salary" type="number" placeholder="Calculated automatically" />
                  <InputField label="INFONAVIT Credit" name="infonavit_credit" />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <InputField label="Tax Regime" name="tax_regime" />
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="organization" className="space-y-4">
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  Shift Assignment
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Work Shift
                  </label>
                  {loadingShifts ? (
                    <div className="flex items-center gap-2 text-slate-400 py-2">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Loading shifts...
                    </div>
                  ) : availableShifts.length === 0 ? (
                    <div className="text-slate-400 py-2">
                      No shifts available for this collar type.
                      <Button
                        variant="link"
                        className="text-blue-400 ml-2 p-0"
                        onClick={() => router.push('/configuration/shifts')}
                      >
                        Configure shifts
                      </Button>
                    </div>
                  ) : (
                    <select
                      value={formData.shift_id}
                      onChange={(e) => setFormData({ ...formData, shift_id: e.target.value })}
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
                    >
                      <option value="">No shift assigned</option>
                      {availableShifts.map(shift => (
                        <option key={shift.id} value={shift.id}>
                          {shift.name} ({shift.code}) - {shift.start_time} to {shift.end_time}
                        </option>
                      ))}
                    </select>
                  )}
                  <p className="text-xs text-slate-500 mt-1">
                    Only shifts compatible with the employee's collar type are shown
                  </p>
                </div>

                {/* Show selected shift details */}
                {formData.shift_id && (
                  <div className="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
                    {(() => {
                      const selectedShift = (shifts || []).find(s => s.id === formData.shift_id)
                      if (!selectedShift) return null

                      const workDays = (() => {
                        try {
                          const days = JSON.parse(selectedShift.work_days)
                          const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
                          return days.map((d: number) => dayNames[d]).join(', ')
                        } catch {
                          return 'M-F'
                        }
                      })()

                      return (
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                          <div>
                            <span className="text-slate-400">Schedule:</span>
                            <p className="text-white font-medium">{selectedShift.start_time} - {selectedShift.end_time}</p>
                          </div>
                          <div>
                            <span className="text-slate-400">Days:</span>
                            <p className="text-white font-medium">{workDays}</p>
                          </div>
                          <div>
                            <span className="text-slate-400">Hours/day:</span>
                            <p className="text-white font-medium">{selectedShift.work_hours_per_day}h</p>
                          </div>
                          <div>
                            <span className="text-slate-400">Break:</span>
                            <p className="text-white font-medium">{selectedShift.break_minutes} min</p>
                          </div>
                        </div>
                      )
                    })()}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Users className="h-5 w-5" />
                  Direct Supervisor
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-1.5">
                    Supervisor
                  </label>
                  {loadingEmployees ? (
                    <div className="flex items-center gap-2 text-slate-400 py-2">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Loading employees...
                    </div>
                  ) : (
                    <select
                      value={formData.supervisor_id}
                      onChange={(e) => setFormData({ ...formData, supervisor_id: e.target.value })}
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
                    >
                      <option value="">No supervisor assigned</option>
                      {(employees || [])
                        .filter(emp => emp.employment_status === 'active')
                        .sort((a, b) => (a.full_name || '').localeCompare(b.full_name || ''))
                        .map(emp => (
                          <option key={emp.id} value={emp.id}>
                            {emp.full_name} ({emp.employee_number}) - {
                              emp.collar_type === 'white_collar' ? 'Administrative' :
                              emp.collar_type === 'blue_collar' ? 'Operative' :
                              'Mixed'
                            }
                          </option>
                        ))}
                    </select>
                  )}
                  <p className="text-xs text-slate-500 mt-1">
                    Select the direct supervisor of this employee for the approval chain
                  </p>
                </div>

                {/* Show selected supervisor details */}
                {formData.supervisor_id && (
                  <div className="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
                    {(() => {
                      const supervisor = (employees || []).find(e => e.id === formData.supervisor_id)
                      if (!supervisor) return null

                      return (
                        <div className="flex items-center gap-4">
                          <div className="w-10 h-10 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold">
                            {supervisor.first_name[0]}{supervisor.last_name[0]}
                          </div>
                          <div>
                            <p className="text-white font-medium">{supervisor.full_name}</p>
                            <p className="text-sm text-slate-400">
                              #{supervisor.employee_number} - {
                                supervisor.collar_type === 'white_collar' ? 'White Collar' :
                                supervisor.collar_type === 'blue_collar' ? 'Blue Collar' :
                                'Gray Collar'
                              }
                            </p>
                          </div>
                        </div>
                      )
                    })()}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="bank" className="space-y-4">
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Bank Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <SelectField
                    label="Payment Method"
                    name="payment_method"
                    options={[
                      { value: "bank_transfer", label: "Bank Transfer" },
                      { value: "cash", label: "Cash" },
                      { value: "check", label: "Check" },
                    ]}
                  />
                  <InputField label="Bank" name="bank_name" placeholder="Bank name" />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <InputField label="Account Number" name="bank_account" />
                  <InputField label="CLABE Interbank Code" name="clabe" placeholder="18 digits" />
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="address" className="space-y-4">
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Address</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="md:col-span-2">
                    <InputField label="Street" name="street" />
                  </div>
                  <InputField label="Exterior Number" name="exterior_number" />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <InputField label="Interior Number" name="interior_number" />
                  <InputField label="Neighborhood" name="neighborhood" />
                  <InputField label="Municipality/District" name="municipality" />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <InputField label="State" name="state" />
                  <InputField label="Postal Code" name="postal_code" />
                  <InputField label="Country" name="country" placeholder="Mexico" />
                </div>
              </CardContent>
            </Card>

            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Emergency Contact</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <InputField label="Contact Name" name="emergency_contact" />
                  <InputField label="Emergency Phone" name="emergency_phone" />
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        <div className="flex justify-end gap-2 pt-6 border-t border-slate-700">
          <Button
            variant="outline"
            onClick={() => router.push(`/employees/${params.id}`)}
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSubmitting}
            className="bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700"
          >
            {isSubmitting ? "Saving..." : "Save Changes"}
          </Button>
        </div>
      </div>
    </DashboardLayout>
  )
}
