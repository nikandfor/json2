package json2

import (
	"bytes"
	"strconv"
)

// MergePatch merges the patch into the base and appends the result to w.
// It works the same way as RFC 7386 JSON Merge Patch:
// objects are merged recursively, null in the patch deletes the key,
// any other value replaces the base one.
func MergePatch(w, base, patch []byte) ([]byte, error) {
	braw, err := rawValue(base)
	if err != nil {
		return w, err
	}

	praw, err := rawValue(patch)
	if err != nil {
		return w, err
	}

	return patchValue(w, braw, praw)
}

// Set sets the keys of the base object to the values and appends the result to w.
// kvs is a list of key and value pairs. Key is a string, nil value deletes the key.
// Value is a bool, a number, a string, []byte for a string, or Value/RawMessage for a raw json.
func Set(w, base []byte, kvs ...any) ([]byte, error) {
	if len(kvs)%2 != 0 {
		panic(len(kvs))
	}

	braw, err := rawValue(base)
	if err != nil {
		return w, err
	}

	if !isObject(braw) {
		braw = nil
	}

	w = append(w, '{')

	w, err = setBase(w, braw, kvs)
	if err != nil {
		return w, err
	}

	w, err = setRest(w, braw, kvs)
	if err != nil {
		return w, err
	}

	return append(w, '}'), nil
}

// setBase writes the base keys with the values replaced by the kvs ones.
func setBase(w, base []byte, kvs []any) ([]byte, error) {
	if base == nil {
		return w, nil
	}

	var k, v []byte

	i, err := dec.Enter(base, 0, Object)

	for err == nil && dec.ForMore(base, &i, Object, &err) {
		k, i, err = dec.Key(base, i)
		if err != nil {
			return w, err
		}

		v, i, err = dec.Raw(base, i)
		if err != nil {
			return w, err
		}

		j := kvsKey(kvs, k)
		if j < 0 {
			w = appendKey(w, k)
			w = append(w, v...)

			continue
		}

		if kvs[j+1] == nil {
			continue
		}

		w = appendKey(w, k)
		w = appendValue(w, kvs[j+1])
	}

	return w, err
}

// setRest writes the kvs keys missing from the base.
func setRest(w, base []byte, kvs []any) ([]byte, error) {
	for j := 0; j+1 < len(kvs); j += 2 {
		k := s2b(kvs[j].(string))

		if kvs[j+1] == nil || kvsKey(kvs, k) != j {
			continue
		}

		_, ok, err := objectKey(base, k)
		if err != nil {
			return w, err
		}
		if ok {
			continue
		}

		w = appendKey(w, k)
		w = appendValue(w, kvs[j+1])
	}

	return w, nil
}

func patchValue(w, base, patch []byte) ([]byte, error) {
	if !isObject(patch) {
		return append(w, patch...), nil
	}
	if !isObject(base) {
		base = nil
	}

	w = append(w, '{')

	w, err := patchBase(w, base, patch)
	if err != nil {
		return w, err
	}

	w, err = patchRest(w, base, patch)
	if err != nil {
		return w, err
	}

	return append(w, '}'), nil
}

// patchBase writes the base keys merged with the patch ones.
func patchBase(w, base, patch []byte) ([]byte, error) {
	if base == nil {
		return w, nil
	}

	var k, v, p []byte
	var ok bool

	i, err := dec.Enter(base, 0, Object)

	for err == nil && dec.ForMore(base, &i, Object, &err) {
		k, i, err = dec.Key(base, i)
		if err != nil {
			return w, err
		}

		v, i, err = dec.Raw(base, i)
		if err != nil {
			return w, err
		}

		p, ok, err = objectKey(patch, k)
		if err != nil {
			return w, err
		}

		if !ok {
			w = appendKey(w, k)
			w = append(w, v...)

			continue
		}

		if isNull(p) {
			continue
		}

		w = appendKey(w, k)

		w, err = patchValue(w, v, p)
		if err != nil {
			return w, err
		}
	}

	return w, err
}

// patchRest writes the patch keys missing from the base.
func patchRest(w, base, patch []byte) ([]byte, error) {
	var k, v []byte
	var ok bool

	i, err := dec.Enter(patch, 0, Object)

	for err == nil && dec.ForMore(patch, &i, Object, &err) {
		k, i, err = dec.Key(patch, i)
		if err != nil {
			return w, err
		}

		v, i, err = dec.Raw(patch, i)
		if err != nil {
			return w, err
		}

		if isNull(v) {
			continue
		}

		_, ok, err = objectKey(base, k)
		if err != nil {
			return w, err
		}
		if ok {
			continue
		}

		w = appendKey(w, k)

		w, err = patchValue(w, nil, v)
		if err != nil {
			return w, err
		}
	}

	return w, err
}

// objectKey looks the key up in the raw object and returns its raw value.
func objectKey(b, key []byte) (v []byte, ok bool, err error) {
	if b == nil {
		return nil, false, nil
	}

	var k []byte

	i, err := dec.Enter(b, 0, Object)

	for err == nil && dec.ForMore(b, &i, Object, &err) {
		k, i, err = dec.Key(b, i)
		if err != nil {
			return nil, false, err
		}

		if bytes.Equal(k, key) {
			v, _, err = dec.Raw(b, i)
			return v, err == nil, err
		}

		i, err = dec.Skip(b, i)
		if err != nil {
			return nil, false, err
		}
	}

	return nil, false, err
}

func kvsKey(kvs []any, key []byte) int {
	for j := 0; j+1 < len(kvs); j += 2 {
		if kvs[j].(string) == string(key) {
			return j
		}
	}

	return -1
}

// rawValue trims spaces and comments around the value. Empty buffer is no value.
func rawValue(b []byte) (raw []byte, err error) {
	if len(b) == 0 {
		return nil, nil
	}

	raw, _, err = dec.Raw(b, 0)

	return raw, err
}

// appendKey adds the comma if the previous value is already written.
func appendKey(w, k []byte) []byte {
	if w[len(w)-1] != '{' {
		w = append(w, ',')
	}

	w = append(w, '"')
	w = append(w, k...)
	w = append(w, '"', ':')

	return w
}

func appendValue(w []byte, v any) []byte {
	switch v := v.(type) {
	case nil:
		return append(w, "null"...)
	case bool:
		return strconv.AppendBool(w, v)
	case string:
		return enc.AppendString(w, s2b(v))
	case []byte:
		return enc.AppendString(w, v)
	case Value:
		return append(w, v...)
	case RawMessage:
		return append(w, v...)
	case int:
		return strconv.AppendInt(w, int64(v), 10)
	case int8:
		return strconv.AppendInt(w, int64(v), 10)
	case int16:
		return strconv.AppendInt(w, int64(v), 10)
	case int32:
		return strconv.AppendInt(w, int64(v), 10)
	case int64:
		return strconv.AppendInt(w, v, 10)
	case uint:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint8:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint16:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint32:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint64:
		return strconv.AppendUint(w, v, 10)
	case float32:
		return strconv.AppendFloat(w, float64(v), 'g', -1, 32)
	case float64:
		return strconv.AppendFloat(w, v, 'g', -1, 64)
	default:
		panic(v)
	}
}

func isObject(raw []byte) bool {
	return len(raw) != 0 && raw[0] == byte(Object)
}

func isNull(raw []byte) bool {
	return len(raw) != 0 && raw[0] == byte(Null)
}
