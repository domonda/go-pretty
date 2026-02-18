package pretty

import (
	"math"
	"testing"
)

func TestAsJSON(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		data, err := asJSON(struct{ Name string }{Name: "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "{\n  \"Name\": \"test\"\n}"
		if got := string(data); got != want {
			t.Errorf("asJSON() = %q, want %q", got, want)
		}
	})

	t.Run("byte slice", func(t *testing.T) {
		data, err := asJSON([]byte(`hello`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `"aGVsbG8="` // base64 encoded
		if got := string(data); got != want {
			t.Errorf("asJSON() = %q, want %q", got, want)
		}
	})

	t.Run("custom indent", func(t *testing.T) {
		data, err := asJSON(map[string]int{"a": 1}, "\t")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "{\n\t\"a\": 1\n}"
		if got := string(data); got != want {
			t.Errorf("asJSON() = %q, want %q", got, want)
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		_, err := asJSON(math.NaN())
		if err == nil {
			t.Fatal("expected error for NaN, got nil")
		}
	})
}
