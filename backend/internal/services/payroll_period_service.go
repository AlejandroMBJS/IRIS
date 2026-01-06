/*
Package services - Payroll Period Service

==============================================================================
FILE: internal/services/payroll_period_service.go
==============================================================================

DESCRIPTION:
    Manages payroll periods (weekly, biweekly, monthly) including creation,
    status tracking, and period lifecycle management.

USER PERSPECTIVE:
    - Create and manage payroll periods for different frequencies
    - Track period status (open, calculated, approved, paid, closed)
    - View historical periods and their associated payrolls
    - Ensure only one active period exists per frequency at a time

DEVELOPER GUIDELINES:
    OK to modify: Period validation rules, add custom period types
    CAUTION: Changing period dates affects all payroll calculations
    DO NOT modify: Period status transitions without updating workflow
    Note: Period status controls which operations are allowed

SYNTAX EXPLANATION:
    - PeriodCode format: "2025-01" for weekly, "2025-Q01" for biweekly
    - Frequency: 'weekly', 'biweekly', 'monthly'
    - Status flow: open -> calculated -> approved -> paid -> closed
    - PeriodType: 'ordinary', 'extraordinary', 'aguinaldo', etc.

==============================================================================
*/
package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"backend/internal/dtos"
	"backend/internal/models"
	"backend/internal/repositories"
)

type PayrollPeriodService struct {
	repo *repositories.PayrollPeriodRepository
}

func NewPayrollPeriodService(db *gorm.DB) *PayrollPeriodService {
	return &PayrollPeriodService{
		repo: repositories.NewPayrollPeriodRepository(db),
	}
}

func (s *PayrollPeriodService) GetPeriods(filters map[string]interface{}) ([]models.PayrollPeriod, error) {
	return s.repo.GetPeriods(filters)
}

func (s *PayrollPeriodService) GetPeriod(id uuid.UUID) (*models.PayrollPeriod, error) {
	return s.repo.FindByID(id)
}

func (s *PayrollPeriodService) CreatePeriod(req dtos.CreatePayrollPeriodRequest) (*models.PayrollPeriod, error) {
	// Validate date relationships
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end date (%s) cannot be before start date (%s)",
			req.EndDate.Format("2006-01-02"), req.StartDate.Format("2006-01-02"))
	}

	if req.PaymentDate.Before(req.EndDate) {
		return nil, fmt.Errorf("payment date (%s) cannot be before end date (%s) - employees must be paid after period ends",
			req.PaymentDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02"))
	}

	// Validate payment date is not too far in the past (prevents backdating)
	if req.PaymentDate.Before(time.Now().AddDate(0, 0, -30)) {
		return nil, fmt.Errorf("payment date (%s) is more than 30 days in the past - backdating is not allowed",
			req.PaymentDate.Format("2006-01-02"))
	}

	// Validate period duration is reasonable
	periodDuration := req.EndDate.Sub(req.StartDate).Hours() / 24
	switch req.Frequency {
	case "weekly":
		if periodDuration < 6 || periodDuration > 8 {
			return nil, fmt.Errorf("weekly payroll period must be 6-8 days, got %.0f days", periodDuration)
		}
	case "biweekly":
		if periodDuration < 13 || periodDuration > 15 {
			return nil, fmt.Errorf("biweekly payroll period must be 13-15 days, got %.0f days", periodDuration)
		}
	case "monthly":
		if periodDuration < 28 || periodDuration > 32 {
			return nil, fmt.Errorf("monthly payroll period must be 28-32 days, got %.0f days", periodDuration)
		}
	}

	// Check for overlapping periods with same frequency
	hasOverlap, err := s.repo.HasOverlappingPeriod(req.Frequency, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("error checking for overlapping periods: %w", err)
	}
	if hasOverlap {
		return nil, fmt.Errorf("payroll period overlaps with existing %s period", req.Frequency)
	}

	period := &models.PayrollPeriod{
		PeriodCode:   req.PeriodCode,
		Year:         req.Year,
		PeriodNumber: req.PeriodNumber,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		PaymentDate:  req.PaymentDate,
		Frequency:    req.Frequency,
		PeriodType:   req.PeriodType,
		Description:  req.Description,
		Status:       "open",
	}

	if err := s.repo.Create(period); err != nil {
		return nil, fmt.Errorf("could not create payroll period: %w", err)
	}

	return period, nil
}

