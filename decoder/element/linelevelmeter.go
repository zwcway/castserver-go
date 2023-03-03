package element

import (
	"math"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder"
)

type LineLevelMeter struct {
	power bool
	line  *speaker.Line
}

const LevelMeterName = "Resampler"

func (r *LineLevelMeter) Name() string {
	return ResampleName
}

func (r *LineLevelMeter) Type() decoder.ElementType {
	return decoder.ET_WholeSamples
}

func (r *LineLevelMeter) Stream(samples *decoder.Samples) {
	var (
		frac float64
		sum  float64
		rms  float64
	)
	if r.line == nil || !r.power {
		return
	}
	for i := 0; i < samples.Size; i += 10 {
		frac = 0
		for ch := 0; ch < samples.Format.Layout.Count; ch++ {
			frac += samples.Buffer[ch][i]
		}
		frac = frac / float64(samples.Format.Layout.Count)
		sum += frac * frac
	}

	rms = math.Sqrt(sum / float64(samples.Size/10))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	r.line.LevelMeter = rms
}

func (r *LineLevelMeter) Sample(*float64, int, int) {}

func (r *LineLevelMeter) On() {
	r.power = true
}

func (r *LineLevelMeter) Off() {
	r.power = false
}

func NewLineLevelMeter(line *speaker.Line) *LineLevelMeter {
	return &LineLevelMeter{power: false, line: line}
}
