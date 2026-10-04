package cmd

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
)

func newTestPrompter(input string) *prompter {
	return &prompter{in: bufio.NewReader(strings.NewReader(input)), out: io.Discard}
}

func TestAskUsesDefaultOnEnter(t *testing.T) {
	p := newTestPrompter("\nhello\n")

	if got, _ := p.ask("X", "fallback"); got != "fallback" {
		t.Errorf("blank answer: got %q, want %q", got, "fallback")
	}
	if got, _ := p.ask("X", "fallback"); got != "hello" {
		t.Errorf("typed answer: got %q, want %q", got, "hello")
	}
}

func TestAskParsedRetriesUntilValid(t *testing.T) {
	p := newTestPrompter("abc\n-5\n7\n")

	got, err := askParsed(p, "ID", "", parseID)

	if err != nil || got != 7 {
		t.Errorf("got %d, %v; want 7, nil", got, err)
	}
}

func TestAskCancelsOnEOF(t *testing.T) {
	p := newTestPrompter("") // like pressing Ctrl-D immediately

	if _, err := p.ask("X", ""); !errors.Is(err, errCancelled) {
		t.Errorf("got %v, want errCancelled", err)
	}
}
