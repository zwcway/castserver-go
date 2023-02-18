package jsonpack

import (
	"reflect"
)

func Marshal(v any) ([]byte, error) {
	e := Encoder{}

	err := e.reflectValue(reflect.ValueOf(v), "")

	return e, err
}

func Unmarshal(b []byte, v any) error {
	d := newDecoder(b)

	return d.reflectDecode(reflect.ValueOf(v))
}
