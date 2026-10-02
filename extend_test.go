package vx_test

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/sevlyar/vx"
)

// MultipleOf is a worked example of writing a leaf check: one with no
// nested schema, built with BindTypeCheck and vx.CheckError. Compare with
// Gt in number.go, which follows the same shape.
func MultipleOf(n int) vx.Schema {
	mustBeInt := func(t reflect.Type) error {
		if t.Kind() != reflect.Int {
			return errors.New("type must be int")
		}
		return nil
	}
	err := &vx.CheckError{
		CheckName: "MultipleOf",
		Param:     fmt.Sprint(n), // rendered into SchemaPath, e.g. "MultipleOf(5)"
		Msg:       fmt.Sprintf("value must be a multiple of %d", n),
	}
	return vx.BindTypeCheck(mustBeInt, func(v reflect.Value) error {
		if v.Int()%int64(n) != 0 {
			return err
		}
		return nil
	})
}

func ExampleMultipleOf() {
	schema := MultipleOf(5).BindAny()
	fmt.Println(schema.Check(10))
	fmt.Println(schema.Check(7))
	// Output:
	// <nil>
	// value must be a multiple of 5
}

// Values is a worked example of writing a compound check: one that
// delegates to a nested schema reached through a transformation of the
// bound type, built with BindCompound. It validates every value of a map
// the way Item (in array.go) validates every element of a slice.
//
// Its DataPath step uses KeyElement, since a map key isn't a struct field
// (FieldElement) or a slice index (IndexElement); Key is the key's own
// rendering, chosen by the check, not by the vx package.
func Values(schema vx.Schema) vx.Schema {
	mustBeMap := func(t reflect.Type) error {
		if t.Kind() != reflect.Map {
			return errors.New("type must be map")
		}
		return nil
	}
	checkValues := func(validate vx.BoundSchema, m reflect.Value) error {
		iter := m.MapRange()
		for iter.Next() {
			if err := validate(iter.Value()); err != nil {
				return &vx.CompoundCheckError{
					CheckName: "Values",
					DataItem:  vx.PathElement{Kind: vx.KeyElement, Name: fmt.Sprintf("%q", iter.Key())},
					Cause:     err,
				}
			}
		}
		return nil
	}
	return vx.BindCompound(mustBeMap,
		func(t reflect.Type) vx.BoundSchema {
			validate := schema(t.Elem())
			return func(v reflect.Value) error { return checkValues(validate, v) }
		},
		func() vx.BoundSchema {
			validate := schema.BindAny()
			return func(v reflect.Value) error { return checkValues(validate, v) }
		},
	)
}

// ExampleValues validates a map with a single invalid value, so the result
// doesn't depend on Go's randomized map iteration order: every valid entry
// is silently skipped, and the one invalid entry is reported regardless of
// when it's visited.
func ExampleValues() {
	schema := Values(vx.Gt(0)).BindAny()
	err := schema.Check(map[string]int{"a": 2, "b": 4, "c": -1})

	fmt.Println(err.Error())
	// Output:
	// Values.Gt(0): invalid ["c"]: value should be greater than 0
}
