"use client"

import { useEffect, useState } from "react"
import { Bell, Calendar, User, CheckCircle } from "lucide-react"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { announcementApi, Announcement } from "@/lib/api-client"
import { useToast } from "@/hooks/use-toast"
import { format } from "date-fns"
import { es } from "date-fns/locale"
import { PortalLayout } from "@/components/layout/portal-layout"

export default function AnnouncementsPage() {
  const { toast } = useToast()
  const [announcements, setAnnouncements] = useState<Announcement[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadAnnouncements()
  }, [])

  async function loadAnnouncements() {
    try {
      setLoading(true)
      const data = await announcementApi.getAll()
      setAnnouncements(data.announcements || data || [])
    } catch (error) {
      console.error("Failed to load announcements:", error)
      toast({
        title: "Error",
        description: "No se pudieron cargar los comunicados",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  async function handleMarkAsRead(id: string) {
    try {
      await announcementApi.markAsRead(id)
      // Update local state to reflect read status
      setAnnouncements(prev =>
        prev.map(a => a.id === id ? { ...a, is_read: true } : a)
      )
    } catch (error) {
      console.error("Failed to mark as read:", error)
    }
  }

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return "Hace un momento"
    if (diffMins < 60) return `Hace ${diffMins} min`
    if (diffHours < 24) return `Hace ${diffHours} hora${diffHours > 1 ? "s" : ""}`
    if (diffDays < 7) return `Hace ${diffDays} dia${diffDays > 1 ? "s" : ""}`
    return format(date, "d 'de' MMM, yyyy", { locale: es })
  }

  if (loading) {
    return (
      <PortalLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="flex items-center gap-2 text-slate-400">
            <div className="w-5 h-5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin" />
            Cargando comunicados...
          </div>
        </div>
      </PortalLayout>
    )
  }

  return (
    <PortalLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold text-white flex items-center gap-3">
            <Bell className="h-8 w-8 text-blue-400" />
            Comunicados
          </h1>
          <p className="text-slate-400 mt-2">Noticias y actualizaciones de la empresa</p>
        </div>

        {/* Announcements List */}
        {announcements.length === 0 ? (
          <Card className="bg-slate-800/50 backdrop-blur-sm border-slate-700 p-12 text-center">
            <Bell className="h-12 w-12 mx-auto text-slate-600 mb-4" />
            <h3 className="text-lg font-medium text-slate-300 mb-2">No hay comunicados</h3>
            <p className="text-slate-500">
              Vuelve mas tarde para ver las actualizaciones de la empresa
            </p>
          </Card>
        ) : (
          <div className="space-y-4">
            {announcements.map((announcement) => (
              <Card
                key={announcement.id}
                className={`bg-slate-800/50 backdrop-blur-sm border-slate-700 p-6 transition-all cursor-pointer hover:border-slate-600 ${
                  !announcement.is_read ? "border-l-4 border-l-blue-500" : ""
                }`}
                onClick={() => !announcement.is_read && handleMarkAsRead(announcement.id)}
              >
                <div className="space-y-3">
                  {/* Header */}
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="text-xl font-semibold text-white">
                          {announcement.title}
                        </h3>
                        {announcement.is_read ? (
                          <CheckCircle className="h-4 w-4 text-green-500" />
                        ) : (
                          <Badge className="bg-blue-600 text-white text-xs">Nuevo</Badge>
                        )}
                      </div>
                      <div className="flex items-center gap-4 text-sm text-slate-400">
                        <div className="flex items-center gap-1.5">
                          <User className="h-4 w-4" />
                          <span>{announcement.created_by?.full_name || "Sistema"}</span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          <Calendar className="h-4 w-4" />
                          <span>{formatRelativeTime(announcement.created_at)}</span>
                        </div>
                        <Badge variant="outline" className="border-slate-600 text-slate-400">
                          {announcement.scope === "ALL" ? "Empresa" : "Equipo"}
                        </Badge>
                      </div>
                    </div>
                  </div>

                  {/* Message */}
                  <div className="prose prose-invert max-w-none">
                    <p className="text-slate-300 whitespace-pre-wrap">{announcement.message}</p>
                  </div>

                  {/* Image */}
                  {announcement.image_data && (
                    <div className="mt-4">
                      <img
                        src={announcement.image_data.startsWith("data:")
                          ? announcement.image_data
                          : `data:image/jpeg;base64,${announcement.image_data}`}
                        alt={announcement.title}
                        className="rounded-lg max-w-md border border-slate-700"
                      />
                    </div>
                  )}

                  {/* Expiration */}
                  {announcement.expires_at && (
                    <div className="flex items-center gap-1.5 text-sm text-amber-400">
                      <Calendar className="h-4 w-4" />
                      <span>Expira el {format(new Date(announcement.expires_at), "d 'de' MMM, yyyy", { locale: es })}</span>
                    </div>
                  )}

                  {/* Click hint for unread */}
                  {!announcement.is_read && (
                    <p className="text-xs text-slate-500 italic">
                      Haz clic para marcar como leido
                    </p>
                  )}
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </PortalLayout>
  )
}
