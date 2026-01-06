/**
 * @file app/incidences/types/page.tsx
 * @description Dynamic incidence type and category configuration page
 *
 * USER PERSPECTIVE:
 *   - Manage categories (parent groupings) with CRUD operations
 *   - Define incidence types under each category
 *   - Configure custom form fields that employees see when creating requests
 *   - Set is_requestable to control which types appear in employee portal
 *   - Configure effect type: positive (income), negative (deduction), or neutral
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: UI layout, form validation, field types
 *   CAUTION: form_fields JSON structure must match backend expectations
 *   DO NOT modify: API endpoints without coordinating with backend
 */

"use client"

import { useState, useEffect, useCallback } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Switch } from "@/components/ui/switch"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Plus,
  Pencil,
  Trash2,
  Tag,
  ArrowUp,
  ArrowDown,
  Minus,
  Paperclip,
  FolderOpen,
  FileText,
  GripVertical,
  X,
  Eye,
  EyeOff,
} from "lucide-react"
import {
  incidenceTypeApi,
  incidenceCategoryApi,
  IncidenceType,
  IncidenceCategory,
  CreateIncidenceTypeRequest,
  CreateIncidenceCategoryRequest,
  FormField,
  FormFieldsConfig,
} from "@/lib/api-client"

const EFFECT_TYPES = [
  { value: "positive", label: "Positive (Income)", icon: ArrowUp, color: "text-green-600" },
  { value: "negative", label: "Negative (Deduction)", icon: ArrowDown, color: "text-red-600" },
  { value: "neutral", label: "Neutral (No effect)", icon: Minus, color: "text-gray-600" },
]

const CALCULATION_METHODS = [
  { value: "daily_rate", label: "Daily Wage" },
  { value: "hourly_rate", label: "Hourly Wage" },
  { value: "fixed_amount", label: "Fixed Amount" },
  { value: "percentage", label: "Percentage" },
]

const FIELD_TYPES = [
  { value: "text", label: "Text" },
  { value: "textarea", label: "Long Text" },
  { value: "number", label: "Number" },
  { value: "date", label: "Date" },
  { value: "time", label: "Time" },
  { value: "boolean", label: "Yes/No" },
  { value: "select", label: "Dropdown List" },
  { value: "shift_select", label: "Shift Selector" },
]

const DEFAULT_CATEGORY_COLORS = [
  { value: "red", label: "Red", class: "bg-red-100 text-red-800" },
  { value: "orange", label: "Orange", class: "bg-orange-100 text-orange-800" },
  { value: "yellow", label: "Yellow", class: "bg-yellow-100 text-yellow-800" },
  { value: "green", label: "Green", class: "bg-green-100 text-green-800" },
  { value: "blue", label: "Blue", class: "bg-blue-100 text-blue-800" },
  { value: "purple", label: "Purple", class: "bg-purple-100 text-purple-800" },
  { value: "pink", label: "Pink", class: "bg-pink-100 text-pink-800" },
  { value: "gray", label: "Gray", class: "bg-gray-100 text-gray-800" },
]

