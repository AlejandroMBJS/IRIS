/**
 * @file app/layout.tsx
 * @description Root layout component that wraps the entire application with providers and global styles
 *
 * USER PERSPECTIVE:
 *   - Provides consistent layout, theming, and toast notifications across all pages
 *   - Sets the application's base styling and font
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: Metadata (title, description), font selection, global CSS imports
 *   CAUTION: Changes to Providers affect entire app, test thoroughly
 *   DO NOT modify: The html/body structure without understanding Next.js app router
 *
 * KEY COMPONENTS:
 *   - Providers: Authentication, theme, and state management context
 *   - SonnerToaster: Toast notification system (Sonner library)
 *   - ShadcnToaster: Alternative toast system (shadcn/ui)
 *   - Inter font from Google Fonts
 *
 * API ENDPOINTS USED:
 *   - None (layout only)
 */

import type React from "react"
import { Inter } from "next/font/google"
import "./globals.css"
import { Providers } from "@/components/providers"
import { Toaster as SonnerToaster } from "@/components/ui/sonner"
import { Toaster as ShadcnToaster } from "@/components/ui/toaster"

const inter = Inter({ subsets: ["latin"] })

export const metadata = {
  title: "IRIS - Portal del Empleado",
  description: "Portal de solicitudes y aprobaciones para empleados",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="es" suppressHydrationWarning>
      <body className={`${inter.className} bg-background text-foreground`}>
        <Providers>
          {children}
          <SonnerToaster />
          <ShadcnToaster />
        </Providers>
      </body>
    </html>
  )
}
