"use client"

import { useState } from "react"
import { Download, FileText, ShieldAlert, Building, TrendingUp } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

interface ComplianceReport {
  id: string
  name: string
  description: string
  period: string
  generatedDate: string
  status: "pending" | "generated" | "submitted"
  downloadUrl?: string
}

export function ComplianceReports() {
  const [selectedPeriod, setSelectedPeriod] = useState("2024-12")
  const [reports, setReports] = useState<ComplianceReport[]>([
    {
      id: "1",
      name: "Reporte IMSS",
      description: "Declaración patronal IMSS",
      period: "2024-12",
      generatedDate: "2024-12-15",
      status: "generated",
      downloadUrl: "/reports/imss-2024-12.pdf"
    },
    {
      id: "2",
      name: "Reporte ISR",
      description: "Retenciones ISR mensuales",
      period: "2024-12",
      generatedDate: "2024-12-10",
      status: "submitted",
      downloadUrl: "/reports/isr-2024-12.pdf"
    },
    {
      id: "3",
      name: "Reporte Infonavit",
      description: "Aportaciones Infonavit",
      period: "2024-12",
      generatedDate: "2024-12-05",
      status: "generated",
      downloadUrl: "/reports/infonavit-2024-12.pdf"
    },
    {
      id: "4",
      name: "DIOT",
      description: "Declaración Informativa de Operaciones con Terceros",
      period: "2024-12",
      generatedDate: "2024-12-20",
      status: "pending",
    }
  ])

  const handleGenerateReport = (reportId: string) => {
    // Simulate report generation
    const updatedReports = reports.map(report =>
      report.id === reportId
        ? {
            ...report,
            status: "generated" as const,
            generatedDate: new Date().toISOString().split('T')[0],
            downloadUrl: `/reports/${report.name.toLowerCase().replace(/\s+/g, '-')}-${selectedPeriod}.pdf`
          }
        : report
    )
    setReports(updatedReports)
  }

  const handleDownload = (reportId: string) => {
    const report = reports.find(r => r.id === reportId)
    if (report?.downloadUrl) {
      // Simulate download
      window.open(report.downloadUrl, '_blank')
    }
  }

  const handleSubmit = (reportId: string) => {
    // Simulate submission
    const updatedReports = reports.map(report =>
      report.id === reportId
        ? { ...report, status: "submitted" as const }
        : report
    )
    setReports(updatedReports)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Reportes de Cumplimiento</h1>
          <p className="text-slate-400 mt-1">Reportes regulatorios para autoridades mexicanas</p>
        </div>
        
        <div className="flex items-center gap-3">
          <Select value={selectedPeriod} onValueChange={setSelectedPeriod}>
            <SelectTrigger className="w-[180px] bg-slate-800 border-slate-700 text-white">
              <SelectValue placeholder="Seleccionar periodo" />
            </SelectTrigger>
            <SelectContent className="bg-slate-800 border-slate-700">
              <SelectItem value="2024-12">Diciembre 2024</SelectItem>
              <SelectItem value="2024-11">Noviembre 2024</SelectItem>
              <SelectItem value="2024-10">Octubre 2024</SelectItem>
              <SelectItem value="2024-09">Septiembre 2024</SelectItem>
            </SelectContent>
          </Select>
          
          <Button className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
            <Download className="h-4 w-4 mr-2" />
            Exportar Todo
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card className="border-slate-700 bg-slate-900">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm">IMSS Pendiente</p>
                <p className="text-3xl font-bold text-white mt-2">$45,280</p>
              </div>
              <div className="p-3 bg-blue-600 rounded-lg">
                <ShieldAlert size={24} className="text-white" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-slate-700 bg-slate-900">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm">ISR Retenido</p>
                <p className="text-3xl font-bold text-white mt-2">$12,450</p>
              </div>
              <div className="p-3 bg-red-600 rounded-lg">
                <FileText size={24} className="text-white" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-slate-700 bg-slate-900">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm">Infonavit</p>
                <p className="text-3xl font-bold text-white mt-2">$8,920</p>
              </div>
              <div className="p-3 bg-green-600 rounded-lg">
                <Building size={24} className="text-white" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-slate-700 bg-slate-900">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm">Crecimiento</p>
                <p className="text-3xl font-bold text-white mt-2">+8.5%</p>
              </div>
              <div className="p-3 bg-amber-600 rounded-lg">
                <TrendingUp size={24} className="text-white" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="all" className="w-full">
        <TabsList className="grid grid-cols-4 w-full bg-slate-800">
          <TabsTrigger value="all" className="data-[state=active]:bg-primary">
            Todos
          </TabsTrigger>
          <TabsTrigger value="imss" className="data-[state=active]:bg-primary">
            IMSS
          </TabsTrigger>
          <TabsTrigger value="isr" className="data-[state=active]:bg-primary">
            ISR
          </TabsTrigger>
          <TabsTrigger value="other" className="data-[state=active]:bg-primary">
            Otros
          </TabsTrigger>
        </TabsList>

        <TabsContent value="all" className="space-y-4">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Reportes Disponibles</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="rounded-lg border border-slate-700 overflow-hidden">
                <Table>
                  <TableHeader className="bg-slate-800">
                    <TableRow>
                      <TableHead className="text-slate-300">Reporte</TableHead>
                      <TableHead className="text-slate-300">Descripción</TableHead>
                      <TableHead className="text-slate-300">Periodo</TableHead>
                      <TableHead className="text-slate-300">Fecha Generación</TableHead>
                      <TableHead className="text-slate-300">Estado</TableHead>
                      <TableHead className="text-slate-300">Acciones</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {reports.map((report) => (
                      <TableRow key={report.id} className="hover:bg-slate-800/50">
                        <TableCell className="font-medium text-white">
                          {report.name}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {report.description}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {report.period}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {report.generatedDate}
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            report.status === "pending" 
                              ? "bg-amber-900/30 text-amber-400" 
                              : report.status === "generated"
                              ? "bg-blue-900/30 text-blue-400"
                              : "bg-green-900/30 text-green-400"
                          }`}>
                            {report.status === "pending" ? "Pendiente" :
                             report.status === "generated" ? "Generado" : "Enviado"}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-2">
                            {report.status === "pending" && (
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleGenerateReport(report.id)}
                                className="h-8 px-3 border-slate-700 text-slate-300 hover:bg-slate-800"
                              >
                                Generar
                              </Button>
                            )}
                            
                            {report.status === "generated" && (
                              <>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => handleDownload(report.id)}
                                  className="h-8 px-3 border-slate-700 text-slate-300 hover:bg-slate-800"
                                >
                                  <Download className="h-3 w-3 mr-1" />
                                  PDF
                                </Button>
                                <Button
                                  size="sm"
                                  onClick={() => handleSubmit(report.id)}
                                  className="h-8 px-3 bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700"
                                >
                                  Enviar
                                </Button>
                              </>
                            )}
                            
                            {report.status === "submitted" && (
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleDownload(report.id)}
                                className="h-8 px-3 border-slate-700 text-slate-300 hover:bg-slate-800"
                              >
                                <Download className="h-3 w-3 mr-1" />
                                Descargar
                              </Button>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="imss">
          <Card className="border-slate-700 bg-slate-900">
            <CardHeader>
              <CardTitle className="text-white">Reportes IMSS</CardTitle>
            </CardHeader>
            <CardContent>
              {/* IMSS specific reports content */}
              <div className="space-y-4">
                <div className="p-4 bg-blue-900/20 border border-blue-800 rounded-lg">
                  <h3 className="font-semibold text-blue-400 mb-2">Formatos IMSS Disponibles</h3>
                  <ul className="space-y-2 text-slate-300">
                    <li>• IMSS-01: Aviso de alta/baja/modificación de salario</li>
                    <li>• IMSS-02: Declaración de cotizaciones</li>
                    <li>• IMSS-03: Aviso de incapacidad</li>
                    <li>• IMSS-04: Movimiento de trabajadores</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Card className="border-slate-700 bg-slate-900">
        <CardHeader>
          <CardTitle className="text-white">Calendario de Vencimientos</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[
              { name: "IMSS Mensual", dueDate: "05/01/2024", status: "pending" },
              { name: "ISR Mensual", dueDate: "17/01/2024", status: "pending" },
              { name: "INFONAVIT", dueDate: "17/01/2024", status: "pending" },
              { name: "DIOT Mensual", dueDate: "20/01/2024", status: "pending" },
              { name: "IMSS Diciembre", dueDate: "15/12/2023", status: "submitted" },
            ].map((item, index) => (
              <div key={index} className="flex items-center justify-between p-3 border border-slate-700 rounded-lg hover:bg-slate-800/50">
                <div className="flex items-center gap-3">
                  <div className={`w-3 h-3 rounded-full ${
                    item.status === "pending" ? "bg-amber-500" : "bg-green-500"
                  }`} />
                  <span className="text-white">{item.name}</span>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-slate-400">Vence: {item.dueDate}</span>
                  <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                    item.status === "pending" 
                      ? "bg-amber-900/30 text-amber-400" 
                      : "bg-green-900/30 text-green-400"
                  }`}>
                    {item.status === "pending" ? "Pendiente" : "Enviado"}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
