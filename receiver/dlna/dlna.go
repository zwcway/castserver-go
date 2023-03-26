package dlna

import (
	"strings"

	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/receiver/dlna/service"
	upnp "github.com/zwcway/fasthttp-upnp"
	upnps "github.com/zwcway/fasthttp-upnp/service"
	"github.com/zwcway/fasthttp-upnp/ssdp"

	"go.uber.org/zap"
)

type DLNAServer struct {
	ctx utils.Context
	log *zap.Logger

	upnp *upnp.DeviceServer
	c    chan int
}

func (s *DLNAServer) ListenAndServe() {
	if s == nil || s.upnp == nil {
		return
	}

	go s.upnp.Serve()
}

func (s *DLNAServer) onError(err error) {
	if upnp.IsIPDenyError(err) {
		s.log.Warn("ip denied", zap.Error(err))
	} else if ssdp.IsRequestError(err) {
	} else {
		s.log.Error("error", zap.Error(err))
	}
}

func (s *DLNAServer) onInfo(err string) {
	if !strings.Contains(err, "request from") {
		s.log.Info(err)
	}
}

func (s *DLNAServer) Close() {
	if s == nil {
		return
	}
	s.c <- 1
	close(s.c)
	s.upnp.Close()
}

func (s *DLNAServer) AddNewInstance(name string, uuid string) string {
	if s == nil || s.upnp == nil {
		return ""
	}
	return s.upnp.AddServer(name, uuid, "")
}

func (s *DLNAServer) ChangeName(uuid string, newName string) {
	if s == nil || s.upnp == nil {
		return
	}
	s.upnp.AddServer(newName, uuid, "")
}

func (s *DLNAServer) DelInstance(uuid string) {
	if s == nil || s.upnp == nil {
		return
	}
	s.upnp.DelServer(uuid)
}

func (s *DLNAServer) newUPnPServer(ctx utils.Context) (err error) {
	s.upnp, err = upnp.NewDeviceServer(ctx)
	if err != nil {
		return
	}
	s.upnp.DeviceType = upnps.DeviceType_MediaRenderer
	s.upnp.Manufacturer = config.APPNAME
	s.upnp.ServerName = config.NameVersion()
	s.upnp.RootDescNamespaces = map[string]string{
		"xmlns:dlna": "urn:schemas-dlna-org:device-1-0",
	}
	s.upnp.ServiceList = service.NewServiceList(ctx)
	s.upnp.ListenInterface = config.DLNAListen.Iface
	s.upnp.ListenPort = config.DLNAListen.AddrPort.Port()
	s.upnp.DenyIps = config.DLNADenyIps
	s.upnp.AllowIps = config.DLNAAllowIps
	s.upnp.ErrorHandler = s.onError
	s.upnp.InfoHandler = s.onInfo

	// s.upnp.BeforeRequestHandle = func(ctx *fasthttp.RequestCtx) bool {
	// 	s.log.Info(ctx.RemoteAddr().String() + " " + string(ctx.Request.Body()))
	// 	return true
	// }
	// s.upnp.AfterRequestHandle = func(ctx *fasthttp.RequestCtx) bool {
	// 	s.log.Info(ctx.RemoteAddr().String() + " " + string(ctx.Response.Body()))
	// 	s.log.Info("----------------------------------------------------------------------------------")
	// 	return true
	// }
	return nil
}

func NewDLNAServer(ctx utils.Context) (s *DLNAServer, err error) {
	s = &DLNAServer{}
	s.ctx = ctx
	s.log = ctx.Logger("dlna")
	s.c = make(chan int, 1)

	err = s.newUPnPServer(ctx)
	if err != nil {
		return
	}

	err = s.upnp.Init()

	return
}
