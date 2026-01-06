/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/incidence.go
==============================================================================

DESCRIPTION:
    Defines incidence-related models. Incidences are events that affect payroll:
    absences, overtime, bonuses, vacations, delays, etc. They are recorded
    before payroll calculation and impact the final pay.

USER PERSPECTIVE:
    - Incidences appear in the "Incidencias" section of the frontend
    - Users can create, approve, and track incidences with evidence
    - Incidence types define what kind of event occurred
    - Each incidence affects payroll positively, negatively, or neutrally

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new incidence categories, new fields
    ⚠️  CAUTION: Status transitions, calculation methods
    ❌  DO NOT modify: Existing category values (breaks database constraints)
    📝  When adding new types: Update incidence_types seed data

SYNTAX EXPLANATION:
    - Status workflow: pending → approved → processed (or rejected)
    - EffectType: 'positive' (adds to pay), 'negative' (deducts), 'neutral'
    - CalculationMethod: How quantity converts to money
        * daily_rate: quantity × daily salary
        * hourly_rate: quantity × hourly rate
        * fixed_amount: use DefaultValue directly
        * percentage: quantity as percentage of base

CATEGORIES:
    - absence: Unexcused absence (negative)
    - sick: Sick leave (may be paid/unpaid)
    - vacation: Vacation days (positive - includes vacation premium)
    - overtime: Extra hours worked (positive)
    - delay: Late arrival (negative)
    - bonus: One-time bonus (positive)
    - deduction: One-time deduction (negative)
    - other: Miscellaneous

EVIDENCE:
    - IncidenceEvidence stores file attachments (medical notes, etc.)
    - Files stored in UPLOAD_PATH, metadata in database

==============================================================================
*/
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Incidence represents an employee incidence (e.g., absence, bonus, overtime).
// Incidences affect payroll calculations based on their type and quantity.
type Incidence struct {
	BaseModel
	EmployeeID       uuid.UUID  `gorm:"type:text;not null" json:"employee_id"`
	PayrollPeriodID  uuid.UUID  `gorm:"type:text;not null" json:"payroll_period_id"`
	IncidenceTypeID  uuid.UUID  `gorm:"type:text;not null" json:"incidence_type_id"`
	StartDate        time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate          time.Time  `gorm:"type:date;not null" json:"end_date"`
	Quantity         float64    `gorm:"type:decimal(8,2);not null" json:"quantity"` // e.g., days, hours, amount
	CalculatedAmount float64    `gorm:"type:decimal(15,2)" json:"calculated_amount"`
	Comments         string     `gorm:"type:text" json:"comments,omitempty"`
	Status           string     `gorm:"type:varchar(50);default:'pending';check:status IN ('pending','approved','rejected','processed')" json:"status"`
	ApprovedBy       *uuid.UUID `gorm:"type:text" json:"approved_by,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`

	// Link back to the original absence request (if created from one)
	AbsenceRequestID *uuid.UUID `gorm:"type:text" json:"absence_request_id,omitempty"`

	// Payroll Export Fields (NEW for dual Excel export system)
	LateApprovalFlag     bool       `gorm:"default:false;index:idx_incidence_export,priority:1" json:"late_approval_flag"`     // Approved after payroll cutoff
	ExcludedFromPayroll  bool       `gorm:"default:false;index:idx_incidence_export,priority:2" json:"excluded_from_payroll"` // HR/GM rejected, exclude from export
	FinalApprovalAt      *time.Time `gorm:"type:timestamp" json:"final_approval_at,omitempty"`                                  // Timestamp of final approval

	// Relations
	Employee       *Employee      `gorm:"foreignKey:EmployeeID;constraint:OnDelete:RESTRICT" json:"employee,omitempty"`
	PayrollPeriod  *PayrollPeriod `gorm:"foreignKey:PayrollPeriodID;constraint:OnDelete:RESTRICT" json:"payroll_period,omitempty"`
	IncidenceType  *IncidenceType `gorm:"foreignKey:IncidenceTypeID;constraint:OnDelete:RESTRICT" json:"incidence_type,omitempty"`
	ApprovedByUser *User          `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL" json:"approved_by_user,omitempty"`
}

// TableName specifies the table name
func (Incidence) TableName() string {
	return "incidences"
}

// FormFieldOption represents an option for select/multiselect field types
type FormFieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FormField defines a custom form field for an incidence type
type FormField struct {
	Name         string            `json:"name"`                    // Field identifier
	Type         string            `json:"type"`                    // text, textarea, number, date, time, boolean, select, multiselect, shift_select
	Label        string            `json:"label"`                   // Display label (Spanish)
	LabelEn      string            `json:"label_en,omitempty"`      // Display label (English)
	Required     bool              `json:"required"`                // Is this field required?
	Min          *float64          `json:"min,omitempty"`           // Minimum value (for number)
	Max          *float64          `json:"max,omitempty"`           // Maximum value (for number)
	Step         *float64          `json:"step,omitempty"`          // Step value (for number)
	Placeholder  string            `json:"placeholder,omitempty"`   // Placeholder text
	DefaultValue interface{}       `json:"default_value,omitempty"` // Default value
	Options      []FormFieldOption `json:"options,omitempty"`       // Options for select/multiselect
	DisplayOrder int               `json:"display_order"`           // Order in form
	HelpText     string            `json:"help_text,omitempty"`     // Help text shown below field
}

