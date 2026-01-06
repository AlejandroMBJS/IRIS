# IRIS HR & Payroll System - Complete Production Sprint Plan

**Version**: 2.0
**Created**: January 2, 2026
**Target**: Full Production-Ready Enterprise HR/Payroll System
**Benchmark**: Odoo HR, Oracle HCM, SAP SuccessFactors, Revo en la Nube

---

## Executive Summary

This document contains the complete sprint plan to transform IRIS from a Mexican Payroll system into a **full-featured Enterprise HR & Payroll Management System** comparable to industry leaders.

### Total Effort Summary

| Category | Sprints | Total Hours | Duration |
|----------|---------|-------------|----------|
| Security Hardening | 1 | 40h | 1 week |
| Code Refactoring | 1 | 74h | 1 week |
| Testing Foundation | 1 | 84h | 1 week |
| New Modules | 12 | 1,160h | 12 weeks |
| Mobile & Accessibility | 1 | 116h | 1 week |
| Performance | 1 | 88h | 1 week |
| Final QA & Launch | 1 | 132h | 1 week |
| **TOTAL** | **18** | **1,694h** | **18 weeks** |

### Industry Comparison Coverage

| Feature | Odoo | Oracle HCM | SAP SF | IRIS (Current) | IRIS (After) |
|---------|------|------------|--------|----------------|--------------|
| Core HR | Yes | Yes | Yes | 70% | 100% |
| Payroll | Yes | Yes | Yes | 95% | 100% |
| Recruitment | Yes | Yes | Yes | 0% | 100% |
| Onboarding | Yes | Yes | Yes | 0% | 100% |
| Training/LMS | Yes | Yes | Yes | 0% | 100% |
| Performance | Yes | Yes | Yes | 0% | 100% |
| Benefits | Yes | Yes | Yes | 10% | 100% |
| Time Tracking | Yes | Yes | Yes | 40% | 100% |
| Expenses | Yes | Yes | Yes | 0% | 100% |
| Documents | Yes | Yes | Yes | 20% | 100% |
| Analytics | Yes | Yes | Yes | 30% | 100% |
| Self-Service | Yes | Yes | Yes | 10% | 100% |
| Mobile | Yes | Yes | Yes | 20% | 100% |

---

# SPRINT 1: Security Hardening

**Duration**: 1 Week (40 hours)
**Priority**: P0 - CRITICAL
**Goal**: Fix all security vulnerabilities before adding new features

---

## [S1-001] Move JWT tokens from localStorage to httpOnly cookies

**Priority**: URGENT | **Effort**: 8h | **Tags**: security, authentication, P0

### Problem
Current implementation stores JWT tokens in localStorage, which is vulnerable to XSS attacks. Any JavaScript can access localStorage and steal tokens.

### Solution Pseudocode

```
// Backend: Cookie Utility Functions
MODULE CookieUtils

  CONSTANT ACCESS_COOKIE_NAME = "iris_access_token"
  CONSTANT REFRESH_COOKIE_NAME = "iris_refresh_token"
  CONSTANT COOKIE_MAX_AGE = 15 * 60          // 15 minutes for access
  CONSTANT REFRESH_MAX_AGE = 7 * 24 * 60 * 60 // 7 days for refresh

  FUNCTION SetAccessTokenCookie(response, token)
    cookie = NEW Cookie()
    cookie.Name = ACCESS_COOKIE_NAME
    cookie.Value = token
    cookie.HttpOnly = TRUE           // Cannot be accessed by JavaScript
    cookie.Secure = TRUE             // Only sent over HTTPS
    cookie.SameSite = "Strict"       // Prevents CSRF
    cookie.Path = "/"
    cookie.MaxAge = COOKIE_MAX_AGE
    response.SetCookie(cookie)
  END FUNCTION

  FUNCTION SetRefreshTokenCookie(response, token)
    cookie = NEW Cookie()
    cookie.Name = REFRESH_COOKIE_NAME
    cookie.Value = token
    cookie.HttpOnly = TRUE
    cookie.Secure = TRUE
    cookie.SameSite = "Strict"
    cookie.Path = "/api/auth/refresh"  // Only sent to refresh endpoint
    cookie.MaxAge = REFRESH_MAX_AGE
    response.SetCookie(cookie)
  END FUNCTION

  FUNCTION ClearAuthCookies(response)
    // Set cookies with immediate expiration
    response.SetCookie(Cookie{Name: ACCESS_COOKIE_NAME, MaxAge: -1})
    response.SetCookie(Cookie{Name: REFRESH_COOKIE_NAME, MaxAge: -1})
  END FUNCTION

  FUNCTION GetAccessToken(request) RETURNS string
    cookie = request.GetCookie(ACCESS_COOKIE_NAME)
    IF cookie IS NULL THEN
      RETURN ""
    END IF
    RETURN cookie.Value
  END FUNCTION

END MODULE

// Backend: Update Auth Handler Login
FUNCTION HandleLogin(request, response)
  credentials = ParseJSON(request.Body)

  // Validate credentials
  user = authService.ValidateCredentials(credentials.email, credentials.password)
  IF user IS NULL THEN
    RETURN ErrorResponse(401, "Invalid credentials")
  END IF

  // Generate tokens
  accessToken = jwtService.GenerateAccessToken(user)
  refreshToken = jwtService.GenerateRefreshToken(user)

  // Set httpOnly cookies instead of returning in body
  CookieUtils.SetAccessTokenCookie(response, accessToken)
  CookieUtils.SetRefreshTokenCookie(response, refreshToken)

  // Return user data (without tokens)
  RETURN JSONResponse({
    "user": {
      "id": user.ID,
      "email": user.Email,
      "role": user.Role,
      "fullName": user.FullName
    }
  })
END FUNCTION

// Backend: Update Auth Middleware
FUNCTION AuthMiddleware(request, response, next)
  // Read token from cookie instead of Authorization header
  token = CookieUtils.GetAccessToken(request)

  IF token IS EMPTY THEN
    RETURN ErrorResponse(401, "Authentication required")
  END IF

  claims = jwtService.ValidateToken(token)
  IF claims IS NULL THEN
    RETURN ErrorResponse(401, "Invalid or expired token")
  END IF

  // Set user context
  request.Context.Set("userID", claims.UserID)
  request.Context.Set("userRole", claims.Role)

  next()
END FUNCTION

// Frontend: Update API Client
CLASS ApiClient

  // No longer need to store/retrieve tokens
  // Cookies are sent automatically with credentials: 'include'

  FUNCTION fetch(url, options)
    defaultOptions = {
      credentials: 'include',  // Send cookies with request
      headers: {
        'Content-Type': 'application/json'
      }
    }

    mergedOptions = MERGE(defaultOptions, options)
    response = AWAIT window.fetch(url, mergedOptions)

    IF response.status == 401 THEN
      // Try to refresh token
      refreshed = AWAIT this.refreshToken()
      IF refreshed THEN
        RETURN AWAIT window.fetch(url, mergedOptions)
      ELSE
        this.redirectToLogin()
      END IF
    END IF

    RETURN response
  END FUNCTION

  FUNCTION refreshToken()
    response = AWAIT fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include'
    })
    RETURN response.ok
  END FUNCTION

END CLASS
```

### Acceptance Criteria
- [ ] Tokens stored in httpOnly cookies, not localStorage
- [ ] SameSite=Strict prevents CSRF
- [ ] Secure flag ensures HTTPS only
- [ ] Frontend cannot access tokens via JavaScript
- [ ] Refresh token only sent to /api/auth/refresh
- [ ] Logout clears all auth cookies

---

## [S1-002] Implement CSRF protection middleware

**Priority**: URGENT | **Effort**: 6h | **Tags**: security, middleware, P0

### Problem
Without CSRF protection, malicious sites can submit requests on behalf of authenticated users.

### Solution Pseudocode

```
// Backend: CSRF Token Service
MODULE CSRFService

  CONSTANT CSRF_COOKIE_NAME = "iris_csrf_token"
  CONSTANT CSRF_HEADER_NAME = "X-CSRF-Token"
  CONSTANT TOKEN_LENGTH = 32

  FUNCTION GenerateToken() RETURNS string
    // Generate cryptographically secure random token
    randomBytes = crypto.RandomBytes(TOKEN_LENGTH)
    RETURN Base64Encode(randomBytes)
  END FUNCTION

  FUNCTION SetCSRFCookie(response)
    token = GenerateToken()

    cookie = NEW Cookie()
    cookie.Name = CSRF_COOKIE_NAME
    cookie.Value = token
    cookie.HttpOnly = FALSE    // JavaScript needs to read this
    cookie.Secure = TRUE
    cookie.SameSite = "Strict"
    cookie.Path = "/"

    response.SetCookie(cookie)
    RETURN token
  END FUNCTION

  FUNCTION ValidateToken(request) RETURNS boolean
    // Get token from cookie
    cookieToken = request.GetCookie(CSRF_COOKIE_NAME)
    IF cookieToken IS NULL THEN
      RETURN FALSE
    END IF

    // Get token from header (frontend must send this)
    headerToken = request.GetHeader(CSRF_HEADER_NAME)
    IF headerToken IS EMPTY THEN
      RETURN FALSE
    END IF

    // Compare using constant-time comparison to prevent timing attacks
    RETURN crypto.ConstantTimeCompare(cookieToken.Value, headerToken)
  END FUNCTION

END MODULE

// Backend: CSRF Middleware
FUNCTION CSRFMiddleware(request, response, next)
  // Skip CSRF for safe methods
  IF request.Method IN ["GET", "HEAD", "OPTIONS"] THEN
    next()
    RETURN
  END IF

  // Skip CSRF for specific public endpoints
  IF request.Path IN ["/api/auth/login", "/api/auth/register"] THEN
    next()
    RETURN
  END IF

  // Validate CSRF token for state-changing requests
  IF NOT CSRFService.ValidateToken(request) THEN
    RETURN ErrorResponse(403, "Invalid CSRF token")
  END IF

  next()
END FUNCTION

// Backend: Login endpoint sets CSRF token
FUNCTION HandleLogin(request, response)
  // ... existing login logic ...

  // Set CSRF cookie after successful login
  CSRFService.SetCSRFCookie(response)

  RETURN JSONResponse(userData)
END FUNCTION

// Frontend: CSRF Token Helper
CLASS CSRFHelper

  FUNCTION getToken() RETURNS string
    // Read CSRF token from cookie
    cookies = document.cookie.split(';')
    FOR EACH cookie IN cookies
      parts = cookie.trim().split('=')
      IF parts[0] == 'iris_csrf_token' THEN
        RETURN parts[1]
      END IF
    END FOR
    RETURN ''
  END FUNCTION

END CLASS

// Frontend: Update API Client
CLASS ApiClient

  FUNCTION fetch(url, options)
    defaultOptions = {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': CSRFHelper.getToken()  // Add CSRF token
      }
    }

    mergedOptions = MERGE(defaultOptions, options)
    RETURN AWAIT window.fetch(url, mergedOptions)
  END FUNCTION

END CLASS
```

