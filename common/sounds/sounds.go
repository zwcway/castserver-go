package sounds

import (
	_ "embed"

	"github.com/zwcway/castserver-go/common/audio"
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

func FrontLeft() []byte {
	return appends(soundsFront, soundsLeft)
}

func FrontRight() []byte {
	return appends(soundsFront, soundsRight)
}

func FrontCenter() []byte {
	return appends(soundsFront, soundsCenter)
}

func LowBass() []byte {
	return soundsLowBass
}

func BackLeft() []byte {
	return appends(soundsBack, soundsSurround, soundsLeft)
}

func BackRight() []byte {
	return appends(soundsBack, soundsSurround, soundsRight)
}

func BackCenter() []byte {
	return appends(soundsBack, soundsCenter)
}

func SideLeft() []byte {
	return appends(soundsSurround, soundsLeft)
}

func SideRight() []byte {
	return appends(soundsSurround, soundsRight)
}

func Here() []byte {
	return soundsHere
}

func appends(a ...[]byte) []byte {
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

	return ret
}

func Format() audio.Format {
	return audio.Format{
		SampleRate: audio.AudioRate_44100,
		Layout:     audio.ChannelLayout10,
		SampleBits: audio.Bits_S16LE,
	}
}
