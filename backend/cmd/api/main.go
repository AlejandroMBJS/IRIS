/*
Package main - IRIS Payroll System Backend Entry Point

==============================================================================
FILE: cmd/api/main.go
==============================================================================

DESCRIPTION:
    This is the main entry point for the IRIS Payroll System backend API server.
    It initializes all core components and starts the HTTP server that handles
    all payroll-related operations.

USER PERSPECTIVE:
    - This file starts the backend server that powers all payroll operations
    - Users interact with this indirectly through the frontend web application
    - When IT staff runs "go build cmd/api/main.go", this creates the server executable
    - The server handles: authentication, employee management, payroll calculations,
      incidences (absences, vacations), and report generation

DEVELOPER GUIDELINES:
    ⚠️  MODIFY WITH CAUTION - This is a critical system file
    ✅  OK to modify: CORS origins, server timeouts, port configuration
    ❌  DO NOT modify: Service initialization order, graceful shutdown logic
    📝  When adding new services: Add them in the initialization section (lines 47-50)
        and pass them to setupRouter()

SYNTAX EXPLANATION:
    - package main: Go requires 'main' package for executable programs
    - import (...): Multi-line import block for external and internal packages
    - func main(): Entry point function, executed when program starts
    - go func() {...}(): Goroutine - runs server in background thread
    - signal.Notify(): Captures OS signals (Ctrl+C) for graceful shutdown
    - context.WithTimeout(): Creates cancellable context with deadline

ARCHITECTURE:
    main() → LoadConfig → SetupLogger → ConnectDB → InitServices → StartServer
                                                                        ↓
    ShutdownServer ← WaitForSignal ← ListenAndServe ← setupRouter()

DEPENDENCIES:
    External:
    - github.com/gin-gonic/gin: HTTP web framework (fast, minimalist)
    - github.com/gin-contrib/cors: Cross-Origin Resource Sharing middleware
    - github.com/sirupsen/logrus: Structured logging library
    - gorm.io/gorm: ORM for database operations

    Internal:
    - backend/internal/api: HTTP handlers and routing
    - backend/internal/config: Application configuration loading
    - backend/internal/database: Database connection and migrations
    - backend/internal/services: Business logic layer

==============================================================================
*/
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"backend/internal/api"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/logger"
	"backend/internal/services"
)
    
    func main() {
            // Load configuration
            cfg, err := config.LoadAppConfig("./configs")
        if err != nil {
            log.Fatalf("Failed to load application configuration: %v", err)
        }
        
        // Setup logger
        appLogger := logger.Setup(cfg.Env)
        
        // Initialize database
        db, err := database.NewConnection(cfg.DatabaseURL, cfg.DBDriver)
        if err != nil {
            appLogger.Fatalf("Failed to connect to database: %v", err)
        }    
    // Auto migrate (only in development)
    if cfg.Env == "development" {
        if err := database.Migrate(db); err != nil {
            appLogger.Warnf("Migration failed: %v", err)
        }
    }

    // Seed default incidence categories
    categoryService := services.NewIncidenceCategoryService(db)
    if err := categoryService.SeedDefaultCategories(); err != nil {
        appLogger.Warnf("Failed to seed incidence categories: %v", err)
    } else {
        appLogger.Info("Incidence categories verified/seeded")
    }

    // Migrate legacy category strings to category_id
    if err := categoryService.MigrateLegacyCategories(); err != nil {
        appLogger.Warnf("Failed to migrate legacy categories: %v", err)
    }

    // Seed default tipo mappings for payroll export
    incidenceConfigService := services.NewIncidenceConfigService(db)
    if err := incidenceConfigService.SeedDefaultMappings(); err != nil {
        appLogger.Warnf("Failed to seed tipo mappings: %v", err)
    } else {
        appLogger.Info("Tipo mappings verified/seeded")
    }

    // Seed default role inheritances for permission management
    roleInheritanceService := services.NewRoleInheritanceService(db)
    if err := roleInheritanceService.SeedDefaultInheritances(); err != nil {
        appLogger.Warnf("Failed to seed role inheritances: %v", err)
    } else {
        appLogger.Info("Role inheritances verified/seeded")
    }

    // Seed default permission matrix for role-based access control
    permissionService := services.NewPermissionService(db)
    if err := permissionService.SeedDefaultPermissions(); err != nil {
        appLogger.Warnf("Failed to seed permissions: %v", err)
    } else {
        appLogger.Info("Permission matrix verified/seeded")
    }

    // Initialize services
    authService := services.NewAuthService(db, cfg)
    employeeService := services.NewEmployeeService(db)
    payrollService := services.NewPayrollService(db, cfg)
    payrollPeriodService := services.NewPayrollPeriodService(db)

    // Auto-generate current payroll periods on startup
    periods, err := payrollPeriodService.GenerateCurrentPeriods()
    if err != nil {
        appLogger.Warnf("Failed to generate current periods: %v", err)
    } else {
        appLogger.Infof("Generated/verified %d payroll periods", len(periods))
    }

    // Setup router
    router := setupRouter(cfg, db, appLogger, authService, employeeService, payrollService)
    
    // Create server
    srv := &http.Server{
        Addr:         ":" + strconv.Itoa(cfg.ServerPort),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server in goroutine
    go func() {
        appLogger.Infof("Starting server on port %s in %s mode", strconv.Itoa(cfg.ServerPort), cfg.Env)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            appLogger.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    appLogger.Info("Shutting down server...")
    
    // Give outstanding requests 30 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        appLogger.Fatalf("Server forced to shutdown: %v", err)
    }
    
    // Close database connection
    sqlDB, err := db.DB()
    if err == nil {
        sqlDB.Close()
    }
    
    appLogger.Info("Server exited properly")
}

func setupRouter(
    cfg *config.AppConfig,
    db *gorm.DB,
    appLogger *logrus.Logger,
    authService *services.AuthService,
    employeeService *services.EmployeeService,
    payrollService *services.PayrollService,
) *gin.Engine {
    // Set Gin mode
    if cfg.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()

    // CORS configuration - must be applied BEFORE routes
    // This allows the frontend to make cross-origin requests to the backend
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:8080", "http://localhost:8081", "http://localhost"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    // Global middleware
    router.Use(logger.GinLogger(appLogger))
    router.Use(gin.Recovery())

    // Health checks
    healthHandler := api.NewHealthHandler(db)
    router.GET("/health", healthHandler.HealthCheck)
    router.GET("/ready", healthHandler.ReadyCheck)
    router.GET("/live", healthHandler.LivenessCheck)

    // API v1 routes
    apiRouter := api.NewRouter(db, cfg) // Pass cfg to NewRouter
    apiRouter.Setup(router.Group("/api/v1"))

    return router
}