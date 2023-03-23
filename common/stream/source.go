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
	From   SourceType
	Format audio.Format

	PipeLine PipeLiner

	Mixer MixerElement

	fs FileStreamer
	rs ReceiverStreamer

	Cover  image.Image
	Title  string
	Artist string
}

func (s *Source) ApplySource(f SourceStreamer) {
	s.Mixer.Add(f)

	if fs, ok := f.(FileStreamer); ok {
		s.From = ST_File
		s.fs = fs
	} else if rs, ok := f.(ReceiverStreamer); ok {
		s.From = ST_Receiver
		s.rs = rs
	}

	s.Format = f.AudioFormat()
}

func (s *Source) FileStreamer() FileStreamer {
	return s.fs
}

func (s *Source) ReceiverStreamer() ReceiverStreamer {
	return s.rs
}

func (s *Source) Duration() time.Duration {
	if s.fs == nil {
		return 0
	}
	return s.fs.Duration()
}

func (s *Source) TotalDuration() time.Duration {
	if s.fs == nil {
		return 0
	}
	return s.fs.TotalDuration()
}
