//go:build windows

package ansi

import (
	"sync"

	"golang.org/x/sys/windows"
)

var (
	ansiErr  error
	ansiOnce sync.Once
)

func EnableANSI() error {
	ansiOnce.Do(func() {
		ansiErr = realEnableANSI()
	})
	return ansiErr
}

func realEnableANSI() error {
	// We want to enable the following modes, if they are not already set:
	// ENABLE_VIRTUAL_TERMINAL_PROCESSING on stdout (color support)
	// ENABLE_VIRTUAL_TERMINAL_INPUT on stdin (ansi input sequences)
	// See https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences
	if err := windowsSetMode(windows.STD_OUTPUT_HANDLE, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return err
	}
	if err := windowsSetMode(windows.STD_INPUT_HANDLE, windows.ENABLE_VIRTUAL_TERMINAL_INPUT); err != nil {
		return err
	}
	return nil
}

func windowsSetMode(stdhandle uint32, modeFlag uint32) (err error) {
	handle, err := windows.GetStdHandle(stdhandle)
	if err != nil {
		return err
	}

	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}

	// Enable the mode if it is not currently set
	if mode&modeFlag != modeFlag {
		mode = mode | modeFlag
		err = windows.SetConsoleMode(handle, mode)
		if err != nil {
			return err
		}
	}

	return nil
}
