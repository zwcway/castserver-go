package audio

import "fmt"

type Format struct {
	SampleRate AudioRate     // 采样率
	Layout     ChannelLayout // 声道数量
	SampleBits Bits          // 采样格式
}

func (s *Format) String() string {
	return fmt.Sprintf("%d/%d/%s", s.SampleRate.ToInt(), s.Layout.Count, s.SampleBits.Name())
}
