package jq

import (
	"testing"
)

func TestCat(t *testing.T) {
	f := Cat{
		Separator: []byte("-"),
	}

	data := `"ama", "ena", "uma", "viva"`

	w, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertNil(t, state)
	assertLen(t, data, i)
	assertEqual(t, `"ama-ena-uma-viva"`, string(w))

	data = `"\nqwe 世"  "界\tend"`

	w, i, state, err = f.Next(nil, []byte(data), 0, nil)
	assertNoError(t, err)
	assertNil(t, state)
	assertLen(t, data, i)
	assertEqual(t, `"\nqwe 世-界\tend"`, string(w))
}
