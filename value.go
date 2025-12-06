package json2

import "strconv"

type (
	Value []byte
)

func (d Iterator) ShouldValue(b []byte, st int) Value {
	val, err := d.Value(b, st)
	if err != nil {
		return nil
	}

	return val
}

func (d Iterator) Value(b []byte, st int) (Value, error) {
	raw, _, err := d.Raw(b, st)
	return Value(raw), err
}

func (raw Value) Type() Type {
	tp, _, _ := dec.Type(raw, 0)
	return tp
}

func (raw Value) TypeErr() (Type, error) {
	tp, _, err := dec.Type(raw, 0)
	return tp, err
}

func (raw Value) ShouldBool() (v bool) {
	v, _ = raw.Bool()
	return v
}

func (raw Value) Bool() (v bool, err error) {
	if raw == nil {
		return false, ErrValue
	}

	_, err = dec.ExpectType(raw, 0, Bool)
	if err != nil {
		return false, err
	}

	v, err = strconv.ParseBool(string(raw))
	if err != nil {
		return false, err
	}

	return v, nil
}

func (raw Value) ShouldString() (v string) {
	v, _ = raw.String()
	return v
}

func (raw Value) String() (v string, err error) {
	if raw == nil {
		return "", ErrValue
	}

	_, err = dec.ExpectType(raw, 0, String)
	if err != nil {
		return "", err
	}

	s, i, err := dec.DecodeString(raw, 0, nil)
	if err != nil {
		return "", err
	}
	if SkipSpaces(raw, i) != len(raw) {
		return "", ErrValue
	}

	return string(s), nil
}

func (raw Value) ShouldInt() (v int) {
	v, _ = raw.Int()
	return v
}

func (raw Value) Int() (v int, err error) {
	x, err := raw.Int64()
	return int(x), err
}

func (raw Value) ShouldInt64() (v int64) {
	v, _ = raw.Int64()
	return v
}

func (raw Value) Int64() (v int64, err error) {
	v, err = strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func (raw Value) ShouldUint64() (v uint64) {
	v, _ = raw.Uint64()
	return v
}

func (raw Value) Uint64() (v uint64, err error) {
	v, err = strconv.ParseUint(string(raw), 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func (raw Value) ShouldFloat64() (v float64) {
	v, _ = raw.Float64()
	return v
}

func (raw Value) Float64() (v float64, err error) {
	v, err = strconv.ParseFloat(string(raw), 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func (raw Value) ShouldFloat32() (v float32) {
	v, _ = raw.Float32()
	return v
}

func (raw Value) Float32() (v float32, err error) {
	x, err := strconv.ParseFloat(string(raw), 32)
	if err != nil {
		return 0, err
	}

	return float32(x), nil
}
