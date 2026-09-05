package json2

import (
	"testing"
)

func TestIteratorSeek(tb *testing.T) {
	var d Iterator

	for _, tc := range []struct {
		In   string
		Out  string
		Err  error
		Path []interface{}
	}{
		{In: `1`, Out: `1`, Path: []interface{}{}},

		{In: `{"a":1}`, Out: `1`, Path: []interface{}{"a"}},
		{In: `{"a":{"b":"c"}}`, Out: `"c"`, Path: []interface{}{"a", "b"}},
		{In: `{"a":{"b":"c", "d": "e", "f": [0, 1]}}`, Out: `[0, 1]`, Path: []interface{}{"a", "f"}},

		{In: `[0,1,2]`, Out: `0`, Path: []interface{}{0}},
		{In: `[0,1,2]`, Out: `1`, Path: []interface{}{1}},
		{In: `[0,1,2, 3, 4, 5]`, Out: `5`, Path: []interface{}{5}},

		{In: `[0,1,2]`, Out: `2`, Path: []interface{}{-1}},
		{In: `["a", "b", "c"]`, Out: `"a"`, Path: []interface{}{-3}},

		{In: `{"a":{"b":[{"c": "d"}, {"c": [6]}]}}`, Out: `[6]`, Path: []interface{}{"a", "b", 1, "c"}},

		{In: `{"a":{"b":"c"}}`, Err: ErrNoSuchKey, Path: []interface{}{"a", "c"}},
	} {
		st, err := d.Seek([]byte(tc.In), 0, tc.Path...)
		if tc.Err == nil {
			assertNoError(tb, err)
		} else {
			assertErrorIs(tb, err, tc.Err)
			continue
		}

		end, err := d.Skip([]byte(tc.In), st)
		assertNoError(tb, err)

		assertEqual(tb, tc.Out, tc.In[st:end])
	}
}
