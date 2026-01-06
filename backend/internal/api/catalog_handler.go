/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/catalog_handler.go
==============================================================================

DESCRIPTION:
    Handles catalog/reference data endpoints for payroll concepts and
    incidence types. These are lookup tables used throughout the system.

USER PERSPECTIVE:
    - View available payroll concepts (salary, deductions, benefits)
    - View incidence types for creating employee incidences
    - Create new payroll concepts as needed

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new catalog endpoints
    ⚠️  CAUTION: Changing existing concept categories
    ❌  DO NOT modify: SAT codes on existing concepts
    📝  Catalogs are seeded on first run

SYNTAX EXPLANATION:
    - PayrollConcept: Income/deduction types for payslips
    - IncidenceType: Categories for employee events
    - Catalog data is shared across all companies

ENDPOINTS:
    GET  /catalogs/concepts - List payroll concepts
    GET  /catalogs/incidence-types - List incidence types
    POST /catalogs/concepts - Create new concept

==============================================================================
*/
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/dtos"
	"backend/internal/services"
)

type CatalogHandler struct {
	service *services.CatalogService
}

func NewCatalogHandler(service *services.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/catalogs")
	group.GET("/concepts", h.GetPayrollConcepts)
	group.GET("/incidence-types", h.GetIncidenceTypes)
	group.POST("/concepts", h.CreatePayrollConcept)
}

func (h *CatalogHandler) GetPayrollConcepts(c *gin.Context) {
	concepts, err := h.service.GetPayrollConcepts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payroll concepts"})
		return
	}
	c.JSON(http.StatusOK, concepts)
}

func (h *CatalogHandler) GetIncidenceTypes(c *gin.Context) {
	incidenceTypes, err := h.service.GetIncidenceTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incidence types"})
		return
	}
	c.JSON(http.StatusOK, incidenceTypes)
}

func (h *CatalogHandler) CreatePayrollConcept(c *gin.Context) {
	var req dtos.CreatePayrollConceptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	concept, err := h.service.CreatePayrollConcept(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll concept"})
		return
	}

	c.JSON(http.StatusCreated, concept)
}
