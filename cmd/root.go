package cmd

import (
	"errors"
	"fmt"

	"todo-cli/todo"
)

type handler func(path string, args []string) error

var commands = map[string]handler{
	"init":   runInit,
	"add":    runAdd,
	"list":   runList,
	"update": runUpdate,
	"done":   runDone,
	"delete": runDelete,
}

func Execute(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: todo <init|add|list|update|done|delete> [flags]")
	}

	h, ok := commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q", args[0])
	}

	path, err := todo.FilePath()
	if err != nil {
		return err
	}
	return h(path, args[1:])
}

var errNoSave = errors.New("no changes to save")

func withList(path string, fn func(*todo.List) error) error {
	l, err := todo.Load(path)
	if err != nil {
		return err
	}
	if err := fn(l); err != nil {
		if errors.Is(err, errNoSave) {
			return nil
		}
		return err
	}
	return todo.Save(path, l)
}
