/*
Package models - IRIS Payroll System Document Management Module Models

==============================================================================
FILE: internal/models/documents.go
==============================================================================

DESCRIPTION:
    Data models for Document Management including document storage,
    templates, e-signatures, and employee documents.

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DocumentStatus represents the status of a document
type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusActive    DocumentStatus = "active"
	DocumentStatusArchived  DocumentStatus = "archived"
	DocumentStatusExpired   DocumentStatus = "expired"
)

// SignatureStatus represents the status of a signature request
type SignatureStatus string

const (
	SignatureStatusPending   SignatureStatus = "pending"
	SignatureStatusSigned    SignatureStatus = "signed"
	SignatureStatusDeclined  SignatureStatus = "declined"
	SignatureStatusExpired   SignatureStatus = "expired"
)

// DocumentCategory represents a document category/folder
type DocumentCategory struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Code         string     `gorm:"size:50;index" json:"code"`
	Description  string     `gorm:"type:text" json:"description"`

	// Parent Category
	ParentID     *uuid.UUID `gorm:"type:text;index" json:"parent_id,omitempty"`

	// Permissions
	ViewRoles    pq.StringArray `gorm:"type:text[]" json:"view_roles"`
	EditRoles    pq.StringArray `gorm:"type:text[]" json:"edit_roles"`
	DeleteRoles  pq.StringArray `gorm:"type:text[]" json:"delete_roles"`

	// Settings
	IsEmployeeFolder bool   `gorm:"default:false" json:"is_employee_folder"` // Auto-created for each employee
	RequiresExpiry   bool   `gorm:"default:false" json:"requires_expiry"`
	DefaultExpiryDays int   `gorm:"default:0" json:"default_expiry_days"`

	// Retention
	RetentionYears int      `gorm:"default:0" json:"retention_years"` // 0 = indefinite

	IsActive     bool       `gorm:"default:true" json:"is_active"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`
	Icon         string     `gorm:"size:50" json:"icon"`
	Color        string     `gorm:"size:7" json:"color"`

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`

	// Relationships
	Parent       *DocumentCategory  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []DocumentCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Documents    []Document         `gorm:"foreignKey:CategoryID" json:"documents,omitempty"`
}

func (DocumentCategory) TableName() string {
	return "document_categories"
}

// Document represents a stored document
type Document struct {
	BaseModel
	CompanyID    uuid.UUID      `gorm:"type:text;not null;index" json:"company_id"`
	CategoryID   *uuid.UUID     `gorm:"type:text;index" json:"category_id,omitempty"`

	// Document Info
	Name         string         `gorm:"size:255;not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	DocumentType string         `gorm:"size:100" json:"document_type"` // contract, policy, form, certificate, etc.

	// File Info
	FileName     string         `gorm:"size:255;not null" json:"file_name"`
	FileExtension string        `gorm:"size:20" json:"file_extension"`
	MimeType     string         `gorm:"size:100" json:"mime_type"`
	FileSize     int64          `gorm:"default:0" json:"file_size"`
	FileURL      string         `gorm:"size:500;not null" json:"file_url"` // S3 or storage path
	StorageKey   string         `gorm:"size:255" json:"storage_key"` // Storage backend key

	// Versioning
	Version      int            `gorm:"default:1" json:"version"`
	IsLatestVersion bool        `gorm:"default:true" json:"is_latest_version"`
	ParentDocID  *uuid.UUID     `gorm:"type:text" json:"parent_doc_id,omitempty"` // Original document for versions

	// Ownership
	EmployeeID   *uuid.UUID     `gorm:"type:text;index" json:"employee_id,omitempty"` // If employee-specific
	DepartmentID *uuid.UUID     `gorm:"type:text" json:"department_id,omitempty"`

	// Dates
	EffectiveDate *time.Time    `json:"effective_date,omitempty"`
	ExpiryDate   *time.Time     `json:"expiry_date,omitempty"`
	ExpiryNotified bool         `gorm:"default:false" json:"expiry_notified"`

	// Status
	Status       DocumentStatus `gorm:"size:50;default:'active'" json:"status"`

	// Security
	IsConfidential bool        `gorm:"default:false" json:"is_confidential"`
	IsEncrypted  bool          `gorm:"default:false" json:"is_encrypted"`
	AccessLevel  string        `gorm:"size:50" json:"access_level"` // public, internal, restricted, confidential

	// Metadata
	Tags         pq.StringArray `gorm:"type:text[]" json:"tags"`
	MetaData     string        `gorm:"type:text" json:"meta_data"` // JSON for custom fields

	// Audit
	UploadedByID *uuid.UUID    `gorm:"type:text" json:"uploaded_by_id,omitempty"`
	LastViewedAt *time.Time    `json:"last_viewed_at,omitempty"`
	LastViewedByID *uuid.UUID  `gorm:"type:text" json:"last_viewed_by_id,omitempty"`
	ViewCount    int           `gorm:"default:0" json:"view_count"`
	DownloadCount int          `gorm:"default:0" json:"download_count"`

	// Signature
	RequiresSignature bool     `gorm:"default:false" json:"requires_signature"`
	SignatureCompleted bool    `gorm:"default:false" json:"signature_completed"`

	// Relationships
	Category     *DocumentCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Employee     *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	SignatureRequests []SignatureRequest `gorm:"foreignKey:DocumentID" json:"signature_requests,omitempty"`
	AccessLog    []DocumentAccessLog `gorm:"foreignKey:DocumentID" json:"access_log,omitempty"`
}

func (Document) TableName() string {
	return "documents"
}

// DocumentTemplate represents a document template for generation
type DocumentTemplate struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Code         string     `gorm:"size:50;uniqueIndex" json:"code"`
	Description  string     `gorm:"type:text" json:"description"`
	TemplateType string     `gorm:"size:100;not null" json:"template_type"` // offer_letter, contract, policy_ack, etc.

	// Template Content
	ContentType  string     `gorm:"size:50" json:"content_type"` // html, docx, pdf
	TemplateURL  string     `gorm:"size:500" json:"template_url"` // Storage path
	TemplateData string     `gorm:"type:text" json:"template_data"` // HTML content or JSON config

	// Variables
	Variables    string     `gorm:"type:text" json:"variables"` // JSON: [{name, type, required, default}]

	// Output Settings
	OutputFormat string     `gorm:"size:20;default:'pdf'" json:"output_format"` // pdf, docx
	PageSize     string     `gorm:"size:20;default:'letter'" json:"page_size"`
	Orientation  string     `gorm:"size:20;default:'portrait'" json:"orientation"`

	// Workflow
	RequiresApproval bool   `gorm:"default:false" json:"requires_approval"`
	RequiresSignature bool  `gorm:"default:false" json:"requires_signature"`
	SignatureRoles   pq.StringArray `gorm:"type:text[]" json:"signature_roles"` // Who needs to sign

	// Applicability
	ApplicableEvents pq.StringArray `gorm:"type:text[]" json:"applicable_events"` // hire, promotion, termination, etc.

	IsActive     bool       `gorm:"default:true" json:"is_active"`
	Version      int        `gorm:"default:1" json:"version"`

	CreatedByID  *uuid.UUID `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (DocumentTemplate) TableName() string {
	return "document_templates"
}

// GeneratedDocument tracks documents generated from templates
type GeneratedDocument struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	TemplateID   uuid.UUID  `gorm:"type:text;not null;index" json:"template_id"`
	DocumentID   uuid.UUID  `gorm:"type:text;not null;index" json:"document_id"` // The generated document
	EmployeeID   *uuid.UUID `gorm:"type:text;index" json:"employee_id,omitempty"`

	// Generation Info
	GeneratedAt  time.Time  `gorm:"autoCreateTime" json:"generated_at"`
	GeneratedByID *uuid.UUID `gorm:"type:text" json:"generated_by_id,omitempty"`

	// Variables used
	VariablesUsed string    `gorm:"type:text" json:"variables_used"` // JSON snapshot of values

	// Event trigger
	TriggerEvent string     `gorm:"size:100" json:"trigger_event"` // hire, promotion, etc.

	// Relationships
	Template     *DocumentTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Document     *Document        `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Employee     *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (GeneratedDocument) TableName() string {
	return "generated_documents"
}

// SignatureRequest represents a request for e-signature
type SignatureRequest struct {
	BaseModel
	CompanyID    uuid.UUID       `gorm:"type:text;not null;index" json:"company_id"`
	DocumentID   uuid.UUID       `gorm:"type:text;not null;index" json:"document_id"`

	// Request Info
	RequestedByID uuid.UUID     `gorm:"type:text;not null" json:"requested_by_id"`
	RequestedAt  time.Time      `gorm:"autoCreateTime" json:"requested_at"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	Message      string         `gorm:"type:text" json:"message"`

	// Signers
	SignerOrder  int            `gorm:"default:0" json:"signer_order"` // For sequential signing

	// Status
	Status       SignatureStatus `gorm:"size:50;default:'pending'" json:"status"`
	CurrentStep  int            `gorm:"default:1" json:"current_step"` // Which signer is active

	// Completion
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	SignedDocURL string         `gorm:"size:500" json:"signed_doc_url"` // Final signed document

	// Reminder
	ReminderSent bool           `gorm:"default:false" json:"reminder_sent"`
	ReminderSentAt *time.Time   `json:"reminder_sent_at,omitempty"`

	// Relationships
	Document     *Document      `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Signers      []Signer       `gorm:"foreignKey:RequestID" json:"signers,omitempty"`
}

func (SignatureRequest) TableName() string {
	return "signature_requests"
}

// Signer represents a person who needs to sign a document
type Signer struct {
	BaseModel
	RequestID    uuid.UUID       `gorm:"type:text;not null;index" json:"request_id"`

	// Signer Info
	EmployeeID   *uuid.UUID      `gorm:"type:text;index" json:"employee_id,omitempty"`
	ExternalEmail string         `gorm:"size:255" json:"external_email"` // For external signers
	ExternalName  string         `gorm:"size:255" json:"external_name"`

	// Signing Info
	SigningOrder int             `gorm:"default:0" json:"signing_order"`
	Role         string          `gorm:"size:100" json:"role"` // employee, manager, hr, witness, etc.

	// Status
	Status       SignatureStatus `gorm:"size:50;default:'pending'" json:"status"`

	// Signature
	SignedAt     *time.Time      `json:"signed_at,omitempty"`
	SignatureData string         `gorm:"type:text" json:"signature_data"` // Base64 signature image or typed
	SignatureType string         `gorm:"size:50" json:"signature_type"` // drawn, typed, uploaded
	SignedIPAddress string       `gorm:"size:50" json:"signed_ip_address"`
	SignedUserAgent string       `gorm:"size:255" json:"signed_user_agent"`

	// Decline
	DeclinedAt   *time.Time      `json:"declined_at,omitempty"`
	DeclineReason string         `gorm:"type:text" json:"decline_reason"`

	// Access
	AccessToken  string          `gorm:"size:255;uniqueIndex" json:"-"` // For external signer access
	ViewedAt     *time.Time      `json:"viewed_at,omitempty"`

	// Relationships
	Request      *SignatureRequest `gorm:"foreignKey:RequestID" json:"request,omitempty"`
	Employee     *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (Signer) TableName() string {
	return "document_signers"
}

// DocumentAccessLog tracks document access for auditing
type DocumentAccessLog struct {
	BaseModel
	DocumentID   uuid.UUID  `gorm:"type:text;not null;index" json:"document_id"`
	UserID       uuid.UUID  `gorm:"type:text;not null;index" json:"user_id"`

	// Access Info
	Action       string     `gorm:"size:50;not null" json:"action"` // view, download, print, edit, share
	AccessedAt   time.Time  `gorm:"autoCreateTime" json:"accessed_at"`

	// Context
	IPAddress    string     `gorm:"size:50" json:"ip_address"`
	UserAgent    string     `gorm:"size:255" json:"user_agent"`
	SessionID    string     `gorm:"size:100" json:"session_id"`

	// Relationships
	Document     *Document  `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

func (DocumentAccessLog) TableName() string {
	return "document_access_logs"
}

// DocumentRequirement represents required documents for employees
type DocumentRequirement struct {
	BaseModel
	CompanyID    uuid.UUID      `gorm:"type:text;not null;index" json:"company_id"`

	Name         string         `gorm:"size:255;not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	CategoryID   *uuid.UUID     `gorm:"type:text" json:"category_id,omitempty"`

	// When Required
	RequiredFor  string         `gorm:"size:100" json:"required_for"` // onboarding, annual, promotion, etc.

	// Applicability
	ApplicableRoles pq.StringArray `gorm:"type:text[]" json:"applicable_roles"`
	ApplicableCollarTypes pq.StringArray `gorm:"type:text[]" json:"applicable_collar_types"`

	// Timing
	DueDaysAfterHire int       `gorm:"default:0" json:"due_days_after_hire"` // For onboarding
	AnnualRenewal    bool      `gorm:"default:false" json:"annual_renewal"`
	RenewalDay       int       `gorm:"default:0" json:"renewal_day"` // Day of year
	RenewalMonth     int       `gorm:"default:0" json:"renewal_month"`

	// Validation
	AllowedFormats   pq.StringArray `gorm:"type:text[]" json:"allowed_formats"` // pdf, jpg, png
	MaxFileSize      int64      `gorm:"default:0" json:"max_file_size"` // bytes

	IsRequired   bool           `gorm:"default:true" json:"is_required"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	DisplayOrder int            `gorm:"default:0" json:"display_order"`

	CreatedByID  *uuid.UUID     `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (DocumentRequirement) TableName() string {
	return "document_requirements"
}

// EmployeeDocument tracks employee-submitted documents
type EmployeeDocument struct {
	BaseModel
	CompanyID     uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID    uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`
	RequirementID *uuid.UUID `gorm:"type:text;index" json:"requirement_id,omitempty"`
	DocumentID    uuid.UUID  `gorm:"type:text;not null;index" json:"document_id"`

	// Submission
	SubmittedAt  time.Time  `gorm:"autoCreateTime" json:"submitted_at"`

	// Verification
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, verified, rejected
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	VerifiedByID *uuid.UUID `gorm:"type:text" json:"verified_by_id,omitempty"`
	RejectionReason string  `gorm:"size:255" json:"rejection_reason"`

	// Expiry
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	ExpiryWarned bool       `gorm:"default:false" json:"expiry_warned"`

	// Notes
	EmployeeNotes string    `gorm:"type:text" json:"employee_notes"`
	VerifierNotes string    `gorm:"type:text" json:"verifier_notes"`

	// Relationships
	Employee     *Employee          `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Requirement  *DocumentRequirement `gorm:"foreignKey:RequirementID" json:"requirement,omitempty"`
	Document     *Document          `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

func (EmployeeDocument) TableName() string {
	return "employee_documents"
}

// SharedDocument tracks document sharing
type SharedDocument struct {
	BaseModel
	DocumentID   uuid.UUID  `gorm:"type:text;not null;index" json:"document_id"`
	SharedByID   uuid.UUID  `gorm:"type:text;not null" json:"shared_by_id"`

	// Share Target
	SharedWithUserID *uuid.UUID `gorm:"type:text;index" json:"shared_with_user_id,omitempty"`
	SharedWithRole   string     `gorm:"size:100" json:"shared_with_role"` // For role-based sharing
	SharedWithDept   *uuid.UUID `gorm:"type:text" json:"shared_with_dept,omitempty"`
	ShareType        string     `gorm:"size:50" json:"share_type"` // user, role, department, public

	// Permissions
	CanView      bool       `gorm:"default:true" json:"can_view"`
	CanDownload  bool       `gorm:"default:true" json:"can_download"`
	CanPrint     bool       `gorm:"default:false" json:"can_print"`
	CanEdit      bool       `gorm:"default:false" json:"can_edit"`
	CanShare     bool       `gorm:"default:false" json:"can_share"`

	// Validity
	SharedAt     time.Time  `gorm:"autoCreateTime" json:"shared_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// Access
	AccessToken  string     `gorm:"size:255;uniqueIndex" json:"-"` // For external share links
	AccessCount  int        `gorm:"default:0" json:"access_count"`
	MaxAccess    int        `gorm:"default:0" json:"max_access"` // 0 = unlimited

	// Message
	ShareMessage string     `gorm:"type:text" json:"share_message"`

	// Relationships
	Document     *Document  `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

func (SharedDocument) TableName() string {
	return "shared_documents"
}
