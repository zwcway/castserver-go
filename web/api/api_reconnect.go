package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/pusher"
	"go.uber.org/zap"
)

type requestReconnect uint32

func apiReconnect(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var sp requestReconnect
	err := req.Unmarshal(&sp)
	if err != nil {
		return nil, &Error{1, err}
	}
	s := speaker.FindSpeakerByID(speaker.ID(sp))
	if s == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", sp)}
	}
	pusher.Disconnect(s)
	pusher.Connect(s)
	return nil, nil
}
