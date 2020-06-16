package pretty

import (
	"errors"
	"testing"
	"time"
	"unsafe"
)

type ErrorStruct struct {
	X   int
	Y   int
	err string
}

func (e ErrorStruct) Error() string {
	return e.err
}

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
		value interface{}
		want  string
	}{
		{name: "nil", value: nil, want: `nil`},
		{name: "nilError", value: nilError, want: `nil`},
		{name: "an error", value: errors.New("An\nError"), want: `error('An\nError')`},
		{name: "ErrorStruct", value: ErrorStruct{X: 1, Y: 2, err: "xxx"}, want: `ErrorStruct{X:1,Y:2}`},
		{name: "ErrorStructPtr", value: &ErrorStruct{X: 1, Y: 2, err: "xxx"}, want: `ErrorStruct{X:1,Y:2}`},
		{name: "ErrorStruct as error", value: (error)(ErrorStruct{X: 1, Y: 2, err: "xxx"}), want: `ErrorStruct{X:1,Y:2}`},
		{name: "nilPtr", value: (*int)(nil), want: `nil`},
		{name: "empty string", value: "", want: `''`},
		{name: "empty bytes string", value: []byte{}, want: `''`},
		{name: "multiline string", value: "Hello\n\tWorld!\n", want: `'Hello\n\tWorld!\n'`},
		{name: "bytes string", value: []byte("Hello World"), want: `'Hello World'`},
		{name: "int", value: 666, want: `666`},
		{name: "struct no sub-init", value: Struct{Int: -1, Str: "xxx"}, want: `Struct{Parent{Map:nil},Int:-1,Str:'xxx',Sub:{Map:nil}}`},
		{name: "struct sub-init", value: Struct{Sub: struct{ Map map[string]struct{} }{Map: map[string]struct{}{"key": {}}}}, want: `Struct{Parent{Map:nil},Int:0,Str:'',Sub:{Map:{'key':{}}}}`},
		{name: "string slice", value: []string{"", `"quoted"`, "hello\nworld"}, want: `['','"quoted"','hello\nworld']`},
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
		// {name: "Array", value: array(x), want: ``},
		// {name: "Interface", value: interface(x), want: ``},
		// {name: "Map", value: map(x), want: ``},
		// {name: "Ptr", value: ptr(x), want: ``},
		// {name: "Slice", value: slice(x), want: ``},

		// TODO
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sprint(tt.value); got != tt.want {
				t.Errorf("PrettySprint() = %v, want %v", got, tt.want)
			}
		})
	}
}
