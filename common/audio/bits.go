package audio

import (
	"errors"
	"strings"
)

type Bits uint8 // uint4

const (
	Bits_NONE Bits = iota

	Bits_S8    // int8
	Bits_S16LE // int16
	Bits_S24LE // int24
	Bits_S32LE // int32
	Bits_U8    // uint8
	Bits_U16LE // uint16
	Bits_U24LE // uint24
	Bits_U32LE // uint32
	Bits_32LEF // float32
	Bits_64LEF // float64

	Bits_16LEF // float16
	Bits_24LEF // float24

	Bits_MAX
)

const Bits_DEFAULT = Bits_64LEF

func NewAudioBits(i int) Bits {
	var a Bits
	a.FromInt(i)
	return a
}

func (a *Bits) FromInt(i int) {
	switch i {
	case 8:
		*a = Bits_U8
	case 16:
		*a = Bits_S16LE
	case 24:
		*a = Bits_S24LE
	case 32:
		*a = Bits_32LEF
	case 64:
		*a = Bits_64LEF
	default:
		*a = Bits_NONE
	}
}

func (a *Bits) FromBytes(i int8) {
	a.FromInt(int(i * 8))
}

func (a Bits) ToInt() int32 {
	switch a {
	case Bits_S8, Bits_U8:
		return 8
	case Bits_S16LE, Bits_U16LE:
		return 16
	case Bits_S24LE, Bits_U24LE:
		return 24
	case Bits_S32LE, Bits_U32LE:
		return 32
	case Bits_32LEF:
		return 32
	case Bits_16LEF:
		return 16
	case Bits_24LEF:
		return 24
	case Bits_64LEF:
		return 64
	default:
		return 0
	}
}

func (a *Bits) FromName(n string) {
	switch strings.ToLower(n) {
	case "u8", "8":
		*a = Bits_U8
	case "s8":
		*a = Bits_S8
	case "s16le", "s16", "16":
		*a = Bits_S16LE
	case "u16le", "u16":
		*a = Bits_U16LE
	case "f16le", "f16", "16f":
		*a = Bits_16LEF
	case "s24le", "s24", "24":
		*a = Bits_S24LE
	case "u24le", "u24":
		*a = Bits_U24LE
	case "f24le", "f24", "24f":
		*a = Bits_24LEF
	case "s32le", "s32", "32":
		*a = Bits_S32LE
	case "u32le", "u32":
		*a = Bits_U32LE
	case "f32le", "f32", "32f", "flt", "fltle", "float", "floatle":
		*a = Bits_32LEF
	case "f64le", "f64", "64f":
		*a = Bits_64LEF
	default:
		*a = Bits_NONE
	}
}

func (a Bits) String() string {
	switch a {
	case Bits_U8:
		return "u8"
	case Bits_S8:
		return "s8"
	case Bits_S16LE:
		return "s16le"
	case Bits_U16LE:
		return "u16le"
	case Bits_16LEF:
		return "f16le"
	case Bits_S24LE:
		return "s24le"
	case Bits_U24LE:
		return "u24le"
	case Bits_24LEF:
		return "f24le"
	case Bits_S32LE:
		return "s32le"
	case Bits_U32LE:
		return "u32le"
	case Bits_32LEF:
		return "f32le"
	case Bits_64LEF:
		return "f64le"
	default:
		return "Unknown"
	}
}

func (a Bits) Bits() int32 {
	return a.ToInt()
}
func (a Bits) Size() int {
	return int(a.ToInt() / 8)
}

func (a Bits) IsValid() bool {
	return a > Bits_NONE && a < Bits_MAX
}

type BitsMask uint16

func NewAudioBitsMask(arr []uint8) (BitsMask, error) {
	var a BitsMask
	err := a.FromSlice(arr)
	return a, err
}

func (m *BitsMask) FromSlice(arr []uint8) error {
	if len(arr) > 16 {
		return errors.New("bits too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m *BitsMask) Isset(a uint8) bool {
	return maskIsset(uint(*m), a)
}

func (m *BitsMask) IssetSlice(a []uint8) bool {
	return maskIssetSlice(uint(*m), a)
}

func (m *BitsMask) CombineSlice(a []uint8) bool {
	r := maskCombineSlice(uint(*m), a)
	*m = BitsMask(r)
	return r > 0
}

func (m *BitsMask) Combine(a []Bits) bool {
	r := maskCombineSlice(uint(*m), toSlice(a))
	*m = BitsMask(r)
	return r > 0
}

func (m *BitsMask) IsValid() bool {
	return *m > 0 && ((*m)>>(Bits_MAX-1)) == 0
}
func (m *BitsMask) Slice() []string {
	s := []string{}
	for i := 0; i < 16; i++ {
		if (*m>>i)&0x01 == 1 {
			b := Bits(i + 1)
			s = append(s, b.String())
		}
	}
	return s
}
