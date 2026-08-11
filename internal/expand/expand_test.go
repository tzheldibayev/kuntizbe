package expand

import (
	"testing"
	"time"

	"github.com/tzheldibayev/kuntizbe/internal/rules"
	"github.com/tzheldibayev/kuntizbe/internal/workdays"
)

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func loadCal(t *testing.T) *workdays.Calendars {
	t.Helper()
	cal, err := workdays.LoadCalendars("../../data/workdays")
	if err != nil {
		t.Fatalf("LoadCalendars: %v", err)
	}
	return cal
}

func loadSamozanyatyRule(t *testing.T) rules.Rule {
	t.Helper()
	rs, err := rules.Load("../../data/rules")
	if err != nil {
		t.Fatalf("rules.Load: %v", err)
	}
	for _, r := range rs {
		if r.ID == "snr-self-employed-social" {
			return r
		}
	}
	t.Fatal("правило snr-self-employed-social не найдено")
	return rules.Rule{}
}

// TestExpandSamozanyaty2026 сверяет 12 вхождений годового диапазона
// с датами, посчитанными вручную по data/workdays/2026.yaml (см. план
// Этапа 1): каждое смещение объясняется выходным, праздником или
// переносом Наурыза.
func TestExpandSamozanyaty2026(t *testing.T) {
	r := loadSamozanyatyRule(t)
	cal := loadCal(t)

	occs, err := Expand(r, cal, date("2026-01-01"), date("2026-12-31"))
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}

	wantDue := []string{
		"2026-01-26", "2026-02-25", "2026-03-26", "2026-04-27",
		"2026-05-25", "2026-06-25", "2026-07-27", "2026-08-25",
		"2026-09-25", "2026-10-27", "2026-11-25", "2026-12-25",
	}
	if len(occs) != len(wantDue) {
		t.Fatalf("получено %d вхождений, ожидалось %d: %+v", len(occs), len(wantDue), occs)
	}
	for i, occ := range occs {
		wantStart := date(wantDue[i])
		if !occ.DTStart.Equal(wantStart) {
			t.Errorf("occ[%d].DTStart = %s, want %s", i, occ.DTStart.Format("2006-01-02"), wantDue[i])
		}
		wantEnd := wantStart.AddDate(0, 0, 1)
		if !occ.DTEnd.Equal(wantEnd) {
			t.Errorf("occ[%d].DTEnd = %s, want %s (I3: DTEND не включается в интервал)", i, occ.DTEnd.Format("2006-01-02"), wantEnd.Format("2006-01-02"))
		}
	}
}

// TestExpandUIDStable — инвариант I1: UID не зависит от note/title.
func TestExpandUIDStable(t *testing.T) {
	r := loadSamozanyatyRule(t)
	cal := loadCal(t)

	before, err := Expand(r, cal, date("2026-01-01"), date("2026-12-31"))
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}

	r.Note = "переписанное пояснение, не влияющее на UID"
	r.Title = "переписанный заголовок"
	after, err := Expand(r, cal, date("2026-01-01"), date("2026-12-31"))
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}

	if len(before) != len(after) {
		t.Fatalf("разное число вхождений: %d vs %d", len(before), len(after))
	}
	for i := range before {
		if before[i].UID != after[i].UID {
			t.Errorf("UID изменился при правке note/title: %q -> %q", before[i].UID, after[i].UID)
		}
	}
}

func TestExpandUIDFormat(t *testing.T) {
	r := loadSamozanyatyRule(t)
	cal := loadCal(t)
	occs, err := Expand(r, cal, date("2026-01-01"), date("2026-01-31"))
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if len(occs) != 1 {
		t.Fatalf("ожидалось 1 вхождение в январе, получено %d", len(occs))
	}
	want := "snr-self-employed-social-2025-12@nalog-cal.kz"
	if occs[0].UID != want {
		t.Errorf("UID = %q, want %q", occs[0].UID, want)
	}
}
