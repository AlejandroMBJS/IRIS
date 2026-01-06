/*
Package config - IRIS Payroll System Application Configuration

==============================================================================
FILE: internal/config/app_config.go
==============================================================================

DESCRIPTION:
    Central application configuration for the IRIS Payroll backend.
    Loads settings from environment variables, .env files, and optionally
    from HashiCorp Vault for production secrets management.

USER PERSPECTIVE:
    - Controls server port, database connection, JWT settings
    - Manages email configuration for notifications
    - Handles payroll-specific settings like currency

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new configuration fields, new env var mappings
    ⚠️  CAUTION: Changing default values (may affect existing deployments)
    ❌  DO NOT modify: Security-critical defaults without review
    📝  Always add new fields with sensible defaults

SYNTAX EXPLANATION:
    - AppConfig struct: Holds all configuration with mapstructure tags
    - LoadAppConfig(): Entry point called from main.go
    - godotenv.Load(): Loads .env file if present
    - Vault integration: Optional, for production secret management

CONFIGURATION SOURCES (priority order):
    1. HashiCorp Vault (if VAULT_ADDR is set)
    2. Environment variables
    3. .env file
    4. Default values in DefaultAppConfig()

==============================================================================
*/
package config

import (
	"context"
	"fmt"

	"os"
	"strconv"

	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"backend/internal/config/payroll"
)

// AppConfig contains all application configuration
type AppConfig struct {
	// Server configuration
	ServerPort int    `mapstructure:"SERVER_PORT"`
	Env        string `mapstructure:"ENVIRONMENT"`

	// Database configuration
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	DBDriver    string `mapstructure:"DB_DRIVER"`

	// JWT configuration
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTExpirationHours int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	JWTRefreshHours    int    `mapstructure:"JWT_REFRESH_HOURS"`

	// Security
	BcryptCost int `mapstructure:"BCRYPT_COST"`

	// Logging
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// CORS
	CORSAllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	// Rate limiting
	RateLimitRequestsPerMinute int `mapstructure:"RATE_LIMIT_REQUESTS_PER_MINUTE"`

	// Payroll specific
	DefaultPayrollCurrency string `mapstructure:"DEFAULT_PAYROLL_CURRENCY"`

	// Email (optional)
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	EmailFrom    string `mapstructure:"EMAIL_FROM"`

	// Payroll configuration (loaded from JSON)
	PayrollConfig *payroll.PayrollConfig

	// Vault client
	VaultClient *api.Client
}

// DefaultAppConfig returns configuration with default values
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ServerPort:                 8080,
		Env:                        "development",
		DatabaseURL:                "./iris_payroll.db",
		DBDriver:                   "sqlite",
		JWTSecret:                  "your-secret-key-change-in-production",
		JWTExpirationHours:         24,
		JWTRefreshHours:            168,
		BcryptCost:                 12,
		LogLevel:                   "info",
		CORSAllowedOrigins:         "*",
		RateLimitRequestsPerMinute: 60,
		DefaultPayrollCurrency:     "MXN",
		SMTPHost:                   "",
		SMTPPort:                   587,
		SMTPUsername:               "",
		SMTPPassword:               "",
		EmailFrom:                  "noreply@iristalent.com",
		PayrollConfig:              nil,
	}
}

// LoadAppConfig loads all application configuration
func LoadAppConfig(configDir string) (*AppConfig, error) {
	// Load environment variables
	_ = godotenv.Load()

	config := DefaultAppConfig()

	// Load from environment variables
	// Use env var or default if not set
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.ServerPort = port
		}
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Env = env
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}
	if dbDriver := os.Getenv("DB_DRIVER"); dbDriver != "" {
		config.DBDriver = dbDriver
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = jwtSecret
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}
	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		config.CORSAllowedOrigins = corsOrigins
	}
	if payrollCurrency := os.Getenv("DEFAULT_PAYROLL_CURRENCY"); payrollCurrency != "" {
		config.DefaultPayrollCurrency = payrollCurrency
	}
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.SMTPHost = smtpHost
	}
	if smtpUsername := os.Getenv("SMTP_USERNAME"); smtpUsername != "" {
		config.SMTPUsername = smtpUsername
	}
	if smtpPassword := os.Getenv("SMTP_PASSWORD"); smtpPassword != "" {
		config.SMTPPassword = smtpPassword
	}
	if emailFrom := os.Getenv("EMAIL_FROM"); emailFrom != "" {
		config.EmailFrom = emailFrom
	}

	// Load secrets from Vault if configured
	if os.Getenv("VAULT_ADDR") != "" {
		if err := loadFromVault(config); err != nil {
			// Log the error but continue, allowing fallback to env vars
			fmt.Printf("Warning: Could not load secrets from Vault: %v\n", err)
		}
	}

	// Load payroll configuration from JSON files
	if configDir != "" {
		payrollLoader := payroll.NewPayrollConfigLoader(configDir)
		payrollConfig, err := payrollLoader.Load()
		if err != nil {
			return nil, fmt.Errorf("error loading payroll config: %w", err)
		}
		config.PayrollConfig = payrollConfig
	}

	return config, nil
}


// loadFromVault connects to Vault and loads secrets.
func loadFromVault(c *AppConfig) error {
	vaultConfig := api.DefaultConfig() // VAULT_ADDR and VAULT_TOKEN are read from env vars

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}
	c.VaultClient = client

	// Example: Reading secrets from kv-v2 engine at path "secret/iris-payroll"
	secretPath := os.Getenv("VAULT_SECRET_PATH")
	if secretPath == "" {
		secretPath = "secret/data/iris-payroll" // Default path
	}

	secret, err := client.KVv2(secretPath).Get(context.Background(), "")
	if err != nil {
		return fmt.Errorf("failed to read secrets from vault path %s: %w", secretPath, err)
	}

	if dbURL, ok := secret.Data["DATABASE_URL"].(string); ok {
		c.DatabaseURL = dbURL
	}
	if jwtSecret, ok := secret.Data["JWT_SECRET"].(string); ok {
		c.JWTSecret = jwtSecret
	}
	if smtpPassword, ok := secret.Data["SMTP_PASSWORD"].(string); ok {
		c.SMTPPassword = smtpPassword
	}

	fmt.Println("Successfully loaded secrets from Vault")
	return nil
}

// IsProduction returns true if environment is production
func (c *AppConfig) IsProduction() bool {
    return c.Env == "production"
}

// IsDevelopment returns true if environment is development
func (c *AppConfig) IsDevelopment() bool {
    return c.Env == "development"
}

// IsTesting returns true if environment is testing
func (c *AppConfig) IsTesting() bool {
    return c.Env == "testing"
}
