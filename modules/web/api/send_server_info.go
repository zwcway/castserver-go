package api

import "github.com/fasthttp/websocket"

type reqSendServerInfo struct {
	ID uint32 `jp:"id"`
}

func SendServerInfo(c *websocket.Conn, req any) {
	switch req.(type) {
	case string:
	default:
		return
	}
}
