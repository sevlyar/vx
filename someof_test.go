package vx

import (
	"errors"
	"slices"
	"testing"
)

func TestAllOf(t *testing.T) {
	t.Run("single schema is returned as-is", func(t *testing.T) {
		schema := Gt(0)
		if got := AllOf(schema).BindAny(); got == nil {
			t.Fatal("AllOf(schema) = nil")
		}
	})

	t.Run("panics without schemas", wantPanic(func() { AllOf() }))

	t.Run("early bind", func(t *testing.T) {
		schema := AllOf(Gt(0), Lt(10)).BindTypeOf(0)

		t.Run("passes all", wantNoError(schema.Check(5)))
		t.Run("fails first", wantError(schema.Check(-1)))
		t.Run("fails second", wantError(schema.Check(20)))
	})

	t.Run("late bind", func(t *testing.T) {
		schema := AllOf(Gt(0), Lt(10)).BindAny()

		t.Run("passes all", wantNoError(schema.Check(5)))
		t.Run("fails first", wantError(schema.Check(-1)))
	})

	t.Run("error carries the first failing schema in SchemaPath", func(t *testing.T) {
		err := AllOf(Gt(0), Lt(10)).BindAny().Check(-1)
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"AllOf", "Gt(0)"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
	})
}

func TestAnyOf(t *testing.T) {
	t.Run("single schema is returned as-is", func(t *testing.T) {
		schema := Gt(0)
		if got := AnyOf(schema).BindAny(); got == nil {
			t.Fatal("AnyOf(schema) = nil")
		}
	})

	t.Run("panics without schemas", wantPanic(func() { AnyOf() }))

	t.Run("early bind", func(t *testing.T) {
		schema := AnyOf(Lt(0), Gt(10)).BindTypeOf(0)

		t.Run("matches first", wantNoError(schema.Check(-1)))
		t.Run("matches second", wantNoError(schema.Check(20)))
		t.Run("matches neither", wantError(schema.Check(5)))
	})

	t.Run("late bind", func(t *testing.T) {
		schema := AnyOf(Lt(0), Gt(10)).BindAny()

		t.Run("matches first", wantNoError(schema.Check(-1)))
		t.Run("matches neither", wantError(schema.Check(5)))
	})

	t.Run("error when none match carries the last failure in SchemaPath", func(t *testing.T) {
		err := AnyOf(Lt(0), Gt(10)).BindAny().Check(5)
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"AnyOf", "Gt(10)"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
	})
}

func TestOneOf(t *testing.T) {
	t.Run("single schema is returned as-is", func(t *testing.T) {
		schema := Gt(0)
		if got := OneOf(schema).BindAny(); got == nil {
			t.Fatal("OneOf(schema) = nil")
		}
	})

	t.Run("panics without schemas", wantPanic(func() { OneOf() }))

	t.Run("early bind", func(t *testing.T) {
		schema := OneOf(Lt(0), Gt(10)).BindTypeOf(0)

		t.Run("matches exactly one", wantNoError(schema.Check(-1)))
		t.Run("matches none", wantError(schema.Check(5)))
	})

	t.Run("fails when more than one schema matches", func(t *testing.T) {
		err := OneOf(Gt(0), Gt(1)).BindAny().Check(5)
		if !errors.Is(err, ErrMultipleSchemasMatch) {
			t.Errorf("errors.Is(err, ErrMultipleSchemasMatch) = false, err = %v", err)
		}
	})

	t.Run("error when none match carries the last failure in SchemaPath", func(t *testing.T) {
		err := OneOf(Lt(0), Gt(10)).BindAny().Check(5)
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"OneOf", "Gt(10)"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
	})
}
