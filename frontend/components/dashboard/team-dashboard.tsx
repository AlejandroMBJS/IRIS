"use client"

import { useRouter } from "next/navigation"
import {
  Users, FileText, Clock, CheckCircle2,
  ChevronRight, UserCheck, Calendar
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Employee } from "@/lib/api-client"

interface TeamDashboardProps {
  employees: Employee[]
  user: any
}

export function TeamDashboard({ employees, user }: TeamDashboardProps) {
  const router = useRouter()

  // Calculate team statistics
  const stats = {
    totalTeam: employees.length,
    activeTeam: employees.filter(e => e.employment_status === "active").length,
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("es-MX", {
      style: "currency",
      currency: "MXN",
      minimumFractionDigits: 2,
    }).format(amount || 0)
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-"
    try {
      return new Date(dateStr).toLocaleDateString("es-MX", {
        day: "2-digit",
        month: "short",
        year: "numeric"
      })
    } catch {
      return dateStr
    }
  }

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">My Team</h1>
          <p className="text-slate-400 mt-1">
            Welcome, {user?.full_name || user?.email} - Supervisor view
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
            onClick={() => router.push("/employees")}
            variant="outline"
            className="border-slate-600 text-slate-300 hover:bg-slate-800"
          >
            <Users size={18} className="mr-2" />
            View Employees
          </Button>
        </div>
      </div>

      {/* Team Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {/* Team Members */}
        <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-blue-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-blue-500/10">
          <div className="flex items-start justify-between">
            <div className="space-y-3">
              <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">My Team</p>
              <p className="text-4xl font-bold text-white">{stats.activeTeam}</p>
              <div className="flex items-center gap-2 text-sm">
                <span className="text-emerald-400 flex items-center">
                  <CheckCircle2 size={14} className="mr-1" />
                  Active employees
                </span>
              </div>
            </div>
            <div className="p-4 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl shadow-lg shadow-blue-500/30">
              <Users size={24} className="text-white" />
            </div>
          </div>
        </div>

        {/* Pending Approvals */}
        <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-amber-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-amber-500/10 cursor-pointer"
          onClick={() => router.push("/approvals")}
        >
          <div className="flex items-start justify-between">
            <div className="space-y-3">
              <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Requests</p>
              <p className="text-4xl font-bold text-white">-</p>
              <div className="flex items-center gap-2 text-sm">
                <span className="text-amber-400 flex items-center">
                  <Clock size={14} className="mr-1" />
                  View pending
                </span>
              </div>
            </div>
            <div className="p-4 bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl shadow-lg shadow-amber-500/30">
              <FileText size={24} className="text-white" />
            </div>
          </div>
        </div>

        {/* Quick Access */}
        <div className="group relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300 hover:shadow-xl hover:shadow-emerald-500/10 cursor-pointer"
          onClick={() => router.push("/announcements")}
        >
          <div className="flex items-start justify-between">
            <div className="space-y-3">
              <p className="text-slate-400 text-sm font-medium uppercase tracking-wider">Announcements</p>
              <p className="text-4xl font-bold text-white">-</p>
              <div className="flex items-center gap-2 text-sm">
                <span className="text-emerald-400 flex items-center">
                  <Calendar size={14} className="mr-1" />
                  View announcements
                </span>
              </div>
            </div>
            <div className="p-4 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/30">
              <UserCheck size={24} className="text-white" />
            </div>
          </div>
        </div>
      </div>

      {/* Team Members Table */}
      <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
        <div className="p-6 border-b border-slate-700/50 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <Users size={20} className="text-blue-400" />
            My Team Members
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
                <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Frequency</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-700/50">
              {employees.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-12 text-center">
                    <Users size={48} className="mx-auto text-slate-600 mb-4" />
                    <p className="text-slate-400">You have no assigned employees</p>
                    <p className="text-sm text-slate-500 mt-2">Employees assigned to your supervision will appear here</p>
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
                      <span className="text-slate-300 text-sm capitalize">
                        {emp.pay_frequency === "weekly" ? "Weekly" : emp.pay_frequency === "biweekly" ? "Biweekly" : emp.pay_frequency || "N/A"}
                      </span>
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
  )
}
