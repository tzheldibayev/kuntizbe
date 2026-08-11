package workdays

import (
	"fmt"
	"time"
)

// NextWorkingDay сдвигает d вперёд, пока день — суббота, воскресенье,
// праздник или перенесённый выходной (SPEC.md §5). Возвращает ошибку,
// если сдвиг выходит за пределы загруженных календарей (инвариант I6):
// недостающий год не достраивается предположениями.
func (c *Calendars) NextWorkingDay(d time.Time) (time.Time, error) {
	d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	for {
		if !c.HasYear(d.Year()) {
			return time.Time{}, fmt.Errorf("workdays: нет производственного календаря на %d год (дата %s)", d.Year(), d.Format("2006-01-02"))
		}
		cal := c.byYear[d.Year()]
		if isWeekend(d) || cal.isHoliday(d) || cal.isTransferredRest(d) {
			d = d.AddDate(0, 0, 1)
			continue
		}
		return d, nil
	}
}

func isWeekend(d time.Time) bool {
	wd := d.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

func (c Calendar) isHoliday(d time.Time) bool {
	for _, h := range c.Holidays {
		if sameDate(h.Date, d) {
			return true
		}
	}
	return false
}

func (c Calendar) isTransferredRest(d time.Time) bool {
	for _, t := range c.Transfers {
		if t.Kind == "rest" && sameDate(t.To, d) {
			return true
		}
	}
	return false
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
