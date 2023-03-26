package element

import (
	"math"

	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
)

// 44100 ~21hz
//
// 192000 ~93hz
const SpectrumCount = 2048

type Spectrum struct {
	power bool

	n          int       // fft 计算量
	pos        int       // 如果样本数量少于 n
	s          []float64 // fft 输入数据
	spectrum   []float64 // 输出幅值，大小 < n/2
	suml       float64   // 音阶绝对值求和
	levelMeter float64   // 音阶
	logAxis    bool
	hasData    bool
}

func (r *Spectrum) Name() string {
	return "Spectrum"
}

func (r *Spectrum) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (r *Spectrum) zeroData() {
	for i := 0; i < len(r.spectrum); i++ {
		r.spectrum[i] = 0
	}
	for i := 0; i < r.n; i++ {
		r.s[i] = 0
	}
	r.levelMeter = 0
	r.suml = 0
	r.pos = 0
}

func (r *Spectrum) Stream(samples *stream.Samples) {
	if !r.power || samples == nil || samples.LastNbSamples == 0 {
		if r.hasData {
			r.zeroData()
		}
		r.hasData = false
		return
	}

	var (
		sam  float64
		frac float64 // 多声道音阶求平均
		rms  float64
		sums float64 // 频谱求和
		// 不同采样率下保证fft的频率宽度约小于20hz，大于20hz的采样率下求平均，模拟转码
		c     int = 20 / (samples.Format.SampleRate.ToInt() / samples.LastNbSamples)
		i, ch int
	)
	r.hasData = true
	if c < 1 {
		c = 1
	}

	for i = 0; i < samples.LastNbSamples && r.pos < r.n; i += c {
		frac = 0
		sums = 0
		for ch = 0; ch < samples.Format.Layout.Count; ch++ {
			// for j = 0; j < c; j++ {
			sam = samples.Data[ch][i]
			sums += sam
			if sam >= 0 {
				frac += sam
			} else {
				frac += -sam
			}
			// }
		}
		frac = frac / float64(samples.Format.Layout.Count)

		r.s[r.pos] = sums / float64(samples.Format.Layout.Count)
		// 加汉宁窗
		r.s[r.pos] *= (1 - math.Cos(2*dsp.Pi*float64(r.pos)/float64(r.n-1)))
		r.pos++

		r.suml += frac * frac
		if r.suml != r.suml {
			// 过滤异常
			r.suml = 0
		}
	}

	if r.pos >= r.n-1 {
		r.pos = 0
		i = dsp.FFTN(r.s, r.n, samples.Format.SampleRate.ToInt(), r.logAxis)
		if len(r.spectrum) != i {
			r.spectrum = make([]float64, i)
		}
		copy(r.spectrum, r.s[:i])

		rms = math.Sqrt(r.suml / float64(r.n>>3))
		rms = math.Max(0.0, rms)
		rms = math.Min(1.0, rms)

		r.levelMeter = rms
		r.suml = 0
	}
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

func (r *Spectrum) SetLogAxis(s bool) {
	r.logAxis = s
}

func (r *Spectrum) LogAxis() bool {
	return r.logAxis
}

func (r *Spectrum) LevelMeter() float64 {
	if r.levelMeter != r.levelMeter {
		return 0
	}
	return r.levelMeter
}

func (r *Spectrum) Spectrum() []float64 {
	return r.spectrum
}

func (r *Spectrum) Close() error {
	r.Off()
	return nil
}

func (r *Spectrum) init() {
	r.s = make([]float64, r.n)
	// r.spectrum = make([]float64, r.n>>1)
}

func NewSpectrum() stream.SpectrumElement {
	s := &Spectrum{
		power:   false,
		n:       SpectrumCount,
		logAxis: true,
	}
	s.init()
	return s
}
