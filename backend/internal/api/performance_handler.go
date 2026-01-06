package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"backend/internal/services"
)

type PerformanceHandler struct {
	performanceService *services.PerformanceService
}

func NewPerformanceHandler(performanceService *services.PerformanceService) *PerformanceHandler {
	return &PerformanceHandler{performanceService: performanceService}
}

// Review Cycle handlers
func (h *PerformanceHandler) CreateReviewCycle(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")
	var dto services.CreateReviewCycleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	uid := userID.(uuid.UUID)
	dto.CreatedByID = &uid
	cycle, err := h.performanceService.CreateReviewCycle(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cycle)
}

func (h *PerformanceHandler) GetReviewCycle(c *gin.Context) {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cycle ID"})
		return
	}
	cycle, err := h.performanceService.GetReviewCycleByID(cycleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review cycle not found"})
		return
	}
	c.JSON(http.StatusOK, cycle)
}

func (h *PerformanceHandler) ListReviewCycles(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	filters := services.ReviewCycleFilters{
		CompanyID: companyID.(uuid.UUID),
		Status:    status,
		Page:      page,
		Limit:     limit,
	}

	result, err := h.performanceService.ListReviewCycles(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *PerformanceHandler) ActivateReviewCycle(c *gin.Context) {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cycle ID"})
		return
	}
	cycle, err := h.performanceService.ActivateReviewCycle(cycleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cycle)
}

// Performance Review handlers
func (h *PerformanceHandler) GetPerformanceReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	review, err := h.performanceService.GetReviewByID(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *PerformanceHandler) GetEmployeeReviews(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	reviews, err := h.performanceService.GetEmployeeReviews(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

func (h *PerformanceHandler) StartSelfReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	review, err := h.performanceService.StartSelfReview(reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *PerformanceHandler) SubmitSelfReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	var req struct {
		Rating   float64 `json:"rating"`
		Comments string  `json:"comments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	review, err := h.performanceService.SubmitSelfReview(reviewID, req.Rating, req.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *PerformanceHandler) SubmitManagerReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	var req struct {
		Rating             float64 `json:"rating"`
		Comments           string  `json:"comments"`
		OverallPerformance string  `json:"overall_performance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	review, err := h.performanceService.SubmitManagerReview(reviewID, req.Rating, req.Comments, req.OverallPerformance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *PerformanceHandler) AcknowledgeReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	var req struct {
		Comments string `json:"comments"`
	}
	c.ShouldBindJSON(&req)
	review, err := h.performanceService.AcknowledgeReview(reviewID, req.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, review)
}

// Goal handlers
func (h *PerformanceHandler) CreateGoal(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateGoalDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	goal, err := h.performanceService.CreateGoal(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, goal)
}

func (h *PerformanceHandler) GetGoal(c *gin.Context) {
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}
	goal, err := h.performanceService.GetGoalByID(goalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal not found"})
		return
	}
	c.JSON(http.StatusOK, goal)
}

func (h *PerformanceHandler) GetEmployeeGoals(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	goals, err := h.performanceService.GetEmployeeGoals(employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goals)
}

func (h *PerformanceHandler) UpdateGoalProgress(c *gin.Context) {
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}
	userID, _ := c.Get("user_id")
	var req struct {
		Progress int     `json:"progress"`
		NewValue float64 `json:"new_value"`
		Note     string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	goal, err := h.performanceService.UpdateGoalProgress(goalID, req.Progress, req.NewValue, req.Note, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goal)
}

// Feedback handlers
func (h *PerformanceHandler) CreateFeedback(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateFeedbackDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	feedback, err := h.performanceService.CreateFeedback(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, feedback)
}

func (h *PerformanceHandler) GetEmployeeFeedback(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	feedbackType := c.Query("type")
	feedback, err := h.performanceService.GetEmployeeFeedback(employeeID, feedbackType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feedback)
}

// One-on-One handlers
func (h *PerformanceHandler) CreateOneOnOne(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	var dto services.CreateOneOnOneDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto.CompanyID = companyID.(uuid.UUID)
	meeting, err := h.performanceService.CreateOneOnOne(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, meeting)
}

func (h *PerformanceHandler) GetEmployeeOneOnOnes(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	status := c.Query("status")
	meetings, err := h.performanceService.GetEmployeeOneOnOnes(employeeID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meetings)
}

func (h *PerformanceHandler) CompleteOneOnOne(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}
	var req struct {
		Notes       string `json:"notes"`
		ActionItems string `json:"action_items"`
		Mood        string `json:"mood"`
		Engagement  int    `json:"engagement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	meeting, err := h.performanceService.CompleteOneOnOne(meetingID, req.Notes, req.ActionItems, req.Mood, req.Engagement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meeting)
}
