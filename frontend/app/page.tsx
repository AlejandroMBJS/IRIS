/**
 * @file app/page.tsx
 * @description Root page that redirects unauthenticated users to login
 *
 * USER PERSPECTIVE:
 *   - Users are automatically redirected to the login page when accessing the root URL
 *   - This is the entry point of the application
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: The redirect destination (/auth/login)
 *   CAUTION: This is a server-side redirect, consider auth state before changing
 *   DO NOT modify: The redirect function import or component structure
 *
 * KEY COMPONENTS:
 *   - Next.js redirect() function for server-side navigation
 *
 * API ENDPOINTS USED:
 *   - None (client-side redirect only)
 */

import { redirect } from "next/navigation"

export default function HomePage() {
  redirect("/auth/login")
}
