package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend/internal/services"
)

type ExpenseHandler struct {
	expenseService *services.ExpenseService
}

func NewExpenseHandler(expenseService *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

// Expense Category handlers
func (h *ExpenseHandler) CreateCategory(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateExpenseCategoryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	category, err := h.expenseService.CreateExpenseCategory(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *ExpenseHandler) ListCategories(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	categories, err := h.expenseService.GetExpenseCategories(companyID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Expense Report handlers
func (h *ExpenseHandler) CreateExpenseReport(c *gin.Context) {
	var dto services.CreateExpenseReportDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.expenseService.CreateExpenseReport(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, report)
}

func (h *ExpenseHandler) GetExpenseReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	report, err := h.expenseService.GetExpenseReportByID(reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ExpenseHandler) ListExpenseReports(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	employeeIDStr := c.Query("employee_id")

	filters := services.ExpenseReportFilters{
		CompanyID: companyID.(uuid.UUID),
		Status:    status,
		Page:      page,
		Limit:     limit,
	}
	if employeeIDStr != "" {
		if id, err := uuid.Parse(employeeIDStr); err == nil {
			filters.EmployeeID = &id
		}
	}

	result, err := h.expenseService.ListExpenseReports(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExpenseHandler) SubmitExpenseReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	report, err := h.expenseService.SubmitExpenseReport(reportID, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ExpenseHandler) ApproveExpenseReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	userID, _ := c.Get("user_id")
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	report, err := h.expenseService.ApproveExpenseReport(reportID, userID.(uuid.UUID), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ExpenseHandler) RejectExpenseReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	userID, _ := c.Get("user_id")
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.expenseService.RejectExpenseReport(reportID, userID.(uuid.UUID), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ExpenseHandler) ProcessReimbursement(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	var req struct {
		Method          string     `json:"method"`
		PayrollPeriodID *uuid.UUID `json:"payroll_period_id"`
	}
	c.ShouldBindJSON(&req)
	report, err := h.expenseService.ProcessReimbursement(reportID, req.Method, req.PayrollPeriodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ExpenseHandler) MarkExpenseReportPaid(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	var req struct {
		PaymentDate time.Time `json:"payment_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.PaymentDate = time.Now()
	}
	report, err := h.expenseService.MarkAsPaid(reportID, req.PaymentDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// Expense Item handlers
func (h *ExpenseHandler) AddExpenseItem(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}
	var dto services.CreateExpenseItemDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.ReportID = reportID
	item, err := h.expenseService.AddExpenseItem(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ExpenseHandler) GetExpenseItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}
	item, err := h.expenseService.GetExpenseItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ExpenseHandler) DeleteExpenseItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}
	if err := h.expenseService.DeleteExpenseItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Expense item deleted"})
}

// Advance Payment handlers
func (h *ExpenseHandler) RequestAdvancePayment(c *gin.Context) {
	var dto services.CreateAdvancePaymentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	advance, err := h.expenseService.RequestAdvancePayment(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, advance)
}

func (h *ExpenseHandler) GetEmployeeAdvancePayments(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	advances, err := h.expenseService.GetEmployeeAdvances(employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, advances)
}

func (h *ExpenseHandler) ApproveAdvancePayment(c *gin.Context) {
	advanceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advance ID"})
		return
	}
	userID, _ := c.Get("user_id")
	advance, err := h.expenseService.ApproveAdvancePayment(advanceID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, advance)
}

func (h *ExpenseHandler) IssueAdvancePayment(c *gin.Context) {
	advanceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advance ID"})
		return
	}
	var req struct {
		Method string `json:"method"`
	}
	c.ShouldBindJSON(&req)
	advance, err := h.expenseService.IssueAdvancePayment(advanceID, req.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, advance)
}

func (h *ExpenseHandler) ReconcileAdvance(c *gin.Context) {
	advanceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advance ID"})
		return
	}
	var req struct {
		ExpenseReportID uuid.UUID `json:"expense_report_id" binding:"required"`
		AmountSpent     float64   `json:"amount_spent"`
		AmountReturned  float64   `json:"amount_returned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	advance, err := h.expenseService.ReconcileAdvance(advanceID, req.ExpenseReportID, req.AmountSpent, req.AmountReturned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, advance)
}
