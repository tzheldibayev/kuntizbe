package ics

import "fmt"

// monthNames — именительный/винительный падеж («за январь», не «за
// января»): «за» с названием месяца управляет винительным падежом,
// который у названий месяцев мужского рода неодушевлённый и совпадает
// с именительным.
var monthNames = [...]string{
	"январь", "февраль", "март", "апрель", "май", "июнь",
	"июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь",
}

// periodLabel строит человекочитаемое «за <период>» из ключа периода
// (SPEC.md §4: 2026-01, 2026Q1, 2026H1, 2026).
func periodLabel(period, periodKey string) string {
	switch period {
	case "month":
		var y, m int
		if _, err := fmt.Sscanf(periodKey, "%d-%d", &y, &m); err == nil && m >= 1 && m <= 12 {
			return fmt.Sprintf("за %s %d", monthNames[m-1], y)
		}
	case "quarter":
		var y, q int
		if _, err := fmt.Sscanf(periodKey, "%dQ%d", &y, &q); err == nil {
			return fmt.Sprintf("за %d квартал %d", q, y)
		}
	case "half-year":
		var y, h int
		if _, err := fmt.Sscanf(periodKey, "%dH%d", &y, &h); err == nil {
			return fmt.Sprintf("за %d полугодие %d", h, y)
		}
	case "year":
		return fmt.Sprintf("за %s год", periodKey)
	}
	return "за " + periodKey
}
