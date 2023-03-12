package dsp

import "github.com/zwcway/castserver-go/common/audio"

type DSPFilter interface {
	Init(*audio.Format)
	Process(float64) float64
}
