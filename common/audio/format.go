package audio

import "fmt"

type Format struct {
	SampleRate Rate          // 采样率
	Layout     ChannelLayout // 声道数量
	SampleBits Bits          // 采样格式
}

func (s *Format) String() string {
	return fmt.Sprintf("%d/%d/%s", s.SampleRate.ToInt(), s.Layout.Count, s.SampleBits.Name())
}

func (s *Format) Equal(r *Format) bool {
	return s.Layout.Count == r.Layout.Count && s.SampleRate == r.SampleRate && s.SampleBits == r.SampleBits
}

func (s *Format) Bytes() int {
	return s.Layout.Count * s.SampleBits.Size()
}

func (s *Format) IsValid() bool {
	return s.Layout.Count > 0 && s.SampleRate.IsValid() && s.SampleBits.IsValid()
}

func (s *Format) Mixed(r *Format) {
	s.SampleRate = r.SampleRate
	s.SampleBits = r.SampleBits
	if s.Layout.Count > r.Layout.Count {
		s.Layout = r.Layout
	}
}
