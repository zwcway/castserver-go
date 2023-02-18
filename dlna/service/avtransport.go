package service

import (
	_ "embed"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/dlna/upnp"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

//go:embed scpd/avtransport.xml
var avtXML []byte

type aVTransport struct {
	ctx       utils.Context
	log       *zap.Logger
	actionXML []byte
	handlers  *ServiceHandler

	events upnp.Event
}

var AVTransport = aVTransport{
	actionXML: avtXML,
}

func (c *aVTransport) Init(ctx utils.Context) error {
	c.ctx = ctx
	c.log = ctx.Logger("dlna aVTransport")
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
		err := checkSubscibe(&c.events, ctx)
		if err != nil {
			c.log.Error("subscribe invalid", zap.Error(err), zap.String("request", ctx.Request.Header.String()))
		}
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}
