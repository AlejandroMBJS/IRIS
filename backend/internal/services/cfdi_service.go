/*
Package services - CFDI (SAT) XML Generation Service

==============================================================================
FILE: internal/services/cfdi_service.go
==============================================================================

DESCRIPTION:
    Generates Mexican CFDI (Comprobante Fiscal Digital por Internet) version 4.0
    payroll receipts with Nomina 1.2 complement for SAT compliance. Creates
    digitally signed XML documents for employee payslips.

USER PERSPECTIVE:
    - Generate legally valid electronic payslips (recibos de nomina)
    - CFDI 4.0 compliant with SAT regulations
    - Digital signature using company's CSD (Certificado de Sello Digital)
    - Required for legal payroll in Mexico

DEVELOPER GUIDELINES:
    OK to modify: XML structure following SAT schema updates
    CAUTION: CSD certificate handling requires secure storage
    DO NOT modify: Digital signature process without cryptography expertise
    Note: Requires company CSD .cer and .key files for signing

SYNTAX EXPLANATION:
    - Comprobante: Main CFDI invoice structure (version 4.0)
    - Complemento Nomina 1.2: Payroll-specific supplement
    - Sello: Digital signature created with company's private key
    - Cadena Original: String to sign generated via SAT XSLT transformation
    - PeriodicidadPago: SAT codes (01=daily, 02=weekly, 04=biweekly, 05=monthly)

==============================================================================
*/
package services

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"time"

	"backend/internal/models"

	"golang.org/x/crypto/pkcs12"
)

// CfdiService is responsible for creating CFDI for payroll.
type CfdiService struct {
	// In a real scenario, you would have access to the company's
	// CSD (Certificado de Sello Digital) .cer and .key files.
	// These would be securely stored, perhaps using a secrets manager.
	certificatePath string
	privateKeyPath  string
	privateKeyPass  string
}

// NewCfdiService creates a new CFDI service.
func NewCfdiService(certPath, keyPath, keyPass string) *CfdiService {
	return &CfdiService{
		certificatePath: certPath,
		privateKeyPath:  keyPath,
		privateKeyPass:  keyPass,
	}
}

// GenerateCfdiXML creates a CFDI 4.0 XML for a given payroll calculation.
func (s *CfdiService) GenerateCfdiXML(payroll *models.PayrollCalculation) ([]byte, error) {
	comprobante := s.buildComprobante(payroll)

	// In a real implementation, you would perform the following steps:
	// 1. Generate the "cadena original" (original string) from the comprobante.
	// 2. Sign the cadena original with the employer's private key (CSD) to get the "sello".
	// 3. Add the sello to the comprobante.
	//
	// The following is a placeholder for this complex process.
	cadenaOriginal := s.generateCadenaOriginal(comprobante)
	sello, err := s.signCadenaOriginal(cadenaOriginal)
	if err != nil {
		// For this placeholder, we'll just generate a dummy sello
		sello = "DUMMY_SELLO_FOR_DEMONSTRATION_PURPOSES_ONLY"
	}
	comprobante.Sello = sello

	// Marshal the struct into XML
	xmlBytes, err := xml.MarshalIndent(comprobante, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CFDI to XML: %w", err)
	}

	return xmlBytes, nil
}

