package stream

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
)

type mixerOutFormatChanged struct{}

func (mixerOutFormatChanged) Dispatch(m MixerElement, format *audio.Format, channelIndex audio.ChannelIndex) error {
	return bus.DispatchObj(m, "mixer out format changed", format, channelIndex)
}
func (mixerOutFormatChanged) Register(m MixerElement, c func(m MixerElement, format *audio.Format, channelIndex audio.ChannelIndex) error) *bus.HandlerData {
	return bus.RegisterObj(m, "mixer out format changed", func(o any, a ...any) error {
		return c(o.(MixerElement), a[0].(*audio.Format), a[1].(audio.ChannelIndex))
	})
}

type mixerInFormatChanged struct{}

func (mixerInFormatChanged) Dispatch(m MixerElement, s SourceStreamer, format *audio.Format, channelIndex audio.ChannelIndex) error {
	return bus.DispatchObj(m, "mixer in format changed", s, format, channelIndex)
}
func (mixerInFormatChanged) Register(m MixerElement, c func(m MixerElement, s SourceStreamer, format *audio.Format, channelIndex audio.ChannelIndex) error) *bus.HandlerData {
	return bus.RegisterObj(m, "mixer in format changed", func(o any, a ...any) error {
		return c(o.(MixerElement), a[0].(SourceStreamer), a[1].(*audio.Format), a[2].(audio.ChannelIndex))
	})
}

var (
	// 混音器输出格式变更
	BusMixerOutFormatChanged = mixerOutFormatChanged{}

	// 混音器输入格式变更
	BusMixerInFormatChanged = mixerInFormatChanged{}
)
