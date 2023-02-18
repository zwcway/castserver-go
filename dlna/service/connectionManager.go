package service

import (
	_ "embed"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

//go:embed scpd/connectionManager.xml
var cmXML []byte

type connectionManager struct {
	actionXML []byte
	handlers  *ServiceHandler
}

var ConnectionManager = connectionManager{
	actionXML: cmXML,
}

func (c *connectionManager) Init(log *zap.Logger) error {

	c.handlers = &ServiceHandler{
		Id: "urn:schemas-upnp-org:service:ConnectionManager",

		SCPD:   Controller{"/ConnectionManager1.xml", c.SCPDHandler},
		Contol: Controller{"/ConnectionManager/control", c.ControlHandler},
		Event:  Controller{"/ConnectionManager/event", nil},
	}
	return nil
}

func (c *connectionManager) Deinit() {

}
func (c *connectionManager) Handlers() *ServiceHandler {
	return c.handlers
}

func (c *connectionManager) SCPDHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)
	ctx.Write(c.actionXML)
}
func (c *connectionManager) ControlHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(`text/xml; charset="utf-8"`)

}
