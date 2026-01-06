/**
 * @file app/configuration/shifts/page.tsx
 * @description Shift configuration page for managing work schedules
 *
 * USER PERSPECTIVE:
 *   - View all configured shifts for the company
 *   - Create new shifts with schedule details
 *   - Edit existing shifts (times, work days, breaks)
 *   - Toggle shift active status
 *   - Delete shifts (only if no employees assigned)
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Add new shift fields, validation rules
 *   CAUTION: Shift code uniqueness, work days array format
 *   DO NOT modify: Shift deletion checks (employee count)
 */

"use client"

import { useState, useEffect } from "react"
import { Plus, Edit, Trash2, Clock, Calendar, Users, Sun, Moon, ToggleLeft, ToggleRight, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { useToast } from "@/hooks/use-toast"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { shiftApi, ShiftResponse, CreateShiftRequest } from "@/lib/api-client"

const DAYS_OF_WEEK = [
  { value: 1, label: "Monday" },
  { value: 2, label: "Tuesday" },
  { value: 3, label: "Wednesday" },
  { value: 4, label: "Thursday" },
  { value: 5, label: "Friday" },
  { value: 6, label: "Saturday" },
  { value: 0, label: "Sunday" },
]

const DEFAULT_COLORS = [
  "#F59E0B", // Amber
  "#3B82F6", // Blue
  "#8B5CF6", // Purple
  "#10B981", // Green
  "#EF4444", // Red
  "#EC4899", // Pink
  "#06B6D4", // Cyan
  "#F97316", // Orange
]

const COLLAR_TYPES = [
  { value: "white_collar", label: "White Collar", description: "Administrative, office" },
  { value: "blue_collar", label: "Blue Collar", description: "Operations, plant" },
  { value: "gray_collar", label: "Gray Collar", description: "Technicians, maintenance" },
]

interface ShiftFormData {
  name: string
  code: string
  description: string
  start_time: string
  end_time: string
  break_minutes: number
  break_start_time: string
  work_hours_per_day: number
  work_days: number[]
  color: string
  display_order: number
  is_active: boolean
  is_night_shift: boolean
  collar_types: string[]
}

const DEFAULT_FORM_DATA: ShiftFormData = {
  name: "",
  code: "",
  description: "",
  start_time: "08:00",
  end_time: "17:00",
  break_minutes: 60,
  break_start_time: "13:00",
  work_hours_per_day: 8,
  work_days: [1, 2, 3, 4, 5], // Mon-Fri
  color: "#3B82F6",
  display_order: 0,
  is_active: true,
  is_night_shift: false,
  collar_types: [], // Empty = available to all
}

export default function ShiftsConfigurationPage() {
  const { toast } = useToast()
  const [shifts, setShifts] = useState<ShiftResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [editingShift, setEditingShift] = useState<ShiftResponse | null>(null)
  const [shiftToDelete, setShiftToDelete] = useState<ShiftResponse | null>(null)
  const [formData, setFormData] = useState<ShiftFormData>(DEFAULT_FORM_DATA)

  // Fetch shifts on mount
  useEffect(() => {
    fetchShifts()
  }, [])

  const fetchShifts = async () => {
    try {
      setIsLoading(true)
      const data = await shiftApi.getAll()
      setShifts(data)
    } catch (error) {
      toast({
        title: "Error",
        description: "Could not load shifts",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleOpenCreate = () => {
    setEditingShift(null)
    setFormData({
      ...DEFAULT_FORM_DATA,
      display_order: shifts.length + 1,
    })
    setIsDialogOpen(true)
  }

  const handleOpenEdit = (shift: ShiftResponse) => {
    setEditingShift(shift)
    // Parse collar_types from JSON string
    let collarTypes: string[] = []
    try {
      if (shift.collar_types && shift.collar_types !== "[]") {
        collarTypes = JSON.parse(shift.collar_types)
      }
    } catch {
      collarTypes = []
    }
    setFormData({
      name: shift.name,
      code: shift.code,
      description: shift.description || "",
      start_time: shift.start_time,
      end_time: shift.end_time,
      break_minutes: shift.break_minutes,
      break_start_time: shift.break_start_time || "",
      work_hours_per_day: shift.work_hours_per_day,
      work_days: JSON.parse(shift.work_days || "[1,2,3,4,5]"),
      color: shift.color,
      display_order: shift.display_order,
      is_active: shift.is_active,
      is_night_shift: shift.is_night_shift,
      collar_types: collarTypes,
    })
    setIsDialogOpen(true)
  }

  const handleSave = async () => {
    if (!formData.name.trim()) {
      toast({
        title: "Error",
        description: "Shift name is required",
        variant: "destructive",
      })
      return
    }
    if (!formData.code.trim()) {
      toast({
        title: "Error",
        description: "Shift code is required",
        variant: "destructive",
      })
      return
    }

    try {
      setIsSaving(true)
      const request: CreateShiftRequest = {
        name: formData.name.trim(),
        code: formData.code.trim().toUpperCase(),
        description: formData.description.trim(),
        start_time: formData.start_time,
        end_time: formData.end_time,
        break_minutes: formData.break_minutes,
        break_start_time: formData.break_start_time || undefined,
        work_hours_per_day: formData.work_hours_per_day,
        work_days: formData.work_days,
        color: formData.color,
        display_order: formData.display_order,
        is_active: formData.is_active,
        is_night_shift: formData.is_night_shift,
        collar_types: formData.collar_types.length > 0 ? formData.collar_types : undefined,
      }

      if (editingShift) {
        await shiftApi.update(editingShift.id, request)
        toast({
          title: "Shift updated",
          description: `The shift "${formData.name}" has been updated`,
        })
      } else {
        await shiftApi.create(request)
        toast({
          title: "Shift created",
          description: `The shift "${formData.name}" has been created`,
        })
      }

      setIsDialogOpen(false)
      fetchShifts()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not save shift",
        variant: "destructive",
      })
    } finally {
      setIsSaving(false)
    }
  }

  const handleConfirmDelete = async () => {
    if (!shiftToDelete) return

    try {
      await shiftApi.delete(shiftToDelete.id)
      toast({
        title: "Shift deleted",
        description: `The shift "${shiftToDelete.name}" has been deleted`,
      })
      setIsDeleteDialogOpen(false)
      setShiftToDelete(null)
      fetchShifts()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not delete shift",
        variant: "destructive",
      })
    }
  }

  const handleToggleActive = async (shift: ShiftResponse) => {
    try {
      await shiftApi.toggleActive(shift.id)
      toast({
        title: shift.is_active ? "Shift deactivated" : "Shift activated",
        description: `The shift "${shift.name}" has been ${shift.is_active ? "deactivated" : "activated"}`,
      })
      fetchShifts()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not change shift status",
        variant: "destructive",
      })
    }
  }

  const handleWorkDayToggle = (day: number) => {
    setFormData(prev => {
      const days = prev.work_days.includes(day)
        ? prev.work_days.filter(d => d !== day)
        : [...prev.work_days, day].sort((a, b) => a - b)
      return { ...prev, work_days: days }
    })
  }

  const formatWorkDays = (workDaysJson: string) => {
    try {
      const days = JSON.parse(workDaysJson)
      return days.map((d: number) => DAYS_OF_WEEK.find(day => day.value === d)?.label?.[0] || d).join(", ")
    } catch {
      return "M-F"
    }
  }

  const formatCollarTypes = (collarTypesJson: string): string[] => {
    try {
      if (!collarTypesJson || collarTypesJson === "[]") return []
      return JSON.parse(collarTypesJson)
    } catch {
      return []
    }
  }

  const getCollarTypeLabel = (type: string) => {
    return COLLAR_TYPES.find(t => t.value === type)?.label || type
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto py-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Shift Configuration</h1>
            <p className="text-muted-foreground">
              Manage the work shifts available for assigning to employees
            </p>
          </div>
          <Button onClick={handleOpenCreate}>
            <Plus className="h-4 w-4 mr-2" />
            New Shift
          </Button>
        </div>

        {/* Shifts Grid */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : shifts.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Clock className="h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold mb-2">No shifts configured</h3>
              <p className="text-muted-foreground text-center mb-4">
                Create work shifts to assign to your employees
              </p>
              <Button onClick={handleOpenCreate}>
                <Plus className="h-4 w-4 mr-2" />
                Create first shift
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {shifts.map((shift) => (
              <Card key={shift.id} className={`relative ${!shift.is_active ? "opacity-60" : ""}`}>
                <div
                  className="absolute top-0 left-0 right-0 h-1 rounded-t-lg"
                  style={{ backgroundColor: shift.color }}
                />
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: shift.color }}
                      />
                      <CardTitle className="text-lg">{shift.name}</CardTitle>
                    </div>
                    <div className="flex items-center gap-1">
                      {shift.is_night_shift && (
                        <Badge variant="secondary" className="gap-1">
                          <Moon className="h-3 w-3" />
                          Night
                        </Badge>
                      )}
                      <Badge variant={shift.is_active ? "default" : "secondary"}>
                        {shift.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </div>
                  </div>
                  <Badge variant="outline" className="w-fit font-mono">
                    {shift.code}
                  </Badge>
                </CardHeader>
                <CardContent className="space-y-3">
                  {/* Schedule */}
                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">{shift.start_time}</span>
                    <span className="text-muted-foreground">-</span>
                    <span className="font-medium">{shift.end_time}</span>
                    <span className="text-muted-foreground">
                      ({shift.work_hours_per_day}h)
                    </span>
                  </div>

                  {/* Work Days */}
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    <span>{formatWorkDays(shift.work_days)}</span>
                  </div>

                  {/* Break */}
                  {shift.break_minutes > 0 && (
                    <div className="text-sm text-muted-foreground">
                      Break: {shift.break_minutes} min
                      {shift.break_start_time && ` (${shift.break_start_time})`}
                    </div>
                  )}

                  {/* Description */}
                  {shift.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {shift.description}
                    </p>
                  )}

                  {/* Collar Types */}
                  {formatCollarTypes(shift.collar_types).length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {formatCollarTypes(shift.collar_types).map(type => (
                        <Badge key={type} variant="outline" className="text-xs">
                          {getCollarTypeLabel(type)}
                        </Badge>
                      ))}
                    </div>
                  )}

                  {/* Employee Count */}
                  <div className="flex items-center gap-2 text-sm text-muted-foreground pt-2 border-t">
                    <Users className="h-4 w-4" />
                    <span>
                      {shift.employee_count} {shift.employee_count === 1 ? "employee" : "employees"}
                    </span>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center justify-between pt-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleToggleActive(shift)}
                      className="gap-1"
                    >
                      {shift.is_active ? (
                        <>
                          <ToggleRight className="h-4 w-4" />
                          Deactivate
                        </>
                      ) : (
                        <>
                          <ToggleLeft className="h-4 w-4" />
                          Activate
                        </>
                      )}
                    </Button>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenEdit(shift)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          setShiftToDelete(shift)
                          setIsDeleteDialogOpen(true)
                        }}
                        disabled={shift.employee_count > 0}
                        className="text-destructive hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Create/Edit Dialog */}
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>
                {editingShift ? "Edit Shift" : "New Shift"}
              </DialogTitle>
              <DialogDescription>
                {editingShift
                  ? "Modify the work shift details"
                  : "Configure a new work shift to assign to employees"}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6 py-4">
              {/* Basic Info */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name *</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="Morning Shift"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="code">Code *</Label>
                  <Input
                    id="code"
                    value={formData.code}
                    onChange={(e) => setFormData({ ...formData, code: e.target.value.toUpperCase() })}
                    placeholder="T1"
                    className="uppercase"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Shift description..."
                  rows={2}
                />
              </div>

              {/* Schedule */}
              <div className="space-y-4">
                <h4 className="font-medium">Schedule</h4>
                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="start_time">Start Time *</Label>
                    <Input
                      id="start_time"
                      type="time"
                      value={formData.start_time}
                      onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="end_time">End Time *</Label>
                    <Input
                      id="end_time"
                      type="time"
                      value={formData.end_time}
                      onChange={(e) => setFormData({ ...formData, end_time: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="work_hours">Work Hours</Label>
                    <Input
                      id="work_hours"
                      type="number"
                      min="1"
                      max="24"
                      step="0.5"
                      value={formData.work_hours_per_day}
                      onChange={(e) => setFormData({ ...formData, work_hours_per_day: parseFloat(e.target.value) || 8 })}
                    />
                  </div>
                </div>
              </div>

              {/* Break */}
              <div className="space-y-4">
                <h4 className="font-medium">Break</h4>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="break_minutes">Duration (minutes)</Label>
                    <Input
                      id="break_minutes"
                      type="number"
                      min="0"
                      max="120"
                      value={formData.break_minutes}
                      onChange={(e) => setFormData({ ...formData, break_minutes: parseInt(e.target.value) || 0 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="break_start">Start Time</Label>
                    <Input
                      id="break_start"
                      type="time"
                      value={formData.break_start_time}
                      onChange={(e) => setFormData({ ...formData, break_start_time: e.target.value })}
                    />
                  </div>
                </div>
              </div>

              {/* Work Days */}
              <div className="space-y-4">
                <h4 className="font-medium">Work Days</h4>
                <div className="flex flex-wrap gap-3">
                  {DAYS_OF_WEEK.map((day) => (
                    <label
                      key={day.value}
                      className="flex items-center gap-2 cursor-pointer"
                    >
                      <Checkbox
                        checked={formData.work_days.includes(day.value)}
                        onCheckedChange={() => handleWorkDayToggle(day.value)}
                      />
                      <span className="text-sm">{day.label}</span>
                    </label>
                  ))}
                </div>
              </div>

              {/* Color */}
              <div className="space-y-4">
                <h4 className="font-medium">Color</h4>
                <div className="flex flex-wrap gap-2">
                  {DEFAULT_COLORS.map((color) => (
                    <button
                      key={color}
                      type="button"
                      className={`w-8 h-8 rounded-full border-2 transition-all ${
                        formData.color === color
                          ? "border-foreground scale-110"
                          : "border-transparent hover:scale-105"
                      }`}
                      style={{ backgroundColor: color }}
                      onClick={() => setFormData({ ...formData, color })}
                    />
                  ))}
                  <Input
                    type="color"
                    value={formData.color}
                    onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                    className="w-8 h-8 p-0 border-0 cursor-pointer"
                  />
                </div>
              </div>

              {/* Options */}
              <div className="space-y-4">
                <h4 className="font-medium">Options</h4>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Active Shift</Label>
                      <p className="text-sm text-muted-foreground">
                        Inactive shifts do not appear in employee selection
                      </p>
                    </div>
                    <Switch
                      checked={formData.is_active}
                      onCheckedChange={(checked) => setFormData({ ...formData, is_active: checked })}
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Night Shift</Label>
                      <p className="text-sm text-muted-foreground">
                        Mark this shift as night schedule
                      </p>
                    </div>
                    <Switch
                      checked={formData.is_night_shift}
                      onCheckedChange={(checked) => setFormData({ ...formData, is_night_shift: checked })}
                    />
                  </div>
                </div>
              </div>

              {/* Display Order */}
              <div className="space-y-2">
                <Label htmlFor="display_order">Display Order</Label>
                <Input
                  id="display_order"
                  type="number"
                  min="0"
                  value={formData.display_order}
                  onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) || 0 })}
                  className="w-24"
                />
              </div>

              {/* Collar Types */}
              <div className="space-y-3">
                <div>
                  <Label>Employee Types (Collar Type)</Label>
                  <p className="text-sm text-muted-foreground">
                    Select which employee types this shift is available for. If none selected, it will be available for all.
                  </p>
                </div>
                <div className="grid grid-cols-3 gap-3">
                  {COLLAR_TYPES.map((type) => (
                    <label
                      key={type.value}
                      className={`flex items-center gap-2 p-3 border rounded-lg cursor-pointer transition-colors ${
                        formData.collar_types.includes(type.value)
                          ? "border-blue-500 bg-blue-500/10"
                          : "border-slate-200 hover:border-slate-300"
                      }`}
                    >
                      <Checkbox
                        checked={formData.collar_types.includes(type.value)}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            setFormData({ ...formData, collar_types: [...formData.collar_types, type.value] })
                          } else {
                            setFormData({ ...formData, collar_types: formData.collar_types.filter(t => t !== type.value) })
                          }
                        }}
                      />
                      <div>
                        <span className="font-medium text-sm">{type.label}</span>
                        <p className="text-xs text-muted-foreground">{type.description}</p>
                      </div>
                    </label>
                  ))}
                </div>
                {formData.collar_types.length === 0 && (
                  <p className="text-xs text-amber-600 bg-amber-50 p-2 rounded">
                    No selection = available for all employee types
                  </p>
                )}
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleSave} disabled={isSaving}>
                {isSaving ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  editingShift ? "Save Changes" : "Create Shift"
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Confirmation Dialog */}
        <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete Shift</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to delete the shift "{shiftToDelete?.name}"?
                This action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={handleConfirmDelete}
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              >
                Delete
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </DashboardLayout>
  )
}
