package vx

import (
	"errors"
	"fmt"
	"reflect"
)

// StructOption configures a Structure schema; see Field and AllowUncheckedFields.
type StructOption func(*structureValidator)

// Structure checks a struct's fields according to the given options.
// ptr must point to the struct being described; it is used only to resolve
// field types and addresses, not validated itself.
//
// Structure panics if any field has no Field option attached to it; use
// AllowUncheckedFields to opt specific fields out of this requirement.
func Structure(ptr any, opts ...StructOption) Schema {
	validator := newStructureValidator(ptr)
	for _, opt := range opts {
		opt(validator)
	}
	return validator.makeSchema()
}

// Field declares how to validate the struct field addressed by fieldPtr.
// fieldPtr must point to a field of the struct passed to Structure.
func Field(fieldPtr any, schema Schema) StructOption {
	return func(validator *structureValidator) {
		validator.addFieldValidator(fieldPtr, schema)
	}
}

// AllowUncheckedFields opts the addressed fields out of Structure's
// requirement that every field have a Field option.
func AllowUncheckedFields(fieldPtrs ...any) StructOption {
	return func(v *structureValidator) {
		for _, ptr := range fieldPtrs {
			v.markFieldChecked(ptr)
		}
	}
}

type structureValidator struct {
	typ     reflect.Type
	base    uintptr // address of the structure, to resolve field pointers by offset
	fields  map[uintptr]*reflect.StructField
	checked map[uintptr]struct{}

	validators []fieldValidator
}

type fieldValidator struct {
	field    *reflect.StructField
	validate BoundSchema
}

func newStructureValidator(ptr any) *structureValidator {
	str := reflect.ValueOf(ptr)
	if str.Kind() != reflect.Ptr || str.Elem().Kind() != reflect.Struct {
		panic("the ptr should be pointer on a structure")
	}
	typ := str.Type().Elem()

	fields := make(map[uintptr]*reflect.StructField, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fields[f.Offset] = &f
	}
	return &structureValidator{
		typ:     typ,
		base:    str.Pointer(),
		fields:  fields,
		checked: make(map[uintptr]struct{}, len(fields)),
	}
}

func (v *structureValidator) addFieldValidator(fieldPtr any, schema Schema) {
	field := v.markFieldChecked(fieldPtr)
	if schema == nil {
		panic("schema required")
	}
	v.validators = append(v.validators, fieldValidator{
		field:    field,
		validate: schema(field.Type),
	})
}

func (v *structureValidator) markFieldChecked(fieldPtr any) *reflect.StructField {
	ptrVal := reflect.ValueOf(fieldPtr)
	if ptrVal.Kind() != reflect.Ptr {
		panic("the fieldPtr should be pointer")
	}
	offset := ptrVal.Pointer() - v.base
	field := v.fields[offset]
	if field == nil {
		panic("the fieldPtr does not keep address of a field value of the structure")
	}
	v.checked[offset] = struct{}{}
	return field
}

func (v *structureValidator) makeSchema() Schema {
	// Walk fields in declaration order so the panic is deterministic.
	for i := 0; i < v.typ.NumField(); i++ {
		f := v.typ.Field(i)
		if _, ok := v.checked[f.Offset]; !ok {
			panic(fmt.Sprintf("validation of '%s' field is missing", f.Name))
		}
	}
	return BindTypeCheck(v.checkType, v.validate)
}

func (v *structureValidator) checkType(t reflect.Type) error {
	if v.typ != t {
		return errors.New("invalid value type, original structure type required")
	}
	return nil
}

func (v *structureValidator) validate(strVal reflect.Value) error {
	for _, fv := range v.validators {
		fieldVal := strVal.FieldByIndex(fv.field.Index)
		if err := fv.validate(fieldVal); err != nil {
			return fieldError(fv.field.Name, err)
		}
	}
	return nil
}