// GenerateCurrentPeriods generates the current payroll periods based on today's date
// Weekly: Pays every Friday (period is Mon-Sun)
// Biweekly: Pays every 2 Fridays (reference: Dec 5, 2025 was a biweekly payday)
// Monthly: Pays every 4 Fridays
func (s *PayrollPeriodService) GenerateCurrentPeriods() ([]models.PayrollPeriod, error) {
	var generatedPeriods []models.PayrollPeriod
	today := time.Now()

	// Reference biweekly payment date: December 5, 2025 (Friday)
	biweeklyReference := time.Date(2025, 12, 5, 0, 0, 0, 0, time.Local)

	// Generate Weekly Period
	weeklyPeriod, err := s.generateWeeklyPeriod(today)
	if err == nil && weeklyPeriod != nil {
		generatedPeriods = append(generatedPeriods, *weeklyPeriod)
	}

	// Generate Biweekly Period
	biweeklyPeriod, err := s.generateBiweeklyPeriod(today, biweeklyReference)
	if err == nil && biweeklyPeriod != nil {
		generatedPeriods = append(generatedPeriods, *biweeklyPeriod)
	}

	// Generate Monthly Period (every 4 weeks)
	monthlyPeriod, err := s.generateMonthlyPeriod(today, biweeklyReference)
	if err == nil && monthlyPeriod != nil {
		generatedPeriods = append(generatedPeriods, *monthlyPeriod)
	}

	return generatedPeriods, nil
}

// generateWeeklyPeriod creates the current weekly period
// Period runs Monday to Friday (working days only), payment on Friday
func (s *PayrollPeriodService) generateWeeklyPeriod(today time.Time) (*models.PayrollPeriod, error) {
	// Find the Friday of the current week (payment day)
	weekday := int(today.Weekday())
	daysUntilFriday := (5 - weekday + 7) % 7
	if weekday == 5 {
		daysUntilFriday = 0 // Today is Friday
	} else if weekday == 6 {
		daysUntilFriday = 6 // Saturday -> next Friday
	} else if weekday == 0 {
		daysUntilFriday = 5 // Sunday -> next Friday
	}

	friday := today.AddDate(0, 0, daysUntilFriday)
	friday = time.Date(friday.Year(), friday.Month(), friday.Day(), 0, 0, 0, 0, time.Local)

	// Period runs Monday to Thursday (payment is Friday, after period ends)
	monday := friday.AddDate(0, 0, -4)   // Start of period (Monday)
	thursday := friday.AddDate(0, 0, -1) // End of period (Thursday before payment)

	// Calculate period number (week of year based on Friday)
	_, periodNumber := friday.ISOWeek()
	year := friday.Year()

	periodCode := fmt.Sprintf("%d-W%02d", year, periodNumber)

	// Check if period already exists
	existing, _ := s.repo.FindByPeriodCode(periodCode)
	if existing != nil {
		return existing, nil // Already exists, return it
	}

	period := &models.PayrollPeriod{
		PeriodCode:   periodCode,
		Year:         year,
		PeriodNumber: periodNumber,
		StartDate:    monday,
		EndDate:      thursday,
		PaymentDate:  friday,
		Frequency:    "weekly",
		PeriodType:   "weekly",
		Description:  fmt.Sprintf("Semana %d - %s al %s", periodNumber, monday.Format("02/01"), thursday.Format("02/01/2006")),
		Status:       "open",
	}

	if err := s.repo.Create(period); err != nil {
		return nil, err
	}

	return period, nil
}

