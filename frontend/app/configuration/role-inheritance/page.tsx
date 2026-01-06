/**
 * @file app/configuration/role-inheritance/page.tsx
 * @description Role inheritance configuration for permission management
 *
 * USER PERSPECTIVE:
 *   - Configure which roles inherit permissions from other roles
 *   - Example: admin inherits from hr, manager, payroll (has all their permissions)
 *   - Visualize role hierarchy tree
 *   - Test resolved permissions for each role
 *   - Enable/disable inheritances without deletion
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: UI layout, visualization style
 *   CAUTION: Role names must match backend enum values
 *   DO NOT modify: Circular dependency prevention logic
 *
 * KEY FEATURES:
 *   - CRUD operations for role inheritances
 *   - Hierarchy visualization
 *   - Circular dependency detection
 *   - Role permission level display
 *   - Active/inactive toggle
 *
 * API ENDPOINTS USED:
 *   - GET /api/role-inheritance (via roleInheritanceApi.getAllInheritances)
 *   - POST /api/role-inheritance (via roleInheritanceApi.createInheritance)
 *   - PUT /api/role-inheritance/:id (via roleInheritanceApi.updateInheritance)
 *   - DELETE /api/role-inheritance/:id (via roleInheritanceApi.deleteInheritance)
 *   - GET /api/role-inheritance/hierarchy (via roleInheritanceApi.getHierarchy)
 *   - GET /api/role-inheritance/resolve/:role (via roleInheritanceApi.resolveRoles)
 *   - GET /api/role-inheritance/valid-roles (via roleInheritanceApi.getValidRoles)
 */

"use client"

