package stream

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
)

var (
	BusSourceFormatChanged = sourceFormatChanged{}
)

type sourceFormatChanged struct{}

func (sourceFormatChanged) Dispatch(ss SourceStreamer, format *audio.Format, channelIndex audio.ChannelIndex) error {
	return bus.DispatchObj(ss, "source format changed", format, channelIndex)
}
func (sourceFormatChanged) Register(ss SourceStreamer, c func(ss SourceStreamer, format *audio.Format, channelIndex audio.ChannelIndex) error) *bus.HandlerData {
	return bus.RegisterObj(ss, "source format changed", func(o any, a ...any) error {
		return c(o.(SourceStreamer), a[0].(*audio.Format), a[1].(audio.ChannelIndex))
	})
}
