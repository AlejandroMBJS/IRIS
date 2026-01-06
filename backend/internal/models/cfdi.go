/*
Package models - IRIS Payroll System Data Models

==============================================================================
FILE: internal/models/cfdi.go
==============================================================================

DESCRIPTION:
    Defines XML structures for CFDI (Comprobante Fiscal Digital por Internet)
    version 4.0 with Nomina 1.2 complement. These structs map to the official
    SAT XML schema for electronic payroll receipts in Mexico.

USER PERSPECTIVE:
    - CFDI is the official electronic invoice required by Mexican law
    - Each payroll payment generates a CFDI Nomina document
    - These are the XML receipts employees receive for tax purposes
    - Required for IMSS, SAT reporting, and employee tax declarations

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new optional elements from SAT schema
    ⚠️  CAUTION: Any changes must match official SAT XSD schemas
    ❌  DO NOT modify: Required fields, namespaces, attribute names
    📝  Test any changes against SAT validation services

SYNTAX EXPLANATION:
    - xml.Name: Defines the XML element name and namespace
    - `xml:"attr"`: Maps struct field to XML attribute
    - `xml:",omitempty"`: Omit element if empty (for optional fields)
    - `xml:"cfdi:..."`: Namespace prefix for CFDI 4.0 elements
    - `xml:"nomina12:..."`: Namespace prefix for Nomina 1.2 elements

CFDI 4.0 + NOMINA 1.2 STRUCTURE:
    Comprobante (root)
    ├── Emisor (company info)
    ├── Receptor (employee info)
    ├── Conceptos (payroll line item)
    └── Complemento
        └── Nomina (payroll-specific data)
            ├── Emisor (employer registration)
            ├── Receptor (employee work details)
            ├── Percepciones (income items)
            ├── Deducciones (deduction items)
            └── OtrosPagos (other payments like subsidies)

NOMINA CONCEPTS:
    - Percepciones: Income items (salary, overtime, bonuses)
    - Deducciones: Deductions (ISR, IMSS, loans)
    - OtrosPagos: Other payments (employment subsidy)
    - TipoNomina: "O" ordinary, "E" extraordinary
    - PeriodicidadPago: "02" weekly, "04" biweekly, "05" monthly

IMPORTANT:
    This file is critical for Mexican tax compliance. Any modifications
    must be validated against the official SAT schemas before deployment.

==============================================================================
*/
package models

import "encoding/xml"

// Comprobante is the root element for CFDI 4.0.
type Comprobante struct {
	XMLName           xml.Name `xml:"cfdi:Comprobante"`
	Cfdi              string   `xml:"xmlns:cfdi,attr"`
	Nomina12          string   `xml:"xmlns:nomina12,attr"`
	Version           string   `xml:"Version,attr"`
	Serie             string   `xml:"Serie,attr,omitempty"`
	Folio             string   `xml:"Folio,attr,omitempty"`
	Fecha             string   `xml:"Fecha,attr"`
	Sello             string   `xml:"Sello,attr,omitempty"`
	FormaPago         string   `xml:"FormaPago,attr,omitempty"`
	NoCertificado     string   `xml:"NoCertificado,attr,omitempty"`
	Certificado       string   `xml:"Certificado,attr,omitempty"`
	CondicionesDePago string   `xml:"CondicionesDePago,attr,omitempty"`
	SubTotal          string   `xml:"SubTotal,attr"`
	Descuento         string   `xml:"Descuento,attr,omitempty"`
	Moneda            string   `xml:"Moneda,attr"`
	TipoCambio        string   `xml:"TipoCambio,attr,omitempty"`
	Total             string   `xml:"Total,attr"`
	TipoDeComprobante string   `xml:"TipoDeComprobante,attr"`
	Exportacion       string   `xml:"Exportacion,attr,omitempty"`
	MetodoPago        string   `xml:"MetodoPago,attr,omitempty"`
	LugarExpedicion   string   `xml:"LugarExpedicion,attr"`

	InformacionGlobal *InformacionGlobal `xml:"cfdi:InformacionGlobal,omitempty"`
	CfdiRelacionados  *CfdiRelacionados  `xml:"cfdi:CfdiRelacionados,omitempty"`
	Emisor            Emisor           `xml:"cfdi:Emisor"`
	Receptor          Receptor         `xml:"cfdi:Receptor"`
	Conceptos         Conceptos        `xml:"cfdi:Conceptos"`
	Impuestos         *Impuestos       `xml:"cfdi:Impuestos,omitempty"`
	Complemento       Complemento      `xml:"cfdi:Complemento"`
	Addenda           *Addenda         `xml:"cfdi:Addenda,omitempty"`
}

// InformacionGlobal for CFDI 4.0
type InformacionGlobal struct {
	Meses string `xml:"Meses,attr"`
	Ano   string `xml:"Ano,attr"`
	Periodicidad string `xml:"Periodicidad,attr"`
}

