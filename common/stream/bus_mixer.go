package stream

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
)

var (
	BusMixerFormatChanged = mixerFormatChanged{}
)

type mixerFormatChanged struct{}

func (mixerFormatChanged) Dispatch(m MixerElement, format *audio.Format, channelIndex *audio.ChannelIndex) error {
	return bus.DispatchObj(m, "mixer format changed", format, channelIndex)
}
func (mixerFormatChanged) Register(m MixerElement, c func(m MixerElement, format *audio.Format, channelIndex *audio.ChannelIndex) error) *bus.HandlerData {
	return bus.RegisterObj(m, "mixer format changed", func(o any, a ...any) error {
		return c(o.(MixerElement), a[0].(*audio.Format), a[1].(*audio.ChannelIndex))
	})
}
