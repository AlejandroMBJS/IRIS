/**
 * @file app/configuration/incidence-mapping/page.tsx
 * @description Incidence tipo mapping configuration for payroll export
 *
 * USER PERSPECTIVE:
 *   - Configure which Excel template each incidence type maps to
 *   - Set Tipo codes (2-9) for payroll system import
 *   - Define Motivo values (EXTRAS, FALTA, PERHORAS)
 *   - Configure hours multipliers for day-to-hour conversion
 *   - View unmapped incidence types that need configuration
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: UI layout, table columns, form validation
 *   CAUTION: Template type values (must match backend enum)
 *   DO NOT modify: Motivo options without updating backend validation
 *
 * KEY FEATURES:
 *   - CRUD operations for tipo mappings
 *   - Modal-based create/edit forms
 *   - Unmapped incidence types alert
 *   - Template type badges (Vacaciones / Faltas y Extras)
 *
 * API ENDPOINTS USED:
 *   - GET /api/incidence-config/tipo-mappings (via incidenceConfigApi.getAllMappings)
 *   - POST /api/incidence-config/tipo-mappings (via incidenceConfigApi.createMapping)
 *   - PUT /api/incidence-config/tipo-mappings/:id (via incidenceConfigApi.updateMapping)
 *   - DELETE /api/incidence-config/tipo-mappings/:id (via incidenceConfigApi.deleteMapping)
 *   - GET /api/incidence-config/unmapped-types (via incidenceConfigApi.getUnmappedTypes)
 */

"use client"

