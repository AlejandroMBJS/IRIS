/**
 * @file app/configuration/payroll-setup/page.tsx
 * @description Payroll system configuration for Mexican tax and social security parameters
 *
 * USER PERSPECTIVE:
 *   - Configure IMSS rates (employer and employee contributions)
 *   - Set ISR (income tax) parameters and subsidies
 *   - Define Infonavit contribution rates
 *   - Configure UMA, minimum wage, and other legal values
 *   - Set aguinaldo, vacation, and overtime rates
 *   - View summary of total employer costs
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Default values, input validation, summary calculations
 *   CAUTION: These values directly affect payroll calculations
 *   DO NOT modify: Rate structures without legal/accounting verification
 *
 * KEY COMPONENTS:
 *   - Tabs: IMSS, ISR, Infonavit, Other configurations
 *   - Rate inputs: Percentage and currency fields
 *   - Summary panel: Total employer contribution calculation
 *   - Reset button: Restore default values
 *
 * API ENDPOINTS USED:
 *   - None (currently client-side only)
 *   - TODO: GET /api/configuration/payroll
 *   - TODO: PUT /api/configuration/payroll
 */

"use client"

import { useState } from "react"
import { Save, Calculator, Shield, Building } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useToast } from "@/hooks/use-toast"
import { DashboardLayout } from "@/components/layout/dashboard-layout"

