package rules

import "testing"

func TestLoadDataRules(t *testing.T) {
	rs, err := Load("../../data/rules")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantIDs := map[string]bool{
		"snr-self-employed-social": true,
		"fno-910-00":               true,
		"fno-910-00-payment":       true,
		"ip-uproshchenka-social":   true,
	}
	if len(rs) != len(wantIDs) {
		t.Fatalf("ожидалось %d правил, получено %d", len(wantIDs), len(rs))
	}
	for _, r := range rs {
		if !wantIDs[r.ID] {
			t.Errorf("неожиданный id %q", r.ID)
		}
		if r.SourceURL == "" {
			t.Errorf("%q: пустой source_url", r.ID)
		}
	}
}

func TestValidateRejectsMissingSourceURL(t *testing.T) {
	r := Rule{
		ID: "x", Title: "t", Source: "норма",
		Period: "month", Shift: "next-working-day",
		Due:         Due{MonthsAfterPeriodEnd: 1, Day: 25},
		AppliesWhen: map[string][]string{"regime": {"snr-self-employed"}},
	}
	if err := Validate([]Rule{r}); err == nil {
		t.Fatal("ожидалась ошибка для правила без source_url")
	}
}

func TestValidateRejectsUnknownRegime(t *testing.T) {
	r := Rule{
		ID: "x", Title: "t", Source: "норма", SourceURL: "https://adilet.zan.kz/rus/docs/x",
		Period: "month", Shift: "next-working-day",
		Due:         Due{MonthsAfterPeriodEnd: 1, Day: 25},
		AppliesWhen: map[string][]string{"regime": {"no-such-regime"}},
	}
	if err := Validate([]Rule{r}); err == nil {
		t.Fatal("ожидалась ошибка для несуществующего режима")
	}
}
