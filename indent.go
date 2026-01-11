package pretty

import (
	"strings"
	"unicode/utf8"
)

// Indent pretty printed source using the passed indent string
// and an optional linePrefix used for every line in case of
// a multiple line result.
// Multiple linePrefix values are concatenated into a single string.
func Indent(source []byte, indent string, linePrefix ...string) []byte {
	const (
		stateDefault = iota
		stateRawString
		stateEscString
	)
	var (
		state         = stateDefault
		newLineIndent = "\n" + strings.Join(linePrefix, "")
		result        = make([]byte, 0, len(source)+256)
		unwritten     = 0
		i             int
		r             rune
		rSize         int

		appendUnwritten = func() {
			next := i + rSize
			result = append(result, source[unwritten:next]...)
			unwritten = next
		}
	)
	for i = 0; i < len(source); i += rSize {
		r, rSize = utf8.DecodeRune(source[i:])
		if r == utf8.RuneError {
			break
		}
		if i == 0 {
			for _, prefix := range linePrefix {
				result = append(result, prefix...)
			}
		}
		switch state {
		case stateDefault:
			switch r {
			case ':':
				appendUnwritten()
				result = append(result, ' ')
			case ';':
				result = append(result, source[unwritten:i]...)
				unwritten = i + 1
				result = append(result, newLineIndent...)
			case '{':
				appendUnwritten()
				if i+1 < len(source) && source[i+1] == '}' {
					// no newLineIndent for {}
					result = append(result, '}')
					unwritten++
					i++
					continue
				}
				newLineIndent += indent
				result = append(result, newLineIndent...)
			case '}':
				result = append(result, source[unwritten:i]...)
				unwritten = i + 1
				if len(newLineIndent) >= len(indent) {
					newLineIndent = newLineIndent[:len(newLineIndent)-len(indent)]
				}
				result = append(result, newLineIndent...)
				result = append(result, '}')
			case '`':
				state = stateRawString
			case '"':
				state = stateEscString
			}

		case stateRawString:
			if r == '`' {
				next := i + rSize
				result = append(result, source[unwritten:next]...)
				unwritten = next
				state = stateDefault
			}

		case stateEscString:
			switch r {
			case '"':
				next := i + rSize
				result = append(result, source[unwritten:next]...)
				unwritten = next
				state = stateDefault

			case '\\':
				next := i + 1
				if next < len(source) && (source[next] == '\\' || source[next] == '"') {
					// Skip next character to prevent interpreting it as string end
					rSize = 2
				}
				// tail0 := string(source[i:])
				// _, _, tail1, err := strconv.UnquoteChar(tail0, '"')
				// if err != nil {
				// 	continue
				// }
				// rSize = len(tail0) - len(tail1)
			}
		}
	}

	// Append any remaining unwritten content
	result = append(result, source[unwritten:]...)

	return result
}
