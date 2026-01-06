package models

import (
	"github.com/google/uuid"
)

// IncidenceTipoMapping maps incidence types to payroll export templates and Tipo codes
// This enables the dual Excel export system (Vacaciones vs Faltas y Extras)
type IncidenceTipoMapping struct {
	BaseModel
	IncidenceTypeID uuid.UUID `gorm:"type:text;not null;uniqueIndex" json:"incidence_type_id"`
	TipoCode        *string   `gorm:"type:varchar(10)" json:"tipo_code"` // Nullable for vacaciones
	Motivo          *string   `gorm:"type:varchar(50);check:motivo IN ('EXTRAS', 'FALTA', 'PERHORAS')" json:"motivo"`
	TemplateType    string    `gorm:"type:varchar(20);not null;check:template_type IN ('vacaciones', 'faltas_extras')" json:"template_type"`
	HoursMultiplier float64   `gorm:"type:decimal(5,2);default:8.0" json:"hours_multiplier"` // For days → hours conversion
	Notes           *string   `gorm:"type:text" json:"notes"`

	// Relationships
	IncidenceType IncidenceType `gorm:"foreignKey:IncidenceTypeID;constraint:OnDelete:CASCADE" json:"incidence_type,omitempty"`
}

// TableName specifies the table name for GORM
func (IncidenceTipoMapping) TableName() string {
	return "incidence_tipo_mappings"
}

// IsVacaciones returns true if this mapping is for the Vacaciones template
func (m *IncidenceTipoMapping) IsVacaciones() bool {
	return m.TemplateType == "vacaciones"
}

// IsFaltasExtras returns true if this mapping is for the Faltas y Extras template
func (m *IncidenceTipoMapping) IsFaltasExtras() bool {
	return m.TemplateType == "faltas_extras"
}

// GetTipoCode returns the Tipo code or empty string if null
func (m *IncidenceTipoMapping) GetTipoCode() string {
	if m.TipoCode == nil {
		return ""
	}
	return *m.TipoCode
}

// GetMotivo returns the Motivo or empty string if null
func (m *IncidenceTipoMapping) GetMotivo() string {
	if m.Motivo == nil {
		return ""
	}
	return *m.Motivo
}

// GetNotes returns the notes or empty string if null
func (m *IncidenceTipoMapping) GetNotes() string {
	if m.Notes == nil {
		return ""
	}
	return *m.Notes
}
