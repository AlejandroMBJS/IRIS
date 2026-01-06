/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/payroll_handler.go
==============================================================================

DESCRIPTION:
    Handles all payroll calculation and payment endpoints: individual and
    bulk calculations, approvals, payments, and payslip generation.

USER PERSPECTIVE:
    - Calculate payroll for individual employees or all at once
    - View calculated payroll by period
    - Approve payroll calculations
    - Process payments
    - Generate and download payslips (PDF, XML/CFDI, HTML)

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new payroll reporting endpoints
    ⚠️  CAUTION: Calculation logic, approval workflow
    ❌  DO NOT modify: CFDI XML structure (SAT compliance)
    📝  All monetary amounts use decimal(15,2)

SYNTAX EXPLANATION:
    - CalculatePayroll: Computes taxes, deductions, net pay
    - GeneratePayslip: Creates PDF/XML output
    - ApprovePayroll: Transitions period to approved status
    - ProcessPayment: Marks payroll as paid

ENDPOINTS:
    POST /payroll/calculate - Calculate single employee payroll
    POST /payroll/bulk-calculate - Calculate for multiple employees
    GET  /payroll/calculation/:period_id/:employee_id - Get calculation
    GET  /payroll/period/:period_id - Get all calculations for period
    GET  /payroll/payslip/:periodId/:employeeId - Download payslip
    GET  /payroll/summary/:periodId - Get period summary
    POST /payroll/approve/:periodId - Approve payroll period
    POST /payroll/payment/:periodId - Process payment
    GET  /payroll/payment/:period_id - Get payment status
    GET  /payroll/concept-totals/:periodId - Get totals by concept

PAYSLIP FORMATS:
    - PDF: Human-readable receipt with all details
    - XML: CFDI Nomina 1.2 compliant for SAT
    - HTML: Web-viewable format

==============================================================================
*/
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/dtos"
	"backend/internal/middleware"
	"backend/internal/services"
)

// PayrollHandler handles payroll endpoints
type PayrollHandler struct {
    payrollService *services.PayrollService
}

// NewPayrollHandler creates new payroll handler
func NewPayrollHandler(payrollService *services.PayrollService) *PayrollHandler {
    return &PayrollHandler{payrollService: payrollService}
}

// RegisterRoutes registers payroll routes
func (h *PayrollHandler) RegisterRoutes(router *gin.RouterGroup) {
    payroll := router.Group("/payroll")
    {
        payroll.POST("/calculate", h.CalculatePayroll)
        payroll.POST("/bulk-calculate", h.BulkCalculatePayroll)
        payroll.GET("/calculation/:period_id/:employee_id", h.GetPayrollCalculation)
        payroll.GET("/period/:period_id", h.GetPayrollByPeriod)
        payroll.GET("/payslip/:periodId/:employeeId", h.GetPayslip)
        payroll.GET("/summary/:periodId", h.GetPayrollSummary)
        payroll.POST("/approve/:periodId", h.ApprovePayroll)
        payroll.POST("/payment/:periodId", h.ProcessPayment)
        payroll.GET("/payment/:period_id", h.GetPaymentStatus)
        payroll.GET("/concept-totals/:periodId", h.GetConceptTotals)
    }
}

