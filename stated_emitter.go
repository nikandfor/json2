package json2

import (
	"strconv"
	"unsafe"
)

type (
	StatedEmitter struct {
		Buf []byte

		Emitter

		state encState
		stack encStack
		depth int
	}

	encState int
	encStack int64
)

const (
	encComma = 1 << iota
	encColon
	encNewline
	encKey
	_
)

func NewStatedEmitter(b []byte) *StatedEmitter {
	return &StatedEmitter{Buf: b}
}

func (e *StatedEmitter) Reset() *StatedEmitter {
	e.Buf = e.Buf[:0]
	e.state = 0
	e.stack = 0
	e.depth = 0

	return e
}

func (e *StatedEmitter) Result() []byte { return e.Buf }

func (e *StatedEmitter) KeyString(k, v string) *StatedEmitter {
	return e.Key(k).String(v)
}

func (e *StatedEmitter) KeyStringBytes(k string, v []byte) *StatedEmitter {
	return e.Key(k).StringBytes(v)
}

func (e *StatedEmitter) KeyInt(k string, v int) *StatedEmitter {
	return e.Key(k).Int(v)
}

func (e *StatedEmitter) KeyInt64(k string, v int64) *StatedEmitter {
	return e.Key(k).Int64(v)
}

func (e *StatedEmitter) ObjStart() *StatedEmitter {
	if e.depth > 62 {
		panic("too deep")
	}

	e.comma()

	e.Buf = append(e.Buf, '{')
	e.state.Set(encKey)

	e.stack.SetBit(e.depth)
	e.depth++

	e.state.Set(encNewline)

	return e
}

func (e *StatedEmitter) ObjEnd() *StatedEmitter {
	e.depth--
	e.stack.UnsetBit(e.depth)

	e.Buf = append(e.Buf, '}')
	e.setcomma()

	return e
}

func (e *StatedEmitter) ArrStart() *StatedEmitter {
	e.comma()

	e.Buf = append(e.Buf, '[')
	e.depth++

	e.state.Set(encNewline)

	return e
}

func (e *StatedEmitter) ArrEnd() *StatedEmitter {
	e.depth--

	e.Buf = append(e.Buf, ']')
	e.setcomma()

	return e
}

func (e *StatedEmitter) Key(s string) *StatedEmitter {
	e.comma()

	e.Buf = e.Emitter.AppendKey(e.Buf, s2b(s))
	e.colon(true)

	return e
}

func (e *StatedEmitter) NextIsKey() *StatedEmitter {
	e.state.Set(encKey)

	return e
}

func (e *StatedEmitter) Raw(raw []byte) *StatedEmitter {
	e.comma()

	e.Buf = append(e.Buf, raw...)
	e.setcomma()

	e.colon(false)

	return e
}

func (e *StatedEmitter) String(s string) *StatedEmitter {
	e.comma()

	e.Buf = e.Emitter.AppendString(e.Buf, s2b(s))
	e.setcomma()

	e.colon(false)

	return e
}

func (e *StatedEmitter) StringBytes(s []byte) *StatedEmitter {
	e.comma()

	e.Buf = e.Emitter.AppendString(e.Buf, s)
	e.setcomma()

	e.colon(false)

	return e
}

func (e *StatedEmitter) Int(v int) *StatedEmitter {
	e.comma()

	e.Buf = strconv.AppendInt(e.Buf, int64(v), 10)
	e.setcomma()

	return e
}

func (e *StatedEmitter) Int64(v int64) *StatedEmitter {
	e.comma()

	e.Buf = strconv.AppendInt(e.Buf, v, 10)
	e.setcomma()

	return e
}

func (e *StatedEmitter) Newline() *StatedEmitter {
	e.newline()

	return e
}

func (e *StatedEmitter) setcomma() {
	e.state.Unset(encComma)

	if e.depth != 0 {
		e.state.Set(encComma)
	}
}

func (e *StatedEmitter) comma() {
	if e.depth == 0 {
		e.newline()
		e.state.Unset(encNewline)

		return
	}

	if e.state.Is(encComma) {
		e.Buf = append(e.Buf, ',')
	}

	e.state.Unset(encComma)
	e.state.Set(encNewline)
}

func (e *StatedEmitter) colon(force bool) {
	if force || e.state.Is(encKey) {
		e.Buf = append(e.Buf, ':')
	}

	e.state.Unset(encKey)
}

func (e *StatedEmitter) newline() {
	if e.state.Is(encNewline) {
		e.Buf = append(e.Buf, '\n')
	}

	e.state.Unset(encNewline)
}

func (s encState) Is(flag encState) bool {
	return s&flag == flag
}

func (s encState) Any(flag encState) bool {
	return s&flag != 0
}

func (s *encState) Set(flag encState) {
	*s |= flag
}

func (s *encState) Unset(flag encState) {
	*s &^= flag
}

func (s *encStack) Bit(n int) bool {
	return *s&(1<<n) != 0
}

func (s *encStack) SetBit(n int) {
	*s |= 1 << n
}

func (s *encStack) UnsetBit(n int) {
	*s &^= 1 << n
}

func s2b(s string) []byte {
	data := unsafe.StringData(s)

	return unsafe.Slice(data, len(s))
}
