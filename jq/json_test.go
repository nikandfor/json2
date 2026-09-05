package jq

import (
	"testing"
)

func TestJSON(t *testing.T) {
	data := `"\"abcd\"" "1" "{\"a\":\"b\"}"`

	var e JSONDecoder

	res, i, _, err := e.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertEqual(t, `"abcd"`, string(res))

	res, i, _, err = e.Next(nil, []byte(data), i, nil)
	assertNoError(t, err)
	assertEqual(t, `1`, string(res))

	res, i, _, err = e.Next(nil, []byte(data), i, nil)
	assertNoError(t, err)
	assertEqual(t, `{"a":"b"}`, string(res))

	assertLen(t, data, i)
}
