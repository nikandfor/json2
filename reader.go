package json2

import (
	"errors"
	"io"

	"nikand.dev/go/skip"
)

type (
	Reader struct {
		r io.Reader

		lock []int

		b []byte
		i int

		off int64
	}
)

// NewReader creates a new Reader.
// It first reads from b and then from rd.
// If you just want to provide a buffer slice it to zero length.
func NewReader(b []byte, rd io.Reader) *Reader {
	return &Reader{
		r: rd,
		b: b,
	}
}

// Reset resets reader.
// As in NewReader it makes reading first from b and then from rd.
func (r *Reader) Reset(b []byte, rd io.Reader) {
	r.r = rd
	r.b = b
	r.lock = r.lock[:0]
	r.i = 0
	r.off = 0
}

// Offset returns current position in the stream.
func (r *Reader) Offset() int64 {
	return r.off + int64(r.i)
}

// Type finds the beginning of the next value and detects its type.
// It doesn't parse the value so it can't detect if it's incorrect.
func (r *Reader) Type() (tp Type, err error) {
again:
	for r.i < len(r.b) {
		if isWhitespace(r.b[r.i]) {
			r.i++
			continue
		}

		switch r.b[r.i] {
		case ',', ':':
			r.i++
			continue
		case '/':
			err = r.skipComment()
			if err != nil {
				return None, err
			}
			continue
		case 't', 'f':
			return Bool, nil
		case '"':
			return String, nil
		case byte(Null), byte(Array), byte(Object):
			return Type(r.b[r.i]), nil
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'+', '-', '.',
			'N',      // NaN
			'i', 'I': // Inf
			return Number, nil
		}

		return None, ErrSyntax
	}

	err = r.more()
	if err != nil {
		return None, err
	}

	goto again
}

// Skip skips the next value.
func (r *Reader) Skip() error {
	return r.Break(0)
}

// Raw skips the next value and returns subslice with the value trimming whitespaces.
//
// Returned buffer is only valid until the next reading method is called.
// It can be reused if more data needed to be read from underlying reader.
func (r *Reader) Raw() ([]byte, error) {
	_, err := r.Type()
	if err != nil {
		return nil, err
	}

	l := r.Lock()
	defer r.Unlock()

	err = r.Break(0)
	if err != nil {
		return nil, nil
	}

	st := r.lock[l-1]

	return r.b[st:r.i], nil
}

// Break breaks from inside the object to the end of it on depth levels.
// As a special case with depth=0 it skips the next value.
// Skip and Raw do exactly that.
//
// It's intended for exiting out of arrays and objects when their content is not needed anymore
// (all the needed indexes or keys are already parsed) and we want to parse the next array or object.
func (r *Reader) Break(depth int) (err error) {
	var d Iterator

again:
	for err == nil && r.i < len(r.b) {
		if isWhitespace(r.b[r.i]) {
			r.i++
			continue
		}

		switch r.b[r.i] {
		case ',', ':':
			r.i++
			continue
		case '/':
			err = r.skipComment()
			if err != nil {
				return err
			}
			continue
		case '"':
			_, _, err = r.skipString()
		case 'n', 't', 'f':
			r.i, err = d.skipLit(r.b, r.i)
			if err == ErrShortBuffer { //nolint:errorlint
				err = nil
				break again
			}
		case '[', '{':
			r.i++
			depth++
		case ']', '}':
			r.i++
			depth--
		default:
			r.i, err = d.skipNum(r.b, r.i)
			if err == ErrBadNumber && r.i == len(r.b) { //nolint:errorlint
				err = nil
				break again
			}
		}

		if depth == 0 {
			return nil
		}
	}
	if err != nil {
		return err
	}

	err = r.more()

	goto again
}

// Key reads the next string removing quotes but not decoding the string value.
// So escape sequences (\n, \uXXXX) are not decoded. They are returned as is.
// This is intended for object keys as they usually contain alpha-numeric symbols only.
// This is faster and does not require additional buffer for decoding.
//
// Returned buffer is only valid until the next reading method is called.
// It can be reused if more data needed to be read from underlying reader.
func (r *Reader) Key() ([]byte, error) {
	tp, err := r.Type()
	if err != nil {
		return nil, err
	}
	if tp != String {
		return nil, ErrType
	}

	l := r.Lock()
	defer r.Unlock()

	if _, _, err := r.skipString(); err != nil {
		return nil, err
	}

	st := r.lock[l-1]

	return r.b[st+1 : r.i-1], nil
}

// DecodeString reads the next string, decodes escape sequences (\n, \uXXXX),
// and appends the result to the buf.
//
// Data is appended to the provided buffer. And the buffer will not be preserved by Reader.
func (r *Reader) DecodeString(buf []byte) (s []byte, err error) {
	tp, err := r.Type()
	if err != nil {
		return buf, err
	}
	if tp != String {
		return buf, ErrType
	}

	return r.decodeString(buf)
}

// DecodedStringLength reads and decodes the next string but only return the result length.
// It doesn't allocate while DecodeString does.
func (r *Reader) DecodedStringLength() (bs, rs int, err error) {
	tp, err := r.Type()
	if err != nil {
		return 0, 0, err
	}
	if tp != String {
		return 0, 0, ErrType
	}

	return r.skipString()
}

// Enter enters an Array or an Object. typ is checked to match with the actual container type.
// Use More or, more convenient form, ForMore to iterate over container.
// See examples to better understand usage pattern.
func (r *Reader) Enter(typ Type) (err error) {
	tp, err := r.Type()
	if err != nil {
		return
	}

	if tp != typ || typ != Array && typ != Object {
		return ErrType
	}

	r.i++

	return
}

