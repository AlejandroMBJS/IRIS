/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/excel_export_service.go
==============================================================================

DESCRIPTION:
    Handles dual Excel export for payroll processing. Generates two templates:
    1. Vacaciones.xlsx - Vacation days with premium calculation
    2. Faltas_y_Extras.xlsx - Absences, overtime, and other incidences

USER PERSPECTIVE:
    - Download ZIP file with both Excel templates for payroll import
    - Multi-day absences appear as separate rows (one per day)
    - Late approvals flagged in Observaciones column
    - Rejected incidences automatically excluded

DEVELOPER GUIDELINES:
    ✅  OK to modify: Column formatting, calculation formulas
    ⚠️  CAUTION: Date formats (must match payroll system: YYYYMMDD)
    ❌  DO NOT modify: Column order (payroll system expects exact structure)
    📝  Template specifications from user's Excel files

EXCEL TEMPLATE SPECIFICATIONS:

Vacaciones.xlsx (6 columns):
- Empleado: Employee number (text)
- Fecha: Start date (YYYYMMDD format)
- FechaRegreso: Return date (YYYYMMDD format)
- Descrip: Description (optional)
- DiasPago: Business days (decimal, excludes weekends/holidays)
- DiasPrima: Vacation premium (25% of DiasPago)

Faltas_y_Extras.xlsx (7 columns):
- Empleado: Employee number (text)
- Fecha: Date (YYYYMMDD format)
- Tipo: Payroll code (2-9, from tipo_mappings)
- Horas: Hours (decimal, or empty for day-based)
- Motivo: EXTRAS, FALTA, PERHORAS (from tipo_mappings)
- Destino: Destination date for permits (YYYYMMDD, optional)
- Observaciones: Notes, includes late approval flag

==============================================================================
*/
package services

import (
	"archive/zip"
	"backend/internal/models"
	"backend/internal/utils"
	"bytes"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ExcelExportService handles dual Excel export for payroll
type ExcelExportService struct {
	db *gorm.DB
}

// NewExcelExportService creates a new Excel export service
func NewExcelExportService(db *gorm.DB) *ExcelExportService {
	return &ExcelExportService{db: db}
}

// ExportResult contains the generated Excel files
type ExportResult struct {
	VacacionesFile    []byte
	FaltasExtrasFile  []byte
	VacacionesCount   int
	FaltasExtrasCount int
}

// GenerateDualExport generates both Excel files for a payroll period
// Returns a ZIP file containing Vacaciones.xlsx and Faltas_y_Extras.xlsx
func (s *ExcelExportService) GenerateDualExport(payrollPeriodID uuid.UUID) (*bytes.Buffer, error) {
	// Get all approved incidences for this payroll period
	// Exclude incidences marked as excluded_from_payroll (HR/GM rejected)
	var incidences []models.Incidence
	err := s.db.
		Preload("Employee").
		Preload("IncidenceType").
		Preload("IncidenceType.IncidenceCategory").
		Joins("LEFT JOIN incidence_tipo_mappings ON incidence_tipo_mappings.incidence_type_id = incidences.incidence_type_id").
		Where("incidences.payroll_period_id = ?", payrollPeriodID).
		Where("incidences.status = ?", "approved").
		Where("incidences.excluded_from_payroll = ?", false).
		Find(&incidences).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load incidences: %w", err)
	}

	// Separate incidences by template type
	var vacacionesIncidences []models.Incidence
	var faltasExtrasIncidences []models.Incidence

	for _, inc := range incidences {
		// Load tipo mapping to determine template type
		var mapping models.IncidenceTipoMapping
		err := s.db.Where("incidence_type_id = ?", inc.IncidenceTypeID).First(&mapping).Error
		if err != nil {
			// If no mapping found, skip this incidence
			continue
		}

		if mapping.TemplateType == "vacaciones" {
			vacacionesIncidences = append(vacacionesIncidences, inc)
		} else if mapping.TemplateType == "faltas_extras" {
			faltasExtrasIncidences = append(faltasExtrasIncidences, inc)
		}
	}

	// Generate both Excel files
	vacacionesBytes, err := s.GenerateVacacionesExcel(vacacionesIncidences)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Vacaciones.xlsx: %w", err)
	}

	faltasExtrasBytes, err := s.GenerateFaltasExtrasExcel(faltasExtrasIncidences)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Faltas_y_Extras.xlsx: %w", err)
	}

	// Create ZIP file containing both Excel files
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// Add Vacaciones.xlsx to ZIP
	vacacionesFile, err := zipWriter.Create("Vacaciones.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create Vacaciones.xlsx in ZIP: %w", err)
	}
	_, err = vacacionesFile.Write(vacacionesBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to write Vacaciones.xlsx: %w", err)
	}

	// Add Faltas_y_Extras.xlsx to ZIP
	faltasExtrasFile, err := zipWriter.Create("Faltas_y_Extras.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create Faltas_y_Extras.xlsx in ZIP: %w", err)
	}
	_, err = faltasExtrasFile.Write(faltasExtrasBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to write Faltas_y_Extras.xlsx: %w", err)
	}

	// Close ZIP writer
	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close ZIP: %w", err)
	}

	return zipBuffer, nil
}

