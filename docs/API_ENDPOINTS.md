# IRIS Payroll System - API Endpoints Documentation

## Overview
This document lists all API endpoints with their corresponding backend handlers, frontend API client functions, and the pages that consume them.

**Base URL:** `http://localhost:8080/api/v1`

---

## 1. Health/Utility Endpoints (4 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/health` | GET | `HealthHandler.HealthCheck` | `healthApi.getHealth()` | Health checks |
| `/ready` | GET | `HealthHandler.ReadyCheck` | - | Kubernetes readiness probe |
| `/live` | GET | `HealthHandler.LivenessCheck` | - | Kubernetes liveness probe |
| `/api/v1/health` | GET | `Router.Setup` (inline) | `healthApi.isServerAvailable()` | Login page, Dashboard |

**Backend File:** `internal/api/health_handler.go`

---

## 2. Authentication Endpoints (9 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/auth/register` | POST | `AuthHandler.Register` | `authApi.register()` | `/auth/signup` |
| `/auth/login` | POST | `AuthHandler.Login` | `authApi.login()` | `/auth/login` |
| `/auth/refresh` | POST | `AuthHandler.RefreshToken` | `authApi.refreshToken()` | Auto (api-client interceptor) |
| `/auth/logout` | POST | `AuthHandler.Logout` | `authApi.logout()` | Navbar (logout action) |
| `/auth/change-password` | POST | `AuthHandler.ChangePassword` | `authApi.changePassword()` | `/settings/security` |
| `/auth/forgot-password` | POST | `AuthHandler.ForgotPassword` | `authApi.forgotPassword()` | `/auth/forgot-password` |
| `/auth/reset-password` | POST | `AuthHandler.ResetPassword` | `authApi.resetPassword()` | `/auth/reset-password` |
| `/auth/profile` | GET | `AuthHandler.GetProfile` | `authApi.getProfile()` | `/settings/profile` |
| `/auth/profile` | PUT | `AuthHandler.UpdateProfile` | `authApi.updateProfile()` | `/settings/profile` |

**Backend File:** `internal/api/auth_handler.go`

---

## 3. User Management Endpoints (4 endpoints) - Admin Only

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/users` | GET | `UserHandler.ListUsers` | `userApi.getUsers()` | `/settings/users` |
| `/users` | POST | `UserHandler.CreateUser` | `userApi.createUser()` | `/settings/users` (create dialog) |
| `/users/:id` | DELETE | `UserHandler.DeleteUser` | `userApi.deleteUser()` | `/settings/users` (delete action) |
| `/users/:id/toggle-active` | PATCH | `UserHandler.ToggleUserActive` | `userApi.toggleUserActive()` | `/settings/users` (toggle action) |

**Backend File:** `internal/api/user_handler.go`

---

## 4. Employee Endpoints (9 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/employees` | GET | `EmployeeHandler.ListEmployees` | `employeeApi.getEmployees()` | `/employees` |
| `/employees` | POST | `EmployeeHandler.CreateEmployee` | `employeeApi.createEmployee()` | `/employees/new` |
| `/employees/:id` | GET | `EmployeeHandler.GetEmployee` | `employeeApi.getEmployee()` | `/employees/[id]` |
| `/employees/:id` | PUT | `EmployeeHandler.UpdateEmployee` | `employeeApi.updateEmployee()` | `/employees/[id]/edit` |
| `/employees/:id` | DELETE | `EmployeeHandler.DeleteEmployee` | `employeeApi.deleteEmployee()` | `/employees` (delete action) |
| `/employees/:id/terminate` | POST | `EmployeeHandler.TerminateEmployee` | `employeeApi.terminateEmployee()` | `/employees/[id]` (terminate action) |
| `/employees/:id/salary` | PUT | `EmployeeHandler.UpdateSalary` | `employeeApi.updateSalary()` | `/employees/[id]` (salary update) |
| `/employees/stats` | GET | `EmployeeHandler.GetEmployeeStats` | `employeeApi.getStats()` | `/dashboard` |
| `/employees/validate-ids` | POST | `EmployeeHandler.ValidateMexicanIDs` | `employeeApi.validateMexicanIds()` | `/employees/new` (validation) |

