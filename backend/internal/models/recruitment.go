/*
Package models - IRIS Payroll System Recruitment Models

==============================================================================
FILE: internal/models/recruitment.go
==============================================================================

DESCRIPTION:
    Defines data models for the recruitment module including job postings,
    candidates, applications, interviews, offers, and evaluation criteria.

USER PERSPECTIVE:
    - HR can create and manage job postings
    - Candidates can apply for open positions
    - Hiring managers can track applicants through the pipeline
    - Interview feedback and offer management supported

DEVELOPER GUIDELINES:
    OK to modify:
      - Add new fields as recruitment features expand
      - Add new validation rules
      - Add helper methods

    DO NOT modify:
      - Primary key structure (ID field)
      - Foreign key relationships without migration

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// JobPostingStatus represents the lifecycle status of a job posting
type JobPostingStatus string

const (
	JobPostingStatusDraft     JobPostingStatus = "draft"
	JobPostingStatusPublished JobPostingStatus = "published"
	JobPostingStatusPaused    JobPostingStatus = "paused"
	JobPostingStatusClosed    JobPostingStatus = "closed"
	JobPostingStatusFilled    JobPostingStatus = "filled"
)

// EmploymentType represents the type of employment offered
type EmploymentType string

const (
	EmploymentTypeFullTime  EmploymentType = "full_time"
	EmploymentTypePartTime  EmploymentType = "part_time"
	EmploymentTypeContract  EmploymentType = "contract"
	EmploymentTypeIntern    EmploymentType = "intern"
	EmploymentTypeTemporary EmploymentType = "temporary"
)

// RemoteType represents remote work arrangement
type RemoteType string

const (
	RemoteTypeFullyRemote RemoteType = "fully_remote"
	RemoteTypeHybrid      RemoteType = "hybrid"
	RemoteTypeOnSite      RemoteType = "on_site"
)

// JobPosting represents a job listing in the recruitment system
type JobPosting struct {
	BaseModel

	CompanyID uuid.UUID `gorm:"type:text;not null;index" json:"company_id"`
	Company   *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	// Basic Information
	Title            string `gorm:"size:255;not null" json:"title"`
	Description      string `gorm:"type:text;not null" json:"description"`
	Requirements     string `gorm:"type:text" json:"requirements,omitempty"`
	Responsibilities string `gorm:"type:text" json:"responsibilities,omitempty"`

	// Classification
	DepartmentID  *uuid.UUID `gorm:"type:text;index" json:"department_id,omitempty"`
	CostCenterID  *uuid.UUID `gorm:"type:text;index" json:"cost_center_id,omitempty"`
	PositionLevel string     `gorm:"size:50" json:"position_level,omitempty"` // entry, mid, senior, manager, director
	EmploymentType EmploymentType `gorm:"size:50;not null" json:"employment_type"`
	CollarType    string     `gorm:"size:20" json:"collar_type,omitempty"` // white_collar, blue_collar, gray_collar

	// Compensation
	SalaryMin       float64 `gorm:"type:decimal(12,2)" json:"salary_min,omitempty"`
	SalaryMax       float64 `gorm:"type:decimal(12,2)" json:"salary_max,omitempty"`
	SalaryCurrency  string  `gorm:"size:3;default:'MXN'" json:"salary_currency"`
	SalaryFrequency string  `gorm:"size:20" json:"salary_frequency,omitempty"` // daily, biweekly, monthly, annual
	ShowSalary      bool    `gorm:"default:false" json:"show_salary"`

	// Headcount
	PositionsAvailable int `gorm:"default:1" json:"positions_available"`
	PositionsFilled    int `gorm:"default:0" json:"positions_filled"`

	// Location
	Location   string     `gorm:"size:255" json:"location,omitempty"`
	IsRemote   bool       `gorm:"default:false" json:"is_remote"`
	RemoteType RemoteType `gorm:"size:50" json:"remote_type,omitempty"`

	// Status & Dates
	Status      JobPostingStatus `gorm:"size:50;default:'draft';not null" json:"status"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
	ClosesAt    *time.Time       `json:"closes_at,omitempty"`
	ClosedAt    *time.Time       `json:"closed_at,omitempty"`
	CloseReason string           `gorm:"size:255" json:"close_reason,omitempty"`

	// Workflow
	HiringManagerID *uuid.UUID `gorm:"type:text;index" json:"hiring_manager_id,omitempty"`
	HiringManager   *User      `gorm:"foreignKey:HiringManagerID" json:"hiring_manager,omitempty"`
	RecruiterID     *uuid.UUID `gorm:"type:text;index" json:"recruiter_id,omitempty"`
	Recruiter       *User      `gorm:"foreignKey:RecruiterID" json:"recruiter,omitempty"`

	// Metadata
	CreatedByID *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
	CreatedBy   *User      `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// Relationships
	Applications []Application `gorm:"foreignKey:JobPostingID" json:"applications,omitempty"`
}

// TableName specifies the table name for JobPosting
func (JobPosting) TableName() string {
	return "job_postings"
}

// IsOpen checks if the job posting is accepting applications
func (jp *JobPosting) IsOpen() bool {
	if jp.Status != JobPostingStatusPublished {
		return false
	}
	if jp.ClosesAt != nil && jp.ClosesAt.Before(time.Now()) {
		return false
	}
	if jp.PositionsFilled >= jp.PositionsAvailable {
		return false
	}
	return true
}

// CandidateStatus represents the status of a candidate in the system
type CandidateStatus string

const (
	CandidateStatusActive    CandidateStatus = "active"
	CandidateStatusHired     CandidateStatus = "hired"
	CandidateStatusRejected  CandidateStatus = "rejected"
	CandidateStatusWithdrawn CandidateStatus = "withdrawn"
)

// Candidate represents a person applying for jobs
type Candidate struct {
	BaseModel

	CompanyID uuid.UUID `gorm:"type:text;not null;index" json:"company_id"`
	Company   *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	// Personal Information
	FirstName string `gorm:"size:100;not null" json:"first_name"`
	LastName  string `gorm:"size:100;not null" json:"last_name"`
	Email     string `gorm:"size:255;not null;index" json:"email"`
	Phone     string `gorm:"size:20" json:"phone,omitempty"`

	// Professional Information
	CurrentTitle   string `gorm:"size:255" json:"current_title,omitempty"`
	CurrentCompany string `gorm:"size:255" json:"current_company,omitempty"`
	YearsExperience int   `json:"years_experience,omitempty"`
	LinkedInURL    string `gorm:"size:255" json:"linkedin_url,omitempty"`
	PortfolioURL   string `gorm:"size:255" json:"portfolio_url,omitempty"`

	// Documents
	ResumeFileID *uuid.UUID `gorm:"type:text" json:"resume_file_id,omitempty"`
	CoverLetter  string     `gorm:"type:text" json:"cover_letter,omitempty"`

	// Source Tracking
	Source                string     `gorm:"size:100" json:"source,omitempty"` // linkedin, referral, job_board, website, agency
	SourceDetails         string     `gorm:"size:255" json:"source_details,omitempty"`
	ReferredByEmployeeID  *uuid.UUID `gorm:"type:text" json:"referred_by_employee_id,omitempty"`
	ReferredByEmployee    *Employee  `gorm:"foreignKey:ReferredByEmployeeID;constraint:OnDelete:SET NULL" json:"referred_by_employee,omitempty"`

	// Status
	Status CandidateStatus `gorm:"size:50;default:'active'" json:"status"`

	// Tags
	Tags pq.StringArray `gorm:"type:text[]" json:"tags,omitempty"`

	// Relationships
	Applications []Application `gorm:"foreignKey:CandidateID" json:"applications,omitempty"`
}

// TableName specifies the table name for Candidate
func (Candidate) TableName() string {
	return "candidates"
}

// FullName returns the candidate's full name
func (c *Candidate) FullName() string {
	return c.FirstName + " " + c.LastName
}

// ApplicationStage represents the recruitment pipeline stage
type ApplicationStage string

const (
	ApplicationStageNew               ApplicationStage = "new"
	ApplicationStageScreening         ApplicationStage = "screening"
	ApplicationStagePhoneInterview    ApplicationStage = "phone_interview"
	ApplicationStageTechnicalInterview ApplicationStage = "technical_interview"
	ApplicationStageOnsiteInterview   ApplicationStage = "onsite_interview"
	ApplicationStageOffer             ApplicationStage = "offer"
	ApplicationStageHired             ApplicationStage = "hired"
	ApplicationStageRejected          ApplicationStage = "rejected"
)

// ApplicationStatus represents the overall application status
type ApplicationStatus string

const (
	ApplicationStatusActive    ApplicationStatus = "active"
	ApplicationStatusWithdrawn ApplicationStatus = "withdrawn"
	ApplicationStatusRejected  ApplicationStatus = "rejected"
	ApplicationStatusHired     ApplicationStatus = "hired"
)

// Application links candidates to job postings with pipeline tracking
type Application struct {
	BaseModel

	CompanyID    uuid.UUID   `gorm:"type:text;not null;index" json:"company_id"`
	CandidateID  uuid.UUID   `gorm:"type:text;not null;index" json:"candidate_id"`
	Candidate    *Candidate  `gorm:"foreignKey:CandidateID" json:"candidate,omitempty"`
	JobPostingID uuid.UUID   `gorm:"type:text;not null;index" json:"job_posting_id"`
	JobPosting   *JobPosting `gorm:"foreignKey:JobPostingID" json:"job_posting,omitempty"`

	// Pipeline Stage
	Stage          ApplicationStage `gorm:"size:50;not null;default:'new'" json:"stage"`
	StageEnteredAt time.Time        `gorm:"autoCreateTime" json:"stage_entered_at"`

	// Status
	Status          ApplicationStatus `gorm:"size:50;default:'active'" json:"status"`
	RejectionReason string            `gorm:"size:255" json:"rejection_reason,omitempty"`
	RejectionNotes  string            `gorm:"type:text" json:"rejection_notes,omitempty"`

	// Screening
	ScreeningScore int        `json:"screening_score,omitempty"` // 1-5 rating
	ScreeningNotes string     `gorm:"type:text" json:"screening_notes,omitempty"`
	ScreenedByID   *uuid.UUID `gorm:"type:text" json:"screened_by_id,omitempty"`
	ScreenedBy     *User      `gorm:"foreignKey:ScreenedByID" json:"screened_by,omitempty"`
	ScreenedAt     *time.Time `json:"screened_at,omitempty"`

	// Overall Rating (aggregated from interviews)
	OverallRating float64 `gorm:"type:decimal(3,2)" json:"overall_rating,omitempty"`

	// Salary Negotiation
	ExpectedSalary float64 `gorm:"type:decimal(12,2)" json:"expected_salary,omitempty"`
	OfferedSalary  float64 `gorm:"type:decimal(12,2)" json:"offered_salary,omitempty"`

	// Timeline
	AppliedAt       time.Time  `gorm:"autoCreateTime" json:"applied_at"`
	LastActivityAt  time.Time  `gorm:"autoUpdateTime" json:"last_activity_at"`
	HiredAt         *time.Time `json:"hired_at,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`

	// Relationships
	Interviews []Interview `gorm:"foreignKey:ApplicationID" json:"interviews,omitempty"`
	Offers     []Offer     `gorm:"foreignKey:ApplicationID" json:"offers,omitempty"`
}

// TableName specifies the table name for Application
func (Application) TableName() string {
	return "applications"
}

// InterviewType represents the type of interview
type InterviewType string

const (
	InterviewTypePhone      InterviewType = "phone"
	InterviewTypeVideo      InterviewType = "video"
	InterviewTypeOnsite     InterviewType = "onsite"
	InterviewTypeTechnical  InterviewType = "technical"
	InterviewTypeBehavioral InterviewType = "behavioral"
	InterviewTypePanel      InterviewType = "panel"
)

// InterviewStatus represents the status of an interview
type InterviewStatus string

const (
	InterviewStatusScheduled InterviewStatus = "scheduled"
	InterviewStatusCompleted InterviewStatus = "completed"
	InterviewStatusCancelled InterviewStatus = "cancelled"
	InterviewStatusNoShow    InterviewStatus = "no_show"
)

// InterviewRecommendation represents the interviewer's hiring recommendation
type InterviewRecommendation string

const (
	InterviewRecommendationStrongHire   InterviewRecommendation = "strong_hire"
	InterviewRecommendationHire         InterviewRecommendation = "hire"
	InterviewRecommendationNoHire       InterviewRecommendation = "no_hire"
	InterviewRecommendationStrongNoHire InterviewRecommendation = "strong_no_hire"
)

// Interview represents an interview session
type Interview struct {
	BaseModel

	ApplicationID uuid.UUID    `gorm:"type:text;not null;index" json:"application_id"`
	Application   *Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`

	// Schedule
	ScheduledAt     time.Time `gorm:"not null" json:"scheduled_at"`
	DurationMinutes int       `gorm:"default:60" json:"duration_minutes"`
	Location        string    `gorm:"size:255" json:"location,omitempty"` // Room, video link, etc.
	InterviewType   InterviewType `gorm:"size:50" json:"interview_type"`

	// Participants
	InterviewerID        uuid.UUID      `gorm:"type:text;not null" json:"interviewer_id"`
	Interviewer          *User          `gorm:"foreignKey:InterviewerID" json:"interviewer,omitempty"`
	AdditionalInterviewers pq.StringArray `gorm:"type:text[]" json:"additional_interviewers,omitempty"` // Array of user IDs

	// Status
	Status InterviewStatus `gorm:"size:50;default:'scheduled'" json:"status"`

	// Feedback (after interview)
	Rating              int                     `json:"rating,omitempty"` // 1-5
	Strengths           string                  `gorm:"type:text" json:"strengths,omitempty"`
	Weaknesses          string                  `gorm:"type:text" json:"weaknesses,omitempty"`
	Recommendation      InterviewRecommendation `gorm:"size:50" json:"recommendation,omitempty"`
	Notes               string                  `gorm:"type:text" json:"notes,omitempty"`
	FeedbackSubmittedAt *time.Time              `json:"feedback_submitted_at,omitempty"`

	// Calendar Integration
	CalendarEventID string `gorm:"size:255" json:"calendar_event_id,omitempty"`
	VideoMeetingURL string `gorm:"size:255" json:"video_meeting_url,omitempty"`

	// Relationships
	Scores []InterviewScore `gorm:"foreignKey:InterviewID" json:"scores,omitempty"`
}

// TableName specifies the table name for Interview
func (Interview) TableName() string {
	return "interviews"
}

// OfferStatus represents the status of a job offer
type OfferStatus string

const (
	OfferStatusDraft     OfferStatus = "draft"
	OfferStatusSent      OfferStatus = "sent"
	OfferStatusAccepted  OfferStatus = "accepted"
	OfferStatusDeclined  OfferStatus = "declined"
	OfferStatusExpired   OfferStatus = "expired"
	OfferStatusWithdrawn OfferStatus = "withdrawn"
)

// Offer represents a job offer to a candidate
type Offer struct {
	BaseModel

	ApplicationID uuid.UUID    `gorm:"type:text;not null;index" json:"application_id"`
	Application   *Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`

	// Offer Details
	PositionTitle string     `gorm:"size:255;not null" json:"position_title"`
	DepartmentID  *uuid.UUID `gorm:"type:text" json:"department_id,omitempty"`

	// Compensation
	SalaryAmount    float64 `gorm:"type:decimal(12,2);not null" json:"salary_amount"`
	SalaryFrequency string  `gorm:"size:20;not null" json:"salary_frequency"`
	BonusAmount     float64 `gorm:"type:decimal(12,2)" json:"bonus_amount,omitempty"`
	BonusType       string  `gorm:"size:50" json:"bonus_type,omitempty"` // signing, annual, quarterly

	// Benefits
	BenefitsPackage string `gorm:"type:text" json:"benefits_package,omitempty"`

	// Dates
	StartDate      time.Time  `gorm:"not null" json:"start_date"`
	OfferExpiresAt *time.Time `json:"offer_expires_at,omitempty"`

	// Status
	Status        OfferStatus `gorm:"size:50;default:'draft'" json:"status"`
	SentAt        *time.Time  `json:"sent_at,omitempty"`
	RespondedAt   *time.Time  `json:"responded_at,omitempty"`
	ResponseNotes string      `gorm:"type:text" json:"response_notes,omitempty"`

	// Documents
	OfferLetterDocumentID *uuid.UUID `gorm:"type:text" json:"offer_letter_document_id,omitempty"`
	SignedDocumentID      *uuid.UUID `gorm:"type:text" json:"signed_document_id,omitempty"`

	// Approval
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedBy   *User      `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`

	// Metadata
	CreatedByID *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
	CreatedBy   *User      `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

// TableName specifies the table name for Offer
func (Offer) TableName() string {
	return "offers"
}

// IsExpired checks if the offer has expired
func (o *Offer) IsExpired() bool {
	if o.OfferExpiresAt == nil {
		return false
	}
	return o.OfferExpiresAt.Before(time.Now())
}

// EvaluationCriteriaCategory represents categories of evaluation criteria
type EvaluationCriteriaCategory string

const (
	EvaluationCriteriaCategoryTechnical  EvaluationCriteriaCategory = "technical"
	EvaluationCriteriaCategoryBehavioral EvaluationCriteriaCategory = "behavioral"
	EvaluationCriteriaCategoryCultural   EvaluationCriteriaCategory = "cultural"
	EvaluationCriteriaCategoryExperience EvaluationCriteriaCategory = "experience"
)

// EvaluationCriteria defines what to evaluate in interviews
type EvaluationCriteria struct {
	BaseModel

	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	JobPostingID *uuid.UUID `gorm:"type:text;index" json:"job_posting_id,omitempty"` // NULL = company-wide

	Name         string                     `gorm:"size:255;not null" json:"name"`
	Description  string                     `gorm:"type:text" json:"description,omitempty"`
	Category     EvaluationCriteriaCategory `gorm:"size:50" json:"category"`
	Weight       int                        `gorm:"default:1" json:"weight"` // For weighted scoring

	IsRequired   bool `gorm:"default:false" json:"is_required"`
	DisplayOrder int  `gorm:"default:0" json:"display_order"`
	IsActive     bool `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for EvaluationCriteria
func (EvaluationCriteria) TableName() string {
	return "evaluation_criteria"
}

// InterviewScore stores individual criteria scores from interviews
type InterviewScore struct {
	BaseModel

	InterviewID uuid.UUID           `gorm:"type:text;not null;index;uniqueIndex:idx_interview_criteria" json:"interview_id"`
	Interview   *Interview          `gorm:"foreignKey:InterviewID" json:"interview,omitempty"`
	CriteriaID  uuid.UUID           `gorm:"type:text;not null;uniqueIndex:idx_interview_criteria" json:"criteria_id"`
	Criteria    *EvaluationCriteria `gorm:"foreignKey:CriteriaID" json:"criteria,omitempty"`

	Score int    `gorm:"not null" json:"score"` // 1-5
	Notes string `gorm:"type:text" json:"notes,omitempty"`
}

// TableName specifies the table name for InterviewScore
func (InterviewScore) TableName() string {
	return "interview_scores"
}
