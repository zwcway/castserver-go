package stream

import "github.com/zwcway/castserver-go/common/audio"

type ChannelSamples []float64

// 对应ffmpeg中的planar类型
type Samples struct {
	Size     int //每声道样本数量
	Format   *audio.Format
	Buffer   []ChannelSamples
	LastErr  error
	LastSize int
}

func (s *Samples) Bytes() int {
	return s.Samples() * s.Format.SampleBits.Size()
}

func (s *Samples) Samples() int {
	return s.Format.Layout.Count * s.Size
}

func (s *Samples) BeZero() int {
	return s.BeZeroLeft(0)
}

func (s *Samples) BeZeroLeft(j int) int {
	for ch := 0; ch < len(s.Buffer); ch++ {
		for i := j; i < len(s.Buffer[ch]); i++ {
			s.Buffer[ch][i] = 0
		}
	}
	if len(s.Buffer) == 0 {
		return 0
	}
	return len(s.Buffer[0])
}

func NewSamples(samples int, format *audio.Format) (s *Samples) {
	chs := format.Layout.Count
	s = &Samples{
		Buffer: make([]ChannelSamples, chs),
		Format: format,
		Size:   samples,
	}

	for i := 0; i < chs; i++ {
		s.Buffer[i] = make(ChannelSamples, samples)
	}
	return
}
