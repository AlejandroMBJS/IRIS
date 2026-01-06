"use client"

import { Incidence } from "@/lib/api-client"

interface IncidenceStatsProps {
  incidences: Incidence[]
}

export function IncidenceStats({ incidences }: IncidenceStatsProps) {
  const pendingCount = incidences.filter(i => i.status === "pending").length
  const approvedCount = incidences.filter(i => i.status === "approved").length
  const rejectedCount = incidences.filter(i => i.status === "rejected").length

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="text-slate-400 text-sm">Total Incidences</div>
        <div className="text-2xl font-bold text-white">{incidences.length}</div>
      </div>
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="text-slate-400 text-sm">Pending</div>
        <div className="text-2xl font-bold text-yellow-400">{pendingCount}</div>
      </div>
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="text-slate-400 text-sm">Approved</div>
        <div className="text-2xl font-bold text-green-400">{approvedCount}</div>
      </div>
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <div className="text-slate-400 text-sm">Rejected</div>
        <div className="text-2xl font-bold text-red-400">{rejectedCount}</div>
      </div>
    </div>
  )
}
