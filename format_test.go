package vx

import (
	"strings"
	"testing"
)

func isUpper(s string) bool { return s == strings.ToUpper(s) }

func TestFormat(t *testing.T) {
	schema := Format("Upper", isUpper).BindAny()

	if err := schema.Check("ABC"); err != nil {
		t.Errorf("Check(ABC) error = %v, want nil", err)
	}
	if err := schema.Check("abc"); err == nil {
		t.Error("Check(abc) error = nil, want error")
	}
}

func TestFormat_WrongType_PanicsWhenBoundEarly(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("BindTypeOf(0) did not panic")
		}
	}()
	Format("Upper", isUpper).BindTypeOf(0)
}

func TestFormat_WrongType_ErrorsWhenBoundLate(t *testing.T) {
	schema := Format("Upper", isUpper).BindAny()

	if err := schema.Check(42); err == nil {
		t.Error("Check(42) error = nil, want error")
	}
}

func TestRegexpFormat(t *testing.T) {
	schema := RegexpFormat(`^\d+$`).BindAny()

	if err := schema.Check("123"); err != nil {
		t.Errorf("Check(123) error = %v, want nil", err)
	}
	if err := schema.Check("abc"); err == nil {
		t.Error("Check(abc) error = nil, want error")
	}
}

func TestRegexpFormat_InvalidPattern_Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("RegexpFormat(\"[\") did not panic")
		}
	}()
	RegexpFormat("[")
}

func TestIsPrintableLine(t *testing.T) {
	cases := map[string]bool{
		"hello world":  true,
		"line1\nline2": false,
		"tab\there":    false,
		"\x00":         false,
		"line1" + string(lineSeparator) + "line2":      false,
		"line1" + string(paragraphSeparator) + "line2": false,
	}
	for s, want := range cases {
		if got := isPrintableLine(s); got != want {
			t.Errorf("isPrintableLine(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestIsPrintableText(t *testing.T) {
	cases := map[string]bool{
		"hello world":  true,
		"line1\nline2": true,
		"tab\there":    true,
		"\x00":         false,
		"line1" + string(lineSeparator) + "line2":      true,
		"line1" + string(paragraphSeparator) + "line2": true,
	}
	for s, want := range cases {
		if got := isPrintableText(s); got != want {
			t.Errorf("isPrintableText(%q) = %v, want %v", s, got, want)
		}
	}
}
