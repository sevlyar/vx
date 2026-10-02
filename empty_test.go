package vx

import (
	"reflect"
	"testing"
)

type emptyCase struct {
	input     any
	inputType reflect.Type
	wantEmpty bool
	wantPanic any
}

var emptyCases = map[string]emptyCase{
	"zero-string":      {input: "", wantEmpty: true},
	"non-zero-string":  {input: "NonEmpty", wantEmpty: false},
	"zero-int":         {input: 0, wantEmpty: true},
	"non-zero-int":     {input: 42, wantEmpty: false},
	"zero-bool":        {input: false, wantEmpty: true},
	"non-zero-bool":    {input: true, wantEmpty: false},
	"zero-struct":      {input: struct{}{}, wantEmpty: true},
	"non-zero-struct":  {input: struct{ int }{42}, wantEmpty: false},
	"zero-slice":       {input: []int{}, wantEmpty: true},
	"non-zero-slice":   {input: []int{42}, wantEmpty: false},
	"non-zero-pointer": {input: new(int), wantEmpty: false},
	"zero-pointer": {
		input:     (*int)(nil),
		inputType: reflect.TypeFor[*int](),
		wantPanic: "is not a valid kind",
	},
	"zero-interface": {
		input:     error(nil),
		inputType: reflect.TypeFor[any](),
		wantPanic: "is not a valid kind",
	},
	"interface-value": {
		input:     struct{ int }{42},
		inputType: reflect.TypeFor[any](),
		wantEmpty: false,
	},
}

func checkEmpty(schema Schema, tc emptyCase) error {
	inputType := tc.inputType
	if inputType == nil {
		inputType = reflect.ValueOf(tc.input).Type()
	}
	return schema(inputType).Check(tc.input)
}

func expectPanic(t *testing.T, want any, f func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Errorf("panic = %v, want %v", got, want)
		}
	}()
	f()
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	for name, tc := range emptyCases {
		t.Run(name, func(t *testing.T) {
			if tc.wantPanic != nil {
				expectPanic(t, tc.wantPanic, func() { _ = checkEmpty(Empty, tc) })
				return
			}
			err := checkEmpty(Empty, tc)
			switch {
			case tc.wantEmpty && err != nil:
				t.Errorf("Empty() error = %v, want nil", err)
			case !tc.wantEmpty && err != ErrNonEmptyValue:
				t.Errorf("Empty() error = %v, want %v", err, ErrNonEmptyValue)
			}
		})
	}
}

func TestNonEmpty(t *testing.T) {
	t.Parallel()
	for name, tc := range emptyCases {
		t.Run(name, func(t *testing.T) {
			if tc.wantPanic != nil {
				expectPanic(t, tc.wantPanic, func() { _ = checkEmpty(NonEmpty, tc) })
				return
			}
			err := checkEmpty(NonEmpty, tc)
			switch {
			case !tc.wantEmpty && err != nil:
				t.Errorf("NonEmpty() error = %v, want nil", err)
			case tc.wantEmpty && err != ErrEmptyValue:
				t.Errorf("NonEmpty() error = %v, want %v", err, ErrEmptyValue)
			}
		})
	}
}
