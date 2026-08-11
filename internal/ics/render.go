package ics

import (
	"fmt"
	"sort"
	"strings"

	gical "github.com/arran4/golang-ical"

	"github.com/tzheldibayev/kuntizbe/internal/expand"
	"github.com/tzheldibayev/kuntizbe/internal/rules"
)

// Header — параметры заголовка фида (SPEC.md §7). Имя файла, на котором
// он публикуется, сюда не входит — это забота вызывающего кода.
type Header struct {
	CalName string
	CalDesc string
}

const prodID = "-//nalog-cal.kz//tax-calendar//RU"

// Render рендерит вхождения в текст VCALENDAR. Сортирует по DTSTART, затем
// по UID (инвариант I2) — порядок входного среза не имеет значения.
func Render(header Header, occs []expand.Occurrence) string {
	sorted := make([]expand.Occurrence, len(occs))
	copy(sorted, occs)
	sort.Slice(sorted, func(i, j int) bool {
		if !sorted[i].DTStart.Equal(sorted[j].DTStart) {
			return sorted[i].DTStart.Before(sorted[j].DTStart)
		}
		return sorted[i].UID < sorted[j].UID
	})

	cal := gical.NewCalendar()
	cal.SetVersion("2.0")
	cal.SetProductId(prodID)
	cal.SetCalscale("GREGORIAN")
	cal.SetMethod(gical.MethodPublish)
	cal.SetXWRCalName(header.CalName)
	cal.SetXWRCalDesc(header.CalDesc)
	cal.SetRefreshInterval("P1D")
	cal.SetXPublishedTTL("P1D")

	for _, occ := range sorted {
		addEvent(cal, occ)
	}

	// RFC 5545 требует CRLF; библиотека по умолчанию на не-Windows сборках
	// пишет \n (os_unix.go) — .gitattributes (*.ics -text) существует
	// именно затем, чтобы git не тронул то, что здесь задано явно.
	return cal.Serialize(gical.WithNewLineWindows)
}

func addEvent(cal *gical.Calendar, occ expand.Occurrence) {
	event := cal.AddEvent(occ.UID)
	event.SetAllDayStartAt(occ.DTStart)
	event.SetAllDayEndAt(occ.DTEnd)
	event.SetDtStampTime(occ.Rule.Updated)
	event.SetSummary(summary(occ))
	event.SetDescription(description(occ.Rule))

	alarm7 := event.AddAlarm()
	alarm7.SetAction(gical.ActionDisplay)
	alarm7.SetTrigger("-P7D")
	alarm7.SetDescription(fmt.Sprintf("Через неделю — %s", occ.Rule.Title))

	alarm1 := event.AddAlarm()
	alarm1.SetAction(gical.ActionDisplay)
	alarm1.SetTrigger("-P1D")
	alarm1.SetDescription(fmt.Sprintf("Завтра — %s", occ.Rule.Title))
}

func summary(occ expand.Occurrence) string {
	return fmt.Sprintf("%s (%s)", occ.Rule.Title, periodLabel(occ.Rule.Period, occ.PeriodKey))
}

func description(r rules.Rule) string {
	var b strings.Builder
	b.WriteString(r.Source)
	b.WriteString("\n")
	b.WriteString(r.SourceURL)
	if r.Note != "" {
		b.WriteString("\n\n")
		b.WriteString(strings.TrimSpace(r.Note))
	}
	return b.String()
}
