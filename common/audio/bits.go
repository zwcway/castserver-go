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

	Bits_16LEF // float16
	Bits_24LEF // float24
	Bits_32LEF // float32
	Bits_64LEF // float64

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
		*a = Bits_S32LE
	case 64:
		*a = Bits_64LEF
	default:
		*a = Bits_NONE
	}
}

func (a *Bits) FromBytes(i int8) {
	a.FromInt(int(i * 8))
}

func (a Bits) ToInt() int {
	switch a {
	case Bits_S8, Bits_U8:
		return 8
	case Bits_S16LE, Bits_U16LE, Bits_16LEF:
		return 16
	case Bits_S24LE, Bits_U24LE, Bits_24LEF:
		return 24
	case Bits_S32LE, Bits_U32LE, Bits_32LEF:
		return 32
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

func (a Bits) IsFloat() bool {
	return a == Bits_16LEF || a == Bits_24LEF || a == Bits_32LEF || a == Bits_64LEF
}

func (a Bits) Bits() int {
	return a.ToInt()
}

func (a Bits) LessThan(r Bits) bool {
	return a.ToInt() < r.ToInt()
}

func (a Bits) Size() int {
	return a.ToInt() / 8
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

func NewAudioBitsMaskFromNames(arr []string) (BitsMask, error) {
	var (
		a   BitsMask
		bit Bits
	)
	barr := make([]uint8, 0)
	for _, b := range arr {
		bit.FromName(b)
		if !bit.IsValid() {
			continue
		}
		barr = append(barr, uint8(bit))
	}
	err := a.FromSlice(barr)
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

func (m BitsMask) IssetInt(a uint8) bool {
	return MaskIsset(uint32(m), a)
}

func (m BitsMask) Isset(a Bits) bool {
	return MaskIsset(uint32(m), uint8(a))
}

func (m BitsMask) IssetSlice(a []uint8) bool {
	return MaskIssetIntSlice(uint32(m), a)
}

func (m *BitsMask) CombineIntSlice(a []uint8) bool {
	r, err := NewAudioBitsMask(a)
	if err != nil {
		return false
	}
	*m &= r
	return m.IsValid()
}

// 默认参数是合法的
func (m *BitsMask) IntersectSlice(a []Bits) bool {
	*m &= BitsMask(MakeMaskFromSlice(a))
	return m.IsValid()
}

func (m *BitsMask) Intersect(a BitsMask) bool {
	*m &= a
	return m.IsValid()
}

// 默认参数是合法的
func (m *BitsMask) CombineSlice(a []Bits) bool {
	*m |= BitsMask(MakeMaskFromSlice(a))
	return m.IsValid()
}

func (m *BitsMask) Combine(a BitsMask) bool {
	*m |= a
	return m.IsValid()
}

func (m BitsMask) IsValid() bool {
	return m > 0 && ((m)>>(Bits_MAX-1)) == 0
}

func (m BitsMask) Max() (max Bits) {
	for _, r := range m.BitsSlice() {
		if max.LessThan(r) {
			max = r
		}
	}

	return
}

func (m BitsMask) StringSlice() []string {
	s := make([]string, 16)
	j := 0
	for i := 0; i < 16; i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = Bits(i + 1).String()
			j++
		}
	}
	return s[:j]
}

func (m BitsMask) Slice() []int {
	s := make([]int, 16)
	j := 0
	for i := 0; i < 16; i++ {
		if (m>>i)&0x01 == 1 {
			b := Bits(i + 1)
			s[j] = int(b.ToInt())
			j++
		}
	}
	return s[:j]
}

func (m BitsMask) BitsSlice() []Bits {
	s := make([]Bits, 16)
	j := 0
	for i := 0; i < 16; i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = Bits(i + 1)
			j++
		}
	}
	return s[:j]
}
