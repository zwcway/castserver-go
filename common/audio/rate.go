package audio

import (
	"errors"
	"fmt"
)

type Rate uint8 // uint4

const (
	AudioRate_NONE Rate = iota
	AudioRate_44100
	AudioRate_48000
	AudioRate_96000
	AudioRate_192000
	AudioRate_384000
	AudioRate_MAX
)

func NewAudioRate(i int) Rate {
	var a Rate
	a.FromInt(i)
	return a
}

func (a *Rate) FromInt(i int) {
	switch i {
	case 44100:
		*a = AudioRate_44100
	case 48000:
		*a = AudioRate_48000
	case 96000:
		*a = AudioRate_96000
	case 192000:
		*a = AudioRate_192000
	case 384000:
		*a = AudioRate_384000
	default:
		*a = AudioRate_NONE
	}
}

func (a Rate) ToInt() int {
	switch a {
	case AudioRate_44100:
		return 44100
	case AudioRate_48000:
		return 48000
	case AudioRate_96000:
		return 96000
	case AudioRate_192000:
		return 192000
	case AudioRate_384000:
		return 384000
	default:
		return 0
	}
}

func (a Rate) String() string {
	return fmt.Sprintf("%d", a.ToInt())
}

func (a Rate) IsValid() bool {
	return a > AudioRate_NONE && a < AudioRate_MAX
}

func (a Rate) ResampleTo(r Rate, s int) int {
	return a.ToInt() / r.ToInt() * s
}

type AudioRateMask uint16

func NewAudioRateMask(arr []uint8) (AudioRateMask, error) {
	var a AudioRateMask
	err := a.FromSlice(arr)
	return a, err
}

func (m *AudioRateMask) FromSlice(arr []uint8) error {
	if len(arr) > 16 {
		return errors.New("rates too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m AudioRateMask) Isset(a uint8) bool {
	return maskIsset(uint(m), a)
}

func (m AudioRateMask) IssetSlice(a []uint8) bool {
	return maskIssetSlice(uint(m), a)
}

func (m *AudioRateMask) CombineSlice(a []uint8) bool {
	r := maskCombineSlice(uint(*m), a)
	*m = AudioRateMask(r)
	return r > 0
}

func (m *AudioRateMask) Combine(a []Rate) bool {
	r := maskCombineSlice(uint(*m), toSlice(a))
	*m = AudioRateMask(r)
	return r > 0
}

func (m AudioRateMask) IsValid() bool {
	return m > 0 && ((m)>>(AudioRate_MAX-1)) == 0
}

func (m AudioRateMask) Slice() []int {
	s := []int{}
	for i := 0; i < 16; i++ {
		if (m>>i)&0x01 == 1 {
			b := Rate(i + 1)
			s = append(s, b.ToInt())
		}
	}
	return s
}
