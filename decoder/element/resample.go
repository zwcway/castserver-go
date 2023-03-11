package element

import (
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
)

type Resample struct {
	power  bool
	format *audio.Format
}

const ResampleName = "Resampler"

func (r *Resample) Name() string {
	return ResampleName
}

func (r *Resample) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (r *Resample) Stream(samples *stream.Samples) {
	if !r.power || r.format == nil {
		return
	}

	switch r.format.SampleBits {
	case audio.AudioBits_S16LE:
		r.toInt16(samples)
	default:
		return
	}

	samples.Format = r.format
}

func (r *Resample) Sample(*float64, int, int) {}

func (r *Resample) toInt16(samples *stream.Samples) {
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

func (r *Resample) On() {
	r.power = true
}

func (r *Resample) Off() {
	r.power = false
}

func (r *Resample) IsOn() bool {
	return r.power
}

func (r *Resample) SetFormat(format *audio.Format) {
	if format == nil {
		return
	}
	*r.format = *format
}

func (r *Resample) Format() *audio.Format {
	return r.format
}

func NewResample(format *audio.Format) stream.ResampleElement {
	if format == nil {
		format = &audio.Format{}
	}
	return &Resample{format: format}
}
