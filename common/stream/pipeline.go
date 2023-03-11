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
}

type ComputeElement interface {
	Element

	On()
	Off()
	IsOn() bool
}

type VolumeElement interface {
	ComputeElement

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
	AddFileStreamer(FileStreamer)
	FileStreamer() FileStreamer
	Clear()
}

// 播放临时 pcm 格式
type RawPlayerElement interface {
	Element

	Add([]float64)
	Len() int
	IsIdle() bool
}

type ResampleElement interface {
	ComputeElement

	SetFormat(*audio.Format)
	Format() *audio.Format
}

type SpectrumElement interface {
	ComputeElement

	LevelMeter() float64
	Spectrum() []float64
}

type EqualizerElement interface {
	ComputeElement

	SetFilterType(dsp.FilterType)
	FilterType() dsp.FilterType

	SetEqualizer([]dsp.FreqEqualizer)
	Equalizer() []dsp.FreqEqualizer

	Add(int, float64, float64)

	SetDelay(time.Duration)
	Delay() time.Duration
}

type FileStreamerOpenFileHandler func(stream FileStreamer, format *audio.Format)

type PipeLiner interface {
	StreamCloser

	Len() int
	SetBuffer(*Samples)
	Buffer() *Samples

	Prepend(s Element)
	Append(s ...Element)
	Clear()

	Format() *audio.Format

	LastCost() time.Duration
	LastMaxCost() time.Duration

	Lock()
	Unlock()
}
