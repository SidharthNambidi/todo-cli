package cmd

import (
	"flag"

	"todo-cli/todo"
)

func runDone(path string, args []string) error {
	fs := flag.NewFlagSet("done", flag.ExitOnError)
	id := fs.Int("id", 0, "task id")
	fs.Parse(args)

	return withList(path, func(l *todo.List) error {
		return l.MarkDone(*id)
	})
}
