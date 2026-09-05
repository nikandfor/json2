package jq

import (
	"testing"

	"nikand.dev/go/json2"
)

type (
	TestFilter [][]byte
)

func TestSimple(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	b, i, _, err := Dot{}.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, data, string(b))

	b, i, _, err = NewQuery("a").Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, `"b"`, string(b))

	b, i, _, err = NewQuery("c").Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, `4`, string(b))

	b, i, _, err = NewQuery("non").Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, `null`, string(b))

	b, i, _, err = NewQuery("d", 2).Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, `true`, string(b))

	b, i, _, err = NewQuery("d", 4, "f").Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, `3.4`, string(b))
}

func BenchmarkNextAll(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`[1,2,3,4,5,6]`)

	var err error
	var state State
	var st int
	var buf []byte

	f := Iter{}

	for i := 0; i < b.N; i++ {
		buf, st, state, err = f.Next(buf, data, st, state)
		if err != nil {
			b.Errorf("pipe: %v", err)
		}

		if state == nil {
			st = 0
			buf = buf[:0]
		}
	}
}

func (f TestFilter) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json2.Iterator

	i, err = p.Skip(r, st)
	if err != nil {
		return w, i, state, err
	}

	ss, _ := state.(int)
	if ss == len(f) {
		return w, i, nil, nil
	}

	w = append(w, f[ss]...)

	ss++
	state = ss

	return w, st, state, nil
}
