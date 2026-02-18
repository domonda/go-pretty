package pretty

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/token"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

// Nullable can be implemented to print "null" instead of
// the representation of the underlying type's value.
type Nullable interface {
	// IsNull returns true if the implementing value is considered null.
	IsNull() bool
}

// Printer holds a pretty-print configuration
type Printer struct {
	// MaxStringLength is the maximum length for escaped strings.
	// Longer strings will be truncated with an ellipsis rune at the end.
	// A value <= 0 will disable truncating.
	MaxStringLength int

	// MaxErrorLength is the maximum length for escaped errors.
	// Longer errors will be truncated with an ellipsis rune at the end.
	// A value <= 0 will disable truncating.
	MaxErrorLength int

	// MaxSliceLength is the maximum length for slices.
	// Longer slices will be truncated with an ellipsis rune as last element.
	// A value <= 0 will disable truncating.
	MaxSliceLength int

	// PrintFuncFor can be used to customize the printing of a value
	// by returning a PrintFunc for a reflect.Value.
	// Returning nil from the function will disable custom printing
	// for the value including checking for and using the Printable interface.
	// If set this function will be used instead of PrintFuncForPrintable,
	// so call PrintFuncForPrintable within a PrintFuncFor function to check for
	// and use the Printable interface.
	// If not set, the PrintFuncForPrintable function will be used instead.
	//
	// Example: Adapting fmt.Stringer types
	//
	//	printer := pretty.DefaultPrinter.WithPrintFuncFor(func(v reflect.Value) pretty.PrintFunc {
	//	    stringer, ok := v.Interface().(fmt.Stringer)
	//	    if !ok && v.CanAddr() {
	//	        stringer, ok = v.Addr().Interface().(fmt.Stringer)
	//	    }
	//	    if ok {
	//	        return func(w io.Writer) (int, error) {
	//	            return fmt.Fprint(w, stringer.String())
	//	        }
	//	    }
	//	    return pretty.PrintFuncForPrintable(v) // Use default
	//	})
	PrintFuncFor func(reflect.Value) PrintFunc
}

// WithPrintFuncFor returns a new Printer with the passed PrintFuncFor
// function set, leaving all other fields unchanged.
func (p *Printer) WithPrintFuncFor(printFuncFor func(reflect.Value) PrintFunc) *Printer {
	return &Printer{
		MaxStringLength: p.MaxStringLength,
		MaxErrorLength:  p.MaxErrorLength,
		MaxSliceLength:  p.MaxSliceLength,
		PrintFuncFor:    printFuncFor,
	}
}

// Println pretty prints a value to os.Stdout followed by a newline.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func (p *Printer) Println(value any, indent ...string) (n int, err error) {
	return p.Fprintln(os.Stdout, value, indent...)
}

// Print pretty prints a value to os.Stdout.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func (p *Printer) Print(value any, indent ...string) (n int, err error) {
	return p.Fprint(os.Stdout, value, indent...)
}

// Fprint pretty prints a value to a io.Writer.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func (p *Printer) Fprint(w io.Writer, value any, indent ...string) (n int, err error) {
	_, n, err = p.fprintIndent(w, value, indent)
	return n, err
}

// Fprintln pretty prints a value to a io.Writer followed by a newline.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func (p *Printer) Fprintln(w io.Writer, value any, indent ...string) (n int, err error) {
	endsWithNewLine, n, err := p.fprintIndent(w, value, indent)
	if err != nil {
		return n, err
	}
	if !endsWithNewLine {
		var n2 int
		n2, err = w.Write([]byte{'\n'})
		n += n2
	}
	return n, err
}

// Sprint pretty prints a value to a string.
// The optional indent parameter controls indentation:
//   - No arguments: prints on a single line without indentation
//   - One argument: uses indent[0] as indent string for nested structures
//   - Two+ arguments: uses indent[0] as indent string and indent[1:] concatenated as line prefix
func (p *Printer) Sprint(value any, indent ...string) string {
	var b strings.Builder
	_, _, err := p.fprintIndent(&b, value, indent)
	if err != nil {
		return fmt.Sprintf("error printing value: %v", err)
	}
	return b.String()
}

