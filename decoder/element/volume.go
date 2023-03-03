package element

import (
	"math"

	"github.com/zwcway/castserver-go/decoder"
)

type Volume struct {
	base   float64
	gain   float64
	volume float64
	mute   bool
}

const VolumeName = "Volume"

func (v *Volume) Name() string {
	return VolumeName
}

func (v *Volume) Type() decoder.ElementType {
	return decoder.ET_OneSample
}

func (v *Volume) Stream(samples *decoder.Samples) {
}

func (v *Volume) Sample(sample *float64, ch int, n int) {
	*sample *= v.gain
}

func (v *Volume) SetMute(b bool) {
	v.mute = b
	v.SetVolume(v.volume)
}

func (v *Volume) Mute() bool {
	return v.mute
}

func (v *Volume) SetVolume(p float64) {
	if v.mute {
		v.gain = 0
	} else {
		v.gain = math.Pow(v.base, v.volume)
	}
}

func (v *Volume) Volume() float64 {
	return v.volume
}

func NewVolume(vol float64) *Volume {
	return &Volume{volume: vol}
}
