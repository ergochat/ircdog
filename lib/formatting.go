// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"strings"

	"github.com/goshuirc/irc-go/ircfmt"
	"github.com/mgutz/ansi"
)

var (
	colourToAnsi = map[string]string{
		"white":       "white+h",
		"black":       "black",
		"blue":        "blue",
		"green":       "green",
		"red":         "red+h",
		"brown":       "red",
		"magenta":     "magenta",
		"orange":      "yellow",
		"yellow":      "yellow+h",
		"light green": "green+h",
		"cyan":        "cyan",
		"light cyan":  "cyan+h",
		"light blue":  "blue+h",
		"pink":        "magenta+h",
		"grey":        "black+h",
		"light grey":  "white",
	}

	italicOn = makeANSI("3")
)

func makeANSI(code string) string {
	return "\033[" + code + "m"
}

// formattedPart is a single section of the line, along with the given attributes it has.
type formattedPart struct {
	ForeColour    string
	BackColour    string
	Bold          bool
	Monospace     bool
	Strikethrough bool
	Underlined    bool
	Italic        bool
	Content       string
}

func formatLinePart(rawPart string, outputItalics bool) string {
	// []rune so that we loop over chars (keeping UTF-8 chars intact)
	remaining := []rune(ircfmt.Escape(rawPart))
	var lineParts []formattedPart
	var buffer string
	var isBold, isItalic, isMonospace, isStrikethrough, isUnderline bool
	var storedFgColour, storedBgColour string

	for 0 < len(remaining) {
		char := remaining[0]
		remaining = remaining[1:]

		if char == '$' {
			char = remaining[0]
			remaining = remaining[1:]

			if char == '$' {
				buffer += "$"
				continue
			}

			if 0 < len(buffer) {
				lineParts = append(lineParts, formattedPart{
					ForeColour:    storedFgColour,
					BackColour:    storedBgColour,
					Bold:          isBold,
					Monospace:     isMonospace,
					Strikethrough: isStrikethrough,
					Underlined:    isUnderline,
					Italic:        isItalic,
					Content:       buffer,
				})
				buffer = ""
			}

			if char == 'b' {
				isBold = !isBold
			} else if char == 'i' {
				isItalic = !isItalic
			} else if char == 's' {
				isStrikethrough = !isStrikethrough
			} else if char == 'u' {
				isUnderline = !isUnderline
			} else if char == 'm' {
				isMonospace = !isMonospace
			} else if char == 'r' {
				isBold = false
				isMonospace = false
				isItalic = false
				isStrikethrough = false
				isUnderline = false
				storedFgColour = ""
				storedBgColour = ""
			} else if char == 'c' {
				// get colours
				var coloursBuffer string
				remaining = remaining[1:] // strip initial '['
				for remaining[0] != ']' {
					coloursBuffer += string(remaining[0])
					remaining = remaining[1:]
				}
				remaining = remaining[1:] // strip final ']'

				colours := strings.Split(coloursBuffer, ",")
				var foreColour, backColour string
				foreColour = colours[0]
				if 1 < len(colours) {
					backColour = colours[1]
				}

				if foreColour == "" {
					storedFgColour = ""
					storedBgColour = ""
					continue
				}
				fore := colourToAnsi[foreColour]
				back := colourToAnsi[backColour]
				if fore != "" && back != "" {
					storedFgColour = fore
					storedBgColour = back
				} else if fore != "" {
					storedFgColour = fore
				}
			}
		} else {
			buffer += string(char)
		}
	}
	// add last buffered part
	if 0 < len(buffer) {
		lineParts = append(lineParts, formattedPart{
			ForeColour:    storedFgColour,
			BackColour:    storedBgColour,
			Bold:          isBold,
			Monospace:     isMonospace,
			Strikethrough: isStrikethrough,
			Underlined:    isUnderline,
			Italic:        isItalic,
			Content:       buffer,
		})
	}

	// assemble message parts
	var message string

	for _, part := range lineParts {
		if part.BackColour == "" && part.ForeColour == "" && part.Bold == false && part.Underlined == false && part.Italic == false {
			message += part.Content
			continue
		}

		// we do italics separately ourselves, since lib doesn't support it.
		// it's got spotty support, so we make it optional
		if part.Italic && outputItalics {
			message += italicOn
		}

		// additional things to request
		var additional string
		if part.Bold {
			additional += "b"
		}
		if part.Strikethrough {
			additional += "s"
		}
		if part.Underlined {
			additional += "u"
		}
		if additional != "" {
			additional = "+" + additional
		}

		// colours
		if part.ForeColour != "" && part.BackColour != "" {
			message += ansi.ColorCode(part.ForeColour + additional + ":" + part.BackColour)
		} else if part.ForeColour != "" {
			message += ansi.ColorCode(part.ForeColour + additional)
		} else if additional != "" {
			message += ansi.ColorCode("default" + additional)
		}

		// paste it all together
		message += part.Content
		message += ansi.Reset
	}

	// // debugging output
	// if 1 < len(lineParts) {
	// 	spew.Dump(lineParts)
	// }
	return message
}

// AnsiFormatLineParts takes line parts and returns a formatted string with ANSI codes.
func AnsiFormatLineParts(lineParts []string, outputItalics bool) string {
	// // print raw line with goshuirc escapes for debugging
	// var slP string
	// for _, rawPart := range lineParts {
	// 	slP += ircfmt.Escape(rawPart)
	// }
	// fmt.Println(slP)

	// get actual line
	var formattedLineParts []string

	for _, rawPart := range lineParts {
		formattedLineParts = append(formattedLineParts, formatLinePart(rawPart, outputItalics))
	}

	return strings.Join(formattedLineParts, "")
}
