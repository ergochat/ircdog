// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ergochat/irc-go/ircutils"
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

func EncodeSASLPlain(username, password string) []string {
	buf := make([]byte, 2*len(username)+len(password)+2)
	pos := buf
	// authzid, optional in most implementations but we'll include it
	copy(pos, username[:])
	pos = pos[len(username):]
	pos[0] = '\x00'
	pos = pos[1:]
	// authcid, required
	copy(pos, username[:])
	pos = pos[len(username):]
	pos[0] = '\x00'
	pos = pos[1:]
	copy(pos, password[:])

	encoded := ircutils.EncodeSASLResponse(buf)
	result := make([]string, len(encoded))
	for i, enc := range encoded {
		result[i] = "AUTHENTICATE " + enc
	}
	return result
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

type ScriptCommandType uint

const (
	ScriptMessage ScriptCommandType = iota
	ScriptSleep
)

type ScriptCommand struct {
	Type    ScriptCommandType
	Message string
	Sleep   time.Duration
}

func ReadScript(filename string) (commands []ScriptCommand, err error) {
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

		switch {
		case command == "":
			// ignore
		case strings.HasPrefix(command, "#"):
			// comment, ignore
		case strings.HasPrefix(command, "*"):
			if sc, pErr := parseStarCommand(command); pErr == nil {
				commands = append(commands, sc...)
			}
		default:
			sc := ScriptCommand{Type: ScriptMessage, Message: command}
			commands = append(commands, sc)
		}

		if err == io.EOF {
			return commands, nil
		} else if err != nil {
			return commands, err
		}
	}
}

var invalidCommand = errors.New("invalid command")

func parseStarCommand(origCommand string) (sc []ScriptCommand, err error) {
	fields := strings.Fields(origCommand)
	if len(fields) > 0 {
		switch strings.ToLower(fields[0]) {
		case "*sleep":
			if len(fields) == 2 {
				durStr := fields[1]
				if dur, err := time.ParseDuration(durStr); err == nil {
					return []ScriptCommand{{Type: ScriptSleep, Sleep: dur}}, nil
				}
				if floatDur, err := strconv.ParseFloat(durStr, 64); err == nil {
					dur := time.Duration(floatDur * float64(time.Second))
					return []ScriptCommand{{Type: ScriptSleep, Sleep: dur}}, nil
				}
			}
		case "*saslplain":
			if len(fields) == 3 {
				var result []ScriptCommand
				for _, str := range EncodeSASLPlain(fields[1], fields[2]) {
					result = append(
						result,
						ScriptCommand{Type: ScriptMessage, Message: str},
					)
				}
				return result, nil
			}
		}
	}
	return sc, invalidCommand
}
