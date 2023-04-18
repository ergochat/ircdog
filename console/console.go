// Copyright (c) 2023 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the ISC license

package console

import (
	"io"
	"os"

	"github.com/ergochat/irc-go/ircreader"

	"github.com/ergochat/ircdog/lib"
)

// Console is an abstract representation of keyboard input and screen output
type Console interface {
	io.Writer

	Readline() (string, error)

	// this is a hook to perform terminal cleanup, as in chzyer/readline
	Close() error
}

type stdioConsole struct {
	reader ircreader.Reader
}

func NewStandardConsole() (Console, error) {
	result := new(stdioConsole)
	result.reader.Initialize(os.Stdin, lib.InitialBufferSize, lib.MaxBufferSize)
	return result, nil
}

func (s *stdioConsole) Readline() (string, error) {
	lineBytes, err := s.reader.ReadLine()
	return string(lineBytes), err
}

func (s *stdioConsole) Write(b []byte) (n int, err error) {
	return os.Stdout.Write(b)
}

func (s *stdioConsole) Close() error {
	return nil
}
