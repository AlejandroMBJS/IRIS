"use client"

import { useState, useEffect } from "react"
import { Input, InputProps } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"

interface MexicanInputProps extends Omit<InputProps, "onChange" | "value"> {
  value: string
  onChange: (value: string) => void
  label?: string
  error?: string
}

// RFC Validation and Formatting
export function RFCInput({ value, onChange, label = "RFC", error, ...props }: MexicanInputProps) {
  const formatRFC = (input: string) => {
    // Remove all non-alphanumeric characters
    let cleaned = input.replace(/[^a-zA-Z0-9]/g, '').toUpperCase()
    
    // Format based on length
    if (cleaned.length <= 3) {
      return cleaned
    } else if (cleaned.length <= 10) {
      return `${cleaned.slice(0, 3)}${cleaned.slice(3, 10)}`
    } else {
      return `${cleaned.slice(0, 3)}${cleaned.slice(3, 10)}${cleaned.slice(10, 13)}`
    }
  }

  const validateRFC = (rfc: string) => {
    const regex = /^[A-ZÑ&]{3,4}\d{6}[A-Z0-9]{2}[A-Z0-9]?$/
    return regex.test(rfc)
  }

  const isValid = validateRFC(value)

  return (
    <div className="space-y-2">
      {label && <Label className="text-slate-300">{label} *</Label>}
      <Input
        {...props}
        value={value}
        onChange={(e) => onChange(formatRFC(e.target.value))}
        className={cn(
          "bg-slate-800 border-slate-700 text-white",
          isValid && "border-green-500",
          error && "border-red-500"
        )}
        placeholder="ABC123456XYZ"
        maxLength={13}
      />
      {error && <p className="text-sm text-red-500">{error}</p>}
      {!error && value && !isValid && (
        <p className="text-sm text-amber-500">RFC no válido</p>
      )}
    </div>
  )
}

// CURP Input Component
export function CURPInput({ value, onChange, label = "CURP", error, ...props }: MexicanInputProps) {
  const formatCURP = (input: string) => {
    return input.replace(/[^a-zA-Z0-9]/g, '').toUpperCase()
  }

  const validateCURP = (curp: string) => {
    const regex = /^[A-Z][AEIOUX][A-Z]{2}\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])[HM](AS|BC|BS|CC|CL|CM|CS|CH|DF|DG|GT|GR|HG|JC|MC|MN|MS|NT|NL|OC|PL|QT|QR|SP|SL|SR|TC|TS|TL|VZ|YN|ZS|NE)[B-DF-HJ-NP-TV-Z]{3}[A-Z\d]\d$/
    return regex.test(curp)
  }

  const isValid = validateCURP(value)

  return (
    <div className="space-y-2">
      {label && <Label className="text-slate-300">{label} *</Label>}
      <Input
        {...props}
        value={value}
        onChange={(e) => onChange(formatCURP(e.target.value))}
        className={cn(
          "bg-slate-800 border-slate-700 text-white",
          isValid && "border-green-500",
          error && "border-red-500"
        )}
        placeholder="ABCD123456HXYZ789"
        maxLength={18}
      />
      {error && <p className="text-sm text-red-500">{error}</p>}
      {!error && value && !isValid && (
        <p className="text-sm text-amber-500">CURP no válido</p>
      )}
    </div>
  )
}

