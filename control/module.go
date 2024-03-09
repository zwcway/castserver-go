package control

import (
	"github.com/zwcway/castserver-go/common/bus"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
)

var (
	log lg.Logger
)

type controlModule struct{}

var Module = controlModule{}

func (controlModule) Init(ctx utils.Context) error {
	log = ctx.Logger("control")

	c := func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		ControlSample(sp)
		return nil
	}
	bus.Register("speaker connected", c)
	bus.Register("speaker format changed", c)

	bus.Register("speaker volume changed", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		ControlSpeakerVolume(sp, float64(sp.Volume), sp.Mute)
		return nil
	})
	return nil
}

func (controlModule) Start() error {
	return nil
}

func (controlModule) DeInit() {

}
