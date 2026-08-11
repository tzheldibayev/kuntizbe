package rules

import "time"

// Due — смещение расчётной даты от конца отчётного периода (SPEC.md §4).
// Используется и для due, и для window_opens — обе величины одной формы:
// «N число месяца через M месяцев после конца периода».
type Due struct {
	MonthsAfterPeriodEnd int
	Day                  int
}

// Rule — одно правило из data/rules/*.yaml (SPEC.md §4).
type Rule struct {
	ID          string
	Title       string
	AppliesWhen map[string][]string
	Period      string // month | quarter | half-year | year
	WindowOpens *Due   // nil, если окно не ограничено снизу
	Due         Due
	Shift       string // next-working-day | none
	Source      string
	SourceURL   string
	Updated     time.Time
	Note        string
}
