package web

import (
	"text/template"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/config"
)

var statusTempl = template.Must(template.New("").Parse(statusHTML))

type tplClient struct {
	IP string
}
type tplData struct {
	Name    string
	Clients []tplClient
}

func statusHandler(ctx *fasthttp.RequestCtx) {
	if !ctx.IsGet() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}
	ctx.SetContentType("text/html; charset=utf-8")

	td := tplData{
		Name: config.NameVersion(),
	}

	for c, _ := range wsHub.clients {
		client := tplClient{
			IP: c.conn.RemoteAddr().String(),
		}
		td.Clients = append(td.Clients, client)
	}
	statusTempl.Execute(ctx, &td)
}

const statusHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<title>{{ .Name }} 服务状态</title>
</head>
<body>
<h1>客户端列表</h1>
<table>
	{{ range .Clients}}
	<tr>
	<td>{{ .IP }} </td>
	</tr>
	{{ end }}
</table>
</body>
</html>`
