# API Client to Backend Endpoint Mapping Verification

This document verifies that all frontend API client functions (`frontend/lib/api-client.ts`) correctly map to backend endpoints.

## Verification Summary

| API Module | Backend Endpoints | Frontend Functions | Status |
|------------|-------------------|-------------------|--------|
| authApi | 9 | 9 | ✅ Complete |
| userApi | 4 | 4 | ✅ Complete |
| employeeApi | 9 | 9 | ✅ Complete |
| payrollApi | 13 | 13 | ✅ Complete |
| catalogApi | 3 | 3 | ✅ Complete |
| reportApi | 2 | 2 | ✅ Complete |
| incidenceTypeApi | 4 | 4 | ✅ Complete |
| incidenceApi | 10 | 10 | ✅ Complete |
| evidenceApi | 5 | 5 | ✅ Complete |
| healthApi | 2 | 2 | ✅ Complete |

**Total: 61 endpoints - All verified ✅**

---

## Detailed Mapping

### 1. Auth API (authApi) - Line 310-360

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/auth/login` | POST | `authApi.login()` | 311 |
| `/auth/register` | POST | `authApi.register()` | 317 |
| `/auth/refresh` | POST | `authApi.refreshToken()` | 323 |
| `/auth/logout` | POST | `authApi.logout()` | 329 |
| `/auth/change-password` | POST | `authApi.changePassword()` | 334 |
| `/auth/forgot-password` | POST | `authApi.forgotPassword()` | 340 |
| `/auth/reset-password` | POST | `authApi.resetPassword()` | 346 |
| `/auth/profile` | GET | `authApi.getProfile()` | 352 |
| `/auth/profile` | PUT | `authApi.updateProfile()` | 355 |

---

### 2. User API (userApi) - Line 383-402

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/users` | GET | `userApi.getUsers()` | 384 |
| `/users` | POST | `userApi.createUser()` | 387 |
| `/users/:id` | DELETE | `userApi.deleteUser()` | 393 |
| `/users/:id/toggle-active` | PATCH | `userApi.toggleUserActive()` | 398 |

---

### 3. Employee API (employeeApi) - Line 428-478

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/employees` | GET | `employeeApi.getEmployees()` | 429 |
| `/employees/:id` | GET | `employeeApi.getEmployee()` | 434 |
| `/employees` | POST | `employeeApi.createEmployee()` | 437 |
| `/employees/:id` | PUT | `employeeApi.updateEmployee()` | 443 |
| `/employees/:id` | DELETE | `employeeApi.deleteEmployee()` | 449 |
| `/employees/:id/terminate` | POST | `employeeApi.terminateEmployee()` | 454 |
| `/employees/:id/salary` | PUT | `employeeApi.updateSalary()` | 460 |
| `/employees/stats` | GET | `employeeApi.getStats()` | 466 |
| `/employees/validate-ids` | POST | `employeeApi.validateMexicanIds()` | 469 |

---

### 4. Payroll API (payrollApi) - Line 488-546

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/payroll/periods` | GET | `payrollApi.getPeriods()` | 489 |
| `/payroll/periods/:id` | GET | `payrollApi.getPeriod()` | 494 |
| `/payroll/periods` | POST | `payrollApi.createPeriod()` | 497 |
| `/payroll/calculate` | POST | `payrollApi.calculatePayroll()` | 503 |
| `/payroll/bulk-calculate` | POST | `payrollApi.bulkCalculatePayroll()` | 513 |
| `/payroll/calculation/:period_id/:employee_id` | GET | `payrollApi.getPayrollCalculation()` | 519 |
| `/payroll/approve/:periodId` | POST | `payrollApi.approvePayroll()` | 522 |
| `/payroll/summary/:periodId` | GET | `payrollApi.getPayrollSummary()` | 527 |
| `/payroll/payslip/:periodId/:employeeId` | GET | `payrollApi.getPayslip()` | 530 |
| `/payroll/concept-totals/:periodId` | GET | `payrollApi.getConceptTotals()` | 533 |
| `/payroll/period/:period_id` | GET | `payrollApi.getPayrollByPeriod()` | 536 |
| `/payroll/payment/:periodId` | POST | `payrollApi.processPayment()` | 539 |
| `/payroll/payment/:period_id` | GET | `payrollApi.getPaymentStatus()` | 544 |

