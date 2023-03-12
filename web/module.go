package web

import (
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/castserver-go/web/api"
	"github.com/zwcway/castserver-go/web/websockets"

	"go.uber.org/zap"
)

var (
	log    *zap.Logger
	Module = webModule{}
)

type webModule struct{}

func (webModule) Init(ctx utils.Context) error {
	log = ctx.Logger("web")

	websockets.Init(ctx)
	api.Init(ctx)
	websockets.ApiDispatch = api.ApiDispatch

	err := startStaticServer(&config.HTTPListen, config.HTTPRoot)
	if err != nil {
		return err
	}

	return nil
}

func (webModule) DeInit() {

}
