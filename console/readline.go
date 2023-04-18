//go:build !minimal

package console

import (
	"io"
	"log"
	"os"
	"syscall"

	"golang.org/x/term"

	"github.com/ergochat/ircdog/ansi"
)

// termConsole uses term.Terminal to implement readline functionality
type termConsole struct {
	state *term.State
	term  *term.Terminal
}

func newTermConsole() (Console, error) {
	if err := ansi.EnableANSI(); err != nil {
		return nil, err
	}
	state, err := term.MakeRaw(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	c := struct {
		io.Reader
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}
	term := term.NewTerminal(c, ">>> ")
	result := &termConsole{
		state: state,
		term: term,
	}
	log.SetOutput(result)
	return result, nil
}

func (c *termConsole) Write(b []byte) (n int, err error) {
	return c.term.Write(b)
}

func (c *termConsole) Readline() (string, error) {
	return c.term.ReadLine()
}

func (c *termConsole) Close() error {
	return term.Restore(int(syscall.Stdin), c.state)
}

func NewConsole(enableReadline bool) (Console, error) {
	if enableReadline && term.IsTerminal(int(syscall.Stdin)) {
		return newTermConsole()
	}
	return NewStandardConsole()
}
