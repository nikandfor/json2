package jq

import (
	"testing"
)

func TestKeyIndex(tb *testing.T) {
	data := []byte(`{"a":[{"b":"c"},{"b":"d"}]}`)

	var w []byte

	l0 := len(w)
	w, _, _, err := Key("a").Next(w, data, 0, nil)
	assertNoError(tb, err)

	tb.Logf("key a: (%s) %v", w, err)

	l1 := len(w)
	w, _, _, err = Index(1).Next(w, w, l0, nil)
	assertNoError(tb, err)

	tb.Logf("key a: (%s) %v", w, err)

	l2 := len(w)
	w, _, _, err = NewEqual(Key("b"), Literal(`"d"`)).Next(w, w, l1, nil)
	assertNoError(tb, err)

	tb.Logf("key a: (%s) %v", w, err)

	assertEqual(tb, `true`, string(w[l2:]))
}
