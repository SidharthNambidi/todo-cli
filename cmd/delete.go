package cmd

import (
	"flag"

	"todo-cli/todo"
)

func runDelete(path string, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	id := fs.Int("id", 0, "task id")
	fs.Parse(args)

	return withList(path, func(l *todo.List) error {
		return l.Delete(*id)
	})
}
