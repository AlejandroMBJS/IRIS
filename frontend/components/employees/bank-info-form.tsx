"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface BankInfoFormProps {
  data: Record<string, unknown>
  onChange: (data: Record<string, unknown>) => void
}

const BANK_OPTIONS = [
  { value: "", label: "Seleccionar..." },
  { value: "BBVA", label: "BBVA México" },
  { value: "Santander", label: "Santander" },
  { value: "Banorte", label: "Banorte" },
  { value: "Citibanamex", label: "Citibanamex" },
  { value: "HSBC", label: "HSBC" },
  { value: "Scotiabank", label: "Scotiabank" },
  { value: "Inbursa", label: "Inbursa" },
  { value: "BanCoppel", label: "BanCoppel" },
  { value: "Azteca", label: "Banco Azteca" },
  { value: "Afirme", label: "Afirme" },
  { value: "BanBajio", label: "BanBajío" },
  { value: "Banregio", label: "Banregio" },
  { value: "Mifel", label: "Banca Mifel" },
  { value: "Multiva", label: "Multiva" },
  { value: "Intercam", label: "Intercam Banco" },
  { value: "CIBanco", label: "CIBanco" },
  { value: "other", label: "Otro" },
]

export function BankInfoForm({ data, onChange }: BankInfoFormProps) {
  const handleChange = (field: string, value: unknown) => {
    onChange({ [field]: value })
  }

  const formatCLABE = (value: string) => {
    // Remove non-digits and limit to 18 characters
    return value.replace(/\D/g, "").slice(0, 18)
  }

  const formatBankAccount = (value: string) => {
    // Remove non-digits and limit to 20 characters
    return value.replace(/\D/g, "").slice(0, 20)
  }

  return (
    <div className="space-y-6">
      {/* Bank Information */}
      <div>
        <h3 className="text-lg font-medium text-slate-200 mb-4">Datos Bancarios</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label htmlFor="bankName" className="text-slate-300">Banco</Label>
            <select
              id="bankName"
              value={(data.bankName as string) || ""}
              onChange={(e) => handleChange("bankName", e.target.value)}
              className="w-full h-10 px-3 rounded-md border border-slate-700 bg-slate-800 text-white"
            >
              {BANK_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="bankAccount" className="text-slate-300">Número de Cuenta</Label>
            <Input
              id="bankAccount"
              placeholder="0123456789"
              value={(data.bankAccount as string) || ""}
              onChange={(e) => handleChange("bankAccount", formatBankAccount(e.target.value))}
              className="bg-slate-800 border-slate-700 text-white"
            />
            <p className="text-xs text-slate-500 mt-1">10-20 dígitos</p>
          </div>
          <div>
            <Label htmlFor="clabe" className="text-slate-300">CLABE Interbancaria</Label>
            <Input
              id="clabe"
              placeholder="012345678901234567"
              value={(data.clabe as string) || ""}
              onChange={(e) => handleChange("clabe", formatCLABE(e.target.value))}
              maxLength={18}
              className="bg-slate-800 border-slate-700 text-white font-mono"
            />
            <p className="text-xs text-slate-500 mt-1">18 dígitos</p>
          </div>
        </div>
      </div>

      {/* CLABE Validation Info */}
      {(data.clabe as string)?.length === 18 && (
        <div className="p-4 bg-slate-800/50 rounded-lg border border-slate-700">
          <h4 className="text-sm font-medium text-slate-300 mb-2">Desglose CLABE</h4>
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div>
              <span className="text-slate-400">Código Banco:</span>
              <span className="ml-2 text-white font-mono">{(data.clabe as string).slice(0, 3)}</span>
            </div>
            <div>
              <span className="text-slate-400">Código Plaza:</span>
              <span className="ml-2 text-white font-mono">{(data.clabe as string).slice(3, 6)}</span>
            </div>
            <div>
              <span className="text-slate-400">Cuenta:</span>
              <span className="ml-2 text-white font-mono">{(data.clabe as string).slice(6, 17)}</span>
            </div>
          </div>
        </div>
      )}

      {/* Additional Info */}
      <div className="p-4 bg-blue-500/10 border border-blue-500/30 rounded-lg">
        <p className="text-sm text-blue-400">
          <strong>Nota:</strong> Asegúrate de verificar los datos bancarios con el empleado.
          Una CLABE incorrecta puede causar rechazos en los depósitos de nómina.
        </p>
      </div>
    </div>
  )
}
