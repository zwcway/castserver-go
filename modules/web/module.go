package web

import (
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

var (
	log    *zap.Logger
	Module = webModule{}
)

type webModule struct{}

func (webModule) Init(ctx utils.Context) error {
	log = ctx.Logger("web")

	err := startStaticServer(config.HTTPAddrPort, config.HTTPRoot)
	if err != nil {
		return err
	}

	return nil
}

func (webModule) DeInit() {

}
