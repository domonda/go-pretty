package pretty

import (
	"reflect"
	"time"
)

// DefaultPrinter is used by the package level print functions
var DefaultPrinter = Printer{
	MaxStringLength: 200,
	MaxErrorLength:  2000,
	MaxSliceLength:  20,
}

// CircularRef is a replacement token CIRCULAR_REF
// that will be printed instad of a circular data reference.
const CircularRef = "CIRCULAR_REF"

var (
	typeOfByte     = reflect.TypeOf(byte(0))
	typeOfRune     = reflect.TypeOf(rune(0))
	typeOfTime     = reflect.TypeOf(time.Time{})
	typeOfDuration = reflect.TypeOf(time.Duration(0))
)
