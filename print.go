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
	"bytes"
	"context"
	"fmt"
	"go/token"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"
)

var (
	// MaxStringLength is the maximum length for escaped strings.
	// Longer strings will be truncated with an ellipsis rune at the end.
	// A value <= 0 will disable truncating.
	MaxStringLength = 1000

	// MaxErrorLength is the maximum length for escaped errors.
	// Longer errors will be truncated with an ellipsis rune at the end.
	// A value <= 0 will disable truncating.
	MaxErrorLength = 10000

	// MaxSliceLength is the maximum length for slices.
	// Longer slices will be truncated with an ellipsis rune as last element.
	// A value <= 0 will disable truncating.
	MaxSliceLength = 1000

	typeOfByte = reflect.TypeOf(byte(0))
	// typeOfError = reflect.TypeOf((*error)(nil)).Elem()
	// typeOfSortInterface = reflect.TypeOf((*sort.Interface)(nil)).Elem()
)

// Printer can be implemented to customize the pretty printing of a type.
type Printer interface {
	// PrettyPrint the implementation's data
	PrettyPrint(io.Writer)
}

// Println pretty prints a value to os.Stderr followed by a newline
func Println(value interface{}, indent ...string) {
	fprintIndent(os.Stderr, value, indent)
	os.Stderr.Write([]byte{'\n'})
}

// Print pretty prints a value to os.Stderr
func Print(value interface{}, indent ...string) {
	fprintIndent(os.Stderr, value, indent)
}

// Fprint pretty prints a value to a io.Writer
func Fprint(w io.Writer, value interface{}, indent ...string) {
	fprintIndent(w, value, indent)
}

// Fprint pretty prints a value to a io.Writer followed by a newline
func Fprintln(w io.Writer, value interface{}, indent ...string) {
	fprintIndent(w, value, indent)
	os.Stderr.Write([]byte{'\n'})
}

// Sprint pretty prints a value to a string
func Sprint(value interface{}, indent ...string) string {
	var b strings.Builder
	fprintIndent(&b, value, indent)
	return b.String()
}

func fprintIndent(w io.Writer, value interface{}, indent []string) {
	switch {
	case value == nil:
		if len(indent) > 1 {
			fmt.Fprint(w, indent[1])
		}
		fmt.Fprint(w, "nil")

	case len(indent) == 0:
		fprint(w, reflect.ValueOf(value))

	default:
		var buf bytes.Buffer
		fprint(&buf, reflect.ValueOf(value))
		indented := make([]byte, 0, buf.Len()+256)
		linePrefx := ""
		if len(indent) > 1 {
			linePrefx = indent[1]
		}
		indented = AppendIndent(indented, buf.Bytes(), indent[0], linePrefx)
		w.Write(indented)
	}
}

