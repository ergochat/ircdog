//go:build !minimal

package console

import (
	"syscall"

	"github.com/ergochat/readline"
	"golang.org/x/term"
)

func NewConsole(enableReadline bool, historyFile string) (Console, error) {
	if !(enableReadline && term.IsTerminal(int(syscall.Stdin))) {
		return NewStandardConsole()
	}
	return readline.NewFromConfig(&readline.Config{
		Prompt:       ">>> ",
		HistoryFile:  historyFile,
		HistoryLimit: 1000,
	})
}
