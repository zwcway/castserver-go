package pusher

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/localspeaker"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"go.uber.org/zap"
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
		log.Error("init local speaker failed. use timer", zap.Error(err))
		lastTrigger = trigger_timer
	} else {
		localspeaker.Play()
		localspeaker.SetCallback(PushLineBuffer)
		lastTrigger = trigger_local
	}

	currentTrigger = lastTrigger
}

func TriggerReceiverIn(b bool) {
	if b {
		currentTrigger = trigger_receiver
	} else {
		currentTrigger = lastTrigger
	}
}

func TriggerAddLine(line *speaker.Line) {
	line.Pusher = NewElement(line)
	pipeline.NewPipeLine(line)
	localspeaker.AddLine(line)
	TimerAddLine(line)
}

func TriggerRemoveLine(line *speaker.Line) {
	localspeaker.RemoveLine(line)
	TimerRemoveLine(line)
	line.Input.PipeLine.Close()
}