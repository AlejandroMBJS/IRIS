/**
 * @file app/auth/login/page.tsx
 * @description User login page with email/password authentication
 *
 * USER PERSPECTIVE:
 *   - Users enter their email and password to access the system
 *   - Password visibility can be toggled
 *   - "Remember me" option available
 *   - Link to register page (contacts admin) and forgot password
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Form validation rules, styling, error messages
 *   CAUTION: Changes to authentication flow require backend coordination
 *   DO NOT modify: The login() function signature without updating @/lib/auth
 *
 * KEY COMPONENTS:
 *   - Card: Container for login form
 *   - Input: Email and password fields
 *   - Button: Submit button
 *   - useToast: For displaying success/error messages
 *
 * API ENDPOINTS USED:
 *   - POST /api/auth/login (via login() helper)
 */

"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Eye, EyeOff, Lock, Mail } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { useToast } from "@/hooks/use-toast"
import { login } from "@/lib/auth"

export default function LoginPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      const result = await login(formData.email, formData.password)
      
      if (result.success) {
        toast({
          title: "Login successful",
          description: "Welcome to the IRIS Talent system",
        })
        router.push("/dashboard")
      } else {
        toast({
          title: "Authentication error",
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
      <Card className="w-full max-w-md border-slate-800 bg-slate-900/50 backdrop-blur-sm">
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
            Log In
          </CardTitle>
          <CardDescription className="text-center text-slate-400">
            Enter your credentials to access the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email" className="text-slate-300">
                Email
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                <Input
                  id="email"
                  type="email"
                  placeholder="admin@empresa.com"
                  className="pl-10 bg-slate-800 border-slate-700 text-white"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  required
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="password" className="text-slate-300">
                Password
              </Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500" size={20} />
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="••••••••"
                  className="pl-10 pr-10 bg-slate-800 border-slate-700 text-white"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-500 hover:text-slate-300"
                >
                  {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="remember"
                  className="rounded border-slate-700 bg-slate-800 text-primary"
                />
                <Label htmlFor="remember" className="text-sm text-slate-400">
                  Remember me
                </Label>
              </div>
              <Link
                href="/auth/forgot-password"
                className="text-sm text-primary hover:text-primary/80 transition-colors"
              >
                Forgot your password?
              </Link>
            </div>

            <Button
              type="submit"
              className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white"
              disabled={isLoading}
            >
              {isLoading ? "Logging in..." : "Log In"}
            </Button>

            <div className="text-center text-sm text-slate-500">
              Don't have an account?{" "}
              <Link href="/auth/register" className="text-primary hover:text-primary/80 transition-colors">
                Contact administrator
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
