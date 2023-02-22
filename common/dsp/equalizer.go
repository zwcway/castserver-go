package dsp

import "time"

type FilterType uint8

const (
	LowPassFilter   FilterType = 1 + iota // 低通滤波器
	HighPassFilter                        // 高通滤波器
	PeakingFilter                         // 尖峰滤波器
	NotchFilter                           // 陷波滤波器
	LowShelfFilter                        // 低切滤波器
	HighShelfFilter                       // 高切滤波器
)

type FreqEqualizer struct {
	Type      FilterType
	Frequency int
	Gain      float32 // 增益大小
	Q         float32 // Q 值
}

type DataProcess struct {
	Delay time.Duration // 延迟
	FEQ   []FreqEqualizer
}

func IsFrequencyValid(freq int) bool {
	return freq >= 20 && freq <= 20000
}
