package service

import "testing"

func TestFormatCents(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{99, "0.99"},
		{100, "1.00"},
		{150, "1.50"},
		{1000, "10.00"},
		{1500, "15.00"},
		{123456, "1234.56"},
		{-1, "-0.01"},
		{-100, "-1.00"},
		{-150, "-1.50"},
		{-123456, "-1234.56"},
	}

	for _, tt := range tests {
		got := FormatCents(tt.cents)
		if got != tt.expected {
			t.Errorf("FormatCents(%d) = %q, want %q", tt.cents, got, tt.expected)
		}
	}
}