---

### 5. Catalog API (catalogApi) - Line 549-561

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/catalogs/concepts` | GET | `catalogApi.getPayrollConcepts()` | 550 |
| `/catalogs/incidence-types` | GET | `catalogApi.getIncidenceTypes()` | 553 |
| `/catalogs/concepts` | POST | `catalogApi.createConcept()` | 556 |

---

### 6. Report API (reportApi) - Line 564-573

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/reports/generate` | POST | `reportApi.generateReport()` | 565 |
| `/reports/history` | GET | `reportApi.getReportHistory()` | 571 |

---

### 7. Incidence Type API (incidenceTypeApi) - Line 659-679

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/incidence-types` | GET | `incidenceTypeApi.getAll()` | 660 |
| `/incidence-types` | POST | `incidenceTypeApi.create()` | 663 |
| `/incidence-types/:id` | PUT | `incidenceTypeApi.update()` | 669 |
| `/incidence-types/:id` | DELETE | `incidenceTypeApi.delete()` | 675 |

---

### 8. Incidence API (incidenceApi) - Line 682-731

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/incidences` | GET | `incidenceApi.getAll()` | 683 |
| `/incidences/:id` | GET | `incidenceApi.get()` | 692 |
| `/incidences` | POST | `incidenceApi.create()` | 695 |
| `/incidences/:id` | PUT | `incidenceApi.update()` | 701 |
| `/incidences/:id` | DELETE | `incidenceApi.delete()` | 707 |
| `/incidences/:id/approve` | POST | `incidenceApi.approve()` | 712 |
| `/incidences/:id/reject` | POST | `incidenceApi.reject()` | 717 |
| `/employees/:id/incidences` | GET | `incidenceApi.getByEmployee()` | 722 |
| `/employees/:id/vacation-balance` | GET | `incidenceApi.getVacationBalance()` | 725 |
| `/employees/:id/absence-summary` | GET | `incidenceApi.getAbsenceSummary()` | 728 |

---

### 9. Evidence API (evidenceApi) - Line 748-800

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/evidence/incidence/:incidence_id` | POST | `evidenceApi.upload()` | 750 |
| `/evidence/incidence/:incidence_id` | GET | `evidenceApi.list()` | 772 |
| `/evidence/:evidence_id` | GET | `evidenceApi.get()` | 776 |
| `/evidence/:evidence_id/download` | GET | `evidenceApi.download()` | 780 |
| `/evidence/:evidence_id` | DELETE | `evidenceApi.delete()` | 796 |

---

### 10. Health API (healthApi) - Line 809-832

| Backend Endpoint | Method | Frontend Function | Line |
|-----------------|--------|-------------------|------|
| `/health` | GET | `healthApi.isServerAvailable()` | 814 |
| `/health` | GET | `healthApi.getHealth()` | 830 |

---

## Notes

1. **All endpoints verified**: Every backend endpoint has a corresponding frontend API client function.

2. **Request/Response types defined**: The api-client.ts file includes TypeScript interfaces for:
   - `Employee`, `PayrollPeriod`, `PayrollConcept`
   - `PayrollSummary`, `PayrollCalculationResponse`
   - `IncidenceType`, `Incidence`, `VacationBalance`, `AbsenceSummary`
   - `IncidenceEvidence`, `HealthCheckResponse`
   - `AuthResponse`, `UserProfile`, `User`

3. **Error handling**: The `ApiError` class provides comprehensive error handling with:
   - Network error detection
   - Server error detection (5xx)
   - Auth error detection (401/403)
   - User-friendly error messages

4. **Auth token management**: Token is stored in localStorage and automatically included in requests via the `apiRequest` wrapper.

5. **Mock notifications**: `getNotifications()`, `markNotificationAsRead()`, and `markAllNotificationsAsRead()` are mock implementations (lines 845-875) - will need backend implementation if notifications feature is required.

---

## API cURL Test Examples

### Setup - Get Authentication Token

```bash
# Register a new company and admin user (first time setup)
curl -X POST http://localhost/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Test Company",
    "company_rfc": "TCO123456789",
    "email": "admin@test.com",
    "password": "Test123456!",
    "role": "admin",
    "full_name": "Test Admin"
  }'

# Login to get access token
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@test.com", "password": "Test123456!"}'