**Backend File:** `internal/api/employee_handler.go`

---

## 5. Payroll Period Endpoints (3 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/payroll/periods` | GET | `PayrollPeriodHandler.GetPeriods` | `payrollApi.getPeriods()` | `/payroll`, Period selector |
| `/payroll/periods` | POST | `PayrollPeriodHandler.CreatePeriod` | `payrollApi.createPeriod()` | `/payroll/periods/new` |
| `/payroll/periods/:id` | GET | `PayrollPeriodHandler.GetPeriod` | `payrollApi.getPeriod()` | `/payroll/periods/[id]` |

**Backend File:** `internal/api/payroll_period_handler.go`

---

## 6. Payroll Calculation Endpoints (10 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/payroll/calculate` | POST | `PayrollHandler.CalculatePayroll` | `payrollApi.calculatePayroll()` | `/payroll` (calculate action) |
| `/payroll/bulk-calculate` | POST | `PayrollHandler.BulkCalculatePayroll` | `payrollApi.bulkCalculatePayroll()` | `/payroll` (batch calculate) |
| `/payroll/calculation/:period_id/:employee_id` | GET | `PayrollHandler.GetPayrollCalculation` | `payrollApi.getPayrollCalculation()` | `/payroll/[id]` |
| `/payroll/period/:period_id` | GET | `PayrollHandler.GetPayrollByPeriod` | `payrollApi.getPayrollByPeriod()` | `/payroll` |
| `/payroll/payslip/:periodId/:employeeId` | GET | `PayrollHandler.GetPayslip` | `payrollApi.getPayslip()` | `/payroll/payslip/[id]` |
| `/payroll/payslip/:periodId/:employeeId?format=pdf` | GET | `PayrollHandler.GetPayslip` | `payrollApi.getPayslip(id, id, 'pdf')` | Download PDF button |
| `/payroll/payslip/:periodId/:employeeId?format=xml` | GET | `PayrollHandler.GetPayslip` | `payrollApi.getPayslip(id, id, 'xml')` | Download CFDI button |
| `/payroll/summary/:periodId` | GET | `PayrollHandler.GetPayrollSummary` | `payrollApi.getPayrollSummary()` | `/payroll`, `/dashboard` |
| `/payroll/approve/:periodId` | POST | `PayrollHandler.ApprovePayroll` | `payrollApi.approvePayroll()` | `/payroll` (approve action) |
| `/payroll/payment/:periodId` | POST | `PayrollHandler.ProcessPayment` | `payrollApi.processPayment()` | `/payroll` (payment action) |
| `/payroll/payment/:period_id` | GET | `PayrollHandler.GetPaymentStatus` | `payrollApi.getPaymentStatus()` | `/payroll` |
| `/payroll/concept-totals/:periodId` | GET | `PayrollHandler.GetConceptTotals` | `payrollApi.getConceptTotals()` | `/payroll` |

**Backend File:** `internal/api/payroll_handler.go`

---

