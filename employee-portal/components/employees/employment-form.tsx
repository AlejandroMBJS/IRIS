"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function EmploymentForm({ data, onChange }: any) {
  const handleChange = (field: string, value: any) => {
    onChange({ [field]: value })
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="position">Puesto</Label>
          <Input
            id="position"
            placeholder="Desarrollador"
            value={data.position || ""}
            onChange={(e) => handleChange("position", e.target.value)}
          />
        </div>
        <div>
          <Label htmlFor="department">Departamento</Label>
          <Input
            id="department"
            placeholder="TI"
            value={data.department || ""}
            onChange={(e) => handleChange("department", e.target.value)}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="collarType">Tipo de Collar</Label>
          <select
            id="collarType"
            className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-900 text-slate-200"
            value={data.collarType || "white_collar"}
            onChange={(e) => {
              const value = e.target.value
              handleChange("collarType", value)
              // Auto-set pay frequency and sindicalizado based on collar type
              if (value === "white_collar") {
                handleChange("paymentFrequency", "biweekly")
                handleChange("isSindicalizado", false)
              } else if (value === "blue_collar") {
                handleChange("paymentFrequency", "weekly")
                handleChange("isSindicalizado", true)
              } else if (value === "gray_collar") {
                handleChange("paymentFrequency", "weekly")
                handleChange("isSindicalizado", false)
              }
            }}
          >
            <option value="white_collar">White Collar (Administrativo)</option>
            <option value="blue_collar">Blue Collar (Obrero Sindicalizado)</option>
            <option value="gray_collar">Gray Collar (Obrero No Sindicalizado)</option>
          </select>
        </div>
        <div>
          <Label htmlFor="paymentFrequency">Frecuencia de Pago</Label>
          <select
            id="paymentFrequency"
            className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-900 text-slate-200"
            value={data.paymentFrequency || "biweekly"}
            onChange={(e) => handleChange("paymentFrequency", e.target.value)}
          >
            <option value="weekly">Semanal</option>
            <option value="biweekly">Quincenal</option>
            <option value="monthly">Mensual</option>
          </select>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="isSindicalizado"
          className="h-4 w-4 rounded border-slate-700 bg-slate-900"
          checked={data.isSindicalizado || false}
          onChange={(e) => handleChange("isSindicalizado", e.target.checked)}
        />
        <Label htmlFor="isSindicalizado">Sindicalizado</Label>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="salary">Salario Diario</Label>
          <Input
            id="salary"
            type="number"
            placeholder="500.00"
            value={data.dailySalary || ""}
            onChange={(e) => handleChange("dailySalary", e.target.value)}
          />
        </div>
        <div>
          <Label htmlFor="hireDate">Fecha de Contratación</Label>
          <Input
            id="hireDate"
            type="date"
            value={data.hireDate || ""}
            onChange={(e) => handleChange("hireDate", e.target.value)}
          />
        </div>
      </div>
    </div>
  )
}
