package vx

import (
	"errors"
	"reflect"
)

// ErrMultipleSchemasMatch is the cause of a OneOf error when more than one schema matches.
var ErrMultipleSchemasMatch = errors.New("value matches multiple schemas")

type composeMode int

const (
	allOf composeMode = iota
	anyOf
	oneOf
)

// AllOf checks the value against every schema. The first failure is returned.
func AllOf(schema ...Schema) Schema { return compose("AllOf", allOf, schema) }

// AnyOf checks the value against each schema until one matches.
func AnyOf(schema ...Schema) Schema { return compose("AnyOf", anyOf, schema) }

// OneOf checks the value matches exactly one schema.
func OneOf(schema ...Schema) Schema { return compose("OneOf", oneOf, schema) }

func compose(name string, mode composeMode, schema []Schema) Schema {
	if len(schema) == 0 {
		panic("schema required")
	}
	if len(schema) == 1 {
		return schema[0]
	}
	return func(t reflect.Type) BoundSchema {
		validators := make([]BoundSchema, len(schema))
		for i := range validators {
			validators[i] = schema[i](t)
		}
		return func(v reflect.Value) error {
			var hits int
			var lastErr error
			for i, validate := range validators {
				err := validate(v)
				switch {
				case mode == allOf && err != nil:
					// AllOf needs every schema to pass, so one failure is final.
					return wrapError(name, err)
				case mode == anyOf && err == nil:
					// AnyOf needs just one match, so one success is final.
					return nil
				case mode == oneOf && err == nil:
					// OneOf needs exactly one match; a second one is already a failure.
					hits++
					if hits > 1 {
						return wrapError(name, ErrMultipleSchemasMatch)
					}
				case err != nil:
					// AnyOf/OneOf failure: keep it in case nothing else matches.
					lastErr = err
				}
				// Last schema checked, still no match for AnyOf/OneOf: report it.
				if i == len(validators)-1 && mode != allOf && hits == 0 {
					return wrapError(name, lastErr)
				}
			}
			return nil
		}
	}
}
