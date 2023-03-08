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
	StreamCloser
	Len() int      // 总长度
	Position() int // 当前位置
	Seek(p time.Duration) error
}

type FileStreamer interface {
	StreamSeekCloser
	OpenFile(string) error
	CurrentFile() string
	AudioFormat() *audio.Format   // 当前音频格式
	Duration() time.Duration      // 当前时长
	TotalDuration() time.Duration // 总时长
	Pause(bool)                   // 暂停解码
	IsPaused() bool               // 是否暂停
}

type ElementType int

const (
	ET_WholeSamples ElementType = iota
	ET_OneSample
)

type Element interface {
	Streamer
	Name() string
	Type() ElementType
	Sample(*float64, int, int)
}

type FileStreamerOpenFileHandler func(stream FileStreamer, format *audio.Format)

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
