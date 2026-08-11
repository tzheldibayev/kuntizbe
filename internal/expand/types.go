package expand

import (
	"time"

	"github.com/tzheldibayev/kuntizbe/internal/rules"
)

// Occurrence — конкретное вхождение правила: один платёж или один срок
// сдачи за один период. DTEnd не включается в интервал (инвариант I3).
type Occurrence struct {
	UID       string
	Rule      rules.Rule
	PeriodKey string
	DTStart   time.Time
	DTEnd     time.Time
}
