/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/absence_request_service.go
==============================================================================

DESCRIPTION:
    Handles the business logic for absence request approval workflow.
    Manages request creation, multi-stage approval, and notifications.

APPROVAL FLOW:
    - Normal: SUPERVISOR → MANAGER → HR (blue collar) → COMPLETED
    - SUPANDGM supervisor: Auto-skip to HR stage
    - White collar: Skip HR after manager approval

==============================================================================
*/
package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"backend/internal/models"
	"backend/internal/models/enums"
)

// AbsenceRequestService handles absence request operations
type AbsenceRequestService struct {
	db *gorm.DB
}

// NewAbsenceRequestService creates a new AbsenceRequestService
func NewAbsenceRequestService(db *gorm.DB) *AbsenceRequestService {
	return &AbsenceRequestService{db: db}
}

// CreateAbsenceRequestInput holds the input data for creating an absence request
type CreateAbsenceRequestInput struct {
	EmployeeID     uuid.UUID
	RequestType    models.RequestType
	StartDate      time.Time
	EndDate        time.Time
	TotalDays      float64
	Reason         string
	HoursPerDay    *float64
	PaidDays       *float64
	UnpaidDays     *float64
	UnpaidComments string
	ShiftDetails   string
	NewShiftID     *uuid.UUID // For SHIFT_CHANGE requests - the target shift
}

// CreateAbsenceRequestResult holds the result of creating an absence request
type CreateAbsenceRequestResult struct {
	Request     *models.AbsenceRequest
	IncidenceID uuid.UUID
}

