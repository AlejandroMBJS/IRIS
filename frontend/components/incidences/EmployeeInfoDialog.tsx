"use client"

import { User, Mail, IdCard, Briefcase, DollarSign } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Employee } from "@/lib/api-client"
import { formatDate } from "./constants"

interface EmployeeInfoDialogProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  employee: Employee | null
}

export function EmployeeInfoDialog({
  isOpen,
  onOpenChange,
  employee,
}: EmployeeInfoDialogProps) {
  if (!employee) return null

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Employee Information
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            Complete employee details
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Personal Info Section */}
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <h3 className="text-sm font-medium text-slate-400 mb-3 flex items-center gap-2">
              <User className="h-4 w-4" />
              Personal Information
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-slate-500 text-xs">Full Name</Label>
                <p className="text-white font-medium">
                  {employee.first_name} {employee.last_name}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Employee No.</Label>
                <p className="text-white font-medium">{employee.employee_number}</p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Date of Birth</Label>
                <p className="text-white">
                  {employee.date_of_birth ? formatDate(employee.date_of_birth) : "N/A"}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Gender</Label>
                <p className="text-white capitalize">{employee.gender || "N/A"}</p>
              </div>
            </div>
          </div>

          {/* Contact Info Section */}
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <h3 className="text-sm font-medium text-slate-400 mb-3 flex items-center gap-2">
              <Mail className="h-4 w-4" />
              Contact
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-slate-500 text-xs">Email</Label>
                <p className="text-white">{employee.personal_email || "N/A"}</p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Phone</Label>
                <p className="text-white">{employee.personal_phone || "N/A"}</p>
              </div>
              <div className="col-span-2">
                <Label className="text-slate-500 text-xs">Address</Label>
                <p className="text-white">
                  {employee.street ? `${employee.street} ${employee.exterior_number || ""}` : "N/A"}
                </p>
              </div>
            </div>
          </div>

          {/* Tax/Legal Info Section */}
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <h3 className="text-sm font-medium text-slate-400 mb-3 flex items-center gap-2">
              <IdCard className="h-4 w-4" />
              Tax / IMSS Information
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-slate-500 text-xs">RFC</Label>
                <p className="text-white font-mono">{employee.rfc || "N/A"}</p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">CURP</Label>
                <p className="text-white font-mono">{employee.curp || "N/A"}</p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">NSS (Social Security Number)</Label>
                <p className="text-white font-mono">{employee.nss || "N/A"}</p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Bank Account (CLABE)</Label>
                <p className="text-white font-mono">{employee.bank_account || "N/A"}</p>
              </div>
            </div>
          </div>

          {/* Employment Info Section */}
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <h3 className="text-sm font-medium text-slate-400 mb-3 flex items-center gap-2">
              <Briefcase className="h-4 w-4" />
              Employment Information
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-slate-500 text-xs">Hire Date</Label>
                <p className="text-white">
                  {employee.hire_date ? formatDate(employee.hire_date) : "N/A"}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Employee Type</Label>
                <p className="text-white capitalize">
                  {employee.employee_type?.replace("_", " ") || "N/A"}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Status</Label>
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                  employee.employment_status === "active"
                    ? "bg-green-100 text-green-800"
                    : "bg-red-100 text-red-800"
                }`}>
                  {employee.employment_status === "active" ? "Active" : "Inactive"}
                </span>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Department</Label>
                <p className="text-white">{employee.department_id || "N/A"}</p>
              </div>
            </div>
          </div>

          {/* Salary Info Section */}
          <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
            <h3 className="text-sm font-medium text-slate-400 mb-3 flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              Salary Information
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-slate-500 text-xs">Daily Salary</Label>
                <p className="text-white font-medium text-lg">
                  ${employee.daily_salary?.toFixed(2) || "0.00"}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">SDI (Integrated Daily Salary)</Label>
                <p className="text-white">
                  ${employee.integrated_daily_salary?.toFixed(2) || "0.00"}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Pay Frequency</Label>
                <p className="text-white capitalize">
                  {employee.pay_frequency?.replace("_", " ") || "N/A"}
                </p>
              </div>
              <div>
                <Label className="text-slate-500 text-xs">Payment Method</Label>
                <p className="text-white capitalize">
                  {employee.payment_method?.replace("_", " ") || "N/A"}
                </p>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="border-slate-600"
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
