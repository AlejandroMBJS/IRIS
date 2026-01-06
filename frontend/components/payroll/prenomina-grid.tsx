"use client"

import { useState } from "react"
import { Save, Calculator, CheckCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { useToast } from "@/hooks/use-toast"

interface PrenominaEntry {
  employeeId: string
  employeeCode: string
  employeeName: string
  regularDays: number
  overtimeHours: number
  overtimeRate: number
  absenceDays: number
  vacationDays: number
  sickDays: number
  bonuses: number
  deductions: number
  status: "pending" | "calculated" | "approved"
}

export function PrenominaGrid({ periodId }: { periodId: string }) {
  const { toast } = useToast()
  const [entries, setEntries] = useState<PrenominaEntry[]>([
    {
      employeeId: "EMP001",
      employeeCode: "001",
      employeeName: "Juan Pérez López",
      regularDays: 15,
      overtimeHours: 8,
      overtimeRate: 2,
      absenceDays: 0,
      vacationDays: 0,
      sickDays: 0,
      bonuses: 1000,
      deductions: 500,
      status: "pending"
    },
    {
      employeeId: "EMP002",
      employeeCode: "002",
      employeeName: "María García Rodríguez",
      regularDays: 15,
      overtimeHours: 4,
      overtimeRate: 2,
      absenceDays: 1,
      vacationDays: 0,
      sickDays: 0,
      bonuses: 0,
      deductions: 300,
      status: "pending"
    },
    {
      employeeId: "EMP003",
      employeeCode: "003",
      employeeName: "Carlos Hernández Martínez",
      regularDays: 15,
      overtimeHours: 12,
      overtimeRate: 2,
      absenceDays: 0,
      vacationDays: 2,
      sickDays: 0,
      bonuses: 1500,
      deductions: 750,
      status: "calculated"
    }
  ])

  const [saving, setSaving] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      toast({
        title: "Successfully Saved",
        description: "The pre-payroll has been saved successfully",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save the pre-payroll",
        variant: "destructive",
      })
    } finally {
      setSaving(false)
    }
  }

  const handleCalculate = async (employeeId: string) => {
    try {
      // Simulate calculation
      await new Promise(resolve => setTimeout(resolve, 500))
      
      setEntries(prev => prev.map(entry => 
        entry.employeeId === employeeId 
          ? { ...entry, status: "calculated" }
          : entry
      ))
      
      toast({
        title: "Calculation Completed",
        description: `Payroll calculated for ${employeeId}`,
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to calculate the payroll",
        variant: "destructive",
      })
    }
  }

  const handleApprove = async (employeeId: string) => {
    try {
      // Simulate approval
      await new Promise(resolve => setTimeout(resolve, 500))
      
      setEntries(prev => prev.map(entry => 
        entry.employeeId === employeeId 
          ? { ...entry, status: "approved" }
          : entry
      ))
      
      toast({
        title: "Approved",
        description: `Payroll approved for ${employeeId}`,
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to approve the payroll",
        variant: "destructive",
      })
    }
  }

  const handleBulkCalculate = async () => {
    try {
      // Simulate bulk calculation
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      setEntries(prev => prev.map(entry => 
        entry.status === "pending" 
          ? { ...entry, status: "calculated" }
          : entry
      ))
      
      toast({
        title: "Bulk Calculation Completed",
        description: "All pending payrolls have been calculated",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to perform bulk calculation",
        variant: "destructive",
      })
    }
  }

  const handleBulkApprove = async () => {
    try {
      // Simulate bulk approval
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      setEntries(prev => prev.map(entry => 
        entry.status === "calculated" 
          ? { ...entry, status: "approved" }
          : entry
      ))
      
      toast({
        title: "Bulk Approval Completed",
        description: "All calculated payrolls have been approved",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to perform bulk approval",
        variant: "destructive",
      })
    }
  }

  const updateEntry = (employeeId: string, field: keyof PrenominaEntry, value: any) => {
    setEntries(prev => prev.map(entry => 
      entry.employeeId === employeeId 
        ? { ...entry, [field]: value, status: "pending" }
        : entry
    ))
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Pre-Payroll</h1>
          <p className="text-slate-400 mt-1">
            Period: {periodId} • Entry of worked days and overtime hours
          </p>
        </div>
        
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleBulkCalculate}
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            <Calculator className="h-4 w-4 mr-2" />
            Calculate All
          </Button>

          <Button
            variant="outline"
            onClick={handleBulkApprove}
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            <CheckCircle className="h-4 w-4 mr-2" />
            Approve All
          </Button>

          <Button
            onClick={handleSave}
            disabled={saving}
            className="bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700"
          >
            <Save className="h-4 w-4 mr-2" />
            {saving ? "Saving..." : "Save Pre-Payroll"}
          </Button>
        </div>
      </div>

      <Card className="border-slate-700 bg-slate-900">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-white">Pre-Payroll Entry</CardTitle>
            <div className="flex items-center gap-3 text-sm">
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-amber-500" />
                <span className="text-slate-400">Pending</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-blue-500" />
                <span className="text-slate-400">Calculated</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-green-500" />
                <span className="text-slate-400">Approved</span>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border border-slate-700 overflow-hidden">
            <Table>
              <TableHeader className="bg-slate-800">
                <TableRow>
                  <TableHead className="text-slate-300">Code</TableHead>
                  <TableHead className="text-slate-300">Employee</TableHead>
                  <TableHead className="text-slate-300">Regular Days</TableHead>
                  <TableHead className="text-slate-300">Overtime Hours</TableHead>
                  <TableHead className="text-slate-300">O.T. Rate</TableHead>
                  <TableHead className="text-slate-300">Absences</TableHead>
                  <TableHead className="text-slate-300">Vacation</TableHead>
                  <TableHead className="text-slate-300">Sick Days</TableHead>
                  <TableHead className="text-slate-300">Bonuses</TableHead>
                  <TableHead className="text-slate-300">Deductions</TableHead>
                  <TableHead className="text-slate-300">Status</TableHead>
                  <TableHead className="text-slate-300">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {entries.map((entry) => (
                  <TableRow 
                    key={entry.employeeId}
                    className={`
                      ${entry.status === "pending" ? "bg-amber-950/20 hover:bg-amber-950/30" : ""}
                      ${entry.status === "calculated" ? "bg-blue-950/20 hover:bg-blue-950/30" : ""}
                      ${entry.status === "approved" ? "bg-green-950/20 hover:bg-green-950/30" : ""}
                    `}
                  >
                    <TableCell className="font-medium text-white">
                      {entry.employeeCode}
                    </TableCell>
                    <TableCell className="text-white">{entry.employeeName}</TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.regularDays}
                        onChange={(e) => updateEntry(entry.employeeId, "regularDays", parseInt(e.target.value) || 0)}
                        className="w-20 bg-slate-800 border-slate-700 text-white"
                        min="0"
                        max="31"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.overtimeHours}
                        onChange={(e) => updateEntry(entry.employeeId, "overtimeHours", parseInt(e.target.value) || 0)}
                        className="w-20 bg-slate-800 border-slate-700 text-white"
                        min="0"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <select
                        value={entry.overtimeRate}
                        onChange={(e) => updateEntry(entry.employeeId, "overtimeRate", parseFloat(e.target.value))}
                        className="w-20 bg-slate-800 border border-slate-700 rounded-md px-2 py-1 text-white"
                      >
                        <option value="1">1x</option>
                        <option value="2">2x</option>
                        <option value="3">3x</option>
                      </select>
                    </TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.absenceDays}
                        onChange={(e) => updateEntry(entry.employeeId, "absenceDays", parseInt(e.target.value) || 0)}
                        className="w-20 bg-slate-800 border-slate-700 text-white"
                        min="0"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.vacationDays}
                        onChange={(e) => updateEntry(entry.employeeId, "vacationDays", parseInt(e.target.value) || 0)}
                        className="w-20 bg-slate-800 border-slate-700 text-white"
                        min="0"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.sickDays}
                        onChange={(e) => updateEntry(entry.employeeId, "sickDays", parseInt(e.target.value) || 0)}
                        className="w-20 bg-slate-800 border-slate-700 text-white"
                        min="0"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.bonuses}
                        onChange={(e) => updateEntry(entry.employeeId, "bonuses", parseFloat(e.target.value) || 0)}
                        className="w-24 bg-slate-800 border-slate-700 text-white"
                        min="0"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Input
                        type="number"
                        value={entry.deductions}
                        onChange={(e) => updateEntry(entry.employeeId, "deductions", parseFloat(e.target.value) || 0)}
                        className="w-24 bg-slate-800 border-slate-700 text-white"
                        min="0"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Badge 
                        variant={
                          entry.status === "pending" ? "outline" :
                          entry.status === "calculated" ? "secondary" :
                          "default"
                        }
                        className={`
                          ${entry.status === "pending" ? "border-amber-500 text-amber-500" : ""}
                          ${entry.status === "calculated" ? "bg-blue-900/50 text-blue-400" : ""}
                          ${entry.status === "approved" ? "bg-green-900/50 text-green-400" : ""}
                        `}
                      >
                        {entry.status === "pending" ? "Pending" :
                         entry.status === "calculated" ? "Calculated" : "Approved"}
                      </Badge>
                    </TableCell>
                    
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleCalculate(entry.employeeId)}
                          disabled={entry.status === "approved"}
                          className="h-8 px-2 border-slate-700 text-slate-300 hover:bg-slate-800"
                        >
                          <Calculator className="h-3 w-3" />
                        </Button>
                        
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleApprove(entry.employeeId)}
                          disabled={entry.status !== "calculated"}
                          className="h-8 px-2 border-slate-700 text-slate-300 hover:bg-slate-800"
                        >
                          <CheckCircle className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          <div className="mt-6 p-4 bg-slate-800/50 border border-slate-700 rounded-lg">
            <h3 className="text-lg font-semibold text-white mb-2">Period Totals</h3>
            <div className="grid grid-cols-4 gap-4">
              <div className="text-center p-3 bg-slate-900 rounded-lg">
                <div className="text-sm text-slate-400">Total Employees</div>
                <div className="text-2xl font-bold text-white">{entries.length}</div>
              </div>

              <div className="text-center p-3 bg-slate-900 rounded-lg">
                <div className="text-sm text-slate-400">Pending</div>
                <div className="text-2xl font-bold text-amber-400">
                  {entries.filter(e => e.status === "pending").length}
                </div>
              </div>

              <div className="text-center p-3 bg-slate-900 rounded-lg">
                <div className="text-sm text-slate-400">Calculated</div>
                <div className="text-2xl font-bold text-blue-400">
                  {entries.filter(e => e.status === "calculated").length}
                </div>
              </div>

              <div className="text-center p-3 bg-slate-900 rounded-lg">
                <div className="text-sm text-slate-400">Approved</div>
                <div className="text-2xl font-bold text-green-400">
                  {entries.filter(e => e.status === "approved").length}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
