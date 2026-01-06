/*
Package dtos - IRIS Payroll System Data Transfer Objects

==============================================================================
FILE: internal/dtos/calendar.go
==============================================================================

DESCRIPTION:
    DTOs for the HR Calendar feature. These structures define the API contract
    for calendar events that aggregate data from AbsenceRequests, Incidences,
    and ShiftExceptions into a unified calendar view.

USER PERSPECTIVE:
    - HR users can view all employee events on a calendar
    - Filter by collar type, employee, event type, date range
    - Each event shows employee info and event details
    - Color-coded by employee for easy identification

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new fields, new event types
    ⚠️  CAUTION: Changing field names (breaks frontend)
    ❌  DO NOT modify: Existing enum values
    📝  Keep in sync with frontend TypeScript interfaces

==============================================================================
*/
package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CalendarEventType categorizes calendar events
type CalendarEventType string

const (
	EventTypeAbsence     CalendarEventType = "absence"
	EventTypeIncidence   CalendarEventType = "incidence"
	EventTypeShiftChange CalendarEventType = "shift_change"
)

// CalendarEventRequest defines query parameters for calendar events
type CalendarEventRequest struct {
	StartDate    string   `form:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate      string   `form:"end_date" binding:"required"`   // YYYY-MM-DD
	EmployeeIDs  []string `form:"employee_ids[]"`                // Filter by employee UUIDs
	CollarTypes  []string `form:"collar_types[]"`                // white_collar, blue_collar, gray_collar
	EventTypes   []string `form:"event_types[]"`                 // absence, incidence, shift_change
	DepartmentID string   `form:"department_id"`                 // Filter by department UUID
	Status       string   `form:"status"`                        // pending, approved, declined, etc.
}

// CalendarEventResponse represents a unified calendar event from any source
type CalendarEventResponse struct {
	ID          uuid.UUID         `json:"id"`
	EventType   CalendarEventType `json:"event_type"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date"`
	AllDay      bool              `json:"all_day"`
	Status      string            `json:"status"`

	// Employee information
	EmployeeID     uuid.UUID  `json:"employee_id"`
	EmployeeName   string     `json:"employee_name"`
	EmployeeNumber string     `json:"employee_number"`
	CollarType     string     `json:"collar_type"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`

	// Source tracking
	SourceID   uuid.UUID `json:"source_id"`   // ID of AbsenceRequest/Incidence/ShiftException
	SourceType string    `json:"source_type"` // absence_request, incidence, shift_exception

	// AbsenceRequest-specific fields
	RequestType   string  `json:"request_type,omitempty"`
	TotalDays     float64 `json:"total_days,omitempty"`
	ApprovalStage string  `json:"approval_stage,omitempty"`
	Reason        string  `json:"reason,omitempty"`

	// Incidence-specific fields
	Category         string  `json:"category,omitempty"`
	EffectType       string  `json:"effect_type,omitempty"` // positive, negative, neutral
	Quantity         float64 `json:"quantity,omitempty"`
	CalculatedAmount float64 `json:"calculated_amount,omitempty"`
	IncidenceType    string  `json:"incidence_type,omitempty"`

	// ShiftChange-specific fields
	ShiftName     string `json:"shift_name,omitempty"`
	ShiftCode     string `json:"shift_code,omitempty"`
	ShiftTime     string `json:"shift_time,omitempty"` // e.g., "07:00 - 15:00"
	OriginalShift string `json:"original_shift,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CalendarSummary provides aggregated statistics for the date range
type CalendarSummary struct {
	TotalAbsences     int `json:"total_absences"`
	TotalIncidences   int `json:"total_incidences"`
	TotalShiftChanges int `json:"total_shift_changes"`

	// By status
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Declined int `json:"declined"`

	// By collar type
	WhiteCollar int `json:"white_collar"`
	BlueCollar  int `json:"blue_collar"`
	GrayCollar  int `json:"gray_collar"`
}

// CalendarEventsResponse wraps the list of events with metadata
type CalendarEventsResponse struct {
	Events     []CalendarEventResponse `json:"events"`
	TotalCount int                     `json:"total_count"`
	Summary    CalendarSummary         `json:"summary"`
}

// CalendarEmployee represents an employee with their assigned calendar color
type CalendarEmployee struct {
	ID             uuid.UUID  `json:"id"`
	EmployeeNumber string     `json:"employee_number"`
	FullName       string     `json:"full_name"`
	CollarType     string     `json:"collar_type"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`
	DepartmentName string     `json:"department_name,omitempty"`
	Color          string     `json:"color"` // Hex color code auto-assigned
}

// CalendarEmployeesResponse wraps the list of employees
type CalendarEmployeesResponse struct {
	Employees []CalendarEmployee `json:"employees"`
}

// Predefined color palette for employee color assignment
// These colors are visually distinct and work well on both light and dark themes
var CalendarColorPalette = []string{
	"#6366f1", // Indigo
	"#8b5cf6", // Violet
	"#ec4899", // Pink
	"#f43f5e", // Rose
	"#ef4444", // Red
	"#f97316", // Orange
	"#f59e0b", // Amber
	"#eab308", // Yellow
	"#84cc16", // Lime
	"#22c55e", // Green
	"#10b981", // Emerald
	"#14b8a6", // Teal
	"#06b6d4", // Cyan
	"#0ea5e9", // Sky
	"#3b82f6", // Blue
	"#a855f7", // Purple
	"#d946ef", // Fuchsia
	"#64748b", // Slate
	"#78716c", // Stone
	"#71717a", // Zinc
}

// GetColorForIndex returns a color from the palette based on index
// Cycles through the palette if index exceeds length
func GetColorForIndex(index int) string {
	return CalendarColorPalette[index%len(CalendarColorPalette)]
}

// Request type labels in Spanish for display
var RequestTypeLabels = map[string]string{
	"PAID_LEAVE":    "Permiso con Goce",
	"UNPAID_LEAVE":  "Permiso sin Goce",
	"VACATION":      "Vacaciones",
	"LATE_ENTRY":    "Pase de Entrada",
	"EARLY_EXIT":    "Pase de Salida",
	"SHIFT_CHANGE":  "Cambio de Turno",
	"TIME_FOR_TIME": "Tiempo por Tiempo",
	"SICK_LEAVE":    "Incapacidad",
	"PERSONAL":      "Personal",
	"OTHER":         "Otro",
}

// GetRequestTypeLabel returns the Spanish label for a request type
func GetRequestTypeLabel(requestType string) string {
	if label, ok := RequestTypeLabels[requestType]; ok {
		return label
	}
	return requestType
}

// Incidence category labels in Spanish
var IncidenceCategoryLabels = map[string]string{
	"absence":   "Ausencia",
	"sick":      "Enfermedad",
	"vacation":  "Vacaciones",
	"overtime":  "Tiempo Extra",
	"delay":     "Retardo",
	"bonus":     "Bono",
	"deduction": "Deduccion",
	"other":     "Otro",
}

// GetIncidenceCategoryLabel returns the Spanish label for a category
func GetIncidenceCategoryLabel(category string) string {
	if label, ok := IncidenceCategoryLabels[category]; ok {
		return label
	}
	return category
}
