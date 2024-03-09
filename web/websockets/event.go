package websockets

type Command uint8

const (
	Command_MIN Command = iota
	Command_SERVER
	Command_SPEAKER
	Command_LINE

	Command_MAX
)

type Event uint8

const (
	Event_MIN Event = 10 + iota
	Event_SP_Detected
	Event_SP_Online
	Event_SP_Offline
	Event_SP_Deleted
	Event_SP_Moved
	Event_SP_Edited
	Event_SP_Spectrum
	Event_SP_LevelMeter

	Event_Line_Created
	Event_Line_Deleted
	Event_Line_Edited
	Event_Line_Speaker
	Event_Line_Spectrum // 频谱图
	Event_Line_LevelMeter
	Event_Line_Input // 有音频信号进入

	Event_SRV_Exited

	Event_MAX
)

type broadcastEvent struct {
	evt Event
	sub Event
	arg int
}

var CommandEventMap = map[Command][]Event{
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
		Event_Line_Input,
	},
	Command_SERVER: {},
}

func isSpectrumEvent(e Event) bool {
	return e == Event_Line_Spectrum || e == Event_Line_LevelMeter || e == Event_SP_Spectrum || e == Event_SP_LevelMeter
}

func FindEvent(cmd Command, e Event) bool {
	es, ok := CommandEventMap[cmd]
	if !ok {
		return false
	}
	for _, ee := range es {
		if ee == e {
			return true
		}
	}

	return false
}

func findEvent(es []Event, e Event) bool {
	for _, ee := range es {
		if ee == e {
			return true
		}
	}

	return false
}
func findBEvent(es []broadcastEvent, evt Event, sub Event, arg int) bool {
	for _, ee := range es {
		if ee.evt == evt && ee.sub == sub && ee.arg == arg {
			return true
		}
	}

	return false
}

func hasSpectrumEvent(list broadcastMap) bool {
	for _, evts := range list {
		for _, e := range evts {
			if isSpectrumEvent(e.evt) {
				return true
			}
		}
	}
	return false
}
