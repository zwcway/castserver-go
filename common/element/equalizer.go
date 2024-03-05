package element

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
)

type Equalizer struct {
	power     bool
	format    audio.Format
	equalizer *dsp.EqualizerProcessor

	filters [][]*dsp.Filter

	newFilter [][]*dsp.Filter
}

func (e *Equalizer) Name() string {
	return "Equalizer"
}

func (e *Equalizer) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (e *Equalizer) Stream(samples *stream.Samples) {
	if !e.power || samples == nil || samples.LastNbSamples == 0 {
		return
	}
	for ch := 0; ch < int(samples.Format.Layout.Count); ch++ {
		for _, f := range e.filters[ch] {
			for i := 0; i < samples.LastNbSamples; i++ {
				samples.Data[ch][i] = f.Process(samples.Data[ch][i])
			}
		}
	}
}

func (e *Equalizer) Sample(sample *float64, ch int, n int) {
}

func (e *Equalizer) OnStarting() {
	if e.newFilter == nil {
		return
	}
	e.filters = e.newFilter
}

func (e *Equalizer) OnEnding() {
}

func (e *Equalizer) OnFormatChanged(newFormat *audio.Format) {
}

func (r *Equalizer) On() {
	r.power = true
}

func (r *Equalizer) Off() {
	r.power = false
}

func (r *Equalizer) IsOn() bool {
	return r.power
}

func (e *Equalizer) SetFilterType(t dsp.FilterType) {
	e.equalizer.Type = t
}
func (e *Equalizer) FilterType() dsp.FilterType {
	return e.equalizer.Type
}

func (e *Equalizer) SetEqualizer(eq []*dsp.FilterParams) {
	e.equalizer.Filters = eq
	e.init(&e.newFilter)
}

func (e *Equalizer) Equalizer() []*dsp.FilterParams {
	return e.equalizer.Filters
}

func (e *Equalizer) Count() int {
	return len(e.equalizer.Filters)
}

func (e *Equalizer) Set(freq int, gain, q float64) {
	e.equalizer.Set(freq, gain, q)

	e.init(&e.newFilter)
	return
}

func (e *Equalizer) Delay() time.Duration {
	return e.equalizer.Delay
}

func (e *Equalizer) SetDelay(delay time.Duration) {
	e.equalizer.Delay = delay
}

func (e *Equalizer) init(filter *[][]*dsp.Filter) {
	*filter = make([][]*dsp.Filter, e.format.Layout.Count)
	chCount := int(e.format.Layout.Count)
	rate := e.format.Rate.ToInt()

	for ch := 0; ch < chCount; ch++ {
		fch := make([]*dsp.Filter, 0)

		for _, f := range e.equalizer.Filters {
			if f == nil {
				continue
			}
			fch = append(fch, dsp.NewFilter(*f, rate))
		}

		(*filter)[ch] = fch
	}
}

func (e *Equalizer) Close() error {
	bus.UnregisterObj(e)

	e.Off()
	if e.filters != nil {
		e.filters = e.filters[:0]
	}
	return nil
}

func (o *Equalizer) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(o, e, a...)
}
func (o *Equalizer) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
}

func NewEqualizer(eq *dsp.EqualizerProcessor) stream.EqualizerElement {
	if eq == nil {
		eq = dsp.NewPeakingFilterEqualizerProcessor(0)
	}
	e := &Equalizer{equalizer: eq}

	return e
}
