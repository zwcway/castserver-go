package websockets

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
)

func register() {

	bus.Register("speaker detected", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		// 触发设备发现事件，通知管理后台
		BroadcastSpeakerEvent(sp, Event_SP_Detected)
		return nil
	}).ASync()
	bus.Register("speaker offline", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		BroadcastSpeakerEvent(sp, Event_SP_Offline)
		return nil
	}).ASync()
	bus.Register("speaker online", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		// 触发设备上线事件，通知管理后台
		BroadcastSpeakerEvent(sp, Event_SP_Online)
		return nil
	}).ASync()
	bus.Register("speaker channel moved", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		och := a[0].(audio.Channel)

		BroadcastSpeakerChannelMovedEvent(sp, och, sp.SampleChannel())
		return nil
	}).ASync()
	bus.Register("line created", func(o any, a ...any) error {
		line := a[0].(*speaker.Line)
		BroadcastLineEvent(line, Event_Line_Created)
		return nil
	}).ASync()
	bus.Register("line deleted", func(o any, a ...any) error {
		line := o.(*speaker.Line)
		BroadcastLineEvent(line, Event_Line_Deleted)
		return nil
	}).ASync()

	bus.Register("line input changed", func(o any, a ...any) error {
		line := o.(*speaker.Line)
		// 通知输入格式
		BroadcastLineInputEvent(line)
		return nil
	}).ASync()
	speaker.BusLineVolumeChanged.Register(func(line *speaker.Line, oldVol float64) error {
		BroadcastLineEvent(line, Event_Line_Edited)
		return nil
	}).ASync()

}
