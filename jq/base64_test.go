package jq

import (
	"encoding/base64"
	"testing"
)

func TestBase64(t *testing.T) {
	data := `"ab\ncd"`

	e := Base64{
		Encoding: base64.RawStdEncoding,
	}

	res1, i, _, err := e.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertLen(t, data, i)
	assertEqual(t, `"YWIKY2Q"`, string(res1))

	d := Base64d{
		Encoding: base64.RawStdEncoding,
	}

	res2, i, _, err := d.Next(nil, res1, 0, nil)
	assertNoError(t, err)
	assertLen(t, res1, i)
	assertEqual(t, data, string(res2))
}
