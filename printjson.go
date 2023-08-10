package pretty

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PrintAsJSON marshalles input as indented JSON
// and calles fmt.Println with the result.
// If indent arguments are given, they are joined into
// a string and used as JSON line indent.
// If no indet argument is given, two spaces will be used
// to indent JSON lines.
// A byte slice as input will be marshalled as json.RawMessage.
func PrintAsJSON(input interface{}, indent ...string) {
	var indentStr string
	if len(indent) == 0 {
		indentStr = "  "
	} else {
		indentStr = strings.Join(indent, "")
	}
	if b, ok := input.([]byte); ok {
		input = json.RawMessage(b)
	}
	data, err := json.MarshalIndent(input, "", indentStr)
	if err != nil {
		_, _ = fmt.Println(fmt.Errorf("%w from input: %#v", err, input))
		return
	}
	_, _ = fmt.Println(string(data))
}