type visitedPtrs map[uintptr]struct{}

func (v visitedPtrs) visit(ptr uintptr) (visited bool) {
	if _, visited = v[ptr]; visited {
		return true
	}
	v[ptr] = struct{}{}
	return false
}

// fprintIndent pretty prints a value to w with optional indentation.
// The indent parameter controls indentation behavior:
//   - Empty slice: prints value without indentation on a single line
//   - One element: uses indent[0] as indent string for nested structures
//   - Two+ elements: uses indent[0] as indent string and indent[1:] concatenated as line prefix
//
// Returns true if the output ends with a newline character.
// For nil values with line prefix, prints the prefix before "nil".
// For non-nil values with indentation, the output is formatted with newlines
// and proper indentation for readability.
func (p *Printer) fprintIndent(w io.Writer, value any, indent []string) (endsWithNewLine bool, n int, err error) {
	switch {
	case value == nil:
		if len(indent) > 1 {
			n, err = fmt.Fprint(w, strings.Join(indent[1:], ""))
			if err != nil {
				return false, n, err
			}
		}
		n2, err := fmt.Fprint(w, "nil")
		return false, n + n2, err

	case len(indent) == 0:
		n, err = p.fprint(w, reflect.ValueOf(value), make(visitedPtrs))
		return false, n, err

	default:
		var buf bytes.Buffer
		_, err = p.fprint(&buf, reflect.ValueOf(value), make(visitedPtrs))
		if err != nil {
			return false, 0, err
		}
		in := Indent(buf.Bytes(), indent[0], indent[1:]...)
		n, err = w.Write(in)
		return len(in) > 0 && in[len(in)-1] == '\n', n, err
	}
}

