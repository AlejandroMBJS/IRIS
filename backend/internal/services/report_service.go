/*
Package services - Report Generation Service

==============================================================================
FILE: internal/services/report_service.go
==============================================================================

DESCRIPTION:
    Generates payroll reports in multiple formats (PDF, CSV, Excel, JSON) including
    financial exports, collar type summaries, and employee lists for accounting
    and compliance purposes.

USER PERSPECTIVE:
    - Generate financial export reports grouped by collar type
    - Export payroll summaries for accounting teams
    - Create employee lists in various formats
    - PDF reports with company branding and Mexican Spanish labels

DEVELOPER GUIDELINES:
    OK to modify: Report formats, add new report types
    CAUTION: Financial export calculations must match payroll exactly
    DO NOT modify: Collar type grouping logic without updating payroll
    Note: Reports use gofpdf for PDF generation with Mexican formatting

SYNTAX EXPLANATION:
    - CollarType groups: white_collar (admin), blue_collar (union), gray_collar (non-union)
    - Financial export includes employer contributions (IMSS, INFONAVIT, SAR)
    - CSV uses Spanish headers for Mexican accounting standards
    - PDF reports display amounts in MXN currency format

==============================================================================
*/
package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"

	"backend/internal/dtos"
	"backend/internal/models"
	"backend/internal/repositories"
)

// ReportService handles business logic for generating reports.
type ReportService struct {
	payrollRepo  *repositories.PayrollRepository
	employeeRepo *repositories.EmployeeRepository
	periodRepo   *repositories.PayrollPeriodRepository
}

// NewReportService creates a new ReportService.
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{
		payrollRepo:  repositories.NewPayrollRepository(db),
		employeeRepo: repositories.NewEmployeeRepository(db),
		periodRepo:   repositories.NewPayrollPeriodRepository(db),
	}
}

// GeneratePayrollReport is a placeholder for generating a payroll report.
func (s *ReportService) GeneratePayrollReport(periodID string) ([]byte, error) {
	// Dummy report generation
	return []byte(fmt.Sprintf("Payroll Report for period: %s", periodID)), nil
}

// GenerateComplianceReport is a placeholder for generating a compliance report.
func (s *ReportService) GenerateComplianceReport() ([]byte, error) {
	// Dummy report generation
	return []byte("Compliance Report Data"), nil
}

// GetReportHistory is a placeholder for retrieving report history.
func (s *ReportService) GetReportHistory() ([]byte, error) {
	return []byte("Report History Data"), nil
}

// GenerateReport generates various reports based on request.
func (s *ReportService) GenerateReport(req dtos.PayrollReportRequest) ([]byte, error) {
	switch req.ReportType {
	case "payroll_summary":
		calculations, err := s.payrollRepo.FindByPeriod(req.PayrollPeriodID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve payroll calculations for report: %w", err)
		}
		// In a real scenario, convert calculations to a summary DTO or structure
		// For now, just a placeholder of the data.

		switch req.Format {
		case "json":
			// Marshal calculations to JSON for now
			// In reality, this would be a dtos.PayrollSummary or similar
			jsonReport, err := json.MarshalIndent(calculations, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payroll summary to JSON: %w", err)
			}
			return jsonReport, nil
		case "pdf":
			// Placeholder for PDF generation
			return []byte(fmt.Sprintf("PDF Payroll Summary Report for period %s", req.PayrollPeriodID)), nil
		case "csv":
			// Placeholder for CSV generation
			return []byte(fmt.Sprintf("CSV Payroll Summary Report for period %s", req.PayrollPeriodID)), nil
		case "excel":
			// Placeholder for Excel generation
			return []byte(fmt.Sprintf("Excel Payroll Summary Report for period %s", req.PayrollPeriodID)), nil
		default:
			return nil, errors.New("unsupported report format")
		}

	case "employee_list":
		// Fetch employee list
		employees, _, err := s.employeeRepo.List(1, 10000, nil) // All employees
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve employee list for report: %w", err)
		}
		switch req.Format {
		case "json":
			jsonReport, err := json.MarshalIndent(employees, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal employee list to JSON: %w", err)
			}
			return jsonReport, nil
		case "pdf":
			return []byte(fmt.Sprintf("PDF Employee List Report")), nil
		case "csv":
			return []byte(fmt.Sprintf("CSV Employee List Report")), nil
		case "excel":
			return []byte(fmt.Sprintf("Excel Employee List Report")), nil
		default:
			return nil, errors.New("unsupported report format")
		}

	case "financial_export":
		return s.GenerateFinancialExport(req)

	case "collar_type_summary":
		return s.GenerateCollarTypeSummary(req)

	default:
		return nil, errors.New("unsupported report type")
	}
}

