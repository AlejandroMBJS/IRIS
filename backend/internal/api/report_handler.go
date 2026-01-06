/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/report_handler.go
==============================================================================

DESCRIPTION:
    Handles report generation and history endpoints. Generates various
    payroll reports and maintains a history of generated reports.

USER PERSPECTIVE:
    - Generate payroll summary reports
    - Export reports in different formats
    - View history of previously generated reports

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new report types
    ⚠️  CAUTION: Report data accuracy, date ranges
    ❌  DO NOT modify: Existing report formats in use
    📝  Consider adding PDF/Excel export

SYNTAX EXPLANATION:
    - GenerateReport: Creates report based on request params
    - GetReportHistory: Returns list of past reports
    - PayrollReportRequest: Specifies period, format, filters

ENDPOINTS:
    POST /reports/generate - Generate new report
    GET  /reports/history - List past reports

REPORT TYPES (Future):
    - Payroll summary by period
    - Employee earnings summary
    - Tax withholdings report
    - IMSS contributions report

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/dtos"
	"backend/internal/services"
)

type ReportHandler struct {
	service *services.ReportService
}

func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/reports")
	group.POST("/generate", h.GenerateReport)
	group.GET("/history", h.GetReportHistory)
}

func (h *ReportHandler) GetReportHistory(c *gin.Context) {
	history, err := h.service.GetReportHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve report history"})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *ReportHandler) GenerateReport(c *gin.Context) {
	var req dtos.PayrollReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.GenerateReport(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	// For now, just return the JSON summary.
	// In a real app, you would convert this to the requested format (PDF, etc.)
	c.JSON(http.StatusOK, report)
}
