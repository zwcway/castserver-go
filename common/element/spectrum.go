package element

import (
	"math"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
)

// 44100 ~21hz
//
// 192000 ~93hz
const SpectrumCount = 2048

// HannWindow 预计算的汉宁窗系数，避免运行时计算
var hannWindow []float64

func init() {
	hannWindow = make([]float64, SpectrumCount)
	for i := 0; i < SpectrumCount; i++ {
		hannWindow[i] = 0.5 * (1 - math.Cos(2*dsp.Pi*float64(i)/float64(SpectrumCount-1)))
	}
}

type Spectrum struct {
	power bool

	n            int       // fft 计算量
	pos          int       // 如果样本数量少于 n
	s            []float64 // fft 输入数据
	spectrum     []float64 // 输出幅值，大小 < n/2
	suml         float64   // 音阶绝对值求和
	levelMeter   float64   // 音阶
	logAxis      bool
	hasData      bool
	rate         int   // 采样率缓存
	logBins      int   // 对数轴变换后的 bin 数量
	logBinsCache []int // 对数轴 bin 映射表
}

func (r *Spectrum) Name() string {
	return "Spectrum"
}

func (r *Spectrum) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (r *Spectrum) zeroData() {
	clear(r.spectrum)
	clear(r.s)
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
		sums float64 // 频谱求和
		// 不同采样率下保证fft的频率宽度约小于20hz，大于20hz的采样率下求平均，模拟转码
		c     int = 20 * samples.LastNbSamples / samples.Format.Rate.ToInt()
		i, ch int
	)
	r.hasData = true
	if c < 1 {
		c = 1
	}

	rate := samples.Format.Rate.ToInt()
	chCount := int(samples.Format.Count)

	// 检查采样率是否变化，重新计算对数轴映射
	if r.rate != rate {
		r.rate = rate
		r.updateLogBins()
	}

	for i = 0; i < samples.LastNbSamples && r.pos < r.n; i += c {
		frac = 0
		sums = 0
		for ch = 0; ch < chCount; ch++ {
			sam = samples.Data[ch][i]
			sums += sam
			if sam >= 0 {
				frac += sam
			} else {
				frac -= sam
			}
		}
		frac /= float64(chCount)

		// 只填充数据，窗函数稍后统一应用
		r.s[r.pos] = sums / float64(chCount)
		r.pos++

		r.suml += frac * frac
		if r.suml != r.suml { // NaN 检查
			r.suml = 0
		}
	}

	if r.pos >= r.n-1 {
		// 在 FFT 前统一应用窗函数
		for i = 0; i < r.n; i++ {
			r.s[i] *= hannWindow[i]
		}

		nbins := dsp.FFTN(r.s, r.n, rate, r.logAxis)
		if len(r.spectrum) != nbins {
			r.spectrum = make([]float64, nbins)
		}
		copy(r.spectrum, r.s[:nbins])

		// 修复 RMS 计算：使用实际 FFT 点数 n，而非 n>>3
		rms := math.Sqrt(r.suml / float64(r.n))
		rms = math.Max(0.0, math.Min(1.0, rms))

		r.levelMeter = rms
		r.suml = 0
		r.pos = 0
	}
}

// updateLogBins 更新对数轴 bin 映射表
func (r *Spectrum) updateLogBins() {
	n := r.n >> 1
	if r.logBinsCache == nil {
		r.logBinsCache = make([]int, n)
	}

	// 使用对数映射：频率轴采用对数刻度，靠近低频分辨率更高
	// bins[i] 表示第 i 个对数 bin 对应线性频谱的起始索引
	scale := float64(n)
	for i := 0; i < n; i++ {
		// 对数映射公式
		r.logBinsCache[i] = int(math.Pow(float64(i+1)/float64(n), 0.8) * scale)
		if r.logBinsCache[i] >= n {
			r.logBinsCache[i] = n - 1
		}
	}
}

func (e *Spectrum) OnStarting() {
}

func (e *Spectrum) OnEnding() {
}

func (e *Spectrum) OnFormatChanged(newFormat *audio.Format) {
}

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
	bus.UnregisterObj(r)

	r.Off()
	return nil
}

func (o *Spectrum) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(o, e, a...)
}
func (o *Spectrum) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
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