func (s *AbsenceRequestService) CreateAbsenceRequest(input CreateAbsenceRequestInput) (*CreateAbsenceRequestResult, error) {
	// Get the employee's user record
	var employee models.User
	if err := s.db.Preload("Employee").First(&employee, "id = ?", input.EmployeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}

	// Validate supervisor configuration
	if employee.SupervisorID == nil {
		return nil, errors.New("employee has no assigned supervisor - please contact HR to configure your supervisor")
	}

	// Validate general manager configuration
	if employee.GeneralManagerID == nil {
		return nil, errors.New("employee has no assigned general manager - please contact HR to configure reporting structure")
	}

	// Verify supervisor exists
	var supervisor models.User
	if err := s.db.First(&supervisor, "id = ?", employee.SupervisorID).Error; err != nil {
		return nil, errors.New("assigned supervisor not found in system - please contact HR")
	}

	// Verify general manager exists
	var generalManager models.User
	if err := s.db.First(&generalManager, "id = ?", employee.GeneralManagerID).Error; err != nil {
		return nil, errors.New("assigned general manager not found in system - please contact HR")
	}

	// Determine initial approval stage based on collar type
	var emp models.Employee
	isBlueOrGrayCollar := false
	if employee.EmployeeID != nil {
		s.db.First(&emp, "id = ?", *employee.EmployeeID)
		isBlueOrGrayCollar = emp.CollarType == "blue_collar" ||
			emp.CollarType == "gray_collar" ||
			emp.IsSindicalizado
	}

	// Initial stage: HR_BLUE_GRAY for blue/gray collar, SUPERVISOR for white collar
	initialStage := models.ApprovalStageSupervisor
	if isBlueOrGrayCollar {
		initialStage = models.ApprovalStageHRBlueGray
	}

	isSupervisorSUPANDGM := supervisor.Role == enums.RoleSupAndGM

	// Create the request
	now := time.Now()
	cutoff := s.calculatePayrollCutoff(&models.AbsenceRequest{
		EmployeeID: input.EmployeeID,
		StartDate:  input.StartDate,
	})

	request := &models.AbsenceRequest{
		EmployeeID:           input.EmployeeID,
		RequestType:          input.RequestType,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		TotalDays:            input.TotalDays,
		Reason:               input.Reason,
		Status:               models.RequestStatusPending,
		CurrentApprovalStage: initialStage,
		HoursPerDay:          input.HoursPerDay,
		PaidDays:             input.PaidDays,
		UnpaidDays:           input.UnpaidDays,
		UnpaidComments:       input.UnpaidComments,
		ShiftDetails:         input.ShiftDetails,
		NewShiftID:           input.NewShiftID,
		LastActionAt:         now,
		PayrollCutoffDate:    &cutoff,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the request
	if err := tx.Create(request).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create corresponding incidence immediately for evidence upload
	incidence, err := s.createInitialIncidence(tx, request, &employee)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create initial incidence: %w", err)
	}

	// Get request type display name
	requestTypeName := string(input.RequestType)

	// If supervisor is SUPANDGM, auto-approve supervisor and manager stages
	if isSupervisorSUPANDGM {
		// Add auto-approval for SUPERVISOR stage
		supervisorApproval := &models.ApprovalHistory{
			RequestID:     request.ID,
			ApproverID:    *employee.SupervisorID,
			ApprovalStage: models.ApprovalStageSupervisor,
			Action:        models.ApprovalActionApproved,
			Comments:      "Auto-approved (Supervisor is also General Manager)",
		}
		if err := tx.Create(supervisorApproval).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Add auto-approval for MANAGER stage
		managerApproval := &models.ApprovalHistory{
			RequestID:     request.ID,
			ApproverID:    *employee.SupervisorID,
			ApprovalStage: models.ApprovalStageManager,
			Action:        models.ApprovalActionApproved,
			Comments:      "Auto-approved (User is both Supervisor and General Manager)",
		}
		if err := tx.Create(managerApproval).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update request to HR stage
		request.CurrentApprovalStage = models.ApprovalStageHR
		if err := tx.Save(request).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Notify HR users (notification only, no inbox message)
		s.notifyHRUsers(tx, request, &employee)
	} else {
		// Normal flow - notify supervisor with both notification AND inbox message
		s.notifyApproverWithMessage(tx, *employee.SupervisorID, request.ID, employee.FullName, requestTypeName)

		// Also notify general manager if configured (with both notification AND inbox message)
		if employee.GeneralManagerID != nil && *employee.GeneralManagerID != *employee.SupervisorID {
			s.notifyApproverWithMessage(tx, *employee.GeneralManagerID, request.ID, employee.FullName, requestTypeName)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &CreateAbsenceRequestResult{
		Request:     request,
		IncidenceID: incidence.ID,
	}, nil
}

// GetMyRequests returns all active requests for an employee
func (s *AbsenceRequestService) GetMyRequests(employeeID uuid.UUID) ([]models.AbsenceRequest, error) {
	var requests []models.AbsenceRequest
	err := s.db.Preload("ApprovalHistory").
		Preload("ApprovalHistory.Approver").
		Where("employee_id = ? AND status != ?", employeeID, models.RequestStatusArchived).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetPendingRequestsForSupervisor returns requests pending supervisor approval
func (s *AbsenceRequestService) GetPendingRequestsForSupervisor(supervisorID uuid.UUID) ([]models.AbsenceRequest, error) {
	var requests []models.AbsenceRequest
	err := s.db.Preload("Employee").
		Preload("ApprovalHistory").
		Joins("JOIN users ON users.id = absence_requests.employee_id").
		Where("users.supervisor_id = ? AND absence_requests.status = ? AND absence_requests.current_approval_stage = ?",
			supervisorID, models.RequestStatusPending, models.ApprovalStageSupervisor).
		Order("absence_requests.created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetPendingRequestsForManager returns requests pending manager approval
func (s *AbsenceRequestService) GetPendingRequestsForManager() ([]models.AbsenceRequest, error) {
	var requests []models.AbsenceRequest
	err := s.db.Preload("Employee").
		Preload("ApprovalHistory").
		Where("status = ? AND current_approval_stage = ?",
			models.RequestStatusPending, models.ApprovalStageManager).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetPendingRequestsForHR returns requests pending HR approval
func (s *AbsenceRequestService) GetPendingRequestsForHR(hrUserID uuid.UUID) ([]models.AbsenceRequest, error) {
	// Get employee types assigned to this HR user
	var assignments []models.HRAssignment
	s.db.Where("hr_user_id = ?", hrUserID).Find(&assignments)

	var employeeTypes []string
	for _, a := range assignments {
		employeeTypes = append(employeeTypes, a.EmployeeType)
	}

	var requests []models.AbsenceRequest
	query := s.db.Preload("Employee").
		Preload("ApprovalHistory").
		Where("status = ? AND current_approval_stage = ?",
			models.RequestStatusPending, models.ApprovalStageHR)

	// Filter by employee types if HR has assignments
	if len(employeeTypes) > 0 {
		// Join with employees to filter by collar type / regimen
		query = query.Joins("JOIN users ON users.id = absence_requests.employee_id").
			Joins("LEFT JOIN employees ON employees.id = users.employee_id")
		// Note: This may need adjustment based on how regimen is stored
	}

	err := query.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// GetPendingRequestsForHRBlueGray returns blue/gray collar requests pending HR approval
func (s *AbsenceRequestService) GetPendingRequestsForHRBlueGray(hrUserID uuid.UUID) ([]models.AbsenceRequest, error) {
	var requests []models.AbsenceRequest
	err := s.db.Preload("Employee").
		Preload("ApprovalHistory").
		Joins("JOIN users ON users.id = absence_requests.employee_id").
		Joins("LEFT JOIN employees ON employees.id = users.employee_id").
		Where("absence_requests.status = ? AND absence_requests.current_approval_stage = ?",
			models.RequestStatusPending, models.ApprovalStageHRBlueGray).
		Where("employees.collar_type IN ? OR employees.is_sindicalizado = ?",
			[]string{"blue_collar", "gray_collar"}, true).
		Order("absence_requests.created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetPendingRequestsForGeneralManager returns all requests pending general manager approval
func (s *AbsenceRequestService) GetPendingRequestsForGeneralManager() ([]models.AbsenceRequest, error) {
	var requests []models.AbsenceRequest
	err := s.db.Preload("Employee").
		Preload("ApprovalHistory").
		Where("status = ? AND current_approval_stage = ?",
			models.RequestStatusPending, models.ApprovalStageGeneralManager).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetApprovedRequests returns all approved absence requests (for /incidences display)
func (s *AbsenceRequestService) GetApprovedRequests(filters map[string]interface{}) ([]models.AbsenceRequest, error) {
	query := s.db.Preload("Employee").
		Preload("ApprovalHistory").
		Preload("ApprovalHistory.Approver").
		Where("status = ?", models.RequestStatusApproved)

	if periodID, ok := filters["period_id"].(string); ok && periodID != "" {
		query = query.Joins("LEFT JOIN incidences ON incidences.absence_request_id = absence_requests.id").
			Where("incidences.payroll_period_id = ?", periodID)
	}

	if employeeID, ok := filters["employee_id"].(string); ok && employeeID != "" {
		query = query.Where("employee_id = ?", employeeID)
	}

	if collarType, ok := filters["collar_type"].(string); ok && collarType != "" {
		query = query.Joins("JOIN employees ON employees.id = absence_requests.employee_id").
			Where("employees.collar_type = ?", collarType)
	}

	var requests []models.AbsenceRequest
	err := query.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// ApproveRequest processes an approval or decline action
type ApproveRequestInput struct {
	RequestID uuid.UUID
	ApproverID uuid.UUID
	Stage     models.ApprovalStage
	Action    models.ApprovalAction
	Comments  string
}

func (s *AbsenceRequestService) ApproveRequest(input ApproveRequestInput) error {
	// Get the request
	var request models.AbsenceRequest
	if err := s.db.Preload("Employee").First(&request, "id = ?", input.RequestID).Error; err != nil {
		return errors.New("request not found")
	}

	// Verify request is pending and at correct stage
	if request.Status != models.RequestStatusPending {
		return errors.New("request is not pending")
	}
	if request.CurrentApprovalStage != input.Stage {
		return errors.New("request is not at the specified approval stage")
	}

	// Get approver for role verification
	var approver models.User
	if err := s.db.First(&approver, "id = ?", input.ApproverID).Error; err != nil {
		return errors.New("approver not found")
	}

	// Verify approver has permission for this stage
	if !s.canApproveStage(approver.Role, input.Stage) {
		return errors.New("unauthorized for this approval stage")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Add approval history
	history := &models.ApprovalHistory{
		RequestID:     input.RequestID,
		ApproverID:    input.ApproverID,
		ApprovalStage: input.Stage,
		Action:        input.Action,
		Comments:      input.Comments,
	}
	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Calculate payroll cutoff if not already set
	now := time.Now()
	if request.PayrollCutoffDate == nil {
		cutoff := s.calculatePayrollCutoff(&request)
		request.PayrollCutoffDate = &cutoff
	}

	// Update last_action_at for escalation tracking
	request.LastActionAt = now

	if input.Action == models.ApprovalActionDeclined {
		// Decline the request
		request.Status = models.RequestStatusDeclined
		request.CurrentApprovalStage = models.ApprovalStageCompleted

		// If rejected by HR or GM, exclude from payroll export
		isHRorGM := (input.Stage == models.ApprovalStageHR ||
			input.Stage == models.ApprovalStageHRBlueGray ||
			input.Stage == models.ApprovalStageGeneralManager)
		if isHRorGM {
			request.ExcludedFromPayroll = true
			// Update linked incidence
			if err := s.updateLinkedIncidenceFlags(tx, &request, false, true); err != nil {
				fmt.Printf("Warning: Failed to update incidence exclusion flag: %v\n", err)
			}
		}

		if err := tx.Save(&request).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Notify employee
		s.createNotification(tx, request.EmployeeID, request.ID,
			fmt.Sprintf("Tu solicitud de ausencia ha sido rechazada por %s", approver.FullName))
	} else {
		// Approved - determine next stage
		nextStage := s.getNextStage(input.Stage, &request)

		// Check if this is a late approval (after payroll cutoff)
		isLate := s.isLateApproval(&request, now)
		if isLate {
			request.LateApprovalFlag = true
			fmt.Printf("⚠️  Late approval detected for request %s (approved at %s, cutoff was %s)\n",
				request.ID, now.Format("2006-01-02 15:04:05"),
				request.PayrollCutoffDate.Format("2006-01-02 15:04:05"))
		}

		if nextStage == "" {
			// Final approval
			request.Status = models.RequestStatusApproved
			request.CurrentApprovalStage = models.ApprovalStageCompleted

			// Update linked incidence with late approval flag
			if err := s.updateLinkedIncidenceFlags(tx, &request, isLate, false); err != nil {
				fmt.Printf("Warning: Failed to update incidence late approval flag: %v\n", err)
			}

			if err := tx.Save(&request).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Create incidence in payroll system
			if err := s.createPayrollIncidence(tx, &request, input.ApproverID); err != nil {
				// Log error but don't fail the approval
				fmt.Printf("Warning: Failed to create payroll incidence for request %s: %v\n", request.ID, err)
			}

			// Create shift exception for SHIFT_CHANGE requests (makes the change reflect on schedule)
			if request.RequestType == models.RequestTypeShiftChange {
				if err := s.createShiftException(tx, &request, input.ApproverID); err != nil {
					fmt.Printf("Warning: Failed to create shift exception for request %s: %v\n", request.ID, err)
				}
			}

			// Notify employee
			s.createNotification(tx, request.EmployeeID, request.ID,
				"¡Tu solicitud de ausencia ha sido aprobada completamente!")

			// Notify payroll
			s.notifyPayrollUsers(tx, &request)
		} else {
			// Move to next stage
			request.CurrentApprovalStage = nextStage

			// Update linked incidence with late approval flag (but not excluded)
			if err := s.updateLinkedIncidenceFlags(tx, &request, isLate, false); err != nil {
				fmt.Printf("Warning: Failed to update incidence late approval flag: %v\n", err)
			}

			if err := tx.Save(&request).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Notify next approvers
			s.notifyNextApprovers(tx, &request, nextStage)
		}
	}

	return tx.Commit().Error
}

// ArchiveRequest archives (soft deletes) a request
func (s *AbsenceRequestService) ArchiveRequest(requestID, employeeID uuid.UUID) error {
	var request models.AbsenceRequest
	if err := s.db.First(&request, "id = ?", requestID).Error; err != nil {
		return errors.New("request not found")
	}

	if request.EmployeeID != employeeID {
		return errors.New("can only archive your own requests")
	}

	if request.Status == models.RequestStatusArchived {
		return errors.New("request is already archived")
	}

	request.Status = models.RequestStatusArchived
	return s.db.Save(&request).Error
}

// DeleteRequest permanently deletes a pending or declined request
func (s *AbsenceRequestService) DeleteRequest(requestID, employeeID uuid.UUID) error {
	var request models.AbsenceRequest
	if err := s.db.First(&request, "id = ?", requestID).Error; err != nil {
		return errors.New("request not found")
	}

	if request.EmployeeID != employeeID {
		return errors.New("can only delete your own requests")
	}

	if request.Status != models.RequestStatusPending && request.Status != models.RequestStatusDeclined {
		return errors.New("only pending or declined requests can be deleted")
	}

	// Delete related records
	tx := s.db.Begin()
	tx.Where("request_id = ?", requestID).Delete(&models.ApprovalHistory{})
	tx.Where("entity_id = ? AND entity_type = ?", requestID, "absence_request").Delete(&models.Notification{})
	tx.Delete(&request)
	return tx.Commit().Error
}

// GetOverlappingAbsences finds approved absences that overlap with the given date range
func (s *AbsenceRequestService) GetOverlappingAbsences(supervisorID uuid.UUID, startDate, endDate time.Time, excludeRequestID *uuid.UUID) ([]models.AbsenceRequest, error) {
	// Get employees under this supervisor
	var employeeIDs []uuid.UUID
	s.db.Model(&models.User{}).
		Where("supervisor_id = ?", supervisorID).
		Pluck("id", &employeeIDs)

	if len(employeeIDs) == 0 {
		return []models.AbsenceRequest{}, nil
	}

	query := s.db.Preload("Employee").
		Where("employee_id IN ? AND status = ?", employeeIDs, models.RequestStatusApproved).
		Where("(start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?) OR (start_date >= ? AND end_date <= ?)",
			endDate, startDate, endDate, startDate, startDate, endDate)

	if excludeRequestID != nil {
		query = query.Where("id != ?", *excludeRequestID)
	}

	var overlapping []models.AbsenceRequest
	err := query.Find(&overlapping).Error
	return overlapping, err
}

// GetPendingCounts returns counts of pending requests by stage
func (s *AbsenceRequestService) GetPendingCounts(userID uuid.UUID, role enums.UserRole) map[string]int64 {
	var supervisorCount, managerCount, hrCount int64

	// Supervisor count
	if role == enums.RoleSupervisor || role == enums.RoleSupAndGM || role == enums.RoleHRAndPR {
		s.db.Model(&models.AbsenceRequest{}).
			Joins("JOIN users ON users.id = absence_requests.employee_id").
			Where("users.supervisor_id = ? AND absence_requests.status = ? AND absence_requests.current_approval_stage = ?",
				userID, models.RequestStatusPending, models.ApprovalStageSupervisor).
			Count(&supervisorCount)
	}

	// Manager count
	if role == enums.RoleManager || role == enums.RoleSupAndGM {
		s.db.Model(&models.AbsenceRequest{}).
			Where("status = ? AND current_approval_stage = ?",
				models.RequestStatusPending, models.ApprovalStageManager).
			Count(&managerCount)
	}

	// HR count
	if role == enums.RoleHR || role == enums.RoleHRAndPR {
		s.db.Model(&models.AbsenceRequest{}).
			Where("status = ? AND current_approval_stage = ?",
				models.RequestStatusPending, models.ApprovalStageHR).
			Count(&hrCount)
	}

	return map[string]int64{
		"supervisor": supervisorCount,
		"manager":    managerCount,
		"hr":         hrCount,
	}
}

// Helper functions

func (s *AbsenceRequestService) canApproveStage(role enums.UserRole, stage models.ApprovalStage) bool {
	switch stage {
	case models.ApprovalStageSupervisor:
		return role == enums.RoleSupervisor || role == enums.RoleSupAndGM
	case models.ApprovalStageManager, models.ApprovalStageGeneralManager:
		return role == enums.RoleManager || role == enums.RoleSupAndGM
	case models.ApprovalStageHRBlueGray, models.ApprovalStageHR:
		return role == enums.RoleHR || role == enums.RoleHRAndPR || role == enums.RoleHRBlueGray
	case models.ApprovalStagePayroll:
		return role == enums.RolePayrollStaff || role == enums.RoleHRAndPR
	}
	return false
}

func (s *AbsenceRequestService) getNextStage(currentStage models.ApprovalStage, request *models.AbsenceRequest) models.ApprovalStage {
	// 6-Level Workflow (NEW):
	// Employee → SUPERVISOR → MANAGER → HR (role-based) → GENERAL_MANAGER → PAYROLL → COMPLETED

	// Get employee collar type for HR routing
	var employee models.Employee
	if request.Employee != nil && request.Employee.ID != uuid.Nil {
		s.db.First(&employee, "id = ?", request.Employee.ID)
	} else {
		// Load employee if not preloaded
		var user models.User
		if s.db.Preload("Employee").First(&user, "id = ?", request.EmployeeID).Error == nil {
			if user.EmployeeID != nil {
				s.db.First(&employee, "id = ?", *user.EmployeeID)
			}
		}
	}

	isBlueOrGrayCollar := employee.CollarType == "blue_collar" ||
		employee.CollarType == "gray_collar" ||
		employee.IsSindicalizado

	switch currentStage {
	case models.ApprovalStageSupervisor:
		// SUPERVISOR → MANAGER
		return models.ApprovalStageManager

	case models.ApprovalStageManager:
		// MANAGER → HR (role-based on collar type)
		if isBlueOrGrayCollar {
			return models.ApprovalStageHRBlueGray
		}
		return models.ApprovalStageHR

	case models.ApprovalStageHRBlueGray, models.ApprovalStageHR:
		// HR → GENERAL_MANAGER
		return models.ApprovalStageGeneralManager

	case models.ApprovalStageGeneralManager:
		// GENERAL_MANAGER → PAYROLL
		return models.ApprovalStagePayroll

	case models.ApprovalStagePayroll:
		// PAYROLL → COMPLETED (final stage)
		return "" // Empty = completed

	default:
		return ""
	}
}

func (s *AbsenceRequestService) createNotification(tx *gorm.DB, targetUserID, entityID uuid.UUID, message string) {
	// Get the target user to find their company
	var targetUser models.User
	if err := tx.First(&targetUser, "id = ?", targetUserID).Error; err != nil {
		return // Skip notification if user not found
	}

	notification := &models.Notification{
		CompanyID:    targetUser.CompanyID,
		ActorUserID:  targetUserID, // In approval workflow, actor is the system/current user
		TargetUserID: &targetUserID,
		Type:         "incidence_created", // Reuse existing notification type
		Title:        "Solicitud de Ausencia",
		Message:      message,
		ResourceType: "absence_request",
		ResourceID:   &entityID,
	}
	tx.Create(notification)
}

// createSystemMessage creates an inbox message from the system to notify approvers
func (s *AbsenceRequestService) createSystemMessage(tx *gorm.DB, recipientID uuid.UUID, subject, body string) {
	// Get the recipient user to find their company
	var recipient models.User
	if err := tx.First(&recipient, "id = ?", recipientID).Error; err != nil {
		return // Skip message if user not found
	}

	// Find or create system user for sending messages
	var systemUser models.User
	if err := tx.Where("email = ?", "sistema@iris.com").First(&systemUser).Error; err != nil {
		// Use the recipient's own ID as sender for system messages (self-notification)
		systemUser.ID = recipientID
	}

	message := &models.Message{
		ID:          uuid.New(),
		CompanyID:   recipient.CompanyID,
		SenderID:    systemUser.ID,
		RecipientID: recipientID,
		Subject:     subject,
		Body:        body,
		Type:        models.MessageTypeSystem,
		Status:      models.MessageStatusUnread,
	}
	tx.Create(message)
}

// notifyApproverWithMessage creates both a notification (bell) and inbox message
func (s *AbsenceRequestService) notifyApproverWithMessage(tx *gorm.DB, approverID, requestID uuid.UUID, employeeName, requestType string) {
	subject := fmt.Sprintf("Nueva solicitud de %s pendiente de aprobación", requestType)
	body := fmt.Sprintf("El empleado %s ha enviado una solicitud de %s que requiere tu aprobación.\n\nPor favor revisa la solicitud en el portal de aprobaciones.", employeeName, requestType)

	// Create bell notification
	s.createNotification(tx, approverID, requestID, subject)

	// Create inbox message
	s.createSystemMessage(tx, approverID, subject, body)
}

func (s *AbsenceRequestService) notifyHRUsers(tx *gorm.DB, request *models.AbsenceRequest, employee *models.User) {
	// HR users only get bell notification, NO inbox message
	requestTypeName := string(request.RequestType)
	var hrUsers []models.User
	tx.Where("role IN ?", []enums.UserRole{enums.RoleHR, enums.RoleHRAndPR}).Find(&hrUsers)
	for _, hr := range hrUsers {
		s.createNotification(tx, hr.ID, request.ID,
			fmt.Sprintf("Nueva solicitud de %s de %s (auto-aprobada por SUPANDGM) requiere aprobación de RH", requestTypeName, employee.FullName))
	}
}

func (s *AbsenceRequestService) notifyPayrollUsers(tx *gorm.DB, request *models.AbsenceRequest) {
	var payrollUsers []models.User
	tx.Where("role IN ?", []enums.UserRole{enums.RolePayrollStaff, enums.RoleHRAndPR}).Find(&payrollUsers)
	for _, user := range payrollUsers {
		s.createNotification(tx, user.ID, request.ID,
			"Nueva solicitud de ausencia aprobada para procesamiento de nómina")
	}
}

func (s *AbsenceRequestService) notifyNextApprovers(tx *gorm.DB, request *models.AbsenceRequest, nextStage models.ApprovalStage) {
	// Get employee name for the message
	var employee models.User
	tx.First(&employee, "id = ?", request.EmployeeID)
	employeeName := employee.FullName
	requestTypeName := string(request.RequestType)

	var roles []enums.UserRole
	switch nextStage {
	case models.ApprovalStageManager:
		roles = []enums.UserRole{enums.RoleManager, enums.RoleSupAndGM}
	case models.ApprovalStageHR:
		roles = []enums.UserRole{enums.RoleHR, enums.RoleHRAndPR}
	}

	if len(roles) > 0 {
		var approvers []models.User
		tx.Where("role IN ?", roles).Find(&approvers)
		for _, approver := range approvers {
			if nextStage == models.ApprovalStageHR {
				// HR only gets notification (no inbox message)
				s.createNotification(tx, approver.ID, request.ID,
					fmt.Sprintf("Nueva solicitud de %s de %s requiere aprobación de RH", requestTypeName, employeeName))
			} else {
				// Manager gets both notification AND inbox message
				s.notifyApproverWithMessage(tx, approver.ID, request.ID, employeeName, requestTypeName)
			}
		}
	}
}

// createInitialIncidence creates an incidence record immediately when absence request is created
// This allows evidence files to be uploaded using the incidence ID
func (s *AbsenceRequestService) createInitialIncidence(tx *gorm.DB, request *models.AbsenceRequest, employee *models.User) (*models.Incidence, error) {
	// User must have an employee_id to create an incidence
	if employee.EmployeeID == nil {
		return nil, fmt.Errorf("user has no associated employee record")
	}

	// Find or create the corresponding incidence type
	var incidenceType models.IncidenceType
	typeName := s.getIncidenceTypeName(request.RequestType)

	err := tx.Where("name = ?", typeName).First(&incidenceType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create the incidence type if it doesn't exist
			incidenceType = models.IncidenceType{
				Name:              typeName,
				Category:          s.mapRequestTypeToCategory(request.RequestType),
				EffectType:        s.getEffectType(request.RequestType),
				IsCalculated:      true,
				CalculationMethod: "daily_rate",
				DefaultValue:      1,
				RequiresEvidence:  false,
			}
			if err := tx.Create(&incidenceType).Error; err != nil {
				return nil, fmt.Errorf("failed to create incidence type: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to find incidence type: %w", err)
		}
	}

	// Find the payroll period that contains the request start date
	var period models.PayrollPeriod
	err = tx.Where("start_date <= ? AND end_date >= ?", request.StartDate, request.StartDate).
		First(&period).Error
	if err != nil {
		// If no period found for start date, try to find any open period
		err = tx.Where("status IN ?", []string{"open", "active"}).
			Order("start_date DESC").
			First(&period).Error
		if err != nil {
			return nil, fmt.Errorf("no payroll period found for incidence: %w", err)
		}
	}

	// Get employee to calculate amount
	var emp models.Employee
	if err := tx.First(&emp, "id = ?", *employee.EmployeeID).Error; err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// Calculate the amount based on request type and days
	calculatedAmount := 0.0
	if incidenceType.IsCalculated {
		switch incidenceType.CalculationMethod {
		case "daily_rate":
			calculatedAmount = emp.DailySalary * request.TotalDays
		case "hourly_rate":
			hoursPerDay := 8.0
			if request.HoursPerDay != nil {
				hoursPerDay = *request.HoursPerDay
			}
			calculatedAmount = (emp.DailySalary / 8) * hoursPerDay * request.TotalDays
		default:
			calculatedAmount = request.TotalDays
		}
	}

	// Create the incidence with 'pending' status
	incidence := &models.Incidence{
		EmployeeID:       *employee.EmployeeID,
		PayrollPeriodID:  period.ID,
		IncidenceTypeID:  incidenceType.ID,
		StartDate:        request.StartDate,
		EndDate:          request.EndDate,
		Quantity:         request.TotalDays,
		CalculatedAmount: calculatedAmount,
		Comments:         fmt.Sprintf("Solicitud de ausencia: %s", request.Reason),
		Status:           "pending", // Will be updated when request is approved
		AbsenceRequestID: &request.ID, // Link back to the absence request
	}

	if err := tx.Create(incidence).Error; err != nil {
		return nil, fmt.Errorf("failed to create incidence: %w", err)
	}

	return incidence, nil
}

// createPayrollIncidence creates an incidence record in the payroll system when an absence request is approved
func (s *AbsenceRequestService) createPayrollIncidence(tx *gorm.DB, request *models.AbsenceRequest, approverID uuid.UUID) error {
	// Get the user to find their employee_id
	var user models.User
	if err := tx.First(&user, "id = ?", request.EmployeeID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// User must have an employee_id to create an incidence
	if user.EmployeeID == nil {
		return fmt.Errorf("user has no associated employee record")
	}

	// Map request type to incidence category
	category := s.mapRequestTypeToCategory(request.RequestType)

	// Find or create the corresponding incidence type
	var incidenceType models.IncidenceType
	typeName := s.getIncidenceTypeName(request.RequestType)

	err := tx.Where("name = ?", typeName).First(&incidenceType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create the incidence type if it doesn't exist
			incidenceType = models.IncidenceType{
				Name:              typeName,
				Category:          category,
				EffectType:        s.getEffectType(request.RequestType),
				IsCalculated:      true,
				CalculationMethod: "daily_rate",
				DefaultValue:      1,
				RequiresEvidence:  false,
				Description:       fmt.Sprintf("Auto-created from %s request", request.RequestType),
			}
			if err := tx.Create(&incidenceType).Error; err != nil {
				return fmt.Errorf("failed to create incidence type: %w", err)
			}
		} else {
			return fmt.Errorf("failed to find incidence type: %w", err)
		}
	}

	// Find the payroll period that contains the request start date
	var period models.PayrollPeriod
	err = tx.Where("start_date <= ? AND end_date >= ?", request.StartDate, request.StartDate).
		First(&period).Error
	if err != nil {
		// If no period found for start date, try to find any open period
		err = tx.Where("status IN ?", []string{"open", "active"}).
			Order("start_date DESC").
			First(&period).Error
		if err != nil {
			return fmt.Errorf("no payroll period found for incidence: %w", err)
		}
	}

	// Get employee to calculate amount
	var employee models.Employee
	if err := tx.First(&employee, "id = ?", *user.EmployeeID).Error; err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	// Calculate the amount based on request type and days
	calculatedAmount := 0.0
	if incidenceType.IsCalculated {
		switch incidenceType.CalculationMethod {
		case "daily_rate":
			calculatedAmount = employee.DailySalary * request.TotalDays
		case "hourly_rate":
			hoursPerDay := 8.0
			if request.HoursPerDay != nil {
				hoursPerDay = *request.HoursPerDay
			}
			calculatedAmount = (employee.DailySalary / 8) * hoursPerDay * request.TotalDays
		default:
			calculatedAmount = request.TotalDays
		}
	}

	// Create the incidence - already approved since the request was approved
	now := time.Now()
	incidence := &models.Incidence{
		EmployeeID:       *user.EmployeeID,
		PayrollPeriodID:  period.ID,
		IncidenceTypeID:  incidenceType.ID,
		StartDate:        request.StartDate,
		EndDate:          request.EndDate,
		Quantity:         request.TotalDays,
		CalculatedAmount: calculatedAmount,
		Comments:         fmt.Sprintf("Solicitud de ausencia aprobada: %s", request.Reason),
		Status:           "approved",
		ApprovedBy:       &approverID,
		ApprovedAt:       &now,
		AbsenceRequestID: &request.ID, // Link back to the original absence request
	}

	if err := tx.Create(incidence).Error; err != nil {
		return fmt.Errorf("failed to create incidence: %w", err)
	}

	fmt.Printf("Created payroll incidence %s for employee %s from absence request %s\n",
		incidence.ID, employee.EmployeeNumber, request.ID)

	return nil
}

// mapRequestTypeToCategory maps absence request types to incidence categories
func (s *AbsenceRequestService) mapRequestTypeToCategory(requestType models.RequestType) string {
	switch requestType {
	case models.RequestTypeVacation:
		return "vacation"
	case models.RequestTypeSickLeave:
		return "sick"
	case models.RequestTypePaidLeave, models.RequestTypeUnpaidLeave,
		 models.RequestTypePersonal, models.RequestTypeOther:
		return "absence"
	case models.RequestTypeLateEntry, models.RequestTypeEarlyExit:
		return "delay"
	case models.RequestTypeShiftChange, models.RequestTypeTimeForTime:
		return "other"
	default:
		return "absence"
	}
}

// getIncidenceTypeName returns the incidence type name for a request type
func (s *AbsenceRequestService) getIncidenceTypeName(requestType models.RequestType) string {
	switch requestType {
	case models.RequestTypeVacation:
		return "Vacaciones"
	case models.RequestTypeSickLeave:
		return "Incapacidad"
	case models.RequestTypePaidLeave:
		return "Permiso con Goce de Sueldo"
	case models.RequestTypeUnpaidLeave:
		return "Permiso sin Goce de Sueldo"
	case models.RequestTypeLateEntry:
		return "Pase de Entrada"
	case models.RequestTypeEarlyExit:
		return "Pase de Salida"
	case models.RequestTypeShiftChange:
		return "Cambio de Turno"
	case models.RequestTypeTimeForTime:
		return "Tiempo por Tiempo"
	case models.RequestTypePersonal:
		return "Permiso Personal"
	case models.RequestTypeOther:
		return "Otro Permiso"
	default:
		return "Permiso General"
	}
}

// getEffectType returns the effect type for a request type
func (s *AbsenceRequestService) getEffectType(requestType models.RequestType) string {
	switch requestType {
	case models.RequestTypeUnpaidLeave:
		return "negative" // Deducts from salary
	case models.RequestTypePaidLeave, models.RequestTypeVacation:
		return "neutral" // No effect on salary
	default:
		return "neutral"
	}
}

// createShiftException creates shift exception records when a SHIFT_CHANGE request is approved
// This makes the shift change actually reflect on the employee's schedule
func (s *AbsenceRequestService) createShiftException(tx *gorm.DB, request *models.AbsenceRequest, approverID uuid.UUID) error {
	// Only process SHIFT_CHANGE requests
	if request.RequestType != models.RequestTypeShiftChange {
		return nil
	}

	// Get the user to find their employee_id
	var user models.User
	if err := tx.First(&user, "id = ?", request.EmployeeID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// User must have an employee_id
	if user.EmployeeID == nil {
		return fmt.Errorf("user has no associated employee record")
	}

	// Determine the new shift ID
	var newShiftID uuid.UUID
	if request.NewShiftID != nil {
		// Use the specified shift from the request
		newShiftID = *request.NewShiftID
	} else {
		// Try to find a shift based on ShiftDetails (legacy compatibility)
		// Look for a shift by name or code in the details
		if request.ShiftDetails != "" {
			var shift models.Shift
			err := tx.Where("name LIKE ? OR code LIKE ?",
				"%"+request.ShiftDetails+"%",
				"%"+request.ShiftDetails+"%").
				First(&shift).Error
			if err == nil {
				newShiftID = shift.ID
			}
		}

		// If still no shift found, we can't create the exception
		if newShiftID == uuid.Nil {
			fmt.Printf("Warning: No shift specified for SHIFT_CHANGE request %s, skipping shift exception creation\n", request.ID)
			return nil
		}
	}

	// Create shift exception for each day in the range
	currentDate := request.StartDate
	for !currentDate.After(request.EndDate) {
		// Check if an exception already exists for this employee/date
		var existing models.ShiftException
		err := tx.Where("employee_id = ? AND date = ?", *user.EmployeeID, currentDate).
			First(&existing).Error

		if err == nil {
			// Update existing exception
			existing.ShiftID = newShiftID
			existing.CreatedByID = approverID
			if err := tx.Save(&existing).Error; err != nil {
				return fmt.Errorf("failed to update shift exception: %w", err)
			}
			fmt.Printf("Updated shift exception for employee %s on %s\n",
				user.EmployeeID, currentDate.Format("2006-01-02"))
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new exception
			exception := &models.ShiftException{
				EmployeeID:  *user.EmployeeID,
				Date:        currentDate,
				ShiftID:     newShiftID,
				CreatedByID: approverID,
			}
			if err := tx.Create(exception).Error; err != nil {
				return fmt.Errorf("failed to create shift exception: %w", err)
			}
			fmt.Printf("Created shift exception for employee %s on %s with shift %s\n",
				user.EmployeeID, currentDate.Format("2006-01-02"), newShiftID)
		} else {
			return fmt.Errorf("failed to check existing shift exception: %w", err)
		}

		currentDate = currentDate.AddDate(0, 0, 1) // Move to next day
	}

	return nil
}

// GenerateApprovedRequestsExcel generates an Excel file with approved absence requests
func (s *AbsenceRequestService) GenerateApprovedRequestsExcel(requests []models.AbsenceRequest) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Approved Requests"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Headers
	headers := []string{"Employee Name", "Employee Number", "Request Type", "Start Date", "End Date", "Total Days", "Reason", "Approved Date"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	// Data rows
	for row, request := range requests {
		rowNum := row + 2

		// Get employee name and number - EmployeeID references employees table directly
		employeeName := ""
		employeeNumber := ""
		if request.Employee != nil {
			employeeName = request.Employee.FirstName + " " + request.Employee.LastName
			employeeNumber = request.Employee.EmployeeNumber
		} else {
			// Fallback: load employee if not preloaded
			var employee models.Employee
			if err := s.db.First(&employee, "id = ?", request.EmployeeID).Error; err == nil {
				employeeName = employee.FirstName + " " + employee.LastName
				employeeNumber = employee.EmployeeNumber
			}
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), employeeName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), employeeNumber)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), request.RequestType)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), request.StartDate.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), request.EndDate.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), request.TotalDays)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), request.Reason)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), request.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

// calculatePayrollCutoff calculates the payroll cutoff date for an absence request
// Based on employee collar type:
// - Blue/Gray collar: Weekly cutoff (Friday 23:59:59 CST)
// - White collar: Biweekly cutoff (every 2nd Friday 23:59:59 CST)
func (s *AbsenceRequestService) calculatePayrollCutoff(request *models.AbsenceRequest) time.Time {
	// Get employee collar type
	var user models.User
	s.db.Preload("Employee").First(&user, "id = ?", request.EmployeeID)

	isBlueOrGrayCollar := false
	if user.EmployeeID != nil {
		var employee models.Employee
		s.db.First(&employee, "id = ?", *user.EmployeeID)
		isBlueOrGrayCollar = employee.CollarType == "blue_collar" ||
			employee.CollarType == "gray_collar" ||
			employee.IsSindicalizado
	}

	// Use CST timezone (America/Mexico_City)
	cst, _ := time.LoadLocation("America/Mexico_City")
	now := time.Now().In(cst)

	// Find the next Friday at 23:59:59 CST
	daysUntilFriday := (5 - int(now.Weekday()) + 7) % 7
	if daysUntilFriday == 0 && now.Hour() >= 23 {
		// If today is Friday after cutoff, next cutoff is next week
		daysUntilFriday = 7
	}

	nextFriday := now.AddDate(0, 0, daysUntilFriday)
	cutoff := time.Date(nextFriday.Year(), nextFriday.Month(), nextFriday.Day(),
		23, 59, 59, 0, cst)

	// For white collar (biweekly), add 7 days if this is an "off" week
	// Simplified logic: Use week number to determine biweekly cycle
	if !isBlueOrGrayCollar {
		_, week := cutoff.ISOWeek()
		if week%2 == 0 {
			cutoff = cutoff.AddDate(0, 0, 7)
		}
	}

	return cutoff
}

// isLateApproval checks if the approval timestamp is after the payroll cutoff
func (s *AbsenceRequestService) isLateApproval(request *models.AbsenceRequest, approvalTime time.Time) bool {
	// Calculate cutoff if not already set
	var cutoff time.Time
	if request.PayrollCutoffDate != nil {
		cutoff = *request.PayrollCutoffDate
	} else {
		cutoff = s.calculatePayrollCutoff(request)
	}

	return approvalTime.After(cutoff)
}

// updateLinkedIncidenceFlags updates the incidence linked to an absence request with export flags
func (s *AbsenceRequestService) updateLinkedIncidenceFlags(tx *gorm.DB, request *models.AbsenceRequest, isLate, isExcluded bool) error {
	// Find incidence linked to this request
	var incidence models.Incidence
	err := tx.Where("absence_request_id = ?", request.ID).First(&incidence).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No linked incidence yet (might be created later), skip
			return nil
		}
		return fmt.Errorf("failed to find linked incidence: %w", err)
	}

	// Update flags
	updates := map[string]interface{}{
		"late_approval_flag":    isLate,
		"excluded_from_payroll": isExcluded,
	}

	if err := tx.Model(&incidence).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update incidence flags: %w", err)
	}

	return nil
}
