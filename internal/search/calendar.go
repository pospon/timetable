package search

import (
	"time"
	"timetable/internal/gtfs"
)

func ActiveServices(calendars []gtfs.Calendar, calendarDates []gtfs.CalendarDate, date time.Time) map[string]bool {
	dateStr := date.Format("20060102")
	weekday := date.Weekday()
	active := make(map[string]bool)

	for _, cal := range calendars {
		if dateStr < cal.StartDate || dateStr > cal.EndDate {
			continue
		}
		if matchesWeekday(cal, weekday) {
			active[cal.ServiceID] = true
		}
	}

	for _, cd := range calendarDates {
		if cd.Date != dateStr {
			continue
		}
		switch cd.ExceptionType {
		case 1:
			active[cd.ServiceID] = true
		case 2:
			delete(active, cd.ServiceID)
		}
	}

	return active
}

func matchesWeekday(cal gtfs.Calendar, wd time.Weekday) bool {
	switch wd {
	case time.Monday:
		return cal.Monday
	case time.Tuesday:
		return cal.Tuesday
	case time.Wednesday:
		return cal.Wednesday
	case time.Thursday:
		return cal.Thursday
	case time.Friday:
		return cal.Friday
	case time.Saturday:
		return cal.Saturday
	case time.Sunday:
		return cal.Sunday
	}
	return false
}