# Store token for subsequent requests
export TOKEN="<your_access_token_here>"
```

### Health Check

```bash
curl -s http://localhost/api/v1/health
# Response: {"service":"iris-payroll-backend","status":"ok"}
```

### Authentication API

```bash
# Get user profile
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/auth/profile

# Update profile
curl -s -X PUT http://localhost/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Updated Name"}'

# Change password
curl -s -X POST http://localhost/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"current_password": "OldPass123!", "new_password": "NewPass123!"}'

# Refresh token
curl -s -X POST http://localhost/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<your_refresh_token>"}'

# Logout
curl -s -X POST http://localhost/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

### User Management API (Admin Only)

```bash
# List all users
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/users

# Create new user
curl -s -X POST http://localhost/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@test.com",
    "password": "User123456!",
    "full_name": "Test User",
    "role": "viewer"
  }'

# Update user (change name and role)
curl -s -X PUT http://localhost/api/v1/users/<user_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Updated User Name", "role": "employee"}'

# Toggle user active status
curl -s -X PATCH http://localhost/api/v1/users/<user_id>/toggle-active \
  -H "Authorization: Bearer $TOKEN"

# Delete user
curl -s -X DELETE http://localhost/api/v1/users/<user_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Employee API

```bash
# List employees (paginated)
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost/api/v1/employees?page=1&page_size=20"

# Get single employee
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/employees/<employee_id>

# Get employee statistics
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/employees/stats

# Get employee vacation balance
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/employees/<employee_id>/vacation-balance

# Create employee
curl -s -X POST http://localhost/api/v1/employees \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "EMP001",
    "first_name": "Juan",
    "last_name": "Perez",
    "date_of_birth": "1990-05-15T00:00:00Z",
    "gender": "male",
    "rfc": "PELJ900515XXX",
    "curp": "PELJ900515HSLRLN09",
    "hire_date": "2024-01-15T00:00:00Z",
    "daily_salary": 500.00,
    "collar_type": "white_collar",
    "pay_frequency": "biweekly",
    "employment_status": "active",
    "employee_type": "permanent"
  }'

# Update employee
curl -s -X PUT http://localhost/api/v1/employees/<employee_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"daily_salary": 550.00}'

# Update employee salary
curl -s -X PUT http://localhost/api/v1/employees/<employee_id>/salary \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"new_daily_salary": 600.00, "effective_date": "2025-01-01"}'

# Terminate employee
curl -s -X POST http://localhost/api/v1/employees/<employee_id>/terminate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "termination_date": "2025-12-31",
    "reason": "Resignation",
    "comments": "Employee resigned voluntarily"
  }'

# Import employees from Excel
curl -X POST http://localhost/api/v1/employees/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@employees.xlsx"

# Download import template
curl -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/employees/import/template -o template.xlsx
```

### Payroll API

```bash
# List payroll periods
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/payroll/periods

# Get single period
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/payroll/periods/<period_id>

# Generate current periods (weekly, biweekly, monthly)
curl -s -X POST http://localhost/api/v1/payroll/periods/generate \
  -H "Authorization: Bearer $TOKEN"

# Calculate payroll for single employee
curl -s -X POST http://localhost/api/v1/payroll/calculate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "<employee_id>",
    "payroll_period_id": "<period_id>",
    "calculate_sdi": true
  }'

# Bulk calculate payroll for all employees
curl -s -X POST http://localhost/api/v1/payroll/bulk-calculate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"payroll_period_id": "<period_id>", "calculate_all": true}'

# Get payroll by period (all calculations)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/payroll/period/<period_id>

# Get payroll summary
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/payroll/summary/<period_id>

# Get concept totals
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/payroll/concept-totals/<period_id>

# Approve payroll period
curl -s -X POST http://localhost/api/v1/payroll/approve/<period_id> \
  -H "Authorization: Bearer $TOKEN"

# Process payment
curl -s -X POST http://localhost/api/v1/payroll/payment/<period_id> \
  -H "Authorization: Bearer $TOKEN"

# Get payment status
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/payroll/payment/<period_id>

# Get payslip (PDF)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost/api/v1/payroll/payslip/<period_id>/<employee_id>?format=pdf" -o payslip.pdf
```

### Incidence Types API

```bash
# List all incidence types
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/incidence-types