## 7. Incidence Type Endpoints (4 endpoints) - Catalog

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/incidence-types` | GET | `IncidenceHandler.ListIncidenceTypes` | `incidenceTypeApi.getAll()` | `/incidences/new`, `/incidences` |
| `/incidence-types` | POST | `IncidenceHandler.CreateIncidenceType` | `incidenceTypeApi.create()` | Admin settings |
| `/incidence-types/:id` | PUT | `IncidenceHandler.UpdateIncidenceType` | `incidenceTypeApi.update()` | Admin settings |
| `/incidence-types/:id` | DELETE | `IncidenceHandler.DeleteIncidenceType` | `incidenceTypeApi.delete()` | Admin settings |

**Backend File:** `internal/api/incidence_handler.go`

---

## 8. Incidence Endpoints (7 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/incidences` | GET | `IncidenceHandler.ListIncidences` | `incidenceApi.getAll()` | `/incidences` |
| `/incidences` | POST | `IncidenceHandler.CreateIncidence` | `incidenceApi.create()` | `/incidences/new` |
| `/incidences/:id` | GET | `IncidenceHandler.GetIncidence` | `incidenceApi.get()` | `/incidences/[id]` |
| `/incidences/:id` | PUT | `IncidenceHandler.UpdateIncidence` | `incidenceApi.update()` | `/incidences/[id]/edit` |
| `/incidences/:id` | DELETE | `IncidenceHandler.DeleteIncidence` | `incidenceApi.delete()` | `/incidences` (delete action) |
| `/incidences/:id/approve` | POST | `IncidenceHandler.ApproveIncidence` | `incidenceApi.approve()` | `/incidences` (approve action) |
| `/incidences/:id/reject` | POST | `IncidenceHandler.RejectIncidence` | `incidenceApi.reject()` | `/incidences` (reject action) |

**Backend File:** `internal/api/incidence_handler.go`

---

## 9. Employee Incidence Endpoints (3 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/employees/:id/incidences` | GET | `IncidenceHandler.GetEmployeeIncidences` | `incidenceApi.getByEmployee()` | `/employees/[id]` |
| `/employees/:id/vacation-balance` | GET | `IncidenceHandler.GetEmployeeVacationBalance` | `incidenceApi.getVacationBalance()` | `/employees/[id]` |
| `/employees/:id/absence-summary` | GET | `IncidenceHandler.GetEmployeeAbsenceSummary` | `incidenceApi.getAbsenceSummary()` | `/employees/[id]` |

**Backend File:** `internal/api/incidence_handler.go`

---

## 10. Evidence/Upload Endpoints (5 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/evidence/incidence/:incidence_id` | POST | `UploadHandler.UploadEvidence` | `evidenceApi.upload()` | `/incidences/[id]` (upload form) |
| `/evidence/incidence/:incidence_id` | GET | `UploadHandler.ListEvidence` | `evidenceApi.list()` | `/incidences/[id]` |
| `/evidence/:evidence_id` | GET | `UploadHandler.GetEvidence` | `evidenceApi.get()` | `/incidences/[id]` |
| `/evidence/:evidence_id/download` | GET | `UploadHandler.DownloadEvidence` | `evidenceApi.download()` | Download button |
| `/evidence/:evidence_id` | DELETE | `UploadHandler.DeleteEvidence` | `evidenceApi.delete()` | Delete button |

**Backend File:** `internal/api/upload_handler.go`

---

## 11. Catalog Endpoints (3 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/catalogs/concepts` | GET | `CatalogHandler.GetPayrollConcepts` | `catalogApi.getPayrollConcepts()` | `/payroll` |
| `/catalogs/concepts` | POST | `CatalogHandler.CreatePayrollConcept` | `catalogApi.createConcept()` | Admin settings |
| `/catalogs/incidence-types` | GET | `CatalogHandler.GetIncidenceTypes` | `catalogApi.getIncidenceTypes()` | `/incidences/new` |

**Backend File:** `internal/api/catalog_handler.go`

---

## 12. Report Endpoints (2 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/reports/generate` | POST | `ReportHandler.GenerateReport` | `reportApi.generateReport()` | `/reports` |
| `/reports/history` | GET | `ReportHandler.GetReportHistory` | `reportApi.getReportHistory()` | `/reports` |

**Backend File:** `internal/api/report_handler.go`

---

