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

	PipeLine PipeLiner

	MixerEle     MixerElement
	VolumeEle    VolumeElement
	SpectrumEle  SpectrumElement
	EqualizerEle EqualizerElement
	PlayerEle    RawPlayerElement
	// ResampleEle  ResampleElement
	// PusherEle    SwitchElement

	fs FileStreamer
	rs ReceiverStreamer

	Cover  image.Image
	Title  string
	Artist string
}

func (s *Source) ApplySource(f SourceStreamer) {

	if fs, ok := f.(FileStreamer); ok {
		s.From = ST_File
		s.fs = fs
	} else if rs, ok := f.(ReceiverStreamer); ok {
		s.From = ST_Receiver
		s.rs = rs
	}

	s.MixerEle.Add(f)
}

func (s *Source) Format() audio.Format {
	if s.fs != nil {
		return s.fs.AudioFormat()
	} else if s.rs != nil {
		return s.rs.AudioFormat()
	}
	return s.MixerEle.Format()
}

func (s *Source) SetFormat(f audio.Format) {
	s.MixerEle.SetFormat(f)
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
