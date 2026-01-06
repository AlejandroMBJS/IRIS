/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/message_handler.go
==============================================================================

DESCRIPTION:
    HTTP handlers for internal messaging system endpoints. Provides REST API
    for inbox, sending messages, marking as read, and asking questions about
    announcements.

ENDPOINTS:
    GET    /messages              - Get inbox (received messages)
    GET    /messages/sent         - Get sent messages
    GET    /messages/unread-count - Get unread message count
    GET    /messages/:id          - Get single message with thread
    POST   /messages              - Send a new message
    POST   /messages/:id/read     - Mark message as read
    POST   /messages/:id/unread   - Mark message as unread
    POST   /messages/:id/archive  - Archive a message
    DELETE /messages/:id          - Delete a message
    GET    /messages/recipients   - Get recipient suggestions
    POST   /messages/announcement-question - Ask question about announcement

==============================================================================
*/
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
)

// MessageHandler handles message endpoints
type MessageHandler struct {
	messageService *services.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// RegisterRoutes registers message routes
func (h *MessageHandler) RegisterRoutes(router *gin.RouterGroup) {
	messages := router.Group("/messages")
	{
		messages.GET("", h.GetInbox)
		messages.GET("/sent", h.GetSentMessages)
		messages.GET("/unread-count", h.GetUnreadCount)
		messages.GET("/recipients", h.GetRecipientSuggestions)
		messages.GET("/:id", h.GetMessage)
		messages.POST("", h.SendMessage)
		messages.POST("/announcement-question", h.AskAnnouncementQuestion)
		messages.POST("/:id/read", h.MarkAsRead)
		messages.POST("/:id/unread", h.MarkAsUnread)
		messages.POST("/:id/archive", h.ArchiveMessage)
		messages.POST("/:id/reply", h.ReplyToMessage)
		messages.DELETE("/:id", h.DeleteMessage)
	}
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	RecipientID string `json:"recipient_id" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	Body        string `json:"body" binding:"required"`
}

// AskQuestionRequest represents the request body for asking a question about an announcement
type AskQuestionRequest struct {
	AnnouncementID string `json:"announcement_id" binding:"required"`
	Question       string `json:"question" binding:"required"`
}

// ReplyRequest represents the request body for replying to a message
type ReplyRequest struct {
	Body string `json:"body" binding:"required"`
}

// GetInbox handles getting inbox messages
func (h *MessageHandler) GetInbox(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.DefaultQuery("status", "all")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	messages, total, err := h.messageService.GetInbox(userID, companyID, page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":    messages,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetSentMessages handles getting sent messages
func (h *MessageHandler) GetSentMessages(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	messages, total, err := h.messageService.GetSentMessages(userID, companyID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":    messages,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetMessage handles getting a single message
func (h *MessageHandler) GetMessage(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	message, err := h.messageService.GetMessage(messageID, userID, companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Auto-mark as read if recipient is viewing
	if message.RecipientID == userID && message.Status == models.MessageStatusUnread {
		_ = h.messageService.MarkAsRead(messageID, userID, companyID)
		message.Status = models.MessageStatusRead
	}

	c.JSON(http.StatusOK, message)
}

// SendMessage handles sending a new message
func (h *MessageHandler) SendMessage(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
		return
	}

	message, err := h.messageService.SendMessage(
		userID,
		recipientID,
		companyID,
		req.Subject,
		req.Body,
		models.MessageTypeDirect,
		nil,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// AskAnnouncementQuestion handles asking a question about an announcement
func (h *MessageHandler) AskAnnouncementQuestion(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req AskQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcementID, err := uuid.Parse(req.AnnouncementID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	// Get the announcement creator to send the question to
	creator, err := h.messageService.GetAnnouncementCreator(announcementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.SendMessage(
		userID,
		creator.ID,
		companyID,
		"Pregunta sobre anuncio",
		req.Question,
		models.MessageTypeAnnouncementQuestion,
		&announcementID,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// ReplyToMessage handles replying to a message
func (h *MessageHandler) ReplyToMessage(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	parentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req ReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the parent message to determine recipient
	parentMsg, err := h.messageService.GetMessage(parentID, userID, companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Determine the recipient (the other party in the conversation)
	var recipientID uuid.UUID
	if parentMsg.SenderID == userID {
		recipientID = parentMsg.RecipientID
	} else {
		recipientID = parentMsg.SenderID
	}

	// Get the root message ID if this is a reply to a reply
	rootID := parentID
	if parentMsg.ParentID != nil {
		rootID = *parentMsg.ParentID
	}

	message, err := h.messageService.SendMessage(
		userID,
		recipientID,
		companyID,
		"Re: "+parentMsg.Subject,
		req.Body,
		parentMsg.Type,
		parentMsg.AnnouncementID,
		&rootID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// MarkAsRead handles marking a message as read
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.messageService.MarkAsRead(messageID, userID, companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

// MarkAsUnread handles marking a message as unread
func (h *MessageHandler) MarkAsUnread(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.messageService.MarkAsUnread(messageID, userID, companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as unread"})
}

// ArchiveMessage handles archiving a message
func (h *MessageHandler) ArchiveMessage(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.messageService.ArchiveMessage(messageID, userID, companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message archived"})
}

// DeleteMessage handles deleting a message
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.messageService.DeleteMessage(messageID, userID, companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

// GetUnreadCount handles getting unread message count
func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := h.messageService.GetUnreadCount(userID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// GetRecipientSuggestions handles getting recipient suggestions
func (h *MessageHandler) GetRecipientSuggestions(c *gin.Context) {
	userID, _, companyID, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")

	users, err := h.messageService.GetRecipientSuggestions(userID, companyID, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
