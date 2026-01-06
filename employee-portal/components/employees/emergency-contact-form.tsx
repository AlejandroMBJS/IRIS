"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function EmergencyContactForm({ data, onChange }: any) {
  return (
    <div className="space-y-4">
      <div>
        <Label htmlFor="contactName">Nombre del Contacto</Label>
        <Input id="contactName" placeholder="María Pérez" />
      </div>
      <div>
        <Label htmlFor="relationship">Relación</Label>
        <Input id="relationship" placeholder="Esposa" />
      </div>
      <div>
        <Label htmlFor="contactPhone">Teléfono</Label>
        <Input id="contactPhone" placeholder="555-1234567" />
      </div>
    </div>
  )
}
