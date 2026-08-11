package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "перезаписать golden-файлы")

const (
	testFrom = "2026-01-01"
	testTo   = "2026-12-31"
)

func generate(t *testing.T, outDir, file string) string {
	t.Helper()
	if err := run("../../data/rules", "../../data/workdays", outDir, testFrom, testTo); err != nil {
		t.Fatalf("run: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, file))
	if err != nil {
		t.Fatalf("чтение результата: %v", err)
	}
	return string(got)
}

// TestGolden сравнивает полный фид за 2026 год с эталоном для каждого
// фида. Любое изменение вывода — библиотеки, формул раскрытия, рендера,
// нового правила — становится видимым в diff (SPEC.md §8).
func TestGolden(t *testing.T) {
	for _, file := range []string{"samozanyaty.ics", "ip-uproshchenka.ics"} {
		t.Run(file, func(t *testing.T) {
			got := generate(t, t.TempDir(), file)
			goldenPath := "../../testdata/golden/" + file

			if *update {
				if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
					t.Fatalf("запись golden-файла: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("чтение golden-файла: %v (запустите go test ./cmd/gen -update)", err)
			}
			if got != string(want) {
				t.Errorf("вывод разошёлся с %s (запустите go test ./cmd/gen -update, если расхождение ожидаемо)\n--- got ---\n%s", goldenPath, got)
			}
		})
	}
}

// TestDeterministic — инвариант I2: два прогона на неизменных входных
// данных дают идентичные байты.
func TestDeterministic(t *testing.T) {
	a := generate(t, t.TempDir(), "samozanyaty.ics")
	b := generate(t, t.TempDir(), "samozanyaty.ics")
	if a != b {
		t.Error("два прогона дали разный вывод — нарушена побайтовая стабильность (I2)")
	}
}

// TestMissingCalendarYearFails — инвариант I6: диапазон вне покрытия
// производственного календаря завершается ошибкой, а не пустым файлом.
func TestMissingCalendarYearFails(t *testing.T) {
	err := run("../../data/rules", "../../data/workdays", t.TempDir(), "2030-01-01", "2030-12-31")
	if err == nil {
		t.Fatal("ожидалась ошибка для диапазона без производственного календаря")
	}
}

// TestFromAfterToFails — --to раньше --from недопустимо.
func TestFromAfterToFails(t *testing.T) {
	err := run("../../data/rules", "../../data/workdays", t.TempDir(), "2026-06-01", "2026-01-01")
	if err == nil {
		t.Fatal("ожидалась ошибка для --to раньше --from")
	}
}
