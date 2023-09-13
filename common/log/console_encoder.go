package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const _hex = "0123456789ABCDEF"
const _timeLayout = "2006-01-02 15:04:05"

var bufferpool = buffer.NewPool()

var _Pool = sync.Pool{New: func() interface{} {
	return &consoleEncoder{}
}}

func getEncoder() *consoleEncoder {
	return _Pool.Get().(*consoleEncoder)
}

func putEncoder(c *consoleEncoder) {
	c.EncoderConfig = nil
	c.key = ""
	c.reflectBuf = nil
	c.reflectEnc = nil
	_Pool.Put(c)
}

var _sliceEncoderPool = sync.Pool{
	New: func() interface{} {
		return &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	},
}

func getSliceEncoder() *sliceArrayEncoder {
	return _sliceEncoderPool.Get().(*sliceArrayEncoder)
}

func putSliceEncoder(e *sliceArrayEncoder) {
	e.elems = e.elems[:0]
	_sliceEncoderPool.Put(e)
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(_timeLayout))
}

func nameEncoder(v string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(YellowBold)
	enc.AppendString(v)
	enc.AppendString(Reset)
}

func levelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var v string
	switch level {
	case zapcore.InfoLevel:
		v = Green
	case zapcore.WarnLevel:
		v = Magenta
	case zapcore.DebugLevel:
		v = Blue
	case zapcore.ErrorLevel:
		v = RedBold
	}
	v += level.CapitalString() + Reset

	enc.AppendString(v)
}

type consoleEncoder struct {
	*zapcore.EncoderConfig
	key        string
	buf        *buffer.Buffer
	reflectBuf *buffer.Buffer
	reflectEnc *json.Encoder
}

func NewConsoleEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	if cfg.ConsoleSeparator == "" {
		cfg.ConsoleSeparator = " "
	}

	return &consoleEncoder{
		EncoderConfig: &cfg,
		buf:           bufferpool.Get(),
		reflectEnc:    json.NewEncoder(bufferpool.Get()),
	}
}

func (c *consoleEncoder) Clone() zapcore.Encoder {
	return c.clone()
}

func (c *consoleEncoder) clone() *consoleEncoder {
	clone := getEncoder()
	clone.EncoderConfig = c.EncoderConfig
	clone.buf = bufferpool.Get()
	// clone.buf.Write(c.buf.Bytes())
	return clone
}

func (c *consoleEncoder) EncodeEntry(ent zapcore.Entry, fields []zap.Field) (*buffer.Buffer, error) {
	line := bufferpool.Get()

	arr := getSliceEncoder()

	if c.TimeKey != "" && c.EncodeTime != nil {
		c.EncodeTime(ent.Time, arr)
	}
	if c.LevelKey != "" && c.EncodeLevel != nil {
		c.EncodeLevel(ent.Level, arr)
	}
	if ent.LoggerName != "" && c.NameKey != "" {
		nameEncoder := c.EncodeName

		if nameEncoder == nil {
			// Fall back to FullNameEncoder for backward compatibility.
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, arr)
	}
	if ent.Caller.Defined {
		if c.CallerKey != "" && c.EncodeCaller != nil {
			c.EncodeCaller(ent.Caller, arr)
		}
		if c.FunctionKey != "" {
			arr.AppendString(ent.Caller.Function)
		}
	}

	for i := range arr.elems {
		if i > 0 {
			line.AppendString(c.ConsoleSeparator)
		}
		fmt.Fprint(line, arr.elems[i])
	}
	putSliceEncoder(arr)

	if line.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}

	line.Write(c.buf.Bytes())

	if c.buf.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}

	c.writeContext(line, fields)

	if c.MessageKey != "" {
		line.AppendString(Yellow)
		line.AppendString(ent.Message)
		line.AppendString(Reset)
		line.AppendString(c.ConsoleSeparator)
	}

	if ent.Stack != "" && c.StacktraceKey != "" {
		line.AppendByte('\n')
		line.AppendString(ent.Stack)
	}

	line.AppendString(c.LineEnding)

	return line, nil
}

func (c consoleEncoder) writeContext(line *buffer.Buffer, extra []zap.Field) {
	context := c.clone()
	defer func() {
		context.buf.Free()
		putEncoder(context)
	}()

	for i := range extra {
		extra[i].AddTo(context)
		context.buf.AppendString(c.ConsoleSeparator)
	}

	context.CloseNamespace()
	if context.buf.Len() == 0 {
		return
	}
	context.buf.AppendString(c.ConsoleSeparator)

	line.Write(context.buf.Bytes())
}

func (c *consoleEncoder) addKey(key string) {
	c.buf.AppendString(Cyan)
	if len(c.key) == 0 {
		c.buf.AppendString(key)
	} else {
		c.buf.AppendString(c.key + "." + key)
	}
	c.buf.AppendString(Reset)
	c.buf.AppendByte('=')
}

