/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/notification_service.go
==============================================================================

DESCRIPTION:
    Business logic for notification management. Creates notifications when
    company events occur and retrieves them for users.

USER PERSPECTIVE:
    - Automatic notifications for changes by other users
    - Unread count in navbar badge
    - Mark notifications as read

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new notification creation methods
    ⚠️  CAUTION: Performance of notification queries
    ❌  DO NOT modify: Exclusion of actor from notifications
    📝  Call CreateNotification from other services when events occur

==============================================================================
*/
package services

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// NotificationService handles notification business logic
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification creates a new notification for the company
func (s *NotificationService) CreateNotification(
	companyID uuid.UUID,
	actorUserID uuid.UUID,
	notifType models.NotificationType,
	title string,
	message string,
	resourceType string,
	resourceID *uuid.UUID,
) error {
	notification := models.Notification{
		CompanyID:    companyID,
		ActorUserID:  actorUserID,
		Type:         notifType,
		Title:        title,
		Message:      message,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now(),
	}

	return s.db.Create(&notification).Error
}

// GetNotificationsForUser returns notifications for a user (excluding their own actions)
func (s *NotificationService) GetNotificationsForUser(userID, companyID uuid.UUID, limit int) ([]models.NotificationWithReadStatus, error) {
	if limit <= 0 {
		limit = 20
	}

	var notifications []models.Notification

	// Get recent notifications for the company, excluding those created by this user
	err := s.db.
		Where("company_id = ? AND actor_user_id != ?", companyID, userID).
		Order("created_at DESC").
		Limit(limit).
		Preload("ActorUser").
		Find(&notifications).Error

	if err != nil {
		return nil, err
	}

	// Get read status for these notifications
	notifIDs := make([]uuid.UUID, len(notifications))
	for i, n := range notifications {
		notifIDs[i] = n.ID
	}

	var readRecords []models.NotificationRead
	if len(notifIDs) > 0 {
		s.db.Where("notification_id IN ? AND user_id = ?", notifIDs, userID).Find(&readRecords)
	}

	// Create a map for quick lookup
	readMap := make(map[uuid.UUID]time.Time)
	for _, r := range readRecords {
		readMap[r.NotificationID] = r.ReadAt
	}

	// Build response with read status
	result := make([]models.NotificationWithReadStatus, len(notifications))
	for i, n := range notifications {
		readAt, isRead := readMap[n.ID]
		result[i] = models.NotificationWithReadStatus{
			Notification: n,
			Read:         isRead,
		}
		if isRead {
			result[i].ReadAt = &readAt
		}
	}

	return result, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID, companyID uuid.UUID) (int64, error) {
	var totalCount int64
	var readCount int64

	// Count total notifications for this company (excluding user's own)
	err := s.db.Model(&models.Notification{}).
		Where("company_id = ? AND actor_user_id != ?", companyID, userID).
		Count(&totalCount).Error
	if err != nil {
		return 0, err
	}

	// Count read notifications
	err = s.db.Model(&models.NotificationRead{}).
		Joins("JOIN notifications ON notifications.id = notification_reads.notification_id").
		Where("notifications.company_id = ? AND notifications.actor_user_id != ? AND notification_reads.user_id = ?",
			companyID, userID, userID).
		Count(&readCount).Error
	if err != nil {
		return 0, err
	}

	return totalCount - readCount, nil
}

// MarkAsRead marks a notification as read for a user
func (s *NotificationService) MarkAsRead(notificationID, userID uuid.UUID) error {
	// Check if already read
	var existing models.NotificationRead
	err := s.db.Where("notification_id = ? AND user_id = ?", notificationID, userID).First(&existing).Error
	if err == nil {
		// Already read
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Create read record
	readRecord := models.NotificationRead{
		NotificationID: notificationID,
		UserID:         userID,
		ReadAt:         time.Now(),
	}

	return s.db.Create(&readRecord).Error
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID, companyID uuid.UUID) error {
	// Get all unread notifications for this user
	var notifications []models.Notification
	err := s.db.
		Where("company_id = ? AND actor_user_id != ?", companyID, userID).
		Find(&notifications).Error
	if err != nil {
		return err
	}

	// Get existing read records
	notifIDs := make([]uuid.UUID, len(notifications))
	for i, n := range notifications {
		notifIDs[i] = n.ID
	}

	var existingReads []models.NotificationRead
	if len(notifIDs) > 0 {
		s.db.Where("notification_id IN ? AND user_id = ?", notifIDs, userID).Find(&existingReads)
	}

	existingMap := make(map[uuid.UUID]bool)
	for _, r := range existingReads {
		existingMap[r.NotificationID] = true
	}

	// Create read records for unread notifications
	now := time.Now()
	var newReads []models.NotificationRead
	for _, n := range notifications {
		if !existingMap[n.ID] {
			newReads = append(newReads, models.NotificationRead{
				ID:             uuid.New(),
				NotificationID: n.ID,
				UserID:         userID,
				ReadAt:         now,
			})
		}
	}

	if len(newReads) > 0 {
		return s.db.Create(&newReads).Error
	}

	return nil
}
