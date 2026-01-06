"use client"

import { useState } from "react"
import { Calculator, TrendingUp, TrendingDown, DollarSign } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Slider } from "@/components/ui/slider"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

interface PayrollCalculation {
  grossSalary: number
  workedDays: number
  overtimeHours: number
  overtimeRate: number
  bonuses: number
  deductions: number
  // Mexican specific
  imssEmployee: number
  imssEmployer: number
  isr: number
  infonavit: number
  netPay: number
  totalEmployerCost: number
}

export function PayrollCalculator() {
  const [calculation, setCalculation] = useState<PayrollCalculation>({
    grossSalary: 10000,
    workedDays: 30,
    overtimeHours: 0,
    overtimeRate: 2,
    bonuses: 0,
    deductions: 0,
    imssEmployee: 0,
    imssEmployer: 0,
    isr: 0,
    infonavit: 0,
    netPay: 0,
    totalEmployerCost: 0,
  })

  const calculatePayroll = () => {
    // Simplified Mexican payroll calculation
    const dailySalary = calculation.grossSalary / 30
    
    // IMSS Calculations (simplified)
    const imssEmployeeRate = 0.025 // 2.5%
    const imssEmployerRate = 0.07  // 7%
    
    const imssEmployee = calculation.grossSalary * imssEmployeeRate
    const imssEmployer = calculation.grossSalary * imssEmployerRate
    
    // ISR Calculation (simplified progressive tax)
    let isr = 0
    if (calculation.grossSalary <= 10000) {
      isr = calculation.grossSalary * 0.02
    } else if (calculation.grossSalary <= 20000) {
      isr = 200 + (calculation.grossSalary - 10000) * 0.10
    } else if (calculation.grossSalary <= 30000) {
      isr = 1200 + (calculation.grossSalary - 20000) * 0.20
    } else {
      isr = 3200 + (calculation.grossSalary - 30000) * 0.30
    }
    
    // Overtime
    const overtimePay = calculation.overtimeHours * (dailySalary / 8) * calculation.overtimeRate
    
    // Infonavit (simplified)
    const infonavit = calculation.grossSalary * 0.05
    
    // Net Pay
    const totalDeductions = imssEmployee + isr + infonavit + calculation.deductions
    const totalIncome = calculation.grossSalary + overtimePay + calculation.bonuses
    const netPay = totalIncome - totalDeductions
    
    // Total Employer Cost
    const totalEmployerCost = calculation.grossSalary + imssEmployer + calculation.bonuses

    setCalculation({
      ...calculation,
      imssEmployee,
      imssEmployer,
      isr,
      infonavit,
      netPay,
      totalEmployerCost,
    })
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('es-MX', {
      style: 'currency',
      currency: 'MXN',
      minimumFractionDigits: 2
    }).format(amount)
  }

  return (
    <div className="space-y-6">
      <Card className="border-slate-700 bg-slate-900">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-white">
            <Calculator className="text-primary" size={24} />
            Calculadora de Nómina Mexicana
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Input Section */}
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-slate-300">Parámetros de Cálculo</h3>
              
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label className="text-slate-400">Salario Mensual (MXN)</Label>
                  <Input
                    type="number"
                    value={calculation.grossSalary}
                    onChange={(e) => setCalculation({
                      ...calculation,
                      grossSalary: parseFloat(e.target.value) || 0
                    })}
                    className="bg-slate-800 border-slate-700 text-white"
                  />
                </div>
                
                <div className="space-y-2">
                  <Label className="text-slate-400">Días Trabajados</Label>
                  <Slider
                    value={[calculation.workedDays]}
                    onValueChange={([value]) => setCalculation({
                      ...calculation,
                      workedDays: value
                    })}
                    min={1}
                    max={30}
                    step={1}
                    className="[&_[role=slider]]:bg-primary"
                  />
                  <div className="flex justify-between text-sm text-slate-400">
                    <span>1 día</span>
                    <span>{calculation.workedDays} días</span>
                    <span>30 días</span>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-slate-400">Horas Extra</Label>
                    <Input
                      type="number"
                      value={calculation.overtimeHours}
                      onChange={(e) => setCalculation({
                        ...calculation,
                        overtimeHours: parseFloat(e.target.value) || 0
                      })}
                      className="bg-slate-800 border-slate-700 text-white"
                    />
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Bonos (MXN)</Label>
                    <Input
                      type="number"
                      value={calculation.bonuses}
                      onChange={(e) => setCalculation({
                        ...calculation,
                        bonuses: parseFloat(e.target.value) || 0
                      })}
                      className="bg-slate-800 border-slate-700 text-white"
                    />
                  </div>
                </div>
                
                <Button
                  onClick={calculatePayroll}
                  className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                >
                  Calcular Nómina
                </Button>
              </div>
            </div>
            
            {/* Results Section */}
            <div className="space-y-4">
              <Tabs defaultValue="employee" className="w-full">
                <TabsList className="grid w-full grid-cols-2 bg-slate-800">
                  <TabsTrigger value="employee" className="data-[state=active]:bg-primary">
                    Vista Empleado
                  </TabsTrigger>
                  <TabsTrigger value="employer" className="data-[state=active]:bg-primary">
                    Vista Empresa
                  </TabsTrigger>
                </TabsList>
                
                <TabsContent value="employee" className="space-y-4">
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-3 bg-green-900/20 border border-green-800 rounded-lg">
                      <div className="flex items-center gap-2">
                        <TrendingUp className="text-green-500" size={20} />
                        <span className="text-green-400">Ingresos Totales</span>
                      </div>
                      <span className="text-xl font-bold text-white">
                        {formatCurrency(calculation.grossSalary + calculation.bonuses)}
                      </span>
                    </div>
                    
                    <div className="space-y-2">
                      <h4 className="text-sm font-medium text-slate-400">Deducciones</h4>
                      
                      <div className="space-y-1">
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">IMSS Empleado</span>
                          <span className="text-red-400">{formatCurrency(calculation.imssEmployee)}</span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">ISR (Impuesto)</span>
                          <span className="text-red-400">{formatCurrency(calculation.isr)}</span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">Infonavit</span>
                          <span className="text-red-400">{formatCurrency(calculation.infonavit)}</span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">Otras Deducciones</span>
                          <span className="text-red-400">{formatCurrency(calculation.deductions)}</span>
                        </div>
                      </div>
                      
                      <div className="pt-2 border-t border-slate-700">
                        <div className="flex items-center justify-between p-3 bg-blue-900/20 border border-blue-800 rounded-lg">
                          <div className="flex items-center gap-2">
                            <DollarSign className="text-blue-500" size={20} />
                            <span className="text-blue-400 font-semibold">Neto a Pagar</span>
                          </div>
                          <span className="text-2xl font-bold text-white">
                            {formatCurrency(calculation.netPay)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </TabsContent>
                
                <TabsContent value="employer" className="space-y-4">
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-3 bg-amber-900/20 border border-amber-800 rounded-lg">
                      <div className="flex items-center gap-2">
                        <TrendingDown className="text-amber-500" size={20} />
                        <span className="text-amber-400">Costo para la Empresa</span>
                      </div>
                      <span className="text-xl font-bold text-white">
                        {formatCurrency(calculation.totalEmployerCost)}
                      </span>
                    </div>
                    
                    <div className="space-y-2">
                      <h4 className="text-sm font-medium text-slate-400">Desglose de Costos</h4>
                      
                      <div className="space-y-1">
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">Salario Base</span>
                          <span className="text-slate-300">{formatCurrency(calculation.grossSalary)}</span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">IMSS Patronal</span>
                          <span className="text-amber-400">{formatCurrency(calculation.imssEmployer)}</span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">Bonos</span>
                          <span className="text-slate-300">{formatCurrency(calculation.bonuses)}</span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">Aguinaldo (Proporcional)</span>
                          <span className="text-amber-400">
                            {formatCurrency(calculation.grossSalary / 12)}
                          </span>
                        </div>
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-400">Prima Vacacional (Proporcional)</span>
                          <span className="text-amber-400">
                            {formatCurrency(calculation.grossSalary * 0.25 / 12)}
                          </span>
                        </div>
                      </div>
                      
                      <div className="pt-2 border-t border-slate-700">
                        <div className="flex items-center justify-between p-3 bg-purple-900/20 border border-purple-800 rounded-lg">
                          <div className="flex items-center gap-2">
                            <Calculator className="text-purple-500" size={20} />
                            <span className="text-purple-400 font-semibold">Costo Total Mensual</span>
                          </div>
                          <span className="text-2xl font-bold text-white">
                            {formatCurrency(calculation.totalEmployerCost + 
                              (calculation.grossSalary / 12) + 
                              (calculation.grossSalary * 0.25 / 12))}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </TabsContent>
              </Tabs>
              
              {/* Summary Stats */}
              <div className="grid grid-cols-2 gap-3 pt-4 border-t border-slate-700">
                <div className="text-center p-3 bg-slate-800 rounded-lg">
                  <div className="text-sm text-slate-400">Salario Diario</div>
                  <div className="text-lg font-semibold text-white">
                    {formatCurrency(calculation.grossSalary / 30)}
                  </div>
                </div>
                
                <div className="text-center p-3 bg-slate-800 rounded-lg">
                  <div className="text-sm text-slate-400">SDI</div>
                  <div className="text-lg font-semibold text-white">
                    {formatCurrency((calculation.grossSalary / 30) * 1.045)}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
