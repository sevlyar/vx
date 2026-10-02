package vx_test

import (
	"fmt"

	"github.com/sevlyar/vx"
)

// Example validates a struct: every field needs a Field option, checks run
// in the order they are declared, and the first failure wins.
func Example() {
	type Person struct {
		Name string
		Age  int
	}

	var p Person
	schema := vx.Structure(&p,
		vx.Field(&p.Name, vx.Len(vx.Gt(0))),
		vx.Field(&p.Age, vx.Ge(0)),
	)

	err := schema.BindAny().Check(Person{Name: "", Age: -1})
	fmt.Println(err)
	// Output:
	// Field(Name).Len.Gt(0): invalid Name: value should be greater than 0
}

// Example_errorPaths shows SchemaPath and DataPath on a failure nested
// through a struct field, a slice item, and a second struct field.
func Example_errorPaths() {
	type Item struct{ Name string }
	type Container struct{ Items []Item }

	var it Item
	itemSchema := vx.Structure(&it, vx.Field(&it.Name, vx.Len(vx.Gt(3))))

	var c Container
	schema := vx.Structure(&c, vx.Field(&c.Items, vx.Item(itemSchema)))

	err := schema.BindAny().Check(Container{Items: []Item{{Name: "abcd"}, {Name: "ab"}}})
	ce := err.(*vx.CompoundCheckError)

	fmt.Println(ce.SchemaPath())
	fmt.Println(vx.RenderPath(ce.DataPath()))
	// Output:
	// [Field(Items) Item Field(Name) Len Gt(3)]
	// Items[1].Name
}

func ExampleGt() {
	schema := vx.Gt(0).BindAny()
	fmt.Println(schema.Check(5))
	fmt.Println(schema.Check(-1))
	// Output:
	// <nil>
	// value should be greater than 0
}

func ExampleIn() {
	schema := vx.In("red", "green", "blue").BindAny()
	fmt.Println(schema.Check("green"))
	fmt.Println(schema.Check("purple"))
	// Output:
	// <nil>
	// value is not one of [red green blue]
}

func ExampleRegexpFormat() {
	schema := vx.RegexpFormat(`^\d+$`).BindAny()
	fmt.Println(schema.Check("12345"))
	fmt.Println(schema.Check("12a45"))
	// Output:
	// <nil>
	// value does not match ^\d+$
}

func ExampleItem() {
	schema := vx.Item(vx.Gt(0)).BindAny()
	fmt.Println(schema.Check([]int{1, 2, 3}))
	fmt.Println(schema.Check([]int{1, -2, 3}))
	// Output:
	// <nil>
	// Item.Gt(0): invalid [1]: value should be greater than 0
}

func ExampleLen() {
	schema := vx.Len(vx.Gt(3)).BindAny()
	fmt.Println(schema.Check("hello"))
	fmt.Println(schema.Check("hi"))
	// Output:
	// <nil>
	// Len.Gt(3): value should be greater than 3
}

// ExampleAllOf shows AllOf failing on the first schema that does not match.
func ExampleAllOf() {
	schema := vx.AllOf(vx.Gt(0), vx.Lt(10)).BindAny()
	fmt.Println(schema.Check(5))
	fmt.Println(schema.Check(15))
	// Output:
	// <nil>
	// AllOf.Lt(10): value should be less than 10
}

// ExampleOneOf shows OneOf requiring exactly one schema to match; here
// neither Lt(0) nor Gt(10) matches 5.
func ExampleOneOf() {
	schema := vx.OneOf(vx.Lt(0), vx.Gt(10)).BindAny()
	fmt.Println(schema.Check(-1))
	fmt.Println(schema.Check(5))
	// Output:
	// <nil>
	// OneOf.Gt(10): value should be greater than 10
}
