package api

import (
	"time"

	"github.com/fasthttp/websocket"
	"go.uber.org/zap"
)

func apiPing(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
	c.WriteMessage(websocket.TextMessage, []byte("pong"))
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	return nil, nil
}
