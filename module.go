package main

import (
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/modules/controller"
	"github.com/zwcway/castserver-go/modules/decoder"
	"github.com/zwcway/castserver-go/modules/detector"
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
	pusher.Module,
	controller.Module,
	receiver.Module,
	web.Module,
}

func initModules(rootCtx utils.Context) (err error) {
	for i := 0; i < len(mods); i++ {
		err = mods[i].Init(rootCtx)
		if err != nil {
			return err
		}
	}
	return nil
}

func startModules() (err error) {
	for i := 0; i < len(mods); i++ {
		err = mods[i].Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func deinitModules() {
	for i := 0; i < len(mods); i++ {
		mods[i].DeInit()
	}
}
