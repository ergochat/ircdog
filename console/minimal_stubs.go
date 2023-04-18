//go:build minimal

package console

func NewConsole(enableReadline bool) (Console, error) {
	return NewStandardConsole()
}
