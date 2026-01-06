/*
Package api - IRIS Payroll System HTTP API Handlers

==============================================================================
FILE: internal/api/employee_handler.go
==============================================================================

DESCRIPTION:
    Handles all employee management endpoints: CRUD operations, salary
    updates, terminations, and bulk import functionality.

USER PERSPECTIVE:
    - View, create, edit, and delete employees
    - Update employee salaries with history tracking
    - Terminate employees with reason and date
    - Bulk import employees from Excel/CSV files
    - Download import template

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new employee-related endpoints
    ⚠️  CAUTION: Validation logic, Mexican ID formats
    ❌  DO NOT modify: Import template column order
    📝  All employee operations track created_by/updated_by

SYNTAX EXPLANATION:
    - dtos.EmployeeRequest: Data transfer object for validation
    - c.ShouldBindQuery(): Parses query parameters
    - c.Request.FormFile(): Handles multipart file upload
    - middleware.GetUserFromContext(): Gets authenticated user

ENDPOINTS:
    GET    /employees - List employees with pagination/filters
    GET    /employees/:id - Get employee details
    POST   /employees - Create new employee
    PUT    /employees/:id - Update employee
    DELETE /employees/:id - Soft delete employee
    POST   /employees/:id/terminate - Terminate with reason
    PUT    /employees/:id/salary - Update salary
    GET    /employees/stats - Get employee statistics
    POST   /employees/validate-ids - Validate RFC/CURP/NSS
    POST   /employees/import - Bulk import from file
    GET    /employees/import/template - Download import template

IMPORT FILE FORMAT:
    - Supports .xlsx, .xls, and .csv
    - Template includes all required columns with examples

==============================================================================
*/
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/dtos"
	"backend/internal/middleware"
	"backend/internal/services"
)

// EmployeeHandler handles employee endpoints
type EmployeeHandler struct {
    employeeService *services.EmployeeService
}

// NewEmployeeHandler creates new employee handler
func NewEmployeeHandler(employeeService *services.EmployeeService) *EmployeeHandler {
    return &EmployeeHandler{employeeService: employeeService}
}

// RegisterRoutes registers employee routes
func (h *EmployeeHandler) RegisterRoutes(router *gin.RouterGroup) {
    employees := router.Group("/employees")
    {
        employees.GET("", h.ListEmployees)
        employees.GET("/:id", h.GetEmployee)
        employees.POST("", h.CreateEmployee)
        employees.PUT("/:id", h.UpdateEmployee)
        employees.DELETE("/:id", h.DeleteEmployee)
        employees.POST("/:id/terminate", h.TerminateEmployee)
        employees.PUT("/:id/salary", h.UpdateSalary)
        employees.GET("/stats", h.GetEmployeeStats)
        employees.POST("/validate-ids", h.ValidateMexicanIDs)
        employees.POST("/import", h.ImportEmployees)
        employees.GET("/import/template", h.DownloadTemplate)
        // Portal credentials management
        employees.GET("/:id/portal-user", h.GetEmployeePortalUser)
        employees.POST("/:id/portal-user", h.CreateEmployeePortalUser)
        employees.PUT("/:id/portal-user", h.UpdateEmployeePortalUser)
        employees.DELETE("/:id/portal-user", h.DeleteEmployeePortalUser)
    }
}

// ListEmployees handles employee listing
// @Summary List employees
// @Description Get paginated list of employees with filtering
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search term"
// @Param status query string false "Employment status"
// @Param employee_type query string false "Employee type"
// @Param department_id query string false "Department ID"
// @Success 200 {object} dtos.EmployeeListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees [get]
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
    var req dtos.EmployeeSearchRequest
    
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Set defaults
    if req.Page == 0 {
        req.Page = 1
    }
    if req.PageSize == 0 {
        req.PageSize = 20
    }
    
    // Build filters
    filters := make(map[string]interface{})
    if req.Search != "" {
        filters["search"] = req.Search
    }
    if req.Status != "" {
        filters["status"] = req.Status
    }
    if req.EmployeeType != "" {
        filters["employee_type"] = req.EmployeeType
    }
    if req.DepartmentID != "" {
        deptID, err := uuid.Parse(req.DepartmentID)
        if err == nil {
            filters["department_id"] = deptID
        }
    }
    
    // Get user info for role-based filtering
    userID, _, _, _ := middleware.GetUserFromContext(c)
    userRole := middleware.GetUserRoleFromContext(c)

    // Get employees (pass role and user ID for supervisor filtering)
    response, err := h.employeeService.ListEmployees(req.Page, req.PageSize, filters, userRole, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Internal Server Error",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, response)
}

// GetEmployee handles getting employee details
// @Summary Get employee
// @Description Get employee details by ID
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Success 200 {object} dtos.EmployeeResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees/{id} [get]
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }

    // Get user role for collar type access control
    userRole := middleware.GetUserRoleFromContext(c)

    employee, err := h.employeeService.GetEmployee(id, userRole)
    if err != nil {
        if err.Error() == "employee not found" {
            c.JSON(http.StatusNotFound, gin.H{
                "error":   "Not Found",
                "message": "Employee not found",
            })
        } else if strings.Contains(err.Error(), "access denied") {
            c.JSON(http.StatusForbidden, gin.H{
                "error":   "Forbidden",
                "message": err.Error(),
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Internal Server Error",
                "message": err.Error(),
            })
        }
        return
    }

    c.JSON(http.StatusOK, employee)
}

