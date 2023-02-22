package streamer

import (
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/decoder"
)

type resample struct {
	format *audio.Format
}

func (r *resample) Name() string {
	return "Resampler"
}

func (r *resample) Type() decoder.ElementType {
	return decoder.ET_WholeSamples
}

func (r *resample) Stream(samples *decoder.Samples) {
	switch r.format.SampleBits.Size() {
	case 2:
		r.to16(samples)
	default:
		return
	}

	samples.Format = r.format
}

func (r *resample) Sample(*float64, int, int) {}

func (r *resample) to16(samples *decoder.Samples) {
	for i := 0; i < samples.Size; i++ {
		for c := 0; c < samples.Format.Layout.Count && c < r.format.Layout.Count; c++ {
			val := samples.Buffer[c][i]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			samples.Buffer[c][i] = (*(*float64)(unsafe.Pointer(&valInt16)))
		}
	}
}

var resampleStreamer = &resample{}

func ResampleStreamer() decoder.Element {
	return resampleStreamer
}

func ResampleSet(fmt *audio.Format) decoder.Element {
	resampleStreamer.format = fmt
	return resampleStreamer
}

func ResampleGet() *audio.Format {
	return resampleStreamer.format
}
