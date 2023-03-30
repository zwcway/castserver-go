package audio

import "unsafe"

type AudioEnum interface {
	FromInt(int)
	ToInt() int
	String() string
	IsValid() bool
}

type AudioMask interface {
	FromSlice([]uint8) error
	IssetInt(uint8) bool
	IsValid() bool
}

func MakeMaskFromSlice[E Rate | Bits | Channel](a []E) (m uint32) {
	for _, aa := range a {
		m |= 1 << (aa - 1)
	}
	return
}

func MakeMaskFromIntSlice(a []uint8) (m uint32) {
	for _, aa := range a {
		m |= 1 << (aa - 1)
	}
	return
}

func MaskCount32(m uint32) int {
	m -= (m >> 1) & 0x55555555
	m = (m & 0x33333333) + ((m >> 2) & 0x33333333)
	m = (m + (m >> 4)) & 0x0F0F0F0F
	m += m >> 8
	return int((m + (m >> 16)) & 0x3F)
}

func MaskIsset(m uint32, a uint8) bool {
	if a == 0 || a > 31 {
		return false
	}
	return ((m) & (1 << (a - 1))) > 0
}

func MaskIssetIntSlice(m uint32, a []uint8) bool {
	r := MakeMaskFromIntSlice(a)
	return (m & r) > 0
}

func MaskToSlice[E RateMask | BitsMask | ChannelMask](m E) []E {
	i := int(unsafe.Sizeof(m))*8 - 1
	s := make([]E, i)
	j := 0
	for ; i >= 0; i-- {
		if (m>>i)&0x01 == 1 {
			s[j] = E(i + 1)
			j++
		}
	}

	return s[:j]
}

func MaskIntersectIntSlice[E RateMask | BitsMask | ChannelMask](m E, a []uint8) E {
	r := MakeMaskFromIntSlice(a)
	return E(uint32(m) & r)
}

func MaskIntersect[E RateMask | BitsMask | ChannelMask](m E, a E) E {
	return m & a
}

func SliceToIntSlice[E Rate | Channel | Bits](s []E) []uint8 {
	ret := make([]uint8, len(s))
	for i, ss := range s {
		ret[i] = uint8(ss)
	}
	return ret
}
