/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/upload_handler.go
==============================================================================

DESCRIPTION:
    Handles file upload endpoints for incidence evidence. Allows uploading
    supporting documents (medical notes, vacation requests, etc.) for
    incidences.

USER PERSPECTIVE:
    - Upload evidence files for incidences (PDFs, images)
    - View list of evidence for an incidence
    - Download or delete evidence files

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new file type support
    ⚠️  CAUTION: File size limits, security validation
    ❌  DO NOT modify: File storage path structure
    📝  Files stored in UPLOAD_PATH from config

SYNTAX EXPLANATION:
    - c.FormFile(): Gets uploaded file from request
    - FileAttachment(): Serves file for download
    - Evidence: Database record linking file to incidence

ENDPOINTS:
    POST   /evidence/incidence/:incidence_id - Upload evidence
    GET    /evidence/incidence/:incidence_id - List evidence
    GET    /evidence/:evidence_id - Get evidence metadata
    GET    /evidence/:evidence_id/download - Download file
    DELETE /evidence/:evidence_id - Delete evidence

FILE VALIDATION:
    - Max size: Configurable (default 10MB)
    - Allowed types: PDF, images, common docs
    - Files renamed with UUID to prevent conflicts

==============================================================================
*/
package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/services"
)

// UploadHandler handles file upload endpoints
type UploadHandler struct {
	uploadService *services.UploadService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// RegisterRoutes registers upload routes
func (h *UploadHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Using /evidence base path to avoid conflicts with /incidences/:id routes
	evidence := router.Group("/evidence")
	{
		evidence.POST("/incidence/:incidence_id", h.UploadEvidence)
		evidence.GET("/incidence/:incidence_id", h.ListEvidence)
		evidence.GET("/:evidence_id", h.GetEvidence)
		evidence.GET("/:evidence_id/download", h.DownloadEvidence)
		evidence.DELETE("/:evidence_id", h.DeleteEvidence)
	}
}

// UploadEvidence handles file upload for incidence evidence
// POST /api/v1/incidences/:incidence_id/evidence
func (h *UploadHandler) UploadEvidence(c *gin.Context) {
	// Parse incidence ID
	incidenceID, err := uuid.Parse(c.Param("incidence_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid incidence ID",
			"message": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, _, _, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No file uploaded",
			"message": "Please provide a file in the 'file' form field",
		})
		return
	}

	// Upload the file
	evidence, err := h.uploadService.UploadEvidence(incidenceID, userID, file)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case services.ErrFileTooLarge:
			status = http.StatusRequestEntityTooLarge
		case services.ErrInvalidFileType, services.ErrInvalidExtension:
			status = http.StatusUnsupportedMediaType
		case services.ErrNoFile:
			status = http.StatusBadRequest
		case services.ErrIncidenceNotFound:
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Upload failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "File uploaded successfully",
		"evidence": evidence,
	})
}

// ListEvidence returns all evidence files for an incidence
// GET /api/v1/incidences/:incidence_id/evidence
func (h *UploadHandler) ListEvidence(c *gin.Context) {
	incidenceID, err := uuid.Parse(c.Param("incidence_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid incidence ID",
			"message": err.Error(),
		})
		return
	}

	evidences, err := h.uploadService.GetEvidenceByIncidence(incidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve evidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, evidences)
}

// GetEvidence returns a specific evidence file metadata
// GET /api/v1/incidences/:incidence_id/evidence/:evidence_id
func (h *UploadHandler) GetEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("evidence_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid evidence ID",
			"message": err.Error(),
		})
		return
	}

	evidence, err := h.uploadService.GetEvidenceByID(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrEvidenceNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Failed to retrieve evidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, evidence)
}

// DownloadEvidence downloads an evidence file
// GET /api/v1/incidences/:incidence_id/evidence/:evidence_id/download
func (h *UploadHandler) DownloadEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("evidence_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid evidence ID",
			"message": err.Error(),
		})
		return
	}

	filePath, originalName, err := h.uploadService.GetFilePath(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrEvidenceNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Failed to retrieve file",
			"message": err.Error(),
		})
		return
	}

	// Serve the file with original filename
	c.FileAttachment(filePath, originalName)
}

// DeleteEvidence deletes an evidence file
// DELETE /api/v1/incidences/:incidence_id/evidence/:evidence_id
func (h *UploadHandler) DeleteEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("evidence_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid evidence ID",
			"message": err.Error(),
		})
		return
	}

	if err := h.uploadService.DeleteEvidence(evidenceID); err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrEvidenceNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Failed to delete evidence",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Evidence deleted successfully"})
}

// getAbsPath returns absolute path for a file
func getAbsPath(path string) string {
	abs, _ := filepath.Abs(path)
	return abs
}