// generateBiweeklyPeriod creates the current biweekly period
// Reference: Dec 5, 2025 was a payday, so we calculate from there
// Period is 2 weeks of work days (Mon-Fri), paid on the 2nd Friday
func (s *PayrollPeriodService) generateBiweeklyPeriod(today time.Time, reference time.Time) (*models.PayrollPeriod, error) {
	// Calculate days since reference
	daysSinceRef := int(today.Sub(reference).Hours() / 24)

	// Each biweekly period is 14 calendar days
	// Find which period we're in
	periodOffset := daysSinceRef / 14
	if daysSinceRef < 0 {
		periodOffset = (daysSinceRef - 13) / 14 // Handle dates before reference
	}

	// Calculate the payment date for current period (Friday)
	currentPaymentDate := reference.AddDate(0, 0, periodOffset*14)

	// If today is after the payment date, move to next period
	if today.After(currentPaymentDate) {
		periodOffset++
		currentPaymentDate = reference.AddDate(0, 0, periodOffset*14)
	}

	// Period starts on Monday 11 days before payment (2 weeks of Mon-Thu work)
	// Week 1: Mon-Thu (4 days), Week 2: Mon-Thu (4 days), paid Friday of week 2
	periodStart := currentPaymentDate.AddDate(0, 0, -11)  // Monday of first week
	periodEnd := currentPaymentDate.AddDate(0, 0, -1)     // Thursday before payment

	year := periodStart.Year()
	// Calculate biweekly period number (1-26 per year)
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	dayOfYear := int(periodStart.Sub(startOfYear).Hours()/24) + 1
	periodNumber := (dayOfYear / 14) + 1
	if periodNumber > 26 {
		periodNumber = 26
	}

	periodCode := fmt.Sprintf("%d-BW%02d", year, periodNumber)

	// Check if period already exists
	existing, _ := s.repo.FindByPeriodCode(periodCode)
	if existing != nil {
		return existing, nil
	}

	period := &models.PayrollPeriod{
		PeriodCode:   periodCode,
		Year:         year,
		PeriodNumber: periodNumber,
		StartDate:    periodStart,
		EndDate:      periodEnd,
		PaymentDate:  currentPaymentDate,
		Frequency:    "biweekly",
		PeriodType:   "biweekly",
		Description:  fmt.Sprintf("Quincena %d - %s al %s", periodNumber, periodStart.Format("02/01"), periodEnd.Format("02/01/2006")),
		Status:       "open",
	}

	if err := s.repo.Create(period); err != nil {
		return nil, err
	}

	return period, nil
}

// generateMonthlyPeriod creates the current monthly period (every 4 weeks)
// Period is 4 weeks of work days (Mon-Thu each week), paid on the 4th Friday
func (s *PayrollPeriodService) generateMonthlyPeriod(today time.Time, reference time.Time) (*models.PayrollPeriod, error) {
	// Monthly is every 4 weeks (28 calendar days)
	daysSinceRef := int(today.Sub(reference).Hours() / 24)

	periodOffset := daysSinceRef / 28
	if daysSinceRef < 0 {
		periodOffset = (daysSinceRef - 27) / 28
	}

	currentPaymentDate := reference.AddDate(0, 0, periodOffset*28)

	if today.After(currentPaymentDate) {
		periodOffset++
		currentPaymentDate = reference.AddDate(0, 0, periodOffset*28)
	}

	// Period starts on Monday 25 days before payment (4 weeks of Mon-Thu work)
	periodStart := currentPaymentDate.AddDate(0, 0, -25) // Monday of first week
	periodEnd := currentPaymentDate.AddDate(0, 0, -1)    // Thursday before payment

	year := periodStart.Year()
	// Calculate monthly period number (1-13 per year)
	periodNumber := int(periodStart.Month())

	periodCode := fmt.Sprintf("%d-M%02d", year, periodNumber)

	// Check if period already exists
	existing, _ := s.repo.FindByPeriodCode(periodCode)
	if existing != nil {
		return existing, nil
	}

	period := &models.PayrollPeriod{
		PeriodCode:   periodCode,
		Year:         year,
		PeriodNumber: periodNumber,
		StartDate:    periodStart,
		EndDate:      periodEnd,
		PaymentDate:  currentPaymentDate,
		Frequency:    "monthly",
		PeriodType:   "monthly",
		Description:  fmt.Sprintf("Mes %d - %s al %s", periodNumber, periodStart.Format("02/01"), periodEnd.Format("02/01/2006")),
		Status:       "open",
	}

	if err := s.repo.Create(period); err != nil {
		return nil, err
	}

	return period, nil
}
