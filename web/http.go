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
	"github.com/zwcway/castserver-go/common/config"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/web/api"
	"github.com/zwcway/castserver-go/web/websockets"
)

//go:embed public/*
var httpfs embed.FS

var (
	goWsIPID   = []byte("***Go-WS-IP***")
	goWsPortID = []byte("***Go-WS-Port***")
	listenAddr netip.AddrPort
	conn       net.Listener
)

func startServer(listen *config.Interface, root string) error {
	root, err := filepath.Abs(root + "/")
	if err != nil {
		return err
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return err
	}

	fasthttp.SetBodySizePoolLimit(1024, 512000)

	s := &fasthttp.Server{
		Handler: requestHandle,
	}
	conn, err = net.Listen("tcp", listen.AddrPort.String())
	if err != nil {
		log.Error("http listen error", lg.Error(err))
		return err
	}

	log.Info("start http on " + conn.Addr().String())

	listenAddr = netip.MustParseAddrPort(conn.Addr().String())

	if listenAddr.Addr().IsUnspecified() {
		defaultAddr := utils.DefaultAddr()
		listenAddr = netip.AddrPortFrom(*defaultAddr, listenAddr.Port())
	}

	log.Info("you can open http://" + listenAddr.String() + " in your web browser.")

	go s.Serve(conn)

	return nil
}

func stopServer() error {
	if conn == nil {
		return nil
	}

	return conn.Close()
}

func requestHandle(ctx *fasthttp.RequestCtx) {
	uri := string(ctx.Path())

	defer func() {
		log.Info("request",
			lg.Int("status", int64(ctx.Response.StatusCode())),
			lg.String("path", uri),
			lg.String("src", ctx.RemoteAddr().String()))
	}()

	if api.ApiDispatchDevel(ctx) {
		return
	}
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
