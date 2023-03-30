package decoder

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg/resample"
)

type Resample struct {
	power  bool
	format audio.Format

	swrCtx *resample.Resample
}

const ResampleName = "Resampler"

func (r *Resample) Name() string {
	return ResampleName
}

func (r *Resample) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (r *Resample) Stream(samples *stream.Samples) {
	if !r.power || !r.format.IsValid() {
		return
	}
	if samples.Format.Equal(r.format) {
		return
	}

	r.swrCtx.SetIn(samples.Format)

	if r.swrCtx == nil || !r.swrCtx.Inited() {
		return
	}
	r.swrCtx.SetOut(r.format)

	r.swrCtx.Stream(samples)
}

func (r *Resample) Sample(*float64, int, int) {}

func (r *Resample) On() {
	r.power = true
}

func (r *Resample) Off() {
	r.power = false
}

func (r *Resample) IsOn() bool {
	return r.power
}

func (r *Resample) SetFormat(format audio.Format) {
	if !format.IsValid() {
		return
	}
	r.format = format
	r.swrCtx.SetOut(r.format)
}

func (r *Resample) Format() audio.Format {
	return r.format
}

func (r *Resample) Close() error {
	r.swrCtx.Close()
	return nil
}

func NewResample(format audio.Format) stream.ResampleElement {
	r := &Resample{
		swrCtx: &resample.Resample{},
	}

	r.SetFormat(format)

	return r
}
