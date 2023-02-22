package decoder

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/utils"
)

type Streamer interface {
	Stream(*Samples)
}

type StreamCloser interface {
	Streamer
	Close() error
}

type StreamSeekCloser interface {
	Streamer
	Len() int
	Position() int
	Seek(p time.Duration) error
	Close() error
}

type FileStreamer interface {
	StreamSeekCloser
	SetFormat(*audio.Format)
	OpenFile(string) error
	CurrentFile() string
	AudioFormat() *audio.Format
	Duration() time.Duration
	Pause(bool)
	IsPaused() bool
}

type ElementType int

const (
	ET_WholeSamples ElementType = iota
	ET_OneSample
)

type Element interface {
	Name() string
	Type() ElementType
	Stream(*Samples)
	Sample(*float64, int, int)
}

type FileStreamerOpenFileHandler func(format *audio.Format)

type ChannelSamples []float64

// 对应ffmpeg中的planar类型
type Samples struct {
	Size     int
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

func ParseDuration(s string) (time.Duration, error) {
	t, err := time.Parse("15:04:05.9999", s)
	if err != nil {
		return 0, err
	}

	return t.Sub(utils.ZeroTime), nil
}
func DurationFormat(d time.Duration) string {
	return utils.ZeroTime.Add(d).Format("15:04:05.9999")
}
