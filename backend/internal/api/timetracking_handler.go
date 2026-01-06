package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend/internal/services"
)

type TimeTrackingHandler struct {
	timeTrackingService *services.TimeTrackingService
}

func NewTimeTrackingHandler(timeTrackingService *services.TimeTrackingService) *TimeTrackingHandler {
	return &TimeTrackingHandler{timeTrackingService: timeTrackingService}
}

// Project handlers
func (h *TimeTrackingHandler) CreateProject(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var dto services.CreateProjectDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	uid := userID.(uuid.UUID)
	dto.CreatedByID = &uid
	project, err := h.timeTrackingService.CreateProject(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (h *TimeTrackingHandler) GetProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	project, err := h.timeTrackingService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *TimeTrackingHandler) ListProjects(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	clientName := c.Query("client_name")
	search := c.Query("search")

	filters := services.ProjectFilters{
		CompanyID:  companyID.(uuid.UUID),
		Status:     status,
		ClientName: clientName,
		Search:     search,
		Page:       page,
		Limit:      limit,
	}

	result, err := h.timeTrackingService.ListProjects(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Project Task handlers
func (h *TimeTrackingHandler) CreateProjectTask(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Code        string  `json:"code"`
		BudgetHours float64 `json:"budget_hours"`
		IsBillable  bool    `json:"is_billable"`
		Order       int     `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.timeTrackingService.CreateProjectTask(projectID, req.Name, req.Description, req.Code, req.BudgetHours, req.IsBillable, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

// Project Member handlers
func (h *TimeTrackingHandler) AddProjectMember(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	var req struct {
		EmployeeID  uuid.UUID `json:"employee_id" binding:"required"`
		Role        string    `json:"role"`
		HourlyRate  float64   `json:"hourly_rate"`
		BudgetHours float64   `json:"budget_hours"`
		CanApprove  bool      `json:"can_approve"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	member, err := h.timeTrackingService.AddProjectMember(projectID, req.EmployeeID, req.Role, req.HourlyRate, req.BudgetHours, req.CanApprove)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, member)
}

// Timesheet handlers
func (h *TimeTrackingHandler) CreateTimesheet(c *gin.Context) {
	var dto services.CreateTimesheetDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	timesheet, err := h.timeTrackingService.CreateTimesheet(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, timesheet)
}

func (h *TimeTrackingHandler) GetTimesheet(c *gin.Context) {
	timesheetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timesheet ID"})
		return
	}
	timesheet, err := h.timeTrackingService.GetTimesheetByID(timesheetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Timesheet not found"})
		return
	}
	c.JSON(http.StatusOK, timesheet)
}

func (h *TimeTrackingHandler) GetEmployeeTimesheets(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	timesheets, err := h.timeTrackingService.GetEmployeeTimesheets(employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timesheets)
}

func (h *TimeTrackingHandler) SubmitTimesheet(c *gin.Context) {
	timesheetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timesheet ID"})
		return
	}
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	timesheet, err := h.timeTrackingService.SubmitTimesheet(timesheetID, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timesheet)
}

func (h *TimeTrackingHandler) ApproveTimesheet(c *gin.Context) {
	timesheetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timesheet ID"})
		return
	}
	userID, _ := c.Get("user_id")
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	timesheet, err := h.timeTrackingService.ApproveTimesheet(timesheetID, userID.(uuid.UUID), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timesheet)
}

func (h *TimeTrackingHandler) RejectTimesheet(c *gin.Context) {
	timesheetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timesheet ID"})
		return
	}
	userID, _ := c.Get("user_id")
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	timesheet, err := h.timeTrackingService.RejectTimesheet(timesheetID, userID.(uuid.UUID), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timesheet)
}

// Time Entry handlers
func (h *TimeTrackingHandler) CreateTimeEntry(c *gin.Context) {
	var dto services.CreateTimeEntryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.timeTrackingService.CreateTimeEntry(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *TimeTrackingHandler) GetTimeEntry(c *gin.Context) {
	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID"})
		return
	}
	entry, err := h.timeTrackingService.GetTimeEntryByID(entryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *TimeTrackingHandler) GetEmployeeTimeEntries(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)

	entries, err := h.timeTrackingService.GetEmployeeTimeEntries(employeeID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// Clock In/Out handlers
func (h *TimeTrackingHandler) ClockIn(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var req struct {
		Source   string `json:"source"`
		Location string `json:"location"`
		Notes    string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	record, err := h.timeTrackingService.ClockIn(companyID.(uuid.UUID), employeeID, req.Source, req.Location, c.ClientIP(), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

func (h *TimeTrackingHandler) ClockOut(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var req struct {
		Source   string `json:"source"`
		Location string `json:"location"`
		Notes    string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	record, err := h.timeTrackingService.ClockOut(employeeID, req.Source, req.Location, c.ClientIP(), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *TimeTrackingHandler) StartBreak(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	record, err := h.timeTrackingService.StartBreak(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *TimeTrackingHandler) EndBreak(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	record, err := h.timeTrackingService.EndBreak(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *TimeTrackingHandler) GetCurrentClockStatus(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	record, err := h.timeTrackingService.GetActiveClockRecord(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active clock record"})
		return
	}
	c.JSON(http.StatusOK, record)
}

// Time-Off Balance handlers
func (h *TimeTrackingHandler) GetTimeOffBalances(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	balance, err := h.timeTrackingService.GetTimeOffBalance(employeeID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, balance)
}

func (h *TimeTrackingHandler) UpdateTimeOffBalance(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var req struct {
		Year        int     `json:"year" binding:"required"`
		BalanceType string  `json:"balance_type" binding:"required"`
		Hours       float64 `json:"hours" binding:"required"`
		IsUsage     bool    `json:"is_usage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	balance, err := h.timeTrackingService.UpdateTimeOffBalance(employeeID, req.Year, req.BalanceType, req.Hours, req.IsUsage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, balance)
}

// Holiday handlers
func (h *TimeTrackingHandler) CreateHoliday(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var req struct {
		Name        string    `json:"name" binding:"required"`
		Date        time.Time `json:"date" binding:"required"`
		HolidayType string    `json:"holiday_type"`
		IsPaid      bool      `json:"is_paid"`
		PaidHours   float64   `json:"paid_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid := userID.(uuid.UUID)
	holiday, err := h.timeTrackingService.CreateHoliday(companyID.(uuid.UUID), req.Name, req.Date, req.HolidayType, req.IsPaid, req.PaidHours, &uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, holiday)
}

func (h *TimeTrackingHandler) ListHolidays(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	holidays, err := h.timeTrackingService.GetHolidays(companyID.(uuid.UUID), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, holidays)
}
