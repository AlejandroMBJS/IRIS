package services

import (
	"backend/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate test tables
	err = db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Employee{},
		&models.AbsenceRequest{},
		&models.EscalationLog{},
		&models.ApprovalHistory{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestEmployee(db *gorm.DB, collarType string) (*models.Employee, *models.User) {
	company := &models.Company{
		Name: "Test Company",
	}
	db.Create(company)

	dateOfBirth := time.Now().AddDate(-30, 0, 0)
	hireDate := dateOfBirth.AddDate(25, 0, 0) // Hired 5 years ago

	employee := &models.Employee{
		CompanyID:        company.ID,
		EmployeeNumber:   "TEST001",
		FirstName:        "Test",
		LastName:         "Employee",
		DateOfBirth:      dateOfBirth,
		HireDate:         hireDate,
		Gender:           "male", // Required: male, female, other
		RFC:              "XAXX010101000", // Valid RFC format
		CURP:             "XAXX010101HDFXXX00", // Valid CURP format
		EmployeeType:     "permanent", // Required: permanent, temporary, contractor, intern
		EmploymentStatus: "active",
		CollarType:       collarType,
		DailySalary:      500.0,
		Regime:           "02",
	}
	empResult := db.Create(employee)
	if empResult.Error != nil {
		panic("Failed to create test employee: " + empResult.Error.Error())
	}

	// Reload employee to ensure we have the saved ID
	var savedEmployee models.Employee
	db.First(&savedEmployee, "employee_number = ?", "TEST001")

	user := &models.User{
		CompanyID:    company.ID,
		EmployeeID:   &savedEmployee.ID,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123", // Required field
		Role:         "employee",           // Required field
		FullName:     "Test Employee",
		IsActive:     true,
	}
	result := db.Create(user)
	if result.Error != nil {
		panic("Failed to create test user: " + result.Error.Error())
	}

	return &savedEmployee, user
}

func TestDetermineNextStage_SupervisorToManager(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	nextStage, err := service.determineNextStage("SUPERVISOR", "blue_collar")

	assert.NoError(t, err)
	assert.Equal(t, "MANAGER", nextStage)
}

func TestDetermineNextStage_ManagerToHR(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	nextStage, err := service.determineNextStage("MANAGER", "blue_collar")

	assert.NoError(t, err)
	assert.Equal(t, "HR", nextStage)
}

func TestDetermineNextStage_HRToGM(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	nextStage, err := service.determineNextStage("HR", "white_collar")

	assert.NoError(t, err)
	assert.Equal(t, "GENERAL_MANAGER", nextStage)
}

func TestDetermineNextStage_GMToPayroll(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	nextStage, err := service.determineNextStage("GENERAL_MANAGER", "white_collar")

	assert.NoError(t, err)
	assert.Equal(t, "PAYROLL", nextStage)
}

func TestDetermineNextStage_PayrollToApproved(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	nextStage, err := service.determineNextStage("PAYROLL", "white_collar")

	assert.NoError(t, err)
	assert.Equal(t, "COMPLETED", nextStage)
}

func TestDetermineNextStage_RejectedCannotEscalate(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	_, err := service.determineNextStage("rejected", "white_collar")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rejected requests cannot be escalated")
}

func TestGetRequiredApproverRole_Supervisor(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	role, err := service.GetRequiredApproverRole("pending_supervisor", "blue_collar")

	assert.NoError(t, err)
	assert.Equal(t, "supervisor", role)
}

func TestGetRequiredApproverRole_Manager(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	role, err := service.GetRequiredApproverRole("pending_manager", "white_collar")

	assert.NoError(t, err)
	assert.Equal(t, "manager", role)
}

func TestGetRequiredApproverRole_HRBlueGray(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	role, err := service.GetRequiredApproverRole("pending_hr", "Blue")

	assert.NoError(t, err)
	assert.Equal(t, "hr_blue_gray", role)
}

func TestGetRequiredApproverRole_HRWhite(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	role, err := service.GetRequiredApproverRole("pending_hr", "White")

	assert.NoError(t, err)
	assert.Equal(t, "hr_white", role)
}

func TestGetRequiredApproverRole_GM(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	role, err := service.GetRequiredApproverRole("pending_gm", "white_collar")

	assert.NoError(t, err)
	assert.Equal(t, "gm", role)
}

func TestGetRequiredApproverRole_Payroll(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	role, err := service.GetRequiredApproverRole("pending_payroll", "blue_collar")

	assert.NoError(t, err)
	assert.Equal(t, "payroll", role)
}

func TestProcessPendingEscalations_NoRequests(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	err := service.ProcessPendingEscalations()

	assert.NoError(t, err)
}

func TestProcessPendingEscalations_EscalatesOldRequest(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	_, user := createTestEmployee(db, "blue_collar")

	// Verify user exists
	var checkUser models.User
	checkErr := db.First(&checkUser, "id = ?", user.ID).Error
	assert.NoError(t, checkErr, "User should exist in database")

	// Create a request that's been pending for 25 hours
	oldTime := time.Now().Add(-25 * time.Hour)
	now := time.Now()
	request := &models.AbsenceRequest{
		EmployeeID:           user.ID,
		RequestType:          models.RequestTypeVacation,
		StartDate:            now,
		EndDate:              now,
		TotalDays:            1.0,
		Reason:               "Test vacation",
		Status:               models.RequestStatusPending,
		CurrentApprovalStage: models.ApprovalStageSupervisor,
		LastActionAt:         oldTime,
	}
	err := db.Create(request).Error
	assert.NoError(t, err)

	// Verify the request was created
	var createdRequest models.AbsenceRequest
	err = db.First(&createdRequest, "id = ?", request.ID).Error
	assert.NoError(t, err)

	// Process escalations
	err = service.ProcessPendingEscalations()
	assert.NoError(t, err)

	// Verify request was escalated
	var updatedRequest models.AbsenceRequest
	db.First(&updatedRequest, "id = ?", request.ID)

	assert.Equal(t, models.ApprovalStageManager, updatedRequest.CurrentApprovalStage)
	assert.Equal(t, 1, updatedRequest.EscalationCount)
	assert.True(t, updatedRequest.IsEscalated)
	assert.True(t, updatedRequest.LastActionAt.After(oldTime))

	// Verify escalation log was created
	var logs []models.EscalationLog
	db.Where("absence_request_id = ?", request.ID).Find(&logs)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, string(models.ApprovalStageSupervisor), logs[0].FromStage)
	assert.Equal(t, string(models.ApprovalStageManager), logs[0].ToStage)
}

func TestProcessPendingEscalations_IgnoresRecentRequest(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	_, user := createTestEmployee(db, "white_collar")

	// Create a request that's only been pending for 5 hours
	recentTime := time.Now().Add(-5 * time.Hour)
	request := &models.AbsenceRequest{
		EmployeeID:           user.ID,
		RequestType:          models.RequestTypeVacation,
		StartDate:            time.Now(),
		EndDate:              time.Now(),
		TotalDays:            1.0,
		Reason:               "Test vacation",
		Status:               models.RequestStatusPending,
		CurrentApprovalStage: models.ApprovalStageSupervisor,
		LastActionAt:         recentTime,
	}
	db.Create(request)

	// Process escalations
	err := service.ProcessPendingEscalations()
	assert.NoError(t, err)

	// Verify request was NOT escalated
	var updatedRequest models.AbsenceRequest
	db.First(&updatedRequest, "id = ?", request.ID)

	assert.Equal(t, models.ApprovalStageSupervisor, updatedRequest.CurrentApprovalStage)
	assert.Equal(t, 0, updatedRequest.EscalationCount)
	assert.False(t, updatedRequest.IsEscalated)
}

func TestGetEscalationHistory(t *testing.T) {
	db := setupTestDB(t)
	service := NewEscalationService(db)

	requestID := uuid.New()

	// Create multiple escalation logs
	log1 := &models.EscalationLog{
		AbsenceRequestID: requestID,
		FromStage:        "pending_supervisor",
		ToStage:          "pending_manager",
		EscalatedAt:      time.Now().Add(-2 * time.Hour),
	}
	db.Create(log1)

	log2 := &models.EscalationLog{
		AbsenceRequestID: requestID,
		FromStage:        "pending_manager",
		ToStage:          "pending_hr",
		EscalatedAt:      time.Now().Add(-1 * time.Hour),
	}
	db.Create(log2)

	// Get history
	logs, err := service.GetEscalationHistory(requestID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(logs))
	// Should be ordered by escalated_at ASC
	assert.Equal(t, "pending_supervisor", logs[0].FromStage)
	assert.Equal(t, "pending_manager", logs[1].FromStage)
}
