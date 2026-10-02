package vx

import (
	"slices"
	"testing"
)

func TestItem(t *testing.T) {
	t.Run("early bind", func(t *testing.T) {
		schema := Item(Gt(3)).BindTypeOf([]int{})

		t.Run("all items valid", wantNoError(schema.Check([]int{4, 5, 6})))
		t.Run("empty slice", wantNoError(schema.Check([]int{})))
		t.Run("item invalid", wantError(schema.Check([]int{4, 2, 6})))
	})

	t.Run("works on arrays too", func(t *testing.T) {
		schema := Item(Gt(3)).BindTypeOf([4]int{})
		t.Run("all items valid", wantNoError(schema.Check([4]int{4, 5, 6, 7})))
		t.Run("item invalid", wantError(schema.Check([4]int{4, 2, 6, 7})))
	})

	t.Run("early bind wrong type panics", wantPanic(func() { Item(Gt(3)).BindTypeOf(0) }))

	t.Run("early bind propagates nested schema panic", wantPanic(func() {
		Item(Gt(3)).BindTypeOf([]string{})
	}))

	t.Run("late bind", func(t *testing.T) {
		schema := Item(Gt(3)).BindAny()

		t.Run("wrong type errors, not panics", wantError(schema.Check(0)))
		t.Run("all items valid", wantNoError(schema.Check([]int{4, 5, 6})))
		t.Run("item invalid", wantError(schema.Check([]int{4, 2, 6})))
	})

	t.Run("error reports index and cause in DataPath/SchemaPath", func(t *testing.T) {
		err := Item(Gt(3)).BindAny().Check([]int{4, 2, 6})
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"Item", "Gt(3)"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
		if got, want := RenderPath(ce.DataPath()), "[1]"; got != want {
			t.Errorf("RenderPath(DataPath()) = %q, want %q", got, want)
		}
	})
}

func TestLen(t *testing.T) {
	t.Run("early bind", func(t *testing.T) {
		schema := Len(Gt(3)).BindTypeOf("")

		t.Run("long enough", wantNoError(schema.Check("hello")))
		t.Run("too short", wantError(schema.Check("hi")))
		t.Run("counts runes, not bytes", wantNoError(schema.Check("héllo")))
	})

	t.Run("works on slice, array and map", func(t *testing.T) {
		schema := Len(Gt(1)).BindAny()

		t.Run("slice", wantNoError(schema.Check([]int{1, 2, 3})))
		t.Run("array", wantNoError(schema.Check([2]int{1, 2})))
		t.Run("map", wantNoError(schema.Check(map[string]int{"a": 1, "b": 2})))
	})

	t.Run("early bind wrong type panics", wantPanic(func() { Len(Gt(3)).BindTypeOf(0) }))

	t.Run("late bind wrong type errors, not panics", wantError(Len(Gt(3)).BindAny().Check(0)))

	t.Run("error reports check chain in SchemaPath, no DataPath", func(t *testing.T) {
		err := Len(Gt(3)).BindAny().Check("hi")
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"Len", "Gt(3)"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
		if got := ce.DataPath(); len(got) != 0 {
			t.Errorf("DataPath() = %v, want empty", got)
		}
	})
}