// #nosec G104 -- We don't check for errors writing to w
func (p *Printer) fprint(w io.Writer, v reflect.Value, ptrs visitedPtrs) (int, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Fprint(w, "nil")
		}
		ptr := v.Pointer()
		if ptrs.visit(ptr) {
			return fmt.Fprint(w, CircularRef)
		}
		defer delete(ptrs, ptr)
	}

	printFuncFor := p.PrintFuncFor
	if printFuncFor == nil {
		printFuncFor = PrintFuncForPrintable
	}
	if printFunc := printFuncFor(v); printFunc != nil {
		return printFunc(w)
	}

	nullable, ok := tryCastReflectValue[Nullable](v)
	if ok && nullable.IsNull() {
		return fmt.Fprint(w, "null")
	}

	ctx, ok := tryCastReflectValue[context.Context](v)
	if ok {
		var inner string
		if ctx.Err() != nil {
			inner = "Err:" + Sprint(ctx.Err().Error())
		}
		return fmt.Fprintf(w, "Context{%s}", inner)
	}

	for (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && !v.IsNil() {
		v = v.Elem()
	}
	t := v.Type()

	switch t {
	case typeOfTime:
		return fmt.Fprintf(w, "Time(`%s`)", v.Interface())

	case typeOfDuration:
		return fmt.Fprintf(w, "Duration(`%s`)", v.Interface())
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Interface:
		// Pointers and interfaces were dereferenced above, so only nil left as possibility
		if !v.IsNil() {
			panic("expected nil")
		}
		return fmt.Fprint(w, "nil")

	case reflect.String:
		err, ok := tryCastReflectValue[error](v)
		if ok {
			return fmt.Fprintf(w, "error(%s)", quoteString(err, p.MaxErrorLength))
		}
		return fmt.Fprint(w, quoteString(v.Interface(), p.MaxStringLength))

	case reflect.Bool:
		return fmt.Fprint(w, v.Interface())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Fprint(w, v.Interface())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Fprint(w, v.Interface())

	case reflect.Uintptr:
		return fmt.Fprintf(w, "%#v", v.Interface())

	case reflect.Float32, reflect.Float64:
		return fmt.Fprint(w, v.Interface())

	case reflect.Complex64, reflect.Complex128:
		return fmt.Fprint(w, v.Interface())

	case reflect.Array:
		n, err := w.Write([]byte{'['})
		if err != nil {
			return n, err
		}
		for i := range v.Len() {
			if i > 0 {
				n2, err := w.Write([]byte{','})
				n += n2
				if err != nil {
					return n, err
				}
			}
			n2, err := p.fprint(w, v.Index(i), ptrs)
			n += n2
			if err != nil {
				return n, err
			}
		}
		n2, err := w.Write([]byte{']'})
		return n + n2, err

	case reflect.Slice:
		if v.IsNil() {
			return fmt.Fprint(w, "nil")
		}
		ptr := v.Pointer()
		if ptrs.visit(ptr) {
			return fmt.Fprint(w, CircularRef)
		}
		defer delete(ptrs, ptr)
		switch t.Elem() {
		case typeOfByte:
			b := v.Bytes()
			if bytes.IndexByte(b, 0) == -1 && utf8.Valid(b) {
				// Bytes are valid UTF-8 without zero, assume it's a string
				return fmt.Fprint(w, quoteString(b, p.MaxStringLength))
			}
			if p.MaxSliceLength > 0 && len(b) > p.MaxSliceLength {
				return fmt.Fprintf(w, "[]byte{len(%d)}", len(b))
			}
		case typeOfRune:
			runes := v.Interface().([]rune)
			valid := true
			for _, r := range runes {
				valid = r > 0 && utf8.ValidRune(r)
				if !valid {
					break
				}
			}
			if valid {
				return fmt.Fprint(w, quoteString(string(runes), p.MaxStringLength))
			}
		}
		n, err := w.Write([]byte{'['})
		if err != nil {
			return n, err
		}
		for i := range v.Len() {
			if i > 0 {
				n2, err := w.Write([]byte{','})
				n += n2
				if err != nil {
					return n, err
				}
			}
			if p.MaxSliceLength > 0 && i >= p.MaxSliceLength {
				n2, err := fmt.Fprint(w, "…")
				n += n2
				if err != nil {
					return n, err
				}
				break
			}
			n2, err := p.fprint(w, v.Index(i), ptrs)
			n += n2
			if err != nil {
				return n, err
			}
		}
		n2, err := w.Write([]byte{']'})
		return n + n2, err

	case reflect.Map:
		if v.IsNil() {
			return fmt.Fprint(w, "nil")
		}
		ptr := v.Pointer()
		if ptrs.visit(ptr) {
			return fmt.Fprint(w, CircularRef)
		}
		defer delete(ptrs, ptr)
		n, err := fmt.Fprintf(w, "%s{", t.Name())
		if err != nil {
			return n, err
		}
		mapKeys := v.MapKeys()
		p.sortReflectValues(mapKeys, t.Key(), ptrs)
		for i, key := range mapKeys {
			if i > 0 {
				n2, err := w.Write([]byte{';'})
				n += n2
				if err != nil {
					return n, err
				}
			}
			n2, err := p.fprint(w, key, ptrs)
			n += n2
			if err != nil {
				return n, err
			}
			n2, err = w.Write([]byte{':'})
			n += n2
			if err != nil {
				return n, err
			}
			n2, err = p.fprint(w, v.MapIndex(key), ptrs)
			n += n2
			if err != nil {
				return n, err
			}
		}
		n2, err := w.Write([]byte{'}'})
		return n + n2, err

	case reflect.Struct:
		hasExportedFields := false
		for i := range t.NumField() {
			if token.IsExported(t.Field(i).Name) {
				hasExportedFields = true
				break
			}
		}
		if !hasExportedFields {
			err, ok := tryCastReflectValue[error](v)
			if ok {
				return fmt.Fprintf(w, "error(%s)", quoteString(err, p.MaxErrorLength))
			}
		}

		n, err := fmt.Fprintf(w, "%s{", t.Name())
		if err != nil {
			return n, err
		}
		first := true
		for i := range t.NumField() {
			f := t.Field(i)
			if !token.IsExported(f.Name) {
				continue
			}
			if first {
				first = false
			} else {
				n2, err := w.Write([]byte{';'})
				n += n2
				if err != nil {
					return n, err
				}
			}
			if !f.Anonymous {
				n2, err := fmt.Fprintf(w, "%s:", f.Name)
				n += n2
				if err != nil {
					return n, err
				}
			}
			n2, err := p.fprint(w, v.Field(i), ptrs)
			n += n2
			if err != nil {
				return n, err
			}
		}
		n2, err := w.Write([]byte{'}'})
		return n + n2, err

	case reflect.Chan, reflect.Func:
		if v.IsNil() {
			return fmt.Fprint(w, "nil")
		}
		return fmt.Fprint(w, t.String())

	case reflect.UnsafePointer:
		if v.IsNil() {
			return fmt.Fprint(w, "nil")
		}
		return fmt.Fprint(w, v.Interface())

	default:
		return 0, fmt.Errorf("unexpected reflect.Kind: %s", t.Kind())
	}
}

