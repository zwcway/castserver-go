package jsonpack

import (
	"io"
	"math"
	"reflect"
	"strings"
)

type structField struct {
	field string
	tags  map[string]bool
	r     reflect.Value
}

type Decoder struct {
	d    []byte
	pos  int
	left int
	r    io.Reader
}

func (d *Decoder) sure(s int64) (left int64, err error) {
	if int64(d.pos)+s <= int64(d.left) {
		return
	}

	if d.r == nil {
		err = &endError{}
		return
	}

	copy(d.d, d.d[d.pos:])
	i := d.left - d.pos
	d.pos = 0

	n, err := d.r.Read(d.d[i:])
	d.left = i + n

	// 剩余的缓存不足，计算出不足的待请求的大小
	left = int64(d.pos) + s - int64(d.left)
	if left < 0 {
		left = 0
	}

	return
}

func (d *Decoder) getType() (uint8, uint8, error) {
	if e, _ := d.sure(1); e > 0 {
		return 0, 0, &endError{}
	}
	by := d.d[d.pos]

	return uint8(by >> 4), uint8(by & 0x0f), nil
}

func (d *Decoder) readType() (uint8, uint8, error) {
	t, s, err := d.getType()
	if err != nil {
		return 0, 0, err
	}
	d.pos++

	return t, s, err
}

func (d *Decoder) readInt64(s uint8) (int64, error) {
	if e, _ := d.sure(int64(s)); e > 0 {
		return 0, &endError{}
	}
	s &= 0x07

	i := d.d[d.pos : d.pos+int(s)]
	d.pos += int(s)

	if s > 7 || s < 1 {
		return 0, &InvalidJsonPackError{d.pos - 1, byte(s), []byte{1, 2, 4}}
	}

	val := int64(0)
	for idx := 0; idx < int(s); idx++ {
		val |= int64(i[idx]) << uint(idx*8)
	}

	return val, nil
}

// 当前数据必须符合指定类型
func (d *Decoder) needType(tt uint8) (uint8, error) {
	t, s, err := d.readType()
	if err != nil {
		return s, err
	}
	if t != tt {
		return s, &InvalidJsonPackError{d.pos - 1, byte(t), []byte{byte(tt)}}
	}

	return s, nil
}

// 当前存储结构必须符合指定类型
func needReflectKind(skip bool, f string, r reflect.Value, ks ...reflect.Kind) (rp reflect.Value, err error) {
	if skip {
		return
	}

	kind := r.Kind()
	rp = r
	if kind == reflect.Pointer {
		rp = r.Elem()
		kind = r.Type().Elem().Kind()
	}
	found := false
	for _, k := range ks {
		if k == kind {
			found = true
			break
		}
	}
	if !found {
		err = &KindUnmarshalError{InvalidUnmarshalError{f, kind}, ks}
		return
	}

	if !r.CanSet() {
		err = &PrivateUnmarshalError{f}
		return
	}

	return
}

func (d *Decoder) parseString(s uint8, skip bool) (strs string, err error) {
	l, err := d.readInt64(s)
	if err != nil {
		return
	}

	var (
		str []byte
		i   int64
		end int
		e   int64
	)
	if !skip {
		str = make([]byte, l)
	}

	for {
		e, err = d.sure(l)
		if err != nil {
			if e > 0 {
				err = &endError{}
				return
			}
			return
		}
		if e > 0 {
			i += l - e
			end = d.left
		} else {
			end = d.pos + int(l)
		}
		l = e
		if !skip {
			copy(str[i:], d.d[d.pos:end])
		}
		d.pos = end

		if l == 0 {
			break
		}
	}

	strs = string(str)

	return
}

func (d *Decoder) decodeString(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.String)
	if err != nil {
		return err
	}

	str, err := d.parseString(s, skip)
	if err != nil {
		return err
	}

	if !skip {
		rp.SetString(str)
	}

	return nil
}

func (d *Decoder) needString(r reflect.Value, f string) (str string, err error) {
	s, err := d.needType(JSONPACK_STRING)
	if err != nil {
		return
	}

	str, err = d.parseString(s, false)

	return
}

