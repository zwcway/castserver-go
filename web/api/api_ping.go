package api

import (
	"time"

	"github.com/fasthttp/websocket"
	"go.uber.org/zap"
)

func apiPing(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	if c == nil {
		return nil, nil
	}
	c.WriteMessage(websocket.TextMessage, []byte("pong"))
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	return nil, nil
}
