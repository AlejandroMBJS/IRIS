/*
Package models - IRIS Payroll System Performance Management Module Models

==============================================================================
FILE: internal/models/performance.go
==============================================================================

DESCRIPTION:
    Data models for Performance Management including reviews, goals,
    competencies, feedback, and development plans.

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ReviewStatus represents the status of a performance review
type ReviewStatus string

const (
	ReviewStatusDraft        ReviewStatus = "draft"
	ReviewStatusInProgress   ReviewStatus = "in_progress"
	ReviewStatusSelfReview   ReviewStatus = "self_review"
	ReviewStatusManagerReview ReviewStatus = "manager_review"
	ReviewStatusCalibration  ReviewStatus = "calibration"
	ReviewStatusCompleted    ReviewStatus = "completed"
	ReviewStatusCancelled    ReviewStatus = "cancelled"
)

// GoalStatus represents the status of a goal
type GoalStatus string

const (
	GoalStatusDraft      GoalStatus = "draft"
	GoalStatusActive     GoalStatus = "active"
	GoalStatusCompleted  GoalStatus = "completed"
	GoalStatusCancelled  GoalStatus = "cancelled"
	GoalStatusDeferred   GoalStatus = "deferred"
)

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackTypePositive    FeedbackType = "positive"
	FeedbackTypeConstructive FeedbackType = "constructive"
	FeedbackTypePeer        FeedbackType = "peer"
	FeedbackType360         FeedbackType = "360"
	FeedbackTypeRecognition FeedbackType = "recognition"
)

// ReviewCycle represents a performance review cycle/period
type ReviewCycle struct {
	BaseModel
	CompanyID    uuid.UUID    `gorm:"type:text;not null;index" json:"company_id"`

	Name         string       `gorm:"size:255;not null" json:"name"`
	Description  string       `gorm:"type:text" json:"description"`

	// Period
	StartDate    time.Time    `gorm:"not null" json:"start_date"`
	EndDate      time.Time    `gorm:"not null" json:"end_date"`
	Year         int          `gorm:"not null" json:"year"`
	Quarter      int          `json:"quarter"` // 0 for annual, 1-4 for quarterly

	// Review Type
	CycleType    string       `gorm:"size:50;not null" json:"cycle_type"` // annual, semi_annual, quarterly

	// Phases
	SelfReviewStart  *time.Time `json:"self_review_start,omitempty"`
	SelfReviewEnd    *time.Time `json:"self_review_end,omitempty"`
	ManagerReviewStart *time.Time `json:"manager_review_start,omitempty"`
	ManagerReviewEnd   *time.Time `json:"manager_review_end,omitempty"`
	CalibrationStart *time.Time `json:"calibration_start,omitempty"`
	CalibrationEnd   *time.Time `json:"calibration_end,omitempty"`

	// Status
	Status       string       `gorm:"size:50;default:'draft'" json:"status"` // draft, active, completed
	IsActive     bool         `gorm:"default:false" json:"is_active"`

	// Settings
	RequireSelfReview bool     `gorm:"default:true" json:"require_self_review"`
	Require360Review  bool     `gorm:"default:false" json:"require_360_review"`
	AllowPeerFeedback bool     `gorm:"default:true" json:"allow_peer_feedback"`

	// Template
	TemplateID   *uuid.UUID   `gorm:"type:text" json:"template_id,omitempty"`

	CreatedByID  *uuid.UUID   `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Reviews      []PerformanceReview `gorm:"foreignKey:CycleID" json:"reviews,omitempty"`
}

func (ReviewCycle) TableName() string {
	return "performance_review_cycles"
}

// ReviewTemplate defines the structure of a performance review
type ReviewTemplate struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	IsDefault    bool       `gorm:"default:false" json:"is_default"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Sections     []ReviewTemplateSection `gorm:"foreignKey:TemplateID" json:"sections,omitempty"`
}

func (ReviewTemplate) TableName() string {
	return "performance_review_templates"
}

// ReviewTemplateSection represents a section in a review template
type ReviewTemplateSection struct {
	BaseModel
	TemplateID   uuid.UUID  `gorm:"type:text;not null;index" json:"template_id"`

	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`
	Weight       int        `gorm:"default:0" json:"weight"` // Percentage weight for scoring

	// Section Type
	SectionType  string     `gorm:"size:50;not null" json:"section_type"` // competencies, goals, free_text, rating

	// Settings
	IsRequired   bool       `gorm:"default:true" json:"is_required"`
	AllowSelfRating bool    `gorm:"default:true" json:"allow_self_rating"`
	AllowManagerRating bool `gorm:"default:true" json:"allow_manager_rating"`

	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Relationships
	Template     *ReviewTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Questions    []ReviewQuestion `gorm:"foreignKey:SectionID" json:"questions,omitempty"`
}

func (ReviewTemplateSection) TableName() string {
	return "performance_review_template_sections"
}

// ReviewQuestion represents a question/criterion in a review section
type ReviewQuestion struct {
	BaseModel
	SectionID    uuid.UUID  `gorm:"type:text;not null;index" json:"section_id"`

	Text         string     `gorm:"type:text;not null" json:"text"`
	HelpText     string     `gorm:"type:text" json:"help_text"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	// Question Type
	QuestionType string     `gorm:"size:50;not null" json:"question_type"` // rating, text, scale, multi_choice

	// Rating Scale
	MinRating    int        `gorm:"default:1" json:"min_rating"`
	MaxRating    int        `gorm:"default:5" json:"max_rating"`
	RatingLabels string     `gorm:"type:text" json:"rating_labels"` // JSON: {"1": "Poor", "5": "Excellent"}

	// Options (for multi_choice)
	Options      string     `gorm:"type:text" json:"options"` // JSON array

	IsRequired   bool       `gorm:"default:true" json:"is_required"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Relationships
	Section      *ReviewTemplateSection `gorm:"foreignKey:SectionID" json:"section,omitempty"`
}

func (ReviewQuestion) TableName() string {
	return "performance_review_questions"
}

// PerformanceReview represents an employee's performance review
type PerformanceReview struct {
	BaseModel
	CompanyID    uuid.UUID    `gorm:"type:text;not null;index" json:"company_id"`
	CycleID      uuid.UUID    `gorm:"type:text;not null;index" json:"cycle_id"`
	EmployeeID   uuid.UUID    `gorm:"type:text;not null;index" json:"employee_id"`
	ManagerID    *uuid.UUID   `gorm:"type:text;index" json:"manager_id,omitempty"`
	TemplateID   *uuid.UUID   `gorm:"type:text" json:"template_id,omitempty"`

	// Status
	Status       ReviewStatus `gorm:"size:50;default:'draft'" json:"status"`

	// Dates
	SelfReviewStartedAt  *time.Time `json:"self_review_started_at,omitempty"`
	SelfReviewCompletedAt *time.Time `json:"self_review_completed_at,omitempty"`
	ManagerReviewStartedAt *time.Time `json:"manager_review_started_at,omitempty"`
	ManagerReviewCompletedAt *time.Time `json:"manager_review_completed_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`

	// Scores
	SelfRating       float64  `gorm:"default:0" json:"self_rating"`
	ManagerRating    float64  `gorm:"default:0" json:"manager_rating"`
	FinalRating      float64  `gorm:"default:0" json:"final_rating"`
	CalibratedRating float64  `gorm:"default:0" json:"calibrated_rating"`

	// Overall Assessment
	OverallPerformance string  `gorm:"size:50" json:"overall_performance"` // exceeds, meets, below
	SelfComments       string  `gorm:"type:text" json:"self_comments"`
	ManagerComments    string  `gorm:"type:text" json:"manager_comments"`

	// Calibration
	CalibrationNotes   string  `gorm:"type:text" json:"calibration_notes"`
	CalibratedByID     *uuid.UUID `gorm:"type:text" json:"calibrated_by_id,omitempty"`
	CalibratedAt       *time.Time `json:"calibrated_at,omitempty"`

	// Acknowledgement
	EmployeeAcknowledged bool   `gorm:"default:false" json:"employee_acknowledged"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgementComments string `gorm:"type:text" json:"acknowledgement_comments"`

	// Relationships
	Cycle        *ReviewCycle    `gorm:"foreignKey:CycleID" json:"cycle,omitempty"`
	Employee     *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Responses    []ReviewResponse `gorm:"foreignKey:ReviewID" json:"responses,omitempty"`
	Goals        []Goal          `gorm:"foreignKey:ReviewID" json:"goals,omitempty"`
}

func (PerformanceReview) TableName() string {
	return "performance_reviews"
}

// ReviewResponse stores responses to review questions
type ReviewResponse struct {
	BaseModel
	ReviewID     uuid.UUID  `gorm:"type:text;not null;index" json:"review_id"`
	QuestionID   uuid.UUID  `gorm:"type:text;not null;index" json:"question_id"`

	// Self Assessment
	SelfRating   int        `gorm:"default:0" json:"self_rating"`
	SelfComment  string     `gorm:"type:text" json:"self_comment"`

	// Manager Assessment
	ManagerRating int       `gorm:"default:0" json:"manager_rating"`
	ManagerComment string   `gorm:"type:text" json:"manager_comment"`

	// Final
	FinalRating  int        `gorm:"default:0" json:"final_rating"`

	// Relationships
	Review       *PerformanceReview `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	Question     *ReviewQuestion    `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (ReviewResponse) TableName() string {
	return "performance_review_responses"
}

// Goal represents a performance goal
type Goal struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	ReviewID     *uuid.UUID `gorm:"type:text;index" json:"review_id,omitempty"`
	CycleID      *uuid.UUID `gorm:"type:text;index" json:"cycle_id,omitempty"`

	// Goal Info
	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`

	// Category
	Category     string     `gorm:"size:100" json:"category"` // business, development, personal
	Priority     string     `gorm:"size:50" json:"priority"` // high, medium, low

	// SMART Goal Details
	Measurable   string     `gorm:"type:text" json:"measurable"` // How will it be measured
	TargetValue  float64    `gorm:"default:0" json:"target_value"`
	CurrentValue float64    `gorm:"default:0" json:"current_value"`
	Unit         string     `gorm:"size:50" json:"unit"` // %, $, count, etc.

	// Timeline
	StartDate    *time.Time `json:"start_date,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`

	// Progress
	Status       GoalStatus `gorm:"size:50;default:'draft'" json:"status"`
	ProgressPercent int     `gorm:"default:0" json:"progress_percent"`

	// Weight for review
	Weight       int        `gorm:"default:0" json:"weight"`

	// Assessment
	SelfRating   int        `gorm:"default:0" json:"self_rating"`
	ManagerRating int       `gorm:"default:0" json:"manager_rating"`
	SelfComment  string     `gorm:"type:text" json:"self_comment"`
	ManagerComment string   `gorm:"type:text" json:"manager_comment"`

	// Alignment
	AlignedToGoalID *uuid.UUID `gorm:"type:text" json:"aligned_to_goal_id,omitempty"` // Parent/company goal

	// Visibility
	IsPrivate    bool       `gorm:"default:false" json:"is_private"`

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Review       *PerformanceReview `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	Updates      []GoalUpdate `gorm:"foreignKey:GoalID" json:"updates,omitempty"`
}

func (Goal) TableName() string {
	return "performance_goals"
}

// GoalUpdate tracks updates/progress on goals
type GoalUpdate struct {
	BaseModel
	GoalID       uuid.UUID  `gorm:"type:text;not null;index" json:"goal_id"`

	// Update Info
	Note         string     `gorm:"type:text;not null" json:"note"`
	NewProgress  int        `gorm:"default:0" json:"new_progress"`
	NewValue     float64    `gorm:"default:0" json:"new_value"`

	// Who updated
	UpdatedByID  uuid.UUID  `gorm:"type:text;not null" json:"updated_by_id"`

	// Relationships
	Goal         *Goal      `gorm:"foreignKey:GoalID" json:"goal,omitempty"`
}

func (GoalUpdate) TableName() string {
	return "performance_goal_updates"
}

// Competency represents a skill/competency to be evaluated
type Competency struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	Category     string     `gorm:"size:100" json:"category"` // core, leadership, technical, behavioral

	// Proficiency Levels
	Levels       string     `gorm:"type:text" json:"levels"` // JSON: [{level: 1, name: "Beginner", description: "..."}, ...]

	// Applicability
	ApplicableRoles pq.StringArray `gorm:"type:text[]" json:"applicable_roles"`
	ApplicableLevels pq.StringArray `gorm:"type:text[]" json:"applicable_levels"` // entry, mid, senior, etc.

	IsActive     bool       `gorm:"default:true" json:"is_active"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (Competency) TableName() string {
	return "performance_competencies"
}

// CompetencyAssessment stores competency ratings in reviews
type CompetencyAssessment struct {
	BaseModel
	ReviewID     uuid.UUID  `gorm:"type:text;not null;index" json:"review_id"`
	CompetencyID uuid.UUID  `gorm:"type:text;not null;index" json:"competency_id"`

	// Ratings
	SelfRating   int        `gorm:"default:0" json:"self_rating"`
	ManagerRating int       `gorm:"default:0" json:"manager_rating"`
	FinalRating  int        `gorm:"default:0" json:"final_rating"`

	// Comments
	SelfComment  string     `gorm:"type:text" json:"self_comment"`
	ManagerComment string   `gorm:"type:text" json:"manager_comment"`

	// Evidence
	Evidence     string     `gorm:"type:text" json:"evidence"` // Examples demonstrating the competency

	// Relationships
	Review       *PerformanceReview `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	Competency   *Competency        `gorm:"foreignKey:CompetencyID" json:"competency,omitempty"`
}

func (CompetencyAssessment) TableName() string {
	return "performance_competency_assessments"
}

// Feedback represents peer/continuous feedback
type Feedback struct {
	BaseModel
	CompanyID    uuid.UUID    `gorm:"type:text;not null;index" json:"company_id"`

	// Participants
	RecipientID  uuid.UUID    `gorm:"type:text;not null;index" json:"recipient_id"`
	GiverID      uuid.UUID    `gorm:"type:text;not null;index" json:"giver_id"`

	// Type
	FeedbackType FeedbackType `gorm:"size:50;not null" json:"feedback_type"`

	// Content
	Subject      string       `gorm:"size:255" json:"subject"`
	Content      string       `gorm:"type:text;not null" json:"content"`

	// Rating (optional)
	Rating       int          `gorm:"default:0" json:"rating"` // 1-5

	// Skills/Competencies mentioned
	Skills       pq.StringArray `gorm:"type:text[]" json:"skills"`

	// Visibility
	IsAnonymous  bool         `gorm:"default:false" json:"is_anonymous"`
	IsPrivate    bool         `gorm:"default:false" json:"is_private"`
	VisibleToManager bool     `gorm:"default:true" json:"visible_to_manager"`

	// Link to review (if part of 360)
	ReviewID     *uuid.UUID   `gorm:"type:text" json:"review_id,omitempty"`

	// Request Info (if requested)
	RequestedAt  *time.Time   `json:"requested_at,omitempty"`
	RequestedByID *uuid.UUID  `gorm:"type:text" json:"requested_by_id,omitempty"`

	// Relationships
	Recipient    *Employee    `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Giver        *Employee    `gorm:"foreignKey:GiverID" json:"giver,omitempty"`
}

func (Feedback) TableName() string {
	return "performance_feedback"
}

// DevelopmentPlan represents an employee development plan
type DevelopmentPlan struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	ReviewID     *uuid.UUID `gorm:"type:text" json:"review_id,omitempty"`

	// Plan Info
	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`

	// Period
	StartDate    time.Time  `gorm:"not null" json:"start_date"`
	EndDate      time.Time  `gorm:"not null" json:"end_date"`

	// Focus Areas
	FocusAreas   pq.StringArray `gorm:"type:text[]" json:"focus_areas"`
	CareerGoal   string     `gorm:"type:text" json:"career_goal"`

	// Status
	Status       string     `gorm:"size:50;default:'active'" json:"status"` // draft, active, completed
	ProgressPercent int     `gorm:"default:0" json:"progress_percent"`

	// Manager Approval
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Review       *PerformanceReview `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	Actions      []DevelopmentAction `gorm:"foreignKey:PlanID" json:"actions,omitempty"`
}

func (DevelopmentPlan) TableName() string {
	return "performance_development_plans"
}

// DevelopmentAction represents an action item in a development plan
type DevelopmentAction struct {
	BaseModel
	PlanID       uuid.UUID  `gorm:"type:text;not null;index" json:"plan_id"`

	// Action Info
	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	ActionType   string     `gorm:"size:50" json:"action_type"` // training, mentoring, assignment, reading, etc.

	// Timeline
	DueDate      *time.Time `json:"due_date,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`

	// Status
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, in_progress, completed
	ProgressPercent int     `gorm:"default:0" json:"progress_percent"`

	// Link to training
	CourseID     *uuid.UUID `gorm:"type:text" json:"course_id,omitempty"`

	// Notes
	Notes        string     `gorm:"type:text" json:"notes"`
	CompletionNotes string  `gorm:"type:text" json:"completion_notes"`

	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	// Relationships
	Plan         *DevelopmentPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (DevelopmentAction) TableName() string {
	return "performance_development_actions"
}

// OneOnOne represents a 1:1 meeting between manager and employee
type OneOnOne struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	ManagerID    uuid.UUID  `gorm:"type:text;not null;index" json:"manager_id"`

	// Schedule
	ScheduledAt  time.Time  `gorm:"not null" json:"scheduled_at"`
	DurationMinutes int     `gorm:"default:30" json:"duration_minutes"`

	// Recurrence
	IsRecurring  bool       `gorm:"default:false" json:"is_recurring"`
	RecurringPattern string `gorm:"size:50" json:"recurring_pattern"` // weekly, biweekly, monthly

	// Status
	Status       string     `gorm:"size:50;default:'scheduled'" json:"status"` // scheduled, completed, cancelled, rescheduled

	// Notes
	EmployeeAgenda string   `gorm:"type:text" json:"employee_agenda"`
	ManagerAgenda  string   `gorm:"type:text" json:"manager_agenda"`
	MeetingNotes   string   `gorm:"type:text" json:"meeting_notes"`
	PrivateNotes   string   `gorm:"type:text" json:"private_notes"` // Manager only

	// Action Items
	ActionItems    string   `gorm:"type:text" json:"action_items"` // JSON array

	// Sentiment
	EmployeeMood   string   `gorm:"size:50" json:"employee_mood"` // great, good, okay, struggling
	Engagement     int      `gorm:"default:0" json:"engagement"` // 1-5 rating

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (OneOnOne) TableName() string {
	return "performance_one_on_ones"
}