// GenerateFinancialExport generates a financial export report grouped by collar type
func (s *ReportService) GenerateFinancialExport(req dtos.PayrollReportRequest) ([]byte, error) {
	period, err := s.periodRepo.FindByID(req.PayrollPeriodID)
	if err != nil {
		return nil, fmt.Errorf("failed to find period: %w", err)
	}

	calculations, err := s.payrollRepo.FindByPeriod(req.PayrollPeriodID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payroll calculations: %w", err)
	}

	// Group by collar type
	collarGroups := s.groupByCollarType(calculations)

	response := s.buildFinancialExportResponse(period, collarGroups)

	switch req.Format {
	case "json":
		return json.MarshalIndent(response, "", "  ")
	case "csv":
		return s.generateFinancialCSV(period, calculations, collarGroups)
	case "pdf":
		return s.generateFinancialPDF(period, collarGroups, response)
	default:
		return json.MarshalIndent(response, "", "  ")
	}
}

// GenerateCollarTypeSummary generates a summary report by collar type
func (s *ReportService) GenerateCollarTypeSummary(req dtos.PayrollReportRequest) ([]byte, error) {
	calculations, err := s.payrollRepo.FindByPeriod(req.PayrollPeriodID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payroll calculations: %w", err)
	}

	// Filter by collar type if specified
	if req.CollarType != "" && req.CollarType != "all" {
		filtered := make([]models.PayrollCalculation, 0)
		for _, calc := range calculations {
			if calc.Employee.CollarType == req.CollarType {
				filtered = append(filtered, calc)
			}
		}
		calculations = filtered
	}

	collarGroups := s.groupByCollarType(calculations)

	switch req.Format {
	case "json":
		return json.MarshalIndent(collarGroups, "", "  ")
	default:
		return json.MarshalIndent(collarGroups, "", "  ")
	}
}

// groupByCollarType groups payroll calculations by collar type
func (s *ReportService) groupByCollarType(calculations []models.PayrollCalculation) map[string][]models.PayrollCalculation {
	groups := make(map[string][]models.PayrollCalculation)

	for _, calc := range calculations {
		collarType := calc.Employee.CollarType
		if collarType == "" {
			collarType = "white_collar"
		}
		groups[collarType] = append(groups[collarType], calc)
	}

	return groups
}

// buildFinancialExportResponse builds the financial export response
func (s *ReportService) buildFinancialExportResponse(period *models.PayrollPeriod, collarGroups map[string][]models.PayrollCalculation) *dtos.FinancialExportResponse {
	response := &dtos.FinancialExportResponse{
		PeriodCode:      period.PeriodCode,
		PeriodStartDate: period.StartDate.Format("2006-01-02"),
		PeriodEndDate:   period.EndDate.Format("2006-01-02"),
		PaymentDate:     period.PaymentDate.Format("2006-01-02"),
		GeneratedAt:     time.Now().Format("2006-01-02 15:04:05"),
		CollarSummaries: make([]dtos.CollarTypeSummary, 0),
	}

	collarLabels := map[string]string{
		"white_collar": "Administrativo (White Collar)",
		"blue_collar":  "Obrero Sindicalizado (Blue Collar)",
		"gray_collar":  "Obrero No Sindicalizado (Gray Collar)",
	}

	// Process each collar type in order
	for _, collarType := range []string{"white_collar", "blue_collar", "gray_collar"} {
		calcs, exists := collarGroups[collarType]
		if !exists || len(calcs) == 0 {
			continue
		}

		summary := dtos.CollarTypeSummary{
			CollarType:      collarType,
			CollarTypeLabel: collarLabels[collarType],
			EmployeeCount:   len(calcs),
		}

		for _, calc := range calcs {
			summary.TotalGross += calc.TotalGrossIncome
			summary.TotalDeductions += calc.TotalStatutoryDeductions + calc.TotalOtherDeductions
			summary.TotalNet += calc.TotalNetPay
			summary.TotalISR += calc.ISRWithholding
			summary.TotalIMSS += calc.IMSSEmployee
			summary.TotalInfonavit += calc.InfonavitEmployee
			summary.EmployerIMSS += calc.IMSSEmployer
			summary.EmployerInfonavit += calc.InfonavitEmployer
			// SAR comes from employer contribution if loaded
			if calc.EmployerContribution != nil {
				summary.EmployerRetirement += calc.EmployerContribution.RetirementSAR
			}
		}

		response.CollarSummaries = append(response.CollarSummaries, summary)
		response.TotalEmployees += summary.EmployeeCount
		response.GrandTotalGross += summary.TotalGross
		response.GrandTotalNet += summary.TotalNet
		response.GrandTotalEmployer += summary.EmployerIMSS + summary.EmployerInfonavit + summary.EmployerRetirement
	}

	return response
}

