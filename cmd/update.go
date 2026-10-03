package cmd

import (
	"flag"
	"time"

	"todo-cli/todo"
)

func runUpdate(path string, args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.Int("id", 0, "task id")
	task := fs.String("task", "", "new description")
	cat := fs.String("cat", "", "new category")
	prio := fs.String("priority", "", "new priority: low, medium or high")
	due := fs.String("due", "", "new due date, or \"none\" to clear it")
	fs.Parse(args)

	// Visit only walks flags that were actually passed on the command line.
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

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
			d, err := todo.ParseDate(*due, time.Now())
			if err != nil {
				return err
			}
			e.Due = &d
		}
	}

	return withList(path, func(l *todo.List) error {
		return l.Update(*id, e)
	})
}
