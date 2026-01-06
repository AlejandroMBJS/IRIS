/**
 * @file app/auth/register/page.tsx
 * @description Company registration page for new organizations to sign up
 *
 * USER PERSPECTIVE:
 *   - New companies can register with company details, contact info, and credentials
 *   - Form includes validation for RFC, email, phone, and password requirements
 *   - Password must meet security requirements (8+ chars, upper, lower, number, special)
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Form fields, validation rules, password requirements display
 *   CAUTION: RFC validation format must match Mexican tax ID standards
 *   DO NOT modify: The registerCompany() function signature without updating backend
 *
 * KEY COMPONENTS:
 *   - Card: Multi-field registration form container
 *   - RFCInput: Custom Mexican RFC input with validation
 *   - Input: Standard form inputs
 *   - Button: Form submission and navigation
 *
 * API ENDPOINTS USED:
 *   - POST /api/auth/register (via registerCompany() helper)
 */

"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Building, Mail, Phone, User, Lock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { useToast } from "@/hooks/use-toast"
import { RFCInput } from "@/components/ui/mexican-inputs"
import { registerCompany } from "@/lib/auth"

export default function RegisterPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    companyName: "",
    rfc: "",
    email: "",
    phone: "",
    contactName: "",
    password: "",
    confirmPassword: "",
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (formData.password !== formData.confirmPassword) {
      toast({
        title: "Error",
        description: "Las contraseñas no coinciden",
        variant: "destructive",
      })
      return
    }

    setIsLoading(true)

    try {
      const result = await registerCompany(formData)
      
      if (result.success) {
        toast({
          title: "Registro exitoso",
          description: "Tu empresa ha sido registrada. Inicia sesion para continuar.",
        })
        router.push("/auth/login")
      } else {
        toast({
          title: "Error en el registro",
          description: result.message,
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Ocurrió un error inesperado",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-950 p-4">
      <Card className="w-full max-w-2xl border-slate-800 bg-slate-900/50 backdrop-blur-sm">
        <CardHeader className="space-y-1">
          <div className="flex items-center justify-center mb-4">
            <div className="flex items-baseline gap-2">
              <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
                IRIS
              </h1>
              <span className="text-xl font-medium text-slate-400">Talent</span>
            </div>
          </div>
          <CardTitle className="text-2xl font-bold text-white text-center">
            Registra tu Empresa
          </CardTitle>
          <CardDescription className="text-center text-slate-400">
            Completa la información para comenzar a usar el sistema de nómina
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="companyName" className="text-slate-300">
                  Nombre de la Empresa *
                </Label>
                <div className="relative">
                  <Building className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="companyName"
                    placeholder="Mi Empresa S.A. de C.V."
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.companyName}
                    onChange={(e) => setFormData({ ...formData, companyName: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="rfc" className="text-slate-300">
                  RFC de la Empresa *
                </Label>
                <RFCInput
                  value={formData.rfc}
                  onChange={(value) => setFormData({ ...formData, rfc: value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="email" className="text-slate-300">
                  Correo Electrónico *
                </Label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="email"
                    type="email"
                    placeholder="contacto@empresa.com"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="phone" className="text-slate-300">
                  Teléfono *
                </Label>
                <div className="relative">
                  <Phone className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="phone"
                    type="tel"
                    placeholder="55 1234 5678"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.phone}
                    onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="contactName" className="text-slate-300">
                  Persona de Contacto *
                </Label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="contactName"
                    placeholder="Juan Pérez"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.contactName}
                    onChange={(e) => setFormData({ ...formData, contactName: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-slate-300">
                  Contraseña *
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    required
                  />
                </div>
                <p className="text-xs text-slate-500">
                  Minimo 8 caracteres, mayuscula, minuscula, numero y caracter especial (!@#$%^&*())
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirmPassword" className="text-slate-300">
                  Confirmar Contraseña *
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="confirmPassword"
                    type="password"
                    placeholder="••••••••"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.confirmPassword}
                    onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                    required
                  />
                </div>
              </div>
            </div>

            <div className="text-sm text-slate-400 space-y-2">
              <p>Al registrarte, aceptas nuestros:</p>
              <ul className="list-disc list-inside space-y-1">
                <li>Términos de Servicio</li>
                <li>Política de Privacidad</li>
                <li>Acuerdo de Confidencialidad de Datos</li>
              </ul>
            </div>

            <div className="flex gap-4">
              <Button
                type="button"
                variant="outline"
                className="flex-1 border-slate-700 text-slate-300 hover:bg-slate-800"
                onClick={() => router.push("/auth/login")}
              >
                Volver al Login
              </Button>
              <Button
                type="submit"
                className="flex-1 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white"
                disabled={isLoading}
              >
                {isLoading ? "Registrando..." : "Registrar Empresa"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
