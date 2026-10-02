package benchmarks

import (
	"regexp"
	"testing"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/invopop/validation"
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/sevlyar/vx"
)

// Both custom-format fields below are checked with the exact same compiled
// regexp in all three libraries, so the comparison measures framework
// overhead, not differences in how thorough each library's built-in email
// validator happens to be.
var (
	codeRe = regexp.MustCompile(`^[A-Z]{2}\d{4}$`)
	zipRe  = regexp.MustCompile(`^\d{5}$`)
)

// ---------- flat struct: Name, Age, Code ----------

type vxPerson struct {
	Name string
	Age  int
	Code string
}

type gpPerson struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=0"`
	Code string `validate:"required,code"`
}

type ivPerson struct {
	Name string
	Age  int
	Code string
}

func (p ivPerson) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.Age, validation.Min(0)),
		validation.Field(&p.Code, validation.Required, validation.Match(codeRe)),
	)
}

var vxPersonSchema = func() vx.Schema {
	var p vxPerson
	return vx.Structure(&p,
		vx.Field(&p.Name, vx.Len(vx.Gt(0))),
		vx.Field(&p.Age, vx.Ge(0)),
		vx.Field(&p.Code, vx.RegexpFormat(codeRe.String())),
	)
}()

type gvPerson struct {
	Name string
	Age  int
	Code string
}

var gvPersonValidator = govy.New(
	govy.For(func(p gvPerson) string { return p.Name }).
		WithName("name").
		Required(),
	govy.For(func(p gvPerson) int { return p.Age }).
		WithName("age").
		Rules(rules.GTE(0)),
	govy.For(func(p gvPerson) string { return p.Code }).
		WithName("code").
		Required().
		Rules(rules.StringMatchRegexp(codeRe)),
)

var gpValidate = func() *govalidator.Validate {
	v := govalidator.New()
	_ = v.RegisterValidation("code", func(fl govalidator.FieldLevel) bool {
		return codeRe.MatchString(fl.Field().String())
	})
	_ = v.RegisterValidation("zip", func(fl govalidator.FieldLevel) bool {
		return zipRe.MatchString(fl.Field().String())
	})
	return v
}()

// ---------- nested struct: Customer{Name, []Address{City, Zip}} ----------

type vxAddress struct {
	City string
	Zip  string
}
type vxCustomer struct {
	Name      string
	Addresses []vxAddress
}

type gpAddress struct {
	City string `validate:"required"`
	Zip  string `validate:"required,zip"`
}
type gpCustomer struct {
	Name      string      `validate:"required"`
	Addresses []gpAddress `validate:"required,dive"`
}

type ivAddress struct {
	City string
	Zip  string
}

func (a ivAddress) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.City, validation.Required),
		validation.Field(&a.Zip, validation.Required, validation.Match(zipRe)),
	)
}

type ivCustomer struct {
	Name      string
	Addresses []ivAddress
}

func (c ivCustomer) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Addresses, validation.Each()),
	)
}

var vxAddressSchema = func() vx.Schema {
	var a vxAddress
	return vx.Structure(&a,
		vx.Field(&a.City, vx.Len(vx.Gt(0))),
		vx.Field(&a.Zip, vx.RegexpFormat(zipRe.String())),
	)
}()

var vxCustomerSchema = func() vx.Schema {
	var c vxCustomer
	return vx.Structure(&c,
		vx.Field(&c.Name, vx.Len(vx.Gt(0))),
		vx.Field(&c.Addresses, vx.Item(vxAddressSchema)),
	)
}()

type gvAddress struct {
	City string
	Zip  string
}
type gvCustomer struct {
	Name      string
	Addresses []gvAddress
}

var gvAddressValidator = govy.New(
	govy.For(func(a gvAddress) string { return a.City }).
		WithName("city").
		Required(),
	govy.For(func(a gvAddress) string { return a.Zip }).
		WithName("zip").
		Required().
		Rules(rules.StringMatchRegexp(zipRe)),
)

