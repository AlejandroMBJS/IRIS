/**
 * @file app/employees/new/page.tsx
 * @description Multi-step form for creating a new employee record
 *
 * USER PERSPECTIVE:
 *   - Fill out employee information across 4 tabs: Personal, Employment, Bank, Emergency Contact
 *   - Navigate between tabs or use next/previous buttons
 *   - Required fields are validated before submission
 *   - SDI (Integrated Daily Salary) is auto-calculated if not provided
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Form fields, tab order, validation rules
 *   CAUTION: Required fields list must match backend validation
 *   DO NOT modify: Tab structure without updating all form components
 *
 * KEY COMPONENTS:
 *   - Tabs: Personal, Employment, Bank Info, Emergency Contact
 *   - EmployeeForm: Personal information fields
 *   - EmploymentForm: Job and salary details
 *   - BankInfoForm: Banking information
 *   - EmergencyContactForm: Emergency contact details
 *
 * API ENDPOINTS USED:
 *   - POST /api/employees (via createEmployee() helper)
 */

"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { ArrowLeft, Calendar, MapPin, Building2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useToast } from "@/hooks/use-toast"
import { createEmployee } from "@/lib/api/employees"
import { EmployeeForm } from "@/components/employees/employee-form"
import { EmploymentForm } from "@/components/employees/employment-form"
import { BankInfoForm } from "@/components/employees/bank-info-form"
import { EmergencyContactForm } from "@/components/employees/emergency-contact-form"
import { DashboardLayout } from "@/components/layout/dashboard-layout"

