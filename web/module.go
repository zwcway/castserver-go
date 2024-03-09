package web

import (
	"github.com/zwcway/castserver-go/common/config"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/web/api"
	"github.com/zwcway/castserver-go/web/websockets"
)

var (
	log    lg.Logger
	Module = webModule{}
)

type webModule struct{}

func (webModule) Init(ctx utils.Context) error {
	log = ctx.Logger("web")

	websockets.Init(ctx)
	api.Init(ctx)
	websockets.ApiDispatch = api.ApiDispatch

	return nil
}

func (webModule) Start() error {
	err := startServer(&config.HTTPListen, config.HTTPRoot)

	return err
}

func (webModule) DeInit() {
	stopServer()
}
