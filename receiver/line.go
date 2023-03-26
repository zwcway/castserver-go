package receiver

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
)

func initDefaultLine() {
	bus.Register("line created", func(a ...any) error {
		line := a[0].(*speaker.Line)

		AddDLNA(line)
		return nil
	})
	bus.Register("line deleted", func(a ...any) error {
		line := a[0].(*speaker.Line)
		DelDLNA(line)
		return nil
	})
}
