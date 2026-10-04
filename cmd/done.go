package cmd

import (
	"flag"
	"fmt"

	"todo-cli/todo"
)

func runDone(path string, args []string) error {
	fs := flag.NewFlagSet("done", flag.ExitOnError)
	id := fs.Int("id", 0, "task id (prompts if omitted)")
	fs.Parse(args)

	var p *prompter // stays nil when stdin is not a terminal
	if isInteractive() {
		p = newPrompter()
	}
	return withList(path, func(l *todo.List) error {
		taskID, err := idOrPrompt(p, l, *id)
		if err != nil {
			return err
		}
		if err := l.MarkDone(taskID); err != nil {
			return err
		}
		fmt.Printf("completed #%d\n", taskID)
		return nil
	})
}
