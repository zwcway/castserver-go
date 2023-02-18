package service

import (
	_ "embed"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/dlna/upnp"
	"go.uber.org/zap"
)

//go:embed scpd/avtransport.xml
var avtXML []byte

type aVTransport struct {
	actionXML []byte
	handlers  *ServiceHandler

	events upnp.Event
}

var AVTransport = aVTransport{
	actionXML: avtXML,
}

func (c *aVTransport) Init(log *zap.Logger) error {
	c.handlers = &ServiceHandler{
		Id: "urn:schemas-upnp-org:service:AVTransport",

		SCPD:   Controller{"/AVTransport1.xml", c.SCPDHandler},
		Contol: Controller{"/AVTransport/control", c.ControlHandler},
		Event:  Controller{"/AVTransport/event", c.EventHandler},
	}
	return nil
}
func (c *aVTransport) Deinit() {

}
func (c *aVTransport) Handlers() *ServiceHandler {
	return c.handlers
}

func (c *aVTransport) SCPDHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)
	ctx.Write(c.actionXML)
}
func (c *aVTransport) ControlHandler(ctx *fasthttp.RequestCtx) {

}
func (c *aVTransport) EventHandler(ctx *fasthttp.RequestCtx) {
	method := string(ctx.Method())
	switch method {
	case "SUBSCRIBE":
	}

}
