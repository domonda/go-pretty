package pretty

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

type ErrorStruct struct {
	X   int
	Y   int
	err string
}

func (e ErrorStruct) Error() string { return e.err }

type StringXer string

func (s StringXer) PrettyPrint(w io.Writer) { fmt.Fprintf(w, "'%sX'", s) }

func TestSprint(t *testing.T) {
	type Parent struct {
		Map map[int]string
	}
	type Struct struct {
		Parent
		Int        int
		unexported bool
		Str        string
		Sub        struct {
			Map map[string]struct{}
		}
	}
	type UUID [16]byte

	var (
		nilUUID  UUID
		nilError error
	)

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: `nil`},
		{name: "nilError", value: nilError, want: `nil`},
		{name: "an error", value: errors.New("An\nError"), want: "error(`An\\nError`)"},
		{name: "ErrorStruct", value: ErrorStruct{X: 1, Y: 2, err: "xxx"}, want: `ErrorStruct{X:1;Y:2}`},
		{name: "ErrorStructPtr", value: &ErrorStruct{X: 1, Y: 2, err: "xxx"}, want: `ErrorStruct{X:1;Y:2}`},
		{name: "ErrorStruct as error", value: (error)(ErrorStruct{X: 1, Y: 2, err: "xxx"}), want: `ErrorStruct{X:1;Y:2}`},
		{name: "Printer", value: StringXer("hello"), want: `'helloX'`},
		{name: "nil Printer", value: (*StringXer)(nil), want: `nil`},
		{name: "nilPtr", value: (*int)(nil), want: `nil`},
		{name: "empty string", value: "", want: "``"},
		{name: "multiline string", value: "Hello\n\"World!\"", want: "`Hello\\n\\\"World!\\\"`"},
		{name: "byte string", value: []byte("Hello World"), want: "`Hello World`"},
		{name: "rune string", value: []rune("Hello World"), want: "`Hello World`"},
		{name: "int", value: 666, want: `666`},
		{name: "struct no sub-init", value: Struct{Int: -1, Str: "xxx"}, want: "Struct{Parent{Map:nil};Int:-1;Str:`xxx`;Sub:{Map:nil}}"},
		{name: "struct sub-init", value: Struct{Sub: struct{ Map map[string]struct{} }{Map: map[string]struct{}{"key": {}}}}, want: "Struct{Parent{Map:nil};Int:0;Str:``;Sub:{Map:{`key`:{}}}}"},
		{name: "string slice", value: []string{"", `"quoted"`, "hello\nworld"}, want: "[``,`\"quoted\"`,`hello\\nworld`]"},
		{name: "Nil UUID", value: nilUUID, want: `[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]`},
		{name: "true", value: true, want: `true`},
		{name: "false", value: false, want: `false`},
		{name: "Int", value: int(-123), want: `-123`},
		{name: "Int8", value: int8(-123), want: `-123`},
		{name: "Int16", value: int16(-123), want: `-123`},
		{name: "Int32", value: int32(-123), want: `-123`},
		{name: "Int64", value: int64(-123), want: `-123`},
		{name: "Uint", value: uint(123), want: `123`},
		{name: "Uint8", value: uint8(123), want: `123`},
		{name: "Uint16", value: uint16(123), want: `123`},
		{name: "Uint32", value: uint32(123), want: `123`},
		{name: "Uint64", value: uint64(123), want: `123`},
		{name: "Uintptr", value: uintptr(0xf0), want: `0xf0`},
		{name: "Float32", value: float32(-1.23), want: `-1.23`},
		{name: "Float64", value: float64(-1.23), want: `-1.23`},
		{name: "Complex64", value: complex64(1 - 2i), want: `(1-2i)`},
		{name: "Complex128", value: complex128(1 - 2i), want: `(1-2i)`},
		{name: "chan int", value: make(chan int), want: `chan int`},
		{name: "<-chan int", value: make(<-chan int), want: `<-chan int`},
		{name: "chan<- int", value: make(chan<- int), want: `chan<- int`},
		{name: "(chan int)(nil)", value: (chan int)(nil), want: `nil`},
		{name: "(<-chan int)(nil)", value: (<-chan int)(nil), want: `nil`},
		{name: "(chan<- int)(nil)", value: (chan<- int)(nil), want: `nil`},
		{name: "func(int) error", value: func(int) error { panic("") }, want: `func(int) error`},
		{name: "func() (<-chan time.Time, error)", value: func() (<-chan time.Time, error) { panic("") }, want: `func() (<-chan time.Time, error)`},
		{name: "(func(int) error)(nil)", value: (func(int) error)(nil), want: `nil`},
		{name: "nil UnsafePointer", value: unsafe.Pointer(nil), want: `nil`},
		{name: "nil byte slice", value: []byte(nil), want: "nil"},
		{name: "empty byte slice", value: []byte{}, want: "``"},
		{name: "1 byte slice", value: make([]byte, 1), want: "[0]"},
		{name: "MaxSliceLength byte slice", value: make([]byte, DefaultPrinter.MaxSliceLength), want: "[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]"},
		{name: "big byte slice", value: make([]byte, DefaultPrinter.MaxSliceLength+1), want: "[]byte{len(21)}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sprint(tt.value); got != tt.want {
				t.Errorf("Sprint() = %v, want %v", got, tt.want)
			}
		})
	}

	// Save and restore DefaultPrinter after modifying it
	savedPrinter := DefaultPrinter
	t.Cleanup(func() { DefaultPrinter = savedPrinter })

	DefaultPrinter.MaxStringLength = 5
	t.Run(fmt.Sprintf("MaxStringLength_%d", DefaultPrinter.MaxStringLength), func(t *testing.T) {
		want := "`Hello…`"
		if got := Sprint("Hello World"); got != want {
			t.Errorf("Sprint() = %v, want %v", got, want)
		}
	})
	DefaultPrinter.MaxStringLength = 1
	t.Run(fmt.Sprintf("MaxStringLength_%d", DefaultPrinter.MaxStringLength), func(t *testing.T) {
		want := "`H…`"
		if got := Sprint("Hello World"); got != want {
			t.Errorf("Sprint() = %v, want %v", got, want)
		}
	})
	DefaultPrinter.MaxStringLength = 0
	t.Run(fmt.Sprintf("MaxStringLength_%d", DefaultPrinter.MaxStringLength), func(t *testing.T) {
		want := "`Hello World`"
		if got := Sprint("Hello World"); got != want {
			t.Errorf("Sprint() = %v, want %v", got, want)
		}
	})
	DefaultPrinter.MaxStringLength = -1
	t.Run(fmt.Sprintf("MaxStringLength_%d", DefaultPrinter.MaxStringLength), func(t *testing.T) {
		want := "`Hello World`"
		if got := Sprint("Hello World"); got != want {
			t.Errorf("Sprint() = %v, want %v", got, want)
		}
	})

	DefaultPrinter.MaxErrorLength = 5
	t.Run("MaxErrorLength", func(t *testing.T) {
		want := "error(`An\\nE…`)"
		if got := Sprint(errors.New("An\nError")); got != want {
			t.Errorf("Sprint() = %v, want %v", got, want)
		}
	})

	DefaultPrinter.MaxSliceLength = 5
	t.Run("MaxSliceLength", func(t *testing.T) {
		want := `[1,2,3,4,5,…]`
		if got := Sprint([]int{1, 2, 3, 4, 5, 6, 7}); got != want {
			t.Errorf("Sprint() = %v, want %v", got, want)
		}
	})
}

