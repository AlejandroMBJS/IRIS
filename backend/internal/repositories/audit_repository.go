/*
Package repositories - IRIS Payroll System Data Access Layer

==============================================================================
FILE: internal/repositories/audit_repository.go
==============================================================================

DESCRIPTION:
    Handles database operations for audit logs, login sessions, and page visits.
    Provides methods to create logs, query audit history, and track user sessions.

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new query methods, add filters
    ⚠️  CAUTION: Database queries performance
    ❌  DO NOT modify: Core CRUD operations without careful testing
    📝  Consider indexes when adding new query methods

==============================================================================
*/
package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// AuditRepository handles database operations for audit logs
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *AuditRepository) CreateAuditLog(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

// GetAuditLogs retrieves audit logs with filters
func (r *AuditRepository) GetAuditLogs(filters map[string]interface{}, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{}).Preload("User")

	// Apply filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if email, ok := filters["email"].(string); ok && email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if eventType, ok := filters["event_type"].(string); ok && eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if success, ok := filters["success"].(bool); ok {
		query = query.Where("success = ?", success)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

// GetLoginAttempts retrieves login attempts (both success and failure)
func (r *AuditRepository) GetLoginAttempts(limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{}).
		Where("event_type IN ?", []models.AuditLogEventType{
			models.EventLoginAttempt,
			models.EventLoginSuccess,
			models.EventLoginFailure,
		}).
		Preload("User")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

// GetLoginHistory retrieves successful login history with session info
func (r *AuditRepository) GetLoginHistory(userID *uuid.UUID, limit, offset int) ([]models.LoginSession, int64, error) {
	var sessions []models.LoginSession
	var total int64

	query := r.db.Model(&models.LoginSession{}).Preload("User")

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("login_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, total, err
}

// CreateLoginSession creates a new login session
func (r *AuditRepository) CreateLoginSession(session *models.LoginSession) error {
	return r.db.Create(session).Error
}

// UpdateLoginSession updates a login session
func (r *AuditRepository) UpdateLoginSession(session *models.LoginSession) error {
	return r.db.Save(session).Error
}

// GetActiveSession retrieves an active session by session ID
func (r *AuditRepository) GetActiveSession(sessionID string) (*models.LoginSession, error) {
	var session models.LoginSession
	err := r.db.Where("session_id = ? AND is_active = ?", sessionID, true).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetActiveSessions retrieves all active sessions
func (r *AuditRepository) GetActiveSessions(userID *uuid.UUID) ([]models.LoginSession, error) {
	var sessions []models.LoginSession
	query := r.db.Where("is_active = ?", true).Preload("User")

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Order("last_activity DESC").Find(&sessions).Error
	return sessions, err
}

// EndSession marks a session as inactive and sets logout time
func (r *AuditRepository) EndSession(sessionID string) error {
	now := time.Now()
	return r.db.Model(&models.LoginSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"is_active": false,
			"logout_at": now,
		}).Error
}

// CreatePageVisit creates a new page visit entry
func (r *AuditRepository) CreatePageVisit(visit *models.PageVisit) error {
	return r.db.Create(visit).Error
}

// UpdatePageVisit updates a page visit entry
func (r *AuditRepository) UpdatePageVisit(visit *models.PageVisit) error {
	return r.db.Save(visit).Error
}

// GetPageVisits retrieves page visits for a user
func (r *AuditRepository) GetPageVisits(userID uuid.UUID, sessionID *string, limit, offset int) ([]models.PageVisit, int64, error) {
	var visits []models.PageVisit
	var total int64

	query := r.db.Model(&models.PageVisit{}).Where("user_id = ?", userID).Preload("User")

	if sessionID != nil {
		query = query.Where("session_id = ?", *sessionID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("entered_at DESC").Limit(limit).Offset(offset).Find(&visits).Error
	return visits, total, err
}

// GetRecentLoginsByEmail retrieves recent login attempts for an email
func (r *AuditRepository) GetRecentLoginsByEmail(email string, hours int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	since := time.Now().Add(time.Duration(-hours) * time.Hour)

	err := r.db.Model(&models.AuditLog{}).
		Where("email = ? AND event_type IN ? AND created_at >= ?",
			email,
			[]models.AuditLogEventType{
				models.EventLoginAttempt,
				models.EventLoginSuccess,
				models.EventLoginFailure,
			},
			since,
		).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, err
}

// GetUserActivityStats retrieves activity statistics for a user
func (r *AuditRepository) GetUserActivityStats(userID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count login attempts
	var loginCount int64
	r.db.Model(&models.AuditLog{}).
		Where("user_id = ? AND event_type IN ? AND created_at BETWEEN ? AND ?",
			userID,
			[]models.AuditLogEventType{models.EventLoginSuccess},
			startDate,
			endDate,
		).
		Count(&loginCount)
	stats["total_logins"] = loginCount

	// Count page visits
	var visitCount int64
	r.db.Model(&models.PageVisit{}).
		Where("user_id = ? AND entered_at BETWEEN ? AND ?", userID, startDate, endDate).
		Count(&visitCount)
	stats["total_page_visits"] = visitCount

	// Average session duration
	var avgDuration float64
	r.db.Model(&models.LoginSession{}).
		Where("user_id = ? AND login_at BETWEEN ? AND ? AND logout_at IS NOT NULL",
			userID, startDate, endDate).
		Select("AVG(EXTRACT(EPOCH FROM (logout_at - login_at)))").
		Scan(&avgDuration)
	stats["avg_session_duration_seconds"] = avgDuration

	return stats, nil
}
