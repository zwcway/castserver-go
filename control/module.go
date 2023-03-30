package control

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/lg"
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

	bus.Registers(func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		ControlSample(sp)
		return nil
	}, "speaker connected", "speaker format changed")

	bus.Register("speaker volume changed", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
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