// CfdiRelacionados for CFDI 4.0
type CfdiRelacionados struct {
	TipoRelacion string         `xml:"TipoRelacion,attr"`
	CfdiRelacion []CfdiRelacion `xml:"cfdi:CfdiRelacion"`
}

// CfdiRelacion for CFDI 4.0
type CfdiRelacion struct {
	UUID string `xml:"UUID,attr"`
}

// Emisor for CFDI 4.0
type Emisor struct {
	Rfc           string `xml:"Rfc,attr"`
	Nombre        string `xml:"Nombre,attr"`
	RegimenFiscal string `xml:"RegimenFiscal,attr"`
	FacAtrAdquirente string `xml:"FacAtrAdquirente,attr,omitempty"`
}

// Receptor for CFDI 4.0
type Receptor struct {
	Rfc_receptor            string `xml:"Rfc,attr"`
	Nombre_receptor         string `xml:"Nombre,attr"`
	DomicilioFiscalReceptor string `xml:"DomicilioFiscalReceptor,attr"`
	RegimenFiscalReceptor   string `xml:"RegimenFiscalReceptor,attr"`
	UsoCFDI                 string `xml:"UsoCFDI,attr"`
}

// Conceptos for CFDI 4.0
type Conceptos struct {
	Concepto []Concepto `xml:"cfdi:Concepto"`
}

// Concepto for CFDI 4.0
type Concepto struct {
	ClaveProdServ string `xml:"ClaveProdServ,attr"`
	NoIdentificacion string `xml:"NoIdentificacion,attr,omitempty"`
	Cantidad        string `xml:"Cantidad,attr"`
	Unidad          string `xml:"Unidad,attr,omitempty"`
	ClaveUnidad     string `xml:"ClaveUnidad,attr"`
	Descripcion     string `xml:"Descripcion,attr"`
	ValorUnitario   string `xml:"ValorUnitario,attr"`
	Importe         string `xml:"Importe,attr"`
	Descuento       string `xml:"Descuento,attr,omitempty"`
	ObjetoImp       string `xml:"ObjetoImp,attr"`
	Impuestos       *ImpuestosConcepto `xml:"cfdi:Impuestos,omitempty"`
	InformacionAduanera *InformacionAduanera `xml:"cfdi:InformacionAduanera,omitempty"`
	CuentaPredial   *CuentaPredial     `xml:"cfdi:CuentaPredial,omitempty"`
	ComplementoConcepto *ComplementoConcepto `xml:"cfdi:ComplementoConcepto,omitempty"`
}

// ImpuestosConcepto for CFDI 4.0
type ImpuestosConcepto struct {
	Traslados *TrasladosConcepto `xml:"cfdi:Traslados,omitempty"`
	Retenciones *RetencionesConcepto `xml:"cfdi:Retenciones,omitempty"`
}

// TrasladosConcepto for CFDI 4.0
type TrasladosConcepto struct {
	Traslado []TrasladoConcepto `xml:"cfdi:Traslado"`
}

// TrasladoConcepto for CFDI 4.0
type TrasladoConcepto struct {
	Base string `xml:"Base,attr"`
	Impuesto string `xml:"Impuesto,attr"`
	TipoFactor string `xml:"TipoFactor,attr"`
	TasaOCuota string `xml:"TasaOCuota,attr"`
	Importe string `xml:"Importe,attr"`
}

// RetencionesConcepto for CFDI 4.0
type RetencionesConcepto struct {
	Retencion []RetencionConcepto `xml:"cfdi:Retencion"`
}

// RetencionConcepto for CFDI 4.0
type RetencionConcepto struct {
	Base string `xml:"Base,attr"`
	Impuesto string `xml:"Impuesto,attr"`
	TipoFactor string `xml:"TipoFactor,attr"`
	TasaOCuota string `xml:"TasaOCuota,attr"`
	Importe string `xml:"Importe,attr"`
}

// InformacionAduanera for CFDI 4.0
type InformacionAduanera struct {
	NumeroPedimento string `xml:"NumeroPedimento,attr"`
}

// CuentaPredial for CFDI 4.0
type CuentaPredial struct {
	Numero string `xml:"Numero,attr"`
}

// ComplementoConcepto for CFDI 4.0
type ComplementoConcepto struct {
	// Any specific complement for concept
}

// Impuestos for CFDI 4.0 (Global)
type Impuestos struct {
	TotalRetenciones string    `xml:"TotalRetenciones,attr,omitempty"`
	TotalTraslados   string    `xml:"TotalTraslados,attr,omitempty"`
	Retenciones      Retenciones `xml:"cfdi:Retenciones,omitempty"`
	Traslados        Traslados `xml:"cfdi:Traslados,omitempty"`
}

