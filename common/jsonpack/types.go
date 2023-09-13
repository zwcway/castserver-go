package jsonpack

import (
	"fmt"
	"reflect"
)

const (
	JSONPACK_NUMBER uint8 = iota + 1
	JSONPACK_BOOLEAN
	JSONPACK_STRING
	JSONPACK_ARRAY
	JSONPACK_MAP
	JSONPACK_FLOAT
	JSONPACK_NULL
)

type InvalidValueError struct {
	field string
	kind  reflect.Kind
}

func (e *InvalidValueError) Error() string {
	return fmt.Sprintf("Unsupport kind(%v) of '%s'", e.kind, e.field)
}

type InvalidUnmarshalError struct {
	field string
	kind  reflect.Kind
}

func (e *InvalidUnmarshalError) Error() string {
	return fmt.Sprintf("unmarshal type '%v' invalid for '%s'", e.kind, e.field)
}

type KindUnmarshalError struct {
	InvalidUnmarshalError
	want []reflect.Kind
}

func (e *KindUnmarshalError) Error() string {
	return e.InvalidUnmarshalError.Error() + fmt.Sprintf(" need %v", e.want)
}

type PrivateUnmarshalError struct {
	field string
}

func (e *PrivateUnmarshalError) Error() string {
	return fmt.Sprintf("unmarshal field '%v' is private", e.field)
}

type EmptyUnmarshalError struct {
	field string
}

func (e *EmptyUnmarshalError) Error() string {
	return fmt.Sprintf("field '%s' can not empty", e.field)
}

type InvalidJsonPackError struct {
	pos  int
	data byte
	want []byte
}

func (e *InvalidJsonPackError) Error() string {
	return fmt.Sprintf("data invalid on %d got %d expect %v", e.pos, e.data, e.want)
}

type endError struct{}

func (e *endError) Error() string {
	return "data eof"
}
