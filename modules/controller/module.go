package controller

import (
	"github.com/zwcway/castserver-go/common/speaker"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

var (
	log *zap.Logger
)

func send(sp *speaker.Speaker) {

}

func syncTime() {

}

type controlModule struct{}

var Module = controlModule{}

func (controlModule) Init(ctx utils.Context) error {
	log = ctx.Logger("control")

	return nil
}

func (controlModule) DeInit() {

}
