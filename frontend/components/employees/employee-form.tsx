"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface EmployeeFormProps {
  data: Record<string, unknown>
  onChange: (data: Record<string, unknown>) => void
}

const GENDER_OPTIONS = [
  { value: "", label: "Select..." },
  { value: "male", label: "Male" },
  { value: "female", label: "Female" },
  { value: "other", label: "Other" },
]

const MARITAL_STATUS_OPTIONS = [
  { value: "", label: "Select..." },
  { value: "single", label: "Single" },
  { value: "married", label: "Married" },
  { value: "divorced", label: "Divorced" },
  { value: "widowed", label: "Widowed" },
  { value: "cohabiting", label: "Cohabiting" },
]

const EDUCATION_OPTIONS = [
  { value: "", label: "Select..." },
  { value: "none", label: "No Education" },
  { value: "primary", label: "Primary" },
  { value: "secondary", label: "Secondary" },
  { value: "high_school", label: "High School" },
  { value: "technical", label: "Technical" },
  { value: "bachelor", label: "Bachelor's Degree" },
  { value: "master", label: "Master's Degree" },
  { value: "doctorate", label: "Doctorate" },
]

const MEXICAN_STATES = [
  "Aguascalientes", "Baja California", "Baja California Sur", "Campeche",
  "Chiapas", "Chihuahua", "Ciudad de México", "Coahuila", "Colima",
  "Durango", "Estado de México", "Guanajuato", "Guerrero", "Hidalgo",
  "Jalisco", "Michoacán", "Morelos", "Nayarit", "Nuevo León", "Oaxaca",
  "Puebla", "Querétaro", "Quintana Roo", "San Luis Potosí", "Sinaloa",
  "Sonora", "Tabasco", "Tamaulipas", "Tlaxcala", "Veracruz", "Yucatán", "Zacatecas"
]

