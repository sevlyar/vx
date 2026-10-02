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

// TestBindTypeCheck_NestedInAnySlot_Works covers the two cases of the
// BindTypeCheck any-slot regression (see TestChecks_NestedInAnySlot for
// the general valid/invalid sweep) that sweep doesn't: a nil value, and a
// struct field typed any rather than a []any element.
func TestBindTypeCheck_NestedInAnySlot_Works(t *testing.T) {
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

// TestChecks_NestedInAnySlot is a sweep, not a per-feature test: every
// check below is driven through exactly one any-typed boundary —
// Item(check).BindAny() over a []any — the same shape of indirection that
// broke BindTypeCheck, BindCompound and valueIsEmpty (see derefInterface
// in schema.go). A struct field typed any goes through the identical code
// path (the value arrives as a Kind-Interface reflect.Value either way),
// so this one mechanism stands in for both.
//
// Add a case here for every new check, so a future regression in this
// class of bug is caught across the whole library at once, not
// rediscovered one check at a time.
func TestChecks_NestedInAnySlot(t *testing.T) {
	type person struct{ Name string }

	personSchema := func() Schema {
		var p person
		return Structure(&p, Field(&p.Name, Len(Gt(0))))
	}()

	cases := []struct {
		name    string
		schema  Schema
		valid   any
		invalid any
	}{
		{"Gt", Gt(0), 5, -1},
		{"Ge", Ge(0), 0, -1},
		{"Lt", Lt(10), 5, 20},
		{"Le", Le(10), 10, 20},
		{"In", In(1, 2, 3), 2, 99},
		{"Empty", Empty, 0, 5},
		{"NonEmpty", NonEmpty, 5, 0},
		{"Format", Format("Upper", isUpper), "ABC", "abc"},
		{"RegexpFormat", RegexpFormat(`^\d+$`), "123", "abc"},
		{"PrintableLine", PrintableLine, "hello", "line1\nline2"},
		{"PrintableText", PrintableText, "line1\nline2", "\x00"},
		{"Len", Len(Gt(3)), "hello", "hi"},
		{"Item", Item(Gt(0)), []int{1, 2, 3}, []int{1, -2, 3}},
		{"AllOf", AllOf(Gt(0), Lt(10)), 5, -1},
		{"AnyOf", AnyOf(Lt(0), Gt(10)), -5, 5},
		{"OneOf", OneOf(Lt(0), Gt(10)), -5, 5},
		{"Structure", personSchema, person{Name: "Alice"}, person{Name: ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			schema := Item(tc.schema).BindAny()

			t.Run("valid passes", wantNoError(schema.Check([]any{tc.valid})))
			t.Run("invalid still fails", wantError(schema.Check([]any{tc.invalid})))
		})
	}
}
