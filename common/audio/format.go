package audio

import "fmt"

type Format struct {
	SampleRate Rate          // 采样率
	Layout     ChannelLayout // 声道数量
	SampleBits Bits          // 采样格式
}

func (s Format) String() string {
	return fmt.Sprintf("%d/%d/%s", s.SampleRate.ToInt(), s.Layout.Count, s.SampleBits.String())
}

func (s *Format) Equal(r *Format) bool {
	return s.Layout.Mask == r.Layout.Mask && s.SampleRate == r.SampleRate && s.SampleBits == r.SampleBits
}

// 仅对比样本格式，忽略声道布局
func (s *Format) SampleEqual(r *Format) bool {
	return s.SampleRate == r.SampleRate && s.SampleBits == r.SampleBits
}

// 仅对比样本格式，忽略声道布局
func (s *Format) SampleLessThan(r *Format) bool {
	return s.SampleRate.ToInt() < r.SampleRate.ToInt() && s.SampleBits.ToInt() == r.SampleBits.ToInt()
}

func (s *Format) LessThan(r *Format) bool {
	return s.Size() < r.Size()
}

func (s *Format) Channels() []Channel {
	return s.Layout.Mask.Slice()
}

// 单样本字节数，包含所有声道
func (s *Format) Size() int {
	return s.Layout.Count * s.SampleBits.Size()
}

func (s *Format) IsValid() bool {
	return s.Layout.IsValid() && s.SampleRate.IsValid() && s.SampleBits.IsValid()
}

// 每声道样本字节数
func (s *Format) SamplesSize(nbSamples int) int {
	return nbSamples * s.SampleBits.Size()
}

// 每声道样本数量
func (s *Format) SamplesCount(nbSamplesSize int) int {
	return nbSamplesSize / s.SampleBits.Size()
}

// 所有声道样本总大小
func (s *Format) AllSamplesSize(nbSamples int) int {
	return s.SamplesSize(nbSamples) * s.Layout.Count
}

func (s *Format) Mixin(r Format) {
	s.SampleRate = r.SampleRate
	s.SampleBits = r.SampleBits
	if s.Layout.Count > r.Layout.Count {
		s.Layout = r.Layout
	}
}
func (s *Format) InitFrom(r Format) {
	if !s.SampleRate.IsValid() {
		s.SampleRate = r.SampleRate
	}
	if !s.SampleBits.IsValid() {
		s.SampleBits = r.SampleBits
	}
	if !s.Layout.IsValid() {
		s.Layout = r.Layout
	}
}

func DefaultFormat() Format {
	return Format{
		SampleRate: AudioRate_44100,
		Layout:     ChannelLayout10,
		SampleBits: Bits_DEFAULT,
	}
}

// 内部处理格式，其他无所谓，位宽必须是float64
func InternalFormat() Format {
	return Format{
		SampleBits: Bits_DEFAULT,
	}
}