### Acceptance Criteria
- [ ] CSRF token generated on login
- [ ] All POST/PUT/DELETE requests require valid CSRF token
- [ ] Token validated using constant-time comparison
- [ ] 403 Forbidden returned for invalid/missing token
- [ ] Frontend automatically includes CSRF header

---

## [S1-003] Add Content Security Policy headers

**Priority**: HIGH | **Effort**: 4h | **Tags**: security, headers, P0

### Solution Pseudocode

```
// Backend: CSP Middleware
MODULE CSPMiddleware

  FUNCTION GetCSPPolicy() RETURNS string
    policy = []

    // Default: only allow same origin
    policy.ADD("default-src 'self'")

    // Scripts: self + specific CDNs if needed
    policy.ADD("script-src 'self'")

    // Styles: self + inline for Tailwind (use nonce in production)
    policy.ADD("style-src 'self' 'unsafe-inline'")

    // Images: self + data URIs for inline images
    policy.ADD("img-src 'self' data: blob:")

    // Fonts: self + Google Fonts if used
    policy.ADD("font-src 'self'")

    // Connect: self + API endpoints
    policy.ADD("connect-src 'self' " + config.APIBaseURL)

    // Frames: none (prevent clickjacking)
    policy.ADD("frame-ancestors 'none'")

    // Form actions: only self
    policy.ADD("form-action 'self'")

    // Base URI: only self
    policy.ADD("base-uri 'self'")

    RETURN policy.JOIN("; ")
  END FUNCTION

  FUNCTION Middleware(request, response, next)
    response.SetHeader("Content-Security-Policy", GetCSPPolicy())
    next()
  END FUNCTION

END MODULE

// Next.js: next.config.js CSP Configuration
CONST CSPConfig = {
  headers: [
    {
      source: '/(.*)',
      headers: [
        {
          key: 'Content-Security-Policy',
          value: [
            "default-src 'self'",
            "script-src 'self'",
            "style-src 'self' 'unsafe-inline'",
            "img-src 'self' data: blob:",
            "font-src 'self'",
            "connect-src 'self' http://localhost:8080",
            "frame-ancestors 'none'",
            "form-action 'self'",
            "base-uri 'self'"
          ].join('; ')
        }
      ]
    }
  ]
}
```

### Acceptance Criteria
- [ ] CSP header present on all responses
- [ ] XSS attacks prevented by script-src
- [ ] Clickjacking prevented by frame-ancestors
- [ ] No inline script violations
- [ ] CSP report endpoint configured for violations

---

## [S1-004] Implement rate limiting for authentication endpoints

**Priority**: HIGH | **Effort**: 6h | **Tags**: security, authentication, rate-limiting

### Solution Pseudocode

```
// Backend: Rate Limiter with Sliding Window
MODULE RateLimiter

  // In-memory store (use Redis in production)
  STORE requestCounts = MAP<string, RateLimitEntry>

  STRUCT RateLimitEntry
    timestamps: LIST<timestamp>
    blockedUntil: timestamp
  END STRUCT

  STRUCT RateLimitConfig
    windowSeconds: integer     // Time window
    maxRequests: integer       // Max requests per window
    blockDurationSeconds: integer // Block duration after limit exceeded
  END STRUCT

  // Different limits for different endpoints
  CONFIGS = {
    "/api/auth/login": { windowSeconds: 60, maxRequests: 5, blockDurationSeconds: 300 },
    "/api/auth/register": { windowSeconds: 3600, maxRequests: 3, blockDurationSeconds: 3600 },
    "/api/auth/forgot-password": { windowSeconds: 3600, maxRequests: 3, blockDurationSeconds: 3600 }
  }

  FUNCTION GetClientIdentifier(request) RETURNS string
    // Use IP + User-Agent for more accurate tracking
    ip = request.GetHeader("X-Forwarded-For") OR request.RemoteIP
    userAgent = request.GetHeader("User-Agent")
    RETURN Hash(ip + userAgent)
  END FUNCTION

  FUNCTION CleanOldEntries(entry, windowSeconds)
    cutoff = NOW() - windowSeconds
    entry.timestamps = entry.timestamps.FILTER(t => t > cutoff)
  END FUNCTION

  FUNCTION CheckRateLimit(request) RETURNS (allowed: boolean, retryAfter: integer)
    path = request.Path
    config = CONFIGS[path]

    IF config IS NULL THEN
      RETURN (TRUE, 0)  // No rate limit for this endpoint
    END IF

    clientKey = path + ":" + GetClientIdentifier(request)
    entry = requestCounts.GET(clientKey) OR NEW RateLimitEntry()

    // Check if currently blocked
    IF entry.blockedUntil > NOW() THEN
      retryAfter = entry.blockedUntil - NOW()
      RETURN (FALSE, retryAfter)
    END IF

    // Clean old entries
    CleanOldEntries(entry, config.windowSeconds)

    // Check if over limit
    IF entry.timestamps.LENGTH >= config.maxRequests THEN
      entry.blockedUntil = NOW() + config.blockDurationSeconds
      requestCounts.SET(clientKey, entry)
      RETURN (FALSE, config.blockDurationSeconds)
    END IF

    // Add current request
    entry.timestamps.ADD(NOW())
    requestCounts.SET(clientKey, entry)

    remaining = config.maxRequests - entry.timestamps.LENGTH
    RETURN (TRUE, 0, remaining)
  END FUNCTION

END MODULE

// Backend: Rate Limit Middleware
FUNCTION RateLimitMiddleware(request, response, next)
  allowed, retryAfter, remaining = RateLimiter.CheckRateLimit(request)

  // Add rate limit headers
  response.SetHeader("X-RateLimit-Remaining", remaining)

  IF NOT allowed THEN
    response.SetHeader("Retry-After", retryAfter)
    RETURN ErrorResponse(429, "Too many requests. Please try again later.")
  END IF

  next()
END FUNCTION
```

### Acceptance Criteria
- [ ] Login limited to 5 attempts per minute
- [ ] Registration limited to 3 per hour
- [ ] Password reset limited to 3 per hour
- [ ] 429 response with Retry-After header when exceeded
- [ ] Block persists for configured duration
- [ ] X-RateLimit-Remaining header on responses

---

## [S1-005] Add input sanitization middleware

**Priority**: HIGH | **Effort**: 6h | **Tags**: security, validation, middleware

### Solution Pseudocode

```
// Backend: Input Sanitizer
MODULE InputSanitizer

  // Dangerous patterns to detect
  SQL_PATTERNS = [
    "'; DROP TABLE",
    "1=1", "1 = 1",
    "UNION SELECT",
    "OR 1=1",
    "--"
  ]

  XSS_PATTERNS = [
    "<script",
    "javascript:",
    "onerror=",
    "onclick=",
    "onload="
  ]

  FUNCTION SanitizeString(input) RETURNS string
    IF input IS NULL THEN
      RETURN ""
    END IF

    result = input

    // HTML entity encoding for XSS prevention
    result = result.REPLACE("&", "&amp;")
    result = result.REPLACE("<", "&lt;")
    result = result.REPLACE(">", "&gt;")
    result = result.REPLACE("\"", "&quot;")
    result = result.REPLACE("'", "&#x27;")

    // Trim whitespace
    result = result.TRIM()

    RETURN result
  END FUNCTION

  FUNCTION DetectMaliciousInput(input) RETURNS boolean
    upperInput = input.UPPERCASE()

    FOR EACH pattern IN SQL_PATTERNS
      IF upperInput.CONTAINS(pattern.UPPERCASE()) THEN
        RETURN TRUE
      END IF
    END FOR

    FOR EACH pattern IN XSS_PATTERNS
      IF input.LOWERCASE().CONTAINS(pattern.LOWERCASE()) THEN
        RETURN TRUE
      END IF
    END FOR

    RETURN FALSE
  END FUNCTION

  FUNCTION SanitizeObject(obj) RETURNS object
    IF obj IS NULL THEN
      RETURN NULL
    END IF

    IF obj IS string THEN
      RETURN SanitizeString(obj)
    END IF

    IF obj IS array THEN
      RETURN obj.MAP(item => SanitizeObject(item))
    END IF

    IF obj IS object THEN
      result = {}
      FOR EACH key, value IN obj
        sanitizedKey = SanitizeString(key)
        sanitizedValue = SanitizeObject(value)
        result[sanitizedKey] = sanitizedValue
      END FOR
      RETURN result
    END IF

    // Numbers, booleans, etc. pass through
    RETURN obj
  END FUNCTION

END MODULE

// Backend: Sanitization Middleware
FUNCTION SanitizationMiddleware(request, response, next)
  // Check query parameters for malicious content
  FOR EACH key, value IN request.QueryParams
    IF InputSanitizer.DetectMaliciousInput(value) THEN
      LogSecurityEvent("Malicious query param detected", request)
      RETURN ErrorResponse(400, "Invalid input detected")
    END IF
  END FOR

  // Sanitize request body
  IF request.Body IS NOT EMPTY THEN
    bodyJSON = ParseJSON(request.Body)

    // Check for malicious content
    bodyString = JSON.Stringify(bodyJSON)
    IF InputSanitizer.DetectMaliciousInput(bodyString) THEN
      LogSecurityEvent("Malicious body content detected", request)
      RETURN ErrorResponse(400, "Invalid input detected")
    END IF

    // Sanitize and replace body
    sanitizedBody = InputSanitizer.SanitizeObject(bodyJSON)
    request.Body = JSON.Stringify(sanitizedBody)
  END IF

  next()
END FUNCTION
```

### Acceptance Criteria
- [ ] All string inputs HTML-encoded
- [ ] SQL injection patterns detected and blocked
- [ ] XSS patterns detected and blocked
- [ ] Security events logged for analysis
- [ ] Nested objects recursively sanitized

---

## [S1-006] Enforce strong JWT secret configuration

**Priority**: MEDIUM | **Effort**: 3h | **Tags**: security, authentication, configuration

### Solution Pseudocode

