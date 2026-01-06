/**
 * @file components/dynamic-form-fields.tsx
 * @description Dynamic form field renderer for incidence request forms
 *
 * USER PERSPECTIVE:
 *   - Renders form fields dynamically based on configuration from backend
 *   - Supports various field types: text, number, date, time, select, etc.
 *   - Validates required fields and field constraints
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Add new field types, styling, validation
 *   CAUTION: Field types must match backend form_fields config
 */

"use client"

import { useState, useEffect } from "react"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { FormField, Shift, shiftApi } from "@/lib/api-client"
import { Loader2 } from "lucide-react"

interface DynamicFormFieldsProps {
  fields: FormField[]
  values: Record<string, unknown>
  onChange: (values: Record<string, unknown>) => void
  disabled?: boolean
}

export function DynamicFormFields({
  fields,
  values,
  onChange,
  disabled = false,
}: DynamicFormFieldsProps) {
  const [shifts, setShifts] = useState<Shift[]>([])
  const [loadingShifts, setLoadingShifts] = useState(false)

  // Load shifts if any field requires shift_select
  useEffect(() => {
    const hasShiftSelect = fields.some((f) => f.type === "shift_select")
    if (hasShiftSelect && shifts.length === 0) {
      loadShifts()
    }
  }, [fields, shifts.length])

  async function loadShifts() {
    try {
      setLoadingShifts(true)
      const data = await shiftApi.getAll(true)
      setShifts(data)
    } catch (error) {
      console.error("Failed to load shifts:", error)
    } finally {
      setLoadingShifts(false)
    }
  }

  const handleChange = (fieldName: string, value: unknown) => {
    onChange({ ...values, [fieldName]: value })
  }

  // Sort fields by display_order
  const sortedFields = [...fields].sort((a, b) => a.display_order - b.display_order)

  if (sortedFields.length === 0) {
    return null
  }

  return (
    <div className="space-y-4">
      {sortedFields.map((field) => (
        <div key={field.name} className="space-y-2">
          <Label htmlFor={field.name} className="text-slate-300">
            {field.label}
            {field.required && <span className="text-red-400 ml-1">*</span>}
          </Label>

          {/* Text input */}
          {field.type === "text" && (
            <Input
              id={field.name}
              value={(values[field.name] as string) || ""}
              onChange={(e) => handleChange(field.name, e.target.value)}
              placeholder={field.placeholder}
              className="bg-slate-800/50 border-slate-700 text-white"
              disabled={disabled}
              required={field.required}
            />
          )}

          {/* Textarea */}
          {field.type === "textarea" && (
            <Textarea
              id={field.name}
              value={(values[field.name] as string) || ""}
              onChange={(e) => handleChange(field.name, e.target.value)}
              placeholder={field.placeholder}
              className="bg-slate-800/50 border-slate-700 text-white"
              disabled={disabled}
              required={field.required}
              rows={3}
            />
          )}

          {/* Number input */}
          {field.type === "number" && (
            <Input
              id={field.name}
              type="number"
              value={(values[field.name] as number) ?? ""}
              onChange={(e) =>
                handleChange(field.name, e.target.value ? parseFloat(e.target.value) : undefined)
              }
              placeholder={field.placeholder}
              className="bg-slate-800/50 border-slate-700 text-white"
              disabled={disabled}
              required={field.required}
              min={field.min}
              max={field.max}
              step={field.step}
            />
          )}

          {/* Date input */}
          {field.type === "date" && (
            <Input
              id={field.name}
              type="date"
              value={(values[field.name] as string) || ""}
              onChange={(e) => handleChange(field.name, e.target.value)}
              className="bg-slate-800/50 border-slate-700 text-white"
              disabled={disabled}
              required={field.required}
            />
          )}

          {/* Time input */}
          {field.type === "time" && (
            <Input
              id={field.name}
              type="time"
              value={(values[field.name] as string) || ""}
              onChange={(e) => handleChange(field.name, e.target.value)}
              className="bg-slate-800/50 border-slate-700 text-white"
              disabled={disabled}
              required={field.required}
            />
          )}

          {/* Boolean switch */}
          {field.type === "boolean" && (
            <div className="flex items-center gap-3">
              <Switch
                id={field.name}
                checked={Boolean(values[field.name])}
                onCheckedChange={(checked) => handleChange(field.name, checked)}
                disabled={disabled}
              />
              <span className="text-slate-400 text-sm">
                {values[field.name] ? "Si" : "No"}
              </span>
            </div>
          )}

          {/* Select dropdown */}
          {field.type === "select" && field.options && (
            <Select
              value={(values[field.name] as string) || ""}
              onValueChange={(value) => handleChange(field.name, value)}
              disabled={disabled}
            >
              <SelectTrigger className="bg-slate-800/50 border-slate-700 text-white">
                <SelectValue placeholder={field.placeholder || "Seleccionar..."} />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                {field.options.map((option) => (
                  <SelectItem
                    key={option.value}
                    value={option.value}
                    className="text-white hover:bg-slate-700"
                  >
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {/* Shift selector */}
          {field.type === "shift_select" && (
            <>
              {loadingShifts ? (
                <div className="flex items-center gap-2 p-3 bg-slate-800/50 rounded-lg">
                  <Loader2 className="h-4 w-4 animate-spin text-blue-400" />
                  <span className="text-slate-400 text-sm">Cargando turnos...</span>
                </div>
              ) : shifts.length > 0 ? (
                <Select
                  value={(values[field.name] as string) || ""}
                  onValueChange={(value) => handleChange(field.name, value)}
                  disabled={disabled}
                >
                  <SelectTrigger className="bg-slate-800/50 border-slate-700 text-white">
                    <SelectValue placeholder="Selecciona un turno" />
                  </SelectTrigger>
                  <SelectContent className="bg-slate-800 border-slate-700">
                    {shifts.map((shift) => (
                      <SelectItem
                        key={shift.id}
                        value={shift.id}
                        className="text-white hover:bg-slate-700"
                      >
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{shift.name}</span>
                          <span className="text-slate-400 text-sm">
                            ({shift.start_time} - {shift.end_time})
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <div className="p-3 bg-amber-500/10 border border-amber-500/30 rounded-lg text-amber-400 text-sm">
                  No hay turnos disponibles.
                </div>
              )}
            </>
          )}

          {/* Help text */}
          {field.help_text && (
            <p className="text-xs text-slate-400">{field.help_text}</p>
          )}
        </div>
      ))}
    </div>
  )
}
