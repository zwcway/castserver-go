package websockets

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
)

func register() {

	bus.Register("speaker detected", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		// 触发设备发现事件，通知管理后台
		BroadcastSpeakerEvent(sp, Event_SP_Detected)
		return nil
	}).ASync()
	bus.Register("speaker offline", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		BroadcastSpeakerEvent(sp, Event_SP_Offline)
		return nil
	}).ASync()
	bus.Register("speaker online", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		// 触发设备上线事件，通知管理后台
		BroadcastSpeakerEvent(sp, Event_SP_Online)
		return nil
	}).ASync()
	bus.Register("speaker channel moved", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		och := a[1].(audio.Channel)

		BroadcastSpeakerChannelMovedEvent(sp, och, sp.SampleChannel())
		return nil
	}).ASync()
	bus.Register("line created", func(a ...any) error {
		line := a[0].(*speaker.Line)
		BroadcastLineEvent(line, Event_Line_Created)
		return nil
	}).ASync()
	bus.Register("line deleted", func(a ...any) error {
		line := a[0].(*speaker.Line)
		BroadcastLineEvent(line, Event_Line_Deleted)
		return nil
	}).ASync()

	bus.Register("line input changed", func(a ...any) error {
		line := a[0].(*speaker.Line)
		// 通知输入格式
		BroadcastLineInputEvent(line)
		return nil
	}).ASync()
	bus.Register("line volume changed", func(a ...any) error {
		line := a[0].(*speaker.Line)

		BroadcastLineEvent(line, Event_Line_Edited)
		return nil
	}).ASync()

}
