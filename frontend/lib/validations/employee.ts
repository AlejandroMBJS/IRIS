/**
 * Employee Form Validation Schemas
 *
 * Comprehensive validation for employee data using Zod
 * Includes Mexican-specific validation for RFC, CURP, NSS
 */

import { z } from "zod"

// RFC validation: 12-13 alphanumeric characters
const rfcRegex = /^[A-ZÑ&]{3,4}\d{6}[A-Z0-9]{3}$/
// CURP validation: 18 characters
const curpRegex = /^[A-Z]{4}\d{6}[HM][A-Z]{5}[0-9A-Z]\d$/
// NSS validation: 11 digits
const nssRegex = /^\d{11}$/
// Phone validation: 10 digits
const phoneRegex = /^\d{10}$/
// Email validation: standard email
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

/**
 * Employee Basic Information Schema
 */
export const employeeBasicInfoSchema = z.object({
  employeeNumber: z.string()
    .min(1, "Employee number is required")
    .max(50, "Employee number must be 50 characters or less")
    .regex(/^[A-Za-z0-9-]+$/, "Employee number can only contain letters, numbers, and hyphens"),

  firstName: z.string()
    .min(1, "First name is required")
    .max(100, "First name must be 100 characters or less")
    .regex(/^[A-Za-zÁÉÍÓÚáéíóúÑñ\s]+$/, "First name can only contain letters and spaces"),

  lastName: z.string()
    .min(1, "Paternal last name is required")
    .max(100, "Last name must be 100 characters or less")
    .regex(/^[A-Za-zÁÉÍÓÚáéíóúÑñ\s]+$/, "Last name can only contain letters and spaces"),

  motherLastName: z.string()
    .max(100, "Maternal last name must be 100 characters or less")
    .regex(/^[A-Za-zÁÉÍÓÚáéíóúÑñ\s]*$/, "Maternal last name can only contain letters and spaces")
    .optional()
    .or(z.literal("")),

  birthDate: z.string()
    .or(z.date())
    .refine((date) => {
      const birthDate = new Date(date)
      const today = new Date()
      const age = today.getFullYear() - birthDate.getFullYear()
      return age >= 16 && age <= 100
    }, "Employee must be between 16 and 100 years old"),

  gender: z.enum(["male", "female", "other"], {
    message: "Please select a valid gender"
  }).optional(),

  maritalStatus: z.enum(["single", "married", "divorced", "widowed", "cohabiting"]).optional(),

  nationality: z.string().default("Mexicana"),

  birthState: z.string().optional(),

  educationLevel: z.enum([
    "none", "primary", "secondary", "high_school",
    "technical", "bachelor", "master", "doctorate"
  ]).optional(),
})

/**
 * Mexican Official IDs Schema
 */
export const employeeIdentificationSchema = z.object({
  rfc: z.string()
    .min(12, "RFC must be at least 12 characters")
    .max(13, "RFC must be at most 13 characters")
    .regex(rfcRegex, "Invalid RFC format. Example: ABCD123456XXX")
    .transform(val => val.toUpperCase()),

  curp: z.string()
    .length(18, "CURP must be exactly 18 characters")
    .regex(curpRegex, "Invalid CURP format. Example: ABCD123456HDFXXX00")
    .transform(val => val.toUpperCase()),

  nss: z.string()
    .regex(nssRegex, "NSS must be exactly 11 digits")
    .optional()
    .or(z.literal("")),

  infonavitCredit: z.string().optional(),
})

/**
 * Contact Information Schema
 */
export const employeeContactSchema = z.object({
  personalEmail: z.string()
    .email("Invalid email format")
    .optional()
    .or(z.literal("")),

  personalPhone: z.string()
    .regex(phoneRegex, "Phone must be exactly 10 digits")
    .optional()
    .or(z.literal("")),

  cellPhone: z.string()
    .regex(phoneRegex, "Cell phone must be exactly 10 digits")
    .optional()
    .or(z.literal("")),

  emergencyContact: z.string()
    .max(255, "Emergency contact name is too long")
    .optional(),

  emergencyPhone: z.string()
    .regex(phoneRegex, "Emergency phone must be exactly 10 digits")
    .optional()
    .or(z.literal("")),

  emergencyRelationship: z.string()
    .max(100, "Relationship description is too long")
    .optional(),
})

