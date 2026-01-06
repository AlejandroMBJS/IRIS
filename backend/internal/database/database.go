/*
Package database - IRIS Payroll Database Connection Management

==============================================================================
FILE: internal/database/database.go
==============================================================================

DESCRIPTION:
    Handles database connection creation and configuration using GORM ORM.
    Supports multiple database drivers (PostgreSQL, SQLite) with connection
    pooling and logging configuration.

USER PERSPECTIVE:
    - Establishes database connection at application startup
    - Configures connection pooling for production performance
    - Supports SQLite for development and PostgreSQL for production

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new database drivers, adjust pool settings
    ⚠️  CAUTION: Changing pool settings (affects performance)
    ❌  DO NOT modify: Connection string handling without security review
    📝  Test with both SQLite and PostgreSQL before deployment

SYNTAX EXPLANATION:
    - NewConnection(): Factory function for database connections
    - gorm.Dialector: Database driver abstraction
    - SetMaxIdleConns/SetMaxOpenConns: Connection pool configuration
    - SetConnMaxLifetime: Prevents stale connections

CONNECTION POOL DEFAULTS:
    - MaxIdleConns: 10 (idle connections kept open)
    - MaxOpenConns: 100 (maximum concurrent connections)
    - ConnMaxLifetime: 1 hour (connection refresh interval)

==============================================================================
*/
package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection creates and returns a new GORM database connection.
func NewConnection(dbURL, dbDriver string) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:           logger.Info, // Log level
			Colorful:           true,        // Disable color
		},
	)

	var dialector gorm.Dialector
	switch dbDriver {
	case "postgres":
		dialector = postgres.Open(dbURL)
	case "sqlite":
		dialector = sqlite.Open(dbURL)
	default:
		log.Fatalf("Unsupported database driver: %s", dbDriver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, err
	}

	// Connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}