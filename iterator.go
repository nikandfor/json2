package json2

import (
	"errors"
	"io"

	"nikand.dev/go/skip"
)

type (
	// Iterator is a group of methods to parse JSON.
	// Iterator is stateless.
	// All the needed state is passed though arguments and return values.
	//
	// Most of the methods take buffer with json and start position
	// and return a value, end position and possible error.
	Iterator struct{}

	Type byte
)

// Value types returned by Iterator.
const (
	None    Type = 0 // never returned in successful case
	Null    Type = 'n'
	Bool    Type = 'b'
	String  Type = 's'
	Array   Type = '['
	Object  Type = '{'
	Number  Type = '1'
	Comment Type = '/'
)

var whitespaces uint64 = 1<<'\n' | 1<<'\r' | 1<<'\t' | 1<<' '

// Iterator errors. Plus Str errors from skip module.
var (
	ErrBadNumber   = errors.New("bad number")
	ErrShortBuffer = io.ErrShortBuffer
	ErrSyntax      = errors.New("syntax error")
	ErrType        = errors.New("incompatible type")
)

func (d *Iterator) ExpectType(b []byte, st int, typ Type) (i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return i, err
	}

	if tp != typ {
		return i, ErrType
	}

	return i, nil
}

// Type finds the beginning of the next value and detects its type.
// It doesn't parse the value so it can't detect if it's incorrect.
func (d *Iterator) Type(b []byte, st int) (tp Type, i int, err error) {
	for i = st; i < len(b); i++ {
		if isWhitespace(b[i]) {
			continue
		}

		switch b[i] {
		case ',', ':':
			continue
		case '/':
			i, err = d.skipComment(b, i)
			if err != nil {
				return None, i, err
			}
			continue
		case 't', 'f':
			return Bool, i, nil
		case '"':
			return String, i, nil
		case byte(Null), byte(Array), byte(Object):
			return Type(b[i]), i, nil
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'+', '-', '.',
			'N',      // NaN
			'i', 'I': // Inf
			return Number, i, nil
		}

		return None, i, ErrSyntax
	}

	return None, i, ErrShortBuffer
}

// Skip skips the next value.
func (d *Iterator) Skip(b []byte, st int) (i int, err error) {
	return d.Break(b, st, 0)
}

// Raw skips the next value and returns subslice with the value trimming whitespaces.
func (d *Iterator) Raw(b []byte, st int) (v []byte, i int, err error) {
	_, st, err = d.Type(b, st)
	if err != nil {
		return nil, st, err
	}

	i, err = d.Break(b, st, 0)
	if err != nil {
		return
	}

	return b[st:i], i, nil
}

// Break breaks from inside the object to the end of it on depth levels.
// As a special case with depth=0 it skips the next value.
// Skip and Raw do exactly that.
//
// It's intended for exiting out of arrays and objects when their content is not needed anymore
// (all the needed indexes or keys are already parsed) and we want to parse the next array or object.
func (d *Iterator) Break(b []byte, st, depth int) (i int, err error) {
	i = st

	for i < len(b) {
		if isWhitespace(b[i]) {
			i++
			continue
		}

		switch b[i] {
		case ',', ':':
			i++
			continue
		case '/':
			i, err = d.skipComment(b, i)
			if err != nil {
				return i, err
			}
			continue
		case '"':
			i, err = d.skipString(b, i)
		case 'n', 't', 'f':
			i, err = d.skipLit(b, i)
		case '[', '{':
			i++
			depth++
		case ']', '}':
			i++
			depth--
		default:
			i, err = d.skipNum(b, i)
		}

		if err != nil {
			return
		}

		if depth == 0 {
			return
		}
	}

	return st, ErrShortBuffer
}

// Key reads the next string removing quotes but not decoding the string value.
// Escape sequences (\n, \uXXXX) are not decoded. They are returned as is.
// This is intended for object keys as they usually contain alpha-numeric symbols only.
// This is faster and does not require additional buffer for decoding.
// It returns subslice of b, so k content must not be edited.
func (d *Iterator) Key(b []byte, st int) (k []byte, i int, err error) {
	i, err = d.ExpectType(b, st, String)
	if err != nil {
		return nil, i, err
	}

	raw, i, err := d.Raw(b, i)
	if err != nil {
		return nil, i, err
	}

	return raw[1 : len(raw)-1], i, nil
}

// DecodeString reads the next string, decodes escape sequences (\n, \uXXXX),
// and appends the result to the buf.
func (d *Iterator) DecodeString(b []byte, st int, buf []byte) (s []byte, i int, err error) {
	i, err = d.ExpectType(b, st, String)
	if err != nil {
		return buf, i, err
	}

	ss, w, _, i := skip.DecodeString(b, i, skip.Quo|skip.ErrRune|skip.ErrEscape, buf)
	if ss.Is(skip.ErrBuffer) {
		return w, st, ErrShortBuffer
	}
	if ss.Err() {
		return w, i, ss
	}

	return w, i, nil
}

