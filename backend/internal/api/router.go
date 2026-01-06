/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/router.go
==============================================================================

DESCRIPTION:
    Central routing configuration for the IRIS Payroll API. Sets up all
    endpoints, middleware chains, and service dependencies.

USER PERSPECTIVE:
    - This file defines all available API endpoints
    - Determines which routes require authentication
    - Sets up role-based access control for admin/hr features

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new route groups, new handlers
    ⚠️  CAUTION: Changing existing route paths (breaks frontend)
    ❌  DO NOT modify: Authentication middleware order
    📝  Follow RESTful conventions for new endpoints

SYNTAX EXPLANATION:
    - Router struct: Holds dependencies for handler creation
    - Setup(): Called from main.go to configure all routes
    - gin.RouterGroup: Groups routes with shared prefix/middleware
    - protected.Use(): Applies middleware to all routes in group

ROUTE STRUCTURE:
    /api/v1
    ├── /health (no auth)
    ├── /auth/* (mixed auth)
    ├── /employees/* (auth required)
    ├── /payroll/* (auth required)
    ├── /incidences/* (auth required)
    ├── /users/* (admin only)
    └── /reports/* (auth required)

==============================================================================
*/
package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/services"
)

// Router sets up all API routes
type Router struct {
    db         *gorm.DB
    appConfig  *config.AppConfig
    authService *services.AuthService
}

// NewRouter creates a new router
func NewRouter(db *gorm.DB, appConfig *config.AppConfig) *Router {
    authService := services.NewAuthService(db, appConfig)
    return &Router{
        db:         db,
        appConfig:  appConfig,
        authService: authService,
    }
}

