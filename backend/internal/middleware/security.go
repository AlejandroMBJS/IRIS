/*
Package middleware - Security Headers Middleware

==============================================================================
FILE: internal/middleware/security.go
==============================================================================

DESCRIPTION:
    Provides HTTP security headers for defense-in-depth protection.
    Implements industry-standard security headers including:
    - Content-Security-Policy (CSP): Prevents XSS and data injection attacks
    - X-Frame-Options: Prevents clickjacking
    - X-Content-Type-Options: Prevents MIME type sniffing
    - Referrer-Policy: Controls referrer information
    - Permissions-Policy: Restricts browser features

USER PERSPECTIVE:
    - Transparent security layer that hardens the application
    - May cause issues if third-party scripts/styles are needed
    - Protects against common web vulnerabilities

DEVELOPER GUIDELINES:
    OK to modify: CSP directives to allow specific sources
    CAUTION: Loosening CSP too much reduces security
    DO NOT modify: X-Content-Type-Options or X-Frame-Options without security review

==============================================================================
*/
package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
)

// SecurityMiddleware provides HTTP security headers
type SecurityMiddleware struct {
	appConfig *config.AppConfig
}

// NewSecurityMiddleware creates a new security headers middleware
func NewSecurityMiddleware(appConfig *config.AppConfig) *SecurityMiddleware {
	return &SecurityMiddleware{appConfig: appConfig}
}

// Headers returns a Gin middleware that sets security headers
func (m *SecurityMiddleware) Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content-Security-Policy
		// Restricts sources for scripts, styles, images, fonts, etc.
		csp := m.buildCSP()
		c.Header("Content-Security-Policy", csp)

		// X-Frame-Options: Prevent clickjacking by disallowing framing
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Referrer-Policy: Control referrer information sent with requests
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// X-XSS-Protection: Enable browser XSS filter (legacy, but still useful)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Permissions-Policy: Restrict browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Strict-Transport-Security (HSTS): Force HTTPS in production
		if m.appConfig.IsProduction() {
			// max-age=31536000 = 1 year, includeSubDomains for all subdomains
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// buildCSP constructs the Content-Security-Policy header value
func (m *SecurityMiddleware) buildCSP() string {
	// Base CSP directives
	directives := []string{
		// Default: Only allow resources from same origin
		"default-src 'self'",

		// Scripts: Allow self and inline for Next.js hydration
		// Note: 'unsafe-inline' is needed for Next.js, consider using nonce in future
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'",

		// Styles: Allow self and inline for Tailwind/styled-components
		"style-src 'self' 'unsafe-inline'",

		// Images: Allow self, data URIs, and blob for file previews
		"img-src 'self' data: blob:",

		// Fonts: Allow self and common CDNs
		"font-src 'self' data:",

		// Connect: Allow API connections to self
		"connect-src 'self'",

		// Forms: Only submit to same origin
		"form-action 'self'",

		// Frame ancestors: Prevent embedding in iframes
		"frame-ancestors 'none'",

		// Base URI: Restrict base tag
		"base-uri 'self'",

		// Object: Block Flash and other plugins
		"object-src 'none'",

		// Upgrade insecure requests in production
	}

	// Add upgrade-insecure-requests in production
	if m.appConfig.IsProduction() {
		directives = append(directives, "upgrade-insecure-requests")
	}

	// Join all directives with semicolons
	csp := ""
	for i, directive := range directives {
		if i > 0 {
			csp += "; "
		}
		csp += directive
	}

	return csp
}

// ReportOnlyHeaders returns CSP in report-only mode for testing
// Use this to test CSP changes before enforcing them
func (m *SecurityMiddleware) ReportOnlyHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		csp := m.buildCSP()
		// Report-Only mode logs violations but doesn't block
		c.Header("Content-Security-Policy-Report-Only", csp)

		// Also add a report-uri if you have a CSP violation endpoint
		if m.appConfig.IsProduction() {
			reportURI := fmt.Sprintf("; report-uri /api/v1/csp-report")
			c.Header("Content-Security-Policy-Report-Only", csp+reportURI)
		}

		c.Next()
	}
}
