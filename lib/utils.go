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
	// multiple escapes are accepted within the same block, e.g. [[\x00\x01\x03]]
	hexEscapeRegex = regexp.MustCompile(`^\[\[(\\x[0-9a-fA-F]{2})+\]\]`)
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
			if matched := hexEscapeRegex.FindString(line); matched != "" {
				// [[\x01\x02]]
				for i := 2; i < len(matched)-2; i += 4 {
					if val, err := strconv.ParseUint(line[i+2:i+4], 16, 8); err == nil {
						buf.WriteByte(byte(val))
					}
				}
				line = line[len(matched):]
				continue LineLoop
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
		command := strings.TrimRight(line, "\r\n")
		command = strings.TrimLeft(command, " \t\v\r")
		if command != "" && !strings.HasPrefix(command, "#") {
			commands = append(commands, command)
		}
		if err == io.EOF {
			return commands, nil
		} else if err != nil {
			return commands, err
		}
	}
}
