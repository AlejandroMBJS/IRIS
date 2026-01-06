"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select } from "@/components/ui/select"

export function EmployeeForm({ data, onChange }: any) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="firstName">Nombre</Label>
          <Input id="firstName" placeholder="Juan" />
        </div>
        <div>
          <Label htmlFor="lastName">Apellido</Label>
          <Input id="lastName" placeholder="Pérez" />
        </div>
      </div>
      <div>
        <Label htmlFor="email">Correo Electrónico</Label>
        <Input id="email" type="email" placeholder="juan@empresa.com" />
      </div>
      <div>
        <Label htmlFor="curp">CURP</Label>
        <Input id="curp" placeholder="ABCD123456HDFXXX00" />
      </div>
      <div>
        <Label htmlFor="rfc">RFC</Label>
        <Input id="rfc" placeholder="ABCD123456XXX" />
      </div>
    </div>
  )
}
