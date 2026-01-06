/**
 * @file components/providers.tsx
 * @description React Context Providers setup for the IRIS Payroll System. Wraps the application
 * with necessary provider components for global state management, theming, and other cross-cutting concerns.
 * This is the top-level provider configuration used in the root layout.
 *
 * USER PERSPECTIVE:
 *   - Enables dark/light theme switching with system preference detection
 *   - Persists theme preference across sessions
 *   - Provides smooth theme transitions without flash of unstyled content
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify:
 *     - Add new providers (QueryClientProvider for React Query, AuthProvider, NotificationProvider, etc.)
 *     - Adjust theme configuration (default theme, enableSystem, etc.)
 *     - Add provider-specific configuration options
 *     - Wrap with additional context providers as needed
 *
 *   CAUTION:
 *     - Provider order matters - some providers may depend on others
 *     - Theme attribute must match Tailwind CSS configuration (currently "class")
 *     - disableTransitionOnChange prevents FOUC but may affect animations
 *     - This component is "use client" - runs in browser, not server
 *
 *   DO NOT modify:
 *     - Remove "use client" directive (providers require client-side React)
 *     - Change theme attribute without updating tailwind.config.js darkMode setting
 *     - Remove children prop (breaks app rendering)
 *
 * EXPORTS:
 *   - Providers: Main provider wrapper component that configures:
 *     - NextThemesProvider: Theme management with dark/light mode support
 *     - Accepts ThemeProviderProps and forwards to NextThemesProvider
 */

"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

type ThemeProviderProps = React.ComponentProps<typeof NextThemesProvider>

export function Providers({ children, ...props }: ThemeProviderProps) {
  return (
    <NextThemesProvider
      attribute="class"
      defaultTheme="dark"
      enableSystem
      disableTransitionOnChange
      {...props}
    >
      {children}
    </NextThemesProvider>
  )
}
