/*
Package services - Message Service for Internal Messaging System

==============================================================================
FILE: internal/services/message_service.go
==============================================================================

DESCRIPTION:
    Provides business logic for the internal messaging system including
    sending messages, retrieving inbox/outbox, marking as read, and
    handling announcement questions.

==============================================================================
*/
package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// MessageService handles message operations
type MessageService struct {
	db *gorm.DB
}

// NewMessageService creates a new message service
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// SendMessage creates a new message
func (s *MessageService) SendMessage(senderID, recipientID, companyID uuid.UUID, subject, body string, msgType models.MessageType, announcementID, parentID *uuid.UUID) (*models.Message, error) {
	// Validate recipient exists
	var recipient models.User
	if err := s.db.Where("id = ? AND company_id = ?", recipientID, companyID).First(&recipient).Error; err != nil {
		return nil, errors.New("recipient not found")
	}

	// If announcement question, validate announcement exists
	if msgType == models.MessageTypeAnnouncementQuestion && announcementID != nil {
		var announcement models.Announcement
		if err := s.db.Where("id = ?", announcementID).First(&announcement).Error; err != nil {
			return nil, errors.New("announcement not found")
		}
	}

	// If reply, validate parent message exists
	if parentID != nil {
		var parent models.Message
		if err := s.db.Where("id = ?", parentID).First(&parent).Error; err != nil {
			return nil, errors.New("parent message not found")
		}
	}

	message := &models.Message{
		CompanyID:      companyID,
		SenderID:       senderID,
		RecipientID:    recipientID,
		Subject:        subject,
		Body:           body,
		Type:           msgType,
		Status:         models.MessageStatusUnread,
		AnnouncementID: announcementID,
		ParentID:       parentID,
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, err
	}

	// Load relationships for response
	s.db.Preload("Sender").Preload("Recipient").First(message, message.ID)

	return message, nil
}

// GetInbox retrieves messages for a user (received messages)
func (s *MessageService) GetInbox(userID, companyID uuid.UUID, page, pageSize int, status string) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := s.db.Model(&models.Message{}).
		Where("recipient_id = ? AND company_id = ?", userID, companyID).
		Where("parent_id IS NULL") // Only top-level messages

	// Filter by status if provided
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.
		Preload("Sender").
		Preload("Announcement").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	// Count replies for each message
	for i := range messages {
		var replyCount int64
		s.db.Model(&models.Message{}).Where("parent_id = ?", messages[i].ID).Count(&replyCount)
		// Note: We can't directly set ReplyCount on Message struct, but handler can calculate
	}

	return messages, total, nil
}

// GetSentMessages retrieves messages sent by a user
func (s *MessageService) GetSentMessages(userID, companyID uuid.UUID, page, pageSize int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := s.db.Model(&models.Message{}).
		Where("sender_id = ? AND company_id = ?", userID, companyID).
		Where("parent_id IS NULL") // Only top-level messages

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.
		Preload("Recipient").
		Preload("Announcement").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetMessage retrieves a single message by ID
func (s *MessageService) GetMessage(messageID, userID, companyID uuid.UUID) (*models.Message, error) {
	var message models.Message

	if err := s.db.
		Preload("Sender").
		Preload("Recipient").
		Preload("Announcement").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Sender").Order("created_at ASC")
		}).
		Where("id = ? AND company_id = ?", messageID, companyID).
		Where("sender_id = ? OR recipient_id = ?", userID, userID). // Must be sender or recipient
		First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	return &message, nil
}

// GetThread retrieves a message thread (parent + all replies)
func (s *MessageService) GetThread(messageID, userID, companyID uuid.UUID) ([]models.Message, error) {
	// First get the root message (could be this message or its parent)
	var rootMessage models.Message
	if err := s.db.Where("id = ? AND company_id = ?", messageID, companyID).First(&rootMessage).Error; err != nil {
		return nil, errors.New("message not found")
	}

	// If this message has a parent, get the parent instead
	rootID := rootMessage.ID
	if rootMessage.ParentID != nil {
		rootID = *rootMessage.ParentID
	}

	// Get all messages in the thread
	var messages []models.Message
	if err := s.db.
		Preload("Sender").
		Where("(id = ? OR parent_id = ?) AND company_id = ?", rootID, rootID, companyID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkAsRead marks a message as read
func (s *MessageService) MarkAsRead(messageID, userID, companyID uuid.UUID) error {
	now := time.Now()
	result := s.db.Model(&models.Message{}).
		Where("id = ? AND recipient_id = ? AND company_id = ?", messageID, userID, companyID).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusRead,
			"read_at": now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("message not found or already read")
	}
	return nil
}

// MarkAsUnread marks a message as unread
func (s *MessageService) MarkAsUnread(messageID, userID, companyID uuid.UUID) error {
	result := s.db.Model(&models.Message{}).
		Where("id = ? AND recipient_id = ? AND company_id = ?", messageID, userID, companyID).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusUnread,
			"read_at": nil,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("message not found")
	}
	return nil
}

// ArchiveMessage archives a message (soft archive, not delete)
func (s *MessageService) ArchiveMessage(messageID, userID, companyID uuid.UUID) error {
	result := s.db.Model(&models.Message{}).
		Where("id = ? AND company_id = ?", messageID, companyID).
		Where("sender_id = ? OR recipient_id = ?", userID, userID).
		Update("status", models.MessageStatusArchived)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("message not found")
	}
	return nil
}

// DeleteMessage soft deletes a message
func (s *MessageService) DeleteMessage(messageID, userID, companyID uuid.UUID) error {
	result := s.db.
		Where("id = ? AND company_id = ?", messageID, companyID).
		Where("sender_id = ? OR recipient_id = ?", userID, userID).
		Delete(&models.Message{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("message not found")
	}
	return nil
}

// GetUnreadCount returns the count of unread messages for a user
func (s *MessageService) GetUnreadCount(userID, companyID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.Message{}).
		Where("recipient_id = ? AND company_id = ? AND status = ?", userID, companyID, models.MessageStatusUnread).
		Count(&count).Error
	return count, err
}

// GetRecipientSuggestions returns users that can be messaged
func (s *MessageService) GetRecipientSuggestions(userID, companyID uuid.UUID, search string) ([]models.User, error) {
	var users []models.User

	query := s.db.Model(&models.User{}).
		Where("company_id = ? AND id != ?", companyID, userID).
		Where("role IN ?", []string{"admin", "hr", "hr_and_pr", "hr_blue_gray", "hr_white", "supervisor", "manager", "sup_and_gm"})

	if search != "" {
		query = query.Where("full_name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetAnnouncementCreator returns the creator of an announcement for sending questions
func (s *MessageService) GetAnnouncementCreator(announcementID uuid.UUID) (*models.User, error) {
	var announcement models.Announcement
	if err := s.db.Preload("CreatedBy").Where("id = ?", announcementID).First(&announcement).Error; err != nil {
		return nil, errors.New("announcement not found")
	}
	return announcement.CreatedBy, nil
}
