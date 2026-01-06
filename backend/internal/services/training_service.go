/*
Package services - IRIS Training/LMS Service

==============================================================================
FILE: internal/services/training_service.go
==============================================================================

DESCRIPTION:
    Business logic for Training and Learning Management System including
    courses, enrollments, progress tracking, and certifications.

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

// TrainingService provides business logic for training/LMS
type TrainingService struct {
	db *gorm.DB
}

// NewTrainingService creates a new TrainingService
func NewTrainingService(db *gorm.DB) *TrainingService {
	return &TrainingService{db: db}
}

// === Course DTOs ===

// CreateCourseDTO contains data for creating a course
type CreateCourseDTO struct {
	CompanyID       uuid.UUID
	CategoryID      *uuid.UUID
	Title           string
	Code            string
	Description     string
	Objectives      string
	CourseType      models.CourseType
	Level           string
	Tags            []string
	DurationMinutes int
	EstimatedWeeks  int
	PassingScore    int
	MaxAttempts     int
	ThumbnailURL    string
	InstructorID    *uuid.UUID
	InstructorName  string
	AllowSelfEnroll bool
	EnrollmentLimit int
	CreatedByID     *uuid.UUID
}

// UpdateCourseDTO contains data for updating a course
type UpdateCourseDTO struct {
	CategoryID      *uuid.UUID
	Title           *string
	Code            *string
	Description     *string
	Objectives      *string
	CourseType      *models.CourseType
	Level           *string
	Tags            []string
	DurationMinutes *int
	EstimatedWeeks  *int
	PassingScore    *int
	MaxAttempts     *int
	ThumbnailURL    *string
	InstructorID    *uuid.UUID
	InstructorName  *string
	AllowSelfEnroll *bool
	EnrollmentLimit *int
}

// CourseFilters contains filters for listing courses
type CourseFilters struct {
	CompanyID  uuid.UUID
	CategoryID *uuid.UUID
	CourseType string
	Status     string
	Level      string
	Search     string
	Page       int
	Limit      int
}

// PaginatedCourses contains paginated course results
type PaginatedCourses struct {
	Data       []models.Course `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// === Course Methods ===

// CreateCourse creates a new course
func (s *TrainingService) CreateCourse(dto CreateCourseDTO) (*models.Course, error) {
	if dto.Title == "" {
		return nil, errors.New("course title is required")
	}

	course := &models.Course{
		CompanyID:       dto.CompanyID,
		CategoryID:      dto.CategoryID,
		Title:           dto.Title,
		Code:            dto.Code,
		Description:     dto.Description,
		Objectives:      dto.Objectives,
		CourseType:      dto.CourseType,
		Level:           dto.Level,
		Tags:            dto.Tags,
		DurationMinutes: dto.DurationMinutes,
		EstimatedWeeks:  dto.EstimatedWeeks,
		PassingScore:    dto.PassingScore,
		MaxAttempts:     dto.MaxAttempts,
		ThumbnailURL:    dto.ThumbnailURL,
		InstructorID:    dto.InstructorID,
		InstructorName:  dto.InstructorName,
		AllowSelfEnroll: dto.AllowSelfEnroll,
		EnrollmentLimit: dto.EnrollmentLimit,
		Status:          models.CourseStatusDraft,
		CreatedByID:     dto.CreatedByID,
	}

	if course.PassingScore <= 0 {
		course.PassingScore = 70
	}
	if course.MaxAttempts <= 0 {
		course.MaxAttempts = 3
	}
	if course.CourseType == "" {
		course.CourseType = models.CourseTypeSelfPaced
	}

	if err := s.db.Create(course).Error; err != nil {
		return nil, err
	}

	return course, nil
}

// GetCourseByID retrieves a course by ID with modules and content
func (s *TrainingService) GetCourseByID(id uuid.UUID) (*models.Course, error) {
	var course models.Course
	err := s.db.Preload("Modules", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Preload("Modules.Contents", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Preload("Category").First(&course, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("course not found")
		}
		return nil, err
	}
	return &course, nil
}

// UpdateCourse updates a course
func (s *TrainingService) UpdateCourse(id uuid.UUID, dto UpdateCourseDTO) (*models.Course, error) {
	course, err := s.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	if dto.CategoryID != nil {
		course.CategoryID = dto.CategoryID
	}
	if dto.Title != nil {
		course.Title = *dto.Title
	}
	if dto.Code != nil {
		course.Code = *dto.Code
	}
	if dto.Description != nil {
		course.Description = *dto.Description
	}
	if dto.Objectives != nil {
		course.Objectives = *dto.Objectives
	}
	if dto.CourseType != nil {
		course.CourseType = *dto.CourseType
	}
	if dto.Level != nil {
		course.Level = *dto.Level
	}
	if dto.Tags != nil {
		course.Tags = dto.Tags
	}
	if dto.DurationMinutes != nil {
		course.DurationMinutes = *dto.DurationMinutes
	}
	if dto.EstimatedWeeks != nil {
		course.EstimatedWeeks = *dto.EstimatedWeeks
	}
	if dto.PassingScore != nil {
		course.PassingScore = *dto.PassingScore
	}
	if dto.MaxAttempts != nil {
		course.MaxAttempts = *dto.MaxAttempts
	}
	if dto.ThumbnailURL != nil {
		course.ThumbnailURL = *dto.ThumbnailURL
	}
	if dto.InstructorID != nil {
		course.InstructorID = dto.InstructorID
	}
	if dto.InstructorName != nil {
		course.InstructorName = *dto.InstructorName
	}
	if dto.AllowSelfEnroll != nil {
		course.AllowSelfEnroll = *dto.AllowSelfEnroll
	}
	if dto.EnrollmentLimit != nil {
		course.EnrollmentLimit = *dto.EnrollmentLimit
	}

	if err := s.db.Save(course).Error; err != nil {
		return nil, err
	}

	return course, nil
}

// DeleteCourse soft-deletes a course
func (s *TrainingService) DeleteCourse(id uuid.UUID) error {
	result := s.db.Delete(&models.Course{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("course not found")
	}
	return nil
}

// ListCourses retrieves courses with filters
func (s *TrainingService) ListCourses(filters CourseFilters) (*PaginatedCourses, error) {
	query := s.db.Model(&models.Course{}).Where("company_id = ?", filters.CompanyID)

	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.CourseType != "" {
		query = query.Where("course_type = ?", filters.CourseType)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Level != "" {
		query = query.Where("level = ?", filters.Level)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR code LIKE ?", searchTerm, searchTerm, searchTerm)
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
	query = query.Offset(offset).Limit(filters.Limit)
	query = query.Order("created_at DESC")

	var courses []models.Course
	if err := query.Preload("Category").Find(&courses).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	return &PaginatedCourses{
		Data:       courses,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// PublishCourse publishes a course
func (s *TrainingService) PublishCourse(id uuid.UUID) (*models.Course, error) {
	course, err := s.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	if course.Status != models.CourseStatusDraft {
		return nil, errors.New("only draft courses can be published")
	}

	now := time.Now()
	course.Status = models.CourseStatusPublished
	course.PublishedAt = &now

	if err := s.db.Save(course).Error; err != nil {
		return nil, err
	}

	return course, nil
}

// ArchiveCourse archives a course
func (s *TrainingService) ArchiveCourse(id uuid.UUID) (*models.Course, error) {
	course, err := s.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	course.Status = models.CourseStatusArchived
	course.ArchivedAt = &now

	if err := s.db.Save(course).Error; err != nil {
		return nil, err
	}

	return course, nil
}

// === Enrollment Methods ===

// EnrollEmployeeDTO contains data for enrolling an employee
type EnrollEmployeeDTO struct {
	CompanyID      uuid.UUID
	CourseID       uuid.UUID
	EmployeeID     uuid.UUID
	EnrolledByID   *uuid.UUID
	EnrollmentType string // self, assigned, required
	DueDate        *time.Time
}

// EnrollEmployee enrolls an employee in a course
func (s *TrainingService) EnrollEmployee(dto EnrollEmployeeDTO) (*models.CourseEnrollment, error) {
	// Check if course exists and is published
	var course models.Course
	if err := s.db.First(&course, "id = ?", dto.CourseID).Error; err != nil {
		return nil, errors.New("course not found")
	}

	if course.Status != models.CourseStatusPublished {
		return nil, errors.New("course is not available for enrollment")
	}

	// Check enrollment limit
	if course.EnrollmentLimit > 0 {
		var count int64
		s.db.Model(&models.CourseEnrollment{}).
			Where("course_id = ? AND status IN (?)", dto.CourseID, []string{"pending", "active"}).
			Count(&count)
		if int(count) >= course.EnrollmentLimit {
			return nil, errors.New("enrollment limit reached")
		}
	}

	// Check if already enrolled
	var existing models.CourseEnrollment
	err := s.db.Where("course_id = ? AND employee_id = ? AND status NOT IN (?)",
		dto.CourseID, dto.EmployeeID, []string{"dropped", "expired"}).
		First(&existing).Error
	if err == nil {
		return nil, errors.New("employee already enrolled in this course")
	}

	enrollment := &models.CourseEnrollment{
		CompanyID:      dto.CompanyID,
		CourseID:       dto.CourseID,
		EmployeeID:     dto.EmployeeID,
		Status:         models.EnrollmentStatusPending,
		ProgressPercent: 0,
		EnrolledByID:   dto.EnrolledByID,
		EnrollmentType: dto.EnrollmentType,
		DueDate:        dto.DueDate,
	}

	if enrollment.EnrollmentType == "" {
		enrollment.EnrollmentType = "self"
	}

	if err := s.db.Create(enrollment).Error; err != nil {
		return nil, err
	}

	return enrollment, nil
}

// StartCourse starts a course for an enrollment
func (s *TrainingService) StartCourse(enrollmentID uuid.UUID) (*models.CourseEnrollment, error) {
	var enrollment models.CourseEnrollment
	if err := s.db.First(&enrollment, "id = ?", enrollmentID).Error; err != nil {
		return nil, errors.New("enrollment not found")
	}

	if enrollment.Status != models.EnrollmentStatusPending {
		return nil, errors.New("course already started")
	}

	now := time.Now()
	enrollment.Status = models.EnrollmentStatusActive
	enrollment.StartedAt = &now

	// Find first module and content
	var firstModule models.CourseModule
	if err := s.db.Where("course_id = ?", enrollment.CourseID).
		Order("display_order ASC").First(&firstModule).Error; err == nil {
		enrollment.CurrentModuleID = &firstModule.ID

		var firstContent models.ModuleContent
		if err := s.db.Where("module_id = ?", firstModule.ID).
			Order("display_order ASC").First(&firstContent).Error; err == nil {
			enrollment.CurrentContentID = &firstContent.ID
		}
	}

	if err := s.db.Save(&enrollment).Error; err != nil {
		return nil, err
	}

	return &enrollment, nil
}

// UpdateContentProgress updates progress for a content item
type UpdateProgressDTO struct {
	EnrollmentID uuid.UUID
	ContentID    uuid.UUID
	TimeSpent    int
	Score        int
	IsCompleted  bool
}

// UpdateContentProgress updates progress on a content item
func (s *TrainingService) UpdateContentProgress(dto UpdateProgressDTO) (*models.ContentProgress, error) {
	// Check enrollment exists and is active
	var enrollment models.CourseEnrollment
	if err := s.db.First(&enrollment, "id = ?", dto.EnrollmentID).Error; err != nil {
		return nil, errors.New("enrollment not found")
	}

	if enrollment.Status != models.EnrollmentStatusActive {
		return nil, errors.New("enrollment is not active")
	}

	// Get or create progress record
	var progress models.ContentProgress
	err := s.db.Where("enrollment_id = ? AND content_id = ?", dto.EnrollmentID, dto.ContentID).
		First(&progress).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		progress = models.ContentProgress{
			EnrollmentID: dto.EnrollmentID,
			ContentID:    dto.ContentID,
		}
	}

	now := time.Now()
	progress.TimeSpentMinutes += dto.TimeSpent
	progress.LastAccessedAt = &now
	progress.Attempts++

	if dto.Score > 0 {
		progress.Score = dto.Score
	}

	if dto.IsCompleted && !progress.IsCompleted {
		progress.IsCompleted = true
		progress.CompletedAt = &now
	}

	if progress.ID == uuid.Nil {
		if err := s.db.Create(&progress).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Save(&progress).Error; err != nil {
			return nil, err
		}
	}

	// Update enrollment progress
	s.updateEnrollmentProgress(dto.EnrollmentID)

	return &progress, nil
}

// updateEnrollmentProgress recalculates enrollment progress
func (s *TrainingService) updateEnrollmentProgress(enrollmentID uuid.UUID) {
	var enrollment models.CourseEnrollment
	if err := s.db.First(&enrollment, "id = ?", enrollmentID).Error; err != nil {
		return
	}

	// Get total required content count
	var totalContent int64
	s.db.Model(&models.ModuleContent{}).
		Joins("JOIN training_course_modules ON training_module_contents.module_id = training_course_modules.id").
		Where("training_course_modules.course_id = ? AND training_module_contents.is_required = ?", enrollment.CourseID, true).
		Count(&totalContent)

	if totalContent == 0 {
		return
	}

	// Get completed content count
	var completedContent int64
	s.db.Model(&models.ContentProgress{}).
		Where("enrollment_id = ? AND is_completed = ?", enrollmentID, true).
		Count(&completedContent)

	enrollment.ProgressPercent = int((float64(completedContent) / float64(totalContent)) * 100)

	// Get total time
	var totalTime struct {
		Total int
	}
	s.db.Model(&models.ContentProgress{}).
		Select("SUM(time_spent_minutes) as total").
		Where("enrollment_id = ?", enrollmentID).
		Scan(&totalTime)
	enrollment.TotalTimeMinutes = totalTime.Total

	// Check if completed
	if enrollment.ProgressPercent >= 100 {
		now := time.Now()
		enrollment.Status = models.EnrollmentStatusCompleted
		enrollment.CompletedAt = &now

		// Calculate final score
		var avgScore struct {
			Avg float64
		}
		s.db.Model(&models.ContentProgress{}).
			Select("AVG(score) as avg").
			Where("enrollment_id = ? AND score > 0", enrollmentID).
			Scan(&avgScore)
		enrollment.FinalScore = int(avgScore.Avg)
	}

	s.db.Save(&enrollment)
}

// GetEnrollmentByID retrieves an enrollment with progress details
func (s *TrainingService) GetEnrollmentByID(id uuid.UUID) (*models.CourseEnrollment, error) {
	var enrollment models.CourseEnrollment
	err := s.db.Preload("Course").Preload("Progress").Preload("Employee").
		First(&enrollment, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("enrollment not found")
		}
		return nil, err
	}
	return &enrollment, nil
}

// GetEmployeeEnrollments gets all enrollments for an employee
func (s *TrainingService) GetEmployeeEnrollments(companyID, employeeID uuid.UUID, status string) ([]models.CourseEnrollment, error) {
	query := s.db.Where("company_id = ? AND employee_id = ?", companyID, employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var enrollments []models.CourseEnrollment
	if err := query.Preload("Course").Find(&enrollments).Error; err != nil {
		return nil, err
	}

	return enrollments, nil
}

// IssueCertificate issues a certificate for a completed enrollment
func (s *TrainingService) IssueCertificate(enrollmentID uuid.UUID) (*models.Certificate, error) {
	var enrollment models.CourseEnrollment
	if err := s.db.Preload("Course").First(&enrollment, "id = ?", enrollmentID).Error; err != nil {
		return nil, errors.New("enrollment not found")
	}

	if enrollment.Status != models.EnrollmentStatusCompleted {
		return nil, errors.New("course not completed")
	}

	// Check if certificate already issued
	if enrollment.CertificateIssued {
		var existing models.Certificate
		if err := s.db.Where("enrollment_id = ?", enrollmentID).First(&existing).Error; err == nil {
			return &existing, nil
		}
	}

	// Generate certificate number
	certNumber := "CERT-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]

	now := time.Now()
	var expiresAt *time.Time
	if enrollment.Course.CertificateValid > 0 {
		exp := now.AddDate(0, enrollment.Course.CertificateValid, 0)
		expiresAt = &exp
	}

	cert := &models.Certificate{
		CompanyID:         enrollment.CompanyID,
		EnrollmentID:      enrollmentID,
		EmployeeID:        enrollment.EmployeeID,
		CourseID:          enrollment.CourseID,
		CertificateNumber: certNumber,
		Title:             enrollment.Course.Title,
		IssuedAt:          now,
		ExpiresAt:         expiresAt,
		FinalScore:        enrollment.FinalScore,
		IsValid:           true,
	}

	if err := s.db.Create(cert).Error; err != nil {
		return nil, err
	}

	// Update enrollment
	enrollment.CertificateIssued = true
	enrollment.CertificateIssuedAt = &now
	if expiresAt != nil {
		enrollment.CertificateExpiresAt = expiresAt
	}
	s.db.Save(&enrollment)

	return cert, nil
}

// === Category Methods ===

// CreateCategory creates a training category
func (s *TrainingService) CreateCategory(companyID uuid.UUID, name, description string, parentID *uuid.UUID) (*models.TrainingCategory, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}

	category := &models.TrainingCategory{
		CompanyID:   companyID,
		Name:        name,
		Description: description,
		ParentID:    parentID,
		IsActive:    true,
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategories retrieves all categories for a company
func (s *TrainingService) GetCategories(companyID uuid.UUID) ([]models.TrainingCategory, error) {
	var categories []models.TrainingCategory
	err := s.db.Where("company_id = ? AND is_active = ?", companyID, true).
		Preload("Children").
		Where("parent_id IS NULL").
		Order("name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// === Module Methods ===

// CreateModule creates a course module
func (s *TrainingService) CreateModule(courseID uuid.UUID, title, description string, order int) (*models.CourseModule, error) {
	if title == "" {
		return nil, errors.New("module title is required")
	}

	module := &models.CourseModule{
		CourseID:     courseID,
		Title:        title,
		Description:  description,
		DisplayOrder: order,
		IsActive:     true,
		IsRequired:   true,
	}

	if err := s.db.Create(module).Error; err != nil {
		return nil, err
	}

	return module, nil
}

// CreateContent creates module content
func (s *TrainingService) CreateContent(moduleID uuid.UUID, title, description string, contentType models.ContentType, contentURL string, order int) (*models.ModuleContent, error) {
	if title == "" {
		return nil, errors.New("content title is required")
	}

	content := &models.ModuleContent{
		ModuleID:     moduleID,
		Title:        title,
		Description:  description,
		ContentType:  contentType,
		ContentURL:   contentURL,
		DisplayOrder: order,
		IsActive:     true,
		IsRequired:   true,
		PassingScore: 70,
		MaxAttempts:  3,
	}

	if err := s.db.Create(content).Error; err != nil {
		return nil, err
	}

	return content, nil
}

// === Statistics ===

// CourseStats contains course statistics
type CourseStats struct {
	TotalEnrollments   int     `json:"total_enrollments"`
	ActiveEnrollments  int     `json:"active_enrollments"`
	CompletedCount     int     `json:"completed_count"`
	AverageScore       float64 `json:"average_score"`
	AverageCompletion  float64 `json:"average_completion_time_days"`
	CompletionRate     float64 `json:"completion_rate"`
}

// GetCourseStats gets statistics for a course
func (s *TrainingService) GetCourseStats(courseID uuid.UUID) (*CourseStats, error) {
	var stats CourseStats

	// Total enrollments
	var total int64
	s.db.Model(&models.CourseEnrollment{}).Where("course_id = ?", courseID).Count(&total)
	stats.TotalEnrollments = int(total)

	// Active enrollments
	var active int64
	s.db.Model(&models.CourseEnrollment{}).
		Where("course_id = ? AND status = ?", courseID, models.EnrollmentStatusActive).
		Count(&active)
	stats.ActiveEnrollments = int(active)

	// Completed count
	var completed int64
	s.db.Model(&models.CourseEnrollment{}).
		Where("course_id = ? AND status = ?", courseID, models.EnrollmentStatusCompleted).
		Count(&completed)
	stats.CompletedCount = int(completed)

	// Average score
	var avgScore struct {
		Avg float64
	}
	s.db.Model(&models.CourseEnrollment{}).
		Select("AVG(final_score) as avg").
		Where("course_id = ? AND status = ?", courseID, models.EnrollmentStatusCompleted).
		Scan(&avgScore)
	stats.AverageScore = avgScore.Avg

	// Completion rate
	if stats.TotalEnrollments > 0 {
		stats.CompletionRate = float64(stats.CompletedCount) / float64(stats.TotalEnrollments) * 100
	}

	return &stats, nil
}
