package cmd

import (
	"flag"
	"fmt"
	"strings"

	"todo-cli/todo"
)

func runDelete(path string, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	id := fs.Int("id", 0, "task id (prompts if omitted)")
	yes := fs.Bool("y", false, "skip the confirmation question")
	fs.Parse(args)

	var p *prompter
	if isInteractive() {
		p = newPrompter()
	}
	return withList(path, func(l *todo.List) error {
		taskID, err := idOrPrompt(p, l, *id)
		if err != nil {
			return err
		}
		t, err := l.Get(taskID)
		if err != nil {
			return err
		}

		if p != nil && !*yes {
			ans, err := p.ask(fmt.Sprintf("Delete #%d %q? (y/N)", t.ID, t.Title), "")
			if err != nil {
				return err
			}
			if a := strings.ToLower(ans); a != "y" && a != "yes" {
				fmt.Println("cancelled")
				return errNoSave
			}
		}

		if err := l.Delete(taskID); err != nil {
			return err
		}
		fmt.Printf("deleted #%d\n", taskID)
		return nil
	})
}
