package websockets

import (
	"context"

	"go.uber.org/zap"
)

const (
	Command_MIN uint8 = iota
	Command_SERVER
	Command_SPEAKER
	Command_LINE

	Command_MAX
)

const (
	Event_MIN uint8 = 10 + iota
	Event_SP_Detected
	Event_SP_Online
	Event_SP_Offline
	Event_SP_Deleted
	Event_SP_Moved
	Event_SP_Edited
	Event_SP_LevelMeter

	Event_Line_Created
	Event_Line_Deleted
	Event_Line_Edited
	Event_Line_LevelMeter
	Event_Line_Spectrum // 频谱图
	Event_Line_Input    // 有音频信号进入

	Event_SRV_Exited

	Event_MAX
)

type broadcastEvent struct {
	evt uint8
	arg int
}

var CommandEventMap = map[uint8][]uint8{
	Command_SPEAKER: {
		Event_SP_Deleted,
		Event_SP_Online,
		Event_SP_Offline,
		Event_SP_Detected,
		Event_SP_Moved,
		Event_SP_Edited,
	},
	Command_LINE: {
		Event_Line_Created,
		Event_Line_Deleted,
		Event_Line_Edited,
	},
	Command_SERVER: {},
}

func findEvent(es []uint8, e uint8) bool {
	for _, ee := range es {
		if ee == e {
			return true
		}
	}

	return false
}
func findBEvent(es []broadcastEvent, e uint8) bool {
	for _, ee := range es {
		if ee.evt == e {
			return true
		}
	}

	return false
}

// 事件开关回调
type EventHandler struct {
	On  func(evt uint8, arg int, ctx context.Context, log *zap.Logger)
	Off func(evt uint8, arg int)
}
