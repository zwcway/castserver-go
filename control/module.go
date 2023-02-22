package control

import (
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

var (
	log *zap.Logger
)

type controlModule struct{}

var Module = controlModule{}

func (controlModule) Init(ctx utils.Context) error {
	log = ctx.Logger("control")

	return nil
}

func (controlModule) DeInit() {

}
