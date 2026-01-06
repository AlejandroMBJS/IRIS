"use client"

import { useEffect, useState, useRef } from "react"
import { useRouter, usePathname } from "next/navigation"
import Link from "next/link"
import {
  Home,
  Users,
  Calendar,
  DollarSign,
  FileText,
  Settings,
  ChevronDown,
  ChevronRight,
  LogOut,
  Menu,
  X,
  UserCog,
  ClipboardList,
  Palmtree,
  UserX,
  Tag,
  Bell,
  Check,
  AlertCircle,
  Info,
  Clock,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { isAuthenticated, logout, getCurrentUser, isAdmin } from "@/lib/auth"
import { notificationApi, Notification, NotificationType } from "@/lib/api-client"

// Map notification types to icon types
type IconType = "success" | "warning" | "info" | "error"
const notificationTypeToIcon: Record<NotificationType, IconType> = {
  employee_created: "info",
  employee_updated: "info",
  incidence_created: "warning",
  incidence_approved: "success",
  incidence_rejected: "error",
  payroll_calculated: "success",
  period_created: "info",
  user_created: "info",
}

// Format time as relative (e.g., "Hace 5 min")
const formatRelativeTime = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMin / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMin < 1) return "Hace un momento"
  if (diffMin < 60) return `Hace ${diffMin} min`
  if (diffHours < 24) return `Hace ${diffHours} hora${diffHours > 1 ? "s" : ""}`
  if (diffDays < 7) return `Hace ${diffDays} dia${diffDays > 1 ? "s" : ""}`
  return date.toLocaleDateString("es-MX", { day: "numeric", month: "short" })
}

interface DashboardLayoutProps {
  children: React.ReactNode
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const router = useRouter()
  const pathname = usePathname()
  const [user, setUser] = useState<any>(null)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [userIsAdmin, setUserIsAdmin] = useState(false)
  const [mounted, setMounted] = useState(false)
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({
    payroll: true,
    catalogs: true,
    incidences: true,
    admin: true,
  })

