package pretty

import (
	"reflect"
	"time"
)

var (
	// DefaultPrinter is used by the package level print functions
	DefaultPrinter = Printer{
		MaxStringLength: 200,
		MaxErrorLength:  2000,
		MaxSliceLength:  20,
	}

	// CircularRef is a replacement token (default "CIRCULAR_REF")
	// that will be printed instead of a circular data reference.
	CircularRef = "CIRCULAR_REF"
)

var (
	typeOfByte     = reflect.TypeFor[byte]()
	typeOfRune     = reflect.TypeFor[rune]()
	typeOfTime     = reflect.TypeFor[time.Time]()
	typeOfDuration = reflect.TypeFor[time.Duration]()
)
