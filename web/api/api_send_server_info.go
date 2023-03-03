package api

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/detector"
	"go.uber.org/zap"
)

type reqSendServerInfo struct {
	ID uint32 `jp:"id"`
}

func apiSendServerInfo(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var spId uint32
	err := req.Unmarshal(&spId)
	if err != nil {
		return nil, err
	}

	sp := speaker.FindSpeakerByID(speaker.ID(spId))
	if sp == nil {
		return nil, nil
	}
	detector.SendServerInfo(sp)

	return nil, nil
}
