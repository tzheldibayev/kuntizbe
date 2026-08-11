package rules

import "fmt"

// Оси и значения из TAXONOMY.md §1. Список сознательно неполный — только
// то, что уже используется правилами; расширяется по мере кодирования
// новых блоков EVENTS.md.
var knownAxes = map[string]map[string]bool{
	"form":      {"individual": true, "self-employed": true, "ip": true, "too": true, "kfh": true},
	"regime":    {"our": true, "snr-self-employed": true, "snr-simplified": true, "snr-farm": true},
	"vat":       {"vat-none": true, "vat-registered": true},
	"employees": {"no-employees": true, "has-employees": true},
	"flags":     {"property": true, "vehicle": true, "land": true, "foreign-trade": true, "non-resident-payments": true},
}

var knownPeriods = map[string]bool{"month": true, "quarter": true, "half-year": true, "year": true}
var knownShifts = map[string]bool{"next-working-day": true, "none": true}

// Validate проверяет обязательные поля и ссылки правил (SPEC.md §4, §8).
// Любое нарушение — ошибка загрузки, а не предупреждение: правило без
// source_url или со ссылкой на несуществующий режим не должно попасть
// в раскрытие и рендер.
func Validate(rs []Rule) error {
	seen := make(map[string]bool, len(rs))
	for _, r := range rs {
		if err := validateOne(r); err != nil {
			return err
		}
		if seen[r.ID] {
			return fmt.Errorf("rules: повторяющийся id %q", r.ID)
		}
		seen[r.ID] = true
	}
	return nil
}

func validateOne(r Rule) error {
	if r.ID == "" {
		return fmt.Errorf("rules: правило без id")
	}
	if r.Title == "" {
		return fmt.Errorf("rules: %q: пустой title", r.ID)
	}
	if r.Source == "" {
		return fmt.Errorf("rules: %q: обязателен source", r.ID)
	}
	if r.SourceURL == "" {
		return fmt.Errorf("rules: %q: обязателен source_url", r.ID)
	}
	if !isAllowedSourceHost(r.SourceURL) {
		return fmt.Errorf("rules: %q: source_url должен вести на adilet.zan.kz или kgd.gov.kz, получено %q", r.ID, r.SourceURL)
	}
	if !knownPeriods[r.Period] {
		return fmt.Errorf("rules: %q: неизвестный period %q", r.ID, r.Period)
	}
	if !knownShifts[r.Shift] {
		return fmt.Errorf("rules: %q: неизвестный shift %q", r.ID, r.Shift)
	}
	if err := validateDue(r.ID, "due", r.Due); err != nil {
		return err
	}
	if r.WindowOpens != nil {
		if err := validateDue(r.ID, "window_opens", *r.WindowOpens); err != nil {
			return err
		}
	}
	if len(r.AppliesWhen) == 0 {
		return fmt.Errorf("rules: %q: пустой applies_when", r.ID)
	}
	for axis, values := range r.AppliesWhen {
		known, axisExists := knownAxes[axis]
		if !axisExists {
			return fmt.Errorf("rules: %q: неизвестная ось %q в applies_when", r.ID, axis)
		}
		for _, v := range values {
			if !known[v] {
				return fmt.Errorf("rules: %q: applies_when.%s ссылается на несуществующее значение %q", r.ID, axis, v)
			}
		}
	}
	return nil
}

func validateDue(ruleID, field string, d Due) error {
	if d.MonthsAfterPeriodEnd < 0 {
		return fmt.Errorf("rules: %q: %s.months_after_period_end отрицателен", ruleID, field)
	}
	if d.Day < 1 || d.Day > 31 {
		return fmt.Errorf("rules: %q: %s.day вне диапазона 1–31", ruleID, field)
	}
	return nil
}

func isAllowedSourceHost(url string) bool {
	const (
		adilet = "https://adilet.zan.kz/"
		kgd    = "https://kgd.gov.kz/"
	)
	return len(url) >= len(adilet) && url[:len(adilet)] == adilet ||
		len(url) >= len(kgd) && url[:len(kgd)] == kgd
}
