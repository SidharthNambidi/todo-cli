package cmd

import "todo-cli/todo"

func runInit(path string, args []string) error {
	return todo.Init(path)
}
