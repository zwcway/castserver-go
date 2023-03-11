package stream

import (
	"image"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
)

type SourceType uint8

const (
	ST_NONE     SourceType = iota
	ST_File                // 从 文件 播放
	ST_DLNA                // 从 dlna 播放
	ST_AirPlay             // 从 Airplay 播放
	ST_Receiver            // 从 源数据 播放
)

type Source struct {
	From SourceType
	audio.Format

	PipeLine     PipeLiner
	FileStreamer FileStreamer

	Cover  image.Image
	Title  string
	Artist string
}

func (s *Source) FromFileStreamer(f FileStreamer) {
	s.FileStreamer = f
	s.From = ST_File
	s.Format = *f.AudioFormat()
}

func (s *Source) Duration() time.Duration {
	if s.FileStreamer == nil {
		return 0
	}

	return s.FileStreamer.Duration()
}

func (s *Source) TotalDuration() time.Duration {
	if s.FileStreamer == nil {
		return 0
	}

	return s.FileStreamer.TotalDuration()
}
