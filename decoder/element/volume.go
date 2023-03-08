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
	return decoder.ET_WholeSamples
}

func (v *Volume) Stream(samples *decoder.Samples) {
	for ch := 0; ch < samples.Format.Layout.Count; ch++ {
		for i := 0; i < samples.Size; i++ {
			samples.Buffer[ch][i] *= v.gain
		}
	}
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
	v.volume = p

	if v.mute || v.volume == 0 {
		v.gain = 0
	} else if v.base == 1 {
		v.gain = v.volume
	} else {
		v.gain = math.Pow(v.base, v.volume)
	}
}

func (v *Volume) Volume() float64 {
	return v.volume
}

func NewVolume(vol float64) *Volume {
	return &Volume{volume: vol, base: 1}
}
