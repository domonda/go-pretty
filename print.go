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

// Println pretty prints a value to os.Stdout followed by a newline
func Println(value interface{}, indent ...string) {
	DefaultPrinter.Println(value, indent...)
}

// Print pretty prints a value to os.Stdout
func Print(value interface{}, indent ...string) {
	DefaultPrinter.Print(value, indent...)
}

// Fprint pretty prints a value to a io.Writer
func Fprint(w io.Writer, value interface{}, indent ...string) {
	DefaultPrinter.Fprint(w, value, indent...)
}

// Fprint pretty prints a value to a io.Writer followed by a newline
func Fprintln(w io.Writer, value interface{}, indent ...string) {
	DefaultPrinter.Fprintln(w, value, indent...)
}

// Sprint pretty prints a value to a string
func Sprint(value interface{}, indent ...string) string {
	return DefaultPrinter.Sprint(value, indent...)
}
