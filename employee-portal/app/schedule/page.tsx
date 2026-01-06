"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Calendar as CalendarIcon,
  ChevronLeft,
  ChevronRight,
  Plus,
  Clock,
  Briefcase,
  Coffee,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { isAuthenticated, getCurrentUser } from "@/lib/auth"
import { PortalLayout } from "@/components/layout/portal-layout"
import {
  absenceRequestApi,
  AbsenceRequest,
  REQUEST_TYPE_LABELS,
} from "@/lib/api-client"
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameMonth, isSameDay, addMonths, subMonths, startOfWeek, endOfWeek, isToday } from "date-fns"
import { es } from "date-fns/locale"

export default function SchedulePage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [currentDate, setCurrentDate] = useState(new Date())
  const [requests, setRequests] = useState<AbsenceRequest[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    setUser(getCurrentUser())
    loadScheduleData()
  }, [router, currentDate])

  async function loadScheduleData() {
    try {
      setLoading(true)
      const response = await absenceRequestApi.getMyRequests()
      setRequests(response.requests || [])
    } catch (error) {
      console.error("Error loading schedule data:", error)
    } finally {
      setLoading(false)
    }
  }

  const monthStart = startOfMonth(currentDate)
  const monthEnd = endOfMonth(currentDate)
  const calendarStart = startOfWeek(monthStart, { weekStartsOn: 1 }) // Monday
  const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 1 })
  const calendarDays = eachDayOfInterval({ start: calendarStart, end: calendarEnd })

  const getRequestsForDay = (day: Date) => {
    return requests.filter((req) => {
      if (!req.start_date || !req.end_date) return false
      const start = new Date(req.start_date)
      const end = new Date(req.end_date)
      return day >= start && day <= end && req.status === 'APPROVED'
    })
  }

  const getDayStatus = (day: Date) => {
    const dayRequests = getRequestsForDay(day)
    if (dayRequests.length === 0) return "work"

    const hasVacation = dayRequests.some(r => r.request_type === 'VACATION')
    const hasSickLeave = dayRequests.some(r => r.request_type === 'SICK_LEAVE')
    const hasPaidLeave = dayRequests.some(r => r.request_type === 'PAID_LEAVE')

    if (hasVacation) return "vacation"
    if (hasSickLeave) return "sick"
    if (hasPaidLeave) return "paid-leave"
    return "other"
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case "vacation":
        return "bg-blue-500/20 border-blue-500/50 text-blue-400"
      case "sick":
        return "bg-red-500/20 border-red-500/50 text-red-400"
      case "paid-leave":
        return "bg-emerald-500/20 border-emerald-500/50 text-emerald-400"
      case "other":
        return "bg-amber-500/20 border-amber-500/50 text-amber-400"
      default:
        return ""
    }
  }

  const weekDays = ["Lun", "Mar", "Mié", "Jue", "Vie", "Sáb", "Dom"]

  if (loading) {
    return (
      <PortalLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Cargando calendario...</p>
          </div>
        </div>
      </PortalLayout>
    )
  }

  return (
    <PortalLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white flex items-center gap-3">
              <CalendarIcon className="h-8 w-8 text-blue-400" />
              Mi Calendario
            </h1>
            <p className="text-slate-400 mt-2">
              Visualiza tus dias de trabajo y ausencias programadas
            </p>
          </div>
          <Button
            onClick={() => router.push("/requests/new")}
            className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800"
          >
            <Plus size={18} className="mr-2" />
            Nueva Solicitud
          </Button>
        </div>

        {/* Calendar Card */}
        <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 overflow-hidden">
          {/* Month Navigation */}
          <div className="bg-slate-900/50 border-b border-slate-700 p-4">
            <div className="flex items-center justify-between">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setCurrentDate(subMonths(currentDate, 1))}
                className="text-slate-400 hover:text-white hover:bg-slate-700"
              >
                <ChevronLeft size={20} />
              </Button>
              <h2 className="text-xl font-semibold text-white">
                {format(currentDate, "MMMM yyyy", { locale: es })}
              </h2>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setCurrentDate(addMonths(currentDate, 1))}
                className="text-slate-400 hover:text-white hover:bg-slate-700"
              >
                <ChevronRight size={20} />
              </Button>
            </div>
          </div>

          {/* Calendar Grid */}
          <div className="p-6">
            {/* Week Day Headers */}
            <div className="grid grid-cols-7 gap-2 mb-2">
              {weekDays.map((day) => (
                <div
                  key={day}
                  className="text-center text-sm font-semibold text-slate-400 py-2"
                >
                  {day}
                </div>
              ))}
            </div>

            {/* Calendar Days */}
            <div className="grid grid-cols-7 gap-2">
              {calendarDays.map((day, idx) => {
                const dayStatus = getDayStatus(day)
                const dayRequests = getRequestsForDay(day)
                const isCurrentMonth = isSameMonth(day, currentDate)
                const isCurrentDay = isToday(day)

                return (
                  <div
                    key={idx}
                    className={`min-h-24 p-2 rounded-lg border transition-all ${
                      isCurrentMonth
                        ? dayStatus !== "work"
                          ? getStatusColor(dayStatus)
                          : "bg-slate-900/30 border-slate-700 hover:bg-slate-800/50"
                        : "bg-slate-900/10 border-slate-800/50"
                    } ${isCurrentDay ? "ring-2 ring-blue-500" : ""}`}
                  >
                    <div className="flex flex-col h-full">
                      <div className={`text-sm font-medium mb-1 ${
                        isCurrentMonth ? "text-white" : "text-slate-600"
                      }`}>
                        {format(day, "d")}
                      </div>

                      {dayRequests.length > 0 && isCurrentMonth && (
                        <div className="flex-1 space-y-1">
                          {dayRequests.slice(0, 2).map((req) => (
                            <div
                              key={req.id}
                              className="text-xs px-1.5 py-0.5 rounded truncate"
                              title={REQUEST_TYPE_LABELS[req.request_type]}
                            >
                              {REQUEST_TYPE_LABELS[req.request_type]}
                            </div>
                          ))}
                          {dayRequests.length > 2 && (
                            <div className="text-xs text-slate-400">
                              +{dayRequests.length - 2} más
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </Card>

        {/* Legend */}
        <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Leyenda</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="flex items-center gap-3">
              <div className="w-4 h-4 rounded bg-slate-900/30 border border-slate-700" />
              <span className="text-sm text-slate-300">Dia laboral</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-4 h-4 rounded bg-blue-500/20 border border-blue-500/50" />
              <span className="text-sm text-slate-300">Vacaciones</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-4 h-4 rounded bg-red-500/20 border border-red-500/50" />
              <span className="text-sm text-slate-300">Incapacidad</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-4 h-4 rounded bg-emerald-500/20 border border-emerald-500/50" />
              <span className="text-sm text-slate-300">Permiso con goce</span>
            </div>
          </div>
        </Card>

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card className="bg-gradient-to-br from-blue-900/30 to-blue-800/30 backdrop-blur-sm border-blue-700/50 p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-blue-300 text-sm font-medium mb-1">Este Mes</p>
                <p className="text-3xl font-bold text-white">
                  {requests.filter(r => {
                    const start = new Date(r.start_date)
                    return isSameMonth(start, currentDate) && r.status === 'APPROVED'
                  }).length}
                </p>
                <p className="text-xs text-blue-300/70 mt-1">Ausencias aprobadas</p>
              </div>
              <div className="p-3 bg-blue-500/20 rounded-xl">
                <CalendarIcon className="h-6 w-6 text-blue-400" />
              </div>
            </div>
          </Card>

          <Card className="bg-gradient-to-br from-emerald-900/30 to-emerald-800/30 backdrop-blur-sm border-emerald-700/50 p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-emerald-300 text-sm font-medium mb-1">Dias Trabajados</p>
                <p className="text-3xl font-bold text-white">
                  {calendarDays.filter(day =>
                    isSameMonth(day, currentDate) &&
                    getDayStatus(day) === 'work' &&
                    day <= new Date()
                  ).length}
                </p>
                <p className="text-xs text-emerald-300/70 mt-1">En {format(currentDate, "MMMM", { locale: es })}</p>
              </div>
              <div className="p-3 bg-emerald-500/20 rounded-xl">
                <Briefcase className="h-6 w-6 text-emerald-400" />
              </div>
            </div>
          </Card>

          <Card className="bg-gradient-to-br from-amber-900/30 to-amber-800/30 backdrop-blur-sm border-amber-700/50 p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-amber-300 text-sm font-medium mb-1">Pendientes</p>
                <p className="text-3xl font-bold text-white">
                  {requests.filter(r => r.status === 'PENDING').length}
                </p>
                <p className="text-xs text-amber-300/70 mt-1">Solicitudes en revisión</p>
              </div>
              <div className="p-3 bg-amber-500/20 rounded-xl">
                <Clock className="h-6 w-6 text-amber-400" />
              </div>
            </div>
          </Card>
        </div>
      </div>
    </PortalLayout>
  )
}
