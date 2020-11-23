package pretty

var (
	// MaxStringLength is the maximum length for escaped strings.
	// Longer strings will be truncated with an ellipsis rune at the end.
	// A value <= 0 will disable truncating.
	MaxStringLength = 200

	// MaxErrorLength is the maximum length for escaped errors.
	// Longer errors will be truncated with an ellipsis rune at the end.
	// A value <= 0 will disable truncating.
	MaxErrorLength = 2000

	// MaxSliceLength is the maximum length for slices.
	// Longer slices will be truncated with an ellipsis rune as last element.
	// A value <= 0 will disable truncating.
	MaxSliceLength = 20
)

const (
	// CircularRef is a replacement token CIRCULAR_REF
	// that will be printed instad of a circular data reference.
	CircularRef = "CIRCULAR_REF"
)
