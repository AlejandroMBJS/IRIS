"use client"

import { useEffect, useState, Suspense, useCallback, useMemo } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  Calendar,
  FileText,
  ArrowLeft,
  ArrowRight,
  Send,
  AlertCircle,
  Clock,
  Loader2,
  CheckCircle2,
  Palmtree,
  Stethoscope,
  UserCheck,
  UserX,
  LogIn,
  LogOut,
  RefreshCw,
  Timer,
  HelpCircle,
  Sparkles,
  FileCheck,
  ChevronRight,
  Grid3X3,
  Briefcase,
  HeartPulse,
  Plane,
  Coffee,
  Zap,
  Star,
  Folder,
  Upload,
  Trash2,
  Paperclip,
  X,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { isAuthenticated, getCurrentUser } from "@/lib/auth"
import { PortalLayout } from "@/components/layout/portal-layout"
import { DynamicFormFields } from "@/components/dynamic-form-fields"
import {
  absenceRequestApi,
  incidenceTypeApi,
  evidenceApi,
  IncidenceType,
  RequestType,
} from "@/lib/api-client"
import { useToast } from "@/hooks/use-toast"

// Icon mapping for request types
const TYPE_ICONS: Record<string, typeof Calendar> = {
  vacation: Palmtree,
  vacaciones: Palmtree,
  sick: Stethoscope,
  incapacidad: Stethoscope,
  enfermedad: Stethoscope,
  "permiso con goce": UserCheck,
  "con goce": UserCheck,
  "permiso sin goce": UserX,
  "sin goce": UserX,
  entrada: LogIn,
  retardo: LogIn,
  salida: LogOut,
  turno: RefreshCw,
  "tiempo por tiempo": Timer,
  personal: HelpCircle,
}

// Color and icon mapping for categories
const CATEGORY_CONFIG: Record<string, {
  bg: string;
  border: string;
  text: string;
  glow: string;
  icon: typeof Calendar;
  gradient: string;
}> = {
  vacation: {
    bg: "from-emerald-500/20 to-emerald-600/10",
    border: "border-emerald-500/50",
    text: "text-emerald-400",
    glow: "shadow-emerald-500/20",
    icon: Palmtree,
    gradient: "from-emerald-500 to-teal-600"
  },
  sick: {
    bg: "from-orange-500/20 to-orange-600/10",
    border: "border-orange-500/50",
    text: "text-orange-400",
    glow: "shadow-orange-500/20",
    icon: HeartPulse,
    gradient: "from-orange-500 to-red-600"
  },
  absence: {
    bg: "from-blue-500/20 to-blue-600/10",
    border: "border-blue-500/50",
    text: "text-blue-400",
    glow: "shadow-blue-500/20",
    icon: Briefcase,
    gradient: "from-blue-500 to-indigo-600"
  },
  delay: {
    bg: "from-amber-500/20 to-amber-600/10",
    border: "border-amber-500/50",
    text: "text-amber-400",
    glow: "shadow-amber-500/20",
    icon: Clock,
    gradient: "from-amber-500 to-orange-600"
  },
  deduction: {
    bg: "from-red-500/20 to-red-600/10",
    border: "border-red-500/50",
    text: "text-red-400",
    glow: "shadow-red-500/20",
    icon: Zap,
    gradient: "from-red-500 to-rose-600"
  },
  bonus: {
    bg: "from-green-500/20 to-green-600/10",
    border: "border-green-500/50",
    text: "text-green-400",
    glow: "shadow-green-500/20",
    icon: Star,
    gradient: "from-green-500 to-emerald-600"
  },
  overtime: {
    bg: "from-cyan-500/20 to-cyan-600/10",
    border: "border-cyan-500/50",
    text: "text-cyan-400",
    glow: "shadow-cyan-500/20",
    icon: Timer,
    gradient: "from-cyan-500 to-blue-600"
  },
  other: {
    bg: "from-purple-500/20 to-purple-600/10",
    border: "border-purple-500/50",
    text: "text-purple-400",
    glow: "shadow-purple-500/20",
    icon: Folder,
    gradient: "from-purple-500 to-violet-600"
  },
}