func (c *consoleEncoder) AddArray(k string, v zapcore.ArrayMarshaler) error {
	c.addKey(k)
	return c.AppendArray(v)
}

func (c *consoleEncoder) AddObject(k string, v zapcore.ObjectMarshaler) error {
	c.addKey(k)
	return c.AppendObject(v)
}

func (c *consoleEncoder) AddBinary(k string, v []byte) {
	c.addKey(k)
	c.buf.AppendString(Blue)
	c.addBinary(v)
	c.buf.AppendString(Reset)
}

func (c *consoleEncoder) addBinary(v []byte) {
	c.buf.AppendByte('[')
	for i, v := range v {
		if i > 0 {
			c.buf.AppendByte(' ')
		}
		c.buf.AppendByte(_hex[v>>4])
		c.buf.AppendByte(_hex[v&0x0f])
	}
	c.buf.AppendByte(']')
}

func (c *consoleEncoder) AddByteString(k string, v []byte) {
	c.addKey(k)
	c.AppendByteString(v)
}

func (c *consoleEncoder) AddBool(k string, v bool) {
	c.addKey(k)
	c.AppendBool(v)
}

func (c *consoleEncoder) AddDuration(k string, v time.Duration) {
	c.addKey(k)
	c.AppendDuration(v)
}

func (c *consoleEncoder) AddComplex128(k string, v complex128) {
	c.addKey(k)
	c.appendComplex((v), 64)
}

func (c *consoleEncoder) AddComplex64(k string, v complex64) {
	c.addKey(k)
	c.appendComplex(complex128(v), 32)
}

func (c *consoleEncoder) AddFloat64(k string, v float64) {
	c.addKey(k)
	c.buf.AppendFloat(float64(v), 32)
}

func (c *consoleEncoder) AddFloat32(k string, v float32) {
	c.addKey(k)
	c.buf.AppendFloat(float64(v), 32)
}

func (c *consoleEncoder) AddInt64(k string, v int64) {
	c.addKey(k)
	c.buf.AppendString(Blue)
	c.AppendInt64(v)
	c.buf.AppendString(Reset)
}

func (c *consoleEncoder) AddUint64(k string, v uint64) {
	c.addKey(k)
	c.buf.AppendUint(v)
	c.buf.AppendString(c.ConsoleSeparator)
}

func (c *consoleEncoder) AddString(k string, v string) {
	c.addKey(k)
	c.AppendString(v)
}

func (c *consoleEncoder) AddTime(k string, v time.Time) {
	c.addKey(k)
	c.AppendTime(v)
}