export function EmployeeForm({ data, onChange }: EmployeeFormProps) {
  const handleChange = (field: string, value: unknown) => {
    onChange({ [field]: value })
  }

  return (
    <div className="space-y-6">
      {/* Basic Info */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Basic Information</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="employeeNumber" className="text-slate-300">
              Employee Number <span className="text-red-400">*</span>
            </Label>
            <Input
              id="employeeNumber"
              placeholder="EMP001"
              value={(data.employeeNumber as string) || ""}
              onChange={(e) => handleChange("employeeNumber", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="firstName" className="text-slate-300">
              First Name(s) <span className="text-red-400">*</span>
            </Label>
            <Input
              id="firstName"
              placeholder="Juan Carlos"
              value={(data.firstName as string) || ""}
              onChange={(e) => handleChange("firstName", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="lastName" className="text-slate-300">
              Paternal Last Name <span className="text-red-400">*</span>
            </Label>
            <Input
              id="lastName"
              placeholder="Pérez"
              value={(data.lastName as string) || ""}
              onChange={(e) => handleChange("lastName", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="motherLastName" className="text-slate-300">Maternal Last Name</Label>
            <Input
              id="motherLastName"
              placeholder="González"
              value={(data.motherLastName as string) || ""}
              onChange={(e) => handleChange("motherLastName", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="birthDate" className="text-slate-300">
              Date of Birth <span className="text-red-400">*</span>
            </Label>
            <Input
              id="birthDate"
              type="date"
              value={(data.birthDate as string) || ""}
              onChange={(e) => handleChange("birthDate", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="gender" className="text-slate-300">Gender</Label>
            <select
              id="gender"
              value={(data.gender as string) || ""}
              onChange={(e) => handleChange("gender", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {GENDER_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="maritalStatus" className="text-slate-300">Marital Status</Label>
            <select
              id="maritalStatus"
              value={(data.maritalStatus as string) || ""}
              onChange={(e) => handleChange("maritalStatus", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {MARITAL_STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="nationality" className="text-slate-300">Nationality</Label>
            <Input
              id="nationality"
              placeholder="Mexican"
              value={(data.nationality as string) || "Mexican"}
              onChange={(e) => handleChange("nationality", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="birthState" className="text-slate-300">Birth State</Label>
            <select
              id="birthState"
              value={(data.birthState as string) || ""}
              onChange={(e) => handleChange("birthState", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              <option value="">Select...</option>
              {MEXICAN_STATES.map((state) => (
                <option key={state} value={state}>{state}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="educationLevel" className="text-slate-300">Education Level</Label>
            <select
              id="educationLevel"
              value={(data.educationLevel as string) || ""}
              onChange={(e) => handleChange("educationLevel", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {EDUCATION_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Identification */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Official Identification</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div>
            <Label htmlFor="curp" className="text-slate-300">
              CURP <span className="text-red-400">*</span>
            </Label>
            <Input
              id="curp"
              placeholder="ABCD123456HDFXXX00"
              value={(data.curp as string) || ""}
              onChange={(e) => handleChange("curp", e.target.value.toUpperCase())}
              maxLength={18}
              className="bg-slate-800 border-slate-700 text-white uppercase"
            />
          </div>
          <div>
            <Label htmlFor="rfc" className="text-slate-300">
              RFC <span className="text-red-400">*</span>
            </Label>
            <Input
              id="rfc"
              placeholder="ABCD123456XXX"
              value={(data.rfc as string) || ""}
              onChange={(e) => handleChange("rfc", e.target.value.toUpperCase())}
              maxLength={13}
              className="bg-slate-800 border-slate-700 text-white uppercase"
            />
          </div>
          <div>
            <Label htmlFor="nss" className="text-slate-300">NSS (IMSS)</Label>
            <Input
              id="nss"
              placeholder="12345678901"
              value={(data.nss as string) || ""}
              onChange={(e) => handleChange("nss", e.target.value.replace(/\D/g, ""))}
              maxLength={11}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="infonavitCredit" className="text-slate-300">Infonavit Credit</Label>
            <Input
              id="infonavitCredit"
              placeholder="Credit number"
              value={(data.infonavitCredit as string) || ""}
              onChange={(e) => handleChange("infonavitCredit", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Contact */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Contact Information</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="personalEmail" className="text-slate-300">Email</Label>
            <Input
              id="personalEmail"
              type="email"
              placeholder="juan@example.com"
              value={(data.personalEmail as string) || ""}
              onChange={(e) => handleChange("personalEmail", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="personalPhone" className="text-slate-300">Home Phone</Label>
            <Input
              id="personalPhone"
              placeholder="444 123 4567"
              value={(data.personalPhone as string) || ""}
              onChange={(e) => handleChange("personalPhone", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="cellPhone" className="text-slate-300">Cell Phone</Label>
            <Input
              id="cellPhone"
              placeholder="444 987 6543"
              value={(data.cellPhone as string) || ""}
              onChange={(e) => handleChange("cellPhone", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Address */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Address</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="md:col-span-2">
            <Label htmlFor="street" className="text-slate-300">Street</Label>
            <Input
              id="street"
              placeholder="Av. Reforma"
              value={(data.street as string) || ""}
              onChange={(e) => handleChange("street", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div>
              <Label htmlFor="exteriorNumber" className="text-slate-300">Ext. No.</Label>
              <Input
                id="exteriorNumber"
                placeholder="123"
                value={(data.exteriorNumber as string) || ""}
                onChange={(e) => handleChange("exteriorNumber", e.target.value)}
                className="bg-slate-800 border-slate-700 text-white"
              />
            </div>
            <div>
              <Label htmlFor="interiorNumber" className="text-slate-300">Int. No.</Label>
              <Input
                id="interiorNumber"
                placeholder="A"
                value={(data.interiorNumber as string) || ""}
                onChange={(e) => handleChange("interiorNumber", e.target.value)}
                className="bg-slate-800 border-slate-700 text-white"
              />
            </div>
          </div>
          <div>
            <Label htmlFor="neighborhood" className="text-slate-300">Neighborhood</Label>
            <Input
              id="neighborhood"
              placeholder="Centro"
              value={(data.neighborhood as string) || ""}
              onChange={(e) => handleChange("neighborhood", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="municipality" className="text-slate-300">Municipality</Label>
            <Input
              id="municipality"
              placeholder="San Luis Potosí"
              value={(data.municipality as string) || ""}
              onChange={(e) => handleChange("municipality", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="state" className="text-slate-300">State</Label>
            <select
              id="state"
              value={(data.state as string) || "San Luis Potosí"}
              onChange={(e) => handleChange("state", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {MEXICAN_STATES.map((state) => (
                <option key={state} value={state}>{state}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="postalCode" className="text-slate-300">Postal Code</Label>
            <Input
              id="postalCode"
              placeholder="78000"
              value={(data.postalCode as string) || ""}
              onChange={(e) => handleChange("postalCode", e.target.value.replace(/\D/g, ""))}
              maxLength={5}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="country" className="text-slate-300">Country</Label>
            <Input
              id="country"
              placeholder="Mexico"
              value={(data.country as string) || "Mexico"}
              onChange={(e) => handleChange("country", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>
    </div>
  )
}
