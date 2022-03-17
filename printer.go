package pretty

import (
	"bytes"
	"context"
	"fmt"
	"go/token"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

// Printable can be implemented to customize the pretty printing of a type.
type Printable interface {
	// PrettyPrint the implementation's data
	PrettyPrint(io.Writer)
}

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
}

// Println pretty prints a value to os.Stdout followed by a newline
func (p *Printer) Println(value any, indent ...string) {
	endsWithNewLine := p.fprintIndent(os.Stdout, value, indent)
	if !endsWithNewLine {
		os.Stdout.Write([]byte{'\n'}) //#nosec G104
	}
}

// Print pretty prints a value to os.Stdout
func (p *Printer) Print(value any, indent ...string) {
	p.fprintIndent(os.Stdout, value, indent)
}

// Fprint pretty prints a value to a io.Writer
func (p *Printer) Fprint(w io.Writer, value any, indent ...string) {
	p.fprintIndent(w, value, indent)
}

// Fprint pretty prints a value to a io.Writer followed by a newline
func (p *Printer) Fprintln(w io.Writer, value any, indent ...string) {
	endsWithNewLine := p.fprintIndent(w, value, indent)
	if !endsWithNewLine {
		os.Stdout.Write([]byte{'\n'}) //#nosec G104
	}
}

// Sprint pretty prints a value to a string
func (p *Printer) Sprint(value any, indent ...string) string {
	var b strings.Builder
	p.fprintIndent(&b, value, indent)
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

func (p *Printer) fprintIndent(w io.Writer, value any, indent []string) (endsWithNewLine bool) {
	switch {
	case value == nil:
		if len(indent) > 1 {
			fmt.Fprint(w, indent[1])
		}
		fmt.Fprint(w, "nil")
		return false

	case len(indent) == 0:
		p.fprint(w, reflect.ValueOf(value), make(visitedPtrs))
		return false

	default:
		var buf bytes.Buffer
		p.fprint(&buf, reflect.ValueOf(value), make(visitedPtrs))
		in := Indent(buf.Bytes(), indent[0], indent[1:]...)
		w.Write(in) //#nosec G104
		return len(in) > 0 && in[len(in)-1] == '\n'
	}
}

//#nosec G104 -- We don't check for errors writing to w
func (p *Printer) fprint(w io.Writer, v reflect.Value, ptrs visitedPtrs) {
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

	printer, _ := v.Interface().(Printable)
	if printer == nil && v.CanAddr() {
		printer, _ = v.Addr().Interface().(Printable)
	}
	if printer != nil {
		printer.PrettyPrint(w)
		return
	}

	nullable, _ := v.Interface().(Nullable)
	if nullable == nil && v.CanAddr() {
		nullable, _ = v.Addr().Interface().(Nullable)
	}
	if nullable != nil && nullable.IsNull() {
		fmt.Fprint(w, "null")
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

	switch t {
	case typeOfTime:
		fmt.Fprintf(w, "Time(`%s`)", v.Interface())
		return
	case typeOfDuration:
		fmt.Fprintf(w, "Duration(`%s`)", v.Interface())
		return
	}

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
			fmt.Fprintf(w, "error(%s)", quoteString(err, p.MaxErrorLength))
			return
		}
		fmt.Fprint(w, quoteString(v.Interface(), p.MaxStringLength))

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
			p.fprint(w, v.Index(i), ptrs)
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
		switch t.Elem() {
		case typeOfByte:
			b := v.Bytes()
			if bytes.IndexByte(b, 0) == -1 && utf8.Valid(b) {
				// Bytes are valid UTF-8 without zero, assume it's a string
				fmt.Fprint(w, quoteString(b, p.MaxStringLength))
				return
			}
			if len(b) > p.MaxSliceLength {
				fmt.Fprintf(w, "[]byte{len(%d)}", len(b))
				return
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
				fmt.Fprint(w, quoteString(string(runes), p.MaxStringLength))
				return
			}
		}
		w.Write([]byte{'['})
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w.Write([]byte{','})
			}
			if p.MaxSliceLength > 0 && i >= p.MaxSliceLength {
				fmt.Fprint(w, "…")
				break
			}
			p.fprint(w, v.Index(i), ptrs)
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
		fmt.Fprintf(w, "%s{", t.Name())
		mapKeys := v.MapKeys()
		p.sortReflectValues(mapKeys, t.Key(), ptrs)
		for i, key := range mapKeys {
			if i > 0 {
				w.Write([]byte{';'})
			}
			p.fprint(w, key, ptrs)
			w.Write([]byte{':'})
			p.fprint(w, v.MapIndex(key), ptrs)
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
				fmt.Fprintf(w, "error(%s)", quoteString(err, p.MaxErrorLength))
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
			p.fprint(w, v.Field(i), ptrs)
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
			return vals[i].Bool() == false && vals[j].Bool() == true
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
		// but then count runes to trim at avalid rune byte position
		for i := range q {
			if i > maxLen {
				q = q[:i] + "…" + q[len(q)-1:]
				break
			}
		}
	}
	// Replace double qoutes
	if q[0] == '"' && q[len(q)-1] == '"' {
		q = "`" + q[1:len(q)-1] + "`"
	}
	return q
}
