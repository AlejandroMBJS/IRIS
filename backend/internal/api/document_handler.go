package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend/internal/services"
)

type DocumentHandler struct {
	documentService *services.DocumentService
}

func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

// Document Category handlers
func (h *DocumentHandler) CreateCategory(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateDocumentCategoryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	category, err := h.documentService.CreateDocumentCategory(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *DocumentHandler) ListCategories(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	categories, err := h.documentService.GetDocumentCategories(companyID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Document handlers
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var dto services.CreateDocumentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	uid := userID.(uuid.UUID)
	dto.UploadedByID = &uid
	document, err := h.documentService.CreateDocument(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, document)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	document, err := h.documentService.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	// Log access
	userID, _ := c.Get("user_id")
	h.documentService.LogDocumentAccess(documentID, userID.(uuid.UUID), "view", c.ClientIP(), c.GetHeader("User-Agent"), "")
	c.JSON(http.StatusOK, document)
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	categoryIDStr := c.Query("category_id")
	status := c.Query("status")
	search := c.Query("search")

	filters := services.DocumentFilters{
		CompanyID: companyID.(uuid.UUID),
		Status:    status,
		Search:    search,
		Page:      page,
		Limit:     limit,
	}
	if categoryIDStr != "" {
		if id, err := uuid.Parse(categoryIDStr); err == nil {
			filters.CategoryID = &id
		}
	}

	result, err := h.documentService.ListDocuments(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	var req struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		Tags        []string   `json:"tags"`
		ExpiryDate  *time.Time `json:"expiry_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	document, err := h.documentService.UpdateDocument(documentID, req.Name, req.Description, req.Tags, req.ExpiryDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, document)
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	if err := h.documentService.DeleteDocument(documentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

func (h *DocumentHandler) ArchiveDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	document, err := h.documentService.ArchiveDocument(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, document)
}

// Document Version handlers
func (h *DocumentHandler) CreateDocumentVersion(c *gin.Context) {
	parentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var dto services.CreateDocumentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	uid := userID.(uuid.UUID)
	dto.UploadedByID = &uid
	document, err := h.documentService.CreateDocumentVersion(parentID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, document)
}

// Document Template handlers
func (h *DocumentHandler) CreateTemplate(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateDocumentTemplateDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	template, err := h.documentService.CreateDocumentTemplate(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, template)
}

func (h *DocumentHandler) ListTemplates(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	templateType := c.Query("type")
	templates, err := h.documentService.GetDocumentTemplates(companyID.(uuid.UUID), templateType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

// Signature Request handlers
func (h *DocumentHandler) CreateSignatureRequest(c *gin.Context) {
	var dto services.CreateSignatureRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request, err := h.documentService.CreateSignatureRequest(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

func (h *DocumentHandler) GetSignatureRequest(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}
	request, err := h.documentService.GetSignatureRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Signature request not found"})
		return
	}
	c.JSON(http.StatusOK, request)
}

func (h *DocumentHandler) SignDocument(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}
	signerID, err := uuid.Parse(c.Param("signerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signer ID"})
		return
	}
	var req struct {
		SignatureData string `json:"signature_data" binding:"required"`
		SignatureType string `json:"signature_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.documentService.SignDocument(requestID, signerID, req.SignatureData, req.SignatureType, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document signed"})
}

func (h *DocumentHandler) DeclineSignature(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}
	signerID, err := uuid.Parse(c.Param("signerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signer ID"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	if err := h.documentService.DeclineSignature(requestID, signerID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Signature declined"})
}

// Document Sharing handlers
func (h *DocumentHandler) ShareDocument(c *gin.Context) {
	var dto services.ShareDocumentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	share, err := h.documentService.ShareDocument(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, share)
}

func (h *DocumentHandler) GetDocumentShares(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	shares, err := h.documentService.GetDocumentShares(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shares)
}

func (h *DocumentHandler) RevokeDocumentShare(c *gin.Context) {
	shareID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}
	if err := h.documentService.RevokeShare(shareID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document share revoked"})
}

// Employee Document handlers
func (h *DocumentHandler) SubmitEmployeeDocument(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var dto services.CreateEmployeeDocumentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.EmployeeID = employeeID
	document, err := h.documentService.SubmitEmployeeDocument(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, document)
}

func (h *DocumentHandler) GetEmployeeDocuments(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	documents, err := h.documentService.GetEmployeeDocuments(employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, documents)
}

func (h *DocumentHandler) VerifyEmployeeDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	userID, _ := c.Get("user_id")
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	document, err := h.documentService.VerifyEmployeeDocument(documentID, userID.(uuid.UUID), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, document)
}

func (h *DocumentHandler) RejectEmployeeDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	document, err := h.documentService.RejectEmployeeDocument(documentID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, document)
}
