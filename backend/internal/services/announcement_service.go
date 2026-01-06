/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/announcement_service.go
==============================================================================

DESCRIPTION:
    Business logic for company announcements and employee communication.
    Handles announcement creation, retrieval, marking as read, and scope-based
    filtering for different user roles.

USER PERSPECTIVE:
    - HR/Managers can create company-wide or team announcements
    - Employees see relevant announcements based on their role/department
    - Track which announcements have been read
    - Support for image attachments and expiration dates

==============================================================================
*/
package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// AnnouncementService handles announcement operations
type AnnouncementService struct {
	db *gorm.DB
}

// NewAnnouncementService creates a new announcement service
func NewAnnouncementService(db *gorm.DB) *AnnouncementService {
	return &AnnouncementService{db: db}
}

// CreateAnnouncement creates a new announcement
func (s *AnnouncementService) CreateAnnouncement(title, message, scope string, createdByID uuid.UUID, imageData []byte, expiresAt *time.Time) (*models.Announcement, error) {
	announcement := &models.Announcement{
		CreatedByID: createdByID,
		Title:       title,
		Message:     message,
		ImageData:   imageData,
		Scope:       scope,
		IsActive:    true,
		ExpiresAt:   expiresAt,
	}

	if err := s.db.Create(announcement).Error; err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}

	// Preload creator information
	if err := s.db.Preload("CreatedBy").First(announcement, "id = ?", announcement.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load announcement: %w", err)
	}

	return announcement, nil
}

// GetActiveAnnouncements retrieves active, non-expired announcements filtered by scope
// - ALL scope: visible to everyone
// - TEAM scope: visible only to creator's subordinates (users whose supervisor_id = creator)
func (s *AnnouncementService) GetActiveAnnouncements(userID uuid.UUID) ([]models.Announcement, error) {
	var announcements []models.Announcement
	now := time.Now()

	// First get the user's supervisor_id to know whose TEAM announcements they can see
	var currentUser models.User
	if err := s.db.First(&currentUser, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Build query with scope filtering:
	// 1. Show ALL scope announcements to everyone
	// 2. Show TEAM scope announcements only if:
	//    - User is the creator (supervisor viewing their own), OR
	//    - User's supervisor_id matches the creator (subordinate seeing their supervisor's announcement)
	query := s.db.Preload("CreatedBy").
		Where("is_active = ?", true).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Where(`(
			scope = 'ALL' OR
			scope = 'COMPANY' OR
			(scope = 'TEAM' AND (created_by_id = ? OR created_by_id = ?))
		)`, userID, currentUser.SupervisorID).
		Order("created_at DESC")

	if err := query.Find(&announcements).Error; err != nil {
		return nil, fmt.Errorf("failed to get announcements: %w", err)
	}

	// Mark which ones have been read by this user
	var readRecords []models.ReadAnnouncement
	s.db.Where("user_id = ?", userID).Find(&readRecords)

	readMap := make(map[uuid.UUID]bool)
	for _, record := range readRecords {
		readMap[record.AnnouncementID] = true
	}

	return announcements, nil
}

// MarkAnnouncementAsRead marks an announcement as read by a user
func (s *AnnouncementService) MarkAnnouncementAsRead(announcementID, userID uuid.UUID) error {
	// Check if already marked as read
	var existing models.ReadAnnouncement
	err := s.db.Where("user_id = ? AND announcement_id = ?", userID, announcementID).First(&existing).Error

	if err == nil {
		// Already read
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check read status: %w", err)
	}

	// Mark as read
	readRecord := models.ReadAnnouncement{
		UserID:         userID,
		AnnouncementID: announcementID,
		ReadAt:         time.Now(),
	}

	if err := s.db.Create(&readRecord).Error; err != nil {
		return fmt.Errorf("failed to mark announcement as read: %w", err)
	}

	return nil
}

// GetUnreadCount returns the count of unread announcements for a user
// Uses the same scope filtering as GetActiveAnnouncements
func (s *AnnouncementService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	now := time.Now()

	// Get user's supervisor to filter TEAM scoped announcements
	var currentUser models.User
	if err := s.db.First(&currentUser, "id = ?", userID).Error; err != nil {
		return 0, fmt.Errorf("failed to get current user: %w", err)
	}

	// Subquery to get read announcement IDs
	subQuery := s.db.Model(&models.ReadAnnouncement{}).
		Select("announcement_id").
		Where("user_id = ?", userID)

	// Count active announcements that haven't been read with scope filtering
	err := s.db.Model(&models.Announcement{}).
		Where("is_active = ?", true).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Where("id NOT IN (?)", subQuery).
		Where(`(
			scope = 'ALL' OR
			scope = 'COMPANY' OR
			(scope = 'TEAM' AND (created_by_id = ? OR created_by_id = ?))
		)`, userID, currentUser.SupervisorID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// DeleteAnnouncement soft deletes an announcement
func (s *AnnouncementService) DeleteAnnouncement(announcementID, userID uuid.UUID) error {
	// Check if announcement exists and user is the creator or has admin rights
	var announcement models.Announcement
	if err := s.db.First(&announcement, "id = ?", announcementID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("announcement not found")
		}
		return fmt.Errorf("failed to get announcement: %w", err)
	}

	// Verify user is creator (additional role check should be done in handler)
	if announcement.CreatedByID != userID {
		// Get user to check role
		var user models.User
		if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
			return fmt.Errorf("unauthorized: not the creator")
		}

		// Allow admin and hr to delete any announcement
		if user.Role != "admin" && user.Role != "hr" {
			return fmt.Errorf("unauthorized: not the creator")
		}
	}

	// Soft delete by setting is_active to false
	if err := s.db.Model(&announcement).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}

	return nil
}

// GetAnnouncementByID retrieves a single announcement by ID
func (s *AnnouncementService) GetAnnouncementByID(announcementID uuid.UUID) (*models.Announcement, error) {
	var announcement models.Announcement

	if err := s.db.Preload("CreatedBy").First(&announcement, "id = ?", announcementID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("announcement not found")
		}
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}

	return &announcement, nil
}
