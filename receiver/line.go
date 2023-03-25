package receiver

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/pusher"
)

func initDefaultLine() {
	defaultLine := speaker.DefaultLine()
	pusher.TriggerAddLine(defaultLine)

	bus.Register("line created", func(a ...any) error {
		line := a[0].(*speaker.Line)

		AddDLNA(line)
		pusher.TriggerAddLine(line)
		return nil
	})
	bus.Register("line deleted", func(a ...any) error {
		line := a[0].(*speaker.Line)
		pusher.TriggerRemoveLine(line)
		DelDLNA(line)
		return nil
	})
}
