package service

import (
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
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
	Init(log *zap.Logger) error
	Deinit()

	Handlers() *ServiceHandler
}
