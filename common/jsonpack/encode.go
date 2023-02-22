package jsonpack

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

type Encoder []byte

func (j *Encoder) intSize(val uint32) uint32 {
	if val <= 0xff {
		return 1
	} else if val <= 0xffff {
		return 2
	} else {
		return 4
	}
}

func (j *Encoder) writeType(t uint32, size uint32) {
	*j = append(*j, byte((t<<4)&0xF0|size&0x0F))
}

func (j *Encoder) writeInteger(t uint32, size uint32) {
	if size == 1 {
		*j = append(*j, byte(t))
	} else if size == 2 {
		*j = append(*j, byte(t), byte(t>>8))
	} else if size == 4 {
		*j = append(*j, byte(t), byte(t>>8), byte(t>>16), byte(t>>24))
	}
}
func (j *Encoder) writeFloat(t float32) {
	bits := math.Float32bits(t)
	j.writeInteger(bits, 4)
}

func (j *Encoder) encodeInt8(val uint8) {
	j.encodeInt32(uint32(val))
}

func (j *Encoder) encodeInt16(val uint16) {
	j.encodeInt32(uint32(val))
}

func (j *Encoder) encodeInt32(val uint32) {
	size := j.intSize(val)
	j.writeType(JSONPACK_NUMBER, size)
	j.writeInteger(val, size)
}

func (j *Encoder) encodeFloat32(val float32) {
	j.writeType(JSONPACK_FLOAT, 4)
	j.writeFloat(val)
}

func (j *Encoder) encodeBool(val bool) {
	if val {
		j.writeType(JSONPACK_BOOLEAN, 1)
	} else {
		j.writeType(JSONPACK_BOOLEAN, 0)
	}
}

func (j *Encoder) encodeString(val string) {
	j.EncodeBinary([]byte(val))
}

func (j *Encoder) EncodeBinary(val []byte) {
	len := uint32(len(val))
	size := j.intSize(len)
	j.writeType(JSONPACK_STRING, size)
	j.writeInteger(len, size)
	*j = append(*j, val...)
}

func (j *Encoder) EncodeArray(len uint32) {
	size := j.intSize(len)
	j.writeType(JSONPACK_ARRAY, size)
	j.writeInteger(len, size)
}

func (j *Encoder) EncodeMap(len uint32) {
	size := j.intSize(len)
	j.writeType(JSONPACK_MAP, size)
	j.writeInteger(len, size)
}

func (j *Encoder) reflectArray(r reflect.Value, t reflect.Type) {
	len := r.Len()
	j.EncodeArray(uint32(len))
	for i := 0; i < len; i++ {
		j.reflectValue(r.Index(i), fmt.Sprintf("%d", i))
	}
}

type structFieldInfo struct {
	name string
	idx  int
}

func (j *Encoder) reflectMap(r reflect.Value, t reflect.Type) {
	l := r.NumField()
	ss := []structFieldInfo{}

	for i := 0; i < l; i++ {
		tf := t.Field(i)
		name := tf.Tag.Get("jp")
		tags := strings.Split(name, ",")
		if name == "" {
			name = tf.Name
		} else {
			name = tags[0]
		}

		omitempty := false
		if len(tags) > 1 {
			for _, t := range tags[1:] {
				if t == "omitempty" {
					omitempty = true
				}
			}
		}
		field := r.Field(i)

		if omitempty {
			switch field.Kind() {
			case reflect.Slice:
				if field.Len() == 0 {
					continue
				}
			case reflect.Pointer:
				if field.IsNil() {
					continue
				}
			}
		}

		ss = append(ss, structFieldInfo{
			name: name,
			idx:  i,
		})
	}

	j.EncodeMap(uint32(len(ss)))
	for _, s := range ss {
		tf := t.Field(s.idx)
		name := s.name

		j.encodeString(name)
		j.reflectValue(r.Field(s.idx), tf.Name)
	}
}

func (j *Encoder) reflectValue(r reflect.Value, field string) error {
	if !r.IsValid() {
		return &InvalidValueError{}
	}
	switch r.Kind() {
	case reflect.Pointer:
		j.reflectValue(r.Elem(), field)
	case reflect.Struct:
		j.reflectMap(r, r.Type())
	case reflect.Array, reflect.Slice:
		j.reflectArray(r, r.Type())
	case reflect.String:
		j.encodeString(r.String())
	case reflect.Bool:
		j.encodeBool(r.Bool())
	case reflect.Int, reflect.Int32:
		j.encodeInt32(uint32(r.Int()))
	case reflect.Uint, reflect.Uint32:
		j.encodeInt32(uint32(r.Uint()))
	case reflect.Int8:
		j.encodeInt8(uint8(r.Int()))
	case reflect.Uint8:
		j.encodeInt8(uint8(r.Uint()))
	case reflect.Int16:
		j.encodeInt16(uint16(r.Int()))
	case reflect.Uint16:
		j.encodeInt16(uint16(r.Uint()))
	case reflect.Float32:
		j.encodeFloat32(float32(r.Float()))
	default:
		return &InvalidValueError{field, r.Kind()}
	}
	return nil
}
