package receiver

import (
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/receiver/dlna"

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
