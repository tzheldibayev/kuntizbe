package workdays

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type yamlHoliday struct {
	Date time.Time `yaml:"date"`
	Name string    `yaml:"name"`
}

type yamlTransfer struct {
	From time.Time `yaml:"from"`
	To   time.Time `yaml:"to"`
	Kind string    `yaml:"kind"`
}

type yamlCalendar struct {
	Year      int            `yaml:"year"`
	Holidays  []yamlHoliday  `yaml:"holidays"`
	Transfers []yamlTransfer `yaml:"transfers"`
}

// LoadCalendars читает все производственные календари *.yaml из каталога.
func LoadCalendars(dir string) (*Calendars, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("workdays: чтение %s: %w", dir, err)
	}

	byYear := make(map[int]Calendar)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("workdays: чтение %s: %w", path, err)
		}
		var yc yamlCalendar
		if err := yaml.Unmarshal(data, &yc); err != nil {
			return nil, fmt.Errorf("workdays: разбор %s: %w", path, err)
		}
		if yc.Year == 0 {
			return nil, fmt.Errorf("workdays: %s: не указан year", path)
		}
		if _, exists := byYear[yc.Year]; exists {
			return nil, fmt.Errorf("workdays: год %d определён более чем в одном файле", yc.Year)
		}

		cal := Calendar{Year: yc.Year}
		for _, h := range yc.Holidays {
			cal.Holidays = append(cal.Holidays, Holiday{Date: h.Date, Name: h.Name})
		}
		for _, t := range yc.Transfers {
			cal.Transfers = append(cal.Transfers, Transfer{From: t.From, To: t.To, Kind: t.Kind})
		}
		byYear[yc.Year] = cal
	}

	return &Calendars{byYear: byYear}, nil
}

// Years возвращает загруженные годы по возрастанию.
func (c *Calendars) Years() []int {
	years := make([]int, 0, len(c.byYear))
	for y := range c.byYear {
		years = append(years, y)
	}
	sort.Ints(years)
	return years
}

// HasYear сообщает, загружен ли календарь на год d.
func (c *Calendars) HasYear(year int) bool {
	_, ok := c.byYear[year]
	return ok
}
