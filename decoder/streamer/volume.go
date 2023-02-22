package streamer

import (
	"math"

	"github.com/zwcway/castserver-go/decoder"
)

type volume struct {
	base   float64
	gain   float64
	volume float64
	silent bool
}

func (v *volume) Name() string {
	return "Volume"
}

func (v *volume) Type() decoder.ElementType {
	return decoder.ET_OneSample
}

func (v *volume) Stream(samples *decoder.Samples) {
}

func (v *volume) Sample(sample *float64, ch int, n int) {
	*sample *= v.gain
}

var volumeStreamer = &volume{}

func VolumeStreamer() decoder.Element {
	return volumeStreamer
}

func VolumeSlient(s bool) {
	volumeStreamer.silent = s
	if !s {
		volumeStreamer.gain = math.Pow(volumeStreamer.base, volumeStreamer.volume)
	} else {
		volumeStreamer.gain = 0.0
	}
}

func VolumeSet(vol float64) {
	volumeStreamer.volume = vol

	VolumeSlient(volumeStreamer.silent)
}

func VolumeGet() float64 {
	return volumeStreamer.volume
}
