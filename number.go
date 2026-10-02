package vx

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"reflect"
)

// Gt checks the value is greater than x.
func Gt(x any) Schema {
	return newComparison("Gt", x, func(sign int) bool { return sign > 0 },
		fmt.Sprintf("value should be greater than %v", x))
}

// Ge checks the value is greater than or equal to x.
func Ge(x any) Schema {
	return newComparison("Ge", x, func(sign int) bool { return sign >= 0 },
		fmt.Sprintf("value should be greater or equal than %v", x))
}

// Lt checks the value is less than x.
func Lt(x any) Schema {
	return newComparison("Lt", x, func(sign int) bool { return sign < 0 },
		fmt.Sprintf("value should be less than %v", x))
}

// Le checks the value is less than or equal to x.
func Le(x any) Schema {
	return newComparison("Le", x, func(sign int) bool { return sign <= 0 },
		fmt.Sprintf("value should be less or equal than %v", x))
}

func newComparison(name string, x any, accept func(sign int) bool, msg string) Schema {
	xVal := reflect.ValueOf(x)
	if !isNumericKind(xVal.Kind()) {
		panic("value must be number")
	}
	err := newCheckError(name, fmt.Sprint(x), msg)
	return BindTypeCheck(mustBeNumber, func(v reflect.Value) error {
		if accept(compareNumeric(v, xVal)) {
			return nil
		}
		return err
	})
}

// compareNumeric returns -1, 0 or 1 as v is less than, equal to, or greater
// than x. v and x may be of different numeric kinds (int/uint/float); an
// out-of-range int-uint comparison is resolved by sign and magnitude rather
// than by a lossy conversion between the two.
func compareNumeric(v, x reflect.Value) int {
	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compareToInt(v, x.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compareToUint(v, x.Uint())
	default:
		return compareToFloat(v, x.Float())
	}
}

func compareToInt(v reflect.Value, x int64) int {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(v.Int(), x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u := v.Uint(); u <= math.MaxInt64 {
			return cmp.Compare(int64(u), x)
		}
		return 1 // v overflows int64, so it is greater than any int64 x
	default:
		return cmp.Compare(v.Float(), float64(x))
	}
}

func compareToUint(v reflect.Value, x uint64) int {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i := v.Int(); i >= 0 {
			return cmp.Compare(uint64(i), x)
		}
		return -1 // negative v is less than any uint64 x
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cmp.Compare(v.Uint(), x)
	default:
		return cmp.Compare(v.Float(), float64(x))
	}
}

func compareToFloat(v reflect.Value, x float64) int {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(float64(v.Int()), x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cmp.Compare(float64(v.Uint()), x)
	default:
		return cmp.Compare(v.Float(), x)
	}
}

func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func mustBeNumber(t reflect.Type) error {
	if !isNumericKind(t.Kind()) {
		return errors.New("type must be int or float")
	}
	return nil
}
