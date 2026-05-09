package tui_test

import (
	"testing"

	"github.com/gjtiquia/vimance/internal/tui"
)

func TestValidateAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid cents", "12.50", true},
		{"valid whole", "100", true},
		{"valid zero", "0", true},
		{"valid decimal", "0.99", true},
		{"empty", "", false},
		{"letters", "abc", false},
		{"three decimals", "12.123", false},
		{"trailing dot", "12.", false},
		{"leading dot", ".50", false},
		{"comma stripped", "1,000", true},
		{"negative", "-50", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tui.ValidateAmount(tc.input)
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseAmountToCents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect int64
		valid  bool
	}{
		{"12.50", "12.50", 1250, true},
		{"100", "100", 10000, true},
		{"0", "0", 0, true},
		{"0.99", "0.99", 99, true},
		{"1.00", "1.00", 100, true},
		{"10.0", "10.0", 1000, true},
		{"invalid", "abc", 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cents, err := tui.ParseAmountToCents(tc.input)
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("expected error")
			}
			if tc.valid && cents != tc.expect {
				t.Errorf("expected %d cents, got %d", tc.expect, cents)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		year, month, day     string
		valid                bool
	}{
		{"valid", "2026", "05", "09", true},
		{"empty year", "", "05", "09", false},
		{"empty month", "2026", "", "09", false},
		{"empty day", "2026", "05", "", false},
		{"month 13", "2026", "13", "01", false},
		{"day 32", "2026", "01", "32", false},
		{"feb 29 non-leap", "2025", "02", "29", false},
		{"feb 29 leap", "2024", "02", "29", true},
		{"garbage", "abc", "def", "ghi", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tui.ValidateDate(tc.year, tc.month, tc.day)
			if tc.valid && err != nil {
				t.Errorf("expected valid, got: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	t.Parallel()
	got := tui.FormatDate("2026", "05", "09")
	want := "2026-05-09"
	if got != want {
		t.Errorf("FormatDate = %q, want %q", got, want)
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 5, "hello"},
		{"hello world", 5, "hello..."},
		{"hi", 5, "hi"},
		{"", 3, ""},
	}
	for _, tc := range tests {
		got := tui.Truncate(tc.input, tc.max)
		if got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
		}
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()
	var errors tui.ValidationErrors
	if errors.HasErrors() {
		t.Error("empty errors should not have errors")
	}

	errors = append(errors, tui.ValidationError{Field: "date", Message: "bad date"})
	if !errors.HasErrors() {
		t.Error("non-empty errors should have errors")
	}

	if msg := errors.Get("date"); msg != "bad date" {
		t.Errorf("Get('date') = %q, want 'bad date'", msg)
	}
	if msg := errors.Get("nonexistent"); msg != "" {
		t.Errorf("Get('nonexistent') = %q, want ''", msg)
	}
}
