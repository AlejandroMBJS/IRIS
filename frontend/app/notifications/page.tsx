/**
 * @file app/notifications/page.tsx
 * @description System notifications center showing payroll events and alerts
 *
 * USER PERSPECTIVE:
 *   - View all system notifications (success, warning, info, error)
 *   - Filter by type and read/unread status
 *   - Mark individual notifications as read
 *   - Mark all notifications as read at once
 *   - Delete individual or all notifications
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Notification types, filter options, display layout
 *   CAUTION: Currently uses mock data, replace with real API when available
 *   DO NOT modify: Notification structure without updating backend schema
 *
 * KEY COMPONENTS:
 *   - Notification list: Grouped by status with visual indicators
 *   - Filter bar: Type and read/unread filters
 *   - Action buttons: Mark all read, clear all
 *   - Notification cards: Color-coded by type (success, warning, error, info)
 *
 * API ENDPOINTS USED:
 *   - None (currently uses mock data)
 *   - TODO: GET /api/notifications
 *   - TODO: PUT /api/notifications/:id/read
 *   - TODO: DELETE /api/notifications/:id
 */

"use client"

import { useState } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import {
  Bell,
  Check,
  AlertCircle,
  Info,
  Clock,
  CheckCheck,
  Trash2,
  Filter,
} from "lucide-react"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface Notification {
  id: string
  type: "success" | "warning" | "info" | "error"
  title: string
  message: string
  time: string
  date: string
  read: boolean
}

