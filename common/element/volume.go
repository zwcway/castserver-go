package element

import (
	"math"

	"github.com/zwcway/castserver-go/common/stream"
)

type Volume struct {
	power  bool
	base   float64
	gain   float64
	volume float64
	mute   bool
}

func (v *Volume) Name() string {
	return "Volume"
}

func (v *Volume) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (v *Volume) Stream(samples *stream.Samples) {
	if !v.power || v.gain == 1 {
		return
	}
	for ch := 0; ch < int(samples.Format.Count); ch++ {
		for i := 0; i < samples.LastNbSamples; i++ {
			samples.Data[ch][i] *= v.gain
		}
	}
}

func (v *Volume) Sample(sample *float64, ch int, n int) {
	*sample *= v.gain
}

func (r *Volume) On() {
	r.power = true
}

func (r *Volume) Off() {
	r.power = false
}

func (r *Volume) IsOn() bool {
	return r.power
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

func (v *Volume) Close() error {
	v.volume = 0
	v.Off()
	return nil
}

func NewVolume(vol float64) stream.VolumeElement {
	v := &Volume{volume: vol, base: 1, power: true}

	v.SetVolume(vol)

	return v
}
