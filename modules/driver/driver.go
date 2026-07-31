package driver

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/modules/driver/local"
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

func initDriver() error {
	err := local.Init()
	if err != nil {
		log.Error("init local speaker failed. use timer", lg.Error(err))
		lastTrigger = trigger_timer
		return err
	} else {
		local.Play()
		// localspeaker.SetCallback(PushLineBuffer)
		lastTrigger = trigger_local
	}

	currentTrigger = lastTrigger

	speaker.BusLineCreated.Register(func(l *speaker.Line) error {
		return local.AddLine(l)
	})

	speaker.BusLineDeleted.Register(func(src, dst *speaker.Line) error {
		return local.RemoveLine(src, dst)
	})

	return nil
}

func Play() {

}

func Stop() {

}
func IsPlaying() bool {
	return local.IsPlaying()
}
func TriggerReceiverIn(b bool) {
	if b {
		currentTrigger = trigger_receiver
	} else {
		currentTrigger = lastTrigger
	}
}

func triggerAddLine(line *speaker.Line) {
	// e := NewElement(line)
	// line.Input.PipeLiner.Append(e)
	// e.On()

	timerAddLine(line)
}

func triggerRemoveLine(line *speaker.Line) {
	timerRemoveLine(line)
}
