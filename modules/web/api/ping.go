package api

import "github.com/fasthttp/websocket"

func Ping(c *websocket.Conn, req any) {
	c.WriteMessage(websocket.TextMessage, []byte("pon pon pon"))

}
