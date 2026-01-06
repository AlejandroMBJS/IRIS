/*
Package services - IRIS Performance Management Service

==============================================================================
FILE: internal/services/performance_service.go
==============================================================================

DESCRIPTION:
    Business logic for Performance Management including reviews, goals,
    competencies, feedback, and development plans.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PerformanceService provides business logic for performance management
type PerformanceService struct {
	db *gorm.DB
}

// NewPerformanceService creates a new PerformanceService
func NewPerformanceService(db *gorm.DB) *PerformanceService {
	return &PerformanceService{db: db}
}

// === Review Cycle DTOs ===

type CreateReviewCycleDTO struct {
	CompanyID          uuid.UUID
	Name               string
	Description        string
	StartDate          time.Time
	EndDate            time.Time
	Year               int
	Quarter            int
	CycleType          string
	SelfReviewStart    *time.Time
	SelfReviewEnd      *time.Time
	ManagerReviewStart *time.Time
	ManagerReviewEnd   *time.Time
	TemplateID         *uuid.UUID
	RequireSelfReview  bool
	CreatedByID        *uuid.UUID
}

type ReviewCycleFilters struct {
	CompanyID uuid.UUID
	Year      int
	Status    string
	Page      int
	Limit     int
}

type PaginatedReviewCycles struct {
	Data       []models.ReviewCycle `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// CreateReviewCycle creates a new review cycle
func (s *PerformanceService) CreateReviewCycle(dto CreateReviewCycleDTO) (*models.ReviewCycle, error) {
	if dto.Name == "" {
		return nil, errors.New("cycle name is required")
	}

	cycle := &models.ReviewCycle{
		CompanyID:          dto.CompanyID,
		Name:               dto.Name,
		Description:        dto.Description,
		StartDate:          dto.StartDate,
		EndDate:            dto.EndDate,
		Year:               dto.Year,
		Quarter:            dto.Quarter,
		CycleType:          dto.CycleType,
		SelfReviewStart:    dto.SelfReviewStart,
		SelfReviewEnd:      dto.SelfReviewEnd,
		ManagerReviewStart: dto.ManagerReviewStart,
		ManagerReviewEnd:   dto.ManagerReviewEnd,
		TemplateID:         dto.TemplateID,
		RequireSelfReview:  dto.RequireSelfReview,
		Status:             "draft",
		IsActive:           false,
		CreatedByID:        dto.CreatedByID,
	}

	if cycle.Year == 0 {
		cycle.Year = time.Now().Year()
	}
	if cycle.CycleType == "" {
		cycle.CycleType = "annual"
	}

	if err := s.db.Create(cycle).Error; err != nil {
		return nil, err
	}

	return cycle, nil
}

// GetReviewCycleByID retrieves a review cycle by ID
func (s *PerformanceService) GetReviewCycleByID(id uuid.UUID) (*models.ReviewCycle, error) {
	var cycle models.ReviewCycle
	err := s.db.First(&cycle, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("review cycle not found")
		}
		return nil, err
	}
	return &cycle, nil
}

// ListReviewCycles retrieves review cycles with filters
func (s *PerformanceService) ListReviewCycles(filters ReviewCycleFilters) (*PaginatedReviewCycles, error) {
	query := s.db.Model(&models.ReviewCycle{}).Where("company_id = ?", filters.CompanyID)

	if filters.Year > 0 {
		query = query.Where("year = ?", filters.Year)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	offset := (filters.Page - 1) * filters.Limit

	var cycles []models.ReviewCycle
	if err := query.Offset(offset).Limit(filters.Limit).Order("year DESC, start_date DESC").Find(&cycles).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	return &PaginatedReviewCycles{
		Data:       cycles,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// ActivateReviewCycle activates a review cycle and creates reviews for employees
func (s *PerformanceService) ActivateReviewCycle(id uuid.UUID) (*models.ReviewCycle, error) {
	cycle, err := s.GetReviewCycleByID(id)
	if err != nil {
		return nil, err
	}

	if cycle.Status != "draft" {
		return nil, errors.New("only draft cycles can be activated")
	}

	// Get all active employees
	var employees []models.Employee
	if err := s.db.Where("company_id = ? AND employment_status = ?", cycle.CompanyID, "active").Find(&employees).Error; err != nil {
		return nil, err
	}

	// Create reviews for each employee
	for _, emp := range employees {
		review := &models.PerformanceReview{
			CompanyID:  cycle.CompanyID,
			CycleID:    cycle.ID,
			EmployeeID: emp.ID,
			TemplateID: cycle.TemplateID,
			Status:     models.ReviewStatusDraft,
		}

		// Set manager if available
		if emp.SupervisorID != nil {
			review.ManagerID = emp.SupervisorID
		}

		s.db.Create(review)
	}

	cycle.Status = "active"
	cycle.IsActive = true
	if err := s.db.Save(cycle).Error; err != nil {
		return nil, err
	}

	return cycle, nil
}

// === Performance Review Methods ===

// GetReviewByID retrieves a performance review by ID
func (s *PerformanceService) GetReviewByID(id uuid.UUID) (*models.PerformanceReview, error) {
	var review models.PerformanceReview
	err := s.db.Preload("Cycle").Preload("Employee").Preload("Responses").Preload("Goals").
		First(&review, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("review not found")
		}
		return nil, err
	}
	return &review, nil
}

// GetEmployeeReviews gets all reviews for an employee
func (s *PerformanceService) GetEmployeeReviews(employeeID uuid.UUID) ([]models.PerformanceReview, error) {
	var reviews []models.PerformanceReview
	err := s.db.Where("employee_id = ?", employeeID).
		Preload("Cycle").
		Order("created_at DESC").
		Find(&reviews).Error
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

// StartSelfReview starts the self-review process
func (s *PerformanceService) StartSelfReview(reviewID uuid.UUID) (*models.PerformanceReview, error) {
	review, err := s.GetReviewByID(reviewID)
	if err != nil {
		return nil, err
	}

	if review.Status != models.ReviewStatusDraft {
		return nil, errors.New("review already started")
	}

	now := time.Now()
	review.Status = models.ReviewStatusSelfReview
	review.SelfReviewStartedAt = &now

	if err := s.db.Save(review).Error; err != nil {
		return nil, err
	}

	return review, nil
}

// SubmitSelfReview submits the self-review
func (s *PerformanceService) SubmitSelfReview(reviewID uuid.UUID, rating float64, comments string) (*models.PerformanceReview, error) {
	review, err := s.GetReviewByID(reviewID)
	if err != nil {
		return nil, err
	}

	if review.Status != models.ReviewStatusSelfReview {
		return nil, errors.New("review is not in self-review phase")
	}

	now := time.Now()
	review.SelfRating = rating
	review.SelfComments = comments
	review.SelfReviewCompletedAt = &now
	review.Status = models.ReviewStatusManagerReview
	review.ManagerReviewStartedAt = &now

	if err := s.db.Save(review).Error; err != nil {
		return nil, err
	}

	return review, nil
}

// SubmitManagerReview submits the manager review
func (s *PerformanceService) SubmitManagerReview(reviewID uuid.UUID, rating float64, comments, overallPerformance string) (*models.PerformanceReview, error) {
	review, err := s.GetReviewByID(reviewID)
	if err != nil {
		return nil, err
	}

	if review.Status != models.ReviewStatusManagerReview {
		return nil, errors.New("review is not in manager review phase")
	}

	now := time.Now()
	review.ManagerRating = rating
	review.ManagerComments = comments
	review.OverallPerformance = overallPerformance
	review.ManagerReviewCompletedAt = &now
	review.FinalRating = rating // Can be calibrated later
	review.Status = models.ReviewStatusCompleted
	review.CompletedAt = &now

	if err := s.db.Save(review).Error; err != nil {
		return nil, err
	}

	return review, nil
}

// AcknowledgeReview employee acknowledges the review
func (s *PerformanceService) AcknowledgeReview(reviewID uuid.UUID, comments string) (*models.PerformanceReview, error) {
	review, err := s.GetReviewByID(reviewID)
	if err != nil {
		return nil, err
	}

	if review.Status != models.ReviewStatusCompleted {
		return nil, errors.New("review is not completed")
	}

	now := time.Now()
	review.EmployeeAcknowledged = true
	review.AcknowledgedAt = &now
	review.AcknowledgementComments = comments

	if err := s.db.Save(review).Error; err != nil {
		return nil, err
	}

	return review, nil
}

// === Goal Methods ===

type CreateGoalDTO struct {
	CompanyID   uuid.UUID
	EmployeeID  uuid.UUID
	ReviewID    *uuid.UUID
	CycleID     *uuid.UUID
	Title       string
	Description string
	Category    string
	Priority    string
	Measurable  string
	TargetValue float64
	Unit        string
	StartDate   *time.Time
	DueDate     *time.Time
	Weight      int
	CreatedByID *uuid.UUID
}

// CreateGoal creates a new goal
func (s *PerformanceService) CreateGoal(dto CreateGoalDTO) (*models.Goal, error) {
	if dto.Title == "" {
		return nil, errors.New("goal title is required")
	}

	goal := &models.Goal{
		CompanyID:   dto.CompanyID,
		EmployeeID:  dto.EmployeeID,
		ReviewID:    dto.ReviewID,
		CycleID:     dto.CycleID,
		Title:       dto.Title,
		Description: dto.Description,
		Category:    dto.Category,
		Priority:    dto.Priority,
		Measurable:  dto.Measurable,
		TargetValue: dto.TargetValue,
		Unit:        dto.Unit,
		StartDate:   dto.StartDate,
		DueDate:     dto.DueDate,
		Weight:      dto.Weight,
		Status:      models.GoalStatusDraft,
		CreatedByID: dto.CreatedByID,
	}

	if goal.Category == "" {
		goal.Category = "business"
	}
	if goal.Priority == "" {
		goal.Priority = "medium"
	}

	if err := s.db.Create(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

// GetGoalByID retrieves a goal by ID
func (s *PerformanceService) GetGoalByID(id uuid.UUID) (*models.Goal, error) {
	var goal models.Goal
	err := s.db.Preload("Updates").First(&goal, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("goal not found")
		}
		return nil, err
	}
	return &goal, nil
}

// GetEmployeeGoals gets all goals for an employee
func (s *PerformanceService) GetEmployeeGoals(employeeID uuid.UUID, status string) ([]models.Goal, error) {
	query := s.db.Where("employee_id = ?", employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var goals []models.Goal
	if err := query.Order("due_date ASC, priority DESC").Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
}

// UpdateGoalProgress updates progress on a goal
func (s *PerformanceService) UpdateGoalProgress(goalID uuid.UUID, progress int, newValue float64, note string, updatedByID uuid.UUID) (*models.Goal, error) {
	goal, err := s.GetGoalByID(goalID)
	if err != nil {
		return nil, err
	}

	goal.ProgressPercent = progress
	goal.CurrentValue = newValue

	// Create update record
	update := &models.GoalUpdate{
		GoalID:      goalID,
		Note:        note,
		NewProgress: progress,
		NewValue:    newValue,
		UpdatedByID: updatedByID,
	}
	s.db.Create(update)

	// Check if completed
	if progress >= 100 {
		now := time.Now()
		goal.Status = models.GoalStatusCompleted
		goal.CompletedAt = &now
	} else if goal.Status == models.GoalStatusDraft {
		goal.Status = models.GoalStatusActive
	}

	if err := s.db.Save(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

// === Feedback Methods ===

type CreateFeedbackDTO struct {
	CompanyID        uuid.UUID
	RecipientID      uuid.UUID
	GiverID          uuid.UUID
	FeedbackType     models.FeedbackType
	Subject          string
	Content          string
	Rating           int
	Skills           []string
	IsAnonymous      bool
	VisibleToManager bool
	ReviewID         *uuid.UUID
}

// CreateFeedback creates new feedback
func (s *PerformanceService) CreateFeedback(dto CreateFeedbackDTO) (*models.Feedback, error) {
	if dto.Content == "" {
		return nil, errors.New("feedback content is required")
	}

	feedback := &models.Feedback{
		CompanyID:        dto.CompanyID,
		RecipientID:      dto.RecipientID,
		GiverID:          dto.GiverID,
		FeedbackType:     dto.FeedbackType,
		Subject:          dto.Subject,
		Content:          dto.Content,
		Rating:           dto.Rating,
		Skills:           dto.Skills,
		IsAnonymous:      dto.IsAnonymous,
		VisibleToManager: dto.VisibleToManager,
		ReviewID:         dto.ReviewID,
	}

	if feedback.FeedbackType == "" {
		feedback.FeedbackType = models.FeedbackTypePositive
	}

	if err := s.db.Create(feedback).Error; err != nil {
		return nil, err
	}

	return feedback, nil
}

// GetEmployeeFeedback gets feedback for an employee
func (s *PerformanceService) GetEmployeeFeedback(employeeID uuid.UUID, feedbackType string) ([]models.Feedback, error) {
	query := s.db.Where("recipient_id = ?", employeeID)
	if feedbackType != "" {
		query = query.Where("feedback_type = ?", feedbackType)
	}

	var feedback []models.Feedback
	if err := query.Order("created_at DESC").Find(&feedback).Error; err != nil {
		return nil, err
	}
	return feedback, nil
}

// === 1:1 Meeting Methods ===

type CreateOneOnOneDTO struct {
	CompanyID       uuid.UUID
	EmployeeID      uuid.UUID
	ManagerID       uuid.UUID
	ScheduledAt     time.Time
	DurationMinutes int
	IsRecurring     bool
	RecurringPattern string
}

// CreateOneOnOne creates a 1:1 meeting
func (s *PerformanceService) CreateOneOnOne(dto CreateOneOnOneDTO) (*models.OneOnOne, error) {
	meeting := &models.OneOnOne{
		CompanyID:        dto.CompanyID,
		EmployeeID:       dto.EmployeeID,
		ManagerID:        dto.ManagerID,
		ScheduledAt:      dto.ScheduledAt,
		DurationMinutes:  dto.DurationMinutes,
		IsRecurring:      dto.IsRecurring,
		RecurringPattern: dto.RecurringPattern,
		Status:           "scheduled",
	}

	if meeting.DurationMinutes == 0 {
		meeting.DurationMinutes = 30
	}

	if err := s.db.Create(meeting).Error; err != nil {
		return nil, err
	}

	return meeting, nil
}

// GetEmployeeOneOnOnes gets 1:1 meetings for an employee
func (s *PerformanceService) GetEmployeeOneOnOnes(employeeID uuid.UUID, status string) ([]models.OneOnOne, error) {
	query := s.db.Where("employee_id = ?", employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var meetings []models.OneOnOne
	if err := query.Order("scheduled_at DESC").Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

// CompleteOneOnOne completes a 1:1 meeting
func (s *PerformanceService) CompleteOneOnOne(id uuid.UUID, notes, actionItems string, mood string, engagement int) (*models.OneOnOne, error) {
	var meeting models.OneOnOne
	if err := s.db.First(&meeting, "id = ?", id).Error; err != nil {
		return nil, errors.New("meeting not found")
	}

	meeting.Status = "completed"
	meeting.MeetingNotes = notes
	meeting.ActionItems = actionItems
	meeting.EmployeeMood = mood
	meeting.Engagement = engagement

	if err := s.db.Save(&meeting).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}
