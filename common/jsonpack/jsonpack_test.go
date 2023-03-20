package jsonpack

import (
	"reflect"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	type args struct {
		v any
	}
	intn := -1
	str256 := strings.Repeat("1", 256)
	bytes256 := append([]byte{50, 0, 1}, []byte(str256)...)

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"uint8", args{uint8(1)}, []byte{0x11, 0x01}, false},
		{"uint8_256", args{uint8(0xff)}, []byte{0x11, 0xff}, false},
		{"uint8-1", args{int8(intn)}, []byte{0x19, 0x01}, false},

		{"uint16", args{uint16(1)}, []byte{0x11, 0x01}, false},
		{"int16-1", args{int16(intn)}, []byte{0x19, 0x01}, false},

		{"uint32", args{uint32(1)}, []byte{0x11, 0x01}, false},
		{"uint32-1144422400", args{uint32(1144422400)}, []byte{0x14, 0x00, 0x80, 0x36, 0x44}, false},
		{"int32-1", args{int32(intn)}, []byte{0x19, 0x01}, false},

		{"bool0", args{false}, []byte{0x20}, false},
		{"bool1", args{true}, []byte{0x21}, false},

		{"string3", args{"123"}, []byte{49, 3, '1', '2', '3'}, false},
		{"string safe", args{"abc\x00abc"}, []byte{0x31, 0x07, 'a', 'b', 'c', 0, 'a', 'b', 'c'}, false},
		{"string256", args{str256}, bytes256, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("map", func(t *testing.T) {
		type ms struct {
			key string
		}
		m := ms{"val"}
		got, _ := Marshal(m)
		if !reflect.DeepEqual(got, []byte("\x51\x01\x31\x03key\x31\x03val")) {
			t.Errorf("Marshal() = %v", got)
		}
	})
	t.Run("map map", func(t *testing.T) {
		type ms struct {
			key struct {
				z int32
			}
		}
		m := ms{struct{ z int32 }{1}}
		got, _ := Marshal(m)
		if !reflect.DeepEqual(got, []byte("\x51\x01\x31\x03key\x51\x01\x31\x01z\x11\x01")) {
			t.Errorf("Marshal() = %v", got)
		}
	})
	t.Run("map rename", func(t *testing.T) {
		type ms struct {
			key string `jp:"kfy"`
		}
		m := ms{"val"}
		got, _ := Marshal(m)
		if !reflect.DeepEqual(got, []byte("\x51\x01\x31\x03kfy\x31\x03val")) {
			t.Errorf("Marshal() = %v", got)
		}
	})
}

func TestUnmarshal(t *testing.T) {
	t.Run("unknown type", func(t *testing.T) {
		var val int
		want := 0
		wantErr := true
		err := Unmarshal([]byte{0x01, 0x01}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("int", func(t *testing.T) {
		var val int
		want := int(1)
		wantErr := false
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("int64", func(t *testing.T) {
		var val int64
		want := int64(1)
		wantErr := false
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("uint8", func(t *testing.T) {
		var val uint8
		want := uint8(1)
		wantErr := false
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("uint32", func(t *testing.T) {
		var val uint32
		want := uint32(1)
		wantErr := false
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("uint16", func(t *testing.T) {
		var val uint16
		want := uint16(1)
		wantErr := false
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})

	t.Run("bool error", func(t *testing.T) {
		var val bool
		want := false
		wantErr := true
		err := Unmarshal([]byte{0x22}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("bool false", func(t *testing.T) {
		var val bool
		want := false
		wantErr := false
		err := Unmarshal([]byte{0x20}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("bool true", func(t *testing.T) {
		var val bool
		want := true
		wantErr := false
		err := Unmarshal([]byte{0x21}, &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("string size error", func(t *testing.T) {
		var val string
		want := ""
		wantErr := true
		err := Unmarshal([]byte("\x32\x03abc"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("string length error", func(t *testing.T) {
		var val string
		want := "ab"
		wantErr := true
		err := Unmarshal([]byte("\x31\x02abc"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("string", func(t *testing.T) {
		var val string
		want := "abc"
		wantErr := false
		err := Unmarshal([]byte("\x31\x03abc"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("struct error", func(t *testing.T) {
		type jsonp struct {
			Key string
			B   string `jp:"b"`
		}
		var val jsonp
		wantErr := true
		err := Unmarshal([]byte("\x51\x02\x31\x03Key\x31\x03val\x31\x01b\x51\x01\x31\x01a\x11\x01"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
	})
	t.Run("struct", func(t *testing.T) {
		type subS struct {
			A int `jp:"a"`
		}
		type jsonp struct {
			Key string
			B   subS `jp:"b"`
		}
		var val jsonp
		want := jsonp{"val", subS{1}}
		wantErr := false
		err := Unmarshal([]byte("\x51\x02\x31\x03Key\x31\x03val\x31\x01b\x51\x01\x31\x01a\x11\x01"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("struct can empty", func(t *testing.T) {
		type jsonp struct {
			Key string `jp:",omitempty"`
		}
		var val jsonp
		want := jsonp{}
		wantErr := false
		err := Unmarshal([]byte("\x51\x01\x31\x03key\x31\x03val"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("struct can not empty", func(t *testing.T) {
		type jsonp struct {
			Key string
		}
		var val jsonp
		want := jsonp{}
		wantErr := true
		err := Unmarshal([]byte("\x51\x01\x31\x03key\x31\x03val"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("array", func(t *testing.T) {
		var val []string
		want := []string{"key", "val"}
		wantErr := false
		err := Unmarshal([]byte("\x41\x02\x31\x03key\x31\x03val"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
	t.Run("struct", func(t *testing.T) {
		type reqSubscribe struct {
			Evt int  `jp:"evt"`
			Act bool `jp:"act"`
			Sub int  `jp:"sub,omitempty"`
		}
		var val reqSubscribe
		want := reqSubscribe{2, true, 0}
		wantErr := false
		err := Unmarshal([]byte("Q\x021\x03evt\x11\x021\x03act!"), &val)
		if (err != nil) != wantErr {
			t.Errorf("Marshal() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(val, want) {
			t.Errorf("Marshal() = %v, want %v", val, want)
		}
	})
}
