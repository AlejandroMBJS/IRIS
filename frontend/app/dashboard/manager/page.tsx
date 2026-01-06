/**
 * @file app/dashboard/manager/page.tsx
 * @description Manager Dashboard - View all subordinates (recursive hierarchy), team analytics
 *
 * USER PERSPECTIVE:
 *   - View entire team hierarchy (direct + indirect reports)
 *   - Team performance metrics and analytics
 *   - Export team reports
 *   - Approve requests from all subordinates
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Add analytics widgets, team performance charts
 *   CAUTION: Shows all subordinates recursively (filtered by backend)
 *   DO NOT modify: Access control - manager role only
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Users, TrendingUp, CheckCircle2, Download, Calendar,
  BarChart3, UserCheck, ChevronRight, FileText, AlertCircle
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { isAuthenticated, getCurrentUser, isManager } from "@/lib/auth"
import { employeeApi, Employee } from "@/lib/api-client"

export default function ManagerDashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    const currentUser = getCurrentUser()
    if (!isManager()) {
      router.push("/dashboard")
      return
    }

    setUser(currentUser)
    loadDashboardData()
  }, [router])

  async function loadDashboardData() {
    try {
      // Get employees - backend automatically filters to show all subordinates recursively
      const empResponse = await employeeApi.getEmployees().catch(() => ({ employees: [] }))
      setEmployees(empResponse.employees || [])
    } catch (error) {
      console.error("Error loading dashboard data:", error)
    } finally {
      setLoading(false)
    }
  }

  const stats = {
    totalTeam: employees.length,
    activeTeam: employees.filter(e => e.employment_status === "active").length,
    whiteCollar: employees.filter(e => e.collar_type === "white_collar").length,
    blueCollar: employees.filter(e => e.collar_type === "blue_collar").length,
    grayCollar: employees.filter(e => e.collar_type === "gray_collar").length,
    avgSalary: employees.length > 0 ? employees.reduce((sum, e) => sum + (e.daily_salary || 0), 0) / employees.length : 0,
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
      minimumFractionDigits: 2,
    }).format(amount || 0)
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Loading dashboard...</p>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-8">
        {/* Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white">Manager Dashboard</h1>
            <p className="text-slate-400 mt-1">
              Welcome, {user?.full_name || user?.email} - Full Team Management
            </p>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={() => router.push("/approvals")}
              className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-lg shadow-blue-500/20"
            >
              <CheckCircle2 size={18} className="mr-2" />
              Approvals
            </Button>
            <Button
              onClick={() => router.push("/reports")}
              variant="outline"
              className="border-slate-600 text-slate-300 hover:bg-slate-800"
            >
              <BarChart3 size={18} className="mr-2" />
              Reports
            </Button>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          {/* Total Team */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-blue-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Total Team</p>
                <p className="text-4xl font-bold text-white">{stats.activeTeam}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-emerald-400 flex items-center">
                    <CheckCircle2 size={14} className="mr-1" />
                    {stats.totalTeam} total
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl shadow-lg shadow-blue-500/30">
                <Users size={24} className="text-white" />
              </div>
            </div>
          </div>

          {/* Average Salary */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-emerald-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Average Salary</p>
                <p className="text-4xl font-bold text-white">{formatCurrency(stats.avgSalary)}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-emerald-400 flex items-center">
                    <TrendingUp size={14} className="mr-1" />
                    Daily
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/30">
                <TrendingUp size={24} className="text-white" />
              </div>
            </div>
          </div>

          {/* White Collar */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-purple-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">White Collar</p>
                <p className="text-4xl font-bold text-white">{stats.whiteCollar}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-purple-400 text-xs">
                    Administrative
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl shadow-lg shadow-purple-500/30">
                <UserCheck size={24} className="text-white" />
              </div>
            </div>
          </div>

          {/* Blue/Gray Collar */}
          <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-indigo-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-indigo-500/10">
            <div className="flex items-start justify-between">
              <div className="space-y-3">
                <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Blue/Gray Collar</p>
                <p className="text-4xl font-bold text-white">{stats.blueCollar + stats.grayCollar}</p>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-indigo-400 text-xs">
                    Operational
                  </span>
                </div>
              </div>
              <div className="p-4 bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-xl shadow-lg shadow-indigo-500/30">
                <Users size={24} className="text-white" />
              </div>
            </div>
          </div>
        </div>

        {/* Two Column Layout */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
          {/* Team Distribution */}
          <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
            <div className="p-6 border-b border-slate-700/50">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <BarChart3 size={20} className="text-emerald-400" />
                Team Distribution
              </h3>
            </div>
            <div className="p-6 space-y-4">
              {/* White Collar */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-300 flex items-center gap-2">
                    <span className="w-3 h-3 rounded-full bg-blue-500"></span>
                    White Collar
                  </span>
                  <span className="text-white font-medium">{stats.whiteCollar}</span>
                </div>
                <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-blue-500 to-blue-400 rounded-full transition-all duration-500"
                    style={{ width: `${stats.totalTeam > 0 ? (stats.whiteCollar / stats.totalTeam) * 100 : 0}%` }}
                  />
                </div>
              </div>

              {/* Blue Collar */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-300 flex items-center gap-2">
                    <span className="w-3 h-3 rounded-full bg-indigo-500"></span>
                    Blue Collar
                  </span>
                  <span className="text-white font-medium">{stats.blueCollar}</span>
                </div>
                <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 rounded-full transition-all duration-500"
                    style={{ width: `${stats.totalTeam > 0 ? (stats.blueCollar / stats.totalTeam) * 100 : 0}%` }}
                  />
                </div>
              </div>

              {/* Gray Collar */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-300 flex items-center gap-2">
                    <span className="w-3 h-3 rounded-full bg-slate-400"></span>
                    Gray Collar
                  </span>
                  <span className="text-white font-medium">{stats.grayCollar}</span>
                </div>
                <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-slate-400 to-slate-300 rounded-full transition-all duration-500"
                    style={{ width: `${stats.totalTeam > 0 ? (stats.grayCollar / stats.totalTeam) * 100 : 0}%` }}
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6">
            <h3 className="text-lg font-semibold text-white mb-6 flex items-center gap-2">
              <FileText size={20} className="text-amber-400" />
              Manager Actions
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <button
                onClick={() => router.push("/approvals")}
                className="group relative flex flex-col items-center p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300"
              >
                <div className="p-3 bg-blue-500/20 rounded-xl mb-3 group-hover:scale-110 transition-transform">
                  <CheckCircle2 size={20} className="text-blue-400" />
                </div>
                <span className="text-white font-medium text-sm">Approve</span>
              </button>

              <button
                onClick={() => router.push("/employees")}
                className="group relative flex flex-col items-center p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300"
              >
                <div className="p-3 bg-emerald-500/20 rounded-xl mb-3 group-hover:scale-110 transition-transform">
                  <Users size={20} className="text-emerald-400" />
                </div>
                <span className="text-white font-medium text-sm">Team</span>
              </button>

              <button
                onClick={() => router.push("/reports")}
                className="group relative flex flex-col items-center p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-purple-500/50 transition-all duration-300"
              >
                <div className="p-3 bg-purple-500/20 rounded-xl mb-3 group-hover:scale-110 transition-transform">
                  <BarChart3 size={20} className="text-purple-400" />
                </div>
                <span className="text-white font-medium text-sm">Reports</span>
              </button>

              <button
                onClick={() => router.push("/incidences")}
                className="group relative flex flex-col items-center p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300"
              >
                <div className="p-3 bg-amber-500/20 rounded-xl mb-3 group-hover:scale-110 transition-transform">
                  <Calendar size={20} className="text-amber-400" />
                </div>
                <span className="text-white font-medium text-sm">Incidences</span>
              </button>
            </div>
          </div>
        </div>

        {/* Recent Team Members */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
          <div className="p-6 border-b border-slate-700/50 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Users size={20} className="text-blue-400" />
              Team Members
            </h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push("/employees")}
              className="text-slate-400 hover:text-white"
            >
              View all
              <ChevronRight size={16} className="ml-1" />
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Employee</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Type</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Salary</th>
                  <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {employees.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-6 py-12 text-center">
                      <Users size={48} className="mx-auto text-slate-600 mb-4" />
                      <p className="text-slate-400">No employees in your team</p>
                    </td>
                  </tr>
                ) : (
                  employees.slice(0, 10).map((emp) => (
                    <tr
                      key={emp.id}
                      className="hover:bg-slate-800/30 transition-colors cursor-pointer"
                      onClick={() => router.push(`/employees/${emp.id}`)}
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white font-semibold">
                            {emp.first_name?.[0]}{emp.last_name?.[0]}
                          </div>
                          <div>
                            <p className="text-white font-medium">{emp.full_name}</p>
                            <p className="text-sm text-slate-400">{emp.employee_number}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-full ${
                          emp.collar_type === "white_collar"
                            ? "bg-blue-500/20 text-blue-400"
                            : emp.collar_type === "blue_collar"
                            ? "bg-indigo-500/20 text-indigo-400"
                            : "bg-slate-500/20 text-slate-400"
                        }`}>
                          {emp.collar_type?.replace("_", " ") || "N/A"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white font-medium">{formatCurrency(emp.daily_salary)}</span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-full ${
                          emp.employment_status === "active"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-slate-500/20 text-slate-400"
                        }`}>
                          <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${
                            emp.employment_status === "active" ? "bg-emerald-400" : "bg-slate-400"
                          }`}></span>
                          {emp.employment_status === "active" ? "Active" : emp.employment_status}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </DashboardLayout>
  )
}
