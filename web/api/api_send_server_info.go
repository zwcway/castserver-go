package api

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/detector"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiSendServerInfo(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
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
