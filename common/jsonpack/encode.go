package jsonpack

import (
	"math"
	"reflect"
	"strconv"
	"strings"
)

type Encoder struct {
	buf []byte
	pos int
}

func (j *Encoder) write(bytes ...byte) {
	j.writeBytes(bytes)
}

func (j *Encoder) writeBytes(bytes []byte) {
	l := j.pos + len(bytes) - len(j.buf)
	if l > 0 {
		ob := j.buf
		j.buf = make([]byte, len(j.buf)*2+l)
		copy(j.buf, ob)
	}
	for _, b := range bytes {
		j.buf[j.pos] = b
		j.pos++
	}
}

func (j *Encoder) intSize(val uint64) uint8 {
	// 对于 number 类型，第4位表示负数，因此最多7个字节
	if val <= 0xff {
		return 1
	} else if val <= 0xff_ff {
		return 2
	} else if val <= 0xff_ff_ff {
		return 3
	} else if val <= 0xff_ff_ff_ff {
		return 4
	} else if val <= 0xff_ff_ff_ff_ff {
		return 5
	} else if val <= 0xff_ff_ff_ff_ff_ff {
		return 6
	} else if val <= 0xff_ff_ff_ff_ff_ff_ff {
		return 7
	} else {
		return 0
	}
}

func (j *Encoder) writeType(t uint8, size uint8) {
	j.write(byte((t<<4)&0xF0 | size&0x0F))
}