// CreateEmployee handles employee creation
// @Summary Create employee
// @Description Create a new employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.EmployeeRequest true "Employee data"
// @Success 201 {object} dtos.EmployeeResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees [post]
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
    var req dtos.EmployeeRequest
    
    // Get user from context
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Create employee
    employee, err := h.employeeService.CreateEmployee(req, userID)
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "already exists") {
            status = http.StatusConflict
        } else if strings.Contains(err.Error(), "validation") {
            status = http.StatusBadRequest
        }
        
        c.JSON(status, gin.H{
            "error":   "Employee Creation Failed",
            "message": err.Error(),
        })
        return
    }
    
    // Convert to response
    response := h.employeeService.ConvertToResponse(employee) // Updated from convertToResponse
    
    c.JSON(http.StatusCreated, response)
}

// UpdateEmployee handles employee updates
// @Summary Update employee
// @Description Update existing employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Param request body dtos.EmployeeRequest true "Employee data"
// @Success 200 {object} dtos.EmployeeResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees/{id} [put]
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }
    
    var req dtos.EmployeeRequest
    
    // Get user from context
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    // Update employee
    response, err := h.employeeService.UpdateEmployee(id, req, userID)
    if err != nil {
        status := http.StatusInternalServerError
        if err.Error() == "employee not found" {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "already exists") {
            status = http.StatusConflict
        } else if strings.Contains(err.Error(), "validation") {
            status = http.StatusBadRequest
        }
        
        c.JSON(status, gin.H{
            "error":   "Employee Update Failed",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// DeleteEmployee handles employee deletion
// @Summary Delete employee
// @Description Soft delete an employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees/{id} [delete]
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }
    
    err = h.employeeService.DeleteEmployee(id)
    if err != nil {
        status := http.StatusInternalServerError
        if err.Error() == "employee not found" {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "cannot delete active") {
            status = http.StatusBadRequest
        }
        
        c.JSON(status, gin.H{
            "error":   "Employee Deletion Failed",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Employee deleted successfully",
    })
}

// TerminateEmployee handles employee termination
// @Summary Terminate employee
// @Description Terminate an employee with reason and date
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Param request body dtos.EmployeeTerminationRequest true "Termination data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees/{id}/terminate [post]
func (h *EmployeeHandler) TerminateEmployee(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }
    
    var req dtos.EmployeeTerminationRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    err = h.employeeService.TerminateEmployee(id, req)
    if err != nil {
        status := http.StatusInternalServerError
        if err.Error() == "employee not found" {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "already terminated") {
            status = http.StatusBadRequest
        } else if strings.Contains(err.Error(), "cannot be before") {
            status = http.StatusBadRequest
        }
        
        c.JSON(status, gin.H{
            "error":   "Employee Termination Failed",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Employee terminated successfully",
    })
}

// UpdateSalary handles salary updates
// @Summary Update employee salary
// @Description Update employee salary with effective date
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Param request body dtos.EmployeeSalaryUpdateRequest true "Salary update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees/{id}/salary [put]
func (h *EmployeeHandler) UpdateSalary(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }
    
    var req dtos.EmployeeSalaryUpdateRequest
    
    // Get user from context
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    err = h.employeeService.UpdateEmployeeSalary(id, req, userID)
    if err != nil {
        status := http.StatusInternalServerError
        if err.Error() == "employee not found" {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "must be positive") {
            status = http.StatusBadRequest
        }
        
        c.JSON(status, gin.H{
            "error":   "Salary Update Failed",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Salary updated successfully",
    })
}

// GetEmployeeStats handles employee statistics
// @Summary Get employee statistics
// @Description Get statistics about employees
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /employees/stats [get]
func (h *EmployeeHandler) GetEmployeeStats(c *gin.Context) {
    // Get user role for collar type filtering
    userRole := middleware.GetUserRoleFromContext(c)

    stats, err := h.employeeService.GetEmployeeStats(userRole)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Internal Server Error",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, stats)
}

// ValidateMexicanIDs handles ID validation
// @Summary Validate Mexican IDs
// @Description Validate RFC, CURP, and NSS formats
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rfc query string false "RFC to validate"
// @Param curp query string false "CURP to validate"
// @Param nss query string false "NSS to validate"
// @Success 200 {object} map[string]bool
// @Failure 500 {object} map[string]string
// @Router /employees/validate-ids [post]
func (h *EmployeeHandler) ValidateMexicanIDs(c *gin.Context) {
    rfc := c.Query("rfc")
    curp := c.Query("curp")
    nss := c.Query("nss")
    
    validation, err := h.employeeService.ValidateMexicanIDs(rfc, curp, nss)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, validation)
}

// ImportEmployees handles bulk employee import from Excel/CSV
// @Summary Import employees from Excel/CSV
// @Description Upload an Excel or CSV file to bulk import employees
// @Tags Employees
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Excel/CSV file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /employees/import [post]
func (h *EmployeeHandler) ImportEmployees(c *gin.Context) {
    // Get user from context
    userID, _, _, err := middleware.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error":   "Unauthorized",
            "message": "User not authenticated",
        })
        return
    }

    // Get file from form
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "File Required",
            "message": "Please upload an Excel (.xlsx, .xls) or CSV file",
        })
        return
    }
    defer file.Close()

    // Validate file type
    filename := header.Filename
    if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") &&
       !strings.HasSuffix(strings.ToLower(filename), ".xls") &&
       !strings.HasSuffix(strings.ToLower(filename), ".csv") {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid File Type",
            "message": "Only Excel (.xlsx, .xls) and CSV files are supported",
        })
        return
    }

    // Import employees
    result, err := h.employeeService.ImportEmployeesFromFile(file, filename, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Import Failed",
            "message": err.Error(),
        })
        return
    }

    // Transform response to match frontend expected format
    total := 0
    if t, ok := result["total"].(int); ok {
        total = t
    }
    created := 0
    if c, ok := result["created"].(int); ok {
        created = c
    }
    updated := 0
    if u, ok := result["updated"].(int); ok {
        updated = u
    }
    failed := 0
    if f, ok := result["failed"].(int); ok {
        failed = f
    }
    errors := result["errors"]

    c.JSON(http.StatusOK, gin.H{
        "message":    "Import completed",
        "total_rows": total,
        "imported":   created + updated,
        "created":    created,
        "updated":    updated,
        "failed":     failed,
        "errors":     errors,
    })
}

