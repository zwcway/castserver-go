package api

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

func apiLineList(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	list := []responseLineInfo{}

	for id, l := range speaker.LineList() {
		list = append(list, responseLineInfo{
			ID:   uint8(id),
			Name: l.Name,
		})
	}

	return list, nil
}
