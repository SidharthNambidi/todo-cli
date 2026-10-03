package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"todo-cli/todo"
)

func runList(path string, args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	done := fs.Int("done", -1, "1 = completed, 0 = pending, -1 = all")
	cat := fs.String("cat", "", "filter by category")
	prio := fs.String("priority", "", "filter by priority")
	overdue := fs.Bool("overdue", false, "only overdue tasks")
	sortBy := fs.String("sort", "id", "sort by id, priority or due")
	fs.Parse(args)

	f := todo.Filter{Category: *cat, Overdue: *overdue}
	if *done != -1 {
		b := *done == 1
		f.Done = &b
	}
	if *prio != "" {
		p, err := todo.ParsePriority(*prio)
		if err != nil {
			return err
		}
		f.Priority = p
	}

	l, err := todo.Load(path)
	if err != nil {
		return err
	}
	now := time.Now()
	tasks := l.Filter(f, now)
	if err := todo.SortTasks(tasks, *sortBy); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDONE\tPRI\tDUE\tCATEGORY\tTASK")
	for _, t := range tasks {
		mark := " "
		if t.Done {
			mark = "x"
		}
		due := "-"
		if t.Due != nil {
			due = t.Due.Format(todo.DateLayout)
			if t.Overdue(now) {
				due += " (overdue)"
			}
		}
		fmt.Fprintf(w, "%d\t[%s]\t%s\t%s\t%s\t%s\n", t.ID, mark, t.Priority, due, t.Category, t.Title)
	}
	return w.Flush()
}