import { useEffect, useState } from "react"
import {
  Settings, Plus, Edit, Trash2, AlertCircle, FileSpreadsheet,
  CheckCircle, X
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import {
  incidenceConfigApi, IncidenceTipoMapping, IncidenceType,
  CreateTipoMappingRequest, UpdateTipoMappingRequest, ApiError
} from "@/lib/api-client"

export default function IncidenceMappingPage() {
  const [mappings, setMappings] = useState<IncidenceTipoMapping[]>([])
  const [unmappedTypes, setUnmappedTypes] = useState<IncidenceType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingMapping, setEditingMapping] = useState<IncidenceTipoMapping | null>(null)

  // Form state
  const [formData, setFormData] = useState<CreateTipoMappingRequest>({
    incidence_type_id: "",
    tipo_code: "",
    motivo: null,
    template_type: "vacaciones",
    hours_multiplier: 8.0,
    notes: "",
  })

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    try {
      setLoading(true)
      setError(null)
      const [mappingsData, unmappedData] = await Promise.all([
        incidenceConfigApi.getAllMappings(),
        incidenceConfigApi.getUnmappedTypes(),
      ])
      setMappings(Array.isArray(mappingsData) ? mappingsData : [])
      setUnmappedTypes(Array.isArray(unmappedData) ? unmappedData : [])
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to load mappings")
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate() {
    try {
      setError(null)
      // Clean empty strings to null
      const cleanedData = {
        ...formData,
        tipo_code: formData.tipo_code?.trim() || null,
        notes: formData.notes?.trim() || null,
      }
      await incidenceConfigApi.createMapping(cleanedData)
      setShowCreateModal(false)
      resetForm()
      await loadData()
      alert("Mapping created successfully")
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to create mapping")
      }
    }
  }

  async function handleUpdate() {
    if (!editingMapping) return

    try {
      setError(null)
      const updates: UpdateTipoMappingRequest = {
        tipo_code: formData.tipo_code?.trim() || null,
        motivo: formData.motivo,
        template_type: formData.template_type,
        hours_multiplier: formData.hours_multiplier,
        notes: formData.notes?.trim() || null,
      }
      await incidenceConfigApi.updateMapping(editingMapping.id, updates)
      setShowEditModal(false)
      setEditingMapping(null)
      resetForm()
      await loadData()
      alert("Mapping updated successfully")
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to update mapping")
      }
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Are you sure you want to delete this mapping?")) return

    try {
      setError(null)
      await incidenceConfigApi.deleteMapping(id)
      await loadData()
      alert("Mapping deleted successfully")
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError("Failed to delete mapping")
      }
    }
  }

  function openCreateModal() {
    resetForm()
    setShowCreateModal(true)
  }

  function openEditModal(mapping: IncidenceTipoMapping) {
    setEditingMapping(mapping)
    setFormData({
      incidence_type_id: mapping.incidence_type_id,
      tipo_code: mapping.tipo_code || "",
      motivo: mapping.motivo as any,
      template_type: mapping.template_type,
      hours_multiplier: mapping.hours_multiplier,
      notes: mapping.notes || "",
    })
    setShowEditModal(true)
  }

  function resetForm() {
    setFormData({
      incidence_type_id: "",
      tipo_code: "",
      motivo: null,
      template_type: "vacaciones",
      hours_multiplier: 8.0,
      notes: "",
    })
  }

  function getTemplateBadge(type: string) {
    if (type === "vacaciones") {
      return (
        <span className="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-700 rounded">
          Vacation
        </span>
      )
    }
    return (
      <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded">
        Absences and Extras
      </span>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-purple-50 rounded-lg">
              <Settings className="w-6 h-6 text-purple-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Incidence Type Mappings</h1>
              <p className="text-sm text-gray-500">Configure Type codes and Excel templates</p>
            </div>
          </div>
          <Button
            onClick={openCreateModal}
            className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            New Mapping
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

        {/* Unmapped Types Warning */}
        {unmappedTypes.length > 0 && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <h3 className="font-semibold text-yellow-900 mb-2">Unmapped Types</h3>
                <p className="text-sm text-yellow-700 mb-3">
                  The following incidence types have no configured mapping and will not appear in the export:
                </p>
                <ul className="list-disc list-inside text-sm text-yellow-700 space-y-1">
                  {unmappedTypes.map((type) => (
                    <li key={type.id}>{type.name}</li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        )}

        {/* Mappings Table */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Incidence Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Template
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Type Code
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Reason
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Multiplier
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {loading ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-4 text-center text-gray-500">
                      Loading...
                    </td>
                  </tr>
                ) : mappings.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-4 text-center text-gray-500">
                      No mappings configured
                    </td>
                  </tr>
                ) : (
                  mappings.map((mapping) => (
                    <tr key={mapping.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {mapping.incidence_type?.name || "No name"}
                        </div>
                        {mapping.incidence_type?.incidence_category && (
                          <div className="text-xs text-gray-500">
                            {mapping.incidence_type.incidence_category.name}
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getTemplateBadge(mapping.template_type)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {mapping.tipo_code || "-"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {mapping.motivo ? (
                          <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-700 rounded">
                            {mapping.motivo}
                          </span>
                        ) : (
                          <span className="text-sm text-gray-400">-</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {mapping.hours_multiplier}h
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => openEditModal(mapping)}
                          className="text-blue-600 hover:text-blue-900 mr-3"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(mapping.id)}
                          className="text-red-600 hover:text-red-900"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Create/Edit Modal */}
        {(showCreateModal || showEditModal) && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
              <div className="p-6">
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-xl font-bold text-gray-900">
                    {showCreateModal ? "Create New Mapping" : "Edit Mapping"}
                  </h2>
                  <button
                    onClick={() => {
                      setShowCreateModal(false)
                      setShowEditModal(false)
                      resetForm()
                    }}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <X className="w-6 h-6" />
                  </button>
                </div>

                <div className="space-y-4">
                  {showCreateModal && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Incidence Type *
                      </label>
                      <select
                        value={formData.incidence_type_id}
                        onChange={(e) => setFormData({ ...formData, incidence_type_id: e.target.value })}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                        required
                      >
                        <option value="">-- Select a type --</option>
                        {unmappedTypes.map((type) => (
                          <option key={type.id} value={type.id}>
                            {type.name} ({type.incidence_category?.name})
                          </option>
                        ))}
                      </select>
                    </div>
                  )}

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Excel Template *
                    </label>
                    <select
                      value={formData.template_type}
                      onChange={(e) => setFormData({ ...formData, template_type: e.target.value as any })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                      required
                    >
                      <option value="vacaciones">Vacation</option>
                      <option value="faltas_extras">Absences and Extras</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Type Code (2-9)
                    </label>
                    <input
                      type="text"
                      value={formData.tipo_code || ""}
                      onChange={(e) => setFormData({ ...formData, tipo_code: e.target.value })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                      placeholder="e.g.: 7"
                    />
                    <p className="text-xs text-gray-500 mt-1">Code used by the payroll system</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Reason
                    </label>
                    <select
                      value={formData.motivo || ""}
                      onChange={(e) => setFormData({ ...formData, motivo: e.target.value as any || null })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                    >
                      <option value="">-- No reason --</option>
                      <option value="EXTRAS">EXTRAS</option>
                      <option value="FALTA">FALTA</option>
                      <option value="PERHORAS">PERHORAS</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Hours Multiplier *
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      value={formData.hours_multiplier}
                      onChange={(e) => setFormData({ ...formData, hours_multiplier: parseFloat(e.target.value) })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                      required
                    />
                    <p className="text-xs text-gray-500 mt-1">Factor to convert days to hours (typically 8.0)</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Notes
                    </label>
                    <textarea
                      value={formData.notes || ""}
                      onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                      rows={3}
                      placeholder="Additional notes (optional)"
                    />
                  </div>
                </div>

                <div className="flex items-center justify-end gap-3 mt-6 pt-6 border-t border-gray-200">
                  <Button
                    onClick={() => {
                      setShowCreateModal(false)
                      setShowEditModal(false)
                      resetForm()
                    }}
                    className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg"
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={showCreateModal ? handleCreate : handleUpdate}
                    className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg flex items-center gap-2"
                  >
                    <CheckCircle className="w-4 h-4" />
                    {showCreateModal ? "Create Mapping" : "Save Changes"}
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