func (c *consoleEncoder) AddInt(k string, v int)         { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddInt32(k string, v int32)     { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddInt16(k string, v int16)     { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddInt8(k string, v int8)       { c.AddInt64(k, int64(v)) }
func (c *consoleEncoder) AddUint(k string, v uint)       { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUint32(k string, v uint32)   { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUint16(k string, v uint16)   { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUint8(k string, v uint8)     { c.AddUint64(k, uint64(v)) }
func (c *consoleEncoder) AddUintptr(k string, v uintptr) { c.AddUint64(k, uint64(v)) }

func (c *consoleEncoder) AddReflected(k string, v interface{}) error {
	c.addKey(k)
	c.AppendReflected(v)
	return nil
}

func (c *consoleEncoder) resetReflectBuf() {
	if c.reflectBuf == nil {
		c.reflectBuf = bufferpool.Get()
		c.reflectEnc = json.NewEncoder(c.reflectBuf)
		c.reflectEnc.SetEscapeHTML(false)
	} else {
		c.reflectBuf.Reset()
	}
}

func (c *consoleEncoder) OpenNamespace(k string) {
	if len(c.key) > 0 {
		c.key += "." + k
	} else {
		c.key = k
	}
}

func (c *consoleEncoder) CloseNamespace() {
	index := strings.LastIndexByte(c.key, '.')
	if index > 0 {
		c.key = c.key[:index]
	}
}

func (c *consoleEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	err := arr.MarshalLogArray(c)
	return err
}

func (c *consoleEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	err := obj.MarshalLogObject(c)
	c.CloseNamespace()
	return err
}

func (c *consoleEncoder) AppendBool(val bool) {
	c.buf.AppendBool(val)
}

func (c *consoleEncoder) AppendByteString(val []byte) {
	c.buf.AppendByte('"')
	c.safeAddByteString(val)
	c.buf.AppendByte('"')
}

func (c *consoleEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if c.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if c.tryAddRuneError(r, size) {
			i++
			continue
		}
		c.buf.AppendString(s[i : i+size])
		i += size
	}
}
func (c *consoleEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if c.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if c.tryAddRuneError(r, size) {
			i++
			continue
		}
		c.buf.Write(s[i : i+size])
		i += size
	}
}

func (c *consoleEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		c.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		c.buf.AppendByte('\\')
		c.buf.AppendByte(b)
	case '\n':
		c.buf.AppendByte('\\')
		c.buf.AppendByte('n')
	case '\r':
		c.buf.AppendByte('\\')
		c.buf.AppendByte('r')
	case '\t':
		c.buf.AppendByte('\\')
		c.buf.AppendByte('t')
	default:
		c.buf.AppendString(`\x`)
		c.buf.AppendByte(_hex[b>>4])
		c.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (c *consoleEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		c.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func (c *consoleEncoder) appendComplex(val complex128, precision int) {
	r, i := float64(real(val)), float64(imag(val))
	c.buf.AppendFloat(r, precision)
	if i >= 0 {
		c.buf.AppendByte('+')
	}
	c.buf.AppendFloat(i, precision)
	c.buf.AppendByte('i')
}

func (c *consoleEncoder) AppendDuration(val time.Duration) {
	cur := c.buf.Len()
	if e := c.EncodeDuration; e != nil {
		e(val, c)
	}
	if cur == c.buf.Len() {
		c.AppendInt64(int64(val))
	}
}

func (c *consoleEncoder) AppendInt64(val int64) {
	c.buf.AppendInt(val)
}

func (c *consoleEncoder) AppendReflected(val interface{}) error {
	c.resetReflectBuf()
	if err := c.reflectEnc.Encode(val); err != nil {
		return err
	}
	c.buf.Write(c.reflectBuf.Bytes())

	return nil
}

func (c *consoleEncoder) AppendString(val string) {
	c.buf.AppendByte('"')
	c.buf.AppendString(Green)
	c.safeAddString(val)
	c.buf.AppendString(Reset)
	c.buf.AppendByte('"')
}

func (c *consoleEncoder) AppendTimeLayout(time time.Time, layout string) {
	c.buf.AppendTime(time, layout)
}

func (c *consoleEncoder) AppendTime(val time.Time) {
	cur := c.buf.Len()
	if e := c.EncodeTime; e != nil {
		e(val, c)
	}
	if cur == c.buf.Len() {
		c.AppendInt64(val.UnixNano())
	}
}

func (c *consoleEncoder) AppendUint64(val uint64) {
	c.buf.AppendUint(val)
}
func (c *consoleEncoder) AppendComplex64(v complex64)   { c.appendComplex(complex128(v), 32) }
func (c *consoleEncoder) AppendComplex128(v complex128) { c.appendComplex(complex128(v), 64) }
func (c *consoleEncoder) AppendFloat64(v float64)       { c.buf.AppendFloat(v, 64) }
func (c *consoleEncoder) AppendFloat32(v float32)       { c.buf.AppendFloat(float64(v), 32) }
func (c *consoleEncoder) AppendInt(v int)               { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendInt32(v int32)           { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendInt16(v int16)           { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendInt8(v int8)             { c.AppendInt64(int64(v)) }
func (c *consoleEncoder) AppendUint(v uint)             { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUint32(v uint32)         { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUint16(v uint16)         { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUint8(v uint8)           { c.AppendUint64(uint64(v)) }
func (c *consoleEncoder) AppendUintptr(v uintptr)       { c.AppendUint64(uint64(v)) }

// sliceArrayEncoder is an ArrayEncoder backed by a simple []interface{}. Like
// the MapObjectEncoder, it's not designed for production use.
type sliceArrayEncoder struct {
	elems []interface{}
}

func (s *sliceArrayEncoder) AppendArray(v zapcore.ArrayMarshaler) error {
	enc := &sliceArrayEncoder{}
	err := v.MarshalLogArray(enc)
	s.elems = append(s.elems, enc.elems)
	return err
}

func (s *sliceArrayEncoder) AppendObject(v zapcore.ObjectMarshaler) error {
	m := zapcore.NewMapObjectEncoder()
	err := v.MarshalLogObject(m)
	s.elems = append(s.elems, m.Fields)
	return err
}

func (s *sliceArrayEncoder) AppendReflected(v interface{}) error {
	s.elems = append(s.elems, v)
	return nil
}

func (s *sliceArrayEncoder) AppendTime(v time.Time)         { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendBool(v bool)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendByteString(v []byte)      { s.elems = append(s.elems, string(v)) }
func (s *sliceArrayEncoder) AppendComplex128(v complex128)  { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendComplex64(v complex64)    { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendDuration(v time.Duration) { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendFloat64(v float64)        { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendFloat32(v float32)        { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt(v int)                { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt64(v int64)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt32(v int32)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt16(v int16)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt8(v int8)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendString(v string)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint(v uint)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint64(v uint64)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint32(v uint32)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint16(v uint16)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint8(v uint8)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUintptr(v uintptr)        { s.elems = append(s.elems, v) }
