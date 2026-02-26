package jalien

import (
	"testing"
)

func TestGetStringField(t *testing.T) {
	m := map[string]any{
		"s":   "value",
		"n":   float64(42),
		"nil": nil,
	}

	if got := getStringField(m, "s"); got != "value" {
		t.Fatalf("getStringField(s) = %q, want %q", got, "value")
	}

	if got := getStringField(m, "n"); got != "42" {
		t.Fatalf("getStringField(n) = %q, want %q", got, "42")
	}

	if got := getStringField(m, "missing"); got != "" {
		t.Fatalf("getStringField(missing) = %q, want empty", got)
	}
}

func TestGetUint64Field(t *testing.T) {
	m := map[string]any{
		"float": float64(123),
		"str":   "456",
	}

	got, err := getUint64Field(m, "float")
	if err != nil || got != 123 {
		t.Fatalf("getUint64Field(float) = (%d,%v), want (123,nil)", got, err)
	}

	got, err = getUint64Field(m, "str")
	if err != nil || got != 456 {
		t.Fatalf("getUint64Field(str) = (%d,%v), want (456,nil)", got, err)
	}

	if _, err = getUint64Field(m, "missing"); err == nil {
		t.Fatalf("getUint64Field(missing) expected error, got nil")
	}
}