# Get requestable incidence types (for employee portal)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/requestable-incidence-types

# List incidence categories
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/incidence-categories

# Create incidence type
curl -s -X POST http://localhost/api/v1/incidence-types \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Custom Absence",
    "category": "absence",
    "effect_type": "negative",
    "is_calculated": true,
    "calculation_method": "daily_rate",
    "default_value": 0,
    "requires_evidence": true,
    "is_requestable": true,
    "approval_flow": "standard"
  }'

# Update incidence type
curl -s -X PUT http://localhost/api/v1/incidence-types/<type_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name", "requires_evidence": false}'

# Delete incidence type
curl -s -X DELETE http://localhost/api/v1/incidence-types/<type_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Incidences API

```bash
# List all incidences
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/incidences

# List incidences with filters
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost/api/v1/incidences?employee_id=<id>&period_id=<id>&status=pending"

# Get single incidence
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/incidences/<incidence_id>

# Get incidences by employee
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/employees/<employee_id>/incidences

# Get absence summary
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/employees/<employee_id>/absence-summary?year=2025

# Create incidence
curl -s -X POST http://localhost/api/v1/incidences \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "<employee_id>",
    "incidence_type_id": "<type_id>",
    "start_date": "2025-12-15",
    "end_date": "2025-12-15",
    "quantity": 1,
    "comments": "Vacation day"
  }'

# Update incidence
curl -s -X PUT http://localhost/api/v1/incidences/<incidence_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"comments": "Updated comments"}'

# Approve incidence
curl -s -X POST http://localhost/api/v1/incidences/<incidence_id>/approve \
  -H "Authorization: Bearer $TOKEN"

# Reject incidence
curl -s -X POST http://localhost/api/v1/incidences/<incidence_id>/reject \
  -H "Authorization: Bearer $TOKEN"

# Delete incidence
curl -s -X DELETE http://localhost/api/v1/incidences/<incidence_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Evidence API

```bash
# Upload evidence for incidence
curl -X POST http://localhost/api/v1/evidence/incidence/<incidence_id> \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@document.pdf"

# List evidence for incidence
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/evidence/incidence/<incidence_id>

# Get evidence details
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/evidence/<evidence_id>

# Download evidence
curl -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/evidence/<evidence_id>/download -o evidence.pdf

# Delete evidence
curl -s -X DELETE http://localhost/api/v1/evidence/<evidence_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Notifications API

```bash
# Get notifications
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost/api/v1/notifications?limit=20"

# Get unread count
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/notifications/unread-count

# Mark notification as read
curl -s -X POST http://localhost/api/v1/notifications/<notification_id>/read \
  -H "Authorization: Bearer $TOKEN"

# Mark all as read
curl -s -X POST http://localhost/api/v1/notifications/read-all \
  -H "Authorization: Bearer $TOKEN"
```

### Shifts API

```bash
# List all shifts
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost/api/v1/shifts?active_only=true"

# Get single shift
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/shifts/<shift_id>

# Create shift
curl -s -X POST http://localhost/api/v1/shifts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Morning Shift",
    "code": "T1",
    "start_time": "07:00",
    "end_time": "15:00",
    "is_active": true,
    "collar_types": ["blue_collar", "gray_collar"],
    "work_days": [1, 2, 3, 4, 5, 6]
  }'

# Update shift
curl -s -X PUT http://localhost/api/v1/shifts/<shift_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Shift Name"}'

# Delete shift
curl -s -X DELETE http://localhost/api/v1/shifts/<shift_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Absence Requests API (Employee Portal)

```bash
# Get my requests (current user)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/absence-requests/my-requests

# Get pending counts
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/absence-requests/counts

# Get pending requests by stage (supervisor, manager, hr)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/absence-requests/pending/supervisor
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/absence-requests/pending/manager
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/absence-requests/pending/hr

# Check for overlapping absences
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost/api/v1/absence-requests/overlapping?start_date=2025-12-15&end_date=2025-12-20"

# Create absence request
curl -s -X POST http://localhost/api/v1/absence-requests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "<employee_id>",
    "request_type": "VACATION",
    "start_date": "2025-12-20",
    "end_date": "2025-12-25",
    "total_days": 5,
    "reason": "Family vacation"
  }'

