package json2

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

var (
	readerValues = []string{
		"null", "true", "false",
		"1", "1.1", "1e1", "+1", "-1", "-1.4", "0x1p+1", "-0x1p-2", "0x3", "0xf", "0XF",
		`""`, `"a"`, `"abc def"`, `"a\"b\nc\td"`, `"Ā Ž 世界 😀"`,
		"[]", "[1, 2, 3]", `[null, "str"]`,
		"{}", `{"key":"val"}`, `{"k": "v", "k2": 3, "k3": [], "k4": {}, "k5": null}`,
	}

	readerStrings = []string{
		`""`, `"a"`, `"a\"b\nc\tde\"f\\g"`,
		//	`"\xab\xac\xf3"`,
		`"\u00ab\u00ac\u00f3"`,
		`"\u0100\u017e"`,
		`"Ā Ž 世界 😀"`,
		//	`"\U00e4b896\U00e7958c"`,
	}
)

func TestReader(t *testing.T) {
	var r Reader

	for _, data := range readerValues {
		r.Reset([]byte(data), nil)

		raw, err := r.Raw()
		if !assertNoError(t, err) || !assertEqual(t, []byte(data), raw) {
			t.Logf("data: %q", data)
		}
	}
}

// TestReaderStream reads the same values a byte at a time so each of them is cut by a buffer refill.
func TestReaderStream(t *testing.T) {
	var r Reader

	for _, data := range readerValues {
		r.Reset(make([]byte, 0, 8), iotest.OneByteReader(strings.NewReader(data)))

		raw, err := r.Raw()
		if !assertNoError(t, err) || !assertEqual(t, []byte(data), raw) {
			t.Logf("data: %q", data)
		}
	}
}

func TestReaderDecodeString(t *testing.T) {
	var r Reader

	for j, data := range readerStrings {
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

func TestReaderStreamDecodeString(t *testing.T) {
	var r Reader

	for _, data := range readerStrings {
		r.Reset(make([]byte, 0, 8), iotest.OneByteReader(strings.NewReader(data)))

		s, err := r.DecodeString(nil)
		if !assertNoError(t, err) {
			t.Logf("data: %q", data)
			continue
		}

		var q string

		err = json.Unmarshal([]byte(data), &q)
		assertNoError(t, err)
		assertEqual(t, q, string(s))
	}
}

func TestReaderStreamError(t *testing.T) {
	errBroken := errors.New("broken pipe")

	var r Reader

	for _, data := range []string{"1.5", `"abc`, `{"a": 1.5`} {
		r.Reset(nil, io.MultiReader(strings.NewReader(data), iotest.ErrReader(errBroken)))

		_, err := r.Raw()
		if !assertErrorIs(t, err, errBroken) {
			t.Logf("data: %q", data)
		}
	}
}
