/*
Package models - IRIS Payroll System Expense Management Module Models

==============================================================================
FILE: internal/models/expense.go
==============================================================================

DESCRIPTION:
    Data models for Expense Management including expense reports, receipts,
    approval workflows, and reimbursements.

==============================================================================
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ExpenseReportStatus represents the status of an expense report
type ExpenseReportStatus string

const (
	ExpenseReportDraft     ExpenseReportStatus = "draft"
	ExpenseReportSubmitted ExpenseReportStatus = "submitted"
	ExpenseReportApproved  ExpenseReportStatus = "approved"
	ExpenseReportRejected  ExpenseReportStatus = "rejected"
	ExpenseReportProcessed ExpenseReportStatus = "processed"
	ExpenseReportPaid      ExpenseReportStatus = "paid"
)

// ExpenseCategory represents an expense category
type ExpenseCategory struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`

	Name         string     `gorm:"size:255;not null" json:"name"`
	Code         string     `gorm:"size:50;index" json:"code"`
	Description  string     `gorm:"type:text" json:"description"`

	// Parent Category
	ParentID     *uuid.UUID `gorm:"type:text;index" json:"parent_id,omitempty"`

	// Accounting
	GLCode       string     `gorm:"size:50" json:"gl_code"` // General Ledger code
	CostCenterID *uuid.UUID `gorm:"type:text" json:"cost_center_id,omitempty"`

	// Limits
	MaxAmount    float64    `gorm:"default:0" json:"max_amount"` // 0 = no limit
	RequiresReceipt bool    `gorm:"default:true" json:"requires_receipt"`
	RequiresApproval bool   `gorm:"default:true" json:"requires_approval"`

	// Tax
	TaxDeductible bool      `gorm:"default:true" json:"tax_deductible"`
	TaxRate      float64    `gorm:"default:0" json:"tax_rate"` // e.g., 16 for IVA

	// Settings
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`
	Icon         string     `gorm:"size:50" json:"icon"`
	Color        string     `gorm:"size:7" json:"color"`

	// Relationships
	Parent       *ExpenseCategory  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []ExpenseCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (ExpenseCategory) TableName() string {
	return "expense_categories"
}

// ExpensePolicy represents company expense policies
type ExpensePolicy struct {
	BaseModel
	CompanyID    uuid.UUID      `gorm:"type:text;not null;index" json:"company_id"`

	Name         string         `gorm:"size:255;not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`

	// Limits
	DailyLimit   float64        `gorm:"default:0" json:"daily_limit"` // 0 = no limit
	MonthlyLimit float64        `gorm:"default:0" json:"monthly_limit"`
	PerTransactionLimit float64 `gorm:"default:0" json:"per_transaction_limit"`

	// Approval Thresholds
	AutoApproveThreshold float64 `gorm:"default:0" json:"auto_approve_threshold"` // Auto-approve below this amount
	ManagerApprovalThreshold float64 `gorm:"default:0" json:"manager_approval_threshold"`
	ExecutiveApprovalThreshold float64 `gorm:"default:0" json:"executive_approval_threshold"`

	// Mileage
	MileageRatePerKm float64    `gorm:"default:0" json:"mileage_rate_per_km"`

	// Meals
	MealDailyLimit float64      `gorm:"default:0" json:"meal_daily_limit"`
	BreakfastLimit float64      `gorm:"default:0" json:"breakfast_limit"`
	LunchLimit     float64      `gorm:"default:0" json:"lunch_limit"`
	DinnerLimit    float64      `gorm:"default:0" json:"dinner_limit"`

	// Travel
	HotelDailyLimit float64     `gorm:"default:0" json:"hotel_daily_limit"`
	FlightClass     string      `gorm:"size:50" json:"flight_class"` // economy, business

	// Receipt Requirements
	ReceiptRequiredThreshold float64 `gorm:"default:0" json:"receipt_required_threshold"`
	RequireOriginalReceipts  bool    `gorm:"default:false" json:"require_original_receipts"`

	// Category Limits (JSON)
	CategoryLimits string       `gorm:"type:text" json:"category_limits"` // JSON: {categoryId: maxAmount}

	// Applicability
	ApplicableRoles pq.StringArray `gorm:"type:text[]" json:"applicable_roles"`
	ApplicablePositions pq.StringArray `gorm:"type:text[]" json:"applicable_positions"`

	IsDefault    bool           `gorm:"default:false" json:"is_default"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`

	CreatedByID  *uuid.UUID     `gorm:"type:text" json:"created_by_id,omitempty"`
}

func (ExpensePolicy) TableName() string {
	return "expense_policies"
}

// ExpenseReport represents an expense report submitted by an employee
type ExpenseReport struct {
	BaseModel
	CompanyID    uuid.UUID           `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID           `gorm:"type:text;not null;index" json:"employee_id"`

	// Report Info
	ReportNumber string              `gorm:"size:50;uniqueIndex" json:"report_number"`
	Title        string              `gorm:"size:255;not null" json:"title"`
	Description  string              `gorm:"type:text" json:"description"`

	// Purpose
	Purpose      string              `gorm:"size:255" json:"purpose"` // business_travel, client_meeting, etc.
	ProjectID    *uuid.UUID          `gorm:"type:text" json:"project_id,omitempty"`
	ClientName   string              `gorm:"size:255" json:"client_name"`

	// Period
	StartDate    time.Time           `gorm:"not null" json:"start_date"`
	EndDate      time.Time           `gorm:"not null" json:"end_date"`

	// Totals
	TotalAmount  float64             `gorm:"default:0" json:"total_amount"`
	TotalTax     float64             `gorm:"default:0" json:"total_tax"`
	TotalApproved float64            `gorm:"default:0" json:"total_approved"`
	TotalRejected float64            `gorm:"default:0" json:"total_rejected"`

	// Currency
	Currency     string              `gorm:"size:3;default:'MXN'" json:"currency"`

	// Status
	Status       ExpenseReportStatus `gorm:"size:50;default:'draft'" json:"status"`

	// Submission
	SubmittedAt  *time.Time          `json:"submitted_at,omitempty"`

	// Approval Chain
	CurrentApproverID *uuid.UUID     `gorm:"type:text" json:"current_approver_id,omitempty"`
	ApprovalLevel int               `gorm:"default:1" json:"approval_level"` // 1 = manager, 2 = director, etc.

	// Final Approval
	FinalApprovedByID *uuid.UUID    `gorm:"type:text" json:"final_approved_by_id,omitempty"`
	FinalApprovedAt *time.Time      `json:"final_approved_at,omitempty"`

	// Rejection
	RejectedByID *uuid.UUID         `gorm:"type:text" json:"rejected_by_id,omitempty"`
	RejectedAt   *time.Time         `json:"rejected_at,omitempty"`
	RejectionReason string          `gorm:"type:text" json:"rejection_reason"`

	// Reimbursement
	ReimbursementStatus string      `gorm:"size:50" json:"reimbursement_status"` // pending, scheduled, paid
	ReimbursementDate *time.Time    `json:"reimbursement_date,omitempty"`
	ReimbursementMethod string      `gorm:"size:50" json:"reimbursement_method"` // payroll, direct_deposit, check
	PayrollPeriodID *uuid.UUID      `gorm:"type:text" json:"payroll_period_id,omitempty"`

	// Notes
	EmployeeNotes string             `gorm:"type:text" json:"employee_notes"`
	ApproverNotes string             `gorm:"type:text" json:"approver_notes"`
	FinanceNotes  string             `gorm:"type:text" json:"finance_notes"`

	// Policy
	PolicyID     *uuid.UUID          `gorm:"type:text" json:"policy_id,omitempty"`
	ViolationsCount int              `gorm:"default:0" json:"violations_count"`

	// Relationships
	Employee     *Employee           `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Items        []ExpenseItem       `gorm:"foreignKey:ReportID" json:"items,omitempty"`
	ApprovalHistory []ExpenseApproval `gorm:"foreignKey:ReportID" json:"approval_history,omitempty"`
}

func (ExpenseReport) TableName() string {
	return "expense_reports"
}

// ExpenseItem represents an individual expense line item
type ExpenseItem struct {
	BaseModel
	ReportID     uuid.UUID  `gorm:"type:text;not null;index" json:"report_id"`
	CategoryID   uuid.UUID  `gorm:"type:text;not null;index" json:"category_id"`

	// Expense Info
	ExpenseDate  time.Time  `gorm:"not null" json:"expense_date"`
	Description  string     `gorm:"type:text;not null" json:"description"`
	Merchant     string     `gorm:"size:255" json:"merchant"`
	MerchantCity string     `gorm:"size:100" json:"merchant_city"`

	// Amount
	Amount       float64    `gorm:"not null" json:"amount"`
	TaxAmount    float64    `gorm:"default:0" json:"tax_amount"`
	TipAmount    float64    `gorm:"default:0" json:"tip_amount"`
	TotalAmount  float64    `gorm:"not null" json:"total_amount"`
	Currency     string     `gorm:"size:3;default:'MXN'" json:"currency"`

	// Exchange (if foreign currency)
	OriginalAmount   float64 `gorm:"default:0" json:"original_amount"`
	OriginalCurrency string  `gorm:"size:3" json:"original_currency"`
	ExchangeRate     float64 `gorm:"default:1" json:"exchange_rate"`

	// Mileage (if applicable)
	IsMileage    bool       `gorm:"default:false" json:"is_mileage"`
	Distance     float64    `gorm:"default:0" json:"distance"` // km
	MileageRate  float64    `gorm:"default:0" json:"mileage_rate"`
	StartLocation string    `gorm:"size:255" json:"start_location"`
	EndLocation  string     `gorm:"size:255" json:"end_location"`

	// Attendees (for meals/entertainment)
	AttendeeCount int       `gorm:"default:1" json:"attendee_count"`
	AttendeeNames string    `gorm:"type:text" json:"attendee_names"` // Comma-separated or JSON

	// Receipt
	ReceiptURL   string     `gorm:"size:500" json:"receipt_url"`
	ReceiptVerified bool    `gorm:"default:false" json:"receipt_verified"`
	ReceiptMissing bool     `gorm:"default:false" json:"receipt_missing"`
	ReceiptMissingReason string `gorm:"size:255" json:"receipt_missing_reason"`

	// Status
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, approved, rejected
	RejectionReason string  `gorm:"size:255" json:"rejection_reason"`

	// Policy Compliance
	ExceedsLimit bool       `gorm:"default:false" json:"exceeds_limit"`
	PolicyViolation string  `gorm:"size:255" json:"policy_violation"`

	// GL Coding
	GLCode       string     `gorm:"size:50" json:"gl_code"`
	CostCenterID *uuid.UUID `gorm:"type:text" json:"cost_center_id,omitempty"`

	// Audit
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovedAmount float64  `gorm:"default:0" json:"approved_amount"`

	DisplayOrder int        `gorm:"default:0" json:"display_order"`

	// Relationships
	Report       *ExpenseReport   `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	Category     *ExpenseCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (ExpenseItem) TableName() string {
	return "expense_items"
}

// ExpenseApproval tracks approval history for expense reports
type ExpenseApproval struct {
	BaseModel
	ReportID     uuid.UUID  `gorm:"type:text;not null;index" json:"report_id"`
	ApproverID   uuid.UUID  `gorm:"type:text;not null;index" json:"approver_id"`

	// Approval Level
	Level        int        `gorm:"default:1" json:"level"` // 1 = manager, 2 = director, etc.
	LevelName    string     `gorm:"size:100" json:"level_name"`

	// Action
	Action       string     `gorm:"size:50;not null" json:"action"` // approved, rejected, returned
	ActionAt     time.Time  `gorm:"autoCreateTime" json:"action_at"`

	// Details
	Comments     string     `gorm:"type:text" json:"comments"`
	AmountApproved float64  `gorm:"default:0" json:"amount_approved"`
	AmountRejected float64  `gorm:"default:0" json:"amount_rejected"`

	// Items affected
	ItemsApproved  int      `gorm:"default:0" json:"items_approved"`
	ItemsRejected  int      `gorm:"default:0" json:"items_rejected"`
	ItemsModified  string   `gorm:"type:text" json:"items_modified"` // JSON: [{itemId, oldAmount, newAmount}]

	// Relationships
	Report       *ExpenseReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`
}

func (ExpenseApproval) TableName() string {
	return "expense_approvals"
}

// CorporateCard represents a company credit card
type CorporateCard struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Card Info
	CardNumber   string     `gorm:"size:20" json:"card_number"` // Last 4 digits only
	CardType     string     `gorm:"size:50" json:"card_type"` // visa, mastercard, amex
	Issuer       string     `gorm:"size:100" json:"issuer"` // Bank name
	ExpiryDate   time.Time  `gorm:"not null" json:"expiry_date"`

	// Limits
	CreditLimit  float64    `gorm:"default:0" json:"credit_limit"`
	MonthlyLimit float64    `gorm:"default:0" json:"monthly_limit"`

	// Status
	Status       string     `gorm:"size:50;default:'active'" json:"status"` // active, suspended, cancelled
	ActivatedAt  *time.Time `json:"activated_at,omitempty"`
	SuspendedAt  *time.Time `json:"suspended_at,omitempty"`
	SuspendReason string    `gorm:"size:255" json:"suspend_reason"`

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Transactions []CardTransaction `gorm:"foreignKey:CardID" json:"transactions,omitempty"`
}

func (CorporateCard) TableName() string {
	return "expense_corporate_cards"
}

// CardTransaction represents a credit card transaction
type CardTransaction struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	CardID       uuid.UUID  `gorm:"type:text;not null;index" json:"card_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Transaction Info
	TransactionDate time.Time `gorm:"not null" json:"transaction_date"`
	PostDate     time.Time  `json:"post_date"`
	MerchantName string     `gorm:"size:255" json:"merchant_name"`
	MerchantCategory string `gorm:"size:100" json:"merchant_category"` // MCC
	MerchantCity string     `gorm:"size:100" json:"merchant_city"`
	MerchantCountry string  `gorm:"size:50" json:"merchant_country"`

	// Amount
	Amount       float64    `gorm:"not null" json:"amount"`
	Currency     string     `gorm:"size:3" json:"currency"`
	OriginalAmount float64  `gorm:"default:0" json:"original_amount"`
	OriginalCurrency string `gorm:"size:3" json:"original_currency"`

	// Reference
	AuthCode     string     `gorm:"size:50" json:"auth_code"`
	ReferenceNumber string  `gorm:"size:100" json:"reference_number"`

	// Status
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, reconciled, disputed

	// Reconciliation
	ExpenseItemID *uuid.UUID `gorm:"type:text" json:"expense_item_id,omitempty"`
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`
	ReconciledByID *uuid.UUID `gorm:"type:text" json:"reconciled_by_id,omitempty"`

	// Relationships
	Card         *CorporateCard `gorm:"foreignKey:CardID" json:"card,omitempty"`
	Employee     *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (CardTransaction) TableName() string {
	return "expense_card_transactions"
}

// AdvancePayment represents a cash advance given to an employee
type AdvancePayment struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:text;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID  `gorm:"type:text;not null;index" json:"employee_id"`

	// Advance Info
	RequestDate  time.Time  `gorm:"not null" json:"request_date"`
	Amount       float64    `gorm:"not null" json:"amount"`
	Currency     string     `gorm:"size:3;default:'MXN'" json:"currency"`
	Purpose      string     `gorm:"type:text;not null" json:"purpose"`
	ProjectID    *uuid.UUID `gorm:"type:text" json:"project_id,omitempty"`

	// Trip Info (if travel advance)
	TripStartDate *time.Time `json:"trip_start_date,omitempty"`
	TripEndDate  *time.Time  `json:"trip_end_date,omitempty"`
	Destination  string      `gorm:"size:255" json:"destination"`

	// Status
	Status       string     `gorm:"size:50;default:'pending'" json:"status"` // pending, approved, issued, reconciled

	// Approval
	ApprovedByID *uuid.UUID `gorm:"type:text" json:"approved_by_id,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`

	// Payment
	IssuedAt     *time.Time `json:"issued_at,omitempty"`
	PaymentMethod string    `gorm:"size:50" json:"payment_method"` // cash, bank_transfer

	// Reconciliation
	ReconciliationDue time.Time `json:"reconciliation_due"`
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`
	ExpenseReportID *uuid.UUID `gorm:"type:text" json:"expense_report_id,omitempty"`

	// Balance
	AmountSpent  float64    `gorm:"default:0" json:"amount_spent"`
	AmountReturned float64  `gorm:"default:0" json:"amount_returned"`
	Balance      float64    `gorm:"default:0" json:"balance"` // Negative = owes company

	// Relationships
	Employee     *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (AdvancePayment) TableName() string {
	return "expense_advance_payments"
}