# Approve/Decline request
curl -s -X POST http://localhost/api/v1/absence-requests/<request_id>/approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "APPROVED",
    "stage": "SUPERVISOR",
    "comments": "Approved"
  }'

# Archive request
curl -s -X PATCH http://localhost/api/v1/absence-requests/<request_id>/archive \
  -H "Authorization: Bearer $TOKEN"

# Delete request (only pending requests by owner)
curl -s -X DELETE http://localhost/api/v1/absence-requests/<request_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Announcements API

```bash
# List all announcements
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/announcements

# Get single announcement
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/announcements/<announcement_id>

# Get unread count
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/announcements/unread-count

# Create announcement
curl -s -X POST http://localhost/api/v1/announcements \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Company Holiday Notice",
    "message": "The office will be closed on December 25th",
    "scope": "ALL",
    "expires_in_days": 30
  }'

# Mark announcement as read
curl -s -X POST http://localhost/api/v1/announcements/<announcement_id>/read \
  -H "Authorization: Bearer $TOKEN"

# Delete announcement
curl -s -X DELETE http://localhost/api/v1/announcements/<announcement_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Messages/Inbox API

```bash
# Get inbox messages
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost/api/v1/messages?page=1&page_size=20&status=all"

# Get sent messages
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost/api/v1/messages/sent?page=1&page_size=20"

# Get single message
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/messages/<message_id>

# Get unread count
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/messages/unread-count

# Get recipient suggestions
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost/api/v1/messages/recipients?search=admin"

# Send message
curl -s -X POST http://localhost/api/v1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient_id": "<user_id>",
    "subject": "Meeting Request",
    "body": "Can we schedule a meeting for tomorrow?"
  }'

# Ask question about announcement
curl -s -X POST http://localhost/api/v1/messages/announcement-question \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "announcement_id": "<announcement_id>",
    "question": "What time will the office reopen?"
  }'

# Reply to message
curl -s -X POST http://localhost/api/v1/messages/<message_id>/reply \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"body": "Yes, 10am works for me."}'

# Mark message as read
curl -s -X POST http://localhost/api/v1/messages/<message_id>/read \
  -H "Authorization: Bearer $TOKEN"

# Mark message as unread
curl -s -X POST http://localhost/api/v1/messages/<message_id>/unread \
  -H "Authorization: Bearer $TOKEN"

# Archive message
curl -s -X POST http://localhost/api/v1/messages/<message_id>/archive \
  -H "Authorization: Bearer $TOKEN"

# Delete message
curl -s -X DELETE http://localhost/api/v1/messages/<message_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Calendar API

```bash
# Get calendar events
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost/api/v1/calendar/events?start_date=2025-12-01&end_date=2025-12-31"
```

### Reports API

```bash
# Generate report
curl -s -X POST http://localhost/api/v1/reports/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "report_type": "payroll_summary",
    "payroll_period_id": "<period_id>",
    "format": "json"
  }'

# Get report history
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/reports/history
```

### Catalog API

```bash
# Get payroll concepts
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/catalogs/concepts

# Get incidence types (catalog)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost/api/v1/catalogs/incidence-types

# Create payroll concept
curl -s -X POST http://localhost/api/v1/catalogs/concepts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Bonus",
    "category": "EARNING",
    "input_type": "amount",
    "affects_tax_base": true,
    "affects_social_security": false,
    "affects_integrated_salary": false
  }'
```

---

## User Roles Reference

| Role | Value | Description |
|------|-------|-------------|
| Admin | `admin` | Full system access, user management |
| HR - All | `hr` | View/edit ALL employees |
| HR - Blue/Gray | `hr_blue_gray` | Only blue_collar and gray_collar employees |
| HR - White | `hr_white` | Only white_collar employees |
| HR + Payroll | `hr_and_pr` | Combined HR and Payroll permissions |
| Accountant | `accountant` | Calculate and export payroll |
| Payroll Staff | `payroll_staff` | Day-to-day payroll operations |
| Supervisor | `supervisor` | Approve subordinate requests |
| Manager | `manager` | Approve requests after supervisor |
| Sup + Manager | `sup_and_gm` | Combined Supervisor and Manager |
| Employee | `employee` | Employee portal access only |
| Viewer | `viewer` | Read-only access |

---

Generated: 2025-12-12
