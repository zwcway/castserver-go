package driver

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
)

var (
	log     lg.Logger
	context utils.Context
	Module  = driverModule{}
)

type driverModule struct{}

func (driverModule) Init(ctx utils.Context) error {
	log = ctx.Logger("pusher")
	context = ctx

	if err := initDriver(); err != nil {
		return err
	}

	return nil
}

func (driverModule) Start() error {
	return nil
}

func (driverModule) DeInit() {
}
