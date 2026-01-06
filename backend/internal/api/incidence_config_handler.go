/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/incidence_config_handler.go
==============================================================================

DESCRIPTION:
    Handles incidence tipo mapping configuration endpoints for payroll export.
    Provides CRUD operations for managing mappings between incidence types
    and Excel template configurations.

USER PERSPECTIVE:
    - Configure which incidence types export to which Excel template
    - Set Tipo codes and Motivo values for payroll system
    - View unmapped incidence types that need configuration

DEVELOPER GUIDELINES:
    ✅  OK to modify: Validation logic, response formats
    ⚠️  CAUTION: Authorization (admin/hr only)
    ❌  DO NOT modify: Template type values (must match ExcelExportService)
    📝  All endpoints require authentication

ENDPOINTS:
    GET    /incidence-config/tipo-mappings - List all mappings
    GET    /incidence-config/tipo-mappings/:id - Get single mapping
    POST   /incidence-config/tipo-mappings - Create mapping
    PUT    /incidence-config/tipo-mappings/:id - Update mapping
    DELETE /incidence-config/tipo-mappings/:id - Delete mapping
    GET    /incidence-config/unmapped-types - List unmapped incidence types

==============================================================================
*/
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/models"
	"backend/internal/services"
)

// IncidenceConfigHandler handles incidence configuration endpoints
type IncidenceConfigHandler struct {
	configService *services.IncidenceConfigService
}

// NewIncidenceConfigHandler creates new incidence config handler
func NewIncidenceConfigHandler(configService *services.IncidenceConfigService) *IncidenceConfigHandler {
	return &IncidenceConfigHandler{configService: configService}
}

// RegisterRoutes registers incidence config routes
func (h *IncidenceConfigHandler) RegisterRoutes(router *gin.RouterGroup) {
	config := router.Group("/incidence-config")
	{
		config.GET("/tipo-mappings", h.GetAllTipoMappings)
		config.GET("/tipo-mappings/:id", h.GetTipoMappingByID)
		config.POST("/tipo-mappings", h.CreateTipoMapping)
		config.PUT("/tipo-mappings/:id", h.UpdateTipoMapping)
		config.DELETE("/tipo-mappings/:id", h.DeleteTipoMapping)
		config.GET("/unmapped-types", h.GetUnmappedIncidenceTypes)
	}
}

// GetAllTipoMappings handles fetching all tipo mappings
// @Summary List all tipo mappings
// @Description Returns all incidence tipo mappings with incidence type details
// @Tags Incidence Config
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.IncidenceTipoMapping
// @Failure 500 {object} map[string]string
// @Router /incidence-config/tipo-mappings [get]
func (h *IncidenceConfigHandler) GetAllTipoMappings(c *gin.Context) {
	mappings, err := h.configService.GetAllTipoMappings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tipo mappings", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mappings)
}

// GetTipoMappingByID handles fetching a single tipo mapping
// @Summary Get tipo mapping by ID
// @Description Returns a single tipo mapping with incidence type details
// @Tags Incidence Config
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tipo Mapping ID (UUID)"
// @Success 200 {object} models.IncidenceTipoMapping
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Mapping not found"
// @Failure 500 {object} map[string]string
// @Router /incidence-config/tipo-mappings/{id} [get]
func (h *IncidenceConfigHandler) GetTipoMappingByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	mapping, err := h.configService.GetTipoMappingByID(id)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to get tipo mapping", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// CreateTipoMappingRequest represents the request body for creating a tipo mapping
type CreateTipoMappingRequest struct {
	IncidenceTypeID uuid.UUID `json:"incidence_type_id" binding:"required"`
	TipoCode        *string   `json:"tipo_code"`
	Motivo          *string   `json:"motivo"`
	TemplateType    string    `json:"template_type" binding:"required"`
	HoursMultiplier float64   `json:"hours_multiplier"`
	Notes           *string   `json:"notes"`
}

