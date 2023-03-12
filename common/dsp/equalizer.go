package dsp

import (
	"time"
)

type FilterType uint8

const (
	LowPassFilter   FilterType = 1 + iota // 低通滤波器
	HighPassFilter                        // 高通滤波器
	PeakingFilter                         // 尖峰滤波器,FIR
	NotchFilter                           // 陷波滤波器,IIR
	LowShelfFilter                        // 低切滤波器
	HighShelfFilter                       // 高切滤波器
)

type Equalizer struct {
	Type      FilterType
	Frequency int
	Gain      float64 // 增益大小
	Q         float64 // Q 值
}

func NewLowPass(freq int) Equalizer {
	return Equalizer{
		Type:      LowPassFilter,
		Frequency: freq,
	}
}

func NewHighPass(freq int) Equalizer {
	return Equalizer{
		Type:      HighPassFilter,
		Frequency: freq,
	}
}

func NewLowShelf(freq int) Equalizer {
	return Equalizer{
		Type:      LowShelfFilter,
		Frequency: freq,
	}
}

func NewHighShelf(freq int) Equalizer {
	return Equalizer{
		Type:      HighShelfFilter,
		Frequency: freq,
	}
}

func NewFIREqualizer(freq int, gain float64, q float64) Equalizer {
	return Equalizer{
		Type:      PeakingFilter,
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	}
}

func NewIIREqualizer(freq int, gain float64, q float64) Equalizer {
	return Equalizer{
		Type:      NotchFilter,
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	}
}

type DataProcess struct {
	Delay time.Duration // 延迟
	Type  FilterType
	FEQ   []Equalizer
}

func IsFrequencyValid(freq int) bool {
	return freq >= 20 && freq <= 20000
}
