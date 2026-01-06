/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/absence_request.go
==============================================================================

DESCRIPTION:
    Defines models for the absence request approval workflow. This system
    handles employee requests for time off (vacations, leaves, passes) and
    routes them through a multi-stage approval process.

USER PERSPECTIVE:
    - Employees create absence requests (vacation, leave, passes)
    - Requests flow through: SUPERVISOR → MANAGER → HR → COMPLETED
    - Each approver can approve or decline with comments
    - Notifications are sent at each stage

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new request types, approval stages
    ⚠️  CAUTION: Status transitions, approval flow logic
    ❌  DO NOT modify: Existing status/stage values (breaks workflows)
    📝  When adding types: Update validation and frontend

APPROVAL FLOW:
    - Normal employees: SUPERVISOR → MANAGER → HR (if blue collar) → COMPLETED
    - SUPANDGM supervisor: Auto-skip SUPERVISOR and MANAGER, start at HR
    - White collar: Skip HR stage after manager approval

REQUEST TYPES:
    - PAID_LEAVE: Permiso con goce de sueldo
    - UNPAID_LEAVE: Permiso sin goce de sueldo
    - VACATION: Vacaciones
    - LATE_ENTRY: Pase de entrada (llegada tarde)
    - EARLY_EXIT: Pase de salida (salida temprano)
    - SHIFT_CHANGE: Cambio de turno
    - TIME_FOR_TIME: Tiempo por tiempo
    - SICK_LEAVE: Incapacidad por enfermedad

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RequestType represents the type of absence request
type RequestType string

const (
	RequestTypePaidLeave    RequestType = "PAID_LEAVE"
	RequestTypeUnpaidLeave  RequestType = "UNPAID_LEAVE"
	RequestTypeVacation     RequestType = "VACATION"
	RequestTypeLateEntry    RequestType = "LATE_ENTRY"
	RequestTypeEarlyExit    RequestType = "EARLY_EXIT"
	RequestTypeShiftChange  RequestType = "SHIFT_CHANGE"
	RequestTypeTimeForTime  RequestType = "TIME_FOR_TIME"
	RequestTypeSickLeave    RequestType = "SICK_LEAVE"
	RequestTypePersonal     RequestType = "PERSONAL"
	RequestTypeOther        RequestType = "OTHER"
)

// RequestStatus represents the status of an absence request
type RequestStatus string

const (
	RequestStatusPending  RequestStatus = "PENDING"
	RequestStatusApproved RequestStatus = "APPROVED"
	RequestStatusDeclined RequestStatus = "DECLINED"
	RequestStatusArchived RequestStatus = "ARCHIVED"
)

// ApprovalStage represents the current stage in the approval workflow
type ApprovalStage string

const (
	ApprovalStageSupervisor     ApprovalStage = "SUPERVISOR"
	ApprovalStageManager        ApprovalStage = "MANAGER"           // Keep for backward compatibility
	ApprovalStageGeneralManager ApprovalStage = "GENERAL_MANAGER"  // NEW: Renamed from MANAGER
	ApprovalStageHR             ApprovalStage = "HR"                // Keep for backward compatibility
	ApprovalStageHRBlueGray     ApprovalStage = "HR_BLUE_GRAY"     // NEW: HR approval for blue/gray collar
	ApprovalStagePayroll        ApprovalStage = "PAYROLL"
	ApprovalStageCompleted      ApprovalStage = "COMPLETED"
)

// ApprovalAction represents the action taken on an approval
type ApprovalAction string

const (
	ApprovalActionApproved ApprovalAction = "APPROVED"
	ApprovalActionDeclined ApprovalAction = "DECLINED"
)

