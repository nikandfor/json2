package json2

import "errors"

// Seek errors.
var (
	ErrNoSuchKey   = errors.New("no such object key")
	ErrOutOfBounds = errors.New("out of array bounds")
)

// Seek seeks to the beginning of the value at the path – list of object keys and array indexes.
// If you parse multiple object and you only need one value from each,
// it's good to use Break(len(path)) to move to the beginning of the next object.
func (d *Iterator) Seek(b []byte, st int, path ...any) (i int, err error) {
	i = st

	for _, p := range path {
		switch p := p.(type) {
		case string:
			i, err = d.seekObj(b, i, p)
		case int:
			i, err = d.seekArr(b, i, p)
		}

		if err != nil {
			return i, err
		}
	}

	return i, nil
}

func (d *Iterator) seekObj(b []byte, st int, key string) (i int, err error) {
	i, err = d.Enter(b, st, Object)
	if err != nil {
		return
	}

	var k []byte

	for err == nil && d.ForMore(b, &i, Object, &err) {
		k, i, err = d.Key(b, i)
		if err != nil {
			return
		}

		if string(k) == key {
			_, i, err = d.Type(b, i)
			return
		}

		i, err = d.Skip(b, i)
	}
	if err != nil {
		return
	}

	return i, ErrNoSuchKey
}

func (d *Iterator) seekArr(b []byte, st, idx int) (i int, err error) {
	if idx < 0 {
		l, i, err := d.Length(b, st)
		if err != nil {
			return i, err
		}

		idx = l + idx
	}
	if idx < 0 {
		return st, ErrOutOfBounds
	}

	i, err = d.Enter(b, st, Array)
	if err != nil {
		return
	}

	j := 0

	for err == nil && d.ForMore(b, &i, Array, &err) {
		if j == idx {
			_, i, err = d.Type(b, i)
			return
		}

		i, err = d.Skip(b, i)
		j++
	}
	if err != nil {
		return
	}

	return i, ErrOutOfBounds
}
