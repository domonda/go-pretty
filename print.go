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

var byteType = reflect.TypeOf(byte(0))

// Printer can be implemented to customize the pretty printing of a type.
type Printer interface {
	// PrettyPrint the implementation's data
	PrettyPrint(io.Writer)
}

// Println pretty prints a value to os.Stdout followed by a newline
func Println(value interface{}, indent ...string) {
	endsWithNewLine := fprintIndent(os.Stdout, value, indent)
	if !endsWithNewLine {
		os.Stdout.Write([]byte{'\n'})
	}
}

// Print pretty prints a value to os.Stdout
func Print(value interface{}, indent ...string) {
	fprintIndent(os.Stdout, value, indent)
}

// Fprint pretty prints a value to a io.Writer
func Fprint(w io.Writer, value interface{}, indent ...string) {
	fprintIndent(w, value, indent)
}

// Fprint pretty prints a value to a io.Writer followed by a newline
func Fprintln(w io.Writer, value interface{}, indent ...string) {
	endsWithNewLine := fprintIndent(w, value, indent)
	if !endsWithNewLine {
		os.Stdout.Write([]byte{'\n'})
	}
}

// Sprint pretty prints a value to a string
func Sprint(value interface{}, indent ...string) string {
	var b strings.Builder
	fprintIndent(&b, value, indent)
	return b.String()
}

func fprintIndent(w io.Writer, value interface{}, indent []string) (endsWithNewLine bool) {
	switch {
	case value == nil:
		if len(indent) > 1 {
			fmt.Fprint(w, indent[1])
		}
		fmt.Fprint(w, "nil")
		return false

	case len(indent) == 0:
		fprint(w, reflect.ValueOf(value), make(visitedPtrs))
		return false

	default:
		var buf bytes.Buffer
		fprint(&buf, reflect.ValueOf(value), make(visitedPtrs))
		in := Indent(buf.Bytes(), indent[0], indent[1:]...)
		w.Write(in)
		return len(in) > 0 && in[len(in)-1] == '\n'
	}
}

type visitedPtrs map[uintptr]struct{}

func (v visitedPtrs) visit(ptr uintptr) (visited bool) {
	if _, visited = v[ptr]; visited {
		return true
	}
	v[ptr] = struct{}{}
	return false
}

func fprint(w io.Writer, v reflect.Value, ptrs visitedPtrs) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		ptr := v.Pointer()
		if ptrs.visit(ptr) {
			fmt.Fprint(w, CircularRef)
			return
		}
		defer delete(ptrs, ptr)
	}

	printer, _ := v.Interface().(Printer)
	if printer == nil && v.CanAddr() {
		printer, _ = v.Addr().Interface().(Printer)
	}
	if printer != nil {
		printer.PrettyPrint(w)
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

	for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
		v = v.Elem()
	}
	t := v.Type()

	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		// Pointers and interfaces were dereferenced above, so only nil left as possibility
		if !v.IsNil() {
			panic("expected nil")
		}
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
			fprint(w, v.Index(i), ptrs)
		}
		w.Write([]byte{']'})

	case reflect.Slice:
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		ptr := v.Pointer()
		if ptrs.visit(ptr) {
			fmt.Fprint(w, CircularRef)
			return
		}
		defer delete(ptrs, ptr)
		if t.Elem() == byteType && utf8.Valid(v.Bytes()) {
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
			fprint(w, v.Index(i), ptrs)
		}
		w.Write([]byte{']'})

	case reflect.Map:
		if v.IsNil() {
			fmt.Fprint(w, "nil")
			return
		}
		ptr := v.Pointer()
		if ptrs.visit(ptr) {
			fmt.Fprint(w, CircularRef)
			return
		}
		defer delete(ptrs, ptr)
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
				w.Write([]byte{';'})
			}
			fprint(w, iter.Key(), ptrs)
			w.Write([]byte{':'})
			fprint(w, iter.Value(), ptrs)
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
				w.Write([]byte{';'})
			}
			if !f.Anonymous {
				fmt.Fprintf(w, "%s:", f.Name)
			}
			fprint(w, v.Field(i), ptrs)
		}
		w.Write([]byte{'}'})

	case reflect.Chan, reflect.Func:
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

// Indent pretty printed source using the passed indent string
// and an optional linePrefix used for every line in case of
// a multiple line result.
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
				newLineIndent = newLineIndent[:len(newLineIndent)-len(indent)]
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

	return result
}
