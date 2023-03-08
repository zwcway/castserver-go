package api

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiPing(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	if c == nil {
		return nil, nil
	}
	c.Conn.WriteMessage(websocket.TextMessage, []byte("pong"))
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	return nil, nil
}
