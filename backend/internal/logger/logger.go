/*
Package logger - Structured logging configuration and HTTP request logging

==============================================================================
FILE: internal/logger/logger.go
==============================================================================

DESCRIPTION:
    This file configures structured logging using logrus for the application.
    It provides environment-based log level configuration and Gin middleware
    for HTTP request/response logging with rich contextual information including
    latency, status codes, client IPs, and error details.

USER PERSPECTIVE:
    - Application logs provide detailed debugging and monitoring information
    - Production logs are optimized (Info level) to reduce noise while development
      logs (Debug level) provide verbose details for troubleshooting
    - HTTP request logs help track API usage patterns and diagnose issues
    - JSON format enables easy parsing by log aggregation tools

DEVELOPER GUIDELINES:
    ✅  OK to modify: Log output destination (stdout, file, log aggregator)
    ✅  OK to modify: Add custom log fields for specific application needs
    ✅  OK to modify: Adjust log levels per environment (staging, production, etc.)
    ⚠️  CAUTION: Changing formatter type - ensure downstream tools can parse the format
    ⚠️  CAUTION: Adding PII to logs - ensure compliance with privacy regulations
    ❌  DO NOT modify: Core logrus fields structure without updating log parsers
    ❌  DO NOT log: Sensitive data (passwords, tokens, SSNs, credit cards)
    📝  Use structured fields (log.WithFields) instead of string concatenation
    📝  Set appropriate log levels: Error (500+), Warn (400+), Info (200-399)

SYNTAX EXPLANATION:
    - logrus.New(): Creates a new logger instance with default configuration
    - JSONFormatter: Outputs logs in JSON format for structured log processing
    - log.SetLevel(): Controls which log messages are output based on severity
    - log.WithFields(): Adds structured key-value pairs to log entries
    - gin.HandlerFunc: Middleware function that intercepts HTTP requests
    - time.Since(start): Calculates request latency by measuring elapsed time
    - c.ClientIP(): Extracts client IP address handling X-Forwarded-For headers
    - c.Errors.ByType(): Retrieves errors of a specific type from Gin's error collection

LOG LEVELS (from most to least severe):
    - Error: System errors, failed operations (500+ status codes)
    - Warn: Potential issues, client errors (400-499 status codes)
    - Info: Normal operations, successful requests (200-399 status codes)
    - Debug: Detailed debugging information (development only)

EXAMPLE LOG OUTPUT (JSON):
    {
        "level": "info",
        "msg": "",
        "latency": 45000000,
        "method": "POST",
        "status": 201,
        "ip": "192.168.1.100",
        "uri": "/api/v1/employees",
        "user_agent": "Mozilla/5.0...",
        "errors": "",
        "time": "2025-12-08T10:30:45Z"
    }

==============================================================================
*/

package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Setup initializes the logger with a given environment.
func Setup(env string) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout) // Or to a file, e.g., os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if env == "production" {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.DebugLevel)
	}

	return log
}

// GinLogger returns a gin.HandlerFunc for logging HTTP requests.
func GinLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		entry := log.WithFields(logrus.Fields{
			"latency":    time.Since(start),
			"method":     c.Request.Method,
			"status":     c.Writer.Status(),
			"ip":         c.ClientIP(),
			"uri":        path,
			"user_agent": c.Request.UserAgent(),
			"errors":     c.Errors.ByType(gin.ErrorTypePrivate).String(),
		})

		if c.Writer.Status() >= 500 {
			entry.Error()
		} else if c.Writer.Status() >= 400 {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
