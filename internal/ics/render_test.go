package ics

import (
	"strings"
	"testing"
	"time"

	"github.com/tzheldibayev/kuntizbe/internal/expand"
	"github.com/tzheldibayev/kuntizbe/internal/rules"
)

func mustDate(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func sampleOccurrence() expand.Occurrence {
	r := rules.Rule{
		ID:        "snr-self-employed-social",
		Title:     "Социальные платежи самозанятого — ОПВ, ОПВР, СО, ВОСМС",
		Source:    "Соцкодекс РК, ст. 101-1",
		SourceURL: "https://adilet.zan.kz/rus/docs/K2300000224#z3538",
		Updated:   mustDate("2026-08-11"),
		Note:      "ОПВ, ОПВР и СО — по 1% от дохода.",
	}
	return expand.Occurrence{
		UID:       "snr-self-employed-social-2025-12@nalog-cal.kz",
		Rule:      r,
		PeriodKey: "2025-12",
		DTStart:   mustDate("2026-01-26"),
		DTEnd:     mustDate("2026-01-27"),
	}
}

func TestRenderProducesDateOnlyEvent(t *testing.T) {
	out := Render(Header{CalName: "Тест", CalDesc: "Тестовое описание"}, []expand.Occurrence{sampleOccurrence()})

	mustContain := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"METHOD:PUBLISH",
		"UID:snr-self-employed-social-2025-12@nalog-cal.kz",
		"DTSTART;VALUE=DATE:20260126",
		"DTEND;VALUE=DATE:20260127",
		"DTSTAMP",
		"TRIGGER:-P7D",
		"TRIGGER:-P1D",
		"ACTION:DISPLAY",
		"END:VCALENDAR",
	}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Errorf("вывод не содержит %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "VTIMEZONE") {
		t.Error("вывод не должен содержать VTIMEZONE (инвариант I3)")
	}
}

func TestRenderSortsByDTStartThenUID(t *testing.T) {
	early := sampleOccurrence()
	early.UID = "b-2026-01@nalog-cal.kz"
	early.DTStart = mustDate("2026-02-01")
	early.DTEnd = mustDate("2026-02-02")

	late := sampleOccurrence()
	late.UID = "a-2026-01@nalog-cal.kz"
	late.DTStart = mustDate("2026-01-01")
	late.DTEnd = mustDate("2026-01-02")

	// Передаём в "неправильном" порядке — Render должен отсортировать сам.
	out := Render(Header{CalName: "Тест"}, []expand.Occurrence{early, late})

	posLate := strings.Index(out, late.UID)
	posEarly := strings.Index(out, early.UID)
	if posLate == -1 || posEarly == -1 {
		t.Fatalf("UID не найдены в выводе")
	}
	if posLate > posEarly {
		t.Errorf("события не отсортированы по DTSTART: %s должен идти раньше %s", late.UID, early.UID)
	}
}
