package main

import (
	"github.com/zwcway/castserver-go/common"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/detector"
	"github.com/zwcway/castserver-go/mutexer"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/receiver"
	"github.com/zwcway/castserver-go/web"
)

type Module interface {
	Init(ctx utils.Context) error
	Start() error
	DeInit()
}

var mods = []Module{
	common.Module,
	mutexer.Module,
	detector.Module,
	pusher.Module,
	control.Module,
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
