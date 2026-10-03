package cmd

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"todo-cli/todo"
)

func runAdd(path string, args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	task := fs.String("task", "", "task description")
	cat := fs.String("cat", "General", "category")
	prio := fs.String("priority", "medium", "low, medium or high")
	due := fs.String("due", "", "due date: YYYY-MM-DD, today, tomorrow or +Nd")
	fs.Parse(args)

	if *task == "" {
		return errors.New("-task is required")
	}
	p, err := todo.ParsePriority(*prio)
	if err != nil {
		return err
	}

	t := todo.Task{Title: *task, Category: *cat, Priority: p}
	if *due != "" {
		d, err := todo.ParseDate(*due, time.Now())
		if err != nil {
			return err
		}
		t.Due = &d
	}

	return withList(path, func(l *todo.List) error {
		added := l.Add(t)
		fmt.Printf("added #%d: %s\n", added.ID, added.Title)
		return nil
	})
}