func (j *Encoder) writeInteger(v uint32, size uint8) {
	switch size & 0x07 {
	case 1:
		j.write(byte(v))
	case 2:
		j.write(byte(v), byte(v>>8))
	case 3:
		j.write(byte(v), byte(v>>8), byte(v>>16))
	case 4:
		j.write(byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
	}
}

func (j *Encoder) writeFloat(t float32) {
	bits := math.Float32bits(t)
	j.writeInteger(bits, 4)
}

func (j *Encoder) writeFloat64(t float64) {
	b := math.Float64bits(t)
	j.writeInteger(uint32(b), 4)
	j.writeInteger(uint32(b>>32), 4)
}

func (j *Encoder) EncodeInt8(val int8) {
	if val >= 0 {
		j.EncodeUint32(uint32(val))
	} else {
		j.EncodeInt32(-int32(-val))
	}
}
func (j *Encoder) EncodeUint8(val uint8) {
	j.EncodeUint32(uint32(val))
}

func (j *Encoder) EncodeInt16(val int16) {
	if val >= 0 {
		j.EncodeUint32(uint32(val))
	} else {
		j.EncodeInt32(-int32(-val))
	}
}
func (j *Encoder) EncodeUint16(val uint16) {
	j.EncodeUint32(uint32(val))
}

func (j *Encoder) EncodeInt32(val int32) {
	j.EncodeInt64(int64(val))
}

func (j *Encoder) EncodeInt64(val int64) {
	var i uint64 = uint64(val)
	var size = j.intSize(i)
	if val < 0 {
		i = uint64(-val)
		size = j.intSize(i) | 0x08
		j.writeType(JSONPACK_NUMBER, size|0x08)
	} else {
		j.writeType(JSONPACK_NUMBER, size)
	}

	size = size & 0x07
	if size > 4 {
		j.writeInteger(uint32(i), 4)
		j.writeInteger(uint32(i>>32), size-4)
	} else {
		j.writeInteger(uint32(i), size)
	}
}

func (j *Encoder) EncodeUint32(val uint32) {
	size := j.intSize(uint64(val))
	j.writeType(JSONPACK_NUMBER, size)
	j.writeInteger(val, size)
}

func (j *Encoder) EncodeFloat32(val float32) {
	j.writeType(JSONPACK_FLOAT, 4)
	j.writeFloat(val)
}

func (j *Encoder) EncodeFloat64(val float64) {
	j.writeType(JSONPACK_FLOAT, 8)
	j.writeFloat64(val)
}

func (j *Encoder) EncodeBool(val bool) {
	if val {
		j.writeType(JSONPACK_BOOLEAN, 1)
	} else {
		j.writeType(JSONPACK_BOOLEAN, 0)
	}
}

func (j *Encoder) EncodeNull() {
	j.writeType(JSONPACK_NULL, 0)
}

func (j *Encoder) EncodeString(val string) {
	j.EncodeBinary([]byte(val))
}

// 4 位：       0-15 类型
// 4 位：       0-7  字符串长度的字节数量
// 0-7 个字节： 支持长度 0 - 0xFF FFFF FFFF FFFF (0 - 72,057,594,037,927,935)
func (j *Encoder) EncodeBinary(val []byte) {
	len := uint32(len(val))
	size := j.intSize(uint64(len))
	j.writeType(JSONPACK_STRING, size)
	j.writeInteger(len, size)
	j.writeBytes(val)
}

func (j *Encoder) EncodeArray(len uint32) {
	size := j.intSize(uint64(len))
	j.writeType(JSONPACK_ARRAY, size)
	j.writeInteger(len, size)
}

func (j *Encoder) EncodeMap(len uint32) {
	size := j.intSize(uint64(len))
	j.writeType(JSONPACK_MAP, size)
	j.writeInteger(len, size)
}

func (j *Encoder) reflectArray(r reflect.Value, t reflect.Type, field string) error {
	len := r.Len()
	j.EncodeArray(uint32(len))
	for i := 0; i < len; i++ {
		err := j.reflectValue(r.Index(i), field+"."+strconv.Itoa(i))
		if err != nil {
			return err
		}
	}
	return nil
}

type structFieldInfo struct {
	name string
	idx  int
	tr   reflect.Value
	tf   reflect.Type
}

func isEmpty(r reflect.Value) bool {
	switch r.Kind() {
	case reflect.Slice:
		return r.Len() == 0
	case reflect.Pointer:
		return r.IsNil()
	}

	return false
}

func (j *Encoder) collectMap(r reflect.Value, t reflect.Type) []structFieldInfo {
	l := r.NumField()
	ss := []structFieldInfo{}

	for i := 0; i < l; i++ {
		tf := t.Field(i)
		if tf.Anonymous {
			as := j.collectMap(r.Field(i), tf.Type)
			ss = append(ss, as...)
			continue
		}
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

		if omitempty && isEmpty(field) {
			continue
		}

		ss = append(ss, structFieldInfo{
			name: name,
			idx:  i,
			tf:   tf.Type,
			tr:   r.Field(i),
		})
	}

	return ss
}

func (j *Encoder) reflectMap(r reflect.Value, t reflect.Type, field string) error {
	ss := j.collectMap(r, t)

	j.EncodeMap(uint32(len(ss)))
	for _, s := range ss {
		tf := s.tf
		j.EncodeString(s.name)
		err := j.reflectValue(s.tr, field+"."+tf.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Encoder) reflectValue(r reflect.Value, field string) (err error) {
	if !r.IsValid() {
		return &InvalidValueError{field: field}
	}
	switch r.Kind() {
	case reflect.Pointer:
		if r.IsNil() {
			j.EncodeNull()
			return
		}
		err = j.reflectValue(r.Elem(), field)
	case reflect.Struct:
		err = j.reflectMap(r, r.Type(), field)
	case reflect.Array, reflect.Slice:
		err = j.reflectArray(r, r.Type(), field)
	case reflect.String:
		j.EncodeString(r.String())
	case reflect.Bool:
		j.EncodeBool(r.Bool())
	case reflect.Int, reflect.Int32:
		j.EncodeInt32(int32(r.Int()))
	case reflect.Uint, reflect.Uint32:
		j.EncodeUint32(uint32(r.Uint()))
	case reflect.Int8:
		j.EncodeInt8(int8(r.Int()))
	case reflect.Uint8:
		j.EncodeUint8(uint8(r.Uint()))
	case reflect.Int16:
		j.EncodeInt16(int16(r.Int()))
	case reflect.Uint16:
		j.EncodeUint16(uint16(r.Uint()))
	case reflect.Int64: // js 不支持64位整数，转为千分字符串
		j.EncodeInt64(r.Int())
	case reflect.Uint64:
		j.EncodeInt64(int64(r.Uint()))
	case reflect.Float32:
		j.EncodeFloat32(float32(r.Float()))
	case reflect.Float64:
		j.EncodeFloat64(r.Float())
	default:
		return &InvalidValueError{field, r.Kind()}
	}
	return
}
