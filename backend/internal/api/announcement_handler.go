/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/announcement_handler.go
==============================================================================

DESCRIPTION:
    HTTP handlers for announcement endpoints. Provides REST API for creating,
    reading, and managing company announcements.

ENDPOINTS:
    GET    /announcements - List active announcements
    GET    /announcements/:id - Get announcement details
    POST   /announcements - Create announcement
    DELETE /announcements/:id - Delete announcement
    POST   /announcements/:id/read - Mark announcement as read
    GET    /announcements/unread-count - Get unread count

==============================================================================
*/
package api

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/services"
)

// AnnouncementHandler handles announcement endpoints
type AnnouncementHandler struct {
	announcementService *services.AnnouncementService
}

// NewAnnouncementHandler creates a new announcement handler
func NewAnnouncementHandler(announcementService *services.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
	}
}

// RegisterRoutes registers announcement routes
func (h *AnnouncementHandler) RegisterRoutes(router *gin.RouterGroup) {
	announcements := router.Group("/announcements")
	{
		announcements.GET("", h.ListAnnouncements)
		announcements.GET("/:id", h.GetAnnouncement)
		announcements.POST("", h.CreateAnnouncement)
		announcements.DELETE("/:id", h.DeleteAnnouncement)
		announcements.POST("/:id/read", h.MarkAsRead)
		announcements.GET("/unread-count", h.GetUnreadCount)
	}
}

// CreateAnnouncementRequest represents the request body for creating an announcement
type CreateAnnouncementRequest struct {
	Title       string  `json:"title" binding:"required"`
	Message     string  `json:"message" binding:"required"`
	Scope       string  `json:"scope" binding:"required"` // ALL, DEPARTMENT, etc.
	ImageBase64 string  `json:"image_base64,omitempty"`
	ExpiresIn   *int    `json:"expires_in_days,omitempty"` // Days until expiration
}

// ListAnnouncements handles listing announcements
// @Summary List announcements
// @Description Get all active, non-expired announcements
// @Tags Announcements
// @Produce json
// @Success 200 {array} models.Announcement
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /announcements [get]
func (h *AnnouncementHandler) ListAnnouncements(c *gin.Context) {
	// Get user from context
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	announcements, err := h.announcementService.GetActiveAnnouncements(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": announcements,
	})
}

// GetAnnouncement handles getting a single announcement
// @Summary Get announcement
// @Description Get announcement details by ID
// @Tags Announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} models.Announcement
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /announcements/{id} [get]
func (h *AnnouncementHandler) GetAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Invalid announcement ID format",
		})
		return
	}

	announcement, err := h.announcementService.GetAnnouncementByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, announcement)
}

// CreateAnnouncement handles creating a new announcement
// @Summary Create announcement
// @Description Create a new company announcement
// @Tags Announcements
// @Accept json
// @Produce json
// @Param request body CreateAnnouncementRequest true "Announcement data"
// @Success 201 {object} models.Announcement
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /announcements [post]
func (h *AnnouncementHandler) CreateAnnouncement(c *gin.Context) {
	// Get user from context
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Get user role
	role := middleware.GetUserRoleFromContext(c)

	// Check if user has permission to create announcements
	allowedRoles := map[string]bool{
		"admin":      true,
		"hr":         true,
		"hr_and_pr":  true,
		"manager":    true,
		"supervisor": true,
		"sup_and_gm": true,
	}

	if !allowedRoles[role] {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to create announcements",
		})
		return
	}

	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation Error",
			"message": err.Error(),
		})
		return
	}

	// Validate scope based on role:
	// - Supervisors and managers can only create TEAM scope (for their subordinates)
	// - HR and admin can create any scope (ALL, COMPANY, TEAM)
	canCreateAllScope := map[string]bool{
		"admin":     true,
		"hr":        true,
		"hr_and_pr": true,
	}

	// Normalize scope - treat ALL and COMPANY as equivalent for company-wide
	normalizedScope := strings.ToUpper(req.Scope)
	isCompanyWide := normalizedScope == "ALL" || normalizedScope == "COMPANY"

	if isCompanyWide && !canCreateAllScope[role] {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Solo HR y administradores pueden crear anuncios para toda la empresa. Como supervisor/gerente, solo puede crear anuncios para su equipo (TEAM).",
		})
		return
	}

	// Decode image if provided
	var imageData []byte
	if req.ImageBase64 != "" {
		// Remove data URI prefix if present
		imageStr := req.ImageBase64
		if strings.Contains(imageStr, ",") {
			parts := strings.Split(imageStr, ",")
			if len(parts) == 2 {
				imageStr = parts[1]
			}
		}

		decoded, err := base64.StdEncoding.DecodeString(imageStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid Image",
				"message": "Failed to decode image data",
			})
			return
		}
		imageData = decoded
	}

	// Calculate expiration date
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expires := time.Now().AddDate(0, 0, *req.ExpiresIn)
		expiresAt = &expires
	}

	announcement, err := h.announcementService.CreateAnnouncement(
		req.Title,
		req.Message,
		req.Scope,
		userID,
		imageData,
		expiresAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, announcement)
}

// DeleteAnnouncement handles deleting an announcement
// @Summary Delete announcement
// @Description Delete an announcement (soft delete)
// @Tags Announcements
// @Param id path string true "Announcement ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /announcements/{id} [delete]
func (h *AnnouncementHandler) DeleteAnnouncement(c *gin.Context) {
	// Get user from context
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Invalid announcement ID format",
		})
		return
	}

	err = h.announcementService.DeleteAnnouncement(id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": err.Error(),
			})
		} else if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Announcement deleted successfully",
	})
}

// MarkAsRead handles marking an announcement as read
// @Summary Mark announcement as read
// @Description Mark an announcement as read by the current user
// @Tags Announcements
// @Param id path string true "Announcement ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /announcements/{id}/read [post]
func (h *AnnouncementHandler) MarkAsRead(c *gin.Context) {
	// Get user from context
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Invalid announcement ID format",
		})
		return
	}

	err = h.announcementService.MarkAnnouncementAsRead(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Announcement marked as read",
	})
}

// GetUnreadCount handles getting unread announcement count
// @Summary Get unread count
// @Description Get the count of unread announcements for the current user
// @Tags Announcements
// @Produce json
// @Success 200 {object} map[string]int64
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /announcements/unread-count [get]
func (h *AnnouncementHandler) GetUnreadCount(c *gin.Context) {
	// Get user from context
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	count, err := h.announcementService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}
