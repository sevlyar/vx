// Package vx validates Go values against schemas built from small,
// composable checks (Gt, Len, Item, Structure, ...).
//
// A Schema is a function that binds to a reflect.Type and returns a
// BoundSchema for that type:
//
//	schema := Gt(0)
//	validate := schema.BindTypeOf(0)
//	err := validate.Check(-1) // error: value should be greater than 0
//
// Binding happens either early, against a concrete type known when the
// schema is built (BindTypeOf, or a struct field's own type inside
// Structure), or late, against any value's type at Check time (BindAny).
// A schema bound early fails fast: a mismatched type panics when the
// schema is built, not when data is checked. A schema bound late can only
// report a type mismatch as a regular error, since any value may arrive
// at Check time.
//
// Checks compose: Structure and Field describe a struct's fields, Item and
// Len reach into a slice or its length, AllOf/AnyOf/OneOf combine several
// schemas. A failure is a *CompoundCheckError (or, for a check with no
// nested schema, a *CheckError); SchemaPath and DataPath reconstruct
// exactly which rule fired and where in the data, e.g. "Field(Name).Item.
// Len.Gt(3)" and "Name[2]".
package vx

import (
	"reflect"
)

// BoundSchema validates a single reflect.Value, already bound to a known type.
type BoundSchema func(v reflect.Value) error

// Check adapts BoundSchema to validate a Go value directly.
func (v BoundSchema) Check(val any) error {
	return Validate(val, v)
}

// Validate runs validate against value, indirecting through a pointer if value is one.
func Validate(value any, validate BoundSchema) error {
	v := reflect.ValueOf(value)
	return validate(reflect.Indirect(v))
}

// Schema builds a BoundSchema for t. It must check t is a type the schema
// supports and panic if it is not, unless it defers that check until
// validation time (see BindAny).
type Schema func(t reflect.Type) BoundSchema

// BindAny binds schema late, against values of any type. A type mismatch
// the schema would otherwise panic on at bind time is instead returned as
// an error from the resulting BoundSchema.
func (schema Schema) BindAny() BoundSchema {
	var v any
	return schema(reflect.TypeOf(&v).Elem())
}

// BindTypeOf binds schema early, against the type of v (or, if v is a
// pointer, the type it points to). A value of a different kind passed to
// the resulting BoundSchema panics.
func (schema Schema) BindTypeOf(v any) BoundSchema {
	buildType := reflect.TypeOf(v)
	if buildType.Kind() == reflect.Ptr {
		buildType = buildType.Elem()
	}
	validate := schema(buildType)
	return func(v reflect.Value) error {
		if v.Kind() != buildType.Kind() {
			panic("invalid type")
		}
		return validate(v)
	}
}

// BindTypeCheck builds a Schema for a check with a fixed validate func that
// requires a specific kind of type (check reports a non-nil error if t is
// unsupported). It is the building block behind most of this package's
// checks (Gt, Format, Structure, ...) and is exported for writing new ones:
// check runs, and panics on failure, at bind time when the type is known
// early; otherwise it runs on every Check call and returns its error
// instead of panicking. See CanCheckEarly.
func BindTypeCheck(check func(t reflect.Type) error, validate BoundSchema) Schema {
	return func(t reflect.Type) BoundSchema {
		if CanCheckEarly(t) {
			if err := check(t); err != nil {
				panic(err)
			}
			return validate
		}
		return func(v reflect.Value) error {
			if err := check(v.Type()); err != nil {
				return err
			}
			return validate(v)
		}
	}
}

// CanCheckEarly reports whether t is a concrete type a Schema can check
// immediately when it is built, rather than having to wait for a value at
// Check time. It is false only for the interface type BindAny binds to.
func CanCheckEarly(t reflect.Type) bool {
	return t.Kind() != reflect.Interface
}
