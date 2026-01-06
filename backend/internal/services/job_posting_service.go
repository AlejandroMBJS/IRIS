/*
Package services - IRIS Payroll System Job Posting Service

==============================================================================
FILE: internal/services/job_posting_service.go
==============================================================================

DESCRIPTION:
    Provides business logic for managing job postings in the recruitment module.
    Handles CRUD operations, publishing workflow, and statistics.

USER PERSPECTIVE:
    - HR can create, edit, and publish job postings
    - Hiring managers can view applicants and statistics
    - Job postings follow a draft -> published -> closed lifecycle

DEVELOPER GUIDELINES:
    OK to modify:
      - Add new validation rules
      - Add new business logic methods
      - Extend filtering capabilities

    DO NOT modify:
      - Core CRUD operations without testing
      - Status transition logic without review

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

// JobPostingService provides business logic for job postings
type JobPostingService struct {
	db *gorm.DB
}

// NewJobPostingService creates a new JobPostingService
func NewJobPostingService(db *gorm.DB) *JobPostingService {
	return &JobPostingService{db: db}
}

// CreateJobPostingDTO contains data for creating a job posting
type CreateJobPostingDTO struct {
	CompanyID          uuid.UUID
	Title              string
	Description        string
	Requirements       string
	Responsibilities   string
	DepartmentID       *uuid.UUID
	CostCenterID       *uuid.UUID
	PositionLevel      string
	EmploymentType     models.EmploymentType
	CollarType         string
	SalaryMin          float64
	SalaryMax          float64
	SalaryCurrency     string
	SalaryFrequency    string
	ShowSalary         bool
	PositionsAvailable int
	Location           string
	IsRemote           bool
	RemoteType         models.RemoteType
	HiringManagerID    *uuid.UUID
	RecruiterID        *uuid.UUID
	ClosesAt           *time.Time
	CreatedByUserID    *uuid.UUID
}

// UpdateJobPostingDTO contains data for updating a job posting
type UpdateJobPostingDTO struct {
	Title              *string
	Description        *string
	Requirements       *string
	Responsibilities   *string
	DepartmentID       *uuid.UUID
	CostCenterID       *uuid.UUID
	PositionLevel      *string
	EmploymentType     *models.EmploymentType
	CollarType         *string
	SalaryMin          *float64
	SalaryMax          *float64
	SalaryCurrency     *string
	SalaryFrequency    *string
	ShowSalary         *bool
	PositionsAvailable *int
	Location           *string
	IsRemote           *bool
	RemoteType         *models.RemoteType
	HiringManagerID    *uuid.UUID
	RecruiterID        *uuid.UUID
	ClosesAt           *time.Time
}

// JobPostingFilters contains filters for listing job postings
type JobPostingFilters struct {
	CompanyID      uuid.UUID
	Status         string
	DepartmentID   *uuid.UUID
	EmploymentType string
	IsRemote       *bool
	Search         string
	Page           int
	Limit          int
	SortBy         string
	SortOrder      string
}

// JobPostingStats contains statistics for a job posting
type JobPostingStats struct {
	PostingID            uuid.UUID                  `json:"posting_id"`
	TotalApplications    int                        `json:"total_applications"`
	ApplicationsByStage  map[string]int             `json:"applications_by_stage"`
	AverageTimeInStage   map[string]float64         `json:"average_time_in_stage_days"`
	ConversionRates      map[string]float64         `json:"conversion_rates"`
	SourceBreakdown      map[string]int             `json:"source_breakdown"`
	AverageRating        float64                    `json:"average_rating"`
	InterviewsScheduled  int                        `json:"interviews_scheduled"`
	OffersSent           int                        `json:"offers_sent"`
	OffersAccepted       int                        `json:"offers_accepted"`
}

// PaginatedJobPostings contains paginated job posting results
type PaginatedJobPostings struct {
	Data       []models.JobPosting `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// Create creates a new job posting
func (s *JobPostingService) Create(dto CreateJobPostingDTO) (*models.JobPosting, error) {
	// Validate required fields
	if dto.Title == "" {
		return nil, errors.New("title is required")
	}
	if dto.Description == "" {
		return nil, errors.New("description is required")
	}
	if dto.EmploymentType == "" {
		return nil, errors.New("employment type is required")
	}

	// Validate salary range if provided
	if dto.SalaryMin > 0 && dto.SalaryMax > 0 {
		if dto.SalaryMin > dto.SalaryMax {
			return nil, errors.New("minimum salary cannot exceed maximum salary")
		}
	}

	// Set defaults
	if dto.SalaryCurrency == "" {
		dto.SalaryCurrency = "MXN"
	}
	if dto.PositionsAvailable <= 0 {
		dto.PositionsAvailable = 1
	}

	posting := &models.JobPosting{
		CompanyID:          dto.CompanyID,
		Title:              dto.Title,
		Description:        dto.Description,
		Requirements:       dto.Requirements,
		Responsibilities:   dto.Responsibilities,
		DepartmentID:       dto.DepartmentID,
		CostCenterID:       dto.CostCenterID,
		PositionLevel:      dto.PositionLevel,
		EmploymentType:     dto.EmploymentType,
		CollarType:         dto.CollarType,
		SalaryMin:          dto.SalaryMin,
		SalaryMax:          dto.SalaryMax,
		SalaryCurrency:     dto.SalaryCurrency,
		SalaryFrequency:    dto.SalaryFrequency,
		ShowSalary:         dto.ShowSalary,
		PositionsAvailable: dto.PositionsAvailable,
		Location:           dto.Location,
		IsRemote:           dto.IsRemote,
		RemoteType:         dto.RemoteType,
		HiringManagerID:    dto.HiringManagerID,
		RecruiterID:        dto.RecruiterID,
		ClosesAt:           dto.ClosesAt,
		Status:             models.JobPostingStatusDraft,
		CreatedByID:        dto.CreatedByUserID,
	}

	if err := s.db.Create(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// GetByID retrieves a job posting by ID
func (s *JobPostingService) GetByID(id uuid.UUID) (*models.JobPosting, error) {
	var posting models.JobPosting
	err := s.db.Preload("HiringManager").
		Preload("Recruiter").
		Preload("CreatedBy").
		First(&posting, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job posting not found")
		}
		return nil, err
	}
	return &posting, nil
}

// Update updates a job posting
func (s *JobPostingService) Update(id uuid.UUID, dto UpdateJobPostingDTO) (*models.JobPosting, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Cannot update published/closed postings (except certain fields)
	if posting.Status == models.JobPostingStatusClosed || posting.Status == models.JobPostingStatusFilled {
		return nil, errors.New("cannot update closed or filled job posting")
	}

	// Apply updates
	if dto.Title != nil {
		posting.Title = *dto.Title
	}
	if dto.Description != nil {
		posting.Description = *dto.Description
	}
	if dto.Requirements != nil {
		posting.Requirements = *dto.Requirements
	}
	if dto.Responsibilities != nil {
		posting.Responsibilities = *dto.Responsibilities
	}
	if dto.DepartmentID != nil {
		posting.DepartmentID = dto.DepartmentID
	}
	if dto.CostCenterID != nil {
		posting.CostCenterID = dto.CostCenterID
	}
	if dto.PositionLevel != nil {
		posting.PositionLevel = *dto.PositionLevel
	}
	if dto.EmploymentType != nil {
		posting.EmploymentType = *dto.EmploymentType
	}
	if dto.CollarType != nil {
		posting.CollarType = *dto.CollarType
	}
	if dto.SalaryMin != nil {
		posting.SalaryMin = *dto.SalaryMin
	}
	if dto.SalaryMax != nil {
		posting.SalaryMax = *dto.SalaryMax
	}
	if dto.SalaryCurrency != nil {
		posting.SalaryCurrency = *dto.SalaryCurrency
	}
	if dto.SalaryFrequency != nil {
		posting.SalaryFrequency = *dto.SalaryFrequency
	}
	if dto.ShowSalary != nil {
		posting.ShowSalary = *dto.ShowSalary
	}
	if dto.PositionsAvailable != nil {
		posting.PositionsAvailable = *dto.PositionsAvailable
	}
	if dto.Location != nil {
		posting.Location = *dto.Location
	}
	if dto.IsRemote != nil {
		posting.IsRemote = *dto.IsRemote
	}
	if dto.RemoteType != nil {
		posting.RemoteType = *dto.RemoteType
	}
	if dto.HiringManagerID != nil {
		posting.HiringManagerID = dto.HiringManagerID
	}
	if dto.RecruiterID != nil {
		posting.RecruiterID = dto.RecruiterID
	}
	if dto.ClosesAt != nil {
		posting.ClosesAt = dto.ClosesAt
	}

	// Validate salary range
	if posting.SalaryMin > 0 && posting.SalaryMax > 0 && posting.SalaryMin > posting.SalaryMax {
		return nil, errors.New("minimum salary cannot exceed maximum salary")
	}

	if err := s.db.Save(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// Delete soft-deletes a job posting
func (s *JobPostingService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.JobPosting{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("job posting not found")
	}
	return nil
}

// List retrieves job postings with filters and pagination
func (s *JobPostingService) List(filters JobPostingFilters) (*PaginatedJobPostings, error) {
	query := s.db.Model(&models.JobPosting{}).Where("company_id = ?", filters.CompanyID)

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.DepartmentID != nil {
		query = query.Where("department_id = ?", *filters.DepartmentID)
	}
	if filters.EmploymentType != "" {
		query = query.Where("employment_type = ?", filters.EmploymentType)
	}
	if filters.IsRemote != nil {
		query = query.Where("is_remote = ?", *filters.IsRemote)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchTerm, searchTerm)
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
	sortBy := "created_at"
	sortOrder := "DESC"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Execute query
	var postings []models.JobPosting
	if err := query.Preload("HiringManager").Preload("Recruiter").Find(&postings).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedJobPostings{
		Data:       postings,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// Publish publishes a job posting (makes it live)
func (s *JobPostingService) Publish(id uuid.UUID) (*models.JobPosting, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate posting is ready to publish
	validationErrors := s.validateForPublish(posting)
	if len(validationErrors) > 0 {
		return nil, errors.New("job posting is not ready to publish: " + validationErrors[0])
	}

	// Update status
	now := time.Now()
	posting.Status = models.JobPostingStatusPublished
	posting.PublishedAt = &now

	if err := s.db.Save(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// validateForPublish checks if a job posting is ready to be published
func (s *JobPostingService) validateForPublish(posting *models.JobPosting) []string {
	var errs []string

	if posting.Title == "" {
		errs = append(errs, "title is required")
	}
	if posting.Description == "" {
		errs = append(errs, "description is required")
	}
	if posting.EmploymentType == "" {
		errs = append(errs, "employment type is required")
	}

	return errs
}

// Pause pauses a published job posting
func (s *JobPostingService) Pause(id uuid.UUID) (*models.JobPosting, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if posting.Status != models.JobPostingStatusPublished {
		return nil, errors.New("can only pause published job postings")
	}

	posting.Status = models.JobPostingStatusPaused

	if err := s.db.Save(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// Resume resumes a paused job posting
func (s *JobPostingService) Resume(id uuid.UUID) (*models.JobPosting, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if posting.Status != models.JobPostingStatusPaused {
		return nil, errors.New("can only resume paused job postings")
	}

	posting.Status = models.JobPostingStatusPublished

	if err := s.db.Save(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// Close closes a job posting
func (s *JobPostingService) Close(id uuid.UUID, reason string) (*models.JobPosting, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if posting.Status == models.JobPostingStatusClosed || posting.Status == models.JobPostingStatusFilled {
		return nil, errors.New("job posting is already closed")
	}

	now := time.Now()
	posting.Status = models.JobPostingStatusClosed
	posting.ClosedAt = &now
	posting.CloseReason = reason

	if err := s.db.Save(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// MarkAsFilled marks a job posting as filled
func (s *JobPostingService) MarkAsFilled(id uuid.UUID) (*models.JobPosting, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	posting.Status = models.JobPostingStatusFilled
	posting.ClosedAt = &now

	if err := s.db.Save(posting).Error; err != nil {
		return nil, err
	}

	return posting, nil
}

// GetStats returns statistics for a job posting
func (s *JobPostingService) GetStats(id uuid.UUID) (*JobPostingStats, error) {
	posting, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	stats := &JobPostingStats{
		PostingID:           posting.ID,
		ApplicationsByStage: make(map[string]int),
		AverageTimeInStage:  make(map[string]float64),
		ConversionRates:     make(map[string]float64),
		SourceBreakdown:     make(map[string]int),
	}

	// Count total applications
	var totalApps int64
	s.db.Model(&models.Application{}).Where("job_posting_id = ?", id).Count(&totalApps)
	stats.TotalApplications = int(totalApps)

	// Count applications by stage
	var stageCounts []struct {
		Stage string
		Count int
	}
	s.db.Model(&models.Application{}).
		Select("stage, count(*) as count").
		Where("job_posting_id = ?", id).
		Group("stage").
		Scan(&stageCounts)

	for _, sc := range stageCounts {
		stats.ApplicationsByStage[sc.Stage] = sc.Count
	}

	// Count interviews scheduled
	var interviewCount int64
	s.db.Model(&models.Interview{}).
		Joins("JOIN applications ON applications.id = interviews.application_id").
		Where("applications.job_posting_id = ?", id).
		Count(&interviewCount)
	stats.InterviewsScheduled = int(interviewCount)

	// Count offers
	var offersSentCount int64
	s.db.Model(&models.Offer{}).
		Joins("JOIN applications ON applications.id = offers.application_id").
		Where("applications.job_posting_id = ?", id).
		Count(&offersSentCount)
	stats.OffersSent = int(offersSentCount)

	// Count accepted offers
	var offersAcceptedCount int64
	s.db.Model(&models.Offer{}).
		Joins("JOIN applications ON applications.id = offers.application_id").
		Where("applications.job_posting_id = ? AND offers.status = ?", id, models.OfferStatusAccepted).
		Count(&offersAcceptedCount)
	stats.OffersAccepted = int(offersAcceptedCount)

	// Get source breakdown
	var sourceCounts []struct {
		Source string
		Count  int
	}
	s.db.Model(&models.Application{}).
		Joins("JOIN candidates ON candidates.id = applications.candidate_id").
		Select("candidates.source, count(*) as count").
		Where("applications.job_posting_id = ?", id).
		Group("candidates.source").
		Scan(&sourceCounts)

	for _, sc := range sourceCounts {
		if sc.Source == "" {
			stats.SourceBreakdown["unknown"] = sc.Count
		} else {
			stats.SourceBreakdown[sc.Source] = sc.Count
		}
	}

	// Calculate average rating from interviews
	var avgRating float64
	s.db.Model(&models.Interview{}).
		Joins("JOIN applications ON applications.id = interviews.application_id").
		Where("applications.job_posting_id = ? AND interviews.rating > 0", id).
		Select("AVG(interviews.rating)").
		Scan(&avgRating)
	stats.AverageRating = avgRating

	return stats, nil
}

// GetPublicPostings returns published job postings for public job board
func (s *JobPostingService) GetPublicPostings(companyID uuid.UUID, filters JobPostingFilters) (*PaginatedJobPostings, error) {
	filters.CompanyID = companyID
	filters.Status = string(models.JobPostingStatusPublished)
	return s.List(filters)
}

// IncrementPositionsFilled increments the positions filled count
func (s *JobPostingService) IncrementPositionsFilled(id uuid.UUID) error {
	result := s.db.Model(&models.JobPosting{}).
		Where("id = ?", id).
		UpdateColumn("positions_filled", gorm.Expr("positions_filled + 1"))

	if result.Error != nil {
		return result.Error
	}

	// Check if all positions are filled
	var posting models.JobPosting
	if err := s.db.First(&posting, "id = ?", id).Error; err != nil {
		return err
	}

	if posting.PositionsFilled >= posting.PositionsAvailable {
		_, err := s.MarkAsFilled(id)
		return err
	}

	return nil
}
