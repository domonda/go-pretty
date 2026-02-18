package pretty

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func asJSON(input any, indent ...string) (data []byte, err error) {
	var indentStr string
	if len(indent) == 0 {
		indentStr = "  "
	} else {
		indentStr = strings.Join(indent, "")
	}
	return json.MarshalIndent(input, "", indentStr)
}

// PrintAsJSON marshals input as indented JSON
// and prints the result via fmt.Print.
// If indent arguments are given, they are joined into
// a string and used as JSON line indent.
// If no indent argument is given, two spaces will be used
// to indent JSON lines.
func PrintAsJSON(input any, indent ...string) (n int, err error) {
	data, err := asJSON(input, indent...)
	if err != nil {
		return 0, err
	}
	return fmt.Print(string(data))
}

// PrintlnAsJSON marshals input as indented JSON
// and prints the result via fmt.Println.
// If indent arguments are given, they are joined into
// a string and used as JSON line indent.
// If no indent argument is given, two spaces will be used
// to indent JSON lines.
func PrintlnAsJSON(input any, indent ...string) (n int, err error) {
	data, err := asJSON(input, indent...)
	if err != nil {
		return 0, err
	}
	return fmt.Println(string(data))
}

// SprintAsJSON marshals input as indented JSON
// and returns the result as a string.
// If indent arguments are given, they are joined into
// a string and used as JSON line indent.
// If no indent argument is given, two spaces will be used
// to indent JSON lines.
func SprintAsJSON(input any, indent ...string) (string, error) {
	data, err := asJSON(input, indent...)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FprintAsJSON marshals input as indented JSON
// and writes the result to w.
// If indent arguments are given, they are joined into
// a string and used as JSON line indent.
// If no indent argument is given, two spaces will be used
// to indent JSON lines.
func FprintAsJSON(w io.Writer, input any, indent ...string) (n int, err error) {
	data, err := asJSON(input, indent...)
	if err != nil {
		return 0, err
	}
	return w.Write(data)
}
