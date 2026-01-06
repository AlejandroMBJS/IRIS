/*
Package services - File Upload Service

==============================================================================
FILE: internal/services/upload_service.go
==============================================================================

DESCRIPTION:
    Handles secure file uploads for incidence evidence (medical certificates,
    overtime requests, etc.) with validation, MIME type checking, and file
    size limits.

USER PERSPECTIVE:
    - Upload evidence files for incidences (PDF, images, documents)
    - Secure file storage with unique filenames
    - Download uploaded evidence files
    - Delete evidence when no longer needed

DEVELOPER GUIDELINES:
    OK to modify: Allowed file types, maximum file size (currently 10MB)
    CAUTION: File validation prevents malicious uploads
    DO NOT modify: File path generation without security review
    Note: Files stored in uploads/evidence directory

SYNTAX EXPLANATION:
    - AllowedMimeTypes: PDF, images (JPG, PNG, GIF), Word docs, plain text
    - MaxFileSize: 10MB limit for uploads
    - Files renamed with UUID to prevent collisions and path traversal
    - MIME type validated using http.DetectContentType (first 512 bytes)
    - Physical file deleted when evidence record is removed

==============================================================================
*/
package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// File upload configuration
const (
	MaxFileSize       = 10 * 1024 * 1024 // 10MB max file size
	UploadDir         = "uploads/evidence"
)

// Allowed file types for evidence upload
var AllowedMimeTypes = map[string]bool{
	"application/pdf":  true,
	"image/jpeg":       true,
	"image/jpg":        true,
	"image/png":        true,
	"image/gif":        true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain":       true,
}

// Allowed file extensions
var AllowedExtensions = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
}

// Upload service errors
var (
	ErrFileTooLarge      = errors.New("file size exceeds maximum allowed (10MB)")
	ErrInvalidFileType   = errors.New("file type not allowed")
	ErrInvalidExtension  = errors.New("file extension not allowed")
	ErrNoFile            = errors.New("no file provided")
	ErrIncidenceNotFound = errors.New("incidence not found")
	ErrEvidenceNotFound  = errors.New("evidence not found")
)

// UploadService handles file uploads
type UploadService struct {
	db *gorm.DB
}

// NewUploadService creates a new upload service
func NewUploadService(db *gorm.DB) *UploadService {
	// Ensure upload directory exists
	os.MkdirAll(UploadDir, 0755)
	return &UploadService{db: db}
}

// UploadEvidence uploads an evidence file for an incidence
func (s *UploadService) UploadEvidence(incidenceID, userID uuid.UUID, fileHeader *multipart.FileHeader) (*models.IncidenceEvidence, error) {
	// Validate incidence exists
	var incidence models.Incidence
	if err := s.db.First(&incidence, "id = ?", incidenceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIncidenceNotFound
		}
		return nil, err
	}

	// Validate file
	if fileHeader == nil {
		return nil, ErrNoFile
	}

	// Check file size
	if fileHeader.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !AllowedExtensions[ext] {
		return nil, ErrInvalidExtension
	}

	// Open the file to validate MIME type
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	contentType := http.DetectContentType(buffer[:n])

	// Validate MIME type
	if !AllowedMimeTypes[contentType] {
		// Some files may have generic types, check extension as fallback
		if !strings.HasPrefix(contentType, "application/octet-stream") || !AllowedExtensions[ext] {
			return nil, ErrInvalidFileType
		}
		// Override with extension-based type for allowed extensions
		switch ext {
		case ".pdf":
			contentType = "application/pdf"
		case ".doc":
			contentType = "application/msword"
		case ".docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		}
	}

	// Reset file position
	file.Seek(0, 0)

	// Generate unique filename
	newFileName := fmt.Sprintf("%s_%s%s", incidenceID.String(), uuid.New().String(), ext)
	filePath := filepath.Join(UploadDir, newFileName)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, file); err != nil {
		// Clean up on failure
		os.Remove(filePath)
		return nil, err
	}

	// Create evidence record
	evidence := &models.IncidenceEvidence{
		IncidenceID:  incidenceID,
		FileName:     newFileName,
		OriginalName: fileHeader.Filename,
		ContentType:  contentType,
		FileSize:     fileHeader.Size,
		FilePath:     filePath,
		UploadedBy:   userID,
	}

	if err := s.db.Create(evidence).Error; err != nil {
		// Clean up file on database error
		os.Remove(filePath)
		return nil, err
	}

	return evidence, nil
}

// GetEvidenceByIncidence retrieves all evidence files for an incidence
func (s *UploadService) GetEvidenceByIncidence(incidenceID uuid.UUID) ([]models.IncidenceEvidence, error) {
	var evidences []models.IncidenceEvidence
	err := s.db.Where("incidence_id = ?", incidenceID).Find(&evidences).Error
	return evidences, err
}

// GetEvidenceByID retrieves a specific evidence file
func (s *UploadService) GetEvidenceByID(evidenceID uuid.UUID) (*models.IncidenceEvidence, error) {
	var evidence models.IncidenceEvidence
	if err := s.db.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEvidenceNotFound
		}
		return nil, err
	}
	return &evidence, nil
}

// DeleteEvidence deletes an evidence file
func (s *UploadService) DeleteEvidence(evidenceID uuid.UUID) error {
	var evidence models.IncidenceEvidence
	if err := s.db.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEvidenceNotFound
		}
		return err
	}

	// Delete physical file
	if err := os.Remove(evidence.FilePath); err != nil && !os.IsNotExist(err) {
		// Log error but continue with database deletion
	}

	// Delete database record
	return s.db.Delete(&evidence).Error
}

// GetFilePath returns the file path for an evidence
func (s *UploadService) GetFilePath(evidenceID uuid.UUID) (string, string, error) {
	evidence, err := s.GetEvidenceByID(evidenceID)
	if err != nil {
		return "", "", err
	}
	return evidence.FilePath, evidence.OriginalName, nil
}
