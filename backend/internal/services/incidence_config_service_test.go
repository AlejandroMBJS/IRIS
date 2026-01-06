package services

import (
	"backend/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupIncidenceConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate test tables
	err = db.AutoMigrate(
		&models.IncidenceCategory{},
		&models.IncidenceType{},
		&models.IncidenceTipoMapping{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestIncidenceType(db *gorm.DB, name string) *models.IncidenceType {
	category := &models.IncidenceCategory{
		Name:     "Test Category",
		Code:     "TEST",
		IsActive: true,
	}
	db.Create(category)

	incidenceType := &models.IncidenceType{
		CategoryID:  &category.ID,
		Name:        name,
		Category:    "vacation",
		EffectType:  "positive",
		Description: "Test type",
	}
	db.Create(incidenceType)

	return incidenceType
}

func TestCreateTipoMapping_Success(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	motivo := "EXTRAS"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		Motivo:          &motivo,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}

	err := service.CreateTipoMapping(mapping)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, mapping.ID)
}

func TestCreateTipoMapping_DuplicateIncidenceType(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	// Create first mapping
	tipoCode := "1"
	motivo := "EXTRAS"
	mapping1 := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		Motivo:          &motivo,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	err := service.CreateTipoMapping(mapping1)
	assert.NoError(t, err)

	// Try to create second mapping for same incidence type
	mapping2 := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		Motivo:          &motivo,
		TemplateType:    "faltas_extras",
		HoursMultiplier: 8.0,
	}
	err = service.CreateTipoMapping(mapping2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCreateTipoMapping_InvalidTemplateType(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		TemplateType:    "invalid_template",
		HoursMultiplier: 8.0,
	}

	err := service.CreateTipoMapping(mapping)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid template type")
}

func TestCreateTipoMapping_InvalidMotivo(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	invalidMotivo := "INVALID"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		Motivo:          &invalidMotivo,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}

	err := service.CreateTipoMapping(mapping)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid motivo")
}

func TestCreateTipoMapping_DefaultHoursMultiplier(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		TemplateType:    "vacaciones",
		// HoursMultiplier not set
	}

	err := service.CreateTipoMapping(mapping)

	assert.NoError(t, err)
	assert.Equal(t, 8.0, mapping.HoursMultiplier)
}

func TestGetAllTipoMappings(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	// Create two mappings
	type1 := createTestIncidenceType(db, "Vacation")
	type2 := createTestIncidenceType(db, "Overtime")

	tipoCode1 := "1"
	motivo1 := "EXTRAS"
	mapping1 := &models.IncidenceTipoMapping{
		IncidenceTypeID: type1.ID,
		TipoCode:        &tipoCode1,
		Motivo:          &motivo1,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	service.CreateTipoMapping(mapping1)

	tipoCode2 := "7"
	motivo2 := "EXTRAS"
	mapping2 := &models.IncidenceTipoMapping{
		IncidenceTypeID: type2.ID,
		TipoCode:        &tipoCode2,
		Motivo:          &motivo2,
		TemplateType:    "faltas_extras",
		HoursMultiplier: 1.0,
	}
	service.CreateTipoMapping(mapping2)

	// Get all mappings
	mappings, err := service.GetAllTipoMappings()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(mappings))
}

func TestGetTipoMappingByID(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	motivo := "EXTRAS"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		Motivo:          &motivo,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	service.CreateTipoMapping(mapping)

	// Get by ID
	retrieved, err := service.GetTipoMappingByID(mapping.ID)

	assert.NoError(t, err)
	assert.Equal(t, mapping.ID, retrieved.ID)
	assert.Equal(t, "vacaciones", retrieved.TemplateType)
	assert.NotNil(t, retrieved.IncidenceType)
}

func TestGetTipoMappingByIncidenceType(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	service.CreateTipoMapping(mapping)

	// Get by incidence type
	retrieved, err := service.GetTipoMappingByIncidenceType(incidenceType.ID)

	assert.NoError(t, err)
	assert.Equal(t, incidenceType.ID, retrieved.IncidenceTypeID)
}

func TestUpdateTipoMapping(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	service.CreateTipoMapping(mapping)

	// Update
	newTipoCode := "2"
	newMotivo := "FALTA"
	updates := &models.IncidenceTipoMapping{
		TipoCode:        &newTipoCode,
		Motivo:          &newMotivo,
		TemplateType:    "faltas_extras",
		HoursMultiplier: 10.0,
	}

	err := service.UpdateTipoMapping(mapping.ID, updates)

	assert.NoError(t, err)

	// Verify update
	updated, _ := service.GetTipoMappingByID(mapping.ID)
	assert.Equal(t, "2", *updated.TipoCode)
	assert.Equal(t, "FALTA", *updated.Motivo)
	assert.Equal(t, "faltas_extras", updated.TemplateType)
	assert.Equal(t, 10.0, updated.HoursMultiplier)
}

func TestDeleteTipoMapping(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	incidenceType := createTestIncidenceType(db, "Vacation")

	tipoCode := "1"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: incidenceType.ID,
		TipoCode:        &tipoCode,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	service.CreateTipoMapping(mapping)

	// Delete
	err := service.DeleteTipoMapping(mapping.ID)

	assert.NoError(t, err)

	// Verify deleted
	_, err = service.GetTipoMappingByID(mapping.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetUnmappedIncidenceTypes(t *testing.T) {
	db := setupIncidenceConfigTestDB(t)
	service := NewIncidenceConfigService(db)

	// Create 3 incidence types
	type1 := createTestIncidenceType(db, "Vacation")
	createTestIncidenceType(db, "Overtime")
	createTestIncidenceType(db, "Sick Leave")

	// Create mapping for only one
	tipoCode := "1"
	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: type1.ID,
		TipoCode:        &tipoCode,
		TemplateType:    "vacaciones",
		HoursMultiplier: 8.0,
	}
	service.CreateTipoMapping(mapping)

	// Get unmapped
	unmapped, err := service.GetUnmappedIncidenceTypes()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(unmapped), "Should have 2 unmapped types")
}
