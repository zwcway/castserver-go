package jsonpack

import (
	"reflect"
	"strings"
)

type structField struct {
	field string
	tags  map[string]bool
	r     reflect.Value
}

type Decoder struct {
	d   []byte
	pos uint32
}

func (d *Decoder) getType() (uint32, uint32, error) {
	if len(d.d) < int(d.pos)+1 {
		return 0, 0, &endError{}
	}
	by := d.d[d.pos]

	return uint32(by >> 4), uint32(by & 0x0f), nil
}
func (d *Decoder) readType() (uint32, uint32, error) {
	t, s, err := d.getType()
	if err != nil {
		return 0, 0, err
	}
	d.pos++

	return t, s, err
}
func (d *Decoder) readInt32(s uint32) (uint32, error) {
	if len(d.d) < int(d.pos)+1 {
		return 0, &endError{}
	}
	i := d.d[d.pos : d.pos+s]

	d.pos += s

	switch s {
	case 1:
		return uint32(i[0]), nil
	case 2:
		return uint32(i[0]) | (uint32(i[1]) << 8), nil
	case 4:
		return uint32(i[0]) | (uint32(i[1]) << 8) | (uint32(i[2]) << 16) | (uint32(i[3]) << 24), nil
	}

	return 0, &InvalidJsonPackError{d.pos - 1, byte(s), []byte{1, 2, 4}}
}

func (d *Decoder) needType(tt uint32) (uint32, error) {
	t, s, err := d.readType()
	if err != nil {
		return s, err
	}
	if t != tt {
		return s, &InvalidJsonPackError{d.pos - 1, byte(t), []byte{byte(tt)}}
	}

	return s, nil
}

func needReflectKind(skip bool, f string, r reflect.Value, ks ...reflect.Kind) error {
	if skip {
		return nil
	}
	kind := r.Kind()
	found := false
	for _, k := range ks {
		if k == kind {
			found = true
			break
		}
	}
	if !found {
		return &KindUnmarshalError{InvalidUnmarshalError{f, kind}, ks}
	}

	if !r.CanSet() {
		return &PrivateUnmarshalError{f}
	}

	return nil
}

func (d *Decoder) decodeString(r reflect.Value, s uint32, f string, skip bool) error {
	err := needReflectKind(skip, f, r, reflect.String)
	if err != nil {
		return err
	}

	l, err := d.readInt32(s)
	if err != nil {
		return err
	}
	if int(l) > len(d.d)-int(d.pos) {
		return &endError{}
	}

	str := string(d.d[d.pos : d.pos+l])
	d.pos += l

	if skip {
		return nil
	}

	r.SetString(str)

	return nil
}

func (d *Decoder) needString(r reflect.Value, f string) (string, error) {
	var str string

	s, err := d.needType(JSONPACK_STRING)
	if err != nil {
		return str, err
	}
	l, err := d.readInt32(s)
	if err != nil {
		return str, err
	}
	if int(l) > len(d.d)-int(d.pos) {
		return str, &endError{}
	}

	str = string(d.d[d.pos : d.pos+l])

	d.pos += l

	return str, nil
}

func (d *Decoder) decodeInt32(r reflect.Value, s uint32, f string, skip bool) error {
	err := needReflectKind(skip, f, r, reflect.Int8, reflect.Uint8, reflect.Int, reflect.Uint,
		reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64)
	if err != nil {
		return err
	}

	i, err := d.readInt32(s)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	switch r.Kind() {
	case reflect.Int8, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		r.SetInt(int64(i))
	case reflect.Uint8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		r.SetUint(uint64(i))
	}

	return nil
}

func (d *Decoder) decodeBool(r reflect.Value, s uint32, f string, skip bool) error {
	err := needReflectKind(skip, f, r, reflect.Bool)
	if err != nil {
		return err
	}

	if s != 0 && s != 1 {
		return &InvalidJsonPackError{d.pos - 1, byte(s), []byte{0, 1}}
	}
	if skip {
		return nil
	}

	r.SetBool(s != 0)

	return nil
}

func convertType(t uint32) []reflect.Kind {
	switch t {
	case JSONPACK_ARRAY:
		return []reflect.Kind{reflect.Slice}
	case JSONPACK_BOOLEAN:
		return []reflect.Kind{reflect.Bool}
	case JSONPACK_MAP:
		return []reflect.Kind{reflect.Struct}
	case JSONPACK_NUMBER:
		return []reflect.Kind{reflect.Int8, reflect.Uint8, reflect.Int, reflect.Uint,
			reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64}
	case JSONPACK_STRING:
		return []reflect.Kind{reflect.String}
	}
	return []reflect.Kind{}
}

func (d *Decoder) decodeArray(r reflect.Value, s uint32, f string, skip bool) error {
	err := needReflectKind(skip, f, r, reflect.Slice)
	if err != nil {
		return err
	}

	l, err := d.readInt32(s)
	if err != nil {
		return err
	}

	if l == 0 {
		// 数组是空的
		return nil
	}

	lastT := uint32(0)
	for i := 0; i < int(l); i++ {
		// 保证数组元素类型一致
		t, _, _ := d.getType()
		if lastT == 0 {
			lastT = t
		} else if lastT != t {
			return &InvalidJsonPackError{d.pos, byte(t), []byte{byte(lastT)}}
		}
		nrv := reflect.New(r.Type().Elem())
		err = d.reflectObject(nrv.Elem(), f, skip)
		if err != nil {
			return err
		}
		if !skip {
			r.Set(reflect.Append(r, nrv.Elem()))
		}
	}

	return nil
}

func (d *Decoder) reflectMap2Struct(r reflect.Value, l uint32, f string, skip bool) error {
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

	// 对于未指定 omit 的键，需要抛出异常
	for k, f := range fields {
		if _, ok := f.tags["omit"]; !ok {
			return &EmptyUnmarshalError{k}
		}
	}

	return nil
}

func (d *Decoder) reflectMap(r reflect.Value, s uint32, f string, skip bool) error {
	err := needReflectKind(skip, f, r, reflect.Struct)
	if err != nil {
		return err
	}
	l, err := d.readInt32(s)
	if err != nil {
		return err
	}

	if r.Kind() == reflect.Struct {
		return d.reflectMap2Struct(r, l, f, skip)
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
		err = d.decodeInt32(r, s, f, skip)
	case JSONPACK_STRING:
		err = d.decodeString(r, s, f, skip)
	default:
		return &InvalidJsonPackError{d.pos - 1, byte(t), []byte{1, 2, 3, 4, 5}}
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

	if int(d.pos) != len(d.d) {
		return &endError{}
	}

	return nil
}

func newDecoder(b []byte) *Decoder {
	return &Decoder{b, 0}
}

func (d *Decoder) Resume(v any) error {
	return d.reflectDecode(reflect.ValueOf(v))
}