// NSS (Social Security Number) Input
export function NSSInput({ value, onChange, label = "NSS", error, ...props }: MexicanInputProps) {
  const formatNSS = (input: string) => {
    let cleaned = input.replace(/\D/g, '')
    if (cleaned.length <= 11) {
      return cleaned
    }
    return cleaned.slice(0, 11)
  }

  const validateNSS = (nss: string) => {
    if (nss.length !== 11) return false
    
    // Basic validation for Mexican NSS format
    const subDelegation = parseInt(nss.slice(0, 2))
    const year = parseInt(nss.slice(2, 4))
    
    return subDelegation >= 1 && subDelegation <= 99 && year >= 0 && year <= 99
  }

  const isValid = validateNSS(value)

  return (
    <div className="space-y-2">
      {label && <Label className="text-slate-300">{label} *</Label>}
      <Input
        {...props}
        value={value}
        onChange={(e) => onChange(formatNSS(e.target.value))}
        className={cn(
          "bg-slate-800 border-slate-700 text-white",
          isValid && "border-green-500",
          error && "border-red-500"
        )}
        placeholder="12345678901"
        maxLength={11}
      />
      {error && <p className="text-sm text-red-500">{error}</p>}
      {!error && value && !isValid && (
        <p className="text-sm text-amber-500">NSS no válido (debe tener 11 dígitos)</p>
      )}
    </div>
  )
}

// CLABE Bank Account Input
export function CLABEInput({ value, onChange, label = "CLABE Bancaria", error, ...props }: MexicanInputProps) {
  const formatCLABE = (input: string) => {
    return input.replace(/\D/g, '').slice(0, 18)
  }

  const validateCLABE = (clabe: string) => {
    if (clabe.length !== 18) return false
    
    // Simple validation (could implement full CLABE validation)
    const bankCode = parseInt(clabe.slice(0, 3))
    return bankCode >= 1 && bankCode <= 999
  }

  const isValid = validateCLABE(value)

  return (
    <div className="space-y-2">
      {label && <Label className="text-slate-300">{label}</Label>}
      <Input
        {...props}
        value={value}
        onChange={(e) => onChange(formatCLABE(e.target.value))}
        className={cn(
          "bg-slate-800 border-slate-700 text-white",
          isValid && "border-green-500",
          error && "border-red-500"
        )}
        placeholder="123456789012345678"
        maxLength={18}
      />
      {error && <p className="text-sm text-red-500">{error}</p>}
      {!error && value && !isValid && (
        <p className="text-sm text-amber-500">CLABE no válida (debe tener 18 dígitos)</p>
      )}
    </div>
  )
}

