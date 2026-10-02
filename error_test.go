package vx

import (
	"errors"
	"slices"
	"testing"
)

// TestCompoundCheckError_FullChain simulates the nested schema
//
//	Structure(Field("Field1", Structure(Field("NestedArray", Item(
//	    Structure(Field("Field2", Len(Gt(3)))),
//	))))
//
// failing on data.Field1.NestedArray[2].Field2, to check that DataPath
// reconstructs the path into the data, and SchemaPath reconstructs the path
// into the schema that rejected it, qualified by each check's own
// schema-time argument (the field names, the threshold) so that two
// sibling Field(...) or Gt(...) checks can be told apart.
func TestCompoundCheckError_FullChain(t *testing.T) {
	leaf := newCheckError("Gt", "3", "value should be greater than threshold")
	err := fieldError("Field1",
		fieldError("NestedArray",
			itemError(2,
				fieldError("Field2",
					wrapError("Len", leaf),
				),
			),
		),
	)

	wantSchema := []string{"Field(Field1)", "Field(NestedArray)", "Item", "Field(Field2)", "Len", "Gt(3)"}
	if got := err.SchemaPath(); !slices.Equal(got, wantSchema) {
		t.Errorf("SchemaPath() = %v, want %v", got, wantSchema)
	}

	if got := RenderPath(err.DataPath()); got != "Field1.NestedArray[2].Field2" {
		t.Errorf("RenderPath(DataPath()) = %q, want %q", got, "Field1.NestedArray[2].Field2")
	}

	wantErr := "Field(Field1).Field(NestedArray).Item.Field(Field2).Len.Gt(3): " +
		"invalid Field1.NestedArray[2].Field2: value should be greater than threshold"
	if got := err.Error(); got != wantErr {
		t.Errorf("Error() = %q, want %q", got, wantErr)
	}

	// The leaf is reachable both via errors.As (standard unwrap chain) and
	// as the formatted root cause.
	var cmp *CheckError
	if !errors.As(err, &cmp) {
		t.Fatal("errors.As(err, *CheckError) = false, want true")
	}
	if cmp != leaf {
		t.Errorf("errors.As found %v, want %v", cmp, leaf)
	}
	if got := rootCause(err); got != error(leaf) {
		t.Errorf("rootCause(err) = %v, want %v", got, leaf)
	}
}

// TestCompoundCheckError_SchemaPathWithoutLeadingField simulates validating
// a slice directly (no enclosing struct field), e.g.
//
//	Item(Structure(Field("Field2", Len(Gt(3)))))
//
// matching the shorter schema-path shape quoted in the issue.
func TestCompoundCheckError_SchemaPathWithoutLeadingField(t *testing.T) {
	leaf := newCheckError("Gt", "3", "value should be greater than threshold")
	err := itemError(2,
		fieldError("Field2",
			wrapError("Len", leaf),
		),
	)

	wantSchema := []string{"Item", "Field(Field2)", "Len", "Gt(3)"}
	if got := err.SchemaPath(); !slices.Equal(got, wantSchema) {
		t.Errorf("SchemaPath() = %v, want %v", got, wantSchema)
	}
	if got := RenderPath(err.DataPath()); got != "[2].Field2" {
		t.Errorf("RenderPath(DataPath()) = %q, want %q", got, "[2].Field2")
	}
}

// TestCompoundCheckError_SiblingFieldsDistinguishable is the direct
// regression test for the ask: two sibling Field checks at the same depth,
// failing on the same kind of leaf check, must still produce distinct
// schema paths once qualified by their field name.
func TestCompoundCheckError_SiblingFieldsDistinguishable(t *testing.T) {
	err1 := fieldError("Name1", newCheckError("Gt", "3", "value should be greater than threshold"))
	err2 := fieldError("Name2", newCheckError("Gt", "3", "value should be greater than threshold"))

	want1 := []string{"Field(Name1)", "Gt(3)"}
	want2 := []string{"Field(Name2)", "Gt(3)"}
	if got := err1.SchemaPath(); !slices.Equal(got, want1) {
		t.Errorf("err1.SchemaPath() = %v, want %v", got, want1)
	}
	if got := err2.SchemaPath(); !slices.Equal(got, want2) {
		t.Errorf("err2.SchemaPath() = %v, want %v", got, want2)
	}
	if slices.Equal(err1.SchemaPath(), err2.SchemaPath()) {
		t.Errorf("err1.SchemaPath() and err2.SchemaPath() must differ, both are %v", err1.SchemaPath())
	}
}

// TestCompoundCheckError_NoDataNavigation checks that checks which validate
// the value itself (Len, AllOf, AnyOf, OneOf) contribute to SchemaPath but
// not to DataPath, and that an unparametrized check renders without
// parentheses.
func TestCompoundCheckError_NoDataNavigation(t *testing.T) {
	leaf := newCheckError("Gt", "3", "value should be greater than threshold")
	err := wrapError("AllOf", wrapError("Len", leaf))

	wantSchema := []string{"AllOf", "Len", "Gt(3)"}
	if got := err.SchemaPath(); !slices.Equal(got, wantSchema) {
		t.Errorf("SchemaPath() = %v, want %v", got, wantSchema)
	}
	if got := err.DataPath(); len(got) != 0 {
		t.Errorf("DataPath() = %v, want empty", got)
	}

	wantErr := "AllOf.Len.Gt(3): value should be greater than threshold"
	if got := err.Error(); got != wantErr {
		t.Errorf("Error() = %q, want %q", got, wantErr)
	}
}

// TestCompoundCheckError_UnwrapChain checks the standard errors.Is/As
// machinery works through the whole chain, so callers that only care about
// the terminal error don't need to know about CompoundCheckError at all.
func TestCompoundCheckError_UnwrapChain(t *testing.T) {
	sentinel := errors.New("sentinel")
	err := fieldError("Field1", wrapError("OneOf", sentinel))

	if !errors.Is(err, sentinel) {
		t.Error("errors.Is(err, sentinel) = false, want true")
	}

	// A sentinel cause is not a leafError, so it does not extend SchemaPath.
	wantSchema := []string{"Field(Field1)", "OneOf"}
	if got := err.SchemaPath(); !slices.Equal(got, wantSchema) {
		t.Errorf("SchemaPath() = %v, want %v", got, wantSchema)
	}
}

func TestCheckError_StandaloneLeaf(t *testing.T) {
	err := newCheckError("Empty", "", "value is not empty")

	if got := err.Error(); got != "value is not empty" {
		t.Errorf("Error() = %q, want %q", got, "value is not empty")
	}
	if got := err.leafSegment(); got != "Empty" {
		t.Errorf("leafSegment() = %q, want %q", got, "Empty")
	}
}

func TestCheckError_ParametrizedLeaf(t *testing.T) {
	err := newCheckError("Gt", "5", "value should be greater than 5")

	if got := err.leafSegment(); got != "Gt(5)" {
		t.Errorf("leafSegment() = %q, want %q", got, "Gt(5)")
	}
}

// TestPathElement_KeyElement checks a custom, map-like check can describe
// its own kind of DataPath step: KeyElement is not used by any check in
// this package, only by extensions.
func TestPathElement_KeyElement(t *testing.T) {
	path := []PathElement{
		{Kind: FieldElement, Name: "Tags"},
		{Kind: KeyElement, Name: `"color"`},
	}
	if got, want := RenderPath(path), `Tags["color"]`; got != want {
		t.Errorf("RenderPath(path) = %q, want %q", got, want)
	}
}
