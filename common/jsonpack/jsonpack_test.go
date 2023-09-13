package jsonpack

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	str256 := strings.Repeat("1", 256)
	bytes256 := append([]byte{0x32, 0x00, 0x01}, []byte(str256)...)

	tests := []struct {
		name    string
		args    any
		want    []byte
		wantErr bool
	}{
		{"uint8_1", uint8(1), []byte{0x11, 0x01}, false},
		{"uint8_256", uint8(0xff), []byte{0x11, 0xff}, false},
		{"uint8_-1", int8(-1), []byte{0x19, 0x01}, false},

		{"uint16_1", uint16(1), []byte{0x11, 0x01}, false},
		{"int16_-1", int16(-1), []byte{0x19, 0x01}, false},

		{"uint32_1", uint32(1), []byte{0x11, 0x01}, false},
		{"uint32_1144422400", uint32(1144422400), []byte{0x14, 0x00, 0x80, 0x36, 0x44}, false},
		{"int32_-1", int32(-1), []byte{0x19, 0x01}, false},

		{"int64_1", int64(1), []byte{0x11, 0x01}, false},
		{"int64_256", int64(256), []byte{0x12, 0x00, 0x01}, false},
		{"int64_65535", int64(65535), []byte{0x12, 0xFF, 0xFF}, false},
		{"int64_65536", int64(65536), []byte{0x13, 0x00, 0x00, 0x01}, false},
		{"int64_16777216", int64(16_777_216), []byte{0x14, 0x00, 0x00, 0x00, 0x01}, false},
		{"int64_4294967296", int64(4_294_967_296), []byte{0x15, 0x00, 0x00, 0x00, 0x00, 0x01}, false},
		{"int64_-4294967296", int64(-4_294_967_296), []byte{0x1D, 0x00, 0x00, 0x00, 0x00, 0x01}, false},
		{"int64_-1", int64(-1), []byte{0x19, 0x01}, false},

		{"bool_0", false, []byte{0x20}, false},
		{"bool_1", true, []byte{0x21}, false},

		{"string_3", "123", []byte{0x31, 3, '1', '2', '3'}, false},
		{"string_safe", "abc\x00abc", []byte{0x31, 0x07, 'a', 'b', 'c', 0, 'a', 'b', 'c'}, false},
		{"string_256", str256, bytes256, false},

		{"null", nil, []byte{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, got, tt.want)
		})
	}

	t.Run("map", func(t *testing.T) {
		type ms struct {
			key string
		}
		m := ms{"val"}
		got, err := Marshal(m)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, got, []byte("\x51\x01\x31\x03key\x31\x03val"))
	})
	t.Run("map map", func(t *testing.T) {
		type ms struct {
			key struct {
				z int32
			}
		}
		m := ms{struct{ z int32 }{1}}
		got, err := Marshal(m)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, got, []byte("\x51\x01\x31\x03key\x51\x01\x31\x01z\x11\x01"))
	})
	t.Run("map rename", func(t *testing.T) {
		type ms struct {
			key string `jp:"kfy"`
		}
		m := ms{"val"}
		got, err := Marshal(m)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, got, []byte("\x51\x01\x31\x03kfy\x31\x03val"))
	})
	t.Run("map null", func(t *testing.T) {
		type ms struct {
			key *string
		}
		m := ms{nil}
		got, err := Marshal(m)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, got, []byte("\x51\x01\x31\x03key\x70"))
	})
	t.Run("map int64", func(t *testing.T) {
		type ms struct {
			key time.Duration
		}
		m := ms{0}
		got, err := Marshal(m)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, got, []byte("\x51\x01\x31\x03key\x11\x00"))
	})
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("unknown type", func(t *testing.T) {
		var val int
		err := Unmarshal([]byte{0x01, 0x01}, &val)
		assert.Equal(t, err != nil, true)
		assert.Equal(t, val, 0)
	})
	{
		for _, tt := range []struct {
			name    string
			want    int
			arg     []byte
			wantErr bool
		}{
			{"size too long", 1, []byte{0x11, 0x01, 0x00}, true},
			{"size too short", 0, []byte{0x12, 0x01}, true},
		} {
			t.Run(tt.name, func(t *testing.T) {
				var val int
				err := Unmarshal(tt.arg, &val)
				assert.Equal(t, err != nil, tt.wantErr)
				assert.Equal(t, val, tt.want)
			})
		}
	}
	{
		for _, tt := range []struct {
			name    string
			want    int
			arg     []byte
			wantErr bool
		}{
			{"int_1", 1, []byte{0x11, 0x01}, false},
			{"int_255", 255, []byte{0x11, 0xFF}, false},
			{"int_256", 256, []byte{0x12, 0x00, 0x01}, false},
			{"int_65535", 65535, []byte{0x12, 0xFF, 0xFF}, false},
			{"int_65536", 65536, []byte{0x13, 0x00, 0x00, 0x01}, false},
			{"int_1677216", 16_777_216, []byte{0x14, 0x00, 0x00, 0x00, 0x01}, false},
			{"int_4294967296", 0, []byte{0x15, 0x00, 0x00, 0x00, 0x00, 0x01}, false},
			{"int_-4294967296", 0, []byte{0x1D, 0x00, 0x00, 0x00, 0x00, 0x01}, false},
		} {
			t.Run(tt.name, func(t *testing.T) {
				var val int
				err := Unmarshal(tt.arg, &val)
				assert.Equal(t, err != nil, tt.wantErr)
				assert.Equal(t, val, tt.want)
			})
		}
	}
	{
		for _, tt := range []struct {
			name string
			want int64
			arg  []byte
		}{
			{"int64_1", 1, []byte{0x11, 0x01}},
			{"int64_255", 255, []byte{0x11, 0xFF}},
			{"int64_256", 256, []byte{0x12, 0x00, 0x01}},
			{"int64_65535", 65535, []byte{0x12, 0xFF, 0xFF}},
			{"int64_65536", 65536, []byte{0x13, 0x00, 0x00, 0x01}},
			{"int64_16777216", 16_777_216, []byte{0x14, 0x00, 0x00, 0x00, 0x01}},
			{"int64_4294967296", 4_294_967_296, []byte{0x15, 0x00, 0x00, 0x00, 0x00, 0x01}},
			{"int64_-4294967296", -4_294_967_296, []byte{0x1D, 0x00, 0x00, 0x00, 0x00, 0x01}},
		} {
			t.Run(tt.name, func(t *testing.T) {
				var val int64
				err := Unmarshal(tt.arg, &val)
				assert.Equal(t, err != nil, false)
				assert.Equal(t, val, tt.want)
			})
		}
	}
	t.Run("uint8", func(t *testing.T) {
		var val uint8
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, uint8(1))
	})
	t.Run("uint32", func(t *testing.T) {
		var val uint32
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, uint32(1))
	})
	t.Run("uint16", func(t *testing.T) {
		var val uint16
		err := Unmarshal([]byte{0x11, 0x01}, &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, uint16(1))
	})

	t.Run("bool error", func(t *testing.T) {
		var val bool
		err := Unmarshal([]byte{0x22}, &val)
		assert.Equal(t, err != nil, true)
		assert.Equal(t, val, false)
	})
	t.Run("bool false", func(t *testing.T) {
		var val bool
		err := Unmarshal([]byte{0x20}, &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, false)
	})
	t.Run("bool true", func(t *testing.T) {
		var val bool
		err := Unmarshal([]byte{0x21}, &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, true)
	})
	t.Run("string size error", func(t *testing.T) {
		var val string
		err := Unmarshal([]byte("\x32\x03abc"), &val)
		assert.Equal(t, err != nil, true)
		assert.Equal(t, val, "")
	})
	t.Run("string length error", func(t *testing.T) {
		var val string
		err := Unmarshal([]byte("\x31\x02abc"), &val)
		assert.Equal(t, err != nil, true)
		assert.Equal(t, val, "ab")
	})
	t.Run("string", func(t *testing.T) {
		var val string
		err := Unmarshal([]byte("\x31\x03abc"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, "abc")
	})
	t.Run("struct error", func(t *testing.T) {
		type jsonp struct {
			Key string
			B   string `jp:"b"`
		}
		var val jsonp
		err := Unmarshal([]byte("\x51\x02\x31\x03Key\x31\x03val\x31\x01b\x51\x01\x31\x01a\x11\x01"), &val)
		assert.Equal(t, err != nil, true)
		assert.Equal(t, val, jsonp{Key: "val"})
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
		err := Unmarshal([]byte("\x51\x02\x31\x03Key\x31\x03val\x31\x01b\x51\x01\x31\x01a\x11\x01"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, jsonp{"val", subS{1}})
	})
	t.Run("struct can empty", func(t *testing.T) {
		type jsonp struct {
			Key string `jp:",omitempty"`
		}
		var val jsonp
		err := Unmarshal([]byte("\x51\x01\x31\x03key\x31\x03val"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, jsonp{})
	})
	t.Run("struct can not empty", func(t *testing.T) {
		type jsonp struct {
			Key string
		}
		var val jsonp
		err := Unmarshal([]byte("\x51\x01\x31\x03key\x31\x03val"), &val)
		assert.Equal(t, err != nil, true)
		assert.Equal(t, val, jsonp{})
	})
	t.Run("array", func(t *testing.T) {
		var val []string
		err := Unmarshal([]byte("\x41\x02\x31\x03key\x31\x03val"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, []string{"key", "val"})
	})
	t.Run("struct", func(t *testing.T) {
		type reqSubscribe struct {
			Evt int  `jp:"evt"`
			Act bool `jp:"act"`
			Sub int  `jp:"sub,omitempty"`
		}
		var val reqSubscribe
		err := Unmarshal([]byte("\x51\x02\x31\x03evt\x11\x02\x31\x03act\x21"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, reqSubscribe{2, true, 0})
	})
	t.Run("struct null", func(t *testing.T) {
		type reqSubscribe struct {
			Evt int   `jp:"evt"`
			Act *bool `jp:"act"`
			Sub *int  `jp:"sub,omitempty"`
			Cmd *int  `jp:"cmd"`
		}
		var (
			val reqSubscribe
			i   int = 1
		)

		err := Unmarshal([]byte("\x51\x03\x31\x03evt\x11\x02\x31\x03act\x70\x31\x03cmd\x11\x01"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, reqSubscribe{2, nil, nil, &i})
	})
	t.Run("struct []byte", func(t *testing.T) {
		type reqSubscribe struct {
			Evt int    `jp:"evt"`
			Act []byte `jp:"act"`
			Sub []byte `jp:"sub,omitempty"`
			Cmd []byte `jp:"cmd"`
		}
		var (
			val reqSubscribe
		)

		err := Unmarshal([]byte("\x51\x03\x31\x03evt\x11\x02\x31\x03act\x70\x31\x03cmd\x11\x01"), &val)
		assert.Equal(t, err != nil, false)
		assert.Equal(t, val, reqSubscribe{2, []byte{0x70}, nil, []byte{0x11, 0x01}})
	})
	t.Run("struct io.Writer", func(t *testing.T) {
		type req struct {
			Evt  int       `jp:"evt"`
			Act  []byte    `jp:"act"`
			Sub  []byte    `jp:"sub,omitempty"`
			Data io.Writer `jp:"data"`
		}
		// var (
		// 	val req
		// )

		// err := Unmarshal([]byte("\x51\x03\x31\x03evt\x11\x02\x31\x03act\x70\x31\x03cmd\x11\x01"), &val)
		// assert.Equal(t, err != nil, false, err)
		// assert.Equal(t, val, req{2, []byte{0x70}, nil, nil})

	})
}