var gvCustomerValidator = govy.New(
	govy.For(func(c gvCustomer) string { return c.Name }).
		WithName("name").
		Required(),
	govy.ForSlice(func(c gvCustomer) []gvAddress { return c.Addresses }).
		WithName("addresses").
		IncludeForEach(gvAddressValidator),
)

func makeAddresses(n int) ([]vxAddress, []gpAddress, []ivAddress, []gvAddress) {
	var vxA []vxAddress
	var gpA []gpAddress
	var ivA []ivAddress
	var gvA []gvAddress
	for i := 0; i < n; i++ {
		vxA = append(vxA, vxAddress{City: "Springfield", Zip: "12345"})
		gpA = append(gpA, gpAddress{City: "Springfield", Zip: "12345"})
		ivA = append(ivA, ivAddress{City: "Springfield", Zip: "12345"})
		gvA = append(gvA, gvAddress{City: "Springfield", Zip: "12345"})
	}
	return vxA, gpA, ivA, gvA
}

// ---------- flat / valid ----------

func BenchmarkFlatValid_VX(b *testing.B) {
	bound := vxPersonSchema.BindTypeOf(vxPerson{})
	data := vxPerson{Name: "Alice", Age: 30, Code: "AB1234"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = bound.Check(&data)
	}
}

func BenchmarkFlatValid_GoPlayground(b *testing.B) {
	data := gpPerson{Name: "Alice", Age: 30, Code: "AB1234"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gpValidate.Struct(data)
	}
}

func BenchmarkFlatValid_Invopop(b *testing.B) {
	data := ivPerson{Name: "Alice", Age: 30, Code: "AB1234"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = data.Validate()
	}
}

func BenchmarkFlatValid_Govy(b *testing.B) {
	data := gvPerson{Name: "Alice", Age: 30, Code: "AB1234"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gvPersonValidator.Validate(data)
	}
}

// ---------- flat / invalid (error path) ----------

func BenchmarkFlatInvalid_VX(b *testing.B) {
	bound := vxPersonSchema.BindTypeOf(vxPerson{})
	data := vxPerson{Name: "", Age: -1, Code: "not-a-code"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = bound.Check(&data)
	}
}

func BenchmarkFlatInvalid_GoPlayground(b *testing.B) {
	data := gpPerson{Name: "", Age: -1, Code: "not-a-code"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gpValidate.Struct(data)
	}
}

func BenchmarkFlatInvalid_Invopop(b *testing.B) {
	data := ivPerson{Name: "", Age: -1, Code: "not-a-code"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = data.Validate()
	}
}

func BenchmarkFlatInvalid_Govy(b *testing.B) {
	data := gvPerson{Name: "", Age: -1, Code: "not-a-code"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gvPersonValidator.Validate(data)
	}
}

// ---------- nested / valid, 5 addresses ----------

func BenchmarkNestedValid_VX(b *testing.B) {
	vxA, _, _, _ := makeAddresses(5)
	bound := vxCustomerSchema.BindTypeOf(vxCustomer{})
	data := vxCustomer{Name: "Acme", Addresses: vxA}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = bound.Check(&data)
	}
}

func BenchmarkNestedValid_GoPlayground(b *testing.B) {
	_, gpA, _, _ := makeAddresses(5)
	data := gpCustomer{Name: "Acme", Addresses: gpA}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gpValidate.Struct(data)
	}
}

func BenchmarkNestedValid_Invopop(b *testing.B) {
	_, _, ivA, _ := makeAddresses(5)
	data := ivCustomer{Name: "Acme", Addresses: ivA}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = data.Validate()
	}
}

func BenchmarkNestedValid_Govy(b *testing.B) {
	_, _, _, gvA := makeAddresses(5)
	data := gvCustomer{Name: "Acme", Addresses: gvA}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gvCustomerValidator.Validate(data)
	}
}
