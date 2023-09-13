package pusher

import (
	"github.com/zwcway/castserver-go/common/bus"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/localspeaker"
)

const (
	trigger_local int = 1 + iota
	trigger_timer
	trigger_receiver
)

var (
	lastTrigger    = trigger_local
	currentTrigger = lastTrigger
)

func initTrigger() {
	err := localspeaker.Init()
	if err != nil {
		log.Error("init local speaker failed. use timer", log1.Error(err))
		lastTrigger = trigger_timer
	} else {
		localspeaker.Play()
		// localspeaker.SetCallback(PushLineBuffer)
		lastTrigger = trigger_local
	}

	currentTrigger = lastTrigger

	bus.Register("line created", func(o any, a ...any) error {
		line := a[0].(*speaker.Line)
		TriggerAddLine(line)
		return nil
	})
}

func TriggerReceiverIn(b bool) {
	if b {
		currentTrigger = trigger_receiver
	} else {
		currentTrigger = lastTrigger
	}
}

func TriggerAddLine(line *speaker.Line) {
	e := NewElement(line)
	line.Input.PipeLine.Append(e)
	e.On()

	TimerAddLine(line)
}

func TriggerRemoveLine(line *speaker.Line) {
	TimerRemoveLine(line)
}
