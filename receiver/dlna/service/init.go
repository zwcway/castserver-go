package service

import (
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/fasthttp-upnp/avtransport1"
	"github.com/zwcway/fasthttp-upnp/service"
)

var log lg.Logger

func NewServiceList(ctx utils.Context) []*service.Controller {
	log = ctx.Logger("dlna srv")

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