// AbsenceRequest represents an employee's request for time off
type AbsenceRequest struct {
	BaseModel
	EmployeeID            uuid.UUID     `gorm:"type:text;not null" json:"employee_id"`
	RequestType           RequestType   `gorm:"type:varchar(50);not null" json:"request_type"`              // Legacy field (keep for backward compatibility)
	IncidenceTypeID       *uuid.UUID    `gorm:"type:text" json:"incidence_type_id,omitempty"`                // New: Link to dynamic incidence type
	StartDate             time.Time     `gorm:"type:date;not null" json:"start_date"`
	EndDate               time.Time     `gorm:"type:date;not null" json:"end_date"`
	TotalDays             float64       `gorm:"type:decimal(5,2);not null" json:"total_days"`
	Reason                string        `gorm:"type:text;not null" json:"reason"`
	Status                RequestStatus `gorm:"type:varchar(50);default:'PENDING'" json:"status"`
	CurrentApprovalStage  ApprovalStage `gorm:"type:varchar(50);default:'SUPERVISOR'" json:"current_approval_stage"`
	CustomFields          datatypes.JSON `gorm:"type:jsonb" json:"custom_fields,omitempty"`                  // New: Stores dynamic form field values

	// Optional fields for specific request types (legacy - kept for backward compatibility)
	HoursPerDay           *float64      `gorm:"type:decimal(4,2)" json:"hours_per_day,omitempty"`
	PaidDays              *float64      `gorm:"type:decimal(5,2)" json:"paid_days,omitempty"`
	UnpaidDays            *float64      `gorm:"type:decimal(5,2)" json:"unpaid_days,omitempty"`
	UnpaidComments        string        `gorm:"type:text" json:"unpaid_comments,omitempty"`
	ShiftDetails          string        `gorm:"type:text" json:"shift_details,omitempty"`

	// Shift change specific fields
	NewShiftID            *uuid.UUID    `gorm:"type:text" json:"new_shift_id,omitempty"`
	NewShift              *Shift        `gorm:"foreignKey:NewShiftID" json:"new_shift,omitempty"`

	// Escalation and Payroll Export Fields (NEW for dual Excel export system)
	LastActionAt          time.Time     `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index" json:"last_action_at"`           // For 24-hour escalation tracking
	EscalationCount       int           `gorm:"default:0" json:"escalation_count"`                                               // Number of times escalated
	IsEscalated           bool          `gorm:"default:false;index:idx_absence_escalation,priority:1" json:"is_escalated"`      // Has this been auto-escalated?
	LateApprovalFlag      bool          `gorm:"default:false" json:"late_approval_flag"`                                         // Approved after payroll cutoff
	ExcludedFromPayroll   bool          `gorm:"default:false" json:"excluded_from_payroll"`                                      // HR/GM rejected, exclude from export
	PayrollCutoffDate     *time.Time    `gorm:"type:timestamp" json:"payroll_cutoff_date,omitempty"`                            // Calculated cutoff for this request

	// Relations
	Employee              *Employee         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	IncidenceType         *IncidenceType    `gorm:"foreignKey:IncidenceTypeID" json:"incidence_type,omitempty"` // New: Relationship to incidence type
	ApprovalHistory       []ApprovalHistory `gorm:"foreignKey:RequestID" json:"approval_history,omitempty"`
	EscalationLogs        []EscalationLog   `gorm:"foreignKey:AbsenceRequestID" json:"escalation_logs,omitempty"` // NEW: Escalation audit trail
}

// TableName specifies the table name
func (AbsenceRequest) TableName() string {
	return "absence_requests"
}

// BeforeCreate hook
func (ar *AbsenceRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if ar.ID == uuid.Nil {
		ar.ID = uuid.New()
	}
	if ar.Status == "" {
		ar.Status = RequestStatusPending
	}
	if ar.CurrentApprovalStage == "" {
		ar.CurrentApprovalStage = ApprovalStageSupervisor
	}
	return nil
}

