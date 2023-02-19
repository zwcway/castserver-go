package service

import (
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/fasthttp-upnp/avtransport1"
	"github.com/zwcway/fasthttp-upnp/service"
	"go.uber.org/zap"
)

var log *zap.Logger
var audioDecoder *decoder.Decoder

func NewServiceList(ctx utils.Context) []*service.Controller {
	log = ctx.Logger("dlna srv")

	audioDecoder = decoder.NewDecoder(ctx)
	
	return []*service.Controller{
		{
			ServiceName: avtransport1.NAME,
			Actions: []*service.Action{
				avtransport1.SetAVTransportURI(setAVTransportURIHandler),
				avtransport1.GetPositionInfo(avtGetPositionInfo),
				avtransport1.Play(avtPlay),
				avtransport1.Pause(avtPause),
				avtransport1.Stop(avtStop),
				avtransport1.Seek(avtSeek),
			},
		},
	}
}
