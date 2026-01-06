/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/audit_handler.go
==============================================================================

DESCRIPTION:
    Handles all audit log related endpoints: viewing logs, login attempts,
    login history, active sessions, and activity statistics.

USER PERSPECTIVE:
    - View detailed audit logs of all system activity
    - Track login attempts (successful and failed)
    - Monitor active user sessions
    - View user activity statistics

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new audit-related endpoints
    ⚠️  CAUTION: Access control - only admins should see all logs
    ❌  DO NOT modify: Core audit log structure
    📝  All responses use consistent pagination format

ENDPOINTS:
    GET  /audit/logs - Get audit logs with filters (admin only)
    GET  /audit/login-attempts - Get login attempts (admin only)
    GET  /audit/login-history - Get login history (admin or self)
    GET  /audit/active-sessions - Get active sessions (admin or self)
    GET  /audit/page-visits - Get page visits (admin or self)
    GET  /audit/stats - Get activity statistics (admin only)
    GET  /audit/user/:user_id/stats - Get user activity stats (admin or self)

==============================================================================
*/
package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/services"
)

// AuditHandler handles audit log endpoints
type AuditHandler struct {
	auditService *services.AuditService
	authService  *services.AuthService
}

// NewAuditHandler creates new audit handler
func NewAuditHandler(auditService *services.AuditService, authService *services.AuthService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		authService:  authService,
	}
}

// RegisterRoutes registers audit routes
func (h *AuditHandler) RegisterRoutes(router *gin.RouterGroup) {
	authMiddleware := middleware.NewAuthMiddleware(h.authService)
	adminMiddleware := middleware.NewRoleMiddleware("admin")

	audit := router.Group("/audit")
	audit.Use(authMiddleware.RequireAuth())
	{
		// Admin only endpoints
		audit.GET("/logs", adminMiddleware.RequireRole(), h.GetAuditLogs)
		audit.GET("/login-attempts", adminMiddleware.RequireRole(), h.GetLoginAttempts)
		audit.GET("/stats", adminMiddleware.RequireRole(), h.GetGlobalStats)

		// User can access their own data, admin can access all
		audit.GET("/login-history", h.GetLoginHistory)
		audit.GET("/active-sessions", h.GetActiveSessions)
		audit.GET("/page-visits", h.GetPageVisits)
		audit.GET("/user/:user_id/stats", h.GetUserStats)
	}
}

// GetAuditLogs retrieves audit logs with filters
// @Summary Get audit logs
// @Description Get audit logs with optional filters (admin only)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param user_id query string false "Filter by user ID"
// @Param email query string false "Filter by email"
// @Param event_type query string false "Filter by event type"
// @Param success query bool false "Filter by success status"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/logs [get]
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Parse filters
	filters := make(map[string]interface{})

	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if email := c.Query("email"); email != "" {
		filters["email"] = email
	}
	if eventType := c.Query("event_type"); eventType != "" {
		filters["event_type"] = eventType
	}
	if successStr := c.Query("success"); successStr != "" {
		success, _ := strconv.ParseBool(successStr)
		filters["success"] = success
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	// Get logs
	logs, total, err := h.auditService.GetAuditLogs(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve audit logs",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetLoginAttempts retrieves login attempts
// @Summary Get login attempts
// @Description Get all login attempts (admin only)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/login-attempts [get]
func (h *AuditHandler) GetLoginAttempts(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Get login attempts
	attempts, total, err := h.auditService.GetLoginAttempts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve login attempts",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": attempts,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetLoginHistory retrieves login history
// @Summary Get login history
// @Description Get login history for current user or specified user (admin can see all)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param user_id query string false "User ID (admin only)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/login-history [get]
func (h *AuditHandler) GetLoginHistory(c *gin.Context) {
	// Get current user
	currentUserID, _, _, err := middleware.GetUserFromContext(c)
	role := middleware.GetUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Determine which user's history to get
	var targetUserID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		// Only admin can view other users' history
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You can only view your own login history",
			})
			return
		}
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user ID",
				"message": err.Error(),
			})
			return
		}
		targetUserID = &parsedID
	} else {
		// View own history
		targetUserID = &currentUserID
	}

	// Get login history
	history, total, err := h.auditService.GetLoginHistory(targetUserID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve login history",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetActiveSessions retrieves active sessions
// @Summary Get active sessions
// @Description Get active sessions for current user or all users (admin only)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "User ID (admin only)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/active-sessions [get]
func (h *AuditHandler) GetActiveSessions(c *gin.Context) {
	// Get current user
	currentUserID, _, _, err := middleware.GetUserFromContext(c)
	role := middleware.GetUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Determine which user's sessions to get
	var targetUserID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		// Only admin can view other users' sessions
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You can only view your own active sessions",
			})
			return
		}
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user ID",
				"message": err.Error(),
			})
			return
		}
		targetUserID = &parsedID
	} else {
		// View own sessions
		targetUserID = &currentUserID
	}

	// Get active sessions
	sessions, err := h.auditService.GetActiveSessions(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve active sessions",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sessions,
	})
}

// GetPageVisits retrieves page visits
// @Summary Get page visits
// @Description Get page visits for current user or specified user (admin can see all)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param user_id query string false "User ID (admin only)"
// @Param session_id query string false "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/page-visits [get]
func (h *AuditHandler) GetPageVisits(c *gin.Context) {
	// Get current user
	currentUserID, _, _, err := middleware.GetUserFromContext(c)
	role := middleware.GetUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Determine which user's page visits to get
	targetUserID := currentUserID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		// Only admin can view other users' page visits
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You can only view your own page visits",
			})
			return
		}
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user ID",
				"message": err.Error(),
			})
			return
		}
		targetUserID = parsedID
	}

	// Get session ID filter if provided
	var sessionID *string
	if sid := c.Query("session_id"); sid != "" {
		sessionID = &sid
	}

	// Get page visits
	visits, total, err := h.auditService.GetPageVisits(targetUserID, sessionID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve page visits",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": visits,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetGlobalStats retrieves global activity statistics
// @Summary Get global statistics
// @Description Get global activity statistics (admin only)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/stats [get]
func (h *AuditHandler) GetGlobalStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	stats, err := h.auditService.GetLoginAttemptsStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserStats retrieves user activity statistics
// @Summary Get user statistics
// @Description Get user activity statistics (admin or self)
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/user/{user_id}/stats [get]
func (h *AuditHandler) GetUserStats(c *gin.Context) {
	// Get current user
	currentUserID, _, _, err := middleware.GetUserFromContext(c)
	role := middleware.GetUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Parse target user ID
	userIDStr := c.Param("user_id")
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": err.Error(),
		})
		return
	}

	// Check permissions
	if role != "admin" && currentUserID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You can only view your own statistics",
		})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	stats, err := h.auditService.GetUserActivityStats(targetUserID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
