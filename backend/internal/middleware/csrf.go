/*
Package middleware - CSRF Protection Middleware

==============================================================================
FILE: internal/middleware/csrf.go
==============================================================================

DESCRIPTION:
    Implements Cross-Site Request Forgery (CSRF) protection using the
    double-submit cookie pattern. This middleware:
    1. Sets a CSRF token in a JavaScript-readable cookie
    2. Validates that state-changing requests include the token in a header
    3. Compares the cookie and header values to prevent CSRF attacks

USER PERSPECTIVE:
    - Transparent security layer that protects against CSRF attacks
    - Frontend automatically includes CSRF token in requests
    - No user action required for normal operation

DEVELOPER GUIDELINES:
    OK to modify: Token length, cookie settings, header name
    CAUTION: Exempt paths list - only truly safe endpoints should be exempt
    DO NOT modify: Core validation logic without security review

SECURITY NOTES:
    - CSRF token cookie is NOT httpOnly (frontend must read it)
    - Token is cryptographically random (32 bytes)
    - Only state-changing methods (POST, PUT, DELETE, PATCH) require validation
    - Login/register endpoints are exempt (no session to protect yet)

==============================================================================
*/
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
)

const (
	// CSRFTokenCookieName is the name of the cookie containing the CSRF token
	CSRFTokenCookieName = "csrf_token"

	// CSRFTokenHeaderName is the name of the header where frontend sends the token
	CSRFTokenHeaderName = "X-CSRF-Token"

	// CSRFTokenLength is the number of bytes in the CSRF token
	CSRFTokenLength = 32
)

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	appConfig *config.AppConfig
}

// NewCSRFMiddleware creates a new CSRF middleware instance
func NewCSRFMiddleware(appConfig *config.AppConfig) *CSRFMiddleware {
	return &CSRFMiddleware{appConfig: appConfig}
}

// Protect returns a Gin middleware that provides CSRF protection
func (m *CSRFMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or retrieve CSRF token
		token := m.getOrCreateToken(c)

		// Set the CSRF token cookie on every response
		// Note: This cookie is NOT httpOnly so JavaScript can read it
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie(
			CSRFTokenCookieName,
			token,
			3600, // 1 hour
			"/",
			"",
			m.appConfig.IsProduction(), // Secure in production
			false,                      // NOT httpOnly - JS must read it
		)

		// Skip validation for safe HTTP methods
		if m.isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Skip validation for exempt paths (login, register, etc.)
		if m.isExemptPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Validate CSRF token for state-changing requests
		headerToken := c.GetHeader(CSRFTokenHeaderName)
		cookieToken, err := c.Cookie(CSRFTokenCookieName)

		if err != nil || headerToken == "" || headerToken != cookieToken {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Invalid or missing CSRF token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getOrCreateToken retrieves existing CSRF token from cookie or creates a new one
func (m *CSRFMiddleware) getOrCreateToken(c *gin.Context) string {
	// Try to get existing token from cookie
	if token, err := c.Cookie(CSRFTokenCookieName); err == nil && token != "" {
		return token
	}

	// Generate new token
	return generateCSRFToken()
}

// isSafeMethod returns true for HTTP methods that don't require CSRF protection
func (m *CSRFMiddleware) isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// isExemptPath returns true for paths that don't require CSRF validation
// These are endpoints that either:
// 1. Don't have an authenticated session yet (login, register)
// 2. Are stateless API endpoints that use other auth mechanisms
func (m *CSRFMiddleware) isExemptPath(path string) bool {
	exemptPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
		"/api/v1/health",
	}

	for _, exempt := range exemptPaths {
		if strings.HasPrefix(path, exempt) {
			return true
		}
	}

	return false
}

// generateCSRFToken creates a cryptographically secure random token
func generateCSRFToken() string {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback - this should never happen in practice
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
