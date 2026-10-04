package cmd

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"todo-cli/todo"
)

func runUpdate(path string, args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.Int("id", 0, "task id (prompts if omitted)")
	task := fs.String("task", "", "new description")
	cat := fs.String("cat", "", "new category")
	prio := fs.String("priority", "", "new priority: low, medium or high")
	due := fs.String("due", "", "new due date, or \"none\" to clear it")
	fs.Parse(args)

	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	now := time.Now()

	// flag mode: at least one field to change was passed
	if set["task"] || set["cat"] || set["priority"] || set["due"] {
		var e todo.Edit
		if set["task"] {
			e.Title = task
		}
		if set["cat"] {
			e.Category = cat
		}
		if set["priority"] {
			p, err := todo.ParsePriority(*prio)
			if err != nil {
				return err
			}
			e.Priority = &p
		}
		if set["due"] {
			if *due == "none" {
				e.ClearDue = true
			} else {
				d, err := todo.ParseDate(*due, now)
				if err != nil {
					return err
				}
				e.Due = &d
			}
		}
		if *id == 0 {
			return errors.New("-id is required")
		}
		return withList(path, func(l *todo.List) error {
			return l.Update(*id, e)
		})
	}

	//prompt mode
	if !isInteractive() {
		return errors.New("nothing to update: pass -task, -cat, -priority or -due")
	}
	p := newPrompter()
	return withList(path, func(l *todo.List) error {
		taskID, err := idOrPrompt(p, l, *id)
		if err != nil {
			return err
		}
		cur, err := l.Get(taskID)
		if err != nil {
			return err
		}
		e, err := promptEdit(p, cur, now)
		if err != nil {
			return err
		}
		if err := l.Update(taskID, e); err != nil {
			return err
		}
		fmt.Printf("updated #%d\n", taskID)
		return nil
	})
}

func promptEdit(p *prompter, cur todo.Task, now time.Time) (todo.Edit, error) {
	var e todo.Edit

	title, err := askParsed(p, "Task", cur.Title, nonEmpty)
	if err != nil {
		return e, err
	}
	if title != cur.Title {
		e.Title = &title
	}

	category, err := p.ask("Category", cur.Category)
	if err != nil {
		return e, err
	}
	if category != cur.Category {
		e.Category = &category
	}

	curPrio := cur.Priority.String() // "-" when unset
	prio, err := askParsed(p, "Priority (low/medium/high)", curPrio,
		func(s string) (todo.Priority, error) {
			if s == curPrio {
				return cur.Priority, nil // Enter pressed: keep as is
			}
			return todo.ParsePriority(s)
		})
	if err != nil {
		return e, err
	}
	if prio != cur.Priority {
		e.Priority = &prio
	}

	curDue := ""
	if cur.Due != nil {
		curDue = cur.Due.Format(todo.DateLayout)
	}
	dueStr, err := askParsed(p, "Due (date, 'none' to clear)", curDue,
		func(s string) (string, error) {
			if s == curDue || s == "none" {
				return s, nil
			}
			_, err := todo.ParseDate(s, now)
			return s, err
		})
	if err != nil {
		return e, err
	}
	switch {
	case dueStr == curDue:
		//unchanged
	case dueStr == "none":
		e.ClearDue = true
	default:
		d, _ := todo.ParseDate(dueStr, now) // already validated above
		e.Due = &d
	}

	return e, nil
}
