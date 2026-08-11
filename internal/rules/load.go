package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type yamlDue struct {
	MonthsAfterPeriodEnd int `yaml:"months_after_period_end"`
	Day                  int `yaml:"day"`
}

type yamlRule struct {
	ID          string              `yaml:"id"`
	Title       string              `yaml:"title"`
	AppliesWhen map[string][]string `yaml:"applies_when"`
	Period      string              `yaml:"period"`
	WindowOpens *yamlDue            `yaml:"window_opens"`
	Due         yamlDue             `yaml:"due"`
	Shift       string              `yaml:"shift"`
	Source      string              `yaml:"source"`
	SourceURL   string              `yaml:"source_url"`
	Updated     string              `yaml:"updated"`
	Note        string              `yaml:"note"`
}

// Load читает и валидирует все правила из dir (SPEC.md §4, §8).
func Load(dir string) ([]Rule, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("rules: чтение %s: %w", dir, err)
	}

	var out []Rule
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("rules: чтение %s: %w", path, err)
		}
		var yrs []yamlRule
		if err := yaml.Unmarshal(data, &yrs); err != nil {
			return nil, fmt.Errorf("rules: разбор %s: %w", path, err)
		}
		for _, yr := range yrs {
			r, err := fromYAML(yr)
			if err != nil {
				return nil, fmt.Errorf("rules: %s: %w", path, err)
			}
			out = append(out, r)
		}
	}

	if err := Validate(out); err != nil {
		return nil, err
	}
	return out, nil
}

func fromYAML(yr yamlRule) (Rule, error) {
	updated, err := time.Parse("2006-01-02", yr.Updated)
	if err != nil {
		return Rule{}, fmt.Errorf("правило %q: некорректная дата updated %q: %w", yr.ID, yr.Updated, err)
	}

	r := Rule{
		ID:          yr.ID,
		Title:       yr.Title,
		AppliesWhen: yr.AppliesWhen,
		Period:      yr.Period,
		Due:         Due(yr.Due),
		Shift:       yr.Shift,
		Source:      yr.Source,
		SourceURL:   yr.SourceURL,
		Updated:     updated,
		Note:        yr.Note,
	}
	if yr.WindowOpens != nil {
		wo := Due(*yr.WindowOpens)
		r.WindowOpens = &wo
	}
	return r, nil
}
