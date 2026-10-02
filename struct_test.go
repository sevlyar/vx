package vx

import (
	"slices"
	"testing"
)

type person struct {
	Name string
	Age  int
}

func personSchema() Schema {
	var p person
	return Structure(&p,
		Field(&p.Name, Len(Gt(0))),
		Field(&p.Age, Ge(0)),
	)
}

func TestStructure(t *testing.T) {
	schema := personSchema().BindTypeOf(person{})

	t.Run("valid", wantNoError(schema.Check(person{Name: "Alice", Age: 30})))
	t.Run("empty name", wantError(schema.Check(person{Name: "", Age: 30})))
	t.Run("negative age", wantError(schema.Check(person{Name: "Alice", Age: -1})))

	t.Run("first failing field wins", func(t *testing.T) {
		err := schema.Check(person{Name: "", Age: -1})
		ce, ok := err.(*CompoundCheckError)
		if !ok {
			t.Fatalf("error type = %T, want *CompoundCheckError", err)
		}
		if got, want := ce.SchemaPath(), []string{"Field(Name)", "Len", "Gt(0)"}; !slices.Equal(got, want) {
			t.Errorf("SchemaPath() = %v, want %v", got, want)
		}
	})

	t.Run("late bind wrong type errors, not panics", wantError(personSchema().BindAny().Check(42)))
}

func TestStructure_PtrNotToStruct_Panics(t *testing.T) {
	t.Run("not a pointer", wantPanic(func() { Structure(person{}) }))
	t.Run("pointer to non-struct", wantPanic(func() { var i int; Structure(&i) }))
}

func TestStructure_MissingFieldOption_Panics(t *testing.T) {
	wantPanic(func() {
		var p person
		Structure(&p, Field(&p.Name, Len(Gt(0))))
	})(t)
}

func TestStructure_AllowUncheckedFields(t *testing.T) {
	schema := func() Schema {
		var p person
		return Structure(&p,
			Field(&p.Name, Len(Gt(0))),
			AllowUncheckedFields(&p.Age),
		)
	}().BindTypeOf(person{})

	t.Run("unchecked field ignored", wantNoError(schema.Check(person{Name: "Alice", Age: -1})))
	t.Run("checked field still validated", wantError(schema.Check(person{Name: "", Age: -1})))
}

func TestStructure_FieldPtr_Invalid(t *testing.T) {
	t.Run("not a pointer", wantPanic(func() {
		var p person
		Structure(&p, Field(p.Name, Len(Gt(0))), Field(&p.Age, Ge(0)))
	}))

	t.Run("points outside the structure", wantPanic(func() {
		var p person
		var other person
		Structure(&p, Field(&other.Name, Len(Gt(0))), Field(&p.Age, Ge(0)))
	}))

	t.Run("nil schema", wantPanic(func() {
		var p person
		Structure(&p, Field(&p.Name, nil), Field(&p.Age, Ge(0)))
	}))
}

// TestStructure_FullChain is the end-to-end case the whole error-path
// design was built for: a struct field holding a slice of structs, each
// checked by Len(Gt(...)) on one of their own fields.
func TestStructure_FullChain(t *testing.T) {
	type inner struct {
		Field2 string
	}
	type nested struct {
		NestedArray []inner
	}
	type outer struct {
		Field1 nested
	}

	var in inner
	innerSchema := Structure(&in, Field(&in.Field2, Len(Gt(3))))

	var n nested
	nestedSchema := Structure(&n, Field(&n.NestedArray, Item(innerSchema)))

	var o outer
	outerSchema := Structure(&o, Field(&o.Field1, nestedSchema))

	data := outer{Field1: nested{NestedArray: []inner{
		{Field2: "abcd"},
		{Field2: "ab"},
		{Field2: "abcd"},
	}}}

	err := outerSchema.BindAny().Check(data)
	ce, ok := err.(*CompoundCheckError)
	if !ok {
		t.Fatalf("error type = %T, want *CompoundCheckError", err)
	}

	wantSchema := []string{"Field(Field1)", "Field(NestedArray)", "Item", "Field(Field2)", "Len", "Gt(3)"}
	if got := ce.SchemaPath(); !slices.Equal(got, wantSchema) {
		t.Errorf("SchemaPath() = %v, want %v", got, wantSchema)
	}

	wantData := "Field1.NestedArray[1].Field2"
	if got := RenderPath(ce.DataPath()); got != wantData {
		t.Errorf("RenderPath(DataPath()) = %q, want %q", got, wantData)
	}
}
