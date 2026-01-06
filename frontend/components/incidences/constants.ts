/**
 * Shared constants for incidence management components
 */

export const STATUS_BADGES = {
  pending: { label: "Pending", color: "bg-yellow-100 text-yellow-800" },
  approved: { label: "Approved", color: "bg-green-100 text-green-800" },
  rejected: { label: "Rejected", color: "bg-red-100 text-red-800" },
  processed: { label: "Processed", color: "bg-blue-100 text-blue-800" },
} as const

export type SelectionMode = "single" | "multiple" | "all"

export const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("es-MX", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  })
}

export const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + " B"
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB"
  return (bytes / 1024 / 1024).toFixed(1) + " MB"
}

export const getStatusBadge = (status: string) => {
  const badge = STATUS_BADGES[status as keyof typeof STATUS_BADGES]
  if (!badge) return status
  return badge
}
