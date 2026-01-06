/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/incidence_config_service.go
==============================================================================

DESCRIPTION:
    Manages incidence tipo mappings configuration for payroll export.
    Provides CRUD operations for mapping incidence types to payroll
    template types, Tipo codes, and Motivo values.

USER PERSPECTIVE:
    - Configure which Excel template each incidence type maps to
    - Set Tipo codes (2-9) for payroll system import
    - Define Motivo (EXTRAS, FALTA, PERHORAS) for each incidence type
    - Configure hours multipliers for day-to-hour conversion

DEVELOPER GUIDELINES:
    ✅  OK to modify: Validation rules, default values
    ⚠️  CAUTION: Template type validation (must match Excel export logic)
    ❌  DO NOT modify: Database constraints without migration
    📝  Tipo codes must match payroll system expectations

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IncidenceConfigService manages incidence tipo mappings
type IncidenceConfigService struct {
	db *gorm.DB
}

// NewIncidenceConfigService creates a new incidence config service
func NewIncidenceConfigService(db *gorm.DB) *IncidenceConfigService {
	return &IncidenceConfigService{db: db}
}

// GetAllTipoMappings retrieves all tipo mappings with incidence type details
func (s *IncidenceConfigService) GetAllTipoMappings() ([]models.IncidenceTipoMapping, error) {
	var mappings []models.IncidenceTipoMapping
	err := s.db.
		Preload("IncidenceType").
		Preload("IncidenceType.IncidenceCategory").
		Find(&mappings).Error

	return mappings, err
}

// GetTipoMappingByID retrieves a single tipo mapping by ID
func (s *IncidenceConfigService) GetTipoMappingByID(id uuid.UUID) (*models.IncidenceTipoMapping, error) {
	var mapping models.IncidenceTipoMapping
	err := s.db.
		Preload("IncidenceType").
		Preload("IncidenceType.IncidenceCategory").
		First(&mapping, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tipo mapping not found")
		}
		return nil, fmt.Errorf("failed to get tipo mapping: %w", err)
	}

	return &mapping, nil
}

// GetTipoMappingByIncidenceType retrieves tipo mapping for a specific incidence type
func (s *IncidenceConfigService) GetTipoMappingByIncidenceType(incidenceTypeID uuid.UUID) (*models.IncidenceTipoMapping, error) {
	var mapping models.IncidenceTipoMapping
	err := s.db.
		Preload("IncidenceType").
		Where("incidence_type_id = ?", incidenceTypeID).
		First(&mapping).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tipo mapping not found for incidence type")
		}
		return nil, fmt.Errorf("failed to get tipo mapping: %w", err)
	}

	return &mapping, nil
}

// CreateTipoMapping creates a new tipo mapping
func (s *IncidenceConfigService) CreateTipoMapping(mapping *models.IncidenceTipoMapping) error {
	// Validate incidence type exists
	var incidenceType models.IncidenceType
	err := s.db.First(&incidenceType, "id = ?", mapping.IncidenceTypeID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("incidence type not found")
		}
		return fmt.Errorf("failed to validate incidence type: %w", err)
	}

	// Check if mapping already exists for this incidence type
	var existingMapping models.IncidenceTipoMapping
	err = s.db.Where("incidence_type_id = ?", mapping.IncidenceTypeID).First(&existingMapping).Error
	if err == nil {
		return fmt.Errorf("tipo mapping already exists for this incidence type")
	}

	// Validate template type
	if mapping.TemplateType != "vacaciones" && mapping.TemplateType != "faltas_extras" {
		return fmt.Errorf("invalid template type: must be 'vacaciones' or 'faltas_extras'")
	}

	// Validate Motivo if provided
	if mapping.Motivo != nil {
		motivo := *mapping.Motivo
		if motivo != "EXTRAS" && motivo != "FALTA" && motivo != "PERHORAS" {
			return fmt.Errorf("invalid motivo: must be 'EXTRAS', 'FALTA', or 'PERHORAS'")
		}
	}

	// Set default hours multiplier if not provided
	if mapping.HoursMultiplier == 0 {
		mapping.HoursMultiplier = 8.0
	}

	return s.db.Create(mapping).Error
}