## 13. Announcement Endpoints (6 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/announcements` | GET | `AnnouncementHandler.ListAnnouncements` | `announcementApi.getAll()` | `/announcements`, Dashboard |
| `/announcements` | POST | `AnnouncementHandler.CreateAnnouncement` | `announcementApi.create()` | `/announcements/new` |
| `/announcements/:id` | GET | `AnnouncementHandler.GetAnnouncement` | `announcementApi.get()` | `/announcements/[id]` |
| `/announcements/:id` | DELETE | `AnnouncementHandler.DeleteAnnouncement` | `announcementApi.delete()` | `/announcements` |
| `/announcements/:id/read` | POST | `AnnouncementHandler.MarkAsRead` | `announcementApi.markAsRead()` | Auto (on view) |
| `/announcements/unread-count` | GET | `AnnouncementHandler.GetUnreadCount` | `announcementApi.getUnreadCount()` | Navbar badge |

**Backend File:** `internal/api/announcement_handler.go`

### Announcement Scope Rules

| Role | Can Create ALL/COMPANY Scope | Can Create TEAM Scope |
|------|------------------------------|------------------------|
| admin | Yes | Yes |
| hr | Yes | Yes |
| hr_and_pr | Yes | Yes |
| supervisor | No | Yes (subordinates only) |
| manager | No | Yes (subordinates only) |
| sup_and_gm | No | Yes (subordinates only) |
| employee | No | No |

### Announcement Visibility Rules

- **ALL/COMPANY scope**: Visible to all employees
- **TEAM scope**: Visible only to the creator and their direct subordinates (users whose `supervisor_id` matches the creator)

---

## 14. Absence Request Endpoints (8 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/absence-requests` | POST | `AbsenceRequestHandler.Create` | `absenceRequestApi.create()` | `/requests/new` |
| `/absence-requests/my-requests` | GET | `AbsenceRequestHandler.GetMyRequests` | `absenceRequestApi.getMyRequests()` | `/requests` |
| `/absence-requests/pending/:stage` | GET | `AbsenceRequestHandler.GetPendingByStage` | `absenceRequestApi.getPendingByStage()` | `/approvals/supervisor`, `/approvals/manager`, `/approvals/hr` |
| `/absence-requests/:id/approve` | POST | `AbsenceRequestHandler.Approve` | `absenceRequestApi.approve()` | Approval dialogs |
| `/absence-requests/:id` | DELETE | `AbsenceRequestHandler.Delete` | `absenceRequestApi.delete()` | `/requests` |
| `/absence-requests/:id/archive` | PATCH | `AbsenceRequestHandler.Archive` | `absenceRequestApi.archive()` | `/requests` |
| `/absence-requests/overlapping` | GET | `AbsenceRequestHandler.GetOverlapping` | `absenceRequestApi.getOverlapping()` | Date conflict check |
| `/absence-requests/counts` | GET | `AbsenceRequestHandler.GetCounts` | `absenceRequestApi.getCounts()` | Dashboard badges |

**Backend File:** `internal/api/absence_request_handler.go`

### Request Types

- `VACATION` - Vacation/time-off request
- `SICK_LEAVE` - Sick leave with medical documentation
- `PERSONAL_LEAVE` - Personal time off
- `SHIFT_CHANGE` - Request to change work shift

### Approval Workflow

1. **SUPERVISOR** stage → Supervisor approves/declines
2. **MANAGER** stage → General Manager approves/declines
3. **HR** stage → HR final approval, creates payroll incidence
4. **COMPLETED** → Request fully approved, incidence active

### Shift Change Auto-Processing

When a `SHIFT_CHANGE` request receives final HR approval:
1. A payroll incidence is created
2. A `ShiftException` record is auto-created for each date in the range
3. The employee's schedule is updated automatically

---

## 15. Shift Endpoints (2 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/shifts` | GET | `ShiftHandler.ListShifts` | `shiftApi.getAll()` | `/requests/new` (shift selector) |
| `/shifts/:id` | GET | `ShiftHandler.GetShift` | `shiftApi.get()` | Shift details |

**Backend File:** `internal/api/shift_handler.go`

---

## 16. HR Calendar Endpoints (2 endpoints)

