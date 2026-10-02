package vx

import (
	"errors"
	"testing"
)

func TestIn(t *testing.T) {
	t.Run("early check", func(t *testing.T) {
		t.Run("different types", wantPanic(func() { In(1, "string", false).BindTypeOf(2) }))

		t.Run("one type", func(t *testing.T) {
			validator := In(1, 2, 3).BindTypeOf(0)
			t.Run("unknown value", wantError(validator.Check(0)))
			t.Run("known value", wantNoError(validator.Check(1)))
		})

		t.Run("slice unwound", func(t *testing.T) {
			validator := In([]int{1, 2, 3}, 4).BindTypeOf(0)
			t.Run("unknown value", wantError(validator.Check(0)))
			t.Run("known value from slice", wantNoError(validator.Check(1)))
			t.Run("known value from variadic", wantNoError(validator.Check(4)))
		})
	})

	t.Run("late check", func(t *testing.T) {
		var valid BoundSchema
		t.Run("bind does not panic", wantNoPanic(func() { valid = In(1, "string", false).BindAny() }))

		t.Run("known int", wantNoError(valid.Check(1)))
		t.Run("known bool", wantNoError(valid.Check(false)))
		t.Run("unknown bool", wantError(valid.Check(true)))
	})
}

// TestIn_LateCheck_UncomparableValue_Errors is a regression test: the
// comparability guard used to run only in the early-bind branch, so a
// late-bound In (via BindAny, or inside an any-typed field) panicked on an
// uncomparable value (slice, map, func) instead of returning an error.
func TestIn_LateCheck_UncomparableValue_Errors(t *testing.T) {
	wantError(In(1, 2, 3).BindAny().Check([]int{1, 2, 3}))(t)
}

// TestIn_NestedInAnySlot_UncomparableValue_Errors is a regression test: the
// late-bind comparability guard checked v.Type(), but an element reached
// through an any-typed slot (e.g. a []any item) is itself interface-kind,
// so Type() reports the always-comparable interface type instead of the
// dynamic value actually being compared, letting an uncomparable value
// (here a []int) slip through to panic on the map lookup.
func TestIn_NestedInAnySlot_UncomparableValue_Errors(t *testing.T) {
	t.Run("uncomparable element", wantError(Item(In(1, 2, 3)).BindAny().Check([]any{1, []int{9}})))
	t.Run("nil element", wantError(Item(In(1, 2, 3)).BindAny().Check([]any{1, nil})))
}

func TestIn_ErrorMessageTruncatesLargeSets(t *testing.T) {
	err := In(1, 2, 3).BindAny().Check(0)
	if got, want := err.Error(), "value is not one of [1 2 3]"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	err = In(1, 2, 3, 4, 5).BindAny().Check(0)
	if got, want := err.Error(), "value is not one of [1 2 3 ...]"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	var uv *UnknownValueError
	if !errors.As(err, &uv) {
		t.Fatal("errors.As(err, *UnknownValueError) = false")
	}
	if got, want := len(uv.KnownValues), 5; got != want {
		t.Errorf("len(KnownValues) = %d, want %d (the full set, not truncated)", got, want)
	}
}
