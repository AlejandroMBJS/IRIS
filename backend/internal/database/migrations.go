/*
Package database - IRIS Payroll Database Migrations

==============================================================================
FILE: internal/database/migrations.go
==============================================================================

DESCRIPTION:
    Handles automatic database schema migrations using GORM AutoMigrate.
    Creates and updates tables for all application models. Called at
    application startup to ensure schema is current.

USER PERSPECTIVE:
    - Automatically creates database tables on first run
    - Updates schema when models change
    - No manual SQL migration scripts needed

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new models to AutoMigrate list
    ⚠️  CAUTION: Removing models (may cause data loss)
    ❌  DO NOT modify: Model order if foreign key dependencies exist
    📝  Add new models at the end of the list

SYNTAX EXPLANATION:
    - Migrate(): Entry point called from main.go
    - AutoMigrate(): GORM function that creates/updates tables
    - &models.XXX{}: Pointer to model struct for schema inference

MODEL LIST (in migration order):
    - Company: Multi-tenant isolation
    - User: Authentication and authorization
    - Employee: Core employee data
    - CostCenter: Department/cost center organization
    - PayrollPeriod: Payroll calculation periods
    - PayrollCalculation: Calculated payroll results
    - PayrollDetail: Line items for each calculation
    - PrenominaMetric: Pre-payroll metrics
    - EmployerContribution: Employer contribution tracking
    - Incidence: HR incidences (absences, vacations, etc.)
    - IncidenceType/Evidence: Incidence categories and attachments
    - PayrollConcept: Configurable payroll concepts
    - SalaryHistory: Salary change tracking
    - Notification/NotificationRead: User notifications

==============================================================================
*/
package database

import (
	"gorm.io/gorm"

	"backend/internal/models"
)

// Migrate performs database migrations.
func Migrate(db *gorm.DB) error {
	// AutoMigrate all models
	return db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Employee{},
		&models.CostCenter{},
		&models.PayrollPeriod{},
		&models.PayrollCalculation{},
		&models.PayrollDetail{},
		&models.PrenominaMetric{},
		&models.EmployerContribution{},
		&models.IncidenceCategory{}, // Must be before IncidenceType (FK dependency)
		&models.Incidence{},
		&models.IncidenceType{},
		&models.IncidenceEvidence{},
		&models.IncidenceTipoMapping{}, // NEW: Maps incidence types to payroll export templates
		&models.PayrollConcept{},
		&models.SalaryHistory{},
		&models.Notification{},
		&models.NotificationRead{},
		// Absence Request Approval Workflow Models
		&models.AbsenceRequest{},
		&models.ApprovalHistory{},
		&models.EscalationLog{}, // NEW: Audit trail for 24-hour auto-escalations
		&models.RoleInheritance{}, // NEW: Role permission inheritance configuration
		&models.HRAssignment{},
		&models.Shift{},
		&models.EmployeeShiftBase{},
		&models.ShiftException{},
		&models.Announcement{},
		&models.ReadAnnouncement{},
		// Internal Messaging System
		&models.Message{},
		// Audit Logging System
		&models.AuditLog{},
		&models.LoginSession{},
		&models.PageVisit{},
		&models.Permission{}, // NEW: Permission matrix for role-based access control
		// Recruitment Module Models
		&models.JobPosting{},
		&models.Candidate{},
		&models.Application{},
		&models.Interview{},
		&models.Offer{},
		&models.EvaluationCriteria{},
		&models.InterviewScore{},
		// Onboarding Module Models
		&models.OnboardingTemplate{},
		&models.OnboardingTaskTemplate{},
		&models.OnboardingChecklist{},
		&models.OnboardingTask{},
		&models.OnboardingNote{},
		// Training/LMS Module Models
		&models.TrainingCategory{},
		&models.Course{},
		&models.CourseModule{},
		&models.ModuleContent{},
		&models.CourseEnrollment{},
		&models.ContentProgress{},
		&models.TrainingSession{},
		&models.SessionAttendee{},
		&models.Certificate{},
		&models.LearningPath{},
		&models.LearningPathCourse{},
		// Performance Management Module Models
		&models.ReviewCycle{},
		&models.ReviewTemplate{},
		&models.ReviewTemplateSection{},
		&models.ReviewQuestion{},
		&models.PerformanceReview{},
		&models.ReviewResponse{},
		&models.Goal{},
		&models.GoalUpdate{},
		&models.Competency{},
		&models.CompetencyAssessment{},
		&models.Feedback{},
		&models.DevelopmentPlan{},
		&models.DevelopmentAction{},
		&models.OneOnOne{},
		// Benefits Administration Module Models
		&models.BenefitPlan{},
		&models.BenefitOption{},
		&models.EnrollmentPeriod{},
		&models.BenefitEnrollment{},
		&models.Dependent{},
		&models.BenefitDependent{},
		&models.LifeEvent{},
		&models.BenefitClaim{},
		&models.Beneficiary{},
		// Time Tracking Module Models
		&models.Project{},
		&models.ProjectTask{},
		&models.ProjectMember{},
		&models.Timesheet{},
		&models.TimeEntry{},
		&models.ClockRecord{},
		&models.OvertimeRequest{},
		&models.TimeOffBalance{},
		&models.TimeOffAccrual{},
		&models.TimePolicy{},
		&models.Holiday{},
		// Expense Management Module Models
		&models.ExpenseCategory{},
		&models.ExpensePolicy{},
		&models.ExpenseReport{},
		&models.ExpenseItem{},
		&models.ExpenseApproval{},
		&models.CorporateCard{},
		&models.CardTransaction{},
		&models.AdvancePayment{},
		// Document Management Module Models
		&models.DocumentCategory{},
		&models.Document{},
		&models.DocumentTemplate{},
		&models.GeneratedDocument{},
		&models.SignatureRequest{},
		&models.Signer{},
		&models.DocumentAccessLog{},
		&models.DocumentRequirement{},
		&models.EmployeeDocument{},
		&models.SharedDocument{},
	)
}
