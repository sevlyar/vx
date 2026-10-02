# vx

[![CI](https://github.com/sevlyar/vx/actions/workflows/ci.yml/badge.svg)](https://github.com/sevlyar/vx/actions/workflows/ci.yml)

Validation for Go values of any shape — scalars, whole structs, and the nested structs, slices, and maps inside them — built from small, composable checks. No dependencies beyond the standard library.

```go
import "github.com/sevlyar/vx"
```

## Why

- **Schemas are values you build once, not graphs rebuilt per call.** `vx.Structure(&zeroValue, vx.Field(&zeroValue.Age, vx.Ge(0)))` only needs `zeroValue` to resolve field addresses by pointer; the result is a `vx.Schema` you can store in a package-level var and bind once, then check any number of actual values. Libraries that take a pointer to the *live* instance being validated (e.g. `ozzo`/`invopop`'s `validation.Field(&a.Age, ...)`) can't do this — their rule graph is reconstructed on every single call.
- **No dependencies beyond the standard library** — not even for format or regexp checks. Compare `go-playground/validator`'s 7 direct dependencies (including a full locale/translation stack) or `invopop/validation`'s 1.
- **Checks compose, and a failure's path shows the whole chain, not one flat rule.** `Len(Gt(3))`, `AllOf`/`AnyOf`/`OneOf` nest arbitrarily; `SchemaPath` reconstructs the full chain that fired, e.g. `Field(Name).Len.Gt(3)`, and `DataPath` the exact place in the data, e.g. `Items[2].Name` — no error-string parsing needed for either.
- **Field selectors and arguments are regular Go code, checked by the compiler.** `vx.Field(&user.Age, vx.Gt(0))` — a typo or type mismatch is a build error, not a validation tag that silently never matches at runtime.
- **Measured faster, with fewer allocations**, against both libraries named below, on every scenario benchmarked so far — see [Performance](#performance).

See [below](#compared-to-other-validators) for how these hold up against `go-playground/validator` and `invopop/validation` specifically, and what vx doesn't have (yet).

## Quick start

```go
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
// Field(Name).Len.Gt(0): invalid Name: value should be greater than 0
```

`Structure` requires a `Field` option for every field; use `vx.AllowUncheckedFields(&p.SomeField)` to opt specific fields out deliberately, rather than leaving them unchecked by accident.

## Binding: early vs. late

A `Schema` is a function from a `reflect.Type` to a `BoundSchema`. How you bind it decides when a type mismatch is caught:

- `schema.BindTypeOf(v)` binds **early**, against the concrete type of `v`. A schema that doesn't support that type panics immediately, at startup — not buried in a request handler. This is what `Structure`/`Field` use internally for each field, since a field's type is always known.
- `schema.BindAny()` binds **late**, against any type. A type mismatch is returned as a regular error from `Check`, since the actual value is only known at validation time (e.g. a field typed `any`, or data decoded from JSON).

## Inspecting a failure

Every failure is either a `*vx.CompoundCheckError` (a check that delegates to a nested schema: `Structure`, `Item`, `Len`, `AllOf`/`AnyOf`/`OneOf`) or a `*vx.CheckError` (a leaf check with nothing to delegate to: `Gt`, `In`, `Format`, ...). Both are reachable through the standard `errors.Is`/`errors.As`.

```go
var ce *vx.CompoundCheckError
if errors.As(err, &ce) {
    ce.SchemaPath() // []string{"Field(Items)", "Item", "Field(Name)", "Len", "Gt(3)"}
    ce.DataPath()   // []vx.PathElement{...} -> vx.RenderPath(...) == "Items[1].Name"
}
```

## Checks

**Navigate the data**

| Check | Checks |
|---|---|
| `Structure(ptr, Field(...), ...)` | a struct's fields, addressed by pointer |
| `Item(schema)` | every element of an array or slice |

**A derived aspect of the value**

| Check | Checks |
|---|---|
| `Len(schema)` | the length of a string (rune count), array, slice or map |

**Leaf predicates**

| Check | Checks |
|---|---|
| `Gt`, `Ge`, `Lt`, `Le` | a number against a threshold |
| `In(values...)` | membership in a fixed set |
| `Empty`, `NonEmpty` | the value is/isn't the zero value for its type |
| `Format(name, isValid)`, `RegexpFormat(pattern)` | a string against a predicate or regexp |
| `PrintableLine`, `PrintableText` | a string has only printable runes (the latter also allows line breaks) |

**Combine schemas**

| Check | Checks |
|---|---|
| `AllOf(schemas...)` | the value against every schema |
| `AnyOf(schemas...)` | the value against each schema until one matches |
| `OneOf(schemas...)` | the value matches exactly one schema |

See the [package documentation](https://pkg.go.dev/github.com/sevlyar/vx) for runnable examples of each.

## Writing your own checks

A check is just a `vx.Schema` — `func(reflect.Type) vx.BoundSchema`. Two exported building blocks cover the two shapes every check in this package follows:

- **`vx.BindTypeCheck(check, validate)`** for a leaf check, one with no nested schema (like `Gt` or `Format`). `check` rejects types the check doesn't support; `validate` does the actual work and returns a `*vx.CheckError` on failure.

  ```go
  func MultipleOf(n int) vx.Schema {
      mustBeInt := func(t reflect.Type) error {
          if t.Kind() != reflect.Int {
              return errors.New("type must be int")
          }
          return nil
      }
      err := &vx.CheckError{
          CheckName: "MultipleOf",
          Param:     fmt.Sprint(n), // -> "MultipleOf(5)" in SchemaPath
          Msg:       fmt.Sprintf("value must be a multiple of %d", n),
      }
      return vx.BindTypeCheck(mustBeInt, func(v reflect.Value) error {
          if v.Int()%int64(n) != 0 {
              return err
          }
          return nil
      })
  }
  ```

- **`vx.BindCompound(check, early, late)`** for a check that delegates to a nested schema reached through some transformation of the value (like `Item` reaching into a slice's elements, or `Len` into its length). Wrap the nested failure in a `*vx.CompoundCheckError`, picking a `vx.PathElement` for `DataItem`: `FieldElement`/`IndexElement` if one fits, or the general-purpose `KeyElement` otherwise (e.g. for a map key, which is neither).

Both keep the early/late binding contract: a type mismatch panics when the type is known at bind time, and is returned as an error when it's only known at `Check` time (see "Binding" above) — exactly what hand-rolling the two branches yourself would get you, without having to get it right by hand.

See [`extend_test.go`](extend_test.go) for both worked end to end, including a `Values` check that validates every value of a map the way `Item` validates a slice.

## Compared to other validators

Checked against [`go-playground/validator`](https://github.com/go-playground/validator) (the dominant struct-tag library) and [`invopop/validation`](https://github.com/invopop/validation) (the maintained `ozzo-validation` fork — closest in spirit: Go code and pointer-addressed fields, no tags), against their own source, not secondhand claims:

✅ clear advantage · ⚠️ partial / limited · ❌ clear gap — unmarked rows are a paradigm choice, not a win or a loss.

| | `go-playground/validator` | `invopop/validation` | `vx` |
|---|---|---|---|
| Rule is | a string tag, `validate:"gt=0"` | Go code | Go code |
| Field addressed by | tag on the field | pointer, per live instance | pointer, resolved once via a throwaway zero value |
| Schema reusable across calls? | ✅ yes (caches parsed tags per type) | ❌ no — rebuilt on every call | ✅ yes, explicitly, as a value |
| Failure path | ⚠️ `Namespace()`, e.g. `"User.Addresses[0].Street"`, but one flat `Tag()`+`Param()` | ❌ none — a (possibly nested) `map[string]error` | ✅ `SchemaPath`/`DataPath`, full nested chain |
| Rule composition | ⚠️ flat, comma-separated (AND), limited `\|` (OR) per tag | ❌ independent rules, no nesting | ✅ arbitrary nesting: `Len(Gt(3))`, `AllOf`/`AnyOf`/`OneOf` |
| Direct dependencies | ❌ 7 (locales, universal-translator, go-urn, mimetype, x/crypto, x/text, assert) | ⚠️ 1 (govalidator) | ✅ 0 |
| Rule checked by the compiler? | ❌ no — a malformed tag fails silently or at runtime | ✅ yes | ✅ yes |

### Performance

Benchmarked in an isolated module — not a dependency of `vx` itself, see "no dependencies" above — against the same two libraries, validating a flat struct and a struct with a nested 5-item slice. The one format check in each scenario uses the *exact same compiled regexp* in all three libraries: an earlier pass instead used each library's own built-in email validator and showed a bigger gap, which wasn't a fair reading — `go-playground`'s and `invopop`'s built-in email checks are more thorough (and so more expensive) than a plain regexp. This version isolates framework overhead from validation thoroughness.

`go1.24.4 darwin/arm64, Apple M3`, `go test -bench=. -benchmem -count=3`, numbers stable across runs:

| Scenario | `vx` | `go-playground/validator` | `invopop/validation` |
|---|---|---|---|
| Flat struct, valid | **84 ns/op, 0 allocs** | 156 ns/op, 0 allocs (1.9×) | 560–590 ns/op, 18 allocs (7×) |
| Flat struct, invalid | **47 ns/op, 2 allocs** | 347 ns/op, 10 allocs (7.4×) | 680–720 ns/op, 22 allocs (15×) |
| Nested struct + 5-item slice, valid | **390 ns/op, 1 alloc** | 1052–1058 ns/op, 22 allocs (2.7×) | 4188–4194 ns/op, 142 allocs (10.7×) |

Two caveats, honestly:

- The "invalid" row isn't purely framework overhead. `vx`'s `Structure` stops at the first failing field by design ("first failure wins"); the other two collect every field's errors by default. Part of that gap is a difference in what's being done, not just how fast.
- One machine, one run of three scenarios — not a broad performance suite. `invopop/validation`'s much higher allocation count does match the architectural gap noted in the table above: it rebuilds its rule graph on every call, `vx` and `go-playground/validator` don't.

What `vx` doesn't have, honestly:

- ❌ **No struct-tag / data-driven mode.** Rules are Go code; if you need to change validation without recompiling (e.g. rules loaded from JSON/config), this isn't it.
- ❌ **No built-in i18n/translation of error messages**, unlike `go-playground/validator`'s locale stack.
- ❌ **Smaller built-in rule set and a much smaller community** — this is a new library, not a battle-tested one with years of edge cases shaken out.

## Install

```sh
go get github.com/sevlyar/vx
```

## License

MIT, see [LICENSE](LICENSE).