| Endpoint | Method | Backend Handler | API Client | Frontend Page |
|----------|--------|-----------------|------------|---------------|
| `/calendar/events` | GET | `CalendarHandler.GetEvents` | `calendarApi.getEvents()` | `/hr/calendar` |
| `/calendar/employees` | GET | `CalendarHandler.GetEmployees` | `calendarApi.getEmployees()` | `/hr/calendar` sidebar |

**Backend File:** `internal/api/calendar_handler.go`

### Calendar Event Query Parameters

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| start_date | string | Yes | YYYY-MM-DD |
| end_date | string | Yes | YYYY-MM-DD |
| employee_ids[] | string[] | No | Filter by specific employees |
| collar_types[] | string[] | No | white_collar, blue_collar, gray_collar |
| event_types[] | string[] | No | absence, incidence, shift_change |
| status | string | No | pending, approved, declined |

### Calendar Event Types

- **absence** - Absence requests (vacations, sick leave, etc.)
- **incidence** - Payroll incidences
- **shift_change** - Shift change records (from ShiftException)

---

## Summary

### Total Endpoints: 81

### By Category:
- **Health/Utility**: 4 endpoints
- **Authentication**: 9 endpoints
- **User Management**: 4 endpoints (Admin only)
- **Employees**: 9 endpoints
- **Payroll Periods**: 3 endpoints
- **Payroll Calculations**: 12 endpoints (including payslip formats)
- **Incidence Types**: 4 endpoints
- **Incidences**: 7 endpoints
- **Employee Incidences**: 3 endpoints
- **Evidence/Upload**: 5 endpoints
- **Catalogs**: 3 endpoints
- **Reports**: 2 endpoints
- **Announcements**: 6 endpoints
- **Absence Requests**: 8 endpoints
- **Shifts**: 2 endpoints
- **HR Calendar**: 2 endpoints

---

## Backend File Locations

| Handler | File Path |
|---------|-----------|
| Router | `internal/api/router.go` |
| Health | `internal/api/health_handler.go` |
| Auth | `internal/api/auth_handler.go` |
| User | `internal/api/user_handler.go` |
| Employee | `internal/api/employee_handler.go` |
| Payroll Period | `internal/api/payroll_period_handler.go` |
| Payroll | `internal/api/payroll_handler.go` |
| Incidence | `internal/api/incidence_handler.go` |
| Upload | `internal/api/upload_handler.go` |
| Catalog | `internal/api/catalog_handler.go` |
| Report | `internal/api/report_handler.go` |
| Announcement | `internal/api/announcement_handler.go` |
| Absence Request | `internal/api/absence_request_handler.go` |
| Shift | `internal/api/shift_handler.go` |
| Calendar | `internal/api/calendar_handler.go` |

## Service Layer Files

| Service | File Path |
|---------|-----------|
| Auth | `internal/services/auth_service.go` |
| User | `internal/services/user_service.go` |
| Employee | `internal/services/employee_service.go` |
| Payroll | `internal/services/payroll_service.go` |
| Payroll Period | `internal/services/payroll_period_service.go` |
| Incidence | `internal/services/incidence_service.go` |
| Upload | `internal/services/upload_service.go` |
| Catalog | `internal/services/catalog_service.go` |
| Report | `internal/services/report_service.go` |
| Tax Calculation | `internal/services/tax_calculation_service.go` |
| CFDI | `internal/services/cfdi_service.go` |
| Announcement | `internal/services/announcement_service.go` |
| Absence Request | `internal/services/absence_request_service.go` |
| Calendar | `internal/services/calendar_service.go` |

## Frontend API Client

**File:** `frontend/lib/api-client.ts`