export default function NewEmployeePage() {
  const router = useRouter()
  const { toast } = useToast()
  const [activeTab, setActiveTab] = useState("personal")
  const [isSubmitting, setIsSubmitting] = useState(false)
  
  const [formData, setFormData] = useState({
    // Personal Information
    employeeNumber: "",
    firstName: "",
    lastName: "",
    motherLastName: "",
    birthDate: "",
    gender: "",
    maritalStatus: "",
    nationality: "Mexicana",
    birthState: "",
    educationLevel: "",

    // Identification
    rfc: "",
    curp: "",
    nss: "",
    infonavitCredit: "",

    // Contact
    personalEmail: "",
    personalPhone: "",
    cellPhone: "",

    // Address
    street: "",
    exteriorNumber: "",
    interiorNumber: "",
    neighborhood: "",
    municipality: "",
    state: "San Luis Potosí",
    postalCode: "",
    country: "México",

    // Employment
    hireDate: "",
    terminationDate: "",
    employmentStatus: "active",
    employeeType: "permanent",
    collarType: "white_collar",
    payFrequency: "biweekly",
    isSindicalizado: false,
    contractStartDate: "",
    contractEndDate: "",
    departmentId: "",
    positionId: "",
    costCenterId: "",
    productionArea: "",
    location: "",
    patronalRegistry: "",
    companyName: "",
    shiftId: "",
    supervisorId: "",
    teamName: "",
    packageCode: "",

    // Operational/Logistics
    route: "",
    transportStop: "",
    recruitmentSource: "",

    // Salary
    dailySalary: "",
    integratedDailySalary: "",
    paymentMethod: "bank_transfer",

    // Bank
    bankName: "",
    bankAccount: "",
    clabe: "",

    // IMSS & Tax
    imssRegistrationDate: "",
    regime: "salary",
    taxRegime: "",
    fiscalPostalCode: "",
    fiscalName: "",

    // Emergency Contact
    emergencyContact: "",
    emergencyPhone: "",
    emergencyRelationship: "",

    // Family Information (Children)
    child1Gender: "",
    child2Gender: "",
    child3Gender: "",
    child4Gender: "",
  })

  const handleSubmit = async () => {
    setIsSubmitting(true)
    
    try {
      // Validate required fields
      const requiredFields = [
        'employeeNumber', 'firstName', 'lastName', 'rfc', 'curp',
        'birthDate', 'hireDate', 'dailySalary'
      ]

      const missingFields = requiredFields.filter(field => !formData[field as keyof typeof formData])
      
      if (missingFields.length > 0) {
        toast({
          title: "Required fields",
          description: `Missing fields: ${missingFields.join(', ')}`,
          variant: "destructive",
        })
        return
      }
      
      // Calculate SDI if not provided
      const sdi = formData.integratedDailySalary || (parseFloat(formData.dailySalary) * 1.045).toString()

      const employeeData = {
        employee_number: formData.employeeNumber,
        first_name: formData.firstName,
        last_name: formData.lastName,
        mother_last_name: formData.motherLastName,
        date_of_birth: formData.birthDate,
        gender: formData.gender,
        marital_status: formData.maritalStatus,
        nationality: formData.nationality,
        birth_state: formData.birthState,
        education_level: formData.educationLevel,
        rfc: formData.rfc,
        curp: formData.curp,
        nss: formData.nss,
        infonavit_credit: formData.infonavitCredit,
        personal_email: formData.personalEmail,
        personal_phone: formData.personalPhone,
        cell_phone: formData.cellPhone,
        street: formData.street,
        exterior_number: formData.exteriorNumber,
        interior_number: formData.interiorNumber,
        neighborhood: formData.neighborhood,
        municipality: formData.municipality,
        state: formData.state,
        postal_code: formData.postalCode,
        country: formData.country,
        hire_date: formData.hireDate,
        termination_date: formData.terminationDate || null,
        employment_status: formData.employmentStatus,
        employee_type: formData.employeeType,
        collar_type: formData.collarType,
        pay_frequency: formData.payFrequency,
        is_sindicalizado: formData.isSindicalizado,
        contract_start_date: formData.contractStartDate || null,
        contract_end_date: formData.contractEndDate || null,
        department_id: formData.departmentId || null,
        position_id: formData.positionId || null,
        cost_center_id: formData.costCenterId || null,
        production_area: formData.productionArea,
        location: formData.location,
        patronal_registry: formData.patronalRegistry,
        company_name: formData.companyName,
        shift_id: formData.shiftId || null,
        supervisor_id: formData.supervisorId || null,
        team_name: formData.teamName,
        package_code: formData.packageCode,
        route: formData.route,
        transport_stop: formData.transportStop,
        recruitment_source: formData.recruitmentSource,
        daily_salary: parseFloat(formData.dailySalary) || 0,
        integrated_daily_salary: parseFloat(sdi) || 0,
        payment_method: formData.paymentMethod,
        bank_name: formData.bankName,
        bank_account: formData.bankAccount,
        clabe: formData.clabe,
        imss_registration_date: formData.imssRegistrationDate || null,
        regime: formData.regime,
        tax_regime: formData.taxRegime,
        fiscal_postal_code: formData.fiscalPostalCode,
        fiscal_name: formData.fiscalName,
        emergency_contact: formData.emergencyContact,
        emergency_phone: formData.emergencyPhone,
        emergency_relationship: formData.emergencyRelationship,
        child1_gender: formData.child1Gender,
        child2_gender: formData.child2Gender,
        child3_gender: formData.child3Gender,
        child4_gender: formData.child4Gender,
      }
      
      const result = await createEmployee(employeeData)
      
      if (result.success) {
        toast({
          title: "Employee created",
          description: "The employee has been registered successfully",
        })
        router.push("/employees")
      } else {
        throw new Error(result.message || "Error creating employee")
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Error creating employee",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <DashboardLayout>
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => router.push("/employees")}
            className="hover:bg-slate-800"
            aria-label="Back to employees list"
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-white">New Employee</h1>
            <p className="text-slate-400">Register a new employee in the system</p>
          </div>
        </div>
        
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => router.push("/employees")}
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSubmitting}
            className="bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700"
          >
            {isSubmitting ? "Saving..." : "Save Employee"}
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid grid-cols-4 w-full bg-slate-800">
          <TabsTrigger value="personal" className="data-[state=active]:bg-primary">
            Personal Information
          </TabsTrigger>
          <TabsTrigger value="employment" className="data-[state=active]:bg-primary">
            <Calendar className="h-4 w-4 mr-2" />
            Employment
          </TabsTrigger>
          <TabsTrigger value="bank" className="data-[state=active]:bg-primary">
            <Building2 className="h-4 w-4 mr-2" />
            Bank Information
          </TabsTrigger>
          <TabsTrigger value="emergency" className="data-[state=active]:bg-primary">
            <MapPin className="h-4 w-4 mr-2" />
            Emergency Contact
          </TabsTrigger>
        </TabsList>

        <TabsContent value="personal" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Personal Information</CardTitle>
            </CardHeader>
            <CardContent>
              <EmployeeForm
                data={formData}
                onChange={(data: Record<string, unknown>) => setFormData({ ...formData, ...data })}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="employment" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Employment Information</CardTitle>
            </CardHeader>
            <CardContent>
              <EmploymentForm
                data={formData}
                onChange={(data: Record<string, unknown>) => setFormData({ ...formData, ...data })}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="bank" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Bank Information</CardTitle>
            </CardHeader>
            <CardContent>
              <BankInfoForm
                data={formData}
                onChange={(data: Record<string, unknown>) => setFormData({ ...formData, ...data })}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="emergency" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Emergency Contact</CardTitle>
            </CardHeader>
            <CardContent>
              <EmergencyContactForm
                data={formData}
                onChange={(data: Record<string, unknown>) => setFormData({ ...formData, ...data })}
              />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <div className="flex justify-between pt-6 border-t border-slate-700">
        <Button
          variant="outline"
          onClick={() => {
            const tabs = ["personal", "employment", "bank", "emergency"]
            const currentIndex = tabs.indexOf(activeTab)
            if (currentIndex > 0) setActiveTab(tabs[currentIndex - 1])
          }}
          className="border-slate-700 text-slate-300 hover:bg-slate-800"
          disabled={activeTab === "personal"}
        >
          Previous
        </Button>
        
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => {
              // Reset form to initial state
              setFormData({
                // Personal Information
                employeeNumber: "",
                firstName: "",
                lastName: "",
                motherLastName: "",
                birthDate: "",
                gender: "",
                maritalStatus: "",
                nationality: "Mexicana",
                birthState: "",
                educationLevel: "",
                // Identification
                rfc: "",
                curp: "",
                nss: "",
                infonavitCredit: "",
                // Contact
                personalEmail: "",
                personalPhone: "",
                cellPhone: "",
                // Address
                street: "",
                exteriorNumber: "",
                interiorNumber: "",
                neighborhood: "",
                municipality: "",
                state: "San Luis Potosí",
                postalCode: "",
                country: "México",
                // Employment
                hireDate: "",
                terminationDate: "",
                employmentStatus: "active",
                employeeType: "permanent",
                collarType: "white_collar",
                payFrequency: "biweekly",
                isSindicalizado: false,
                contractStartDate: "",
                contractEndDate: "",
                departmentId: "",
                positionId: "",
                costCenterId: "",
                productionArea: "",
                location: "",
                patronalRegistry: "",
                companyName: "",
                shiftId: "",
                supervisorId: "",
                teamName: "",
                packageCode: "",
                // Operational/Logistics
                route: "",
                transportStop: "",
                recruitmentSource: "",
                // Salary
                dailySalary: "",
                integratedDailySalary: "",
                paymentMethod: "bank_transfer",
                // Bank
                bankName: "",
                bankAccount: "",
                clabe: "",
                // IMSS & Tax
                imssRegistrationDate: "",
                regime: "salary",
                taxRegime: "",
                fiscalPostalCode: "",
                fiscalName: "",
                // Emergency Contact
                emergencyContact: "",
                emergencyPhone: "",
                emergencyRelationship: "",
                // Family Information (Children)
                child1Gender: "",
                child2Gender: "",
                child3Gender: "",
                child4Gender: "",
              })
              setActiveTab("personal")
            }}
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            Clear Form
          </Button>

          <Button
            onClick={() => {
              const tabs = ["personal", "employment", "bank", "emergency"]
              const currentIndex = tabs.indexOf(activeTab)
              if (currentIndex < tabs.length - 1) {
                setActiveTab(tabs[currentIndex + 1])
              } else {
                handleSubmit()
              }
            }}
            className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
          >
            {activeTab === "emergency" ? "Save Employee" : "Next"}
          </Button>
        </div>
      </div>
    </div>
    </DashboardLayout>
  )
}
