package services

import (
	"backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupExcelExportTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate test tables
	err = db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Employee{},
		&models.PayrollPeriod{},
		&models.IncidenceCategory{},
		&models.IncidenceType{},
		&models.Incidence{},
		&models.IncidenceTipoMapping{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestIncidence(db *gorm.DB, templateType string, startDate, endDate time.Time, quantity float64) *models.Incidence {
	// Create company
	company := &models.Company{Name: "Test Company"}
	db.Create(company)

	// Create employee
	dateOfBirth := time.Now().AddDate(-30, 0, 0)
	hireDate := dateOfBirth.AddDate(25, 0, 0)
	employee := &models.Employee{
		CompanyID:        company.ID,
		EmployeeNumber:   "EMP001",
		FirstName:        "John",
		LastName:         "Doe",
		DateOfBirth:      dateOfBirth,
		HireDate:         hireDate,
		Gender:           "male",
		RFC:              "XAXX010101000",
		CURP:             "XAXX010101HDFXXX00",
		EmployeeType:     "permanent",
		EmploymentStatus: "active",
		CollarType:       "blue_collar",
		DailySalary:      500.0,
		Regime:           "02",
	}
	db.Create(employee)

	// Create user
	user := &models.User{
		CompanyID:    company.ID,
		EmployeeID:   &employee.ID,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         "employee",
		FullName:     "John Doe",
		IsActive:     true,
	}
	db.Create(user)

	// Create payroll period
	period := &models.PayrollPeriod{
		PeriodCode:   "2024-W01",
		PeriodType:   "weekly",
		Frequency:    "weekly",
		StartDate:    startDate.AddDate(0, 0, -7),
		EndDate:      endDate.AddDate(0, 0, 7),
		PaymentDate:  endDate.AddDate(0, 0, 10),
		Status:       "open",
		Year:         startDate.Year(),
		PeriodNumber: 1,
	}
	db.Create(period)

	// Create incidence category and type
	category := &models.IncidenceCategory{
		Name:        "Vacations",
		Description: "Vacation days",
		IsActive:    true,
	}
	db.Create(category)

	incidenceType := &models.IncidenceType{
		CategoryID:  &category.ID,
		Name:        "Vacation",
		Category:    "vacation",
		EffectType:  "positive",
		Description: "Paid vacation",
	}
	db.Create(incidenceType)

	// Create tipo mapping
	tipoCode := "7"
	motivo := "EXTRAS"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		Motivo:          &motivo,
		TemplateType:    templateType,
		HoursMultiplier: 8.0,
	}
	db.Create(mapping)

	// Create incidence
	incidence := &models.Incidence{
		EmployeeID:          employee.ID,
		IncidenceTypeID:     incidenceType.ID,
		PayrollPeriodID:     period.ID,
		StartDate:           startDate,
		EndDate:             endDate,
		Quantity:            quantity,
		Status:              "approved",
		Comments:            "Test incidence",
		LateApprovalFlag:    false,
		ExcludedFromPayroll: false,
	}
	db.Create(incidence)

	// Reload with relationships
	db.Preload("Employee").Preload("IncidenceType").First(incidence, incidence.ID)

	return incidence
}

func TestGenerateVacacionesExcel_SingleRow(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Create a single vacation incidence (5 days)
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)   // Friday
	incidence := createTestIncidence(db, "vacaciones", startDate, endDate, 5.0)

	incidences := []models.Incidence{*incidence}

	// Generate Excel
	excelBytes, err := service.GenerateVacacionesExcel(incidences)

	assert.NoError(t, err)
	assert.NotNil(t, excelBytes)
	assert.Greater(t, len(excelBytes), 0, "Excel file should have content")
}

func TestGenerateFaltasExtrasExcel_SingleDay(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Create a single-day absence
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	incidence := createTestIncidence(db, "faltas_extras", startDate, endDate, 1.0)

	incidences := []models.Incidence{*incidence}

	// Generate Excel
	excelBytes, err := service.GenerateFaltasExtrasExcel(incidences)

	assert.NoError(t, err)
	assert.NotNil(t, excelBytes)
	assert.Greater(t, len(excelBytes), 0, "Excel file should have content")
}

func TestExpandMultiDayAbsence_SingleDay(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Single day incidence
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	incidence := createTestIncidence(db, "faltas_extras", startDate, endDate, 1.0)

	// Load mapping
	var mapping models.IncidenceTipoMapping
	db.Where("incidence_type_id = ?", incidence.IncidenceTypeID).First(&mapping)

	// Expand
	rows := service.expandMultiDayAbsence(*incidence, mapping)

	assert.Equal(t, 1, len(rows))
	assert.Equal(t, startDate, rows[0].Date)
	assert.Equal(t, 8.0, rows[0].Hours) // 1 quantity * 8 multiplier
}

func TestExpandMultiDayAbsence_ThreeDays(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Three-day sick leave
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)   // Wednesday
	incidence := createTestIncidence(db, "faltas_extras", startDate, endDate, 3.0)

	// Load mapping
	var mapping models.IncidenceTipoMapping
	db.Where("incidence_type_id = ?", incidence.IncidenceTypeID).First(&mapping)

	// Expand
	rows := service.expandMultiDayAbsence(*incidence, mapping)

	assert.Equal(t, 3, len(rows), "Should expand to 3 separate rows")

	// Verify dates
	assert.Equal(t, startDate, rows[0].Date)
	assert.Equal(t, startDate.AddDate(0, 0, 1), rows[1].Date)
	assert.Equal(t, startDate.AddDate(0, 0, 2), rows[2].Date)

	// Verify hours are divided equally
	expectedHoursPerDay := (3.0 * 8.0) / 3.0 // (quantity * multiplier) / days
	assert.Equal(t, expectedHoursPerDay, rows[0].Hours)
	assert.Equal(t, expectedHoursPerDay, rows[1].Hours)
	assert.Equal(t, expectedHoursPerDay, rows[2].Hours)
}

