package rules

import "testing"

func TestLoadSamozanyaty(t *testing.T) {
	rs, err := Load("../../data/rules")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("ожидалось 1 правило, получено %d", len(rs))
	}
	r := rs[0]
	if r.ID != "snr-self-employed-social" {
		t.Errorf("id = %q", r.ID)
	}
	if r.SourceURL == "" {
		t.Error("пустой source_url")
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
