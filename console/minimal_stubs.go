//go:build minimal

package console

func NewConsole(enableReadline bool, historyFile string) (Console, error) {
	return NewStandardConsole()
}
