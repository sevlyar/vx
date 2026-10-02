package vx

import "testing"

func TestGt(t *testing.T) {
	t.Parallel()

	t.Run("greater of int value", func(t *testing.T) {
		schema := Gt(2).BindAny()

		t.Run("check int", wantNoError(schema.Check(4)))
		t.Run("check uint", wantNoError(schema.Check(uint(4))))
		t.Run("check float", wantNoError(schema.Check(4.0)))
		t.Run("check greater of max int", wantNoError(schema.Check(uint(9223372036854775808))))
		t.Run("check equal values", wantError(schema.Check(2)))
	})

	t.Run("greater of uint value", func(t *testing.T) {
		schema := Gt(uint(2)).BindAny()

		t.Run("check int", wantNoError(schema.Check(4)))
		t.Run("check negative int", wantError(schema.Check(-4)))
		t.Run("check uint", wantNoError(schema.Check(uint(4))))
		t.Run("check float", wantNoError(schema.Check(4.0)))
		t.Run("check equal values", wantError(schema.Check(2)))
	})

	t.Run("greater of float value", func(t *testing.T) {
		schema := Gt(2.0).BindAny()

		t.Run("check int", wantNoError(schema.Check(4)))
		t.Run("check uint", wantNoError(schema.Check(uint(4))))
		t.Run("check float", wantNoError(schema.Check(4.0)))
		t.Run("check equal values", wantError(schema.Check(2)))
	})
}

func TestGe(t *testing.T) {
	t.Parallel()

	t.Run("greater or equal of int value", func(t *testing.T) {
		schema := Ge(2).BindAny()

		t.Run("check int", wantNoError(schema.Check(4)))
		t.Run("check uint", wantNoError(schema.Check(uint(4))))
		t.Run("check float", wantNoError(schema.Check(4.0)))
		t.Run("check greater of max int", wantNoError(schema.Check(uint(9223372036854775808))))
		t.Run("check equal values", wantNoError(schema.Check(2)))
	})

	t.Run("greater or equal of uint value", func(t *testing.T) {
		schema := Ge(uint(2)).BindAny()

		t.Run("check int", wantNoError(schema.Check(4)))
		t.Run("check negative int", wantError(schema.Check(-4)))
		t.Run("check uint", wantNoError(schema.Check(uint(4))))
		t.Run("check float", wantNoError(schema.Check(4.0)))
		t.Run("check equal values", wantNoError(schema.Check(2)))
	})

	t.Run("greater or equal of float value", func(t *testing.T) {
		schema := Ge(2.0).BindAny()

		t.Run("check int", wantNoError(schema.Check(4)))
		t.Run("check uint", wantNoError(schema.Check(uint(4))))
		t.Run("check float", wantNoError(schema.Check(4.0)))
		t.Run("check equal values", wantNoError(schema.Check(2)))
	})
}

func TestLt(t *testing.T) {
	t.Parallel()

	t.Run("less of int value", func(t *testing.T) {
		schema := Lt(4).BindAny()

		t.Run("check int", wantNoError(schema.Check(2)))
		t.Run("check uint", wantNoError(schema.Check(uint(2))))
		t.Run("check float", wantNoError(schema.Check(2.0)))
		t.Run("check equal values", wantError(schema.Check(4)))
	})

	t.Run("less of uint value", func(t *testing.T) {
		schema := Lt(uint(4)).BindAny()

		t.Run("check int", wantNoError(schema.Check(2)))
		t.Run("check uint", wantNoError(schema.Check(uint(2))))
		t.Run("check float", wantNoError(schema.Check(2.0)))
		t.Run("check equal values", wantError(schema.Check(4)))
	})

	t.Run("less of float value", func(t *testing.T) {
		schema := Lt(4.0).BindAny()

		t.Run("check int", wantNoError(schema.Check(2)))
		t.Run("check uint", wantNoError(schema.Check(uint(2))))
		t.Run("check float", wantNoError(schema.Check(2.0)))
		t.Run("check equal values", wantError(schema.Check(4)))
	})
}

func TestLe(t *testing.T) {
	t.Parallel()

	t.Run("less or equal of int value", func(t *testing.T) {
		schema := Le(4).BindAny()

		t.Run("check int", wantNoError(schema.Check(2)))
		t.Run("check uint", wantNoError(schema.Check(uint(2))))
		t.Run("check float", wantNoError(schema.Check(2.0)))
		t.Run("check equal values", wantNoError(schema.Check(4)))
	})

	t.Run("less or equal of uint value", func(t *testing.T) {
		schema := Le(uint(4)).BindAny()

		t.Run("check int", wantNoError(schema.Check(2)))
		t.Run("check uint", wantNoError(schema.Check(uint(2))))
		t.Run("check float", wantNoError(schema.Check(2.0)))
		t.Run("check equal values", wantNoError(schema.Check(4)))
	})

	t.Run("less or equal of float value", func(t *testing.T) {
		schema := Le(4.0).BindAny()

		t.Run("check int", wantNoError(schema.Check(2)))
		t.Run("check uint", wantNoError(schema.Check(uint(2))))
		t.Run("check float", wantNoError(schema.Check(2.0)))
		t.Run("check equal values", wantNoError(schema.Check(4)))
	})
}

// TestComparison_WrongThresholdType_Panics covers a case missing from the
// original test suite: constructing a comparison with a non-numeric
// threshold must panic at schema-build time, not at check time.
func TestComparison_WrongThresholdType_Panics(t *testing.T) {
	t.Run("Gt(string)", wantPanic(func() { Gt("not a number") }))
}

// TestComparison_WrongValueType_Errors covers a case missing from the
// original test suite: checking a non-numeric value against a comparison
// schema bound late (through an interface) must return an error, not panic.
func TestComparison_WrongValueType_Errors(t *testing.T) {
	t.Run("Gt(2).Check(string)", wantError(Gt(2).BindAny().Check("not a number")))
}

// TestSchemaPath_Comparison checks the schema path of a failing comparison
// carries the threshold, e.g. "Gt(2)".
func TestSchemaPath_Comparison(t *testing.T) {
	err := Gt(2).BindAny().Check(1)
	cmp, ok := err.(*CheckError)
	if !ok {
		t.Fatalf("error type = %T, want *CheckError", err)
	}
	if got := cmp.leafSegment(); got != "Gt(2)" {
		t.Errorf("leafSegment() = %q, want %q", got, "Gt(2)")
	}
}