// sortReflectValues sorts a slice of reflected values.
// All values must be of the same type passed as valType.
// The < operator is used if the value's type supports it,
// else the pretty printed string representations are compared.
func (p *Printer) sortReflectValues(vals []reflect.Value, valType reflect.Type, ptrs visitedPtrs) {
	if len(vals) < 2 {
		return
	}
	switch valType.Kind() {
	case reflect.String:
		sort.Slice(vals, func(i, j int) bool {
			return vals[i].String() < vals[j].String()
		})
		return
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sort.Slice(vals, func(i, j int) bool {
			return vals[i].Int() < vals[j].Int()
		})
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sort.Slice(vals, func(i, j int) bool {
			return vals[i].Uint() < vals[j].Uint()
		})
		return
	case reflect.Float32, reflect.Float64:
		sort.Slice(vals, func(i, j int) bool {
			return vals[i].Float() < vals[j].Float()
		})
		return
	case reflect.Bool:
		sort.Slice(vals, func(i, j int) bool {
			return !vals[i].Bool() && vals[j].Bool()
		})
		return
	case reflect.Slice:
		if valType.Elem().Kind() == reflect.Uint8 {
			sort.Slice(vals, func(i, j int) bool {
				return bytes.Compare(vals[i].Bytes(), vals[j].Bytes()) < 0
			})
			return
		}
	}
	sort.Slice(vals, func(i, j int) bool {
		var ip, jp strings.Builder
		p.fprint(&ip, vals[i], ptrs)
		p.fprint(&jp, vals[j], ptrs)
		return ip.String() < jp.String()
	})
}

func quoteString(s any, maxLen int) string {
	q := fmt.Sprintf("%#q", s)
	if maxLen > 0 && len(q)-2 > maxLen {
		// Compare byte length as first approximation,
		// but then count runes to trim at a valid rune byte position
		for i := range q {
			if i > maxLen {
				q = q[:i] + "…" + q[len(q)-1:]
				break
			}
		}
	}
	// Replace double quotes
	if q[0] == '"' && q[len(q)-1] == '"' {
		q = "`" + q[1:len(q)-1] + "`"
	}
	return q
}

type countingWriter struct {
	writer io.Writer

	n   int
	err error
}

func newCountingWriter(writer io.Writer) *countingWriter {
	return &countingWriter{writer: writer}
}

func (w *countingWriter) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	w.n += n
	w.err = errors.Join(w.err, err)
	return n, err
}

func (w *countingWriter) Result() (n int, err error) {
	return w.n, w.err
}
