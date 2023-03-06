package api

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiLineList(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	list := []*websockets.ResponseLineList{}

	for _, l := range speaker.LineList() {
		list = append(list, websockets.NewResponseLineList(l))
	}

	return list, nil
}
