package workdays

import "time"

// Holiday — нерабочий день из производственного календаря.
type Holiday struct {
	Date time.Time
	Name string
}

// Transfer — перенос выходного дня с одной даты на другую (SPEC.md §5).
// From перестаёт быть выходным, To становится выходным вместо него.
type Transfer struct {
	From time.Time
	To   time.Time
	Kind string
}

// Calendar — производственный календарь одного года.
type Calendar struct {
	Year      int
	Holidays  []Holiday
	Transfers []Transfer
}

// Calendars — производственные календари, загруженные по годам.
type Calendars struct {
	byYear map[int]Calendar
}
