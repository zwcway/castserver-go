package stream

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
)

var (
	BusResample = busResample{}
)

type busResample struct{}

func (busResample) GetInstance(m any, resample *ResampleElement, format *audio.Format) error {
	return bus.DispatchObj(m, "get resample element", resample, format)
}
func (busResample) Register(c func(resample *ResampleElement, format *audio.Format) error) *bus.HandlerData {
	return bus.Register("get resample element", func(o any, a ...any) error {
		return c(a[0].(*ResampleElement), a[1].(*audio.Format))
	})
}
