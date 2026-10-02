package vx

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"unicode"
)

// Format checks a string value against isValid, requiring the bound type to be string.
func Format(name string, isValid func(str string) bool) Schema {
	err := newCheckError("Format", name, fmt.Sprintf("value format is not %s", name))
	return newFormatCheck(err, isValid)
}

// RegexpFormat checks a string value matches pattern, requiring the bound type to be string.
func RegexpFormat(pattern string) Schema {
	re := regexp.MustCompile(pattern)
	err := newCheckError("RegexpFormat", pattern, fmt.Sprintf("value does not match %s", pattern))
	return newFormatCheck(err, re.MatchString)
}

func newFormatCheck(err error, isValid func(str string) bool) Schema {
	return BindTypeCheck(mustBeString, func(v reflect.Value) error {
		if !isValid(v.String()) {
			return err
		}
		return nil
	})
}

func mustBeString(t reflect.Type) error {
	if t.Kind() != reflect.String {
		return errors.New("type must be string")
	}
	return nil
}

// lineSeparator and paragraphSeparator are written as rune(0x...), not a
// '\u...' rune literal: gofmt unescapes a printable \u escape into the
// literal rune, leaving an invisible character sitting in the source.
const (
	lineSeparator      = rune(0x2028) // U+2028 LINE SEPARATOR
	paragraphSeparator = rune(0x2029) // U+2029 PARAGRAPH SEPARATOR
)

// PrintableLine checks a string has no non-printable runes, including line breaks.
var PrintableLine = Format("PrintableLine", isPrintableLine)

func isPrintableLine(value string) bool {
	for _, r := range value {
		// Neither IsPrint (control-like category) nor the White_Space
		// property IsSpace/unicode.Space both use rejects these two.
		if r == lineSeparator || r == paragraphSeparator {
			return false
		}
		if !unicode.IsPrint(r) && !unicode.Is(unicode.Space, r) {
			return false
		}
	}
	return true
}

// PrintableText is like PrintableLine but allows line breaks and other whitespace control runes.
var PrintableText = Format("PrintableText", isPrintableText)

func isPrintableText(value string) bool {
	for _, r := range value {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