// Setup configures all routes
func (r *Router) Setup(routerGroup *gin.RouterGroup) { // Changed return type to avoid confusion in main.go
    // Set Gin mode
    if r.appConfig.IsProduction() {
        gin.SetMode(gin.ReleaseMode)
    }

    // Apply security headers to all routes
    securityMiddleware := middleware.NewSecurityMiddleware(r.appConfig)
    routerGroup.Use(securityMiddleware.Headers())

    // Apply CSRF protection to all routes
    csrfMiddleware := middleware.NewCSRFMiddleware(r.appConfig)
    routerGroup.Use(csrfMiddleware.Protect())

    // Health check endpoint
    routerGroup.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":  "ok",
            "service": "iris-payroll-backend",
        })
    })

    // API v1 routes
    api := routerGroup.Group("") // Group under the provided routerGroup
    {
        // Initialize audit service
        auditService := services.NewAuditService(r.db)

        // Authentication routes (no auth required)
        authHandler := NewAuthHandler(r.authService, auditService, r.appConfig)
        authHandler.RegisterRoutes(api)
        
        // Protected routes
        protected := api.Group("")
        protected.Use(middleware.NewAuthMiddleware(r.authService).RequireAuth())
        {
            // Payroll Period Routes
            payrollPeriodService := services.NewPayrollPeriodService(r.db)
            payrollPeriodHandler := NewPayrollPeriodHandler(payrollPeriodService)
            payrollPeriodHandler.RegisterRoutes(protected)

            // Catalog Routes
            catalogService := services.NewCatalogService(r.db)
            catalogHandler := NewCatalogHandler(catalogService)
            catalogHandler.RegisterRoutes(protected)

            // Report Routes
            reportService := services.NewReportService(r.db)
            reportHandler := NewReportHandler(reportService)
            reportHandler.RegisterRoutes(protected)

            // Employee Routes
            employeeService := services.NewEmployeeService(r.db)
            employeeHandler := NewEmployeeHandler(employeeService)
            employeeHandler.RegisterRoutes(protected)

            // Payroll Routes
            payrollService := services.NewPayrollService(r.db, r.appConfig)
            payrollHandler := NewPayrollHandler(payrollService)
            payrollHandler.RegisterRoutes(protected)

            // Payroll Export Routes (Dual Excel export for payroll processing)
            excelExportService := services.NewExcelExportService(r.db)
            payrollExportHandler := NewPayrollExportHandler(excelExportService)
            payrollExportHandler.RegisterRoutes(protected)

            // Incidence Config Routes (Tipo mappings for payroll export)
            incidenceConfigService := services.NewIncidenceConfigService(r.db)
            incidenceConfigHandler := NewIncidenceConfigHandler(incidenceConfigService)
            incidenceConfigHandler.RegisterRoutes(protected)

            // Role Inheritance Routes (Admin only - Role permission inheritance configuration)
            roleInheritanceService := services.NewRoleInheritanceService(r.db)
            roleInheritanceHandler := NewRoleInheritanceHandler(roleInheritanceService)
            roleInheritanceHandler.RegisterRoutes(protected, middleware.NewAuthMiddleware(r.authService))

            // Incidence Category Routes (for organizing incidence types)
            categoryService := services.NewIncidenceCategoryService(r.db)
            categoryHandler := NewIncidenceCategoryHandler(categoryService)
            categoryHandler.RegisterRoutes(protected)

            // Incidence Routes (for HR tracking absences, vacations, etc.)
            incidenceService := services.NewIncidenceService(r.db)
            incidenceHandler := NewIncidenceHandler(incidenceService)
            incidenceHandler.RegisterRoutes(protected)

            // Upload/Evidence Routes
            uploadService := services.NewUploadService(r.db)
            uploadHandler := NewUploadHandler(uploadService)
            uploadHandler.RegisterRoutes(protected)

            // Notification Routes
            notificationService := services.NewNotificationService(r.db)
            notificationHandler := NewNotificationHandler(notificationService)
            notificationHandler.RegisterRoutes(protected)

            // Absence Request Approval Workflow Routes
            absenceRequestService := services.NewAbsenceRequestService(r.db)
            absenceRequestHandler := NewAbsenceRequestHandler(absenceRequestService)
            absenceRequestHandler.RegisterRoutes(protected)

            // Announcement Routes
            announcementService := services.NewAnnouncementService(r.db)
            announcementHandler := NewAnnouncementHandler(announcementService)
            announcementHandler.RegisterRoutes(protected)

            // Calendar Routes (HR Calendar feature)
            calendarService := services.NewCalendarService(r.db)
            calendarHandler := NewCalendarHandler(calendarService)
            calendarHandler.RegisterRoutes(protected)

            // Shift Routes (for schedule management)
            shiftService := services.NewShiftService(r.db)
            shiftHandler := NewShiftHandler(shiftService)
            shiftHandler.RegisterRoutes(protected, middleware.NewAuthMiddleware(r.authService))

            // Message Routes (Inbox/Messaging system)
            messageService := services.NewMessageService(r.db)
            messageHandler := NewMessageHandler(messageService)
            messageHandler.RegisterRoutes(protected)

            // Audit Log Routes (Admin can see all, users can see their own)
            auditHandler := NewAuditHandler(auditService, r.authService)
            auditHandler.RegisterRoutes(protected)

            // Permission Matrix Routes (Admin only - for role-based access control)
            permissionService := services.NewPermissionService(r.db)
            permissionHandler := NewPermissionHandler(permissionService)
            permissionHandler.RegisterRoutes(protected, middleware.NewAuthMiddleware(r.authService))

            // Recruitment Module Routes
            recruitment := protected.Group("/recruitment")
            {
                // Job Posting Routes
                jobPostingService := services.NewJobPostingService(r.db)
                jobPostingHandler := NewJobPostingHandler(jobPostingService, r.authService)
                jobPostingHandler.RegisterRoutes(recruitment)

                // Candidate Routes
                candidateService := services.NewCandidateService(r.db)
                candidateHandler := NewCandidateHandler(candidateService, r.authService)
                candidateHandler.RegisterRoutes(recruitment)

                // Application Routes
                applicationService := services.NewApplicationService(r.db)
                applicationHandler := NewApplicationHandler(applicationService, candidateService, r.authService)
                applicationHandler.RegisterRoutes(recruitment)
            }

            // Onboarding Module Routes
            onboardingTemplateService := services.NewOnboardingTemplateService(r.db)
            onboardingChecklistService := services.NewOnboardingChecklistService(r.db)
            onboardingHandler := NewOnboardingHandler(onboardingTemplateService, onboardingChecklistService, r.authService)
            onboardingHandler.RegisterRoutes(protected)

            // Training/LMS Module Routes
            training := protected.Group("/training")
            {
                trainingService := services.NewTrainingService(r.db)
                trainingHandler := NewTrainingHandler(trainingService)

                // Categories
                training.POST("/categories", trainingHandler.CreateCategory)
                training.GET("/categories", trainingHandler.GetCategories)

                // Courses
                training.POST("/courses", trainingHandler.CreateCourse)
                training.GET("/courses", trainingHandler.ListCourses)
                training.GET("/courses/:id", trainingHandler.GetCourse)
                training.PUT("/courses/:id", trainingHandler.UpdateCourse)
                training.POST("/courses/:id/publish", trainingHandler.PublishCourse)
                training.POST("/courses/:id/archive", trainingHandler.ArchiveCourse)
                training.GET("/courses/:id/statistics", trainingHandler.GetCourseStatistics)

                // Modules and Content
                training.POST("/courses/:id/modules", trainingHandler.CreateModule)
                training.POST("/modules/:moduleId/content", trainingHandler.CreateContent)

                // Enrollments
                training.POST("/enrollments", trainingHandler.EnrollEmployee)
                training.GET("/enrollments/:id", trainingHandler.GetEnrollment)
                training.GET("/employees/:employeeId/enrollments", trainingHandler.GetEmployeeEnrollments)
                training.POST("/enrollments/:id/start", trainingHandler.StartCourse)
                training.PUT("/progress", trainingHandler.UpdateContentProgress)

                // Certificates
                training.POST("/enrollments/:id/certificate", trainingHandler.IssueCertificate)
            }

            // Performance Management Module Routes
            performance := protected.Group("/performance")
            {
                performanceService := services.NewPerformanceService(r.db)
                performanceHandler := NewPerformanceHandler(performanceService)

                // Review Cycles
                performance.POST("/cycles", performanceHandler.CreateReviewCycle)
                performance.GET("/cycles", performanceHandler.ListReviewCycles)
                performance.GET("/cycles/:id", performanceHandler.GetReviewCycle)
                performance.POST("/cycles/:id/activate", performanceHandler.ActivateReviewCycle)

                // Performance Reviews
                performance.GET("/reviews/:id", performanceHandler.GetPerformanceReview)
                performance.GET("/employees/:employeeId/reviews", performanceHandler.GetEmployeeReviews)
                performance.POST("/reviews/:id/start-self", performanceHandler.StartSelfReview)
                performance.POST("/reviews/:id/self-review", performanceHandler.SubmitSelfReview)
                performance.POST("/reviews/:id/manager-review", performanceHandler.SubmitManagerReview)
                performance.POST("/reviews/:id/acknowledge", performanceHandler.AcknowledgeReview)

                // Goals
                performance.POST("/goals", performanceHandler.CreateGoal)
                performance.GET("/goals/:id", performanceHandler.GetGoal)
                performance.GET("/employees/:employeeId/goals", performanceHandler.GetEmployeeGoals)
                performance.PUT("/goals/:id/progress", performanceHandler.UpdateGoalProgress)

                // Feedback
                performance.POST("/feedback", performanceHandler.CreateFeedback)
                performance.GET("/employees/:employeeId/feedback", performanceHandler.GetEmployeeFeedback)

                // One-on-Ones
                performance.POST("/one-on-ones", performanceHandler.CreateOneOnOne)
                performance.GET("/employees/:employeeId/one-on-ones", performanceHandler.GetEmployeeOneOnOnes)
                performance.POST("/one-on-ones/:id/complete", performanceHandler.CompleteOneOnOne)
            }

            // Benefits Administration Module Routes
            benefits := protected.Group("/benefits")
            {
                benefitsService := services.NewBenefitsService(r.db)
                benefitsHandler := NewBenefitsHandler(benefitsService)

                // Benefit Plans
                benefits.POST("/plans", benefitsHandler.CreateBenefitPlan)
                benefits.GET("/plans", benefitsHandler.ListBenefitPlans)
                benefits.GET("/plans/:id", benefitsHandler.GetBenefitPlan)

                // Enrollment Periods
                benefits.POST("/enrollment-periods", benefitsHandler.CreateEnrollmentPeriod)
                benefits.GET("/enrollment-periods/active", benefitsHandler.GetActiveEnrollmentPeriod)
                benefits.POST("/enrollment-periods/:id/open", benefitsHandler.OpenEnrollmentPeriod)

                // Enrollments
                benefits.POST("/enrollments", benefitsHandler.EnrollInBenefit)
                benefits.GET("/enrollments/:id", benefitsHandler.GetEnrollment)
                benefits.GET("/employees/:employeeId/enrollments", benefitsHandler.GetEmployeeEnrollments)
                benefits.POST("/enrollments/:id/approve", benefitsHandler.ApproveEnrollment)
                benefits.POST("/enrollments/:id/decline", benefitsHandler.DeclineEnrollment)
                benefits.POST("/enrollments/:id/terminate", benefitsHandler.TerminateEnrollment)
                benefits.POST("/enrollments/:id/waive", benefitsHandler.WaiveBenefit)

                // Dependents
                benefits.POST("/employees/:employeeId/dependents", benefitsHandler.CreateDependent)
                benefits.GET("/employees/:employeeId/dependents", benefitsHandler.GetEmployeeDependents)
                benefits.POST("/dependents/:id/verify", benefitsHandler.VerifyDependent)

                // Life Events
                benefits.POST("/employees/:employeeId/life-events", benefitsHandler.CreateLifeEvent)
                benefits.GET("/employees/:employeeId/life-events", benefitsHandler.GetEmployeeLifeEvents)
                benefits.POST("/life-events/:id/approve", benefitsHandler.ApproveLifeEvent)

                // Beneficiaries
                benefits.POST("/beneficiaries", benefitsHandler.CreateBeneficiary)
                benefits.GET("/enrollments/:id/beneficiaries", benefitsHandler.GetEnrollmentBeneficiaries)
            }

            // Time Tracking Module Routes
            timeTracking := protected.Group("/time-tracking")
            {
                timeTrackingService := services.NewTimeTrackingService(r.db)
                timeTrackingHandler := NewTimeTrackingHandler(timeTrackingService)

                // Projects
                timeTracking.POST("/projects", timeTrackingHandler.CreateProject)
                timeTracking.GET("/projects", timeTrackingHandler.ListProjects)
                timeTracking.GET("/projects/:id", timeTrackingHandler.GetProject)
                timeTracking.POST("/projects/:id/tasks", timeTrackingHandler.CreateProjectTask)
                timeTracking.POST("/projects/:id/members", timeTrackingHandler.AddProjectMember)

                // Timesheets
                timeTracking.POST("/timesheets", timeTrackingHandler.CreateTimesheet)
                timeTracking.GET("/timesheets/:id", timeTrackingHandler.GetTimesheet)
                timeTracking.GET("/employees/:employeeId/timesheets", timeTrackingHandler.GetEmployeeTimesheets)
                timeTracking.POST("/timesheets/:id/submit", timeTrackingHandler.SubmitTimesheet)
                timeTracking.POST("/timesheets/:id/approve", timeTrackingHandler.ApproveTimesheet)
                timeTracking.POST("/timesheets/:id/reject", timeTrackingHandler.RejectTimesheet)

                // Time Entries
                timeTracking.POST("/time-entries", timeTrackingHandler.CreateTimeEntry)
                timeTracking.GET("/time-entries/:id", timeTrackingHandler.GetTimeEntry)
                timeTracking.GET("/employees/:employeeId/time-entries", timeTrackingHandler.GetEmployeeTimeEntries)

                // Clock In/Out
                timeTracking.POST("/employees/:employeeId/clock-in", timeTrackingHandler.ClockIn)
                timeTracking.POST("/employees/:employeeId/clock-out", timeTrackingHandler.ClockOut)
                timeTracking.POST("/employees/:employeeId/start-break", timeTrackingHandler.StartBreak)
                timeTracking.POST("/employees/:employeeId/end-break", timeTrackingHandler.EndBreak)
                timeTracking.GET("/employees/:employeeId/clock-status", timeTrackingHandler.GetCurrentClockStatus)

                // Time-Off Balances
                timeTracking.GET("/employees/:employeeId/time-off-balances", timeTrackingHandler.GetTimeOffBalances)
                timeTracking.PUT("/time-off-balances/:id", timeTrackingHandler.UpdateTimeOffBalance)

                // Holidays
                timeTracking.POST("/holidays", timeTrackingHandler.CreateHoliday)
                timeTracking.GET("/holidays", timeTrackingHandler.ListHolidays)
            }

            // Expense Management Module Routes
            expenses := protected.Group("/expenses")
            {
                expenseService := services.NewExpenseService(r.db)
                expenseHandler := NewExpenseHandler(expenseService)

                // Categories
                expenses.POST("/categories", expenseHandler.CreateCategory)
                expenses.GET("/categories", expenseHandler.ListCategories)

                // Expense Reports
                expenses.POST("/reports", expenseHandler.CreateExpenseReport)
                expenses.GET("/reports", expenseHandler.ListExpenseReports)
                expenses.GET("/reports/:id", expenseHandler.GetExpenseReport)
                expenses.POST("/reports/:id/submit", expenseHandler.SubmitExpenseReport)
                expenses.POST("/reports/:id/approve", expenseHandler.ApproveExpenseReport)
                expenses.POST("/reports/:id/reject", expenseHandler.RejectExpenseReport)
                expenses.POST("/reports/:id/reimburse", expenseHandler.ProcessReimbursement)
                expenses.POST("/reports/:id/mark-paid", expenseHandler.MarkExpenseReportPaid)

                // Expense Items
                expenses.POST("/reports/:id/items", expenseHandler.AddExpenseItem)
                expenses.GET("/items/:id", expenseHandler.GetExpenseItem)
                expenses.DELETE("/items/:id", expenseHandler.DeleteExpenseItem)

                // Advance Payments
                expenses.POST("/advance-payments", expenseHandler.RequestAdvancePayment)
                expenses.GET("/employees/:employeeId/advance-payments", expenseHandler.GetEmployeeAdvancePayments)
                expenses.POST("/advance-payments/:id/approve", expenseHandler.ApproveAdvancePayment)
                expenses.POST("/advance-payments/:id/issue", expenseHandler.IssueAdvancePayment)
                expenses.POST("/advance-payments/:id/reconcile", expenseHandler.ReconcileAdvance)
            }

            // Document Management Module Routes
            documents := protected.Group("/documents")
            {
                documentService := services.NewDocumentService(r.db)
                documentHandler := NewDocumentHandler(documentService)

                // Categories
                documents.POST("/categories", documentHandler.CreateCategory)
                documents.GET("/categories", documentHandler.ListCategories)

                // Documents
                documents.POST("", documentHandler.CreateDocument)
                documents.GET("", documentHandler.ListDocuments)
                documents.GET("/:id", documentHandler.GetDocument)
                documents.PUT("/:id", documentHandler.UpdateDocument)
                documents.DELETE("/:id", documentHandler.DeleteDocument)
                documents.POST("/:id/archive", documentHandler.ArchiveDocument)

                // Document Versions
                documents.POST("/:id/versions", documentHandler.CreateDocumentVersion)

                // Templates
                documents.POST("/templates", documentHandler.CreateTemplate)
                documents.GET("/templates", documentHandler.ListTemplates)

                // Signature Requests
                documents.POST("/signature-requests", documentHandler.CreateSignatureRequest)
                documents.GET("/signature-requests/:id", documentHandler.GetSignatureRequest)
                documents.POST("/signature-requests/:id/signers/:signerId/sign", documentHandler.SignDocument)
                documents.POST("/signature-requests/:id/signers/:signerId/decline", documentHandler.DeclineSignature)

                // Document Sharing
                documents.POST("/share", documentHandler.ShareDocument)
                documents.GET("/:id/shares", documentHandler.GetDocumentShares)
                documents.DELETE("/shares/:id", documentHandler.RevokeDocumentShare)

                // Employee Documents
                documents.POST("/employees/:employeeId/submit", documentHandler.SubmitEmployeeDocument)
                documents.GET("/employees/:employeeId/documents", documentHandler.GetEmployeeDocuments)
                documents.POST("/employee-documents/:id/verify", documentHandler.VerifyEmployeeDocument)
                documents.POST("/employee-documents/:id/reject", documentHandler.RejectEmployeeDocument)
            }

            // Admin only routes - User Management
            userService := services.NewUserService(r.db)
            userHandler := NewUserHandler(userService, r.authService)
            userHandler.RegisterRoutes(protected, middleware.NewAuthMiddleware(r.authService))
            
            // HR only routes (includes all HR sub-roles with collar-type filtering)
            hr := protected.Group("")
            hr.Use(middleware.NewAuthMiddleware(r.authService).RequireRole("admin", "hr", "hr_and_pr", "hr_blue_gray", "hr_white"))
            {
                // HR endpoints here
            }
        }
    }
}
