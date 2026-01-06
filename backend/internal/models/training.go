/*
Package models - IRIS Payroll System Training/LMS Module Models

==============================================================================
FILE: internal/models/training.go
==============================================================================

DESCRIPTION:
    Data models for the Training and Learning Management System (LMS).
    Supports course creation, enrollments, progress tracking, and certifications.

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CourseStatus represents the status of a training course
type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusPublished CourseStatus = "published"
	CourseStatusArchived  CourseStatus = "archived"
)

// CourseType represents the type of training course
type CourseType string

const (
	CourseTypeOnline      CourseType = "online"
	CourseTypeInPerson    CourseType = "in_person"
	CourseTypeHybrid      CourseType = "hybrid"
	CourseTypeSelfPaced   CourseType = "self_paced"
	CourseTypeInstructor  CourseType = "instructor_led"
)

// ContentType represents the type of course content
type ContentType string

const (
	ContentTypeVideo       ContentType = "video"
	ContentTypeDocument    ContentType = "document"
	ContentTypeQuiz        ContentType = "quiz"
	ContentTypeAssignment  ContentType = "assignment"
	ContentTypeScorm       ContentType = "scorm"
	ContentTypeInteractive ContentType = "interactive"
	ContentTypeWebinar     ContentType = "webinar"
)

// EnrollmentStatus represents the status of a course enrollment
type EnrollmentStatus string

const (
	EnrollmentStatusPending    EnrollmentStatus = "pending"
	EnrollmentStatusActive     EnrollmentStatus = "active"
	EnrollmentStatusCompleted  EnrollmentStatus = "completed"
	EnrollmentStatusFailed     EnrollmentStatus = "failed"
	EnrollmentStatusDropped    EnrollmentStatus = "dropped"
	EnrollmentStatusExpired    EnrollmentStatus = "expired"
)

// TrainingCategory represents a category for organizing courses
type TrainingCategory struct {
	BaseModel
	CompanyID   uuid.UUID `gorm:"type:text;not null;index" json:"company_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:text;index" json:"parent_id,omitempty"`
	Color       string    `gorm:"size:7" json:"color"`
	Icon        string    `gorm:"size:50" json:"icon"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`

	// Relationships
	Parent     *TrainingCategory  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children   []TrainingCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Courses    []Course           `gorm:"foreignKey:CategoryID" json:"courses,omitempty"`
}

func (TrainingCategory) TableName() string {
	return "training_categories"
}

// Course represents a training course
type Course struct {
	BaseModel
	CompanyID   uuid.UUID    `gorm:"type:text;not null;index" json:"company_id"`
	CategoryID  *uuid.UUID   `gorm:"type:text;index" json:"category_id,omitempty"`

	// Basic Info
	Title       string       `gorm:"size:255;not null" json:"title"`
	Code        string       `gorm:"size:50;index" json:"code"`
	Description string       `gorm:"type:text" json:"description"`
	Objectives  string       `gorm:"type:text" json:"objectives"`

	// Classification
	CourseType  CourseType   `gorm:"size:50;not null;default:'self_paced'" json:"course_type"`
	Level       string       `gorm:"size:50" json:"level"` // beginner, intermediate, advanced
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`

	// Duration & Schedule
	DurationMinutes   int       `gorm:"default:0" json:"duration_minutes"`
	EstimatedWeeks    int       `gorm:"default:0" json:"estimated_weeks"`
	ScheduledStartDate *time.Time `json:"scheduled_start_date,omitempty"`
	ScheduledEndDate   *time.Time `json:"scheduled_end_date,omitempty"`

	// Requirements
	Prerequisites     string    `gorm:"type:text" json:"prerequisites"`
	RequiredForRoles  pq.StringArray `gorm:"type:text[]" json:"required_for_roles"`
	RequiredForDepts  pq.StringArray `gorm:"type:text[]" json:"required_for_departments"`
	IsRequired        bool      `gorm:"default:false" json:"is_required"`
	IsMandatory       bool      `gorm:"default:false" json:"is_mandatory"` // Required for compliance

	// Completion
	PassingScore      int       `gorm:"default:70" json:"passing_score"` // Percentage
	MaxAttempts       int       `gorm:"default:3" json:"max_attempts"`
	CertificateValid  int       `gorm:"default:0" json:"certificate_valid_months"` // 0 = never expires

	// Media
	ThumbnailURL      string    `gorm:"size:500" json:"thumbnail_url"`
	IntroVideoURL     string    `gorm:"size:500" json:"intro_video_url"`

	// Instructor
	InstructorID      *uuid.UUID `gorm:"type:text" json:"instructor_id,omitempty"`
	InstructorName    string    `gorm:"size:255" json:"instructor_name"`

	// Status
	Status            CourseStatus `gorm:"size:50;default:'draft'" json:"status"`
	PublishedAt       *time.Time   `json:"published_at,omitempty"`
	ArchivedAt        *time.Time   `json:"archived_at,omitempty"`

	// Settings
	AllowSelfEnroll   bool      `gorm:"default:true" json:"allow_self_enroll"`
	EnrollmentLimit   int       `gorm:"default:0" json:"enrollment_limit"` // 0 = unlimited
	ShowProgress      bool      `gorm:"default:true" json:"show_progress"`

	// Metadata
	CreatedByID       *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Category          *TrainingCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Modules           []CourseModule    `gorm:"foreignKey:CourseID" json:"modules,omitempty"`
	Enrollments       []CourseEnrollment `gorm:"foreignKey:CourseID" json:"enrollments,omitempty"`
}

func (Course) TableName() string {
	return "training_courses"
}

// CourseModule represents a module/section within a course
type CourseModule struct {
	BaseModel
	CourseID     uuid.UUID `gorm:"type:text;not null;index" json:"course_id"`

	Title        string    `gorm:"size:255;not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`

	// Duration
	DurationMinutes int    `gorm:"default:0" json:"duration_minutes"`

	// Requirements
	IsRequired       bool   `gorm:"default:true" json:"is_required"`
	UnlockAfterDays  int    `gorm:"default:0" json:"unlock_after_days"`
	RequiresModuleID *uuid.UUID `gorm:"type:text" json:"requires_module_id,omitempty"`

	IsActive     bool      `gorm:"default:true" json:"is_active"`

	// Relationships
	Course       *Course         `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Contents     []ModuleContent `gorm:"foreignKey:ModuleID" json:"contents,omitempty"`
}

func (CourseModule) TableName() string {
	return "training_course_modules"
}

// ModuleContent represents content items within a module
type ModuleContent struct {
	BaseModel
	ModuleID     uuid.UUID   `gorm:"type:text;not null;index" json:"module_id"`

	Title        string      `gorm:"size:255;not null" json:"title"`
	Description  string      `gorm:"type:text" json:"description"`
	ContentType  ContentType `gorm:"size:50;not null" json:"content_type"`
	DisplayOrder int         `gorm:"default:0" json:"display_order"`

	// Content Reference
	ContentURL   string      `gorm:"size:500" json:"content_url"`
	ContentData  string      `gorm:"type:text" json:"content_data"` // JSON for quizzes, etc.
	FileSize     int64       `gorm:"default:0" json:"file_size"`
	MimeType     string      `gorm:"size:100" json:"mime_type"`

	// Duration & Requirements
	DurationMinutes int     `gorm:"default:0" json:"duration_minutes"`
	IsRequired      bool    `gorm:"default:true" json:"is_required"`
	PassingScore    int     `gorm:"default:70" json:"passing_score"`
	MaxAttempts     int     `gorm:"default:3" json:"max_attempts"`

	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Relationships
	Module       *CourseModule `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
}

func (ModuleContent) TableName() string {
	return "training_module_contents"
}

// CourseEnrollment represents a user's enrollment in a course
type CourseEnrollment struct {
	BaseModel
	CompanyID    uuid.UUID        `gorm:"type:text;not null;index" json:"company_id"`
	CourseID     uuid.UUID        `gorm:"type:text;not null;index" json:"course_id"`
	EmployeeID   uuid.UUID        `gorm:"type:text;not null;index" json:"employee_id"`

	// Status
	Status       EnrollmentStatus `gorm:"size:50;default:'pending'" json:"status"`

	// Progress
	ProgressPercent int           `gorm:"default:0" json:"progress_percent"`
	CurrentModuleID *uuid.UUID    `gorm:"type:text" json:"current_module_id,omitempty"`
	CurrentContentID *uuid.UUID   `gorm:"type:text" json:"current_content_id,omitempty"`

	// Dates
	EnrolledAt   time.Time        `gorm:"autoCreateTime" json:"enrolled_at"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"`
	DueDate      *time.Time       `json:"due_date,omitempty"`
	ExpiresAt    *time.Time       `json:"expires_at,omitempty"`

	// Score
	FinalScore   int              `gorm:"default:0" json:"final_score"`
	TotalTimeMinutes int          `gorm:"default:0" json:"total_time_minutes"`

	// Enrollment Source
	EnrolledByID *uuid.UUID       `gorm:"type:text" json:"enrolled_by_id,omitempty"`
	EnrollmentType string         `gorm:"size:50" json:"enrollment_type"` // self, assigned, required

	// Certificate
	CertificateIssued   bool      `gorm:"default:false" json:"certificate_issued"`
	CertificateIssuedAt *time.Time `json:"certificate_issued_at,omitempty"`
	CertificateURL      string    `gorm:"size:500" json:"certificate_url"`
	CertificateExpiresAt *time.Time `json:"certificate_expires_at,omitempty"`

	// Relationships
	Course      *Course           `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Employee    *Employee         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Progress    []ContentProgress `gorm:"foreignKey:EnrollmentID" json:"progress,omitempty"`
}

func (CourseEnrollment) TableName() string {
	return "training_enrollments"
}

// ContentProgress tracks progress on individual content items
type ContentProgress struct {
	BaseModel
	EnrollmentID  uuid.UUID  `gorm:"type:text;not null;index" json:"enrollment_id"`
	ContentID     uuid.UUID  `gorm:"type:text;not null;index" json:"content_id"`

	// Progress
	IsCompleted   bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`

	// Time
	TimeSpentMinutes int     `gorm:"default:0" json:"time_spent_minutes"`
	LastAccessedAt   *time.Time `json:"last_accessed_at,omitempty"`

	// Score (for quizzes/assignments)
	Score         int        `gorm:"default:0" json:"score"`
	MaxScore      int        `gorm:"default:100" json:"max_score"`
	Attempts      int        `gorm:"default:0" json:"attempts"`

	// Response Data (for quizzes)
	ResponseData  string     `gorm:"type:text" json:"response_data"` // JSON

	// Relationships
	Enrollment    *CourseEnrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
	Content       *ModuleContent    `gorm:"foreignKey:ContentID" json:"content,omitempty"`
}

func (ContentProgress) TableName() string {
	return "training_content_progress"
}

// TrainingSession represents an instructor-led training session
type TrainingSession struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	CourseID     uuid.UUID  `gorm:"type:text;not null;index" json:"course_id"`

	// Session Info
	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`

	// Schedule
	ScheduledAt  time.Time  `gorm:"not null" json:"scheduled_at"`
	DurationMinutes int     `gorm:"default:60" json:"duration_minutes"`

	// Location
	Location     string     `gorm:"size:255" json:"location"`
	IsVirtual    bool       `gorm:"default:false" json:"is_virtual"`
	MeetingURL   string     `gorm:"size:500" json:"meeting_url"`

	// Instructor
	InstructorID uuid.UUID  `gorm:"type:text;not null" json:"instructor_id"`

	// Capacity
	MaxAttendees int        `gorm:"default:0" json:"max_attendees"` // 0 = unlimited

	// Status
	Status       string     `gorm:"size:50;default:'scheduled'" json:"status"` // scheduled, in_progress, completed, cancelled

	// Relationships
	Course       *Course    `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Attendees    []SessionAttendee `gorm:"foreignKey:SessionID" json:"attendees,omitempty"`
}

func (TrainingSession) TableName() string {
	return "training_sessions"
}

// SessionAttendee tracks attendance for training sessions
type SessionAttendee struct {
	BaseModel
	SessionID    uuid.UUID  `gorm:"type:text;not null;index" json:"session_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Registration
	RegisteredAt time.Time  `gorm:"autoCreateTime" json:"registered_at"`

	// Attendance
	Attended     bool       `gorm:"default:false" json:"attended"`
	CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`
	CheckedOutAt *time.Time `json:"checked_out_at,omitempty"`

	// Feedback
	Rating       int        `gorm:"default:0" json:"rating"` // 1-5
	Feedback     string     `gorm:"type:text" json:"feedback"`

	// Relationships
	Session      *TrainingSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Employee     *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (SessionAttendee) TableName() string {
	return "training_session_attendees"
}

// Certificate represents a training certificate
type Certificate struct {
	BaseModel
	CompanyID      uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EnrollmentID   uuid.UUID  `gorm:"type:text;not null;index" json:"enrollment_id"`
	EmployeeID     uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	CourseID       uuid.UUID  `gorm:"type:text;not null;index" json:"course_id"`

	// Certificate Info
	CertificateNumber string   `gorm:"size:100;uniqueIndex" json:"certificate_number"`
	Title           string     `gorm:"size:255;not null" json:"title"`

	// Dates
	IssuedAt        time.Time  `gorm:"not null" json:"issued_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`

	// Score
	FinalScore      int        `gorm:"default:0" json:"final_score"`

	// Document
	DocumentURL     string     `gorm:"size:500" json:"document_url"`

	// Status
	IsValid         bool       `gorm:"default:true" json:"is_valid"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	RevokedReason   string     `gorm:"size:255" json:"revoked_reason"`

	// Relationships
	Enrollment      *CourseEnrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
	Employee        *Employee         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Course          *Course           `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

func (Certificate) TableName() string {
	return "training_certificates"
}

// LearningPath represents a collection of courses
type LearningPath struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`

	// Target Audience
	TargetRoles  pq.StringArray `gorm:"type:text[]" json:"target_roles"`
	TargetDepts  pq.StringArray `gorm:"type:text[]" json:"target_departments"`

	// Completion
	EstimatedWeeks   int     `gorm:"default:0" json:"estimated_weeks"`
	TotalCourses     int     `gorm:"default:0" json:"total_courses"`

	// Status
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Metadata
	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Courses      []LearningPathCourse `gorm:"foreignKey:LearningPathID" json:"courses,omitempty"`
}

func (LearningPath) TableName() string {
	return "training_learning_paths"
}

// LearningPathCourse links courses to learning paths
type LearningPathCourse struct {
	BaseModel
	LearningPathID uuid.UUID `gorm:"type:text;not null;index" json:"learning_path_id"`
	CourseID       uuid.UUID `gorm:"type:text;not null;index" json:"course_id"`

	DisplayOrder   int       `gorm:"default:0" json:"display_order"`
	IsRequired     bool      `gorm:"default:true" json:"is_required"`

	// Relationships
	LearningPath   *LearningPath `gorm:"foreignKey:LearningPathID" json:"learning_path,omitempty"`
	Course         *Course       `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

func (LearningPathCourse) TableName() string {
	return "training_learning_path_courses"
}