// More iterates over an Array or an Object elements entered by the Enter method.
func (r *Reader) More(typ Type) (more bool, err error) {
again:
	for r.i < len(r.b) {
		if isWhitespace(r.b[r.i]) || r.b[r.i] == ',' {
			r.i++
			continue
		}

		break
	}

	if r.i == len(r.b) {
		if err := r.more(); err != nil {
			return false, err
		}

		goto again
	}

	if r.b[r.i] == byte(typ)+2 {
		r.i++
		return false, nil
	}

	tp, err := r.Type()
	if err != nil {
		return false, err
	}

	if typ == Object && tp != String {
		return false, ErrSyntax
	}

	return true, nil
}

// ForMore is a convenient wrapper for More which makes iterating code shorter and simpler.
func (r *Reader) ForMore(typ Type, errp *error) bool {
	more, err := r.More(typ)
	if err != nil {
		*errp = err
	}

	return more
}

// Length calculates number of elements in Array or Object.
func (r *Reader) Length() (n int, err error) {
	tp, err := r.Type()
	if err != nil {
		return 0, err
	}

	switch tp {
	case Array, Object:
	default:
		return 0, ErrType
	}

	err = r.Enter(tp)
	if err != nil {
		return 0, err
	}

	for r.ForMore(tp, &err) {
		if tp == Object {
			_, err = r.Key()
			if err != nil {
				return n, err
			}
		}

		err = r.Skip()
		if err != nil {
			return n, err
		}

		n++
	}
	if err != nil {
		return n, err
	}

	return n, nil
}

// Lock locks the buffer from rewriting by reading more data from Reader.
// Lock also remembers position in the stream and allows rewinding to it.

// Lock locks internal buffer so the data is not overwritten when more data is read from underlaying reader.
// It's used to return to the locked position in a stream and reread some part of it.
// Internal buffer grows to the size of data locked plus additional space for the next Read.
// Lock must be followed by Unlock just like for sync.Mutex.
// Rewind is used to return to the latest Lock position.
// Multiple nested locks are allowed.
// It returns the number of locks acquired and not released so far; kinda Lock depth.
func (r *Reader) Lock() int {
	r.lock = append(r.lock, r.i)

	return len(r.lock)
}

// Unlock releases the latest buffer Lock.
// It returns the number of remaining active Locks.
func (r *Reader) Unlock() int {
	r.lock = r.lock[:len(r.lock)-1]

	return len(r.lock)
}

// Rewind returns stream position to the latest Lock.
func (r *Reader) Rewind() {
	r.i = r.lock[len(r.lock)-1]
}

func (r *Reader) skipString() (bs, rs int, err error) {
	var bs0, rs0 int
	s := skip.Quo

	for {
		if r.i >= len(r.b) {
			if err := r.more(); err != nil {
				return bs, rs, err
			}
		}

		s, bs0, rs0, r.i = skip.String(r.b, r.i, s)
		bs += bs0
		rs += rs0
		if !s.Err() {
			return bs, rs, nil
		}
		if s.Err() && !s.Is(skip.ErrBuffer) {
			return bs, rs, s
		}
	}
}

func (r *Reader) decodeString(w []byte) (_ []byte, err error) {
	s := skip.Quo

	for {
		if r.i >= len(r.b) {
			if err := r.more(); err != nil {
				return w, err
			}
		}

		s, w, _, r.i = skip.DecodeString(r.b, r.i, s, w)
		if !s.Err() {
			return w, nil
		}
		if s.Err() && !s.Is(skip.ErrBuffer) {
			return w, s
		}
	}
}

func (r *Reader) skipComment() (err error) { //nolint:gocognit
	state := byte(0)

	for {
		if r.i >= len(r.b) {
			if err := r.more(); err != nil {
				return err
			}
		}

		if state == 0 {
			if r.i+1 >= len(r.b) {
				continue
			}

			if r.b[r.i] != '/' {
				return ErrSyntax
			}

			state = r.b[r.i+1]

			if state != '/' && state != '*' {
				return ErrSyntax
			}

			r.i += 2

			continue
		}

		if state == '/' {
			for r.i < len(r.b) && r.b[r.i] != '\n' {
				r.i++
			}
			if r.i == len(r.b) {
				continue
			}

			r.i++

			return
		}

		// state == '*'

		for {
			for r.i < len(r.b) && r.b[r.i] != '*' {
				r.i++
			}
			if r.i+1 >= len(r.b) {
				break
			}
			r.i++

			if r.b[r.i] == '/' {
				r.i++

				return
			}
		}
	}
}

func (r *Reader) more() error {
	if r.r == nil && r.i == len(r.b) {
		return io.EOF
	}

	if r.r == nil {
		return io.ErrUnexpectedEOF
	}

	end := len(r.b)

	st := r.i
	if len(r.lock) > 0 {
		st = r.lock[0]
	}

	if st != 0 {
		r.off += int64(st)

		copy(r.b, r.b[st:end])
		end -= st
		r.i -= st

		for i := range r.lock {
			r.lock[i] -= st
		}
	}

	if cap(r.b) == 0 {
		r.b = make([]byte, 16<<10)
	}

	r.b = r.b[:cap(r.b)]

	n, err := r.r.Read(r.b[end:])
	end += n

	r.b = r.b[:end]

	if n != 0 && errors.Is(err, io.EOF) {
		err = nil
	}

	return err
}