// UpdateTipoMapping updates an existing tipo mapping
func (s *IncidenceConfigService) UpdateTipoMapping(id uuid.UUID, updates *models.IncidenceTipoMapping) error {
	// Check if mapping exists
	var existing models.IncidenceTipoMapping
	err := s.db.First(&existing, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("tipo mapping not found")
		}
		return fmt.Errorf("failed to get tipo mapping: %w", err)
	}

	// Validate template type
	if updates.TemplateType != "" {
		if updates.TemplateType != "vacaciones" && updates.TemplateType != "faltas_extras" {
			return fmt.Errorf("invalid template type: must be 'vacaciones' or 'faltas_extras'")
		}
		existing.TemplateType = updates.TemplateType
	}

	// Update fields
	if updates.TipoCode != nil {
		existing.TipoCode = updates.TipoCode
	}

	if updates.Motivo != nil {
		motivo := *updates.Motivo
		if motivo != "EXTRAS" && motivo != "FALTA" && motivo != "PERHORAS" && motivo != "" {
			return fmt.Errorf("invalid motivo: must be 'EXTRAS', 'FALTA', or 'PERHORAS'")
		}
		existing.Motivo = updates.Motivo
	}

	if updates.HoursMultiplier > 0 {
		existing.HoursMultiplier = updates.HoursMultiplier
	}

	if updates.Notes != nil {
		existing.Notes = updates.Notes
	}

	return s.db.Save(&existing).Error
}

// DeleteTipoMapping deletes a tipo mapping
func (s *IncidenceConfigService) DeleteTipoMapping(id uuid.UUID) error {
	result := s.db.Delete(&models.IncidenceTipoMapping{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete tipo mapping: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tipo mapping not found")
	}

	return nil
}

// GetUnmappedIncidenceTypes returns incidence types that don't have a tipo mapping
func (s *IncidenceConfigService) GetUnmappedIncidenceTypes() ([]models.IncidenceType, error) {
	var incidenceTypes []models.IncidenceType

	// Get all incidence types that don't have a mapping
	err := s.db.
		Preload("IncidenceCategory").
		Where("id NOT IN (SELECT incidence_type_id FROM incidence_tipo_mappings)").
		Find(&incidenceTypes).Error

	return incidenceTypes, err
}

// SeedDefaultMappings creates default tipo mappings for common incidence types
// This should be called during application setup
func (s *IncidenceConfigService) SeedDefaultMappings() error {
	// Check if any mappings already exist
	var count int64
	s.db.Model(&models.IncidenceTipoMapping{}).Count(&count)
	if count > 0 {
		// Mappings already exist, skip seeding
		return nil
	}

	// Get incidence types for default mappings
	var vacationType, overtimeType, absenceType models.IncidenceType

	// Try to find Vacation type
	err := s.db.Where("name LIKE ?", "%Vacation%").Or("name LIKE ?", "%Vacacion%").First(&vacationType).Error
	if err == nil {
		// Create Vacaciones mapping
		tipoCodeVac := ""
		motivoVac := ""
		vacMapping := &models.IncidenceTipoMapping{
			IncidenceTypeID: vacationType.ID,
			TipoCode:        &tipoCodeVac,
			Motivo:          &motivoVac,
			TemplateType:    "vacaciones",
			HoursMultiplier: 8.0,
		}
		s.db.Create(vacMapping)
	}

	// Try to find Overtime type
	err = s.db.Where("name LIKE ?", "%Overtime%").Or("name LIKE ?", "%Extra%").First(&overtimeType).Error
	if err == nil {
		// Create Overtime mapping
		tipoCodeOT := "7"
		motivoOT := "EXTRAS"
		otMapping := &models.IncidenceTipoMapping{
			IncidenceTypeID: overtimeType.ID,
			TipoCode:        &tipoCodeOT,
			Motivo:          &motivoOT,
			TemplateType:    "faltas_extras",
			HoursMultiplier: 1.0,
		}
		s.db.Create(otMapping)
	}

	// Try to find Absence type
	err = s.db.Where("name LIKE ?", "%Absence%").Or("name LIKE ?", "%Falta%").First(&absenceType).Error
	if err == nil {
		// Create Absence mapping
		tipoCodeAbs := "5"
		motivoAbs := "FALTA"
		absMapping := &models.IncidenceTipoMapping{
			IncidenceTypeID: absenceType.ID,
			TipoCode:        &tipoCodeAbs,
			Motivo:          &motivoAbs,
			TemplateType:    "faltas_extras",
			HoursMultiplier: 8.0,
		}
		s.db.Create(absMapping)
	}

	return nil
}
