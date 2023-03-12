package stream

import (
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
)

type ChannelSamples []float64

// 对应ffmpeg中的planar类型
type Samples struct {
	Size     int //每声道样本数量
	Format   *audio.Format
	Buffer   []ChannelSamples
	LastErr  error
	LastSize int
}

func (s *Samples) TotalSize() int {
	return s.AllSamples() * s.Format.SampleBits.Size()
}

func (s *Samples) AllSamples() int {
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

func (s *Samples) ToPacked(ch int) []byte {
	bits := s.Format.SampleBits.Size()
	bs := make([]byte, s.Size*bits)
	planr := s.Buffer[ch]

	switch bits {
	case 1:
		for i := 0; i < s.LastSize; i++ {
			bs[i] = *(*byte)(unsafe.Pointer(&planr[i]))
		}
	case 2:
		j := 0
		for i := 0; i < s.LastSize; i++ {
			valuint16 := *(*uint16)(unsafe.Pointer(&planr[i]))
			bs[j] = byte(valuint16)
			j++
			bs[j] = byte(valuint16 >> 8)
			j++
		}
	case 3:
		j := 0
		for i := 0; i < s.LastSize; i++ {
			valuint24 := *(*uint32)(unsafe.Pointer(&planr[i]))
			bs[j] = byte(valuint24)
			j++
			bs[j] = byte(valuint24 >> 8)
			j++
			bs[j] = byte(valuint24 >> 16)
			j++
		}
	case 4:
		j := 0
		for i := 0; i < s.LastSize; i++ {
			valuint32 := *(*uint32)(unsafe.Pointer(&planr[i]))
			bs[j] = byte(valuint32)
			j++
			bs[j] = byte(valuint32 >> 8)
			j++
			bs[j] = byte(valuint32 >> 16)
			j++
			bs[j] = byte(valuint32 >> 24)
			j++
		}
	}

	return bs
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
