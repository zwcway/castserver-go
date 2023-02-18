package dlna

import (
	"encoding/xml"
	"net"
	"net/url"

	config "github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/dlna/service"
	"github.com/zwcway/castserver-go/dlna/ssdp"
	"github.com/zwcway/castserver-go/dlna/upnp"
	utils "github.com/zwcway/castserver-go/utils"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var serviceList = []service.ServiceController{
	&service.AVTransport,
	&service.ConnectionManager,
	&service.RenderingControl,
	&service.MediaReceiverRegistrar,
}

type DLNAServer interface {
	ListenAndServe()
	Close()
}

type dlnaServer struct {
	ctx  utils.Context
	log  *zap.Logger
	conn net.Listener
	addr *net.TCPAddr
	name string
	uuid string

	rootDescXML []byte

	ssdpList []ssdp.SSDPServer
}

func (s *dlnaServer) makeServices() (srvs []upnp.Service) {
	for i := range serviceList {
		handle := serviceList[i].Handlers()
		srv := upnp.Service{
			ServiceType: handle.Id + ":1",
			ServiceId:   handle.Id,
			SCPDURL:     handle.SCPD.Url,
			ControlURL:  handle.Contol.Url,
			EventSubURL: handle.Event.Url,
		}

		srvs = append(srvs, srv)
	}

	return
}
func (s *dlnaServer) makeDevice() upnp.DeviceDesc {
	return upnp.DeviceDesc{
		NSDLNA:      "urn:schemas-dlna-org:device-1-0",
		NSSEC:       "http://www.sec.co.kr/dlna",
		SpecVersion: upnp.SpecVersion{Major: 1, Minor: 0},
		Device: upnp.Device{
			DeviceType:   "urn:schemas-upnp-org:device:MediaRenderer:1",
			FriendlyName: s.name,
			Manufacturer: config.APPNAME,
			ModelName:    config.NameVersion(),
			UDN:          s.uuid,
			ServiceList:  s.makeServices(),
		},
	}
}

func getListenAddress() string {
	if config.ReceiveInterface != nil {
		addr := utils.InterfaceAddr(config.ReceiveInterface, config.ReceiveUseIPV6)
		if addr != nil {
			return addr.IP.String()
		}
		return ""
	} else if config.ReceiveAddrPort.Port() > 0 {
		return config.ReceiveAddrPort.String()
	} else {
		return config.ReceiveAddrPort.String()
	}
}

func NewDLNAServer(ctx utils.Context) (s *dlnaServer, err error) {
	s = &dlnaServer{}
	s.ctx = ctx
	s.log = ctx.Logger("dlna")
	s.name = config.APPNAME + "-DLNA"
	s.uuid = utils.MakeUUID(s.name)

	for _, ss := range serviceList {
		ss.Init(s.ctx)
	}

	s.conn, err = net.Listen("tcp", getListenAddress())
	if err != nil {
		return
	}
	s.log.Info("listen on " + s.conn.Addr().String())

	s.addr = s.conn.Addr().(*net.TCPAddr)

	s.rootDescXML, err = xml.MarshalIndent(s.makeDevice(), " ", "  ")

	return
}

func (s *dlnaServer) Close() {
	for _, ss := range s.ssdpList {
		ss.Close()
	}
	for _, ss := range serviceList {
		ss.Deinit()
	}
	s.conn.Close()
}

func (s *dlnaServer) ListenAndServe() {
	s.startSSDP()

	server := fasthttp.Server{Handler: s.httpHandler}
	server.Serve(s.conn)
}

func (s *dlnaServer) httpHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetServer(config.NameVersion())
	// ctx.Response.Header.Set()
	uri := string(ctx.Path())

	for i := range serviceList {
		handle := serviceList[i].Handlers()
		switch uri {
		case handle.SCPD.Url:
			if handle.SCPD.Handler != nil {
				handle.SCPD.Handler(ctx)
			}
			return
		case handle.Contol.Url:
			if handle.Event.Handler != nil {
				handle.Event.Handler(ctx)
			}
			return
		case handle.Event.Url:
			if handle.Event.Handler != nil {
				handle.Event.Handler(ctx)
			}
			return
		}
	}

	if uri == scpdUrlRootDesc {
		ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)
		ctx.Write(s.rootDescXML)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
}

func (s *dlnaServer) makeSSDPLocation(ip net.IP) string {
	var host net.IP
	if s.addr.IP.IsUnspecified() {
		host = ip
	} else {
		host = s.addr.IP
	}
	url := url.URL{
		Scheme: "http",
		Host: (&net.TCPAddr{
			IP:   host,
			Port: s.addr.Port,
		}).String(),
		Path: scpdUrlRootDesc,
	}
	return url.String()
}

func (s *dlnaServer) startSSDP() {
	ifaces := make([]*net.Interface, 1)
	if config.ReceiveInterface != nil {
		ifaces = append(ifaces, config.ReceiveInterface)
	} else {
		ifaces = utils.Interfaces()
	}
	devices := []string{}

	services := []string{}
	for _, srv := range s.makeServices() {
		services = append(services, srv.ServiceType)
	}

	for _, iface := range ifaces {
		go func(iface *net.Interface) {
			ss, err := ssdp.NewSSDPServer(s.ctx, iface)
			if err != nil {
				return
			}
			s.ssdpList = append(s.ssdpList, ss)
			defer ss.Close()

			ss.Location = s.makeSSDPLocation
			ss.Server = "Linux/3.14.29 DLNADOC/1.50 UPnP/1.0 " + config.NameVersion()
			ss.UUID = s.uuid
			ss.Devices = devices
			ss.Services = services

			err = ss.ListenAndServe()
			if err != nil {
				s.log.Error("start ssdp failed on "+iface.Name, zap.Error(err))
				return
			}
		}(iface)
	}
}