// generateFinancialCSV generates a CSV export for financial team
func (s *ReportService) generateFinancialCSV(period *models.PayrollPeriod, calculations []models.PayrollCalculation, collarGroups map[string][]models.PayrollCalculation) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	writer.Write([]string{
		"Tipo de Collar", "Num. Empleado", "Nombre", "RFC", "CURP",
		"Salario Bruto", "ISR", "IMSS Empleado", "INFONAVIT Empleado",
		"Total Deducciones", "Neto a Pagar",
		"IMSS Patron", "INFONAVIT Patron", "SAR/AFORE Patron",
	})

	collarLabels := map[string]string{
		"white_collar": "Administrativo",
		"blue_collar":  "Obrero Sindicalizado",
		"gray_collar":  "Obrero No Sindicalizado",
	}

	// Write data for each collar type
	for _, collarType := range []string{"white_collar", "blue_collar", "gray_collar"} {
		calcs, exists := collarGroups[collarType]
		if !exists {
			continue
		}

		for _, calc := range calcs {
			// SAR comes from employer contribution if loaded
			sarEmployer := 0.0
			if calc.EmployerContribution != nil {
				sarEmployer = calc.EmployerContribution.RetirementSAR
			}
			writer.Write([]string{
				collarLabels[collarType],
				calc.Employee.EmployeeNumber,
				fmt.Sprintf("%s %s", calc.Employee.FirstName, calc.Employee.LastName),
				calc.Employee.RFC,
				calc.Employee.CURP,
				fmt.Sprintf("%.2f", calc.TotalGrossIncome),
				fmt.Sprintf("%.2f", calc.ISRWithholding),
				fmt.Sprintf("%.2f", calc.IMSSEmployee),
				fmt.Sprintf("%.2f", calc.InfonavitEmployee),
				fmt.Sprintf("%.2f", calc.TotalStatutoryDeductions+calc.TotalOtherDeductions),
				fmt.Sprintf("%.2f", calc.TotalNetPay),
				fmt.Sprintf("%.2f", calc.IMSSEmployer),
				fmt.Sprintf("%.2f", calc.InfonavitEmployer),
				fmt.Sprintf("%.2f", sarEmployer),
			})
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// generateFinancialPDF generates a PDF export for financial team
func (s *ReportService) generateFinancialPDF(period *models.PayrollPeriod, collarGroups map[string][]models.PayrollCalculation, response *dtos.FinancialExportResponse) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for more columns
	pdf.AddPage()

	// Header
	headerR, headerG, headerB := 30, 58, 138
	pdf.SetFillColor(headerR, headerG, headerB)
	pdf.Rect(0, 0, 297, 30, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetXY(10, 8)
	pdf.Cell(200, 10, "REPORTE FINANCIERO DE NOMINA")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(10, 18)
	pdf.Cell(200, 6, fmt.Sprintf("Periodo: %s | Fecha de Pago: %s", period.PeriodCode, period.PaymentDate.Format("02/01/2006")))

	pdf.SetTextColor(0, 0, 0)

	// Summary by collar type
	y := 40.0
	pdf.SetXY(10, y)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(277, 8, "RESUMEN POR TIPO DE COLLAR", "1", 1, "C", true, 0, "")

	// Summary table header
	y = pdf.GetY()
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)
	pdf.SetXY(10, y)
	pdf.CellFormat(60, 7, "Tipo de Collar", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Empleados", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Total Bruto", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Deducciones", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Neto", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "IMSS Patron", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "INFONAVIT", "1", 0, "C", true, 0, "")
	pdf.CellFormat(27, 7, "SAR/AFORE", "1", 1, "C", true, 0, "")

	// Summary data
	pdf.SetFont("Arial", "", 9)
	for _, summary := range response.CollarSummaries {
		pdf.SetXY(10, pdf.GetY())
		pdf.CellFormat(60, 6, summary.CollarTypeLabel, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", summary.EmployeeCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("$%.2f", summary.TotalGross), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("$%.2f", summary.TotalDeductions), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("$%.2f", summary.TotalNet), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("$%.2f", summary.EmployerIMSS), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("$%.2f", summary.EmployerInfonavit), "1", 0, "R", false, 0, "")
		pdf.CellFormat(27, 6, fmt.Sprintf("$%.2f", summary.EmployerRetirement), "1", 1, "R", false, 0, "")
	}

	// Grand totals
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(headerR, headerG, headerB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(10, pdf.GetY())
	pdf.CellFormat(60, 7, "TOTALES", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, fmt.Sprintf("%d", response.TotalEmployees), "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("$%.2f", response.GrandTotalGross), "1", 0, "R", true, 0, "")
	totalDeductions := response.GrandTotalGross - response.GrandTotalNet
	pdf.CellFormat(35, 7, fmt.Sprintf("$%.2f", totalDeductions), "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("$%.2f", response.GrandTotalNet), "1", 0, "R", true, 0, "")
	pdf.CellFormat(87, 7, fmt.Sprintf("$%.2f", response.GrandTotalEmployer), "1", 1, "R", true, 0, "")

	pdf.SetTextColor(0, 0, 0)

	// Footer
	pdf.SetXY(10, 190)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(277, 5, fmt.Sprintf("Generado: %s | Este documento es para uso interno del equipo financiero", time.Now().Format("02/01/2006 15:04")))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// GetCollarTypeLabel returns the Spanish label for a collar type
func GetCollarTypeLabel(collarType string) string {
	labels := map[string]string{
		"white_collar": "Administrativo (White Collar)",
		"blue_collar":  "Obrero Sindicalizado (Blue Collar)",
		"gray_collar":  "Obrero No Sindicalizado (Gray Collar)",
	}
	if label, ok := labels[collarType]; ok {
		return label
	}
	return collarType
}

// Suppress unused variable warnings
var (
	_ = uuid.UUID{}
	_ = models.Employee{}
)
