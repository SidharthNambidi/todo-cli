package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"todo-cli/todo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: todo <init|add|list|update|done|delete> [flags]")
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(cmd string, args []string) error {
	path, err := todo.FilePath()
	if err != nil {
		return err
	}

	if cmd == "init" {
		return todo.Init(path)
	}

	switch cmd {
	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		task := fs.String("task", "", "task description")
		cat := fs.String("cat", "General", "category")
		fs.Parse(args)
		if *task == "" {
			return errors.New("-task is required")
		}
		return withList(path, func(l *todo.List) error {
			t := l.Add(*task, *cat)
			fmt.Printf("added #%d: %s\n", t.ID, t.Title)
			return nil
		})

	case "list":
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		done := fs.Int("done", -1, "1 = completed, 0 = pending, -1 = all")
		cat := fs.String("cat", "", "filter by category")
		fs.Parse(args)
		l, err := todo.Load(path)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tDONE\tCATEGORY\tTASK")
		for _, t := range l.Tasks {
			if *done != -1 && (*done == 1) != t.Done {
				continue
			}
			if *cat != "" && t.Category != *cat {
				continue
			}
			mark := " "
			if t.Done {
				mark = "x"
			}
			fmt.Fprintf(w, "%d\t[%s]\t%s\t%s\n", t.ID, mark, t.Category, t.Title)
		}
		return w.Flush()

	case "update":
		fs := flag.NewFlagSet("update", flag.ExitOnError)
		id := fs.Int("id", 0, "task id")
		task := fs.String("task", "", "new description")
		cat := fs.String("cat", "", "new category")
		fs.Parse(args)
		return withList(path, func(l *todo.List) error {
			return l.Update(*id, *task, *cat)
		})

	case "done":
		fs := flag.NewFlagSet("done", flag.ExitOnError)
		id := fs.Int("id", 0, "task id")
		fs.Parse(args)
		return withList(path, func(l *todo.List) error {
			return l.MarkDone(*id)
		})

	case "delete":
		fs := flag.NewFlagSet("delete", flag.ExitOnError)
		id := fs.Int("id", 0, "task id")
		fs.Parse(args)
		return withList(path, func(l *todo.List) error {
			return l.Delete(*id)
		})
	}

	return fmt.Errorf("unknown command %q", cmd)
}

// withList loads the file, runs fn on it, then saves.
func withList(path string, fn func(*todo.List) error) error {
	l, err := todo.Load(path)
	if err != nil {
		return err
	}
	if err := fn(l); err != nil {
		return err
	}
	return todo.Save(path, l)
}
