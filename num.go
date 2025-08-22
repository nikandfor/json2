package json2

import "strconv"

func (d *Iterator) Num(b []byte, st int) (raw []byte, i int, err error) {
	i, err = d.ExpectType(b, st, Number)
	if err != nil {
		return nil, st, err
	}

	return d.Raw(b, i)
}

func (d *Iterator) Int(b []byte, st int) (v, i int, err error) {
	x, i, err := d.Int64(b, st)
	if err != nil {
		return 0, i, err
	}

	return int(x), i, nil
}

func (d *Iterator) Int64(b []byte, st int) (v int64, i int, err error) {
	raw, i, err := d.Num(b, st)
	if err != nil {
		return 0, i, err
	}

	v, err = strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return 0, i, err
	}

	return v, i, nil
}

func (d *Iterator) Uint64(b []byte, st int) (v uint64, i int, err error) {
	raw, i, err := d.Num(b, st)
	if err != nil {
		return 0, i, err
	}

	v, err = strconv.ParseUint(string(raw), 10, 64)
	if err != nil {
		return 0, i, err
	}

	return v, i, nil
}

func (d *Iterator) Float64(b []byte, st int) (v float64, i int, err error) {
	raw, i, err := d.Num(b, st)
	if err != nil {
		return 0, i, err
	}

	v, err = strconv.ParseFloat(string(raw), 64)
	if err != nil {
		return 0, i, err
	}

	return v, i, nil
}

func (d *Iterator) Float32(b []byte, st int) (v float32, i int, err error) {
	f, i, err := d.Float64(b, st)

	return float32(f), i, err
}
