package labeler

import (
	"testing"
)

func TestFieldFlag_Has(t *testing.T) {
	tests := []struct {
		name     string
		flag     FieldFlag
		check    FieldFlag
		expected bool
	}{
		{"Has title", FieldTitle, FieldTitle, true},
		{"Has body", FieldBody, FieldBody, true},
		{"Has both", AllFieldFlags, FieldTitle, true},
		{"Has both (body)", AllFieldFlags, FieldBody, true},
		{"Does not have", FieldTitle, FieldBody, false},
		{"Zero flag", 0, FieldTitle, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flag.Has(tt.check)
			if got != tt.expected {
				t.Errorf("FieldFlag(%d).Has(%d) = %v; want %v", tt.flag, tt.check, got, tt.expected)
			}
		})
	}
}

func TestFieldFlag_OrDefault(t *testing.T) {
	tests := []struct {
		name     string
		flag     FieldFlag
		expected FieldFlag
	}{
		{"Non-zero returns self", FieldTitle, FieldTitle},
		{"AllFieldFlags returns self", AllFieldFlags, AllFieldFlags},
		{"Zero returns AllFieldFlags", 0, AllFieldFlags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flag.OrDefault()
			if got != tt.expected {
				t.Errorf("FieldFlag(%d).OrDefault() = %d; want %d", tt.flag, got, tt.expected)
			}
		})
	}
}

func TestParseFieldFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected FieldFlag
	}{
		{"Empty slice", []string{}, 0},
		{"Single title", []string{"title"}, FieldTitle},
		{"Single body", []string{"body"}, FieldBody},
		{"Single description (alternate)", []string{"description"}, FieldBody},
		{"Both fields", []string{"title", "body"}, AllFieldFlags},
		{"Duplicate fields", []string{"title", "title"}, FieldTitle},
		{"Unknown field", []string{"unknown"}, 0},
		{"Mixed known and unknown", []string{"title", "unknown"}, FieldTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFieldFlags(tt.input)
			if got != tt.expected {
				t.Errorf("ParseFieldFlags(%v) = %d; want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAllFieldFlagsValue(t *testing.T) {
	expected := FieldTitle | FieldBody
	// Ensure AllFieldFlags is set to the correct value of OR'd flags.
	if AllFieldFlags != expected {
		t.Errorf("AllFieldFlags = %d; want %d", AllFieldFlags, expected)
	}
}
