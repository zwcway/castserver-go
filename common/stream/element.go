package stream

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
)

type ElementType int

const (
	ET_WholeSamples ElementType = iota
	ET_OneSample
)

// Element 音频流元
type Element interface {
	Streamer
	bus.Eventer

	// Type 元类型
	Type() ElementType
	// Sample 开始流处理
	// Sample(*float64, int, int)
	// Close 释放元
	Close() error

	// OnStarting 开始处理数据
	OnStarting()
	// OnEnding 结束处理数据
	OnEnding()

	// OnFormatChanged 音频格式被改变后触发
	OnFormatChanged(*audio.Format)
}

// SwitchElement 开关元
type SwitchElement interface {
	Element

	On()
	Off()
	IsOn() bool
}

// VolumeElement 音量元
type VolumeElement interface {
	SwitchElement

	SetVolume(float64)
	Volume() float64

	SetMute(bool)
	Mute() bool
}

// MixerElement 音频混音元
type MixerElement interface {
	Element

	Len() int
	Del(SourceStreamer)
	Has(SourceStreamer) bool
	Add(...SourceStreamer)
	Clear()

	// 是否开启转码。
	// 如果开启转码，SetFormat 将作用于转码器
	// 否则通知 SourceStreamer 自己转码
	SetResample(bool)

	SetFormat(audio.Format) // 设置输出格式
	Format() audio.Format   // 获取输出格式

	Buffer() *Samples
}

// ChannelMixerElement 声道混音元
type ChannelMixerElement interface {
	Element

	SetRoute([]audio.ChannelRoute)
	Route() []audio.ChannelRoute
}

// RawPlayerElement 临时播放元
type RawPlayerElement interface {
	Element
	SourceStreamer

	AddPCM(audio.Format, []byte)
	AddPCMWithChannel(audio.Channel, audio.Format, []byte)
}

// ResampleElement 转码元
type ResampleElement interface {
	SwitchElement

	SetFormat(audio.Format) // 设置转码目标格式
	Format() audio.Format   // 获取转码目标格式
}

// SpectrumElement 频谱元
type SpectrumElement interface {
	SwitchElement

	SetLogAxis(bool)
	LogAxis() bool

	LevelMeter() float64
	Spectrum() []float64
}

// EqualizerElement 均衡器元
type EqualizerElement interface {
	SwitchElement

	SetFilterType(dsp.FilterType)
	FilterType() dsp.FilterType

	SetEqualizer([]*dsp.FilterParams)
	Equalizer() []*dsp.FilterParams

	Set(int, float64, float64)

	Count() int

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
}
