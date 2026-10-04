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
	task := fs.String("task", "", "task description (prompts if omitted)")
	cat := fs.String("cat", "General", "category")
	prio := fs.String("priority", "medium", "low, medium or high")
	due := fs.String("due", "", "due date: YYYY-MM-DD, today, tomorrow or +Nd")
	fs.Parse(args)

	now := time.Now()
	var t todo.Task

	if *task != "" {
		// Flag mode, same as before.
		pr, err := todo.ParsePriority(*prio)
		if err != nil {
			return err
		}
		d, err := parseOptDate(*due, now)
		if err != nil {
			return err
		}
		t = todo.Task{Title: *task, Category: *cat, Priority: pr, Due: d}
	} else {
		if !isInteractive() {
			return errors.New("-task is required when stdin is not a terminal")
		}
		var err error
		// Any flags that were passed become the prompt defaults.
		t, err = promptNewTask(newPrompter(), *cat, *prio, *due, now)
		if err != nil {
			return err
		}
	}

	return withList(path, func(l *todo.List) error {
		added := l.Add(t)
		fmt.Printf("added #%d: %s\n", added.ID, added.Title)
		return nil
	})
}

func promptNewTask(p *prompter, cat, prio, due string, now time.Time) (todo.Task, error) {
	title, err := askParsed(p, "Task", "", nonEmpty)
	if err != nil {
		return todo.Task{}, err
	}
	category, err := p.ask("Category", cat)
	if err != nil {
		return todo.Task{}, err
	}
	priority, err := askParsed(p, "Priority (low/medium/high)", prio, todo.ParsePriority)
	if err != nil {
		return todo.Task{}, err
	}
	dueDate, err := askParsed(p, "Due (YYYY-MM-DD, today, tomorrow, +Nd; blank for none)", due,
		func(s string) (*time.Time, error) { return parseOptDate(s, now) })
	if err != nil {
		return todo.Task{}, err
	}
	return todo.Task{Title: title, Category: category, Priority: priority, Due: dueDate}, nil
}
