package pretty

import (
	"errors"
	"fmt"
	"io"
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
	t.Run("MaxErrorLength", func(t *testing.T) {
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