All API client functions are organized into namespaced objects:
- `authApi` - Authentication functions
- `userApi` - User management functions
- `employeeApi` - Employee CRUD functions
- `payrollApi` - Payroll calculation and period functions
- `incidenceTypeApi` - Incidence type catalog functions
- `incidenceApi` - Incidence management functions
- `evidenceApi` - Evidence upload functions
- `catalogApi` - General catalog functions
- `reportApi` - Report generation functions
- `healthApi` - Health check functions
- `announcementApi` - Announcement functions
- `absenceRequestApi` - Absence/time-off request functions
- `shiftApi` - Shift catalog functions
- `calendarApi` - HR calendar functions

---

## Test Commands

```bash
# Register admin user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"company_name":"IRIS Test","company_rfc":"ABC123456XYZ","email":"admin@iris.com","password":"Admin123!","role":"admin","full_name":"Admin User"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@iris.com","password":"Admin123!"}'

# List employees (with token)
curl http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer <TOKEN>"

# Create employee
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number":"EMP001",
    "first_name":"Juan",
    "last_name":"Perez",
    "date_of_birth":"1990-01-15T00:00:00Z",
    "gender":"male",
    "rfc":"PERJ900115XXX",
    "curp":"PERJ900115HDFRRN01",
    "hire_date":"2024-01-01T00:00:00Z",
    "daily_salary":500,
    "employment_status":"active",
    "employee_type":"permanent",
    "collar_type":"white"
  }'

# Create incidence type
curl -X POST http://localhost:8080/api/v1/incidence-types \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Vacaciones","category":"vacation","effect_type":"neutral","is_calculated":true,"calculation_method":"daily_rate"}'

# Create payroll period
curl -X POST http://localhost:8080/api/v1/payroll/periods \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "period_code":"2025-W49",
    "year":2025,
    "period_number":49,
    "frequency":"weekly",
    "period_type":"weekly",
    "start_date":"2025-12-01T00:00:00Z",
    "end_date":"2025-12-07T00:00:00Z",
    "payment_date":"2025-12-10T00:00:00Z"
  }'

# Calculate payroll
curl -X POST http://localhost:8080/api/v1/payroll/calculate \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"employee_id":"<EMPLOYEE_ID>","payroll_period_id":"<PERIOD_ID>"}'

# Get payslip PDF
curl -X GET "http://localhost:8080/api/v1/payroll/payslip/<PERIOD_ID>/<EMPLOYEE_ID>?format=pdf" \
  -H "Authorization: Bearer <TOKEN>" \
  -o payslip.pdf

# Get payslip XML (CFDI)
curl -X GET "http://localhost:8080/api/v1/payroll/payslip/<PERIOD_ID>/<EMPLOYEE_ID>?format=xml" \
  -H "Authorization: Bearer <TOKEN>"
```

### Announcement Commands

```bash
# Login and get token
TOKEN=$(curl -s -X POST "http://localhost/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"david.lu@iris.com","password":"Password123@"}' | jq -r '.access_token')

# Create company-wide announcement (HR/Admin only)
curl -X POST "http://localhost/api/v1/announcements" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Company Update",
    "message": "Important announcement for all employees",
    "scope": "ALL",
    "expires_in_days": 7
  }'

# Create team announcement (Supervisors/Managers - subordinates only)
curl -X POST "http://localhost/api/v1/announcements" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Meeting",
    "message": "Reminder: Team meeting tomorrow at 10am",
    "scope": "TEAM",
    "expires_in_days": 1
  }'

# List announcements (filtered by user's visibility)
curl "http://localhost/api/v1/announcements" \
  -H "Authorization: Bearer $TOKEN"

# Get unread count
curl "http://localhost/api/v1/announcements/unread-count" \
  -H "Authorization: Bearer $TOKEN"

# Mark announcement as read
curl -X POST "http://localhost/api/v1/announcements/<ANNOUNCEMENT_ID>/read" \
  -H "Authorization: Bearer $TOKEN"

# Delete announcement
curl -X DELETE "http://localhost/api/v1/announcements/<ANNOUNCEMENT_ID>" \
  -H "Authorization: Bearer $TOKEN"
```

### Absence Request Commands