func fprint(w io.Writer, v reflect.Value) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		fmt.Fprint(w, "nil")
		return
	}

	printer, _ := v.Interface().(Printer)
	if printer == nil && v.CanAddr() {
		printer, _ = v.Addr().Interface().(Printer)
	}
	if printer != nil {
		printer.PrettyPrint(w)
		return
	}

	goStringer, _ := v.Interface().(fmt.GoStringer)
	if goStringer == nil && v.CanAddr() {
		goStringer, _ = v.Addr().Interface().(fmt.GoStringer)
	}
	if goStringer != nil {
		fmt.Fprint(w, goStringer.GoString())
		return
	}

	ctx, _ := v.Interface().(context.Context)
	if ctx == nil && v.CanAddr() {
		ctx, _ = v.Addr().Interface().(context.Context)
	}
	if ctx != nil {
		var inner string
		if ctx.Err() != nil {
			inner = "Err:" + Sprint(ctx.Err().Error())
		}
		fmt.Fprintf(w, "Context{%s}", inner)
		return
	}

	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	t := v.Type()

	switch t.Kind() {
	case reflect.Ptr:
		// Pointers were dereferenced above, so only nil left as possibility
		fmt.Fprint(w, "nil")

	case reflect.String:
		err, _ := v.Interface().(error)
		if err == nil && v.CanAddr() {
			err, _ = v.Addr().Interface().(error)
		}
		if err != nil {
			fmt.Fprintf(w, "error(%s)", quoteString(err, MaxErrorLength))
			return
		}
		fmt.Fprint(w, quoteString(v.Interface(), MaxStringLength))

	case reflect.Bool:
		fmt.Fprint(w, v.Interface())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(w, v.Interface())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprint(w, v.Interface())

	case reflect.Uintptr:
		fmt.Fprintf(w, "%#v", v.Interface())

	case reflect.Float32, reflect.Float64:
		fmt.Fprint(w, v.Interface())

	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(w, v.Interface())

	case reflect.Array:
		w.Write([]byte{'['})
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w.Write([]byte{','})
			}
			fprint(w, v.Index(i))
		}
		w.Write([]byte{']'})

	case reflect.Slice:
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		if t.Elem() == typeOfByte && utf8.Valid(v.Bytes()) {
			fmt.Fprint(w, quoteString(v.Interface(), MaxStringLength))
			return
		}
		w.Write([]byte{'['})
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w.Write([]byte{','})
			}
			if MaxSliceLength > 0 && i >= MaxSliceLength {
				fmt.Fprint(w, "…")
				break
			}
			fprint(w, v.Index(i))
		}
		w.Write([]byte{']'})

	case reflect.Map:
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		// TODO sort map if possible
		// if t.Key().Implements(typeOfSortInterface) {
		// 	// TODO Need to make a temp sorted copy
		// }
		// switch t.Key().Kind() {
		// case reflect.String:
		// case reflect.Slice, reflect.Array:
		// }
		fmt.Fprintf(w, "%s{", t.Name())
		for i, iter := 0, v.MapRange(); iter.Next(); i++ {
			if i > 0 {
				w.Write([]byte{','})
			}
			fprint(w, iter.Key())
			w.Write([]byte{':'})
			fprint(w, iter.Value())
		}
		w.Write([]byte{'}'})

	case reflect.Struct:
		hasExportedFields := false
		for i := 0; i < t.NumField(); i++ {
			if token.IsExported(t.Field(i).Name) {
				hasExportedFields = true
				break
			}
		}
		if !hasExportedFields {
			err, _ := v.Interface().(error)
			if err == nil && v.CanAddr() {
				err, _ = v.Addr().Interface().(error)
			}
			if err != nil {
				fmt.Fprintf(w, "error(%s)", quoteString(err, MaxErrorLength))
				return
			}
		}

		fmt.Fprintf(w, "%s{", t.Name())
		first := true
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !token.IsExported(f.Name) {
				continue
			}
			if first {
				first = false
			} else {
				w.Write([]byte{','})
			}
			if !f.Anonymous {
				fmt.Fprintf(w, "%s:", f.Name)
			}
			fprint(w, v.Field(i))
		}
		w.Write([]byte{'}'})

	case reflect.Chan, reflect.Func, reflect.Interface:
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		fmt.Fprint(w, t.String())

	case reflect.UnsafePointer:
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		fmt.Fprint(w, v.Interface())

	default:
		panic("invalid kind: " + t.Kind().String())
	}
}

func quoteString(s interface{}, maxLen int) string {
	q := fmt.Sprintf("%#q", s)
	if maxLen > 0 && len(q)-2 > maxLen {
		// Compare byte length as first approximation,
		// but then count runes to trim at avalid rune byte position
		for i := range q {
			if i > maxLen {
				return q[:i] + "…" + q[len(q)-1:]
			}
		}
	}
	return q
}

func AppendIndent(dest, pretty []byte, linePrefix, indent string) []byte {
	panic("todo")

	// const (
	// 	stateNormal = iota
	// 	stateRawString
	// 	stateEscString
	// 	stateEscStringRune
	// )
	// var (
	// 	state         = stateNormal
	// 	newLineIndent = append([]byte{'\n'}, linePrefix...)
	// 	current       = 0
	// 	r             rune
	// 	size          int
	// )
	// for i := 0; i < len(pretty) && r != utf8.RuneError; i += size {
	// 	r, size = utf8.DecodeRune(pretty[i:])
	// 	if r == utf8.RuneError {
	// 		break
	// 	}
	// 	switch state {
	// 	case stateNormal:
	// 		if current == 0 {
	// 			dest = append(dest, linePrefix...)
	// 		}
	// 		next := i + size
	// 		dest = append(dest, pretty[current:next]...)
	// 		switch r {
	// 		case ':':
	// 			dest = append(dest, ' ')
	// 		case '{':
	// 			newLineIndent = append(newLineIndent, indent...)
	// 			dest = append(dest, newLineIndent...)
	// 		case '}':
	// 			newLineIndent = newLineIndent[:len(newLineIndent)-len(indent)]
	// 			dest = append(dest, newLineIndent...)
	// 			dest = append(dest, '}')
	// 			dest = append(dest, newLineIndent...)
	// 		case '`':
	// 			state = stateRawString
	// 		case '"':
	// 			state = stateEscString
	// 		}
	// 		current = next

	// 	case stateRawString:
	// 		dest = append(dest, pretty[i:i+size]...)
	// 		if r == '`' {
	// 			state = stateNormal
	// 		}

	// 	case stateEscString:

	// 	case stateEscStringRune:

	// 	}
	// }

	// return dest
}
