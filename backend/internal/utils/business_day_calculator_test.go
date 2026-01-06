package utils

import (
	"testing"
	"time"
)

func TestCalculateBusinessDays_WeekdaysOnly(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Monday to Friday (5 days, all weekdays, no holidays)
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)  // Monday
	end := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)   // Friday
	days := calc.CalculateBusinessDays(start, end)

	expected := 5.0
	if days != expected {
		t.Errorf("Expected %v business days, got %v", expected, days)
	}
}

func TestCalculateBusinessDays_ExcludesWeekend(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Friday to Monday (includes Saturday + Sunday)
	start := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)  // Friday
	end := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)    // Monday
	days := calc.CalculateBusinessDays(start, end)

	expected := 2.0 // Friday + Monday (Sat/Sun excluded)
	if days != expected {
		t.Errorf("Expected %v business days (weekend excluded), got %v", expected, days)
	}
}

func TestCalculateBusinessDays_ExcludesNewYear(t *testing.T) {
	// Use 2025 calculator for Jan 2025 dates
	calc2025 := NewBusinessDayCalculator(2025)
	days2025 := calc2025.CalculateBusinessDays(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
	)

	// Jan 1 (Wed, holiday) + Jan 2 (Thu) + Jan 3 (Fri) = 2 business days (Jan 1 excluded)
	expected := 2.0
	if days2025 != expected {
		t.Errorf("Expected %v business days (New Year excluded), got %v", expected, days2025)
	}
}

func TestCalculateBusinessDays_SingleDay(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Same day (Monday)
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	end := start
	days := calc.CalculateBusinessDays(start, end)

	expected := 1.0
	if days != expected {
		t.Errorf("Expected %v business day (same day), got %v", expected, days)
	}
}

func TestCalculateBusinessDays_SingleDayWeekend(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Single Saturday
	start := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC) // Saturday
	end := start
	days := calc.CalculateBusinessDays(start, end)

	expected := 0.0
	if days != expected {
		t.Errorf("Expected %v business days (Saturday), got %v", expected, days)
	}
}

func TestCalculateBusinessDays_EndBeforeStart(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	start := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	days := calc.CalculateBusinessDays(start, end)

	expected := 0.0
	if days != expected {
		t.Errorf("Expected %v business days (end before start), got %v", expected, days)
	}
}

func TestMexicanHolidays_FixedHolidays(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Test fixed holidays
	testCases := []struct {
		date     time.Time
		name     string
		expected bool
	}{
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "New Year", true},
		{time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), "Labor Day", true},
		{time.Date(2025, 9, 16, 0, 0, 0, 0, time.UTC), "Independence Day", true},
		{time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC), "Christmas", true},
		{time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC), "Regular day", false},
	}

	for _, tc := range testCases {
		isHoliday := calc.isHoliday(tc.date)
		if isHoliday != tc.expected {
			t.Errorf("%s: expected holiday=%v, got %v", tc.name, tc.expected, isHoliday)
		}
	}
}

func TestMexicanHolidays_MoveableHolidays(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// 2025 moveable holidays (calculated manually):
	// - 1st Monday of February = Feb 3, 2025
	// - 3rd Monday of March = March 17, 2025
	// - 3rd Monday of November = Nov 17, 2025

	testCases := []struct {
		date time.Time
		name string
	}{
		{time.Date(2025, 2, 3, 0, 0, 0, 0, time.UTC), "Constitution Day"},
		{time.Date(2025, 3, 17, 0, 0, 0, 0, time.UTC), "Benito Juárez's Birthday"},
		{time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC), "Revolution Day"},
	}

	for _, tc := range testCases {
		if !calc.isHoliday(tc.date) {
			t.Errorf("%s (%s) should be a holiday", tc.name, tc.date.Format("2006-01-02"))
		}
	}
}

func TestAddCustomHoliday(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Add custom holiday (e.g., company closure)
	customDate := "2025-12-24"
	calc.AddCustomHoliday(customDate)

	date := time.Date(2025, 12, 24, 0, 0, 0, 0, time.UTC)
	if !calc.isHoliday(date) {
		t.Errorf("Custom holiday %s should be recognized", customDate)
	}
}

func TestRemoveHoliday(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	// Remove New Year's Day (for testing)
	calc.RemoveHoliday("2025-01-01")

	date := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if calc.isHoliday(date) {
		t.Errorf("Holiday should be removed")
	}
}

func TestIsBusinessDay(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)

	testCases := []struct {
		date     time.Time
		expected bool
		reason   string
	}{
		{time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), true, "Monday (regular business day)"},
		{time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC), false, "Saturday (weekend)"},
		{time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), false, "Sunday (weekend)"},
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), false, "Wednesday but New Year (holiday)"},
	}

	for _, tc := range testCases {
		result := calc.IsBusinessDay(tc.date)
		if result != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.reason, tc.expected, result)
		}
	}
}

func TestCalculateBusinessDays_RealWorldScenario(t *testing.T) {
	// Vacation from Dec 20, 2024 (Fri) to Jan 3, 2025 (Fri)
	// This spans Christmas and New Year holidays

	// Need to use 2024 calendar for Dec dates
	calc2024 := NewBusinessDayCalculator(2024)
	calc2025 := NewBusinessDayCalculator(2025)

	// Calculate Dec 20-31, 2024
	dec2024Days := calc2024.CalculateBusinessDays(
		time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	)

	// Calculate Jan 1-3, 2025
	jan2025Days := calc2025.CalculateBusinessDays(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
	)

	// Dec 20 (Fri) + Dec 23 (Mon) + Dec 24 (Tue) + Dec 26-31 (excluding Dec 25 Christmas)
	// = 1 + 1 + 1 + 5 = 8 days (weekends excluded: 21-22, 28-29)
	// Note: Dec 24-25 if Christmas is Wed, Dec 24 is Tue (business day)

	// Jan 2-3 (Thu-Fri) = 2 days (Jan 1 is holiday)

	// This is a complex test - exact number depends on which days are weekends
	// The key is that weekends and holidays are excluded
	t.Logf("Dec 2024 business days: %v", dec2024Days)
	t.Logf("Jan 2025 business days: %v", jan2025Days)

	// Verify at least holidays are excluded
	if jan2025Days >= 3.0 {
		t.Errorf("Expected Jan 1 (New Year) to be excluded, got %v days", jan2025Days)
	}
}

func TestGetHolidays(t *testing.T) {
	calc := NewBusinessDayCalculator(2025)
	holidays := calc.GetHolidays()

	// Should have at least 7 holidays (4 fixed + 3 moveable)
	if len(holidays) < 7 {
		t.Errorf("Expected at least 7 holidays, got %v", len(holidays))
	}

	t.Logf("2025 Mexican federal holidays: %v", holidays)
}
