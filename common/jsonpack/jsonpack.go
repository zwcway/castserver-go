package jsonpack

import (
	"reflect"
)

func Marshal(v any) ([]byte, error) {
	e := Encoder{buf: make([]byte, 512)}

	err := e.reflectValue(reflect.ValueOf(v), "")

	return e.buf[:e.pos], err
}

func Unmarshal(b []byte, v any) error {
	d := newDecoder(b)

	return d.reflectDecode(reflect.ValueOf(v))
}
