package dsp

import (
	"time"
)

// FilterType 滤波器类型
type FilterType = uint8

const (
	LowPassFilter   FilterType = 1 + iota // 低通滤波器
	HighPassFilter                        // 高通滤波器
	PeakingFilter                         // 尖峰滤波器,FIR
	NotchFilter                           // 陷波滤波器,IIR
	LowShelfFilter                        // 低切滤波器
	HighShelfFilter                       // 高切滤波器
)

// FilterParams 滤波器
type FilterParams struct {
	Frequency int     `jp:"freq"`
	Gain      float64 `jp:"g"` // 增益大小
	Q         float64 `jp:"q"` // Q 值
}

const FEQ_MAX_SIZE uint8 = 31

// EqualizerProcessor 均衡器处理器
type EqualizerProcessor struct {
	Delay   time.Duration   `jp:"d"` // 延迟
	Type    FilterType      `jp:"t"`
	Filters []*FilterParams `jp:"fs"`
}

func (d *EqualizerProcessor) Set(freq int, gain, q float64) {
	for i, eq := range d.Filters {
		if eq == nil {
			continue
		}
		if eq.Frequency == freq {
			d.Filters[i] = &FilterParams{freq, gain, q}
			return
		}
	}
	for i, eq := range d.Filters {
		if eq == nil {
			d.Filters[i] = &FilterParams{freq, gain, q}
			return
		}
	}
}

func (d *EqualizerProcessor) Clear(size uint8) bool {
	if size > FEQ_MAX_SIZE {
		return false
	}
	d.Filters = make([]*FilterParams, size)

	return true
}

func NewPeakingFilterEqualizerProcessor(size uint8) *EqualizerProcessor {
	d := &EqualizerProcessor{
		Delay: 0,
		Type:  PeakingFilter,
	}
	d.Clear(size)
	return d
}

func IsFrequencyValid(freq int) bool {
	return freq >= 20 && freq <= 20000
}
