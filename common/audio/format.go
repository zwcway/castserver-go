package audio

import "fmt"

type Sample struct {
	Rate // 采样率
	Bits // 采样格式
}

type Format struct {
	Sample // 样本格式
	Layout // 声道数量
}

func (s *Format) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("%d/%d/%s", s.Sample.Rate.ToInt(), s.Layout.Count, s.Sample.Bits.String())
}

func (s Format) Equal(r Format) bool {
	return s.Layout.ChannelMask == r.Layout.ChannelMask && s.Sample.Rate == r.Sample.Rate && s.Sample.Bits == r.Sample.Bits
}

// 仅对比样本格式，忽略声道布局
func (s Format) SampleEqual(r Format) bool {
	return s.Sample.Rate == r.Sample.Rate && s.Sample.Bits == r.Sample.Bits
}

// 仅对比样本格式，忽略声道布局
func (s Format) SampleLessThan(r Format) bool {
	return s.Sample.Rate.ToInt() < r.Sample.Rate.ToInt() && s.Sample.Bits.ToInt() == r.Sample.Bits.ToInt()
}

func (s Format) LessThan(r Format) bool {
	if s.Layout.Count < r.Layout.Count {
		return true
	}
	if s.Bits.Size() < r.Bits.Size() {
		return true
	}
	if s.Rate.ToInt() < r.Bits.ToInt() {
		return true
	}
	return false
}

func (s Format) Channels() []Channel {
	return s.Layout.Slice()
}

// 单样本字节数，包含所有声道
func (s Format) Size() int {
	return int(s.Layout.Count) * s.Sample.Bits.Size()
}

func (s Format) IsValid() bool {
	return s.Layout.IsValid() && s.Sample.Rate.IsValid() && s.Sample.Bits.IsValid()
}

// 每声道样本字节数
func (s Format) SamplesSize(nbSamples int) int {
	return nbSamples * s.Sample.Bits.Size()
}

// 每声道样本数量
func (s Format) SamplesCount(nbSamplesSize int) int {
	return nbSamplesSize / s.Sample.Bits.Size()
}

// 所有声道样本总大小
func (s Format) AllSamplesSize(nbSamples int) int {
	return s.SamplesSize(nbSamples) * int(s.Layout.Count)
}

func (s *Format) Mixin(r Format) {
	s.Sample.Rate = r.Sample.Rate
	s.Sample.Bits = r.Sample.Bits
	if s.Layout.Count > r.Layout.Count {
		s.Layout = r.Layout
	}
}
func (s *Format) InitFrom(r Format) {
	if !s.Sample.Rate.IsValid() {
		s.Sample.Rate = r.Sample.Rate
	}
	if !s.Sample.Bits.IsValid() {
		s.Sample.Bits = r.Sample.Bits
	}
	if !s.Layout.IsValid() {
		s.Layout = r.Layout
	}
}

func InternalFormatCombine(dst Format) Format {
	src := InternalFormat()
	if !dst.Sample.Rate.IsValid() {
		dst.Sample.Rate = src.Sample.Rate
	}
	if !dst.Layout.IsValid() {
		dst.Layout = src.Layout
	}
	return dst
}

func DefaultFormat() Format {
	return Format{
		Sample: Sample{
			Rate: AudioRate_44100,
			Bits: Bits_DEFAULT,
		},
		Layout: Layout10,
	}
}

// 内部处理格式，其他无所谓，位宽必须是float64
func InternalFormat() Format {
	return Format{
		Sample: Sample{
			Bits: Bits_DEFAULT,
		},
	}
}
