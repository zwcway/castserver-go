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

func FrontLeft() []float64 {
	sound := append(soundsFront, soundsLeft...)
	return resample(sound)
}

func FrontRight() []float64 {
	sound := append(soundsFront, soundsRight...)
	return resample(sound)
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
