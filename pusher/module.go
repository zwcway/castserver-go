package pusher

import (
	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"

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

	// 初始化并启动发送队列
	queueList = make([]chan QueueData, config.SendRoutinesMax)
	for i := range queueList {
		queueList[i] = make(chan QueueData, config.SendQueueSize)
		go pushRoutine(&queueList[i])
	}

	receiveQueue = make(chan QueueData, config.ReadQueueSize)
	queueSpeaker = make(map[*speaker.Speaker]*chan QueueData)

	return nil
}

func (pusherModule) DeInit() {
	// 关闭所有设备的连接
	speaker.All(func(s *speaker.Speaker) {
		Disconnect(s)
	})
}
