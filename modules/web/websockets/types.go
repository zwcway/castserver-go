package websockets

type Event uint
type Command uint

const (
	Command_NONE Command = iota
	Command_SERVER
	Command_SPEAKER
	Command_LINE
)

const (
	Event_NONE Event = iota
	Event_SP_DETECTED
	Event_SP_ONLINE
	Event_SP_OFFLINE
	Event_SP_DELETED
)
