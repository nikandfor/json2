package json2

import (
	"bytes"
	"errors"
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

func assertErrorIs(tb testing.TB, err, target error) bool {
	tb.Helper()

	if errors.Is(err, target) {
		return true
	}

	tb.Errorf("expected error %v, got %v", target, err)

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
