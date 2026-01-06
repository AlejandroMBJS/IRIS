/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/payroll_export_handler.go
==============================================================================

DESCRIPTION:
    Handles dual Excel export for payroll processing. Provides endpoints for
    downloading ZIP file containing Vacaciones.xlsx and Faltas_y_Extras.xlsx,
    plus preview endpoint for UI display.

USER PERSPECTIVE:
    - Select payroll period and download dual Excel export
    - Preview export counts before downloading
    - See warnings for late approvals and rejected incidences

DEVELOPER GUIDELINES:
    ✅  OK to modify: Response formats, additional preview data
    ⚠️  CAUTION: File download headers, ZIP compression
    ❌  DO NOT modify: Excel format structure (payroll system expects exact structure)
    📝  Always filter excluded_from_payroll = false

ENDPOINTS:
    GET /payroll-export/dual/:periodID - Download ZIP with both Excel files
    GET /payroll-export/preview/:periodID - Get export preview (counts)

==============================================================================
*/
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/services"
)

// PayrollExportHandler handles dual Excel export endpoints
type PayrollExportHandler struct {
	excelExportService *services.ExcelExportService
}

// NewPayrollExportHandler creates new payroll export handler
func NewPayrollExportHandler(excelExportService *services.ExcelExportService) *PayrollExportHandler {
	return &PayrollExportHandler{excelExportService: excelExportService}
}

// RegisterRoutes registers payroll export routes
func (h *PayrollExportHandler) RegisterRoutes(router *gin.RouterGroup) {
	export := router.Group("/payroll-export")
	{
		export.GET("/dual/:periodID", h.GetDualExport)
		export.GET("/preview/:periodID", h.GetExportPreview)
	}
}

// GetDualExport handles downloading the dual Excel export as ZIP
// @Summary Download dual Excel export
// @Description Downloads ZIP file containing Vacaciones.xlsx and Faltas_y_Extras.xlsx
// @Tags Payroll Export
// @Produce application/zip
// @Security BearerAuth
// @Param periodID path string true "Payroll Period ID (UUID)"
// @Success 200 {file} file "ZIP file with Excel templates"
// @Failure 400 {object} map[string]string "Invalid period ID"
// @Failure 404 {object} map[string]string "Period not found or no incidences"
// @Failure 500 {object} map[string]string "Export generation failed"
// @Router /payroll-export/dual/{periodID} [get]
func (h *PayrollExportHandler) GetDualExport(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("periodID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Generate dual export (ZIP with both Excel files)
	zipBuffer, err := h.excelExportService.GenerateDualExport(periodID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to generate Excel export", "message": err.Error()})
		return
	}

	// Set ZIP download headers
	// Filename format: payroll_export_{periodID}.zip
	fileName := "payroll_export_" + periodID.String() + ".zip"
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/zip")
	c.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())
}

// GetExportPreview handles fetching export preview data
// @Summary Get export preview
// @Description Returns counts and metadata for the dual Excel export without generating files
// @Tags Payroll Export
// @Produce json
// @Security BearerAuth
// @Param periodID path string true "Payroll Period ID (UUID)"
// @Success 200 {object} map[string]interface{} "Export preview data"
// @Failure 400 {object} map[string]string "Invalid period ID"
// @Failure 404 {object} map[string]string "Period not found"
// @Failure 500 {object} map[string]string "Preview generation failed"
// @Router /payroll-export/preview/{periodID} [get]
func (h *PayrollExportHandler) GetExportPreview(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("periodID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Get preview data (counts, warnings, etc.)
	preview, err := h.excelExportService.GetExportPreview(periodID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to generate export preview", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}
