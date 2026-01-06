"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function BankInfoForm({ data, onChange }: any) {
  return (
    <div className="space-y-4">
      <div>
        <Label htmlFor="bankName">Banco</Label>
        <Input id="bankName" placeholder="BBVA" />
      </div>
      <div>
        <Label htmlFor="accountNumber">Número de Cuenta</Label>
        <Input id="accountNumber" placeholder="1234567890" />
      </div>
      <div>
        <Label htmlFor="clabe">CLABE</Label>
        <Input id="clabe" placeholder="012345678901234567" maxLength={18} />
      </div>
    </div>
  )
}
