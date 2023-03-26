package dsp

import (
	"time"
)

type FilterType = uint8

const (
	LowPassFilter   FilterType = 1 + iota // 低通滤波器
	HighPassFilter                        // 高通滤波器
	PeakingFilter                         // 尖峰滤波器,FIR
	NotchFilter                           // 陷波滤波器,IIR
	LowShelfFilter                        // 低切滤波器
	HighShelfFilter                       // 高切滤波器
)

type Equalizer struct {
	Type      FilterType `jq:"t"`
	Frequency int        `jp:"freq"`
	Gain      float64    `jp:"g"` // 增益大小
	Q         float64    `jp:"q"` // Q 值
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

func NewFIREqualizer(freq int, gain float64, q float64) *Equalizer {
	return &Equalizer{
		Type:      PeakingFilter,
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	}
}

func NewIIREqualizer(freq int, gain float64, q float64) *Equalizer {
	return &Equalizer{
		Type:      NotchFilter,
		Frequency: freq,
		Gain:      gain,
		Q:         q,
	}
}

const FEQ_MAX_SIZE uint8 = 31

type DataProcess struct {
	Delay time.Duration `jp:"d"` // 延迟
	Type  FilterType    `jp:"t"`
	FEQ   []*Equalizer  `jp:"fs"`
}

func (d *DataProcess) AddFIR(freq int, gain, q float64) bool {
	for i, eq := range d.FEQ {
		if eq == nil {
			continue
		}
		if eq.Frequency == freq {
			d.FEQ[i] = NewFIREqualizer(freq, gain, q)
			return true
		}
	}
	for i, eq := range d.FEQ {
		if eq == nil {
			d.FEQ[i] = NewFIREqualizer(freq, gain, q)
			return true
		}
	}
	return false
}

func (d *DataProcess) Clear(size uint8) bool {
	if size > FEQ_MAX_SIZE {
		return false
	}
	d.FEQ = make([]*Equalizer, size)

	return true
}

func NewDataProcess(size uint8) *DataProcess {
	d := &DataProcess{
		Delay: 0,
		Type:  PeakingFilter,
	}
	d.Clear(size)
	return d
}

func IsFrequencyValid(freq int) bool {
	return freq >= 20 && freq <= 20000
}
