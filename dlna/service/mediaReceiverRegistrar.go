package service

import (
	_ "embed"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/utils"
)

//go:embed scpd/connectionManager.xml
var mrrXML []byte

type mediaReceiverRegistrar struct {
	actionXML []byte
	handlers  *ServiceHandler
}

var MediaReceiverRegistrar = mediaReceiverRegistrar{
	actionXML: mrrXML,
}

func (c *mediaReceiverRegistrar) Init(ctx utils.Context) error {

	c.handlers = &ServiceHandler{
		Id: "urn:schemas-upnp-org:service:X-MS-MediaReceiverRegistrar",

		SCPD:   Controller{"/MediaReceiverRegistrar1.xml", c.SCPDHandler},
		Contol: Controller{"/MediaReceiverRegistrar/control", c.ControlHandler},
		Event:  Controller{"/MediaReceiverRegistrar/event", nil},
	}
	return nil
}

func (c *mediaReceiverRegistrar) Deinit() {

}
func (c *mediaReceiverRegistrar) Handlers() *ServiceHandler {
	return c.handlers
}

func (c *mediaReceiverRegistrar) SCPDHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)
	ctx.Write(c.actionXML)
}
func (c *mediaReceiverRegistrar) ControlHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)

}