func (d *Decoder) decodeInt64(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Int8, reflect.Uint8, reflect.Int, reflect.Uint,
		reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64,
		reflect.Float32, reflect.Float64)
	if err != nil {
		return err
	}
	var (
		i    int64
		size uint8 = s & 0x07
		neg  bool  = s&0x08 > 0
	)
	if size > 4 {
		il, err := d.readInt64(4)
		if err != nil {
			return err
		}
		ih, err := d.readInt64(size - 4)
		if err != nil {
			return err
		}
		i = int64(il) | (int64(ih) << 32)
	} else {
		i32, err := d.readInt64(size)
		if err != nil {
			return err
		}
		i = int64(i32)
	}
	if skip {
		return nil
	}
	switch rp.Kind() {
	case reflect.Int8:
		if neg {
			rp.SetInt(int64(int8(-i)))
		} else {
			rp.SetInt(int64(int8(i)))
		}
	case reflect.Int, reflect.Int32:
		if neg {
			rp.SetInt(int64(int32(-i)))
		} else {
			rp.SetInt(int64(int32(i)))
		}
	case reflect.Int16:
		if neg {
			rp.SetInt(int64(int16(-i)))
		} else {
			rp.SetInt(int64(int16(i)))
		}
	case reflect.Int64:
		if neg {
			rp.SetInt(-i)
		} else {
			rp.SetInt(i)
		}
	case reflect.Uint8:
		if neg {
			rp.SetUint(uint64(uint8(-i)))
		} else {
			rp.SetUint(uint64(uint8(i)))
		}
	case reflect.Uint16:
		if neg {
			rp.SetUint(uint64(uint16(-i)))
		} else {
			rp.SetUint(uint64(uint16(i)))
		}
	case reflect.Uint, reflect.Uint32:
		if neg {
			rp.SetUint(uint64(uint32(-i)))
		} else {
			rp.SetUint(uint64(uint32(i)))
		}
	case reflect.Uint64:
		if neg {
			rp.SetUint(uint64(-i))
		} else {
			rp.SetUint(uint64(i))
		}
	case reflect.Float32:
		if neg {
			rp.SetFloat(float64(float32(-i)))
		} else {
			rp.SetFloat(float64(float32(i)))
		}
	case reflect.Float64:
		if neg {
			rp.SetFloat(float64(-i))
		} else {
			rp.SetFloat(float64(i))
		}
	}

	return nil
}

func (d *Decoder) decodeFloat32(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Float32, reflect.Float64)
	if err != nil {
		return err
	}

	i, err := d.readInt64(s)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	bits := math.Float32frombits(uint32(i))
	rp.SetFloat(float64(bits))

	return nil
}

func (d *Decoder) decodeFloat64(r reflect.Value, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Float32, reflect.Float64)
	if err != nil {
		return err
	}

	il, err := d.readInt64(4)
	if err != nil {
		return err
	}
	ih, err := d.readInt64(4)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	bits := math.Float64frombits(uint64(il) | (uint64(ih) << 32))
	rp.SetFloat(bits)

	return nil
}

func (d *Decoder) decodeBool(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Bool)
	if err != nil {
		return err
	}

	if s != 0 && s != 1 {
		return &InvalidJsonPackError{d.pos - 1, byte(s), []byte{0, 1}}
	}
	if skip {
		return nil
	}

	rp.SetBool(s != 0)

	return nil
}

func (d *Decoder) decodeNull(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Pointer)
	if err != nil {
		return err
	}

	if s != 0 {
		return &InvalidJsonPackError{d.pos - 1, byte(s), []byte{0}}
	}
	if skip {
		return nil
	}

	rp.Set(reflect.Zero(rp.Type()))

	return nil
}

func (d *Decoder) decodeArray(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Slice, reflect.Array)
	if err != nil {
		return err
	}

	l, err := d.readInt64(s)
	if err != nil {
		return err
	}

	if l == 0 {
		// 数组是空的
		return nil
	}

	lastT := uint8(0)
	for i := 0; i < int(l); i++ {
		// 保证数组元素类型一致
		t, _, _ := d.getType()
		if lastT == 0 {
			lastT = t
		} else if lastT != t {
			return &InvalidJsonPackError{d.pos, byte(t), []byte{byte(lastT)}}
		}
		nrv := reflect.New(rp.Type().Elem())
		err = d.reflectObject(nrv.Elem(), f, skip)
		if err != nil {
			return err
		}
		if !skip {
			rp.Set(reflect.Append(r, nrv.Elem()))
		}
	}

	return nil
}