/**
 * Employment Information Schema
 */
export const employmentInfoSchema = z.object({
  hireDate: z.string()
    .or(z.date())
    .refine((date) => {
      const hireDate = new Date(date)
      const today = new Date()
      return hireDate <= today
    }, "Hire date cannot be in the future"),

  dailySalary: z.number()
    .min(200, "Daily salary must be at least 200 MXN (below minimum wage)")
    .max(50000, "Daily salary seems unrealistic. Please verify the amount"),

  collarType: z.enum(["white_collar", "blue_collar", "gray_collar"], {
    message: "Please select a valid collar type"
  }),

  position: z.string()
    .min(1, "Position is required")
    .max(255, "Position name is too long"),

  department: z.string()
    .min(1, "Department is required")
    .max(255, "Department name is too long"),

  employmentStatus: z.enum(["active", "terminated", "suspended", "leave"], {
    message: "Please select a valid employment status"
  }),

  contractType: z.enum([
    "indefinite", "fixed_term", "project_based", "seasonal",
    "trial", "training", "internship"
  ]).optional(),
})

/**
 * Payroll Period Creation Schema
 */
export const payrollPeriodSchema = z.object({
  periodCode: z.string()
    .min(1, "Period code is required")
    .regex(/^\d{4}-\w+\d{2}$/, "Period code must follow format: YYYY-XXnn (e.g., 2025-BW01)"),

  year: z.number()
    .int("Year must be a whole number")
    .min(2020, "Year must be 2020 or later")
    .max(2100, "Year must be 2100 or earlier"),

  periodNumber: z.number()
    .int("Period number must be a whole number")
    .min(1, "Period number must be at least 1")
    .max(53, "Period number cannot exceed 53"),

  frequency: z.enum(["weekly", "biweekly", "monthly"], {
    message: "Please select a valid frequency"
  }),

  startDate: z.date(),
  endDate: z.date(),
  paymentDate: z.date(),

  description: z.string().optional(),
}).refine((data) => {
  return data.endDate >= data.startDate
}, {
  message: "End date must be on or after start date",
  path: ["endDate"]
}).refine((data) => {
  return data.paymentDate >= data.endDate
}, {
  message: "Payment date must be on or after end date",
  path: ["paymentDate"]
}).refine((data) => {
  const duration = (data.endDate.getTime() - data.startDate.getTime()) / (1000 * 60 * 60 * 24)

  if (data.frequency === "weekly") {
    return duration >= 6 && duration <= 8
  } else if (data.frequency === "biweekly") {
    return duration >= 13 && duration <= 15
  } else if (data.frequency === "monthly") {
    return duration >= 28 && duration <= 32
  }
  return true
}, {
  message: "Period duration doesn't match the selected frequency",
  path: ["endDate"]
})

/**
 * Salary Update Schema
 */
export const salaryUpdateSchema = z.object({
  newDailySalary: z.number()
    .positive("Salary must be positive")
    .min(200, "Salary is below minimum wage (200 MXN/day)")
    .max(50000, "Salary seems unrealistic. Please verify."),

  effectiveDate: z.date()
    .refine((date) => date >= new Date(), {
      message: "Effective date cannot be in the past"
    }),

  reason: z.string()
    .min(10, "Please provide a reason for the salary change (min 10 characters)")
    .max(500, "Reason is too long (max 500 characters)")
    .optional(),
})

/**
 * Complete Employee Form Schema
 * Combines all employee-related schemas
 */
export const completeEmployeeSchema = employeeBasicInfoSchema
  .merge(employeeIdentificationSchema)
  .merge(employeeContactSchema)
  .merge(employmentInfoSchema)

/**
 * Type exports for TypeScript
 */
export type EmployeeBasicInfo = z.infer<typeof employeeBasicInfoSchema>
export type EmployeeIdentification = z.infer<typeof employeeIdentificationSchema>
export type EmployeeContact = z.infer<typeof employeeContactSchema>
export type EmploymentInfo = z.infer<typeof employmentInfoSchema>
export type CompleteEmployee = z.infer<typeof completeEmployeeSchema>
export type PayrollPeriod = z.infer<typeof payrollPeriodSchema>
export type SalaryUpdate = z.infer<typeof salaryUpdateSchema>
