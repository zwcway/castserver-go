package dlna

import (
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/dlna/service"
	"github.com/zwcway/castserver-go/utils"
	upnp "github.com/zwcway/fasthttp-upnp"
	upnps "github.com/zwcway/fasthttp-upnp/service"
	"github.com/zwcway/fasthttp-upnp/ssdp"

	"go.uber.org/zap"
)

type DLNAServer struct {
	ctx utils.Context
	log *zap.Logger

	defaultUUID string
	upnp        *upnp.DeviceServer
	c           chan int
}

func (s *DLNAServer) ListenAndServe() {
	if s.upnp == nil {
		return
	}

	go s.upnp.Serve()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.c:
			return
		case err := <-s.upnp.ErrorChan:
			if upnp.IsIPDenyError(err) {
				s.log.Warn("ip denied", zap.Error(err))
			} else if ssdp.IsRequestError(err) {
			} else {
				s.log.Error("error", zap.Error(err))
			}
		case err := <-s.upnp.InfoChan:
			s.log.Info(err)
		}
	}
}

func (s *DLNAServer) Close() {
	s.c <- 1
	close(s.c)
	s.upnp.Close()
}

func (s *DLNAServer) AddNewInstance(name string) string {
	return s.upnp.AddServer(name, "", "")
}

func (s *DLNAServer) ChangeName(uuid string, newName string) {
	s.upnp.AddServer(newName, uuid, "")
}

func (s *DLNAServer) DelInstance(uuid string) {
	s.upnp.DelServer(uuid)
}

func (s *DLNAServer) newUPnPServer(ctx utils.Context, name string) (err error) {
	s.upnp, s.defaultUUID, err = upnp.NewDeviceServer(ctx, name)
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
	s.upnp.ListenInterface = config.DLNAInterface
	s.upnp.ListenPort = config.DLNAAddrPort.Port()
	s.upnp.DenyIps = config.DLNADenyIps
	s.upnp.AllowIps = config.DLNAAllowIps
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

func NewDLNAServer(ctx utils.Context, name string) (s *DLNAServer, uuid string, err error) {
	if name == "" {
		name = config.APPNAME
	}
	s = &DLNAServer{}
	s.ctx = ctx
	s.log = ctx.Logger("dlna")
	s.c = make(chan int, 1)

	err = s.newUPnPServer(ctx, name)
	if err != nil {
		return
	}

	uuid = s.defaultUUID

	err = s.upnp.Init()

	return
}