```
// Backend: JWT Secret Validator
MODULE JWTSecretValidator

  CONSTANT MIN_SECRET_LENGTH = 32    // 256 bits minimum
  CONSTANT MIN_ENTROPY_BITS = 128

  FUNCTION CalculateEntropy(secret) RETURNS float
    // Shannon entropy calculation
    charCounts = MAP<char, int>
    FOR EACH char IN secret
      charCounts[char] = (charCounts[char] OR 0) + 1
    END FOR

    entropy = 0
    length = secret.LENGTH
    FOR EACH count IN charCounts.VALUES
      probability = count / length
      entropy = entropy - (probability * LOG2(probability))
    END FOR

    RETURN entropy * length  // Total bits of entropy
  END FUNCTION

  FUNCTION ValidateSecret(secret) RETURNS (valid: boolean, errors: list)
    errors = []

    IF secret IS EMPTY THEN
      errors.ADD("JWT secret cannot be empty")
      RETURN (FALSE, errors)
    END IF

    IF secret.LENGTH < MIN_SECRET_LENGTH THEN
      errors.ADD("JWT secret must be at least " + MIN_SECRET_LENGTH + " characters")
    END IF

    entropy = CalculateEntropy(secret)
    IF entropy < MIN_ENTROPY_BITS THEN
      errors.ADD("JWT secret has insufficient entropy. Use a randomly generated secret.")
    END IF

    // Check for common weak secrets
    weakSecrets = ["secret", "password", "jwt_secret", "changeme", "12345"]
    IF secret.LOWERCASE() IN weakSecrets THEN
      errors.ADD("JWT secret is too common. Use a unique, random secret.")
    END IF

    RETURN (errors.LENGTH == 0, errors)
  END FUNCTION

  FUNCTION GenerateSecureSecret() RETURNS string
    // Generate 256-bit cryptographically secure random secret
    randomBytes = crypto.RandomBytes(32)
    RETURN Base64Encode(randomBytes)
  END FUNCTION

END MODULE

// Backend: Application Startup Validation
FUNCTION ValidateConfiguration()
  jwtSecret = config.Get("JWT_SECRET")

  valid, errors = JWTSecretValidator.ValidateSecret(jwtSecret)

  IF NOT valid THEN
    IF config.IsProduction() THEN
      // Fatal error in production
      FOR EACH error IN errors
        Logger.Error("JWT Configuration Error: " + error)
      END FOR
      PANIC("Invalid JWT configuration. Cannot start in production.")
    ELSE
      // Warning in development
      FOR EACH error IN errors
        Logger.Warn("JWT Configuration Warning: " + error)
      END FOR
      Logger.Warn("Suggested secure secret: " + JWTSecretValidator.GenerateSecureSecret())
    END IF
  END IF
END FUNCTION
```

### Acceptance Criteria
- [ ] Minimum 256-bit secret enforced
- [ ] Entropy check prevents weak secrets
- [ ] Production fails to start with weak secret
- [ ] Development shows warnings
- [ ] Helper generates secure secrets

---

## [S1-007] Add security headers middleware

**Priority**: MEDIUM | **Effort**: 3h | **Tags**: security, headers, middleware

### Solution Pseudocode

```
// Backend: Security Headers Middleware
MODULE SecurityHeadersMiddleware

  FUNCTION Apply(request, response, next)
    // Prevent MIME type sniffing
    response.SetHeader("X-Content-Type-Options", "nosniff")

    // Prevent clickjacking (backup for CSP frame-ancestors)
    response.SetHeader("X-Frame-Options", "DENY")

    // Enable XSS filter in older browsers
    response.SetHeader("X-XSS-Protection", "1; mode=block")

    // Enforce HTTPS (1 year, include subdomains, preload)
    IF config.IsProduction() THEN
      response.SetHeader(
        "Strict-Transport-Security",
        "max-age=31536000; includeSubDomains; preload"
      )
    END IF

    // Control referrer information
    response.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")

    // Permissions Policy (formerly Feature-Policy)
    response.SetHeader(
      "Permissions-Policy",
      "geolocation=(self), microphone=(), camera=(), payment=()"
    )

    // Prevent browser from caching sensitive pages
    IF request.Path.STARTS_WITH("/api/") THEN
      response.SetHeader("Cache-Control", "no-store, no-cache, must-revalidate")
      response.SetHeader("Pragma", "no-cache")
    END IF

    next()
  END FUNCTION

END MODULE
```

### Acceptance Criteria
- [ ] X-Content-Type-Options: nosniff
- [ ] X-Frame-Options: DENY
- [ ] HSTS header in production
- [ ] Referrer-Policy set
- [ ] API responses not cached

---

## [S1-008] Implement password strength validation

**Priority**: MEDIUM | **Effort**: 4h | **Tags**: security, authentication, validation

### Solution Pseudocode

```
// Backend: Password Validator
MODULE PasswordValidator

  STRUCT ValidationResult
    valid: boolean
    score: integer        // 0-4 strength score
    feedback: list<string>
    suggestions: list<string>
  END STRUCT

  FUNCTION Validate(password) RETURNS ValidationResult
    result = NEW ValidationResult()
    result.feedback = []
    result.suggestions = []
    result.score = 0

    // Minimum length check
    IF password.LENGTH < 12 THEN
      result.feedback.ADD("Password must be at least 12 characters")
    ELSE
      result.score = result.score + 1
    END IF

    // Uppercase check
    IF NOT password.MATCHES(/[A-Z]/) THEN
      result.feedback.ADD("Include at least one uppercase letter")
    ELSE
      result.score = result.score + 1
    END IF

    // Lowercase check
    IF NOT password.MATCHES(/[a-z]/) THEN
      result.feedback.ADD("Include at least one lowercase letter")
    ELSE
      result.score = result.score + 1
    END IF

    // Number check
    IF NOT password.MATCHES(/[0-9]/) THEN
      result.feedback.ADD("Include at least one number")
    ELSE
      result.score = result.score + 1
    END IF

    // Special character check
    IF NOT password.MATCHES(/[!@#$%^&*(),.?":{}|<>]/) THEN
      result.feedback.ADD("Include at least one special character")
    ELSE
      result.score = result.score + 1
    END IF

    // Common password check
    IF IsCommonPassword(password) THEN
      result.feedback.ADD("This password is too common")
      result.score = 0
    END IF

    // Check for personal info patterns
    IF ContainsSequentialChars(password) THEN
      result.suggestions.ADD("Avoid sequential characters like '123' or 'abc'")
    END IF

    IF ContainsRepeatedChars(password) THEN
      result.suggestions.ADD("Avoid repeated characters like 'aaa' or '111'")
    END IF

    result.valid = result.feedback.LENGTH == 0

    RETURN result
  END FUNCTION

  FUNCTION CheckHaveIBeenPwned(password) RETURNS boolean
    // Check password against HaveIBeenPwned API using k-anonymity
    hash = SHA1(password)
    prefix = hash.SUBSTRING(0, 5)
    suffix = hash.SUBSTRING(5)

    response = HTTP.GET("https://api.pwnedpasswords.com/range/" + prefix)

    FOR EACH line IN response.SPLIT("\n")
      parts = line.SPLIT(":")
      IF parts[0].UPPERCASE() == suffix.UPPERCASE() THEN
        RETURN TRUE  // Password has been breached
      END IF
    END FOR

    RETURN FALSE
  END FUNCTION

  FUNCTION IsCommonPassword(password) RETURNS boolean
    commonPasswords = LoadCommonPasswordsList()  // Top 10,000 common passwords
    RETURN password.LOWERCASE() IN commonPasswords
  END FUNCTION

END MODULE

// Backend: Registration/Password Change Handler
FUNCTION HandlePasswordChange(request, response)
  newPassword = request.Body.newPassword

  // Validate password strength
  validation = PasswordValidator.Validate(newPassword)
  IF NOT validation.valid THEN
    RETURN ErrorResponse(400, {
      "message": "Password does not meet requirements",
      "feedback": validation.feedback,
      "suggestions": validation.suggestions
    })
  END IF

  // Check if password has been breached
  IF PasswordValidator.CheckHaveIBeenPwned(newPassword) THEN
    RETURN ErrorResponse(400, {
      "message": "This password has been found in data breaches. Please choose a different password."
    })
  END IF

  // Continue with password change...
END FUNCTION
```

### Acceptance Criteria
- [ ] Minimum 12 character requirement
- [ ] Requires uppercase, lowercase, number, special char
- [ ] Common passwords rejected
- [ ] HaveIBeenPwned integration
- [ ] Clear feedback messages
- [ ] Password strength score displayed

---

# SPRINT 2: Code Refactoring

**Duration**: 1 Week (74 hours)
**Priority**: HIGH
**Goal**: Clean up technical debt before adding new features

---

## [S2-001] Refactor IncidencesPage component (1453 lines)

**Priority**: HIGH | **Effort**: 16h | **Tags**: refactor, frontend, components

### Problem
The IncidencesPage component is 1453 lines, violating single responsibility principle and making maintenance difficult.

### Solution Pseudocode