// FormFieldsConfig is the JSON structure stored in form_fields column
type FormFieldsConfig struct {
	Fields []FormField `json:"fields"`
}

// IncidenceType defines the type of incidence (e.g., "absence", "overtime", "bonus").
type IncidenceType struct {
	BaseModel
	Name              string         `gorm:"type:varchar(255);not null" json:"name"`
	CategoryID        *uuid.UUID     `gorm:"type:text" json:"category_id,omitempty"`                                                           // FK to incidence_categories (new)
	Category          string         `gorm:"type:varchar(50);not null;check:category IN ('absence','sick','vacation','overtime','delay','bonus','deduction','other')" json:"category"` // Legacy field for backward compatibility
	EffectType        string         `gorm:"type:varchar(50);not null;check:effect_type IN ('positive','negative','neutral')" json:"effect_type"` // Affects income positively, negatively, or not at all
	IsCalculated      bool           `gorm:"default:true" json:"is_calculated"`                                                                   // If true, system calculates amount based on quantity
	CalculationMethod string         `gorm:"type:varchar(50)" json:"calculation_method,omitempty"`                                                // e.g., "daily_rate", "hourly_rate", "fixed_amount", "percentage"
	DefaultValue      float64        `gorm:"type:decimal(15,2)" json:"default_value"`
	RequiresEvidence  bool           `gorm:"default:false" json:"requires_evidence"`                                                              // If true, evidence file upload is mandatory
	Description       string         `gorm:"type:text" json:"description,omitempty"`
	FormFields        datatypes.JSON `gorm:"type:jsonb" json:"form_fields,omitempty"`                                                             // Custom form field definitions (new)
	IsRequestable     bool           `gorm:"default:false" json:"is_requestable"`                                                                 // Can employees request this type? (new)
	ApprovalFlow      string         `gorm:"type:varchar(50);default:'standard'" json:"approval_flow"`                                            // Approval workflow type (new)
	DisplayOrder      int            `gorm:"default:0" json:"display_order"`                                                                      // Order in UI (new)

	// Relations
	IncidenceCategory *IncidenceCategory `gorm:"foreignKey:CategoryID" json:"incidence_category,omitempty"`
}

// TableName specifies the table name
func (IncidenceType) TableName() string {
	return "incidence_types"
}

// IncidenceEvidence represents a file attachment for incidence evidence
type IncidenceEvidence struct {
	BaseModel
	IncidenceID  uuid.UUID `gorm:"type:text;not null" json:"incidence_id"`
	FileName     string    `gorm:"type:varchar(255);not null" json:"file_name"`
	OriginalName string    `gorm:"type:varchar(255);not null" json:"original_name"`
	ContentType  string    `gorm:"type:varchar(100);not null" json:"content_type"`
	FileSize     int64     `gorm:"not null" json:"file_size"`
	FilePath     string    `gorm:"type:varchar(500);not null" json:"file_path"`
	UploadedBy   uuid.UUID `gorm:"type:text;not null" json:"uploaded_by"`

	// Relations
	Incidence      *Incidence `gorm:"foreignKey:IncidenceID" json:"incidence,omitempty"`
	UploadedByUser *User      `gorm:"foreignKey:UploadedBy" json:"uploaded_by_user,omitempty"`
}

// TableName specifies the table name
func (IncidenceEvidence) TableName() string {
	return "incidence_evidences"
}

// BeforeCreate hook for IncidenceType
func (it *IncidenceType) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate UUID if not set (call base model logic)
	if it.ID == uuid.Nil {
		it.ID = uuid.New()
	}
	if it.Name == "" {
		return ErrNameRequired
	}
	if it.Category == "" {
		return ErrCategoryRequired
	}
	if it.EffectType == "" {
		return ErrEffectTypeRequired
	}
	return
}

// GetFormFieldsConfig parses the FormFields JSON into a FormFieldsConfig struct
func (it *IncidenceType) GetFormFieldsConfig() (*FormFieldsConfig, error) {
	if it.FormFields == nil || len(it.FormFields) == 0 {
		return &FormFieldsConfig{Fields: []FormField{}}, nil
	}

	var config FormFieldsConfig
	if err := json.Unmarshal(it.FormFields, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetFormFieldsConfig converts a FormFieldsConfig to JSON and stores it
func (it *IncidenceType) SetFormFieldsConfig(config *FormFieldsConfig) error {
	if config == nil {
		it.FormFields = nil
		return nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	it.FormFields = data
	return nil
}
