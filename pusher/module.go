package pusher

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"

	"go.uber.org/zap"
)

var (
	log     *zap.Logger
	context utils.Context
	Module  = pusherModule{}
)

type pusherModule struct{}

func (pusherModule) Init(ctx utils.Context) error {
	log = ctx.Logger("pusher")
	context = ctx

	bus.Register("speaker format changed", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)

		refreshPushQueue(sp, sp.EqualizerEle.Delay())
		return nil
	})
	bus.Register("speaker online", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		Disconnect(sp)
		Connect(sp)
		return nil
	}).ASync()
	bus.Register("speaker detected", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		Connect(sp)
		return nil
	}).ASync()
	receiveQueue = make(chan speaker.QueueData, config.ReadQueueSize)

	initTrigger()

	return nil
}

func (pusherModule) DeInit() {
	// 关闭所有设备的连接
	speaker.All(func(s *speaker.Speaker) {
		Disconnect(s)
	})
}
