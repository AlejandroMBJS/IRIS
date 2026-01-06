/**
 * @file app/admin/permissions/page.tsx
 * @description Permission Matrix Configuration Page - Admin can configure role-based access control
 *
 * USER PERSPECTIVE:
 *   - Admin views the full permission matrix (roles × resources)
 *   - Toggle permissions for each role/resource combination
 *   - See at a glance what each role can do
 *   - Changes take effect immediately
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Add new permission types, improve UI layout
 *   CAUTION: Permission changes affect all users with that role
 *   DO NOT modify: Only admin should access this page
 *
 * KEY FEATURES:
 *   - Matrix view: Roles as rows, Resources as columns
 *   - Checkboxes for each permission type (view, create, edit, delete, export, approve)
 *   - Color-coded roles and resources
 *   - Real-time updates
 *   - Filter by role or resource
 */

"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Shield, Eye, Plus, Edit, Trash2, Download, CheckCircle,
  Search, Filter, Save, AlertCircle, RefreshCw
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { isAuthenticated, isAdmin } from "@/lib/auth"
import { permissionApi, Permission, PermissionSet } from "@/lib/api-client"

export default function PermissionsPage() {
  const router = useRouter()
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [roles, setRoles] = useState<string[]>([])
  const [resources, setResources] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [selectedRole, setSelectedRole] = useState<string>("")
  const [selectedResource, setSelectedResource] = useState<string>("")
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null)

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }

    if (!isAdmin()) {
      router.push("/dashboard")
      return
    }

    loadPermissions()
  }, [router])

  async function loadPermissions() {
    try {
      const [permsData, rolesData, resourcesData] = await Promise.all([
        permissionApi.getAllPermissions(),
        permissionApi.getRoles(),
        permissionApi.getResources(),
      ])

      // Handle null/undefined responses
      setPermissions(permsData || [])
      setRoles(rolesData?.roles || [])
      setResources(resourcesData?.resources || [])
    } catch (error) {
      console.error("Error loading permissions:", error)
      showMessage("error", "Error loading permissions. Please check if the backend is running.")
      // Set empty arrays on error
      setPermissions([])
      setRoles([])
      setResources([])
    } finally {
      setLoading(false)
    }
  }

  function showMessage(type: "success" | "error", text: string) {
    setMessage({ type, text })
    setTimeout(() => setMessage(null), 3000)
  }

  function getPermissionForRoleResource(role: string, resource: string): Permission | undefined {
    return (permissions || []).find(p => p.role === role && p.resource === resource)
  }

  // Safe getter for permission value with null checks
  function getPermissionValue(perm: Permission | undefined, key: keyof PermissionSet): boolean {
    if (!perm || !perm.permissions) return false
    return perm.permissions[key] ?? false
  }

  async function togglePermission(role: string, resource: string, permissionType: keyof PermissionSet) {
    const existing = getPermissionForRoleResource(role, resource)

    if (!existing) {
      showMessage("error", "Permission entry not found. Try refreshing the page.")
      return
    }

    setSaving(true)
    try {
      // Ensure we have a valid permissions object
      const currentPerms = existing.permissions || {
        can_view: false,
        can_create: false,
        can_edit: false,
        can_delete: false,
        can_export: false,
        can_approve: false,
      }

      const updatedPerms: PermissionSet = {
        can_view: currentPerms.can_view ?? false,
        can_create: currentPerms.can_create ?? false,
        can_edit: currentPerms.can_edit ?? false,
        can_delete: currentPerms.can_delete ?? false,
        can_export: currentPerms.can_export ?? false,
        can_approve: currentPerms.can_approve ?? false,
        [permissionType]: !getPermissionValue(existing, permissionType),
      }

      const updated = await permissionApi.updatePermission(existing.id, {
        permissions: updatedPerms,
        description: existing.description || "",
        is_active: existing.is_active ?? true,
      })

      // Update local state
      setPermissions((permissions || []).map(p =>
        p.id === updated.id ? updated : p
      ))

      showMessage("success", "Permission updated")
    } catch (error: any) {
      console.error("Error updating permission:", error)
      showMessage("error", error?.message || "Failed to update permission")
    } finally {
      setSaving(false)
    }
  }

  const filteredPermissions = permissions.filter(p => {
    if (selectedRole && p.role !== selectedRole) return false
    if (selectedResource && p.resource !== selectedResource) return false
    return true
  })

  const filteredRoles = selectedRole ? [selectedRole] : roles
  const filteredResources = selectedResource ? [selectedResource] : resources

  const getRoleBadgeColor = (role: string) => {
    const colors: Record<string, string> = {
      admin: "bg-red-500/20 text-red-400 border-red-500/30",
      hr: "bg-blue-500/20 text-blue-400 border-blue-500/30",
      hr_and_pr: "bg-purple-500/20 text-purple-400 border-purple-500/30",
      hr_blue_gray: "bg-indigo-500/20 text-indigo-400 border-indigo-500/30",
      hr_white: "bg-cyan-500/20 text-cyan-400 border-cyan-500/30",
      manager: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
      supervisor: "bg-amber-500/20 text-amber-400 border-amber-500/30",
      payroll_staff: "bg-violet-500/20 text-violet-400 border-violet-500/30",
      accountant: "bg-teal-500/20 text-teal-400 border-teal-500/30",
      employee: "bg-slate-500/20 text-slate-400 border-slate-500/30",
    }
    return colors[role] || "bg-slate-500/20 text-slate-400 border-slate-500/30"
  }

  const formatRoleName = (role: string) => {
    return role.split("_").map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(" ")
  }

  const formatResourceName = (resource: string) => {
    return resource.charAt(0).toUpperCase() + resource.slice(1)
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-slate-400 text-lg">Loading permissions...</p>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white flex items-center gap-3">
              <Shield className="w-8 h-8 text-blue-400" />
              Permission Matrix
            </h1>
            <p className="text-slate-400 mt-1">
              Configure role-based access control for the system
            </p>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={loadPermissions}
              variant="outline"
              className="border-slate-600 text-slate-300 hover:bg-slate-800"
              disabled={saving}
            >
              <RefreshCw size={18} className="mr-2" />
              Refresh
            </Button>
          </div>
        </div>

        {/* Message */}
        {message && (
          <div className={`p-4 rounded-lg border ${
            message.type === "success"
              ? "bg-emerald-500/10 border-emerald-500/30 text-emerald-400"
              : "bg-red-500/10 border-red-500/30 text-red-400"
          }`}>
            <div className="flex items-center gap-2">
              {message.type === "success" ? <CheckCircle size={20} /> : <AlertCircle size={20} />}
              {message.text}
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6">
          <div className="flex items-center gap-4 mb-4">
            <Filter size={20} className="text-slate-400" />
            <h3 className="text-white font-semibold">Filters</h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-2">Role</label>
              <select
                value={selectedRole}
                onChange={(e) => setSelectedRole(e.target.value)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="">All Roles</option>
                {roles.map(role => (
                  <option key={role} value={role}>{formatRoleName(role)}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-2">Resource</label>
              <select
                value={selectedResource}
                onChange={(e) => setSelectedResource(e.target.value)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="">All Resources</option>
                {resources.map(resource => (
                  <option key={resource} value={resource}>{formatResourceName(resource)}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Permission Matrix */}
        <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden">
          <div className="p-6 border-b border-slate-700/50">
            <h3 className="text-lg font-semibold text-white">Permission Matrix</h3>
            <p className="text-sm text-slate-400 mt-1">
              Click checkboxes to toggle permissions. Changes are saved immediately.
            </p>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-900/50">
                <tr>
                  <th className="sticky left-0 z-10 bg-slate-900 px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider border-r border-slate-700/50">
                    Role / Resource
                  </th>
                  {filteredResources.map(resource => (
                    <th key={resource} className="px-4 py-4 text-center">
                      <div className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                        {formatResourceName(resource)}
                      </div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700/50">
                {filteredRoles.map(role => (
                  <tr key={role} className="hover:bg-slate-800/30 transition-colors">
                    <td className="sticky left-0 z-10 bg-slate-900 px-6 py-4 border-r border-slate-700/50">
                      <span className={`inline-flex px-3 py-1.5 text-xs font-medium rounded-full border ${getRoleBadgeColor(role)}`}>
                        {formatRoleName(role)}
                      </span>
                    </td>
                    {filteredResources.map(resource => {
                      const perm = getPermissionForRoleResource(role, resource)
                      if (!perm) return <td key={resource} className="px-4 py-4 text-center text-slate-500">-</td>

                      return (
                        <td key={resource} className="px-4 py-4">
                          <div className="flex flex-col gap-1 items-center">
                            {/* View */}
                            <label className="flex items-center gap-1.5 cursor-pointer group">
                              <input
                                type="checkbox"
                                checked={getPermissionValue(perm, "can_view")}
                                onChange={() => togglePermission(role, resource, "can_view")}
                                disabled={saving}
                                className="w-4 h-4 rounded border-slate-600 text-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-offset-0 bg-slate-800"
                              />
                              <Eye size={12} className="text-slate-400 group-hover:text-blue-400 transition-colors" />
                            </label>
                            {/* Create */}
                            <label className="flex items-center gap-1.5 cursor-pointer group">
                              <input
                                type="checkbox"
                                checked={getPermissionValue(perm, "can_create")}
                                onChange={() => togglePermission(role, resource, "can_create")}
                                disabled={saving}
                                className="w-4 h-4 rounded border-slate-600 text-emerald-500 focus:ring-2 focus:ring-emerald-500 focus:ring-offset-0 bg-slate-800"
                              />
                              <Plus size={12} className="text-slate-400 group-hover:text-emerald-400 transition-colors" />
                            </label>
                            {/* Edit */}
                            <label className="flex items-center gap-1.5 cursor-pointer group">
                              <input
                                type="checkbox"
                                checked={getPermissionValue(perm, "can_edit")}
                                onChange={() => togglePermission(role, resource, "can_edit")}
                                disabled={saving}
                                className="w-4 h-4 rounded border-slate-600 text-amber-500 focus:ring-2 focus:ring-amber-500 focus:ring-offset-0 bg-slate-800"
                              />
                              <Edit size={12} className="text-slate-400 group-hover:text-amber-400 transition-colors" />
                            </label>
                            {/* Delete */}
                            <label className="flex items-center gap-1.5 cursor-pointer group">
                              <input
                                type="checkbox"
                                checked={getPermissionValue(perm, "can_delete")}
                                onChange={() => togglePermission(role, resource, "can_delete")}
                                disabled={saving}
                                className="w-4 h-4 rounded border-slate-600 text-red-500 focus:ring-2 focus:ring-red-500 focus:ring-offset-0 bg-slate-800"
                              />
                              <Trash2 size={12} className="text-slate-400 group-hover:text-red-400 transition-colors" />
                            </label>
                            {/* Export */}
                            <label className="flex items-center gap-1.5 cursor-pointer group">
                              <input
                                type="checkbox"
                                checked={getPermissionValue(perm, "can_export")}
                                onChange={() => togglePermission(role, resource, "can_export")}
                                disabled={saving}
                                className="w-4 h-4 rounded border-slate-600 text-purple-500 focus:ring-2 focus:ring-purple-500 focus:ring-offset-0 bg-slate-800"
                              />
                              <Download size={12} className="text-slate-400 group-hover:text-purple-400 transition-colors" />
                            </label>
                            {/* Approve */}
                            <label className="flex items-center gap-1.5 cursor-pointer group">
                              <input
                                type="checkbox"
                                checked={getPermissionValue(perm, "can_approve")}
                                onChange={() => togglePermission(role, resource, "can_approve")}
                                disabled={saving}
                                className="w-4 h-4 rounded border-slate-600 text-cyan-500 focus:ring-2 focus:ring-cyan-500 focus:ring-offset-0 bg-slate-800"
                              />
                              <CheckCircle size={12} className="text-slate-400 group-hover:text-cyan-400 transition-colors" />
                            </label>
                          </div>
                        </td>
                      )
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Legend */}
          <div className="p-6 border-t border-slate-700/50 bg-slate-900/30">
            <h4 className="text-sm font-semibold text-slate-300 mb-3">Permission Types</h4>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
              <div className="flex items-center gap-2">
                <Eye size={16} className="text-blue-400" />
                <span className="text-xs text-slate-400">View</span>
              </div>
              <div className="flex items-center gap-2">
                <Plus size={16} className="text-emerald-400" />
                <span className="text-xs text-slate-400">Create</span>
              </div>
              <div className="flex items-center gap-2">
                <Edit size={16} className="text-amber-400" />
                <span className="text-xs text-slate-400">Edit</span>
              </div>
              <div className="flex items-center gap-2">
                <Trash2 size={16} className="text-red-400" />
                <span className="text-xs text-slate-400">Delete</span>
              </div>
              <div className="flex items-center gap-2">
                <Download size={16} className="text-purple-400" />
                <span className="text-xs text-slate-400">Export</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle size={16} className="text-cyan-400" />
                <span className="text-xs text-slate-400">Approve</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  )
}
