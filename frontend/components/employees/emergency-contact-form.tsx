"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface EmergencyContactFormProps {
  data: Record<string, unknown>
  onChange: (data: Record<string, unknown>) => void
}

const RELATIONSHIP_OPTIONS = [
  { value: "", label: "Seleccionar..." },
  { value: "spouse", label: "Esposo/a" },
  { value: "parent", label: "Padre/Madre" },
  { value: "sibling", label: "Hermano/a" },
  { value: "child", label: "Hijo/a" },
  { value: "uncle_aunt", label: "Tío/a" },
  { value: "grandparent", label: "Abuelo/a" },
  { value: "cousin", label: "Primo/a" },
  { value: "friend", label: "Amigo/a" },
  { value: "neighbor", label: "Vecino/a" },
  { value: "other", label: "Otro" },
]

export function EmergencyContactForm({ data, onChange }: EmergencyContactFormProps) {
  const handleChange = (field: string, value: unknown) => {
    onChange({ [field]: value })
  }

  const formatPhone = (value: string) => {
    // Remove non-digits
    const digits = value.replace(/\D/g, "")
    // Format as XXX XXX XXXX if enough digits
    if (digits.length <= 3) return digits
    if (digits.length <= 6) return `${digits.slice(0, 3)} ${digits.slice(3)}`
    return `${digits.slice(0, 3)} ${digits.slice(3, 6)} ${digits.slice(6, 10)}`
  }

  return (
    <div className="space-y-6">
      {/* Emergency Contact */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Contacto de Emergencia</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="emergencyContact" className="text-slate-300">
              Nombre Completo
            </Label>
            <Input
              id="emergencyContact"
              placeholder="María Pérez González"
              value={(data.emergencyContact as string) || ""}
              onChange={(e) => handleChange("emergencyContact", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
          <div>
            <Label htmlFor="emergencyRelationship" className="text-slate-300">Parentesco</Label>
            <select
              id="emergencyRelationship"
              value={(data.emergencyRelationship as string) || ""}
              onChange={(e) => handleChange("emergencyRelationship", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {RELATIONSHIP_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="emergencyPhone" className="text-slate-300">Teléfono</Label>
            <Input
              id="emergencyPhone"
              placeholder="444 123 4567"
              value={(data.emergencyPhone as string) || ""}
              onChange={(e) => handleChange("emergencyPhone", formatPhone(e.target.value))}
              className="bg-slate-800 border-slate-700 text-white"
            />
          </div>
        </div>
      </div>

      {/* Important Notice */}
      <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-lg">
        <p className="text-sm text-amber-400">
          <strong>Importante:</strong> Este contacto será notificado en caso de emergencia médica o accidente laboral.
          Asegúrate de que la persona esté informada y que el número telefónico esté actualizado.
        </p>
      </div>

      {/* Additional Notes */}
      <div className="p-4 bg-slate-800/50 rounded-lg border border-slate-700">
        <h4 className="text-sm font-medium text-slate-300 mb-2">Recomendaciones</h4>
        <ul className="text-sm text-slate-400 space-y-1 list-disc list-inside">
          <li>Preferentemente registrar un familiar directo</li>
          <li>El contacto debe estar disponible durante horario laboral</li>
          <li>Actualizar esta información si cambia el contacto o número</li>
        </ul>
      </div>
    </div>
  )
}