// ApprovalHistory records each approval/decline action in the workflow
type ApprovalHistory struct {
	BaseModel
	RequestID     uuid.UUID      `gorm:"type:text;not null" json:"request_id"`
	ApproverID    uuid.UUID      `gorm:"type:text;not null" json:"approver_id"`
	ApprovalStage ApprovalStage  `gorm:"type:varchar(50);not null" json:"approval_stage"`
	Action        ApprovalAction `gorm:"type:varchar(50);not null" json:"action"`
	Comments      string         `gorm:"type:text" json:"comments,omitempty"`

	// Relations
	Request       *AbsenceRequest `gorm:"foreignKey:RequestID" json:"request,omitempty"`
	Approver      *User           `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}

// TableName specifies the table name
func (ApprovalHistory) TableName() string {
	return "approval_history"
}

// BeforeCreate hook
func (ah *ApprovalHistory) BeforeCreate(tx *gorm.DB) (err error) {
	if ah.ID == uuid.Nil {
		ah.ID = uuid.New()
	}
	return nil
}

// HRAssignment maps HR users to employee types they manage
type HRAssignment struct {
	BaseModel
	HRUserID     uuid.UUID `gorm:"type:text;not null" json:"hr_user_id"`
	EmployeeType string    `gorm:"type:varchar(50);not null" json:"employee_type"` // Sindicalizado, No sindicalizado

	// Relations
	HRUser       *User     `gorm:"foreignKey:HRUserID" json:"hr_user,omitempty"`
}

// TableName specifies the table name
func (HRAssignment) TableName() string {
	return "hr_assignments"
}

// BeforeCreate hook
func (ha *HRAssignment) BeforeCreate(tx *gorm.DB) (err error) {
	if ha.ID == uuid.Nil {
		ha.ID = uuid.New()
	}
	return nil
}

// Shift represents a work shift (turno)
type Shift struct {
	BaseModel
	// Basic Info
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Code        string `gorm:"type:varchar(50);not null" json:"code"`
	Description string `gorm:"type:varchar(255)" json:"description,omitempty"`

	// Schedule Times (24-hour format HH:MM)
	StartTime string `gorm:"type:varchar(10);not null" json:"start_time"` // e.g., "07:00"
	EndTime   string `gorm:"type:varchar(10);not null" json:"end_time"`   // e.g., "15:00"

	// Break Configuration
	BreakMinutes   int    `gorm:"default:30" json:"break_minutes"`
	BreakStartTime string `gorm:"type:varchar(10)" json:"break_start_time,omitempty"`

	// Work Hours
	WorkHoursPerDay float64 `gorm:"type:decimal(4,2);default:8.0" json:"work_hours_per_day"`

	// Work Days (stored as JSON array of day numbers: 0=Sun, 1=Mon, etc.)
	WorkDays string `gorm:"type:text;default:'[1,2,3,4,5]'" json:"work_days"`

	// Display
	Color        string `gorm:"type:varchar(7);default:'#3B82F6'" json:"color"`
	DisplayOrder int    `gorm:"default:0" json:"display_order"`

	// Status
	IsRestDay    bool `gorm:"default:false" json:"is_rest_day"`
	IsActive     bool `gorm:"default:true" json:"is_active"`
	IsNightShift bool `gorm:"default:false" json:"is_night_shift"`

	// Collar Types (JSON array: ["white_collar", "blue_collar", "gray_collar"])
	// Empty means available for all collar types
	CollarTypes string `gorm:"type:text;default:'[]'" json:"collar_types"`

	// Foreign Keys (for multi-tenant)
	CompanyID uuid.UUID `gorm:"type:text" json:"company_id,omitempty"`

	// Relations
	Company   *Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Employees []Employee `gorm:"foreignKey:ShiftID" json:"employees,omitempty"`
}

// TableName specifies the table name
func (Shift) TableName() string {
	return "shifts"
}

// BeforeCreate hook
func (s *Shift) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// EmployeeShiftBase represents the regular weekly schedule for an employee
type EmployeeShiftBase struct {
	BaseModel
	EmployeeID uuid.UUID `gorm:"type:text;not null" json:"employee_id"`
	ShiftID    uuid.UUID `gorm:"type:text;not null" json:"shift_id"`
	DayOfWeek  int       `gorm:"not null" json:"day_of_week"` // 0=Monday, 6=Sunday

	// Relations
	Employee   *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Shift      *Shift     `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
}

// TableName specifies the table name
func (EmployeeShiftBase) TableName() string {
	return "employee_shift_bases"
}

// ShiftException represents a one-time schedule override for a specific date
type ShiftException struct {
	BaseModel
	EmployeeID  uuid.UUID `gorm:"type:text;not null" json:"employee_id"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	ShiftID     uuid.UUID `gorm:"type:text;not null" json:"shift_id"`
	CreatedByID uuid.UUID `gorm:"type:text;not null" json:"created_by_id"`

	// Relations
	Employee   *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Shift      *Shift     `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
	CreatedBy  *User  `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

// TableName specifies the table name
func (ShiftException) TableName() string {
	return "shift_exceptions"
}

// Announcement represents a company announcement
type Announcement struct {
	BaseModel
	CreatedByID uuid.UUID  `gorm:"type:text;not null" json:"created_by_id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Message     string     `gorm:"type:text;not null" json:"message"`
	ImageData   []byte     `gorm:"type:blob" json:"image_data,omitempty"`
	Scope       string     `gorm:"type:varchar(50);not null" json:"scope"` // ALL, DEPARTMENT, etc.
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`

	// Relations
	CreatedBy   *User      `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

// TableName specifies the table name
func (Announcement) TableName() string {
	return "announcements"
}

// ReadAnnouncement tracks which users have read which announcements
type ReadAnnouncement struct {
	UserID         uuid.UUID `gorm:"type:text;primaryKey" json:"user_id"`
	AnnouncementID uuid.UUID `gorm:"type:text;primaryKey" json:"announcement_id"`
	ReadAt         time.Time `gorm:"autoCreateTime" json:"read_at"`

	// Relations
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Announcement *Announcement `gorm:"foreignKey:AnnouncementID" json:"announcement,omitempty"`
}

// TableName specifies the table name
func (ReadAnnouncement) TableName() string {
	return "read_announcements"
}
