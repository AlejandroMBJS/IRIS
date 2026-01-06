package database

import (
	"backend/internal/models"
	"log"

	"gorm.io/gorm"
)

// SeedTipoMappings seeds the incidence_tipo_mappings table with payroll export configurations
// This connects existing incidence types in the system to Tipo codes for Excel export
func SeedTipoMappings(db *gorm.DB) error {
	log.Println("Seeding Tipo mappings for payroll export...")

	// Define mappings by incidence CATEGORY (matches existing system categories)
	// Based on user's Tipo code table:
	// Tipo 2: Horas extras (AEXT)
	// Tipo 3: Permiso con goce (PERCG)
	// Tipo 4: Permiso sin goce (PERSG)
	// Tipo 5: Pago de Extras (EXTRAS)
	// Tipo 6: Pago de Ordinarias (HORAS)
	// Tipo 7: Falta (FALTA)
	// Tipo 8: Retardo (TARDE)
	// Tipo 9: Permuta (PERHRA)

	categoryMappings := []struct {
		Category        string  // Match by category field in incidence_types
		TemplateType    string  // "vacaciones" or "faltas_extras"
		TipoCode        *string // Payroll Tipo code
		Motivo          *string // EXTRAS, FALTA, PERHORAS
		HoursMultiplier float64
		Notes           string
	}{
		// VACACIONES TEMPLATE - Category: vacation
		{
			Category:        "vacation",
			TemplateType:    "vacaciones",
			TipoCode:        nil,
			Motivo:          nil,
			HoursMultiplier: 8.0,
			Notes:           "Vacaciones - exports to Vacaciones.xlsx with DiasPago and DiasPrima (25%)",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: overtime
		{
			Category:        "overtime",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("2"), // or "5" - using Tipo 2 (Horas extras AEXT) as default
			Motivo:          stringPtr("EXTRAS"),
			HoursMultiplier: 1.0,
			Notes:           "Horas extras / Overtime - Tipo 2 (AEXT) - hours entered directly",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: absence
		{
			Category:        "absence",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("7"),
			Motivo:          stringPtr("FALTA"),
			HoursMultiplier: 8.0,
			Notes:           "Falta / Unexcused absence - Tipo 7 (FALTA) - converts days to hours",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: sick
		{
			Category:        "sick",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("7"),
			Motivo:          stringPtr("FALTA"),
			HoursMultiplier: 8.0,
			Notes:           "Incapacidad / Sick leave - Tipo 7 (FALTA) - treated as absence",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: delay
		{
			Category:        "delay",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("8"),
			Motivo:          stringPtr("FALTA"),
			HoursMultiplier: 1.0,
			Notes:           "Retardo / Tardiness - Tipo 8 (TARDE) - hours/minutes entered directly",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: bonus
		{
			Category:        "bonus",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("5"),
			Motivo:          stringPtr("EXTRAS"),
			HoursMultiplier: 1.0,
			Notes:           "Bonus / Extra payment - Tipo 5 (Pago de Extras EXTRAS)",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: deduction
		{
			Category:        "deduction",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("7"),
			Motivo:          stringPtr("FALTA"),
			HoursMultiplier: 1.0,
			Notes:           "Deducción / Deduction - Tipo 7 (FALTA) - negative impact",
		},

		// FALTAS Y EXTRAS TEMPLATE - Category: other
		{
			Category:        "other",
			TemplateType:    "faltas_extras",
			TipoCode:        stringPtr("9"),
			Motivo:          stringPtr("PERHORAS"),
			HoursMultiplier: 1.0,
			Notes:           "Permuta/Other - Tipo 9 (PERHRA) - shift swaps and miscellaneous",
		},
	}

	// Process each category mapping
	for _, m := range categoryMappings {
		// Find ALL incidence types with this category
		var incidenceTypes []models.IncidenceType
		result := db.Where("category = ?", m.Category).Find(&incidenceTypes)

		if result.Error != nil {
			log.Printf("Error querying incidence types for category '%s': %v", m.Category, result.Error)
			continue
		}

		if len(incidenceTypes) == 0 {
			log.Printf("Warning: No incidence types found for category '%s' - skipping", m.Category)
			continue
		}

		// Create mapping for each incidence type in this category
		for _, incidenceType := range incidenceTypes {
			// Check if mapping already exists
			var existing models.IncidenceTipoMapping
			existResult := db.Where("incidence_type_id = ?", incidenceType.ID).First(&existing)

			mapping := models.IncidenceTipoMapping{
				IncidenceTypeID: incidenceType.ID,
				TipoCode:        m.TipoCode,
				Motivo:          m.Motivo,
				TemplateType:    m.TemplateType,
				HoursMultiplier: m.HoursMultiplier,
				Notes:           &m.Notes,
			}

			if existResult.Error == gorm.ErrRecordNotFound {
				// Create new mapping
				if err := db.Create(&mapping).Error; err != nil {
					log.Printf("Error creating mapping for '%s' (ID: %s): %v",
						incidenceType.Name, incidenceType.ID, err)
					continue
				}
				log.Printf("✓ Created mapping: %s (category: %s) → %s [Tipo: %s, Motivo: %s]",
					incidenceType.Name, m.Category, m.TemplateType,
					ptrToString(m.TipoCode), ptrToString(m.Motivo))
			} else {
				// Update existing mapping
				if err := db.Model(&existing).Updates(mapping).Error; err != nil {
					log.Printf("Error updating mapping for '%s' (ID: %s): %v",
						incidenceType.Name, incidenceType.ID, err)
					continue
				}
				log.Printf("✓ Updated mapping: %s (category: %s) → %s [Tipo: %s, Motivo: %s]",
					incidenceType.Name, m.Category, m.TemplateType,
					ptrToString(m.TipoCode), ptrToString(m.Motivo))
			}
		}
	}

	log.Println("Tipo mappings seeding completed")
	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to safely dereference string pointers for logging
func ptrToString(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
