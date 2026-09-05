package jq

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func assertNoError(tb testing.TB, err error) bool {
	tb.Helper()

	if err == nil {
		return true
	}

	tb.Errorf("unexpected error: %v", err)

	return false
}

func assertEqual(tb testing.TB, exp, act any) bool {
	tb.Helper()

	if equalValues(exp, act) {
		return true
	}

	tb.Errorf("expected %v, got %v", printable(exp), printable(act))

	return false
}

func assertTrue(tb testing.TB, v bool, args ...any) bool {
	tb.Helper()

	if v {
		return true
	}

	tb.Errorf("expected true%v", message(args))

	return false
}

func assertNil(tb testing.TB, x any) bool {
	tb.Helper()

	if isNilValue(x) {
		return true
	}

	tb.Errorf("expected nil, got %v", printable(x))

	return false
}

func assertNotNil(tb testing.TB, x any) bool {
	tb.Helper()

	if !isNilValue(x) {
		return true
	}

	tb.Errorf("expected not nil")

	return false
}

func assertLen(tb testing.TB, x any, n int) bool {
	tb.Helper()

	l := reflect.ValueOf(x).Len()
	if l == n {
		return true
	}

	tb.Errorf("expected length %d, got %d", n, l)

	return false
}

func assertEmpty(tb testing.TB, x any) bool {
	tb.Helper()

	if isNilValue(x) || reflect.ValueOf(x).Len() == 0 {
		return true
	}

	tb.Errorf("expected empty, got %v", printable(x))

	return false
}

// message formats the optional assertion message: a format string and its arguments.
func message(args []any) string {
	if len(args) == 0 {
		return ""
	}

	return "; " + fmt.Sprintf(args[0].(string), args[1:]...)
}

func isNilValue(x any) bool {
	if x == nil {
		return true
	}

	r := reflect.ValueOf(x)

	switch r.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return r.IsNil()
	}

	return false
}

func equalValues(x, y any) bool {
	if xb, ok := x.([]byte); ok {
		yb, ok := y.([]byte)

		return ok && bytes.Equal(xb, yb)
	}

	return reflect.DeepEqual(x, y)
}

func printable(x any) any {
	if b, ok := x.([]byte); ok {
		return string(b)
	}

	return x
}
