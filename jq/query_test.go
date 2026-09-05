package jq

import (
	"testing"
)

func TestQuery(tb *testing.T) {
	r := []byte(`{"a": [{"b":[1, "2"]}, {"b":[3]}, {"b": {"c": 4, "d": "5"}}]}`)

	f := NewQuery("a", Iter{}, "b", Iter{})

	var w []byte
	var err error
	var state any
	i := 0

	for j, exp := range []string{`1`, `"2"`, `3`, `4`, `"5"`} {
		assertEqual(tb, j == 0, state == nil)

		w, i, state, err = f.Next(w[:0], r, i, state)
		assertBytesErr(tb, r, i, err)

		assertEqual(tb, exp, string(w))

		if tb.Failed() {
			return
		}
	}

	w, i, state, err = f.Next(w[:0], r, i, state)
	assertBytesErr(tb, r, i, err)

	assertNil(tb, state)
	assertEmpty(tb, w)
	assertEqual(tb, len(r), i)
}

func TestQueryEmpty(tb *testing.T) {
	r := []byte(`{"results":[],"key":"b"}`)
	f := NewQuery()

	w, i, err := NextAll(f, nil, r, 0, nil)
	assertBytesErr(tb, r, i, err)
	assertEqual(tb, len(r), i)
	assertEqual(tb, r, w)
}

func TestQueryIterEmpty(tb *testing.T) {
	var w []byte
	r := []byte(`{"a":[],"b":{}}`)

	f := NewQuery("a", Iter{})

	w, _, state, err := f.Next(w[:0], r, 0, nil)
	assertNoError(tb, err)
	assertNil(tb, state)
	assertLen(tb, w, 0)

	f = NewQuery("b", Iter{})

	w, _, state, err = f.Next(w[:0], r, 0, nil)
	assertNoError(tb, err)
	assertNil(tb, state)
	assertLen(tb, w, 0)
}

func assertBytesErr(tb *testing.T, r []byte, i int, err error) {
	if assertNoError(tb, err) {
		return
	}

	pref := r[:i]
	suff := r[i:]

	tb.Logf("%s|%s", pref, suff)
}
