package pretty

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
)

const (
	// CircularRef is a replacement token CIRCULAR_REF
	// that will be printed instad of a circular data reference.
	CircularRef = "CIRCULAR_REF"
)
