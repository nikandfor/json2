package json2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergePatch(tb *testing.T) {
	for _, tc := range []struct {
		Base  string
		Patch string
		Out   string
	}{
		{Base: `{"a":1}`, Patch: `2`, Out: `2`},
		{Base: `1`, Patch: `{"a":2}`, Out: `{"a":2}`},
		{Base: ``, Patch: `{"a":2,"b":null}`, Out: `{"a":2}`},

		{Base: `{"a":1,"b":2}`, Patch: `{"b":3,"c":4}`, Out: `{"a":1,"b":3,"c":4}`},
		{Base: `{"a":1,"b":2}`, Patch: `{"b":null}`, Out: `{"a":1}`},
		{Base: `{"a":1}`, Patch: `{"b":null}`, Out: `{"a":1}`},

		{Base: `{"a":{"b":1,"c":2}}`, Patch: `{"a":{"c":3,"d":4}}`, Out: `{"a":{"b":1,"c":3,"d":4}}`},
		{Base: `{"a":{"b":1}}`, Patch: `{"a":[1,2]}`, Out: `{"a":[1,2]}`},
		{Base: `{"a":[1,2]}`, Patch: `{"a":{"b":1}}`, Out: `{"a":{"b":1}}`},
		{Base: `{"a":1}`, Patch: `{"b":{"c":2,"d":null}}`, Out: `{"a":1,"b":{"c":2}}`},

		{Base: ` { "a" : 1 } `, Patch: ` { "b" : { } } `, Out: `{"a":1,"b":{}}`},
	} {
		w, err := MergePatch(nil, []byte(tc.Base), []byte(tc.Patch))
		assert.NoError(tb, err)
		assert.Equal(tb, tc.Out, string(w))
	}
}

func TestSet(tb *testing.T) {
	for _, tc := range []struct {
		Base string
		KVs  []any
		Out  string
	}{
		{Base: `{"a":1}`, Out: `{"a":1}`},
		{Base: ``, KVs: []any{"a", 1}, Out: `{"a":1}`},
		{Base: `4`, KVs: []any{"a", 1}, Out: `{"a":1}`},

		{Base: `{"a":1,"b":2}`, KVs: []any{"b", "c", "d", true}, Out: `{"a":1,"b":"c","d":true}`},
		{Base: `{"a":1,"b":2}`, KVs: []any{"a", nil}, Out: `{"b":2}`},
		{Base: `{"a":{"b":1}}`, KVs: []any{"a", RawMessage(`{"c":2}`)}, Out: `{"a":{"c":2}}`},

		{Base: `{"a":1}`, KVs: []any{"b", 1.5, "c", int64(-3), "d", uint(4)}, Out: `{"a":1,"b":1.5,"c":-3,"d":4}`},
		{Base: `{"a":1}`, KVs: []any{"b", "q\"w", "c", []byte("e\nf")}, Out: `{"a":1,"b":"q\"w","c":"e\nf"}`},
	} {
		w, err := Set(nil, []byte(tc.Base), tc.KVs...)
		assert.NoError(tb, err)
		assert.Equal(tb, tc.Out, string(w))
	}
}

func TestMergePatchAlloc(tb *testing.T) {
	base := []byte(`{"a":1,"b":{"c":2,"d":[3,4]},"e":"f"}`)
	patch := []byte(`{"b":{"c":5,"g":null},"h":true}`)
	w, q := make([]byte, 0, 256), make([]byte, 0, 256)
	var err error

	n := testing.AllocsPerRun(100, func() {
		w, err = MergePatch(w[:0], base, patch)
		if err != nil {
			tb.Fatalf("patch: %v", err)
		}

		q, err = Set(q[:0], base, "a", 2)
		if err != nil {
			tb.Fatalf("set: %v", err)
		}
	})

	assert.Equal(tb, 0., n)
}
