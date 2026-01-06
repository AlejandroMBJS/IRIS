/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/escalation_service.go
==============================================================================

DESCRIPTION:
    Handles automatic escalation of pending absence requests after 24 hours.
    This service is called by a scheduled worker (cron job) to find requests
    that have been pending for more than 24 hours and automatically advance
    them to the next approval stage in the workflow.

USER PERSPECTIVE:
    - If a manager doesn't approve within 24 hours, request auto-escalates
    - Escalated requests notify the next level approver
    - Employees see escalation status in their request history

DEVELOPER GUIDELINES:
    ✅  OK to modify: Escalation time threshold, notification templates
    ⚠️  CAUTION: Workflow stage transitions (must match AbsenceRequestService)
    ❌  DO NOT modify: Database transaction logic without thorough testing
    📝  Always log escalations for audit trail

WORKFLOW STAGES:
    1. pending_supervisor (Direct Supervisor)
    2. pending_manager (Manager)
    3. pending_hr (HR - role depends on collar type)
    4. pending_gm (General Manager)
    5. pending_payroll (Payroll)
    6. approved (Final state)

ESCALATION LOGIC:
    - Query: Find all requests where status = 'pending_*' AND last_action_at < NOW() - 24 hours
    - Determine next stage based on current stage and employee collar type
    - Update absence_request: status, escalation_count, is_escalated, last_action_at
    - Create escalation_log entry
    - Send notification to next approver(s)

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EscalationService handles automatic escalation of pending absence requests
type EscalationService struct {
	db *gorm.DB
}

// NewEscalationService creates a new escalation service instance
func NewEscalationService(db *gorm.DB) *EscalationService {
	return &EscalationService{db: db}
}

// ProcessPendingEscalations finds and processes all requests pending for 24+ hours
// This is called by the scheduled worker (hourly cron job)
func (s *EscalationService) ProcessPendingEscalations() error {
	log.Println("Starting escalation processing...")

	// Find all absence requests pending for more than 24 hours
	cutoffTime := time.Now().Add(-24 * time.Hour)

	var pendingRequests []models.AbsenceRequest
	err := s.db.
		Where("status = ?", models.RequestStatusPending).
		Where("current_approval_stage != ?", models.ApprovalStageCompleted).
		Where("last_action_at < ?", cutoffTime).
		Find(&pendingRequests).Error

	if err != nil {
		log.Printf("Error querying pending requests: %v", err)
		return err
	}

	log.Printf("Found %d requests to escalate", len(pendingRequests))

	// Process each request
	for i := range pendingRequests {
		request := &pendingRequests[i]

		// Load employee data: AbsenceRequest.EmployeeID → User.ID → User.EmployeeID → Employee
		var user models.User
		if err := s.db.First(&user, "id = ?", request.EmployeeID).Error; err != nil {
			log.Printf("Error loading user for request %s: %v", request.ID, err)
			continue
		}

		if user.EmployeeID == nil {
			log.Printf("User %s has no associated employee record", user.ID)
			continue
		}

		var emp models.Employee
		if err := s.db.First(&emp, "id = ?", *user.EmployeeID).Error; err != nil {
			log.Printf("Error loading employee %s: %v", *user.EmployeeID, err)
			continue
		}

		request.Employee = &emp

		if err := s.escalateRequest(request); err != nil {
			log.Printf("Error escalating request %s: %v", request.ID, err)
			// Continue with other requests even if one fails
			continue
		}
	}

	log.Println("Escalation processing completed")
	return nil
}