func TestGenerateDualExport_EmptyPeriod(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Create empty payroll period
	company := &models.Company{Name: "Test Company"}
	db.Create(company)

	period := &models.PayrollPeriod{
		PeriodCode:   "2024-W02",
		PeriodType:   "weekly",
		Frequency:    "weekly",
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(0, 0, 7),
		PaymentDate:  time.Now().AddDate(0, 0, 10),
		Status:       "open",
		Year:         2024,
		PeriodNumber: 1,
	}
	db.Create(period)

	// Generate export
	zipBuffer, err := service.GenerateDualExport(period.ID)

	assert.NoError(t, err)
	assert.NotNil(t, zipBuffer)
	assert.Greater(t, zipBuffer.Len(), 0, "ZIP should be created even if empty")
}

func TestGetExportPreview_WithIncidences(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Create 2 vacation incidences
	startDate1 := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	endDate1 := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	inc1 := createTestIncidence(db, "vacaciones", startDate1, endDate1, 5.0)

	// Update to use same period as inc1
	var period models.PayrollPeriod
	db.First(&period, inc1.PayrollPeriodID)

	// Create absence incidence (3 days - should expand to 3 rows)
	startDate2 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)

	// Create new incidence type for absence
	employee := inc1.Employee
	var category models.IncidenceCategory
	db.First(&category)

	absenceType := &models.IncidenceType{
		CategoryID:  &category.ID,
		Name:        "Sick Leave",
		Category:    "sick",
		EffectType:  "negative",
		Description: "Sick leave absence",
	}
	db.Create(absenceType)

	// Create tipo mapping for absence (faltas_extras)
	tipoCode2 := "5"
	motivo2 := "FALTA"
	mapping2 := &models.IncidenceTipoMapping{
		IncidenceTypeID: absenceType.ID,
		TipoCode:        &tipoCode2,
		Motivo:          &motivo2,
		TemplateType:    "faltas_extras",
		HoursMultiplier: 8.0,
	}
	db.Create(mapping2)

	// Create second incidence
	inc2 := &models.Incidence{
		EmployeeID:          employee.ID,
		IncidenceTypeID:     absenceType.ID,
		PayrollPeriodID:     period.ID,
		StartDate:           startDate2,
		EndDate:             endDate2,
		Quantity:            3.0,
		Status:              "approved",
		Comments:            "Test absence",
		LateApprovalFlag:    false,
		ExcludedFromPayroll: false,
	}
	db.Create(inc2)

	// Get preview
	preview, err := service.GetExportPreview(period.ID)

	assert.NoError(t, err)
	assert.NotNil(t, preview)

	// Check counts
	vacacionesCount, ok := preview["vacaciones_count"].(int)
	assert.True(t, ok, "vacaciones_count should be an int")
	assert.Equal(t, 1, vacacionesCount, "Should have 1 vacation incidence")

	faltasExtrasCount, ok := preview["faltas_extras_count"].(int)
	assert.True(t, ok, "faltas_extras_count should be an int")
	assert.Equal(t, 3, faltasExtrasCount, "Should have 3 rows (3-day absence expanded)")

	totalIncidences, ok := preview["total_incidences"].(int)
	assert.True(t, ok, "total_incidences should be an int")
	assert.Equal(t, 2, totalIncidences, "Should have 2 total incidences")
}

func TestGenerateFaltasExtrasExcel_LateApprovalFlag(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Create incidence with late approval flag
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	incidence := createTestIncidence(db, "faltas_extras", startDate, endDate, 1.0)

	// Set late approval flag
	incidence.LateApprovalFlag = true
	db.Save(incidence)

	incidences := []models.Incidence{*incidence}

	// Generate Excel
	excelBytes, err := service.GenerateFaltasExtrasExcel(incidences)

	assert.NoError(t, err)
	assert.NotNil(t, excelBytes)
	assert.Greater(t, len(excelBytes), 0, "Excel file should have content")

	// Note: Full validation of "⚠️ APROBACIÓN TARDÍA" text in Observaciones column
	// would require parsing the Excel file, which is out of scope for this test
}

func TestGenerateDualExport_ExcludesRejectedIncidences(t *testing.T) {
	db := setupExcelExportTestDB(t)
	service := NewExcelExportService(db)

	// Create approved incidence
	startDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	incApproved := createTestIncidence(db, "vacaciones", startDate, endDate, 5.0)

	// Create rejected incidence (excluded_from_payroll = true)
	startDate2 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)
	incRejected := createTestIncidence(db, "vacaciones", startDate2, endDate2, 5.0)
	incRejected.ExcludedFromPayroll = true
	incRejected.Status = "rejected"
	db.Save(incRejected)

	// Generate export
	zipBuffer, err := service.GenerateDualExport(incApproved.PayrollPeriodID)

	assert.NoError(t, err)
	assert.NotNil(t, zipBuffer)

	// Get preview to verify only approved incidence is included
	preview, err := service.GetExportPreview(incApproved.PayrollPeriodID)
	assert.NoError(t, err)

	totalIncidences, ok := preview["total_incidences"].(int)
	assert.True(t, ok)
	assert.Equal(t, 1, totalIncidences, "Should only include 1 approved incidence, not the rejected one")
}