func (d *Decoder) reflectMap2Struct(r reflect.Value, l int64, f string, skip bool) error {
	// 预先整理出struct中的Tag
	fc := r.NumField()
	fields := make(map[string]structField, l)
	rt := r.Type()

	if skip {
		fc = 0
	}
	for i := 0; i < fc; i++ {
		tf := rt.Field(i)
		tags := strings.Split(tf.Tag.Get("jp"), ",")
		name := tf.Name
		if len(tags) > 0 && len(tags[0]) > 0 {
			name = tags[0]
		}
		tm := make(map[string]bool, len(tags)-1)
		for _, t := range tags[1:] {
			tm[t] = true
		}
		fields[name] = structField{tf.Name, tm, r.Field(i)}
	}

	for i := 0; i < int(l); i++ {
		key, err := d.needString(r, f)
		if err != nil {
			return err
		}
		if sf, ok := fields[key]; ok {
			if _, ok := sf.tags["stop"]; ok {
				return nil
			}
			err = d.reflectObject(sf.r, key, skip)
			delete(fields, key)
		} else {
			// struct中没有该键，但是解码需要继续
			var sr reflect.Value
			err = d.reflectObject(sr, key, true)
		}
		if err != nil {
			return err
		}
	}

	// 对于未指定 omitempty 的键，需要抛出异常
	for k, f := range fields {
		if _, ok := f.tags["omitempty"]; !ok {
			return &EmptyUnmarshalError{k}
		}
	}

	return nil
}

func (d *Decoder) reflectMap(r reflect.Value, s uint8, f string, skip bool) error {
	rp, err := needReflectKind(skip, f, r, reflect.Struct)
	if err != nil {
		return err
	}
	l, err := d.readInt64(s)
	if err != nil {
		return err
	}

	if r.Kind() == reflect.Struct {
		return d.reflectMap2Struct(rp, l, f, skip)
	}

	// kk := r.Type().Key().Kind()
	// if kk != reflect.String {
	// 	return &InvalidUnmarshalError{f, kk}
	// }
	// vk := r.Type().Elem().Kind()

	// for i := 0; i < int(l); i++ {
	// 	key, err := d.needString(r, f)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	t, _, err := d.getType()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = d.reflectObject(val, key, skip)

	// 	if !skip {
	// 		r.SetMapIndex(reflect.ValueOf(key), val)
	// 	}

	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (d *Decoder) reflectObject(r reflect.Value, f string, skip bool) error {

	if !skip {
		if ok, err := d.checkAndReflectBytes(r); ok {
			return err
		}
	}

	t, s, err := d.readType()
	if err != nil {
		return err
	}

	switch t {
	case JSONPACK_MAP:
		err = d.reflectMap(r, s, f, skip)
	case JSONPACK_ARRAY:
		err = d.decodeArray(r, s, f, skip)
	case JSONPACK_BOOLEAN:
		err = d.decodeBool(r, s, f, skip)
	case JSONPACK_NUMBER:
		err = d.decodeInt64(r, s, f, skip)
	case JSONPACK_FLOAT:
		if s == 8 {
			err = d.decodeFloat64(r, f, skip)
		} else if s == 4 {
			err = d.decodeFloat32(r, s, f, skip)
		} else {
			err = &InvalidJsonPackError{d.pos - 1, byte(t), []byte{1, 2, 3, 4, 5, 6}}
		}
	case JSONPACK_STRING:
		err = d.decodeString(r, s, f, skip)
	case JSONPACK_NULL:
		err = d.decodeNull(r, s, f, skip)
	default:
		return &InvalidJsonPackError{d.pos - 1, byte(t), []byte{1, 2, 3, 4, 5, 6}}
	}

	return err
}

func (d *Decoder) reflectDecode(r reflect.Value) error {
	if r.Kind() != reflect.Pointer || r.IsNil() {
		return &InvalidUnmarshalError{"", r.Kind()}
	}
	r = r.Elem()

	err := d.reflectObject(r, "", false)
	if err != nil {
		return err
	}

	if int(d.pos) != d.left {
		return &endError{}
	}

	return nil
}

func (d *Decoder) checkAndReflectBytes(r reflect.Value) (ok bool, err error) {
	if r.Kind() != reflect.Slice || r.Type().Elem().Kind() != reflect.Uint8 {
		return
	}
	ok = true
	pos := d.pos

	err = d.reflectObject(r, "", true)
	if err != nil {
		return
	}

	r.Set(reflect.ValueOf(d.d[pos:d.pos]))
	return
}

func (d *Decoder) ReadFrom(rd io.Reader, v any) (s int, err error) {
	var n int
	for {
		n, err = rd.Read(d.d)
		s += n
		if err != nil {
			return
		}
		err = d.reflectDecode(reflect.ValueOf(v))

		if _, ok := err.(*endError); ok {
			copy(d.d[:len(d.d)-int(d.pos)], d.d[int(d.pos):])
			d.pos = 0
			continue
		}
	}
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{
		d:   b,
		pos: 0,
		r:   nil,
	}
}

func NewDecoderFromReader(rd io.Reader, bufSize int) *Decoder {
	return &Decoder{
		d:   make([]byte, bufSize),
		pos: 0,
		r:   rd,
	}
}

func (d *Decoder) Resume(v any) error {
	return d.reflectDecode(reflect.ValueOf(v))
}
