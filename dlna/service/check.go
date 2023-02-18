package service

import (
	"time"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/dlna/upnp"
	"github.com/zwcway/castserver-go/utils"
)

func checkSubscibe(e *upnp.Event, ctx *fasthttp.RequestCtx) error {
	sid := utils.MakeUUID(time.Now().String())
	var (
		err     error
		timeout string
	)

	ctx.Response.Header.Add("SID", "uuid:"+sid)

	defer func() {
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
		} else {
			ctx.Response.Header.Add("TIMEOUT", timeout)
		}
	}()

	callback := string(ctx.Request.Header.Peek("CALLBACK"))
	urls, err := upnp.ParseCallback(callback)
	if err != nil {
		return err
	}
	timeout = string(ctx.Request.Header.Peek("TIMEOUT"))
	t, err := upnp.ParseTimeout(timeout)
	if err != nil {
		return err
	}

	e.Subscribe(sid, urls, t)

	return nil
}