import { useEffect, useState } from "react"
import {
  Shield, Plus, Edit, Trash2, AlertCircle, CheckCircle, X,
  GitBranch, Eye, ToggleLeft, ToggleRight
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import {
  roleInheritanceApi, RoleInheritance, ValidRole,
  CreateRoleInheritanceRequest, ApiError
} from "@/lib/api-client"

export default function RoleInheritancePage() {
  const [inheritances, setInheritances] = useState<RoleInheritance[]>([])
  const [validRoles, setValidRoles] = useState<ValidRole[]>([])
  const [hierarchy, setHierarchy] = useState<Record<string, string[]>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showResolveModal, setShowResolveModal] = useState(false)
  const [resolvedRole, setResolvedRole] = useState<string>("")
  const [resolvedRoles, setResolvedRoles] = useState<string[]>([])

  // Form state
  const [formData, setFormData] = useState<CreateRoleInheritanceRequest>({
    child_role: "",
    parent_role: "",
    priority: 5,
    notes: "",
  })

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    try {
      setLoading(true)
      setError(null)
      const [inheritancesData, rolesData, hierarchyData] = await Promise.all([
        roleInheritanceApi.getAllInheritances(),
        roleInheritanceApi.getValidRoles(),
        roleInheritanceApi.getHierarchy(),
      ])
      setInheritances(Array.isArray(inheritancesData) ? inheritancesData : [])
      setValidRoles(rolesData.roles || [])
      setHierarchy(hierarchyData || {})
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load role inheritances")
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate() {
    try {
      setError(null)
      await roleInheritanceApi.createInheritance(formData)
      setShowCreateModal(false)
      resetForm()
      await loadData()
      alert("Herencia creada exitosamente")
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to create inheritance")
      }
    }
  }

  async function handleToggleActive(inheritance: RoleInheritance) {
    try {
      setError(null)
      await roleInheritanceApi.updateInheritance(inheritance.id, {
        is_active: !inheritance.is_active,
        priority: inheritance.priority,
        notes: inheritance.notes || "",
      })
      await loadData()
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to update inheritance")
      }
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("¿Está seguro que desea eliminar esta herencia?")) return

    try {
      setError(null)
      await roleInheritanceApi.deleteInheritance(id)
      await loadData()
      alert("Herencia eliminada exitosamente")
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to delete inheritance")
      }
    }
  }

  async function handleResolveRoles(role: string) {
    try {
      setError(null)
      const result = await roleInheritanceApi.resolveRoles(role)
      setResolvedRole(role)
      setResolvedRoles(result.inherited_roles)
      setShowResolveModal(true)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to resolve roles")
      }
    }
  }

  function resetForm() {
    setFormData({
      child_role: "",
      parent_role: "",
      priority: 5,
      notes: "",
    })
  }

  function getRoleLabel(roleName: string): string {
    const labels: Record<string, string> = {
      admin: "Admin",
      hr: "HR",
      hr_and_pr: "HR & Payroll",
      hr_blue_gray: "HR Blue/Gray",
      hr_white: "HR White",
      manager: "Manager",
      gm: "General Manager",
      supervisor: "Supervisor",
      payroll: "Payroll",
      accountant: "Accountant",
      employee: "Employee",
    }
    return labels[roleName] || roleName
  }

  function getRoleBadge(roleName: string, level: number) {
    const colors: Record<number, string> = {
      10: "bg-red-100 text-red-700",      // admin
      7: "bg-purple-100 text-purple-700", // gm
      6: "bg-indigo-100 text-indigo-700", // hr_and_pr
      5: "bg-blue-100 text-blue-700",     // payroll, hr, manager
      4: "bg-cyan-100 text-cyan-700",     // accountant
      3: "bg-green-100 text-green-700",   // hr sub-roles
      2: "bg-yellow-100 text-yellow-700", // supervisor
      1: "bg-gray-100 text-gray-700",     // employee
    }
    const colorClass = colors[level] || "bg-gray-100 text-gray-700"
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded ${colorClass}`}>
        {getRoleLabel(roleName)} (L{level})
      </span>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-indigo-50 rounded-lg">
              <Shield className="w-6 h-6 text-indigo-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Configuración de Herencia de Roles</h1>
              <p className="text-sm text-gray-500">Los roles heredan permisos de roles padre</p>
            </div>
          </div>
          <Button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Nueva Herencia
          </Button>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-red-900">Error</h3>
              <p className="text-sm text-red-700">{error}</p>
            </div>
          </div>
        )}

        {/* Info Panel */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <GitBranch className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-blue-900 mb-2">¿Cómo funciona la herencia de roles?</h3>
              <ul className="text-sm text-blue-700 space-y-1">
                <li>• Un rol hijo <strong>hereda todos los permisos</strong> de sus roles padre</li>
                <li>• Ejemplo: Si <strong>admin</strong> hereda de <strong>hr</strong>, los admins pueden hacer todo lo que HR puede hacer</li>
                <li>• La herencia es <strong>transitiva</strong>: si A hereda de B, y B de C, entonces A hereda de C</li>
                <li>• El sistema <strong>previene dependencias circulares</strong> automáticamente</li>
                <li>• Nivel de permiso (L1-L10): Indica jerarquía sugerida, mayor nivel = más permisos</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Inheritances Table */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Rol Hijo (hereda)
                  </th>
                  <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    →
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Rol Padre (de)
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Prioridad
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Estado
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Notas
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Acciones
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {loading ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-4 text-center text-gray-500">
                      Cargando...
                    </td>
                  </tr>
                ) : inheritances.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-4 text-center text-gray-500">
                      No hay herencias configuradas
                    </td>
                  </tr>
                ) : (
                  inheritances.map((inheritance) => {
                    const childRole = validRoles.find(r => r.name === inheritance.child_role)
                    const parentRole = validRoles.find(r => r.name === inheritance.parent_role)
                    return (
                      <tr key={inheritance.id} className={`hover:bg-gray-50 ${!inheritance.is_active ? 'opacity-50' : ''}`}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {getRoleBadge(inheritance.child_role, childRole?.level || 0)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center">
                          <GitBranch className="w-4 h-4 text-gray-400 mx-auto" />
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {getRoleBadge(inheritance.parent_role, parentRole?.level || 0)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className="text-sm text-gray-900">{inheritance.priority}</span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <button
                            onClick={() => handleToggleActive(inheritance)}
                            className="flex items-center gap-1"
                          >
                            {inheritance.is_active ? (
                              <>
                                <ToggleRight className="w-5 h-5 text-green-600" />
                                <span className="text-xs font-medium text-green-700">Activo</span>
                              </>
                            ) : (
                              <>
                                <ToggleLeft className="w-5 h-5 text-gray-400" />
                                <span className="text-xs font-medium text-gray-500">Inactivo</span>
                              </>
                            )}
                          </button>
                        </td>
                        <td className="px-6 py-4">
                          <span className="text-sm text-gray-600">{inheritance.notes || "-"}</span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <button
                            onClick={() => handleDelete(inheritance.id)}
                            className="text-red-600 hover:text-red-900"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Role Permission Resolver */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Eye className="w-5 h-5" />
            Probar Resolución de Permisos
          </h3>
          <p className="text-sm text-gray-600 mb-4">
            Selecciona un rol para ver todos los roles que hereda (directo e indirecto)
          </p>
          <div className="flex gap-2">
            <select
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              onChange={(e) => e.target.value && handleResolveRoles(e.target.value)}
            >
              <option value="">-- Selecciona un rol --</option>
              {validRoles.sort((a, b) => b.level - a.level).map((role) => (
                <option key={role.name} value={role.name}>
                  {getRoleLabel(role.name)} (Nivel {role.level})
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Create Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full">
              <div className="p-6">
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-xl font-bold text-gray-900">Crear Nueva Herencia</h2>
                  <button
                    onClick={() => {
                      setShowCreateModal(false)
                      resetForm()
                    }}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <X className="w-6 h-6" />
                  </button>
                </div>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Rol Hijo (que heredará permisos) *
                    </label>
                    <select
                      value={formData.child_role}
                      onChange={(e) => setFormData({ ...formData, child_role: e.target.value })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                      required
                    >
                      <option value="">-- Seleccione un rol --</option>
                      {validRoles.sort((a, b) => b.level - a.level).map((role) => (
                        <option key={role.name} value={role.name}>
                          {getRoleLabel(role.name)} (Nivel {role.level})
                        </option>
                      ))}
                    </select>
                    <p className="text-xs text-gray-500 mt-1">Este rol heredará los permisos del rol padre</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Rol Padre (del que heredar) *
                    </label>
                    <select
                      value={formData.parent_role}
                      onChange={(e) => setFormData({ ...formData, parent_role: e.target.value })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                      required
                    >
                      <option value="">-- Seleccione un rol --</option>
                      {validRoles.sort((a, b) => b.level - a.level).map((role) => (
                        <option key={role.name} value={role.name}>
                          {getRoleLabel(role.name)} (Nivel {role.level})
                        </option>
                      ))}
                    </select>
                    <p className="text-xs text-gray-500 mt-1">Los permisos de este rol serán heredados</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Prioridad (1-10)
                    </label>
                    <input
                      type="number"
                      min="1"
                      max="10"
                      value={formData.priority}
                      onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 5 })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    />
                    <p className="text-xs text-gray-500 mt-1">Mayor prioridad = se verifica primero (para resolución de conflictos)</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Notas
                    </label>
                    <textarea
                      value={formData.notes}
                      onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                      rows={3}
                      placeholder="Notas sobre esta herencia (opcional)"
                    />
                  </div>
                </div>

                <div className="flex items-center justify-end gap-3 mt-6 pt-6 border-t border-gray-200">
                  <Button
                    onClick={() => {
                      setShowCreateModal(false)
                      resetForm()
                    }}
                    className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg"
                  >
                    Cancelar
                  </Button>
                  <Button
                    onClick={handleCreate}
                    disabled={!formData.child_role || !formData.parent_role}
                    className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <CheckCircle className="w-4 h-4" />
                    Crear Herencia
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Resolve Modal */}
        {showResolveModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full">
              <div className="p-6">
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-xl font-bold text-gray-900">Permisos Resueltos</h2>
                  <button
                    onClick={() => setShowResolveModal(false)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <X className="w-6 h-6" />
                  </button>
                </div>

                <div className="space-y-4">
                  <div>
                    <p className="text-sm text-gray-600 mb-3">
                      Un usuario con el rol <strong className="text-indigo-600">{getRoleLabel(resolvedRole)}</strong> tiene efectivamente los siguientes roles (directo + heredado):
                    </p>
                    <div className="flex flex-wrap gap-2">
                      {resolvedRoles.map((role) => {
                        const roleData = validRoles.find(r => r.name === role)
                        return getRoleBadge(role, roleData?.level || 0)
                      })}
                    </div>
                  </div>

                  {resolvedRoles.length > 1 && (
                    <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                      <p className="text-sm text-green-700">
                        ✓ Este rol hereda permisos de {resolvedRoles.length - 1} rol(es) adicional(es)
                      </p>
                    </div>
                  )}

                  {resolvedRoles.length === 1 && (
                    <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                      <p className="text-sm text-yellow-700">
                        ⚠ Este rol no hereda de ningún otro rol (solo tiene sus permisos propios)
                      </p>
                    </div>
                  )}
                </div>

                <div className="flex justify-end mt-6 pt-6 border-t border-gray-200">
                  <Button
                    onClick={() => setShowResolveModal(false)}
                    className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg"
                  >
                    Cerrar
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}
