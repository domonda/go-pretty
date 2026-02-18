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

// PrintableWithResult is like Printable but returns
// the number of bytes written and any error.
type PrintableWithResult interface {
	PrettyPrint(io.Writer) (n int, err error)
}

// printableWithResultAdapter wraps a Printable as a PrintableWithResult
// by counting the bytes written via a countingWriter.
type printableWithResultAdapter struct {
	printable Printable
}

func (p printableWithResultAdapter) PrettyPrint(w io.Writer) (n int, err error) {
	cw := newCountingWriter(w)
	p.printable.PrettyPrint(cw)
	return cw.Result()
}

// Stringer can be implemented to return a pretty printed string representation.
type Stringer interface {
	PrettyString() string
}

// PrintFunc is used to customize the pretty printing of a value.
type PrintFunc func(io.Writer) (n int, err error)

// PrintFuncForPrintable returns a PrintFunc for a reflect.Value
// if the value implements PrintableWithResult, Printable, or Stringer
// (checked for both value and pointer receivers).
// Returns nil if none of the interfaces is implemented.
func PrintFuncForPrintable(v reflect.Value) PrintFunc {
	printableWithResult, ok := tryCastReflectValue[PrintableWithResult](v)
	if ok {
		return func(w io.Writer) (n int, err error) {
			return printableWithResult.PrettyPrint(w)
		}
	}

	printable, ok := tryCastReflectValue[Printable](v)
	if ok {
		return func(w io.Writer) (n int, err error) {
			return printableWithResultAdapter{printable}.PrettyPrint(w)
		}
	}

	stringer, ok := tryCastReflectValue[Stringer](v)
	if ok {
		return func(w io.Writer) (n int, err error) {
			return io.WriteString(w, stringer.PrettyString())
		}
	}

	return nil
}

func tryCastReflectValue[T any](v reflect.Value) (T, bool) {
	t, ok := v.Interface().(T)
	if !ok && v.CanAddr() {
		t, ok = v.Addr().Interface().(T)
	}
	return t, ok
}