// GenerateVacacionesExcel generates the Vacaciones.xlsx file
// Columns: Empleado, Fecha, FechaRegreso, Descrip, DiasPago, DiasPrima
func (s *ExcelExportService) GenerateVacacionesExcel(incidences []models.Incidence) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Set headers
	headers := []string{"Empleado", "Fecha", "FechaRegreso", "Descrip", "DiasPago", "DiasPrima"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Process each vacation incidence
	row := 2
	for _, inc := range incidences {
		// Calculate business days (DiasPago)
		year := inc.StartDate.Year()
		calc := utils.NewBusinessDayCalculator(year)
		diasPago := calc.CalculateBusinessDays(inc.StartDate, inc.EndDate)

		// Calculate vacation premium (DiasPrima) - 25% of DiasPago
		diasPrima := diasPago * 0.25

		// Format dates as YYYYMMDD
		fechaStr := inc.StartDate.Format("20060102")
		fechaRegresoStr := inc.EndDate.AddDate(0, 0, 1).Format("20060102") // Return date is day after end date

		// Get employee number
		employeeNumber := ""
		if inc.Employee != nil {
			employeeNumber = inc.Employee.EmployeeNumber
		}

		// Description
		descrip := inc.Comments

		// Set row data
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), employeeNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fechaStr)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fechaRegresoStr)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), descrip)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("%.2f", diasPago))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("%.2f", diasPrima))

		row++
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

// GenerateFaltasExtrasExcel generates the Faltas_y_Extras.xlsx file
// Columns: Empleado, Fecha, Tipo, Horas, Motivo, Destino, Observaciones
// Multi-day absences are expanded into separate rows (one per day)
func (s *ExcelExportService) GenerateFaltasExtrasExcel(incidences []models.Incidence) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Set headers
	headers := []string{"Empleado", "Fecha", "Tipo", "Horas", "Motivo", "Destino", "Observaciones"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Process each incidence
	row := 2
	for _, inc := range incidences {
		// Load tipo mapping
		var mapping models.IncidenceTipoMapping
		err := s.db.Where("incidence_type_id = ?", inc.IncidenceTypeID).First(&mapping).Error
		if err != nil {
			// Skip if no mapping found
			continue
		}

		// Expand multi-day absences into separate rows
		expandedRows := s.expandMultiDayAbsence(inc, mapping)

		for _, expandedRow := range expandedRows {
			// Get employee number
			employeeNumber := ""
			if inc.Employee != nil {
				employeeNumber = inc.Employee.EmployeeNumber
			}

			// Format date as YYYYMMDD
			fechaStr := expandedRow.Date.Format("20060102")

			// Tipo code
			tipoCode := ""
			if mapping.TipoCode != nil {
				tipoCode = *mapping.TipoCode
			}

			// Hours (may be empty for day-based absences)
			horasStr := ""
			if expandedRow.Hours > 0 {
				horasStr = fmt.Sprintf("%.2f", expandedRow.Hours)
			}

			// Motivo
			motivo := ""
			if mapping.Motivo != nil {
				motivo = *mapping.Motivo
			}

			// Destino (optional, for permits)
			destinoStr := ""

			// Observaciones - include late approval flag if applicable
			observaciones := inc.Comments
			if inc.LateApprovalFlag {
				if observaciones != "" {
					observaciones += " | "
				}
				observaciones += "⚠️ APROBACIÓN TARDÍA"
			}

			// Set row data
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), employeeNumber)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fechaStr)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), tipoCode)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), horasStr)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), motivo)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), destinoStr)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), observaciones)

			row++
		}
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

