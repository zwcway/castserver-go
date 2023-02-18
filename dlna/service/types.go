package service

import (
	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/utils"
)

type httpHandler func(ctx *fasthttp.RequestCtx)

type Controller struct {
	Url     string
	Handler httpHandler
}

type ServiceHandler struct {
	Id     string
	SCPD   Controller
	Contol Controller
	Event  Controller
}

type ServiceController interface {
	Init(lctx utils.Context) error
	Deinit()

	Handlers() *ServiceHandler
}