```bash
# Create vacation request
curl -X POST "http://localhost/api/v1/absence-requests" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "<EMPLOYEE_ID>",
    "request_type": "VACATION",
    "start_date": "2025-12-23",
    "end_date": "2025-12-27",
    "total_days": 5,
    "reason": "Family vacation",
    "paid_days": 5
  }'

# Create shift change request
curl -X POST "http://localhost/api/v1/absence-requests" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "<EMPLOYEE_ID>",
    "request_type": "SHIFT_CHANGE",
    "start_date": "2025-12-15",
    "end_date": "2025-12-15",
    "total_days": 1,
    "reason": "Need to change to morning shift",
    "new_shift_id": "<SHIFT_ID>"
  }'

# Get my requests
curl "http://localhost/api/v1/absence-requests/my-requests" \
  -H "Authorization: Bearer $TOKEN"

# Get pending requests for supervisor approval
curl "http://localhost/api/v1/absence-requests/pending/supervisor" \
  -H "Authorization: Bearer $TOKEN"

# Get pending requests for manager approval
curl "http://localhost/api/v1/absence-requests/pending/manager" \
  -H "Authorization: Bearer $TOKEN"

# Get pending requests for HR approval
curl "http://localhost/api/v1/absence-requests/pending/hr" \
  -H "Authorization: Bearer $TOKEN"

# Approve request
curl -X POST "http://localhost/api/v1/absence-requests/<REQUEST_ID>/approve" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "APPROVED",
    "stage": "SUPERVISOR",
    "comments": "Approved"
  }'

# Decline request
curl -X POST "http://localhost/api/v1/absence-requests/<REQUEST_ID>/approve" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "DECLINED",
    "stage": "SUPERVISOR",
    "comments": "Cannot approve during busy period"
  }'

# Get pending counts (for dashboard badges)
curl "http://localhost/api/v1/absence-requests/counts" \
  -H "Authorization: Bearer $TOKEN"

# Check overlapping absences
curl "http://localhost/api/v1/absence-requests/overlapping?start_date=2025-12-23&end_date=2025-12-27" \
  -H "Authorization: Bearer $TOKEN"
```

### Shift Commands

```bash
# List all active shifts
curl "http://localhost/api/v1/shifts?active_only=true" \
  -H "Authorization: Bearer $TOKEN"

# Get specific shift
curl "http://localhost/api/v1/shifts/<SHIFT_ID>" \
  -H "Authorization: Bearer $TOKEN"
```

### HR Calendar Commands

```bash
# Get calendar events for a month
curl "http://localhost/api/v1/calendar/events?start_date=2025-12-01&end_date=2025-12-31" \
  -H "Authorization: Bearer $TOKEN"

# Get calendar events filtered by collar type
curl "http://localhost/api/v1/calendar/events?start_date=2025-12-01&end_date=2025-12-31&collar_types[]=white_collar" \
  -H "Authorization: Bearer $TOKEN"

# Get calendar events for specific employees
curl "http://localhost/api/v1/calendar/events?start_date=2025-12-01&end_date=2025-12-31&employee_ids[]=<EMP_ID_1>&employee_ids[]=<EMP_ID_2>" \
  -H "Authorization: Bearer $TOKEN"

# Get calendar events by event type
curl "http://localhost/api/v1/calendar/events?start_date=2025-12-01&end_date=2025-12-31&event_types[]=absence&event_types[]=shift_change" \
  -H "Authorization: Bearer $TOKEN"

# Get employees for calendar sidebar
curl "http://localhost/api/v1/calendar/employees" \
  -H "Authorization: Bearer $TOKEN"

# Get employees filtered by collar type
curl "http://localhost/api/v1/calendar/employees?collar_types[]=blue_collar" \
  -H "Authorization: Bearer $TOKEN"
```

---

Generated: 2025-12-10 (Updated with Announcements, Absence Requests, Shifts, and HR Calendar endpoints)
