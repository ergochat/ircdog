//go:build !minimal

package readline

import (
	"github.com/ergochat/ircdog/lib"

	"github.com/chzyer/readline"
)

func NewReadline(historyFile string) (lib.Console, error) {
	return readline.NewEx(&readline.Config{
		Prompt:       ">>> ",
		HistoryFile:  historyFile,
		HistoryLimit: 1000,
	})
}