// Mexican Phone Input
export function MexicanPhoneInput({ value, onChange, label = "Teléfono", error, ...props }: MexicanInputProps) {
  const formatPhone = (input: string) => {
    let cleaned = input.replace(/\D/g, '')
    
    if (cleaned.length === 10) {
      return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`
    }
    return cleaned
  }

  const validatePhone = (phone: string) => {
    const cleaned = phone.replace(/\D/g, '')
    return cleaned.length === 10
  }

  const isValid = validatePhone(value)

  return (
    <div className="space-y-2">
      {label && <Label className="text-slate-300">{label} *</Label>}
      <Input
        {...props}
        value={value}
        onChange={(e) => onChange(formatPhone(e.target.value))}
        className={cn(
          "bg-slate-800 border-slate-700 text-white",
          isValid && "border-green-500",
          error && "border-red-500"
        )}
        placeholder="(55) 1234-5678"
        maxLength={15}
      />
      {error && <p className="text-sm text-red-500">{error}</p>}
      {!error && value && !isValid && (
        <p className="text-sm text-amber-500">Teléfono no válido (debe tener 10 dígitos)</p>
      )}
    </div>
  )
}

// MXN Currency Input
export function CurrencyInput({ value, onChange, label, error, ...props }: MexicanInputProps) {
  const formatCurrency = (input: string) => {
    // Remove all non-numeric characters except decimal point
    let cleaned = input.replace(/[^\d.]/g, '')
    
    // Only allow one decimal point
    const parts = cleaned.split('.')
    if (parts.length > 2) {
      cleaned = parts[0] + '.' + parts.slice(1).join('')
    }
    
    // Limit to 2 decimal places
    if (parts[1] && parts[1].length > 2) {
      cleaned = parts[0] + '.' + parts[1].slice(0, 2)
    }
    
    return cleaned
  }

  const displayValue = value ? `$${parseFloat(value).toLocaleString('es-MX', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })}` : ""

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatCurrency(e.target.value)
    onChange(formatted)
  }

  return (
    <div className="space-y-2">
      {label && <Label className="text-slate-300">{label}</Label>}
      <Input
        {...props}
        value={displayValue}
        onChange={handleChange}
        className={cn(
          "bg-slate-800 border-slate-700 text-white",
          error && "border-red-500"
        )}
        placeholder="$0.00"
      />
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  )
}

// Mexican Address Input Component
interface MexicanAddressInputProps {
  value: {
    street: string
    exteriorNumber: string
    interiorNumber?: string
    neighborhood: string
    municipality: string
    state: string
    postalCode: string
    country?: string
  }
  onChange: (value: any) => void
}

export function MexicanAddressInput({ value, onChange }: MexicanAddressInputProps) {
  const handleChange = (field: string, fieldValue: string) => {
    onChange({
      ...value,
      [field]: fieldValue
    })
  }

  return (
    <div className="space-y-4 p-4 border border-slate-700 rounded-lg bg-slate-800/50">
      <h4 className="font-medium text-slate-300">Domicilio en México</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label className="text-slate-400 text-sm">Calle *</Label>
          <Input
            value={value.street}
            onChange={(e) => handleChange("street", e.target.value)}
            className="bg-slate-800 border-slate-700 text-white"
            placeholder="Av. Principal"
          />
        </div>
        
        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-2">
            <Label className="text-slate-400 text-sm">Núm. Ext. *</Label>
            <Input
              value={value.exteriorNumber}
              onChange={(e) => handleChange("exteriorNumber", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
              placeholder="123"
            />
          </div>
          <div className="space-y-2">
            <Label className="text-slate-400 text-sm">Núm. Int.</Label>
            <Input
              value={value.interiorNumber || ""}
              onChange={(e) => handleChange("interiorNumber", e.target.value)}
              className="bg-slate-800 border-slate-700 text-white"
              placeholder="A"
            />
          </div>
        </div>
        
        <div className="space-y-2">
          <Label className="text-slate-400 text-sm">Colonia *</Label>
          <Input
            value={value.neighborhood}
            onChange={(e) => handleChange("neighborhood", e.target.value)}
            className="bg-slate-800 border-slate-700 text-white"
            placeholder="Centro"
          />
        </div>
        
        <div className="space-y-2">
          <Label className="text-slate-400 text-sm">Código Postal *</Label>
          <Input
            value={value.postalCode}
            onChange={(e) => handleChange("postalCode", e.target.value)}
            className="bg-slate-800 border-slate-700 text-white"
            placeholder="12345"
            maxLength={5}
          />
        </div>
        
        <div className="space-y-2">
          <Label className="text-slate-400 text-sm">Municipio/Alcaldía *</Label>
          <Input
            value={value.municipality}
            onChange={(e) => handleChange("municipality", e.target.value)}
            className="bg-slate-800 border-slate-700 text-white"
            placeholder="Benito Juárez"
          />
        </div>
        
        <div className="space-y-2">
          <Label className="text-slate-400 text-sm">Estado *</Label>
          <select
            value={value.state}
            onChange={(e) => handleChange("state", e.target.value)}
            className="w-full bg-slate-800 border border-slate-700 rounded-md px-3 py-2 text-white"
          >
            <option value="">Seleccionar estado</option>
            {[
              "Aguascalientes", "Baja California", "Baja California Sur", "Campeche",
              "Chiapas", "Chihuahua", "Ciudad de México", "Coahuila", "Colima",
              "Durango", "Estado de México", "Guanajuato", "Guerrero", "Hidalgo",
              "Jalisco", "Michoacán", "Morelos", "Nayarit", "Nuevo León", "Oaxaca",
              "Puebla", "Querétaro", "Quintana Roo", "San Luis Potosí", "Sinaloa",
              "Sonora", "Tabasco", "Tamaulipas", "Tlaxcala", "Veracruz", "Yucatán", "Zacatecas"
            ].map(state => (
              <option key={state} value={state}>{state}</option>
            ))}
          </select>
        </div>
      </div>
    </div>
  )
}
