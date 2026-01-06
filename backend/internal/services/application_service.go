/*
Package services - IRIS Payroll System Application Service

==============================================================================
FILE: internal/services/application_service.go
==============================================================================

DESCRIPTION:
    Provides business logic for managing job applications in the recruitment module.
    Handles the application pipeline, stage transitions, and hiring workflow.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationService provides business logic for applications
type ApplicationService struct {
	db *gorm.DB
}

// NewApplicationService creates a new ApplicationService
func NewApplicationService(db *gorm.DB) *ApplicationService {
	return &ApplicationService{db: db}
}

// CreateApplicationDTO contains data for creating an application
type CreateApplicationDTO struct {
	CompanyID      uuid.UUID
	CandidateID    uuid.UUID
	JobPostingID   uuid.UUID
	ExpectedSalary float64
}

// UpdateApplicationDTO contains data for updating an application
type UpdateApplicationDTO struct {
	ExpectedSalary *float64
	OfferedSalary  *float64
	StartDate      *time.Time
}

// ApplicationFilters contains filters for listing applications
type ApplicationFilters struct {
	CompanyID    uuid.UUID
	JobPostingID *uuid.UUID
	CandidateID  *uuid.UUID
	Stage        string
	Status       string
	Search       string
	Page         int
	Limit        int
	SortBy       string
	SortOrder    string
}

// PaginatedApplications contains paginated application results
type PaginatedApplications struct {
	Data       []models.Application `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// ScreeningDTO contains data for screening an application
type ScreeningDTO struct {
	Score  int
	Notes  string
	UserID uuid.UUID
}

// RejectionDTO contains data for rejecting an application
type RejectionDTO struct {
	Reason string
	Notes  string
}

// ApplicationStats contains statistics for applications
type ApplicationStats struct {
	TotalApplications int64            `json:"total_applications"`
	ByStage           map[string]int64 `json:"by_stage"`
	ByStatus          map[string]int64 `json:"by_status"`
	AverageRating     float64          `json:"average_rating"`
}

// Create creates a new application
func (s *ApplicationService) Create(dto CreateApplicationDTO) (*models.Application, error) {
	// Verify job posting exists and is open
	var jobPosting models.JobPosting
	err := s.db.First(&jobPosting, "id = ?", dto.JobPostingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job posting not found")
		}
		return nil, err
	}

	if !jobPosting.IsOpen() {
		return nil, errors.New("job posting is not accepting applications")
	}

	// Verify candidate exists
	var candidate models.Candidate
	err = s.db.First(&candidate, "id = ?", dto.CandidateID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("candidate not found")
		}
		return nil, err
	}

	// Check for duplicate application
	var existing models.Application
	err = s.db.Where("candidate_id = ? AND job_posting_id = ?", dto.CandidateID, dto.JobPostingID).First(&existing).Error
	if err == nil {
		return nil, errors.New("candidate has already applied for this position")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()
	application := &models.Application{
		CompanyID:      dto.CompanyID,
		CandidateID:    dto.CandidateID,
		JobPostingID:   dto.JobPostingID,
		Stage:          models.ApplicationStageNew,
		Status:         models.ApplicationStatusActive,
		ExpectedSalary: dto.ExpectedSalary,
		AppliedAt:      now,
		StageEnteredAt: now,
	}

	if err := s.db.Create(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// GetByID retrieves an application by ID with related data
func (s *ApplicationService) GetByID(id uuid.UUID) (*models.Application, error) {
	var application models.Application
	err := s.db.Preload("Candidate").
		Preload("JobPosting").
		Preload("ScreenedBy").
		Preload("Interviews").
		Preload("Interviews.Interviewer").
		Preload("Offers").
		First(&application, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("application not found")
		}
		return nil, err
	}
	return &application, nil
}

// Update updates an application
func (s *ApplicationService) Update(id uuid.UUID, dto UpdateApplicationDTO) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.ExpectedSalary != nil {
		application.ExpectedSalary = *dto.ExpectedSalary
	}
	if dto.OfferedSalary != nil {
		application.OfferedSalary = *dto.OfferedSalary
	}
	if dto.StartDate != nil {
		application.StartDate = dto.StartDate
	}

	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// Delete removes an application
func (s *ApplicationService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Application{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("application not found")
	}
	return nil
}

// List retrieves applications with filters and pagination
func (s *ApplicationService) List(filters ApplicationFilters) (*PaginatedApplications, error) {
	query := s.db.Model(&models.Application{}).Where("company_id = ?", filters.CompanyID)

	// Apply filters
	if filters.JobPostingID != nil {
		query = query.Where("job_posting_id = ?", *filters.JobPostingID)
	}
	if filters.CandidateID != nil {
		query = query.Where("candidate_id = ?", *filters.CandidateID)
	}
	if filters.Stage != "" {
		query = query.Where("stage = ?", filters.Stage)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		// Join with candidates to search by name
		query = query.Joins("LEFT JOIN candidates ON candidates.id = applications.candidate_id").
			Where("candidates.first_name LIKE ? OR candidates.last_name LIKE ? OR candidates.email LIKE ?",
				"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Apply sorting
	sortBy := "applied_at"
	sortOrder := "DESC"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Execute query with preloads
	var applications []models.Application
	if err := query.Preload("Candidate").Preload("JobPosting").Find(&applications).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedApplications{
		Data:       applications,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// MoveToStage moves an application to a new pipeline stage
func (s *ApplicationService) MoveToStage(id uuid.UUID, stage models.ApplicationStage) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if application.Status != models.ApplicationStatusActive {
		return nil, errors.New("can only change stage for active applications")
	}

	// Validate stage transition
	if !s.isValidStageTransition(application.Stage, stage) {
		return nil, errors.New("invalid stage transition")
	}

	application.Stage = stage
	application.StageEnteredAt = time.Now()

	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// isValidStageTransition checks if a stage transition is valid
func (s *ApplicationService) isValidStageTransition(from, to models.ApplicationStage) bool {
	// Define valid transitions
	validTransitions := map[models.ApplicationStage][]models.ApplicationStage{
		models.ApplicationStageNew: {
			models.ApplicationStageScreening,
			models.ApplicationStageRejected,
		},
		models.ApplicationStageScreening: {
			models.ApplicationStagePhoneInterview,
			models.ApplicationStageRejected,
		},
		models.ApplicationStagePhoneInterview: {
			models.ApplicationStageTechnicalInterview,
			models.ApplicationStageOnsiteInterview,
			models.ApplicationStageRejected,
		},
		models.ApplicationStageTechnicalInterview: {
			models.ApplicationStageOnsiteInterview,
			models.ApplicationStageOffer,
			models.ApplicationStageRejected,
		},
		models.ApplicationStageOnsiteInterview: {
			models.ApplicationStageOffer,
			models.ApplicationStageRejected,
		},
		models.ApplicationStageOffer: {
			models.ApplicationStageHired,
			models.ApplicationStageRejected,
		},
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedStage := range allowed {
		if allowedStage == to {
			return true
		}
	}

	return false
}

// Screen records screening information for an application
func (s *ApplicationService) Screen(id uuid.UUID, dto ScreeningDTO) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Score < 1 || dto.Score > 5 {
		return nil, errors.New("screening score must be between 1 and 5")
	}

	now := time.Now()
	application.ScreeningScore = dto.Score
	application.ScreeningNotes = dto.Notes
	application.ScreenedByID = &dto.UserID
	application.ScreenedAt = &now

	// Automatically move to screening stage if still new
	if application.Stage == models.ApplicationStageNew {
		application.Stage = models.ApplicationStageScreening
		application.StageEnteredAt = now
	}

	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// Reject rejects an application
func (s *ApplicationService) Reject(id uuid.UUID, dto RejectionDTO) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if application.Status != models.ApplicationStatusActive {
		return nil, errors.New("application is not active")
	}

	application.Stage = models.ApplicationStageRejected
	application.Status = models.ApplicationStatusRejected
	application.RejectionReason = dto.Reason
	application.RejectionNotes = dto.Notes
	application.StageEnteredAt = time.Now()

	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// Withdraw marks an application as withdrawn by the candidate
func (s *ApplicationService) Withdraw(id uuid.UUID) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if application.Status != models.ApplicationStatusActive {
		return nil, errors.New("application is not active")
	}

	application.Status = models.ApplicationStatusWithdrawn

	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// Hire marks an application as hired and updates related records
func (s *ApplicationService) Hire(id uuid.UUID, startDate time.Time) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if application.Status != models.ApplicationStatusActive {
		return nil, errors.New("application is not active")
	}

	if application.Stage != models.ApplicationStageOffer {
		return nil, errors.New("application must be at offer stage to hire")
	}

	now := time.Now()
	application.Stage = models.ApplicationStageHired
	application.Status = models.ApplicationStatusHired
	application.HiredAt = &now
	application.StartDate = &startDate
	application.StageEnteredAt = now

	// Use transaction to update application, candidate, and job posting
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(application).Error; err != nil {
			return err
		}

		// Update candidate status
		if err := tx.Model(&models.Candidate{}).
			Where("id = ?", application.CandidateID).
			Update("status", models.CandidateStatusHired).Error; err != nil {
			return err
		}

		// Increment positions filled on job posting
		if err := tx.Model(&models.JobPosting{}).
			Where("id = ?", application.JobPostingID).
			UpdateColumn("positions_filled", gorm.Expr("positions_filled + 1")).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return application, nil
}

// GetStats returns statistics for applications of a job posting
func (s *ApplicationService) GetStats(jobPostingID uuid.UUID) (*ApplicationStats, error) {
	stats := &ApplicationStats{
		ByStage:  make(map[string]int64),
		ByStatus: make(map[string]int64),
	}

	// Total count
	err := s.db.Model(&models.Application{}).
		Where("job_posting_id = ?", jobPostingID).
		Count(&stats.TotalApplications).Error
	if err != nil {
		return nil, err
	}

	// Count by stage
	var stageResults []struct {
		Stage string
		Count int64
	}
	err = s.db.Model(&models.Application{}).
		Select("stage, count(*) as count").
		Where("job_posting_id = ?", jobPostingID).
		Group("stage").
		Scan(&stageResults).Error
	if err != nil {
		return nil, err
	}
	for _, r := range stageResults {
		stats.ByStage[r.Stage] = r.Count
	}

	// Count by status
	var statusResults []struct {
		Status string
		Count  int64
	}
	err = s.db.Model(&models.Application{}).
		Select("status, count(*) as count").
		Where("job_posting_id = ?", jobPostingID).
		Group("status").
		Scan(&statusResults).Error
	if err != nil {
		return nil, err
	}
	for _, r := range statusResults {
		stats.ByStatus[r.Status] = r.Count
	}

	// Average rating
	err = s.db.Model(&models.Application{}).
		Select("COALESCE(AVG(overall_rating), 0)").
		Where("job_posting_id = ? AND overall_rating > 0", jobPostingID).
		Scan(&stats.AverageRating).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// UpdateOverallRating recalculates the overall rating from interview scores
func (s *ApplicationService) UpdateOverallRating(id uuid.UUID) (*models.Application, error) {
	application, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Calculate average from completed interviews
	var avgRating float64
	err = s.db.Model(&models.Interview{}).
		Select("COALESCE(AVG(rating), 0)").
		Where("application_id = ? AND status = ? AND rating > 0",
			id, models.InterviewStatusCompleted).
		Scan(&avgRating).Error
	if err != nil {
		return nil, err
	}

	application.OverallRating = avgRating

	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// GetByJobPostingAndCandidate retrieves an application by job posting and candidate
func (s *ApplicationService) GetByJobPostingAndCandidate(jobPostingID, candidateID uuid.UUID) (*models.Application, error) {
	var application models.Application
	err := s.db.Where("job_posting_id = ? AND candidate_id = ?", jobPostingID, candidateID).
		Preload("Candidate").
		Preload("JobPosting").
		First(&application).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &application, nil
}