// GetPayrollCalculation handles fetching a single payroll calculation.
func (h *PayrollHandler) GetPayrollCalculation(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employee_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	periodID, err := uuid.Parse(c.Param("period_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	payrollCalculation, err := h.payrollService.GetPayrollCalculation(employeeID, periodID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to get payroll calculation", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payrollCalculation)
}

// GetConceptTotals handles fetching payroll concept totals for a period.
func (h *PayrollHandler) GetConceptTotals(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("periodId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	totals, err := h.payrollService.GetConceptTotals(periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get concept totals", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, totals)
}

// ApprovePayroll handles payroll approval
func (h *PayrollHandler) ApprovePayroll(c *gin.Context) {
    periodID, err := uuid.Parse(c.Param("periodId"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
        return
    }

    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    err = h.payrollService.ApprovePayroll(periodID, userID)
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "not found") {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "already approved") {
            status = http.StatusBadRequest
        }
        c.JSON(status, gin.H{"error": "Failed to approve payroll", "message": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Payroll period approved successfully"})
}

func (h *PayrollHandler) GetPayrollSummary(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("periodId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	summary, err := h.payrollService.GetPayrollSummary(periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payroll summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// CalculatePayroll handles payroll calculation
// @Summary Calculate payroll
// @Description Calculate payroll for an employee
// @Tags Payroll
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.PayrollCalculationRequest true "Payroll calculation request"
// @Success 200 {object} dtos.PayrollCalculationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payroll/calculate [post]
func (h *PayrollHandler) CalculatePayroll(c *gin.Context) {
    var req dtos.PayrollCalculationRequest
    
    // Get user from context
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    result, err := h.payrollService.CalculatePayroll(
        req.EmployeeID,
        req.PayrollPeriodID,
        req.CalculateSDI,
        userID,
    )
    
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "not found") {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "not open") {
            status = http.StatusBadRequest
        } else if strings.Contains(err.Error(), "must be approved") {
            status = http.StatusBadRequest
        }
        
        c.JSON(status, gin.H{
            "error":   "Payroll Calculation Failed",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// BulkCalculatePayroll handles bulk payroll calculation
// @Summary Bulk calculate payroll
// @Description Calculate payroll for multiple employees
// @Tags Payroll
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.PayrollBulkCalculateRequest true "Bulk calculation request"
// @Success 200 {object} dtos.PayrollBulkCalculationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payroll/bulk-calculate [post]
func (h *PayrollHandler) BulkCalculatePayroll(c *gin.Context) {
    var req dtos.PayrollBulkCalculateRequest
    
    // Get user from context
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    result, err := h.payrollService.BulkCalculatePayroll(
        req.PayrollPeriodID,
        req.EmployeeIDs,
        req.CalculateAll,
        userID,
    )
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Bulk Calculation Failed",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// PostGeneratePayslip handles payslip generation via POST
// @Summary Generate payslip
// @Description Generate payslip in specified format
// @Tags Payroll
// @Accept json
// @Produce octet-stream
// @Security BearerAuth
// @Param request body dtos.PayslipRequest true "Payslip request"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payroll/payslip [post]
func (h *PayrollHandler) PostGeneratePayslip(c *gin.Context) {
    var req dtos.PayslipRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    content, err := h.payrollService.GeneratePayslip(
        req.EmployeeID,
        req.PayrollPeriodID,
        req.Format,
    )
    
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "not found") {
            status = http.StatusNotFound
        }
        
        c.JSON(status, gin.H{
            "error":   "Payslip Generation Failed",
            "message": err.Error(),
        })
        return
    }
    
    // Set appropriate headers
    contentType := "application/octet-stream"
    fileName := "payslip"
    
    switch req.Format {
    case "pdf":
        contentType = "application/pdf"
        fileName = "recibo_nomina.pdf"
    case "xml":
        contentType = "application/xml"
        fileName = "cfdi_nomina.xml"
    case "html":
        contentType = "text/html"
        fileName = "recibo_nomina.html"
    }
    
    c.Header("Content-Disposition", "attachment; filename="+fileName)
    c.Data(http.StatusOK, contentType, content)
}

func (h *PayrollHandler) GetPayrollByPeriod(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("period_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	calculations, err := h.payrollService.GetPayrollByPeriod(periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payroll calculations for period", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, calculations)
}

// GetPayslip handles fetching payslip
func (h *PayrollHandler) GetPayslip(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	periodID, err := uuid.Parse(c.Param("periodId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Get format from query parameter, default to PDF
	format := c.DefaultQuery("format", "pdf")
	if format != "pdf" && format != "xml" && format != "html" {
		format = "pdf"
	}

	// Get payroll calculation for filename metadata
	payrollCalc, err := h.payrollService.GetPayrollCalculation(employeeID, periodID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to get payroll calculation", "message": err.Error()})
		return
	}

	content, err := h.payrollService.GeneratePayslip(employeeID, periodID, format)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to generate payslip", "message": err.Error()})
		return
	}

	// Build filename with employee number and payment date
	// Format: {employee_number}_{payment_date}.{ext}
	employeeNumber := payrollCalc.EmployeeNumber
	paymentDate := payrollCalc.PaymentDate.Format("2006-01-02")

	// Set appropriate headers based on format
	switch format {
	case "xml":
		fileName := employeeNumber + "_" + paymentDate + "_cfdi.xml"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Data(http.StatusOK, "application/xml", content)
	case "html":
		fileName := employeeNumber + "_" + paymentDate + "_recibo.html"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Data(http.StatusOK, "text/html", content)
	default:
		fileName := employeeNumber + "_" + paymentDate + "_recibo.pdf"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Data(http.StatusOK, "application/pdf", content)
	}
}

func (h *PayrollHandler) ProcessPayment(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("periodId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

	err = h.payrollService.ProcessPayment(periodID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "not approved") {
            status = http.StatusBadRequest
        } else if strings.Contains(err.Error(), "no payroll calculations found") {
            status = http.StatusNotFound
        }
		c.JSON(status, gin.H{"error": "Failed to process payment", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment processed successfully"})
}

func (h *PayrollHandler) GetPaymentStatus(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("period_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	status, err := h.payrollService.GetPaymentStatus(periodID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to retrieve payment status", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}
