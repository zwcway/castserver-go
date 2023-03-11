package dsp

import "time"

type FilterType uint8

const (
	LowPassFilter   FilterType = 1 + iota // 低通滤波器
	HighPassFilter                        // 高通滤波器
	PeakingFilter                         // 尖峰滤波器,FIR
	NotchFilter                           // 陷波滤波器,IIR
	LowShelfFilter                        // 低切滤波器
	HighShelfFilter                       // 高切滤波器
)

type FreqEqualizer struct {
	Type      FilterType
	Frequency int
	Gain      float64 // 增益大小
	Q         float64 // Q 值
}

func NewLowPass(freq int) FreqEqualizer {
	return FreqEqualizer{
		Type:      LowPassFilter,
		Frequency: freq,
	}
}

func NewHighPass(freq int) FreqEqualizer {
	return FreqEqualizer{
		Type:      HighPassFilter,
		Frequency: freq,
	}
}

func NewLowShelf(freq int) FreqEqualizer {
	return FreqEqualizer{
		Type:      LowShelfFilter,
		Frequency: freq,
	}
}

func NewHighShelf(freq int) FreqEqualizer {
	return FreqEqualizer{
		Type:      HighShelfFilter,
		Frequency: freq,
	}
}

func NewFIREqualizer(freq int, gain float64, q float64) FreqEqualizer {
	return FreqEqualizer{
		Type:      PeakingFilter,
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	}
}

func NewIIREqualizer(freq int, gain float64, q float64) FreqEqualizer {
	return FreqEqualizer{
		Type:      NotchFilter,
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	}
}

type DataProcess struct {
	Delay time.Duration // 延迟
	Type  FilterType
	FEQ   []FreqEqualizer
}

func IsFrequencyValid(freq int) bool {
	return freq >= 20 && freq <= 20000
}
