package lib

// this is copied and pasted from https://github.com/jwalton/go-supportscolor
// to avoid a direct dependency of our lib/ package on the upstream library

// ColorLevel represents the ANSI color level supported by the terminal.
type ColorLevel int

const (
	// None represents a terminal that does not support color at all.
	ColorLevelNone ColorLevel = 0
	// Basic represents a terminal with basic 16 color support.
	ColorLevelBasic ColorLevel = 1
	// Ansi256 represents a terminal with 256 color support.
	ColorLevelAnsi256 ColorLevel = 2
	// Ansi16m represents a terminal with full true color support.
	ColorLevelAnsi16m ColorLevel = 3
)
