// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

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
