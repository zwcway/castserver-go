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
	Clear()

	Buffer() *Samples
}

type ChannelMixerElement interface {
	Element

	SetRoute([]audio.ChannelRoute)
	Route() []audio.ChannelRoute
}

// 播放临时 pcm 格式
type RawPlayerElement interface {
	Element

	Add(audio.Format, []byte)
	AddToChannel(audio.Channel, audio.Format, []byte)
	Len() int
	IsIdle() bool
}

type ResampleElement interface {
	SwitchElement

	SetFormat(audio.Format) // 设置转码目标格式
	Format() audio.Format   // 获取转码目标格式
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

	SetEqualizer([]*dsp.Equalizer)
	Equalizer() []*dsp.Equalizer

	Add(int, float64, float64)

	SetDelay(time.Duration)
	Delay() time.Duration
}

const EqualizerDelayMax time.Duration = 300 * time.Millisecond // 大约 102 米

type PipeLiner interface {
	StreamCloser

	Len() int
	SetBuffer(*Samples)
	Buffer() *Samples

	Prepend(s Element)
	Append(s ...Element)
	Clear()

	LastCost() time.Duration
	LastMaxCost() time.Duration

	Lock()
	Unlock()
}