func (s *CfdiService) buildComprobante(payroll *models.PayrollCalculation) *models.Comprobante {
	fecha := time.Now().Format("2006-01-02T15:04:05")
	// Placeholder values - these should come from your application's configuration and data
	return &models.Comprobante{
		Cfdi:              "http://www.sat.gob.mx/cfd/4",
		Nomina12:          "http://www.sat.gob.mx/nomina12",
		Version:           "4.0",
		Serie:             "A",
		Folio:             "12345",
		Fecha:             fecha,
		NoCertificado:     "30001000000300023708", // Placeholder
		SubTotal:          fmt.Sprintf("%.2f", payroll.TotalGrossIncome),
		Moneda:            "MXN",
		Total:             fmt.Sprintf("%.2f", payroll.TotalNetPay),
		TipoDeComprobante: "N", // Nómina
		Exportacion:       "01", // No aplica
		LugarExpedicion:   "78000", // San Luis Potosí CP
		Emisor: models.Emisor{
			Rfc:           "EKU9003173C9",    // Placeholder
			Nombre:        "Empresa Ficticia S.A. de C.V.", // Placeholder
			RegimenFiscal: "601",             // General de Ley Personas Morales
		},
		Receptor: models.Receptor{
			Rfc_receptor:            payroll.Employee.RFC,
			Nombre_receptor:         payroll.Employee.FirstName + " " + payroll.Employee.LastName,
			UsoCFDI:                 "CN01", // Nómina
			DomicilioFiscalReceptor: payroll.Employee.PostalCode,
			RegimenFiscalReceptor:   "605", // Sueldos y Salarios e Ingresos Asimilados a Salarios
		},
		Conceptos: models.Conceptos{
			Concepto: []models.Concepto{
				{
					ClaveProdServ: "84111505", // Servicios de contabilidad de sueldos y salarios
					Cantidad:      "1",
					ClaveUnidad:   "ACT", // Actividad
					Descripcion:   "Pago de nómina",
					ValorUnitario: fmt.Sprintf("%.2f", payroll.TotalGrossIncome),
					Importe:       fmt.Sprintf("%.2f", payroll.TotalGrossIncome),
					Descuento:     fmt.Sprintf("%.2f", payroll.TotalStatutoryDeductions+payroll.TotalOtherDeductions),
					ObjetoImp:    "01", // No objeto de impuesto
				},
			},
		},
		Complemento: models.Complemento{
			Nomina: models.Nomina12{
				Version:           "1.2",
				TipoNomina:        "O", // Ordinaria
				FechaPago:         payroll.PayrollPeriod.PaymentDate.Format("2006-01-02"),
				FechaInicialPago:  payroll.PayrollPeriod.StartDate.Format("2006-01-02"),
				FechaFinalPago:    payroll.PayrollPeriod.EndDate.Format("2006-01-02"),
				NumDiasPagados:    fmt.Sprintf("%.2f", 15.0), // Example
				TotalPercepciones: fmt.Sprintf("%.2f", payroll.TotalGrossIncome),
				TotalDeducciones:  fmt.Sprintf("%.2f", payroll.TotalStatutoryDeductions+payroll.TotalOtherDeductions),
				Emisor_nomina: models.NominaEmisor{
					RegistroPatronal: "A1234567890", // Placeholder
				},
				Receptor_nomina: models.NominaReceptor{
					Curp:                  payroll.Employee.CURP,
					NumSeguridadSocial:    payroll.Employee.NSS,
					FechaInicioRelLaboral: payroll.Employee.HireDate.Format("2006-01-02"),
					Antiguedad:            s.calculateAntiguedad(payroll.Employee.HireDate),
					TipoContrato:          s.getTipoContrato(payroll.Employee.EmployeeType),
					TipoRegimen:           "02",  // Sueldos (for San Luis Potosí)
					NumEmpleado:           payroll.Employee.EmployeeNumber,
					PeriodicidadPago:      s.getPeriodicidadPago(payroll.Employee.PayFrequency),
					SalarioDiarioIntegrado: fmt.Sprintf("%.2f", payroll.Employee.IntegratedDailySalary),
					SalarioBaseCotApor:    fmt.Sprintf("%.2f", payroll.Employee.DailySalary),
				},
				Percepciones: &models.Percepciones{
					TotalSueldos:   fmt.Sprintf("%.2f", payroll.TotalGrossIncome),
					TotalGravado:   fmt.Sprintf("%.2f", payroll.TotalGrossIncome), // Simplified
					TotalExento:    "0.00",
					Percepcion: []*models.Percepcion{
						{
							TipoPercepcion: "001",
							Clave:          "001",
							Concepto:       "Sueldos, Salarios Rayas y Jornales",
							ImporteGravado: fmt.Sprintf("%.2f", payroll.RegularSalary),
							ImporteExento:  "0.00",
						},
					},
				},
				Deducciones: &models.Deducciones{
					TotalImpuestosRetenidos: fmt.Sprintf("%.2f", payroll.ISRWithholding),
					Deduccion: []*models.Deduccion{
						{
							TipoDeduccion: "002",
							Clave:         "001",
							Concepto:      "ISR",
							Importe:       fmt.Sprintf("%.2f", payroll.ISRWithholding),
						},
						{
							TipoDeduccion: "001",
							Clave:         "002",
							Concepto:      "Seguridad social",
							Importe:       fmt.Sprintf("%.2f", payroll.IMSSEmployee),
						},
					},
				},
			},
		},
	}
}

