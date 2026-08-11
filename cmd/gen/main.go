// Command gen раскрывает правила из data/rules на диапазон дат и пишет
// готовые .ics-фиды в выходной каталог (SPEC.md §2).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tzheldibayev/kuntizbe/internal/expand"
	"github.com/tzheldibayev/kuntizbe/internal/ics"
	"github.com/tzheldibayev/kuntizbe/internal/rules"
	"github.com/tzheldibayev/kuntizbe/internal/workdays"
)

// feed — один выходной файл (SPEC.md §7) и позиция по осям TAXONOMY.md §2,
// на которую он подписан. Список пока захардкожен: до Этапа 4 общий код
// не выносится (PLAN.md), а фидов — один. Когда их станет больше, это
// содержимое переезжает в data-файл без изменения остальной цепочки.
type feed struct {
	file    string
	calName string
	calDesc string
	axes    map[string]string
}

var feeds = []feed{
	{
		file:    "samozanyaty.ics",
		calName: "Налоги — самозанятые",
		calDesc: "Сроки налоговой отчётности РК. Источник указан в каждом событии.",
		axes:    map[string]string{"regime": "snr-self-employed"},
	},
	{
		file:    "ip-uproshchenka.ics",
		calName: "Налоги — ИП на упрощёнке",
		calDesc: "Сроки налоговой отчётности РК. Источник указан в каждом событии.",
		axes: map[string]string{
			"form":      "ip",
			"regime":    "snr-simplified",
			"vat":       "vat-none",
			"employees": "no-employees",
		},
	},
}

func main() {
	rulesDir := flag.String("rules", "data/rules", "каталог правил")
	workdaysDir := flag.String("workdays", "data/workdays", "каталог производственного календаря")
	outDir := flag.String("out", "docs", "каталог для .ics")
	fromFlag := flag.String("from", "", "начало диапазона, YYYY-MM-DD")
	toFlag := flag.String("to", "", "конец диапазона, YYYY-MM-DD")
	flag.Parse()

	if err := run(*rulesDir, *workdaysDir, *outDir, *fromFlag, *toFlag); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(rulesDir, workdaysDir, outDir, fromFlag, toFlag string) error {
	from, err := time.Parse("2006-01-02", fromFlag)
	if err != nil {
		return fmt.Errorf("--from: %w", err)
	}
	to, err := time.Parse("2006-01-02", toFlag)
	if err != nil {
		return fmt.Errorf("--to: %w", err)
	}
	if !to.After(from) {
		return fmt.Errorf("--to (%s) должен быть позже --from (%s)", toFlag, fromFlag)
	}

	allRules, err := rules.Load(rulesDir)
	if err != nil {
		return err
	}
	cal, err := workdays.LoadCalendars(workdaysDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("создание %s: %w", outDir, err)
	}

	for _, f := range feeds {
		var occs []expand.Occurrence
		for _, r := range allRules {
			if !appliesToFeed(r, f) {
				continue
			}
			o, err := expand.Expand(r, cal, from, to)
			if err != nil {
				return err
			}
			occs = append(occs, o...)
		}

		out := ics.Render(ics.Header{CalName: f.calName, CalDesc: f.calDesc}, occs)
		path := filepath.Join(outDir, f.file)
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return fmt.Errorf("запись %s: %w", path, err)
		}
		fmt.Fprintf(os.Stdout, "gen: %s — %d событий\n", path, len(occs))
	}
	return nil
}

// appliesToFeed разрешает applies_when правила против позиции фида по
// осям (TAXONOMY.md §2) — при каждом запуске заново, а не один раз при
// подписке: появившееся новое правило подхватывается автоматически.
func appliesToFeed(r rules.Rule, f feed) bool {
	for axis, allowed := range r.AppliesWhen {
		feedValue, ok := f.axes[axis]
		if !ok {
			return false
		}
		if !contains(allowed, feedValue) {
			return false
		}
	}
	return true
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
