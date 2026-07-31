package main

import (
	"github.com/zwcway/castserver-go/common"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/modules/controller"
	"github.com/zwcway/castserver-go/modules/decoder"
	"github.com/zwcway/castserver-go/modules/detector"
	"github.com/zwcway/castserver-go/modules/driver"
	"github.com/zwcway/castserver-go/modules/mutexer"
	"github.com/zwcway/castserver-go/modules/pusher"
	"github.com/zwcway/castserver-go/modules/receiver"
	"github.com/zwcway/castserver-go/modules/web"
)

type Module interface {
	Init(ctx utils.Context) error
	Start() error
	DeInit()
}

var mods = []Module{
	decoder.Module,
	mutexer.Module,
	detector.Module,
	driver.Module,
	pusher.Module,
	controller.Module,
	receiver.Module,
	web.Module,
}

func initModules(rootCtx utils.Context) (err error) {
	log = rootCtx.Logger("module")

	common.Init(rootCtx)

	for i := range mods {
		err = mods[i].Init(rootCtx)
		if err != nil {
			log.Error("module init error", lg.Error(err))
			return err
		}
	}
	return nil
}

func startModules() (err error) {
	common.Start()

	for i := range mods {
		err = mods[i].Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func deinitModules() {
	for i := len(mods) - 1; i >= 0; i-- {
		mods[i].DeInit()
	}
}