```
// BEFORE: Monolithic Component (1453 lines)
// app/incidences/page.tsx - Everything in one file

// AFTER: Split into focused components

// 1. components/incidences/IncidenceFilters.tsx
COMPONENT IncidenceFilters
  PROPS:
    filters: FilterState
    onFilterChange: (filters) => void
    periods: PayrollPeriod[]
    employees: Employee[]
    incidenceTypes: IncidenceType[]

  RENDER:
    <Card>
      <CardHeader>Filters</CardHeader>
      <CardContent>
        <SelectPeriod
          value={filters.periodId}
          options={periods}
          onChange={(id) => onFilterChange({...filters, periodId: id})}
        />
        <SelectEmployee
          value={filters.employeeId}
          options={employees}
          onChange={(id) => onFilterChange({...filters, employeeId: id})}
        />
        <SelectStatus
          value={filters.status}
          onChange={(status) => onFilterChange({...filters, status})}
        />
        <SelectType
          value={filters.typeId}
          options={incidenceTypes}
          onChange={(id) => onFilterChange({...filters, typeId: id})}
        />
      </CardContent>
    </Card>
  END RENDER
END COMPONENT

// 2. components/incidences/IncidenceTable.tsx
COMPONENT IncidenceTable
  PROPS:
    incidences: Incidence[]
    onApprove: (id) => void
    onReject: (id) => void
    onDelete: (id) => void
    onViewEvidence: (id) => void
    loading: boolean

  RENDER:
    <Table>
      <TableHeader>
        <Column>Employee</Column>
        <Column>Type</Column>
        <Column>Dates</Column>
        <Column>Quantity</Column>
        <Column>Status</Column>
        <Column>Actions</Column>
      </TableHeader>
      <TableBody>
        {incidences.map(incidence => (
          <IncidenceRow
            key={incidence.id}
            incidence={incidence}
            onApprove={onApprove}
            onReject={onReject}
            onDelete={onDelete}
            onViewEvidence={onViewEvidence}
          />
        ))}
      </TableBody>
    </Table>
  END RENDER
END COMPONENT

// 3. components/incidences/IncidenceForm.tsx
COMPONENT IncidenceForm
  PROPS:
    mode: 'single' | 'multiple' | 'all'
    employees: Employee[]
    incidenceTypes: IncidenceType[]
    selectedPeriod: PayrollPeriod
    onSubmit: (data) => Promise<void>
    onCancel: () => void

  STATE:
    formData: IncidenceFormData
    validationErrors: ValidationErrors
    submitting: boolean

  FUNCTION handleSubmit()
    SET submitting = TRUE

    errors = validateForm(formData)
    IF errors.hasErrors THEN
      SET validationErrors = errors
      SET submitting = FALSE
      RETURN
    END IF

    TRY
      AWAIT onSubmit(formData)
      toast.success("Incidence created successfully")
      resetForm()
    CATCH error
      toast.error("Failed to create incidence")
    FINALLY
      SET submitting = FALSE
    END TRY
  END FUNCTION

  RENDER:
    <Dialog>
      <DialogContent>
        <Form onSubmit={handleSubmit}>
          <ModeSelector mode={mode} />

          {mode === 'single' && (
            <EmployeeSelector
              employees={employees}
              value={formData.employeeId}
              onChange={(id) => updateFormData('employeeId', id)}
            />
          )}

          <IncidenceTypeSelector
            types={incidenceTypes}
            value={formData.typeId}
            onChange={(id) => updateFormData('typeId', id)}
          />

          <DateRangePicker
            startDate={formData.startDate}
            endDate={formData.endDate}
            onChange={(start, end) => {
              updateFormData('startDate', start)
              updateFormData('endDate', end)
            }}
          />

          <QuantityInput
            value={formData.quantity}
            onChange={(q) => updateFormData('quantity', q)}
          />

          {mode === 'single' && (
            <EvidenceUploader
              files={formData.evidence}
              onFilesChange={(files) => updateFormData('evidence', files)}
            />
          )}

          <FormActions>
            <Button variant="outline" onClick={onCancel}>Cancel</Button>
            <Button type="submit" loading={submitting}>Create</Button>
          </FormActions>
        </Form>
      </DialogContent>
    </Dialog>
  END RENDER
END COMPONENT

// 4. hooks/useIncidences.ts - Custom Hook for Logic
HOOK useIncidences(periodId)
  STATE:
    incidences: Incidence[]
    loading: boolean
    error: Error | null

  FUNCTION fetchIncidences()
    SET loading = TRUE
    TRY
      data = AWAIT api.getIncidences({ periodId })
      SET incidences = data
    CATCH error
      SET error = error
    FINALLY
      SET loading = FALSE
    END TRY
  END FUNCTION

  FUNCTION approveIncidence(id)
    TRY
      AWAIT api.approveIncidence(id)
      SET incidences = incidences.map(i =>
        i.id === id ? {...i, status: 'approved'} : i
      )
    CATCH error
      THROW error
    END TRY
  END FUNCTION

  FUNCTION rejectIncidence(id, reason)
    TRY
      AWAIT api.rejectIncidence(id, reason)
      SET incidences = incidences.map(i =>
        i.id === id ? {...i, status: 'rejected'} : i
      )
    CATCH error
      THROW error
    END TRY
  END FUNCTION

  FUNCTION createIncidence(data)
    TRY
      newIncidence = AWAIT api.createIncidence(data)
      SET incidences = [...incidences, newIncidence]
      RETURN newIncidence
    CATCH error
      THROW error
    END TRY
  END FUNCTION

  // Auto-fetch when periodId changes
  useEffect(() => {
    IF periodId THEN
      fetchIncidences()
    END IF
  }, [periodId])

  RETURN {
    incidences,
    loading,
    error,
    approveIncidence,
    rejectIncidence,
    createIncidence,
    refetch: fetchIncidences
  }
END HOOK

// 5. Main Page Component (Now ~100 lines)
PAGE IncidencesPage
  STATE:
    filters: FilterState
    showForm: boolean
    formMode: 'single' | 'multiple' | 'all'

  // Use custom hooks
  periods = usePeriods()
  employees = useEmployees()
  incidenceTypes = useIncidenceTypes()
  { incidences, loading, approveIncidence, rejectIncidence, createIncidence }
    = useIncidences(filters.periodId)

  RENDER:
    <PageLayout title="Incidence Management">
      <IncidenceFilters
        filters={filters}
        onFilterChange={setFilters}
        periods={periods}
        employees={employees}
        incidenceTypes={incidenceTypes}
      />

      <ActionBar>
        <Button onClick={() => { setFormMode('single'); setShowForm(true) }}>
          Single Employee
        </Button>
        <Button onClick={() => { setFormMode('multiple'); setShowForm(true) }}>
          Multiple Employees
        </Button>
        <Button onClick={() => { setFormMode('all'); setShowForm(true) }}>
          All Employees
        </Button>
      </ActionBar>

      <IncidenceTable
        incidences={incidences}
        loading={loading}
        onApprove={approveIncidence}
        onReject={rejectIncidence}
        onDelete={deleteIncidence}
      />

      {showForm && (
        <IncidenceForm
          mode={formMode}
          employees={employees}
          incidenceTypes={incidenceTypes}
          selectedPeriod={selectedPeriod}
          onSubmit={createIncidence}
          onCancel={() => setShowForm(false)}
        />
      )}
    </PageLayout>
  END RENDER
END PAGE
```

### File Structure After Refactoring

```
components/
  incidences/
    index.ts                    // Re-exports
    IncidenceFilters.tsx        // ~80 lines
    IncidenceTable.tsx          // ~120 lines
    IncidenceRow.tsx            // ~60 lines
    IncidenceForm.tsx           // ~200 lines
    IncidenceActions.tsx        // ~80 lines
    EvidenceUploader.tsx        // ~100 lines
    EvidenceViewer.tsx          // ~80 lines

hooks/
  useIncidences.ts              // ~100 lines
  useIncidenceForm.ts           // ~80 lines

app/incidences/
  page.tsx                      // ~100 lines (down from 1453!)
```

### Acceptance Criteria
- [ ] No component exceeds 200 lines
- [ ] Each component has single responsibility
- [ ] Business logic in custom hooks
- [ ] Reusable across other pages
- [ ] All existing functionality preserved
- [ ] Unit tests for each new component

---

## [S2-003] Implement custom error types in Go backend

**Priority**: HIGH | **Effort**: 8h | **Tags**: refactor, backend, errors

### Problem
Current code uses string comparisons for error handling, which is brittle and error-prone.

### Solution Pseudocode

```
// internal/errors/errors.go
MODULE Errors

  // Error categories
  ENUM ErrorCategory
    VALIDATION
    AUTHENTICATION
    AUTHORIZATION
    NOT_FOUND
    CONFLICT
    INTERNAL
    EXTERNAL_SERVICE
  END ENUM

  // Application Error struct
  STRUCT AppError
    Category: ErrorCategory
    Code: string              // e.g., "AUTH001", "VAL002"
    Message: string           // User-friendly message
    Details: map<string, any> // Additional context
    Cause: error              // Underlying error
  END STRUCT

  // Implement error interface
  FUNCTION (e *AppError) Error() RETURNS string
    IF e.Cause IS NOT NULL THEN
      RETURN e.Message + ": " + e.Cause.Error()
    END IF
    RETURN e.Message
  END FUNCTION

  // Implement Unwrap for error chain
  FUNCTION (e *AppError) Unwrap() RETURNS error
    RETURN e.Cause
  END FUNCTION

  // Helper to check error type
  FUNCTION (e *AppError) Is(target error) RETURNS boolean
    t, ok = target.(*AppError)
    IF NOT ok THEN
      RETURN FALSE
    END IF
    RETURN e.Code == t.Code
  END FUNCTION

  // Error constructors
  FUNCTION NewValidationError(message string, details map) RETURNS *AppError
    RETURN &AppError{
      Category: VALIDATION,
      Code: "VAL001",
      Message: message,
      Details: details
    }
  END FUNCTION

  FUNCTION NewNotFoundError(resource string, id string) RETURNS *AppError
    RETURN &AppError{
      Category: NOT_FOUND,
      Code: "NF001",
      Message: resource + " not found",
      Details: {"resource": resource, "id": id}
    }
  END FUNCTION

  FUNCTION NewAuthenticationError(message string) RETURNS *AppError
    RETURN &AppError{
      Category: AUTHENTICATION,
      Code: "AUTH001",
      Message: message
    }
  END FUNCTION

  FUNCTION NewAuthorizationError(action string, resource string) RETURNS *AppError
    RETURN &AppError{
      Category: AUTHORIZATION,
      Code: "AUTHZ001",
      Message: "Not authorized to " + action + " " + resource,
      Details: {"action": action, "resource": resource}
    }
  END FUNCTION

  FUNCTION NewConflictError(message string) RETURNS *AppError
    RETURN &AppError{
      Category: CONFLICT,
      Code: "CONF001",
      Message: message
    }
  END FUNCTION

  FUNCTION NewInternalError(message string, cause error) RETURNS *AppError
    RETURN &AppError{
      Category: INTERNAL,
      Code: "INT001",
      Message: message,
      Cause: cause
    }
  END FUNCTION

END MODULE

// Error to HTTP status mapping
FUNCTION GetHTTPStatus(err error) RETURNS int
  appErr, ok = err.(*AppError)
  IF NOT ok THEN
    RETURN 500  // Internal server error for unknown errors
  END IF

  SWITCH appErr.Category
    CASE VALIDATION:
      RETURN 400
    CASE AUTHENTICATION:
      RETURN 401
    CASE AUTHORIZATION:
      RETURN 403
    CASE NOT_FOUND:
      RETURN 404
    CASE CONFLICT:
      RETURN 409
    CASE EXTERNAL_SERVICE:
      RETURN 502
    DEFAULT:
      RETURN 500
  END SWITCH
END FUNCTION

// Usage in Service Layer
FUNCTION (s *EmployeeService) GetEmployee(id string) (*Employee, error)
  employee, err = s.repo.FindByID(id)

  IF err IS gorm.ErrRecordNotFound THEN
    RETURN NULL, Errors.NewNotFoundError("Employee", id)
  END IF

  IF err IS NOT NULL THEN
    RETURN NULL, Errors.NewInternalError("Failed to fetch employee", err)
  END IF

  RETURN employee, NULL
END FUNCTION

// Usage in Handler Layer
FUNCTION (h *EmployeeHandler) GetEmployee(c *gin.Context)
  id = c.Param("id")

  employee, err = h.service.GetEmployee(id)
  IF err IS NOT NULL THEN
    status = Errors.GetHTTPStatus(err)
    c.JSON(status, gin.H{
      "error": err.Error(),
      "code": err.(*AppError).Code
    })
    RETURN
  END IF

  c.JSON(200, employee)
END FUNCTION

// Error handling middleware
FUNCTION ErrorMiddleware(c *gin.Context)
  c.Next()

  IF c.Errors.LENGTH > 0 THEN
    err = c.Errors.Last().Err
    status = Errors.GetHTTPStatus(err)

    response = gin.H{
      "success": FALSE,
      "error": err.Error()
    }

    appErr, ok = err.(*AppError)
    IF ok THEN
      response["code"] = appErr.Code
      IF appErr.Details IS NOT NULL THEN
        response["details"] = appErr.Details
      END IF
    END IF

    c.JSON(status, response)
  END IF
END FUNCTION
```

### Acceptance Criteria
- [ ] All services use typed errors
- [ ] HTTP status mapped automatically
- [ ] Error codes for client handling
- [ ] Error details preserved
- [ ] No string error comparisons
- [ ] Middleware handles all errors

---

# SPRINT 3: Testing Foundation

**Duration**: 1 Week (84 hours)
**Priority**: HIGH
**Goal**: Establish comprehensive test coverage

---

## [S3-003] Write unit tests for PayrollService

**Priority**: HIGH | **Effort**: 16h | **Tags**: testing, backend, payroll

### Solution Pseudocode

