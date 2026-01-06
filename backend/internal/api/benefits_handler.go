package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend/internal/services"
)

type BenefitsHandler struct {
	benefitsService *services.BenefitsService
}

func NewBenefitsHandler(benefitsService *services.BenefitsService) *BenefitsHandler {
	return &BenefitsHandler{benefitsService: benefitsService}
}

// Benefit Plan handlers
func (h *BenefitsHandler) CreateBenefitPlan(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var dto services.CreateBenefitPlanDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	uid := userID.(uuid.UUID)
	dto.CreatedByID = &uid
	plan, err := h.benefitsService.CreateBenefitPlan(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *BenefitsHandler) GetBenefitPlan(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}
	plan, err := h.benefitsService.GetBenefitPlanByID(planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Benefit plan not found"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *BenefitsHandler) ListBenefitPlans(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	benefitType := c.Query("type")

	filters := services.BenefitPlanFilters{
		CompanyID:   companyID.(uuid.UUID),
		BenefitType: benefitType,
		Page:        page,
		Limit:       limit,
	}

	result, err := h.benefitsService.ListBenefitPlans(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Enrollment Period handlers
func (h *BenefitsHandler) CreateEnrollmentPeriod(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateEnrollmentPeriodDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	period, err := h.benefitsService.CreateEnrollmentPeriod(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, period)
}

func (h *BenefitsHandler) GetActiveEnrollmentPeriod(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	period, err := h.benefitsService.GetActiveEnrollmentPeriod(companyID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active enrollment period"})
		return
	}
	c.JSON(http.StatusOK, period)
}

func (h *BenefitsHandler) OpenEnrollmentPeriod(c *gin.Context) {
	periodID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}
	period, err := h.benefitsService.OpenEnrollmentPeriod(periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, period)
}

// Enrollment handlers
func (h *BenefitsHandler) EnrollInBenefit(c *gin.Context) {
	var dto services.EnrollInBenefitDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	enrollment, err := h.benefitsService.EnrollInBenefit(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, enrollment)
}

func (h *BenefitsHandler) GetEnrollment(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	enrollment, err := h.benefitsService.GetEnrollmentByID(enrollmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enrollment not found"})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *BenefitsHandler) GetEmployeeEnrollments(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	enrollments, err := h.benefitsService.GetEmployeeEnrollmentsByBenefit(employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollments)
}

func (h *BenefitsHandler) ApproveEnrollment(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	userID, _ := c.Get("user_id")
	enrollment, err := h.benefitsService.ApproveEnrollment(enrollmentID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *BenefitsHandler) DeclineEnrollment(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	enrollment, err := h.benefitsService.DeclineEnrollment(enrollmentID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *BenefitsHandler) TerminateEnrollment(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	var req struct {
		TerminationDate time.Time `json:"termination_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.TerminationDate = time.Now()
	}
	enrollment, err := h.benefitsService.TerminateEnrollment(enrollmentID, req.TerminationDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *BenefitsHandler) WaiveBenefit(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var req struct {
		EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
		PlanID     uuid.UUID `json:"plan_id" binding:"required"`
		Reason     string    `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	enrollment, err := h.benefitsService.WaiveBenefit(companyID.(uuid.UUID), req.EmployeeID, req.PlanID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, enrollment)
}

// Dependent handlers
func (h *BenefitsHandler) CreateDependent(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var dto services.CreateDependentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.EmployeeID = employeeID
	dependent, err := h.benefitsService.CreateDependent(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dependent)
}

func (h *BenefitsHandler) GetEmployeeDependents(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	dependents, err := h.benefitsService.GetEmployeeDependents(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dependents)
}

func (h *BenefitsHandler) VerifyDependent(c *gin.Context) {
	dependentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dependent ID"})
		return
	}
	userID, _ := c.Get("user_id")
	dependent, err := h.benefitsService.VerifyDependent(dependentID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dependent)
}

// Life Event handlers
func (h *BenefitsHandler) CreateLifeEvent(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var dto services.CreateLifeEventDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.EmployeeID = employeeID
	event, err := h.benefitsService.CreateLifeEvent(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *BenefitsHandler) GetEmployeeLifeEvents(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	events, err := h.benefitsService.GetEmployeeLifeEvents(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *BenefitsHandler) ApproveLifeEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	userID, _ := c.Get("user_id")
	event, err := h.benefitsService.ApproveLifeEvent(eventID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// Beneficiary handlers
func (h *BenefitsHandler) CreateBeneficiary(c *gin.Context) {
	var dto services.CreateBeneficiaryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	beneficiary, err := h.benefitsService.CreateBeneficiary(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, beneficiary)
}

func (h *BenefitsHandler) GetEnrollmentBeneficiaries(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}
	beneficiaries, err := h.benefitsService.GetEnrollmentBeneficiaries(enrollmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, beneficiaries)
}
