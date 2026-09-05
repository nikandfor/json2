package jq

import (
	"testing"
)

func TestComma(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewComma(
		NewQuery("a"),
		NewQuery("c"),
		NewQuery("d", 0),
	)

	var state State

	b, i, state, err := f.Next(nil, []byte(data), 0, state)
	assertNoError(t, err)
	assertNotNil(t, state)
	assertEqual(t, `"b"`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assertNoError(t, err)
	assertNotNil(t, state)
	assertEqual(t, `4`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assertNoError(t, err)
	assertNil(t, state)
	assertLen(t, data, i)
	assertEqual(t, `"e"`, string(b))
}

func TestCommaPipe(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewComma(
		NewPipe(NewQuery("a")),
		NewPipe(NewQuery("c")),
		NewPipe(NewQuery("d"), NewQuery(0)),
	)

	b, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertNotNil(t, state)
	//	assertLen(t, data, i)
	assertEqual(t, `"b"`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assertNoError(t, err)
	assertNotNil(t, state)
	//	assertLen(t, data, i)
	assertEqual(t, `4`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assertNoError(t, err)
	assertNil(t, state)
	assertLen(t, data, i)
	assertEqual(t, `"e"`, string(b))
}

func TestCommaIter(t *testing.T) {
	r := []byte(`{"a":[1,2,3],"b":[4,5]}`)

	//	f := NewArray(
	f := NewComma(
		NewQuery("a", Iter{}),
		NewQuery("b", Iter{}),
	//	),
	)

	var err error
	var state State
	var w []byte
	i := 0

	for _, tc := range []string{`1`, `2`, `3`, `4`, `5`} {
		w, i, state, err = f.Next(w[:0], r, i, state)
		t.Logf("comma iter  %s  %+v  %v", w, state, err)
		assertNoError(t, err)
		assertNotNil(t, state)
		assertEqual(t, tc, string(w))
	}

	w, i, state, err = f.Next(w[:0], r, i, state)
	assertNoError(t, err)
	assertNil(t, state)
	assertEqual(t, len(r), i)
	assertEmpty(t, w)
}