```
// tests/services/payroll_service_test.go
MODULE PayrollServiceTests

  // Test fixtures
  FUNCTION createTestEmployee(collarType, dailySalary) RETURNS Employee
    RETURN Employee{
      ID: UUID(),
      EmployeeNumber: "EMP001",
      FirstName: "Test",
      LastName: "Employee",
      CollarType: collarType,
      DailySalary: dailySalary,
      HireDate: "2020-01-15",
      EmploymentStatus: "active"
    }
  END FUNCTION

  FUNCTION createTestPeriod(periodType, startDate, endDate) RETURNS PayrollPeriod
    RETURN PayrollPeriod{
      ID: UUID(),
      PeriodCode: "2025-BW01",
      PeriodType: periodType,
      StartDate: startDate,
      EndDate: endDate,
      Status: "open"
    }
  END FUNCTION

  // ISR Calculation Tests
  TEST "TestCalculateISR_LowerBracket"
    service = NewPayrollService(mockRepo, taxTables)

    // Employee earning 500/day * 15 days = 7,500 biweekly
    taxableIncome = 7500.00
    periodType = "biweekly"

    isr = service.CalculateISR(taxableIncome, periodType)

    // Based on 2025 ISR tables for biweekly:
    // Lower limit: 6,224.69, Fixed fee: 353.22, Rate: 10.88%
    expectedLowerLimit = 6224.69
    expectedFixedFee = 353.22
    expectedRate = 0.1088
    expectedExcess = taxableIncome - expectedLowerLimit  // 1,275.31
    expectedISR = expectedFixedFee + (expectedExcess * expectedRate)  // 492.05

    ASSERT isr.APPROXIMATELY_EQUALS(expectedISR, 0.01)
  END TEST

  TEST "TestCalculateISR_HigherBracket"
    service = NewPayrollService(mockRepo, taxTables)

    taxableIncome = 50000.00
    periodType = "biweekly"

    isr = service.CalculateISR(taxableIncome, periodType)

    // Higher bracket calculation
    expectedISR = 8721.45  // Pre-calculated from tables

    ASSERT isr.APPROXIMATELY_EQUALS(expectedISR, 0.01)
  END TEST

  TEST "TestCalculateISR_MonthlyPeriod"
    service = NewPayrollService(mockRepo, taxTables)

    taxableIncome = 30000.00
    periodType = "monthly"

    isr = service.CalculateISR(taxableIncome, periodType)

    // Should use monthly tables, not biweekly
    ASSERT isr > 0
    ASSERT isr != service.CalculateISR(taxableIncome, "biweekly")
  END TEST

  // Employment Subsidy Tests
  TEST "TestCalculateEmploymentSubsidy_Eligible"
    service = NewPayrollService(mockRepo, taxTables)

    // Low income employee eligible for subsidy
    taxableIncome = 4500.00
    periodType = "biweekly"

    subsidy = service.CalculateEmploymentSubsidy(taxableIncome, periodType)

    // Based on subsidy tables
    ASSERT subsidy > 0
    ASSERT subsidy < taxableIncome  // Subsidy cannot exceed income
  END TEST

  TEST "TestCalculateEmploymentSubsidy_NotEligible"
    service = NewPayrollService(mockRepo, taxTables)

    // High income employee not eligible
    taxableIncome = 25000.00
    periodType = "biweekly"

    subsidy = service.CalculateEmploymentSubsidy(taxableIncome, periodType)

    ASSERT subsidy == 0
  END TEST

  // IMSS Calculation Tests
  TEST "TestCalculateIMSSEmployee_StandardRate"
    service = NewPayrollService(mockRepo, taxTables)

    sdi = 500.00  // Salario Diario Integrado
    workedDays = 15

    imss = service.CalculateIMSSEmployee(sdi, workedDays)

    // IMSS employee contributions:
    // - Enfermedad y Maternidad: variable based on SDI vs 3 UMA
    // - Invalidez y Vida: 0.625%
    // - Cesantía y Vejez: 1.125%
    // - Retiro: 0% (employer only)

    expectedBase = sdi * workedDays  // 7,500
    expectedRate = 0.02775  // Combined employee rate
    expectedIMSS = expectedBase * expectedRate

    ASSERT imss.APPROXIMATELY_EQUALS(expectedIMSS, 0.50)
  END TEST

  TEST "TestCalculateIMSSEmployee_ExceedingTopaz"
    service = NewPayrollService(mockRepo, taxTables)

    // Very high SDI - should be capped at 25 UMA
    sdi = 5000.00
    workedDays = 15

    imss = service.CalculateIMSSEmployee(sdi, workedDays)

    // SDI capped at 25 * UMA (108.57) = 2,714.25
    cappedSDI = MIN(sdi, 25 * 108.57)

    ASSERT imss <= cappedSDI * workedDays * 0.05  // Max possible rate
  END TEST

  // INFONAVIT Tests
  TEST "TestCalculateINFONAVITEmployee_NoCredit"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("white_collar", 500)
    employee.InfonavitCredit = ""

    infonavit = service.CalculateINFONAVITEmployee(employee, 15)

    // No credit = no employee deduction
    ASSERT infonavit == 0
  END TEST

  TEST "TestCalculateINFONAVITEmployee_WithCredit"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("white_collar", 500)
    employee.InfonavitCredit = "123456789"
    employee.InfonavitDeductionType = "percentage"
    employee.InfonavitDeductionValue = 25  // 25% of SDI

    sdi = 525.00  // SDI with integration factor
    workedDays = 15

    infonavit = service.CalculateINFONAVITEmployee(employee, workedDays)

    expectedBase = sdi * workedDays
    expectedDeduction = expectedBase * 0.25

    ASSERT infonavit.APPROXIMATELY_EQUALS(expectedDeduction, 0.50)
  END TEST

  // SDI (Salario Diario Integrado) Tests
  TEST "TestCalculateSDI_FirstYear"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("white_collar", 500)
    employee.HireDate = TODAY().SUBTRACT(6 MONTHS)  // Less than 1 year

    sdi = service.CalculateSDI(employee)

    // First year: 12 vacation days, 25% prima vacacional, 15 days aguinaldo
    // Factor = 1 + (12 * 0.25 + 15) / 365 = 1.0493
    expectedFactor = 1.0493
    expectedSDI = 500 * expectedFactor

    ASSERT sdi.APPROXIMATELY_EQUALS(expectedSDI, 0.50)
  END TEST

  TEST "TestCalculateSDI_FiveYears"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("white_collar", 500)
    employee.HireDate = TODAY().SUBTRACT(5 YEARS)

    sdi = service.CalculateSDI(employee)

    // 5 years: 20 vacation days
    // Factor = 1 + (20 * 0.25 + 15) / 365 = 1.0548
    expectedFactor = 1.0548
    expectedSDI = 500 * expectedFactor

    ASSERT sdi.APPROXIMATELY_EQUALS(expectedSDI, 0.50)
  END TEST

  // Full Payroll Calculation Integration
  TEST "TestCalculatePayroll_WhiteCollar_Biweekly"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("white_collar", 800)
    period = createTestPeriod("biweekly", "2025-01-01", "2025-01-15")

    prenomina = PrenominaMetric{
      WorkedDays: 15,
      RegularHours: 120,
      OvertimeHours: 5,
      AbsenceDays: 0,
      VacationDays: 0
    }

    result = service.CalculatePayroll(employee, period, prenomina)

    // Verify all components calculated
    ASSERT result.RegularSalary == 800 * 15  // 12,000
    ASSERT result.OvertimeAmount > 0
    ASSERT result.ISRWithholding > 0
    ASSERT result.IMSSEmployee > 0
    ASSERT result.TotalGrossIncome > result.RegularSalary
    ASSERT result.TotalNetPay < result.TotalGrossIncome
    ASSERT result.TotalNetPay > 0
  END TEST

  TEST "TestCalculatePayroll_BlueCollar_Weekly"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("blue_collar", 300)
    employee.IsSindicalizado = TRUE
    period = createTestPeriod("weekly", "2025-01-01", "2025-01-07")

    prenomina = PrenominaMetric{
      WorkedDays: 6,  // 6 days worked, 1 rest day
      RegularHours: 48,
      OvertimeHours: 4,  // Extra hours
      AbsenceDays: 0
    }

    result = service.CalculatePayroll(employee, period, prenomina)

    // Blue collar weekly calculations
    ASSERT result.RegularSalary == 300 * 6
    ASSERT result.TotalNetPay > 0
  END TEST

  // Edge Cases
  TEST "TestCalculatePayroll_ZeroWorkedDays"
    service = NewPayrollService(mockRepo, taxTables)

    employee = createTestEmployee("white_collar", 500)
    period = createTestPeriod("biweekly", "2025-01-01", "2025-01-15")

    prenomina = PrenominaMetric{
      WorkedDays: 0,
      AbsenceDays: 15  // Full period absent
    }

    result = service.CalculatePayroll(employee, period, prenomina)

    ASSERT result.RegularSalary == 0
    ASSERT result.TotalNetPay == 0
    ASSERT result.CalculationStatus == "calculated"
  END TEST

  TEST "TestCalculatePayroll_NegativeNetPay_ShouldNotOccur"
    service = NewPayrollService(mockRepo, taxTables)

    // Employee with very low income and high deductions
    employee = createTestEmployee("white_collar", 100)
    period = createTestPeriod("biweekly", "2025-01-01", "2025-01-15")

    prenomina = PrenominaMetric{
      WorkedDays: 15,
      LoanDeduction: 2000  // High loan
    }

    result = service.CalculatePayroll(employee, period, prenomina)

    // Net pay should never go negative
    ASSERT result.TotalNetPay >= 0

    // If deductions exceed income, they should be capped
    IF result.TotalNetPay == 0 THEN
      ASSERT result.LoanDeductions < prenomina.LoanDeduction
    END IF
  END TEST

END MODULE
```

### Acceptance Criteria
- [ ] ISR calculation tests for all brackets
- [ ] IMSS calculation tests including caps
- [ ] INFONAVIT tests with/without credit
- [ ] SDI tests for different seniority levels
- [ ] Integration tests for full payroll
- [ ] Edge case tests (zero days, negative net)
- [ ] 80% code coverage for PayrollService

---

# SPRINT 4: Recruitment Module

**Duration**: 1 Week (104 hours)
**Priority**: HIGH
**Goal**: Implement complete recruitment pipeline

---

## [S4-001] Design recruitment module database schema

**Priority**: HIGH | **Effort**: 6h | **Tags**: recruitment, database, design

### Solution Pseudocode

