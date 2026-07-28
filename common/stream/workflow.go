package stream

import (
	"image"
	"time"
)

type SourceType uint8

const (
	ST_NONE     SourceType = iota
	ST_File                // 从 文件 播放
	ST_DLNA                // 从 dlna 播放
	ST_AirPlay             // 从 Airplay 播放
	ST_Receiver            // 从 源数据 播放
)

type Workflow struct {
	From SourceType

	PipeLiner

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

func (s *Workflow) ApplySource(f SourceStreamer) {

	if fs, ok := f.(FileStreamer); ok {
		s.From = ST_File
		s.fs = fs
	} else if rs, ok := f.(ReceiverStreamer); ok {
		s.From = ST_Receiver
		s.rs = rs
	}

	s.MixerEle.Add(f)
}

func (s *Workflow) FileStreamer() FileStreamer {
	return s.fs
}

func (s *Workflow) ReceiverStreamer() ReceiverStreamer {
	return s.rs
}

func (s *Workflow) Duration() time.Duration {
	if s.fs == nil {
		return 0
	}
	return s.fs.Duration()
}

func (s *Workflow) TotalDuration() time.Duration {
	if s.fs == nil {
		return 0
	}
	return s.fs.TotalDuration()
}
