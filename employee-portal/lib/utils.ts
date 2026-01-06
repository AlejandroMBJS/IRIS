/**
 * @file lib/utils.ts
 * @description Utility functions for the IRIS Payroll System frontend. Provides helper functions
 * for common operations like CSS class merging for Tailwind CSS styling.
 *
 * USER PERSPECTIVE:
 *   - Ensures consistent and conflict-free styling across UI components
 *   - Enables smooth theme transitions and responsive design
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify:
 *     - Add new utility functions for formatting (dates, currency, phone numbers)
 *     - Add validation helpers (RFC, CURP, NSS format validation)
 *     - Add data transformation utilities (camelCase to snake_case, etc.)
 *     - Add common calculation helpers (age from birthdate, seniority, etc.)
 *
 *   CAUTION:
 *     - cn() function is used extensively in all components for className merging
 *     - Changes to cn() behavior could affect styling across entire application
 *
 *   DO NOT modify:
 *     - cn() function signature or behavior (breaks all component styling)
 *     - Remove dependencies (clsx, tailwind-merge) without proper alternatives
 *
 * EXPORTS:
 *   - cn(): Merges Tailwind CSS classes with conflict resolution (combines clsx and twMerge)
 */

import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
