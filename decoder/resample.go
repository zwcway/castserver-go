package decoder

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg/resample"
)

type Resample struct {
	power  bool
	format audio.Format

	swrCtx *resample.Resample
}

func (r *Resample) Name() string {
	return "Resampler"
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
func (e *Resample) OnStarting() {
}

func (e *Resample) OnEnding() {
}

func (e *Resample) OnFormatChanged(newFormat *audio.Format) {
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
	bus.UnregisterObj(r)

	r.swrCtx.Close()
	return nil
}

func (o *Resample) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(o, e, a...)
}
func (o *Resample) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
}

func NewResample(format audio.Format) stream.ResampleElement {
	r := &Resample{
		swrCtx: &resample.Resample{},
	}

	r.SetFormat(format)

	return r
}