// Retenciones for CFDI 4.0 (Global)
type Retenciones struct {
	Retencion []Retencion `xml:"cfdi:Retencion"`
}

// Retencion for CFDI 4.0 (Global)
type Retencion struct {
	Impuesto string `xml:"Impuesto,attr"`
	Importe  string `xml:"Importe,attr"`
}

// Traslados for CFDI 4.0 (Global)
type Traslados struct {
	Traslado []Traslado `xml:"cfdi:Traslado"`
}

// Traslado for CFDI 4.0 (Global)
type Traslado struct {
	Base       string `xml:"Base,attr"`
	Impuesto   string `xml:"Impuesto,attr"`
	TipoFactor string `xml:"TipoFactor,attr"`
	TasaOCuota string `xml:"TasaOCuota,attr"`
	Importe    string `xml:"Importe,attr"`
}

// Complemento for CFDI 4.0
type Complemento struct {
	Nomina Nomina12 `xml:"nomina12:Nomina"`
}

// Nomina12 is the Nomina Complement for CFDI 4.0
type Nomina12 struct {
	XMLName              xml.Name    `xml:"nomina12:Nomina"`
	Version              string      `xml:"Version,attr"`
	TipoNomina           string      `xml:"TipoNomina,attr"`
	FechaPago            string      `xml:"FechaPago,attr"`
	FechaInicialPago     string      `xml:"FechaInicialPago,attr"`
	FechaFinalPago       string      `xml:"FechaFinalPago,attr"`
	NumDiasPagados       string      `xml:"NumDiasPagados,attr"`
	TotalPercepciones    string      `xml:"TotalPercepciones,attr,omitempty"`
	TotalDeducciones     string      `xml:"TotalDeducciones,attr,omitempty"`
	TotalOtrosPagos      string      `xml:"TotalOtrosPagos,attr,omitempty"`
	Emisor_nomina        NominaEmisor `xml:"nomina12:Emisor,omitempty"`
	Receptor_nomina      NominaReceptor `xml:"nomina12:Receptor"`
	Percepciones         *Percepciones `xml:"nomina12:Percepciones,omitempty"`
	Deducciones          *Deducciones  `xml:"nomina12:Deducciones,omitempty"`
	OtrosPagos           *OtrosPagos   `xml:"nomina12:OtrosPagos,omitempty"`
	Incapacidades        *Incapacidades `xml:"nomina12:Incapacidades,omitempty"`
	HorasExtras          *HorasExtras   `xml:"nomina12:HorasExtras,omitempty"`
}

// NominaEmisor for Nomina 1.2
type NominaEmisor struct {
	RegistroPatronal string `xml:"RegistroPatronal,attr,omitempty"`
	Curp             string `xml:"CURP,attr,omitempty"` // Added for CFDI 4.0
}

// NominaReceptor for Nomina 1.2
type NominaReceptor struct {
	Curp                  string `xml:"Curp,attr"`
	NumSeguridadSocial    string `xml:"NumSeguridadSocial,attr,omitempty"`
	FechaInicioRelLaboral string `xml:"FechaInicioRelLaboral,attr"`
	Antiguedad            string `xml:"Antiguedad,attr,omitempty"`
	TipoContrato          string `xml:"TipoContrato,attr"`
	TipoJornada           string `xml:"TipoJornada,attr,omitempty"`
	TipoRegimen           string `xml:"TipoRegimen,attr"`
	NumEmpleado           string `xml:"NumEmpleado,attr,omitempty"`
	Departamento          string `xml:"Departamento,attr,omitempty"`
	Puesto                string `xml:"Puesto,attr,omitempty"`
	RiesgoPuesto          string `xml:"RiesgoPuesto,attr,omitempty"`
	PeriodicidadPago      string `xml:"PeriodicidadPago,attr"`
	Banco                 string `xml:"Banco,attr,omitempty"`
	CuentaBancaria        string `xml:"CuentaBancaria,attr,omitempty"`
	SalarioBaseCotApor    string `xml:"SalarioBaseCotApor,attr,omitempty"`
	SalarioDiarioIntegrado string `xml:"SalarioDiarioIntegrado,attr,omitempty"`
	ClaveEntFed           string `xml:"ClaveEntFed,attr,omitempty"`
}

// Percepciones for Nomina 1.2
type Percepciones struct {
	TotalSueldos            string       `xml:"TotalSueldos,attr,omitempty"`
	TotalSeparacionIndemnizacion string `xml:"TotalSeparacionIndemnizacion,attr,omitempty"`
	TotalJubilacionPensionRetiro string `xml:"TotalJubilacionPensionRetiro,attr,omitempty"`
	TotalGravado            string       `xml:"TotalGravado,attr"`
	TotalExento             string       `xml:"TotalExento,attr"`
	Percepcion              []*Percepcion `xml:"nomina12:Percepcion"`
	SeparacionIndemnizacion *SeparacionIndemnizacion `xml:"nomina12:SeparacionIndemnizacion,omitempty"`
	JubilacionPensionRetiro *JubilacionPensionRetiro `xml:"nomina12:JubilacionPensionRetiro,omitempty"`
}

