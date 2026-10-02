package vx

import (
	"fmt"
	"strings"
)

// PathElementKind identifies what a PathElement refers to in the validated data.
type PathElementKind int

const (
	// NoElement marks a check that validates the value itself rather than
	// navigating into one of its sub-elements (e.g. Len, AllOf, AnyOf, OneOf).
	NoElement PathElementKind = iota
	// FieldElement marks a struct field, identified by Name.
	FieldElement
	// IndexElement marks an array/slice element, identified by Index.
	IndexElement
	// KeyElement marks any other keyed element (e.g. a map value),
	// identified by Name. Name is the key's already-rendered text, e.g.
	// fmt.Sprintf("%q", k) for a string key; a custom check picks the
	// rendering appropriate to its own key type.
	KeyElement
)

// PathElement is a single step of a DataPath: a struct field, an
// array/slice index, or some other keyed element a custom check describes
// with KeyElement.
type PathElement struct {
	Kind  PathElementKind
	Name  string // set when Kind == FieldElement or KeyElement
	Index int    // set when Kind == IndexElement
}

// String renders the element the way it is appended to a path, e.g.
// ".Name" for a field, "[2]" for an index, and "[key]" for a KeyElement.
func (p PathElement) String() string {
	switch p.Kind {
	case FieldElement:
		return "." + p.Name
	case IndexElement:
		return fmt.Sprintf("[%d]", p.Index)
	case KeyElement:
		return "[" + p.Name + "]"
	default:
		return ""
	}
}

// RenderPath renders a DataPath as a single string,
// e.g. "Field1.NestedArray[2].Field2".
func RenderPath(path []PathElement) string {
	var b strings.Builder
	for _, p := range path {
		s := p.String()
		if strings.HasPrefix(s, ".") && b.Len() == 0 {
			s = s[1:]
		}
		b.WriteString(s)
	}
	return b.String()
}

// leafError is implemented by every terminal validation error (CheckError
// and the types that embed it) to expose the schema path segment it
// contributes, so SchemaPath can be reconstructed across the whole chain.
type leafError interface {
	error
	leafSegment() string
}

// CheckError is a leaf validation error: it is returned directly by a
// single, non-compound check (Gt, In, Format, Empty, ...) and terminates
// both the schema path and the data path at the point it occurs.
//
// Param, when set, is the schema-time argument of the check (e.g. "5" for
// Gt(5)) and is rendered as part of the schema path so that two uses of the
// same check can be told apart there, e.g. "Gt(5)" vs "Gt(10)".
//
// Checks that need to carry extra, programmatically accessible details
// (e.g. the threshold of a Gt check) embed CheckError in a more specific
// type; embedding promotes Error and leafSegment, so the specific type
// satisfies leafError too.
type CheckError struct {
	CheckName string // the check that failed, e.g. "Gt"
	Param     string // optional schema-time argument, e.g. "5"
	Msg       string
}

func newCheckError(check, param, msg string) *CheckError {
	return &CheckError{CheckName: check, Param: param, Msg: msg}
}

func (err *CheckError) Error() string { return err.Msg }

func (err *CheckError) leafSegment() string { return formatSegment(err.CheckName, err.Param) }

// formatSegment renders a single schema path segment, e.g. "Field" when
// param is empty or "Field(Name1)" when it is not.
func formatSegment(name, param string) string {
	if param == "" {
		return name
	}
	return name + "(" + param + ")"
}

// CompoundCheckError is returned by checks that delegate to a nested schema
// (Structure, Item, Len, AllOf, AnyOf, OneOf, ...). It attributes the
// failure of Cause to a place in the schema (CheckName, qualified by Param
// when the check takes a schema-time argument that cannot otherwise be
// told apart in the schema path, e.g. the field name of Field) and,
// optionally, a place in the validated data (DataItem).
type CompoundCheckError struct {
	CheckName string      // e.g. "Field", "Item", "Len", "AllOf"
	Param     string      // optional schema-time argument, e.g. the field name for Field
	DataItem  PathElement // zero value (NoElement) if this check validates the value itself
	Cause     error
}

// fieldError wraps cause as having occurred in the struct field named name.
func fieldError(name string, cause error) *CompoundCheckError {
	return &CompoundCheckError{
		CheckName: "Field",
		Param:     name,
		DataItem:  PathElement{Kind: FieldElement, Name: name},
		Cause:     cause,
	}
}

// itemError wraps cause as having occurred at the given index of a slice or array.
func itemError(index int, cause error) *CompoundCheckError {
	return &CompoundCheckError{
		CheckName: "Item",
		DataItem:  PathElement{Kind: IndexElement, Index: index},
		Cause:     cause,
	}
}

// wrapError wraps cause under checkName without navigating into a
// sub-element of the data (e.g. Len, AllOf, AnyOf, OneOf).
func wrapError(checkName string, cause error) *CompoundCheckError {
	return &CompoundCheckError{CheckName: checkName, Cause: cause}
}

func (err *CompoundCheckError) Error() string {
	schema := strings.Join(err.SchemaPath(), ".")
	data := RenderPath(err.DataPath())
	if data == "" {
		return fmt.Sprintf("%s: %v", schema, rootCause(err))
	}
	return fmt.Sprintf("%s: invalid %s: %v", schema, data, rootCause(err))
}

func (err *CompoundCheckError) Unwrap() error {
	return err.Cause
}

// SchemaPath returns the chain of check names that led to the failure,
// qualified by their schema-time argument where one exists, e.g.
// []string{"Item", "Field(Name1)", "Len", "Gt(5)"}.
func (err *CompoundCheckError) SchemaPath() []string {
	var path []string
	cur := error(err)
	for {
		ce, ok := cur.(*CompoundCheckError)
		if !ok {
			if le, ok := cur.(leafError); ok {
				path = append(path, le.leafSegment())
			}
			return path
		}
		path = append(path, formatSegment(ce.CheckName, ce.Param))
		if ce.Cause == nil {
			return path
		}
		cur = ce.Cause
	}
}

// DataPath returns the chain of field names and indices that led to the
// invalid value, e.g. []PathElement{Field("NestedArray"), Index(2), Field("Field2")}.
func (err *CompoundCheckError) DataPath() []PathElement {
	var path []PathElement
	cur := error(err)
	for {
		ce, ok := cur.(*CompoundCheckError)
		if !ok {
			return path
		}
		if ce.DataItem.Kind != NoElement {
			path = append(path, ce.DataItem)
		}
		if ce.Cause == nil {
			return path
		}
		cur = ce.Cause
	}
}

// rootCause returns the innermost, non-*CompoundCheckError cause of err.
func rootCause(err error) error {
	cur := err
	for {
		ce, ok := cur.(*CompoundCheckError)
		if !ok || ce.Cause == nil {
			return cur
		}
		cur = ce.Cause
	}
}
