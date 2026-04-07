package task

import (
	"fmt"
	"sort"
	"time"
)

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceTypeDaily, RecurrenceTypeMonthly, RecurrenceTypeSpecificDates, RecurrenceTypeOddDays, RecurrenceTypeEvenDays:
		return true
	default:
		return false
	}
}

func normalizeDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func (r *RecurrenceInput) NormalizedDates() ([]time.Time, error) {
	if r == nil {
		return nil, nil
	}

	if !r.Type.Valid() {
		return nil, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch r.Type {
	case RecurrenceTypeSpecificDates:
		return r.normalizeSpecificDates()
	case RecurrenceTypeDaily:
		if r.EveryNDays <= 0 {
			return nil, fmt.Errorf("%w: every_n_days must be positive", ErrInvalidInput)
		}
	case RecurrenceTypeMonthly:
		if r.DayOfMonth < 1 || r.DayOfMonth > 30 {
			return nil, fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}
	}

	start, end, err := r.normalizedRange()
	if err != nil {
		return nil, err
	}

	var dates []time.Time

	for current := start; !current.After(end); current.AddDate(0, 0, 1) {
		switch r.Type {
		case RecurrenceTypeDaily:
			diffDays := int(current.Sub(start).Hours() / 24)
			if diffDays%r.EveryNDays == 0 {
				dates = append(dates, current)
			}
		case RecurrenceTypeOddDays:
			if current.Day() == r.DayOfMonth {
				dates = append(dates, current)
			}
		case RecurrenceTypeEvenDays:
			if current.Day()%2 == 0 {
				dates = append(dates, current)
			}
		}
	}

	if len(dates) == 0 {
		return nil, fmt.Errorf("%w: recurrence does not create any tasks in selected period", ErrInvalidInput)
	}

	return dates, nil
}

func (r *RecurrenceInput) normalizeSpecificDates() ([]time.Time, error) {
	if len(r.Dates) == 0 {
		return nil, fmt.Errorf("%w: dates are required for specific_dates recurrence", ErrInvalidInput)
	}

	unique := make(map[string]time.Time, len(r.Dates))
	for _, raw := range r.Dates {
		normalized := normalizeDate(raw)
		unique[normalized.Format(time.DateOnly)] = normalized
	}

	dates := make([]time.Time, 0, len(unique))
	for _, date := range unique {
		dates = append(dates, date)
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates, nil
}

func (r *RecurrenceInput) normalizedRange() (time.Time, time.Time, error) {
	if r.StartDate == nil || r.EndDate == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: syart_date and end_date are required", ErrInvalidInput)
	}

	start := normalizeDate(*r.StartDate)
	end := normalizeDate(*r.EndDate)
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: start_date must be less than end_date", ErrInvalidInput)
	}
	return start, end, nil
}
