package json2

import (
	"unicode/utf8"
)

type (
	Emitter struct{}
)

const hex = "0123456789abcdef"

var jsonStringSafe [2]uint64

var enc Emitter

func init() {
	for b := byte(0x20); b < utf8.RuneSelf; b++ {
		if b == '"' || b == '\\' {
			continue
		}

		i, j := safeIJ(b)

		jsonStringSafe[i] |= 1 << j
	}
}

func (e *Emitter) AppendKey(w, s []byte) []byte {
	w = append(w, '"')
	w = append(w, s...)
	w = append(w, '"')

	return w
}

// EncodeString encodes string in a JSON compatible way.
func (e *Emitter) AppendString(w, s []byte) []byte {
	w = append(w, '"')
	w = e.AppendStringContent(w, s)
	w = append(w, '"')

	return w
}

// EncodeStringContent does the same as EncodeString but does not add quotes.
// It can be used to generate the string from multiple parts.
// Yet if a symbol designated to be escaped is split between parts
// it encodes each part of the symbol separately.
func (e *Emitter) AppendStringContent(w, s []byte) []byte {
	done := 0

	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if isSafeJSON(b) {
				i++
				continue
			}

			w = append(w, s[done:i]...)

			switch b {
			case '"', '\\', '/':
				w = append(w, '\\', b)
			case '\n':
				w = append(w, '\\', 'n')
			case '\r':
				w = append(w, '\\', 'r')
			case '\t':
				w = append(w, '\\', 't')
				//	case '\f':
				//		w = append(w, '\\', 'f')
				//	case '\b':
				//		w = append(w, '\\', 'b')
			default:
				w = append(w, '\\', 'u', '0', '0', hex[b>>4&0xf], hex[b&0xf])
			}

			i++
			done = i

			continue
		}

		r, size := utf8.DecodeRune(s[i:])

		if r == utf8.RuneError && size == 1 || r == '\u2028' || r == '\u2029' {
			w = append(w, s[done:i]...)
			w = append(w, '\\', 'u', hex[r>>12&0xf], hex[r>>8&0xf], hex[r>>4&0xf], hex[r&0xf])

			//	w = append(w, `\ufffd`...)
			//	w = append(w, '\\', 'u', '2', '0', 2, hex[r&0xf])

			i += size
			done = i

			continue
		}

		i += size
	}

	w = append(w, s[done:]...)

	return w
}

func safeIJ(b byte) (i, j int) {
	return int(b) / 64, int(b) % 64
}

func isSafeJSON(b byte) bool {
	i, j := safeIJ(b)

	return jsonStringSafe[i]&(1<<j) != 0
}
