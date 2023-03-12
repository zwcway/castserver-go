package element

import (
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
)

type Equalizer struct {
	power     bool
	format    audio.Format
	equalizer *dsp.DataProcess
	filters   [][]*dsp.Filter

	locker sync.Mutex
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

	if !e.locker.TryLock() {
		return
	}
	defer e.locker.Unlock()

	if !e.format.Equal(&samples.Format) {
		e.format = samples.Format
		e.init()
	}

	for ch := 0; ch < samples.Format.Layout.Count; ch++ {
		for f := 0; f < len(e.filters[ch]); f++ {
			for i := 0; i < samples.LastNbSamples; i++ {
				samples.Data[ch][i] = e.filters[ch][f].Process(samples.Data[ch][i])
			}
		}
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

func (e *Equalizer) changing(f func()) {
	e.locker.Lock()
	defer e.locker.Unlock()

	f()

	e.init()
}
func (e *Equalizer) SetEqualizer(eq []dsp.Equalizer) {
	e.changing(func() {
		e.equalizer.FEQ = eq
	})
}

func (e *Equalizer) Equalizer() []dsp.Equalizer {
	return e.equalizer.FEQ
}

func (e *Equalizer) Add(freq int, gain, q float64) {
	e.changing(func() {
		e.equalizer.FEQ = append(e.equalizer.FEQ, dsp.NewFIREqualizer(freq, gain, q))
	})
}

func (e *Equalizer) Delay() time.Duration {
	return e.equalizer.Delay
}

func (e *Equalizer) SetDelay(delay time.Duration) {
	e.equalizer.Delay = delay
}

func (e *Equalizer) init() {
	e.filters = make([][]*dsp.Filter, e.format.Layout.Count)

	for ch := 0; ch < len(e.filters); ch++ {
		e.filters[ch] = make([]*dsp.Filter, len(e.equalizer.FEQ))

		for i := 0; i < len(e.equalizer.FEQ); i++ {
			e.filters[ch][i] = dsp.NewFilter(&e.equalizer.FEQ[i], &e.format)
		}
	}
}

func (e *Equalizer) Close() error {
	e.Off()
	if e.filters != nil {
		e.filters = e.filters[:0]
	}
	return nil
}

func NewEqualizer(eq *dsp.DataProcess) stream.EqualizerElement {
	if eq == nil {
		eq = &dsp.DataProcess{}
	}
	e := &Equalizer{equalizer: eq}
	return e
}
