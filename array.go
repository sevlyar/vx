package vx

import (
	"errors"
	"reflect"
	"unicode/utf8"
)

// Item checks every element of an array or slice against schema.
func Item(schema Schema) Schema {
	return BindCompound(shouldHaveItems,
		func(t reflect.Type) BoundSchema {
			validateItem := schema(t.Elem())
			return func(v reflect.Value) error { return checkItems(validateItem, v) }
		},
		func() BoundSchema {
			validateItem := schema.BindAny()
			return func(v reflect.Value) error { return checkItems(validateItem, v) }
		},
	)
}

func checkItems(validate BoundSchema, list reflect.Value) error {
	for i := 0; i < list.Len(); i++ {
		if err := validate(list.Index(i)); err != nil {
			return itemError(i, err)
		}
	}
	return nil
}

// Len checks the length of a string, array, slice or map against schema.
// A string's length is its rune count, not its byte count.
func Len(schema Schema) Schema {
	return BindCompound(shouldHaveLength,
		func(reflect.Type) BoundSchema {
			validateLen := schema(reflect.TypeOf(0))
			return func(v reflect.Value) error { return checkLen(validateLen, v) }
		},
		func() BoundSchema {
			validateLen := schema.BindAny()
			return func(v reflect.Value) error { return checkLen(validateLen, v) }
		},
	)
}

func checkLen(validateLen BoundSchema, v reflect.Value) error {
	if err := validateLen(reflect.ValueOf(getLen(v))); err != nil {
		return wrapError("Len", err)
	}
	return nil
}

// BindCompound builds a Schema for a check that delegates to a nested
// schema reached through some transformation of the bound type (Item's
// element type, Len's length, ...). It is exported for writing new checks
// of this shape: early builds the BoundSchema when the type is known at
// bind time (so the nested schema's own checks run, and panic, then too);
// late builds it once, against "any", when it is not (so the nested
// schema's checks defer to validation time and return errors instead of
// panicking). See BindTypeCheck for checks with no such transformation.
func BindCompound(
	check func(t reflect.Type) error,
	early func(t reflect.Type) BoundSchema,
	late func() BoundSchema,
) Schema {
	return func(t reflect.Type) BoundSchema {
		if CanCheckEarly(t) {
			if err := check(t); err != nil {
				panic(err)
			}
			return early(t)
		}
		validate := late()
		return func(v reflect.Value) error {
			v = derefInterface(v)
			if !v.IsValid() {
				return ErrNilValue
			}
			if err := check(v.Type()); err != nil {
				return err
			}
			return validate(v)
		}
	}
}

func shouldHaveLength(t reflect.Type) error {
	switch t.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return nil
	default:
		return errors.New("type must be string, array, slice or map")
	}
}

func shouldHaveItems(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return nil
	default:
		return errors.New("type must be array or slice")
	}
}

func getLen(v reflect.Value) int {
	if v.Kind() == reflect.String {
		return utf8.RuneCountInString(v.String())
	}
	return v.Len()
}
