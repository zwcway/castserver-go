package pusher

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
)

var (
	log     log1.Logger
	context utils.Context
	Module  = pusherModule{}
)

type pusherModule struct{}

func (pusherModule) Init(ctx utils.Context) error {
	log = ctx.Logger("pusher")
	context = ctx

	bus.Register("speaker format changed", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)

		refreshPushQueue(sp, sp.EqualizerEle.Delay())
		return nil
	})
	speaker.BusSpeakerOnline.Register(func(sp *speaker.Speaker) error {
		Disconnect(sp)
		Connect(sp)
		return nil
	}).ASync()
	speaker.BusSpeakerDetected.Register(func(sp *speaker.Speaker) error {
		Connect(sp)
		return nil
	}).ASync()
	bus.Register("speaker reonline", func(o any, a ...any) error {
		sp := o.(*speaker.Speaker)
		Connect(sp)
		return nil
	}).ASync()
	speaker.BusLineRefresh.Register(func(line *speaker.Line) error {
		log.Debug("line output format changed", log1.String("line", line.LineName), log1.String("format", line.Output.String()))
		return nil
	}).ASync()
	receiveQueue = make(chan speaker.QueueData, config.ReadQueueSize)

	initTrigger()

	return nil
}

func (pusherModule) Start() error {
	return nil
}

func (pusherModule) DeInit() {
	// 关闭所有设备的连接
	speaker.All(func(s *speaker.Speaker) {
		Disconnect(s)
	})
}
