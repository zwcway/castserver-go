package stream

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/dsp"
)

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
	Close() error
}

type SwitchElement interface {
	Element

	On()
	Off()
	IsOn() bool
}

type VolumeElement interface {
	SwitchElement

	SetVolume(float64)
	Volume() float64

	SetMute(bool)
	Mute() bool
}

type MixerElement interface {
	Element

	Len() int
	Del(Streamer)
	Has(Streamer) bool
	Add(...Streamer)
	PreAdd(...Streamer)
	SetFileStreamer(FileStreamer)
	FileStreamer() FileStreamer
	Clear()
}

type ChannelMixerElement interface {
	Element

	SetRoute([]audio.ChannelRoute)
	Route() []audio.ChannelRoute
}

// 播放临时 pcm 格式
type RawPlayerElement interface {
	Element

	Add([]float64)
	Len() int
	IsIdle() bool
}

type ResampleElement interface {
	SwitchElement

	SetFormat(audio.Format)
	Format() audio.Format
}

type SpectrumElement interface {
	SwitchElement

	SetLogAxis(bool)
	LogAxis() bool

	LevelMeter() float64
	Spectrum() []float64
}

type EqualizerElement interface {
	SwitchElement

	SetFilterType(dsp.FilterType)
	FilterType() dsp.FilterType

	SetEqualizer([]dsp.Equalizer)
	Equalizer() []dsp.Equalizer

	Add(int, float64, float64)

	SetDelay(time.Duration)
	Delay() time.Duration
}

type FileStreamerOpenFileHandler func(stream FileStreamer, inFormat, outFormat audio.Format)

type PipeLiner interface {
	StreamCloser

	Len() int
	SetBuffer(*Samples)
	Buffer() *Samples

	Prepend(s Element)
	Append(s ...Element)
	Clear()

	Format() audio.Format // 获取输入格式

	LastCost() time.Duration
	LastMaxCost() time.Duration

	Lock()
	Unlock()
}