```
// Database Schema Design

// TABLE: job_postings
// Purpose: Job listings with requirements
TABLE job_postings (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  company_id UUID REFERENCES companies(id) NOT NULL

  -- Basic Info
  title VARCHAR(255) NOT NULL
  description TEXT NOT NULL
  requirements TEXT
  responsibilities TEXT

  -- Classification
  department_id UUID REFERENCES departments(id)
  cost_center_id UUID REFERENCES cost_centers(id)
  position_level VARCHAR(50)  -- entry, mid, senior, manager, director
  employment_type VARCHAR(50) NOT NULL  -- full_time, part_time, contract, intern
  collar_type VARCHAR(20)  -- white_collar, blue_collar, gray_collar

  -- Compensation
  salary_min DECIMAL(12,2)
  salary_max DECIMAL(12,2)
  salary_currency VARCHAR(3) DEFAULT 'MXN'
  salary_frequency VARCHAR(20)  -- daily, biweekly, monthly, annual
  show_salary BOOLEAN DEFAULT FALSE

  -- Headcount
  positions_available INTEGER DEFAULT 1
  positions_filled INTEGER DEFAULT 0

  -- Location
  location VARCHAR(255)
  is_remote BOOLEAN DEFAULT FALSE
  remote_type VARCHAR(50)  -- fully_remote, hybrid, on_site

  -- Status & Dates
  status VARCHAR(50) DEFAULT 'draft'  -- draft, published, paused, closed, filled
  published_at TIMESTAMP
  closes_at TIMESTAMP

  -- Workflow
  hiring_manager_id UUID REFERENCES users(id)
  recruiter_id UUID REFERENCES users(id)

  -- Metadata
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  created_by UUID REFERENCES users(id)
  deleted_at TIMESTAMP
)

// TABLE: candidates
// Purpose: People applying for jobs
TABLE candidates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  company_id UUID REFERENCES companies(id) NOT NULL

  -- Personal Info
  first_name VARCHAR(100) NOT NULL
  last_name VARCHAR(100) NOT NULL
  email VARCHAR(255) NOT NULL
  phone VARCHAR(20)

  -- Professional Info
  current_title VARCHAR(255)
  current_company VARCHAR(255)
  years_experience INTEGER
  linkedin_url VARCHAR(255)
  portfolio_url VARCHAR(255)

  -- Documents
  resume_file_id UUID REFERENCES documents(id)
  cover_letter TEXT

  -- Source Tracking
  source VARCHAR(100)  -- linkedin, referral, job_board, website, agency
  source_details VARCHAR(255)
  referred_by_employee_id UUID REFERENCES employees(id)

  -- Status
  status VARCHAR(50) DEFAULT 'active'  -- active, hired, rejected, withdrawn

  -- Tags for filtering
  tags TEXT[]

  -- Metadata
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  deleted_at TIMESTAMP
)

// TABLE: applications
// Purpose: Links candidates to job postings with pipeline tracking
TABLE applications (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  company_id UUID REFERENCES companies(id) NOT NULL
  candidate_id UUID REFERENCES candidates(id) NOT NULL
  job_posting_id UUID REFERENCES job_postings(id) NOT NULL

  -- Pipeline Stage
  stage VARCHAR(50) NOT NULL DEFAULT 'new'
    -- new, screening, phone_interview, technical_interview,
    -- onsite_interview, offer, hired, rejected
  stage_entered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

  -- Status
  status VARCHAR(50) DEFAULT 'active'  -- active, withdrawn, rejected, hired
  rejection_reason VARCHAR(255)
  rejection_notes TEXT

  -- Screening
  screening_score INTEGER  -- 1-5 rating from screening
  screening_notes TEXT
  screened_by UUID REFERENCES users(id)
  screened_at TIMESTAMP

  -- Overall Rating (aggregated from interviews)
  overall_rating DECIMAL(3,2)  -- Average of all interview ratings

  -- Salary Negotiation
  expected_salary DECIMAL(12,2)
  offered_salary DECIMAL(12,2)

  -- Timeline
  applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  hired_at TIMESTAMP
  start_date DATE

  -- Metadata
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  deleted_at TIMESTAMP

  UNIQUE(candidate_id, job_posting_id)  -- One application per job
)

// TABLE: interviews
// Purpose: Individual interview sessions
TABLE interviews (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  application_id UUID REFERENCES applications(id) NOT NULL

  -- Schedule
  scheduled_at TIMESTAMP NOT NULL
  duration_minutes INTEGER DEFAULT 60
  location VARCHAR(255)  -- Room, video link, etc.
  interview_type VARCHAR(50)  -- phone, video, onsite, technical, behavioral

  -- Participants
  interviewer_id UUID REFERENCES users(id) NOT NULL
  additional_interviewers UUID[]  -- Array of user IDs

  -- Status
  status VARCHAR(50) DEFAULT 'scheduled'  -- scheduled, completed, cancelled, no_show

  -- Feedback (after interview)
  rating INTEGER  -- 1-5
  strengths TEXT
  weaknesses TEXT
  recommendation VARCHAR(50)  -- strong_hire, hire, no_hire, strong_no_hire
  notes TEXT
  feedback_submitted_at TIMESTAMP

  -- Calendar Integration
  calendar_event_id VARCHAR(255)
  video_meeting_url VARCHAR(255)

  -- Metadata
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  deleted_at TIMESTAMP
)

// TABLE: offers
// Purpose: Job offers to candidates
TABLE offers (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  application_id UUID REFERENCES applications(id) NOT NULL

  -- Offer Details
  position_title VARCHAR(255) NOT NULL
  department_id UUID REFERENCES departments(id)

  -- Compensation
  salary_amount DECIMAL(12,2) NOT NULL
  salary_frequency VARCHAR(20) NOT NULL
  bonus_amount DECIMAL(12,2)
  bonus_type VARCHAR(50)  -- signing, annual, quarterly

  -- Benefits
  benefits_package TEXT  -- JSON or text description

  -- Dates
  start_date DATE NOT NULL
  offer_expires_at TIMESTAMP

  -- Status
  status VARCHAR(50) DEFAULT 'draft'  -- draft, sent, accepted, declined, expired, withdrawn
  sent_at TIMESTAMP
  responded_at TIMESTAMP
  response_notes TEXT

  -- Document
  offer_letter_document_id UUID REFERENCES documents(id)
  signed_document_id UUID REFERENCES documents(id)

  -- Approval
  approved_by UUID REFERENCES users(id)
  approved_at TIMESTAMP

  -- Metadata
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  created_by UUID REFERENCES users(id)
  deleted_at TIMESTAMP
)

// TABLE: evaluation_criteria
// Purpose: Define what to evaluate in interviews
TABLE evaluation_criteria (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  company_id UUID REFERENCES companies(id) NOT NULL
  job_posting_id UUID REFERENCES job_postings(id)  -- NULL = company-wide

  name VARCHAR(255) NOT NULL
  description TEXT
  category VARCHAR(50)  -- technical, behavioral, cultural, experience
  weight INTEGER DEFAULT 1  -- For weighted scoring

  is_required BOOLEAN DEFAULT FALSE
  display_order INTEGER DEFAULT 0
  is_active BOOLEAN DEFAULT TRUE

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  deleted_at TIMESTAMP
)

// TABLE: interview_scores
// Purpose: Individual criteria scores from interviews
TABLE interview_scores (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
  interview_id UUID REFERENCES interviews(id) NOT NULL
  criteria_id UUID REFERENCES evaluation_criteria(id) NOT NULL

  score INTEGER NOT NULL  -- 1-5
  notes TEXT

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

  UNIQUE(interview_id, criteria_id)
)
```

### Acceptance Criteria
- [ ] All tables created with proper indexes
- [ ] Foreign key relationships established
- [ ] Soft delete support on all tables
- [ ] Migration scripts created
- [ ] Seed data for testing
- [ ] Documentation complete

---

## [S4-002] Implement JobPosting service and API

**Priority**: HIGH | **Effort**: 10h | **Tags**: recruitment, backend, api

### Solution Pseudocode

