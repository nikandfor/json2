package json2

import (
	"encoding/json"
	"testing"
)

func TestReader(t *testing.T) {
	var r Reader

	for _, data := range []string{
		"null", "true", "false",
		"1", "1.1", "1e1", "+1", "-1", "-1.4", "0x1p+1", "-0x1p-2", "0x3", "0xf", "0XF",
		`""`, `"a"`, `"abc def"`, `"a\"b\nc\td"`,
		"[]", "[1, 2, 3]", `[null, "str"]`,
		"{}", `{"key":"val"}`, `{"k": "v", "k2": 3, "k3": [], "k4": {}, "k5": null}`,
	} {
		r.Reset([]byte(data), nil)

		raw, err := r.Raw()
		if !assertNoError(t, err) || !assertEqual(t, []byte(data), raw) {
			t.Logf("data: %q", data)
		}
	}
}

func TestReaderDecodeString(t *testing.T) {
	var r Reader

	for j, data := range []string{
		`""`, `"a"`, `"a\"b\nc\tde\"f\\g"`,
		//	`"\xab\xac\xf3"`,
		`"\u00ab\u00ac\u00f3"`,
		`"\u0100\u017e"`,
		//	`"\U00e4b896\U00e7958c"`,
	} {
		r.Reset([]byte(data), nil)

		s, err := r.DecodeString(nil)
		if !assertNoError(t, err) || !assertEqual(t, len(data), r.i) {
			t.Logf("pos: %d (%[1]x)  data: %d %q", r.i, j, data)
			continue
		}

		var q string

		err = json.Unmarshal([]byte(data), &q)
		assertNoError(t, err)
		assertEqual(t, q, string(s))
	}
}
