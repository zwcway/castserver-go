package audio

import "errors"

type Bits uint8 // uint4

const (
	AudioBits_NONE Bits = iota

	AudioBits_U8    // uint8
	AudioBits_S16LE // int16
	AudioBits_S24LE // int24
	AudioBits_S32LE // int32
	AudioBits_32LEF // float32
	AudioBits_64LEF // float64

	AudioBits_16LEF // float16
	AudioBits_24LEF // float24

	AudioBits_MAX
)

func NewAudioBits(i int32) Bits {
	var a Bits
	a.FromInt(i)
	return a
}

func (a *Bits) FromInt(i int32) {
	switch i {
	case 8:
		*a = AudioBits_U8
	case 16:
		*a = AudioBits_S16LE
	case 24:
		*a = AudioBits_S24LE
	case 32:
		*a = AudioBits_S32LE
	default:
		*a = AudioBits_NONE
	}
}

func (Bits) FromIntSlice(i []int32) []Bits {
	var (
		s []Bits
		a Bits
	)
	for _, v := range i {
		a.FromInt(v)
		if !a.IsValid() {
			continue
		}
		s = append(s, a)
	}

	return s
}

func (a *Bits) FromBytes(i int8) {
	a.FromInt(int32(i * 8))
}

func (a *Bits) ToInt() int32 {
	switch *a {
	case AudioBits_U8:
		return 8
	case AudioBits_S16LE:
		return 16
	case AudioBits_S24LE:
		return 24
	case AudioBits_S32LE:
		return 32
	case AudioBits_32LEF:
		return 32
	case AudioBits_16LEF:
		return 16
	case AudioBits_24LEF:
		return 24
	case AudioBits_64LEF:
		return 64
	default:
		return 0
	}
}

func (a Bits) FromName(n string) Bits {
	switch n {
	case "u8", "8":
		return AudioBits_U8
	case "s16le", "s16", "16":
		return AudioBits_S16LE
	case "f16le", "f16", "16f":
		return AudioBits_16LEF
	case "s24le", "s24", "24":
		return AudioBits_S24LE
	case "f24le", "f24", "24f":
		return AudioBits_24LEF
	case "s32le", "s32", "32":
		return AudioBits_S32LE
	case "f32le", "f32", "32f":
		return AudioBits_32LEF
	default:
		return AudioBits_NONE
	}
}

func (a *Bits) Name() string {
	switch *a {
	case AudioBits_U8:
		return "u8"
	case AudioBits_S16LE:
		return "s16le"
	case AudioBits_16LEF:
		return "f16le"
	case AudioBits_S24LE:
		return "s24le"
	case AudioBits_24LEF:
		return "f24le"
	case AudioBits_S32LE:
		return "s32le"
	case AudioBits_32LEF:
		return "f32le"
	default:
		return "Unknown"
	}
}

func (a *Bits) Bits() int32 {
	return a.ToInt()
}
func (a *Bits) Size() int {
	return int(a.ToInt() / 8)
}

func (a *Bits) IsValid() bool {
	return *a > AudioBits_NONE && *a < AudioBits_MAX
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

func (m *BitsMask) IsValid() bool {
	return *m > 0 && ((*m)>>(AudioBits_MAX-1)) == 0
}
func (m *BitsMask) Slice() []string {
	s := []string{}
	for i := 0; i < 16; i++ {
		if (*m>>i)&0x01 == 1 {
			b := Bits(i + 1)
			s = append(s, b.Name())
		}
	}
	return s
}
