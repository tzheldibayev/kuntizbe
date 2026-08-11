package workdays

import (
	"testing"
	"time"
)

func loadTestCalendars(t *testing.T) *Calendars {
	t.Helper()
	cals, err := LoadCalendars("../../data/workdays")
	if err != nil {
		t.Fatalf("LoadCalendars: %v", err)
	}
	return cals
}

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

// TestNextWorkingDay покрывает четыре контрольных случая EVENTS.md: два
// обычных выходных и один праздничный перенос (Наурыз), который ловит
// ошибку, пропускаемую проверкой «не суббота и не воскресенье».
func TestNextWorkingDay(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"выходной, 15 февраля (воскресенье)", "2026-02-15", "2026-02-16"},
		{"праздник, 25 марта (перенос Наурыза)", "2026-03-25", "2026-03-26"},
		{"выходной, 15 августа (суббота)", "2026-08-15", "2026-08-17"},
		{"выходной, 15 ноября (воскресенье)", "2026-11-15", "2026-11-16"},
		{"уже рабочий день не сдвигается", "2026-02-25", "2026-02-25"},
	}

	cals := loadTestCalendars(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cals.NextWorkingDay(date(tc.in))
			if err != nil {
				t.Fatalf("NextWorkingDay(%s): %v", tc.in, err)
			}
			want := date(tc.want)
			if !got.Equal(want) {
				t.Errorf("NextWorkingDay(%s) = %s, want %s", tc.in, got.Format("2006-01-02"), want.Format("2006-01-02"))
			}
		})
	}
}

// TestNextWorkingDayAllHolidaysAndTransfers проверяет каждую запись
// data/workdays/2026.yaml: даты из holidays и transfers.to не должны
// оставаться результатом NextWorkingDay.
func TestNextWorkingDayAllHolidaysAndTransfers(t *testing.T) {
	cals := loadTestCalendars(t)
	cal := cals.byYear[2026]

	for _, h := range cal.Holidays {
		got, err := cals.NextWorkingDay(h.Date)
		if err != nil {
			t.Fatalf("NextWorkingDay(%s, %s): %v", h.Date.Format("2006-01-02"), h.Name, err)
		}
		if got.Equal(h.Date) {
			t.Errorf("праздник %s (%s) не сдвинут", h.Date.Format("2006-01-02"), h.Name)
		}
	}

	for _, tr := range cal.Transfers {
		if tr.Kind != "rest" {
			continue
		}
		got, err := cals.NextWorkingDay(tr.To)
		if err != nil {
			t.Fatalf("NextWorkingDay(%s): %v", tr.To.Format("2006-01-02"), err)
		}
		if got.Equal(tr.To) {
			t.Errorf("перенесённый выходной %s не сдвинут", tr.To.Format("2006-01-02"))
		}
	}
}

// TestNextWorkingDayMissingYear — инвариант I6: диапазон вне покрытия
// календаря завершается ошибкой, а не подстановкой.
func TestNextWorkingDayMissingYear(t *testing.T) {
	cals := loadTestCalendars(t)
	if _, err := cals.NextWorkingDay(date("2030-01-01")); err == nil {
		t.Fatal("ожидалась ошибка для года без производственного календаря")
	}
}
