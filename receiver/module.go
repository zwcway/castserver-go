package receiver

import (
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/receiver/dlna"
)

var (
	ctx    utils.Context
	log    lg.Logger
	Module receiveModel

	dlnaInstance *dlna.DLNAServer
)

type receiveModel struct {
}

func (receiveModel) Init(uctx utils.Context) error {
	ctx = uctx
	log = ctx.Logger("receiver")

	registerBus()

	return nil
}

func (receiveModel) Start() error {
	err := initDlna()
	return err
}

func (receiveModel) DeInit() {
	dlnaInstance.Close()
}