// Percepcion for Nomina 1.2
type Percepcion struct {
	TipoPercepcion string `xml:"TipoPercepcion,attr"`
	Clave          string `xml:"Clave,attr"`
	Concepto       string `xml:"Concepto,attr"`
	ImporteGravado string `xml:"ImporteGravado,attr"`
	ImporteExento  string `xml:"ImporteExento,attr"`
}

// SeparacionIndemnizacion for Nomina 1.2
type SeparacionIndemnizacion struct {
	TotalPagado          string `xml:"TotalPagado,attr"`
	NumAnosServicio      string `xml:"NumAnosServicio,attr"`
	UltimoSueldoMensOrd  string `xml:"UltimoSueldoMensOrd,attr"`
	IngresoAcumulable    string `xml:"IngresoAcumulable,attr"`
	IngresoNoAcumulable string `xml:"IngresoNoAcumulable,attr"`
}

// JubilacionPensionRetiro for Nomina 1.2
type JubilacionPensionRetiro struct {
	TotalUnaExhibicion string `xml:"TotalUnaExhibicion,attr"`
	TotalParcialidad   string `xml:"TotalParcialidad,attr"`
	MontoDiario        string `xml:"MontoDiario,attr"`
	IngresoAcumulable  string `xml:"IngresoAcumulable,attr"`
	IngresoNoAcumulable string `xml:"IngresoNoAcumulable,attr"`
}

// Deducciones for Nomina 1.2
type Deducciones struct {
	TotalOtrasDeducciones   string      `xml:"TotalOtrasDeducciones,attr,omitempty"`
	TotalImpuestosRetenidos string      `xml:"TotalImpuestosRetenidos,attr,omitempty"`
	Deduccion               []*Deduccion `xml:"nomina12:Deduccion"`
}

// Deduccion for Nomina 1.2
type Deduccion struct {
	TipoDeduccion string `xml:"TipoDeduccion,attr"`
	Clave         string `xml:"Clave,attr"`
	Concepto      string `xml:"Concepto,attr"`
	Importe       string `xml:"Importe,attr"`
}

// OtrosPagos for Nomina 1.2
type OtrosPagos struct {
	OtroPago []*OtroPago `xml:"nomina12:OtroPago"`
}

// OtroPago for Nomina 1.2
type OtroPago struct {
	TipoOtroPago string `xml:"TipoOtroPago,attr"`
	Clave        string `xml:"Clave,attr"`
	Concepto     string `xml:"Concepto,attr"`
	Importe      string `xml:"Importe,attr"`
	SubsidioAlEmpleo *SubsidioAlEmpleo `xml:"nomina12:SubsidioAlEmpleo,omitempty"`
	CompensacionSaldosAFavor *CompensacionSaldosAFavor `xml:"nomina12:CompensacionSaldosAFavor,omitempty"`
}

// SubsidioAlEmpleo for Nomina 1.2
type SubsidioAlEmpleo struct {
	Importe string `xml:"Importe,attr"`
}

// CompensacionSaldosAFavor for Nomina 1.2
type CompensacionSaldosAFavor struct {
	SaldoAFavor     string `xml:"SaldoAFavor,attr"`
	Ano             string `xml:"Ano,attr"`
	RemanenteSalFav string `xml:"RemanenteSalFav,attr"`
}

// Incapacidades for Nomina 1.2
type Incapacidades struct {
	Incapacidad []*Incapacidad `xml:"nomina12:Incapacidad"`
}

// Incapacidad for Nomina 1.2
type Incapacidad struct {
	TipoIncapacidad string `xml:"TipoIncapacidad,attr"`
	DiasIncapacidad string `xml:"DiasIncapacidad,attr"`
	ImporteMonetario string `xml:"ImporteMonetario,attr,omitempty"`
}

// HorasExtras for Nomina 1.2
type HorasExtras struct {
	HorasExtra []*HorasExtra `xml:"nomina12:HorasExtra"`
}

// HorasExtra for Nomina 1.2
type HorasExtra struct {
	Dias         string `xml:"Dias,attr"`
	TipoHoras    string `xml:"TipoHoras,attr"`
	HorasExtra   string `xml:"HorasExtra,attr"`
	ImportePagado string `xml:"ImportePagado,attr"`
}

// Addenda for CFDI 4.0
type Addenda struct {
	XMLName xml.Name `xml:"cfdi:Addenda"`
	// Custom elements can go here
}
