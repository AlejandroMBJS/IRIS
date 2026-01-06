/*
Package services - IRIS Expense Management Service

==============================================================================
FILE: internal/services/expense_service.go
==============================================================================

DESCRIPTION:
    Business logic for Expense Management including expense reports, receipts,
    approval workflows, and reimbursements.

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExpenseService provides business logic for expense management
type ExpenseService struct {
	db *gorm.DB
}

// NewExpenseService creates a new ExpenseService
func NewExpenseService(db *gorm.DB) *ExpenseService {
	return &ExpenseService{db: db}
}

// === Expense Category DTOs ===

type CreateExpenseCategoryDTO struct {
	CompanyID        uuid.UUID
	Name             string
	Code             string
	Description      string
	ParentID         *uuid.UUID
	GLCode           string
	CostCenterID     *uuid.UUID
	MaxAmount        float64
	RequiresReceipt  bool
	RequiresApproval bool
	TaxDeductible    bool
	TaxRate          float64
	Icon             string
	Color            string
}

// CreateExpenseCategory creates an expense category
func (s *ExpenseService) CreateExpenseCategory(dto CreateExpenseCategoryDTO) (*models.ExpenseCategory, error) {
	if dto.Name == "" {
		return nil, errors.New("category name is required")
	}

	category := &models.ExpenseCategory{
		CompanyID:        dto.CompanyID,
		Name:             dto.Name,
		Code:             dto.Code,
		Description:      dto.Description,
		ParentID:         dto.ParentID,
		GLCode:           dto.GLCode,
		CostCenterID:     dto.CostCenterID,
		MaxAmount:        dto.MaxAmount,
		RequiresReceipt:  dto.RequiresReceipt,
		RequiresApproval: dto.RequiresApproval,
		TaxDeductible:    dto.TaxDeductible,
		TaxRate:          dto.TaxRate,
		Icon:             dto.Icon,
		Color:            dto.Color,
		IsActive:         true,
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// GetExpenseCategories gets all expense categories for a company
func (s *ExpenseService) GetExpenseCategories(companyID uuid.UUID) ([]models.ExpenseCategory, error) {
	var categories []models.ExpenseCategory
	if err := s.db.Where("company_id = ? AND is_active = ?", companyID, true).
		Preload("Children").
		Where("parent_id IS NULL").
		Order("display_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// === Expense Report DTOs ===

type CreateExpenseReportDTO struct {
	CompanyID   uuid.UUID
	EmployeeID  uuid.UUID
	Title       string
	Description string
	Purpose     string
	ProjectID   *uuid.UUID
	ClientName  string
	StartDate   time.Time
	EndDate     time.Time
	Currency    string
	PolicyID    *uuid.UUID
}

type ExpenseReportFilters struct {
	CompanyID  uuid.UUID
	EmployeeID *uuid.UUID
	Status     string
	StartDate  *time.Time
	EndDate    *time.Time
	Search     string
	Page       int
	Limit      int
}

type PaginatedExpenseReports struct {
	Data       []models.ExpenseReport `json:"data"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// CreateExpenseReport creates an expense report
func (s *ExpenseService) CreateExpenseReport(dto CreateExpenseReportDTO) (*models.ExpenseReport, error) {
	if dto.Title == "" {
		return nil, errors.New("report title is required")
	}

	// Generate report number
	reportNumber := fmt.Sprintf("EXP-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])

	report := &models.ExpenseReport{
		CompanyID:    dto.CompanyID,
		EmployeeID:   dto.EmployeeID,
		ReportNumber: reportNumber,
		Title:        dto.Title,
		Description:  dto.Description,
		Purpose:      dto.Purpose,
		ProjectID:    dto.ProjectID,
		ClientName:   dto.ClientName,
		StartDate:    dto.StartDate,
		EndDate:      dto.EndDate,
		Currency:     dto.Currency,
		PolicyID:     dto.PolicyID,
		Status:       models.ExpenseReportDraft,
	}

	if report.Currency == "" {
		report.Currency = "MXN"
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

// GetExpenseReportByID retrieves an expense report by ID
func (s *ExpenseService) GetExpenseReportByID(id uuid.UUID) (*models.ExpenseReport, error) {
	var report models.ExpenseReport
	err := s.db.Preload("Items").Preload("Items.Category").Preload("Employee").Preload("ApprovalHistory").
		First(&report, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense report not found")
		}
		return nil, err
	}
	return &report, nil
}

// ListExpenseReports retrieves expense reports with filters
func (s *ExpenseService) ListExpenseReports(filters ExpenseReportFilters) (*PaginatedExpenseReports, error) {
	query := s.db.Model(&models.ExpenseReport{}).Where("company_id = ?", filters.CompanyID)

	if filters.EmployeeID != nil {
		query = query.Where("employee_id = ?", *filters.EmployeeID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.StartDate != nil {
		query = query.Where("start_date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("end_date <= ?", *filters.EndDate)
	}
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("title LIKE ? OR report_number LIKE ?", searchTerm, searchTerm)
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

	var reports []models.ExpenseReport
	if err := query.Offset(offset).Limit(filters.Limit).Preload("Employee").Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	return &PaginatedExpenseReports{
		Data:       reports,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// SubmitExpenseReport submits an expense report for approval
func (s *ExpenseService) SubmitExpenseReport(id uuid.UUID, notes string) (*models.ExpenseReport, error) {
	report, err := s.GetExpenseReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.Status != models.ExpenseReportDraft {
		return nil, errors.New("report is not in draft status")
	}

	// Validate report has items
	if len(report.Items) == 0 {
		return nil, errors.New("report has no expense items")
	}

	// Calculate totals
	s.calculateReportTotals(report)

	now := time.Now()
	report.Status = models.ExpenseReportSubmitted
	report.SubmittedAt = &now
	report.EmployeeNotes = notes

	// Set manager as approver (find employee's supervisor)
	var employee models.Employee
	if err := s.db.First(&employee, "id = ?", report.EmployeeID).Error; err == nil {
		report.CurrentApproverID = employee.SupervisorID
		report.ApprovalLevel = 1
	}

	if err := s.db.Save(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

// ApproveExpenseReport approves an expense report
func (s *ExpenseService) ApproveExpenseReport(id uuid.UUID, approverID uuid.UUID, notes string) (*models.ExpenseReport, error) {
	report, err := s.GetExpenseReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.Status != models.ExpenseReportSubmitted {
		return nil, errors.New("report is not submitted")
	}

	now := time.Now()

	// Create approval record
	approval := &models.ExpenseApproval{
		ReportID:       id,
		ApproverID:     approverID,
		Level:          report.ApprovalLevel,
		Action:         "approved",
		Comments:       notes,
		AmountApproved: report.TotalAmount,
	}
	s.db.Create(approval)

	// Check if needs higher approval
	if report.TotalAmount > 10000 && report.ApprovalLevel == 1 {
		// Escalate to director
		report.ApprovalLevel = 2
		report.ApproverNotes = notes
		// Would set next approver here
	} else {
		report.Status = models.ExpenseReportApproved
		report.FinalApprovedByID = &approverID
		report.FinalApprovedAt = &now
		report.TotalApproved = report.TotalAmount
		report.ApproverNotes = notes
	}

	if err := s.db.Save(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

// RejectExpenseReport rejects an expense report
func (s *ExpenseService) RejectExpenseReport(id uuid.UUID, rejectorID uuid.UUID, reason string) (*models.ExpenseReport, error) {
	report, err := s.GetExpenseReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.Status != models.ExpenseReportSubmitted {
		return nil, errors.New("report is not submitted")
	}

	now := time.Now()

	// Create approval record
	approval := &models.ExpenseApproval{
		ReportID:       id,
		ApproverID:     rejectorID,
		Level:          report.ApprovalLevel,
		Action:         "rejected",
		Comments:       reason,
		AmountRejected: report.TotalAmount,
	}
	s.db.Create(approval)

	report.Status = models.ExpenseReportRejected
	report.RejectedByID = &rejectorID
	report.RejectedAt = &now
	report.RejectionReason = reason

	if err := s.db.Save(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

// ProcessReimbursement marks report as processed for reimbursement
func (s *ExpenseService) ProcessReimbursement(id uuid.UUID, method string, payrollPeriodID *uuid.UUID) (*models.ExpenseReport, error) {
	report, err := s.GetExpenseReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.Status != models.ExpenseReportApproved {
		return nil, errors.New("report is not approved")
	}

	report.Status = models.ExpenseReportProcessed
	report.ReimbursementStatus = "scheduled"
	report.ReimbursementMethod = method
	report.PayrollPeriodID = payrollPeriodID

	if err := s.db.Save(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

// MarkAsPaid marks report as paid
func (s *ExpenseService) MarkAsPaid(id uuid.UUID, paymentDate time.Time) (*models.ExpenseReport, error) {
	report, err := s.GetExpenseReportByID(id)
	if err != nil {
		return nil, err
	}

	report.Status = models.ExpenseReportPaid
	report.ReimbursementStatus = "paid"
	report.ReimbursementDate = &paymentDate

	if err := s.db.Save(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

func (s *ExpenseService) calculateReportTotals(report *models.ExpenseReport) {
	var totalAmount, totalTax float64
	for _, item := range report.Items {
		totalAmount += item.TotalAmount
		totalTax += item.TaxAmount
	}
	report.TotalAmount = totalAmount
	report.TotalTax = totalTax
}

// === Expense Item Methods ===

type CreateExpenseItemDTO struct {
	ReportID     uuid.UUID
	CategoryID   uuid.UUID
	ExpenseDate  time.Time
	Description  string
	Merchant     string
	MerchantCity string
	Amount       float64
	TaxAmount    float64
	TipAmount    float64
	Currency     string
	IsMileage    bool
	Distance     float64
	MileageRate  float64
	StartLocation string
	EndLocation  string
	AttendeeCount int
	AttendeeNames string
	ReceiptURL   string
	GLCode       string
	CostCenterID *uuid.UUID
}

// AddExpenseItem adds an item to an expense report
func (s *ExpenseService) AddExpenseItem(dto CreateExpenseItemDTO) (*models.ExpenseItem, error) {
	if dto.Description == "" {
		return nil, errors.New("description is required")
	}

	// Get report to verify it's in draft
	var report models.ExpenseReport
	if err := s.db.First(&report, "id = ?", dto.ReportID).Error; err != nil {
		return nil, errors.New("expense report not found")
	}
	if report.Status != models.ExpenseReportDraft {
		return nil, errors.New("can only add items to draft reports")
	}

	item := &models.ExpenseItem{
		ReportID:      dto.ReportID,
		CategoryID:    dto.CategoryID,
		ExpenseDate:   dto.ExpenseDate,
		Description:   dto.Description,
		Merchant:      dto.Merchant,
		MerchantCity:  dto.MerchantCity,
		Amount:        dto.Amount,
		TaxAmount:     dto.TaxAmount,
		TipAmount:     dto.TipAmount,
		Currency:      dto.Currency,
		IsMileage:     dto.IsMileage,
		Distance:      dto.Distance,
		MileageRate:   dto.MileageRate,
		StartLocation: dto.StartLocation,
		EndLocation:   dto.EndLocation,
		AttendeeCount: dto.AttendeeCount,
		AttendeeNames: dto.AttendeeNames,
		ReceiptURL:    dto.ReceiptURL,
		GLCode:        dto.GLCode,
		CostCenterID:  dto.CostCenterID,
		Status:        "pending",
	}

	if item.Currency == "" {
		item.Currency = "MXN"
	}
	if item.AttendeeCount == 0 {
		item.AttendeeCount = 1
	}

	// Calculate total
	item.TotalAmount = item.Amount + item.TaxAmount + item.TipAmount

	// Calculate mileage if applicable
	if item.IsMileage && item.Distance > 0 && item.MileageRate > 0 {
		item.Amount = item.Distance * item.MileageRate
		item.TotalAmount = item.Amount
	}

	// Check category limits
	var category models.ExpenseCategory
	if err := s.db.First(&category, "id = ?", dto.CategoryID).Error; err == nil {
		if category.MaxAmount > 0 && item.TotalAmount > category.MaxAmount {
			item.ExceedsLimit = true
			item.PolicyViolation = fmt.Sprintf("Amount exceeds category limit of %.2f", category.MaxAmount)
		}
		item.ReceiptMissing = category.RequiresReceipt && item.ReceiptURL == ""
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	// Update report totals
	s.db.First(&report, "id = ?", dto.ReportID)
	s.calculateReportTotals(&report)
	s.db.Save(&report)

	return item, nil
}

// GetExpenseItemByID retrieves an expense item by ID
func (s *ExpenseService) GetExpenseItemByID(id uuid.UUID) (*models.ExpenseItem, error) {
	var item models.ExpenseItem
	err := s.db.Preload("Category").First(&item, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense item not found")
		}
		return nil, err
	}
	return &item, nil
}

// DeleteExpenseItem deletes an expense item
func (s *ExpenseService) DeleteExpenseItem(id uuid.UUID) error {
	var item models.ExpenseItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return errors.New("expense item not found")
	}

	// Check report is in draft
	var report models.ExpenseReport
	if err := s.db.First(&report, "id = ?", item.ReportID).Error; err == nil {
		if report.Status != models.ExpenseReportDraft {
			return errors.New("can only delete items from draft reports")
		}
	}

	if err := s.db.Delete(&item).Error; err != nil {
		return err
	}

	// Update report totals
	s.db.First(&report, "id = ?", item.ReportID)
	s.calculateReportTotals(&report)
	s.db.Save(&report)

	return nil
}

// === Advance Payment Methods ===

type CreateAdvancePaymentDTO struct {
	CompanyID     uuid.UUID
	EmployeeID    uuid.UUID
	Amount        float64
	Currency      string
	Purpose       string
	ProjectID     *uuid.UUID
	TripStartDate *time.Time
	TripEndDate   *time.Time
	Destination   string
}

// RequestAdvancePayment creates an advance payment request
func (s *ExpenseService) RequestAdvancePayment(dto CreateAdvancePaymentDTO) (*models.AdvancePayment, error) {
	if dto.Amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if dto.Purpose == "" {
		return nil, errors.New("purpose is required")
	}

	// Default reconciliation due date (30 days)
	reconciliationDue := time.Now().AddDate(0, 0, 30)
	if dto.TripEndDate != nil {
		reconciliationDue = dto.TripEndDate.AddDate(0, 0, 7) // 7 days after trip
	}

	advance := &models.AdvancePayment{
		CompanyID:         dto.CompanyID,
		EmployeeID:        dto.EmployeeID,
		RequestDate:       time.Now(),
		Amount:            dto.Amount,
		Currency:          dto.Currency,
		Purpose:           dto.Purpose,
		ProjectID:         dto.ProjectID,
		TripStartDate:     dto.TripStartDate,
		TripEndDate:       dto.TripEndDate,
		Destination:       dto.Destination,
		ReconciliationDue: reconciliationDue,
		Status:            "pending",
		Balance:           dto.Amount,
	}

	if advance.Currency == "" {
		advance.Currency = "MXN"
	}

	if err := s.db.Create(advance).Error; err != nil {
		return nil, err
	}

	return advance, nil
}

// ApproveAdvancePayment approves an advance payment
func (s *ExpenseService) ApproveAdvancePayment(id uuid.UUID, approverID uuid.UUID) (*models.AdvancePayment, error) {
	var advance models.AdvancePayment
	if err := s.db.First(&advance, "id = ?", id).Error; err != nil {
		return nil, errors.New("advance payment not found")
	}

	if advance.Status != "pending" {
		return nil, errors.New("advance is not pending")
	}

	now := time.Now()
	advance.Status = "approved"
	advance.ApprovedByID = &approverID
	advance.ApprovedAt = &now

	if err := s.db.Save(&advance).Error; err != nil {
		return nil, err
	}

	return &advance, nil
}

// IssueAdvancePayment marks advance as issued
func (s *ExpenseService) IssueAdvancePayment(id uuid.UUID, method string) (*models.AdvancePayment, error) {
	var advance models.AdvancePayment
	if err := s.db.First(&advance, "id = ?", id).Error; err != nil {
		return nil, errors.New("advance payment not found")
	}

	if advance.Status != "approved" {
		return nil, errors.New("advance is not approved")
	}

	now := time.Now()
	advance.Status = "issued"
	advance.IssuedAt = &now
	advance.PaymentMethod = method

	if err := s.db.Save(&advance).Error; err != nil {
		return nil, err
	}

	return &advance, nil
}

// ReconcileAdvance reconciles advance with expense report
func (s *ExpenseService) ReconcileAdvance(id uuid.UUID, expenseReportID uuid.UUID, amountSpent, amountReturned float64) (*models.AdvancePayment, error) {
	var advance models.AdvancePayment
	if err := s.db.First(&advance, "id = ?", id).Error; err != nil {
		return nil, errors.New("advance payment not found")
	}

	if advance.Status != "issued" {
		return nil, errors.New("advance is not issued")
	}

	now := time.Now()
	advance.Status = "reconciled"
	advance.ReconciledAt = &now
	advance.ExpenseReportID = &expenseReportID
	advance.AmountSpent = amountSpent
	advance.AmountReturned = amountReturned
	advance.Balance = advance.Amount - amountSpent - amountReturned

	if err := s.db.Save(&advance).Error; err != nil {
		return nil, err
	}

	return &advance, nil
}

// GetEmployeeAdvances gets advance payments for an employee
func (s *ExpenseService) GetEmployeeAdvances(employeeID uuid.UUID, status string) ([]models.AdvancePayment, error) {
	query := s.db.Where("employee_id = ?", employeeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var advances []models.AdvancePayment
	if err := query.Order("request_date DESC").Find(&advances).Error; err != nil {
		return nil, err
	}
	return advances, nil
}
