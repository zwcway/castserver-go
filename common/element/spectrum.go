package element

import (
	"math"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/stream"
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

type spectrum struct {
	S  []float64 // 输出幅值，大小 < n/2
	LL float64   // 音阶
}

type Spectrum struct {
	power bool

	n            int        // fft 计算量
	pos          int        // 环形缓冲池写入位置
	s            []float64  // 环形缓冲池
	spectrum     []spectrum // 输出幅值，大小 < n/2
	suml         float64    // 音阶绝对值求和
	logAxis      bool
	hasData      bool
	rate         int   // 采样率缓存
	logBins      int   // 对数轴变换后的 bin 数量
	logBinsCache []int // 对数轴 bin 映射表
	samples      *stream.Samples

	writePos int
	readPos  int
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
	// 不同采样率下保证fft的频率宽度约小于20hz，大于20hz的采样率下求平均，模拟转码
	)
	r.hasData = true
	rate := samples.Format.Rate.ToInt()

	// 检查采样率是否变化，重新计算对数轴映射
	if r.rate != rate {
		r.rate = rate
		r.samples.Resize(samples.LastNbSamples, samples.Format)
		r.updateLogBins()
	}

	// 通知开始计算

	var (
		sam  float64
		frac float64 // 多声道音阶求平均
		sums float64 // 频谱求和
		// 不同采样率下保证fft的频率宽度约小于20hz，大于20hz的采样率下求平均，模拟转码
		c       int = 20 * samples.LastNbSamples / samples.Format.Rate.ToInt()
		i, ch   int
		chCount = int(samples.Format.Count)
	)
	if c < 1 {
		c = 1
	}

	for i = 0; i < samples.LastNbSamples; i += c {
		frac = 0
		sums = 0

		// 多声道混合计算
		for ch = range chCount {
			sam = samples.Data[ch][i]
			sums += sam
			if sam >= 0 {
				frac += sam
			} else {
				frac -= sam
			}
		}
		// 多声道音阶求平均
		frac /= float64(chCount)
		sums /= float64(chCount)

		r.suml += frac * frac
		if r.suml != r.suml { // NaN 检查
			r.suml = 0
		}

		// 填充环形缓冲池
		r.s[r.pos] = sums
		r.pos++

		if r.pos >= r.n {
			// 应用窗函数
			for i := 0; i < r.n; i++ {
				r.s[i] *= hannWindow[i]
			}

			nbins := dsp.FFTN(r.s, r.n, r.rate, r.logAxis)

			size := samples.Format.Rate.ToInt() * 2 / r.n

			if len(r.spectrum) == 0 || len(r.spectrum[r.writePos].S) != nbins {
				// ml.Debug("spectrum length change", lg.Int("nbins", int64(nbins)), lg.Int("old", int64(len(r.spectrum))))
				r.spectrum = make([]spectrum, size)
				for c := range size {
					r.spectrum[c].S = make([]float64, nbins)
				}
			}
			copy(r.spectrum[r.writePos].S, r.s[:nbins])

			// 修复 RMS 计算：使用实际 FFT 点数 n，而非 n>>3
			rms := math.Sqrt(r.suml / float64(r.n))
			rms = math.Max(0.0, math.Min(1.0, rms))

			r.spectrum[r.writePos].LL = rms

			// FFT 完成后重置缓冲池
			r.suml = 0
			r.pos = 0

			// r.readPos = 0
			r.writePos = (r.writePos + 1) % len(r.spectrum)
			// break
		}
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
	if r.readPos < len(r.spectrum) {
		return r.spectrum[r.readPos].LL
	}
	return 0
}

func (r *Spectrum) Spectrum() (ret []float64) {
	if r.readPos < len(r.spectrum) {
		ret = r.spectrum[r.readPos].S

		r.readPos = (r.readPos + 1) % len(r.spectrum)
	}
	return
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
	r.samples = stream.NewSamples(r.n, audio.DefaultFormat())
	r.s = make([]float64, r.n)
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
