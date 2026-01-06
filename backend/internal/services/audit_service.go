/*
Package services - IRIS Payroll System Business Logic Layer

==============================================================================
FILE: internal/services/audit_service.go
==============================================================================

DESCRIPTION:
    Business logic for audit logging, session tracking, and activity monitoring.
    Provides methods to log events, track sessions, and generate audit reports.

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new logging methods, add analytics functions
    ⚠️  CAUTION: Session management logic
    ❌  DO NOT modify: Core logging without considering security implications
    📝  Ensure all sensitive operations are logged

==============================================================================
*/
package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
	"backend/internal/repositories"
)

// AuditService handles audit logging business logic
type AuditService struct {
	auditRepo *repositories.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		auditRepo: repositories.NewAuditRepository(db),
	}
}

// LogLoginAttempt logs a login attempt
func (s *AuditService) LogLoginAttempt(c *gin.Context, email string, success bool, failureReason *string, userID *uuid.UUID) error {
	eventType := models.EventLoginAttempt
	if success {
		eventType = models.EventLoginSuccess
	} else if failureReason != nil {
		eventType = models.EventLoginFailure
	}

	log := &models.AuditLog{
		EventType:     eventType,
		UserID:        userID,
		Email:         email,
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		Success:       success,
		FailureReason: failureReason,
	}

	return s.auditRepo.CreateAuditLog(log)
}

// LogTokenRefresh logs a token refresh event
func (s *AuditService) LogTokenRefresh(c *gin.Context, userID uuid.UUID, email string, success bool) error {
	log := &models.AuditLog{
		EventType: models.EventTokenRefresh,
		UserID:    &userID,
		Email:     email,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   success,
	}

	return s.auditRepo.CreateAuditLog(log)
}

// LogLogout logs a logout event
func (s *AuditService) LogLogout(c *gin.Context, userID uuid.UUID, email string, sessionID string) error {
	// Create audit log
	log := &models.AuditLog{
		EventType: models.EventLogout,
		UserID:    &userID,
		Email:     email,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
		SessionID: &sessionID,
	}

	if err := s.auditRepo.CreateAuditLog(log); err != nil {
		return err
	}

	// End the session
	return s.auditRepo.EndSession(sessionID)
}

// LogPasswordChange logs a password change event
func (s *AuditService) LogPasswordChange(c *gin.Context, userID uuid.UUID, email string, success bool) error {
	log := &models.AuditLog{
		EventType: models.EventPasswordChange,
		UserID:    &userID,
		Email:     email,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   success,
	}

	return s.auditRepo.CreateAuditLog(log)
}

// CreateLoginSession creates a new login session
func (s *AuditService) CreateLoginSession(userID uuid.UUID, email string, ipAddress, userAgent string) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &models.LoginSession{
		UserID:       userID,
		Email:        email,
		LoginAt:      now,
		LastActivity: now,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		SessionID:    sessionID,
		IsActive:     true,
	}

	err := s.auditRepo.CreateLoginSession(session)
	return sessionID, err
}

// UpdateSessionActivity updates the last activity timestamp for a session
func (s *AuditService) UpdateSessionActivity(sessionID string) error {
	session, err := s.auditRepo.GetActiveSession(sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()
	return s.auditRepo.UpdateLoginSession(session)
}

// LogPageVisit logs a page visit
func (s *AuditService) LogPageVisit(userID uuid.UUID, sessionID, pageURL, pageTitle string) error {
	visit := &models.PageVisit{
		UserID:    userID,
		SessionID: sessionID,
		PageURL:   pageURL,
		PageTitle: &pageTitle,
		EnteredAt: time.Now(),
	}

	return s.auditRepo.CreatePageVisit(visit)
}

// GetAuditLogs retrieves audit logs with filters
func (s *AuditService) GetAuditLogs(filters map[string]interface{}, page, pageSize int) ([]models.AuditLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.auditRepo.GetAuditLogs(filters, pageSize, offset)
}

// GetLoginAttempts retrieves login attempts
func (s *AuditService) GetLoginAttempts(page, pageSize int) ([]models.AuditLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.auditRepo.GetLoginAttempts(pageSize, offset)
}

// GetLoginHistory retrieves login history
func (s *AuditService) GetLoginHistory(userID *uuid.UUID, page, pageSize int) ([]models.LoginSession, int64, error) {
	offset := (page - 1) * pageSize
	return s.auditRepo.GetLoginHistory(userID, pageSize, offset)
}

// GetActiveSessions retrieves active sessions
func (s *AuditService) GetActiveSessions(userID *uuid.UUID) ([]models.LoginSession, error) {
	return s.auditRepo.GetActiveSessions(userID)
}

// GetPageVisits retrieves page visits for a user
func (s *AuditService) GetPageVisits(userID uuid.UUID, sessionID *string, page, pageSize int) ([]models.PageVisit, int64, error) {
	offset := (page - 1) * pageSize
	return s.auditRepo.GetPageVisits(userID, sessionID, pageSize, offset)
}

// GetUserActivityStats retrieves activity statistics for a user
func (s *AuditService) GetUserActivityStats(userID uuid.UUID, days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	return s.auditRepo.GetUserActivityStats(userID, startDate, endDate)
}

// GetRecentLoginsByEmail retrieves recent login attempts for an email
func (s *AuditService) GetRecentLoginsByEmail(email string, hours int) ([]models.AuditLog, error) {
	return s.auditRepo.GetRecentLoginsByEmail(email, hours)
}

// GetLoginAttemptsStats retrieves statistics about login attempts
func (s *AuditService) GetLoginAttemptsStats(days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	filters := map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	}

	// Get all login attempts
	allAttempts, _, err := s.auditRepo.GetAuditLogs(filters, 10000, 0)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	totalAttempts := 0
	successfulAttempts := 0
	failedAttempts := 0
	uniqueUsers := make(map[string]bool)
	failureReasons := make(map[string]int)

	for _, attempt := range allAttempts {
		if attempt.EventType == models.EventLoginAttempt ||
		   attempt.EventType == models.EventLoginSuccess ||
		   attempt.EventType == models.EventLoginFailure {
			totalAttempts++

			if attempt.Success {
				successfulAttempts++
			} else {
				failedAttempts++
				if attempt.FailureReason != nil {
					failureReasons[*attempt.FailureReason]++
				}
			}

			uniqueUsers[attempt.Email] = true
		}
	}

	stats["total_attempts"] = totalAttempts
	stats["successful_attempts"] = successfulAttempts
	stats["failed_attempts"] = failedAttempts
	stats["unique_users"] = len(uniqueUsers)
	stats["failure_reasons"] = failureReasons
	stats["success_rate"] = float64(0)
	if totalAttempts > 0 {
		stats["success_rate"] = float64(successfulAttempts) / float64(totalAttempts) * 100
	}

	return stats, nil
}

// Helper function to convert metadata to JSON string
func (s *AuditService) metadataToJSON(metadata interface{}) (*string, error) {
	if metadata == nil {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	jsonStr := string(jsonBytes)
	return &jsonStr, nil
}
