package vx

import "reflect"

var (
	// ErrNonEmptyValue is returned by Empty when the value is not empty.
	ErrNonEmptyValue = newCheckError("Empty", "", "value is not empty")
	// ErrEmptyValue is returned by NonEmpty when the value is empty.
	ErrEmptyValue = newCheckError("NonEmpty", "", "value is empty")
)

// Empty checks the value is the zero value for its type.
func Empty(t reflect.Type) BoundSchema {
	zero := reflect.Zero(t)
	return func(v reflect.Value) error {
		if !valueIsEmpty(zero, v) {
			return ErrNonEmptyValue
		}
		return nil
	}
}

// NonEmpty checks the value is not the zero value for its type.
func NonEmpty(t reflect.Type) BoundSchema {
	zero := reflect.Zero(t)
	return func(v reflect.Value) error {
		if valueIsEmpty(zero, v) {
			return ErrEmptyValue
		}
		return nil
	}
}

func valueIsEmpty(zero, v reflect.Value) bool {
	if v.Kind() == reflect.Interface {
		v = derefInterface(v)
		if !v.IsValid() {
			return true // a nil interface is empty
		}
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Pointer:
		return v.IsNil()
	case reflect.Invalid:
		panic("is not a valid kind")
	default:
		if zero.Kind() == reflect.Interface {
			zero = reflect.Zero(v.Type())
		}
		return reflect.DeepEqual(v.Interface(), zero.Interface())
	}
}
