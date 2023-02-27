// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// e.g., [[\x00]] for \x00, [[\xFF]] or [[\xff]] for \xff
	hexEscapeRegex = regexp.MustCompile(`^\[\[\\x[0-9a-fA-F]{2}\]\]`)
)

var controlCodeReplacements = []struct {
	escape string
	value  byte
}{
	{"[[CTCP]]", '\x01'},
	{"[[B]]", '\x02'},
	{"[[C]]", '\x03'},
	{"[[M]]", '\x11'},
	{"[[I]]", '\x1d'},
	{"[[S]]", '\x1e'},
	{"[[U]]", '\x1f'},
	{"[[R]]", '\x0f'},
}

// ReplaceControlCodes applies our control code replacements to the line.
func ReplaceControlCodes(line string) string {
	if idx := strings.Index(line, "[["); idx == -1 {
		return line
	}

	var buf strings.Builder

LineLoop:
	for line != "" {
		if line[0] == '[' {
			for _, replacement := range controlCodeReplacements {
				if strings.HasPrefix(line, replacement.escape) {
					buf.WriteByte(replacement.value)
					line = line[len(replacement.escape):]
					continue LineLoop
				}
			}
			if hexEscapeRegex.MatchString(line) {
				if val, err := strconv.ParseUint(strings.ToLower(line[4:6]), 16, 8); err == nil {
					buf.WriteByte(byte(val))
					line = line[8:]
					continue LineLoop
				}
			}
		}
		buf.WriteByte(line[0])
		line = line[1:]
	}

	return buf.String()
}

func ReadScript(filename string) (commands []string, err error) {
	infile, err := os.Open(filename)
	if err != nil {
		return
	}
	defer infile.Close()
	reader := bufio.NewReader(infile)
	for {
		line, err := reader.ReadString('\n')
		if command := strings.TrimRight(line, "\r\n"); command != "" {
			commands = append(commands, command)
		}
		if err == io.EOF {
			return commands, nil
		} else if err != nil {
			return commands, err
		}
	}
}