```
// internal/services/job_posting_service.go
MODULE JobPostingService

  STRUCT JobPostingService
    repo: JobPostingRepository
    companyRepo: CompanyRepository
    notificationService: NotificationService
  END STRUCT

  // Create new job posting
  FUNCTION (s *JobPostingService) Create(dto CreateJobPostingDTO) (*JobPosting, error)
    // Validate required fields
    IF dto.Title IS EMPTY THEN
      RETURN NULL, Errors.NewValidationError("Title is required", NULL)
    END IF

    IF dto.Description IS EMPTY THEN
      RETURN NULL, Errors.NewValidationError("Description is required", NULL)
    END IF

    // Validate salary range if provided
    IF dto.SalaryMin > 0 AND dto.SalaryMax > 0 THEN
      IF dto.SalaryMin > dto.SalaryMax THEN
        RETURN NULL, Errors.NewValidationError(
          "Minimum salary cannot exceed maximum salary",
          {"salaryMin": dto.SalaryMin, "salaryMax": dto.SalaryMax}
        )
      END IF
    END IF

    // Create job posting
    posting = &JobPosting{
      ID: UUID(),
      CompanyID: dto.CompanyID,
      Title: dto.Title,
      Description: dto.Description,
      Requirements: dto.Requirements,
      Responsibilities: dto.Responsibilities,
      DepartmentID: dto.DepartmentID,
      PositionLevel: dto.PositionLevel,
      EmploymentType: dto.EmploymentType,
      CollarType: dto.CollarType,
      SalaryMin: dto.SalaryMin,
      SalaryMax: dto.SalaryMax,
      SalaryCurrency: COALESCE(dto.SalaryCurrency, "MXN"),
      ShowSalary: dto.ShowSalary,
      PositionsAvailable: COALESCE(dto.PositionsAvailable, 1),
      Location: dto.Location,
      IsRemote: dto.IsRemote,
      RemoteType: dto.RemoteType,
      HiringManagerID: dto.HiringManagerID,
      RecruiterID: dto.RecruiterID,
      Status: "draft",
      CreatedBy: dto.CreatedByUserID
    }

    err = s.repo.Create(posting)
    IF err IS NOT NULL THEN
      RETURN NULL, Errors.NewInternalError("Failed to create job posting", err)
    END IF

    RETURN posting, NULL
  END FUNCTION

  // Publish job posting (make it live)
  FUNCTION (s *JobPostingService) Publish(id UUID, userID UUID) (*JobPosting, error)
    posting, err = s.repo.FindByID(id)
    IF err IS NOT NULL THEN
      RETURN NULL, Errors.NewNotFoundError("JobPosting", id.String())
    END IF

    // Validate posting is ready to publish
    validationErrors = s.validateForPublish(posting)
    IF validationErrors.LENGTH > 0 THEN
      RETURN NULL, Errors.NewValidationError(
        "Job posting is not ready to publish",
        {"errors": validationErrors}
      )
    END IF

    // Update status
    posting.Status = "published"
    posting.PublishedAt = NOW()

    err = s.repo.Update(posting)
    IF err IS NOT NULL THEN
      RETURN NULL, Errors.NewInternalError("Failed to publish job posting", err)
    END IF

    // Notify hiring manager
    s.notificationService.CreateNotification(Notification{
      TargetUserID: posting.HiringManagerID,
      Type: "job_published",
      Title: "Job Posted",
      Message: "Your job posting '" + posting.Title + "' is now live"
    })

    RETURN posting, NULL
  END FUNCTION

  FUNCTION (s *JobPostingService) validateForPublish(posting *JobPosting) []string
    errors = []

    IF posting.Title IS EMPTY THEN
      errors.ADD("Title is required")
    END IF

    IF posting.Description IS EMPTY THEN
      errors.ADD("Description is required")
    END IF

    IF posting.EmploymentType IS EMPTY THEN
      errors.ADD("Employment type is required")
    END IF

    IF posting.HiringManagerID IS NULL THEN
      errors.ADD("Hiring manager must be assigned")
    END IF

    RETURN errors
  END FUNCTION

  // List job postings with filters
  FUNCTION (s *JobPostingService) List(filters JobPostingFilters) (*PaginatedResult, error)
    query = s.repo.NewQuery()

    // Apply filters
    IF filters.Status IS NOT EMPTY THEN
      query = query.WHERE("status = ?", filters.Status)
    END IF

    IF filters.DepartmentID IS NOT NULL THEN
      query = query.WHERE("department_id = ?", filters.DepartmentID)
    END IF

    IF filters.EmploymentType IS NOT EMPTY THEN
      query = query.WHERE("employment_type = ?", filters.EmploymentType)
    END IF

    IF filters.IsRemote IS NOT NULL THEN
      query = query.WHERE("is_remote = ?", filters.IsRemote)
    END IF

    IF filters.Search IS NOT EMPTY THEN
      searchTerm = "%" + filters.Search + "%"
      query = query.WHERE(
        "title ILIKE ? OR description ILIKE ?",
        searchTerm, searchTerm
      )
    END IF

    // Get total count
    total = query.Count()

    // Apply pagination
    query = query.OFFSET(filters.Offset).LIMIT(filters.Limit)

    // Apply sorting
    query = query.ORDER_BY(COALESCE(filters.SortBy, "created_at") + " " +
                           COALESCE(filters.SortOrder, "DESC"))

    postings = query.Find()

    RETURN &PaginatedResult{
      Data: postings,
      Total: total,
      Page: filters.Page,
      PageSize: filters.Limit
    }, NULL
  END FUNCTION

  // Get posting statistics
  FUNCTION (s *JobPostingService) GetStats(postingID UUID) (*JobPostingStats, error)
    posting, err = s.repo.FindByID(postingID)
    IF err IS NOT NULL THEN
      RETURN NULL, Errors.NewNotFoundError("JobPosting", postingID.String())
    END IF

    stats = &JobPostingStats{
      PostingID: postingID,
      TotalApplications: s.countApplications(postingID),
      ApplicationsByStage: s.countApplicationsByStage(postingID),
      AverageTimeInStage: s.calculateAverageTimeInStage(postingID),
      ConversionRates: s.calculateConversionRates(postingID),
      SourceBreakdown: s.getSourceBreakdown(postingID)
    }

    RETURN stats, NULL
  END FUNCTION

  // Close job posting
  FUNCTION (s *JobPostingService) Close(id UUID, reason string) (*JobPosting, error)
    posting, err = s.repo.FindByID(id)
    IF err IS NOT NULL THEN
      RETURN NULL, Errors.NewNotFoundError("JobPosting", id.String())
    END IF

    posting.Status = "closed"
    posting.ClosedAt = NOW()
    posting.CloseReason = reason

    err = s.repo.Update(posting)
    IF err IS NOT NULL THEN
      RETURN NULL, Errors.NewInternalError("Failed to close job posting", err)
    END IF

    // Notify candidates with pending applications
    s.notifyPendingCandidates(posting)

    RETURN posting, NULL
  END FUNCTION

END MODULE

// API Handler
// internal/api/job_posting_handler.go
MODULE JobPostingHandler

  // GET /api/job-postings
  FUNCTION (h *JobPostingHandler) List(c *gin.Context)
    filters = JobPostingFilters{
      Status: c.Query("status"),
      DepartmentID: ParseUUID(c.Query("departmentId")),
      EmploymentType: c.Query("employmentType"),
      IsRemote: ParseBool(c.Query("isRemote")),
      Search: c.Query("search"),
      Page: ParseInt(c.Query("page"), 1),
      Limit: ParseInt(c.Query("limit"), 20)
    }

    result, err = h.service.List(filters)
    IF err IS NOT NULL THEN
      c.Error(err)
      RETURN
    END IF

    c.JSON(200, result)
  END FUNCTION

  // POST /api/job-postings
  FUNCTION (h *JobPostingHandler) Create(c *gin.Context)
    dto = CreateJobPostingDTO{}
    IF err = c.ShouldBindJSON(&dto); err IS NOT NULL THEN
      c.Error(Errors.NewValidationError("Invalid request body", NULL))
      RETURN
    END IF

    dto.CompanyID = GetCompanyIDFromContext(c)
    dto.CreatedByUserID = GetUserIDFromContext(c)

    posting, err = h.service.Create(dto)
    IF err IS NOT NULL THEN
      c.Error(err)
      RETURN
    END IF

    c.JSON(201, posting)
  END FUNCTION

  // POST /api/job-postings/:id/publish
  FUNCTION (h *JobPostingHandler) Publish(c *gin.Context)
    id = ParseUUID(c.Param("id"))
    userID = GetUserIDFromContext(c)

    posting, err = h.service.Publish(id, userID)
    IF err IS NOT NULL THEN
      c.Error(err)
      RETURN
    END IF

    c.JSON(200, posting)
  END FUNCTION

  // GET /api/job-postings/:id/stats
  FUNCTION (h *JobPostingHandler) GetStats(c *gin.Context)
    id = ParseUUID(c.Param("id"))

    stats, err = h.service.GetStats(id)
    IF err IS NOT NULL THEN
      c.Error(err)
      RETURN
    END IF

    c.JSON(200, stats)
  END FUNCTION

END MODULE
```

### Acceptance Criteria
- [ ] CRUD operations for job postings
- [ ] Publish/unpublish workflow
- [ ] Status transitions validated
- [ ] Filter and search capability
- [ ] Statistics endpoint
- [ ] Proper error handling
- [ ] API documentation

---

# SPRINT 5-17: Additional Modules

[Content continues with same detailed format for remaining sprints...]

---

# SPRINT 18: Final QA & Launch Prep

**Duration**: 1 Week (132 hours)
**Priority**: URGENT
**Goal**: Ensure production readiness

---

## [S18-001] Final security audit

**Priority**: URGENT | **Effort**: 16h | **Tags**: security, audit, qa

### Solution Pseudocode

```
// Security Audit Checklist and Process

MODULE SecurityAudit

  // 1. Automated Security Scanning
  FUNCTION RunAutomatedScans()
    results = []

    // OWASP ZAP Scan
    zapResults = RunZAPScan({
      target: PRODUCTION_URL,
      scanType: "full",
      contexts: ["authenticated", "unauthenticated"]
    })
    results.ADD(zapResults)

    // Dependency vulnerability scan
    depResults = RunDependencyCheck({
      backend: {
        tool: "govulncheck",
        path: "/backend"
      },
      frontend: {
        tool: "npm audit",
        paths: ["/frontend", "/employee-portal"]
      }
    })
    results.ADD(depResults)

    // Static code analysis
    codeResults = RunStaticAnalysis({
      backend: {
        tool: "gosec",
        path: "/backend"
      },
      frontend: {
        tool: "eslint-plugin-security",
        paths: ["/frontend", "/employee-portal"]
      }
    })
    results.ADD(codeResults)

    RETURN results
  END FUNCTION

  // 2. Manual Security Checklist
  CHECKLIST ManualSecurityReview

    // Authentication
    [_] JWT tokens stored in httpOnly cookies
    [_] Refresh tokens properly isolated
    [_] Password hashing uses bcrypt with cost >= 12
    [_] Password strength requirements enforced
    [_] Rate limiting on auth endpoints
    [_] Account lockout after failed attempts
    [_] Session timeout configured

    // Authorization
    [_] All endpoints require authentication (except public)
    [_] Role-based access properly enforced
    [_] No privilege escalation possible
    [_] Resource ownership verified before operations
    [_] Admin functions properly protected

    // Input Validation
    [_] All inputs sanitized
    [_] SQL injection prevented (parameterized queries)
    [_] XSS prevented (output encoding)
    [_] CSRF protection enabled
    [_] File upload validation and restrictions
    [_] Request size limits configured

    // Data Protection
    [_] Sensitive data encrypted at rest
    [_] TLS 1.2+ enforced for all connections
    [_] No sensitive data in logs
    [_] No sensitive data in URLs
    [_] Proper data masking in responses
    [_] PII handling compliant

    // Headers & Configuration
    [_] Security headers present (CSP, HSTS, etc.)
    [_] CORS properly configured
    [_] Debug mode disabled in production
    [_] Error messages don't leak info
    [_] Server version headers removed

    // API Security
    [_] Rate limiting on all endpoints
    [_] Request validation on all inputs
    [_] Response data minimized (no over-fetching)
    [_] API versioning implemented
    [_] Deprecation warnings for old versions

    // Infrastructure
    [_] Firewall rules reviewed
    [_] Database not publicly accessible
    [_] Secrets in environment variables
    [_] No hardcoded credentials
    [_] Backup encryption enabled

  END CHECKLIST

  // 3. Penetration Testing Scenarios
  FUNCTION PenetrationTests()
    scenarios = [
      // Authentication Attacks
      {
        name: "Brute Force Login",
        steps: [
          "Attempt 100 logins with wrong password",
          "Verify rate limiting triggers",
          "Verify account lockout works"
        ],
        expected: "Blocked after 5 attempts, 5-minute lockout"
      },
      {
        name: "Session Hijacking",
        steps: [
          "Capture JWT token",
          "Attempt to use from different IP",
          "Attempt to use expired token"
        ],
        expected: "Token invalid or session terminated"
      },
      {
        name: "Token Manipulation",
        steps: [
          "Modify JWT payload (change role to admin)",
          "Submit modified token",
          "Verify rejection"
        ],
        expected: "Invalid signature error"
      },

      // Injection Attacks
      {
        name: "SQL Injection",
        steps: [
          "Submit SQL payload in search: ' OR 1=1 --",
          "Submit SQL payload in ID parameter",
          "Verify no data leak"
        ],
        expected: "Input sanitized, no SQL execution"
      },
      {
        name: "XSS Stored",
        steps: [
          "Submit <script>alert('xss')</script> in name field",
          "View the record",
          "Verify script not executed"
        ],
        expected: "Content escaped, no script execution"
      },

      // Authorization Attacks
      {
        name: "IDOR - Access Other User Data",
        steps: [
          "Login as Employee A",
          "Request Employee B's payslip by ID",
          "Verify access denied"
        ],
        expected: "403 Forbidden"
      },
      {
        name: "Privilege Escalation",
        steps: [
          "Login as regular user",
          "Attempt to access /admin endpoints",
          "Attempt to create admin user"
        ],
        expected: "403 Forbidden for all attempts"
      },

      // File Upload Attacks
      {
        name: "Malicious File Upload",
        steps: [
          "Upload PHP file renamed as .jpg",
          "Upload file with double extension .jpg.php",
          "Upload oversized file"
        ],
        expected: "All rejected, file type validated"
      }
    ]

    RETURN ExecuteScenarios(scenarios)
  END FUNCTION

  // 4. Generate Security Report
  FUNCTION GenerateSecurityReport(scanResults, manualResults, penTestResults)
    report = SecurityReport{
      GeneratedAt: NOW(),
      OverallScore: CalculateSecurityScore(scanResults, manualResults),

      CriticalFindings: FilterBySeverity(scanResults, "CRITICAL"),
      HighFindings: FilterBySeverity(scanResults, "HIGH"),
      MediumFindings: FilterBySeverity(scanResults, "MEDIUM"),
      LowFindings: FilterBySeverity(scanResults, "LOW"),

      DependencyVulnerabilities: scanResults.Dependencies,
      PenetrationTestResults: penTestResults,

      Recommendations: GenerateRecommendations(scanResults),
      ComplianceStatus: CheckCompliance(["OWASP", "PCI-DSS", "GDPR"])
    }

    // All critical and high issues must be resolved before launch
    IF report.CriticalFindings.LENGTH > 0 OR report.HighFindings.LENGTH > 0 THEN
      report.LaunchApproved = FALSE
      report.BlockingIssues = report.CriticalFindings + report.HighFindings
    ELSE
      report.LaunchApproved = TRUE
    END IF

    RETURN report
  END FUNCTION

END MODULE
```

