package json2

// Seek seeks to the beginning of the value at the path – list of object keys and array indexes.
// If you parse multiple object and you only need one value from each,
// it's good to use Break(len(path)) to move to the beginning of the next object.
func (r *Reader) Seek(path ...any) (err error) {
	for _, p := range path {
		switch p := p.(type) {
		case string:
			err = r.seekObj(p)
		case int:
			err = r.seekArr(p)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Reader) seekObj(key string) (err error) {
	err = r.Enter(Object)
	if err != nil {
		return
	}

	var k []byte

	for err == nil && r.ForMore(Object, &err) {
		k, err = r.Key()
		if err != nil {
			return
		}

		if string(k) == key {
			_, err = r.Type()
			return
		}

		err = r.Skip()
	}
	if err != nil {
		return
	}

	return ErrNoSuchKey
}

func (r *Reader) seekArr(idx int) (err error) {
	if idx < 0 {
		r.Lock()

		l, err := r.Length()
		if err != nil {
			return err
		}

		r.Rewind()
		r.Unlock()

		idx = l + idx
	}
	if idx < 0 {
		return ErrOutOfBounds
	}

	err = r.Enter(Array)
	if err != nil {
		return
	}

	j := 0

	for err == nil && r.ForMore(Array, &err) {
		if j == idx {
			_, err = r.Type()
			return
		}

		err = r.Skip()
		j++
	}
	if err != nil {
		return
	}

	return ErrOutOfBounds
}