export default function IncidenceTypesPage() {
  const [activeTab, setActiveTab] = useState("types")
  const [categories, setCategories] = useState<IncidenceCategory[]>([])
  const [incidenceTypes, setIncidenceTypes] = useState<IncidenceType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  // Category dialog state
  const [isCategoryDialogOpen, setIsCategoryDialogOpen] = useState(false)
  const [editingCategory, setEditingCategory] = useState<IncidenceCategory | null>(null)
  const [categoryFormData, setCategoryFormData] = useState<CreateIncidenceCategoryRequest>({
    name: "",
    code: "",
    description: "",
    color: "gray",
    is_requestable: false,
    is_active: true,
  })

  // Type dialog state
  const [isTypeDialogOpen, setIsTypeDialogOpen] = useState(false)
  const [editingType, setEditingType] = useState<IncidenceType | null>(null)
  const [typeFormData, setTypeFormData] = useState<CreateIncidenceTypeRequest>({
    name: "",
    category_id: "",
    category: "other",
    effect_type: "neutral",
    is_calculated: false,
    calculation_method: "",
    default_value: 0,
    requires_evidence: false,
    description: "",
    is_requestable: false,
    form_fields: { fields: [] },
  })

  // Form field builder state
  const [formFields, setFormFields] = useState<FormField[]>([])

  const [saving, setSaving] = useState(false)

  const loadData = useCallback(async () => {
    try {
      setLoading(true)
      const [categoriesData, typesData] = await Promise.all([
        incidenceCategoryApi.getAll(),
        incidenceTypeApi.getAll(),
      ])
      setCategories(categoriesData)
      setIncidenceTypes(typesData)
      setError("")
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Error loading data"
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  // Category handlers
  const handleOpenCategoryDialog = (category?: IncidenceCategory) => {
    if (category) {
      setEditingCategory(category)
      setCategoryFormData({
        name: category.name,
        code: category.code,
        description: category.description || "",
        color: category.color || "gray",
        is_requestable: category.is_requestable,
        is_active: category.is_active,
      })
    } else {
      setEditingCategory(null)
      setCategoryFormData({
        name: "",
        code: "",
        description: "",
        color: "gray",
        is_requestable: false,
        is_active: true,
      })
    }
    setIsCategoryDialogOpen(true)
  }

  const handleSaveCategory = async () => {
    try {
      setSaving(true)
      if (editingCategory) {
        await incidenceCategoryApi.update(editingCategory.id, categoryFormData)
      } else {
        await incidenceCategoryApi.create(categoryFormData)
      }
      await loadData()
      setIsCategoryDialogOpen(false)
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Error saving category"
      setError(errorMessage)
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteCategory = async (id: string, isSystem: boolean) => {
    if (isSystem) {
      setError("Cannot delete a system category")
      return
    }
    if (!confirm("Are you sure you want to delete this category? Associated types will be left without a category.")) return

    try {
      await incidenceCategoryApi.delete(id)
      await loadData()
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Error deleting category"
      setError(errorMessage)
    }
  }

  // Type handlers
  const handleOpenTypeDialog = (type?: IncidenceType) => {
    if (type) {
      setEditingType(type)
      setTypeFormData({
        name: type.name,
        category_id: type.category_id || "",
        category: type.category,
        effect_type: type.effect_type,
        is_calculated: type.is_calculated,
        calculation_method: type.calculation_method || "",
        default_value: type.default_value,
        requires_evidence: type.requires_evidence || false,
        description: type.description || "",
        is_requestable: type.is_requestable,
        form_fields: type.form_fields || { fields: [] },
      })
      setFormFields(type.form_fields?.fields || [])
    } else {
      setEditingType(null)
      setTypeFormData({
        name: "",
        category_id: "",
        category: "other",
        effect_type: "neutral",
        is_calculated: false,
        calculation_method: "",
        default_value: 0,
        requires_evidence: false,
        description: "",
        is_requestable: false,
        form_fields: { fields: [] },
      })
      setFormFields([])
    }
    setIsTypeDialogOpen(true)
  }

  const handleCategorySelectForType = (categoryId: string) => {
    const category = categories.find((c) => c.id === categoryId)
    setTypeFormData({
      ...typeFormData,
      category_id: categoryId,
      category: category?.code || "other",
    })
  }

  const handleSaveType = async () => {
    try {
      setSaving(true)
      const dataToSave: CreateIncidenceTypeRequest = {
        ...typeFormData,
        form_fields: { fields: formFields },
      }

      if (editingType) {
        await incidenceTypeApi.update(editingType.id, dataToSave)
      } else {
        await incidenceTypeApi.create(dataToSave)
      }
      await loadData()
      setIsTypeDialogOpen(false)
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Error saving incidence type"
      setError(errorMessage)
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteType = async (id: string) => {
    if (!confirm("Are you sure you want to delete this incidence type?")) return

    try {
      await incidenceTypeApi.delete(id)
      await loadData()
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Error deleting incidence type"
      setError(errorMessage)
    }
  }

  // Form field builder handlers
  const addFormField = () => {
    const newField: FormField = {
      name: `field_${formFields.length + 1}`,
      type: "text",
      label: "",
      required: false,
      display_order: formFields.length + 1,
    }
    setFormFields([...formFields, newField])
  }

  const updateFormField = (index: number, updates: Partial<FormField>) => {
    const updated = [...formFields]
    updated[index] = { ...updated[index], ...updates }
    setFormFields(updated)
  }

  const removeFormField = (index: number) => {
    setFormFields(formFields.filter((_, i) => i !== index))
  }

  const moveFormField = (index: number, direction: "up" | "down") => {
    const newIndex = direction === "up" ? index - 1 : index + 1
    if (newIndex < 0 || newIndex >= formFields.length) return

    const updated = [...formFields]
    const temp = updated[index]
    updated[index] = updated[newIndex]
    updated[newIndex] = temp

    // Update display_order
    updated.forEach((field, i) => {
      field.display_order = i + 1
    })

    setFormFields(updated)
  }

  const getCategoryBadge = (category: string, categoryId?: string) => {
    const cat = categories.find((c) => c.id === categoryId || c.code === category)
    if (!cat) return <span className="text-slate-400">{category}</span>

    const colorClass = DEFAULT_CATEGORY_COLORS.find((c) => c.value === cat.color)?.class || "bg-gray-100 text-gray-800"
    return (
      <span className={`px-2 py-1 rounded-full text-xs font-medium ${colorClass}`}>
        {cat.name}
      </span>
    )
  }

  const getEffectIcon = (effectType: string) => {
    const effect = EFFECT_TYPES.find((e) => e.value === effectType)
    if (!effect) return null
    const Icon = effect.icon
    return <Icon className={`h-4 w-4 ${effect.color}`} />
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white flex items-center gap-2">
              <Tag className="h-6 w-6" />
              Incidence Configuration
            </h1>
            <p className="text-slate-400 mt-1">
              Manage system categories and incidence types
            </p>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/50 rounded-lg p-4 text-red-400">
            {error}
            <button onClick={() => setError("")} className="ml-4 underline">
              Close
            </button>
          </div>
        )}

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="bg-slate-800 border-slate-700">
            <TabsTrigger value="types" className="data-[state=active]:bg-slate-700">
              <FileText className="h-4 w-4 mr-2" />
              Incidence Types
            </TabsTrigger>
            <TabsTrigger value="categories" className="data-[state=active]:bg-slate-700">
              <FolderOpen className="h-4 w-4 mr-2" />
              Categories
            </TabsTrigger>
          </TabsList>

          {/* Types Tab */}
          <TabsContent value="types" className="mt-4">
            <div className="flex justify-end mb-4">
              <Button onClick={() => handleOpenTypeDialog()} className="bg-blue-600 hover:bg-blue-700">
                <Plus className="h-4 w-4 mr-2" />
                Add Type
              </Button>
            </div>

            <div className="bg-slate-800/50 rounded-lg border border-slate-700">
              <Table>
                <TableHeader>
                  <TableRow className="border-slate-700 hover:bg-slate-800/50">
                    <TableHead className="text-slate-300">Name</TableHead>
                    <TableHead className="text-slate-300">Category</TableHead>
                    <TableHead className="text-slate-300">Effect</TableHead>
                    <TableHead className="text-slate-300">Request.</TableHead>
                    <TableHead className="text-slate-300">Fields</TableHead>
                    <TableHead className="text-slate-300">Evidence</TableHead>
                    <TableHead className="text-slate-300 text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loading ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center text-slate-400 py-8">
                        Loading...
                      </TableCell>
                    </TableRow>
                  ) : incidenceTypes.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center text-slate-400 py-8">
                        No incidence types registered
                      </TableCell>
                    </TableRow>
                  ) : (
                    incidenceTypes.map((type) => (
                      <TableRow key={type.id} className="border-slate-700 hover:bg-slate-800/30">
                        <TableCell className="text-white font-medium">
                          <div className="flex flex-col">
                            <span>{type.name}</span>
                            {type.description && (
                              <span className="text-xs text-slate-400">{type.description}</span>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>{getCategoryBadge(type.category, type.category_id)}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getEffectIcon(type.effect_type)}
                            <span className="text-slate-300 text-sm">
                              {EFFECT_TYPES.find((e) => e.value === type.effect_type)?.label}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          {type.is_requestable ? (
                            <span className="flex items-center gap-1 text-green-400">
                              <Eye className="h-4 w-4" />
                              Yes
                            </span>
                          ) : (
                            <span className="flex items-center gap-1 text-slate-500">
                              <EyeOff className="h-4 w-4" />
                              No
                            </span>
                          )}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {type.form_fields?.fields?.length || 0} fields
                        </TableCell>
                        <TableCell>
                          {type.requires_evidence ? (
                            <span className="flex items-center gap-1 text-orange-400">
                              <Paperclip className="h-4 w-4" />
                              Req.
                            </span>
                          ) : (
                            <span className="text-slate-500">Opt.</span>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleOpenTypeDialog(type)}
                              className="text-slate-400 hover:text-white"
                            >
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleDeleteType(type.id)}
                              className="text-red-400 hover:text-red-300"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          {/* Categories Tab */}
          <TabsContent value="categories" className="mt-4">
            <div className="flex justify-end mb-4">
              <Button onClick={() => handleOpenCategoryDialog()} className="bg-blue-600 hover:bg-blue-700">
                <Plus className="h-4 w-4 mr-2" />
                Add Category
              </Button>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {loading ? (
                <div className="col-span-full text-center text-slate-400 py-8">
                  Loading...
                </div>
              ) : categories.length === 0 ? (
                <div className="col-span-full text-center text-slate-400 py-8">
                  No categories registered
                </div>
              ) : (
                categories.map((category) => {
                  const colorClass =
                    DEFAULT_CATEGORY_COLORS.find((c) => c.value === category.color)?.class ||
                    "bg-gray-100 text-gray-800"
                  const typeCount = incidenceTypes.filter(
                    (t) => t.category_id === category.id || t.category === category.code
                  ).length

                  return (
                    <Card
                      key={category.id}
                      className="bg-slate-800/50 border-slate-700"
                    >
                      <CardHeader className="pb-2">
                        <div className="flex items-center justify-between">
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${colorClass}`}>
                            {category.name}
                          </span>
                          <div className="flex items-center gap-1">
                            {category.is_system && (
                              <span className="text-xs text-slate-500 mr-2">System</span>
                            )}
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleOpenCategoryDialog(category)}
                              className="text-slate-400 hover:text-white h-8 w-8 p-0"
                            >
                              <Pencil className="h-3 w-3" />
                            </Button>
                            {!category.is_system && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteCategory(category.id, category.is_system)}
                                className="text-red-400 hover:text-red-300 h-8 w-8 p-0"
                              >
                                <Trash2 className="h-3 w-3" />
                              </Button>
                            )}
                          </div>
                        </div>
                        <CardTitle className="text-white text-lg">{category.code}</CardTitle>
                        <CardDescription className="text-slate-400">
                          {category.description || "No description"}
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-slate-400">{typeCount} types</span>
                          <div className="flex items-center gap-4">
                            <span
                              className={`${
                                category.is_requestable ? "text-green-400" : "text-slate-500"
                              }`}
                            >
                              {category.is_requestable ? "Requestable" : "Not requestable"}
                            </span>
                            <span
                              className={`${category.is_active ? "text-green-400" : "text-red-400"}`}
                            >
                              {category.is_active ? "Active" : "Inactive"}
                            </span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  )
                })
              )}
            </div>
          </TabsContent>
        </Tabs>

        {/* Category Dialog */}
        <Dialog open={isCategoryDialogOpen} onOpenChange={setIsCategoryDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>
                {editingCategory ? "Edit Category" : "New Category"}
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                {editingCategory
                  ? "Modify the category details"
                  : "Create a new category to group incidence types"}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="cat-name">Name</Label>
                  <Input
                    id="cat-name"
                    value={categoryFormData.name}
                    onChange={(e) =>
                      setCategoryFormData({ ...categoryFormData, name: e.target.value })
                    }
                    placeholder="Ex: Vacation"
                    className="bg-slate-800 border-slate-600"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="cat-code">Code</Label>
                  <Input
                    id="cat-code"
                    value={categoryFormData.code}
                    onChange={(e) =>
                      setCategoryFormData({
                        ...categoryFormData,
                        code: e.target.value.toLowerCase().replace(/\s+/g, "_"),
                      })
                    }
                    placeholder="Ex: vacation"
                    className="bg-slate-800 border-slate-600"
                    disabled={editingCategory?.is_system}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="cat-description">Description</Label>
                <Textarea
                  id="cat-description"
                  value={categoryFormData.description}
                  onChange={(e) =>
                    setCategoryFormData({ ...categoryFormData, description: e.target.value })
                  }
                  placeholder="Category description"
                  className="bg-slate-800 border-slate-600"
                />
              </div>

              <div className="space-y-2">
                <Label>Color</Label>
                <div className="flex flex-wrap gap-2">
                  {DEFAULT_CATEGORY_COLORS.map((color) => (
                    <button
                      key={color.value}
                      type="button"
                      onClick={() =>
                        setCategoryFormData({ ...categoryFormData, color: color.value })
                      }
                      className={`px-3 py-1 rounded-full text-xs font-medium ${color.class} ${
                        categoryFormData.color === color.value
                          ? "ring-2 ring-white ring-offset-2 ring-offset-slate-900"
                          : ""
                      }`}
                    >
                      {color.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Switch
                    id="cat-requestable"
                    checked={categoryFormData.is_requestable}
                    onCheckedChange={(checked) =>
                      setCategoryFormData({ ...categoryFormData, is_requestable: checked })
                    }
                  />
                  <Label htmlFor="cat-requestable">Requestable by employees</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    id="cat-active"
                    checked={categoryFormData.is_active}
                    onCheckedChange={(checked) =>
                      setCategoryFormData({ ...categoryFormData, is_active: checked })
                    }
                  />
                  <Label htmlFor="cat-active">Active</Label>
                </div>
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsCategoryDialogOpen(false)}
                className="border-slate-600"
              >
                Cancel
              </Button>
              <Button
                onClick={handleSaveCategory}
                disabled={saving || !categoryFormData.name || !categoryFormData.code}
                className="bg-blue-600 hover:bg-blue-700"
              >
                {saving ? "Saving..." : editingCategory ? "Update" : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Type Dialog */}
        <Dialog open={isTypeDialogOpen} onOpenChange={setIsTypeDialogOpen}>
          <DialogContent className="bg-slate-900 border-slate-700 text-white sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>
                {editingType ? "Edit Incidence Type" : "New Incidence Type"}
              </DialogTitle>
              <DialogDescription className="text-slate-400">
                {editingType
                  ? "Modify details and custom fields"
                  : "Create a new incidence type with custom fields"}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6 py-4">
              {/* Basic Info */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium text-slate-300">Basic Information</h3>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="type-name">Name</Label>
                    <Input
                      id="type-name"
                      value={typeFormData.name}
                      onChange={(e) => setTypeFormData({ ...typeFormData, name: e.target.value })}
                      placeholder="Ex: Vacation"
                      className="bg-slate-800 border-slate-600"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Category</Label>
                    <Select
                      value={typeFormData.category_id}
                      onValueChange={handleCategorySelectForType}
                    >
                      <SelectTrigger className="bg-slate-800 border-slate-600">
                        <SelectValue placeholder="Select category" />
                      </SelectTrigger>
                      <SelectContent className="bg-slate-800 border-slate-600">
                        {categories.map((cat) => (
                          <SelectItem key={cat.id} value={cat.id}>
                            {cat.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="type-description">Description</Label>
                  <Textarea
                    id="type-description"
                    value={typeFormData.description}
                    onChange={(e) =>
                      setTypeFormData({ ...typeFormData, description: e.target.value })
                    }
                    placeholder="Incidence type description"
                    className="bg-slate-800 border-slate-600"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Effect Type</Label>
                    <Select
                      value={typeFormData.effect_type}
                      onValueChange={(value) =>
                        setTypeFormData({ ...typeFormData, effect_type: value })
                      }
                    >
                      <SelectTrigger className="bg-slate-800 border-slate-600">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent className="bg-slate-800 border-slate-600">
                        {EFFECT_TYPES.map((effect) => (
                          <SelectItem key={effect.value} value={effect.value}>
                            {effect.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="flex flex-wrap gap-6">
                  <div className="flex items-center gap-2">
                    <Switch
                      id="type-requestable"
                      checked={typeFormData.is_requestable}
                      onCheckedChange={(checked) =>
                        setTypeFormData({ ...typeFormData, is_requestable: checked })
                      }
                    />
                    <Label htmlFor="type-requestable">Requestable by employees</Label>
                  </div>
                  <div className="flex items-center gap-2">
                    <Switch
                      id="type-evidence"
                      checked={typeFormData.requires_evidence}
                      onCheckedChange={(checked) =>
                        setTypeFormData({ ...typeFormData, requires_evidence: checked })
                      }
                    />
                    <Label htmlFor="type-evidence" className="flex items-center gap-1">
                      <Paperclip className="h-4 w-4" />
                      Requires evidence
                    </Label>
                  </div>
                  <div className="flex items-center gap-2">
                    <Switch
                      id="type-calculated"
                      checked={typeFormData.is_calculated}
                      onCheckedChange={(checked) =>
                        setTypeFormData({ ...typeFormData, is_calculated: checked })
                      }
                    />
                    <Label htmlFor="type-calculated">Calculate automatically</Label>
                  </div>
                </div>

                {typeFormData.is_calculated && (
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Calculation Method</Label>
                      <Select
                        value={typeFormData.calculation_method}
                        onValueChange={(value) =>
                          setTypeFormData({ ...typeFormData, calculation_method: value })
                        }
                      >
                        <SelectTrigger className="bg-slate-800 border-slate-600">
                          <SelectValue placeholder="Select..." />
                        </SelectTrigger>
                        <SelectContent className="bg-slate-800 border-slate-600">
                          {CALCULATION_METHODS.map((method) => (
                            <SelectItem key={method.value} value={method.value}>
                              {method.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="type-default-value">Default Value</Label>
                      <Input
                        id="type-default-value"
                        type="number"
                        value={typeFormData.default_value}
                        onChange={(e) =>
                          setTypeFormData({
                            ...typeFormData,
                            default_value: parseFloat(e.target.value) || 0,
                          })
                        }
                        className="bg-slate-800 border-slate-600"
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* Form Field Builder */}
              <div className="space-y-4 border-t border-slate-700 pt-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-medium text-slate-300">
                    Custom Form Fields
                  </h3>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={addFormField}
                    className="border-slate-600"
                  >
                    <Plus className="h-4 w-4 mr-1" />
                    Add Field
                  </Button>
                </div>

                {formFields.length === 0 ? (
                  <div className="text-center py-8 text-slate-400 border border-dashed border-slate-600 rounded-lg">
                    No custom fields.
                    <br />
                    Basic fields (start date, end date, reason) are included automatically.
                  </div>
                ) : (
                  <div className="space-y-3">
                    {formFields.map((field, index) => (
                      <Card key={index} className="bg-slate-800/50 border-slate-700">
                        <CardContent className="p-3">
                          <div className="flex items-start gap-2">
                            <div className="flex flex-col gap-1 pt-2">
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => moveFormField(index, "up")}
                                disabled={index === 0}
                                className="h-6 w-6 p-0"
                              >
                                <ArrowUp className="h-3 w-3" />
                              </Button>
                              <GripVertical className="h-4 w-4 text-slate-500" />
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => moveFormField(index, "down")}
                                disabled={index === formFields.length - 1}
                                className="h-6 w-6 p-0"
                              >
                                <ArrowDown className="h-3 w-3" />
                              </Button>
                            </div>

                            <div className="flex-1 grid grid-cols-2 md:grid-cols-4 gap-2">
                              <div>
                                <Label className="text-xs">Field Name</Label>
                                <Input
                                  value={field.name}
                                  onChange={(e) =>
                                    updateFormField(index, {
                                      name: e.target.value.toLowerCase().replace(/\s+/g, "_"),
                                    })
                                  }
                                  placeholder="field_1"
                                  className="bg-slate-700 border-slate-600 h-8 text-sm"
                                />
                              </div>
                              <div>
                                <Label className="text-xs">Label</Label>
                                <Input
                                  value={field.label}
                                  onChange={(e) => updateFormField(index, { label: e.target.value })}
                                  placeholder="Visible Name"
                                  className="bg-slate-700 border-slate-600 h-8 text-sm"
                                />
                              </div>
                              <div>
                                <Label className="text-xs">Type</Label>
                                <Select
                                  value={field.type}
                                  onValueChange={(value) =>
                                    updateFormField(index, {
                                      type: value as FormField["type"],
                                    })
                                  }
                                >
                                  <SelectTrigger className="bg-slate-700 border-slate-600 h-8 text-sm">
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent className="bg-slate-700 border-slate-600">
                                    {FIELD_TYPES.map((ft) => (
                                      <SelectItem key={ft.value} value={ft.value}>
                                        {ft.label}
                                      </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                              </div>
                              <div className="flex items-end gap-2">
                                <div className="flex items-center gap-1">
                                  <Switch
                                    id={`field-required-${index}`}
                                    checked={field.required}
                                    onCheckedChange={(checked) =>
                                      updateFormField(index, { required: checked })
                                    }
                                  />
                                  <Label htmlFor={`field-required-${index}`} className="text-xs">
                                    Req.
                                  </Label>
                                </div>
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => removeFormField(index)}
                                  className="h-8 w-8 p-0 text-red-400 hover:text-red-300"
                                >
                                  <X className="h-4 w-4" />
                                </Button>
                              </div>
                            </div>
                          </div>

                          {/* Additional options for number fields */}
                          {field.type === "number" && (
                            <div className="mt-2 ml-10 grid grid-cols-3 gap-2">
                              <div>
                                <Label className="text-xs">Min</Label>
                                <Input
                                  type="number"
                                  value={field.min ?? ""}
                                  onChange={(e) =>
                                    updateFormField(index, {
                                      min: e.target.value ? parseFloat(e.target.value) : undefined,
                                    })
                                  }
                                  className="bg-slate-700 border-slate-600 h-8 text-sm"
                                />
                              </div>
                              <div>
                                <Label className="text-xs">Max</Label>
                                <Input
                                  type="number"
                                  value={field.max ?? ""}
                                  onChange={(e) =>
                                    updateFormField(index, {
                                      max: e.target.value ? parseFloat(e.target.value) : undefined,
                                    })
                                  }
                                  className="bg-slate-700 border-slate-600 h-8 text-sm"
                                />
                              </div>
                              <div>
                                <Label className="text-xs">Step</Label>
                                <Input
                                  type="number"
                                  value={field.step ?? ""}
                                  onChange={(e) =>
                                    updateFormField(index, {
                                      step: e.target.value ? parseFloat(e.target.value) : undefined,
                                    })
                                  }
                                  className="bg-slate-700 border-slate-600 h-8 text-sm"
                                />
                              </div>
                            </div>
                          )}

                          {/* Options for select fields */}
                          {(field.type === "select" || field.type === "multiselect") && (
                            <div className="mt-2 ml-10">
                              <Label className="text-xs">
                                Options (one per line: value|label)
                              </Label>
                              <Textarea
                                value={
                                  field.options?.map((o) => `${o.value}|${o.label}`).join("\n") || ""
                                }
                                onChange={(e) => {
                                  const lines = e.target.value.split("\n")
                                  const options = lines
                                    .filter((line) => line.trim())
                                    .map((line) => {
                                      const [value, label] = line.split("|")
                                      return { value: value?.trim() || "", label: label?.trim() || value?.trim() || "" }
                                    })
                                  updateFormField(index, { options })
                                }}
                                placeholder="option1|Option 1&#10;option2|Option 2"
                                className="bg-slate-700 border-slate-600 text-sm h-20"
                              />
                            </div>
                          )}

                          {/* Help text */}
                          <div className="mt-2 ml-10">
                            <Label className="text-xs">Help Text (optional)</Label>
                            <Input
                              value={field.help_text || ""}
                              onChange={(e) => updateFormField(index, { help_text: e.target.value })}
                              placeholder="Ex: Enter the number of hours"
                              className="bg-slate-700 border-slate-600 h-8 text-sm"
                            />
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsTypeDialogOpen(false)}
                className="border-slate-600"
              >
                Cancel
              </Button>
              <Button
                onClick={handleSaveType}
                disabled={saving || !typeFormData.name}
                className="bg-blue-600 hover:bg-blue-700"
              >
                {saving ? "Saving..." : editingType ? "Update" : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </DashboardLayout>
  )
}
