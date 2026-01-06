/*
Package services - IRIS Payroll System Candidate Service

==============================================================================
FILE: internal/services/candidate_service.go
==============================================================================

DESCRIPTION:
    Provides business logic for managing candidates in the recruitment module.
    Handles candidate creation, updates, and search functionality.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// CandidateService provides business logic for candidates
type CandidateService struct {
	db *gorm.DB
}

// NewCandidateService creates a new CandidateService
func NewCandidateService(db *gorm.DB) *CandidateService {
	return &CandidateService{db: db}
}

// CreateCandidateDTO contains data for creating a candidate
type CreateCandidateDTO struct {
	CompanyID            uuid.UUID
	FirstName            string
	LastName             string
	Email                string
	Phone                string
	CurrentTitle         string
	CurrentCompany       string
	YearsExperience      int
	LinkedInURL          string
	PortfolioURL         string
	ResumeFileID         *uuid.UUID
	CoverLetter          string
	Source               string
	SourceDetails        string
	ReferredByEmployeeID *uuid.UUID
	Tags                 []string
}

// UpdateCandidateDTO contains data for updating a candidate
type UpdateCandidateDTO struct {
	FirstName            *string
	LastName             *string
	Email                *string
	Phone                *string
	CurrentTitle         *string
	CurrentCompany       *string
	YearsExperience      *int
	LinkedInURL          *string
	PortfolioURL         *string
	ResumeFileID         *uuid.UUID
	CoverLetter          *string
	Source               *string
	SourceDetails        *string
	ReferredByEmployeeID *uuid.UUID
	Tags                 []string
	Status               *models.CandidateStatus
}

// CandidateFilters contains filters for listing candidates
type CandidateFilters struct {
	CompanyID uuid.UUID
	Status    string
	Source    string
	Search    string
	Tags      []string
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}

// PaginatedCandidates contains paginated candidate results
type PaginatedCandidates struct {
	Data       []models.Candidate `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// Create creates a new candidate
func (s *CandidateService) Create(dto CreateCandidateDTO) (*models.Candidate, error) {
	// Validate required fields
	if dto.FirstName == "" {
		return nil, errors.New("first name is required")
	}
	if dto.LastName == "" {
		return nil, errors.New("last name is required")
	}
	if dto.Email == "" {
		return nil, errors.New("email is required")
	}

	// Check for duplicate email in same company
	var existing models.Candidate
	err := s.db.Where("company_id = ? AND email = ?", dto.CompanyID, dto.Email).First(&existing).Error
	if err == nil {
		return nil, errors.New("candidate with this email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	candidate := &models.Candidate{
		CompanyID:            dto.CompanyID,
		FirstName:            dto.FirstName,
		LastName:             dto.LastName,
		Email:                dto.Email,
		Phone:                dto.Phone,
		CurrentTitle:         dto.CurrentTitle,
		CurrentCompany:       dto.CurrentCompany,
		YearsExperience:      dto.YearsExperience,
		LinkedInURL:          dto.LinkedInURL,
		PortfolioURL:         dto.PortfolioURL,
		ResumeFileID:         dto.ResumeFileID,
		CoverLetter:          dto.CoverLetter,
		Source:               dto.Source,
		SourceDetails:        dto.SourceDetails,
		ReferredByEmployeeID: dto.ReferredByEmployeeID,
		Tags:                 pq.StringArray(dto.Tags),
		Status:               models.CandidateStatusActive,
	}

	if err := s.db.Create(candidate).Error; err != nil {
		return nil, err
	}

	return candidate, nil
}

// GetByID retrieves a candidate by ID
func (s *CandidateService) GetByID(id uuid.UUID) (*models.Candidate, error) {
	var candidate models.Candidate
	err := s.db.Preload("ReferredByEmployee").
		Preload("Applications").
		First(&candidate, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("candidate not found")
		}
		return nil, err
	}
	return &candidate, nil
}

// GetByEmail retrieves a candidate by email within a company
func (s *CandidateService) GetByEmail(companyID uuid.UUID, email string) (*models.Candidate, error) {
	var candidate models.Candidate
	err := s.db.Where("company_id = ? AND email = ?", companyID, email).First(&candidate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &candidate, nil
}

// Update updates a candidate
func (s *CandidateService) Update(id uuid.UUID, dto UpdateCandidateDTO) (*models.Candidate, error) {
	candidate, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if dto.FirstName != nil {
		candidate.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		candidate.LastName = *dto.LastName
	}
	if dto.Email != nil {
		// Check for duplicate email
		var existing models.Candidate
		err := s.db.Where("company_id = ? AND email = ? AND id != ?", candidate.CompanyID, *dto.Email, id).First(&existing).Error
		if err == nil {
			return nil, errors.New("candidate with this email already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		candidate.Email = *dto.Email
	}
	if dto.Phone != nil {
		candidate.Phone = *dto.Phone
	}
	if dto.CurrentTitle != nil {
		candidate.CurrentTitle = *dto.CurrentTitle
	}
	if dto.CurrentCompany != nil {
		candidate.CurrentCompany = *dto.CurrentCompany
	}
	if dto.YearsExperience != nil {
		candidate.YearsExperience = *dto.YearsExperience
	}
	if dto.LinkedInURL != nil {
		candidate.LinkedInURL = *dto.LinkedInURL
	}
	if dto.PortfolioURL != nil {
		candidate.PortfolioURL = *dto.PortfolioURL
	}
	if dto.ResumeFileID != nil {
		candidate.ResumeFileID = dto.ResumeFileID
	}
	if dto.CoverLetter != nil {
		candidate.CoverLetter = *dto.CoverLetter
	}
	if dto.Source != nil {
		candidate.Source = *dto.Source
	}
	if dto.SourceDetails != nil {
		candidate.SourceDetails = *dto.SourceDetails
	}
	if dto.ReferredByEmployeeID != nil {
		candidate.ReferredByEmployeeID = dto.ReferredByEmployeeID
	}
	if dto.Tags != nil {
		candidate.Tags = pq.StringArray(dto.Tags)
	}
	if dto.Status != nil {
		candidate.Status = *dto.Status
	}

	if err := s.db.Save(candidate).Error; err != nil {
		return nil, err
	}

	return candidate, nil
}

// Delete soft-deletes a candidate
func (s *CandidateService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Candidate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("candidate not found")
	}
	return nil
}

// List retrieves candidates with filters and pagination
func (s *CandidateService) List(filters CandidateFilters) (*PaginatedCandidates, error) {
	query := s.db.Model(&models.Candidate{}).Where("company_id = ?", filters.CompanyID)

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Source != "" {
		query = query.Where("source = ?", filters.Source)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where(
			"first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR current_title LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}
	if len(filters.Tags) > 0 {
		query = query.Where("tags && ?", pq.StringArray(filters.Tags))
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
	var candidates []models.Candidate
	if err := query.Find(&candidates).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedCandidates{
		Data:       candidates,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// MarkAsHired marks a candidate as hired
func (s *CandidateService) MarkAsHired(id uuid.UUID) (*models.Candidate, error) {
	candidate, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	candidate.Status = models.CandidateStatusHired

	if err := s.db.Save(candidate).Error; err != nil {
		return nil, err
	}

	return candidate, nil
}

// MarkAsRejected marks a candidate as rejected
func (s *CandidateService) MarkAsRejected(id uuid.UUID) (*models.Candidate, error) {
	candidate, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	candidate.Status = models.CandidateStatusRejected

	if err := s.db.Save(candidate).Error; err != nil {
		return nil, err
	}

	return candidate, nil
}

// GetOrCreate gets an existing candidate by email or creates a new one
func (s *CandidateService) GetOrCreate(dto CreateCandidateDTO) (*models.Candidate, bool, error) {
	existing, err := s.GetByEmail(dto.CompanyID, dto.Email)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	candidate, err := s.Create(dto)
	if err != nil {
		return nil, false, err
	}

	return candidate, true, nil
}
