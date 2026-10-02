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
