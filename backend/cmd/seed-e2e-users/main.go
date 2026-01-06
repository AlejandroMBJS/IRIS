package main

import (
	"fmt"
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/models/enums"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

/*
E2E Test Users Seeder

Creates test users for ALL roles to enable comprehensive E2E testing:
- admin, hr, accountant, payroll_staff, viewer
- supervisor, manager, employee
- hr_and_pr, sup_and_gm, hr_blue_gray, hr_white

USAGE:
	go run cmd/seed-e2e-users/main.go

PASSWORD:
	All users have password: "Test123456!"

OUTPUT:
	Prints credentials for .env.test configuration
*/

var testPassword = "Test123456!"

func main() {
	fmt.Println("🌱 Seeding E2E Test Users...")

	// Load config
	cfg, err := config.LoadAppConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg.DatabaseURL, cfg.DBDriver)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Hash password once (same for all users)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Get or create test company
	company := getOrCreateTestCompany(db)

	// Define all test users
	testUsers := []struct {
		Email    string
		FullName string
		Role     enums.UserRole
	}{
		// Core roles
		{"e2e.admin@test.com", "E2E Admin User", enums.RoleAdmin},
		{"e2e.hr@test.com", "E2E HR User", enums.RoleHR},
		{"e2e.accountant@test.com", "E2E Accountant User", enums.RoleAccountant},
		{"e2e.payroll@test.com", "E2E Payroll Staff User", enums.RolePayrollStaff},
		{"e2e.viewer@test.com", "E2E Viewer User", enums.RoleViewer},

		// Approval workflow roles
		{"e2e.supervisor@test.com", "E2E Supervisor User", enums.RoleSupervisor},
		{"e2e.manager@test.com", "E2E Manager User", enums.RoleManager},
		{"e2e.employee@test.com", "E2E Employee User", enums.RoleEmployee},

		// Combined roles
		{"e2e.hr_and_pr@test.com", "E2E HR+Payroll User", enums.RoleHRAndPR},
		{"e2e.sup_and_gm@test.com", "E2E Supervisor+GM User", enums.RoleSupAndGM},

		// Collar-specific HR
		{"e2e.hr_blue_gray@test.com", "E2E HR Blue/Gray Collar", enums.RoleHRBlueGray},
		{"e2e.hr_white@test.com", "E2E HR White Collar", enums.RoleHRWhite},
	}

	fmt.Println("\n📋 Creating test users...")
	createdCount := 0
	skippedCount := 0

	for _, userData := range testUsers {
		// Check if user already exists
		var existingUser models.User
		result := db.Where("email = ?", userData.Email).First(&existingUser)

		if result.Error == nil {
			// User exists, update it
			existingUser.PasswordHash = string(hashedPassword)
			existingUser.Role = userData.Role
			existingUser.FullName = userData.FullName
			existingUser.IsActive = true
			existingUser.CompanyID = company.ID

			if err := db.Save(&existingUser).Error; err != nil {
				log.Printf("⚠️  Failed to update %s: %v", userData.Email, err)
				continue
			}
			fmt.Printf("   ✓ Updated: %-30s (Role: %-15s)\n", userData.Email, userData.Role)
			skippedCount++
		} else if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			user := models.User{
				BaseModel: models.BaseModel{
					ID: uuid.New(),
				},
				Email:        userData.Email,
				PasswordHash: string(hashedPassword),
				Role:         userData.Role,
				FullName:     userData.FullName,
				IsActive:     true,
				CompanyID:    company.ID,
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("⚠️  Failed to create %s: %v", userData.Email, err)
				continue
			}
			fmt.Printf("   ✓ Created: %-30s (Role: %-15s)\n", userData.Email, userData.Role)
			createdCount++
		} else {
			log.Printf("⚠️  Error checking %s: %v", userData.Email, result.Error)
		}
	}

	fmt.Printf("\n✅ Seed completed!\n")
	fmt.Printf("   Created: %d users\n", createdCount)
	fmt.Printf("   Updated: %d users\n", skippedCount)
	fmt.Printf("   Total: %d users\n\n", len(testUsers))

	// Print .env.test configuration
	printEnvConfig(testUsers)

	// Print test credentials
	printTestCredentials(testUsers)
}

func getOrCreateTestCompany(db *gorm.DB) *models.Company {
	var company models.Company
	result := db.Where("name = ?", "E2E Test Company").First(&company)

	if result.Error == gorm.ErrRecordNotFound {
		// Create test company
		company = models.Company{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Name:     "E2E Test Company",
			RFC:      "E2E123456789",
			Address:  "Test Street 123, Test City",
			Phone:    "5551234567",
			Email:    "contact@e2etest.com",
			IsActive: true,
		}

		if err := db.Create(&company).Error; err != nil {
			log.Fatalf("Failed to create test company: %v", err)
		}
		fmt.Println("✓ Created test company: E2E Test Company")
	} else if result.Error != nil {
		log.Fatalf("Failed to query company: %v", result.Error)
	} else {
		fmt.Println("✓ Using existing test company: E2E Test Company")
	}

	return &company
}

func printEnvConfig(users []struct {
	Email    string
	FullName string
	Role     enums.UserRole
}) {
	fmt.Println("📝 Add this to /frontend/.env.test:")
	fmt.Println("=" + string(make([]byte, 78)))
	fmt.Println("# E2E Test Users - All roles")
	fmt.Printf("E2E_PASSWORD=%s\n\n", testPassword)

	for _, user := range users {
		envVar := getEnvVarName(user.Role)
		fmt.Printf("E2E_EMAIL_%s=%s\n", envVar, user.Email)
	}
	fmt.Println("=" + string(make([]byte, 78)))
	fmt.Println()
}

func printTestCredentials(users []struct {
	Email    string
	FullName string
	Role     enums.UserRole
}) {
	fmt.Println("🔑 Test Credentials (all have same password):")
	fmt.Println("=" + string(make([]byte, 78)))
	fmt.Printf("Password: %s\n\n", testPassword)

	for _, user := range users {
		fmt.Printf("%-30s  →  %-15s  (%s)\n", user.Email, user.Role, user.FullName)
	}
	fmt.Println("=" + string(make([]byte, 78)))
}

func getEnvVarName(role enums.UserRole) string {
	switch role {
	case enums.RoleAdmin:
		return "ADMIN"
	case enums.RoleHR:
		return "HR"
	case enums.RoleAccountant:
		return "ACCOUNTANT"
	case enums.RolePayrollStaff:
		return "PAYROLL_STAFF"
	case enums.RoleViewer:
		return "VIEWER"
	case enums.RoleSupervisor:
		return "SUPERVISOR"
	case enums.RoleManager:
		return "MANAGER"
	case enums.RoleEmployee:
		return "EMPLOYEE"
	case enums.RoleHRAndPR:
		return "HR_AND_PR"
	case enums.RoleSupAndGM:
		return "SUP_AND_GM"
	case enums.RoleHRBlueGray:
		return "HR_BLUE_GRAY"
	case enums.RoleHRWhite:
		return "HR_WHITE"
	default:
		return string(role)
	}
}
