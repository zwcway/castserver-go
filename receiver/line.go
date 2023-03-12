package receiver

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/pusher"
)

func initDefaultLine() {
	defaultLine := speaker.DefaultLine()
	pusher.TriggerAddLine(defaultLine)
}

func AddLine(line *speaker.Line) {
	AddDLNA(line)
	pusher.TriggerAddLine(line)
}

func DelLine(line *speaker.Line) {
	pusher.TriggerRemoveLine(line)
	DelDLNA(line)
}
