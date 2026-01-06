/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/notification_handler.go
==============================================================================

DESCRIPTION:
    Handles notification endpoints. Users can retrieve their notifications,
    mark them as read, and get unread counts for the navbar badge.

USER PERSPECTIVE:
    - View list of recent notifications
    - See unread count in bell icon
    - Mark individual or all notifications as read

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add pagination, filtering
    ⚠️  CAUTION: Performance with large notification counts
    ❌  DO NOT modify: Company isolation (users only see company notifications)
    📝  Notifications auto-exclude user's own actions

ENDPOINTS:
    GET    /notifications - List recent notifications
    GET    /notifications/unread-count - Get unread count
    POST   /notifications/:id/read - Mark one as read
    POST   /notifications/read-all - Mark all as read

==============================================================================
*/
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/services"
)

// NotificationHandler handles notification endpoints
type NotificationHandler struct {
	service *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// RegisterRoutes registers notification routes
func (h *NotificationHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/notifications")
	{
		group.GET("", h.GetNotifications)
		group.GET("/unread-count", h.GetUnreadCount)
		group.POST("/:id/read", h.MarkAsRead)
		group.POST("/read-all", h.MarkAllAsRead)
	}
}

// GetNotifications returns recent notifications for the current user
// GET /api/v1/notifications?limit=20
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse limit from query param (default 20)
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	notifications, err := h.service.GetNotificationsForUser(userID, companyID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get notifications",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
	})
}

// GetUnreadCount returns the count of unread notifications
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := h.service.GetUnreadCount(userID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get unread count",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// MarkAsRead marks a notification as read
// POST /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid notification ID",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.MarkAsRead(notificationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark notification as read",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead marks all notifications as read
// POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.service.MarkAllAsRead(userID, companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark all notifications as read",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}