// Backwards compatibility
const CATEGORY_COLORS = CATEGORY_CONFIG

function getTypeIcon(type: IncidenceType) {
  const nameLower = type.name.toLowerCase()
  for (const [key, Icon] of Object.entries(TYPE_ICONS)) {
    if (nameLower.includes(key)) return Icon
  }
  return FileText
}

function getCategoryColors(category: string) {
  return CATEGORY_COLORS[category] || CATEGORY_COLORS.other
}

function getCategoryConfig(category: string) {
  return CATEGORY_CONFIG[category] || CATEGORY_CONFIG.other
}

function mapCategoryToRequestType(incidenceType: IncidenceType): RequestType {
  const categoryMap: Record<string, RequestType> = {
    vacation: "VACATION",
    sick: "SICK_LEAVE",
    absence: "PAID_LEAVE",
    delay: "LATE_ENTRY",
    other: "OTHER",
  }
  const nameLower = incidenceType.name.toLowerCase()
  if (nameLower.includes("vacacion")) return "VACATION"
  if (nameLower.includes("incapacidad")) return "SICK_LEAVE"
  if (nameLower.includes("sin goce")) return "UNPAID_LEAVE"
  if (nameLower.includes("con goce")) return "PAID_LEAVE"
  if (nameLower.includes("entrada")) return "LATE_ENTRY"
  if (nameLower.includes("salida")) return "EARLY_EXIT"
  if (nameLower.includes("turno")) return "SHIFT_CHANGE"
  if (nameLower.includes("tiempo por tiempo")) return "TIME_FOR_TIME"
  if (nameLower.includes("personal")) return "PERSONAL"
  return categoryMap[incidenceType.category] || "OTHER"
}

