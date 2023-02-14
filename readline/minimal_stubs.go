//go:build minimal

package readline

import (
	"errors"

	"github.com/ergochat/ircdog/lib"
)

func NewReadline(historyFile string) (lib.Console, error) {
	return nil, errors.New("readline support disabled at compile time")
}