  // Navbar state
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [showNotifications, setShowNotifications] = useState(false)
  const notificationRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setMounted(true)
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }
    setUser(getCurrentUser())
    setUserIsAdmin(isAdmin())
    loadNotifications()
    // Poll for new notifications every 30 seconds
    const interval = setInterval(loadNotifications, 30000)
    return () => clearInterval(interval)
  }, [router])

  // Load notifications from API
  const loadNotifications = async () => {
    try {
      const [notifs, count] = await Promise.all([
        notificationApi.getNotifications(20),
        notificationApi.getUnreadCount(),
      ])
      setNotifications(notifs)
      setUnreadCount(count)
    } catch (err) {
      console.error("Failed to load notifications:", err)
    }
  }

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (notificationRef.current && !notificationRef.current.contains(event.target as Node)) {
        setShowNotifications(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const handleLogout = () => {
    logout()
  }

  const markAsRead = async (id: string) => {
    try {
      await notificationApi.markAsRead(id)
      setNotifications(prev =>
        prev.map(n => (n.id === id ? { ...n, read: true } : n))
      )
      setUnreadCount(prev => Math.max(0, prev - 1))
    } catch (err) {
      console.error("Failed to mark notification as read:", err)
    }
  }

  const markAllAsRead = async () => {
    try {
      await notificationApi.markAllAsRead()
      setNotifications(prev => prev.map(n => ({ ...n, read: true })))
      setUnreadCount(0)
    } catch (err) {
      console.error("Failed to mark all notifications as read:", err)
    }
  }

  const getNotificationIcon = (type: NotificationType) => {
    const iconType = notificationTypeToIcon[type] || "info"
    switch (iconType) {
      case "success":
        return <Check className="h-4 w-4 text-green-400" />
      case "warning":
        return <AlertCircle className="h-4 w-4 text-yellow-400" />
      case "error":
        return <AlertCircle className="h-4 w-4 text-red-400" />
      default:
        return <Info className="h-4 w-4 text-blue-400" />
    }
  }

  const toggleSection = (section: string) => {
    setOpenSections((prev) => ({ ...prev, [section]: !prev[section] }))
  }

  // Check if a path is active - for exact matches only, not parent paths
  // This prevents /incidences from being highlighted when on /incidences/types
  const isActive = (path: string) => pathname === path

  const NavLink = ({ href, children, indent = false }: { href: string; children: React.ReactNode; indent?: boolean }) => (
    <Link
      href={href}
      className={`w-full flex items-center gap-3 px-4 py-2.5 rounded-lg transition-all duration-200 ${
        indent ? "ml-6 text-sm" : ""
      } ${
        isActive(href)
          ? "bg-blue-600 text-white shadow-lg"
          : "text-slate-300 hover:bg-slate-700 hover:text-white"
      }`}
    >
      {children}
    </Link>
  )

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-950 flex">
      {/* Sidebar */}
      <aside
        className={`${
          sidebarOpen ? "w-64" : "w-0 overflow-hidden"
        } transition-all duration-300 bg-slate-900/50 border-r border-slate-800 flex flex-col`}
      >
        {/* Logo */}
        <div className="p-6 border-b border-slate-800">
          <div className="flex items-baseline gap-2">
            <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
              IRIS
            </h1>
            <span className="text-lg font-medium text-slate-400">Talent</span>
          </div>
          <p className="text-xs text-slate-500 mt-1">HR & Payroll System</p>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto p-4 space-y-1">
          <NavLink href="/dashboard">
            <Home size={20} />
            <span className="font-medium">Dashboard</span>
          </NavLink>

          {/* Catalogs Section */}
          <div className="pt-4">
            <button
              onClick={() => toggleSection("catalogs")}
              className="w-full flex items-center justify-between px-4 py-2 text-slate-400 hover:text-white transition-all"
            >
              <div className="flex items-center gap-3">
                <Settings size={20} />
                <span className="font-semibold">Catalogs</span>
              </div>
              {openSections.catalogs ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
            </button>
            {openSections.catalogs && (
              <div className="mt-1 space-y-1">
                <NavLink href="/employees" indent>
                  <Users size={18} />
                  <span>Employees</span>
                </NavLink>
                <NavLink href="/payroll/periods" indent>
                  <Calendar size={18} />
                  <span>Payroll Periods</span>
                </NavLink>
              </div>
            )}
          </div>

          {/* Incidences Section (HR) */}
          <div className="pt-4">
            <button
              onClick={() => toggleSection("incidences")}
              className="w-full flex items-center justify-between px-4 py-2 text-slate-400 hover:text-white transition-all"
            >
              <div className="flex items-center gap-3">
                <ClipboardList size={20} />
                <span className="font-semibold">Incidencias</span>
              </div>
              {openSections.incidences ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
            </button>
            {openSections.incidences && (
              <div className="mt-1 space-y-1">
                <NavLink href="/incidences" indent>
                  <UserX size={18} />
                  <span>Incidencias</span>
                </NavLink>
                <NavLink href="/incidences/types" indent>
                  <Tag size={18} />
                  <span>Tipos de Incidencia</span>
                </NavLink>
                <NavLink href="/incidences/vacations" indent>
                  <Palmtree size={18} />
                  <span>Vacaciones</span>
                </NavLink>
              </div>
            )}
          </div>

          {/* Payroll Section */}
          <div className="pt-4">
            <button
              onClick={() => toggleSection("payroll")}
              className="w-full flex items-center justify-between px-4 py-2 text-slate-400 hover:text-white transition-all"
            >
              <div className="flex items-center gap-3">
                <DollarSign size={20} />
                <span className="font-semibold">Payroll</span>
              </div>
              {openSections.payroll ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
            </button>
            {openSections.payroll && (
              <div className="mt-1 space-y-1">
                <NavLink href="/payroll" indent>
                  <FileText size={18} />
                  <span>Run Payroll</span>
                </NavLink>
              </div>
            )}
          </div>

          {/* Admin Only Section */}
          {mounted && userIsAdmin && (
            <div className="pt-4">
              <button
                onClick={() => toggleSection("admin")}
                className="w-full flex items-center justify-between px-4 py-2 text-slate-400 hover:text-white transition-all"
              >
                <div className="flex items-center gap-3">
                  <Settings size={20} />
                  <span className="font-semibold">Admin</span>
                </div>
                {openSections.admin ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
              </button>
              {openSections.admin && (
                <div className="mt-1 space-y-1">
                  <NavLink href="/admin/users" indent>
                    <UserCog size={18} />
                    <span>Usuarios</span>
                  </NavLink>
                  <NavLink href="/configuration/payroll-setup" indent>
                    <Settings size={18} />
                    <span>Configuracion</span>
                  </NavLink>
                </div>
              )}
            </div>
          )}
        </nav>

        {/* Version */}
        <div className="p-4 border-t border-slate-800">
          <div className="text-xs text-slate-500 text-center">Version 1.0.0</div>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <header className="bg-slate-900/50 border-b border-slate-800 px-6 py-3">
          <div className="flex items-center justify-between">
            {/* Left: Menu toggle */}
            <div className="flex items-center gap-4">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
              >
                {sidebarOpen ? <X size={20} /> : <Menu size={20} />}
              </button>
            </div>

            {/* Right: Notifications, User */}
            <div className="flex items-center gap-3">
              {/* Notifications */}
              <div ref={notificationRef} className="relative">
                <button
                  onClick={() => setShowNotifications(!showNotifications)}
                  className="relative p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
                >
                  <Bell size={20} />
                  {unreadCount > 0 && (
                    <span className="absolute top-1 right-1 w-4 h-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                      {unreadCount}
                    </span>
                  )}
                </button>

                {/* Notifications Dropdown */}
                {showNotifications && (
                  <div className="absolute right-0 top-full mt-2 w-80 bg-slate-800 border border-slate-700 rounded-lg shadow-xl z-50">
                    <div className="flex items-center justify-between p-3 border-b border-slate-700">
                      <h3 className="font-semibold text-white">Notificaciones</h3>
                      {unreadCount > 0 && (
                        <button
                          onClick={markAllAsRead}
                          className="text-xs text-blue-400 hover:text-blue-300"
                        >
                          Marcar todas como leidas
                        </button>
                      )}
                    </div>
                    <div className="max-h-80 overflow-y-auto">
                      {notifications.length === 0 ? (
                        <div className="p-4 text-center text-slate-500 text-sm">
                          No hay notificaciones
                        </div>
                      ) : (
                        notifications.map((notification) => (
                          <div
                            key={notification.id}
                            onClick={() => markAsRead(notification.id)}
                            className={`p-3 border-b border-slate-700 hover:bg-slate-700/50 cursor-pointer transition-colors ${
                              !notification.read ? "bg-slate-700/30" : ""
                            }`}
                          >
                            <div className="flex items-start gap-3">
                              <div className="mt-0.5">
                                {getNotificationIcon(notification.type)}
                              </div>
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center justify-between">
                                  <p className="font-medium text-white text-sm">
                                    {notification.title}
                                  </p>
                                  {!notification.read && (
                                    <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0" />
                                  )}
                                </div>
                                <p className="text-slate-400 text-xs mt-0.5 truncate">
                                  {notification.message}
                                </p>
                                <div className="flex items-center justify-between mt-1 text-slate-500 text-xs">
                                  <div className="flex items-center gap-1">
                                    <Clock className="h-3 w-3" />
                                    {formatRelativeTime(notification.created_at)}
                                  </div>
                                  {notification.actor_user && (
                                    <span className="text-slate-600">
                                      por {notification.actor_user.full_name}
                                    </span>
                                  )}
                                </div>
                              </div>
                            </div>
                          </div>
                        ))
                      )}
                    </div>
                    <div className="p-2 border-t border-slate-700">
                      <Link
                        href="/notifications"
                        className="block w-full text-center text-sm text-blue-400 hover:text-blue-300 py-2"
                      >
                        Ver todas las notificaciones
                      </Link>
                    </div>
                  </div>
                )}
              </div>

              {/* User info */}
              <div className="hidden sm:flex items-center gap-3 pl-3 border-l border-slate-700">
                <span className="text-slate-400 text-sm">
                  <span className="text-white font-medium">{user?.full_name || user?.email}</span>
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleLogout}
                  className="text-slate-400 hover:text-white hover:bg-slate-800"
                >
                  <LogOut size={18} />
                </Button>
              </div>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 overflow-auto p-6">{children}</main>
      </div>
    </div>
  )
}
