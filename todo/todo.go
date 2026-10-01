package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Done     bool   `json:"done"`
}

type List struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

// FilePath returns ~/.todo.json
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

func (l *List) Add(title, category string) Task {
	t := Task{ID: l.NextID, Title: title, Category: category}
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

func (l *List) Update(id int, title, category string) error {
	i, err := l.index(id)
	if err != nil {
		return err
	}
	if title != "" {
		l.Tasks[i].Title = title
	}
	if category != "" {
		l.Tasks[i].Category = category
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