// Mock notifications for demo
const initialNotifications: Notification[] = [
  {
    id: "1",
    type: "success",
    title: "Payroll Processed",
    message: "The payroll for period 48 has been processed successfully. All receipts have been generated.",
    time: "5 min ago",
    date: "2024-12-07",
    read: false,
  },
  {
    id: "2",
    type: "warning",
    title: "Pending Incident",
    message: "There are 3 pending incidents to approve for the current period. Please review and approve before closing.",
    time: "1 hour ago",
    date: "2024-12-07",
    read: false,
  },
  {
    id: "3",
    type: "info",
    title: "New Employee Registered",
    message: "A new employee has been registered: Juan Perez Martinez in the Operations department.",
    time: "2 hours ago",
    date: "2024-12-07",
    read: true,
  },
  {
    id: "4",
    type: "success",
    title: "Period Closed",
    message: "Period 47 has been closed correctly. No further modifications can be made.",
    time: "1 day ago",
    date: "2024-12-06",
    read: true,
  },
  {
    id: "5",
    type: "error",
    title: "Calculation Error",
    message: "There was an error calculating the payroll for employee EMP-045. Please review the data.",
    time: "2 days ago",
    date: "2024-12-05",
    read: true,
  },
  {
    id: "6",
    type: "info",
    title: "System Update",
    message: "The system will be updated next Sunday at 2:00 AM. There may be temporary interruptions.",
    time: "3 days ago",
    date: "2024-12-04",
    read: true,
  },
  {
    id: "7",
    type: "warning",
    title: "Upcoming Vacations",
    message: "5 employees have vacations scheduled for next week. Verify staff coverage.",
    time: "4 days ago",
    date: "2024-12-03",
    read: true,
  },
]

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>(initialNotifications)
  const [filterType, setFilterType] = useState<string>("all")
  const [filterRead, setFilterRead] = useState<string>("all")

  const unreadCount = notifications.filter(n => !n.read).length

  const filteredNotifications = notifications.filter(n => {
    if (filterType !== "all" && n.type !== filterType) return false
    if (filterRead === "unread" && n.read) return false
    if (filterRead === "read" && !n.read) return false
    return true
  })

  const markAsRead = (id: string) => {
    setNotifications(prev =>
      prev.map(n => (n.id === id ? { ...n, read: true } : n))
    )
  }

  const markAllAsRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }

  const deleteNotification = (id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  const clearAll = () => {
    if (confirm("Are you sure you want to delete all notifications?")) {
      setNotifications([])
    }
  }

  const getNotificationIcon = (type: Notification["type"]) => {
    switch (type) {
      case "success":
        return <Check className="h-5 w-5 text-green-400" />
      case "warning":
        return <AlertCircle className="h-5 w-5 text-yellow-400" />
      case "error":
        return <AlertCircle className="h-5 w-5 text-red-400" />
      default:
        return <Info className="h-5 w-5 text-blue-400" />
    }
  }

  const getNotificationBg = (type: Notification["type"], read: boolean) => {
    if (read) return "bg-slate-800/30"
    switch (type) {
      case "success":
        return "bg-green-500/10 border-l-4 border-l-green-500"
      case "warning":
        return "bg-yellow-500/10 border-l-4 border-l-yellow-500"
      case "error":
        return "bg-red-500/10 border-l-4 border-l-red-500"
      default:
        return "bg-blue-500/10 border-l-4 border-l-blue-500"
    }
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white flex items-center gap-2">
              <Bell className="h-6 w-6" />
              Notifications
              {unreadCount > 0 && (
                <span className="ml-2 px-2 py-0.5 bg-red-500 text-white text-sm rounded-full">
                  {unreadCount} new
                </span>
              )}
            </h1>
            <p className="text-slate-400 mt-1">
              System notifications center
            </p>
          </div>
          <div className="flex gap-2">
            {unreadCount > 0 && (
              <Button
                onClick={markAllAsRead}
                variant="outline"
                className="border-slate-600 text-slate-300 hover:bg-slate-800"
              >
                <CheckCheck className="h-4 w-4 mr-2" />
                Mark all as read
              </Button>
            )}
            {notifications.length > 0 && (
              <Button
                onClick={clearAll}
                variant="outline"
                className="border-red-600 text-red-400 hover:bg-red-600/20"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Clear all
              </Button>
            )}
          </div>
        </div>

        {/* Filters */}
        <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
          <div className="flex items-center gap-4">
            <Filter className="h-5 w-5 text-slate-400" />
            <div className="flex gap-4">
              <Select value={filterType} onValueChange={setFilterType}>
                <SelectTrigger className="w-[180px] bg-slate-900 border-slate-600">
                  <SelectValue placeholder="Type..." />
                </SelectTrigger>
                <SelectContent className="bg-slate-800 border-slate-600">
                  <SelectItem value="all">All types</SelectItem>
                  <SelectItem value="success">Success</SelectItem>
                  <SelectItem value="warning">Warning</SelectItem>
                  <SelectItem value="info">Information</SelectItem>
                  <SelectItem value="error">Error</SelectItem>
                </SelectContent>
              </Select>

              <Select value={filterRead} onValueChange={setFilterRead}>
                <SelectTrigger className="w-[180px] bg-slate-900 border-slate-600">
                  <SelectValue placeholder="Status..." />
                </SelectTrigger>
                <SelectContent className="bg-slate-800 border-slate-600">
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="unread">Unread</SelectItem>
                  <SelectItem value="read">Read</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Notifications List */}
        <div className="space-y-3">
          {filteredNotifications.length === 0 ? (
            <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-12 text-center">
              <Bell className="h-12 w-12 text-slate-600 mx-auto mb-4" />
              <p className="text-slate-400">No notifications</p>
            </div>
          ) : (
            filteredNotifications.map((notification) => (
              <div
                key={notification.id}
                className={`rounded-lg border border-slate-700 p-4 transition-colors ${getNotificationBg(
                  notification.type,
                  notification.read
                )}`}
              >
                <div className="flex items-start gap-4">
                  <div className="mt-0.5">
                    {getNotificationIcon(notification.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-white">
                        {notification.title}
                      </h3>
                      <div className="flex items-center gap-2">
                        {!notification.read && (
                          <span className="w-2 h-2 bg-blue-500 rounded-full" />
                        )}
                        <div className="flex items-center gap-1 text-slate-500 text-sm">
                          <Clock className="h-3 w-3" />
                          {notification.time}
                        </div>
                      </div>
                    </div>
                    <p className="text-slate-400 mt-1">{notification.message}</p>
                    <div className="flex items-center justify-between mt-3">
                      <span className="text-xs text-slate-500">
                        {notification.date}
                      </span>
                      <div className="flex gap-2">
                        {!notification.read && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => markAsRead(notification.id)}
                            className="text-blue-400 hover:text-blue-300 text-xs"
                          >
                            <Check className="h-3 w-3 mr-1" />
                            Mark as read
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => deleteNotification(notification.id)}
                          className="text-red-400 hover:text-red-300 text-xs"
                        >
                          <Trash2 className="h-3 w-3 mr-1" />
                          Delete
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </DashboardLayout>
  )
}
