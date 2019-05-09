package json

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

func Marshal(r interface{}) ([]byte, error) {
	w := NewWriter(nil)
	err := w.Marshal(r)
	return w.Bytes(), err
}

func (w *Writer) Marshal(r interface{}) error {
	if r == nil {
		w.Null()
		return w.Err()
	}
	return w.marshal(reflect.ValueOf(r))
}

func (w *Writer) marshal(rv reflect.Value) error {
	//	log.Printf("marshal: %v %v", rv.Type(), rv)
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			w.Null()
			return w.Err()
		}
		rv = rv.Elem()
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			w.Null()
			return w.Err()
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return w.marshalStruct(rv)
	case reflect.Array, reflect.Slice:
		return w.marshalSlice(rv)
	case reflect.String:
		w.SafeString(UnsafeStringToBytes(rv.String()))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i := rv.Int()
		s := strconv.FormatInt(i, 10)
		w.Number(UnsafeStringToBytes(s))
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
		reflect.Uintptr:
		i := rv.Uint()
		s := strconv.FormatUint(i, 10)
		w.Number(UnsafeStringToBytes(s))
	case reflect.Float64, reflect.Float32:
		bits := 64
		if rv.Kind() == reflect.Float32 {
			bits = 32
		}
		f := rv.Float()
		s := strconv.FormatFloat(f, 'g', -1, bits)
		w.Number(UnsafeStringToBytes(s))
	case reflect.Bool:
		w.Bool(rv.Bool())
	case reflect.Map:
		return w.marshalMap(rv)
	default:
		panic(rv.Kind())
	}
	return w.Err()
}

func (w *Writer) marshalMap(rv reflect.Value) error {
	w.ObjStart()
	keys := rv.MapKeys()
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		for k.Kind() == reflect.Interface {
			k = k.Elem()
		}
		switch k.Kind() {
		case reflect.String:
			w.ObjKey([]byte(k.String()))
		default:
			panic(k)
		}
		v := rv.MapIndex(k)
		if v.Kind() == reflect.Interface {
			if v.IsNil() {
				w.Null()
				continue
			}
			v = v.Elem()
		}
		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				w.Null()
				continue
			}
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.String:
			w.SafeStringString(v.String())
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
			reflect.Float64, reflect.Float32,
			reflect.Uintptr:
			s := fmt.Sprintf("%v", v.Interface())
			w.Number(UnsafeStringToBytes(s))
		case reflect.Bool:
			w.Bool(v.Bool())
		default:
			if err := w.marshal(v); err != nil {
				return err
			}
		}
	}
	w.ObjEnd()

	return w.Err()
}

func (w *Writer) marshalStruct(rv reflect.Value) error {
	m := getStructMap(rv.Type())
	var ptr, fptr uintptr
	if rv.CanAddr() {
		ptr = rv.UnsafeAddr()
	}
	//	log.Printf("struct: %+v", m)

	w.ObjStart()
	for i, f := range m.s {
		if ptr != 0 {
			fptr = ptr + f.Ptr
		}

		switch f.Kind {
		case reflect.String:
			w.ObjKey(f.Name)
			val := rv.Field(i).String()
			if val == "" && f.OmitEmpty {
				continue
			}
			w.SafeStringString(val)
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			var q int64
			if fptr == 0 {
				q = rv.Field(i).Int()
			} else {
				switch f.Kind {
				case reflect.Int:
					q = (int64)(*(*int)(unsafe.Pointer(fptr)))
				case reflect.Int64:
					q = (int64)(*(*int64)(unsafe.Pointer(fptr)))
				case reflect.Int32:
					q = (int64)(*(*int32)(unsafe.Pointer(fptr)))
				case reflect.Int16:
					q = (int64)(*(*int16)(unsafe.Pointer(fptr)))
				case reflect.Int8:
					q = (int64)(*(*int8)(unsafe.Pointer(fptr)))
				}
			}
			if q == 0 && f.OmitEmpty {
				continue
			}
			s := strconv.FormatInt(q, 10)
			w.ObjKey(f.Name)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
			reflect.Uintptr:
			var q uint64
			if fptr == 0 {
				q = rv.Field(i).Uint()
			} else {
				switch f.Kind {
				case reflect.Uint:
					q = (uint64)(*(*uint)(unsafe.Pointer(fptr)))
				case reflect.Uintptr:
					q = (uint64)(*(*uintptr)(unsafe.Pointer(fptr)))
				case reflect.Uint64:
					q = (uint64)(*(*uint64)(unsafe.Pointer(fptr)))
				case reflect.Uint32:
					q = (uint64)(*(*uint32)(unsafe.Pointer(fptr)))
				case reflect.Uint16:
					q = (uint64)(*(*uint16)(unsafe.Pointer(fptr)))
				case reflect.Uint8:
					q = (uint64)(*(*uint8)(unsafe.Pointer(fptr)))
				}
			}
			if q == 0 && f.OmitEmpty {
				continue
			}
			s := strconv.FormatUint(q, 10)
			w.ObjKey(f.Name)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Float64, reflect.Float32:
			bits := 64
			if f.Kind == reflect.Float32 {
				bits = 32
			}
			var q float64
			if fptr == 0 {
				q = rv.Field(i).Float()
			} else {
				if f.Kind == reflect.Float64 {
					q = (float64)(*(*float64)(unsafe.Pointer(fptr)))
				} else {
					q = (float64)(*(*float32)(unsafe.Pointer(fptr)))
				}
			}
			if q == 0 && f.OmitEmpty {
				continue
			}
			s := strconv.FormatFloat(q, 'g', -1, bits)
			w.ObjKey(f.Name)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Bool:
			var q bool
			if fptr == 0 {
				q = rv.Field(i).Bool()
			} else {
				q = *(*bool)(unsafe.Pointer(fptr))
			}
			if !q && f.OmitEmpty {
				continue
			}
			w.ObjKey(f.Name)
			w.Bool(q)
		case reflect.Slice, reflect.Ptr, reflect.Map:
			if f.OmitEmpty && rv.Field(i).IsNil() {
				continue
			}
			fallthrough
		default:
			w.ObjKey(f.Name)
			w.marshal(rv.Field(i))
		}
	}
	w.ObjEnd()

	return w.Err()
}

func (w *Writer) marshalSlice(rv reflect.Value) error {
	if rv.IsNil() {
		w.Null()
		return w.Err()
	}

	elk := rv.Type().Elem().Kind()
	if elk == reflect.Uint8 {
		sw := w.Base64Writer(base64.RawStdEncoding)
		sw.Write(rv.Bytes())
		return sw.Close()
	}
	w.ArrayStart()
	l := rv.Len()
	for i := 0; i < l; i++ {
		vi := rv.Index(i)
		switch elk {
		case reflect.String:
			w.SafeString(UnsafeStringToBytes(vi.String()))
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			i := vi.Int()
			s := strconv.FormatInt(i, 10)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
			reflect.Uintptr:
			i := vi.Uint()
			s := strconv.FormatUint(i, 10)
			w.Number(UnsafeStringToBytes(s))
		case reflect.Float64, reflect.Float32:
			bits := 64
			if rv.Kind() == reflect.Float32 {
				bits = 32
			}
			f := vi.Float()
			s := strconv.FormatFloat(f, 'g', -1, bits)
			w.Number(UnsafeStringToBytes(s))
		default:
			w.marshal(vi)
		}
	}
	w.ArrayEnd()
	return w.Err()
}
