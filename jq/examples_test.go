package jq_test

import (
	"encoding/base64"
	"fmt"

	"nikand.dev/go/json/jq"
)

func ExampleIndex() {
	data := []byte(`{"key0":"skip it", "key1": {"next_key": ["array", null, {"obj":"val"}, "trailing element"]}}  "next"`)

	f := jq.Index{"key1", "next_key", 2} // string keys and int array indexes are supported

	var res []byte // reusable buffer
	var i int      // start index

	// Most filters only parse single value and return index where the value ended.
	// Use jq.ApplyToAll(f, res[:0], data, 0, []byte("\n")) to process all values in a buffer.
	res, i, _, err := f.Next(res[:0], data, i, nil)
	if err != nil {
		// i is an index in a source buffer where the error occurred.
	}

	fmt.Printf("value: %s\n", res)
	fmt.Printf("final position: %d of %d\n", i, len(data)) // filter only parsed first value in the buffer
	_ = i < len(data)                                      // and stopped immideately after it

	// Output:
	// value: {"obj":"val"}
	// final position: 92 of 100
}

func ExampleBase64d() {
	// generated by command
	// jq -nc '{key3: "value"} | {key2: (. | tojson)} | @base64 | {key1: .}'
	data := []byte(`{"key1":"eyJrZXkyIjoie1wia2V5M1wiOlwidmFsdWVcIn0ifQ=="}`)

	f := jq.NewPipe(
		jq.Index{"key1"},
		&jq.Base64d{
			Encoding: base64.StdEncoding,
		},
		&jq.JSONDecoder{},
		jq.Index{"key2"},
		&jq.JSONDecoder{},
		jq.Index{"key3"},
	)

	res, _, _, err := f.Next(nil, data, 0, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", res)

	// Output:
	// "value"
}