func TestCircularData(t *testing.T) {
	type Struct struct {
		Int int
		Ref *Struct
	}
	circStruct := &Struct{Int: 666}
	circStruct.Ref = circStruct

	circStructsNotNested := [...]*Struct{circStruct, circStruct}

	circSlice := make([]any, 1)
	circSlice[0] = circSlice

	// Test for indirect circular reference bug with maps
	// Map A -> Slice B -> Map A (should detect circular reference)
	circMap := make(map[string]any)
	indirectSlice := make([]any, 1)
	circMap["slice"] = indirectSlice
	indirectSlice[0] = circMap

	// Test for indirect circular reference bug with nested maps
	// Map A -> Map B -> Map A (should detect circular reference)
	mapA := make(map[string]any)
	mapB := make(map[string]any)
	mapA["b"] = mapB
	mapB["a"] = mapA

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "circStruct",
			value: circStruct,
			want:  `Struct{Int:666;Ref:CIRCULAR_REF}`,
		},
		{
			name:  "circStructsNotNested",
			value: circStructsNotNested,
			want:  `[Struct{Int:666;Ref:CIRCULAR_REF},Struct{Int:666;Ref:CIRCULAR_REF}]`,
		},
		{
			name:  "circSlice",
			value: circSlice,
			want:  `[CIRCULAR_REF]`,
		},
		{
			name:  "circMapViaSlice",
			value: circMap,
			want:  `{` + "`slice`" + `:[CIRCULAR_REF]}`,
		},
		{
			name:  "circMapViaMap",
			value: mapA,
			want:  `{` + "`b`" + `:{` + "`a`" + `:CIRCULAR_REF}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sprint(tt.value); got != tt.want {
				t.Errorf("Sprint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpecialTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "time.Time",
			value: time.Date(2020, 07, 14, 12, 9, 34, 0, time.UTC),
			want:  "Time(`2020-07-14 12:09:34 +0000 UTC`)",
		},
		{
			name:  "time.Duration",
			value: time.Duration(time.Hour*11 + time.Minute*59 + time.Millisecond*666),
			want:  "Duration(`11h59m0.666s`)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sprint(tt.value); got != tt.want {
				t.Errorf("Sprint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFprint(t *testing.T) {
	var b strings.Builder
	n, err := Fprint(&b, "hello")
	if err != nil {
		t.Fatalf("Fprint returned error: %v", err)
	}
	want := "`hello`"
	if got := b.String(); got != want {
		t.Errorf("Fprint wrote %q, want %q", got, want)
	}
	if n != len(want) {
		t.Errorf("Fprint returned n=%d, want %d", n, len(want))
	}
}

func TestFprintln(t *testing.T) {
	var b strings.Builder
	n, err := Fprintln(&b, 42)
	if err != nil {
		t.Fatalf("Fprintln returned error: %v", err)
	}
	want := "42\n"
	if got := b.String(); got != want {
		t.Errorf("Fprintln wrote %q, want %q", got, want)
	}
	if n != len(want) {
		t.Errorf("Fprintln returned n=%d, want %d", n, len(want))
	}
}

func TestFprintNil(t *testing.T) {
	var b strings.Builder
	n, err := Fprint(&b, nil)
	if err != nil {
		t.Fatalf("Fprint returned error: %v", err)
	}
	want := "nil"
	if got := b.String(); got != want {
		t.Errorf("Fprint wrote %q, want %q", got, want)
	}
	if n != len(want) {
		t.Errorf("Fprint returned n=%d, want %d", n, len(want))
	}
}

type nullableValue struct {
	isNull bool
}

func (n nullableValue) IsNull() bool { return n.isNull }

func TestNullable(t *testing.T) {
	t.Run("null value", func(t *testing.T) {
		got := Sprint(nullableValue{isNull: true})
		if got != "null" {
			t.Errorf("Sprint(null) = %q, want %q", got, "null")
		}
	})
	t.Run("non-null value", func(t *testing.T) {
		got := Sprint(nullableValue{isNull: false})
		want := "nullableValue{}"
		if got != want {
			t.Errorf("Sprint(non-null) = %q, want %q", got, want)
		}
	})
}

func TestContext(t *testing.T) {
	t.Run("background context", func(t *testing.T) {
		got := Sprint(context.Background())
		if got != "Context{}" {
			t.Errorf("Sprint(context.Background()) = %q, want %q", got, "Context{}")
		}
	})
	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		got := Sprint(ctx)
		want := "Context{Err:`context canceled`}"
		if got != want {
			t.Errorf("Sprint(cancelled context) = %q, want %q", got, want)
		}
	})
}

func TestWithPrintFuncFor(t *testing.T) {
	printer := DefaultPrinter.WithPrintFuncFor(func(v reflect.Value) PrintFunc {
		if v.Kind() == reflect.String {
			return func(w io.Writer) (int, error) {
				return fmt.Fprintf(w, "CUSTOM(%s)", v.String())
			}
		}
		return PrintFuncForPrintable(v)
	})

	t.Run("custom func used", func(t *testing.T) {
		got := printer.Sprint("test")
		want := "CUSTOM(test)"
		if got != want {
			t.Errorf("Sprint() = %q, want %q", got, want)
		}
	})
	t.Run("default for non-string", func(t *testing.T) {
		got := printer.Sprint(42)
		if got != "42" {
			t.Errorf("Sprint() = %q, want %q", got, "42")
		}
	})
	t.Run("preserves MaxStringLength", func(t *testing.T) {
		if printer.MaxStringLength != DefaultPrinter.MaxStringLength {
			t.Errorf("MaxStringLength = %d, want %d", printer.MaxStringLength, DefaultPrinter.MaxStringLength)
		}
	})
}

func TestQuoteString(t *testing.T) {
	tests := []struct {
		name   string
		s      any
		maxLen int
		want   string
	}{
		{name: "simple", s: "hello", maxLen: 0, want: "`hello`"},
		{name: "empty", s: "", maxLen: 0, want: "``"},
		{name: "newline", s: "a\nb", maxLen: 0, want: "`a\\nb`"},
		{name: "backtick", s: "a`b", maxLen: 0, want: "`a`b`"},
		{name: "truncate", s: "Hello World", maxLen: 5, want: "`Hello…`"},
		{name: "no truncate at limit", s: "Hello", maxLen: 5, want: "`Hello`"},
		{name: "maxLen disabled", s: "Hello World", maxLen: -1, want: "`Hello World`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteString(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("quoteString(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func ExamplePrintln() {
	type Parent struct {
		Map map[int]string
	}

	type Struct struct {
		Parent
		Int        int
		unexported bool
		Str        string
		Sub        struct {
			Map map[string]string
		}
	}

	value := &Struct{
		Sub: struct{ Map map[string]string }{
			Map: map[string]string{
				"key": "value",
				// Note that the resulting `Multi\nLine` is not a valid Go string.
				// Double quotes are avoided for better readability of
				// pretty printed strings in JSON.
				"Multi\nLine": "true",
			},
		},
	}

	Println(value)
	Println(value, "  ")
	Println(value, "  ", "    ")

	// Output:
	// Struct{Parent{Map:nil};Int:0;Str:``;Sub:{Map:{`Multi\nLine`:`true`;`key`:`value`}}}
	// Struct{
	//   Parent{
	//     Map: nil
	//   }
	//   Int: 0
	//   Str: ``
	//   Sub: {
	//     Map: {
	//       `Multi\nLine`: `true`
	//       `key`: `value`
	//     }
	//   }
	// }
	//     Struct{
	//       Parent{
	//         Map: nil
	//       }
	//       Int: 0
	//       Str: ``
	//       Sub: {
	//         Map: {
	//           `Multi\nLine`: `true`
	//           `key`: `value`
	//         }
	//       }
	//     }
}
