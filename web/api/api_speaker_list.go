package api

import (
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

func apiSpeakerList(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
	list := []*websockets.ResponseSpeakerList{}

	speaker.All(func(s *speaker.Speaker) {
		list = append(list, websockets.NewResponseSpeakerList(s))
	})

	return list, nil
}
