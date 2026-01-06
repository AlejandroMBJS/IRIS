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
        description: "Passwords do not match",
        variant: "destructive",
      })
      return
    }

    setIsLoading(true)

    try {
      const result = await registerCompany(formData)
      
      if (result.success) {
        toast({
          title: "Registration successful",
          description: "Your company has been registered. Log in to continue.",
        })
        router.push("/auth/login")
      } else {
        toast({
          title: "Registration error",
          description: result.message,
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "An unexpected error occurred",
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
            Register Your Company
          </CardTitle>
          <CardDescription className="text-center text-slate-400">
            Complete the information to start using the payroll system
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="companyName" className="text-slate-300">
                  Company Name *
                </Label>
                <div className="relative">
                  <Building className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="companyName"
                    placeholder="My Company Inc."
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.companyName}
                    onChange={(e) => setFormData({ ...formData, companyName: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="rfc" className="text-slate-300">
                  Company RFC *
                </Label>
                <RFCInput
                  value={formData.rfc}
                  onChange={(value) => setFormData({ ...formData, rfc: value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="email" className="text-slate-300">
                  Email Address *
                </Label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="email"
                    type="email"
                    placeholder="contact@company.com"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="phone" className="text-slate-300">
                  Phone *
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
                  Contact Person *
                </Label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                  <Input
                    id="contactName"
                    placeholder="John Doe"
                    className="pl-10 bg-slate-800 border-slate-700 text-white"
                    value={formData.contactName}
                    onChange={(e) => setFormData({ ...formData, contactName: e.target.value })}
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-slate-300">
                  Password *
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
                  Minimum 8 characters, uppercase, lowercase, number and special character (!@#$%^&*())
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirmPassword" className="text-slate-300">
                  Confirm Password *
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
              <p>By registering, you agree to our:</p>
              <ul className="list-disc list-inside space-y-1">
                <li>Terms of Service</li>
                <li>Privacy Policy</li>
                <li>Data Confidentiality Agreement</li>
              </ul>
            </div>

            <div className="flex gap-4">
              <Button
                type="button"
                variant="outline"
                className="flex-1 border-slate-700 text-slate-300 hover:bg-slate-800"
                onClick={() => router.push("/auth/login")}
              >
                Back to Login
              </Button>
              <Button
                type="submit"
                className="flex-1 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white"
                disabled={isLoading}
              >
                {isLoading ? "Registering..." : "Register Company"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