### Acceptance Criteria
- [ ] OWASP ZAP scan shows no critical/high issues
- [ ] Dependency scan shows no critical vulnerabilities
- [ ] All penetration test scenarios pass
- [ ] Security checklist 100% complete
- [ ] Security report generated and approved
- [ ] All blocking issues resolved

---

## [S18-010] Production deployment and monitoring

**Priority**: URGENT | **Effort**: 8h | **Tags**: deployment, production, launch

### Solution Pseudocode

```
// Production Deployment Runbook

MODULE ProductionDeployment

  // Pre-Deployment Checklist
  CHECKLIST PreDeployment
    [_] All tests passing in CI/CD
    [_] Security audit approved
    [_] Load testing completed
    [_] UAT sign-off received
    [_] Rollback plan documented
    [_] Database backup completed
    [_] Team on standby for support
    [_] Communication sent to stakeholders
    [_] Monitoring dashboards ready
    [_] On-call schedule confirmed
  END CHECKLIST

  // Deployment Steps
  FUNCTION ExecuteDeployment()
    LOG "=== IRIS Production Deployment Started ==="
    LOG "Time: " + NOW()

    // Step 1: Final backup
    LOG "Step 1: Creating pre-deployment backup..."
    backupResult = CreateDatabaseBackup({
      name: "pre-deployment-" + DATE_FORMAT(NOW(), "YYYYMMDD-HHmmss"),
      includeFiles: TRUE
    })
    IF NOT backupResult.success THEN
      ABORT "Backup failed: " + backupResult.error
    END IF
    LOG "Backup created: " + backupResult.backupId

    // Step 2: Enable maintenance mode
    LOG "Step 2: Enabling maintenance mode..."
    SetMaintenanceMode(TRUE)

    // Step 3: Run database migrations
    LOG "Step 3: Running database migrations..."
    migrationResult = RunMigrations({
      direction: "up",
      dryRun: FALSE
    })
    IF NOT migrationResult.success THEN
      LOG "Migration failed, initiating rollback..."
      RollbackMigrations()
      SetMaintenanceMode(FALSE)
      ABORT "Migration failed: " + migrationResult.error
    END IF
    LOG "Migrations completed: " + migrationResult.migrationsRun + " applied"

    // Step 4: Deploy backend
    LOG "Step 4: Deploying backend services..."
    backendResult = DeployBackend({
      image: BACKEND_IMAGE_TAG,
      replicas: 3,
      strategy: "rolling",
      healthCheck: "/api/health"
    })
    IF NOT backendResult.healthy THEN
      LOG "Backend deployment failed, initiating rollback..."
      RollbackBackend()
      RollbackMigrations()
      SetMaintenanceMode(FALSE)
      ABORT "Backend deployment failed"
    END IF
    LOG "Backend deployed and healthy"

    // Step 5: Deploy frontends
    LOG "Step 5: Deploying frontend applications..."
    frontendResult = DeployFrontends({
      admin: { path: "/frontend", target: CDN_ADMIN },
      portal: { path: "/employee-portal", target: CDN_PORTAL }
    })
    IF NOT frontendResult.success THEN
      LOG "Frontend deployment failed, initiating rollback..."
      RollbackFrontends()
      RollbackBackend()
      RollbackMigrations()
      SetMaintenanceMode(FALSE)
      ABORT "Frontend deployment failed"
    END IF
    LOG "Frontends deployed"

    // Step 6: Warm up caches
    LOG "Step 6: Warming up caches..."
    WarmUpCaches([
      "/api/catalogs",
      "/api/payroll-concepts",
      "/api/incidence-types"
    ])

    // Step 7: Run smoke tests
    LOG "Step 7: Running smoke tests..."
    smokeResult = RunSmokeTests({
      endpoints: [
        { method: "GET", path: "/api/health", expectedStatus: 200 },
        { method: "POST", path: "/api/auth/login", body: TEST_CREDENTIALS, expectedStatus: 200 },
        { method: "GET", path: "/api/employees", auth: TRUE, expectedStatus: 200 }
      ]
    })
    IF NOT smokeResult.allPassed THEN
      LOG "Smoke tests failed, initiating rollback..."
      FullRollback()
      ABORT "Smoke tests failed: " + smokeResult.failures
    END IF
    LOG "Smoke tests passed: " + smokeResult.passed + "/" + smokeResult.total

    // Step 8: Disable maintenance mode
    LOG "Step 8: Disabling maintenance mode..."
    SetMaintenanceMode(FALSE)

    // Step 9: Verify production access
    LOG "Step 9: Verifying production access..."
    verifyResult = VerifyProductionAccess({
      urls: [
        PRODUCTION_ADMIN_URL,
        PRODUCTION_PORTAL_URL,
        PRODUCTION_API_URL
      ]
    })
    IF NOT verifyResult.allAccessible THEN
      LOG "WARNING: Some URLs not accessible"
      AlertOpsTeam(verifyResult)
    END IF

    LOG "=== IRIS Production Deployment Completed Successfully ==="
    LOG "Time: " + NOW()

    // Notify stakeholders
    SendNotification({
      channel: "deployment-alerts",
      message: "IRIS v2.0 deployed to production successfully",
      details: {
        version: VERSION,
        deployedAt: NOW(),
        deployedBy: CURRENT_USER
      }
    })

    RETURN DeploymentResult{
      success: TRUE,
      version: VERSION,
      deployedAt: NOW()
    }
  END FUNCTION

  // Post-Deployment Monitoring
  FUNCTION MonitorPostDeployment(durationMinutes)
    LOG "Starting post-deployment monitoring for " + durationMinutes + " minutes"

    metrics = {
      errorRate: [],
      responseTime: [],
      activeUsers: [],
      databaseConnections: []
    }

    startTime = NOW()
    WHILE (NOW() - startTime) < durationMinutes * 60 SECONDS
      // Collect metrics every 30 seconds
      currentMetrics = CollectMetrics()

      metrics.errorRate.ADD(currentMetrics.errorRate)
      metrics.responseTime.ADD(currentMetrics.avgResponseTime)
      metrics.activeUsers.ADD(currentMetrics.activeUsers)
      metrics.databaseConnections.ADD(currentMetrics.dbConnections)

      // Check for anomalies
      IF currentMetrics.errorRate > ERROR_RATE_THRESHOLD THEN
        AlertOpsTeam({
          severity: "high",
          message: "Error rate spike detected: " + currentMetrics.errorRate + "%",
          recommendation: "Consider rollback if persists"
        })
      END IF

      IF currentMetrics.avgResponseTime > RESPONSE_TIME_THRESHOLD THEN
        AlertOpsTeam({
          severity: "medium",
          message: "Response time degradation: " + currentMetrics.avgResponseTime + "ms",
          recommendation: "Check database and cache performance"
        })
      END IF

      WAIT 30 SECONDS
    END WHILE

    // Generate post-deployment report
    report = GeneratePostDeploymentReport(metrics)
    LOG "Post-deployment monitoring completed"
    LOG "Average error rate: " + report.avgErrorRate + "%"
    LOG "Average response time: " + report.avgResponseTime + "ms"
    LOG "Peak active users: " + report.peakActiveUsers

    RETURN report
  END FUNCTION

  // Rollback Procedure
  FUNCTION FullRollback()
    LOG "=== Initiating Full Rollback ==="

    SetMaintenanceMode(TRUE)

    // Rollback in reverse order
    RollbackFrontends()
    RollbackBackend()
    RollbackMigrations()
    RestoreFromBackup(LATEST_BACKUP_ID)

    SetMaintenanceMode(FALSE)

    AlertOpsTeam({
      severity: "critical",
      message: "Production rollback completed",
      action: "Investigate deployment failure"
    })

    LOG "=== Rollback Completed ==="
  END FUNCTION

END MODULE
```

### Acceptance Criteria
- [ ] Deployment executed without errors
- [ ] All smoke tests passing
- [ ] Monitoring shows no anomalies for 30 minutes
- [ ] Rollback procedure tested and documented
- [ ] Stakeholders notified
- [ ] Production access verified

---

# Appendix A: Technology Dependencies

## Backend Dependencies
- Go 1.24+
- Gin Framework
- GORM ORM
- golang-jwt
- bcrypt
- logrus
- testify (testing)

## Frontend Dependencies
- Next.js 14
- React 18
- TypeScript 5
- Tailwind CSS 4
- Radix UI
- React Hook Form
- Zod
- Recharts
- date-fns
- Jest + React Testing Library
- Playwright (E2E)

## Infrastructure
- Docker
- PostgreSQL (upgrade from SQLite for production)
- Redis (caching)
- Nginx
- Prometheus + Grafana

---

# Appendix B: Reference Links

## ClickUp CSV Import
- [Prepare a spreadsheet for import](https://help.clickup.com/hc/en-us/articles/6310821748759-Prepare-a-spreadsheet-for-import)
- [Fields supported by the Spreadsheets importer](https://help.clickup.com/hc/en-us/articles/6310876671255-Fields-supported-by-the-Spreadsheets-importer)

## Industry Benchmarks
- [Odoo HR Payroll Features](https://www.odoo.com/documentation/19.0/applications/hr/payroll.html)
- [Oracle HCM Cloud](https://www.oracle.com/human-capital-management/)
- [SAP SuccessFactors](https://www.sap.com/products/hcm.html)
- [Revo en la Nube](https://tress.com.mx/en/revolution-on-the-cloud/)

## Mexican Payroll Compliance
- [ISR Tables 2025](https://www.sat.gob.mx)
- [IMSS Contribution Rates](https://www.imss.gob.mx)
- [CFDI 4.0 Specification](https://www.sat.gob.mx/consultas/35025/formato-de-factura-electronica-(anexo-20))

---

**Document Version**: 2.0
**Last Updated**: January 2, 2026
**Next Review**: Before each sprint start
