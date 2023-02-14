// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ergochat/irc-go/ircfmt"
)

const (
	// [CTCP] in bold with inverted colors
	ctcpMarker = "\x1b[1;7m[CTCP]\x1b[0m"

	ansiStart         = "\x1b["
	ansiBold          = "1;"
	ansiUnderline     = "4;"
	ansiReverseColor  = "7;"
	ansiStrikethrough = "9;"
	ansiItalic        = "3;"
	ansiEnd           = "m"
	ansiReset         = "\x1b[0m"
)

func IRCMessageToAnsi(message string, colorLevel ColorLevel, outputItalics bool) string {
	if colorLevel == ColorLevelNone {
		return ircfmt.Strip(message)
	}

	isCTCP := false
	if len(message) > 2 && message[0] == '\x01' && message[len(message)-1] == '\x01' {
		isCTCP = true
		message = message[1 : len(message)-1]
	}

	chunks := ircfmt.Split(message)
	if !isCTCP {
		// fast paths for messages with no formatting characters
		if len(chunks) == 0 {
			return message
		} else if len(chunks) == 1 && !chunks[0].IsFormatted() {
			return chunks[0].Content
		}
	}

	var buf bytes.Buffer
	if isCTCP {
		buf.WriteString(ctcpMarker)
	}
	for _, chunk := range chunks {
		writeChunkAsAnsi(&buf, chunk, colorLevel, outputItalics)
	}
	if isCTCP {
		buf.WriteString(ctcpMarker)
	}
	return buf.String()
}

// normalizeChunk wipes out undisplayable formatting, so we can detect when
// we don't need to emit an ANSI escape code at all
func normalizeChunk(chunk ircfmt.FormattedSubstring, colorLevel ColorLevel, outputItalics bool) (result ircfmt.FormattedSubstring) {
	chunk.Monospace = false // assume terminal is monospace
	if !outputItalics {
		chunk.Italic = false
	}
	if colorLevel <= ColorLevelBasic {
		if chunk.ForegroundColor.Value >= 16 {
			chunk.ForegroundColor = ircfmt.ColorCode{}
		}
		if chunk.BackgroundColor.Value >= 16 {
			chunk.BackgroundColor = ircfmt.ColorCode{}
		}
	}
	return chunk
}

func writeChunkAsAnsi(buf *bytes.Buffer, chunk ircfmt.FormattedSubstring, colorLevel ColorLevel, outputItalics bool) {
	chunk = normalizeChunk(chunk, colorLevel, outputItalics)
	if !chunk.IsFormatted() {
		buf.WriteString(chunk.Content)
		return
	}

	// ANSI uses ; as the delimiter between flags:
	// `truncate` is whether we need to cut off the final trailing ;
	truncate := false
	buf.WriteString(ansiStart)
	if chunk.Bold {
		buf.WriteString(ansiBold)
		truncate = true
	}
	if chunk.Underline {
		buf.WriteString(ansiUnderline)
		truncate = true
	}
	if chunk.Strikethrough {
		buf.WriteString(ansiStrikethrough)
		truncate = true
	}
	if chunk.Italic {
		buf.WriteString(ansiItalic)
		truncate = true
	}
	if chunk.ReverseColor {
		buf.WriteString(ansiReverseColor)
		truncate = true
	}
	if chunk.ForegroundColor.IsSet {
		truncate = writeAnsiColorCode(buf, colorLevel, chunk.ForegroundColor.Value, false) || truncate
	}
	if chunk.BackgroundColor.IsSet {
		truncate = writeAnsiColorCode(buf, colorLevel, chunk.BackgroundColor.Value, true) || truncate
	}
	if truncate {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteString(ansiEnd)
	buf.WriteString(chunk.Content)
	buf.WriteString(ansiReset)
}

func writeAnsiColorCode(buf *bytes.Buffer, colorLevel ColorLevel, ircColor uint8, background bool) (ok bool) {
	if colorLevel >= ColorLevelAnsi256 {
		// unclear whether there is any potential benefit from true-color support
		return writeAnsiColorCode256(buf, ircColor, background)
	} else {
		return writeAnsiColorCode16(buf, ircColor, background)
	}
}

func writeAnsiColorCode16(buf *bytes.Buffer, ircColor uint8, background bool) (ok bool) {
	if ircColor >= 16 {
		return false
	}
	code := ircColorToAnsiForeground[ircColor]

	if background {
		// the normal-intensity foreground codes in the [30-37] block have background
		// counterparts from [40-47]; similarly, the high-intensity foreground codes
		// in the [90-97] block have background counterparts from [100-107]
		code += 10
	}
	fmt.Fprintf(buf, "%d;", code)
	return true
}

func writeAnsiColorCode256(buf *bytes.Buffer, ircColor uint8, background bool) (ok bool) {
	if ircColor >= 99 {
		return false
	}
	code, ok := ircColorToAnsi256[ircColor]
	if !ok {
		return writeAnsiColorCode16(buf, ircColor, background)
	}
	if !background {
		fmt.Fprintf(buf, "38;5;%d;", code)
	} else {
		fmt.Fprintf(buf, "48;5;%d;", code)
	}
	return true
}

// slice off any amount of ' ' from the front of the string
func trimInitialSpaces(str string) string {
	var i int
	for i = 0; i < len(str) && str[i] == ' '; i++ {
	}
	return str[i:]
}

func IRCLineToAnsi(line string, colorLevel ColorLevel, outputItalics bool) string {
	// skip to the last parameter without actually parsing the message;
	// the rationale is that we don't want to destroy any idiosyncrasies
	// of the original IRC line (tag order, extra spaces between params, etc.)
	remaining := line

	skipPastNextSpace := func(in string) (ok bool, out string) {
		if idx := strings.IndexByte(in, ' '); idx != -1 {
			return true, trimInitialSpaces(in[idx+1:])
		} else {
			return false, in
		}
	}

	var ok bool
	// skip the tags section
	if len(remaining) == 0 {
		return line
	}
	if remaining[0] == '@' {
		ok, remaining = skipPastNextSpace(remaining)
		if !ok {
			return line
		}
	}

	if len(remaining) == 0 {
		return line
	}

	// skip the source
	if remaining[0] == ':' {
		ok, remaining = skipPastNextSpace(remaining)
		if !ok {
			return line
		}
	}

	// skip the command
	ok, remaining = skipPastNextSpace(remaining)
	if !ok {
		return line
	}

	// parameters section: we want the final parameter
	for {
		if len(remaining) == 0 {
			return line
		}
		if remaining[0] == ':' {
			// trailing
			remaining = remaining[1:]
			break
		}
		ok, remaining = skipPastNextSpace(remaining)
		if !ok {
			break
		}
	}

	conv := IRCMessageToAnsi(remaining, colorLevel, outputItalics)
	if conv != remaining {
		return line[:len(line)-len(remaining)] + conv
	} else {
		return line
	}
}

func ansiDebugEscape(in string) (out string) {
	return strings.Replace(in, "\x1b", "{ESC}", -1)
}