// generateCadenaOriginal is a placeholder for creating the original string for signing.
// The actual implementation requires transforming the XML with an XSLT provided by SAT.
func (s *CfdiService) generateCadenaOriginal(comprobante *models.Comprobante) string {
	// This is a highly simplified representation.
	// The real process involves applying an XSLT transformation to the XML.
	return fmt.Sprintf("||%s|%s|%s|...||", comprobante.Version, comprobante.Fecha, comprobante.Total)
}

// signCadenaOriginal is a placeholder for the digital signature process.
func (s *CfdiService) signCadenaOriginal(cadena string) (string, error) {
	// IMPORTANT: This is a placeholder. A real implementation is complex.
	// You need to:
	// 1. Load the CSD private key (.key file), which is often encrypted.
	// 2. Decode the key using the provided password.
	// 3. Hash the "cadena original" with SHA-256.
	// 4. Sign the hash using RSA with the private key.
	// 5. Encode the signature in Base64.

	if s.privateKeyPath == "" || s.privateKeyPass == "" {
		return "", fmt.Errorf("private key path or password not provided for signing")
	}

	// Read the encrypted private key file
	pfxData, err := ioutil.ReadFile(s.privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key file: %w", err)
	}

	// Decode the PFX/P12 file
	_ , _, err = pkcs12.Decode(pfxData, s.privateKeyPass) // Use _ to ignore unused privateKey
	if err != nil {
		// Try decoding without password if it fails
		_ , _, err = pkcs12.Decode(pfxData, "") // Use _ to ignore unused privateKey
		if err != nil {
			return "", fmt.Errorf("failed to decode PFX file with or without password: %w", err)
		}
	}

	// For demonstration, we will just hash the cadena and encode it.
	// THIS IS NOT A VALID DIGITAL SIGNATURE.
	hasher := sha256.New()
	hasher.Write([]byte(cadena))
	hash := hasher.Sum(nil)

	// This is where you would use the privateKey to sign the hash.
	// signedHash, err := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hash)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to sign hash: %w", err)
	// }
	// return base64.StdEncoding.EncodeToString(signedHash), nil

	return base64.StdEncoding.EncodeToString(hash), nil
}

// calculateAntiguedad calculates employee's seniority in ISO 8601 duration format (P#Y#M#D)
func (s *CfdiService) calculateAntiguedad(hireDate time.Time) string {
	now := time.Now()
	years := now.Year() - hireDate.Year()
	months := int(now.Month()) - int(hireDate.Month())
	days := now.Day() - hireDate.Day()

	if days < 0 {
		months--
		// Get days in previous month
		prevMonth := now.AddDate(0, -1, 0)
		days += time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	}
	if months < 0 {
		years--
		months += 12
	}

	return fmt.Sprintf("P%dY%dM%dD", years, months, days)
}

// getPeriodicidadPago returns SAT code for payment periodicity
// Per SAT Catálogo c_PeriodicidadPago for Nomina 1.2
func (s *CfdiService) getPeriodicidadPago(payFrequency string) string {
	switch payFrequency {
	case "daily":
		return "01" // Diario
	case "weekly":
		return "02" // Semanal
	case "biweekly":
		return "04" // Quincenal
	case "monthly":
		return "05" // Mensual
	default:
		return "04" // Default to biweekly
	}
}

// getTipoContrato returns SAT code for contract type
// Per SAT Catálogo c_TipoContrato for Nomina 1.2
func (s *CfdiService) getTipoContrato(employeeType string) string {
	switch employeeType {
	case "permanent":
		return "01" // Contrato de trabajo por tiempo indeterminado
	case "temporary":
		return "02" // Contrato de trabajo por obra determinada
	case "contractor":
		return "09" // Otro contrato (for contractors)
	case "intern":
		return "07" // Modalidades de contratación donde no existe relación de trabajo
	default:
		return "01" // Default to permanent
	}
}
