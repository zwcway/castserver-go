package sounds

import (
	_ "embed"
	"math"
)

/*
 s16le
*/

//go:embed front.pcm
var soundsFront []byte

//go:embed left.pcm
var soundsLeft []byte

//go:embed right.pcm
var soundsRight []byte

//go:embed center.pcm
var soundsCenter []byte

//go:embed back.pcm
var soundsBack []byte

//go:embed surround.pcm
var soundsSurround []byte

//go:embed lowbass.pcm
var soundsLowBass []byte

//go:embed iamhere.pcm
var soundsHere []byte

func FrontLeft() []float64 {
	return appends(soundsFront, soundsLeft)
}

func FrontRight() []float64 {
	return appends(soundsFront, soundsRight)
}

func FrontCenter() []float64 {
	return appends(soundsFront, soundsCenter)
}

func LowBass() []float64 {
	return resample(soundsLowBass)
}

func BackLeft() []float64 {
	return appends(soundsBack, soundsSurround, soundsLeft)
}

func BackRight() []float64 {
	return appends(soundsBack, soundsSurround, soundsRight)
}

func BackCenter() []float64 {
	return appends(soundsBack, soundsCenter)
}

func SideLeft() []float64 {
	return appends(soundsSurround, soundsLeft)
}

func SideRight() []float64 {
	return appends(soundsSurround, soundsRight)
}

func Here() []float64 {
	return resample(soundsHere)
}

func appends(a ...[]byte) []float64 {
	l := 0
	for _, s := range a {
		l += len(s)
	}

	ret := make([]byte, l)

	l = 0
	for i := 0; i < len(a); i++ {
		for b := 0; b < len(a[i]); b++ {
			ret[l] = a[i][b]
			l++
		}
	}

	return resample(ret)
}

func resample(s []byte) []float64 {
	sounds := make([]float64, len(s)/2)

	j := 0
	for i := 0; i < len(s); i += 2 {
		varInt16 := int16(s[i]) | (int16(s[i+1]) << 8)

		sounds[j] = float64(varInt16) / (math.Exp2(15) - 1)
		j++
	}
	return sounds
}
