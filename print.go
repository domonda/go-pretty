// Package pretty offers print functions
// that format values of any Go type in a
// compact single line string
// suitable for logging and debugging.
// Strings are escaped to be single line
// with fmt.Sprintf("%#q", s).
// %#q is used instead of %q to minimize
// the number of double quotes that would
// have to be escaped in JSON logs.
//
// MaxStringLength, MaxErrorLength, MaxSliceLength
// can be set to values greater zero to prevent excessive log sizes.
// An ellipsis rune is used as last element to represent
// the truncated elements.
package pretty

import (
	"io"
)

// Println pretty prints a value to os.Stdout followed by a newline.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func Println(value any, indent ...string) (n int, err error) {
	return DefaultPrinter.Println(value, indent...)
}

// Print pretty prints a value to os.Stdout.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func Print(value any, indent ...string) (n int, err error) {
	return DefaultPrinter.Print(value, indent...)
}

// Fprint pretty prints a value to a io.Writer.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func Fprint(w io.Writer, value any, indent ...string) (n int, err error) {
	return DefaultPrinter.Fprint(w, value, indent...)
}

// Fprintln pretty prints a value to a io.Writer followed by a newline.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func Fprintln(w io.Writer, value any, indent ...string) (n int, err error) {
	return DefaultPrinter.Fprintln(w, value, indent...)
}

// Sprint pretty prints a value to a string.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func Sprint(value any, indent ...string) string {
	return DefaultPrinter.Sprint(value, indent...)
}