// Step indicator component
function StepIndicator({ currentStep, totalSteps }: { currentStep: number; totalSteps: number }) {
  return (
    <div className="flex items-center justify-center gap-2 mb-8">
      {Array.from({ length: totalSteps }).map((_, i) => (
        <div key={i} className="flex items-center">
          <div
            className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold text-sm transition-all duration-300 ${
              i < currentStep
                ? "bg-emerald-500 text-white shadow-lg shadow-emerald-500/30"
                : i === currentStep
                ? "bg-blue-500 text-white shadow-lg shadow-blue-500/30 scale-110"
                : "bg-slate-700 text-slate-400"
            }`}
          >
            {i < currentStep ? <CheckCircle2 size={20} /> : i + 1}
          </div>
          {i < totalSteps - 1 && (
            <div
              className={`w-12 h-1 mx-1 rounded transition-all duration-300 ${
                i < currentStep ? "bg-emerald-500" : "bg-slate-700"
              }`}
            />
          )}
        </div>
      ))}
    </div>
  )
}

function NewRequestForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { toast } = useToast()
  const [user, setUser] = useState<{ id: string } | null>(null)
  const [loading, setLoading] = useState(false)
  const [loadingTypes, setLoadingTypes] = useState(true)
  const [requestableTypes, setRequestableTypes] = useState<IncidenceType[]>([])

  // Multi-step form state
  const [currentStep, setCurrentStep] = useState(0)
  const [selectedType, setSelectedType] = useState<IncidenceType | null>(null)
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [startDate, setStartDate] = useState("")
  const [endDate, setEndDate] = useState("")
  const [reason, setReason] = useState("")
  const [customFieldValues, setCustomFieldValues] = useState<Record<string, unknown>>({})
  const [evidenceFiles, setEvidenceFiles] = useState<File[]>([])
  const [uploadingEvidence, setUploadingEvidence] = useState(false)

  const hasCustomFields = selectedType?.form_fields?.fields && selectedType.form_fields.fields.length > 0
  const totalSteps = hasCustomFields ? 4 : 3

  const calculateTotalDays = () => {
    if (!startDate || !endDate) return 0
    const start = new Date(startDate)
    const end = new Date(endDate)
    const diffTime = end.getTime() - start.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1
    return diffDays > 0 ? diffDays : 0
  }

  const totalDays = calculateTotalDays()

  const loadRequestableTypes = useCallback(async () => {
    try {
      setLoadingTypes(true)
      const types = await incidenceTypeApi.getRequestable()
      setRequestableTypes(types)
    } catch (error) {
      console.error("Failed to load requestable types:", error)
      toast({
        title: "Error",
        description: "No se pudieron cargar los tipos de solicitud",
        variant: "destructive",
      })
    } finally {
      setLoadingTypes(false)
    }
  }, [toast])

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }
    setUser(getCurrentUser())
    loadRequestableTypes()
  }, [router, loadRequestableTypes])

  useEffect(() => {
    const typeId = searchParams.get("type_id")
    if (typeId && requestableTypes.length > 0 && !selectedType) {
      const foundType = requestableTypes.find((t) => t.id === typeId)
      if (foundType) {
        setSelectedType(foundType)
        setCurrentStep(1)
      }
    }
  }, [searchParams, requestableTypes, selectedType])

  useEffect(() => {
    setCustomFieldValues({})
  }, [selectedType?.id])

  const handleTypeSelect = (type: IncidenceType) => {
    setSelectedType(type)
    setTimeout(() => setCurrentStep(1), 150)
  }

  const handleNext = () => {
    if (currentStep === 1 && (!startDate || !endDate || totalDays <= 0)) {
      toast({
        title: "Fechas requeridas",
        description: "Por favor selecciona las fechas de inicio y fin",
        variant: "destructive",
      })
      return
    }
    if (currentStep < totalSteps - 1) {
      setCurrentStep(currentStep + 1)
    }
  }

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleSubmit = async () => {
    if (!selectedType || !startDate || !endDate || !reason) {
      toast({
        title: "Error",
        description: "Por favor completa todos los campos requeridos",
        variant: "destructive",
      })
      return
    }

    // Validate evidence files are provided when required
    if (selectedType.requires_evidence && evidenceFiles.length === 0) {
      toast({
        title: "Error",
        description: "Esta solicitud requiere evidencia. Por favor adjunta al menos un archivo.",
        variant: "destructive",
      })
      return
    }

    const requiredFields = selectedType.form_fields?.fields?.filter((f) => f.required) || []
    for (const field of requiredFields) {
      const value = customFieldValues[field.name]
      if (value === undefined || value === null || value === "") {
        toast({
          title: "Error",
          description: `El campo "${field.label}" es requerido`,
          variant: "destructive",
        })
        return
      }
    }

    setLoading(true)

    try {
      const data = {
        employee_id: user?.id || "",
        request_type: mapCategoryToRequestType(selectedType),
        incidence_type_id: selectedType.id,
        start_date: startDate,
        end_date: endDate,
        total_days: totalDays,
        reason: reason,
        custom_fields: Object.keys(customFieldValues).length > 0 ? customFieldValues : undefined,
      }

      const result = await absenceRequestApi.create(data)

      // Upload evidence files if any were selected
      if (evidenceFiles.length > 0 && result.requestId) {
        setUploadingEvidence(true)
        let uploadErrors = 0

        for (const file of evidenceFiles) {
          try {
            await evidenceApi.upload(result.requestId, file)
          } catch (uploadError) {
            console.error("Failed to upload evidence:", uploadError)
            uploadErrors++
          }
        }

        setUploadingEvidence(false)

        if (uploadErrors > 0) {
          toast({
            title: "Solicitud Enviada",
            description: `Tu solicitud fue enviada, pero ${uploadErrors} archivo(s) no se pudieron subir.`,
            variant: "default",
          })
        } else {
          toast({
            title: "Solicitud Enviada",
            description: "Tu solicitud y evidencia han sido enviadas para aprobacion",
          })
        }
      } else {
        toast({
          title: "Solicitud Enviada",
          description: "Tu solicitud ha sido enviada para aprobacion",
        })
      }

      router.push("/requests")
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "No se pudo enviar la solicitud"
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      })
    } finally {
      setLoading(false)
      setUploadingEvidence(false)
    }
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return ""
    const date = new Date(dateStr + "T00:00:00")
    return date.toLocaleDateString("es-MX", { weekday: "long", year: "numeric", month: "long", day: "numeric" })
  }

  // Group types by category
  const groupedTypes = requestableTypes.reduce((groups, type) => {
    const categoryName = type.incidence_category?.name || type.category || "Otros"
    if (!groups[categoryName]) {
      groups[categoryName] = { types: [], category: type.category }
    }
    groups[categoryName].types.push(type)
    return groups
  }, {} as Record<string, { types: IncidenceType[]; category: string }>)

  return (
    <PortalLayout>
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-6">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => (currentStep > 0 ? handleBack() : router.back())}
            className="text-slate-400 hover:text-white hover:bg-slate-800 rounded-full"
          >
            <ArrowLeft size={20} />
          </Button>
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-white">Nueva Solicitud</h1>
            <p className="text-slate-400 text-sm">
              {currentStep === 0 && "Selecciona el tipo de solicitud"}
              {currentStep === 1 && "Define el periodo de tu solicitud"}
              {currentStep === 2 && hasCustomFields && "Completa los detalles adicionales"}
              {((currentStep === 2 && !hasCustomFields) || currentStep === 3) && "Revisa y envia tu solicitud"}
            </p>
          </div>
          {selectedType && (
            <div className={`px-4 py-2 rounded-full bg-gradient-to-r ${getCategoryColors(selectedType.category).bg} ${getCategoryColors(selectedType.category).border} border`}>
              <span className={`text-sm font-medium ${getCategoryColors(selectedType.category).text}`}>
                {selectedType.name}
              </span>
            </div>
          )}
        </div>

        {/* Step Indicator */}
        <StepIndicator currentStep={currentStep} totalSteps={totalSteps} />

        {/* Step Content */}
        <div className="min-h-[400px]">
          {/* Step 0: Type Selection - Horizontal Grid Layout */}
          {currentStep === 0 && (
            <div className="animate-in fade-in slide-in-from-right-4 duration-300">
              {loadingTypes ? (
                <div className="flex flex-col items-center justify-center py-20">
                  <div className="relative">
                    <div className="w-16 h-16 rounded-full border-4 border-slate-700 border-t-blue-500 animate-spin" />
                    <Sparkles className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-blue-400" size={24} />
                  </div>
                  <p className="text-slate-400 mt-4">Cargando tipos de solicitud...</p>
                </div>
              ) : requestableTypes.length === 0 ? (
                <div className="text-center py-20">
                  <div className="w-20 h-20 mx-auto mb-4 rounded-full bg-slate-800 flex items-center justify-center">
                    <FileText size={40} className="text-slate-600" />
                  </div>
                  <h3 className="text-xl font-semibold text-white mb-2">Sin tipos disponibles</h3>
                  <p className="text-slate-400 max-w-md mx-auto">
                    No hay tipos de solicitud configurados. Contacta al administrador.
                  </p>
                </div>
              ) : (
                <div className="space-y-6">
                  {/* Category Tabs - Horizontal Scrollable */}
                  <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
                    <button
                      type="button"
                      onClick={() => setSelectedCategory(null)}
                      className={`flex items-center gap-2 px-4 py-2.5 rounded-xl font-medium text-sm whitespace-nowrap transition-all duration-200 ${
                        selectedCategory === null
                          ? "bg-gradient-to-r from-blue-600 to-indigo-600 text-white shadow-lg shadow-blue-500/25"
                          : "bg-slate-800/80 text-slate-400 hover:bg-slate-700 hover:text-slate-200"
                      }`}
                    >
                      <Grid3X3 size={18} />
                      Todos
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        selectedCategory === null ? "bg-white/20" : "bg-slate-700"
                      }`}>
                        {requestableTypes.length}
                      </span>
                    </button>
                    {Object.entries(groupedTypes).map(([categoryName, { types, category }]) => {
                      const config = getCategoryConfig(category)
                      const CategoryIcon = config.icon
                      const isActive = selectedCategory === categoryName
                      return (
                        <button
                          key={categoryName}
                          type="button"
                          onClick={() => setSelectedCategory(categoryName)}
                          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl font-medium text-sm whitespace-nowrap transition-all duration-200 ${
                            isActive
                              ? `bg-gradient-to-r ${config.gradient} text-white shadow-lg ${config.glow}`
                              : "bg-slate-800/80 text-slate-400 hover:bg-slate-700 hover:text-slate-200"
                          }`}
                        >
                          <CategoryIcon size={18} />
                          {categoryName}
                          <span className={`px-2 py-0.5 rounded-full text-xs ${
                            isActive ? "bg-white/20" : "bg-slate-700"
                          }`}>
                            {types.length}
                          </span>
                        </button>
                      )
                    })}
                  </div>

                  {/* Types Grid - Horizontal Cards */}
                  <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
                    {(selectedCategory
                      ? groupedTypes[selectedCategory]?.types || []
                      : requestableTypes
                    ).map((type) => {
                      const Icon = getTypeIcon(type)
                      const config = getCategoryConfig(type.category)
                      const isSelected = selectedType?.id === type.id
                      return (
                        <button
                          key={type.id}
                          type="button"
                          onClick={() => handleTypeSelect(type)}
                          className={`group relative p-4 rounded-2xl border-2 transition-all duration-300 text-center overflow-hidden ${
                            isSelected
                              ? `bg-gradient-to-br ${config.bg} ${config.border} shadow-lg ${config.glow} scale-[1.02]`
                              : "bg-slate-800/50 border-slate-700/50 hover:border-slate-600 hover:bg-slate-800 hover:scale-[1.02]"
                          }`}
                        >
                          {/* Gradient overlay on hover */}
                          <div className={`absolute inset-0 bg-gradient-to-br ${config.gradient} opacity-0 group-hover:opacity-5 transition-opacity`} />

                          {/* Icon */}
                          <div className={`mx-auto w-12 h-12 rounded-xl flex items-center justify-center mb-3 transition-all ${
                            isSelected
                              ? `bg-gradient-to-br ${config.gradient} text-white shadow-lg`
                              : "bg-slate-700/50 text-slate-400 group-hover:bg-slate-700 group-hover:text-slate-300"
                          }`}>
                            <Icon size={24} />
                          </div>

                          {/* Name */}
                          <h4 className={`font-semibold text-sm mb-1 transition-colors ${
                            isSelected ? "text-white" : "text-slate-200 group-hover:text-white"
                          }`}>
                            {type.name}
                          </h4>

                          {/* Category badge */}
                          <span className={`inline-block px-2 py-0.5 rounded-full text-xs ${
                            isSelected
                              ? "bg-white/20 text-white"
                              : `${config.bg} ${config.text}`
                          }`}>
                            {type.incidence_category?.name || type.category}
                          </span>

                          {/* Requires evidence indicator */}
                          {type.requires_evidence && (
                            <div className="absolute top-2 right-2">
                              <div className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" title="Requiere evidencia" />
                            </div>
                          )}

                          {/* Selection indicator */}
                          {isSelected && (
                            <div className="absolute top-2 left-2">
                              <CheckCircle2 size={18} className={config.text} />
                            </div>
                          )}
                        </button>
                      )
                    })}
                  </div>

                  {/* Selected type preview */}
                  {selectedType && (
                    <div className={`mt-4 p-4 rounded-2xl border ${getCategoryColors(selectedType.category).border} bg-gradient-to-r ${getCategoryColors(selectedType.category).bg}`}>
                      <div className="flex items-center gap-4">
                        <div className={`p-3 rounded-xl bg-gradient-to-br ${getCategoryConfig(selectedType.category).gradient}`}>
                          {(() => {
                            const Icon = getTypeIcon(selectedType)
                            return <Icon size={24} className="text-white" />
                          })()}
                        </div>
                        <div className="flex-1">
                          <h4 className="font-semibold text-white">{selectedType.name}</h4>
                          <p className="text-sm text-slate-400">{selectedType.description || "Haz clic en Siguiente para continuar"}</p>
                        </div>
                        <ChevronRight className={`${getCategoryColors(selectedType.category).text}`} size={24} />
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}

          {/* Step 1: Date Selection */}
          {currentStep === 1 && (
            <div className="animate-in fade-in slide-in-from-right-4 duration-300">
              <div className="bg-slate-800/50 rounded-3xl border border-slate-700/50 p-8">
                <div className="flex items-center gap-3 mb-6">
                  <div className="p-3 rounded-xl bg-emerald-500/20">
                    <Calendar size={24} className="text-emerald-400" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-white">Periodo de la solicitud</h3>
                    <p className="text-sm text-slate-400">Selecciona las fechas de inicio y fin</p>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                  <div className="space-y-2">
                    <Label htmlFor="startDate" className="text-slate-300 flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-emerald-400" />
                      Fecha de Inicio
                    </Label>
                    <Input
                      id="startDate"
                      type="date"
                      value={startDate}
                      onChange={(e) => setStartDate(e.target.value)}
                      className="bg-slate-900/50 border-slate-600 text-white h-12 text-lg rounded-xl focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="endDate" className="text-slate-300 flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-blue-400" />
                      Fecha de Fin
                    </Label>
                    <Input
                      id="endDate"
                      type="date"
                      value={endDate}
                      onChange={(e) => setEndDate(e.target.value)}
                      min={startDate}
                      className="bg-slate-900/50 border-slate-600 text-white h-12 text-lg rounded-xl focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500"
                    />
                  </div>
                </div>

                {totalDays > 0 && (
                  <div className="bg-gradient-to-r from-blue-500/10 to-emerald-500/10 border border-blue-500/30 rounded-2xl p-6">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Clock size={24} className="text-blue-400" />
                        <div>
                          <p className="text-sm text-slate-400">Duracion total</p>
                          <p className="text-2xl font-bold text-white">
                            {totalDays} dia{totalDays !== 1 ? "s" : ""}
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <p className="text-xs text-slate-500 uppercase tracking-wider">Periodo</p>
                        <p className="text-sm text-slate-300">
                          {startDate && new Date(startDate + "T00:00:00").toLocaleDateString("es-MX", { month: "short", day: "numeric" })}
                          {" - "}
                          {endDate && new Date(endDate + "T00:00:00").toLocaleDateString("es-MX", { month: "short", day: "numeric" })}
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Step 2: Custom Fields (if any) */}
          {currentStep === 2 && hasCustomFields && (
            <div className="animate-in fade-in slide-in-from-right-4 duration-300">
              <div className="bg-slate-800/50 rounded-3xl border border-slate-700/50 p-8">
                <div className="flex items-center gap-3 mb-6">
                  <div className="p-3 rounded-xl bg-purple-500/20">
                    <FileText size={24} className="text-purple-400" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-white">Detalles adicionales</h3>
                    <p className="text-sm text-slate-400">Completa la informacion requerida</p>
                  </div>
                </div>

                <DynamicFormFields
                  fields={selectedType?.form_fields?.fields || []}
                  values={customFieldValues}
                  onChange={setCustomFieldValues}
                  disabled={loading}
                />
              </div>
            </div>
          )}

          {/* Final Step: Reason & Review */}
          {((currentStep === 2 && !hasCustomFields) || currentStep === 3) && (
            <div className="animate-in fade-in slide-in-from-right-4 duration-300 space-y-6">
              {/* Reason */}
              <div className="bg-slate-800/50 rounded-3xl border border-slate-700/50 p-8">
                <div className="flex items-center gap-3 mb-6">
                  <div className="p-3 rounded-xl bg-amber-500/20">
                    <AlertCircle size={24} className="text-amber-400" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-white">Motivo de la solicitud</h3>
                    <p className="text-sm text-slate-400">Describe brevemente el motivo</p>
                  </div>
                </div>

                <Textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="Escribe el motivo de tu solicitud..."
                  className="bg-slate-900/50 border-slate-600 text-white min-h-[120px] rounded-xl focus:ring-2 focus:ring-amber-500/50 focus:border-amber-500 resize-none"
                />
              </div>

              {/* Summary Card */}
              {selectedType && startDate && endDate && (
                <div className="bg-gradient-to-br from-slate-800/80 to-slate-900/80 rounded-3xl border border-slate-600/50 p-8 shadow-xl">
                  <div className="flex items-center gap-3 mb-6">
                    <div className="p-3 rounded-xl bg-emerald-500/20">
                      <FileCheck size={24} className="text-emerald-400" />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-white">Resumen de tu solicitud</h3>
                      <p className="text-sm text-slate-400">Verifica que los datos sean correctos</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="bg-slate-900/50 rounded-xl p-4">
                      <p className="text-xs text-slate-500 uppercase tracking-wider mb-1">Tipo</p>
                      <p className="text-white font-medium">{selectedType.name}</p>
                    </div>
                    <div className="bg-slate-900/50 rounded-xl p-4">
                      <p className="text-xs text-slate-500 uppercase tracking-wider mb-1">Duracion</p>
                      <p className="text-white font-medium">{totalDays} dia{totalDays !== 1 ? "s" : ""}</p>
                    </div>
                    <div className="bg-slate-900/50 rounded-xl p-4">
                      <p className="text-xs text-slate-500 uppercase tracking-wider mb-1">Fecha inicio</p>
                      <p className="text-white font-medium capitalize">{formatDate(startDate)}</p>
                    </div>
                    <div className="bg-slate-900/50 rounded-xl p-4">
                      <p className="text-xs text-slate-500 uppercase tracking-wider mb-1">Fecha fin</p>
                      <p className="text-white font-medium capitalize">{formatDate(endDate)}</p>
                    </div>
                  </div>

                  {selectedType.requires_evidence && (
                    <div className="mt-6 space-y-4">
                      <div className="flex items-center gap-2">
                        <Paperclip size={18} className="text-amber-400" />
                        <h4 className="text-white font-medium">Evidencia Requerida</h4>
                        <span className="text-xs bg-amber-500/20 text-amber-400 px-2 py-0.5 rounded-full">
                          Obligatorio
                        </span>
                      </div>

                      {/* File Upload Area */}
                      <div className="bg-slate-900/50 border-2 border-dashed border-slate-700 rounded-xl p-6 text-center hover:border-blue-500/50 transition-colors">
                        <input
                          type="file"
                          id="evidence-upload"
                          multiple
                          accept="image/*,.pdf,.doc,.docx"
                          className="hidden"
                          onChange={(e) => {
                            const files = Array.from(e.target.files || [])
                            setEvidenceFiles(prev => [...prev, ...files])
                            e.target.value = "" // Reset to allow re-selecting same file
                          }}
                        />
                        <label
                          htmlFor="evidence-upload"
                          className="cursor-pointer flex flex-col items-center gap-3"
                        >
                          <div className="w-14 h-14 rounded-full bg-blue-500/20 flex items-center justify-center">
                            <Upload size={24} className="text-blue-400" />
                          </div>
                          <div>
                            <p className="text-white font-medium">Arrastra archivos aqui o haz clic para seleccionar</p>
                            <p className="text-slate-500 text-sm mt-1">
                              Imagenes, PDF o documentos Word (max 10MB cada uno)
                            </p>
                          </div>
                        </label>
                      </div>

                      {/* Selected Files List */}
                      {evidenceFiles.length > 0 && (
                        <div className="space-y-2">
                          <p className="text-sm text-slate-400">{evidenceFiles.length} archivo(s) seleccionado(s):</p>
                          {evidenceFiles.map((file, index) => (
                            <div
                              key={`${file.name}-${index}`}
                              className="flex items-center justify-between bg-slate-800/50 rounded-lg px-4 py-3"
                            >
                              <div className="flex items-center gap-3">
                                <FileCheck size={18} className="text-emerald-400" />
                                <div>
                                  <p className="text-white text-sm font-medium truncate max-w-[200px]">{file.name}</p>
                                  <p className="text-slate-500 text-xs">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
                                </div>
                              </div>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => setEvidenceFiles(prev => prev.filter((_, i) => i !== index))}
                                className="text-slate-400 hover:text-red-400 hover:bg-red-500/10"
                              >
                                <X size={16} />
                              </Button>
                            </div>
                          ))}
                        </div>
                      )}

                      {evidenceFiles.length === 0 && (
                        <p className="text-amber-400/80 text-sm flex items-center gap-2">
                          <AlertCircle size={14} />
                          Debes adjuntar al menos un archivo de evidencia para esta solicitud
                        </p>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Navigation Buttons */}
        <div className="flex gap-4 mt-8 pt-6 border-t border-slate-800">
          {currentStep > 0 && (
            <Button
              type="button"
              variant="outline"
              onClick={handleBack}
              className="flex-1 h-14 rounded-xl border-slate-600 text-slate-300 hover:bg-slate-800 text-lg"
            >
              <ArrowLeft size={20} className="mr-2" />
              Anterior
            </Button>
          )}

          {currentStep === 0 && (
            <Button
              type="button"
              variant="outline"
              onClick={() => router.back()}
              className="flex-1 h-14 rounded-xl border-slate-600 text-slate-300 hover:bg-slate-800 text-lg"
            >
              Cancelar
            </Button>
          )}

          {currentStep < totalSteps - 1 ? (
            <Button
              type="button"
              onClick={handleNext}
              disabled={currentStep === 0 && !selectedType}
              className="flex-1 h-14 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-lg shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Siguiente
              <ArrowRight size={20} className="ml-2" />
            </Button>
          ) : (
            <Button
              type="button"
              onClick={handleSubmit}
              disabled={loading || uploadingEvidence || !reason || (selectedType?.requires_evidence && evidenceFiles.length === 0)}
              className="flex-1 h-14 rounded-xl bg-gradient-to-r from-emerald-600 to-emerald-700 hover:from-emerald-700 hover:to-emerald-800 text-lg shadow-lg shadow-emerald-500/20 disabled:opacity-50"
            >
              {loading || uploadingEvidence ? (
                <>
                  <Loader2 size={20} className="mr-2 animate-spin" />
                  {uploadingEvidence ? "Subiendo evidencia..." : "Enviando..."}
                </>
              ) : (
                <>
                  <Send size={20} className="mr-2" />
                  Enviar Solicitud
                </>
              )}
            </Button>
          )}
        </div>
      </div>
    </PortalLayout>
  )
}

function LoadingFallback() {
  return (
    <PortalLayout>
      <div className="flex items-center justify-center h-96">
        <div className="flex flex-col items-center gap-4">
          <div className="relative">
            <div className="w-16 h-16 rounded-full border-4 border-slate-700 border-t-blue-500 animate-spin" />
            <Sparkles className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-blue-400" size={24} />
          </div>
          <p className="text-slate-400 text-lg">Cargando...</p>
        </div>
      </div>
    </PortalLayout>
  )
}

export default function NewRequestPage() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <NewRequestForm />
    </Suspense>
  )
}
