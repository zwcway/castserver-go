package pusher

import (
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

	queueList = make([]chan sendQueue, config.SendRoutinesMax)
	for i := range queueList {
		queueList[i] = make(chan sendQueue, config.SendQueueSize)
	}

	return nil
}

func (pusherModule) DeInit() {

}
