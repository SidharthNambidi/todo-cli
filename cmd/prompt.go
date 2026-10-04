package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"todo-cli/todo"
)

var errCancelled = errors.New("cancelled")

//reports whether stdin is a terminal rather than a pipe or file.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// prompter reads answers from in and writes questions to out. both are
// fields (not os.Stdin/os.Stdout hardcoded) so tests can swap them
type prompter struct {
	in  *bufio.Reader
	out io.Writer
}

func newPrompter() *prompter {
	return &prompter{in: bufio.NewReader(os.Stdin), out: os.Stdout}
}

// ask prints the label (with the default in brackets) and returns the
// trimmed answer, or def if the user just presses enter
func (p *prompter) ask(label, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(p.out, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(p.out, "%s: ", label)
	}

	line, err := p.in.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line == "" { // Ctrl-D on an empty line
			fmt.Fprintln(p.out)
			return "", errCancelled
		}
		if !errors.Is(err, io.EOF) {
			return "", err
		}
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

//keeps asking until parse accepts the answer
func askParsed[T any](p *prompter, label, def string, parse func(string) (T, error)) (T, error) {
	for {
		s, err := p.ask(label, def)
		if err != nil {
			var zero T
			return zero, err
		}
		v, err := parse(s)
		if err == nil {
			return v, nil
		}
		fmt.Fprintf(p.out, "  %v\n", err)
	}
}

func nonEmpty(s string) (string, error) {
	if s == "" {
		return "", errors.New("this field is required")
	}
	return s, nil
}

func parseID(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%q is not a valid id", s)
	}
	return n, nil
}

func parseOptDate(s string, now time.Time) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	d, err := todo.ParseDate(s, now)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func idOrPrompt(p *prompter, l *todo.List, id int) (int, error) {
	if id != 0 {
		return id, nil
	}
	if p == nil {
		return 0, errors.New("-id is required")
	}
	now := time.Now()
	tasks := l.Filter(todo.Filter{}, now)
	if len(tasks) == 0 {
		return 0, errors.New("no tasks yet")
	}
	if err := printTasks(os.Stdout, tasks, now); err != nil {
		return 0, err
	}
	fmt.Println()
	return askParsed(p, "Task ID", "", parseID)
}
