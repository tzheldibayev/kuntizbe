package expand

import (
	"fmt"
	"time"

	"github.com/tzheldibayev/kuntizbe/internal/rules"
	"github.com/tzheldibayev/kuntizbe/internal/workdays"
)

var periodStepMonths = map[string]int{
	"month":     1,
	"quarter":   3,
	"half-year": 6,
	"year":      12,
}

// Expand раскрывает правило на диапазон [from, to] в конкретные вхождения.
//
// Перебираются периоды с запасом в год до from — обязательство за декабрь
// прошлого года может быть уплачено уже в январе запрошенного диапазона
// (EVENTS.md: универсальный ритм «15/25 число второго месяца»), поэтому
// граница диапазона проверяется по датам самого вхождения, а не по началу
// периода, который его породил.
func Expand(r rules.Rule, cal *workdays.Calendars, from, to time.Time) ([]Occurrence, error) {
	stepMonths, ok := periodStepMonths[r.Period]
	if !ok {
		return nil, fmt.Errorf("expand: %q: неизвестный period %q", r.ID, r.Period)
	}

	rangeStart := from
	rangeEnd := to.AddDate(0, 0, 1) // граница исключается, интервал [from, rangeEnd)

	var out []Occurrence
	for y := from.Year() - 1; y <= to.Year(); y++ {
		for m := 1; m <= 12; m += stepMonths {
			periodStart := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
			periodEndMonth := periodStart.AddDate(0, stepMonths-1, 0)

			dueDate := computeDate(periodEndMonth, r.Due)
			// Перенос сдвигает дату максимум на несколько дней вперёд
			// (выходные вокруг праздничного блока). Отсекаем периоды,
			// заведомо не пересекающие диапазон, до вызова shift — иначе
			// пришлось бы держать календарь на каждый год, через который
			// проходит перебор периодов, а не только на годы диапазона.
			// Перенос двигает дату вперёд не дальше длины самого длинного
			// блока подряд идущих нерабочих дней — в РК это блок Наурыза,
			// 21-25 марта, 5 дней (data/workdays/2026.yaml). Берём запас
			// в один день. Периоды за этой границей отсекаются до вызова
			// shift, чтобы не требовать календарь на год, который явно
			// не может пересечься с диапазоном.
			const shiftMargin = 6
			if dueDate.Before(rangeStart.AddDate(0, 0, -shiftMargin)) || dueDate.After(rangeEnd.AddDate(0, 0, shiftMargin)) {
				continue
			}

			shifted, err := shift(cal, dueDate, r.Shift)
			if err != nil {
				return nil, fmt.Errorf("expand: %q: %w", r.ID, err)
			}

			dtStart := shifted
			if r.WindowOpens != nil {
				dtStart = computeDate(periodEndMonth, *r.WindowOpens)
			}
			dtEnd := shifted.AddDate(0, 0, 1)

			if dtStart.Before(rangeEnd) && dtEnd.After(rangeStart) {
				periodKey := formatPeriodKey(r.Period, periodStart)
				out = append(out, Occurrence{
					UID:       fmt.Sprintf("%s-%s@nalog-cal.kz", r.ID, periodKey),
					Rule:      r,
					PeriodKey: periodKey,
					DTStart:   dtStart,
					DTEnd:     dtEnd,
				})
			}
		}
	}
	return out, nil
}

// computeDate — «day число через monthsAfterPeriodEnd месяцев после
// periodEndMonth» (SPEC.md §4). periodEndMonth — первое число последнего
// месяца периода.
func computeDate(periodEndMonth time.Time, d rules.Due) time.Time {
	target := periodEndMonth.AddDate(0, d.MonthsAfterPeriodEnd, 0)
	return time.Date(target.Year(), target.Month(), d.Day, 0, 0, 0, 0, time.UTC)
}

func shift(cal *workdays.Calendars, d time.Time, kind string) (time.Time, error) {
	switch kind {
	case "next-working-day":
		return cal.NextWorkingDay(d)
	case "none":
		return d, nil
	default:
		return time.Time{}, fmt.Errorf("неизвестный shift %q", kind)
	}
}

func formatPeriodKey(period string, periodStart time.Time) string {
	switch period {
	case "month":
		return periodStart.Format("2006-01")
	case "quarter":
		q := (int(periodStart.Month())-1)/3 + 1
		return fmt.Sprintf("%dQ%d", periodStart.Year(), q)
	case "half-year":
		h := (int(periodStart.Month())-1)/6 + 1
		return fmt.Sprintf("%dH%d", periodStart.Year(), h)
	default: // year
		return fmt.Sprintf("%d", periodStart.Year())
	}
}
