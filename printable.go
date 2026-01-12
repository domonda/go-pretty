package pretty

import (
	"io"
	"reflect"
)

// Printable can be implemented to customize the pretty printing of a type.
type Printable interface {
	// PrettyPrint the implementation's data
	PrettyPrint(io.Writer)
}

// PrintFunc is used to customize the pretty printing of a value.
type PrintFunc func(io.Writer)

// PrintFuncForPrintable returns a PrintFunc for a reflect.Value
// if the value implements the Printable interface.
// Returns nil if the value does not implement Printable.
func PrintFuncForPrintable(v reflect.Value) PrintFunc {
	p, ok := v.Interface().(Printable)
	if !ok && v.CanAddr() {
		p, ok = v.Addr().Interface().(Printable)
	}
	if !ok {
		return nil
	}
	return func(w io.Writer) {
		p.PrettyPrint(w)
	}
}
