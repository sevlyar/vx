package vx

import (
	"fmt"
	"reflect"
)

// UnknownValueError is returned by In when the value is not one of KnownValues.
type UnknownValueError struct {
	CheckError
	KnownValues []any
}

// In checks the value is one of allowed. A slice argument is unwound into
// its elements.
//
// allowed is captured once, into a map, when In is called, not re-read
// afterward: pass a small, fixed set of immutable values (e.g. enum
// constants), not a large or later-mutated collection.
func In(allowed ...any) Schema {
	uniq := uniqueValues(allowed)
	set := make(map[any]struct{}, len(uniq))
	for _, v := range uniq {
		set[v] = struct{}{}
	}
	sv := summarizeValues(uniq)
	err := &UnknownValueError{
		CheckError: CheckError{
			CheckName: "In",
			Param:     sv,
			Msg:       "value is not one of " + sv,
		},
		KnownValues: uniq,
	}
	return func(t reflect.Type) BoundSchema {
		if CanCheckEarly(t) {
			if !t.Comparable() {
				panic("value type must be comparable")
			}
			for _, v := range uniq {
				if reflect.TypeOf(v) != t {
					panic("the value type doesn't match of known values")
				}
			}
		}
		return func(v reflect.Value) error {
			if _, ok := set[v.Interface()]; ok {
				return nil
			}
			return err
		}
	}
}

// maxListedValues caps how many of In's allowed values are rendered into
// its error message; KnownValues always holds the full set.
const maxListedValues = 3

// summarizeValues renders values for an error message, truncating with "..."
// past maxListedValues so a large allowed set can't blow up the message size.
func summarizeValues(values []any) string {
	if len(values) <= maxListedValues {
		return fmt.Sprint(values)
	}
	shown := append(append([]any{}, values[:maxListedValues]...), "...")
	return fmt.Sprint(shown)
}

// uniqueValues unwinds any slice in allowed into its elements and dedupes the result.
func uniqueValues(allowed []any) []any {
	list := make([]any, 0, len(allowed))
	for _, elem := range allowed {
		v := reflect.ValueOf(elem)
		if v.Kind() != reflect.Slice {
			list = append(list, elem)
			continue
		}
		for i := 0; i < v.Len(); i++ {
			list = append(list, v.Index(i).Interface())
		}
	}
	set := make(map[any]struct{}, len(list))
	uniq := make([]any, 0, len(list))
	for _, elem := range list {
		if _, ok := set[elem]; !ok {
			set[elem] = struct{}{}
			uniq = append(uniq, elem)
		}
	}
	return uniq
}
