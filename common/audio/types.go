package audio

type AudioEnum interface {
	FromInt(int32)
	ToInt() int32
	Name() string
	IsValid() bool
}

type AudioMask interface {
	FromSlice([]uint8) error
	Isset(uint8) bool
	IsValid() bool
}

func maskIsset(m uint, a uint8) bool {
	return ((m) & (1 << a)) > 0
}

func maskIssetSlice(m uint, a []uint8) bool {
	for _, v := range a {
		if !maskIsset(m, uint8(v)) {
			return false
		}
	}

	return true
}
func maskCombineSlice(m uint, a []uint8) uint {
	r := uint(0)
	for _, v := range a {
		if maskIsset(m, uint8(v)) {
			r |= 1 << v
		}
	}

	return r
}
