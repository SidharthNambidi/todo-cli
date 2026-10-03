package todo

import (
	"path/filepath"
	"slices"
	"testing"
)

// newList builds a list with the given titles, all in category "General".
// IDs will be 1, 2, 3, ...
func newList(titles ...string) *List {
	l := &List{NextID: 1}
	for _, title := range titles {
		l.Add(title, "General")
	}
	return l
}

// ids returns the task IDs in order, for easy comparison.
func ids(l *List) []int {
	out := make([]int, 0, len(l.Tasks))
	for _, t := range l.Tasks {
		out = append(out, t.ID)
	}
	return out
}

func TestAdd(t *testing.T) {
	l := &List{NextID: 1}
	a := l.Add("first", "Uni")
	b := l.Add("second", "Home")

	if a.ID != 1 || b.ID != 2 {
		t.Errorf("expected IDs 1 and 2, got %d and %d", a.ID, b.ID)
	}
	if len(l.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(l.Tasks))
	}
	first := l.Tasks[0]
	if first.Title != "first" || first.Category != "Uni" || first.Done {
		t.Errorf("unexpected first task: %+v", first)
	}
	if l.NextID != 3 {
		t.Errorf("expected NextID 3, got %d", l.NextID)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantIDs []int
		wantErr bool
	}{
		{"first", 1, []int{2, 3}, false},
		{"middle", 2, []int{1, 3}, false},
		{"last", 3, []int{1, 2}, false},
		{"missing id", 99, []int{1, 2, 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newList("a", "b", "c")

			err := l.Delete(tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete(%d) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
			if got := ids(l); !slices.Equal(got, tt.wantIDs) {
				t.Errorf("remaining IDs = %v, want %v", got, tt.wantIDs)
			}
		})
	}
}

func TestDeleteDoesNotReuseID(t *testing.T) {
	l := newList("a", "b")

	if err := l.Delete(2); err != nil {
		t.Fatal(err)
	}
	got := l.Add("c", "General")

	if got.ID != 3 {
		t.Errorf("expected new task to get ID 3, got %d", got.ID)
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		title     string
		category  string
		wantTitle string
		wantCat   string
		wantErr   bool
	}{
		{"title only", 1, "new", "", "new", "General", false},
		{"category only", 1, "", "Uni", "a", "Uni", false},
		{"both", 1, "new", "Uni", "new", "Uni", false},
		{"empty values leave task unchanged", 1, "", "", "a", "General", false},
		{"missing id", 99, "x", "y", "a", "General", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newList("a") // one task: ID 1, title "a", category "General"

			err := l.Update(tt.id, tt.title, tt.category)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Update error = %v, wantErr %v", err, tt.wantErr)
			}
			got := l.Tasks[0]
			if got.Title != tt.wantTitle || got.Category != tt.wantCat {
				t.Errorf("task = %+v, want title %q category %q", got, tt.wantTitle, tt.wantCat)
			}
		})
	}
}

func TestMarkDone(t *testing.T) {
	l := newList("a", "b")

	if err := l.MarkDone(2); err != nil {
		t.Fatal(err)
	}
	if l.Tasks[0].Done {
		t.Error("task 1 should not be done")
	}
	if !l.Tasks[1].Done {
		t.Error("task 2 should be done")
	}
	if err := l.MarkDone(99); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	want := newList("a", "b")
	want.Tasks[0].Done = true

	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.NextID != want.NextID {
		t.Errorf("NextID = %d, want %d", got.NextID, want.NextID)
	}
	if len(got.Tasks) != len(want.Tasks) {
		t.Fatalf("got %d tasks, want %d", len(got.Tasks), len(want.Tasks))
	}
	for i := range want.Tasks {
		if got.Tasks[i] != want.Tasks[i] {
			t.Errorf("task %d = %+v, want %+v", i, got.Tasks[i], want.Tasks[i])
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")

	if _, err := Load(path); err == nil {
		t.Error("expected error when file is missing")
	}
}

func TestInit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")

	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	l, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if l.NextID != 1 || len(l.Tasks) != 0 {
		t.Errorf("expected empty list with NextID 1, got %+v", l)
	}

	if err := Init(path); err == nil {
		t.Error("second Init should fail because the file already exists")
	}
}
