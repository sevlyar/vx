package vx

import (
	"errors"
	"slices"
	"testing"
)

// TestBindTypeOf_DifferentTypeSameKind_Panics is a regression test: the
// runtime guard in BindTypeOf used to compare only Kind(), so a value of a
// different type with the same Kind (e.g. two different structs) slipped
// through and corrupted reflect's view of the data instead of panicking
// cleanly.
func TestBindTypeOf_DifferentTypeSameKind_Panics(t *testing.T) {
	type a struct{ X int }
	type b struct {
		X string
		Y int
	}

	var av a
	bound := Structure(&av, Field(&av.X, Gt(0))).BindTypeOf(a{})

	wantPanic(func() { _ = bound.Check(b{X: "hello"}) })(t)
}

// TestBindTypeCheck_NestedInAnySlot_Works is a regression test: the
// late-bind branch of BindTypeCheck used to inspect v.Type() without
// unwrapping an interface-kind Value first, so any leaf check (Gt, Format,
// ...) used inside an any-typed slot (a []any element, a struct field
// typed any) rejected every value with a false type-mismatch error.
func TestBindTypeCheck_NestedInAnySlot_Works(t *testing.T) {
	t.Run("valid value passes", wantNoError(Item(Gt(0)).BindAny().Check([]any{5, 10})))
	t.Run("invalid value still fails", wantError(Item(Gt(0)).BindAny().Check([]any{5, -1})))

	t.Run("nil element returns ErrNilValue, not a panic", func(t *testing.T) {
		err := Item(Gt(0)).BindAny().Check([]any{5, nil})
		if !errors.Is(err, ErrNilValue) {
			t.Errorf("errors.Is(err, ErrNilValue) = false, err = %v", err)
		}
		// ErrNilValue must still contribute its own SchemaPath segment,
		// like every other leaf failure (Gt, Format, Empty, ...) does.
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"Item", "NotNil"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
	})

	t.Run("struct field typed any", func(t *testing.T) {
		type holder struct{ X any }
		var h holder
		schema := Structure(&h, Field(&h.X, Gt(0))).BindAny()
		wantNoError(schema.Check(holder{X: 5}))(t)
	})
}
