package receiver

import (
	config "github.com/zwcway/castserver-go/config"
	dlna "github.com/zwcway/castserver-go/dlna"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

var (
	log    *zap.Logger
	Module receiveModel

	dlnaInstance *dlna.DLNAServer
)

type receiveModel struct {
}

func (receiveModel) Init(ctx utils.Context) error {
	var err error

	log = ctx.Logger("receiver")

	if config.EnableDLNA {
		dlnaInstance, err = dlna.NewDLNAServer(ctx, "")
		if err != nil {
			return err
		}
		dlnaInstance.ListenAndServe()
	}
	return nil
}

func (receiveModel) DeInit() {
	dlnaInstance.Close()
}
