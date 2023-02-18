package service

import (
	_ "embed"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/utils"
)

//go:embed scpd/connectionManager.xml
var rcXML []byte

type renderingControl struct {
	actionXML []byte
	handlers  *ServiceHandler
}

var RenderingControl = renderingControl{
	actionXML: rcXML,
}

func (c *renderingControl) Init(ctx utils.Context) error {

	c.handlers = &ServiceHandler{
		Id: "urn:schemas-upnp-org:service:RenderingControl",

		SCPD:   Controller{"/RenderingControl1.xml", c.SCPDHandler},
		Contol: Controller{"/RenderingControl/control", c.ControlHandler},
		Event:  Controller{"/RenderingControl/event", nil},
	}
	return nil
}

func (c *renderingControl) Deinit() {

}
func (c *renderingControl) Handlers() *ServiceHandler {
	return c.handlers
}

func (c *renderingControl) SCPDHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)
	ctx.Write(c.actionXML)
}
func (c *renderingControl) ControlHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)

}