export default function PayrollSetupPage() {
  const { toast } = useToast()
  const [isSaving, setIsSaving] = useState(false)
  
  const [config, setConfig] = useState({
    // IMSS Rates
    imssEmployerRate: 0.0732, // 7.32%
    imssEmployeeRate: 0.0250, // 2.50%
    imssDisabilityRate: 0.0070, // 0.70%
    imssRetirementRate: 0.0110, // 1.10%
    
    // Infonavit
    infonavitRate: 0.05, // 5%
    
    // ISR
    isrSubsidy: 0.0, // Subsidio para el empleo
    isrDailyLimit: 10000, // Límite diario para subsidio
    
    // UMA & Minimum Wage
    umaValue: 108.57,
    minimumWage: 248.93,
    
    // Aguinaldo & Vacations
    aguinaldoDays: 15,
    vacationDaysFirstYear: 6,
    vacationPremium: 0.25, // 25%
    
    // Overtime
    overtimeDoubleRate: 2.0,
    overtimeTripleRate: 3.0,
    
    // Company Contributions
    sarEmployerRate: 0.02, // 2%
    sarEmployeeRate: 0.00,
  })

  const handleSave = async () => {
    setIsSaving(true)
    try {
      // Simulate API save
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      toast({
        title: "Configuración guardada",
        description: "Los cambios se han guardado exitosamente",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "No se pudo guardar la configuración",
        variant: "destructive",
      })
    } finally {
      setIsSaving(false)
    }
  }

  const handleReset = () => {
    // Reset to default values
    setConfig({
      imssEmployerRate: 0.0732,
      imssEmployeeRate: 0.0250,
      imssDisabilityRate: 0.0070,
      imssRetirementRate: 0.0110,
      infonavitRate: 0.05,
      isrSubsidy: 0.0,
      isrDailyLimit: 10000,
      umaValue: 108.57,
      minimumWage: 248.93,
      aguinaldoDays: 15,
      vacationDaysFirstYear: 6,
      vacationPremium: 0.25,
      overtimeDoubleRate: 2.0,
      overtimeTripleRate: 3.0,
      sarEmployerRate: 0.02,
      sarEmployeeRate: 0.00,
    })
    
    toast({
      title: "Configuración restablecida",
      description: "Se han cargado los valores por defecto",
    })
  }

  const updateConfig = (key: string, value: any) => {
    setConfig(prev => ({
      ...prev,
      [key]: parseFloat(value) || value
    }))
  }

  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(2)}%`
  }

  return (
    <DashboardLayout>
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Configuración de Nómina</h1>
          <p className="text-slate-400 mt-1">Configura los parámetros del sistema de nómina mexicano</p>
        </div>
        
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleReset}
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            Restablecer
          </Button>
          <Button
            onClick={handleSave}
            disabled={isSaving}
            className="bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700"
          >
            <Save className="h-4 w-4 mr-2" />
            {isSaving ? "Guardando..." : "Guardar Cambios"}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="imss" className="w-full">
        <TabsList className="grid grid-cols-4 w-full bg-slate-800">
          <TabsTrigger value="imss" className="data-[state=active]:bg-primary">
            <Shield className="h-4 w-4 mr-2" />
            IMSS
          </TabsTrigger>
          <TabsTrigger value="isr" className="data-[state=active]:bg-primary">
            <Calculator className="h-4 w-4 mr-2" />
            ISR
          </TabsTrigger>
          <TabsTrigger value="infonavit" className="data-[state=active]:bg-primary">
            <Building className="h-4 w-4 mr-2" />
            Infonavit
          </TabsTrigger>
          <TabsTrigger value="other" className="data-[state=active]:bg-primary">
            Otros
          </TabsTrigger>
        </TabsList>

        <TabsContent value="imss" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Configuración IMSS</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-slate-300">Cuotas Patronales</h3>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Riesgo de Trabajo</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.0001"
                        value={config.imssEmployerRate}
                        onChange={(e) => updateConfig("imssEmployerRate", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">
                        ({formatPercentage(config.imssEmployerRate)})
                      </span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Invalidez y Vida</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.0001"
                        value={config.imssDisabilityRate}
                        onChange={(e) => updateConfig("imssDisabilityRate", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">
                        ({formatPercentage(config.imssDisabilityRate)})
                      </span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Retiro</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.0001"
                        value={config.imssRetirementRate}
                        onChange={(e) => updateConfig("imssRetirementRate", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">
                        ({formatPercentage(config.imssRetirementRate)})
                      </span>
                    </div>
                  </div>
                </div>
                
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-slate-300">Cuotas Obreras</h3>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Enfermedad y Maternidad</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.0001"
                        value={config.imssEmployeeRate}
                        onChange={(e) => updateConfig("imssEmployeeRate", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">
                        ({formatPercentage(config.imssEmployeeRate)})
                      </span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">SAR (Aportación Patronal)</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.0001"
                        value={config.sarEmployerRate}
                        onChange={(e) => updateConfig("sarEmployerRate", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">
                        ({formatPercentage(config.sarEmployerRate)})
                      </span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">SAR (Aportación Obrera)</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.0001"
                        value={config.sarEmployeeRate}
                        onChange={(e) => updateConfig("sarEmployeeRate", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">
                        ({formatPercentage(config.sarEmployeeRate)})
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="isr" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Configuración ISR</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label className="text-slate-400">Subsidio para el Empleo</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.01"
                        value={config.isrSubsidy}
                        onChange={(e) => updateConfig("isrSubsidy", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">MXN diarios</span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Límite Diario para Subsidio</Label>
                    <div className="flex items-center">
                      <Input
                        type="number"
                        step="0.01"
                        value={config.isrDailyLimit}
                        onChange={(e) => updateConfig("isrDailyLimit", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                      <span className="ml-2 text-slate-300">MXN</span>
                    </div>
                  </div>
                </div>
                
                <div className="p-4 bg-slate-800/50 border border-slate-700 rounded-lg">
                  <h4 className="font-semibold text-slate-300 mb-2">Tablas ISR Vigentes</h4>
                  <div className="text-sm text-slate-400 space-y-1">
                    <p>• Los cálculos ISR utilizan las tablas del ejercicio fiscal actual</p>
                    <p>• Se aplican límites inferiores y superiores según la LISR</p>
                    <p>• Se considera subsidio para el empleo cuando aplica</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="infonavit" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Configuración Infonavit</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label className="text-slate-400">Porcentaje de Aportación</Label>
                  <div className="flex items-center">
                    <Input
                      type="number"
                      step="0.0001"
                      value={config.infonavitRate}
                      onChange={(e) => updateConfig("infonavitRate", e.target.value)}
                      className="bg-slate-800 border-slate-700 text-white"
                    />
                    <span className="ml-2 text-slate-300">
                      ({formatPercentage(config.infonavitRate)} del SDI)
                    </span>
                  </div>
                  <p className="text-sm text-slate-500">
                    Este porcentaje se aplica sobre el Salario Diario Integrado (SDI)
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="other" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Prestaciones de Ley</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label className="text-slate-400">Días de Aguinaldo</Label>
                    <Input
                      type="number"
                      value={config.aguinaldoDays}
                      onChange={(e) => updateConfig("aguinaldoDays", e.target.value)}
                      className="bg-slate-800 border-slate-700 text-white"
                    />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-slate-400">Días de Vacaciones (1er año)</Label>
                      <Input
                        type="number"
                        value={config.vacationDaysFirstYear}
                        onChange={(e) => updateConfig("vacationDaysFirstYear", e.target.value)}
                        className="bg-slate-800 border-slate-700 text-white"
                      />
                    </div>
                    
                    <div className="space-y-2">
                      <Label className="text-slate-400">Prima Vacacional</Label>
                      <div className="flex items-center">
                        <Input
                          type="number"
                          step="0.01"
                          value={config.vacationPremium}
                          onChange={(e) => updateConfig("vacationPremium", e.target.value)}
                          className="bg-slate-800 border-slate-700 text-white"
                        />
                        <span className="ml-2 text-slate-300">
                          ({formatPercentage(config.vacationPremium)})
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            <Card className="border-slate-700 bg-slate-900">
              <CardHeader>
                <CardTitle className="text-white">Valores de Referencia</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label className="text-slate-400">Valor UMA (MXN)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={config.umaValue}
                      onChange={(e) => updateConfig("umaValue", e.target.value)}
                      className="bg-slate-800 border-slate-700 text-white"
                    />
                  </div>
                  
                  <div className="space-y-2">
                    <Label className="text-slate-400">Salario Mínimo (MXN diario)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={config.minimumWage}
                      onChange={(e) => updateConfig("minimumWage", e.target.value)}
                      className="bg-slate-800 border-slate-700 text-white"
                    />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-slate-400">Horas Extra Dobles</Label>
                      <div className="flex items-center">
                        <Input
                          type="number"
                          step="0.1"
                          value={config.overtimeDoubleRate}
                          onChange={(e) => updateConfig("overtimeDoubleRate", e.target.value)}
                          className="bg-slate-800 border-slate-700 text-white"
                        />
                        <span className="ml-2 text-slate-300">x salario</span>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      <Label className="text-slate-400">Horas Extra Triples</Label>
                      <div className="flex items-center">
                        <Input
                          type="number"
                          step="0.1"
                          value={config.overtimeTripleRate}
                          onChange={(e) => updateConfig("overtimeTripleRate", e.target.value)}
                          className="bg-slate-800 border-slate-700 text-white"
                        />
                        <span className="ml-2 text-slate-300">x salario</span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      <div className="p-4 bg-slate-800/30 border border-slate-700 rounded-lg">
        <h3 className="font-semibold text-white mb-2">Resumen de Configuración</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center p-3 bg-slate-900 rounded-lg">
            <div className="text-sm text-slate-400">Total IMSS Patronal</div>
            <div className="text-lg font-bold text-white">
              {formatPercentage(config.imssEmployerRate + config.imssDisabilityRate + config.imssRetirementRate)}
            </div>
          </div>
          
          <div className="text-center p-3 bg-slate-900 rounded-lg">
            <div className="text-sm text-slate-400">Total IMSS Obrero</div>
            <div className="text-lg font-bold text-white">
              {formatPercentage(config.imssEmployeeRate)}
            </div>
          </div>
          
          <div className="text-center p-3 bg-slate-900 rounded-lg">
            <div className="text-sm text-slate-400">Infonavit</div>
            <div className="text-lg font-bold text-white">
              {formatPercentage(config.infonavitRate)}
            </div>
          </div>
          
          <div className="text-center p-3 bg-slate-900 rounded-lg">
            <div className="text-sm text-slate-400">Costo Total Patrón</div>
            <div className="text-lg font-bold text-white">
              {formatPercentage(
                config.imssEmployerRate + 
                config.imssDisabilityRate + 
                config.imssRetirementRate + 
                config.sarEmployerRate +
                config.infonavitRate +
                config.vacationPremium
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
    </DashboardLayout>
  )
}
