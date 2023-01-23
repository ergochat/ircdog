// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"regexp"
	"strconv"
	"strings"
)

// SplitLineIntoParts splits the given IRC line into separate parts.
func SplitLineIntoParts(line string) []string {
	var lineParts []string
	var buffer string
	var isSpace bool
	var isTrailing bool
	var haveHadCommand bool

	for i, charRune := range line {
		char := string(charRune)

		if i == 0 && char == " " {
			isSpace = true
		}

		// trailing behaviour
		if isTrailing {
			buffer += char
			continue
		}

		// check for changes
		if isSpace && char != " " {
			isSpace = false
			lineParts = append(lineParts, buffer)
			buffer = char
			if haveHadCommand && char == ":" {
				isTrailing = true
			}
			continue
		} else if !isSpace && char == " " {
			isSpace = true
			lineParts = append(lineParts, buffer)
			if !haveHadCommand && buffer[0] != '@' && buffer[0] != ':' {
				haveHadCommand = true
			}
			buffer = char
			continue
		}

		// no changes, just append
		buffer += char
	}
	lineParts = append(lineParts, buffer)

	return lineParts
}

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
