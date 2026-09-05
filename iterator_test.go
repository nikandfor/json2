package json2

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestIterator(t *testing.T) {
	var d Iterator

	for _, data := range []string{
		"null", "true", "false",
		"1", "1.1", "1e1", "+1", "-1", "-1.4", "0x1p+1", "-0x1p-2", "0x3", "0xf", "0XF",
		`""`, `"a"`, `"abc def"`, `"a\"b\nc\td"`,
		"[]", "[1, 2, 3]", `[null, "str"]`,
		"{}", `{"key":"val"}`, `{"k": "v", "k2": 3, "k3": [], "k4": {}, "k5": null}`,
	} {
		raw, i, err := d.Raw([]byte(data), 0)
		if !assertNoError(t, err) || !assertEqual(t, len(data), i) || !assertEqual(t, []byte(data), raw) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, data)
		}
	}
}

func TestIteratorString(t *testing.T) {
	var d Iterator

	for j, data := range []string{
		`""`, `"a"`, `"a\"b\nc\tde\"f\\g"`,
		//	"\"a\u0016b\"",
		`"\u00ab\u00ac\u00f3"`,
		`"\u0100\u017e"`,
	} {
		var q string

		err := json.Unmarshal([]byte(data), &q)
		assertNoError(t, err)

		i, err := d.Skip([]byte(data), 0)
		if !assertNoError(t, err) || !assertEqual(t, len(data), i) {
			t.Logf("pos: %d (%[1]x)  data: %d %q", i, j, data)
			continue
		}

		s, i, err := d.DecodeString([]byte(data), 0, nil)
		if !assertNoError(t, err) || !assertEqual(t, len(data), i) {
			t.Logf("pos: %d (%[1]x)  data: %d %q", i, j, data)
			continue
		}

		assertEqual(t, q, string(s))

		if false {
			q, err := strconv.Unquote(data)
			assertNoError(t, err)

			assertEqual(t, q, string(s))
		}
	}
}

func TestIteratorEnterMore(t *testing.T) {
	var d Iterator

cases:
	for _, data := range []string{
		`{}`, `{ }`, `{"a":"b"}`, `{"a": "b", "c": 4}`,
		`[]`, `[ ]`, `["a"]`, `[2]`, `[ "a", 2, null ]`,
	} {
		b := []byte(data)

		i, err := d.Enter(b, 0, Type(data[0]))
		if !assertNoError(t, err) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, data)
			continue
		}

		for d.ForMore(b, &i, Type(data[0]), &err) {
			if data[0] == '{' {
				_, i, err = d.Key(b, i)
				if !assertNoError(t, err) {
					continue cases
				}
			}

			i, err = d.Skip(b, i)
			if !assertNoError(t, err) {
				continue cases
			}
		}

		if !assertNoError(t, err) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, data)
			continue
		}
	}
}
