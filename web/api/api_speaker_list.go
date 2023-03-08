package api

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiSpeakerList(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	list := []*websockets.ResponseSpeakerList{}

	speaker.All(func(s *speaker.Speaker) {
		list = append(list, websockets.NewResponseSpeakerList(s))
	})

	return list, nil
}
