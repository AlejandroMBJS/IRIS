/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/health_handler.go
==============================================================================

DESCRIPTION:
    Handles health check endpoints for monitoring and container orchestration.
    Provides liveness, readiness, and general health status.

USER PERSPECTIVE:
    - Not user-facing - used by infrastructure
    - Kubernetes/Docker health checks
    - Load balancer health monitoring

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add more health checks (Redis, etc.)
    ⚠️  CAUTION: ReadyCheck database ping
    ❌  DO NOT modify: Response format (breaks monitoring)
    📝  Keep health checks lightweight and fast

SYNTAX EXPLANATION:
    - HealthCheck: Basic "is the service running" check
    - ReadyCheck: "Is the service ready to handle requests"
    - LivenessCheck: "Is the process alive" (for restart decisions)

ENDPOINTS:
    GET /health - General health status
    GET /ready - Database connectivity check
    GET /live - Process liveness check

KUBERNETES USAGE:
    livenessProbe:
      httpGet:
        path: /api/v1/live
    readinessProbe:
      httpGet:
        path: /api/v1/ready

==============================================================================
*/
package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "iris-payroll-backend",
	})
}

func (h *HealthHandler) ReadyCheck(c *gin.Context) {
	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"database": "unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"database": "available",
	})
}

func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "live",
	})
}
