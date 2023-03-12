package element

import (
	"math"

	"github.com/zwcway/castserver-go/common/stream"
)

type Spectrum struct {
	power bool

	spectrum   []float64
	levelMeter float64
}

const LineSpectrumName = "Line Spectrum"

func (r *Spectrum) Name() string {
	return LineSpectrumName
}

func (r *Spectrum) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (r *Spectrum) Stream(samples *stream.Samples) {

	r.processLevelMeter(samples)
}

func (r *Spectrum) processLevelMeter(samples *stream.Samples) {
	if !r.power {
		return
	}
	var (
		sam  float64
		frac float64
		sum  float64
		rms  float64
	)

	for i := 0; i < samples.Size; i += 10 {
		frac = 0
		for ch := 0; ch < samples.Format.Layout.Count; ch++ {
			sam = samples.Buffer[ch][i]
			if sam > 0 {
				frac += sam
			} else {
				frac += -sam
			}
		}
		frac = frac / float64(samples.Format.Layout.Count)
		sum += frac * frac
	}

	rms = math.Sqrt(sum / float64(samples.Size/10))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	r.levelMeter = rms
}

func (r *Spectrum) Sample(*float64, int, int) {}

func (r *Spectrum) On() {
	r.power = true
}

func (r *Spectrum) Off() {
	r.power = false
}

func (r *Spectrum) IsOn() bool {
	return r.power
}

func (r *Spectrum) LevelMeter() float64 {
	return r.levelMeter
}

func (r *Spectrum) Spectrum() []float64 {
	return r.spectrum
}

func NewSpectrum() stream.SpectrumElement {
	return &Spectrum{power: false}
}
