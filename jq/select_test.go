package jq

import (
	"testing"
)

func TestSelect(t *testing.T) {
	var state State
	var err error
	var w []byte
	var i int

	data := `1 null "a" [4] {"b":true}`

	for _, exp := range []string{
		"1", ``, `"a"`, `[4]`, `{"b":true}`,
	} {
		w, i, state, err = NewSelect(nil).Next(w[:0], []byte(data), i, state)
		assertNoError(t, err)
		//	assertLen(t, data, i)
		assertEqual(t, exp, string(w))
	}

	assertNil(t, state)
}

func TestMap(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	f := Map{
		Filter: NewComma(
			Literal(`5`),
		),
	}

	w, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertNil(t, state)
	assertEqual(t, len(data), i)
	assertEqual(t, `[5,5,5,5,5]`, string(w))

	f = Map{
		Filter: NewComma(
			Literal(`5`),
			Literal(`6`),
		),
	}

	w, i, state, err = f.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertNil(t, state)
	assertEqual(t, len(data), i)
	assertEqual(t, `[5,6,5,6,5,6,5,6,5,6]`, string(w))
}

func TestMapSelectEqual(t *testing.T) {
	data := `[{"a":"b"},{"a":"c"},{"a":"b"}]`

	f := NewMap(NewSelect(NewNotEqual(
		NewQuery("a"),
		Literal(`"c"`),
	)))

	w, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertNil(t, state)
	assertEqual(t, len(data), i)
	assertEqual(t, `[{"a":"b"},{"a":"b"}]`, string(w))
}