// CreateTipoMapping handles creating a new tipo mapping
// @Summary Create tipo mapping
// @Description Creates a new incidence tipo mapping
// @Tags Incidence Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTipoMappingRequest true "Tipo mapping data"
// @Success 201 {object} models.IncidenceTipoMapping
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 409 {object} map[string]string "Mapping already exists"
// @Failure 500 {object} map[string]string
// @Router /incidence-config/tipo-mappings [post]
func (h *IncidenceConfigHandler) CreateTipoMapping(c *gin.Context) {
	var req CreateTipoMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "message": err.Error()})
		return
	}

	mapping := &models.IncidenceTipoMapping{
		IncidenceTypeID: req.IncidenceTypeID,
		TipoCode:        req.TipoCode,
		Motivo:          req.Motivo,
		TemplateType:    req.TemplateType,
		HoursMultiplier: req.HoursMultiplier,
		Notes:           req.Notes,
	}

	err := h.configService.CreateTipoMapping(mapping)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not found") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": "Failed to create tipo mapping", "message": err.Error()})
		return
	}

	// Reload with relationships
	createdMapping, _ := h.configService.GetTipoMappingByID(mapping.ID)
	c.JSON(http.StatusCreated, createdMapping)
}

// UpdateTipoMappingRequest represents the request body for updating a tipo mapping
type UpdateTipoMappingRequest struct {
	TipoCode        *string `json:"tipo_code"`
	Motivo          *string `json:"motivo"`
	TemplateType    string  `json:"template_type"`
	HoursMultiplier float64 `json:"hours_multiplier"`
	Notes           *string `json:"notes"`
}

// UpdateTipoMapping handles updating an existing tipo mapping
// @Summary Update tipo mapping
// @Description Updates an existing incidence tipo mapping
// @Tags Incidence Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tipo Mapping ID (UUID)"
// @Param request body UpdateTipoMappingRequest true "Updated mapping data"
// @Success 200 {object} models.IncidenceTipoMapping
// @Failure 400 {object} map[string]string "Invalid ID or validation error"
// @Failure 404 {object} map[string]string "Mapping not found"
// @Failure 500 {object} map[string]string
// @Router /incidence-config/tipo-mappings/{id} [put]
func (h *IncidenceConfigHandler) UpdateTipoMapping(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	var req UpdateTipoMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "message": err.Error()})
		return
	}

	updates := &models.IncidenceTipoMapping{
		TipoCode:        req.TipoCode,
		Motivo:          req.Motivo,
		TemplateType:    req.TemplateType,
		HoursMultiplier: req.HoursMultiplier,
		Notes:           req.Notes,
	}

	err = h.configService.UpdateTipoMapping(id, updates)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": "Failed to update tipo mapping", "message": err.Error()})
		return
	}

	// Reload with relationships
	updatedMapping, _ := h.configService.GetTipoMappingByID(id)
	c.JSON(http.StatusOK, updatedMapping)
}

// DeleteTipoMapping handles deleting a tipo mapping
// @Summary Delete tipo mapping
// @Description Deletes an incidence tipo mapping
// @Tags Incidence Config
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tipo Mapping ID (UUID)"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Mapping not found"
// @Failure 500 {object} map[string]string
// @Router /incidence-config/tipo-mappings/{id} [delete]
func (h *IncidenceConfigHandler) DeleteTipoMapping(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	err = h.configService.DeleteTipoMapping(id)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "Failed to delete tipo mapping", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipo mapping deleted successfully"})
}

// GetUnmappedIncidenceTypes handles fetching incidence types without tipo mappings
// @Summary List unmapped incidence types
// @Description Returns incidence types that don't have a tipo mapping configured
// @Tags Incidence Config
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.IncidenceType
// @Failure 500 {object} map[string]string
// @Router /incidence-config/unmapped-types [get]
func (h *IncidenceConfigHandler) GetUnmappedIncidenceTypes(c *gin.Context) {
	types, err := h.configService.GetUnmappedIncidenceTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unmapped types", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types)
}