// ExpandedRow represents a single row in the Faltas_y_Extras export
// Multi-day absences are expanded into multiple ExpandedRows (one per day)
type ExpandedRow struct {
	Date  time.Time
	Hours float64
}

// expandMultiDayAbsence expands a multi-day incidence into separate rows (one per day)
// Example: 3-day sick leave (Jan 1-3) → 3 rows with dates Jan 1, Jan 2, Jan 3
func (s *ExcelExportService) expandMultiDayAbsence(inc models.Incidence, mapping models.IncidenceTipoMapping) []ExpandedRow {
	var rows []ExpandedRow

	// Calculate number of days
	days := inc.EndDate.Sub(inc.StartDate).Hours()/24 + 1

	// If single day, return one row
	if days <= 1 {
		hours := inc.Quantity * mapping.HoursMultiplier
		rows = append(rows, ExpandedRow{
			Date:  inc.StartDate,
			Hours: hours,
		})
		return rows
	}

	// Multi-day: create one row per day
	hoursPerDay := (inc.Quantity * mapping.HoursMultiplier) / days
	currentDate := inc.StartDate

	for i := 0; i < int(days); i++ {
		rows = append(rows, ExpandedRow{
			Date:  currentDate,
			Hours: hoursPerDay,
		})
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return rows
}

// GetExportPreview returns a preview of the export without generating files
// Used for UI preview before download
func (s *ExcelExportService) GetExportPreview(payrollPeriodID uuid.UUID) (map[string]interface{}, error) {
	// Get all approved incidences
	var incidences []models.Incidence
	err := s.db.
		Preload("Employee").
		Preload("IncidenceType").
		Where("payroll_period_id = ?", payrollPeriodID).
		Where("status = ?", "approved").
		Where("excluded_from_payroll = ?", false).
		Find(&incidences).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load incidences: %w", err)
	}

	// Count by template type
	vacacionesCount := 0
	faltasExtrasCount := 0
	lateApprovalCount := 0

	for _, inc := range incidences {
		var mapping models.IncidenceTipoMapping
		err := s.db.Where("incidence_type_id = ?", inc.IncidenceTypeID).First(&mapping).Error
		if err != nil {
			continue
		}

		if mapping.TemplateType == "vacaciones" {
			vacacionesCount++
		} else if mapping.TemplateType == "faltas_extras" {
			// Count expanded rows for multi-day absences
			expandedRows := s.expandMultiDayAbsence(inc, mapping)
			faltasExtrasCount += len(expandedRows)
		}

		if inc.LateApprovalFlag {
			lateApprovalCount++
		}
	}

	preview := map[string]interface{}{
		"vacaciones_count":     vacacionesCount,
		"faltas_extras_count":  faltasExtrasCount,
		"late_approval_count":  lateApprovalCount,
		"total_incidences":     len(incidences),
		"payroll_period_id":    payrollPeriodID,
	}

	return preview, nil
}
