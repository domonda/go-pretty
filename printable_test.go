package pretty

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// Test types with value receivers

type valPrintableWithResult struct{}

func (valPrintableWithResult) PrettyPrint(w io.Writer) (int, error) {
	return fmt.Fprint(w, "valPWR")
}

type valPrintable struct{}

func (valPrintable) PrettyPrint(w io.Writer) {
	fmt.Fprint(w, "valP")
}

type valStringer struct{}

func (valStringer) PrettyString() string {
	return "valPS"
}

// Test types with pointer receivers

type ptrPrintableWithResult struct{}

func (*ptrPrintableWithResult) PrettyPrint(w io.Writer) (int, error) {
	return fmt.Fprint(w, "ptrPWR")
}

type ptrPrintable struct{}

func (*ptrPrintable) PrettyPrint(w io.Writer) {
	fmt.Fprint(w, "ptrP")
}

type ptrStringer struct{}

func (*ptrStringer) PrettyString() string {
	return "ptrPS"
}

// Type implementing both PrintableWithResult and Stringer
// to test that PrintableWithResult takes priority

type printableWithResultAndStringer struct{}

func (printableWithResultAndStringer) PrettyPrint(w io.Writer) (int, error) {
	return fmt.Fprint(w, "PWR-wins")
}

func (printableWithResultAndStringer) PrettyString() string {
	return "PS-loses"
}

// Type implementing both Printable and Stringer
// to test that Printable takes priority over Stringer

type printableAndStringer struct{}

func (printableAndStringer) PrettyPrint(w io.Writer) {
	fmt.Fprint(w, "P-wins")
}

func (printableAndStringer) PrettyString() string {
	return "PS-loses"
}

// Type implementing no interfaces

type plainType struct{}

func TestPrintFuncForPrintable(t *testing.T) {
	tests := []struct {
		name string
		v    reflect.Value
		want string // empty means nil expected
	}{
		{
			name: "PrintableWithResult value receiver",
			v:    reflect.ValueOf(valPrintableWithResult{}),
			want: "valPWR",
		},
		{
			name: "PrintableWithResult pointer receiver via pointer",
			v:    reflect.ValueOf(&ptrPrintableWithResult{}),
			want: "ptrPWR",
		},
		{
			name: "PrintableWithResult pointer receiver via addressable value",
			v:    reflect.ValueOf([]ptrPrintableWithResult{{}}).Index(0),
			want: "ptrPWR",
		},
		{
			name: "Printable value receiver",
			v:    reflect.ValueOf(valPrintable{}),
			want: "valP",
		},
		{
			name: "Printable pointer receiver via pointer",
			v:    reflect.ValueOf(&ptrPrintable{}),
			want: "ptrP",
		},
		{
			name: "Printable pointer receiver via addressable value",
			v:    reflect.ValueOf([]ptrPrintable{{}}).Index(0),
			want: "ptrP",
		},
		{
			name: "Stringer value receiver",
			v:    reflect.ValueOf(valStringer{}),
			want: "valPS",
		},
		{
			name: "Stringer pointer receiver via pointer",
			v:    reflect.ValueOf(&ptrStringer{}),
			want: "ptrPS",
		},
		{
			name: "Stringer pointer receiver via addressable value",
			v:    reflect.ValueOf([]ptrStringer{{}}).Index(0),
			want: "ptrPS",
		},
		{
			name: "no interface returns nil",
			v:    reflect.ValueOf(plainType{}),
			want: "",
		},
		{
			name: "PrintableWithResult takes priority over Stringer",
			v:    reflect.ValueOf(printableWithResultAndStringer{}),
			want: "PWR-wins",
		},
		{
			name: "Printable takes priority over Stringer",
			v:    reflect.ValueOf(printableAndStringer{}),
			want: "P-wins",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := PrintFuncForPrintable(tt.v)
			if tt.want == "" {
				if fn != nil {
					t.Error("PrintFuncForPrintable() should return nil")
				}
				return
			}
			if fn == nil {
				t.Fatal("PrintFuncForPrintable() returned nil")
			}
			var b strings.Builder
			n, err := fn(&b)
			if err != nil {
				t.Fatalf("PrintFunc returned error: %v", err)
			}
			if got := b.String(); got != tt.want {
				t.Errorf("PrintFunc wrote %q, want %q", got, tt.want)
			}
			if n != b.Len() {
				t.Errorf("PrintFunc returned n=%d, but wrote %d bytes", n, b.Len())
			}
		})
	}
}
