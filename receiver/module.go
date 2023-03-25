package receiver

import (
	"github.com/zwcway/castserver-go/receiver/dlna"
	"github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

var (
	ctx    utils.Context
	log    *zap.Logger
	Module receiveModel

	dlnaInstance *dlna.DLNAServer
)

type receiveModel struct {
}

func (receiveModel) Init(uctx utils.Context) error {
	ctx = uctx
	log = ctx.Logger("receiver")

	initDefaultLine()
	err := initDlna()

	return err
}

func (receiveModel) DeInit() {
	dlnaInstance.Close()
}