// DecodedStringLength reads and decodes the next string but only return the result length.
// It doesn't allocate while DecodeString does.
func (d *Iterator) DecodedStringLength(b []byte, st int) (bs, rs, i int, err error) {
	i, err = d.ExpectType(b, st, String)
	if err != nil {
		return 0, 0, i, err
	}

	ss, bs, rs, i := skip.String(b, i, skip.Quo|skip.ErrRune)
	if ss.Is(skip.ErrBuffer) {
		return bs, rs, st, ErrShortBuffer
	}
	if ss.Err() {
		return bs, rs, i, ss
	}

	return bs, rs, i, nil
}

// Enter enters an Array or an Object. typ is checked to match with the actual container type.
// Use More or, more convenient form, ForMore to iterate over container.
// See examples to better understand usage pattern.
func (d *Iterator) Enter(b []byte, st int, typ Type) (i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return
	}

	if tp != typ || typ != Array && typ != Object {
		return i, ErrType
	}

	i++

	return
}

// More iterates over an Array or an Object elements entered by the Enter method.
func (d *Iterator) More(b []byte, st int, typ Type) (more bool, i int, err error) {
	for i = st; i < len(b); i++ {
		if isWhitespace(b[i]) || b[i] == ',' {
			continue
		}

		break
	}

	if i == len(b) {
		return false, i, ErrShortBuffer
	}

	if b[i] == byte(typ)+2 {
		i++
		return false, i, nil
	}

	tp, i, err := d.Type(b, i)
	if err != nil {
		return false, i, err
	}

	if typ == Object && tp != String {
		return false, i, ErrSyntax
	}

	return true, i, nil
}

// ForMore is a convenient wrapper for More which makes iterating code shorter and simpler.
func (d *Iterator) ForMore(b []byte, i *int, typ Type, errp *error) bool { //nolint:gocritic
	more, j, err := d.More(b, *i, typ)
	*i = j

	if errp != nil {
		*errp = err
	}

	return more
}

// Length calculates number of elements in Array or Object.
func (d *Iterator) Length(b []byte, st int) (n, i int, err error) {
	tp, i, err := d.Type(b, st)
	if err != nil {
		return 0, i, err
	}

	switch tp {
	case Array, Object:
	default:
		return 0, i, ErrType
	}

	i, err = d.Enter(b, i, tp)
	if err != nil {
		return 0, i, err
	}

	for d.ForMore(b, &i, tp, &err) {
		if tp == Object {
			_, i, err = d.Key(b, i)
			if err != nil {
				return n, i, err
			}
		}

		i, err = d.Skip(b, i)
		if err != nil {
			return n, i, err
		}

		n++
	}
	if err != nil {
		return n, i, err
	}

	return n, i, nil
}

// SkipSpaces skips whitespaces.
func (d *Iterator) SkipSpaces(b []byte, i int) int {
	for i < len(b) && isWhitespace(b[i]) {
		i++
	}

	return i
}

func (d *Iterator) skipString(b []byte, st int) (i int, err error) {
	ss, _, _, i := skip.String(b, st, skip.Quo)
	if ss.Is(skip.ErrBuffer) {
		return st, ErrShortBuffer
	}
	if ss.Err() {
		return i, ss
	}

	return i, nil
}

func (d *Iterator) skipNum(b []byte, st int) (i int, err error) {
	n, i := skip.Number(b, st, 0)
	if !n.Ok() {
		return i, ErrBadNumber
	}

	return i, nil
}

func (d *Iterator) skipLit(b []byte, st int) (i int, err error) {
	var lit string

	switch b[st] {
	case 't':
		lit = "true"
	case 'f':
		lit = "false"
	case 'n':
		lit = "null"
	}

	return d.skipVal(b, st, lit)
}

func (d *Iterator) skipVal(b []byte, st int, val string) (i int, err error) {
	end := st + len(val)

	if end <= len(b) && string(b[st:end]) == val {
		return end, nil
	}

	if end > len(b) && string(b[st:]) == val[:len(b)-st] {
		return st, ErrShortBuffer
	}

	return st, ErrSyntax
}

func (d *Iterator) skipComment(b []byte, st int) (i int, err error) {
	i = st

	if i+1 >= len(b) {
		return st, ErrShortBuffer
	}

	if b[i] != '/' {
		return st, ErrSyntax
	}
	i++

	switch b[i] {
	case '/':
		for i < len(b) && b[i] != '\n' {
			i++
		}

		return i, nil
	case '*':
		//
	default:
		return st, ErrSyntax
	}

	for {
		for i < len(b) && b[i] != '*' {
			i++
		}
		if i+1 >= len(b) {
			return st, ErrShortBuffer
		}
		i++

		if b[i] == '/' {
			i++
			break
		}
	}

	return i, nil
}

// SkipSpaces skips whitespaces.
func SkipSpaces(b []byte, i int) int {
	return skip.Spaces(b, i)
}

func isWhitespace(b byte) bool {
	return b <= 0x20 && whitespaces&(1<<b) != 0
}
