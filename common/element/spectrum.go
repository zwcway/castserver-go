package element

import (
	"math"
	"sync"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
	"gonum.org/v1/gonum/dsp/fourier"
)

const (
	// 44100 ~21hz
	//
	// 192000 ~93hz
	SpectrumCount = 2048

	BufSize = 50
)

// HannWindow 预计算的汉宁窗系数，避免运行时计算
var hannWindow []float64

func init() {
	hannWindow = make([]float64, SpectrumCount)
	for i := 0; i < SpectrumCount; i++ {
		hannWindow[i] = 0.5 * (1 - math.Cos(2*dsp.Pi*float64(i)/float64(SpectrumCount-1)))
	}
}

type freqScale struct {
	frequency  float64 // 频谱图的核心频率
	start, end int     // FFT 结果对应的 索引范围，0 -> n/2+1
	f0, f1     float64 // 索引范围对应的频率范围
}

type Spectrum struct {
	power bool

	n          int          // fft 计算量
	fft        *fourier.FFT // fft 计算器
	fftCots    []complex128 // fft 输入
	spectrum   []float64    // 输出幅值
	levelMeter float64      // 音阶
	bins       int          // 输出大小
	logAxis    bool
	melFilter  *dsp.MelFilterBank // 梅尔滤波器组
	samples    *stream.Samples

	mu sync.Mutex
}

func (r *Spectrum) Name() string {
	return "Spectrum"
}

func (r *Spectrum) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (r *Spectrum) Stream(samples *stream.Samples) {
	if !r.power || samples == nil || samples.LastNbSamples == 0 || r.samples == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; r.samples.LastNbSamples < r.samples.RequestNbSamples && i < samples.RequestNbSamples; i++ {
		r.samples.Data[0][r.samples.LastNbSamples] = samples.Data[0][i]
		r.samples.LastNbSamples++
	}
	if r.samples.LastNbSamples >= r.samples.RequestNbSamples {
		r.samples.LastNbSamples = 0
	}

}

func (r *Spectrum) OnStarting() {
}

func (r *Spectrum) OnEnding() {
}

func (r *Spectrum) OnFormatChanged(newFormat *audio.Format) {
	f := *newFormat
	f.Bits = r.samples.Format.Bits
	r.samples.Resize(r.samples.RequestNbSamples, f)
	r.initMelFilter()
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logAxis = s
	r.initMelFilter()
}

func (r *Spectrum) LogAxis() bool {
	return r.logAxis
}

func (r *Spectrum) SetBins(b int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b <= 10 {
		b = 10
	}
	r.bins = b
	r.initMelFilter()
}

func (r *Spectrum) LevelMeter() float64 {
	return r.levelMeter
}

func (r *Spectrum) Spectrum() []float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.samples.LastNbSamples != 0 {
		return nil
	}
	var (
		suml        float64 // 音阶绝对值求和
		samplesData = r.samples.Data[0]
	)

	// 加汉宁窗
	for i := 0; i < r.n; i++ {
		frac := 0.0

		// 仅使用第一个声道，通常是左声道
		sam := samplesData[i]
		if sam >= 0 {
			frac += sam
		} else {
			frac -= sam
		}

		suml += frac * frac
		if suml != suml { // NaN 检查
			suml = 0
		}

		samplesData[i] *= hannWindow[i]
	}

	// 修复 RMS 计算
	rms := math.Sqrt(suml / float64(r.n))
	rms = math.Max(0.0, math.Min(1.0, rms))

	r.levelMeter = rms

	// FFT 计算
	r.fft.Coefficients(r.fftCots, samplesData[:r.n])

	// 归一化并转换为dB幅值
	for i, c := range r.fftCots {
		samplesData[i] = math.Hypot(real(c), imag(c))
	}
	nbins := len(r.fftCots)

	// 对数轴变换
	if !r.logAxis {
		return samplesData[:nbins]
	}

	nbins = r.melFilter.Apply(r.spectrum, samplesData[:nbins])

	return r.spectrum[:nbins]
}

func (r *Spectrum) FreqiencyList() []int {
	freqList := make([]int, r.bins)
	for i, f := range r.melFilter.CenterFreqs {
		freqList[i] = int(f)
	}
	return freqList
}

func (r *Spectrum) Close() error {
	bus.UnregisterObj(r)
	r.Off()
	return nil
}

func (r *Spectrum) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(r, e, a...)
}
func (r *Spectrum) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(r, e, c)
}

// initMelFilter 初始化梅尔频谱轴 bin 映射表
//
// 梅尔尺度公式:
//
//	mel(f) = 2595 * ln(1 + f/700)
//	f = 700 * (e^(m/2595) - 1)
//
// 输入参数：
//   - fftLen: FFT 数组长度（默认 2048）
//   - rate: 采样率
//   - binsCount: 期望输出的 bin 数量
//
// 输出：
//   - 梅尔频率分布的 bin 映射表，实际数量可能小于 binsCount
func (r *Spectrum) initMelFilter() {
	fftSize := r.n>>1 + 1 // 1025
	r.melFilter = dsp.NewMelFilterBank(fftSize, r.samples.Format.Rate.ToInt(), r.bins)
}

func (r *Spectrum) init() {
	if r.n <= r.bins {
		ml.Error("count too big")
		return
	}
	r.fft = fourier.NewFFT(r.n)
	r.fftCots = make([]complex128, r.n>>1+1)
	r.samples = stream.NewSamples(r.n, audio.DefaultFormat())
	r.spectrum = make([]float64, r.n)

	r.initMelFilter()

}

func NewSpectrum() stream.SpectrumElement {
	s := &Spectrum{
		power:   false,
		n:       SpectrumCount,
		logAxis: true,
		bins:    128,
	}
	s.init()

	return s
}
