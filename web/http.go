package web

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"mime"
	"net"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

//go:embed public/*
var httpfs embed.FS

func startStaticServer(address string, root string) error {
	root, err := filepath.Abs(root + "/")
	if err != nil {
		return err
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return err
	}

	log.Info("start http on " + address)

	goWsIPID := []byte("***Go-WS-IP***")
	goWsPortID := []byte("***Go-WS-Port***")
	var listenAddr netip.AddrPort

	requestHandle := func(ctx *fasthttp.RequestCtx) {
		uri := string(ctx.Path())

		log.Info("request", zap.String("path", uri), zap.String("src", ctx.RemoteAddr().String()))

		switch uri {
		case "/api":
			websockets.WSHandler(ctx)
		case "/status":
			statusHandler(ctx)
		default:
			if uri == "/" {
				uri += "index.html"
			}
			if strings.Contains(uri, "/.") {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				return
			}

			uri = "public" + uri

			if strings.HasPrefix(uri, "public/js/index.") {
				js, err := httpfs.ReadFile(uri)
				if err != nil {
					ctx.SetStatusCode(fasthttp.StatusNotFound)
					return
				}
				js = bytes.Replace(js, goWsIPID, []byte(listenAddr.Addr().String()), 1)
				js = bytes.Replace(js, goWsPortID, []byte(fmt.Sprint(listenAddr.Port())), 1)
				ctx.Write(js)
				return
			}

			fp, err := httpfs.Open(uri)
			if err != nil {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				return
			}
			defer fp.Close()

			ctx.SetContentType(mime.TypeByExtension(path.Ext(uri)))
			io.Copy(ctx, fp)
		}
	}

	fasthttp.SetBodySizePoolLimit(1024, 512000)

	s := &fasthttp.Server{
		Handler: requestHandle,
	}
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("http listen error", zap.Error(err))
		return err
	}
	listenAddr = netip.MustParseAddrPort(ln.Addr().String())
	if listenAddr.Addr().IsUnspecified() {
		defaultAddr := utils.DefaultAddr()
		listenAddr = netip.AddrPortFrom(*defaultAddr, listenAddr.Port())
	}

	err = s.Serve(ln)
	if err != nil {
		log.Error("http start error", zap.Error(err))
		return err
	}

	return nil
}
