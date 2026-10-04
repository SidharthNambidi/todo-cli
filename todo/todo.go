package todo

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

// ---------- Priority ----------

type Priority int

const (
	PriorityNone Priority = iota // 0: unset, also what old tasks decode to
	PriorityLow
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	}
	return "-"
}

func ParsePriority(s string) (Priority, error) {
	switch strings.ToLower(s) {
	case "low", "l":
		return PriorityLow, nil
	case "medium", "med", "m":
		return PriorityMedium, nil
	case "high", "h":
		return PriorityHigh, nil
	}
	return PriorityNone, fmt.Errorf("invalid priority %q (use low, medium or high)", s)
}

// ---------- Dates ----------

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// ParseDate accepts YYYY-MM-DD, today, tomorrow, or +Nd (N days from today).
// now is passed in so tests don't depend on the real clock.
func ParseDate(s string, now time.Time) (time.Time, error) {
	today := startOfDay(now)
	switch s {
	case "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	}
	if strings.HasPrefix(s, "+") && strings.HasSuffix(s, "d") {
		if n, err := strconv.Atoi(s[1 : len(s)-1]); err == nil && n >= 0 {
			return today.AddDate(0, 0, n), nil
		}
	}
	d, err := time.ParseInLocation(DateLayout, s, now.Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (use YYYY-MM-DD, today, tomorrow or +Nd)", s)
	}
	return d, nil
}

// ---------- Task and List ----------

type Task struct {
	ID       int        `json:"id"`
	Title    string     `json:"title"`
	Category string     `json:"category"`
	Done     bool       `json:"done"`
	Priority Priority   `json:"priority"`
	Due      *time.Time `json:"due,omitempty"`
}

// Overdue reports whether the task is unfinished and its due date is before today.
// A task due today is not overdue.
func (t Task) Overdue(now time.Time) bool {
	return !t.Done && t.Due != nil && t.Due.Before(startOfDay(now))
}

type List struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

func FilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".todo.json"), nil
}

func Init(path string) error {
	if _, err := os.Stat(path); err == nil {
		return errors.New("already initialized")
	}
	return Save(path, &List{NextID: 1})
}

func Load(path string) (*List, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("no todo file found, run `todo init` first")
		}
		return nil, err
	}
	var l List
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func Save(path string, l *List) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Add assigns the next ID to t and appends it.
func (l *List) Add(t Task) Task {
	t.ID = l.NextID
	l.NextID++
	l.Tasks = append(l.Tasks, t)
	return t
}

func (l *List) index(id int) (int, error) {
	for i, t := range l.Tasks {
		if t.ID == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no task with id %d", id)
}

func (l *List) Get(id int) (Task, error) {
	i, err := l.index(id)
	if err != nil {
		return Task{}, err
	}
	return l.Tasks[i], nil
}

// Edit describes changes to a task. A nil field means "leave unchanged".
type Edit struct {
	Title    *string
	Category *string
	Priority *Priority
	Due      *time.Time
	ClearDue bool // remove the due date
}

func (l *List) Update(id int, e Edit) error {
	i, err := l.index(id)
	if err != nil {
		return err
	}
	t := &l.Tasks[i]
	if e.Title != nil {
		t.Title = *e.Title
	}
	if e.Category != nil {
		t.Category = *e.Category
	}
	if e.Priority != nil {
		t.Priority = *e.Priority
	}
	if e.ClearDue {
		t.Due = nil
	} else if e.Due != nil {
		t.Due = e.Due
	}
	return nil
}

func (l *List) MarkDone(id int) error {
	i, err := l.index(id)
	if err != nil {
		return err
	}
	l.Tasks[i].Done = true
	return nil
}

func (l *List) Delete(id int) error {
	i, err := l.index(id)
	if err != nil {
		return err
	}
	l.Tasks = append(l.Tasks[:i], l.Tasks[i+1:]...)
	return nil
}

// ---------- Filtering and sorting ----------

// Filter selects tasks. Zero values mean "don't filter on this".
type Filter struct {
	Done     *bool // nil = any, true = completed, false = pending
	Category string
	Priority Priority
	Overdue  bool
}

func (l *List) Filter(f Filter, now time.Time) []Task {
	var out []Task
	for _, t := range l.Tasks {
		if f.Done != nil && *f.Done != t.Done {
			continue
		}
		if f.Category != "" && t.Category != f.Category {
			continue
		}
		if f.Priority != PriorityNone && t.Priority != f.Priority {
			continue
		}
		if f.Overdue && !t.Overdue(now) {
			continue
		}
		out = append(out, t)
	}
	return out
}

// SortTasks sorts in place by "id", "priority" (high first) or "due" (soonest
// first, tasks with no due date last). The sort is stable, so ties keep ID order.
func SortTasks(tasks []Task, by string) error {
	var less func(a, b Task) int
	switch by {
	case "", "id":
		less = func(a, b Task) int { return cmp.Compare(a.ID, b.ID) }
	case "priority":
		less = func(a, b Task) int { return cmp.Compare(b.Priority, a.Priority) }
	case "due":
		less = func(a, b Task) int {
			switch {
			case a.Due == nil && b.Due == nil:
				return 0
			case a.Due == nil:
				return 1
			case b.Due == nil:
				return -1
			}
			return a.Due.Compare(*b.Due)
		}
	default:
		return fmt.Errorf("unknown sort key %q (use id, priority or due)", by)
	}
	slices.SortStableFunc(tasks, less)
	return nil
}
