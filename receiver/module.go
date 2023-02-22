package receiver

import (
	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/decoder/pipeline"
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

	// 为默认线路添加音频工作流
	line := speaker.DefaultLine()
	pipeline.NewPipeLine(line)

	if config.EnableDLNA {
		dlnaInstance, err = dlna.NewDLNAServer(ctx, line.Name)
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
