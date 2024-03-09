package api

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

func apiSpeakerList(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	list := []*websockets.ResponseSpeakerItem{}

	speaker.All(func(s *speaker.Speaker) {
		list = append(list, websockets.NewResponseSpeakerItem(s))
	})

	return list, nil
}