// escalateRequest escalates a single absence request to the next stage
func (s *EscalationService) escalateRequest(request *models.AbsenceRequest) error {
	// Ensure employee data is available
	if request.Employee == nil {
		return fmt.Errorf("employee data not loaded for request %s", request.ID)
	}

	// Determine next stage based on current stage and employee collar type
	nextStage, err := s.determineNextStage(string(request.CurrentApprovalStage), request.Employee.CollarType)
	if err != nil {
		return fmt.Errorf("cannot determine next stage: %w", err)
	}

	// If already at final stage, skip
	if nextStage == "approved" {
		log.Printf("Request %s is already at final stage, skipping", request.ID)
		return nil
	}

	employeeName := request.Employee.FirstName + " " + request.Employee.LastName
	log.Printf("Escalating request %s: %s → %s (employee: %s, collar: %s)",
		request.ID, request.CurrentApprovalStage, nextStage, employeeName, request.Employee.CollarType)

	// Start database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update absence request
	oldStage := string(request.CurrentApprovalStage)
	now := time.Now()

	// Update the request object directly
	request.CurrentApprovalStage = models.ApprovalStage(nextStage)
	request.EscalationCount = request.EscalationCount + 1
	request.IsEscalated = true
	request.LastActionAt = now

	if err := tx.Save(&request).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update absence request: %w", err)
	}

	// Create escalation log entry
	escalationLog := models.EscalationLog{
		AbsenceRequestID: request.ID,
		FromStage:        oldStage,
		ToStage:          nextStage,
		EscalatedAt:      now,
		EscalationReason: "24 hours without approval",
	}

	if err := tx.Create(&escalationLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create escalation log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit escalation transaction: %w", err)
	}

	// Send notification to next approver(s)
	// TODO: Implement notification sending in Phase 5
	// For now, just log the action
	log.Printf("✓ Escalated request %s from %s to %s", request.ID, oldStage, nextStage)

	return nil
}

// determineNextStage determines the next approval stage based on current stage and collar type
// Workflow: Employee → Supervisor → Manager → HR (role-based) → GM → Payroll → Approved
func (s *EscalationService) determineNextStage(currentStage, collarType string) (string, error) {
	switch currentStage {
	case "SUPERVISOR", "pending_supervisor":
		// Supervisor → Manager
		return "MANAGER", nil

	case "MANAGER", "pending_manager":
		// Manager → HR (role depends on collar type)
		return "HR", nil

	case "HR", "HR_BLUE_GRAY", "pending_hr":
		// HR → GM
		return "GENERAL_MANAGER", nil

	case "GENERAL_MANAGER", "pending_gm":
		// GM → Payroll
		return "PAYROLL", nil

	case "PAYROLL", "pending_payroll":
		// Payroll → Completed (final stage)
		return "COMPLETED", nil

	case "COMPLETED", "approved":
		// Already completed, no next stage
		return "COMPLETED", nil

	case "DECLINED", "rejected":
		// Rejected requests cannot escalate
		return "", fmt.Errorf("rejected requests cannot be escalated")

	default:
		return "", fmt.Errorf("unknown status: %s", currentStage)
	}
}

// GetRequiredApproverRole returns the role needed to approve at a given stage
// Used for notification routing
func (s *EscalationService) GetRequiredApproverRole(stage, collarType string) (string, error) {
	switch stage {
	case "pending_supervisor":
		return "supervisor", nil

	case "pending_manager":
		return "manager", nil

	case "pending_hr":
		// Role-based HR routing based on collar type
		switch collarType {
		case "Blue", "Gray":
			return "hr_blue_gray", nil
		case "White":
			return "hr_white", nil
		default:
			return "hr_all", nil
		}

	case "pending_gm":
		return "gm", nil

	case "pending_payroll":
		return "payroll", nil

	default:
		return "", fmt.Errorf("unknown stage: %s", stage)
	}
}

// GetEscalationHistory returns the escalation history for a request
func (s *EscalationService) GetEscalationHistory(requestID uuid.UUID) ([]models.EscalationLog, error) {
	var logs []models.EscalationLog
	err := s.db.
		Where("absence_request_id = ?", requestID).
		Order("escalated_at ASC").
		Find(&logs).Error

	return logs, err
}
