package utils

import (
	"time"
)

// BusinessDayCalculator calculates business days excluding weekends and Mexican federal holidays
// Used for DiasPago calculation in the Vacaciones payroll export template
type BusinessDayCalculator struct {
	holidays map[string]bool // Map of "YYYY-MM-DD" → true for quick lookup
	year     int
}

// NewBusinessDayCalculator creates a calculator for a specific year
// Automatically loads Mexican federal holidays for that year
func NewBusinessDayCalculator(year int) *BusinessDayCalculator {
	calc := &BusinessDayCalculator{
		holidays: make(map[string]bool),
		year:     year,
	}
	calc.loadMexicanHolidays(year)
	return calc
}

// CalculateBusinessDays returns the number of business days between start and end dates (inclusive)
// Excludes weekends (Saturday, Sunday) and Mexican federal holidays
// Example: Friday to Monday (3 calendar days) = 1 business day (only Friday counts)
func (c *BusinessDayCalculator) CalculateBusinessDays(start, end time.Time) float64 {
	if start.After(end) {
		return 0.0
	}

	days := 0.0
	current := start

	// Iterate through each day
	for !current.After(end) {
		if !c.isWeekend(current) && !c.isHoliday(current) {
			days += 1.0
		}
		current = current.AddDate(0, 0, 1) // Next day
	}

	return days
}

// isWeekend returns true if the date is Saturday or Sunday
func (c *BusinessDayCalculator) isWeekend(date time.Time) bool {
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// isHoliday returns true if the date is a Mexican federal holiday
func (c *BusinessDayCalculator) isHoliday(date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	return c.holidays[dateStr]
}

// loadMexicanHolidays loads Mexican federal holidays for the given year
// Based on Mexican labor law (Ley Federal del Trabajo)
func (c *BusinessDayCalculator) loadMexicanHolidays(year int) {
	// Fixed holidays (Article 74 of Mexican Labor Law)
	fixedHolidays := []time.Time{
		time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),   // New Year's Day (Año Nuevo)
		time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC),   // Labor Day (Día del Trabajo)
		time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC),  // Independence Day (Independencia de México)
		time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), // Christmas (Navidad)
	}

	// Add fixed holidays to map
	for _, holiday := range fixedHolidays {
		c.holidays[holiday.Format("2006-01-02")] = true
	}

	// Moveable holidays (observed on specific Mondays)
	c.addFirstMondayOf(year, time.February)  // Constitution Day (Día de la Constitución) - 1st Monday of February
	c.addThirdMondayOf(year, time.March)     // Benito Juárez's Birthday - 3rd Monday of March
	c.addThirdMondayOf(year, time.November)  // Revolution Day (Revolución Mexicana) - 3rd Monday of November

	// President's Inauguration Day (only on inauguration years - every 6 years, e.g., 2024, 2030)
	// December 1st on inauguration years - for simplicity, not adding special logic here
	// Can be added if needed: if year%6 == 0 { c.holidays["YYYY-12-01"] = true }
}

// addFirstMondayOf adds the first Monday of a given month as a holiday
func (c *BusinessDayCalculator) addFirstMondayOf(year int, month time.Month) {
	monday := c.findNthWeekdayOfMonth(year, month, time.Monday, 1)
	c.holidays[monday.Format("2006-01-02")] = true
}

// addThirdMondayOf adds the third Monday of a given month as a holiday
func (c *BusinessDayCalculator) addThirdMondayOf(year int, month time.Month) {
	monday := c.findNthWeekdayOfMonth(year, month, time.Monday, 3)
	c.holidays[monday.Format("2006-01-02")] = true
}

// findNthWeekdayOfMonth finds the Nth occurrence of a weekday in a month
// Example: 3rd Monday of March 2024
func (c *BusinessDayCalculator) findNthWeekdayOfMonth(year int, month time.Month, weekday time.Weekday, n int) time.Time {
	// Start at the 1st of the month
	current := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	count := 0
	for current.Month() == month {
		if current.Weekday() == weekday {
			count++
			if count == n {
				return current
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	// Fallback (should never happen if n is valid)
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

// AddCustomHoliday allows adding company-specific holidays (e.g., local plant closures)
// Date format: "2024-12-24" (YYYY-MM-DD)
func (c *BusinessDayCalculator) AddCustomHoliday(dateStr string) {
	c.holidays[dateStr] = true
}

// RemoveHoliday removes a holiday (useful for testing or special cases)
func (c *BusinessDayCalculator) RemoveHoliday(dateStr string) {
	delete(c.holidays, dateStr)
}

// GetHolidays returns all holidays as a slice of date strings (for debugging)
func (c *BusinessDayCalculator) GetHolidays() []string {
	holidays := make([]string, 0, len(c.holidays))
	for date := range c.holidays {
		holidays = append(holidays, date)
	}
	return holidays
}

// IsBusinessDay returns true if the given date is a business day (not weekend, not holiday)
func (c *BusinessDayCalculator) IsBusinessDay(date time.Time) bool {
	return !c.isWeekend(date) && !c.isHoliday(date)
}
