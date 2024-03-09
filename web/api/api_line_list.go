package api

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

func apiLineList(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	list := []*websockets.ResponseLineList{}

	for _, l := range speaker.LineList() {
		list = append(list, websockets.NewResponseLineList(l))
	}

	return list, nil
}
