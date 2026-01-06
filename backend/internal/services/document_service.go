/*
Package services - IRIS Document Management Service

==============================================================================
FILE: internal/services/document_service.go
==============================================================================

DESCRIPTION:
    Business logic for Document Management including document storage,
    templates, e-signatures, and employee documents.

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

// DocumentService provides business logic for document management
type DocumentService struct {
	db *gorm.DB
}

// NewDocumentService creates a new DocumentService
func NewDocumentService(db *gorm.DB) *DocumentService {
	return &DocumentService{db: db}
}

// === Document Category Methods ===

type CreateDocumentCategoryDTO struct {
	CompanyID         uuid.UUID
	Name              string
	Code              string
	Description       string
	ParentID          *uuid.UUID
	ViewRoles         []string
	EditRoles         []string
	DeleteRoles       []string
	IsEmployeeFolder  bool
	RequiresExpiry    bool
	DefaultExpiryDays int
	RetentionYears    int
	Icon              string
	Color             string
}

// CreateDocumentCategory creates a document category
func (s *DocumentService) CreateDocumentCategory(dto CreateDocumentCategoryDTO) (*models.DocumentCategory, error) {
	if dto.Name == "" {
		return nil, errors.New("category name is required")
	}

	category := &models.DocumentCategory{
		CompanyID:         dto.CompanyID,
		Name:              dto.Name,
		Code:              dto.Code,
		Description:       dto.Description,
		ParentID:          dto.ParentID,
		ViewRoles:         dto.ViewRoles,
		EditRoles:         dto.EditRoles,
		DeleteRoles:       dto.DeleteRoles,
		IsEmployeeFolder:  dto.IsEmployeeFolder,
		RequiresExpiry:    dto.RequiresExpiry,
		DefaultExpiryDays: dto.DefaultExpiryDays,
		RetentionYears:    dto.RetentionYears,
		Icon:              dto.Icon,
		Color:             dto.Color,
		IsActive:          true,
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// GetDocumentCategories gets all categories for a company
func (s *DocumentService) GetDocumentCategories(companyID uuid.UUID) ([]models.DocumentCategory, error) {
	var categories []models.DocumentCategory
	if err := s.db.Where("company_id = ? AND is_active = ?", companyID, true).
		Preload("Children").
		Where("parent_id IS NULL").
		Order("display_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// === Document Methods ===

type CreateDocumentDTO struct {
	CompanyID       uuid.UUID
	CategoryID      *uuid.UUID
	Name            string
	Description     string
	DocumentType    string
	FileName        string
	FileExtension   string
	MimeType        string
	FileSize        int64
	FileURL         string
	StorageKey      string
	EmployeeID      *uuid.UUID
	DepartmentID    *uuid.UUID
	EffectiveDate   *time.Time
	ExpiryDate      *time.Time
	IsConfidential  bool
	AccessLevel     string
	Tags            []string
	RequiresSignature bool
	UploadedByID    *uuid.UUID
}

type DocumentFilters struct {
	CompanyID    uuid.UUID
	CategoryID   *uuid.UUID
	EmployeeID   *uuid.UUID
	DocumentType string
	Status       string
	Search       string
	Tags         []string
	Page         int
	Limit        int
}

type PaginatedDocuments struct {
	Data       []models.Document `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// CreateDocument creates a new document
func (s *DocumentService) CreateDocument(dto CreateDocumentDTO) (*models.Document, error) {
	if dto.Name == "" {
		return nil, errors.New("document name is required")
	}
	if dto.FileURL == "" {
		return nil, errors.New("file URL is required")
	}

	doc := &models.Document{
		CompanyID:         dto.CompanyID,
		CategoryID:        dto.CategoryID,
		Name:              dto.Name,
		Description:       dto.Description,
		DocumentType:      dto.DocumentType,
		FileName:          dto.FileName,
		FileExtension:     dto.FileExtension,
		MimeType:          dto.MimeType,
		FileSize:          dto.FileSize,
		FileURL:           dto.FileURL,
		StorageKey:        dto.StorageKey,
		EmployeeID:        dto.EmployeeID,
		DepartmentID:      dto.DepartmentID,
		EffectiveDate:     dto.EffectiveDate,
		ExpiryDate:        dto.ExpiryDate,
		IsConfidential:    dto.IsConfidential,
		AccessLevel:       dto.AccessLevel,
		Tags:              dto.Tags,
		RequiresSignature: dto.RequiresSignature,
		UploadedByID:      dto.UploadedByID,
		Status:            models.DocumentStatusActive,
		Version:           1,
		IsLatestVersion:   true,
	}

	if doc.AccessLevel == "" {
		doc.AccessLevel = "internal"
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// GetDocumentByID retrieves a document by ID
func (s *DocumentService) GetDocumentByID(id uuid.UUID) (*models.Document, error) {
	var doc models.Document
	err := s.db.Preload("Category").Preload("SignatureRequests").Preload("Employee").
		First(&doc, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}
	return &doc, nil
}

// ListDocuments retrieves documents with filters
func (s *DocumentService) ListDocuments(filters DocumentFilters) (*PaginatedDocuments, error) {
	query := s.db.Model(&models.Document{}).Where("company_id = ?", filters.CompanyID)

	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.EmployeeID != nil {
		query = query.Where("employee_id = ?", *filters.EmployeeID)
	}
	if filters.DocumentType != "" {
		query = query.Where("document_type = ?", filters.DocumentType)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR file_name LIKE ?", searchTerm, searchTerm, searchTerm)
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

	var docs []models.Document
	if err := query.Offset(offset).Limit(filters.Limit).
		Preload("Category").
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	return &PaginatedDocuments{
		Data:       docs,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateDocument updates a document
func (s *DocumentService) UpdateDocument(id uuid.UUID, name, description string, tags []string, expiryDate *time.Time) (*models.Document, error) {
	doc, err := s.GetDocumentByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		doc.Name = name
	}
	if description != "" {
		doc.Description = description
	}
	if tags != nil {
		doc.Tags = tags
	}
	if expiryDate != nil {
		doc.ExpiryDate = expiryDate
	}

	if err := s.db.Save(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// ArchiveDocument archives a document
func (s *DocumentService) ArchiveDocument(id uuid.UUID) (*models.Document, error) {
	doc, err := s.GetDocumentByID(id)
	if err != nil {
		return nil, err
	}

	doc.Status = models.DocumentStatusArchived

	if err := s.db.Save(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// DeleteDocument soft deletes a document
func (s *DocumentService) DeleteDocument(id uuid.UUID) error {
	result := s.db.Delete(&models.Document{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("document not found")
	}
	return nil
}

// CreateDocumentVersion creates a new version of a document
func (s *DocumentService) CreateDocumentVersion(parentID uuid.UUID, dto CreateDocumentDTO) (*models.Document, error) {
	parent, err := s.GetDocumentByID(parentID)
	if err != nil {
		return nil, err
	}

	// Mark parent as not latest
	parent.IsLatestVersion = false
	s.db.Save(parent)

	// Create new version
	doc := &models.Document{
		CompanyID:       parent.CompanyID,
		CategoryID:      parent.CategoryID,
		Name:            parent.Name,
		Description:     dto.Description,
		DocumentType:    parent.DocumentType,
		FileName:        dto.FileName,
		FileExtension:   dto.FileExtension,
		MimeType:        dto.MimeType,
		FileSize:        dto.FileSize,
		FileURL:         dto.FileURL,
		StorageKey:      dto.StorageKey,
		EmployeeID:      parent.EmployeeID,
		DepartmentID:    parent.DepartmentID,
		EffectiveDate:   dto.EffectiveDate,
		ExpiryDate:      dto.ExpiryDate,
		IsConfidential:  parent.IsConfidential,
		AccessLevel:     parent.AccessLevel,
		Tags:            parent.Tags,
		UploadedByID:    dto.UploadedByID,
		Status:          models.DocumentStatusActive,
		Version:         parent.Version + 1,
		IsLatestVersion: true,
		ParentDocID:     &parentID,
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// LogDocumentAccess logs document access for auditing
func (s *DocumentService) LogDocumentAccess(docID, userID uuid.UUID, action, ipAddress, userAgent, sessionID string) error {
	log := &models.DocumentAccessLog{
		DocumentID: docID,
		UserID:     userID,
		Action:     action,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		SessionID:  sessionID,
	}

	if err := s.db.Create(log).Error; err != nil {
		return err
	}

	// Update document stats
	switch action {
	case "view":
		s.db.Model(&models.Document{}).Where("id = ?", docID).
			Updates(map[string]interface{}{
				"view_count":        gorm.Expr("view_count + 1"),
				"last_viewed_at":    time.Now(),
				"last_viewed_by_id": userID,
			})
	case "download":
		s.db.Model(&models.Document{}).Where("id = ?", docID).
			Update("download_count", gorm.Expr("download_count + 1"))
	}

	return nil
}

// === Signature Request Methods ===

type CreateSignatureRequestDTO struct {
	CompanyID     uuid.UUID
	DocumentID    uuid.UUID
	RequestedByID uuid.UUID
	Message       string
	ExpiresAt     *time.Time
	Signers       []SignerDTO
}

type SignerDTO struct {
	EmployeeID    *uuid.UUID
	ExternalEmail string
	ExternalName  string
	SigningOrder  int
	Role          string
}

// CreateSignatureRequest creates a signature request for a document
func (s *DocumentService) CreateSignatureRequest(dto CreateSignatureRequestDTO) (*models.SignatureRequest, error) {
	if len(dto.Signers) == 0 {
		return nil, errors.New("at least one signer is required")
	}

	// Verify document exists
	var doc models.Document
	if err := s.db.First(&doc, "id = ?", dto.DocumentID).Error; err != nil {
		return nil, errors.New("document not found")
	}

	request := &models.SignatureRequest{
		CompanyID:     dto.CompanyID,
		DocumentID:    dto.DocumentID,
		RequestedByID: dto.RequestedByID,
		Message:       dto.Message,
		ExpiresAt:     dto.ExpiresAt,
		Status:        models.SignatureStatusPending,
		SignerOrder:   len(dto.Signers),
		CurrentStep:   1,
	}

	if err := s.db.Create(request).Error; err != nil {
		return nil, err
	}

	// Create signers
	for _, signerDTO := range dto.Signers {
		accessToken := uuid.New().String()
		signer := &models.Signer{
			RequestID:     request.ID,
			EmployeeID:    signerDTO.EmployeeID,
			ExternalEmail: signerDTO.ExternalEmail,
			ExternalName:  signerDTO.ExternalName,
			SigningOrder:  signerDTO.SigningOrder,
			Role:          signerDTO.Role,
			Status:        models.SignatureStatusPending,
			AccessToken:   accessToken,
		}
		s.db.Create(signer)
	}

	// Update document
	doc.RequiresSignature = true
	s.db.Save(&doc)

	return request, nil
}

// GetSignatureRequestByID retrieves a signature request by ID
func (s *DocumentService) GetSignatureRequestByID(id uuid.UUID) (*models.SignatureRequest, error) {
	var request models.SignatureRequest
	err := s.db.Preload("Document").Preload("Signers").First(&request, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("signature request not found")
		}
		return nil, err
	}
	return &request, nil
}

// SignDocument signs a document
func (s *DocumentService) SignDocument(requestID, signerID uuid.UUID, signatureData, signatureType, ipAddress, userAgent string) error {
	var signer models.Signer
	if err := s.db.Where("request_id = ? AND id = ?", requestID, signerID).First(&signer).Error; err != nil {
		return errors.New("signer not found")
	}

	if signer.Status != models.SignatureStatusPending {
		return errors.New("already signed or declined")
	}

	now := time.Now()
	signer.Status = models.SignatureStatusSigned
	signer.SignedAt = &now
	signer.SignatureData = signatureData
	signer.SignatureType = signatureType
	signer.SignedIPAddress = ipAddress
	signer.SignedUserAgent = userAgent

	if err := s.db.Save(&signer).Error; err != nil {
		return err
	}

	// Check if all signers have signed
	var request models.SignatureRequest
	s.db.Preload("Signers").First(&request, "id = ?", requestID)

	allSigned := true
	for _, sig := range request.Signers {
		if sig.Status != models.SignatureStatusSigned {
			allSigned = false
			// Move to next signer if sequential
			if sig.SigningOrder == request.CurrentStep+1 {
				request.CurrentStep++
				s.db.Save(&request)
			}
			break
		}
	}

	if allSigned {
		request.Status = models.SignatureStatusSigned
		request.CompletedAt = &now
		s.db.Save(&request)

		// Update document
		var doc models.Document
		if s.db.First(&doc, "id = ?", request.DocumentID).Error == nil {
			doc.SignatureCompleted = true
			s.db.Save(&doc)
		}
	}

	return nil
}

// DeclineSignature declines a signature request
func (s *DocumentService) DeclineSignature(requestID, signerID uuid.UUID, reason string) error {
	var signer models.Signer
	if err := s.db.Where("request_id = ? AND id = ?", requestID, signerID).First(&signer).Error; err != nil {
		return errors.New("signer not found")
	}

	now := time.Now()
	signer.Status = models.SignatureStatusDeclined
	signer.DeclinedAt = &now
	signer.DeclineReason = reason

	if err := s.db.Save(&signer).Error; err != nil {
		return err
	}

	// Update request status
	var request models.SignatureRequest
	s.db.First(&request, "id = ?", requestID)
	request.Status = models.SignatureStatusDeclined
	s.db.Save(&request)

	return nil
}

// === Document Template Methods ===

type CreateDocumentTemplateDTO struct {
	CompanyID         uuid.UUID
	Name              string
	Code              string
	Description       string
	TemplateType      string
	ContentType       string
	TemplateURL       string
	TemplateData      string
	Variables         string
	OutputFormat      string
	RequiresApproval  bool
	RequiresSignature bool
	SignatureRoles    []string
	ApplicableEvents  []string
	CreatedByID       *uuid.UUID
}

// CreateDocumentTemplate creates a document template
func (s *DocumentService) CreateDocumentTemplate(dto CreateDocumentTemplateDTO) (*models.DocumentTemplate, error) {
	if dto.Name == "" {
		return nil, errors.New("template name is required")
	}

	template := &models.DocumentTemplate{
		CompanyID:         dto.CompanyID,
		Name:              dto.Name,
		Code:              dto.Code,
		Description:       dto.Description,
		TemplateType:      dto.TemplateType,
		ContentType:       dto.ContentType,
		TemplateURL:       dto.TemplateURL,
		TemplateData:      dto.TemplateData,
		Variables:         dto.Variables,
		OutputFormat:      dto.OutputFormat,
		RequiresApproval:  dto.RequiresApproval,
		RequiresSignature: dto.RequiresSignature,
		SignatureRoles:    dto.SignatureRoles,
		ApplicableEvents:  dto.ApplicableEvents,
		IsActive:          true,
		Version:           1,
		CreatedByID:       dto.CreatedByID,
	}

	if template.OutputFormat == "" {
		template.OutputFormat = "pdf"
	}
	if template.ContentType == "" {
		template.ContentType = "html"
	}

	if err := s.db.Create(template).Error; err != nil {
		return nil, err
	}

	return template, nil
}

// GetDocumentTemplates gets all templates for a company
func (s *DocumentService) GetDocumentTemplates(companyID uuid.UUID, templateType string) ([]models.DocumentTemplate, error) {
	query := s.db.Where("company_id = ? AND is_active = ?", companyID, true)
	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}

	var templates []models.DocumentTemplate
	if err := query.Order("name ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// === Employee Document Methods ===

type CreateEmployeeDocumentDTO struct {
	CompanyID     uuid.UUID
	EmployeeID    uuid.UUID
	RequirementID *uuid.UUID
	DocumentID    uuid.UUID
	EmployeeNotes string
	ExpiresAt     *time.Time
}

// SubmitEmployeeDocument submits a document for an employee requirement
func (s *DocumentService) SubmitEmployeeDocument(dto CreateEmployeeDocumentDTO) (*models.EmployeeDocument, error) {
	empDoc := &models.EmployeeDocument{
		CompanyID:     dto.CompanyID,
		EmployeeID:    dto.EmployeeID,
		RequirementID: dto.RequirementID,
		DocumentID:    dto.DocumentID,
		EmployeeNotes: dto.EmployeeNotes,
		ExpiresAt:     dto.ExpiresAt,
		Status:        "pending",
	}

	if err := s.db.Create(empDoc).Error; err != nil {
		return nil, err
	}

	return empDoc, nil
}

// VerifyEmployeeDocument verifies an employee document
func (s *DocumentService) VerifyEmployeeDocument(id uuid.UUID, verifierID uuid.UUID, notes string) (*models.EmployeeDocument, error) {
	var empDoc models.EmployeeDocument
	if err := s.db.First(&empDoc, "id = ?", id).Error; err != nil {
		return nil, errors.New("employee document not found")
	}

	now := time.Now()
	empDoc.Status = "verified"
	empDoc.VerifiedAt = &now
	empDoc.VerifiedByID = &verifierID
	empDoc.VerifierNotes = notes

	if err := s.db.Save(&empDoc).Error; err != nil {
		return nil, err
	}

	return &empDoc, nil
}

// RejectEmployeeDocument rejects an employee document
func (s *DocumentService) RejectEmployeeDocument(id uuid.UUID, reason string) (*models.EmployeeDocument, error) {
	var empDoc models.EmployeeDocument
	if err := s.db.First(&empDoc, "id = ?", id).Error; err != nil {
		return nil, errors.New("employee document not found")
	}

	empDoc.Status = "rejected"
	empDoc.RejectionReason = reason

	if err := s.db.Save(&empDoc).Error; err != nil {
		return nil, err
	}

	return &empDoc, nil
}

// GetEmployeeDocuments gets documents submitted by an employee
func (s *DocumentService) GetEmployeeDocuments(employeeID uuid.UUID, status string) ([]models.EmployeeDocument, error) {
	query := s.db.Where("employee_id = ?", employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var docs []models.EmployeeDocument
	if err := query.Preload("Document").Preload("Requirement").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// === Sharing Methods ===

type ShareDocumentDTO struct {
	DocumentID       uuid.UUID
	SharedByID       uuid.UUID
	SharedWithUserID *uuid.UUID
	SharedWithRole   string
	SharedWithDept   *uuid.UUID
	ShareType        string
	CanView          bool
	CanDownload      bool
	CanPrint         bool
	CanEdit          bool
	CanShare         bool
	ExpiresAt        *time.Time
	ShareMessage     string
	MaxAccess        int
}

// ShareDocument shares a document
func (s *DocumentService) ShareDocument(dto ShareDocumentDTO) (*models.SharedDocument, error) {
	accessToken := uuid.New().String()

	share := &models.SharedDocument{
		DocumentID:       dto.DocumentID,
		SharedByID:       dto.SharedByID,
		SharedWithUserID: dto.SharedWithUserID,
		SharedWithRole:   dto.SharedWithRole,
		SharedWithDept:   dto.SharedWithDept,
		ShareType:        dto.ShareType,
		CanView:          dto.CanView,
		CanDownload:      dto.CanDownload,
		CanPrint:         dto.CanPrint,
		CanEdit:          dto.CanEdit,
		CanShare:         dto.CanShare,
		ExpiresAt:        dto.ExpiresAt,
		ShareMessage:     dto.ShareMessage,
		MaxAccess:        dto.MaxAccess,
		AccessToken:      accessToken,
		IsActive:         true,
	}

	if share.ShareType == "" {
		share.ShareType = "user"
	}

	if err := s.db.Create(share).Error; err != nil {
		return nil, err
	}

	return share, nil
}

// GetDocumentShares gets all shares for a document
func (s *DocumentService) GetDocumentShares(documentID uuid.UUID) ([]models.SharedDocument, error) {
	var shares []models.SharedDocument
	if err := s.db.Where("document_id = ? AND is_active = ?", documentID, true).Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// RevokeShare revokes a document share
func (s *DocumentService) RevokeShare(id uuid.UUID) error {
	result := s.db.Model(&models.SharedDocument{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("share not found")
	}
	return nil
}
