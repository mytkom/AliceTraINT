package utils

import (
	"testing"
)

func TestFormatSizePretty(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"Bytes case", 500, "500 bytes"},
		{"Kilobytes case", 1500, "1.46 KB"},
		{"Megabytes case", 5 * 1024 * 1024, "5.00 MB"},
		{"Gigabytes case", 3 * 1024 * 1024 * 1024, "3.00 GB"},
		{"Terabytes case", 2 * 1024 * 1024 * 1024 * 1024, "2.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSizePretty(tt.input)
			if result != tt.expected {
				t.Errorf("FormatSizePretty(%d) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
