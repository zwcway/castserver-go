package element

import (
	"time"

	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
)

type Equalizer struct {
	power     bool
	equalizer *dsp.DataProcess
}

func (e *Equalizer) Name() string {
	return "Equalizer"
}

func (e *Equalizer) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (e *Equalizer) Stream(samples *stream.Samples) {
	if !e.power {
		return
	}

}

func (e *Equalizer) Sample(sample *float64, ch int, n int) {
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

func (e *Equalizer) SetEqualizer(eq []dsp.FreqEqualizer) {
	e.equalizer.FEQ = eq
}

func (e *Equalizer) Equalizer() []dsp.FreqEqualizer {
	return e.equalizer.FEQ
}

func (e *Equalizer) Add(freq int, gain, q float64) {
	e.equalizer.FEQ = append(e.equalizer.FEQ, dsp.FreqEqualizer{
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	})
}

func (e *Equalizer) Delay() time.Duration {
	return e.equalizer.Delay
}

func (e *Equalizer) SetDelay(delay time.Duration) {
	e.equalizer.Delay = delay
}

func NewEqualizer(eq *dsp.DataProcess) stream.EqualizerElement {
	if eq == nil {
		eq = &dsp.DataProcess{}
	}
	return &Equalizer{equalizer: eq}
}