// DownloadTemplate returns the employee import template
// @Summary Download employee import template
// @Description Download an Excel template for bulk employee import
// @Tags Employees
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} binary
// @Failure 500 {object} map[string]string
// @Router /employees/import/template [get]
func (h *EmployeeHandler) DownloadTemplate(c *gin.Context) {
    templateData, err := h.employeeService.GenerateImportTemplate()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Template Generation Failed",
            "message": err.Error(),
        })
        return
    }

    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Header("Content-Disposition", "attachment; filename=employee_import_template.xlsx")
    c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", templateData)
}

// =========================================================================
// Portal User Management Endpoints
// =========================================================================

// GetEmployeePortalUser gets the portal user account for an employee
// @Summary Get employee portal user
// @Description Get the portal user account linked to an employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /employees/{id}/portal-user [get]
func (h *EmployeeHandler) GetEmployeePortalUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }

    user, err := h.employeeService.GetEmployeePortalUser(id)
    if err != nil {
        if strings.Contains(err.Error(), "not found") {
            c.JSON(http.StatusNotFound, gin.H{
                "error":   "Not Found",
                "message": err.Error(),
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Internal Server Error",
                "message": err.Error(),
            })
        }
        return
    }

    c.JSON(http.StatusOK, user)
}

// CreateEmployeePortalUser creates a portal user account for an employee
// @Summary Create employee portal user
// @Description Create a new portal user account for an employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Param request body dtos.CreatePortalUserRequest true "Portal user data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /employees/{id}/portal-user [post]
func (h *EmployeeHandler) CreateEmployeePortalUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }

    var req dtos.CreatePortalUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }

    user, err := h.employeeService.CreateEmployeePortalUser(id, req)
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "already has") {
            status = http.StatusConflict
        } else if strings.Contains(err.Error(), "not found") {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "email") {
            status = http.StatusBadRequest
        }

        c.JSON(status, gin.H{
            "error":   "Portal User Creation Failed",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusCreated, user)
}

// UpdateEmployeePortalUser updates a portal user account for an employee
// @Summary Update employee portal user
// @Description Update the portal user account for an employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Param request body dtos.UpdatePortalUserRequest true "Portal user data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /employees/{id}/portal-user [put]
func (h *EmployeeHandler) UpdateEmployeePortalUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }

    var req dtos.UpdatePortalUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Validation Error",
            "message": err.Error(),
        })
        return
    }

    user, err := h.employeeService.UpdateEmployeePortalUser(id, req)
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "not found") {
            status = http.StatusNotFound
        } else if strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "email") {
            status = http.StatusBadRequest
        }

        c.JSON(status, gin.H{
            "error":   "Portal User Update Failed",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, user)
}

// DeleteEmployeePortalUser deletes a portal user account for an employee
// @Summary Delete employee portal user
// @Description Delete the portal user account for an employee
// @Tags Employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /employees/{id}/portal-user [delete]
func (h *EmployeeHandler) DeleteEmployeePortalUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid ID",
            "message": "Invalid employee ID format",
        })
        return
    }

    err = h.employeeService.DeleteEmployeePortalUser(id)
    if err != nil {
        status := http.StatusInternalServerError
        if strings.Contains(err.Error(), "not found") {
            status = http.StatusNotFound
        }

        c.JSON(status, gin.H{
            "error":   "Portal User Deletion Failed",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Portal user deleted successfully",
    })
}
