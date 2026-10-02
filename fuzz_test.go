package vx

import (
	"math/big"
	"reflect"
	"testing"
)

// fuzzValueGen turns a byte recipe into a nested any value: the kind of
// shape is irrelevant to correctness, only that late-bound checks (BindAny)
// must never panic no matter what they're handed.
type fuzzValueGen struct {
	recipe []byte
	pos    int
	leafI  int64
	leafS  string
}

func (g *fuzzValueGen) next() byte {
	if g.pos >= len(g.recipe) {
		return 0
	}
	b := g.recipe[g.pos]
	g.pos++
	return b
}

// build constructs a value from the recipe, capped at depth to guarantee
// termination regardless of what the recipe says.
func (g *fuzzValueGen) build(depth int) any {
	kind := g.next() % 8
	if depth <= 0 {
		kind %= 3 // force a leaf once the depth budget is spent
	}
	switch kind {
	case 0:
		return g.leafI
	case 1:
		return g.leafS
	case 2:
		return nil
	case 3:
		n := int(g.next() % 4)
		s := make([]any, n)
		for i := range s {
			s[i] = g.build(depth - 1)
		}
		return s
	case 4:
		n := int(g.next() % 4)
		m := make(map[string]any, n)
		for i := 0; i < n; i++ {
			m[string(rune('a'+i))] = g.build(depth - 1)
		}
		return m
	case 5:
		return struct{ X any }{X: g.build(depth - 1)}
	case 6:
		return []int{int(g.leafI), int(g.leafI) + 1}
	default:
		return g.leafI != 0
	}
}

// FuzzChecksNoPanic is the generalization of TestChecks_NestedInAnySlot:
// instead of two fixed (valid, invalid) values per check, it throws
// generated, arbitrarily-shaped values at every check through the same
// any-typed boundary (Item(check).BindAny() over a []any). Every check here
// is bound late, so by the package's own documented contract a type
// mismatch must come back as an error, never a panic — that is the one
// invariant this fuzz target asserts; it does not know what the "right"
// pass/fail answer is for random data.
func FuzzChecksNoPanic(f *testing.F) {
	f.Add([]byte{3, 2, 0, 1}, "hello", int64(5))
	f.Add([]byte{4, 1, 5, 0}, "", int64(0))
	f.Add([]byte{5, 5, 5, 5, 5}, "x", int64(-1))
	f.Add([]byte{6}, "", int64(0))
	f.Add([]byte{}, "", int64(0))

	type fuzzPerson struct{ Name any }
	personSchema := func() Schema {
		var p fuzzPerson
		return Structure(&p, Field(&p.Name, Gt(0)))
	}()

	checks := []Schema{
		Gt(0), Ge(0), Lt(10), Le(10),
		In(1, 2, 3),
		Empty, NonEmpty,
		Format("Upper", isUpper),
		RegexpFormat(`^\d+$`),
		PrintableLine, PrintableText,
		Len(Gt(3)),
		Item(Gt(0)),
		AllOf(Gt(0), Lt(10)),
		AnyOf(Lt(0), Gt(10)),
		OneOf(Lt(0), Gt(10)),
		personSchema,
	}

	f.Fuzz(func(t *testing.T, recipe []byte, leafS string, leafI int64) {
		g := &fuzzValueGen{recipe: recipe, leafI: leafI, leafS: leafS}
		value := g.build(4)

		for _, check := range checks {
			_ = Item(check).BindAny().Check([]any{value})
		}
	})
}

// FuzzNumericCompare checks compareNumeric against an independent oracle
// (math/big) for the exact case its own hand-written overflow handling
// exists for: an int64 compared against a uint64 threshold, which can't be
// done by converting either side to the other's native type without risking
// exactly the kind of overflow bug this is meant to catch.
func FuzzNumericCompare(f *testing.F) {
	f.Add(int64(5), uint64(3))
	f.Add(int64(-1), uint64(1))
	f.Add(int64(0), uint64(0))
	f.Add(int64(9223372036854775807), uint64(18446744073709551615))
	f.Add(int64(-9223372036854775808), uint64(0))

	f.Fuzz(func(t *testing.T, i int64, u uint64) {
		got := compareNumeric(reflect.ValueOf(i), reflect.ValueOf(u))

		want := big.NewInt(i).Cmp(new(big.Int).SetUint64(u))

		if sign(got) != sign(want) {
			t.Errorf("compareNumeric(int64(%d), uint64(%d)) sign = %d, want %d", i, u, sign(got), sign(want))
		}
	})
}

func sign(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	default:
		return 0
	}
}

// FuzzPrintableLineImpliesPrintableText checks a property that holds
// independently of either function's implementation: PrintableText is
// documented as accepting everything PrintableLine does, plus line breaks,
// so anything PrintableLine accepts, PrintableText must accept too.
func FuzzPrintableLineImpliesPrintableText(f *testing.F) {
	f.Add("hello")
	f.Add("line1\nline2")
	f.Add("line1 line2")
	f.Add("\x00")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		if isPrintableLine(s) && !isPrintableText(s) {
			t.Errorf("isPrintableLine(%q) = true but isPrintableText(%q) = false", s, s)
		}
	})
}
