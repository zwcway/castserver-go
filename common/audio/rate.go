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
	AudioRate_88200
	AudioRate_96000
	AudioRate_192000
	AudioRate_352800
	AudioRate_384000
	AudioRate_2822400
	AudioRate_5644800
	AudioRate_MAX
)

func NewAudioRate(i int) Rate {
	var a Rate
	a.FromInt(i)
	return a
}

func (a *Rate) FromInt(i int) {
	switch i {
	case 44_100:
		*a = AudioRate_44100
	case 48_000:
		*a = AudioRate_48000
	case 88_200:
		*a = AudioRate_88200
	case 96_000:
		*a = AudioRate_96000
	case 192_000:
		*a = AudioRate_192000
	case 352_800:
		*a = AudioRate_352800
	case 384_000:
		*a = AudioRate_384000
	case 2_822_400:
		*a = AudioRate_2822400
	case 5_644_800:
		*a = AudioRate_5644800
	default:
		*a = AudioRate_NONE
	}
}

func (a Rate) ToInt() int {
	switch a {
	case AudioRate_44100:
		return 44_100
	case AudioRate_48000:
		return 48_000
	case AudioRate_88200:
		return 88_200
	case AudioRate_96000:
		return 96_000
	case AudioRate_192000:
		return 192_000
	case AudioRate_352800:
		return 352_800
	case AudioRate_384000:
		return 384_000
	case AudioRate_2822400:
		return 2_822_400
	case AudioRate_5644800:
		return 5_644_800
	default:
		return 0
	}
}

func (a Rate) String() string {
	return fmt.Sprintf("%d", a.ToInt())
}

func (a Rate) LessThan(r Rate) bool {
	return a.ToInt() < r.ToInt()
}

func (a Rate) IsValid() bool {
	return a > AudioRate_NONE && a < AudioRate_MAX
}

func (a Rate) ResampleTo(r Rate, s int) int {
	return a.ToInt() / r.ToInt() * s
}

type RateMask uint16

func NewAudioRateMask(arr []uint8) (RateMask, error) {
	var a RateMask
	err := a.FromSlice(arr)
	return a, err
}

func (m *RateMask) FromSlice(arr []uint8) error {
	if len(arr) > 16 {
		return errors.New("rates too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m RateMask) IssetInt(a uint8) bool {
	return MaskIsset(uint32(m), a)
}

func (m RateMask) Isset(a Rate) bool {
	return MaskIsset(uint32(m), uint8(a))
}

func (m RateMask) IssetSlice(a []uint8) bool {
	return MaskIssetIntSlice(uint32(m), a)
}

func (m *RateMask) IntersectIntSlice(a []uint8) bool {
	n, err := NewAudioRateMask(a)
	if err != nil {
		return false
	}

	*m &= n
	return m.IsValid()
}

// 默认参数是合法的
func (m *RateMask) IntersectSlice(a []Rate) bool {
	*m &= RateMask(MakeMaskFromSlice(a))
	return m.IsValid()
}

func (m *RateMask) Intersect(a RateMask) bool {
	*m &= a
	return m.IsValid()
}

// 默认参数是合法的
func (m *RateMask) CombineSlice(a []Rate) bool {
	*m |= RateMask(MakeMaskFromSlice(a))
	return m.IsValid()
}

func (m *RateMask) Combine(a RateMask) bool {
	*m |= a
	return m.IsValid()
}

func (m RateMask) Max() (max Rate) {
	for _, r := range m.RateSlice() {
		if max.LessThan(r) {
			max = r
		}
	}
	return
}

func (m RateMask) IsValid() bool {
	return m > 0 && ((m)>>(AudioRate_MAX-1)) == 0
}

func (m RateMask) Slice() []int {
	s := make([]int, 16)
	j := 0
	for i := 0; i < 16; i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = Rate(i + 1).ToInt()
			j++
		}
	}
	return s[:j]
}

func (m RateMask) RateSlice() []Rate {
	s := make([]Rate, 16)
	j := 0
	for i := 0; i < 16; i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = Rate(i + 1)
			j++
		}
	}
	return s[:j]
}
