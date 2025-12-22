package query

import (
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// ByBirthMonth filters individuals born in a specific month (1-12).
func (fq *FilterQuery) ByBirthMonth(month int) *FilterQuery {
	if month < 1 || month > 12 {
		return fq // Invalid month, return unchanged
	}
	
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthDate, err := indi.GetBirthDateParsed()
		if err != nil || birthDate == nil || !birthDate.IsValid() {
			return false
		}
		
		// Check if month matches
		if birthDate.Type == gedcom.DateTypeExact {
			return birthDate.Month == month
		}
		
		// For range dates, check if month falls within range
		if birthDate.Type == gedcom.DateTypeBetween || birthDate.Type == gedcom.DateTypeFromTo {
			// Check if month is within the range
			startMonth := birthDate.StartMonth
			endMonth := birthDate.EndMonth
			
			// Handle year wraparound
			if startMonth <= endMonth {
				return month >= startMonth && month <= endMonth
			} else {
				// Range spans year boundary
				return month >= startMonth || month <= endMonth
			}
		}
		
		// For ABOUT, BEFORE, AFTER - use the month if available
		if birthDate.Month > 0 {
			return birthDate.Month == month
		}
		
		return false
	})
}

// ByBirthDay filters individuals born on a specific day of month (1-31).
// Note: This checks the day component only, not the month/year.
func (fq *FilterQuery) ByBirthDay(day int) *FilterQuery {
	if day < 1 || day > 31 {
		return fq // Invalid day, return unchanged
	}
	
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthDate, err := indi.GetBirthDateParsed()
		if err != nil || birthDate == nil || !birthDate.IsValid() {
			return false
		}
		
		// Check if day matches
		if birthDate.Type == gedcom.DateTypeExact {
			return birthDate.Day == day
		}
		
		// For range dates, check if day falls within range
		if birthDate.Type == gedcom.DateTypeBetween || birthDate.Type == gedcom.DateTypeFromTo {
			startDay := birthDate.StartDay
			endDay := birthDate.EndDay
			
			if startDay <= endDay {
				return day >= startDay && day <= endDay
			} else {
				// Range spans month boundary
				return day >= startDay || day <= endDay
			}
		}
		
		// For ABOUT, BEFORE, AFTER - use the day if available
		if birthDate.Day > 0 {
			return birthDate.Day == day
		}
		
		return false
	})
}

// ByBirthMonthAndDay filters individuals born on a specific month and day.
func (fq *FilterQuery) ByBirthMonthAndDay(month int, day int) *FilterQuery {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return fq // Invalid, return unchanged
	}
	
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthDate, err := indi.GetBirthDateParsed()
		if err != nil || birthDate == nil || !birthDate.IsValid() {
			return false
		}
		
		// For exact dates, check both month and day
		if birthDate.Type == gedcom.DateTypeExact {
			return birthDate.Month == month && birthDate.Day == day
		}
		
		// For range dates, check if month/day falls within range
		if birthDate.Type == gedcom.DateTypeBetween || birthDate.Type == gedcom.DateTypeFromTo {
			startTime := time.Date(birthDate.StartYear, time.Month(birthDate.StartMonth), birthDate.StartDay, 0, 0, 0, 0, time.UTC)
			endTime := time.Date(birthDate.EndYear, time.Month(birthDate.EndMonth), birthDate.EndDay, 23, 59, 59, 999999999, time.UTC)
			
			// Check all years in the range
			for year := birthDate.StartYear; year <= birthDate.EndYear; year++ {
				targetTime := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
				if !targetTime.Before(startTime) && !targetTime.After(endTime) {
					return true
				}
			}
			return false
		}
		
		// For ABOUT, BEFORE, AFTER - check month and day if available
		if birthDate.Month > 0 && birthDate.Day > 0 {
			return birthDate.Month == month && birthDate.Day == day
		}
		
		return false
	})
}

// ByBirthDateRange filters individuals born within a date range.
// This is an alias for ByBirthDate for consistency.
func (fq *FilterQuery) ByBirthDateRange(start, end time.Time) *FilterQuery {
	return fq.ByBirthDate(start, end)
}

