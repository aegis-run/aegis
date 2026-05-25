package assert

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// Equal asserts that two values are equal.
func Equal[T any](tb testing.TB, actual, expected T) {
	tb.Helper()

	if areEqual(actual, expected) {
		return
	}

	tb.Errorf("not equal:\n  actual:   %#v\n  expected: %#v", actual, expected)
}

// Err asserts that an error matches expectations.
//
// Supported expectations:
//   - nil → expect no error
//   - string → substring match
//   - error → errors.Is
//   - reflect.Type → errors.As
func Err(tb testing.TB, actual error, expecteds ...any) {
	tb.Helper()

	if len(expecteds) > 1 {
		tb.Fatalf("assert.Err accepts at most one expectation, got %d", len(expecteds))
	}

	if len(expecteds) == 0 {
		if actual == nil {
			tb.Error("expected an error, got nil")
		}
		return
	}

	matchError(tb, actual, expecteds[0])
}

func matchError(tb testing.TB, actual error, expected any) {
	tb.Helper()

	switch exp := expected.(type) {
	case nil:
		if actual != nil {
			tb.Fatalf("unexpected error: %v", actual)
		}

	case string:
		assertErrorContains(tb, actual, exp)

	case error:
		assertErrorIs(tb, actual, exp)

	case reflect.Type:
		assertErrorAs(tb, actual, exp)

	default:
		tb.Fatalf("unsupported expected type: %T", expected)
	}
}

func assertErrorContains(tb testing.TB, actual error, expected string) {
	tb.Helper()
	if actual == nil {
		tb.Fatalf("expected error containing %q, got nil", expected)
	}
	if !strings.Contains(actual.Error(), expected) {
		tb.Errorf("error mismatch:\n  actual:   %q\n  expected: %q", actual.Error(), expected)
	}
}

func assertErrorIs(tb testing.TB, actual error, expected error) {
	tb.Helper()
	if actual == nil {
		tb.Fatalf("expected error %v, got nil", expected)
	}
	if !errors.Is(actual, expected) {
		tb.Errorf("error mismatch:\n  actual:   %v\n  expected: %v", actual, expected)
	}
}

func assertErrorAs(tb testing.TB, actual error, expected reflect.Type) {
	tb.Helper()
	if actual == nil {
		tb.Fatalf("expected error of type %v, got nil", expected)
	}
	target := reflect.New(expected).Interface()
	if !errors.As(actual, target) {
		tb.Errorf("error type mismatch:\n  actual:   %T\n  expected: %v", actual, expected)
	}
}

// True asserts that a boolean value is true.
func True(tb testing.TB, actual bool) {
	tb.Helper()

	if !actual {
		tb.Error("expected true, got false")
	}
}

// --- internals ---

type equaler[T any] interface {
	Equal(T) bool
}

func areEqual[T any](lhs, rhs T) bool {
	if isNil(lhs) && isNil(rhs) {
		return true
	}

	if eq, ok := any(lhs).(equaler[T]); ok {
		return eq.Equal(rhs)
	}

	if lhsBytes, ok := any(lhs).([]byte); ok {
		if rhsBytes, ok := any(rhs).([]byte); ok {
			return bytes.Equal(lhsBytes, rhsBytes)
		}
		return false
	}

	return reflect.DeepEqual(lhs, rhs)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice,
		reflect.UnsafePointer:
		return rv.IsNil()
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.Array, reflect.String, reflect.Struct:
		return false
	default:
		return false
	}
}
