"use client"

import { Check, X, Paperclip, User } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Incidence, Employee } from "@/lib/api-client"
import { STATUS_BADGES, formatDate } from "./constants"

interface IncidenceTableProps {
  incidences: Incidence[]
  loading: boolean
  onApprove: (id: string) => void
  onReject: (id: string) => void
  onDelete: (id: string) => void
  onOpenEvidence: (incidence: Incidence) => void
  onOpenEmployeeInfo: (employee: Employee) => void
}

export function IncidenceTable({
  incidences,
  loading,
  onApprove,
  onReject,
  onDelete,
  onOpenEvidence,
  onOpenEmployeeInfo,
}: IncidenceTableProps) {
  const getStatusBadge = (status: string) => {
    const badge = STATUS_BADGES[status as keyof typeof STATUS_BADGES]
    if (!badge) return status
    return (
      <span className={`px-2 py-1 rounded-full text-xs font-medium ${badge.color}`}>
        {badge.label}
      </span>
    )
  }

  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow className="border-slate-700 hover:bg-slate-800/50">
            <TableHead className="text-slate-300 hidden lg:table-cell">Period</TableHead>
            <TableHead className="text-slate-300">Employee</TableHead>
            <TableHead className="text-slate-300">Type</TableHead>
            <TableHead className="text-slate-300 hidden md:table-cell">Dates</TableHead>
            <TableHead className="text-slate-300 hidden lg:table-cell">Quantity</TableHead>
            <TableHead className="text-slate-300 hidden xl:table-cell">Amount</TableHead>
            <TableHead className="text-slate-300">Status</TableHead>
            <TableHead className="text-slate-300 text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            <TableRow>
              <TableCell colSpan={8} className="text-center text-slate-400 py-8">
                Loading...
              </TableCell>
            </TableRow>
          ) : incidences.length === 0 ? (
            <TableRow>
              <TableCell colSpan={8} className="text-center text-slate-400 py-8">
                No incidences registered
              </TableCell>
            </TableRow>
          ) : (
            incidences.map((incidence) => (
              <TableRow key={incidence.id} className="border-slate-700 hover:bg-slate-800/30">
                <TableCell className="hidden lg:table-cell">
                  <span className="text-blue-400 text-sm font-medium">
                    {incidence.payroll_period
                      ? `Week ${incidence.payroll_period.period_number}`
                      : "N/A"}
                  </span>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <User className="h-4 w-4 text-slate-400 hidden sm:block" />
                    {incidence.employee ? (
                      <button
                        onClick={() => onOpenEmployeeInfo(incidence.employee!)}
                        className="text-blue-400 hover:text-blue-300 hover:underline cursor-pointer text-left text-sm"
                        title="View employee information"
                      >
                        {incidence.employee.first_name} {incidence.employee.last_name}
                      </button>
                    ) : (
                      <span className="text-white">N/A</span>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-slate-300 text-sm">
                  {incidence.incidence_type?.name || "N/A"}
                </TableCell>
                <TableCell className="text-slate-300 hidden md:table-cell">
                  <div className="text-xs">
                    <div>{formatDate(incidence.start_date)}</div>
                    <div className="text-slate-500">to {formatDate(incidence.end_date)}</div>
                  </div>
                </TableCell>
                <TableCell className="text-slate-300 hidden lg:table-cell">
                  {incidence.quantity}
                </TableCell>
                <TableCell className="text-slate-300 hidden xl:table-cell">
                  {incidence.calculated_amount > 0
                    ? `$${incidence.calculated_amount.toFixed(2)}`
                    : "-"}
                </TableCell>
                <TableCell>{getStatusBadge(incidence.status)}</TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onOpenEvidence(incidence)}
                      className="text-blue-400 hover:text-blue-300"
                      title="View/Attach Evidence"
                    >
                      <Paperclip className="h-4 w-4" />
                    </Button>
                    {incidence.status === "pending" && (
                      <>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onApprove(incidence.id)}
                          className="text-green-400 hover:text-green-300"
                          title="Approve"
                        >
                          <Check className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onReject(incidence.id)}
                          className="text-red-400 hover:text-red-300"
                          title="Reject"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </>
                    )}
                    {incidence.status !== "processed" && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(incidence.id)}
                        className="text-slate-400 hover:text-red-400"
                        title="Delete"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  )
}
