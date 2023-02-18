package web

import (
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func startStaticServer(address string, root string) error {
	root, err := filepath.Abs(root + "/")
	if err != nil {
		return err
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return err
	}

	fsHandler := fasthttp.FSHandler(root, 0)

	log.Info("start http on " + address + " " + root)

	requestHandle := func(ctx *fasthttp.RequestCtx) {
		log.Info("request", zap.String("path", string(ctx.Path())), zap.String("src", ctx.RemoteAddr().String()))

		switch string(ctx.Path()) {
		case "/api":
			wsHandler(ctx)
		case "/status":
			statusHandler(ctx)
		default:
			fsHandler(ctx)
		}
	}

	fasthttp.SetBodySizePoolLimit(1024, 102400)

	err = fasthttp.ListenAndServe(address, requestHandle)
	if err != nil {
		log.Fatal("http start error", zap.Error(err))
		return err
	}

	return nil
}
